package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port          int    `yaml:"port"`
	OllamaURL     string `yaml:"ollama_url"`
	ClaudeAPIKey  string `yaml:"claude_api_key"`
	ClaudeModel   string `yaml:"claude_model"`
	LlamaCppURL   string `yaml:"llamacpp_url"`
	LlamaCppModel string `yaml:"llamacpp_model"`
	// NaN (nan.builders) cloud engine — an OpenAI-compatible gateway serving the
	// Nous-family models. NanModels is the ordered fallback chain tried in turn.
	NanAPIKey  string   `yaml:"nan_api_key"`
	NanBaseURL string   `yaml:"nan_base_url"`
	NanModels  []string `yaml:"nan_models"`
	PromptPath string   `yaml:"prompt_path"`
	APIKey     string   `yaml:"api_key"`
}

func defaults() Config {
	return Config{
		Port:        8090,
		ClaudeModel: "claude-sonnet-4-5-20250929",
		NanModels:   []string{"mimo-v2.5", "qwen3.6", "gemma4"},
		PromptPath:  "prompts/polish.txt",
	}
}

// splitAndTrim parses a comma-separated env value into a clean slice.
func splitAndTrim(csv string) []string {
	parts := strings.Split(csv, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// Load reads optional YAML then applies POLLEX_* env var overrides.
func Load(path string) (Config, error) {
	cfg := defaults()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("config: read file: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("config: parse yaml: %w", err)
		}
	}

	if v := os.Getenv("POLLEX_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("config: invalid POLLEX_PORT %q: %w", v, err)
		}
		cfg.Port = p
	}
	if v := os.Getenv("POLLEX_OLLAMA_URL"); v != "" {
		cfg.OllamaURL = v
	}
	if v := os.Getenv("POLLEX_CLAUDE_API_KEY"); v != "" {
		cfg.ClaudeAPIKey = v
	}
	if v := os.Getenv("POLLEX_CLAUDE_MODEL"); v != "" {
		cfg.ClaudeModel = v
	}
	if v := os.Getenv("POLLEX_LLAMACPP_URL"); v != "" {
		cfg.LlamaCppURL = v
	}
	if v := os.Getenv("POLLEX_LLAMACPP_MODEL"); v != "" {
		cfg.LlamaCppModel = v
	}
	if v := os.Getenv("POLLEX_NAN_API_KEY"); v != "" {
		cfg.NanAPIKey = v
	}
	if v := os.Getenv("POLLEX_NAN_BASE_URL"); v != "" {
		cfg.NanBaseURL = v
	}
	if v := os.Getenv("POLLEX_NAN_MODELS"); v != "" {
		cfg.NanModels = splitAndTrim(v)
	}
	if v := os.Getenv("POLLEX_PROMPT_PATH"); v != "" {
		cfg.PromptPath = v
	}
	if v := os.Getenv("POLLEX_API_KEY"); v != "" {
		cfg.APIKey = v
	}

	return cfg, nil
}
