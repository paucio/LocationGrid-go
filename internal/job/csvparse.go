package job

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/paucio/LocationGrid-go/internal/model"
)

func parseRow(row []string) (model.Point, error) {
	if len(row) < 4 {
		return model.Point{}, fmt.Errorf("row has insufficient columns: %v", row)
	}

	name := strings.TrimSpace(row[0])
	x, err := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
	if err != nil {
		return model.Point{}, fmt.Errorf("invalid X coordinate: %v", err)
	}

	y, err := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
	if err != nil {
		return model.Point{}, fmt.Errorf("invalid Y coordinate: %v", err)
	}

	return model.Point{
		Name: name,
		X:    x,
		Y:    y,
	}, nil
}
