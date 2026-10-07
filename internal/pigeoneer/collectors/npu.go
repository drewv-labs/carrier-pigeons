package collectors

import (
	"log"
	"runtime"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// monitorNPU detects local neural processing units and routes to the correct sensor binary.
func monitorNPU(pub core.TelemetryPublisher, nodeID string) {
	if runtime.GOOS == "darwin" {
		log.Println("[NPU Router] Apple Neural Engine detected. Booting ANE collector.")
		go monitorNPUDarwin(pub, nodeID)
	}
	if isInstalled("hailortcli") {
		log.Println("[NPU Router] Hailo NPU detected. Booting Hailo collector.")
		go monitorNPUHailo(pub, nodeID)

	}
	if isInstalled("edgetpu_compiler") {
		log.Println("[NPU Router] Google Coral TPU detected. Booting Coral collector.")
		go monitorNPUCoral(pub, nodeID)
	}
	return
}
