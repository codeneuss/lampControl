package main

import (
	"context"
	"fmt"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/bluetooth"
	"github.com/spf13/cobra"
)

var powerCmd = &cobra.Command{
	Use:   "power [on|off]",
	Short: "Control power state",
	Long:  `Turn the LED lamp on or off.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if deviceAddress == "" {
			return fmt.Errorf("device address required (use --device or -d flag)")
		}

		state := args[0]
		if state != "on" && state != "off" {
			return fmt.Errorf("invalid state: %s (must be 'on' or 'off')", state)
		}

		on := state == "on"

		// Create BLE adapter
		adapter, err := bluetooth.NewAdapter()
		if err != nil {
			return fmt.Errorf("failed to initialize Bluetooth adapter: %w", err)
		}

		// Create device service
		service := application.NewDeviceService(adapter)
		defer service.DisconnectAll()

		fmt.Printf("Turning %s device %s...\n", state, deviceAddress)

		// Set power
		ctx := context.Background()
		if err := service.SetPower(ctx, deviceAddress, on); err != nil {
			return fmt.Errorf("failed to set power: %w", err)
		}

		fmt.Printf("Device turned %s successfully\n", state)

		return nil
	},
}
