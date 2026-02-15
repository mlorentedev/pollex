package middleware

import (
	"net/http"
	"time"
)

// Chain wraps the handler with the full middleware stack.
// Order: CORS → RequestID → Logging → Metrics → APIKey → RateLimit → MaxBytes → Timeout → mux
// APIKey runs before RateLimit so that: (1) invalid keys are rejected without
// consuming rate limit budget, and (2) authenticated requests skip rate limiting.
func Chain(handler http.Handler, rl *RateLimiter, apiKey string) http.Handler {
	h := handler
	h = http.TimeoutHandler(h, 120*time.Second, `{"error":"request timeout"}`)
	h = MaxBytes(64 * 1024)(h)
	h = RateLimit(rl)(h)
	h = APIKey(apiKey)(h)
	h = Metrics(h)
	h = Logging(h)
	h = RequestID(h)
	h = CORS(h)
	return h
}
