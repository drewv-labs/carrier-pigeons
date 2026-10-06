package collectors

import (
	"encoding/json"
	"log"
	"os/exec"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

func monitorIntelGPU(pub core.TelemetryPublisher, nodeID string) {
	// 1. Launch the process to stream JSON (-J) every 5000ms (-s)
	cmd := exec.Command("intel_gpu_top", "-J", "-s", "5000")

	// 2. Attach a pipe directly to the live standard output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[Intel Collector] Failed to open stdout pipe: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[Intel Collector] Failed to start intel_gpu_top: %v", err)
		return
	}

	// Ensure the child process dies if this goroutine ever crashes
	defer cmd.Process.Kill()

	decoder := json.NewDecoder(stdout)

	// 3. The stream starts with an opening JSON array bracket '['.
	// We must consume this first token so the decoder can read the objects inside.
	if _, err := decoder.Token(); err != nil {
		log.Printf("[Intel Collector] Failed to read opening JSON bracket: %v", err)
		return
	}

	log.Println("[Intel Collector] Successfully attached to intel_gpu_top stream.")

	// 4. Continuously block and wait for the next JSON object in the stream
	for decoder.More() {
		var data map[string]any
		if err := decoder.Decode(&data); err != nil {
			log.Printf("[Intel Collector] JSON stream decode error: %v", err)
			break
		}

		event := core.NewCTDPayload(nodeID, "session-live", "intel-gpu").
			UpdateStatus("active")

		// Extract Power metrics if the specific Intel architecture supports it
		if powerMap, ok := data["power"].(map[string]any); ok {
			if gpuPower, exists := powerMap["GPU"].(float64); exists {
				event.AddMetric("power_w", gpuPower)
			}
		}

		// Extract the Render/3D engine utilization
		if engines, ok := data["engines"].(map[string]any); ok {
			if render, exists := engines["Render/3D/0"].(map[string]any); exists {
				if busy, isFloat := render["busy"].(float64); isFloat {
					event.AddMetric("utilization_pct", busy)
				}
			}
		}

		if err := pub.Publish(event); err != nil {
			log.Printf("[Intel Collector] Publish failed: %v", err)
		}
	}

	log.Println("[Intel Collector] intel_gpu_top stream unexpectedly ended.")
}
