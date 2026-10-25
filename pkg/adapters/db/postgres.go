// File: pkg/adapters/db/postgres.go
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore initializes a connection pool to the PostgreSQL CTD Ledger.
func NewPostgresStore(ctx context.Context, connStr string) (*PostgresStore, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}
	cfg.MaxConns = 16
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	return &PostgresStore{pool: pool}, nil
}

// Pool exposes the underlying pgxpool.Pool for any legacy internal callers during migration.
func (s *PostgresStore) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *PostgresStore) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// EnsureSchema migrates the 3 core CTD tables (Ledger, Artifacts, Traces) and GIN indexes.
func (s *PostgresStore) EnsureSchema(ctx context.Context) error {
	ddl := `
	CREATE TABLE IF NOT EXISTS telemetric_ledger (
		id BIGSERIAL PRIMARY KEY,
		node_id TEXT NOT NULL,
		pillar TEXT NOT NULL DEFAULT 'observer',
		component TEXT NOT NULL,
		status TEXT NOT NULL,
		event_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		metrics JSONB DEFAULT '{}'::jsonb,
		params JSONB DEFAULT '{}'::jsonb
	);

	CREATE INDEX IF NOT EXISTS idx_ledger_node_ts ON telemetric_ledger (node_id, event_timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_ledger_pillar ON telemetric_ledger (pillar);
	CREATE INDEX IF NOT EXISTS idx_ledger_metrics_gin ON telemetric_ledger USING GIN (metrics);

	CREATE TABLE IF NOT EXISTS ctd_artifacts (
		id TEXT PRIMARY KEY,
		kind TEXT NOT NULL,
		node_id TEXT,
		meta JSONB DEFAULT '{}'::jsonb,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_artifacts_kind ON ctd_artifacts (kind);

	CREATE TABLE IF NOT EXISTS ctd_traces (
		id BIGSERIAL PRIMARY KEY,
		source_id TEXT NOT NULL,
		verb TEXT NOT NULL,
		target_id TEXT NOT NULL,
		node_id TEXT,
		run_id TEXT,
		meta JSONB DEFAULT '{}'::jsonb,
		event_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_traces_source ON ctd_traces (source_id);
	CREATE INDEX IF NOT EXISTS idx_traces_target ON ctd_traces (target_id);
	`
	_, err := s.pool.Exec(ctx, ddl)
	return err
}

// InsertEvent writes the CTD envelope and automatically shreds embedded Artifacts/Traces.
func (s *PostgresStore) InsertEvent(ctx context.Context, p *core.CTDPayload) error {
	if p.Pillar == "" {
		p.Pillar = core.PillarObserver
	}
	if p.EventTimestamp.IsZero() {
		p.EventTimestamp = time.Now().UTC()
	}

	metricsJSON, err := json.Marshal(p.Metrics)
	if err != nil {
		return fmt.Errorf("marshal metrics: %w", err)
	}
	paramsJSON, err := json.Marshal(p.Params)
	if err != nil {
		return fmt.Errorf("marshal params: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO telemetric_ledger (node_id, pillar, component, status, event_timestamp, metrics, params)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, p.NodeID, string(p.Pillar), p.Component, p.Status, p.EventTimestamp, metricsJSON, paramsJSON)
	if err != nil {
		return fmt.Errorf("insert telemetric_ledger: %w", err)
	}

	if p.Artifact != nil {
		metaJSON, _ := json.Marshal(p.Artifact.Meta)
		_, err = tx.Exec(ctx, `
			INSERT INTO ctd_artifacts (id, kind, node_id, meta, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
			ON CONFLICT (id) DO UPDATE SET
				kind = EXCLUDED.kind,
				node_id = COALESCE(EXCLUDED.node_id, ctd_artifacts.node_id),
				meta = ctd_artifacts.meta || EXCLUDED.meta,
				updated_at = NOW()
		`, p.Artifact.ID, p.Artifact.Kind, p.Artifact.NodeID, metaJSON, p.Artifact.CreatedAt)
		if err != nil {
			return fmt.Errorf("upsert artifact: %w", err)
		}
	}

	if p.Trace != nil {
		traceMetaJSON, _ := json.Marshal(p.Trace.Meta)
		_, err = tx.Exec(ctx, `
			INSERT INTO ctd_traces (source_id, verb, target_id, node_id, run_id, meta, event_timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, p.Trace.SourceID, p.Trace.Verb, p.Trace.TargetID, p.Trace.NodeID, p.Trace.RunID, traceMetaJSON, p.Trace.Timestamp)
		if err != nil {
			return fmt.Errorf("insert trace: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *PostgresStore) UpsertArtifact(ctx context.Context, a core.Artifact) error {
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	metaJSON, _ := json.Marshal(a.Meta)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO ctd_artifacts (id, kind, node_id, meta, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO UPDATE SET
			meta = ctd_artifacts.meta || EXCLUDED.meta,
			updated_at = NOW()
	`, a.ID, a.Kind, a.NodeID, metaJSON, a.CreatedAt)
	return err
}

func (s *PostgresStore) InsertTrace(ctx context.Context, t core.TraceEdge) error {
	if t.Timestamp.IsZero() {
		t.Timestamp = time.Now().UTC()
	}
	metaJSON, _ := json.Marshal(t.Meta)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO ctd_traces (source_id, verb, target_id, node_id, run_id, meta, event_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, t.SourceID, t.Verb, t.TargetID, t.NodeID, t.RunID, metaJSON, t.Timestamp)
	return err
}

// GetClusterStatus returns the latest state per node using Postgres DISTINCT ON.
func (s *PostgresStore) GetClusterStatus(ctx context.Context, staleThreshold time.Duration) ([]NodeSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ON (node_id)
			node_id, event_timestamp, status, metrics
		FROM telemetric_ledger
		ORDER BY node_id, event_timestamp DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []NodeSummary
	cutoff := time.Now().UTC().Add(-staleThreshold)

	for rows.Next() {
		var ns NodeSummary
		var metricsBytes []byte
		if err := rows.Scan(&ns.NodeID, &ns.LastSeen, &ns.Status, &metricsBytes); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(metricsBytes, &ns.LatestMetrics)
		ns.IsHealthy = ns.LastSeen.After(cutoff) && ns.Status != "error" && ns.Status != "critical"
		summaries = append(summaries, ns)
	}
	return summaries, rows.Err()
}

// GetLineageDAG walks upstream and downstream edges from any Artifact ID using a recursive CTE.
func (s *PostgresStore) GetLineageDAG(ctx context.Context, artifactID string, maxDepth int) (*DAG, error) {
	if maxDepth <= 0 {
		maxDepth = 10
	}

	rows, err := s.pool.Query(ctx, `
		WITH RECURSIVE lineage AS (
			SELECT source_id, verb, target_id, node_id, COALESCE(run_id, ''), meta, event_timestamp, 1 AS depth
			FROM ctd_traces
			WHERE source_id = $1 OR target_id = $1
			UNION
			SELECT t.source_id, t.verb, t.target_id, t.node_id, COALESCE(t.run_id, ''), t.meta, t.event_timestamp, l.depth + 1
			FROM ctd_traces t
			INNER JOIN lineage l ON (t.source_id = l.target_id OR t.target_id = l.source_id)
			WHERE l.depth < $2
		)
		SELECT DISTINCT source_id, verb, target_id, node_id, run_id, meta, event_timestamp
		FROM lineage
		ORDER BY event_timestamp ASC
	`, artifactID, maxDepth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dag := &DAG{RootID: artifactID}
	artifactSet := map[string]bool{artifactID: true}

	for rows.Next() {
		var edge core.TraceEdge
		var metaBytes []byte
		if err := rows.Scan(&edge.SourceID, &edge.Verb, &edge.TargetID, &edge.NodeID, &edge.RunID, &metaBytes, &edge.Timestamp); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(metaBytes, &edge.Meta)
		dag.Edges = append(dag.Edges, edge)
		artifactSet[edge.SourceID] = true
		artifactSet[edge.TargetID] = true
	}

	// Hydrate all discovered Artifacts
	ids := make([]string, 0, len(artifactSet))
	for id := range artifactSet {
		ids = append(ids, id)
	}

	aRows, err := s.pool.Query(ctx, `
		SELECT id, kind, COALESCE(node_id, ''), meta, created_at
		FROM ctd_artifacts
		WHERE id = ANY($1)
	`, ids)
	if err != nil {
		return dag, nil // Return edges even if some artifacts weren't explicitly registered
	}
	defer aRows.Close()

	for aRows.Next() {
		var a core.Artifact
		var metaBytes []byte
		if err := aRows.Scan(&a.ID, &a.Kind, &a.NodeID, &metaBytes, &a.CreatedAt); err != nil {
			continue
		}
		_ = json.Unmarshal(metaBytes, &a.Meta)
		dag.Artifacts = append(dag.Artifacts, a)
	}

	return dag, nil
}

// FetchForUnload pulls filtered ledger events ready to be shredded into Snappy Parquet.
func (s *PostgresStore) FetchForUnload(ctx context.Context, f UnloadFilter) ([]core.CTDPayload, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 100000
	}
	nodePattern := f.NodeIDPattern
	if nodePattern == "" {
		nodePattern = "%"
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, node_id, pillar, component, status, event_timestamp, metrics, params
		FROM telemetric_ledger
		WHERE node_id LIKE $1
		  AND ($2 = '' OR pillar = $2)
		  AND ($3::timestamptz IS NULL OR event_timestamp >= $3)
		ORDER BY event_timestamp ASC
		LIMIT $4
	`, nodePattern, f.Pillar, f.Since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []core.CTDPayload
	for rows.Next() {
		var p core.CTDPayload
		var pillarStr string
		var metricsBytes, paramsBytes []byte
		if err := rows.Scan(&p.EventID, &p.NodeID, &pillarStr, &p.Component, &p.Status, &p.EventTimestamp, &metricsBytes, &paramsBytes); err != nil {
			return nil, err
		}
		p.Pillar = core.Pillar(pillarStr)
		_ = json.Unmarshal(metricsBytes, &p.Metrics)
		_ = json.Unmarshal(paramsBytes, &p.Params)
		results = append(results, p)
	}
	return results, rows.Err()
}

// RegisterView powers PigeonMath's QueryModelAdapter ("CREATE OR REPLACE VIEW").
func (s *PostgresStore) RegisterView(ctx context.Context, viewName string, selectQuery string) error {
	if !validIdentifier.MatchString(viewName) {
		return fmt.Errorf("invalid SQL view identifier: %s", viewName)
	}
	sql := fmt.Sprintf("CREATE OR REPLACE VIEW %s AS %s", viewName, selectQuery)
	_, err := s.pool.Exec(ctx, sql)
	return err
}

func (s *PostgresStore) ExecRaw(ctx context.Context, sql string, args ...any) error {
	_, err := s.pool.Exec(ctx, sql, args...)
	return err
}

// QueryMaps executes arbitrary read queries and returns a driver-agnostic slice of maps.
func (s *PostgresStore) QueryMaps(ctx context.Context, sql string, args ...any) ([]map[string]any, error) {
	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToMap)
}
