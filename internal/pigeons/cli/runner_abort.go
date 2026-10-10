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

var runnerAbortCmd = &cobra.Command{
	Use:   "abort [runner-name] [node-ids...]",
	Short: "Send a hard kill signal to an active runner on edge nodes",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runnerName := args[0]
		targetNodes := args[1:]

		opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883").SetClientID("pigeons-cli-abort")
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			fmt.Printf("Fatal: Could not reach broker: %v\n", token.Error())
			os.Exit(1)
		}
		defer client.Disconnect(250)

		env := runner.Envelope{
			Action: "abort",
			Runner: runnerName,
		}
		payload, _ := json.Marshal(env)

		for _, nodeID := range targetNodes {
			topic := "pigeons/control/" + nodeID
			token := client.Publish(topic, 1, false, payload)
			token.Wait()
			if token.Error() == nil {
				fmt.Printf("[🛑] Dispatched kill signal for %q to %s\n", runnerName, nodeID)
			} else {
				log.Printf("Failed to dispatch to %s: %v", nodeID, token.Error())
			}
		}
	},
}

func init() {
	runnerCmd.AddCommand(runnerAbortCmd)
}
