package benchmark

import (
	"crypto/sha256"
	"net"
	"os/exec"
	"runtime"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

type Targets struct {
	System bool `json:"system"`
	GPU    bool `json:"gpu"`
	NPU    bool `json:"npu"`
	Net    bool `json:"net"`
}

func RunSyntheticLoad(nodeID string, durationSec int, targets Targets, pub core.TelemetryPublisher) {
	runID := time.Now().UnixNano()

	startEvent := core.NewCTDPayload(nodeID, "benchmark", "sys-bench").
		UpdateStatus("benchmark-start").
		AddMetric("run_id", runID).
		AddMetric("duration_sec", durationSec).
		AddMetric("target_system", targets.System).
		AddMetric("target_gpu", targets.GPU).
		AddMetric("target_npu", targets.NPU).
		AddMetric("target_net", targets.Net)
	pub.Publish(startEvent)

	done := make(chan struct{})

	// 1. CPU/System Load
	if targets.System {
		for i := 0; i < runtime.NumCPU(); i++ {
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						hash := sha256.Sum256([]byte("carrier-pigeons-stress-test"))
						_ = hash
					}
				}
			}()
		}
	}

	// 2. Network Load (TCP Flood to loopback/gateway)
	if targets.Net {
		for i := 0; i < 10; i++ {
			go func() {
				for {
					select {
					case <-done:
						return
					default:
						// Fast dialing loop to spike network adapter I/O
						conn, err := net.DialTimeout("tcp", "127.0.0.1:80", 500*time.Millisecond)
						if err == nil {
							conn.Write(make([]byte, 1024))
							conn.Close()
						}
					}
				}
			}()
		}
	}

	// 3. GPU Load (Shell hook to native tools)
	if targets.GPU {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					// Example: Hitting a local metal-backed LLM to keep the GPU pegged
					cmd := exec.Command("curl", "-s", "-d", `{"model":"qwen2.5-coder","prompt":"benchmark loop"}`, "http://localhost:11434/api/generate")
					_ = cmd.Run()
				}
			}
		}()
	}

	// 4. NPU Load (Shell hook to Hailo/Coral CLI)
	if targets.NPU {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					// Shell out to native NPU benchmark utility
					// E.g., Hailo: exec.Command("hailortcli", "benchmark", "network.hef")
					// E.g., Apple ANE: exec.Command("sudo", "powermetrics", "-i", "10")
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	time.Sleep(time.Duration(durationSec) * time.Second)
	close(done)

	endEvent := core.NewCTDPayload(nodeID, "benchmark", "sys-bench").
		UpdateStatus("benchmark-end").
		AddMetric("run_id", runID)
	pub.Publish(endEvent)
}
