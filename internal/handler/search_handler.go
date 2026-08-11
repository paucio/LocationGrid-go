package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/paucio/LocationGrid-go/internal/model"
)

type idLookup interface {
	GetPointByCoordinates(ctx context.Context, x, y float64) ([]int64, error)
}

type pointFinder interface {
	FindByIds(ctx context.Context, ids []int64) ([]*model.Point, error)
}

type SearchHandler struct {
	lookup idLookup
	finder pointFinder
}

func NewSearchHandler(lookup idLookup, finder pointFinder) *SearchHandler {
	return &SearchHandler{
		lookup: lookup,
		finder: finder,
	}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	x, err := strconv.ParseFloat(r.URL.Query().Get("x"), 64)
	if err != nil {
		http.Error(w, "Invalid x coordinate", http.StatusBadRequest)
		return
	}

	y, err := strconv.ParseFloat(r.URL.Query().Get("y"), 64)
	if err != nil {
		http.Error(w, "Invalid y coordinate", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	pointIDs, err := h.lookup.GetPointByCoordinates(ctx, x, y)
	if err != nil {
		http.Error(w, "Failed to lookup points", http.StatusInternalServerError)
		return
	}

	points, err := h.finder.FindByIds(ctx, pointIDs)
	if err != nil {
		http.Error(w, "Failed to find points", http.StatusInternalServerError)
		return
	}

	// Serialize and return the found points
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(points); err != nil {
		http.Error(w, "Failed to encode points", http.StatusInternalServerError)
		return
	}
}
