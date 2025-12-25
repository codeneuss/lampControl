package domain

import (
	"time"
)

// CooldownState tracks cooldowns
type CooldownState struct {
	LastGlobalCommand time.Time
	UserLastCommand   map[string]time.Time // username -> last command time
}

// NewCooldownState creates a new cooldown state
func NewCooldownState() *CooldownState {
	return &CooldownState{
		UserLastCommand: make(map[string]time.Time),
	}
}

// CheckGlobalCooldown checks if global cooldown has expired
func (c *CooldownState) CheckGlobalCooldown(cooldown time.Duration) (bool, time.Duration) {
	if cooldown == 0 {
		return true, 0
	}

	elapsed := time.Since(c.LastGlobalCommand)
	if elapsed < cooldown {
		return false, cooldown - elapsed
	}

	return true, 0
}

// CheckUserCooldown checks if user cooldown has expired
func (c *CooldownState) CheckUserCooldown(username string, cooldown time.Duration) (bool, time.Duration) {
	if cooldown == 0 {
		return true, 0
	}

	lastCmd, exists := c.UserLastCommand[username]
	if !exists {
		return true, 0
	}

	elapsed := time.Since(lastCmd)
	if elapsed < cooldown {
		return false, cooldown - elapsed
	}

	return true, 0
}

// RecordCommand records a command execution
func (c *CooldownState) RecordCommand(username string) {
	now := time.Now()
	c.LastGlobalCommand = now
	c.UserLastCommand[username] = now
}

// Reset resets all cooldowns
func (c *CooldownState) Reset() {
	c.LastGlobalCommand = time.Time{}
	c.UserLastCommand = make(map[string]time.Time)
}
