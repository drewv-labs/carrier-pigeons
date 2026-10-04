package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"
)

func main() {
	log.Println("Starting PigeonCoop: Telemetric Ledger Injector...")

	// 1. Set up a channel to listen for interrupt signals (Ctrl+C, systemd stop)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// 2. (Placeholder) Initialize PostgreSQL Connection Pool
	log.Println("Connecting to PostgreSQL Ledger...")

	// 3. (Placeholder) Initialize MQTT Client
	topic := "drewv/ctd/v1/#"
	log.Printf("Subscribed to MQTT topic wildcard: %s", topic)

	// Simulated incoming MQTT message callback
	handleIncomingMQTT := func(payloadBytes []byte) {
		// Allocate an empty CTDPayload struct in memory
		var event core.CTDPayload

		// Pass the memory address (&event) to the unmarshaler
		err := json.Unmarshal(payloadBytes, &event)
		if err != nil {
			log.Printf("Error: Malformed CTD payload dropped: %v", err)
			return
		}

		// The payload is now strictly typed and safe to use!
		log.Printf("Ingested - Node: %s | Component: %s | Status: %s | Metrics: %v",
			event.NodeID, event.Component, event.Status, event.Metrics)
	}

	// Simulating an incoming byte stream from the MQTT broker
	dummyPayload := []byte(`{"node_id":"ada-1","session_id":"sess-123","timestamp":"2026-10-04T12:00:00Z","component":"hailo-npu","status":"inference_complete","metrics":{"temperature_c":45.5}}`)
	handleIncomingMQTT(dummyPayload)

	// 4. Block the main thread until a shutdown signal is received
	log.Println("PigeonCoop is running. Waiting for telemetry...")
	<-sigs

	log.Println("Shutting down PigeonCoop gracefully...")
}
