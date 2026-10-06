package collectors

import (
	"bytes"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

func monitorGPUNvidia(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Include 'index' so we can tag each physical card uniquely
		cmd := exec.Command(
			"nvidia-smi",
			"--query-gpu=index,temperature.gpu,utilization.gpu,memory.used,power.draw",
			"--format=csv,noheader,nounits",
		)
		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			log.Printf("[Nvidia Collector] Sensor timeout: %v", err)
			continue
		}

		rawOutput := strings.TrimSpace(out.String())
		if rawOutput == "" {
			continue
		}

		// 1. Split into lines (one line per physical GPU)
		lines := strings.Split(rawOutput, "\n")
		for _, line := range lines {
			fields := strings.Split(line, ",")
			if len(fields) < 4 {
				continue
			}

			gpuIdx := strings.TrimSpace(fields[0])
			temp, _ := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
			util, _ := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
			memUsed, _ := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64)
			powerW, _ := strconv.ParseFloat(strings.TrimSpace(fields[4]), 64)

			// 2. Publish with a unique component tag: nvidia-gpu-0, nvidia-gpu-1, etc.
			componentID := "nvidia-gpu-" + gpuIdx

			event := core.NewCTDPayload(nodeID, "session-live", componentID).
				UpdateStatus("active").
				AddMetric("temperature_c", temp).
				AddMetric("utilization_pct", util).
				AddMetric("memory_used_mb", memUsed).
				AddMetric("power_w", powerW) // Now we are logging watts!

			if err := pub.Publish(event); err != nil {
				log.Printf("[Nvidia Collector] Publish failed for %s: %v", componentID, err)
			}
		}
	}
}
