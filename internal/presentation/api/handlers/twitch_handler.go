package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/domain"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/storage"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/dto"
	"github.com/joho/godotenv"
)

// TwitchHandler handles Twitch configuration endpoints
type TwitchHandler struct {
	twitchService *application.TwitchService
	storage       *storage.TwitchStorage
}

// NewTwitchHandler creates a new Twitch handler
func NewTwitchHandler(twitchService *application.TwitchService, storage *storage.TwitchStorage) *TwitchHandler {
	return &TwitchHandler{
		twitchService: twitchService,
		storage:       storage,
	}
}

// GetConfig returns current Twitch configuration
func (h *TwitchHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.storage.Get()
	configDTO := dto.FromDomainTwitchConfig(config)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configDTO)
}

// UpdateConfig updates Twitch configuration
func (h *TwitchHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var updateDTO dto.TwitchConfigUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current config
	config := h.storage.Get()

	// Apply updates
	updateDTO.ApplyUpdate(config)

	// Validate and save
	if err := h.storage.Save(config); err != nil {
		log.Printf("Failed to save Twitch config: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Restart Twitch service if enabled
	if config.Enabled {
		h.twitchService.Stop()
		if err := h.twitchService.Start(r.Context()); err != nil {
			log.Printf("Failed to start Twitch service: %v", err)
			http.Error(w, "Failed to connect to Twitch", http.StatusInternalServerError)
			return
		}
	} else {
		h.twitchService.Stop()
	}

	// Return updated config
	configDTO := dto.FromDomainTwitchConfig(config)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(configDTO)
}

// GetStatus returns Twitch connection status
func (h *TwitchHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	config := h.storage.Get()

	status := dto.TwitchStatusDTO{
		Connected: h.twitchService.IsConnected(),
		Channel:   config.Channel,
	}

	// Add active effect if any
	if activeEffect := h.twitchService.GetActiveEffect(); activeEffect != nil {
		status.ActiveEffect = dto.FromActiveEffect(activeEffect, config.EffectDuration)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// GetAvailableCommands returns list of available commands
func (h *TwitchHandler) GetAvailableCommands(w http.ResponseWriter, r *http.Request) {
	colors := make([]string, 0, len(domain.ColorMap))
	for color := range domain.ColorMap {
		colors = append(colors, color)
	}

	effects := make([]string, 0, len(domain.EffectMap))
	for effect := range domain.EffectMap {
		effects = append(effects, effect)
	}

	commandList := dto.TwitchCommandListDTO{
		Colors:  colors,
		Effects: effects,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commandList)
}

// GetOAuthURL returns the Twitch OAuth URL for token generation
func (h *TwitchHandler) GetOAuthURL(w http.ResponseWriter, r *http.Request) {
	// .env laden
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	clientID := os.Getenv("TWITCH_CLIENT_ID")
	if clientID == "" {
		clientID = "YOUR_CLIENT_ID"
	}
	// Twitch OAuth URL for chat scope
	// Note: You'll need to register a Twitch app and replace YOUR_CLIENT_ID
	oauthURL := fmt.Sprintf("https://id.twitch.tv/oauth2/authorize?client_id=%s&redirect_uri=http://localhost:8080&response_type=token&scope=chat:read+chat:edit", clientID)

	response := map[string]string{
		"oauth_url":    oauthURL,
		"instructions": "1. Register a Twitch app at https://dev.twitch.tv/console/apps\n2. Set the OAuth Redirect URL to http://localhost:8080\n3. Copy your Client ID and replace YOUR_CLIENT_ID in the URL above\n4. Click the link, authorize, and copy the token from the URL",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
