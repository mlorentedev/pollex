package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClaudeAdapterPolish(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/messages" {
			t.Errorf("expected /v1/messages, got %s", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "sk-test" {
			t.Errorf("x-api-key: got %q, want %q", got, "sk-test")
		}
		if got := r.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Errorf("anthropic-version: got %q, want %q", got, "2023-06-01")
		}

		var req claudeMessagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Model != "claude-sonnet-4-5-20250929" {
			t.Errorf("model: got %q, want %q", req.Model, "claude-sonnet-4-5-20250929")
		}
		if req.System != "Fix grammar." {
			t.Errorf("system: got %q, want %q", req.System, "Fix grammar.")
		}
		if len(req.Messages) != 1 {
			t.Fatalf("expected 1 message, got %d", len(req.Messages))
		}
		if req.Messages[0].Role != "user" {
			t.Errorf("message role: got %q, want %q", req.Messages[0].Role, "user")
		}
		if req.MaxTokens != 4096 {
			t.Errorf("max_tokens: got %d, want 4096", req.MaxTokens)
		}

		resp := claudeMessagesResponse{
			Content: []claudeContentBlock{
				{Type: "text", Text: "I went to the store."},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	adapter := &ClaudeAdapter{
		BaseURL: srv.URL,
		APIKey:  "sk-test",
		Model:   "claude-sonnet-4-5-20250929",
		Client:  &http.Client{Timeout: 5 * time.Second},
	}

	got, err := adapter.Polish(context.Background(), "i goes to store", "Fix grammar.")
	if err != nil {
		t.Fatalf("Polish: %v", err)
	}
	if got != "I went to the store." {
		t.Errorf("got %q, want %q", got, "I went to the store.")
	}
}

func TestClaudeAdapterPolishServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"type": "error",
			"error": map[string]any{
				"type":    "invalid_request_error",
				"message": "bad request",
			},
		})
	}))
	defer srv.Close()

	adapter := &ClaudeAdapter{
		BaseURL: srv.URL,
		APIKey:  "sk-test",
		Model:   "claude-sonnet-4-5-20250929",
		Client:  &http.Client{Timeout: 5 * time.Second},
	}

	_, err := adapter.Polish(context.Background(), "hello", "prompt")
	if err == nil {
		t.Error("expected error on 400 response, got nil")
	}
}

func TestClaudeAdapterPolishEmptyContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := claudeMessagesResponse{Content: []claudeContentBlock{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	adapter := &ClaudeAdapter{
		BaseURL: srv.URL,
		APIKey:  "sk-test",
		Model:   "claude-sonnet-4-5-20250929",
		Client:  &http.Client{Timeout: 5 * time.Second},
	}

	_, err := adapter.Polish(context.Background(), "hello", "prompt")
	if err == nil {
		t.Error("expected error on empty content, got nil")
	}
}

func TestClaudeAdapterContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer srv.Close()

	adapter := &ClaudeAdapter{
		BaseURL: srv.URL,
		APIKey:  "sk-test",
		Model:   "claude-sonnet-4-5-20250929",
		Client:  &http.Client{Timeout: 5 * time.Second},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := adapter.Polish(ctx, "hello", "prompt")
	if err == nil {
		t.Error("expected error on cancelled context, got nil")
	}
}

func TestClaudeAdapterAvailable(t *testing.T) {
	adapter := &ClaudeAdapter{APIKey: "sk-test"}
	if !adapter.Available() {
		t.Error("expected available when API key is set")
	}
}

func TestClaudeAdapterNotAvailable(t *testing.T) {
	adapter := &ClaudeAdapter{APIKey: ""}
	if adapter.Available() {
		t.Error("expected not available when API key is empty")
	}
}

func TestClaudeAdapterName(t *testing.T) {
	adapter := &ClaudeAdapter{Model: "claude-sonnet-4-5-20250929"}
	want := "Claude (claude-sonnet-4-5-20250929)"
	if adapter.Name() != want {
		t.Errorf("got %q, want %q", adapter.Name(), want)
	}
}
