package collectors

import (
	"log"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// monitorNPU detects local neural processing units and routes to the correct sensor binary.
func monitorNPU(pub core.TelemetryPublisher, nodeID string) {
	if isInstalled("hailortcli") {
		log.Println("[NPU Router] Hailo architecture detected. Booting Hailo collector.")
		go monitorNPUHailo(pub, nodeID)

	}
	// if isInstalled("edgetpu_compiler") {
	// 	log.Println("[NPU Router] Coral Edge TPU detected. Booting Coral collector.")
	// 	go monitorCoralNpu(pub, nodeID)
	// }
}
