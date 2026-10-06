package broker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Ledger defines the contract for storing ingested telemetry.
type Ledger interface {
	InsertEvent(ctx context.Context, event *core.CTDPayload) error
}

// Subscriber manages the MQTT connection and routes data to the Ledger.
type Subscriber struct {
	client mqtt.Client
	ledger Ledger
}

// NewSubscriber creates a new MQTT listener and injects the storage dependency.
func NewSubscriber(brokerURL string, ledger Ledger) (*Subscriber, error) {
	sub := &Subscriber{ledger: ledger}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID("pigeoncoop-ledger-injector")
	opts.SetDefaultPublishHandler(sub.handleMessage)
	opts.SetAutoReconnect(true)

	opts.OnConnect = func(client mqtt.Client) {
		// 1. Subscribe to the telemetry firehose
		client.Subscribe("drewv/ctd/v1/#", 1, nil)

		// 2. Subscribe to the registry state changes with the specific Ntfy handler
		client.Subscribe("drewv/ctd/v1/registry/+/status", 1, sub.handleRegistryUpdate)

		log.Println("Subscribed to MQTT wildcard & registry topics")
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	sub.client = client
	return sub, nil
}

// handleMessage unmarshals the incoming JSON and passes it to the Ledger.
func (s *Subscriber) handleMessage(client mqtt.Client, msg mqtt.Message) {
	var event core.CTDPayload

	if err := json.Unmarshal(msg.Payload(), &event); err != nil {
		log.Printf("ERROR: Malformed CTD payload dropped: %v", err)
		return
	}

	// 5-second timeout for the database insert
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.ledger.InsertEvent(ctx, &event); err != nil {
		log.Printf("DB INSERT ERROR for %s: %v", event.NodeID, err)
		return
	}

	log.Printf("[LEDGER INJECT] %s/%s -> PostgreSQL Insert Success", event.NodeID, event.Component)
}

// Disconnect cleanly closes the network socket.
func (s *Subscriber) Disconnect() {
	s.client.Disconnect(250)
}
