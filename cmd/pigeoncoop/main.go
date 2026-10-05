package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoncoop/broker"
	"github.com/drewv-labs/carrier-pigeons/internal/pigeoncoop/ledger"
)

func main() {
	log.Println("Starting PigeonCoop: Telemetric Ledger Injector...")

	// 1. Initialize the PostgreSQL Store
	dbURL := "postgres://drewv:ctd_password@localhost:5432/edge_ledger"

	store, err := ledger.NewStore(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL ledger: %v", err)
	}
	defer store.Close()
	log.Println("Connected to PostgreSQL Ledger.")

	// 2. Initialize the MQTT Subscriber, injecting the 'store' as the Ledger interface
	mqttSub, err := broker.NewSubscriber("tcp://localhost:1883", store)
	if err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttSub.Disconnect()

	// 3. Block for Graceful Shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("PigeonCoop is running. Waiting for telemetry...")
	<-sigs

	log.Println("Shutting down gracefully...")
}
