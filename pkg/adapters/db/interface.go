// File: pkg/adapters/db/interface.go
package db

import (
	"context"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// NodeSummary represents the materialized state of an edge node for `pigeons list` and `ps`.
type NodeSummary struct {
	NodeID        string                 `json:"node_id"`
	LastSeen      time.Time              `json:"last_seen"`
	Status        string                 `json:"status"`
	LatestMetrics map[string]interface{} `json:"latest_metrics"`
	IsHealthy     bool                   `json:"is_healthy"`
}

// DAG represents a reconstructed Artifact & Tracer graph for PigeonMath / lineage authentication.
type DAG struct {
	RootID    string           `json:"root_id"`
	Artifacts []core.Artifact  `json:"artifacts"`
	Edges     []core.TraceEdge `json:"edges"`
}

// UnloadFilter scopes which rows are pulled when exporting to Snappy Parquet.
type UnloadFilter struct {
	NodeIDPattern string
	Pillar        string
	Since         time.Time
	Until         time.Time
	Limit         int
}

// BaseDB is the storage adapter contract for CarrierPigeons and any white-labeled AppOS.
type BaseDB interface {
	EnsureSchema(ctx context.Context) error
	InsertEvent(ctx context.Context, payload *core.CTDPayload) error
	UpsertArtifact(ctx context.Context, a core.Artifact) error
	InsertTrace(ctx context.Context, t core.TraceEdge) error

	// High-level semantic queries for CLI, TUI, and PigeonMath
	GetClusterStatus(ctx context.Context, staleThreshold time.Duration) ([]NodeSummary, error)
	GetLineageDAG(ctx context.Context, artifactID string, maxDepth int) (*DAG, error)
	FetchForUnload(ctx context.Context, filter UnloadFilter) ([]core.CTDPayload, error)

	// PigeonMath semantic view builder & raw SQL escape hatch for LLM tools
	RegisterView(ctx context.Context, viewName string, selectQuery string) error
	ExecRaw(ctx context.Context, sql string, args ...any) error
	QueryMaps(ctx context.Context, sql string, args ...any) ([]map[string]any, error)

	Close()
}
