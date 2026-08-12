package job

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paucio/LocationGrid-go/internal/model"
)

type fakeBulkInserter struct {
	calls [][]model.Point
	err   error
}

func (f *fakeBulkInserter) BulkInsert(ctx context.Context, points []model.Point) ([]model.Point, error) {
	cp := make([]model.Point, len(points))
	copy(cp, points)
	f.calls = append(f.calls, cp)
	return cp, f.err
}

type fakeRedisBulkInserter struct {
	calls [][]model.Point
	err   error
}

func (f *fakeRedisBulkInserter) AddIDsBulk(ctx context.Context, points []model.Point) error {
	cp := make([]model.Point, len(points))
	copy(cp, points)
	f.calls = append(f.calls, cp)
	return f.err
}

func newTestImporter() (*Importer, *fakeBulkInserter, *fakeRedisBulkInserter) {
	repo := &fakeBulkInserter{}
	index := &fakeRedisBulkInserter{}
	return NewImporter(repo, index), repo, index
}

func TestImportPointsImportsValidRowsAndSkipsInvalidOnes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := "Alpha,1.5,2.5,101\n" +
			"Beta,3.5,4.5,102\n" +
			"Gamma,5.5,6.5,103\n" +
			"too,few\n" +
			"Delta,notanumber,6.5,104\n"
		w.Write([]byte(body))
	}))
	defer server.Close()

	importer, repo, index := newTestImporter()

	stats, err := importer.ImportPoints(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.TotalPointsImported != 3 {
		t.Fatalf("expected 3 imported points, got %d", stats.TotalPointsImported)
	}
	if stats.SkippedPoints != 2 {
		t.Fatalf("expected 2 skipped points, got %d", stats.SkippedPoints)
	}

	if len(repo.calls) != 1 || len(repo.calls[0]) != 3 {
		t.Fatalf("expected a single batch of 3 points sent to repo, got %v", repo.calls)
	}
	if len(index.calls) != 1 || len(index.calls[0]) != 3 {
		t.Fatalf("expected a single batch of 3 points sent to index, got %v", index.calls)
	}
}

func TestImportPointsReturnsErrorForNon200Status(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	importer, _, _ := newTestImporter()

	_, err := importer.ImportPoints(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to fetch points") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestImportPointsReturnsErrorForInvalidURL(t *testing.T) {
	importer, _, _ := newTestImporter()

	_, err := importer.ImportPoints(context.Background(), "://bad-url")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestImportPointsReturnsErrorWhenRequestFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close() // nothing is listening on url anymore

	importer, _, _ := newTestImporter()

	_, err := importer.ImportPoints(context.Background(), url)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestImportPointsReturnsErrorWhenBulkInsertFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Alpha,1.5,2.5,101\n"))
	}))
	defer server.Close()

	importer, repo, index := newTestImporter()
	repo.err = errors.New("boom")

	stats, err := importer.ImportPoints(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to insert remaining points") {
		t.Fatalf("unexpected error message: %v", err)
	}
	if stats.TotalPointsImported != 0 {
		t.Fatalf("expected 0 imported points on failure, got %d", stats.TotalPointsImported)
	}
	if len(index.calls) != 0 {
		t.Fatalf("expected index not to be called when repo insert fails, got %v", index.calls)
	}
}

func TestImportPointsReturnsErrorWhenAddIDsBulkFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Alpha,1.5,2.5,101\n"))
	}))
	defer server.Close()

	importer, _, index := newTestImporter()
	index.err = errors.New("boom")

	_, err := importer.ImportPoints(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to add IDs to Redis") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestImportPointsFlushesOnBatchSizeBoundary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var b strings.Builder
		for i := 0; i < batchSize+1; i++ {
			fmt.Fprintf(&b, "Point%d,%d,%d,%d\n", i, i, i, i)
		}
		w.Write([]byte(b.String()))
	}))
	defer server.Close()

	importer, repo, index := newTestImporter()

	stats, err := importer.ImportPoints(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.TotalPointsImported != batchSize+1 {
		t.Fatalf("expected %d imported points, got %d", batchSize+1, stats.TotalPointsImported)
	}
	if stats.SkippedPoints != 0 {
		t.Fatalf("expected 0 skipped points, got %d", stats.SkippedPoints)
	}

	if len(repo.calls) != 2 {
		t.Fatalf("expected 2 flush calls, got %d", len(repo.calls))
	}
	if len(repo.calls[0]) != batchSize {
		t.Fatalf("expected first batch of %d, got %d", batchSize, len(repo.calls[0]))
	}
	if len(repo.calls[1]) != 1 {
		t.Fatalf("expected second batch of 1, got %d", len(repo.calls[1]))
	}

	if len(index.calls) != 2 {
		t.Fatalf("expected 2 index flush calls, got %d", len(index.calls))
	}
}
