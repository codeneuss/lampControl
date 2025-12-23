package dto

import (
	"time"

	"github.com/codeneuss/lampcontrol/internal/domain"
)

// DeviceDTO represents a device for API responses
type DeviceDTO struct {
	Address     string           `json:"address"`
	Name        string           `json:"name"`
	RSSI        int16            `json:"rssi"`
	Connected   bool             `json:"connected"`
	State       DeviceStateDTO   `json:"state"`
	LastSeen    time.Time        `json:"last_seen"`
	LastUpdated time.Time        `json:"last_updated"`
}

// DeviceStateDTO represents device state for API responses
type DeviceStateDTO struct {
	PowerOn      bool                  `json:"power_on"`
	Brightness   uint8                 `json:"brightness"`
	RGB          *domain.RGB           `json:"rgb,omitempty"`
	WhiteBalance *domain.WhiteBalance  `json:"white_balance,omitempty"`
	Effect       *int                  `json:"effect,omitempty"`
	EffectSpeed  *uint8                `json:"effect_speed,omitempty"`
	LastUpdated  time.Time             `json:"last_updated"`
}

// ScanRequestDTO represents a request to scan for devices
type ScanRequestDTO struct {
	Timeout string `json:"timeout"` // Duration string (e.g., "10s")
}

// SelectDeviceRequestDTO represents a request to select a device
type SelectDeviceRequestDTO struct {
	Address string `json:"address"` // Device MAC address
}

// HealthResponseDTO represents health check response
type HealthResponseDTO struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// FromDomain converts domain.Device to DeviceDTO
func FromDomain(device *domain.Device) DeviceDTO {
	return DeviceDTO{
		Address:     device.Address,
		Name:        device.Name,
		RSSI:        device.RSSI,
		Connected:   device.Connected,
		State:       FromDomainState(device.State),
		LastSeen:    device.LastSeen,
		LastUpdated: device.LastUpdated,
	}
}

// FromDomainState converts domain.DeviceState to DeviceStateDTO
func FromDomainState(state domain.DeviceState) DeviceStateDTO {
	return DeviceStateDTO{
		PowerOn:      state.PowerOn,
		Brightness:   state.Brightness,
		RGB:          state.RGB,
		WhiteBalance: state.WhiteBalance,
		Effect:       state.Effect,
		EffectSpeed:  state.EffectSpeed,
		LastUpdated:  state.LastUpdated,
	}
}

// FromDomainList converts a list of domain.Device to DeviceDTO list
func FromDomainList(devices []*domain.Device) []DeviceDTO {
	dtos := make([]DeviceDTO, len(devices))
	for i, device := range devices {
		dtos[i] = FromDomain(device)
	}
	return dtos
}
