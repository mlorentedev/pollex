package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// nousTestServer returns a server that records call count and responds with the
// given status. For 200 it returns a valid chat completion carrying content.
func nousTestServer(code int, content string, calls *int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls != nil {
			*calls++
		}
		if code == http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(nousChatResponse{
				Choices: []nousChoice{{Message: nousMessageResp{Content: content}}},
			})
			return
		}
		w.WriteHeader(code)
	}))
}

func chainAdapter(url, model string) *NousAdapter {
	return &NousAdapter{BaseURL: url, APIKey: "k", Model: model, Client: &http.Client{Timeout: 2 * time.Second}}
}

// Quota/availability failures must advance: mimo 429, qwen 500, gemma 200.
func TestFallbackChain_AdvancesOnQuotaThenSucceeds(t *testing.T) {
	var c1, c2, c3 int
	s1 := nousTestServer(http.StatusTooManyRequests, "", &c1)
	s2 := nousTestServer(http.StatusInternalServerError, "", &c2)
	s3 := nousTestServer(http.StatusOK, "polished text", &c3)
	defer s1.Close()
	defer s2.Close()
	defer s3.Close()

	chain := &FallbackChain{Adapters: []LLMAdapter{
		chainAdapter(s1.URL, "mimo-v2.5"),
		chainAdapter(s2.URL, "qwen3.6"),
		chainAdapter(s3.URL, "gemma4"),
	}}

	got, err := chain.Polish(context.Background(), "x", "p")
	if err != nil {
		t.Fatalf("Polish: %v", err)
	}
	if got != "polished text" {
		t.Errorf("got %q, want %q", got, "polished text")
	}
	if c1 != 1 || c2 != 1 || c3 != 1 {
		t.Errorf("call counts: mimo=%d qwen=%d gemma=%d, want 1/1/1", c1, c2, c3)
	}
}

// A network error (unreachable host) on the first model must advance.
func TestFallbackChain_AdvancesOnNetworkError(t *testing.T) {
	var c2 int
	s2 := nousTestServer(http.StatusOK, "recovered", &c2)
	defer s2.Close()

	chain := &FallbackChain{Adapters: []LLMAdapter{
		chainAdapter("http://127.0.0.1:1", "mimo-v2.5"), // connection refused
		chainAdapter(s2.URL, "qwen3.6"),
	}}

	got, err := chain.Polish(context.Background(), "x", "p")
	if err != nil {
		t.Fatalf("Polish: %v", err)
	}
	if got != "recovered" || c2 != 1 {
		t.Errorf("got %q (qwen calls=%d), want %q / 1", got, c2, "recovered")
	}
}

// A client error (400) recurs on every model, so the chain must fail fast and
// NOT call the next adapter.
func TestFallbackChain_FailFastOnClientError(t *testing.T) {
	var c1, c2 int
	s1 := nousTestServer(http.StatusBadRequest, "", &c1)
	s2 := nousTestServer(http.StatusOK, "should-not-reach", &c2)
	defer s1.Close()
	defer s2.Close()

	chain := &FallbackChain{Adapters: []LLMAdapter{
		chainAdapter(s1.URL, "mimo-v2.5"),
		chainAdapter(s2.URL, "qwen3.6"),
	}}

	_, err := chain.Polish(context.Background(), "x", "p")
	if err == nil {
		t.Fatal("expected error on 400, got nil")
	}
	if c2 != 0 {
		t.Errorf("qwen should not be called on a 400, but was called %d times", c2)
	}
}

// All models failing returns a wrapped error after trying every one.
func TestFallbackChain_AllFail(t *testing.T) {
	var c1, c2, c3 int
	s1 := nousTestServer(http.StatusServiceUnavailable, "", &c1)
	s2 := nousTestServer(http.StatusServiceUnavailable, "", &c2)
	s3 := nousTestServer(http.StatusServiceUnavailable, "", &c3)
	defer s1.Close()
	defer s2.Close()
	defer s3.Close()

	chain := &FallbackChain{Adapters: []LLMAdapter{
		chainAdapter(s1.URL, "mimo-v2.5"),
		chainAdapter(s2.URL, "qwen3.6"),
		chainAdapter(s3.URL, "gemma4"),
	}}

	_, err := chain.Polish(context.Background(), "x", "p")
	if err == nil {
		t.Fatal("expected error when all adapters fail, got nil")
	}
	if !strings.Contains(err.Error(), "all 3 adapters failed") {
		t.Errorf("error should mention all adapters failed: %v", err)
	}
	if c1 != 1 || c2 != 1 || c3 != 1 {
		t.Errorf("each adapter should be tried once: %d/%d/%d", c1, c2, c3)
	}
}

// First model succeeds -> no fallback.
func TestFallbackChain_FirstSucceeds(t *testing.T) {
	var c1, c2 int
	s1 := nousTestServer(http.StatusOK, "first", &c1)
	s2 := nousTestServer(http.StatusOK, "second", &c2)
	defer s1.Close()
	defer s2.Close()

	chain := &FallbackChain{Adapters: []LLMAdapter{
		chainAdapter(s1.URL, "mimo-v2.5"),
		chainAdapter(s2.URL, "qwen3.6"),
	}}

	got, err := chain.Polish(context.Background(), "x", "p")
	if err != nil {
		t.Fatalf("Polish: %v", err)
	}
	if got != "first" || c2 != 0 {
		t.Errorf("got %q (qwen calls=%d), want %q / 0", got, c2, "first")
	}
}

func TestFallbackChain_Available(t *testing.T) {
	chain := &FallbackChain{Adapters: []LLMAdapter{
		&NousAdapter{APIKey: ""},
		&NousAdapter{APIKey: "k"},
	}}
	if !chain.Available() {
		t.Error("expected available when at least one adapter has a key")
	}
	empty := &FallbackChain{Adapters: []LLMAdapter{&NousAdapter{APIKey: ""}}}
	if empty.Available() {
		t.Error("expected not available when no adapter has a key")
	}
}

func TestFallbackChain_Name(t *testing.T) {
	chain := &FallbackChain{Label: "NaN Cloud (auto)"}
	if chain.Name() != "NaN Cloud (auto)" {
		t.Errorf("got %q, want %q", chain.Name(), "NaN Cloud (auto)")
	}
}

// With no Label, Name composes the children's names.
func TestFallbackChain_NameComposed(t *testing.T) {
	chain := &FallbackChain{Adapters: []LLMAdapter{
		&NousAdapter{Model: "mimo-v2.5"},
		&NousAdapter{Model: "qwen3.6"},
	}}
	want := "FallbackChain(Nous (mimo-v2.5) -> Nous (qwen3.6))"
	if chain.Name() != want {
		t.Errorf("got %q, want %q", chain.Name(), want)
	}
}

func TestFallbackChain_EmptyAdaptersError(t *testing.T) {
	chain := &FallbackChain{}
	if _, err := chain.Polish(context.Background(), "x", "p"); err == nil {
		t.Error("expected error when chain has no adapters, got nil")
	}
}
