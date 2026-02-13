package middleware

import (
	"net/http"
	"time"
)

// Chain wraps the handler with the full middleware stack.
func Chain(handler http.Handler, rl *RateLimiter) http.Handler {
	h := handler
	h = http.TimeoutHandler(h, 65*time.Second, `{"error":"request timeout"}`)
	h = MaxBytes(64 * 1024)(h)
	h = RateLimit(rl)(h)
	h = Logging(h)
	h = RequestID(h)
	h = CORS(h)
	return h
}
