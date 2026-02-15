package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
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
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	configPath := flag.String("config", "", "path to config.yaml")
	useMock := flag.Bool("mock", false, "use mock adapter instead of real LLM backends")
	port := flag.Int("port", 0, "override listen port")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}
	if *port > 0 {
		cfg.Port = *port
	}

	promptData, err := os.ReadFile(cfg.PromptPath)
	if err != nil {
		slog.Error("prompt read failed", "path", cfg.PromptPath, "error", err)
		os.Exit(1)
	}
	systemPrompt := string(promptData)

	adapters, models := buildAdapters(cfg, *useMock)
	handler := server.SetupMux(adapters, models, systemPrompt, cfg.APIKey)

	if cfg.APIKey != "" {
		slog.Info("auth enabled", "mode", "X-API-Key header")
	} else {
		slog.Info("auth disabled", "reason", "no api_key configured")
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server stopped")
}

func buildAdapters(cfg config.Config, useMock bool) (map[string]adapter.LLMAdapter, []adapter.ModelInfo) {
	adapters := make(map[string]adapter.LLMAdapter)
	var models []adapter.ModelInfo

	if useMock {
		adapters["mock"] = &adapter.MockAdapter{Delay: 500 * time.Millisecond}
		models = append(models, adapter.ModelInfo{ID: "mock", Name: "Mock (dev)", Provider: "mock"})
		slog.Info("adapter registered", "adapter", "mock")
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
		slog.Info("adapter registered", "adapter", "llamacpp", "url", cfg.LlamaCppURL, "model", model)
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
		slog.Info("adapter registered", "adapter", "claude", "model", cfg.ClaudeModel)
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
		slog.Info("adapter registered", "adapter", "ollama", "url", cfg.OllamaURL)
	}

	return adapters, models
}
