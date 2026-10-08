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

// Automatically bind to the execution registry on binary boot
func init() {
	runner.Register(&BenchmarkRunner{})
}

func (b *BenchmarkRunner) Name() string {
	return "benchmark"
}

func (b *BenchmarkRunner) Run(ctx context.Context, nodeID string, params map[string]any, pub core.TelemetryPublisher) error {
	// Parse dynamic JSON parameters passed over MQTT
	durationSec := 60
	if d, ok := params["duration_sec"].(float64); ok {
		durationSec = int(d)
	}

	targetSystem, _ := params["target_system"].(bool)
	targetNPU, _ := params["target_npu"].(bool) // Placeholder for Hailo/Coral CLI hooks
	targetGPU, _ := params["target_gpu"].(bool)

	runID := time.Now().UnixNano()

	// 1. Emit CTD Start Bookend
	startEvent := core.NewCTDPayload(nodeID, "session-live", "runner:benchmark").
		UpdateStatus("runner-start").
		AddMetric("run_id", runID).
		AddMetric("duration_sec", durationSec).
		AddMetric("target_system", targetSystem).
		AddMetric("target_npu", targetNPU).
		AddMetric("target_gpu", targetGPU)
	pub.Publish(startEvent)

	// Enforce the requested duration without blocking the main runner dispatcher
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(durationSec)*time.Second)
	defer cancel()

	// 2. Execute Hardware Workloads
	if targetSystem {
		for range runtime.NumCPU() {
			go func() {
				for {
					select {
					case <-runCtx.Done():
						return
					default:
						// CPU Burner: Force thermal and power draw spike
						_ = sha256.Sum256([]byte("carrier-pigeons-burn"))
					}
				}
			}()
		}
	}

	// 3. Block until duration expires
	<-runCtx.Done()

	// 4. Emit CTD End Bookend
	endEvent := core.NewCTDPayload(nodeID, "session-live", "runner:benchmark").
		UpdateStatus("runner-end").
		AddMetric("run_id", runID)
	pub.Publish(endEvent)

	return nil
}
