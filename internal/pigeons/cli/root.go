package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pigeons",
	Short: "PigeonCommand: The Edge Observability TUI",
	Long:  `CarrierPigeons Master Control. Manage distributed telemetry, benchmark edge nodes, and interface with the PigeonCoop ledger.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Placeholder: This is where we will boot the full Bubble Tea TUI
		fmt.Println("Welcome to PigeonCommand. Type 'pigeons --help' for commands.")
	},
}

// Execute is called by main.go to bootstrap the CLI.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
