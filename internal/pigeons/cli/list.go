package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoncoop/ledger"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered edge nodes and their last known activity",
	Run: func(cmd *cobra.Command, args []string) {
		dbURL := os.Getenv("PIGEONS_DB_URL")
		if dbURL == "" {
			dbURL = "postgres://drewv:ctd_password@localhost:5432/edge_ledger"
		}

		store, err := ledger.NewStore(context.Background(), dbURL)
		if err != nil {
			fmt.Printf("Fatal: Could not connect to ledger: %v\n", err)
			os.Exit(1)
		}
		defer store.Close()

		// Retrieve the last active timestamp for every node that has ever reported
		query := `
			SELECT node_id, MAX(event_timestamp) as last_seen
			FROM telemetric_ledger
			GROUP BY node_id
			ORDER BY node_id ASC;
		`

		rows, err := store.Query(context.Background(), query)
		if err != nil {
			fmt.Printf("Fatal: Failed to query ledger: %v\n", err)
			os.Exit(1)
		}
		defer rows.Close()

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "node\t| last seen\t| status")
		fmt.Fprintln(w, "________________________________________________________\t")

		for rows.Next() {
			var nodeID string
			var lastSeen time.Time

			if err := rows.Scan(&nodeID, &lastSeen); err == nil {
				status := "active"
				if time.Since(lastSeen) > 5*time.Minute {
					status = "offline"
				}
				fmt.Fprintf(w, "%s\t| %s\t| %s\n", nodeID, lastSeen.Format(time.RFC3339), status)
			}
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
