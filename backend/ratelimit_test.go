package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterUnderLimit(t *testing.T) {
	rl := newRateLimiter(5, time.Minute)
	for i := 0; i < 5; i++ {
		if !rl.allow("192.168.1.1") {
			t.Fatalf("request %d: should be allowed", i+1)
		}
	}
}

func TestRateLimiterOverLimit(t *testing.T) {
	rl := newRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		rl.allow("192.168.1.1")
	}
	if rl.allow("192.168.1.1") {
		t.Error("4th request should be denied")
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	rl := newRateLimiter(2, time.Minute)

	// Exhaust limit for IP1
	rl.allow("10.0.0.1")
	rl.allow("10.0.0.1")
	if rl.allow("10.0.0.1") {
		t.Error("IP1: 3rd request should be denied")
	}

	// IP2 should still be allowed
	if !rl.allow("10.0.0.2") {
		t.Error("IP2: first request should be allowed")
	}
}

func TestRateLimiterWindowExpiry(t *testing.T) {
	rl := newRateLimiter(2, 50*time.Millisecond)

	rl.allow("192.168.1.1")
	rl.allow("192.168.1.1")
	if rl.allow("192.168.1.1") {
		t.Error("should be denied before window expires")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.allow("192.168.1.1") {
		t.Error("should be allowed after window expires")
	}
}

func TestRateLimitMiddleware429(t *testing.T) {
	rl := newRateLimiter(1, time.Minute)
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rateLimitMiddleware(rl)(inner)

	// First request — allowed
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("first request: got %d, want %d", w.Code, http.StatusOK)
	}

	// Second request — denied
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:12346"
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("second request: got %d, want %d", w.Code, http.StatusTooManyRequests)
	}

	var resp errorResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != "rate limit exceeded" {
		t.Errorf("error: got %q, want %q", resp.Error, "rate limit exceeded")
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{"ip:port", "192.168.1.1:12345", "192.168.1.1"},
		{"ip only", "192.168.1.1", "192.168.1.1"},
		{"ipv6", "[::1]:8080", "::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{RemoteAddr: tt.remoteAddr}
			if got := clientIP(r); got != tt.want {
				t.Errorf("clientIP(%q): got %q, want %q", tt.remoteAddr, got, tt.want)
			}
		})
	}
}
