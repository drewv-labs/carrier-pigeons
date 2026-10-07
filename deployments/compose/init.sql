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
CREATE INDEX idx_ledger_timestamp ON telemetric_ledger (event_timestamp DESC);

-- Index for filtering by specific edge nodes or hardware components
CREATE INDEX idx_ledger_node_component ON telemetric_ledger (node_id, component, event_timestamp DESC);

-- GIN (Generalized Inverted Index) allows lightning-fast queries inside the JSONB metrics
-- e.g., SELECT * FROM telemetric_ledger WHERE metrics->>'temperature_c' > '45';
CREATE INDEX idx_ledger_metrics ON telemetric_ledger USING GIN (metrics);
