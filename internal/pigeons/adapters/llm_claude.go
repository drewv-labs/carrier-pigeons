package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ClaudeAdapter struct {
	apiKey string
	model  string
	client *http.Client
}

func NewClaudeAdapter(apiKey, model string) *ClaudeAdapter {
	if model == "" {
		model = "claude-3-5-sonnet-latest"
	}
	return &ClaudeAdapter{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

func (a *ClaudeAdapter) Chat(ctx context.Context, systemPrompt string, history []Message) (string, error) {
	payload := map[string]any{
		"model":      a.model,
		"max_tokens": 1024,
		"system":     systemPrompt, // Claude extracts system context here
		"messages":   history,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Content) == 0 {
		return "", fmt.Errorf("failed to decode claude response")
	}

	return result.Content[0].Text, nil
}
