package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()

	handleHealth().ServeHTTP(w, req)

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
}

func TestHandleModels(t *testing.T) {
	models := []ModelInfo{
		{ID: "mock", Name: "Mock", Provider: "mock"},
		{ID: "qwen2.5:1.5b", Name: "Qwen 2.5 1.5B", Provider: "ollama"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	w := httptest.NewRecorder()

	handleModels(models).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp []ModelInfo
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
	adapters := map[string]LLMAdapter{
		"mock": &MockAdapter{},
	}

	tests := []struct {
		name      string
		method    string
		body      any
		wantCode  int
		wantField string // field to check in response
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

			handlePolish(adapters, "system prompt").ServeHTTP(w, req)

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

func TestHandlePolishInvalidJSON(t *testing.T) {
	adapters := map[string]LLMAdapter{"mock": &MockAdapter{}}

	req := httptest.NewRequest(http.MethodPost, "/api/polish", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlePolish(adapters, "prompt").ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandlePolishResponseHasElapsedMs(t *testing.T) {
	adapters := map[string]LLMAdapter{"mock": &MockAdapter{}}

	body, _ := json.Marshal(polishRequest{Text: "hello", ModelID: "mock"})
	req := httptest.NewRequest(http.MethodPost, "/api/polish", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlePolish(adapters, "prompt").ServeHTTP(w, req)

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
