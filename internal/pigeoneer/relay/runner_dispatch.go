package relay

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
	"github.com/drewv-labs/carrier-pigeons/pkg/runner"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ListenForRunners binds the MQTT client to the node's exclusive control topic.
func (c *MQTTClient) ListenForRunners(nodeID string, pub core.TelemetryPublisher) error {
	topic := "pigeons/control/" + nodeID

	// Thread-safe map to hold active runner cancellation functions
	var activeRunners sync.Map

	token := c.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		var env runner.Envelope
		if err := json.Unmarshal(msg.Payload(), &env); err != nil {
			log.Printf("[Runner Dispatch] Invalid envelope dropped: %v", err)
			return
		}

		if err := env.Validate(); err != nil {
			log.Printf("[Runner Dispatch] Envelope validation failed: %v", err)
			return
		}

		// Default to start if no action is provided (backwards compatibility)
		action := env.Action
		if action == "" {
			action = "start"
		}

		// HANDLE CANCELLATION
		if action == "abort" {
			if cancelRaw, active := activeRunners.Load(env.Runner); active {
				log.Printf("[Runner Dispatch] 🛑 ABORT SIGNAL RECEIVED: Killing %q on %s", env.Runner, nodeID)
				cancel := cancelRaw.(context.CancelFunc)
				cancel()
				activeRunners.Delete(env.Runner)
			} else {
				log.Printf("[Runner Dispatch] Abort ignored: %q is not currently running.", env.Runner)
			}
			return
		}

		// HANDLE EXECUTION
		task, exists := runner.Get(env.Runner)
		if !exists {
			log.Printf("[Runner Dispatch] Unknown runner requested: %q", env.Runner)
			return
		}

		// Prevent overlapping executions of the exact same runner
		if _, active := activeRunners.Load(env.Runner); active {
			log.Printf("[Runner Dispatch] Runner %q is already active. Ignoring duplicate start.", env.Runner)
			return
		}

		// Create a cancellable context and store it in the registry
		ctx, cancel := context.WithCancel(context.Background())
		activeRunners.Store(env.Runner, cancel)

		// Fire in an isolated goroutine
		go func() {
			defer activeRunners.Delete(env.Runner) // Clean up registry when finished

			log.Printf("[Runner Dispatch] 🚀 Executing %q on %s", env.Runner, nodeID)
			if err := task.Run(ctx, nodeID, env.Params, pub); err != nil {
				log.Printf("[Runner Dispatch] Runner %q failed: %v", env.Runner, err)
			}
		}()
	})

	token.Wait()
	return token.Error()
}
