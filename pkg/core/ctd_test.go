package core

import (
	"strings"
	"testing"
)

func TestNewCTDPayload(t *testing.T) {
	payload := NewCTDPayload("ada-1", "sess-884", "cpu")

	if payload.NodeID != "ada-1" {
		t.Errorf("expected NodeID 'ada-1', got '%s'", payload.NodeID)
	}
	if payload.Status != "initialized" {
		t.Errorf("expected Status 'initialized', got '%s'", payload.Status)
	}
	if payload.Metrics == nil {
		t.Fatal("expected Metrics map to be initialized, got nil")
	}
}

func TestPayloadChaining(t *testing.T) {
	payload := NewCTDPayload("ada-1", "sess-884", "hailo-npu").
		UpdateStatus("inference_complete").
		AddMetric("temperature_c", 45.5)

	if payload.Status != "inference_complete" {
		t.Errorf("expected Status 'inference_complete', got '%s'", payload.Status)
	}

	// Type assertion is required when pulling from map[string]interface{}
	if temp, ok := payload.Metrics["temperature_c"].(float64); !ok || temp != 45.5 {
		t.Errorf("expected temperature_c to be float64 45.5, got %v", payload.Metrics["temperature_c"])
	}
}

func TestRoutingTopic(t *testing.T) {
	payload := NewCTDPayload("ada-node-1", "sess-1", "hailo-npu")
	expected := "drewv/ctd/v1/ada-node-1/hailo-npu"

	if got := payload.RoutingTopic(); got != expected {
		t.Errorf("expected topic '%s', got '%s'", expected, got)
	}
}

func TestToJSON(t *testing.T) {
	payload := NewCTDPayload("ada-1", "sess-1", "cpu").
		AddMetric("load", 2.5)

	b, err := payload.ToJSON()
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	jsonStr := string(b)

	if !strings.Contains(jsonStr, `"node_id":"ada-1"`) {
		t.Errorf("JSON missing node_id: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"metrics":{"load":2.5}`) {
		t.Errorf("JSON missing metrics: %s", jsonStr)
	}
}
