package collectors

import (
	"log"
	"runtime"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// MonitorOS detects the host operating system and routes to the correct system poller.
func MonitorOS(pub core.TelemetryPublisher, nodeID string) {
	switch runtime.GOOS {
	case "linux":
		if isInstalled("ubus") {
			log.Println("[OS Router] OpenWrt host detected.")
			go monitorOSOpenWrt(pub, nodeID)
		} else {
			log.Println("[OS Router] Standard Unix host detected.")
			go monitorOSUnix(pub, nodeID)
		}
	case "windows":
		log.Println("[OS Router] Windows host detected. Booting PowerShell CIM collector.")
		go monitorOSWindows(pub, nodeID)
	default:
		// Graceful fallback for macOS (darwin) or BSD
		log.Printf("[OS Router] Generic Unix host detected for %s.", runtime.GOOS)
		go monitorOSUnix(pub, nodeID)
	}
}
