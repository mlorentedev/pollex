package main

import "context"

// LLMAdapter abstracts an LLM backend for text polishing.
type LLMAdapter interface {
	// Name returns a human-readable name for this adapter.
	Name() string

	// Polish sends text to the LLM with the given system prompt and returns polished text.
	Polish(ctx context.Context, text, systemPrompt string) (string, error)

	// Available reports whether this adapter is ready to serve requests.
	Available() bool
}

// ModelInfo describes a model exposed via /api/models.
type ModelInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
}
