package relay

import (
	"context"
	"encoding/json"
	"log"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
	"github.com/drewv-labs/carrier-pigeons/pkg/runner"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ListenForRunners binds the MQTT client to the node's exclusive control topic.
func (c *MQTTClient) ListenForRunners(nodeID string, pub core.TelemetryPublisher) error {
	topic := "pigeons/control/" + nodeID

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

		task, exists := runner.Get(env.Runner)
		if !exists {
			log.Printf("[Runner Dispatch] Unknown runner requested: %q", env.Runner)
			return
		}

		// Fire in an isolated goroutine to prevent blocking the MQTT receiver thread
		go func() {
			log.Printf("[Runner Dispatch] Executing %q on %s", env.Runner, nodeID)

			// We inject a background context, which could easily be swapped for a
			// context with a timeout if the Envelope provides one.
			ctx := context.Background()

			if err := task.Run(ctx, nodeID, env.Params, pub); err != nil {
				log.Printf("[Runner Dispatch] Runner %q failed: %v", env.Runner, err)
			}
		}()
	})

	token.Wait()
	return token.Error()
}
