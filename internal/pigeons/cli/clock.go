package cli

import (
	"encoding/json"
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
)

var (
	benchmarkSecs int
	targetSystem  bool
	targetGPU     bool
	targetNPU     bool
	targetNet     bool
)

var clockCmd = &cobra.Command{
	Use:   "clock [node-ids...]",
	Short: "Trigger a remote benchmark on targeted edge nodes",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Default to system if no flags are provided
		if !targetSystem && !targetGPU && !targetNPU && !targetNet {
			targetSystem = true
		}

		opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			fmt.Printf("Fatal: Could not connect to MQTT broker: %v\n", token.Error())
			os.Exit(1)
		}
		defer client.Disconnect(250)

		payload := map[string]any{
			"action":       "benchmark-start",
			"duration_sec": benchmarkSecs,
			"targets": map[string]bool{
				"system": targetSystem,
				"gpu":    targetGPU,
				"npu":    targetNPU,
				"net":    targetNet,
			},
		}
		body, _ := json.Marshal(payload)

		for _, nodeID := range args {
			topic := "pigeons/control/" + nodeID
			token := client.Publish(topic, 1, false, body)
			token.Wait()
			fmt.Printf("[+] Dispatching %ds synthetic load to %s (Sys:%v GPU:%v NPU:%v Net:%v)\n",
				benchmarkSecs, nodeID, targetSystem, targetGPU, targetNPU, targetNet)
		}
	},
}

func init() {
	rootCmd.AddCommand(clockCmd)
	clockCmd.Flags().IntVar(&benchmarkSecs, "secs", 60, "Duration of the synthetic workload in seconds")
	clockCmd.Flags().BoolVar(&targetSystem, "system", false, "Pin the CPU with concurrent SHA-256 hashing")
	clockCmd.Flags().BoolVar(&targetGPU, "gpu", false, "Trigger GPU inference / compute load")
	clockCmd.Flags().BoolVar(&targetNPU, "npu", false, "Trigger NPU accelerator load (Hailo/Coral/ANE)")
	clockCmd.Flags().BoolVar(&targetNet, "net", false, "Flood local network interfaces with dummy packets")
}
