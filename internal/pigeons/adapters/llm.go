package adapters

import (
	"context"
)

// Message represents a single turn in the chat history.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMAdapter is the universal interface for all inference backends.
type LLMAdapter interface {
	Chat(ctx context.Context, systemPrompt string, history []Message) (string, error)
}

// Factory configuration for dynamic instantiation.
type LLMConfig struct {
	Provider string // ollama, claude, openai, llamacpp, vllm
	Model    string
	Endpoint string
	APIKey   string
}

// NewLLMAdapter routes the configuration to the correct silicon backend.
func NewLLMAdapter(cfg LLMConfig) LLMAdapter {
	switch cfg.Provider {
	case "ollama":
		return NewOllamaAdapter(cfg.Endpoint, cfg.Model)
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
		return NewOllamaAdapter("http://localhost:11434", "qwen2.5-coder")
	}
}
