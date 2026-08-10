package cache

import (
	"fmt"
	"math"
)

const (
	cellSize = 50 // Define the size of each cell in the grid
)

func RedisKey(x, y int) string {
	return "point_x:" + fmt.Sprintf("%.6f", float64(x)) + ":y:" + fmt.Sprintf("%.6f", float64(y))
}

func CellForCoordinates(x, y float64) (int, int) {
	cellX := int(math.Floor(x / cellSize))
	cellY := int(math.Floor(y / cellSize))
	return cellX, cellY
}