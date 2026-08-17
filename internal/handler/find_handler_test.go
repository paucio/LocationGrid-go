package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/paucio/LocationGrid-go/internal/model"
)

type fakePointLookup struct {
	point  *model.Point
	err    error
	gotX   float64
	gotY   float64
	called bool
}

func (f *fakePointLookup) FindByID(ctx context.Context, x, y float64) (*model.Point, error) {
	f.called = true
	f.gotX = x
	f.gotY = y
	return f.point, f.err
}

func TestFindReturnsBadRequestForInvalidX(t *testing.T) {
	finder := &fakePointLookup{}
	h := NewFindHandler(finder)

	req := httptest.NewRequest(http.MethodGet, "/find?x=notanumber&y=1.0", nil)
	rec := httptest.NewRecorder()

	h.Find(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if finder.called {
		t.Fatal("expected finder not to be called for invalid x")
	}
}

func TestFindReturnsBadRequestForInvalidY(t *testing.T) {
	finder := &fakePointLookup{}
	h := NewFindHandler(finder)

	req := httptest.NewRequest(http.MethodGet, "/find?x=1.0&y=notanumber", nil)
	rec := httptest.NewRecorder()

	h.Find(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if finder.called {
		t.Fatal("expected finder not to be called for invalid y")
	}
}

func TestFindReturnsInternalServerErrorWhenFinderFails(t *testing.T) {
	finder := &fakePointLookup{err: errors.New("boom")}
	h := NewFindHandler(finder)

	req := httptest.NewRequest(http.MethodGet, "/find?x=1.0&y=2.0", nil)
	rec := httptest.NewRecorder()

	h.Find(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestFindPassesParsedCoordinatesToFinder(t *testing.T) {
	finder := &fakePointLookup{point: &model.Point{ID: 1, Name: "Alpha", X: 1.5, Y: 2.5}}
	h := NewFindHandler(finder)

	req := httptest.NewRequest(http.MethodGet, "/find?x=1.5&y=2.5", nil)
	rec := httptest.NewRecorder()

	h.Find(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if finder.gotX != 1.5 || finder.gotY != 2.5 {
		t.Fatalf("expected finder called with (1.5, 2.5), got (%v, %v)", finder.gotX, finder.gotY)
	}
}

func TestFindReturnsPointAsJSON(t *testing.T) {
	expected := &model.Point{ID: 1, Name: "Alpha", X: 1.5, Y: 2.5}
	finder := &fakePointLookup{point: expected}
	h := NewFindHandler(finder)

	req := httptest.NewRequest(http.MethodGet, "/find?x=1.5&y=2.5", nil)
	rec := httptest.NewRecorder()

	h.Find(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var got model.Point
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if !reflect.DeepEqual(&got, expected) {
		t.Fatalf("expected %+v, got %+v", expected, got)
	}
}

func TestFindReturnsNullWhenPointNotFound(t *testing.T) {
	finder := &fakePointLookup{point: nil, err: nil}
	h := NewFindHandler(finder)

	req := httptest.NewRequest(http.MethodGet, "/find?x=1.5&y=2.5", nil)
	rec := httptest.NewRecorder()

	h.Find(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != "null" {
		t.Fatalf("expected body %q, got %q", "null", got)
	}
}
