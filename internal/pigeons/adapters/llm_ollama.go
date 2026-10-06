package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OllamaAdapter struct {
	endpoint string
	model    string
	client   *http.Client
}

func NewOllamaAdapter(endpoint, model string) *OllamaAdapter {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	return &OllamaAdapter{
		endpoint: endpoint + "/api/chat",
		model:    model,
		client:   &http.Client{},
	}
}

func (a *OllamaAdapter) Chat(ctx context.Context, systemPrompt string, history []Message) (string, error) {
	messages := []Message{{Role: "system", Content: systemPrompt}}
	messages = append(messages, history...)

	payload := map[string]any{
		"model":    a.model,
		"messages": messages,
		"stream":   false,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", a.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Message Message `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode ollama response")
	}

	return result.Message.Content, nil
}
