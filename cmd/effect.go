package main

import (
	"context"
	"fmt"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/bluetooth"
	"github.com/spf13/cobra"
)

var (
	effectIndex int
	effectSpeed int
)

var effectCmd = &cobra.Command{
	Use:   "effect",
	Short: "Set built-in effect/scene",
	Long:  `Set a built-in effect or scene on the LED lamp.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if deviceAddress == "" {
			return fmt.Errorf("device address required (use --device or -d flag)")
		}

		if effectIndex < 0 || effectIndex > 255 {
			return fmt.Errorf("effect index must be between 0 and 255")
		}

		if effectSpeed < 0 || effectSpeed > 255 {
			return fmt.Errorf("effect speed must be between 0 and 255")
		}

		// Create BLE adapter
		adapter, err := bluetooth.NewAdapter()
		if err != nil {
			return fmt.Errorf("failed to initialize Bluetooth adapter: %w", err)
		}

		// Create device service
		service := application.NewDeviceService(adapter)
		defer service.DisconnectAll()

		fmt.Printf("Setting effect %d with speed %d on device %s...\n", effectIndex, effectSpeed, deviceAddress)

		// Set effect
		ctx := context.Background()
		if err := service.SetEffect(ctx, deviceAddress, uint8(effectIndex), uint8(effectSpeed)); err != nil {
			return fmt.Errorf("failed to set effect: %w", err)
		}

		fmt.Println("Effect set successfully")

		return nil
	},
}

func init() {
	effectCmd.Flags().IntVarP(&effectIndex, "index", "i", 1, "Effect index (0-255)")
	effectCmd.Flags().IntVarP(&effectSpeed, "speed", "s", 50, "Effect speed (0-255, higher is faster)")
}
