package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var coopCmd = &cobra.Command{
	Use:   "coop",
	Short: "Manage the PigeonCoop injection engine and PostgreSQL ledger",
}

var coopStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "View live ingestion metrics and registry health",
	Run: func(cmd *cobra.Command, args []string) {
		// Placeholder: We will wire Bubble Tea here to draw the live terminal dashboard
		fmt.Println("Spawning PigeonCoop Status Dashboard...")
	},
}

var coopRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Gracefully restart the PigeonCoop daemon",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Sending restart signal to PigeonCoop engine...")
		// Logic to restart the systemd service or Docker container goes here
	},
}

func init() {
	// Wire the hierarchy together
	rootCmd.AddCommand(coopCmd)
	coopCmd.AddCommand(coopStatusCmd)
	coopCmd.AddCommand(coopRestartCmd)
}
