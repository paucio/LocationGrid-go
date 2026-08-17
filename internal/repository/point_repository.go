package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/paucio/LocationGrid-go/internal/model"
)

// Querier is satisfied by *pgxpool.Pool; it exists so tests can substitute a mock.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arg ...any) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

type PointRepository struct {
	pool Querier
}

func NewPointRepository(pool *pgxpool.Pool) *PointRepository {
	return &PointRepository{pool: pool}
}

func (r *PointRepository) FindByIds(ctx context.Context, ids []int64) ([]*model.Point, error) {
	if len(ids) == 0 {
		return []*model.Point{}, nil
	}

	const query = `
		SELECT id,name, x, y
		FROM points
		WHERE id = ANY($1)
	`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var points []*model.Point
	for rows.Next() {
		var p model.Point
		if err := rows.Scan(&p.ID, &p.Name, &p.X, &p.Y); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		points = append(points, &p)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", rows.Err())
	}

	return points, nil
}

func (r *PointRepository) BulkInsert(ctx context.Context, points []model.Point) ([]model.Point, error) {
	if len(points) == 0 {
		return points, nil
	}

	const query = `
		INSERT INTO points (name, x, y, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	batch := &pgx.Batch{}
	for _, p := range points {
		batch.Queue(query, p.Name, p.X, p.Y, p.Type)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	insertedPoints := make([]model.Point, len(points))
	for i, p := range points {
		if err := br.QueryRow().Scan(&p.ID); err != nil {
			return nil, fmt.Errorf("failed to scan inserted point: %w", err)
		}
		insertedPoints[i] = p
	}

	if len(insertedPoints) == 0 {
		return nil, fmt.Errorf("no points were inserted")
	}

	return insertedPoints, nil
}

func (r *PointRepository) FindByID(ctx context.Context, x, y float64) (*model.Point, error) {
	const query = `
		SELECT id,name, x, y
		FROM points
		WHERE x = $1 AND y = $2
	`

	var p model.Point
	err := r.pool.QueryRow(ctx, query, x, y).Scan(&p.ID, &p.Name, &p.X, &p.Y)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query point: %w", err)
	}

	return &p, nil
}

func (r *PointRepository) CreatePoint(ctx context.Context, p *model.Point) error {
	const query = `
		INSERT INTO points(name, x, y, type)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, &p.Name, &p.X, &p.Y, &p.Type)
	if err != nil {
		return fmt.Errorf("failed to create point: %w", err)
	}

	return nil
}
