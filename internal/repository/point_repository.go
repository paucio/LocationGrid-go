package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/paucio/LocationGrid-go/internal/model"
)

// Querier is satisfied by *pgxpool.Pool; it exists so tests can substitute a mock.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type PointRepository struct {
	pool Querier
}

func NewPointRepository(pool *pgxpool.Pool) *PointRepository {
	return &PointRepository{pool: pool}
}

func (r *PointRepository) FindByIds(ctx context.Context, ids []int64) ([]*model.Point, error){
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