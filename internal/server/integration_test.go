package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mlorentedev/pollex/internal/adapter"
)

type failingAdapter struct{}

func (f *failingAdapter) Name() string { return "failing" }
func (f *failingAdapter) Polish(ctx context.Context, text, systemPrompt string) (string, error) {
	return "", fmt.Errorf("intentional failure")
}
func (f *failingAdapter) Available() bool { return true }

type polishRequest struct {
	Text    string `json:"text"`
	ModelID string `json:"model_id"`
}

type polishResponse struct {
	Polished  string `json:"polished"`
	Model     string `json:"model"`
	ElapsedMs int64  `json:"elapsed_ms"`
}

type healthResponse struct {
	Status   string                      `json:"status"`
	Adapters map[string]json.RawMessage  `json:"adapters"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func newTestServer(t *testing.T, adapters map[string]adapter.LLMAdapter, models []adapter.ModelInfo) *httptest.Server {
	t.Helper()
	h := SetupMux(adapters, models, "test system prompt", "")
	return httptest.NewServer(h)
}

func newTestServerWithAPIKey(t *testing.T, adapters map[string]adapter.LLMAdapter, models []adapter.ModelInfo, apiKey string) *httptest.Server {
	t.Helper()
	h := SetupMux(adapters, models, "test system prompt", apiKey)
	return httptest.NewServer(h)
}

func defaultTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	adapters := map[string]adapter.LLMAdapter{"mock": &adapter.MockAdapter{}}
	models := []adapter.ModelInfo{{ID: "mock", Name: "Mock (dev)", Provider: "mock"}}
	return newTestServer(t, adapters, models)
}

func TestIntegration_PolishFullFlow(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	body, _ := json.Marshal(polishRequest{Text: "hello world", ModelID: "mock"})
	resp, err := http.Post(ts.URL+"/api/polish", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("CORS Allow-Origin: got %q, want %q", got, "*")
	}

	reqID := resp.Header.Get("X-Request-ID")
	if reqID == "" {
		t.Error("X-Request-ID header not set")
	}
	if len(reqID) != 32 {
		t.Errorf("X-Request-ID length: got %d, want 32", len(reqID))
	}

	var pr polishResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if pr.Polished != "Hello world" {
		t.Errorf("polished: got %q, want %q", pr.Polished, "Hello world")
	}
	if pr.Model != "mock" {
		t.Errorf("model: got %q, want %q", pr.Model, "mock")
	}
	if pr.ElapsedMs < 0 {
		t.Errorf("elapsed_ms: got %d, want >= 0", pr.ElapsedMs)
	}
}

func TestIntegration_HealthFullFlow(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("CORS Allow-Origin: got %q, want %q", got, "*")
	}
}

func TestIntegration_ModelsFullFlow(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/models")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var models []adapter.ModelInfo
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(models) != 1 {
		t.Fatalf("models count: got %d, want 1", len(models))
	}
	if models[0].ID != "mock" {
		t.Errorf("model id: got %q, want %q", models[0].ID, "mock")
	}
}

func TestIntegration_OptionsPreflightCORS(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodOptions, ts.URL+"/api/polish", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if got := resp.Header.Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Errorf("Allow-Methods: got %q, want %q", got, "GET, POST, OPTIONS")
	}
	if got := resp.Header.Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-API-Key") {
		t.Errorf("Allow-Headers: got %q, want to contain X-API-Key", got)
	}
}

func TestIntegration_UnknownRoute(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestIntegration_ConcurrentPolish(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	const n = 10
	var wg sync.WaitGroup
	errs := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body, _ := json.Marshal(polishRequest{
				Text:    fmt.Sprintf("message %d", i),
				ModelID: "mock",
			})
			resp, err := http.Post(ts.URL+"/api/polish", "application/json", bytes.NewReader(body))
			if err != nil {
				errs <- fmt.Errorf("request %d: %w", i, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				errs <- fmt.Errorf("request %d: status %d", i, resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

func TestIntegration_ContextCancellation(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{
		"mock": &adapter.MockAdapter{Delay: 5 * time.Second},
	}
	models := []adapter.ModelInfo{{ID: "mock", Name: "Mock", Provider: "mock"}}
	ts := newTestServer(t, adapters, models)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	body, _ := json.Marshal(polishRequest{Text: "hello", ModelID: "mock"})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, ts.URL+"/api/polish", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	_, err := http.DefaultClient.Do(req)
	if err == nil {
		t.Error("expected error from cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context") {
		t.Errorf("expected context error, got: %v", err)
	}
}

func TestIntegration_AdapterErrorPropagation(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{
		"failing": &failingAdapter{},
	}
	models := []adapter.ModelInfo{{ID: "failing", Name: "Failing", Provider: "test"}}
	ts := newTestServer(t, adapters, models)
	defer ts.Close()

	body, _ := json.Marshal(polishRequest{Text: "hello", ModelID: "failing"})
	resp, err := http.Post(ts.URL+"/api/polish", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusBadGateway)
	}

	var er errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.Contains(er.Error, "intentional failure") {
		t.Errorf("error: got %q, want to contain %q", er.Error, "intentional failure")
	}
}

func TestIntegration_RateLimit(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	for i := 0; i < 11; i++ {
		resp, err := http.Get(ts.URL + "/api/health")
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		resp.Body.Close()

		if i < 10 {
			if resp.StatusCode != http.StatusOK {
				t.Errorf("request %d: got %d, want %d", i, resp.StatusCode, http.StatusOK)
			}
		} else {
			if resp.StatusCode != http.StatusTooManyRequests {
				t.Errorf("request %d: got %d, want %d", i, resp.StatusCode, http.StatusTooManyRequests)
			}
		}
	}
}

func TestIntegration_OversizedBody(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	bigBody := strings.Repeat("x", 100*1024)
	payload := fmt.Sprintf(`{"text":"%s","model_id":"mock"}`, bigBody)
	resp, err := http.Post(ts.URL+"/api/polish", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusRequestEntityTooLarge)
	}
}

func TestIntegration_TextTooLong(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	longText := strings.Repeat("a", 10001)
	body, _ := json.Marshal(polishRequest{Text: longText, ModelID: "mock"})
	resp, err := http.Post(ts.URL+"/api/polish", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	var er errorResponse
	if err := json.NewDecoder(resp.Body).Decode(&er); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.Contains(er.Error, "too long") {
		t.Errorf("error: got %q, want to contain 'too long'", er.Error)
	}
}

func TestIntegration_APIKeyRequired(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{"mock": &adapter.MockAdapter{}}
	models := []adapter.ModelInfo{{ID: "mock", Name: "Mock (dev)", Provider: "mock"}}
	ts := newTestServerWithAPIKey(t, adapters, models, "test-key-123")
	defer ts.Close()

	t.Run("polish without key returns 401", func(t *testing.T) {
		body, _ := json.Marshal(polishRequest{Text: "hello", ModelID: "mock"})
		resp, err := http.Post(ts.URL+"/api/polish", "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("polish with valid key returns 200", func(t *testing.T) {
		body, _ := json.Marshal(polishRequest{Text: "hello", ModelID: "mock"})
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/polish", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "test-key-123")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("health exempt from auth", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/health")
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("models without key returns 401", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/api/models")
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("wrong key returns 401", func(t *testing.T) {
		body, _ := json.Marshal(polishRequest{Text: "hello", ModelID: "mock"})
		req, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/polish", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "wrong-key")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})
}

func TestIntegration_MetricsEndpoint(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	text := string(body)

	if !strings.Contains(text, "pollex_requests_total") {
		t.Error("metrics body missing pollex_requests_total")
	}
	if !strings.Contains(text, "go_goroutines") {
		t.Error("metrics body missing go_goroutines")
	}
}

func TestIntegration_MetricsExemptFromAuth(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{"mock": &adapter.MockAdapter{}}
	models := []adapter.ModelInfo{{ID: "mock", Name: "Mock (dev)", Provider: "mock"}}
	ts := newTestServerWithAPIKey(t, adapters, models, "secret-key")
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestIntegration_MetricsAfterPolish(t *testing.T) {
	ts := defaultTestServer(t)
	defer ts.Close()

	// POST a polish request
	body, _ := json.Marshal(polishRequest{Text: "test input", ModelID: "mock"})
	resp, err := http.Post(ts.URL+"/api/polish", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("polish request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("polish status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// GET /metrics and check pollex_* metrics are present
	resp, err = http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("metrics request: %v", err)
	}
	defer resp.Body.Close()

	metricsBody, _ := io.ReadAll(resp.Body)
	text := string(metricsBody)

	if !strings.Contains(text, `pollex_requests_total`) {
		t.Error("missing pollex_requests_total")
	}
	if !strings.Contains(text, "pollex_polish_duration_seconds") {
		t.Error("missing pollex_polish_duration_seconds")
	}
	if !strings.Contains(text, "pollex_input_chars") {
		t.Error("missing pollex_input_chars")
	}
}
