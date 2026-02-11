package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig with no file: %v", err)
	}

	if cfg.Port != 8090 {
		t.Errorf("default port: got %d, want 8090", cfg.Port)
	}
	if cfg.OllamaURL != "http://localhost:11434" {
		t.Errorf("default ollama_url: got %q, want %q", cfg.OllamaURL, "http://localhost:11434")
	}
	if cfg.PromptPath != "../prompts/polish.txt" {
		t.Errorf("default prompt_path: got %q, want %q", cfg.PromptPath, "../prompts/polish.txt")
	}
	if cfg.ClaudeAPIKey != "" {
		t.Errorf("default claude_api_key: got %q, want empty", cfg.ClaudeAPIKey)
	}
	if cfg.ClaudeModel != "claude-sonnet-4-5-20250929" {
		t.Errorf("default claude_model: got %q, want %q", cfg.ClaudeModel, "claude-sonnet-4-5-20250929")
	}
}

func TestLoadConfigFromYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	content := `port: 9999
ollama_url: "http://jetson.local:11434"
claude_api_key: "sk-test-key"
claude_model: "claude-opus-4-6"
prompt_path: "/etc/pollex/polish.txt"
`
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	cfg, err := LoadConfig(yamlPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"port", cfg.Port, 9999},
		{"ollama_url", cfg.OllamaURL, "http://jetson.local:11434"},
		{"claude_api_key", cfg.ClaudeAPIKey, "sk-test-key"},
		{"claude_model", cfg.ClaudeModel, "claude-opus-4-6"},
		{"prompt_path", cfg.PromptPath, "/etc/pollex/polish.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoadConfigEnvOverrides(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	content := `port: 9999
ollama_url: "http://from-yaml:11434"
`
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	t.Setenv("POLLEX_PORT", "7777")
	t.Setenv("POLLEX_OLLAMA_URL", "http://from-env:11434")
	t.Setenv("POLLEX_CLAUDE_API_KEY", "sk-env-key")

	cfg, err := LoadConfig(yamlPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"port from env", cfg.Port, 7777},
		{"ollama_url from env", cfg.OllamaURL, "http://from-env:11434"},
		{"claude_api_key from env", cfg.ClaudeAPIKey, "sk-env-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(yamlPath, []byte("{{invalid"), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	_, err := LoadConfig(yamlPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
