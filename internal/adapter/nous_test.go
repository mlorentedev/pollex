package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNousAdapterPolish covers the happy path plus the two NaN-specific
// contract points: the Bearer auth header and enable_thinking:false in the body.
func TestNousAdapterPolish(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("expected /v1/chat/completions, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization: got %q, want %q", got, "Bearer test-key")
		}

		var req nousChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "mimo-v2.5" {
			t.Errorf("model: got %q, want %q", req.Model, "mimo-v2.5")
		}
		if len(req.Messages) != 2 || req.Messages[0].Role != "system" || req.Messages[1].Role != "user" {
			t.Fatalf("expected [system,user] messages, got %+v", req.Messages)
		}
		// Reasoning must be suppressed to keep polishing fast + quota-cheap.
		if v, ok := req.ChatTemplateKwargs["enable_thinking"]; !ok || v != false {
			t.Errorf("chat_template_kwargs.enable_thinking: got %v (present=%v), want false", v, ok)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nousChatResponse{
			Choices: []nousChoice{
				{Message: nousMessageResp{Role: "assistant", Content: "I went to the store."}},
			},
		})
	}))
	defer srv.Close()

	a := &NousAdapter{
		BaseURL: srv.URL,
		APIKey:  "test-key",
		Model:   "mimo-v2.5",
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

// TestNousAdapterIgnoresReasoningContent: NaN reasoning models stream the
// chain-of-thought in the non-standard reasoning_content field. The polished
// output must be choices[0].message.content ONLY — never the reasoning.
func TestNousAdapterIgnoresReasoningContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nousChatResponse{
			Choices: []nousChoice{
				{Message: nousMessageResp{
					Role:             "assistant",
					ReasoningContent: "Let me think... the verb tense is wrong, I should fix it.",
					Content:          "I went to the store.",
				}},
			},
		})
	}))
	defer srv.Close()

	a := &NousAdapter{BaseURL: srv.URL, APIKey: "k", Model: "mimo-v2.5", Client: &http.Client{Timeout: 5 * time.Second}}

	got, err := a.Polish(context.Background(), "i goes to store", "Fix grammar.")
	if err != nil {
		t.Fatalf("Polish: %v", err)
	}
	if got != "I went to the store." {
		t.Errorf("output leaked reasoning or wrong content: got %q", got)
	}
}

func TestNousAdapterPolishServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	a := &NousAdapter{BaseURL: srv.URL, APIKey: "k", Model: "mimo-v2.5", Client: &http.Client{Timeout: 5 * time.Second}}

	if _, err := a.Polish(context.Background(), "hello", "prompt"); err == nil {
		t.Error("expected error on 500 response, got nil")
	}
}

func TestNousAdapterPolishAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(nousErrorResponse{Error: struct {
			Message string `json:"message"`
		}{Message: "model not found"}})
	}))
	defer srv.Close()

	a := &NousAdapter{BaseURL: srv.URL, APIKey: "k", Model: "no-such-model", Client: &http.Client{Timeout: 5 * time.Second}}

	_, err := a.Polish(context.Background(), "hello", "prompt")
	if err == nil {
		t.Fatal("expected error on 404 response, got nil")
	}
}

func TestNousAdapterPolishEmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nousChatResponse{Choices: []nousChoice{}})
	}))
	defer srv.Close()

	a := &NousAdapter{BaseURL: srv.URL, APIKey: "k", Model: "mimo-v2.5", Client: &http.Client{Timeout: 5 * time.Second}}

	if _, err := a.Polish(context.Background(), "hello", "prompt"); err == nil {
		t.Error("expected error on empty choices, got nil")
	}
}

func TestNousAdapterAvailable(t *testing.T) {
	if (&NousAdapter{APIKey: "k"}).Available() != true {
		t.Error("expected available when API key is set")
	}
	if (&NousAdapter{APIKey: ""}).Available() != false {
		t.Error("expected not available when API key is empty")
	}
}

func TestNousAdapterName(t *testing.T) {
	a := &NousAdapter{Model: "mimo-v2.5"}
	want := "Nous (mimo-v2.5)"
	if a.Name() != want {
		t.Errorf("got %q, want %q", a.Name(), want)
	}
}

// TestNousAdapterBaseURLWithV1 verifies a base URL carrying a trailing /v1
// (the NAN_BASE_URL convention) does not produce a doubled /v1/v1 path.
func TestNousAdapterBaseURLWithV1(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path: got %s, want /v1/chat/completions", r.URL.Path)
		}
		hit = true
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nousChatResponse{Choices: []nousChoice{{Message: nousMessageResp{Content: "ok"}}}})
	}))
	defer srv.Close()

	a := &NousAdapter{BaseURL: srv.URL + "/v1", APIKey: "k", Model: "mimo-v2.5", Client: &http.Client{Timeout: 5 * time.Second}}
	if _, err := a.Polish(context.Background(), "x", "p"); err != nil {
		t.Fatalf("Polish: %v", err)
	}
	if !hit {
		t.Error("server was never hit at the expected path")
	}
}
