package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/drewv-labs/carrier-pigeons/pkg/runner"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
)

var (
	benchSecs   int
	benchSystem bool
	benchGPU    bool
	benchNPU    bool
)

var runnerBenchmarkCmd = &cobra.Command{
	Use:   "benchmark [node-ids...]",
	Short: "Dispatch a targeted synthetic benchmark to edge nodes",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Default to system CPU burn if no targets are specified
		if !benchSystem && !benchGPU && !benchNPU {
			benchSystem = true
		}

		// Spin up a transient MQTT client just for the CLI dispatch
		opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("pigeons-cli-dispatcher")
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			fmt.Printf("Fatal: Could not reach local MQTT broker: %v\n", token.Error())
			os.Exit(1)
		}
		defer client.Disconnect(250)

		// Construct the rigid protocol envelope
		env := runner.Envelope{
			Runner: "benchmark",
			Params: map[string]any{
				"duration_sec":  benchSecs,
				"target_system": benchSystem,
				"target_gpu":    benchGPU,
				"target_npu":    benchNPU,
			},
		}

		payload, err := json.Marshal(env)
		if err != nil {
			log.Fatalf("Failed to marshal runner envelope: %v", err)
		}

		// Dispatch to all targeted nodes
		for _, nodeID := range args {
			topic := "pigeons/control/" + nodeID
			token := client.Publish(topic, 1, false, payload)
			token.Wait()
			if token.Error() != nil {
				fmt.Printf("[-] Failed to dispatch to %s: %v\n", nodeID, token.Error())
			} else {
				fmt.Printf("[+] Dispatched 'benchmark' runner to %s (Duration: %ds, Sys: %v, GPU: %v, NPU: %v)\n",
					nodeID, benchSecs, benchSystem, benchGPU, benchNPU)
			}
		}
	},
}

func init() {
	runnerCmd.AddCommand(runnerBenchmarkCmd)

	runnerBenchmarkCmd.Flags().IntVar(&benchSecs, "secs", 60, "Duration of the workload in seconds")
	runnerBenchmarkCmd.Flags().BoolVar(&benchSystem, "system", false, "Pin the CPU cores")
	runnerBenchmarkCmd.Flags().BoolVar(&benchGPU, "gpu", false, "Trigger GPU load")
	runnerBenchmarkCmd.Flags().BoolVar(&benchNPU, "npu", false, "Trigger NPU accelerator load")
}
