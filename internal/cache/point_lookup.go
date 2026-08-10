package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
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