package cache

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(options)

	// Ping the Redis server to ensure the connection is established
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis server: %w", err)
	}

	return client, nil
}
