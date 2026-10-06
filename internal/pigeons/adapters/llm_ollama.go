package adapters

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoncoop/ledger"
)

type OllamaAdapter struct {
	endpoint string
	model    string
	store    *ledger.Store
	client   *http.Client
}

func NewOllamaAdapter(endpoint, model string, store *ledger.Store) *OllamaAdapter {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	return &OllamaAdapter{
		endpoint: endpoint + "/api/chat",
		model:    model,
		store:    store,
		client:   &http.Client{},
	}
}

func (a *OllamaAdapter) StreamChat(ctx context.Context, systemPrompt string, history []Message) (<-chan string, error) {
	tokenChan := make(chan string)

	go func() {
		defer close(tokenChan)
		messages := []Message{{Role: "system", Content: systemPrompt}}
		messages = append(messages, history...)

		for {
			payload := map[string]any{
				"model":    a.model,
				"messages": messages,
				"tools":    PigeonTools,
				"stream":   true, // Enable real-time token streaming
			}

			body, _ := json.Marshal(payload)
			req, err := http.NewRequestWithContext(ctx, "POST", a.endpoint, bytes.NewBuffer(body))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := a.client.Do(req)
			if err != nil {
				return
			}

			scanner := bufio.NewScanner(resp.Body)
			var toolCalls []ToolCall
			var fullContent string

			// Parse the NDJSON stream chunk by chunk
			for scanner.Scan() {
				var chunk struct {
					Message Message `json:"message"`
				}
				if err := json.Unmarshal(scanner.Bytes(), &chunk); err == nil {
					// If there is text, push it immediately to the UI channel
					if chunk.Message.Content != "" {
						tokenChan <- chunk.Message.Content
						fullContent += chunk.Message.Content
					}
					// If the chunk contains a tool call, capture it
					if len(chunk.Message.ToolCalls) > 0 {
						toolCalls = chunk.Message.ToolCalls
					}
				}
			}
			resp.Body.Close()

			// Base Case: The LLM finished talking and did not request a tool. We are done.
			if len(toolCalls) == 0 {
				return
			}

			// Recursive Case: The LLM wants to run a PostgreSQL query.
			messages = append(messages, Message{
				Role:      "assistant",
				Content:   fullContent,
				ToolCalls: toolCalls,
			})

			for _, tc := range toolCalls {
				// We pipe a status update to the UI so you know what it's doing behind the scenes
				tokenChan <- fmt.Sprintf("\n*(Running tool: %s...)*\n", tc.Function.Name)

				observation := ExecuteTool(ctx, tc.Function.Name, tc.Function.Arguments, a.store)
				messages = append(messages, Message{
					Role:    "tool",
					Content: observation,
				})
			}
			// The loop restarts, pushing the new DB context back to Ollama
		}
	}()

	return tokenChan, nil
}
