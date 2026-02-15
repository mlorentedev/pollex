package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIKeyMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("disabled when key is empty", func(t *testing.T) {
		handler := APIKey("")(inner)
		req := httptest.NewRequest(http.MethodPost, "/api/polish", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("valid key passes", func(t *testing.T) {
		handler := APIKey("secret-123")(inner)
		req := httptest.NewRequest(http.MethodPost, "/api/polish", nil)
		req.Header.Set("X-API-Key", "secret-123")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("missing key returns 401", func(t *testing.T) {
		handler := APIKey("secret-123")(inner)
		req := httptest.NewRequest(http.MethodPost, "/api/polish", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusUnauthorized)
		}

		var body map[string]string
		json.NewDecoder(w.Body).Decode(&body)
		if body["error"] != "missing API key" {
			t.Errorf("error: got %q, want %q", body["error"], "missing API key")
		}
	})

	t.Run("wrong key returns 401", func(t *testing.T) {
		handler := APIKey("secret-123")(inner)
		req := httptest.NewRequest(http.MethodPost, "/api/polish", nil)
		req.Header.Set("X-API-Key", "wrong-key")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusUnauthorized)
		}

		var body map[string]string
		json.NewDecoder(w.Body).Decode(&body)
		if body["error"] != "invalid API key" {
			t.Errorf("error: got %q, want %q", body["error"], "invalid API key")
		}
	})

	t.Run("health endpoint exempt", func(t *testing.T) {
		handler := APIKey("secret-123")(inner)
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("metrics endpoint exempt", func(t *testing.T) {
		handler := APIKey("secret-123")(inner)
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("models endpoint requires auth", func(t *testing.T) {
		handler := APIKey("secret-123")(inner)
		req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}
