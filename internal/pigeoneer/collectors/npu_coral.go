package collectors

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

func monitorNPUCoral(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// The Linux apex driver mounts Coral PCIe/M.2 modules here
	basePath := "/sys/class/apex"

	for range ticker.C {
		// Glob allows us to dynamically support multiple Coral TPUs on a single node (e.g., dual Edge TPU M.2)
		devices, err := filepath.Glob(filepath.Join(basePath, "apex_*"))
		if err != nil || len(devices) == 0 {
			continue // Module unloaded or hardware disconnected
		}

		for _, devPath := range devices {
			devName := filepath.Base(devPath) // e.g., apex_0

			// The apex driver exposes temperature in millidegrees Celsius
			tempBytes, err := os.ReadFile(filepath.Join(devPath, "temp"))
			if err != nil {
				continue
			}

			tempStr := strings.TrimSpace(string(tempBytes))
			tempMilli, err := strconv.ParseFloat(tempStr, 64)
			if err != nil {
				continue
			}

			event := core.NewCTDPayload(nodeID, "session-live", "npu-coral-"+devName).
				UpdateStatus("active").
				AddMetric("npu_temp_c", tempMilli/1000.0) // Convert to standard Celsius

			if err := pub.Publish(event); err != nil {
				log.Printf("[Coral NPU Collector] Publish failed for %s: %v", devName, err)
			}
		}
	}
}
