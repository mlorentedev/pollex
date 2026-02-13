package adapter

import "context"

// LLMAdapter defines the contract for LLM backends.
type LLMAdapter interface {
	Name() string
	Polish(ctx context.Context, text, systemPrompt string) (string, error)
	Available() bool
}

// ModelInfo is exposed via GET /api/models.
type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}
