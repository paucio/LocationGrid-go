package model

// Point represents a point in a 2D space with an ID, name, and coordinates (X, Y).
type Point struct {
	ID   int64   `json:"id"`
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}
