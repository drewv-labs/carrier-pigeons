package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoncoop/ledger"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 1. Connect to the existing PostgreSQL Ledger (Read-Only)
	dbURL := "postgres://drewv:ctd_password@localhost:5432/edge_ledger"
	store, err := ledger.NewStore(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("MCP failed to connect to ledger: %v", err)
	}
	defer store.Close()

	// 2. Initialize the MCP Server
	s := server.NewMCPServer(
		"Ada Lovespace Telemetry",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	// 3. Tool: Get Node Telemetry
	telemetryTool := mcp.NewTool("get_node_telemetry",
		mcp.WithDescription("Fetch the latest telemetry metrics for a specific edge node."),
		mcp.WithString("node_id", mcp.Required(), mcp.Description("The hardware node ID (e.g., ada-node-1)")),
	)

	s.AddTool(telemetryTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nodeID, _ := req.RequireString("node_id")

		// Query Postgres directly for the latest JSONB metrics
		query := `
			SELECT component, metrics, event_timestamp
			FROM telemetric_ledger
			WHERE node_id = $1
			ORDER BY event_timestamp DESC LIMIT 5
		`
		rows, err := store.Query(ctx, query, nodeID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Database error: %v", err)), nil
		}
		defer rows.Close()

		var results string
		for rows.Next() {
			var component string
			var metrics []byte
			var timestamp time.Time
			rows.Scan(&component, &metrics, &timestamp)
			results += fmt.Sprintf("[%s] %s: %s\n", timestamp.Format(time.RFC3339), component, string(metrics))
		}

		if results == "" {
			return mcp.NewToolResultText("No telemetry found for " + nodeID), nil
		}
		return mcp.NewToolResultText(results), nil
	})

	// 4. Tool: Check Offline Nodes
	offlineTool := mcp.NewTool("check_offline_nodes",
		mcp.WithDescription("Identify any edge nodes that have stopped reporting data in the last 60 seconds."),
	)

	s.AddTool(offlineTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := `
			SELECT node_id, MAX(event_timestamp) as last_seen
			FROM telemetric_ledger
			GROUP BY node_id
			HAVING MAX(event_timestamp) < NOW() - INTERVAL '60 seconds'
		`
		rows, err := store.Query(ctx, query)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Database error: %v", err)), nil
		}
		defer rows.Close()

		var results string
		for rows.Next() {
			var nodeID string
			var lastSeen time.Time
			rows.Scan(&nodeID, &lastSeen)
			results += fmt.Sprintf("⚠️ %s (Last seen: %s)\n", nodeID, time.Since(lastSeen).Round(time.Second))
		}

		if results == "" {
			return mcp.NewToolResultText("All nodes are currently online and reporting."), nil
		}
		return mcp.NewToolResultText(results), nil
	})

	// 5. Start the server over Standard I/O for the editor to consume
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
