package twitch

import (
	"context"
	"log"
	"sync"

	"github.com/codeneuss/lampcontrol/internal/domain"
	"github.com/gempir/go-twitch-irc/v4"
)

// MessageHandler is called when a chat message is received
type MessageHandler func(cmd *domain.TwitchCommand)

// IRCClient wraps Twitch IRC functionality
type IRCClient struct {
	client         *twitch.Client
	channel        string
	messageHandler MessageHandler
	connected      bool
	mu             sync.RWMutex
}

// NewIRCClient creates a new Twitch IRC client
func NewIRCClient(username, token, channel string, handler MessageHandler) *IRCClient {
	client := twitch.NewClient(username, token)

	ircClient := &IRCClient{
		client:         client,
		channel:        channel,
		messageHandler: handler,
		connected:      false,
	}

	// Set up message handler
	client.OnPrivateMessage(ircClient.onMessage)

	// Set up connection handlers
	client.OnConnect(func() {
		log.Printf("[Twitch] Connected to channel: %s", channel)
		ircClient.mu.Lock()
		ircClient.connected = true
		ircClient.mu.Unlock()
	})

	return ircClient
}

// Connect connects to Twitch IRC
func (c *IRCClient) Connect(ctx context.Context) error {
	c.client.Join(c.channel)

	go func() {
		if err := c.client.Connect(); err != nil {
			log.Printf("[Twitch] Connection error: %v", err)
		}
	}()

	return nil
}

// Disconnect disconnects from Twitch IRC
func (c *IRCClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		c.client.Disconnect()
		c.connected = false
	}

	return nil
}

// IsConnected returns connection status
func (c *IRCClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// SendMessage sends a message to chat
func (c *IRCClient) SendMessage(message string) {
	c.client.Say(c.channel, message)
}

// onMessage handles incoming chat messages
func (c *IRCClient) onMessage(message twitch.PrivateMessage) {
	// Parse command
	command, err := domain.ParseTwitchCommand(message.Message)
	if err != nil {
		return // Not a lamp command
	}

	// Extract user badges
	badges := extractBadges(message)

	// Create command
	cmd := &domain.TwitchCommand{
		Username:    message.User.Name,
		DisplayName: message.User.DisplayName,
		Command:     command,
		IsVIP:       badges.IsVIP,
		IsSub:       badges.IsSub,
		IsMod:       badges.IsMod,
		Timestamp:   message.Time,
	}

	// Call handler
	if c.messageHandler != nil {
		c.messageHandler(cmd)
	}
}

// extractBadges extracts user privilege information
func extractBadges(message twitch.PrivateMessage) domain.UserBadges {
	badges := domain.UserBadges{}

	for badge := range message.User.Badges {
		switch badge {
		case "vip":
			badges.IsVIP = true
		case "subscriber":
			badges.IsSub = true
		case "moderator":
			badges.IsMod = true
		case "broadcaster":
			badges.IsMod = true // Broadcaster has mod privileges
		}
	}

	return badges
}
