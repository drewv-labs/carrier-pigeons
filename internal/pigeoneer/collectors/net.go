package collectors

import (
	"log"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// monitorNet detects the networking stack and routes to the correct interface poller.
func monitorNet(pub core.TelemetryPublisher, nodeID string) {
	if isInstalled("ubus") {
		log.Println("[Net Router] OpenWrt ubus detected. Booting embedded network collector.")
		go monitorNetOpenWrt(pub, nodeID)
	} else {
		log.Println("[Net Router] Standard Linux networking detected. Booting sysfs network collector.")
		// go monitorNetLinux(pub, nodeID) // Placeholder for future /proc/net/dev parser
	}
}
