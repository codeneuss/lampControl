package main

import (
	"context"
	"fmt"
	"log"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/bluetooth"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/storage"
	"github.com/codeneuss/lampcontrol/internal/presentation/api"
	"github.com/codeneuss/lampcontrol/internal/presentation/api/state"
	"github.com/spf13/cobra"
)

var (
	webPort int
	webHost string
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start web server for lamp control",
	Long:  `Start a web server with REST API and WebSocket support for controlling LED lamps through a browser interface.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create BLE adapter
		adapter, err := bluetooth.NewAdapter()
		if err != nil {
			return fmt.Errorf("failed to initialize Bluetooth adapter: %w", err)
		}

		// Create device service
		deviceService := application.NewDeviceService(adapter)
		defer deviceService.DisconnectAll()

		// Create effect storage
		effectStorage, err := storage.NewEffectStorage()
		if err != nil {
			return fmt.Errorf("failed to initialize effect storage: %w", err)
		}

		// Create Twitch storage
		twitchStorage, err := storage.NewTwitchStorage()
		if err != nil {
			return fmt.Errorf("failed to initialize twitch storage: %w", err)
		}

		// Create Twitch service
		twitchService := application.NewTwitchService(deviceService, twitchStorage)

		// Create server state (with Twitch service)
		serverState := state.NewServerState(deviceService, twitchService)

		// Create and start server
		server := api.NewServer(webHost, webPort, serverState, effectStorage, twitchStorage)

		// Auto-start Twitch if enabled
		twitchConfig := twitchStorage.Get()
		if twitchConfig.Enabled {
			if err := twitchService.Start(context.Background()); err != nil {
				log.Printf("Failed to auto-start Twitch integration: %v", err)
			} else {
				log.Printf("Twitch integration auto-started for channel: %s", twitchConfig.Channel)
			}
		}

		log.Printf("Starting LampControl Web Server")
		log.Printf("  Host: %s", webHost)
		log.Printf("  Port: %d", webPort)
		log.Printf("  Web UI: http://%s:%d", webHost, webPort)
		log.Printf("  API: http://%s:%d/api", webHost, webPort)
		log.Printf("  WebSocket: ws://%s:%d/ws", webHost, webPort)

		if err := server.Start(); err != nil {
			return fmt.Errorf("server error: %w", err)
		}

		return nil
	},
}

func init() {
	webCmd.Flags().IntVarP(&webPort, "port", "p", 8080, "HTTP server port")
	webCmd.Flags().StringVarP(&webHost, "host", "H", "localhost", "HTTP server host")
}
