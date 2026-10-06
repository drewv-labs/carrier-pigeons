package adapters

import (
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

func (a *OllamaAdapter) Chat(ctx context.Context, systemPrompt string, history []Message) (string, error) {
	// 1. Build the initial context window
	messages := []Message{{Role: "system", Content: systemPrompt}}
	messages = append(messages, history...)

	// 2. The Agentic Loop
	for {
		payload := map[string]any{
			"model":    a.model,
			"messages": messages,
			"tools":    PigeonTools, // Inject the capabilities
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

		var result struct {
			Message Message `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return "", fmt.Errorf("failed to decode ollama response")
		}
		resp.Body.Close()

		// 3. Base Case: The LLM returned natural language (No tools requested)
		if len(result.Message.ToolCalls) == 0 {
			return result.Message.Content, nil
		}

		// 4. Recursive Case: The LLM requested a tool execution
		// Append the LLM's tool request to history so it remembers its own thought process
		messages = append(messages, result.Message)

		for _, toolCall := range result.Message.ToolCalls {
			// Execute the SQL query against the ledger
			observation := ExecuteTool(ctx, toolCall.Function.Name, toolCall.Function.Arguments, a.store)

			// Append the raw database observation as a new "tool" role message
			messages = append(messages, Message{
				Role:    "tool",
				Content: observation,
			})
		}

		// The loop repeats, posting the updated context window (including the DB results) back to Ollama
	}
}
