package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LlamaCppAdapter connects to llama-server's OpenAI-compatible /v1/chat/completions.
type LlamaCppAdapter struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

type llamaCppMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type llamaCppChatRequest struct {
	Model    string            `json:"model"`
	Messages []llamaCppMessage `json:"messages"`
}

type llamaCppChoice struct {
	Message llamaCppMessage `json:"message"`
}

type llamaCppChatResponse struct {
	Choices []llamaCppChoice `json:"choices"`
}

func (l *LlamaCppAdapter) Name() string {
	return fmt.Sprintf("llama.cpp (%s)", l.Model)
}

func (l *LlamaCppAdapter) Polish(ctx context.Context, text, systemPrompt string) (string, error) {
	reqBody := llamaCppChatRequest{
		Model: l.Model,
		Messages: []llamaCppMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: text},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("llamacpp: marshal request: %w", err)
	}

	url := strings.TrimRight(l.BaseURL, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("llamacpp: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := l.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llamacpp: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llamacpp: unexpected status %d", resp.StatusCode)
	}

	var chatResp llamaCppChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("llamacpp: decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("llamacpp: empty response choices")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

func (l *LlamaCppAdapter) Available() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(l.BaseURL, "/")+"/health", nil)
	if err != nil {
		return false
	}

	resp, err := l.Client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
