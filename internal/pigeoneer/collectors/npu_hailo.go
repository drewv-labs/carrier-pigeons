package collectors

import (
	"bytes"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

var (
	tempRegex  = regexp.MustCompile(`([0-9.]+)\s*C`)
	powerRegex = regexp.MustCompile(`([0-9.]+)\s*W`)
	// Regex to extract device PCIe addresses (e.g., "0000:01:00.0") from the scan command
	deviceRegex = regexp.MustCompile(`(?i)(?:Device|PCIe)\s+([0-9a-fA-F:\.]+)`)
)

func monitorNPUHailo(pub core.TelemetryPublisher, nodeID string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 1. Scan for all attached Hailo devices on every tick (handles hot-swapping/crashes)
		devices := scanHailoDevices()
		if len(devices) == 0 {
			continue
		}

		// 2. Poll each device independently
		for i, device := range devices {
			var temp, power float64

			// Measure Temp for specific device
			tempCmd := exec.Command("hailortcli", "measure-temp", "-d", device)
			var tempOut bytes.Buffer
			tempCmd.Stdout = &tempOut
			if err := tempCmd.Run(); err == nil {
				if matches := tempRegex.FindStringSubmatch(tempOut.String()); len(matches) > 1 {
					temp, _ = strconv.ParseFloat(matches[1], 64)
				}
			}

			// Measure Power for specific device
			powerCmd := exec.Command("hailortcli", "measure-power", "-d", device)
			var powerOut bytes.Buffer
			powerCmd.Stdout = &powerOut
			if err := powerCmd.Run(); err == nil {
				if matches := powerRegex.FindStringSubmatch(powerOut.String()); len(matches) > 1 {
					power, _ = strconv.ParseFloat(matches[1], 64)
				}
			}

			if temp == 0 && power == 0 {
				continue
			}

			// 3. Tag component uniquely (e.g., hailo-npu-0, hailo-npu-1)
			componentID := "hailo-npu-" + strconv.Itoa(i)

			event := core.NewCTDPayload(nodeID, "session-live", componentID).
				UpdateStatus("active").
				AddMetric("device_id", device). // Store the PCIe address as metadata
				AddMetric("temperature_c", temp).
				AddMetric("power_w", power)

			if err := pub.Publish(event); err != nil {
				log.Printf("[Hailo Collector] Publish failed for %s: %v", componentID, err)
			}
		}
	}
}

// scanHailoDevices asks the Hailo driver to list all attached NPUs.
func scanHailoDevices() []string {
	cmd := exec.Command("hailortcli", "scan")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil
	}

	var devices []string
	// hailortcli scan outputs multiple lines. We extract the PCIe device ID from each.
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		matches := deviceRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			devices = append(devices, strings.TrimSpace(matches[1]))
		}
	}
	return devices
}
