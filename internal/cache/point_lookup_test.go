package cache

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func newTestPointLookup(t *testing.T) (*PointLookup, *miniredis.Miniredis) {
	t.Helper()

	redisServer, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	return NewPointLookup(client), redisServer
}

func TestGetPointByCoordinatesReturnsNilWhenNoPoints(t *testing.T) {
	pl, redisServer := newTestPointLookup(t)
	defer redisServer.Close()

	pointIDs, err := pl.GetPointByCoordinates(context.Background(), 124.5, 25.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pointIDs != nil {
		t.Fatalf("expected nil point IDs, got %v", pointIDs)
	}
}

func TestGetPointByCoordinatesReturnsPointIDs(t *testing.T) {
	pl, redisServer := newTestPointLookup(t)
	defer redisServer.Close()

	x, y := 124.5, 150.1
	cellX, cellY := CellForCoordinates(x, y)
	key := RedisKey(cellX, cellY)

	expected := []int64{101, 202, 303}
	encoded, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("failed to marshal expected point IDs: %v", err)
	}

	redisServer.Set(key, string(encoded))

	pointIDs, err := pl.GetPointByCoordinates(context.Background(), x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(pointIDs, expected) {
		t.Fatalf("expected %v, got %v", expected, pointIDs)
	}
}

func TestGetPointByCoordinatesReturnsErrorForInvalidJSON(t *testing.T) {
	pl, redisServer := newTestPointLookup(t)
	defer redisServer.Close()

	x, y := 124.5, 150.1
	cellX, cellY := CellForCoordinates(x, y)
	key := RedisKey(cellX, cellY)
	redisServer.Set(key, "not-json")

	_, err := pl.GetPointByCoordinates(context.Background(), x, y)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal point IDs") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
