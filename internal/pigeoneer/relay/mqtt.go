package relay

import (
	"encoding/json"
	"log"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoneer/benchmark"
	"github.com/drewv-labs/carrier-pigeons/pkg/core"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTClient wraps the Paho broker connection.
type MQTTClient struct {
	client mqtt.Client
}

// NewMQTTClient initializes and connects to the broker.
func NewMQTTClient(brokerURL, clientID string) *MQTTClient {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(clientID)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Relay failed to connect to broker: %v", token.Error())
	}

	return &MQTTClient{client: c}
}

// Publish serializes the event and fires it over the wire.
func (m *MQTTClient) Publish(event *core.CTDPayload) error {
	payloadBytes, err := event.ToJSON()
	if err != nil {
		return err
	}

	// QoS 1, non-retained message
	token := m.client.Publish(event.RoutingTopic(), 1, false, payloadBytes)
	token.Wait()
	return token.Error()
}

// Disconnect safely closes the network socket.
func (m *MQTTClient) Disconnect() {
	m.client.Disconnect(250)
}

func (c *MQTTClient) SubscribeControl(nodeID string, pub core.TelemetryPublisher) error {
	topic := "pigeons/control/" + nodeID

	token := c.client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		var cmd struct {
			Action   string            `json:"action"`
			Duration int               `json:"duration_sec"`
			Targets  benchmark.Targets `json:"targets"`
		}

		if err := json.Unmarshal(msg.Payload(), &cmd); err == nil {
			if cmd.Action == "benchmark-start" {
				go benchmark.RunSyntheticLoad(nodeID, cmd.Duration, cmd.Targets, pub)
			}
		}
	})

	token.Wait()
	return token.Error()
}
