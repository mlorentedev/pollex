package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOllamaAdapterPolish(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/chat" {
			t.Errorf("expected /api/chat, got %s", r.URL.Path)
		}

		var req ollamaChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Model != "qwen2.5:1.5b" {
			t.Errorf("model: got %q, want %q", req.Model, "qwen2.5:1.5b")
		}
		if req.Stream {
			t.Error("expected stream=false")
		}
		if len(req.Messages) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(req.Messages))
		}
		if req.Messages[0].Role != "system" {
			t.Errorf("first message role: got %q, want %q", req.Messages[0].Role, "system")
		}
		if req.Messages[1].Role != "user" {
			t.Errorf("second message role: got %q, want %q", req.Messages[1].Role, "user")
		}

		resp := ollamaChatResponse{
			Message: ollamaMessage{Role: "assistant", Content: "I went to the store."},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	a := &OllamaAdapter{
		BaseURL: srv.URL,
		Model:   "qwen2.5:1.5b",
		Client:  &http.Client{Timeout: 5 * time.Second},
	}

	got, err := a.Polish(context.Background(), "i goes to store", "Fix grammar.")
	if err != nil {
		t.Fatalf("Polish: %v", err)
	}
	if got != "I went to the store." {
		t.Errorf("got %q, want %q", got, "I went to the store.")
	}
}

func TestOllamaAdapterPolishServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	a := &OllamaAdapter{
		BaseURL: srv.URL,
		Model:   "qwen2.5:1.5b",
		Client:  &http.Client{Timeout: 5 * time.Second},
	}

	_, err := a.Polish(context.Background(), "hello", "prompt")
	if err == nil {
		t.Error("expected error on 500 response, got nil")
	}
}

func TestOllamaAdapterPolishContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer srv.Close()

	a := &OllamaAdapter{
		BaseURL: srv.URL,
		Model:   "qwen2.5:1.5b",
		Client:  &http.Client{Timeout: 5 * time.Second},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := a.Polish(ctx, "hello", "prompt")
	if err == nil {
		t.Error("expected error on cancelled context, got nil")
	}
}

func TestOllamaAdapterAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	a := &OllamaAdapter{
		BaseURL: srv.URL,
		Model:   "qwen2.5:1.5b",
		Client:  &http.Client{Timeout: 1 * time.Second},
	}

	if !a.Available() {
		t.Error("expected available when server is up")
	}
}

func TestOllamaAdapterNotAvailable(t *testing.T) {
	a := &OllamaAdapter{
		BaseURL: "http://localhost:99999",
		Model:   "qwen2.5:1.5b",
		Client:  &http.Client{Timeout: 1 * time.Second},
	}

	if a.Available() {
		t.Error("expected not available when server is unreachable")
	}
}

func TestOllamaAdapterName(t *testing.T) {
	a := &OllamaAdapter{Model: "qwen2.5:1.5b"}
	if a.Name() != "Ollama (qwen2.5:1.5b)" {
		t.Errorf("got %q, want %q", a.Name(), "Ollama (qwen2.5:1.5b)")
	}
}
