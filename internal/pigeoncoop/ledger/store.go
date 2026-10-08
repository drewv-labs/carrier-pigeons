package ledger

import (
	"context"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps the PostgreSQL connection pool.
type Store struct {
	pool *pgxpool.Pool
}

// ExecRaw allows the migration engine to run DDL statements against the database.
func (s *Store) ExecRaw(ctx context.Context, sql string) error {
	_, err := s.pool.Exec(ctx, sql)
	return err
}

// NewStore initializes the connection to the database.
func NewStore(ctx context.Context, dbURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &Store{pool: pool}, nil
}

// Close gracefully terminates the connection pool.
func (s *Store) Close() {
	s.pool.Close()
}

// InsertEvent writes a verified CTDPayload into the JSONB ledger.
func (s *Store) InsertEvent(ctx context.Context, event *core.CTDPayload) error {
	query := `
		INSERT INTO telemetric_ledger
		(node_id, node_group, session_id, event_timestamp, component, status, metrics)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.pool.Exec(ctx, query,
		event.NodeID,
		event.NodeGroup,
		event.SessionID,
		event.EventTimestamp,
		event.Component,
		event.Status,
		event.Metrics,
	)
	return err
}

// Query executes a query against the connection pool and returns pgx.Rows.
func (s *Store) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return s.pool.Query(ctx, sql, args...)
}
