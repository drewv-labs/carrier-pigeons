package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// monitorHailo runs as an independent goroutine, polling the NPU.
func monitorHailo(client mqtt.Client, nodeID string) {
	// A Ticker acts like a metronome, firing a signal on a channel every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C // Block this specific goroutine until the ticker fires

		// Generate the strictly typed payload
		event := core.NewCTDPayload(nodeID, "session-live", "hailo-npu").
			UpdateStatus("active").
			AddMetric("temperature_c", 40.0+(rand.Float64()*10.0)). // Fake thermal data
			AddMetric("power_w", 5.2)

		payloadBytes, _ := event.ToJSON()

		// Publish asynchronously (QoS 1)
		client.Publish(event.RoutingTopic(), 1, false, payloadBytes)
		log.Printf("[Pigeoneer] Deployed CTD -> %s", event.RoutingTopic())
	}
}

// monitorSystem runs concurrently, polling basic OS stats.
func monitorSystem(client mqtt.Client, nodeID string) {
	ticker := time.NewTicker(10 * time.Second) // Wakes up on a different schedule
	defer ticker.Stop()

	for {
		<-ticker.C

		event := core.NewCTDPayload(nodeID, "session-live", "system-os").
			UpdateStatus("active").
			AddMetric("ram_usage_mb", 1024+(rand.Intn(500)))

		payloadBytes, _ := event.ToJSON()
		client.Publish(event.RoutingTopic(), 1, false, payloadBytes)
		log.Printf("[Pigeoneer] Deployed CTD -> %s", event.RoutingTopic())
	}
}

func main() {
	log.Println("Starting Pigeoneer: CTD Edge Agent...")

	// 1. Connect to the MQTT Broker
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883") // We will point this to your lab's IP later
	opts.SetClientID("pigeoneer-ada-node-1")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to broker: %v", token.Error())
	}
	log.Println("Connected to MQTT Broker.")

	// 2. Spawn the concurrent hardware monitors!
	// The 'go' keyword detaches these functions into their own lightweight threads.
	nodeID := "ada-node-1"
	go monitorHailo(client, nodeID)
	go monitorSystem(client, nodeID)

	// 3. Block the main thread so the program doesn't instantly exit
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Telemetry agents deployed. Press Ctrl+C to stop.")
	<-sigs

	log.Println("Shutting down Pigeoneer...")
	client.Disconnect(250)
}
