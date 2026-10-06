package collectors

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

func monitorNetOpenWrt(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// OpenWrt ubus returns a nested JSON object for network devices
	type UbusDevice struct {
		Up         bool `json:"up"`
		Statistics struct {
			RxBytes float64 `json:"rx_bytes"`
			TxBytes float64 `json:"tx_bytes"`
		} `json:"statistics"`
	}

	for range ticker.C {
		cmd := exec.Command("ubus", "call", "network.device", "status")
		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			continue
		}

		var devices map[string]UbusDevice
		if err := json.Unmarshal(out.Bytes(), &devices); err != nil {
			log.Printf("[Net Collector] Failed to parse ubus output: %v", err)
			continue
		}

		for iface, data := range devices {
			// Skip loopback and inactive interfaces to save ledger space
			if !data.Up || iface == "lo" {
				continue
			}

			// Uniquely tag each interface adapter (e.g., net-eth0, net-wlan1)
			componentID := "net-" + iface

			event := core.NewCTDPayload(nodeID, "session-live", componentID).
				UpdateStatus("active").
				AddMetric("rx_bytes", data.Statistics.RxBytes).
				AddMetric("tx_bytes", data.Statistics.TxBytes)

			if err := pub.Publish(event); err != nil {
				log.Printf("[Net Collector] Publish failed for %s: %v", componentID, err)
			}
		}
	}
}
