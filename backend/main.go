package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func buildAdapters(cfg Config, useMock bool) (map[string]LLMAdapter, []ModelInfo) {
	adapters := make(map[string]LLMAdapter)
	var models []ModelInfo

	if useMock {
		adapters["mock"] = &MockAdapter{Delay: 500 * time.Millisecond}
		models = append(models, ModelInfo{ID: "mock", Name: "Mock (dev)", Provider: "mock"})
		log.Println("mode: mock adapter enabled")
	} else {
		ollama := &OllamaAdapter{
			BaseURL: cfg.OllamaURL,
			Model:   "qwen2.5:1.5b",
			Client:  &http.Client{Timeout: 60 * time.Second},
		}
		adapters["qwen2.5:1.5b"] = ollama
		models = append(models, ModelInfo{ID: "qwen2.5:1.5b", Name: "Qwen 2.5 1.5B", Provider: "ollama"})
		log.Printf("mode: ollama at %s", cfg.OllamaURL)
	}

	// Claude adapter (optional, when API key is configured)
	if cfg.ClaudeAPIKey != "" {
		claude := &ClaudeAdapter{
			APIKey: cfg.ClaudeAPIKey,
			Model:  cfg.ClaudeModel,
			Client: &http.Client{Timeout: 60 * time.Second},
		}
		adapters[cfg.ClaudeModel] = claude
		models = append(models, ModelInfo{ID: cfg.ClaudeModel, Name: "Claude (" + cfg.ClaudeModel + ")", Provider: "claude"})
		log.Printf("mode: claude enabled (model: %s)", cfg.ClaudeModel)
	}

	return adapters, models
}

func setupMux(adapters map[string]LLMAdapter, models []ModelInfo, systemPrompt string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth(adapters))
	mux.HandleFunc("/api/models", handleModels(models))
	mux.HandleFunc("/api/polish", handlePolish(adapters, systemPrompt))

	rl := newRateLimiter(10, time.Minute)

	// Middleware stack (outermost → innermost):
	// CORS → requestID → logging → rateLimit → maxBytes → timeout → mux
	var handler http.Handler = mux
	handler = http.TimeoutHandler(handler, 65*time.Second, `{"error":"request timeout"}`)
	handler = maxBytesMiddleware(64 * 1024)(handler)
	handler = rateLimitMiddleware(rl)(handler)
	handler = loggingMiddleware(handler)
	handler = requestIDMiddleware(handler)
	handler = corsMiddleware(handler)

	return handler
}

func main() {
	configPath := flag.String("config", "", "path to config.yaml")
	useMock := flag.Bool("mock", false, "use mock adapter instead of real LLM backends")
	port := flag.Int("port", 0, "override listen port")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if *port > 0 {
		cfg.Port = *port
	}

	// Load system prompt
	promptData, err := os.ReadFile(cfg.PromptPath)
	if err != nil {
		log.Fatalf("prompt: read %s: %v", cfg.PromptPath, err)
	}
	systemPrompt := string(promptData)

	adapters, models := buildAdapters(cfg, *useMock)
	handler := setupMux(adapters, models, systemPrompt)

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("pollex api listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-done
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("server stopped")
}
