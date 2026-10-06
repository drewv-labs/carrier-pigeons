package collectors

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

func monitorOSOpenWrt(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	type UbusSystemInfo struct {
		Uptime float64 `json:"uptime"`
		Memory struct {
			Total float64 `json:"total"`
			Free  float64 `json:"free"`
		} `json:"memory"`
	}

	for range ticker.C {
		cmd := exec.Command("ubus", "call", "system", "info")
		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			continue
		}

		var sysInfo UbusSystemInfo
		if err := json.Unmarshal(out.Bytes(), &sysInfo); err != nil {
			log.Printf("[OS Collector] Failed to parse ubus system info: %v", err)
			continue
		}

		event := core.NewCTDPayload(nodeID, "session-live", "host-os").
			UpdateStatus("active").
			AddMetric("uptime_s", sysInfo.Uptime).
			AddMetric("mem_total_mb", sysInfo.Memory.Total/1024/1024).
			AddMetric("mem_free_mb", sysInfo.Memory.Free/1024/1024)

		if err := pub.Publish(event); err != nil {
			log.Printf("[OS Collector] Publish failed: %v", err)
		}
	}
}
