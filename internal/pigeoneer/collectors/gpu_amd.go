package collectors

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

func monitorAmdGpu(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 1. Use the correct AMD flags to request specific hardware data
		cmd := exec.Command("rocm-smi", "--showtemp", "--showuse", "--showmeminfo", "vram", "--json")
		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			log.Printf("[AMD GPU Collector] Sensor timeout: %v", err)
			continue
		}

		// 2. rocm-smi outputs a JSON object keyed by the card name
		// Example: {"card0": {"Temperature (Sensor edge) (C)": "45.0", "GPU use (%)": "99", ...}}
		var amdData map[string]map[string]any
		if err := json.Unmarshal(out.Bytes(), &amdData); err != nil {
			log.Printf("[AMD GPU Collector] Failed to parse JSON: %v", err)
			continue
		}

		// 3. Loop through any AMD cards found on the system
		for cardName, metrics := range amdData {
			// AMD returns the values as strings inside the JSON, so we type-assert and parse them
			tempStr, _ := metrics["Temperature (Sensor edge) (C)"].(string)
			utilStr, _ := metrics["GPU use (%)"].(string)
			memStr, _ := metrics["VRAM Total Used Memory (B)"].(string)

			temp, _ := strconv.ParseFloat(tempStr, 64)
			util, _ := strconv.ParseFloat(utilStr, 64)
			memBytes, _ := strconv.ParseFloat(memStr, 64)

			event := core.NewCTDPayload(nodeID, "session-live", "amd-gpu-"+cardName).
				UpdateStatus("active").
				AddMetric("temperature_c", temp).
				AddMetric("utilization_pct", util).
				AddMetric("memory_used_mb", memBytes/1024/1024)

			if err := pub.Publish(event); err != nil {
				log.Printf("[AMD GPU Collector] Publish failed for %s: %v", cardName, err)
			}
		}
	}
}
