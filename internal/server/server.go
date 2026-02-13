package server

import (
	"net/http"
	"time"

	"github.com/mlorentedev/pollex/internal/adapter"
	"github.com/mlorentedev/pollex/internal/handler"
	"github.com/mlorentedev/pollex/internal/middleware"
)

// SetupMux wires handlers with the full middleware chain.
func SetupMux(adapters map[string]adapter.LLMAdapter, models []adapter.ModelInfo, systemPrompt string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handler.Health(adapters))
	mux.HandleFunc("/api/models", handler.Models(models))
	mux.HandleFunc("/api/polish", handler.Polish(adapters, systemPrompt))

	rl := middleware.NewRateLimiter(10, time.Minute)
	return middleware.Chain(mux, rl)
}
