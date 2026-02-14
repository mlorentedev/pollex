package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
)

// APIKey returns middleware that requires a valid X-API-Key header.
// If expectedKey is empty, the middleware is a no-op (backward compatible).
// /api/health is exempt so monitoring works without credentials.
func APIKey(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			if r.URL.Path == "/api/health" {
				next.ServeHTTP(w, r)
				return
			}

			provided := r.Header.Get("X-API-Key")
			if provided == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "missing API key"})
				return
			}

			if subtle.ConstantTimeCompare([]byte(provided), []byte(expectedKey)) != 1 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid API key"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
