package main

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	Port         int    `yaml:"port"`
	OllamaURL    string `yaml:"ollama_url"`
	ClaudeAPIKey string `yaml:"claude_api_key"`
	ClaudeModel  string `yaml:"claude_model"`
	PromptPath   string `yaml:"prompt_path"`
}

func configDefaults() Config {
	return Config{
		Port:        8090,
		OllamaURL:   "http://localhost:11434",
		ClaudeModel: "claude-sonnet-4-5-20250929",
		PromptPath:  "../prompts/polish.txt",
	}
}

// LoadConfig loads configuration from a YAML file (if path is non-empty),
// then applies environment variable overrides. An empty path returns defaults + env overrides.
func LoadConfig(path string) (Config, error) {
	cfg := configDefaults()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return Config{}, fmt.Errorf("config: read file: %w", err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("config: parse yaml: %w", err)
		}
	}

	// Environment variable overrides
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
	if v := os.Getenv("POLLEX_PROMPT_PATH"); v != "" {
		cfg.PromptPath = v
	}

	return cfg, nil
}
