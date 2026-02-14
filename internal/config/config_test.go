package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("POLLEX_API_KEY", "")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load with no file: %v", err)
	}

	if cfg.Port != 8090 {
		t.Errorf("default port: got %d, want 8090", cfg.Port)
	}
	if cfg.OllamaURL != "" {
		t.Errorf("default ollama_url: got %q, want empty", cfg.OllamaURL)
	}
	if cfg.PromptPath != "prompts/polish.txt" {
		t.Errorf("default prompt_path: got %q, want %q", cfg.PromptPath, "prompts/polish.txt")
	}
	if cfg.ClaudeAPIKey != "" {
		t.Errorf("default claude_api_key: got %q, want empty", cfg.ClaudeAPIKey)
	}
	if cfg.ClaudeModel != "claude-sonnet-4-5-20250929" {
		t.Errorf("default claude_model: got %q, want %q", cfg.ClaudeModel, "claude-sonnet-4-5-20250929")
	}
	if cfg.LlamaCppURL != "" {
		t.Errorf("default llamacpp_url: got %q, want empty", cfg.LlamaCppURL)
	}
	if cfg.LlamaCppModel != "" {
		t.Errorf("default llamacpp_model: got %q, want empty", cfg.LlamaCppModel)
	}
	if cfg.APIKey != "" {
		t.Errorf("default api_key: got %q, want empty", cfg.APIKey)
	}
}

func TestLoadFromYAML(t *testing.T) {
	t.Setenv("POLLEX_API_KEY", "")

	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	content := `port: 9999
ollama_url: "http://jetson.local:11434"
claude_api_key: "sk-test-key"
claude_model: "claude-opus-4-6"
llamacpp_url: "http://localhost:8080"
llamacpp_model: "qwen2.5-1.5b"
prompt_path: "/etc/pollex/polish.txt"
api_key: "my-secret-key"
`
	if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	cfg, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
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
		{"llamacpp_url", cfg.LlamaCppURL, "http://localhost:8080"},
		{"llamacpp_model", cfg.LlamaCppModel, "qwen2.5-1.5b"},
		{"api_key", cfg.APIKey, "my-secret-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoadEnvOverrides(t *testing.T) {
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
	t.Setenv("POLLEX_LLAMACPP_URL", "http://from-env:8080")
	t.Setenv("POLLEX_LLAMACPP_MODEL", "custom-model")
	t.Setenv("POLLEX_API_KEY", "env-api-key")

	cfg, err := Load(yamlPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"port from env", cfg.Port, 7777},
		{"ollama_url from env", cfg.OllamaURL, "http://from-env:11434"},
		{"claude_api_key from env", cfg.ClaudeAPIKey, "sk-env-key"},
		{"llamacpp_url from env", cfg.LlamaCppURL, "http://from-env:8080"},
		{"llamacpp_model from env", cfg.LlamaCppModel, "custom-model"},
		{"api_key from env", cfg.APIKey, "env-api-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(yamlPath, []byte("{{invalid"), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	_, err := Load(yamlPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}
