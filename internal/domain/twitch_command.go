package domain

import (
	"fmt"
	"strings"
	"time"
)

// TwitchCommand represents a parsed Twitch chat command
type TwitchCommand struct {
	Username    string
	DisplayName string
	Command     string // "red", "rainbow", etc.
	IsVIP       bool
	IsSub       bool
	IsMod       bool
	Timestamp   time.Time
}

// UserBadges represents user privileges
type UserBadges struct {
	IsVIP bool
	IsSub bool
	IsMod bool
}

// ColorMap maps color names to RGB values
var ColorMap = map[string]RGB{
	"red":     {R: 255, G: 0, B: 0},
	"green":   {R: 0, G: 255, B: 0},
	"blue":    {R: 0, G: 0, B: 255},
	"yellow":  {R: 255, G: 255, B: 0},
	"cyan":    {R: 0, G: 255, B: 255},
	"magenta": {R: 255, G: 0, B: 255},
	"purple":  {R: 128, G: 0, B: 128},
	"orange":  {R: 255, G: 165, B: 0},
	"pink":    {R: 255, G: 192, B: 203},
	"white":   {R: 255, G: 255, B: 255},
}

// EffectMap maps effect names to effect indices
// These are example indices - adjust based on ELK-BLEDOM actual effect numbers
var EffectMap = map[string]uint8{
	"rainbow": 0x25,
	"strobe":  0x26,
	"fade":    0x27,
	"pulse":   0x28,
}

// ParseTwitchCommand parses a chat message like "!lamp red"
func ParseTwitchCommand(message string) (string, error) {
	message = strings.TrimSpace(strings.ToLower(message))

	if !strings.HasPrefix(message, "!lamp ") {
		return "", fmt.Errorf("not a lamp command")
	}

	command := strings.TrimPrefix(message, "!lamp ")
	command = strings.TrimSpace(command)

	if command == "" {
		return "", fmt.Errorf("empty command")
	}

	return command, nil
}

// IsColor checks if command is a color
func IsColor(command string) bool {
	_, exists := ColorMap[strings.ToLower(command)]
	return exists
}

// IsEffect checks if command is an effect
func IsEffect(command string) bool {
	_, exists := EffectMap[strings.ToLower(command)]
	return exists
}

// GetRGB returns RGB values for a color command
func GetRGB(command string) (RGB, error) {
	rgb, exists := ColorMap[strings.ToLower(command)]
	if !exists {
		return RGB{}, fmt.Errorf("unknown color: %s", command)
	}
	return rgb, nil
}

// GetEffect returns effect index for an effect command
func GetEffect(command string) (uint8, error) {
	effect, exists := EffectMap[strings.ToLower(command)]
	if !exists {
		return 0, fmt.Errorf("unknown effect: %s", command)
	}
	return effect, nil
}
