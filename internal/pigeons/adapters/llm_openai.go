package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type OpenAIAdapter struct {
	endpoint string
	apiKey   string
	model    string
	client   *http.Client
}

func NewOpenAIAdapter(endpoint, apiKey, model string) *OpenAIAdapter {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}
	return &OpenAIAdapter{
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
		client:   &http.Client{},
	}
}

func (a *OpenAIAdapter) Chat(ctx context.Context, systemPrompt string, history []Message) (string, error) {
	messages := []Message{{Role: "system", Content: systemPrompt}}
	messages = append(messages, history...)

	payload := map[string]any{
		"model":    a.model,
		"messages": messages,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", a.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	if a.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+a.apiKey)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Choices) == 0 {
		return "", fmt.Errorf("failed to decode response")
	}

	return result.Choices[0].Message.Content, nil
}
