package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoneer/collectors"
	"github.com/drewv-labs/carrier-pigeons/internal/pigeoneer/relay"

	// Blank import forces init() to execute, registering the 'benchmark' runner
	_ "github.com/drewv-labs/carrier-pigeons/internal/pigeoneer/runners"
)

func main() {
	log.Println("Starting Pigeoneer: CTD Edge Agent...")
	nodeID := "ada-node-1" // Eventually pull this from env vars or config

	// 1. Initialize the network relay with the offline spooler
	// FIX: Pass nodeID cleanly, NewMQTTClient handles the Client ID prefix now
	mqttRelay := relay.NewMQTTClient("tcp://localhost:1883", nodeID)
	defer mqttRelay.Disconnect()
	log.Println("Network relay established.")

	// 2. Bind the control topic to listen for dynamic Runner dispatches
	go func() {
		log.Printf("Binding runner control topic for node: %s", nodeID)
		if err := mqttRelay.ListenForRunners(nodeID, mqttRelay); err != nil {
			log.Printf("Fatal error on runner control topic: %v", err)
		}
	}()

	// 3. Deploy concurrent passive hardware collectors
	go collectors.MonitorOS(mqttRelay, nodeID)
	go collectors.MonitorAdapters(mqttRelay, nodeID)

	// 4. Block until shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Telemetry & Runner agents deployed. Press Ctrl+C to stop.")
	<-sigs

	log.Println("Shutting down Pigeoneer...")
}
