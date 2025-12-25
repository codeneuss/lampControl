package domain

import "time"

// CustomEffect represents a user-defined lighting effect
type CustomEffect struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Colors      []RGBColor  `json:"colors"`
	Pattern     string      `json:"pattern"` // "fade", "strobe", "jump", "pulse"
	Speed       uint8       `json:"speed"`
	CreatedAt   time.Time   `json:"created_at"`
}

// RGBColor represents an RGB color value
type RGBColor struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// NewCustomEffect creates a new custom effect
func NewCustomEffect(name string, colors []RGBColor, pattern string, speed uint8) *CustomEffect {
	return &CustomEffect{
		ID:        generateID(),
		Name:      name,
		Colors:    colors,
		Pattern:   pattern,
		Speed:     speed,
		CreatedAt: time.Now(),
	}
}

// generateID generates a simple ID based on timestamp
func generateID() string {
	return time.Now().Format("20060102150405")
}
