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

const bulkInsertQuery = `INSERT INTO points`

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

func TestBulkInsertReturnsInputUnchangedForEmptyPoints(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	points := []model.Point{}
	got, err := repo.BulkInsert(context.Background(), points)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, points) {
		t.Fatalf("expected %v, got %v", points, got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected batch executed: %v", err)
	}
}

func TestBulkInsertReturnsInsertedPointsWithIDs(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	points := []model.Point{
		{Name: "Alpha", X: 1.5, Y: 2.5, Type: "shop"},
		{Name: "Beta", X: 3.5, Y: 4.5, Type: "restaurant"},
	}

	eb := mock.ExpectBatch()
	eb.ExpectQuery(bulkInsertQuery).
		WithArgs("Alpha", 1.5, 2.5, "shop").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(1)))
	eb.ExpectQuery(bulkInsertQuery).
		WithArgs("Beta", 3.5, 4.5, "restaurant").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(2)))

	got, err := repo.BulkInsert(context.Background(), points)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []model.Point{
		{ID: 1, Name: "Alpha", X: 1.5, Y: 2.5, Type: "shop"},
		{ID: 2, Name: "Beta", X: 3.5, Y: 4.5, Type: "restaurant"},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %+v, got %+v", expected, got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBulkInsertReturnsErrorWhenScanFails(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	points := []model.Point{
		{Name: "Alpha", X: 1.5, Y: 2.5, Type: "shop"},
	}

	eb := mock.ExpectBatch()
	eb.ExpectQuery(bulkInsertQuery).
		WithArgs("Alpha", 1.5, 2.5, "shop").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow("not-an-id"))

	got, err := repo.BulkInsert(context.Background(), points)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got != nil {
		t.Fatalf("expected nil points, got %v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindByIDReturnsPointWhenFound(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	x, y := 1.5, 2.5
	rows := pgxmock.NewRows([]string{"id", "name", "x", "y"}).
		AddRow(int64(1), "Alpha", x, y)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,name, x, y")).
		WithArgs(x, y).
		WillReturnRows(rows)

	point, err := repo.FindByID(context.Background(), x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := &model.Point{ID: 1, Name: "Alpha", X: x, Y: y}
	if !reflect.DeepEqual(point, expected) {
		t.Fatalf("expected %+v, got %+v", expected, point)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindByIDReturnsNilWhenNoRowsMatch(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	x, y := 1.5, 2.5
	rows := pgxmock.NewRows([]string{"id", "name", "x", "y"})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,name, x, y")).
		WithArgs(x, y).
		WillReturnRows(rows)

	point, err := repo.FindByID(context.Background(), x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if point != nil {
		t.Fatalf("expected nil point, got %+v", point)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindByIDReturnsErrorWhenQueryFails(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	x, y := 1.5, 2.5
	queryErr := errors.New("boom")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,name, x, y")).
		WithArgs(x, y).
		WillReturnError(queryErr)

	point, err := repo.FindByID(context.Background(), x, y)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, queryErr) {
		t.Fatalf("expected wrapped query error, got %v", err)
	}
	if point != nil {
		t.Fatalf("expected nil point, got %+v", point)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestFindByIDReturnsErrorOnScanFailure(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	x, y := 1.5, 2.5
	rows := pgxmock.NewRows([]string{"id", "name", "x", "y"}).
		AddRow("not-an-id", "Alpha", x, y)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,name, x, y")).
		WithArgs(x, y).
		WillReturnRows(rows)

	point, err := repo.FindByID(context.Background(), x, y)
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}
	if point != nil {
		t.Fatalf("expected nil point, got %+v", point)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreatePointExecutesInsert(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	p := &model.Point{Name: "Alpha", X: 1.5, Y: 2.5, Type: "shop"}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO points(name, x, y, type)")).
		WithArgs(&p.Name, &p.X, &p.Y, &p.Type).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repo.CreatePoint(context.Background(), p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}

	// NOTE: CreatePoint's INSERT has no RETURNING clause, so the
	// Postgres-generated ID is never scanned back into p. This test
	// documents that current behavior rather than the presumably-intended
	// one (see point_repository.go).
	if p.ID != 0 {
		t.Fatalf("expected ID to remain unset (0), got %d", p.ID)
	}
}

func TestCreatePointReturnsErrorWhenExecFails(t *testing.T) {
	repo, mock := newTestPointRepository(t)

	p := &model.Point{Name: "Alpha", X: 1.5, Y: 2.5, Type: "shop"}
	execErr := errors.New("boom")

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO points(name, x, y, type)")).
		WithArgs(&p.Name, &p.X, &p.Y, &p.Type).
		WillReturnError(execErr)

	err := repo.CreatePoint(context.Background(), p)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, execErr) {
		t.Fatalf("expected wrapped exec error, got %v", err)
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
