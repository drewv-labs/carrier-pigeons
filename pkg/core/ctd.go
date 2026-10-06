package core

import (
	"encoding/json"
	"fmt"
	"time"
)

// CTDPayload represents a single telemetry event captured at the edge.
type CTDPayload struct {
	NodeID    string         `json:"node_id"`
	NodeGroup string         `json:"node_group"` // Added topological context
	SessionID string         `json:"session_id"`
	Timestamp time.Time      `json:"timestamp"`
	Component string         `json:"component"`
	Status    string         `json:"status"`
	Metrics   map[string]any `json:"metrics,omitempty"`
}

// NewCTDPayload initializes a new telemetry event with safe defaults.
func NewCTDPayload(nodeID, sessionID, component string) *CTDPayload {
	return &CTDPayload{
		NodeID:    nodeID,
		SessionID: sessionID,
		Timestamp: time.Now().UTC(),
		Component: component,
		Status:    "initialized",
		Metrics:   make(map[string]any),
	}
}

// AddMetric attaches a key-value data point to the payload's metrics map.
// It returns the payload pointer to allow for method chaining.
func (p *CTDPayload) AddMetric(key string, value any) *CTDPayload {
	p.Metrics[key] = value
	return p
}

// UpdateStatus changes the state of the payload.
func (p *CTDPayload) UpdateStatus(status string) *CTDPayload {
	p.Status = status
	return p
}

// RoutingTopic generates the hierarchical MQTT topic for this specific payload.
func (p *CTDPayload) RoutingTopic() string {
	return fmt.Sprintf("drewv/ctd/v1/%s/%s", p.NodeID, p.Component)
}

// ToJSON serializes the CTDPayload into a JSON byte array for MQTT transmission.
func (p *CTDPayload) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
