package main

import (
	"context"
	"fmt"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/bluetooth"
	"github.com/spf13/cobra"
)

var (
	warmLevel int
	coldLevel int
)

var whiteCmd = &cobra.Command{
	Use:   "white",
	Short: "Set white balance",
	Long:  `Set the white balance (warm/cold) of the LED lamp.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if deviceAddress == "" {
			return fmt.Errorf("device address required (use --device or -d flag)")
		}

		if warmLevel < 0 || warmLevel > 255 {
			return fmt.Errorf("warm level must be between 0 and 255")
		}

		if coldLevel < 0 || coldLevel > 255 {
			return fmt.Errorf("cold level must be between 0 and 255")
		}

		// Create BLE adapter
		adapter, err := bluetooth.NewAdapter()
		if err != nil {
			return fmt.Errorf("failed to initialize Bluetooth adapter: %w", err)
		}

		// Create device service
		service := application.NewDeviceService(adapter)
		defer service.DisconnectAll()

		fmt.Printf("Setting white balance to warm=%d, cold=%d on device %s...\n", warmLevel, coldLevel, deviceAddress)

		// Set white balance
		ctx := context.Background()
		if err := service.SetWhiteBalance(ctx, deviceAddress, uint8(warmLevel), uint8(coldLevel)); err != nil {
			return fmt.Errorf("failed to set white balance: %w", err)
		}

		fmt.Println("White balance set successfully")

		return nil
	},
}

func init() {
	whiteCmd.Flags().IntVarP(&warmLevel, "warm", "w", 128, "Warm white level (0-255)")
	whiteCmd.Flags().IntVarP(&coldLevel, "cold", "c", 128, "Cold white level (0-255)")
}
