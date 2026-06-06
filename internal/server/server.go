package server

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mlorentedev/pollex/internal/adapter"
	"github.com/mlorentedev/pollex/internal/handler"
	"github.com/mlorentedev/pollex/internal/middleware"
)

// SetupMux wires handlers with the full middleware chain.
// promptSelector maps a model ID to its system prompt, allowing local and cloud adapters
// to use different prompts without changing the adapter interface.
func SetupMux(adapters map[string]adapter.LLMAdapter, models []adapter.ModelInfo, promptSelector func(string) string, apiKey, version string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handler.Health(adapters, version))
	mux.HandleFunc("/api/models", handler.Models(models))
	mux.HandleFunc("/api/polish", handler.Polish(adapters, promptSelector))
	mux.Handle("/metrics", promhttp.Handler())

	rl := middleware.NewRateLimiter(10, time.Minute)
	return middleware.Chain(mux, rl, apiKey)
}
