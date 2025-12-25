package application

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/codeneuss/lampcontrol/internal/domain"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/storage"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/twitch"
)

// TwitchService orchestrates Twitch chat integration
type TwitchService struct {
	deviceService   *DeviceService
	snapshotService *StateSnapshotService
	storage         *storage.TwitchStorage
	ircClient       *twitch.IRCClient
	cooldownManager *CooldownManager

	activeEffect *ActiveEffect
	mu           sync.RWMutex

	// Callbacks
	onStatusChange    func(connected bool)
	onCommandSuccess  func(username, command string)
	getSelectedDevice func() (string, error)
}

// ActiveEffect tracks currently active viewer effect
type ActiveEffect struct {
	Username  string
	Command   string
	StartedAt time.Time
	Timer     *time.Timer
}

// NewTwitchService creates a new Twitch service
func NewTwitchService(
	deviceService *DeviceService,
	storage *storage.TwitchStorage,
) *TwitchService {
	return &TwitchService{
		deviceService:   deviceService,
		snapshotService: NewStateSnapshotService(),
		storage:         storage,
		cooldownManager: NewCooldownManager(),
	}
}

// Start starts the Twitch integration
func (s *TwitchService) Start(ctx context.Context) error {
	config := s.storage.Get()

	if !config.Enabled {
		return fmt.Errorf("twitch integration is disabled")
	}

	if err := config.Validate(); err != nil {
		return err
	}

	// Create IRC client
	s.ircClient = twitch.NewIRCClient(
		config.BotUsername,
		config.AccessToken,
		config.Channel,
		s.handleCommand,
	)

	fmt.Println(config)

	// Connect to Twitch
	if err := s.ircClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to Twitch: %w", err)
	}

	log.Printf("[Twitch] Started integration for channel: %s", config.Channel)

	if s.onStatusChange != nil {
		s.onStatusChange(true)
	}

	return nil
}

// Stop stops the Twitch integration
func (s *TwitchService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel active effect timer
	if s.activeEffect != nil && s.activeEffect.Timer != nil {
		s.activeEffect.Timer.Stop()
		s.activeEffect = nil
	}

	// Disconnect from IRC
	if s.ircClient != nil {
		if err := s.ircClient.Disconnect(); err != nil {
			return err
		}
	}

	if s.onStatusChange != nil {
		s.onStatusChange(false)
	}

	return nil
}

// IsConnected returns connection status
func (s *TwitchService) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ircClient != nil && s.ircClient.IsConnected()
}

// handleCommand processes a Twitch chat command
func (s *TwitchService) handleCommand(cmd *domain.TwitchCommand) {
	config := s.storage.Get()

	// Check if user bypasses cooldown
	bypassCooldown := (cmd.IsVIP && config.VIPBypassCooldown) ||
		(cmd.IsSub && config.SubBypassCooldown) ||
		(cmd.IsMod && config.ModBypassCooldown)

	// Check cooldowns
	if !bypassCooldown {
		if ok, remaining := s.cooldownManager.CheckGlobal(config.GlobalCooldown); !ok {
			s.sendCooldownMessage(cmd.Username, remaining, "global")
			return
		}

		if ok, remaining := s.cooldownManager.CheckUser(cmd.Username, config.UserCooldown); !ok {
			s.sendCooldownMessage(cmd.Username, remaining, "personal")
			return
		}
	}

	// Execute command
	if err := s.executeCommand(cmd, config); err != nil {
		log.Printf("[Twitch] Command failed for %s: %v", cmd.Username, err)
		s.ircClient.SendMessage(fmt.Sprintf("@%s Sorry, that command failed: %v", cmd.DisplayName, err))
		return
	}

	// Record cooldown
	s.cooldownManager.RecordCommand(cmd.Username)

	// Send success message
	s.ircClient.SendMessage(fmt.Sprintf("@%s Lamp set to %s for %d seconds!",
		cmd.DisplayName, cmd.Command, int(config.EffectDuration.Seconds())))

	if s.onCommandSuccess != nil {
		s.onCommandSuccess(cmd.Username, cmd.Command)
	}
}

// executeCommand executes a lamp command
func (s *TwitchService) executeCommand(cmd *domain.TwitchCommand, config *domain.TwitchConfig) error {
	ctx := context.Background()

	// Get selected device
	if s.getSelectedDevice == nil {
		return fmt.Errorf("no device selection callback configured")
	}

	deviceAddr, err := s.getSelectedDevice()
	if err != nil || deviceAddr == "" {
		return fmt.Errorf("no device selected")
	}

	// Cancel existing effect timer if any
	s.mu.Lock()
	if s.activeEffect != nil && s.activeEffect.Timer != nil {
		s.activeEffect.Timer.Stop()
	}
	s.mu.Unlock()

	// Save current state (only if no active effect)
	s.mu.RLock()
	shouldSnapshot := s.activeEffect == nil
	s.mu.RUnlock()

	if shouldSnapshot {
		device, err := s.deviceService.GetDevice(deviceAddr)
		if err != nil {
			return err
		}
		s.snapshotService.SaveSnapshot(deviceAddr, device.State, "twitch_viewer_command")
	}

	// Execute the command
	if domain.IsColor(cmd.Command) {
		rgb, _ := domain.GetRGB(cmd.Command)
		if err := s.deviceService.SetColor(ctx, deviceAddr, rgb.R, rgb.G, rgb.B); err != nil {
			return err
		}
	} else if domain.IsEffect(cmd.Command) {
		effect, _ := domain.GetEffect(cmd.Command)
		if err := s.deviceService.SetEffect(ctx, deviceAddr, effect, 128); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown command: %s", cmd.Command)
	}

	// Set timer to restore state
	timer := time.AfterFunc(config.EffectDuration, func() {
		s.restoreStreamerState(deviceAddr)
	})

	s.mu.Lock()
	s.activeEffect = &ActiveEffect{
		Username:  cmd.Username,
		Command:   cmd.Command,
		StartedAt: time.Now(),
		Timer:     timer,
	}
	s.mu.Unlock()

	return nil
}

// restoreStreamerState restores the saved state
func (s *TwitchService) restoreStreamerState(deviceAddr string) {
	snapshot := s.snapshotService.GetLatestSnapshot(deviceAddr)
	if snapshot == nil {
		log.Printf("[Twitch] No snapshot to restore for device: %s", deviceAddr)
		return
	}

	ctx := context.Background()
	state := snapshot.State

	// Restore state based on what was active
	if state.RGB != nil {
		s.deviceService.SetColor(ctx, deviceAddr, state.RGB.R, state.RGB.G, state.RGB.B)
	} else if state.WhiteBalance != nil {
		s.deviceService.SetWhiteBalance(ctx, deviceAddr, state.WhiteBalance.Warm, state.WhiteBalance.Cold)
	} else if state.Effect != nil {
		speed := uint8(128)
		if state.EffectSpeed != nil {
			speed = *state.EffectSpeed
		}
		s.deviceService.SetEffect(ctx, deviceAddr, uint8(*state.Effect), speed)
	}

	log.Printf("[Twitch] Restored state for device: %s", deviceAddr)

	s.mu.Lock()
	s.activeEffect = nil
	s.mu.Unlock()
}

// sendCooldownMessage sends a cooldown message to chat
func (s *TwitchService) sendCooldownMessage(username string, remaining time.Duration, cooldownType string) {
	seconds := int(remaining.Seconds())
	if s.ircClient != nil {
		s.ircClient.SendMessage(fmt.Sprintf("@%s Please wait %d seconds (%s cooldown)",
			username, seconds, cooldownType))
	}
}

// SetStatusChangeCallback sets callback for connection status changes
func (s *TwitchService) SetStatusChangeCallback(callback func(bool)) {
	s.onStatusChange = callback
}

// SetCommandSuccessCallback sets callback for successful commands
func (s *TwitchService) SetCommandSuccessCallback(callback func(string, string)) {
	s.onCommandSuccess = callback
}

// SetGetSelectedDeviceFunc sets the function to get selected device address
func (s *TwitchService) SetGetSelectedDeviceFunc(fn func() (string, error)) {
	s.getSelectedDevice = fn
}

// GetActiveEffect returns the currently active effect
func (s *TwitchService) GetActiveEffect() *ActiveEffect {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.activeEffect
}
