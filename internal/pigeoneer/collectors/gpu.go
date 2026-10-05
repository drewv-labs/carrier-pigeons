package collectors

import (
	"log"
	"os/exec"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// MonitorGPU detects the local hardware and routes to the correct sensor binary.
func MonitorGPU(pub core.TelemetryPublisher, nodeID string) {
	if isInstalled("nvidia-smi") {
		log.Println("[GPU Router] Nvidia hardware detected. Booting NVML collector.")
		go monitorNvidiaGPU(pub, nodeID)
	}
	if isInstalled("rocm-smi") {
		log.Println("[GPU Router] AMD hardware detected. Booting ROCm collector.")
		go monitorAMDGPU(pub, nodeID)
	}
	if isInstalled("intel_gpu_top") {
		log.Println("[GPU Router] Intel hardware detected. Booting sysfs collector.")
		go monitorIntelGPU(pub, nodeID)
	}
}

// isInstalled checks if a command exists in the system PATH.
func isInstalled(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
