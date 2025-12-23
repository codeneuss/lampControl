package domain

import (
	"fmt"
	"time"

	"tinygo.org/x/bluetooth"
)

// RGB represents an RGB color value
type RGB struct {
	R uint8 `json:"r"` // Red (0-255)
	G uint8 `json:"g"` // Green (0-255)
	B uint8 `json:"b"` // Blue (0-255)
}

// NewRGB creates a new RGB color with validation
func NewRGB(r, g, b uint8) (RGB, error) {
	// Values are uint8, so they're already constrained to 0-255
	return RGB{R: r, G: g, B: b}, nil
}

// String returns a string representation of the RGB color
func (rgb RGB) String() string {
	return fmt.Sprintf("RGB(%d,%d,%d)", rgb.R, rgb.G, rgb.B)
}

// WhiteBalance represents warm/cold white balance
type WhiteBalance struct {
	Warm uint8 `json:"warm"` // Warm white intensity (0-255)
	Cold uint8 `json:"cold"` // Cold white intensity (0-255)
}

// DeviceState represents the current state of a device
// Note: This is client-side state as ELK-BLEDOM devices don't report state
type DeviceState struct {
	PowerOn      bool          `json:"power_on"`
	Brightness   uint8         `json:"brightness"`    // 0-255
	RGB          *RGB          `json:"rgb,omitempty"` // Current RGB color (if in RGB mode)
	WhiteBalance *WhiteBalance `json:"white_balance,omitempty"`
	Effect       *int          `json:"effect,omitempty"`       // Current effect index
	EffectSpeed  *uint8        `json:"effect_speed,omitempty"` // Effect speed
	LastUpdated  time.Time     `json:"last_updated"`
}

// NewDeviceState creates a new device state with defaults
func NewDeviceState() DeviceState {
	return DeviceState{
		PowerOn:     false,
		Brightness:  255, // Full brightness by default
		RGB:         &RGB{R: 255, G: 255, B: 255},
		LastUpdated: time.Now(),
	}
}

// Device represents an ELK-BLEDOM LED device
type Device struct {
	Address     string      `json:"address"`      // Bluetooth MAC address
	Name        string      `json:"name"`         // Device name
	RSSI        int16       `json:"rssi"`         // Signal strength
	Connected   bool        `json:"connected"`    // Connection status
	State       DeviceState `json:"state"`        // Current state (assumed)
	LastSeen    time.Time   `json:"last_seen"`    // Last time device was seen
	LastUpdated time.Time   `json:"last_updated"` // Last time state was updated
	// === NEU: ELK-BLEDOM Characteristics ===
	WriteCharacteristic  *bluetooth.DeviceCharacteristic
	NotifyCharacteristic *bluetooth.DeviceCharacteristic
}

// NewDevice creates a new device with the given address and name
func NewDevice(address, name string, rssi int16) *Device {
	return &Device{
		Address:     address,
		Name:        name,
		RSSI:        rssi,
		Connected:   false,
		State:       NewDeviceState(),
		LastSeen:    time.Now(),
		LastUpdated: time.Now(),
	}
}

// UpdateState updates the device state and timestamp
func (d *Device) UpdateState(state DeviceState) {
	state.LastUpdated = time.Now()
	d.State = state
	d.LastUpdated = time.Now()
}

// MarkConnected marks the device as connected
func (d *Device) MarkConnected() {
	d.Connected = true
	d.LastSeen = time.Now()
}

// MarkDisconnected marks the device as disconnected
func (d *Device) MarkDisconnected() {
	d.Connected = false
}

// IsConnected returns whether the device is currently connected
func (d *Device) IsConnected() bool {
	return d.Connected
}

// Validate validates the device data
func (d *Device) Validate() error {
	if d.Address == "" {
		return ErrInvalidAddress
	}
	return nil
}

// String returns a string representation of the device
func (d *Device) String() string {
	return fmt.Sprintf("Device{Address: %s, Name: %s, Connected: %t}",
		d.Address, d.Name, d.Connected)
}
