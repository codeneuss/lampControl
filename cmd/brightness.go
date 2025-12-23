package main

import (
	"context"
	"fmt"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/bluetooth"
	"github.com/spf13/cobra"
)

var (
	brightnessLevel int
)

var brightnessCmd = &cobra.Command{
	Use:   "brightness",
	Short: "Set brightness level",
	Long:  `Set the brightness level of the LED lamp (0-255).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if deviceAddress == "" {
			return fmt.Errorf("device address required (use --device or -d flag)")
		}

		if brightnessLevel < 0 || brightnessLevel > 255 {
			return fmt.Errorf("brightness must be between 0 and 255")
		}

		// Create BLE adapter
		adapter, err := bluetooth.NewAdapter()
		if err != nil {
			return fmt.Errorf("failed to initialize Bluetooth adapter: %w", err)
		}

		// Create device service
		service := application.NewDeviceService(adapter)
		defer service.DisconnectAll()

		fmt.Printf("Setting brightness to %d on device %s...\n", brightnessLevel, deviceAddress)

		// Set brightness
		ctx := context.Background()
		if err := service.SetBrightness(ctx, deviceAddress, uint8(brightnessLevel)); err != nil {
			return fmt.Errorf("failed to set brightness: %w", err)
		}

		fmt.Println("Brightness set successfully")

		return nil
	},
}

func init() {
	brightnessCmd.Flags().IntVarP(&brightnessLevel, "level", "l", 255, "Brightness level (0-255)")
}
