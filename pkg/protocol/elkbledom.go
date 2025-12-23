package protocol

// ELK-BLEDOM Protocol Implementation
// Reference: https://github.com/FergusInLondon/ELK-BLEDOM/blob/master/PROTCOL.md

// BLE Service and Characteristic UUIDs
const (
	ServiceUUID        = "0000fff0-0000-1000-8000-00805f9b34fb"
	CharacteristicUUID = "0000fff3-0000-1000-8000-00805f9b34fb"
)

// Command frame constants
const (
	StartByte = 0x7E // Start of frame
	EndByte   = 0xEF // End of frame
	SeqByte   = 0x00 // Sequence byte (always 0x00 in basic implementation)
)

// Command codes
const (
	CmdBrightness = 0x01 // Set brightness level
	CmdPower      = 0x04 // Power on/off
	CmdColor      = 0x05 // Set color (RGB, single color, or white balance)
	CmdEffect     = 0x03 // Built-in effects/scenes
	CmdCustom     = 0x06 // Custom program
)

// Color modes
const (
	ColorModeSingle = 0x01 // Single preset color
	ColorModeWhite  = 0x02 // White balance (warm/cold)
	ColorModeRGB    = 0x03 // RGB color
)

// Power states
const (
	PowerOff = 0x00
	PowerOn  = 0x01
)

// Command represents a 9-byte ELK-BLEDOM command packet
type Command [9]byte

func NewPowerCommand(on bool) Command {
	var cmd Command
	cmd[0] = StartByte
	cmd[1] = SeqByte
	cmd[2] = CmdPower

	if on {
		cmd[3] = 0xF0
		cmd[4] = 0x00
		cmd[5] = 0x01
	} else {
		cmd[3] = 0x00
		cmd[4] = 0x00
		cmd[5] = 0x00
	}

	cmd[6] = 0xFF
	cmd[7] = 0x00
	cmd[8] = EndByte
	return cmd
}

// NewRGBCommand creates an RGB color command
// Example: Red = [0x7E, 0x00, 0x05, 0x03, 0xFF, 0x00, 0x00, 0x00, 0xEF]
//
//	Green = [0x7E, 0x00, 0x05, 0x03, 0x00, 0xFF, 0x00, 0x00, 0xEF]
//	Blue = [0x7E, 0x00, 0x05, 0x03, 0x00, 0x00, 0xFF, 0x00, 0xEF]
func NewRGBCommand(r, g, b uint8) Command {
	return Command{
		StartByte,
		SeqByte,
		CmdColor,
		ColorModeRGB,
		r, g, b,
		0x00,
		EndByte,
	}
}

// NewBrightnessCommand creates a brightness command
// level: 0-255 (0x00-0xFF)
// Example: 100% brightness = [0x7E, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0xEF]
//
//	50% brightness = [0x7E, 0x00, 0x01, 0x7F, 0xFF, 0xFF, 0xFF, 0x00, 0xEF]
func NewBrightnessCommand(level uint8) Command {
	return Command{
		StartByte,
		SeqByte,
		CmdBrightness,
		level,
		0xFF,
		0xFF,
		0xFF,
		0x00,
		EndByte,
	}
}

// NewWhiteBalanceCommand creates a white balance command
// warm: 0-255 (warm white intensity)
// cold: 0-255 (cold white intensity)
// Example: 50/50 = [0x7E, 0x00, 0x05, 0x02, 0x50, 0x50, 0xFF, 0x00, 0xEF]
func NewWhiteBalanceCommand(warm, cold uint8) Command {
	return Command{
		StartByte,
		SeqByte,
		CmdColor,
		ColorModeWhite,
		warm,
		cold,
		0xFF,
		0x00,
		EndByte,
	}
}

// NewSingleColorCommand creates a single preset color command
// index: Color index (0-255)
// Example: Color 0 = [0x7E, 0x00, 0x05, 0x01, 0x00, 0xFF, 0xFF, 0x00, 0xEF]
func NewSingleColorCommand(index uint8) Command {
	return Command{
		StartByte,
		SeqByte,
		CmdColor,
		ColorModeSingle,
		index,
		0xFF,
		0xFF,
		0x00,
		EndByte,
	}
}

// NewEffectCommand creates a built-in effect/scene command
// effect: Effect index (0-255)
// speed: Effect speed (0-255, higher is faster)
// Example: Effect 1, speed 50 = [0x7E, 0x00, 0x03, 0x01, 0x32, 0xFF, 0xFF, 0x00, 0xEF]
func NewEffectCommand(effect, speed uint8) Command {
	return Command{
		StartByte,
		SeqByte,
		CmdEffect,
		effect,
		speed,
		0xFF,
		0xFF,
		0x00,
		EndByte,
	}
}

// Bytes returns the command as a byte slice
func (c Command) Bytes() []byte {
	return c[:]
}

// String returns a hex representation of the command for debugging
func (c Command) String() string {
	return bytesToHex(c[:])
}

// bytesToHex converts bytes to hex string for debugging
func bytesToHex(data []byte) string {
	hex := ""
	for i, b := range data {
		if i > 0 {
			hex += " "
		}
		hex += byteToHex(b)
	}
	return hex
}

// byteToHex converts a single byte to hex string
func byteToHex(b byte) string {
	const hexChars = "0123456789ABCDEF"
	return string([]byte{hexChars[b>>4], hexChars[b&0x0F]})
}
