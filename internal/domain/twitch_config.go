package domain

import (
	"fmt"
	"time"
)

// TwitchConfig represents Twitch integration configuration
type TwitchConfig struct {
	Enabled      bool      `json:"enabled"`
	Channel      string    `json:"channel"`       // Twitch channel name (without #)
	BotUsername  string    `json:"bot_username"`  // Bot username
	AccessToken  string    `json:"access_token"`  // OAuth token (encrypted in storage)
	RefreshToken string    `json:"refresh_token"` // OAuth refresh token (encrypted in storage)
	TokenExpiry  time.Time `json:"token_expiry"`

	// Effect duration
	EffectDuration time.Duration `json:"effect_duration"` // How long viewer effects last (default: 30s)

	// Cooldown settings
	GlobalCooldown time.Duration `json:"global_cooldown"` // Cooldown between ANY commands (default: 5s)
	UserCooldown   time.Duration `json:"user_cooldown"`   // Per-user cooldown (default: 30s)

	// Privilege settings
	VIPBypassCooldown bool `json:"vip_bypass_cooldown"` // VIPs bypass cooldown
	SubBypassCooldown bool `json:"sub_bypass_cooldown"` // Subscribers bypass cooldown
	ModBypassCooldown bool `json:"mod_bypass_cooldown"` // Moderators bypass cooldown

	UpdatedAt time.Time `json:"updated_at"`
}

// NewTwitchConfig creates default Twitch configuration
func NewTwitchConfig() *TwitchConfig {
	return &TwitchConfig{
		Enabled:           false,
		EffectDuration:    30 * time.Second,
		GlobalCooldown:    5 * time.Second,
		UserCooldown:      30 * time.Second,
		VIPBypassCooldown: true,
		SubBypassCooldown: true,
		ModBypassCooldown: true,
		UpdatedAt:         time.Now(),
	}
}

// Validate validates the Twitch configuration
func (c *TwitchConfig) Validate() error {
	if c.Enabled {
		if c.Channel == "" {
			return fmt.Errorf("channel name is required")
		}
		if c.BotUsername == "" {
			return fmt.Errorf("bot username is required")
		}
		if c.AccessToken == "" {
			return fmt.Errorf("access token is required")
		}
	}

	if c.EffectDuration < 5*time.Second {
		return fmt.Errorf("effect duration must be at least 5 seconds")
	}

	if c.GlobalCooldown < 0 {
		return fmt.Errorf("global cooldown cannot be negative")
	}

	if c.UserCooldown < 0 {
		return fmt.Errorf("user cooldown cannot be negative")
	}

	return nil
}
