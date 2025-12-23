package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/dto"
)

// ClientMessage represents a message from a client
type ClientMessage struct {
	client  *Client
	message []byte
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from clients
	process chan *ClientMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast channel for sending messages to all clients
	broadcast chan []byte

	// Device service for handling commands
	deviceService *application.DeviceService

	// Function to get selected device address
	getSelectedDevice func() (string, error)
}

// NewHub creates a new WebSocket hub
func NewHub(deviceService *application.DeviceService, getSelectedDevice func() (string, error)) *Hub {
	return &Hub{
		clients:           make(map[*Client]bool),
		process:           make(chan *ClientMessage, 256),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan []byte, 256),
		deviceService:     deviceService,
		getSelectedDevice: getSelectedDevice,
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected. Total clients: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			// Broadcast to all clients
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Client buffer full, disconnect
					close(client.send)
					delete(h.clients, client)
				}
			}

		case clientMsg := <-h.process:
			// Process command from client
			h.handleCommand(clientMsg.client, clientMsg.message)
		}
	}
}

// handleCommand processes a command from a client
func (h *Hub) handleCommand(client *Client, message []byte) {
	var cmd dto.CommandMessage
	if err := json.Unmarshal(message, &cmd); err != nil {
		log.Printf("Failed to unmarshal command: %v", err)
		client.SendJSON(dto.NewErrorMessage("Invalid command format", "INVALID_FORMAT"))
		return
	}

	// Get selected device address
	deviceAddr, err := h.getSelectedDevice()
	if err != nil {
		client.SendJSON(dto.NewErrorMessage("No device selected", "DEVICE_NOT_SELECTED"))
		return
	}

	ctx := context.Background()

	// Process command based on action
	switch cmd.Action {
	case dto.CommandActionPower:
		var payload dto.PowerPayload
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			client.SendJSON(dto.NewErrorMessage("Invalid power payload", "INVALID_PAYLOAD"))
			return
		}
		err = h.deviceService.SetPower(ctx, deviceAddr, payload.On)

	case dto.CommandActionColor:
		var payload dto.ColorPayload
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			client.SendJSON(dto.NewErrorMessage("Invalid color payload", "INVALID_PAYLOAD"))
			return
		}
		err = h.deviceService.SetColor(ctx, deviceAddr, payload.R, payload.G, payload.B)

	case dto.CommandActionBrightness:
		var payload dto.BrightnessPayload
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			client.SendJSON(dto.NewErrorMessage("Invalid brightness payload", "INVALID_PAYLOAD"))
			return
		}
		err = h.deviceService.SetBrightness(ctx, deviceAddr, payload.Level)

	case dto.CommandActionWhiteBalance:
		var payload dto.WhiteBalancePayload
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			client.SendJSON(dto.NewErrorMessage("Invalid white balance payload", "INVALID_PAYLOAD"))
			return
		}
		err = h.deviceService.SetWhiteBalance(ctx, deviceAddr, payload.Warm, payload.Cold)

	case dto.CommandActionEffect:
		var payload dto.EffectPayload
		if err := json.Unmarshal(cmd.Payload, &payload); err != nil {
			client.SendJSON(dto.NewErrorMessage("Invalid effect payload", "INVALID_PAYLOAD"))
			return
		}
		err = h.deviceService.SetEffect(ctx, deviceAddr, payload.Effect, payload.Speed)

	default:
		client.SendJSON(dto.NewErrorMessage("Unknown command action", "UNKNOWN_ACTION"))
		return
	}

	if err != nil {
		log.Printf("Command failed: %v", err)
		client.SendJSON(dto.NewErrorMessage(fmt.Sprintf("Command failed: %v", err), "COMMAND_FAILED"))
		return
	}

	// Broadcast updated state to all clients
	h.BroadcastDeviceState()
}

// BroadcastDeviceState sends the current device state to all clients
func (h *Hub) BroadcastDeviceState() {
	deviceAddr, err := h.getSelectedDevice()
	if err != nil {
		return
	}

	device, err := h.deviceService.GetDevice(deviceAddr)
	if err != nil {
		return
	}

	deviceDTO := dto.FromDomain(device)
	message := dto.NewStateUpdateMessage(deviceDTO)

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal state update: %v", err)
		return
	}

	h.broadcast <- data
}

// BroadcastMessage sends a message to all clients
func (h *Hub) BroadcastMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	h.broadcast <- data
	return nil
}

// RegisterClient registers a new client with the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}
