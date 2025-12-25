package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/codeneuss/lampcontrol/internal/infrastructure/storage"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/dto"
	"github.com/go-chi/chi/v5"
)

// EffectHandler handles custom effect-related HTTP requests
type EffectHandler struct {
	storage *storage.EffectStorage
}

// NewEffectHandler creates a new effect handler
func NewEffectHandler(storage *storage.EffectStorage) *EffectHandler {
	return &EffectHandler{
		storage: storage,
	}
}

// ListEffects handles GET /api/effects
func (h *EffectHandler) ListEffects(w http.ResponseWriter, r *http.Request) {
	effects := h.storage.GetAll()
	effectDTOs := dto.CustomEffectListFromDomain(effects)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(effectDTOs)
}

// CreateEffect handles POST /api/effects
func (h *EffectHandler) CreateEffect(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateEffectRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		http.Error(w, "Effect name is required", http.StatusBadRequest)
		return
	}

	if len(req.Colors) == 0 {
		http.Error(w, "At least one color is required", http.StatusBadRequest)
		return
	}

	// Create effect
	effect := req.ToDomain()

	// Save to storage
	if err := h.storage.Save(effect); err != nil {
		log.Printf("Failed to save effect: %v", err)
		http.Error(w, "Failed to save effect", http.StatusInternalServerError)
		return
	}

	// Return created effect
	effectDTO := dto.CustomEffectFromDomain(effect)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(effectDTO)
}

// DeleteEffect handles DELETE /api/effects/:id
func (h *EffectHandler) DeleteEffect(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Effect ID is required", http.StatusBadRequest)
		return
	}

	if err := h.storage.Delete(id); err != nil {
		log.Printf("Failed to delete effect: %v", err)
		http.Error(w, "Effect not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
