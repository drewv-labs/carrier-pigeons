package migrations

import (
	"context"
	"log"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoncoop/ledger"
)

const telemetricLedgerTableSQL = `
CREATE TABLE IF NOT EXISTS telemetric_ledger (
	id SERIAL PRIMARY KEY,
	node_id VARCHAR(255) NOT NULL,
	node_group TEXT NOT NULL,
	session_id VARCHAR(255) NOT NULL,
	event_timestamp TIMESTAMPTZ NOT NULL,
	component VARCHAR(255) NOT NULL,
	status VARCHAR(50) NOT NULL,
	metrics JSONB,
	ingested_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Index for time-series queries (e.g., "show me the last hour of telemetry")
CREATE INDEX IF NOT EXISTS idx_ledger_timestamp ON telemetric_ledger (event_timestamp DESC);

-- Create a B-Tree index on the node and component for fast time-series dashboarding
CREATE INDEX IF NOT EXISTS idx_telemetric_node_component ON telemetric_ledger (node_id, component, event_timestamp DESC);

-- Create a GIN index on the JSONB column for querying deeply nested edge metrics
CREATE INDEX IF NOT EXISTS idx_telemetric_metrics ON telemetric_ledger USING GIN (metrics);
`

const benchmarkViewSQL = `
CREATE OR REPLACE VIEW benchmark_telemetry AS
WITH benchmark_bounds AS (
    SELECT
        node_id,
        (metrics->>'run_id')::bigint AS run_id,
        MIN(event_timestamp) FILTER (WHERE status = 'benchmark-start') AS start_time,
        MAX(event_timestamp) FILTER (WHERE status = 'benchmark-end') AS end_time
    FROM telemetric_ledger
    WHERE component = 'sys-bench'
    GROUP BY node_id, metrics->>'run_id'
)
SELECT
    b.node_id,
    b.run_id,
    t.component,
    t.event_timestamp,
    t.metrics
FROM benchmark_bounds b
JOIN telemetric_ledger t
  ON t.node_id = b.node_id
 AND t.event_timestamp >= b.start_time
 AND t.event_timestamp <= COALESCE(b.end_time, NOW())
WHERE t.component != 'sys-bench';
`

// EnsureSchema checks for the existence of the required tables and builds them if missing.
func EnsureSchema(ctx context.Context, store *ledger.Store) error {
	log.Println("[Migrations] Verifying PostgreSQL schema...")
	if err := store.ExecRaw(ctx, telemetricLedgerTableSQL); err != nil {
		return err
	}
	if err := store.ExecRaw(ctx, benchmarkViewSQL); err != nil {
		return err
	}
	log.Println("[Migrations] Schema verification complete.")
	return nil
}
