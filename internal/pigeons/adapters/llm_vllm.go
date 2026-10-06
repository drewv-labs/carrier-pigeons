package adapters

// VLLMAdapter wraps the OpenAI client to target a local or edge vLLM cluster.
type VLLMAdapter struct {
	*OpenAIAdapter
}

func NewVLLMAdapter(endpoint, model string) *VLLMAdapter {
	if endpoint == "" {
		endpoint = "http://localhost:8000/v1/chat/completions"
	}

	// vLLM requires the explicit model name (e.g., "Qwen/Qwen2.5-Coder-32B-Instruct")
	return &VLLMAdapter{
		OpenAIAdapter: NewOpenAIAdapter(endpoint, "vllm-no-key", model),
	}
}
