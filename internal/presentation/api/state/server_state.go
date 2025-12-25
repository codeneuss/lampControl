package state

import (
	"fmt"
	"sync"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/domain"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/dto"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/websocket"
)

// ServerState manages the server's global state
type ServerState struct {
	mu             sync.RWMutex
	selectedDevice string                       // Currently selected device address
	deviceService  *application.DeviceService
	twitchService  *application.TwitchService
	wsHub          *websocket.Hub
}

// NewServerState creates a new server state
func NewServerState(deviceService *application.DeviceService, twitchService *application.TwitchService) *ServerState {
	state := &ServerState{
		deviceService: deviceService,
		twitchService: twitchService,
	}

	// Create WebSocket hub with reference to state
	state.wsHub = websocket.NewHub(deviceService, state.GetSelectedDeviceAddress)

	// Set Twitch callbacks if Twitch service is provided
	if twitchService != nil {
		twitchService.SetStatusChangeCallback(func(connected bool) {
			state.BroadcastTwitchStatus()
		})

		twitchService.SetCommandSuccessCallback(func(username, command string) {
			state.BroadcastTwitchCommand(username, command)
		})

		twitchService.SetGetSelectedDeviceFunc(state.GetSelectedDeviceAddress)
	}

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

// GetTwitchService returns the Twitch service
func (s *ServerState) GetTwitchService() *application.TwitchService {
	return s.twitchService
}

// BroadcastTwitchStatus broadcasts Twitch connection status to all WebSocket clients
func (s *ServerState) BroadcastTwitchStatus() {
	if s.twitchService == nil {
		return
	}

	// This will be called by the status change callback
	// The actual status will be fetched by the frontend via API
	// For now, just signal that status changed
	// We could enhance this to include the status in the message
}

// BroadcastTwitchCommand broadcasts a Twitch command execution to all WebSocket clients
func (s *ServerState) BroadcastTwitchCommand(username, command string) {
	if s.wsHub == nil {
		return
	}

	message := dto.NewTwitchCommandMessage(username, command)
	s.wsHub.BroadcastMessage(message)
}
