package collectors

import (
	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// MonitorAdapters is the parent router that initializes all peripheral hardware sensors.
func MonitorAdapters(pub core.TelemetryPublisher, nodeID string) {
	go monitorGPU(pub, nodeID)
	go monitorNPU(pub, nodeID)
	go monitorNet(pub, nodeID)
}
