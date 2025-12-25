package application

import (
	"sync"

	"github.com/codeneuss/lampcontrol/internal/domain"
)

// StateSnapshotService manages state snapshots for restoration
type StateSnapshotService struct {
	snapshots map[string]*domain.StateSnapshot // deviceAddr -> snapshot
	mu        sync.RWMutex
}

// NewStateSnapshotService creates a new state snapshot service
func NewStateSnapshotService() *StateSnapshotService {
	return &StateSnapshotService{
		snapshots: make(map[string]*domain.StateSnapshot),
	}
}

// SaveSnapshot saves a device state snapshot
func (s *StateSnapshotService) SaveSnapshot(deviceAddr string, state domain.DeviceState, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.snapshots[deviceAddr] = domain.NewStateSnapshot(deviceAddr, state, reason)
}

// GetLatestSnapshot retrieves the latest snapshot for a device
func (s *StateSnapshotService) GetLatestSnapshot(deviceAddr string) *domain.StateSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.snapshots[deviceAddr]
}

// ClearSnapshot clears the snapshot for a device
func (s *StateSnapshotService) ClearSnapshot(deviceAddr string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.snapshots, deviceAddr)
}
