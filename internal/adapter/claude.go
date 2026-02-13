package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const claudeDefaultBaseURL = "https://api.anthropic.com"

// ClaudeAdapter connects to the Anthropic Messages API.
type ClaudeAdapter struct {
	BaseURL string
	APIKey  string
	Model   string
	Client  *http.Client
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeMessagesRequest struct {
	Model     string          `json:"model"`
	System    string          `json:"system"`
	Messages  []claudeMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

type claudeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type claudeMessagesResponse struct {
	Content []claudeContentBlock `json:"content"`
}

type claudeErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *ClaudeAdapter) Name() string {
	return fmt.Sprintf("Claude (%s)", c.Model)
}

func (c *ClaudeAdapter) Polish(ctx context.Context, text, systemPrompt string) (string, error) {
	reqBody := claudeMessagesRequest{
		Model:  c.Model,
		System: systemPrompt,
		Messages: []claudeMessage{
			{Role: "user", Content: text},
		},
		MaxTokens: 4096,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("claude: marshal request: %w", err)
	}

	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = claudeDefaultBaseURL
	}
	url := strings.TrimRight(baseURL, "/") + "/v1/messages"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("claude: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp claudeErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil || errResp.Error.Message == "" {
			return "", fmt.Errorf("claude: unexpected status %d", resp.StatusCode)
		}
		return "", fmt.Errorf("claude: API error: %s", errResp.Error.Message)
	}

	var msgResp claudeMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return "", fmt.Errorf("claude: decode response: %w", err)
	}

	if len(msgResp.Content) == 0 {
		return "", fmt.Errorf("claude: empty response content")
	}

	var result strings.Builder
	for _, block := range msgResp.Content {
		if block.Type == "text" {
			result.WriteString(block.Text)
		}
	}

	return strings.TrimSpace(result.String()), nil
}

func (c *ClaudeAdapter) Available() bool {
	return c.APIKey != ""
}
