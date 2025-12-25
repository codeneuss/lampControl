package bluetooth

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"tinygo.org/x/bluetooth"
)

// Adapter wraps the tinygo bluetooth adapter and provides high-level operations
type Adapter struct {
	adapter *bluetooth.Adapter
}

// NewAdapter creates a new Bluetooth adapter
func NewAdapter() (*Adapter, error) {
	adapter := bluetooth.DefaultAdapter
	if adapter == nil {
		return nil, ErrAdapterNotReady
	}

	// Enable the adapter
	if err := adapter.Enable(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAdapterEnableFailed, err)
	}

	return &Adapter{
		adapter: adapter,
	}, nil
}

// ScanResult represents a discovered device
type ScanResult struct {
	Address string
	Name    string
	RSSI    int16
}

// Scan scans for ELK-BLEDOM devices
func (a *Adapter) Scan(ctx context.Context, timeout time.Duration) ([]ScanResult, error) {
	results := make([]ScanResult, 0)
	seen := make(map[string]bool) // Track seen devices to avoid duplicates

	scanCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := a.adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		// Check if context is done
		select {
		case <-scanCtx.Done():
			// Stop scanning
			adapter.StopScan()
			return
		default:
		}

		address := result.Address.String()

		// Skip if we've already seen this device
		if seen[address] {
			return
		}

		// Get device name
		name := result.LocalName()

		// Filter for ELK-BLEDOM devices or devices with recognizable names
		// ELK-BLEDOM devices often advertise as "ELK-BLEDOM", "LEDBLE", etc.
		if name != "" && (strings.Contains(strings.ToUpper(name), "ELK") ||
			strings.Contains(strings.ToUpper(name), "BLEDOM") ||
			strings.Contains(strings.ToUpper(name), "LED") ||
			strings.Contains(strings.ToUpper(name), "STRIP")) {

			results = append(results, ScanResult{
				Address: address,
				Name:    name,
				RSSI:    result.RSSI,
			})
			seen[address] = true
		}
	})

	// Stop scanning
	a.adapter.StopScan()

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrScanFailed, err)
	}

	// Return empty array if no devices found (not an error)
	return results, nil
}

// Connection represents an active connection to a device
type Connection struct {
	device         bluetooth.Device
	characteristic bluetooth.DeviceCharacteristic
	address        string
}

func (a *Adapter) Connect(ctx context.Context, address string, timeout time.Duration) (*Connection, error) {
	fmt.Println("🔍 CONNECT START", address)

	var addr bluetooth.Address
	addr.Set(address)

	dev, err := a.adapter.Connect(addr, bluetooth.ConnectionParams{})
	if err != nil {
		fmt.Println("❌ CONNECT FAILED:", err)
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}
	fmt.Println("✓ CONNECT OK!")

	services, err := dev.DiscoverServices(nil)
	if err != nil {
		fmt.Println("❌ SERVICES FAILED:", err)
		return nil, fmt.Errorf("%w: %v", ErrServiceNotFound, err)
	}

	fmt.Println("Found", len(services), "services")

	var writeChar bluetooth.DeviceCharacteristic

	for i, svc := range services {
		fmt.Println("SERVICE", i, ":", svc.UUID().String())

		chars, err := svc.DiscoverCharacteristics(nil)
		if err != nil {
			fmt.Println("  ❌ CHARS FAILED:", err)
			continue
		}
		fmt.Println("  Found", len(chars), "chars")

		for j, char := range chars {
			uuidStr := char.UUID().String()
			fmt.Println("    CHAR", j, ":", uuidStr)

			// NOTIFY AUF fff4 AKTIVIEREN (wichtig!)
			if strings.Contains(uuidStr, "fff4") {
				char.EnableNotifications(func(buf []byte) {
					fmt.Println("NOTIFY:", hex.EncodeToString(buf))
				})
				fmt.Println("    ✓ NOTIFY ENABLED!")
				time.Sleep(50 * time.Millisecond) // Brief handshake wait
			}

			if j == 1 {
				writeChar = char
				fmt.Println("    → WRITE CHAR:", uuidStr)
			}
		}

	}

	return &Connection{
		device:         dev,
		characteristic: writeChar,
		address:        address,
	}, nil
}

// Write writes data to the device characteristic
func (a *Adapter) Write(ctx context.Context, conn *Connection, data []byte) error {
	fmt.Println("Sending:", hex.EncodeToString(data))

	for i := 0; i < 3; i++ {
		_, err := conn.characteristic.WriteWithoutResponse(data)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrWriteFailed, err)
		}
		fmt.Println("Write", i+1, "OK")
		time.Sleep(20 * time.Millisecond) // Small delay between writes
	}

	fmt.Println("✓ All writes successful!")
	time.Sleep(50 * time.Millisecond) // Minimal delay for effect
	return nil
}

// Disconnect disconnects from a device
func (a *Adapter) Disconnect(conn *Connection) error {
	if conn == nil {
		return nil
	}

	if err := conn.device.Disconnect(); err != nil {
		return fmt.Errorf("%w: %v", ErrDisconnectFailed, err)
	}

	return nil
}

// parseServiceUUID converts a string UUID to bluetooth.UUID format
// ELK-BLEDOM uses standard 128-bit UUIDs
func parseServiceUUID(uuidStr string) [16]byte {
	// Parse UUID string (format: 0000fff0-0000-1000-8000-00805f9b34fb)
	// This is a simplified parser
	var uuid [16]byte

	// Remove dashes
	cleaned := strings.ReplaceAll(uuidStr, "-", "")

	// Parse hex string into bytes
	for i := 0; i < 16 && i*2 < len(cleaned); i++ {
		fmt.Sscanf(cleaned[i*2:i*2+2], "%02x", &uuid[i])
	}

	return uuid
}

// Address returns the connection's device address
func (c *Connection) Address() string {
	return c.address
}
