// File: pkg/core/ctd.go
package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// Pillar defines the core SDK primitive type emitting the event.
type Pillar string

const (
	PillarObserver Pillar = "observer"
	PillarRunner   Pillar = "runner"
	PillarTracer   Pillar = "tracer"
)

// Artifact represents any tangible Noun (hardware node, dataset, FITS frame, script) in the DAG.
type Artifact struct {
	ID        string            `json:"id"`
	Kind      string            `json:"kind"` // e.g., "node", "dataset", "process", "model"
	NodeID    string            `json:"node_id,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// TraceEdge represents a directed Verb connecting two Artifacts across time.
type TraceEdge struct {
	SourceID  string            `json:"source_id"`
	Verb      string            `json:"verb"` // e.g., "calibrated_with", "stacked_into", "connected_to"
	TargetID  string            `json:"target_id"`
	NodeID    string            `json:"node_id,omitempty"`
	RunID     string            `json:"run_id,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// CTDPayload is the immutable wire envelope transmitted over MQTT and stored in the Ledger.
type CTDPayload struct {
	EventID        string                 `json:"event_id,omitempty"`
	NodeID         string                 `json:"node_id"`
	Pillar         Pillar                 `json:"pillar"`
	Component      string                 `json:"component"`
	Status         string                 `json:"status"`
	EventTimestamp time.Time              `json:"event_timestamp"`
	Metrics        map[string]interface{} `json:"metrics,omitempty"`
	Params         map[string]string      `json:"params,omitempty"`
	Artifact       *Artifact              `json:"artifact,omitempty"`
	Trace          *TraceEdge             `json:"trace,omitempty"`
}

// NewCTDPayload initializes a standard Observer/Runner payload (backward-compatible).
func NewCTDPayload(nodeID, component, status string) *CTDPayload {
	return &CTDPayload{
		NodeID:         nodeID,
		Pillar:         PillarObserver,
		Component:      component,
		Status:         status,
		EventTimestamp: time.Now().UTC(),
		Metrics:        make(map[string]interface{}),
		Params:         make(map[string]string),
	}
}

func (c *CTDPayload) WithPillar(p Pillar) *CTDPayload {
	c.Pillar = p
	return c
}

func (c *CTDPayload) WithArtifact(a Artifact) *CTDPayload {
	if a.CreatedAt.IsZero() {
		a.CreatedAt = c.EventTimestamp
	}
	c.Artifact = &a
	return c
}

func (c *CTDPayload) WithTrace(sourceID, verb, targetID string, meta map[string]string) *CTDPayload {
	c.Pillar = PillarTracer
	c.Trace = &TraceEdge{
		SourceID:  sourceID,
		Verb:      verb,
		TargetID:  targetID,
		NodeID:    c.NodeID,
		Meta:      meta,
		Timestamp: c.EventTimestamp,
	}
	return c
}

func (c *CTDPayload) AddMetric(key string, val interface{}) *CTDPayload {
	if c.Metrics == nil {
		c.Metrics = make(map[string]interface{})
	}
	c.Metrics[key] = val
	return c
}

func (c *CTDPayload) UpdateStatus(status string) *CTDPayload {
	c.Status = status
	c.EventTimestamp = time.Now().UTC()
	return c
}

func (c *CTDPayload) RoutingTopic(prefix string) string {
	if prefix == "" {
		prefix = "pigeons/ctd"
	}
	return fmt.Sprintf("%s/%s/%s/%s", prefix, c.Pillar, c.NodeID, c.Component)
}

func (c *CTDPayload) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}
