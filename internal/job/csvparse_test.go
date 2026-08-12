package job

import (
	"testing"

	"github.com/paucio/LocationGrid-go/internal/model"
)

func TestParseRowReturnsPointForValidRow(t *testing.T) {
	row := []string{"Alpha", "1.5", "2.5", "101"}

	got, err := parseRow(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := model.Point{Name: "Alpha", X: 1.5, Y: 2.5}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func TestParseRowTrimsWhitespace(t *testing.T) {
	row := []string{"  Alpha  ", " 1.5 ", " 2.5 ", "101"}

	got, err := parseRow(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := model.Point{Name: "Alpha", X: 1.5, Y: 2.5}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}

func TestParseRowReturnsErrorForTooFewColumns(t *testing.T) {
	row := []string{"Alpha", "1.5", "2.5"}

	_, err := parseRow(row)
	if err == nil {
		t.Fatal("expected error for row with too few columns, got nil")
	}
}

func TestParseRowReturnsErrorForEmptyRow(t *testing.T) {
	_, err := parseRow(nil)
	if err == nil {
		t.Fatal("expected error for empty row, got nil")
	}
}

func TestParseRowReturnsErrorForInvalidX(t *testing.T) {
	row := []string{"Alpha", "not-a-number", "2.5", "101"}

	_, err := parseRow(row)
	if err == nil {
		t.Fatal("expected error for invalid X coordinate, got nil")
	}
}

func TestParseRowReturnsErrorForInvalidY(t *testing.T) {
	row := []string{"Alpha", "1.5", "not-a-number", "101"}

	_, err := parseRow(row)
	if err == nil {
		t.Fatal("expected error for invalid Y coordinate, got nil")
	}
}

func TestParseRowIgnoresExtraColumns(t *testing.T) {
	row := []string{"Alpha", "1.5", "2.5", "101", "extra", "columns"}

	got, err := parseRow(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := model.Point{Name: "Alpha", X: 1.5, Y: 2.5}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}
}
