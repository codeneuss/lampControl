package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/storage"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/handlers"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/middleware"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/state"
)

// Server represents the HTTP server
type Server struct {
	httpServer    *http.Server
	state         *state.ServerState
	effectStorage *storage.EffectStorage
	twitchStorage *storage.TwitchStorage
}

// NewServer creates a new HTTP server
func NewServer(host string, port int, serverState *state.ServerState, effectStorage *storage.EffectStorage, twitchStorage *storage.TwitchStorage) *Server {
	server := &Server{
		state:         serverState,
		effectStorage: effectStorage,
		twitchStorage: twitchStorage,
	}

	// Create router
	router := server.setupRouter()

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

// setupRouter configures the HTTP router with all routes and middleware
func (s *Server) setupRouter() http.Handler {
	r := chi.NewRouter()

	// Apply middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(middleware.CORS)

	// Create handlers
	deviceHandler := handlers.NewDeviceHandler(s.state)
	wsHandler := handlers.NewWebSocketHandler(s.state)
	effectHandler := handlers.NewEffectHandler(s.effectStorage)
	twitchHandler := handlers.NewTwitchHandler(s.state.GetTwitchService(), s.twitchStorage)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", deviceHandler.Health)
		r.Get("/devices", deviceHandler.ListDevices)
		r.Post("/scan", deviceHandler.ScanDevices)
		r.Post("/device/select", deviceHandler.SelectDevice)
		r.Get("/device/current", deviceHandler.GetCurrentDevice)

		// Effect routes
		r.Get("/effects", effectHandler.ListEffects)
		r.Post("/effects", effectHandler.CreateEffect)
		r.Delete("/effects/{id}", effectHandler.DeleteEffect)

		// Twitch routes
		r.Get("/twitch/config", twitchHandler.GetConfig)
		r.Put("/twitch/config", twitchHandler.UpdateConfig)
		r.Get("/twitch/status", twitchHandler.GetStatus)
		r.Get("/twitch/commands", twitchHandler.GetAvailableCommands)
		r.Get("/twitch/oauth", twitchHandler.GetOAuthURL)
	})

	// WebSocket route
	r.Get("/ws", wsHandler.HandleWebSocket)

	// Static file serving
	staticDir := "./web/static"
	if absPath, err := filepath.Abs(staticDir); err == nil {
		staticDir = absPath
	}

	// Serve static files
	fileServer := http.FileServer(http.Dir(staticDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Serve index.html for root
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	return r
}

// Start starts the HTTP server and WebSocket hub
func (s *Server) Start() error {
	// Start WebSocket hub in background
	go s.state.GetWebSocketHub().Run()

	// Start HTTP server
	log.Printf("Starting web server on %s", s.httpServer.Addr)
	log.Printf("Web UI available at http://%s", s.httpServer.Addr)

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Server stopped")
	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
