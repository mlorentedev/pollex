package middleware

import (
	"testing"
	"time"
)

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
