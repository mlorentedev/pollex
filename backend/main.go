package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

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

	// Build adapter registry
	adapters := make(map[string]LLMAdapter)
	var models []ModelInfo

	if *useMock {
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

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth())
	mux.HandleFunc("/api/models", handleModels(models))
	mux.HandleFunc("/api/polish", handlePolish(adapters, systemPrompt))

	// Middleware stack: CORS → logging → timeout → mux
	handler := corsMiddleware(loggingMiddleware(
		http.TimeoutHandler(mux, 65*time.Second, `{"error":"request timeout"}`),
	))

	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("pollex api listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
