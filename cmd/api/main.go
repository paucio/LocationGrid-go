package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/paucio/LocationGrid-go/internal/cache"
	"github.com/paucio/LocationGrid-go/internal/handler"
	"github.com/paucio/LocationGrid-go/internal/repository"
)

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}
	defer pool.Close()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR environment variable is not set")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	pointRepo := repository.NewPointRepository(pool)
	pointLookup := cache.NewPointLookup(redisClient)

	searchHandler := handler.NewSearchHandler(pointLookup, pointRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("/search", searchHandler.Search)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
