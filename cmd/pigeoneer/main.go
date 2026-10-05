package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoneer/collectors"
	"github.com/drewv-labs/carrier-pigeons/internal/pigeoneer/relay"
)

func main() {
	log.Println("Starting Pigeoneer: CTD Edge Agent...")
	nodeID := "ada-node-1"

	// 1. Initialize the network relay
	mqttRelay := relay.NewMQTTClient("tcp://localhost:1883", "pigeoneer-"+nodeID)
	defer mqttRelay.Disconnect()
	log.Println("Network relay established.")

	// 2. Deploy concurrent hardware collectors
	go collectors.MonitorHailoNpu(mqttRelay, nodeID)
	go collectors.MonitorSystem(mqttRelay, nodeID)

	// 3. Block until shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Telemetry agents deployed. Press Ctrl+C to stop.")
	<-sigs

	log.Println("Shutting down Pigeoneer...")
}
