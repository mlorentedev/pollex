package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mlorentedev/pollex/internal/adapter"
)

func TestHandleHealth(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{
		"mock": &adapter.MockAdapter{},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	Health(adapters, "test").ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp healthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("status: got %q, want %q", resp.Status, "ok")
	}
	if resp.Adapters == nil {
		t.Fatal("adapters: got nil")
	}
	mockStatus, ok := resp.Adapters["mock"]
	if !ok {
		t.Fatal("adapters: missing mock")
	}
	if !mockStatus.Available {
		t.Error("mock adapter: got unavailable, want available")
	}
}

func TestHandleHealthUnavailableClaude(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{
		"mock":   &adapter.MockAdapter{},
		"claude": &adapter.ClaudeAdapter{APIKey: "", Model: "claude-sonnet"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	Health(adapters, "test").ServeHTTP(w, req)

	var resp healthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	claudeStatus := resp.Adapters["claude"]
	if claudeStatus.Available {
		t.Error("claude adapter: got available, want unavailable")
	}
	if claudeStatus.Reason != "no API key" {
		t.Errorf("claude reason: got %q, want %q", claudeStatus.Reason, "no API key")
	}
}

func TestHandleHealthMixedAdapters(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{
		"mock":          &adapter.MockAdapter{},
		"claude-sonnet": &adapter.ClaudeAdapter{APIKey: "sk-test", Model: "claude-sonnet"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	Health(adapters, "test").ServeHTTP(w, req)

	var resp healthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Adapters) != 2 {
		t.Fatalf("adapters count: got %d, want 2", len(resp.Adapters))
	}
	if !resp.Adapters["mock"].Available {
		t.Error("mock: got unavailable")
	}
	if !resp.Adapters["claude-sonnet"].Available {
		t.Error("claude-sonnet: got unavailable")
	}
}

func TestHandleModels(t *testing.T) {
	models := []adapter.ModelInfo{
		{ID: "mock", Name: "Mock", Provider: "mock"},
		{ID: "qwen2.5:1.5b", Name: "Qwen 2.5 1.5B", Provider: "ollama"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	w := httptest.NewRecorder()

	Models(models).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp []adapter.ModelInfo
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("models count: got %d, want 2", len(resp))
	}
	if resp[0].ID != "mock" {
		t.Errorf("first model id: got %q, want %q", resp[0].ID, "mock")
	}
}

func TestHandlePolish(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{
		"mock": &adapter.MockAdapter{},
	}

	tests := []struct {
		name      string
		method    string
		body      any
		wantCode  int
		wantField string
		wantValue string
	}{
		{
			name:      "success",
			method:    http.MethodPost,
			body:      polishRequest{Text: "hello world", ModelID: "mock"},
			wantCode:  http.StatusOK,
			wantField: "polished",
			wantValue: "Hello world",
		},
		{
			name:      "wrong method",
			method:    http.MethodGet,
			body:      nil,
			wantCode:  http.StatusMethodNotAllowed,
			wantField: "error",
			wantValue: "method not allowed",
		},
		{
			name:      "empty text",
			method:    http.MethodPost,
			body:      polishRequest{Text: "", ModelID: "mock"},
			wantCode:  http.StatusBadRequest,
			wantField: "error",
			wantValue: "text is required",
		},
		{
			name:      "empty model_id",
			method:    http.MethodPost,
			body:      polishRequest{Text: "hello", ModelID: ""},
			wantCode:  http.StatusBadRequest,
			wantField: "error",
			wantValue: "model_id is required",
		},
		{
			name:      "unknown model",
			method:    http.MethodPost,
			body:      polishRequest{Text: "hello", ModelID: "nonexistent"},
			wantCode:  http.StatusBadRequest,
			wantField: "error",
			wantValue: "unknown model: nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body != nil {
				var err error
				bodyBytes, err = json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("marshal: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/polish", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			Polish(adapters, "system prompt").ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantCode)
			}

			var resp map[string]any
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}

			got, ok := resp[tt.wantField]
			if !ok {
				t.Fatalf("response missing field %q: %v", tt.wantField, resp)
			}
			if got != tt.wantValue {
				t.Errorf("%s: got %q, want %q", tt.wantField, got, tt.wantValue)
			}
		})
	}
}

func TestHandlePolishTextTooLong(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{"mock": &adapter.MockAdapter{}}

	t.Run("over limit", func(t *testing.T) {
		longText := strings.Repeat("a", maxTextLength+1)
		body, _ := json.Marshal(polishRequest{Text: longText, ModelID: "mock"})
		req := httptest.NewRequest(http.MethodPost, "/api/polish", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		Polish(adapters, "prompt").ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
		}
		var resp errorResponse
		json.NewDecoder(w.Body).Decode(&resp)
		if !strings.Contains(resp.Error, "too long") {
			t.Errorf("error: got %q, want to contain 'too long'", resp.Error)
		}
	})

	t.Run("at limit", func(t *testing.T) {
		exactText := strings.Repeat("b", maxTextLength)
		body, _ := json.Marshal(polishRequest{Text: exactText, ModelID: "mock"})
		req := httptest.NewRequest(http.MethodPost, "/api/polish", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		Polish(adapters, "prompt").ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestHandlePolishInvalidJSON(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{"mock": &adapter.MockAdapter{}}

	req := httptest.NewRequest(http.MethodPost, "/api/polish", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Polish(adapters, "prompt").ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePolishResponseHasElapsedMs(t *testing.T) {
	adapters := map[string]adapter.LLMAdapter{"mock": &adapter.MockAdapter{}}

	body, _ := json.Marshal(polishRequest{Text: "hello", ModelID: "mock"})
	req := httptest.NewRequest(http.MethodPost, "/api/polish", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	Polish(adapters, "prompt").ServeHTTP(w, req)

	var resp polishResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Model != "mock" {
		t.Errorf("model: got %q, want %q", resp.Model, "mock")
	}
	if resp.ElapsedMs < 0 {
		t.Errorf("elapsed_ms should be >= 0, got %d", resp.ElapsedMs)
	}
}
