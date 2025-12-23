package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	deviceAddress string
	verbose       bool
)

var rootCmd = &cobra.Command{
	Use:   "lamp",
	Short: "Control duoCo StripX LED lamps via Bluetooth",
	Long: `A CLI tool and REST API for controlling duoCo StripX LED lamps
using the ELK-BLEDOM protocol over Bluetooth Low Energy.`,
	Version: "1.0.0",
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVarP(&deviceAddress, "device", "d", "", "Device MAC address")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Add subcommands
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(powerCmd)
	rootCmd.AddCommand(colorCmd)
	rootCmd.AddCommand(brightnessCmd)
	rootCmd.AddCommand(whiteCmd)
	rootCmd.AddCommand(effectCmd)
	rootCmd.AddCommand(webCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
