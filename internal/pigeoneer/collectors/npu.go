package collectors

import (
	"log"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// MonitorNPU detects local neural processing units and routes to the correct sensor binary.
func MonitorNPU(pub core.TelemetryPublisher, nodeID string) {
	if isInstalled("hailortcli") {
		log.Println("[NPU Router] Hailo architecture detected. Booting Hailo collector.")
		go monitorHailoNPU(pub, nodeID)

	}
	// if isInstalled("edgetpu_compiler") {
	// 	log.Println("[NPU Router] Coral Edge TPU detected. Booting Coral collector.")
	// 	go monitorCoralNpu(pub, nodeID)
	// }
}
