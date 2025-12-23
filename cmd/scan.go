package main

import (
	"context"
	"fmt"
	"time"

	"github.com/codeneuss/lampcontrol/internal/application"
	"github.com/codeneuss/lampcontrol/internal/infrastructure/bluetooth"
	"github.com/spf13/cobra"
)

var (
	scanTimeout time.Duration
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for ELK-BLEDOM devices",
	Long:  `Scan for available ELK-BLEDOM LED devices in range.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create BLE adapter
		adapter, err := bluetooth.NewAdapter()
		if err != nil {
			return fmt.Errorf("failed to initialize Bluetooth adapter: %w", err)
		}

		// Create device service
		service := application.NewDeviceService(adapter)

		fmt.Printf("Scanning for devices (timeout: %v)...\n", scanTimeout)

		// Scan for devices
		ctx := context.Background()
		devices, err := service.Scan(ctx, scanTimeout)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		if len(devices) == 0 {
			fmt.Println("No devices found")
			return nil
		}

		fmt.Printf("\nFound %d device(s):\n\n", len(devices))

		for i, dev := range devices {
			fmt.Printf("%d. %s\n", i+1, dev.Name)
			fmt.Printf("   Address: %s\n", dev.Address)
			fmt.Printf("   RSSI: %d dBm\n", dev.RSSI)
			if verbose {
				fmt.Printf("   Last Seen: %s\n", dev.LastSeen.Format(time.RFC3339))
			}
			fmt.Println()
		}

		return nil
	},
}

func init() {
	scanCmd.Flags().DurationVarP(&scanTimeout, "timeout", "t", 10*time.Second, "Scan timeout")
}
