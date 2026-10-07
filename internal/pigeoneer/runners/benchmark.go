package runners

import (
	"context"
	"crypto/sha256"
	"runtime"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
	"github.com/drewv-labs/carrier-pigeons/pkg/runner"
)

type BenchmarkRunner struct{}

// Auto-register upon binary execution
func init() {
	runner.Register(&BenchmarkRunner{})
}

func (b *BenchmarkRunner) Name() string {
	return "benchmark"
}

func (b *BenchmarkRunner) Run(ctx context.Context, nodeID string, params map[string]any, pub core.TelemetryPublisher) error {
	// Parse dynamic JSON parameters passed from the CLI
	durationSec := 60
	if d, ok := params["duration_sec"].(float64); ok {
		durationSec = int(d)
	}

	targetSystem, _ := params["target_system"].(bool)

	runID := time.Now().UnixNano()

	// 1. Emit CTD Start Bookend
	startEvent := core.NewCTDPayload(nodeID, "session-live", "runner:benchmark").
		UpdateStatus("runner-start").
		AddMetric("run_id", runID).
		AddMetric("duration_sec", durationSec).
		AddMetric("target_system", targetSystem)
	pub.Publish(startEvent)

	// Enforce the duration
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(durationSec)*time.Second)
	defer cancel()

	// 2. Execute Workload
	if targetSystem {
		for i := 0; i < runtime.NumCPU(); i++ {
			go func() {
				for {
					select {
					case <-runCtx.Done():
						return
					default:
						// CPU Burner
						_ = sha256.Sum256([]byte("carrier-pigeons-burn"))
					}
				}
			}()
		}
	}

	// Wait for the duration to elapse
	<-runCtx.Done()

	// 3. Emit CTD End Bookend
	endEvent := core.NewCTDPayload(nodeID, "session-live", "runner:benchmark").
		UpdateStatus("runner-end").
		AddMetric("run_id", runID)
	pub.Publish(endEvent)

	return nil
}
