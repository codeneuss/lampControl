package bluetooth

import "errors"

// Infrastructure errors - errors related to BLE operations
var (
	// Connection errors
	ErrConnectionTimeout  = errors.New("connection timeout")
	ErrConnectionFailed   = errors.New("connection failed")
	ErrWriteFailed        = errors.New("characteristic write failed")
	ErrDisconnectFailed   = errors.New("disconnect failed")

	// Scanning errors
	ErrScanFailed         = errors.New("device scan failed")
	ErrScanTimeout        = errors.New("scan timeout")
	ErrNoDevicesFound     = errors.New("no devices found")

	// Adapter errors
	ErrAdapterNotReady    = errors.New("bluetooth adapter not ready")
	ErrAdapterEnableFailed = errors.New("failed to enable bluetooth adapter")

	// Service/Characteristic errors
	ErrServiceNotFound    = errors.New("service not found")
	ErrCharacteristicNotFound = errors.New("characteristic not found")
)
