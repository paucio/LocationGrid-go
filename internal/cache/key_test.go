package cache

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
)

func TestRedisKeyFormatsCoordinates(t *testing.T) {
	got := RedisKey(2, -3, "shop")
	want := "point_shop_x:2.000000:y:-3.000000"

	if got != want {
		t.Fatalf("RedisKey() = %q, want %q", got, want)
	}
}

func TestCellForCoordinatesPositiveAndNegative(t *testing.T) {
	cases := []struct {
		x, y         float64
		wantX, wantY int
	}{
		{124.5, 25.0, 2, 0},
		{-124.5, 25.0, -3, 0},
		{25.0, -124.5, 0, -3},
		{-124.5, -124.5, -3, -3},
	}

	for _, tc := range cases {
		gotX, gotY := CellForCoordinates(tc.x, tc.y)
		if gotX != tc.wantX || gotY != tc.wantY {
			t.Fatalf("CellForCoordinates(%v, %v) = (%d, %d), want (%d, %d)", tc.x, tc.y, gotX, gotY, tc.wantX, tc.wantY)
		}
	}
}

func TestNewRedisClientParsesURLAndPings(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer server.Close()

	redisURL := "redis://" + server.Addr()
	client, err := NewRedisClient(context.Background(), redisURL)
	if err != nil {
		t.Fatalf("NewRedisClient() unexpected error: %v", err)
	}
	defer client.Close()
}

func TestNewRedisClientInvalidURL(t *testing.T) {
	_, err := NewRedisClient(context.Background(), "not-a-valid-url")
	if err == nil {
		t.Fatal("expected error for invalid Redis URL, got nil")
	}
}
