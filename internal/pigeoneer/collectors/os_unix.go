package collectors

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

func monitorOSUnix(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		event := core.NewCTDPayload(nodeID, "session-live", "host-os").
			UpdateStatus("active")

		// Parse System Load Average
		if loadBytes, err := os.ReadFile("/proc/loadavg"); err == nil {
			fields := strings.Fields(string(loadBytes))
			if len(fields) > 0 {
				if load1, err := strconv.ParseFloat(fields[0], 64); err == nil {
					event.AddMetric("load_1m", load1)
				}
			}
		}

		// Parse Total Memory Capacity
		if memBytes, err := os.ReadFile("/proc/meminfo"); err == nil {
			lines := strings.Split(string(memBytes), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "MemTotal:") {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						if kb, err := strconv.ParseFloat(fields[1], 64); err == nil {
							event.AddMetric("mem_total_mb", kb/1024)
						}
					}
					break
				}
			}
		}

		if err := pub.Publish(event); err != nil {
			log.Printf("[OS Collector] Publish failed: %v", err)
		}
	}
}
