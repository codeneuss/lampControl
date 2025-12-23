package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/bluetooth"
	"github.com/spf13/cobra"
)

var (
	rgbColor string
)

var colorCmd = &cobra.Command{
	Use:   "color",
	Short: "Set RGB color",
	Long:  `Set the RGB color of the LED lamp.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if deviceAddress == "" {
			return fmt.Errorf("device address required (use --device or -d flag)")
		}

		if rgbColor == "" {
			return fmt.Errorf("RGB color required (use --rgb flag)")
		}

		// Parse RGB values
		parts := strings.Split(rgbColor, ",")
		if len(parts) != 3 {
			return fmt.Errorf("invalid RGB format (expected: R,G,B where each value is 0-255)")
		}

		r, err := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 8)
		if err != nil {
			return fmt.Errorf("invalid red value: %w", err)
		}

		g, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 8)
		if err != nil {
			return fmt.Errorf("invalid green value: %w", err)
		}

		b, err := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 8)
		if err != nil {
			return fmt.Errorf("invalid blue value: %w", err)
		}

		// Create BLE adapter
		adapter, err := bluetooth.NewAdapter()
		if err != nil {
			return fmt.Errorf("failed to initialize Bluetooth adapter: %w", err)
		}

		// Create device service
		service := application.NewDeviceService(adapter)
		defer service.DisconnectAll()

		fmt.Printf("Setting color to RGB(%d,%d,%d) on device %s...\n", r, g, b, deviceAddress)

		// Set color
		ctx := context.Background()
		if err := service.SetColor(ctx, deviceAddress, uint8(r), uint8(g), uint8(b)); err != nil {
			return fmt.Errorf("failed to set color: %w", err)
		}

		fmt.Println("Color set successfully")

		return nil
	},
}

func init() {
	colorCmd.Flags().StringVarP(&rgbColor, "rgb", "r", "", "RGB color (format: R,G,B where each is 0-255)")
}
