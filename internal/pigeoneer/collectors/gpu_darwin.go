package collectors

import (
	"bytes"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

func monitorGPUDarwin(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Parse the specific text blocks returned by Apple's powermetrics
	powerRe := regexp.MustCompile(`GPU Power:\s+(\d+)\s+mW`)
	residencyRe := regexp.MustCompile(`GPU active residency:\s+([\d\.]+)\s*%`)

	for range ticker.C {
		// Run a 100ms sample (-i 100) exactly once (-n 1) for the GPU
		cmd := exec.Command("sudo", "powermetrics", "--samplers", "gpu_power", "-n", "1", "-i", "100")

		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			log.Printf("[Darwin GPU] Failed to read powermetrics (is pigeoneer running as root?): %v", err)
			continue
		}

		output := out.String()
		event := core.NewCTDPayload(nodeID, "session-live", "gpu-darwin").UpdateStatus("active")
		metricAdded := false

		// Convert mW to Watts
		if match := powerRe.FindStringSubmatch(output); len(match) > 1 {
			if mw, err := strconv.ParseFloat(match[1], 64); err == nil {
				event.AddMetric("gpu_power_w", mw/1000.0)
				metricAdded = true
			}
		}

		// Parse the utilization percentage
		if match := residencyRe.FindStringSubmatch(output); len(match) > 1 {
			if res, err := strconv.ParseFloat(match[1], 64); err == nil {
				event.AddMetric("gpu_utilization_pct", res)
				metricAdded = true
			}
		}

		if metricAdded {
			if err := pub.Publish(event); err != nil {
				log.Printf("[Darwin GPU] Publish failed: %v", err)
			}
		}
	}
}
