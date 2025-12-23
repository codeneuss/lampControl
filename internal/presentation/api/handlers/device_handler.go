package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/codeneuss/lampcontrol/internal/presentation/api/dto"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/state"
)

// DeviceHandler handles device-related HTTP requests
type DeviceHandler struct {
	state *state.ServerState
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(state *state.ServerState) *DeviceHandler {
	return &DeviceHandler{
		state: state,
	}
}

// Health handles GET /api/health
func (h *DeviceHandler) Health(w http.ResponseWriter, r *http.Request) {
	response := dto.HealthResponseDTO{
		Status:    "ok",
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListDevices handles GET /api/devices
func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	devices := h.state.GetDeviceService().ListDevices()
	deviceDTOs := dto.FromDomainList(devices)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deviceDTOs)
}

// ScanDevices handles POST /api/scan
func (h *DeviceHandler) ScanDevices(w http.ResponseWriter, r *http.Request) {
	var req dto.ScanRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Use default timeout if not provided
		req.Timeout = "10s"
	}

	// Parse timeout duration
	timeout, err := time.ParseDuration(req.Timeout)
	if err != nil {
		timeout = 10 * time.Second
	}

	// Perform scan
	ctx := context.Background()
	devices, err := h.state.GetDeviceService().Scan(ctx, timeout)
	if err != nil {
		log.Printf("Scan failed: %v", err)
		http.Error(w, "Scan failed", http.StatusInternalServerError)
		return
	}

	deviceDTOs := dto.FromDomainList(devices)

	// Broadcast scan results to WebSocket clients
	h.state.GetWebSocketHub().BroadcastMessage(dto.NewScanResultMessage(deviceDTOs))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deviceDTOs)
}

// SelectDevice handles POST /api/device/select
func (h *DeviceHandler) SelectDevice(w http.ResponseWriter, r *http.Request) {
	var req dto.SelectDeviceRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Address == "" {
		http.Error(w, "Device address is required", http.StatusBadRequest)
		return
	}

	// Select the device
	if err := h.state.SelectDevice(req.Address); err != nil {
		log.Printf("Failed to select device: %v", err)
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Get the selected device
	device, err := h.state.GetSelectedDevice()
	if err != nil {
		http.Error(w, "Failed to get device", http.StatusInternalServerError)
		return
	}

	deviceDTO := dto.FromDomain(device)

	// Broadcast state update to WebSocket clients
	h.state.BroadcastState()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"device":  deviceDTO,
	})
}

// GetCurrentDevice handles GET /api/device/current
func (h *DeviceHandler) GetCurrentDevice(w http.ResponseWriter, r *http.Request) {
	device, err := h.state.GetSelectedDevice()
	if err != nil {
		http.Error(w, "No device selected", http.StatusNotFound)
		return
	}

	deviceDTO := dto.FromDomain(device)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deviceDTO)
}
