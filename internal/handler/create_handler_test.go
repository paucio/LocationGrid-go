package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/paucio/LocationGrid-go/internal/model"
)

type fakePointCreate struct {
	err      error
	called   bool
	gotPoint *model.Point
}

func (f *fakePointCreate) CreatePoint(ctx context.Context, p *model.Point) error {
	f.called = true
	f.gotPoint = p
	return f.err
}

func TestCreateReturnsBadRequestForInvalidX(t *testing.T) {
	create := &fakePointCreate{}
	h := NewCreateHandler(create)

	req := httptest.NewRequest(http.MethodPost, "/create?x=notanumber&y=1.0&type=shop&name=Alpha", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if create.called {
		t.Fatal("expected create not to be called for invalid x")
	}
}

func TestCreateReturnsBadRequestForInvalidY(t *testing.T) {
	create := &fakePointCreate{}
	h := NewCreateHandler(create)

	req := httptest.NewRequest(http.MethodPost, "/create?x=1.0&y=notanumber&type=shop&name=Alpha", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if create.called {
		t.Fatal("expected create not to be called for invalid y")
	}
}

func TestCreateReturnsBadRequestForMissingType(t *testing.T) {
	create := &fakePointCreate{}
	h := NewCreateHandler(create)

	req := httptest.NewRequest(http.MethodPost, "/create?x=1.0&y=2.0&name=Alpha", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if create.called {
		t.Fatal("expected create not to be called for missing type")
	}
}

func TestCreateReturnsBadRequestForMissingName(t *testing.T) {
	create := &fakePointCreate{}
	h := NewCreateHandler(create)

	req := httptest.NewRequest(http.MethodPost, "/create?x=1.0&y=2.0&type=shop", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if create.called {
		t.Fatal("expected create not to be called for missing name")
	}
}

func TestCreateReturnsInternalServerErrorWhenCreateFails(t *testing.T) {
	create := &fakePointCreate{err: errors.New("boom")}
	h := NewCreateHandler(create)

	req := httptest.NewRequest(http.MethodPost, "/create?x=1.0&y=2.0&type=shop&name=Alpha", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestCreatePassesParsedPointToCreate(t *testing.T) {
	create := &fakePointCreate{}
	h := NewCreateHandler(create)

	req := httptest.NewRequest(http.MethodPost, "/create?x=1.5&y=2.5&type=shop&name=Alpha", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if !create.called {
		t.Fatal("expected create to be called")
	}

	expected := &model.Point{Name: "Alpha", Type: "shop", X: 1.5, Y: 2.5}
	if !reflect.DeepEqual(create.gotPoint, expected) {
		t.Fatalf("expected point %+v, got %+v", expected, create.gotPoint)
	}
}

func TestCreateReturnsEmptyBodyOnSuccess(t *testing.T) {
	create := &fakePointCreate{}
	h := NewCreateHandler(create)

	req := httptest.NewRequest(http.MethodPost, "/create?x=1.5&y=2.5&type=shop&name=Alpha", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
	if body := rec.Body.String(); body != "" {
		t.Fatalf("expected empty body, got %q", body)
	}
}
