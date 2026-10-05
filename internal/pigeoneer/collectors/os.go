package collectors

import (
	"log"
	"math/rand"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// monitorSystem runs concurrently, polling basic OS stats.
func MonitorSystem(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second) // Wakes up on a different schedule
	defer ticker.Stop()

	for range ticker.C {
		event := core.NewCTDPayload(nodeID, "session-live", "system-os").
			UpdateStatus("active").
			AddMetric("ram_usage_mb", 1024+(rand.Intn(500)))

		if err := pub.Publish(event); err != nil {
			log.Printf("[System Collector] Failed to publish CTD: %v", err)
		}
		log.Printf("[System Collector] Deployed CTD -> %s", event.RoutingTopic())
	}
}
