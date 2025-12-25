package application

import (
	"sync"
	"time"

	"github.com/codeneuss/lampcontrol/internal/domain"
)

// CooldownManager manages command cooldowns
type CooldownManager struct {
	state *domain.CooldownState
	mu    sync.RWMutex
}

// NewCooldownManager creates a new cooldown manager
func NewCooldownManager() *CooldownManager {
	return &CooldownManager{
		state: domain.NewCooldownState(),
	}
}

// CheckGlobal checks global cooldown
func (m *CooldownManager) CheckGlobal(cooldown time.Duration) (bool, time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.state.CheckGlobalCooldown(cooldown)
}

// CheckUser checks user cooldown
func (m *CooldownManager) CheckUser(username string, cooldown time.Duration) (bool, time.Duration) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.state.CheckUserCooldown(username, cooldown)
}

// RecordCommand records a command execution
func (m *CooldownManager) RecordCommand(username string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.RecordCommand(username)
}

// Reset resets all cooldowns
func (m *CooldownManager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.Reset()
}
