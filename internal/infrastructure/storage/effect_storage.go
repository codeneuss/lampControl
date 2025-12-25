package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/codeneuss/lampcontrol/internal/domain"
)

// EffectStorage handles persistent storage of custom effects
type EffectStorage struct {
	filePath string
	mu       sync.RWMutex
	effects  map[string]*domain.CustomEffect
}

// NewEffectStorage creates a new effect storage instance
func NewEffectStorage() (*EffectStorage, error) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create .lampcontrol directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".lampcontrol")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	filePath := filepath.Join(configDir, "custom_effects.json")

	storage := &EffectStorage{
		filePath: filePath,
		effects:  make(map[string]*domain.CustomEffect),
	}

	// Load existing effects
	if err := storage.load(); err != nil {
		// If file doesn't exist, that's okay - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load effects: %w", err)
		}
	}

	return storage, nil
}

// GetAll returns all custom effects
func (s *EffectStorage) GetAll() []*domain.CustomEffect {
	s.mu.RLock()
	defer s.mu.RUnlock()

	effects := make([]*domain.CustomEffect, 0, len(s.effects))
	for _, effect := range s.effects {
		effects = append(effects, effect)
	}

	return effects
}

// Get returns a custom effect by ID
func (s *EffectStorage) Get(id string) (*domain.CustomEffect, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	effect, exists := s.effects[id]
	if !exists {
		return nil, fmt.Errorf("effect not found")
	}

	return effect, nil
}

// Save saves a custom effect
func (s *EffectStorage) Save(effect *domain.CustomEffect) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.effects[effect.ID] = effect

	return s.persist()
}

// Delete deletes a custom effect by ID
func (s *EffectStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.effects[id]; !exists {
		return fmt.Errorf("effect not found")
	}

	delete(s.effects, id)

	return s.persist()
}

// load loads effects from file
func (s *EffectStorage) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var effects []*domain.CustomEffect
	if err := json.Unmarshal(data, &effects); err != nil {
		return fmt.Errorf("failed to unmarshal effects: %w", err)
	}

	for _, effect := range effects {
		s.effects[effect.ID] = effect
	}

	return nil
}

// persist saves effects to file
func (s *EffectStorage) persist() error {
	effects := make([]*domain.CustomEffect, 0, len(s.effects))
	for _, effect := range s.effects {
		effects = append(effects, effect)
	}

	data, err := json.MarshalIndent(effects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal effects: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write effects file: %w", err)
	}

	return nil
}
