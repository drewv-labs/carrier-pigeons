package migrations

import (
	"context"
	"log"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoncoop/ledger"
)

// EnsureSchema checks for the existence of the required tables and builds them if missing.
func EnsureSchema(ctx context.Context, store *ledger.Store) error {
	log.Println("[Migrations] Verifying PostgreSQL schema...")

	query := `
		CREATE TABLE IF NOT EXISTS telemetric_ledger (
			id SERIAL PRIMARY KEY,
			node_id VARCHAR(255) NOT NULL,
			session_id VARCHAR(255) NOT NULL,
			event_timestamp TIMESTAMPTZ NOT NULL,
			component VARCHAR(255) NOT NULL,
			status VARCHAR(50) NOT NULL,
			metrics JSONB,
			ingested_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);

		-- Create a GIN index on the JSONB column for querying deeply nested edge metrics
		CREATE INDEX IF NOT EXISTS idx_telemetric_metrics ON telemetric_ledger USING GIN (metrics);

		-- Create a B-Tree index on the node and component for fast time-series dashboarding
		CREATE INDEX IF NOT EXISTS idx_telemetric_node_comp ON telemetric_ledger (node_id, component, event_timestamp DESC);
	`

	if err := store.ExecRaw(ctx, query); err != nil {
		return err
	}

	log.Println("[Migrations] Schema verification complete.")
	return nil
}
