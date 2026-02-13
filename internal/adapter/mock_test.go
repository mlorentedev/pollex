package adapter

import (
	"context"
	"testing"
	"time"
)

func TestMockAdapterPolish(t *testing.T) {
	m := &MockAdapter{}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"capitalizes first letter", "hello world", "Hello world"},
		{"trims whitespace", "  hello world  ", "Hello world"},
		{"already capitalized", "Hello world", "Hello world"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.Polish(context.Background(), tt.input, "system prompt")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMockAdapterContextCancel(t *testing.T) {
	m := &MockAdapter{Delay: 5 * time.Second}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := m.Polish(ctx, "hello", "prompt")
	if err == nil {
		t.Error("expected error on cancelled context, got nil")
	}
}

func TestMockAdapterAvailable(t *testing.T) {
	m := &MockAdapter{}
	if !m.Available() {
		t.Error("mock adapter should always be available")
	}
}

func TestMockAdapterName(t *testing.T) {
	m := &MockAdapter{}
	if m.Name() != "Mock" {
		t.Errorf("got %q, want %q", m.Name(), "Mock")
	}
}
