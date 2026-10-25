// File: pkg/adapters/llm/llm.go (or internal/pigeons/adapters/llm.go during transition)
package adapters

import (
	"context"

	"github.com/drewv-labs/carrier-pigeons/pkg/adapters/db"
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

// BaseLLM (aliased as LLMAdapter for backward compatibility) defines the LLM mounting point.
type BaseLLM interface {
	// StreamChat returns a channel that yields tokens as they are generated.
	StreamChat(ctx context.Context, systemPrompt string, history []Message) (<-chan string, error)
}

type LLMAdapter = BaseLLM

type LLMConfig struct {
	Provider string
	Model    string
	Endpoint string
	APIKey   string
	Store    db.BaseDB
}

// NewLLMAdapter routes the configuration to the correct silicon backend.
func NewLLMAdapter(cfg LLMConfig) BaseLLM {
	switch cfg.Provider {
	case "ollama":
		return NewOllamaAdapter(cfg.Endpoint, cfg.Model, cfg.Store)

	// case "claude":
	// 	return NewClaudeAdapter(cfg.APIKey, cfg.Model, cfg.Store)
	//
	// case "openai":
	// 	return NewOpenAIAdapter(cfg.Endpoint, cfg.APIKey, cfg.Model, cfg.Store)
	//
	// case "llamacpp":
	// 	return NewLlamaCPPAdapter(cfg.Endpoint, cfg.Store)
	//
	// case "vllm":
	// 	return NewVLLMAdapter(cfg.Endpoint, cfg.Model, cfg.Store)

	default:
		// Default to local open-weights for edge architecture
		return NewOllamaAdapter("http://localhost:11434", "qwen3.6:35b-a3b-q8_0", cfg.Store)
	}
}
