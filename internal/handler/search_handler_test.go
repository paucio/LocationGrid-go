package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/paucio/LocationGrid-go/internal/model"
)

type fakeIDLookup struct {
	ids      []int64
	err      error
	gotX     float64
	gotY     float64
	gotType  string
	gotLimit int
}

func (f *fakeIDLookup) GetPointByCoordinates(ctx context.Context, x, y float64, pointType string, limit int) ([]int64, error) {
	f.gotX = x
	f.gotY = y
	f.gotType = pointType
	f.gotLimit = limit
	return f.ids, f.err
}

type fakePointFinder struct {
	points      []*model.Point
	err         error
	receivedIDs []int64
}

func (f *fakePointFinder) FindByIds(ctx context.Context, ids []int64) ([]*model.Point, error) {
	f.receivedIDs = ids
	return f.points, f.err
}

func TestSearchReturnsBadRequestForInvalidX(t *testing.T) {
	h := NewSearchHandler(&fakeIDLookup{}, &fakePointFinder{})

	req := httptest.NewRequest(http.MethodGet, "/search?x=notanumber&y=1.0&type=1&limit=10", nil)
	rec := httptest.NewRecorder()

	h.Nearest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSearchReturnsBadRequestForInvalidY(t *testing.T) {
	h := NewSearchHandler(&fakeIDLookup{}, &fakePointFinder{})

	req := httptest.NewRequest(http.MethodGet, "/search?x=1.0&y=notanumber&type=1&limit=10", nil)
	rec := httptest.NewRecorder()

	h.Nearest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSearchReturnsBadRequestForMissingType(t *testing.T) {
	h := NewSearchHandler(&fakeIDLookup{}, &fakePointFinder{})

	req := httptest.NewRequest(http.MethodGet, "/search?x=1.0&y=2.0&limit=10", nil)
	rec := httptest.NewRecorder()

	h.Nearest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSearchReturnsBadRequestForNonNumericType(t *testing.T) {
	h := NewSearchHandler(&fakeIDLookup{}, &fakePointFinder{})

	req := httptest.NewRequest(http.MethodGet, "/search?x=1.0&y=2.0&type=shop&limit=10", nil)
	rec := httptest.NewRecorder()

	h.Nearest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSearchReturnsBadRequestForMissingLimit(t *testing.T) {
	h := NewSearchHandler(&fakeIDLookup{}, &fakePointFinder{})

	req := httptest.NewRequest(http.MethodGet, "/search?x=1.0&y=2.0&type=1", nil)
	rec := httptest.NewRecorder()

	h.Nearest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSearchReturnsBadRequestForZeroOrNegativeLimit(t *testing.T) {
	for _, limit := range []string{"0", "-5"} {
		req := httptest.NewRequest(http.MethodGet, "/search?x=1.0&y=2.0&type=1&limit="+limit, nil)
		rec := httptest.NewRecorder()

		h := NewSearchHandler(&fakeIDLookup{}, &fakePointFinder{})
		h.Nearest(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("limit=%s: expected status %d, got %d", limit, http.StatusBadRequest, rec.Code)
		}
	}
}

func TestSearchPassesParsedCoordinatesTypeAndLimitToLookup(t *testing.T) {
	lookup := &fakeIDLookup{ids: []int64{1, 2}}
	finder := &fakePointFinder{}
	h := NewSearchHandler(lookup, finder)

	req := httptest.NewRequest(http.MethodGet, "/search?x=1.5&y=2.5&type=7&limit=10", nil)
	rec := httptest.NewRecorder()

	h.Nearest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if lookup.gotX != 1.5 || lookup.gotY != 2.5 {
		t.Fatalf("expected lookup called with (1.5, 2.5), got (%v, %v)", lookup.gotX, lookup.gotY)
	}
	if lookup.gotType != "7" {
		t.Fatalf("expected lookup called with type %q, got %q", "7", lookup.gotType)
	}
	if lookup.gotLimit != 10 {
		t.Fatalf("expected lookup called with limit 10, got %d", lookup.gotLimit)
	}
}

func TestSearchReturnsInternalServerErrorWhenLookupFails(t *testing.T) {
	lookup := &fakeIDLookup{err: errors.New("boom")}
	finder := &fakePointFinder{}
	h := NewSearchHandler(lookup, finder)

	req := httptest.NewRequest(http.MethodGet, "/search?x=1.0&y=2.0&type=1&limit=10", nil)
	rec := httptest.NewRecorder()

	h.Nearest(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestSearchReturnsInternalServerErrorWhenFinderFails(t *testing.T) {
	lookup := &fakeIDLookup{ids: []int64{1, 2}}
	finder := &fakePointFinder{err: errors.New("boom")}
	h := NewSearchHandler(lookup, finder)

	req := httptest.NewRequest(http.MethodGet, "/search?x=1.0&y=2.0&type=1&limit=10", nil)
	rec := httptest.NewRecorder()

	h.Nearest(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestSearchReturnsPointsAsJSON(t *testing.T) {
	expected := []*model.Point{
		{ID: 1, Name: "Alpha", X: 1.5, Y: 2.5},
		{ID: 2, Name: "Beta", X: 3.5, Y: 4.5},
	}
	lookup := &fakeIDLookup{ids: []int64{1, 2}}
	finder := &fakePointFinder{points: expected}
	h := NewSearchHandler(lookup, finder)

	req := httptest.NewRequest(http.MethodGet, "/search?x=1.5&y=2.5&type=1&limit=10", nil)
	rec := httptest.NewRecorder()

	h.Nearest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
	if !reflect.DeepEqual(finder.receivedIDs, []int64{1, 2}) {
		t.Fatalf("expected finder to receive ids [1 2], got %v", finder.receivedIDs)
	}

	var got []*model.Point
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %+v, got %+v", expected, got)
	}
}

func TestSearchReturnsEmptyJSONArrayWhenNoPointsFound(t *testing.T) {
	lookup := &fakeIDLookup{ids: []int64{}}
	finder := &fakePointFinder{points: []*model.Point{}}
	h := NewSearchHandler(lookup, finder)

	req := httptest.NewRequest(http.MethodGet, "/search?x=1.5&y=2.5&type=1&limit=10", nil)
	rec := httptest.NewRecorder()

	h.Nearest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var got []*model.Point
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no points, got %v", got)
	}
}
