package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"

	"github.com/paucio/LocationGrid-go/internal/model"
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

func TestAddIDsBulkNoopForEmptyPoints(t *testing.T) {
	pl, redisServer := newTestPointLookup(t)
	defer redisServer.Close()

	if err := pl.AddIDsBulk(context.Background(), []model.Point{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(redisServer.Keys()) != 0 {
		t.Fatalf("expected no keys to be written, got %v", redisServer.Keys())
	}
}

func TestAddIDsBulkPushesIDsToSameCell(t *testing.T) {
	pl, redisServer := newTestPointLookup(t)
	defer redisServer.Close()

	points := []model.Point{
		{ID: 101, X: 124.5, Y: 150.1},
		{ID: 202, X: 130.0, Y: 155.0},
	}

	if err := pl.AddIDsBulk(context.Background(), points); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cellX, cellY := CellForCoordinates(points[0].X, points[0].Y)
	key := RedisKey(cellX, cellY)

	got, err := redisServer.List(key)
	if err != nil {
		t.Fatalf("failed to read list %q: %v", key, err)
	}

	expected := []string{fmt.Sprint(points[0].ID), fmt.Sprint(points[1].ID)}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestAddIDsBulkPushesIDsToDistinctCells(t *testing.T) {
	pl, redisServer := newTestPointLookup(t)
	defer redisServer.Close()

	points := []model.Point{
		{ID: 101, X: 10.0, Y: 10.0},
		{ID: 202, X: 5000.0, Y: 5000.0},
	}

	if err := pl.AddIDsBulk(context.Background(), points); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range points {
		cellX, cellY := CellForCoordinates(p.X, p.Y)
		key := RedisKey(cellX, cellY)

		got, err := redisServer.List(key)
		if err != nil {
			t.Fatalf("failed to read list %q: %v", key, err)
		}

		expected := []string{fmt.Sprint(p.ID)}
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("expected %v, got %v", expected, got)
		}
	}
}

func TestAddIDsBulkReturnsErrorWhenPipelineFails(t *testing.T) {
	pl, redisServer := newTestPointLookup(t)
	redisServer.Close()

	points := []model.Point{{ID: 101, X: 124.5, Y: 150.1}}

	err := pl.AddIDsBulk(context.Background(), points)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to execute Redis pipeline") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
