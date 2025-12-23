package handlers

import (
	"log"
	"net/http"

	"github.com/codeneuss/lampcontrol/internal/presentation/api/state"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/websocket"
	gorillaws "github.com/gorilla/websocket"
)

var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// TODO: Restrict this in production
		return true
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	state *state.ServerState
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(state *state.ServerState) *WebSocketHandler {
	return &WebSocketHandler{
		state: state,
	}
}

// HandleWebSocket handles GET /ws
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	hub := h.state.GetWebSocketHub()
	client := websocket.NewClient(hub, conn)

	// Register client with hub
	hub.RegisterClient(client)

	// Start client read/write pumps in separate goroutines
	go client.WritePump()
	go client.ReadPump()

	log.Printf("WebSocket client connected from %s", r.RemoteAddr)
}
