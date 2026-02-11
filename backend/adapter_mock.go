package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MockAdapter returns simulated responses with a configurable delay.
// Used for development and testing without a real LLM backend.
type MockAdapter struct {
	Delay time.Duration
}

func (m *MockAdapter) Name() string { return "Mock" }

func (m *MockAdapter) Polish(ctx context.Context, text, systemPrompt string) (string, error) {
	if m.Delay > 0 {
		select {
		case <-time.After(m.Delay):
		case <-ctx.Done():
			return "", fmt.Errorf("mock: %w", ctx.Err())
		}
	}

	// Simple mock: capitalize first letter of each sentence, trim whitespace.
	polished := strings.TrimSpace(text)
	if len(polished) > 0 && polished[0] >= 'a' && polished[0] <= 'z' {
		polished = strings.ToUpper(polished[:1]) + polished[1:]
	}

	return polished, nil
}

func (m *MockAdapter) Available() bool { return true }
