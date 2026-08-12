package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paucio/LocationGrid-go/internal/cache"
	"github.com/paucio/LocationGrid-go/internal/job"
	"github.com/paucio/LocationGrid-go/internal/repository"
)

func main() {
	url := flag.String("url", "", "URL of the CSV file to import")
	flag.Parse()

	if *url == "" {
		log.Fatal("Please provide a URL using the -url flag.")
	}

	ctx := context.Background()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		log.Fatal("Please set the DATABASE_DSN environment variable.")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("Please set the REDIS_ADDR environment variable.")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	redisClient, err := cache.NewRedisClient(ctx, redisAddr)
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	repo := repository.NewPointRepository(pool)
	lookup := cache.NewPointLookup(redisClient)

	importer := job.NewImporter(repo, lookup)

	stats, err := importer.ImportPoints(ctx, *url)
	if err != nil {
		log.Fatalf("Import failed: %v", err)
	}

	log.Printf("Import completed. Total points imported: %d, Skipped points: %d", stats.TotalPointsImported, stats.SkippedPoints)
}
