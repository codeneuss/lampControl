package application

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/codeneuss/lampcontrol/internal/domain"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/bluetooth"
	"github.com/codeneuss/lampcontrol/pkg/protocol"
)

// DeviceService orchestrates device control operations
type DeviceService struct {
	bleAdapter     *bluetooth.Adapter
	connections    map[string]*bluetooth.Connection // address -> connection
	devices        map[string]*domain.Device        // address -> device
	mu             sync.RWMutex
	connectTimeout time.Duration
	writeTimeout   time.Duration
	retryAttempts  int
}

// NewDeviceService creates a new device service
func NewDeviceService(adapter *bluetooth.Adapter) *DeviceService {
	return &DeviceService{
		bleAdapter:     adapter,
		connections:    make(map[string]*bluetooth.Connection),
		devices:        make(map[string]*domain.Device),
		connectTimeout: 10 * time.Second,
		writeTimeout:   5 * time.Second,
		retryAttempts:  3,
	}
}

// Scan scans for available devices
func (s *DeviceService) Scan(ctx context.Context, timeout time.Duration) ([]*domain.Device, error) {
	results, err := s.bleAdapter.Scan(ctx, timeout)
	if err != nil {
		return nil, err
	}

	devices := make([]*domain.Device, 0, len(results))
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, result := range results {
		// Update or create device
		if dev, exists := s.devices[result.Address]; exists {
			dev.LastSeen = time.Now()
			dev.RSSI = result.RSSI
			devices = append(devices, dev)
		} else {
			dev := domain.NewDevice(result.Address, result.Name, result.RSSI)
			s.devices[result.Address] = dev
			devices = append(devices, dev)
		}
	}

	return devices, nil
}

// GetDevice returns a device by address
func (s *DeviceService) GetDevice(address string) (*domain.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dev, exists := s.devices[address]
	if !exists {
		return nil, domain.ErrDeviceNotFound
	}

	return dev, nil
}

// ListDevices returns all known devices
func (s *DeviceService) ListDevices() []*domain.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]*domain.Device, 0, len(s.devices))
	for _, dev := range s.devices {
		devices = append(devices, dev)
	}

	return devices
}

func (s *DeviceService) connect(ctx context.Context, address string) (*bluetooth.Connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, exists := s.connections[address]; exists {
		return conn, nil
	}

	conn, err := s.bleAdapter.Connect(ctx, address, s.connectTimeout)
	if err != nil {
		return nil, err
	}
	s.connections[address] = conn

	// === ELK-BLEDOM DISCOVERY ===
	if dev, exists := s.devices[address]; exists {
		dev.MarkConnected()
	}

	return conn, nil
}

// disconnect closes a connection to a device
func (s *DeviceService) Disconnect(address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.connections[address]
	if !exists {
		return nil // Already disconnected
	}

	if err := s.bleAdapter.Disconnect(conn); err != nil {
		return err
	}

	delete(s.connections, address)

	// Update device status
	if dev, exists := s.devices[address]; exists {
		dev.MarkDisconnected()
	}

	return nil
}

func (s *DeviceService) writeCommand(ctx context.Context, address string, cmd protocol.Command) error {
	var lastErr error

	for attempt := 0; attempt < s.retryAttempts; attempt++ {
		// Connect + get Connection (nicht Device!)
		conn, err := s.connect(ctx, address)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Adapter.Write() verwendet die Connection.characteristic!
		writeCtx, cancel := context.WithTimeout(ctx, s.writeTimeout)
		err = s.bleAdapter.Write(writeCtx, conn, cmd.Bytes())
		cancel()

		if err == nil {
			fmt.Println("✓ Command gesendet:", hex.EncodeToString(cmd.Bytes()))
			return nil
		}

		lastErr = err
		s.Disconnect(address)
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("failed after %d attempts: %w", s.retryAttempts, lastErr)
}

// SetPower sets the power state of a device
func (s *DeviceService) SetPower(ctx context.Context, address string, on bool) error {
	cmd := protocol.NewPowerCommand(on)

	if err := s.writeCommand(ctx, address, cmd); err != nil {
		return err
	}

	// Update local state
	s.mu.Lock()
	defer s.mu.Unlock()

	if dev, exists := s.devices[address]; exists {
		state := dev.State
		state.PowerOn = on
		state.LastUpdated = time.Now()
		dev.UpdateState(state)
	}

	return nil
}

// SetColor sets the RGB color of a device
func (s *DeviceService) SetColor(ctx context.Context, address string, r, g, b uint8) error {
	cmd := protocol.NewRGBCommand(r, g, b)

	if err := s.writeCommand(ctx, address, cmd); err != nil {
		return err
	}

	// Update local state
	s.mu.Lock()
	defer s.mu.Unlock()

	if dev, exists := s.devices[address]; exists {
		state := dev.State
		rgb, _ := domain.NewRGB(r, g, b)
		state.RGB = &rgb
		state.WhiteBalance = nil // Clear white balance when setting RGB
		state.Effect = nil       // Clear effect when setting RGB
		state.LastUpdated = time.Now()
		dev.UpdateState(state)
	}

	return nil
}

// SetBrightness sets the brightness of a device
func (s *DeviceService) SetBrightness(ctx context.Context, address string, level uint8) error {
	cmd := protocol.NewBrightnessCommand(level)

	if err := s.writeCommand(ctx, address, cmd); err != nil {
		return err
	}

	// Update local state
	s.mu.Lock()
	defer s.mu.Unlock()

	if dev, exists := s.devices[address]; exists {
		state := dev.State
		state.Brightness = level
		state.LastUpdated = time.Now()
		dev.UpdateState(state)
	}

	return nil
}

// SetWhiteBalance sets the white balance of a device
func (s *DeviceService) SetWhiteBalance(ctx context.Context, address string, warm, cold uint8) error {
	cmd := protocol.NewWhiteBalanceCommand(warm, cold)

	if err := s.writeCommand(ctx, address, cmd); err != nil {
		return err
	}

	// Update local state
	s.mu.Lock()
	defer s.mu.Unlock()

	if dev, exists := s.devices[address]; exists {
		state := dev.State
		state.WhiteBalance = &domain.WhiteBalance{Warm: warm, Cold: cold}
		state.RGB = nil    // Clear RGB when setting white balance
		state.Effect = nil // Clear effect when setting white balance
		state.LastUpdated = time.Now()
		dev.UpdateState(state)
	}

	return nil
}

// SetEffect sets an effect/scene on a device
func (s *DeviceService) SetEffect(ctx context.Context, address string, effect, speed uint8) error {
	cmd := protocol.NewEffectCommand(effect, speed)

	if err := s.writeCommand(ctx, address, cmd); err != nil {
		return err
	}

	// Update local state
	s.mu.Lock()
	defer s.mu.Unlock()

	if dev, exists := s.devices[address]; exists {
		state := dev.State
		effectInt := int(effect)
		state.Effect = &effectInt
		state.EffectSpeed = &speed
		state.LastUpdated = time.Now()
		dev.UpdateState(state)
	}

	return nil
}

// DisconnectAll disconnects from all devices
func (s *DeviceService) DisconnectAll() error {
	s.mu.Lock()
	addresses := make([]string, 0, len(s.connections))
	for addr := range s.connections {
		addresses = append(addresses, addr)
	}
	s.mu.Unlock()

	var lastErr error
	for _, addr := range addresses {
		if err := s.Disconnect(addr); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
