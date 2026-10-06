package broker

import (
	"bytes"
	"log"
	"net/http"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// handleRegistryUpdate intercepts LWT messages and triggers Ntfy alerts.
func (s *Subscriber) handleRegistryUpdate(client mqtt.Client, msg mqtt.Message) {
	// Topic structure: drewv/ctd/v1/registry/pigeoneer-ada-node-1/status
	parts := strings.Split(msg.Topic(), "/")
	if len(parts) < 6 {
		return
	}

	nodeID := parts[4]
	status := string(msg.Payload())

	log.Printf("[Registry] %s state changed to: %s", nodeID, status)

	if status == "OFFLINE" || status == "ONLINE" {
		go sendNtfyAlert(nodeID, status)
	}
}

// sendNtfyAlert pushes a notification to your local Ntfy instance.
func sendNtfyAlert(nodeID, status string) {
	// Adjust this URL to point to the Ntfy server in your homelab edge architecture
	url := "http://ntfy.drewv.local/carrier-pigeons"

	var message string
	if status == "OFFLINE" {
		message = "🚨 CRITICAL: " + nodeID + " has dropped off the network."
	} else {
		message = "✅ RECOVERY: " + nodeID + " is back online."
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBufferString(message))
	req.Header.Set("Title", "Ada Lovespace Hardware Alert")

	if status == "OFFLINE" {
		req.Header.Set("Tags", "warning,skull")
		req.Header.Set("Priority", "high")
	} else {
		req.Header.Set("Tags", "white_check_mark")
	}

	client := &http.Client{}
	if _, err := client.Do(req); err != nil {
		log.Printf("[Ntfy] Failed to send alert: %v", err)
	}
}
