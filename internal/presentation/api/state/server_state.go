package state

import (
	"fmt"
	"sync"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/domain"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/websocket"
)

// ServerState manages the server's global state
type ServerState struct {
	mu             sync.RWMutex
	selectedDevice string                       // Currently selected device address
	deviceService  *application.DeviceService
	wsHub          *websocket.Hub
}

// NewServerState creates a new server state
func NewServerState(deviceService *application.DeviceService) *ServerState {
	state := &ServerState{
		deviceService: deviceService,
	}

	// Create WebSocket hub with reference to state
	state.wsHub = websocket.NewHub(deviceService, state.GetSelectedDeviceAddress)

	return state
}

// SelectDevice sets the currently selected device
func (s *ServerState) SelectDevice(address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate that device exists
	_, err := s.deviceService.GetDevice(address)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	s.selectedDevice = address
	return nil
}

// GetSelectedDeviceAddress returns the currently selected device address
func (s *ServerState) GetSelectedDeviceAddress() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.selectedDevice == "" {
		return "", fmt.Errorf("no device selected")
	}

	return s.selectedDevice, nil
}

// GetSelectedDevice returns the currently selected device
func (s *ServerState) GetSelectedDevice() (*domain.Device, error) {
	addr, err := s.GetSelectedDeviceAddress()
	if err != nil {
		return nil, err
	}

	return s.deviceService.GetDevice(addr)
}

// GetDeviceService returns the device service
func (s *ServerState) GetDeviceService() *application.DeviceService {
	return s.deviceService
}

// GetWebSocketHub returns the WebSocket hub
func (s *ServerState) GetWebSocketHub() *websocket.Hub {
	return s.wsHub
}

// BroadcastState broadcasts the current device state to all WebSocket clients
func (s *ServerState) BroadcastState() {
	s.wsHub.BroadcastDeviceState()
}
