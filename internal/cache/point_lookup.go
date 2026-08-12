package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/paucio/LocationGrid-go/internal/model"
)

type PointLookup struct {
	client *redis.Client
}

func NewPointLookup(client *redis.Client) *PointLookup {
	return &PointLookup{
		client: client,
	}
}

func (pl *PointLookup) GetPointByCoordinates(ctx context.Context, x, y float64) ([]int64, error) {
	cellX, cellY := CellForCoordinates(x, y)
	key := RedisKey(cellX, cellY)

	raw, err := pl.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No points found for this cell
		}
		return nil, fmt.Errorf("failed to get points from Redis: %w", err)
	}

	var pointIDs []int64
	if err := json.Unmarshal([]byte(raw), &pointIDs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal point IDs: %w", err)
	}

	return pointIDs, nil
}

func (pl *PointLookup) AddIDsBulk(ctx context.Context, points []model.Point) error {
	if len(points) == 0 {
		return nil
	}

	idsByCell := make(map[string][]interface{})
	for _, point := range points {
		cellX, cellY := CellForCoordinates(point.X, point.Y)
		key := RedisKey(cellX, cellY)
		idsByCell[key] = append(idsByCell[key], point.ID)
	}

	pipeline := pl.client.Pipeline()
	for key, ids := range idsByCell {
		pipeline.RPush(ctx, key, ids...)
	}
	_, err := pipeline.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute Redis pipeline: %w", err)
	}

	return nil
}
