package relay

import (
	"encoding/json"
	"log"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTClient manages the connection and the offline telemetry spool
type MQTTClient struct {
	client mqtt.Client
	nodeID string
	spool  chan *core.CTDPayload
}

// NewMQTTClient initializes the broker connection and starts the background flusher
func NewMQTTClient(brokerURL, nodeID string) *MQTTClient {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID("pigeoneer-" + nodeID)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetMaxReconnectInterval(10 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("[Relay] Warning: Broker unreachable on boot, telemetry will spool. (%v)", token.Error())
	}

	c := &MQTTClient{
		client: client,
		nodeID: nodeID,
		// 5000 event buffer gives you roughly 1.5 hours of offline capacity at 1Hz
		spool: make(chan *core.CTDPayload, 5000),
	}

	// Spin up the background telemetry flusher
	go c.flushLoop()

	return c
}

// Publish drops the payload into the spool channel. It will NEVER block the collectors.
func (c *MQTTClient) Publish(payload *core.CTDPayload) error {
	select {
	case c.spool <- payload:
		// Successfully spooled
	default:
		// The buffer is completely full. We drop the oldest event to make room for the new one.
		<-c.spool
		c.spool <- payload
		log.Printf("[Relay] Spool capacity reached. Dropping oldest telemetry.")
	}

	// Return nil to satisfy the core.TelemetryPublisher interface
	return nil
}

// flushLoop runs continuously, pulling from the spool and publishing when connected
func (c *MQTTClient) flushLoop() {
	for payload := range c.spool {
		// If we are disconnected, sleep briefly and push the payload back to the front of the queue
		if !c.client.IsConnected() {
			time.Sleep(2 * time.Second)
			c.Publish(payload) // Requeue
			continue
		}

		data, err := json.Marshal(payload)
		if err != nil {
			log.Printf("[Relay] Failed to marshal payload: %v", err)
			continue
		}

		topic := payload.RoutingTopic()
		token := c.client.Publish(topic, 1, false, data)

		// If the publish fails mid-flight, requeue it
		if token.Wait() && token.Error() != nil {
			log.Printf("[Relay] Flight error, requeuing payload: %v", token.Error())
			time.Sleep(1 * time.Second)
			c.Publish(payload)
		}
	}
}

// Disconnect gracefully shuts down the client
func (c *MQTTClient) Disconnect() {
	log.Println("[Relay] Disconnecting from broker...")
	c.client.Disconnect(250)
}
