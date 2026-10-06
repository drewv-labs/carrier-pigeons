package adapters

// LlamaCPPAdapter wraps the OpenAI client to target the local llama.cpp server.
type LlamaCPPAdapter struct {
	*OpenAIAdapter
}

func NewLlamaCPPAdapter(endpoint string) *LlamaCPPAdapter {
	if endpoint == "" {
		endpoint = "http://localhost:8080/v1/chat/completions"
	}

	// llama.cpp doesn't strictly require a model name or API key for single-model local instances
	return &LlamaCPPAdapter{
		OpenAIAdapter: NewOpenAIAdapter(endpoint, "local-no-key", "default"),
	}
}
