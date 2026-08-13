package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/paucio/LocationGrid-go/internal/model"
)

const (
	maxRadius = 20
)

type PointLookup struct {
	client *redis.Client
}

func NewPointLookup(client *redis.Client) *PointLookup {
	return &PointLookup{
		client: client,
	}
}

func (pl *PointLookup) GetPointByCoordinates(ctx context.Context, x, y float64, pointType string, limit int) ([]int64, error) {
	cellX, cellY := CellForCoordinates(x, y)

	var pointIDs []int64

	for i := 0; i <= maxRadius; i++ {
		for dx := -i; dx <= i; dx++ {
			for dy := -i; dy <= i; dy++ {
				if abs(dx) != i && abs(dy) != i {
					continue // Skip inner cells, only check the border of the square
				}

				neighborKey := RedisKey(cellX+dx, cellY+dy, pointType)
				raw, err := pl.client.Get(ctx, neighborKey).Result()
				if err != nil {
					if err == redis.Nil {
						continue // No points found for this cell
					}
					return nil, fmt.Errorf("failed to get points from Redis: %w", err)
				}

				var neighborPointIDs []int64
				if err := json.Unmarshal([]byte(raw), &neighborPointIDs); err != nil {
					return nil, fmt.Errorf("failed to unmarshal point IDs: %w", err)
				}

				pointIDs = append(pointIDs, neighborPointIDs...)
			}
		}
		if len(pointIDs) >= limit {
			break // Stop searching if we found the requested number of points
		}
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
		key := RedisKey(cellX, cellY, point.Type)
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
