package domain

import "errors"

// Domain errors - recoverable errors related to business logic validation
var (
	// Device errors
	ErrDeviceNotFound    = errors.New("device not found")
	ErrDeviceDisconnected = errors.New("device disconnected")
	ErrDeviceInUse       = errors.New("device already in use")

	// Validation errors
	ErrInvalidColor      = errors.New("invalid color value (must be 0-255)")
	ErrInvalidBrightness = errors.New("invalid brightness (must be 0-255)")
	ErrInvalidAddress    = errors.New("invalid device address")
	ErrInvalidEffect     = errors.New("invalid effect index")
	ErrInvalidSpeed      = errors.New("invalid speed value (must be 0-255)")

	// State errors
	ErrDeviceNotReady    = errors.New("device not ready")
	ErrInvalidState      = errors.New("invalid device state")
)
