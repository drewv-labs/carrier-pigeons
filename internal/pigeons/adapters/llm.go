package adapters

import (
	"context"

	"github.com/drewv-labs/carrier-pigeons/internal/pigeoncoop/ledger"
)

// Tool schemas mapped to the OpenAI/Ollama specification
type ToolCallFunction struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type LLMAdapter interface {
	Chat(ctx context.Context, systemPrompt string, history []Message) (string, error)
}

type LLMConfig struct {
	Provider string
	Model    string
	Endpoint string
	APIKey   string
	Store    *ledger.Store // Injected so the LLM can query Postgres
}

// NewLLMAdapter routes the configuration to the correct silicon backend.
func NewLLMAdapter(cfg LLMConfig) LLMAdapter {
	switch cfg.Provider {

	case "ollama":
		return NewOllamaAdapter(cfg.Endpoint, cfg.Model, cfg.Store)

	case "claude":
		return NewClaudeAdapter(cfg.APIKey, cfg.Model)

	case "openai":
		return NewOpenAIAdapter(cfg.Endpoint, cfg.APIKey, cfg.Model)

	case "llamacpp":
		return NewLlamaCPPAdapter(cfg.Endpoint)

	case "vllm":
		return NewVLLMAdapter(cfg.Endpoint, cfg.Model)

	default:
		// Default to local open-weights for edge architecture
		return NewOllamaAdapter("http://localhost:11434", "qwen3.6:35b-a3b-q8_0", cfg.Store)
	}
}
