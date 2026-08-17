package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/paucio/LocationGrid-go/internal/model"
)

type pointLookup interface {
	FindByID(ctx context.Context, x, y float64) (*model.Point, error)
}

type FindHandler struct {
	finder pointLookup
}

func NewFindHandler(finder pointLookup) *FindHandler {
	return &FindHandler{
		finder: finder,
	}
}

func (h *FindHandler) Find(w http.ResponseWriter, r *http.Request) {
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
	point, err := h.finder.FindByID(ctx, x, y)
	if err != nil {
		http.Error(w, "Failed to find point", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(point); err != nil {
		http.Error(w, "Failed to encode points", http.StatusInternalServerError)
		return
	}
}
