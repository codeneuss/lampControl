package dto

import "encoding/json"

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeCommand      MessageType = "command"
	MessageTypeStateUpdate  MessageType = "state_update"
	MessageTypeError        MessageType = "error"
	MessageTypeScanResult   MessageType = "scan_result"
	MessageTypeTwitchStatus MessageType = "twitch_status"
	MessageTypeTwitchCommand MessageType = "twitch_command"
)

// CommandAction represents the action to perform
type CommandAction string

const (
	CommandActionPower        CommandAction = "power"
	CommandActionColor        CommandAction = "color"
	CommandActionBrightness   CommandAction = "brightness"
	CommandActionWhiteBalance CommandAction = "white_balance"
	CommandActionEffect       CommandAction = "effect"
)

// CommandMessage represents a command from client to server
type CommandMessage struct {
	Type    MessageType     `json:"type"`
	Action  CommandAction   `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// PowerPayload represents power command payload
type PowerPayload struct {
	On bool `json:"on"`
}

// ColorPayload represents color command payload
type ColorPayload struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

// BrightnessPayload represents brightness command payload
type BrightnessPayload struct {
	Level uint8 `json:"level"`
}

// WhiteBalancePayload represents white balance command payload
type WhiteBalancePayload struct {
	Warm uint8 `json:"warm"`
	Cold uint8 `json:"cold"`
}

// EffectPayload represents effect command payload
type EffectPayload struct {
	Effect uint8 `json:"effect"`
	Speed  uint8 `json:"speed"`
}

// StateUpdateMessage represents a state update from server to client
type StateUpdateMessage struct {
	Type   MessageType `json:"type"`
	Device DeviceDTO   `json:"device"`
}

// ErrorMessage represents an error message from server to client
type ErrorMessage struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
	Code    string      `json:"code,omitempty"`
}

// ScanResultMessage represents scan results from server to client
type ScanResultMessage struct {
	Type    MessageType `json:"type"`
	Devices []DeviceDTO `json:"devices"`
}

// NewStateUpdateMessage creates a new state update message
func NewStateUpdateMessage(device DeviceDTO) StateUpdateMessage {
	return StateUpdateMessage{
		Type:   MessageTypeStateUpdate,
		Device: device,
	}
}

// NewErrorMessage creates a new error message
func NewErrorMessage(message, code string) ErrorMessage {
	return ErrorMessage{
		Type:    MessageTypeError,
		Message: message,
		Code:    code,
	}
}

// NewScanResultMessage creates a new scan result message
func NewScanResultMessage(devices []DeviceDTO) ScanResultMessage {
	return ScanResultMessage{
		Type:    MessageTypeScanResult,
		Devices: devices,
	}
}

// TwitchStatusMessage represents Twitch status update
type TwitchStatusMessage struct {
	Type   MessageType     `json:"type"`
	Status TwitchStatusDTO `json:"status"`
}

// TwitchCommandMessage represents a Twitch command execution
type TwitchCommandMessage struct {
	Type     MessageType `json:"type"`
	Username string      `json:"username"`
	Command  string      `json:"command"`
}

// NewTwitchStatusMessage creates a Twitch status message
func NewTwitchStatusMessage(status TwitchStatusDTO) TwitchStatusMessage {
	return TwitchStatusMessage{
		Type:   MessageTypeTwitchStatus,
		Status: status,
	}
}

// NewTwitchCommandMessage creates a Twitch command message
func NewTwitchCommandMessage(username, command string) TwitchCommandMessage {
	return TwitchCommandMessage{
		Type:     MessageTypeTwitchCommand,
		Username: username,
		Command:  command,
	}
}
