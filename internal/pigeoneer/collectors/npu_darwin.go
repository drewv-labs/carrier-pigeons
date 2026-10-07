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

func monitorNPUDarwin(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	powerRe := regexp.MustCompile(`ANE Power:\s+(\d+)\s+mW`)

	for range ticker.C {
		// Sample only the Apple Neural Engine
		cmd := exec.Command("sudo", "powermetrics", "--samplers", "ane_power", "-n", "1", "-i", "100")

		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			continue // Fail silently if root isn't available to avoid log spam
		}

		if match := powerRe.FindStringSubmatch(out.String()); len(match) > 1 {
			if mw, err := strconv.ParseFloat(match[1], 64); err == nil {

				// Only publish if the NPU is actually doing something (> 0 Watts)
				if mw > 0 {
					event := core.NewCTDPayload(nodeID, "session-live", "npu-ane").
						UpdateStatus("active").
						AddMetric("npu_power_w", mw/1000.0)

					if err := pub.Publish(event); err != nil {
						log.Printf("[Darwin NPU] Publish failed: %v", err)
					}
				}
			}
		}
	}
}
