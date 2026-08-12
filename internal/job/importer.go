package job

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/paucio/LocationGrid-go/internal/model"
)

const batchSize = 1000

type bulkInserter interface {
	BulkInsert(ctx context.Context, points []model.Point) ([]model.Point, error)
}

type RedisBulkInserter interface {
	AddIDsBulk(ctx context.Context, points []model.Point) error
}

type Importer struct {
	repo  bulkInserter
	index RedisBulkInserter
}

func NewImporter(bulkInserter bulkInserter, redisBulkInserter RedisBulkInserter) *Importer {
	return &Importer{
		repo:  bulkInserter,
		index: redisBulkInserter,
	}
}

type Stats struct {
	TotalPointsImported int
	SkippedPoints       int
}

func (i *Importer) ImportPoints(ctx context.Context, url string) (Stats, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Stats{}, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Stats{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Stats{}, fmt.Errorf("failed to fetch points: %s", resp.Status)
	}

	reader := csv.NewReader(resp.Body)
	reader.FieldsPerRecord = -1 // Allow variable number of fields per record

	var stats Stats
	batch := make([]model.Point, 0, batchSize) // Batch size of 1000

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return stats, fmt.Errorf("failed to read record: %w", err)
		}

		if len(record) < 4 {
			stats.SkippedPoints++
			continue // Skip invalid records
		}

		point, err := parseRow(record)
		if err != nil {
			stats.SkippedPoints++
			continue // Skip invalid records
		}

		batch = append(batch, point)
		if len(batch) == batchSize {
			if err := i.flush(ctx, batch); err != nil {
				return stats, fmt.Errorf("failed to insert batch: %w", err)
			}
			stats.TotalPointsImported += len(batch)
			batch = batch[:0] // Reset the batch
		}
	}

	// Insert any remaining points
	if len(batch) > 0 {
		if err := i.flush(ctx, batch); err != nil {
			return stats, fmt.Errorf("failed to insert remaining points: %w", err)
		}
		stats.TotalPointsImported += len(batch)
	}

	return stats, nil
}

func (i *Importer) flush(ctx context.Context, batch []model.Point) error {
	_, err := i.repo.BulkInsert(ctx, batch)
	if err != nil {
		return fmt.Errorf("failed to insert batch: %w", err)
	}

	if err := i.index.AddIDsBulk(ctx, batch); err != nil {
		return fmt.Errorf("failed to add IDs to Redis: %w", err)
	}

	return nil
}
