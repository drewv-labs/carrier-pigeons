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
	{
		"type": "function",
		"function": map[string]any{
			"name":        "analyze_latest_benchmark",
			"description": "Fetch the hardware telemetry strictly bounded by the most recent runner execution for a given node.",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"node_id": map[string]any{
						"type":        "string",
						"description": "The edge node ID (e.g., ada-node-1, caroline_hailo)",
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

	case "analyze_latest_benchmark":
		nodeID, ok := args["node_id"].(string)
		if !ok {
			return "Error: Missing node_id"
		}

		// Use a CTE to grab the start bookend, calculate the end time dynamically from the JSONB,
		// and join it against the main ledger to extract only the sandwiched telemetry.
		query := `
				WITH latest_run AS (
					SELECT
						(metrics->>'run_id')::bigint AS run_id,
						event_timestamp AS start_time,
						event_timestamp + ((metrics->>'duration_sec')::numeric || ' seconds')::interval AS end_time
					FROM telemetric_ledger
					WHERE node_id = $1 AND status = 'runner-start'
					ORDER BY event_timestamp DESC
					LIMIT 1
				)
				SELECT t.component, t.metrics, t.event_timestamp
				FROM telemetric_ledger t
				JOIN latest_run lr ON true
				WHERE t.node_id = $1
				  AND t.event_timestamp >= lr.start_time
				  AND t.event_timestamp <= lr.end_time
				  AND t.component != 'runner:benchmark'
				ORDER BY t.event_timestamp ASC;
			`

		rows, err := store.Query(ctx, query, nodeID)
		if err != nil {
			return fmt.Sprintf("Ledger query failed: %v", err)
		}
		defer rows.Close()

		var results string
		recordCount := 0

		// Unpack the JSONB payloads for the LLM
		for rows.Next() {
			var component string
			var metrics []byte
			var timestamp time.Time

			if err := rows.Scan(&component, &metrics, &timestamp); err == nil {
				results += fmt.Sprintf("[%s] %s: %s\n", timestamp.Format("15:04:05.000"), component, string(metrics))
				recordCount++
			}
		}

		if recordCount == 0 {
			return "No telemetry was recorded during the last benchmark run for " + nodeID
		}
		return fmt.Sprintf("Found %d hardware telemetry events during the benchmark:\n%s", recordCount, results)

	default:
		return fmt.Sprintf("Error: Unknown tool %s", name)
	}
}
