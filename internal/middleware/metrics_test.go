package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/mlorentedev/pollex/internal/metrics"
)

func TestMetricsMiddleware(t *testing.T) {
	t.Run("increments counter on 200", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := Metrics(inner)

		before := testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("GET", "/api/health", "200"))

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		after := testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("GET", "/api/health", "200"))
		if after != before+1 {
			t.Errorf("counter: got %f, want %f", after, before+1)
		}
	})

	t.Run("tracks different status codes", func(t *testing.T) {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		handler := Metrics(inner)

		before := testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("GET", "/missing", "404"))

		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		after := testutil.ToFloat64(metrics.RequestsTotal.WithLabelValues("GET", "/missing", "404"))
		if after != before+1 {
			t.Errorf("counter: got %f, want %f", after, before+1)
		}
	})
}
