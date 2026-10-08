package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// Runner represents any targeted remote execution task.
type Runner interface {
	Name() string
	Run(ctx context.Context, nodeID string, params map[string]any, pub core.TelemetryPublisher) error
}

var (
	mu       sync.RWMutex
	registry = make(map[string]Runner)
)

// Register is called in init() functions to bind a runner to the edge daemon.
func Register(r Runner) {
	mu.Lock()
	defer mu.Unlock()
	registry[r.Name()] = r
}

// Get safely retrieves a registered runner by name.
func Get(name string) (Runner, bool) {
	mu.RLock()
	defer mu.RUnlock()
	r, ok := registry[name]
	return r, ok
}

// RegistryNames returns a list of all currently registered runners.
func RegistryNames() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// Envelope defines the wire protocol for MQTT control dispatches.
type Envelope struct {
	Runner string         `json:"runner"`
	Params map[string]any `json:"params"`
}

// Validate ensures the payload has the minimum required routing data.
func (e Envelope) Validate() error {
	if e.Runner == "" {
		return fmt.Errorf("missing runner execution name")
	}
	return nil
}
