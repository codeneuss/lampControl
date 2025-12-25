package dto

import (
	"time"

	"github.com/codeneuss/lampcontrol/internal/domain"
)

// CustomEffectDTO represents a custom effect for API responses
type CustomEffectDTO struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Colors    []RGBColorDTO  `json:"colors"`
	Pattern   string         `json:"pattern"`
	Speed     uint8          `json:"speed"`
	CreatedAt time.Time      `json:"created_at"`
}

// RGBColorDTO represents an RGB color
type RGBColorDTO struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// CreateEffectRequestDTO represents a request to create a custom effect
type CreateEffectRequestDTO struct {
	Name    string         `json:"name"`
	Colors  []RGBColorDTO  `json:"colors"`
	Pattern string         `json:"pattern"`
	Speed   uint8          `json:"speed"`
}

// CustomEffectFromDomain converts a domain CustomEffect to DTO
func CustomEffectFromDomain(effect *domain.CustomEffect) CustomEffectDTO {
	colors := make([]RGBColorDTO, len(effect.Colors))
	for i, c := range effect.Colors {
		colors[i] = RGBColorDTO{R: c.R, G: c.G, B: c.B}
	}

	return CustomEffectDTO{
		ID:        effect.ID,
		Name:      effect.Name,
		Colors:    colors,
		Pattern:   effect.Pattern,
		Speed:     effect.Speed,
		CreatedAt: effect.CreatedAt,
	}
}

// CustomEffectListFromDomain converts a list of domain CustomEffects to DTOs
func CustomEffectListFromDomain(effects []*domain.CustomEffect) []CustomEffectDTO {
	dtos := make([]CustomEffectDTO, len(effects))
	for i, effect := range effects {
		dtos[i] = CustomEffectFromDomain(effect)
	}
	return dtos
}

// ToDomain converts a CreateEffectRequestDTO to domain model
func (r *CreateEffectRequestDTO) ToDomain() *domain.CustomEffect {
	colors := make([]domain.RGBColor, len(r.Colors))
	for i, c := range r.Colors {
		colors[i] = domain.RGBColor{R: c.R, G: c.G, B: c.B}
	}

	return domain.NewCustomEffect(r.Name, colors, r.Pattern, r.Speed)
}
