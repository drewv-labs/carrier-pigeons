package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node [node_id]",
	Short: "Inspect or configure a specific Pigeoneer edge node",
	Args:  cobra.ExactArgs(1), // Enforces that the user MUST provide the node_id
	Run: func(cmd *cobra.Command, args []string) {
		nodeID := args[0]
		fmt.Printf("Fetching live telemetry and config for %s...\n", nodeID)
		// Logic to query the MCP server or Postgres directly goes here
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}
