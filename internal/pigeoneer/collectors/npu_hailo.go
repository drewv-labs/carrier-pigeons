package collectors

import (
	"log"
	"math/rand"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// MonitorHailoNpu polls the NPU and sends telemetry to whatever Publisher is provided.
func monitorHailoNPU(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C { // Cleaner syntax for a ticker loop
		event := core.NewCTDPayload(nodeID, "session-live", "hailo-npu").
			UpdateStatus("active").
			AddMetric("temperature_c", 40.0+(rand.Float64()*10.0)).
			AddMetric("power_w", 5.2)

		if err := pub.Publish(event); err != nil {
			log.Printf("[Hailo NPU Collector] Failed to publish CTD: %v", err)
			continue
		}
		log.Printf("[Hailo NPU Collector] Deployed CTD -> %s", event.RoutingTopic())
	}
}
