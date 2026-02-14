package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientIPFromCfHeader(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rl := NewRateLimiter(1, time.Minute)
	handler := RateLimit(rl)(inner)

	t.Run("uses Cf-Connecting-Ip when present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Cf-Connecting-Ip", "1.2.3.4")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("different Cf IPs get separate buckets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Cf-Connecting-Ip", "5.6.7.8")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("same Cf IP hits rate limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Cf-Connecting-Ip", "1.2.3.4")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusTooManyRequests {
			t.Errorf("status: got %d, want %d", w.Code, http.StatusTooManyRequests)
		}
	})
}

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !rl.Allow("127.0.0.1") {
			t.Errorf("request %d should be allowed", i)
		}
	}

	if rl.Allow("127.0.0.1") {
		t.Error("4th request should be denied")
	}
}

func TestRateLimiterDifferentKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	if !rl.Allow("10.0.0.1") {
		t.Error("first IP should be allowed")
	}
	if !rl.Allow("10.0.0.2") {
		t.Error("second IP should be allowed")
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)

	if !rl.Allow("127.0.0.1") {
		t.Error("first request should be allowed")
	}
	if rl.Allow("127.0.0.1") {
		t.Error("second request should be denied")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("127.0.0.1") {
		t.Error("request after window should be allowed")
	}
}
