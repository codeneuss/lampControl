package domain

import "time"

// StateSnapshot captures the lamp state for restoration
type StateSnapshot struct {
	DeviceAddress string      `json:"device_address"`
	State         DeviceState `json:"state"`
	CapturedAt    time.Time   `json:"captured_at"`
	Reason        string      `json:"reason"` // "twitch_viewer_command"
}

// NewStateSnapshot creates a new state snapshot
func NewStateSnapshot(deviceAddr string, state DeviceState, reason string) *StateSnapshot {
	return &StateSnapshot{
		DeviceAddress: deviceAddr,
		State:         state,
		CapturedAt:    time.Now(),
		Reason:        reason,
	}
}
