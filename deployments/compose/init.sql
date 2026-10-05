CREATE TABLE IF NOT EXISTS telemetric_ledger (
    id BIGSERIAL PRIMARY KEY,
    node_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    event_timestamp TIMESTAMPTZ NOT NULL,
    component TEXT NOT NULL,
    status TEXT NOT NULL,
    metrics JSONB NOT NULL DEFAULT '{}'::jsonb,
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for time-series queries (e.g., "show me the last hour of telemetry")
CREATE INDEX idx_ledger_timestamp ON telemetric_ledger (event_timestamp DESC);

-- Index for filtering by specific edge nodes or hardware components
CREATE INDEX idx_ledger_node_component ON telemetric_ledger (node_id, component);

-- GIN (Generalized Inverted Index) allows lightning-fast queries inside the JSONB metrics
-- e.g., SELECT * FROM telemetric_ledger WHERE metrics->>'temperature_c' > '45';
CREATE INDEX idx_ledger_metrics ON telemetric_ledger USING GIN (metrics);
