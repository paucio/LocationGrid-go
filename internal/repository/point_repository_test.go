package repository

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/pashagolub/pgxmock/v4"

	"github.com/paucio/LocationGrid-go/internal/model"
)

func newTestPointRepository(t *testing.T) (*PointRepository, pgxmock.PgxPoolIface) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock pool: %v", err)
	}
	t.Cleanup(mock.Close)

	return &PointRepository{pool: mock}, mock
}

func TestFindByIdsReturnsEmptySliceForEmptyIds(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	points, err := repo.FindByIds(context.Background(), []int64{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if points == nil || len(points) != 0 {
		t.Fatalf("expected empty slice, got %v", points)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected query executed: %v", err)
	}
}

func TestFindByIdsReturnsPoints(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	ids := []int64{1, 2}
	rows := pgxmock.NewRows([]string{"id", "name", "x", "y"}).
		AddRow(int64(1), "Alpha", 1.5, 2.5).
		AddRow(int64(2), "Beta", 3.5, 4.5)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,name, x, y")).
		WithArgs(ids).
		WillReturnRows(rows)

	points, err := repo.FindByIds(context.Background(), ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []*model.Point{
		{ID: 1, Name: "Alpha", X: 1.5, Y: 2.5},
		{ID: 2, Name: "Beta", X: 3.5, Y: 4.5},
	}
	if !reflect.DeepEqual(points, expected) {
		t.Fatalf("expected %+v, got %+v", expected, points)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindByIdsReturnsErrorWhenQueryFails(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	ids := []int64{1}
	queryErr := errors.New("boom")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,name, x, y")).
		WithArgs(ids).
		WillReturnError(queryErr)

	points, err := repo.FindByIds(context.Background(), ids)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, queryErr) {
		t.Fatalf("expected wrapped query error, got %v", err)
	}
	if points != nil {
		t.Fatalf("expected nil points, got %v", points)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindByIdsReturnsErrorOnScanFailure(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	ids := []int64{1}
	rows := pgxmock.NewRows([]string{"id", "name", "x", "y"}).
		AddRow("not-an-id", "Alpha", 1.5, 2.5)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,name, x, y")).
		WithArgs(ids).
		WillReturnRows(rows)

	points, err := repo.FindByIds(context.Background(), ids)
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}
	if points != nil {
		t.Fatalf("expected nil points, got %v", points)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindByIdsReturnsEmptySliceWhenNoRowsMatch(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	ids := []int64{999}
	rows := pgxmock.NewRows([]string{"id", "name", "x", "y"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,name, x, y")).
		WithArgs(ids).
		WillReturnRows(rows)

	points, err := repo.FindByIds(context.Background(), ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(points) != 0 {
		t.Fatalf("expected no points, got %v", points)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
