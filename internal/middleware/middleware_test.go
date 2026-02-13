package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := CORS(inner)

	t.Run("adds CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("Allow-Origin: got %q, want %q", got, "*")
		}
		if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
			t.Errorf("Allow-Methods: got %q, want %q", got, "GET, POST, OPTIONS")
		}
		if w.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("OPTIONS preflight returns 204", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusNoContent)
		}
	})
}

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("sets X-Request-ID header", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(inner)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		id := w.Header().Get("X-Request-ID")
		if id == "" {
			t.Error("X-Request-ID header not set")
		}
		if len(id) != 32 {
			t.Errorf("X-Request-ID length: got %d, want 32", len(id))
		}
	})

	t.Run("stores request ID in context", func(t *testing.T) {
		var gotID string
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotID = RequestIDFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		handler := RequestID(inner)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if gotID == "" {
			t.Error("request ID not stored in context")
		}
		headerID := w.Header().Get("X-Request-ID")
		if gotID != headerID {
			t.Errorf("context ID %q != header ID %q", gotID, headerID)
		}
	})
}

func TestLoggingMiddleware(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := Logging(inner)

	req := httptest.NewRequest(http.MethodPost, "/api/polish", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestStatusWriterCapturesStatus(t *testing.T) {
	w := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
	sw.WriteHeader(http.StatusNotFound)

	if sw.status != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", sw.status, http.StatusNotFound)
	}
}

func TestMaxBytesMiddleware(t *testing.T) {
	t.Run("allows small body", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		handler := MaxBytes(1024)(inner)
		body := strings.NewReader("small body")
		req := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("rejects oversized body", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "body too large", http.StatusRequestEntityTooLarge)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		handler := MaxBytes(10)(inner)
		body := strings.NewReader(strings.Repeat("x", 100))
		req := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusRequestEntityTooLarge)
		}
	})
}
