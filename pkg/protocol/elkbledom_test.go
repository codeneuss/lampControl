package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPowerCommand(t *testing.T) {
	tests := []struct {
		name     string
		on       bool
		expected Command
	}{
		{
			name: "power on",
			on:   true,
			expected: Command{
				0x7E, 0x00, 0x04, 0xF0, 0x00, 0x01, 0xFF, 0x00, 0xEF,
			},
		},
		{
			name: "power off",
			on:   false,
			expected: Command{
				0x7E, 0x00, 0x04, 0x00, 0x00, 0x00, 0xFF, 0x00, 0xEF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewPowerCommand(tt.on)
			assert.Equal(t, tt.expected, cmd)
		})
	}
}

func TestNewRGBCommand(t *testing.T) {
	tests := []struct {
		name     string
		r, g, b  uint8
		expected Command
	}{
		{
			name: "red color",
			r:    255, g: 0, b: 0,
			expected: Command{
				0x7E, 0x00, 0x05, 0x03, 0xFF, 0x00, 0x00, 0x00, 0xEF,
			},
		},
		{
			name: "green color",
			r:    0, g: 255, b: 0,
			expected: Command{
				0x7E, 0x00, 0x05, 0x03, 0x00, 0xFF, 0x00, 0x00, 0xEF,
			},
		},
		{
			name: "blue color",
			r:    0, g: 0, b: 255,
			expected: Command{
				0x7E, 0x00, 0x05, 0x03, 0x00, 0x00, 0xFF, 0x00, 0xEF,
			},
		},
		{
			name: "white color",
			r:    255, g: 255, b: 255,
			expected: Command{
				0x7E, 0x00, 0x05, 0x03, 0xFF, 0xFF, 0xFF, 0x00, 0xEF,
			},
		},
		{
			name: "purple color",
			r:    128, g: 0, b: 128,
			expected: Command{
				0x7E, 0x00, 0x05, 0x03, 0x80, 0x00, 0x80, 0x00, 0xEF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRGBCommand(tt.r, tt.g, tt.b)
			assert.Equal(t, tt.expected, cmd)
		})
	}
}

func TestNewBrightnessCommand(t *testing.T) {
	tests := []struct {
		name     string
		level    uint8
		expected Command
	}{
		{
			name:  "full brightness",
			level: 255,
			expected: Command{
				0x7E, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0xEF,
			},
		},
		{
			name:  "half brightness",
			level: 127,
			expected: Command{
				0x7E, 0x00, 0x01, 0x7F, 0xFF, 0xFF, 0xFF, 0x00, 0xEF,
			},
		},
		{
			name:  "minimum brightness",
			level: 0,
			expected: Command{
				0x7E, 0x00, 0x01, 0x00, 0xFF, 0xFF, 0xFF, 0x00, 0xEF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewBrightnessCommand(tt.level)
			assert.Equal(t, tt.expected, cmd)
		})
	}
}

func TestNewWhiteBalanceCommand(t *testing.T) {
	tests := []struct {
		name       string
		warm, cold uint8
		expected   Command
	}{
		{
			name: "50/50 balance",
			warm: 128, cold: 128,
			expected: Command{
				0x7E, 0x00, 0x05, 0x02, 0x80, 0x80, 0xFF, 0x00, 0xEF,
			},
		},
		{
			name: "full warm",
			warm: 255, cold: 0,
			expected: Command{
				0x7E, 0x00, 0x05, 0x02, 0xFF, 0x00, 0xFF, 0x00, 0xEF,
			},
		},
		{
			name: "full cold",
			warm: 0, cold: 255,
			expected: Command{
				0x7E, 0x00, 0x05, 0x02, 0x00, 0xFF, 0xFF, 0x00, 0xEF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewWhiteBalanceCommand(tt.warm, tt.cold)
			assert.Equal(t, tt.expected, cmd)
		})
	}
}

func TestNewSingleColorCommand(t *testing.T) {
	tests := []struct {
		name     string
		index    uint8
		expected Command
	}{
		{
			name:  "color index 0",
			index: 0,
			expected: Command{
				0x7E, 0x00, 0x05, 0x01, 0x00, 0xFF, 0xFF, 0x00, 0xEF,
			},
		},
		{
			name:  "color index 5",
			index: 5,
			expected: Command{
				0x7E, 0x00, 0x05, 0x01, 0x05, 0xFF, 0xFF, 0x00, 0xEF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewSingleColorCommand(tt.index)
			assert.Equal(t, tt.expected, cmd)
		})
	}
}

func TestNewEffectCommand(t *testing.T) {
	tests := []struct {
		name     string
		effect   uint8
		speed    uint8
		expected Command
	}{
		{
			name:   "effect 1, speed 50",
			effect: 1, speed: 50,
			expected: Command{
				0x7E, 0x00, 0x03, 0x01, 0x32, 0xFF, 0xFF, 0x00, 0xEF,
			},
		},
		{
			name:   "effect 10, speed 100",
			effect: 10, speed: 100,
			expected: Command{
				0x7E, 0x00, 0x03, 0x0A, 0x64, 0xFF, 0xFF, 0x00, 0xEF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewEffectCommand(tt.effect, tt.speed)
			assert.Equal(t, tt.expected, cmd)
		})
	}
}

func TestCommandBytes(t *testing.T) {
	cmd := NewPowerCommand(true)
	bytes := cmd.Bytes()

	assert.Equal(t, 9, len(bytes))
	assert.Equal(t, byte(0x7E), bytes[0])
	assert.Equal(t, byte(0xEF), bytes[8])
}

func TestCommandString(t *testing.T) {
	cmd := NewPowerCommand(true)
	str := cmd.String()

	// Should be hex representation
	assert.Contains(t, str, "7E")
	assert.Contains(t, str, "EF")
}

func TestCommandFrameStructure(t *testing.T) {
	// Test that all commands have correct frame structure
	commands := []Command{
		NewPowerCommand(true),
		NewRGBCommand(255, 0, 0),
		NewBrightnessCommand(128),
		NewWhiteBalanceCommand(128, 128),
		NewEffectCommand(1, 50),
	}

	for _, cmd := range commands {
		// Check start byte
		assert.Equal(t, byte(StartByte), cmd[0], "Start byte should be 0x7E")

		// Check sequence byte
		assert.Equal(t, byte(SeqByte), cmd[1], "Sequence byte should be 0x00")

		// Check end byte
		assert.Equal(t, byte(EndByte), cmd[8], "End byte should be 0xEF")
	}
}
