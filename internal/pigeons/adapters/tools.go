package adapters

import (
	"context"
	"fmt"
	"time"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoncoop/ledger"
)

// PigeonTools defines the capabilities we expose to the local LLM
var PigeonTools = []map[string]any{
	{
		"type": "function",
		"function": map[string]any{
			"name":        "get_node_telemetry",
			"description": "Fetch the latest telemetry metrics for a specific edge node from the PostgreSQL ledger.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"node_id": map[string]any{
						"type":        "string",
						"description": "The hardware node ID (e.g., adactrl)",
					},
				},
				"required": []string{"node_id"},
			},
		},
	},
}

// ExecuteTool routes the LLM's requested tool call directly to PostgreSQL
func ExecuteTool(ctx context.Context, name string, args map[string]any, store *ledger.Store) string {
	if store == nil {
		return "Error: Database ledger is not connected."
	}

	switch name {
	case "get_node_telemetry":
		nodeID, ok := args["node_id"].(string)
		if !ok {
			return "Error: Missing required argument 'node_id'"
		}

		query := `
			SELECT component, metrics, event_timestamp
			FROM telemetric_ledger
			WHERE node_id = $1
			ORDER BY event_timestamp DESC LIMIT 5
		`
		rows, err := store.Query(ctx, query, nodeID)
		if err != nil {
			return fmt.Sprintf("Database error: %v", err)
		}
		defer rows.Close()

		var results string
		for rows.Next() {
			var component string
			var metrics []byte
			var timestamp time.Time
			rows.Scan(&component, &metrics, &timestamp)
			results += fmt.Sprintf("[%s] %s: %s\n", timestamp.Format("15:04:05"), component, string(metrics))
		}

		if results == "" {
			return "No recent telemetry found for " + nodeID
		}
		return results

	default:
		return fmt.Sprintf("Error: Unknown tool %s", name)
	}
}
