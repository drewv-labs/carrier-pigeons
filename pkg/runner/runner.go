package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// Runner defines a targeted execution workload that can be dispatched to an edge node.
type Runner interface {
	Name() string
	Run(ctx context.Context, nodeID string, params map[string]any, pub core.TelemetryPublisher) error
}

var (
	mu       sync.RWMutex
	registry = make(map[string]Runner)
)

// Register binds a new runner into the active Pigeoneer engine.
// Wrapper projects will call this in their init() functions.
func Register(r Runner) {
	mu.Lock()
	defer mu.Unlock()
	registry[r.Name()] = r
}

// Get retrieves a registered runner by its exact name.
func Get(name string) (Runner, bool) {
	mu.RLock()
	defer mu.RUnlock()
	r, ok := registry[name]
	return r, ok
}

// Envelope defines the expected MQTT JSON payload for remote dispatch.
type Envelope struct {
	Runner string         `json:"runner"`
	Params map[string]any `json:"params"`
}

func (e Envelope) Validate() error {
	if e.Runner == "" {
		return fmt.Errorf("missing runner execution name")
	}
	return nil
}
