package dto

import (
	"time"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/domain"
)

// TwitchConfigDTO represents Twitch configuration for API
type TwitchConfigDTO struct {
	Enabled bool   `json:"enabled"`
	Channel string `json:"channel"`
	BotUsername string `json:"bot_username"`
	HasToken bool   `json:"has_token"` // Don't expose actual token

	EffectDurationSec int `json:"effect_duration_sec"`
	GlobalCooldownSec int `json:"global_cooldown_sec"`
	UserCooldownSec   int `json:"user_cooldown_sec"`

	VIPBypassCooldown bool `json:"vip_bypass_cooldown"`
	SubBypassCooldown bool `json:"sub_bypass_cooldown"`
	ModBypassCooldown bool `json:"mod_bypass_cooldown"`
}

// TwitchConfigUpdateDTO represents update request
type TwitchConfigUpdateDTO struct {
	Enabled       *bool   `json:"enabled,omitempty"`
	Channel       *string `json:"channel,omitempty"`
	BotUsername   *string `json:"bot_username,omitempty"`
	AccessToken   *string `json:"access_token,omitempty"` // Only for updates

	EffectDurationSec *int `json:"effect_duration_sec,omitempty"`
	GlobalCooldownSec *int `json:"global_cooldown_sec,omitempty"`
	UserCooldownSec   *int `json:"user_cooldown_sec,omitempty"`

	VIPBypassCooldown *bool `json:"vip_bypass_cooldown,omitempty"`
	SubBypassCooldown *bool `json:"sub_bypass_cooldown,omitempty"`
	ModBypassCooldown *bool `json:"mod_bypass_cooldown,omitempty"`
}

// TwitchStatusDTO represents Twitch connection status
type TwitchStatusDTO struct {
	Connected    bool             `json:"connected"`
	Channel      string           `json:"channel,omitempty"`
	ActiveEffect *ActiveEffectDTO `json:"active_effect,omitempty"`
}

// ActiveEffectDTO represents currently active viewer effect
type ActiveEffectDTO struct {
	Username         string `json:"username"`
	Command          string `json:"command"`
	StartedAt        string `json:"started_at"`
	RemainingTimeSec int    `json:"remaining_time_sec"`
}

// TwitchCommandListDTO represents available commands
type TwitchCommandListDTO struct {
	Colors  []string `json:"colors"`
	Effects []string `json:"effects"`
}

// FromDomainTwitchConfig converts domain config to DTO
func FromDomainTwitchConfig(config *domain.TwitchConfig) TwitchConfigDTO {
	return TwitchConfigDTO{
		Enabled:           config.Enabled,
		Channel:           config.Channel,
		BotUsername:       config.BotUsername,
		HasToken:          config.AccessToken != "",
		EffectDurationSec: int(config.EffectDuration.Seconds()),
		GlobalCooldownSec: int(config.GlobalCooldown.Seconds()),
		UserCooldownSec:   int(config.UserCooldown.Seconds()),
		VIPBypassCooldown: config.VIPBypassCooldown,
		SubBypassCooldown: config.SubBypassCooldown,
		ModBypassCooldown: config.ModBypassCooldown,
	}
}

// ApplyUpdate applies update DTO to domain config
func (dto *TwitchConfigUpdateDTO) ApplyUpdate(config *domain.TwitchConfig) {
	if dto.Enabled != nil {
		config.Enabled = *dto.Enabled
	}
	if dto.Channel != nil {
		config.Channel = *dto.Channel
	}
	if dto.BotUsername != nil {
		config.BotUsername = *dto.BotUsername
	}
	if dto.AccessToken != nil {
		config.AccessToken = *dto.AccessToken
	}
	if dto.EffectDurationSec != nil {
		config.EffectDuration = time.Duration(*dto.EffectDurationSec) * time.Second
	}
	if dto.GlobalCooldownSec != nil {
		config.GlobalCooldown = time.Duration(*dto.GlobalCooldownSec) * time.Second
	}
	if dto.UserCooldownSec != nil {
		config.UserCooldown = time.Duration(*dto.UserCooldownSec) * time.Second
	}
	if dto.VIPBypassCooldown != nil {
		config.VIPBypassCooldown = *dto.VIPBypassCooldown
	}
	if dto.SubBypassCooldown != nil {
		config.SubBypassCooldown = *dto.SubBypassCooldown
	}
	if dto.ModBypassCooldown != nil {
		config.ModBypassCooldown = *dto.ModBypassCooldown
	}

	config.UpdatedAt = time.Now()
}

// FromActiveEffect converts active effect to DTO
func FromActiveEffect(effect *application.ActiveEffect, duration time.Duration) *ActiveEffectDTO {
	if effect == nil {
		return nil
	}

	elapsed := time.Since(effect.StartedAt)
	remaining := duration - elapsed
	if remaining < 0 {
		remaining = 0
	}

	return &ActiveEffectDTO{
		Username:         effect.Username,
		Command:          effect.Command,
		StartedAt:        effect.StartedAt.Format(time.RFC3339),
		RemainingTimeSec: int(remaining.Seconds()),
	}
}
