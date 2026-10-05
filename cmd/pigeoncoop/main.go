package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drewv-labs/carrier-pigeons/pkg/core"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx/v5/pgxpool"
)

// makeMessageHandler is a closure (factory) that injects the database connection
// pool into the MQTT callback.
func makeMessageHandler(db *pgxpool.Pool) mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		var event core.CTDPayload

		if err := json.Unmarshal(msg.Payload(), &event); err != nil {
			log.Printf("ERROR: Malformed CTD payload dropped from topic %s: %v", msg.Topic(), err)
			return
		}

		// Create a strict 5-second timeout context for the database transaction
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		query := `
			INSERT INTO telemetric_ledger
			(node_id, session_id, event_timestamp, component, status, metrics)
			VALUES ($1, $2, $3, $4, $5, $6)
		`

		// Execute the SQL statement. pgx automatically marshals the Go map into JSONB.
		_, err := db.Exec(ctx, query,
			event.NodeID,
			event.SessionID,
			event.Timestamp,
			event.Component,
			event.Status,
			event.Metrics,
		)

		if err != nil {
			log.Printf("DB INSERT ERROR: Failed to write event from %s: %v", event.NodeID, err)
			return
		}

		log.Printf("[LEDGER INJECT] %s/%s -> PostgreSQL Insert Success", event.NodeID, event.Component)
	}
}

func main() {
	log.Println("Starting PigeonCoop: Telemetric Ledger Injector...")

	// 1. Initialize PostgreSQL Connection Pool
	// In production, this URL would come from os.Getenv("DATABASE_URL")
	dbURL := "postgres://drewv:ctd_password@localhost:5432/edge_ledger"

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatalf("Database connection dropped: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL Ledger.")

	// 2. Configure MQTT Client
	opts := mqtt.NewClientOptions()
	opts.AddBroker("tcp://localhost:1883")
	opts.SetClientID("pigeoncoop-ledger-injector-01")

	// Inject the database pool into our message handler
	opts.SetDefaultPublishHandler(makeMessageHandler(dbPool))

	opts.OnConnect = func(client mqtt.Client) {
		topic := "drewv/ctd/v1/#"
		client.Subscribe(topic, 1, nil).Wait()
		log.Printf("Subscribed to MQTT wildcard: %s", topic)
	}
	opts.SetAutoReconnect(true)

	// 3. Connect to MQTT Broker
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", token.Error())
	}

	// 4. Block for Graceful Shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	log.Println("PigeonCoop is running. Waiting for telemetry...")
	<-sigs

	log.Println("Shutting down gracefully...")
	client.Disconnect(250)
	// dbPool.Close() is automatically called by the 'defer' statement above
}
