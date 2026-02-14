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

	"github.com/mlorentedev/pollex/internal/adapter"
	"github.com/mlorentedev/pollex/internal/config"
	"github.com/mlorentedev/pollex/internal/server"
)

func main() {
	configPath := flag.String("config", "", "path to config.yaml")
	useMock := flag.Bool("mock", false, "use mock adapter instead of real LLM backends")
	port := flag.Int("port", 0, "override listen port")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if *port > 0 {
		cfg.Port = *port
	}

	promptData, err := os.ReadFile(cfg.PromptPath)
	if err != nil {
		log.Fatalf("prompt: read %s: %v", cfg.PromptPath, err)
	}
	systemPrompt := string(promptData)

	adapters, models := buildAdapters(cfg, *useMock)
	handler := server.SetupMux(adapters, models, systemPrompt, cfg.APIKey)

	if cfg.APIKey != "" {
		log.Println("auth: API key required (X-API-Key header)")
	} else {
		log.Println("auth: disabled (no api_key configured)")
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

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

func buildAdapters(cfg config.Config, useMock bool) (map[string]adapter.LLMAdapter, []adapter.ModelInfo) {
	adapters := make(map[string]adapter.LLMAdapter)
	var models []adapter.ModelInfo

	if useMock {
		adapters["mock"] = &adapter.MockAdapter{Delay: 500 * time.Millisecond}
		models = append(models, adapter.ModelInfo{ID: "mock", Name: "Mock (dev)", Provider: "mock"})
		log.Println("mode: mock adapter enabled")
		return adapters, models
	}

	// 1. llama.cpp (Highest priority for local GPU)
	if cfg.LlamaCppURL != "" {
		model := cfg.LlamaCppModel
		if model == "" {
			model = "qwen2.5-1.5b-gpu"
		}
		llama := &adapter.LlamaCppAdapter{
			BaseURL: cfg.LlamaCppURL,
			Model:   model,
			Client:  &http.Client{Timeout: 120 * time.Second},
		}
		adapters[model] = llama
		models = append(models, adapter.ModelInfo{ID: model, Name: "llama.cpp (" + model + ")", Provider: "llamacpp"})
		log.Printf("mode: llama.cpp at %s (model: %s)", cfg.LlamaCppURL, model)
	}

	// 2. Claude (Optional cloud fallback)
	if cfg.ClaudeAPIKey != "" {
		claude := &adapter.ClaudeAdapter{
			APIKey: cfg.ClaudeAPIKey,
			Model:  cfg.ClaudeModel,
			Client: &http.Client{Timeout: 60 * time.Second},
		}
		adapters[cfg.ClaudeModel] = claude
		models = append(models, adapter.ModelInfo{ID: cfg.ClaudeModel, Name: "Claude (" + cfg.ClaudeModel + ")", Provider: "claude"})
		log.Printf("mode: claude enabled (model: %s)", cfg.ClaudeModel)
	}

	// 3. Ollama (Optional or legacy fallback)
	if cfg.OllamaURL != "" {
		model := "qwen2.5:1.5b"
		ollama := &adapter.OllamaAdapter{
			BaseURL: cfg.OllamaURL,
			Model:   model,
			Client:  &http.Client{Timeout: 60 * time.Second},
		}
		adapters[model] = ollama
		models = append(models, adapter.ModelInfo{ID: model, Name: "Qwen 2.5 1.5B", Provider: "ollama"})
		log.Printf("mode: ollama at %s", cfg.OllamaURL)
	}

	return adapters, models
}
