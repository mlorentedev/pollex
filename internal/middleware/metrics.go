package middleware

import (
	"net/http"
	"strconv"

	"github.com/mlorentedev/pollex/internal/metrics"
)

// Metrics records request count by method, path, and status code.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		metrics.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(sw.status)).Inc()
	})
}
