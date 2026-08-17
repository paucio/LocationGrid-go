package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/paucio/LocationGrid-go/internal/model"
)

type pointCreate interface {
	CreatePoint(ctx context.Context, p *model.Point) error
}

type CreateHandler struct {
	create pointCreate
}

func NewCreateHandler(create pointCreate) *CreateHandler {
	return &CreateHandler{
		create: create,
	}
}

func (c *CreateHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	pointType := r.URL.Query().Get("type")
	if pointType == "" {
		http.Error(w, "Invalid type parameter", http.StatusBadRequest)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Invalid name parameter", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	point := &model.Point{
		Name: name,
		Type: pointType,
		X:    x,
		Y:    y,
	}

	if err := c.create.CreatePoint(ctx, point); err != nil {
		http.Error(w, "Failed to find point", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
