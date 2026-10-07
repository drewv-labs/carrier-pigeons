package collectors

import (
	"log"
	"os/exec"
	"runtime"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

// monitorGPU detects the local hardware and routes to the correct sensor binary.
func monitorGPU(pub core.TelemetryPublisher, nodeID string) {
	if runtime.GOOS == "darwin" {
		log.Println("[GPU Router] Apple/Darwin hardware detected. Booting powermetrics collector.")
		go monitorGPUDarwin(pub, nodeID)
		return
	}
	if isInstalled("nvidia-smi") {
		log.Println("[GPU Router] Nvidia hardware detected. Booting NVML collector.")
		go monitorGPUNvidia(pub, nodeID)
	}
	if isInstalled("rocm-smi") {
		log.Println("[GPU Router] AMD hardware detected. Booting ROCm collector.")
		go monitorGPUAMD(pub, nodeID)
	}
	if isInstalled("intel_gpu_top") {
		log.Println("[GPU Router] Intel hardware detected. Booting sysfs collector.")
		go monitorGPUIntel(pub, nodeID)
	}
}

// isInstalled checks if a command exists in the system PATH.
func isInstalled(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
