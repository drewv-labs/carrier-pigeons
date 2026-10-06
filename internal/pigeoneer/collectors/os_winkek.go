package collectors

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

func monitorOSWindows(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Struct to catch the JSON output from PowerShell
	type WinMem struct {
		TotalVisibleMemorySize uint64 `json:"TotalVisibleMemorySize"`
		FreePhysicalMemory     uint64 `json:"FreePhysicalMemory"`
	}

	for range ticker.C {
		// Ask Windows for memory stats and format the output as JSON
		cmd := exec.Command("powershell", "-NoProfile", "-Command",
			"Get-CimInstance Win32_OperatingSystem | Select-Object TotalVisibleMemorySize, FreePhysicalMemory | ConvertTo-Json")

		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			continue
		}

		var mem WinMem
		if err := json.Unmarshal(out.Bytes(), &mem); err != nil {
			log.Printf("[Windows OS Collector] JSON parse error: %v", err)
			continue
		}

		// Windows CIM returns memory in Kilobytes; convert to Megabytes
		event := core.NewCTDPayload(nodeID, "session-live", "host-os").
			UpdateStatus("active").
			AddMetric("mem_total_mb", float64(mem.TotalVisibleMemorySize)/1024).
			AddMetric("mem_free_mb", float64(mem.FreePhysicalMemory)/1024)

		if err := pub.Publish(event); err != nil {
			log.Printf("[Windows OS Collector] Publish failed: %v", err)
		}
	}
}
