package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// nousDefaultBaseURL is the NaN community gateway (nan.builders), an
// OpenAI-compatible inference API. The "/v1" path segment is appended per call,
// mirroring LlamaCppAdapter, so BaseURL is the host root.
const nousDefaultBaseURL = "https://api.nan.builders"

// nousPolishTemperature keeps polishing low-variance without being fully rigid
// (empirically good across mimo-v2.5 / qwen3.6 / gemma4 on the NaN gateway).
const nousPolishTemperature = 0.3

// nousMaxTokens caps generation; polishing output is roughly input-sized and the
// handler already bounds input length.
const nousMaxTokens = 4096

// NousAdapter connects to the NaN (nan.builders) OpenAI-compatible
// /v1/chat/completions endpoint with Bearer auth. It reads the answer from
// message.content and deliberately ignores the non-standard reasoning_content
// field; it sends enable_thinking:false to suppress reasoning latency and the
// silent monthly token quota on reasoning models.
type NousAdapter struct {
	BaseURL string // host root; defaults to nousDefaultBaseURL
	APIKey  string
	Model   string
	Client  *http.Client
}

type nousMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type nousChatRequest struct {
	Model              string         `json:"model"`
	Messages           []nousMessage  `json:"messages"`
	Temperature        float32        `json:"temperature"`
	MaxTokens          int            `json:"max_tokens,omitempty"`
	ChatTemplateKwargs map[string]any `json:"chat_template_kwargs,omitempty"`
}

// nousMessageResp carries both the answer (Content) and, for reasoning models,
// the non-standard ReasoningContent chain — which we never surface to the user.
type nousMessageResp struct {
	Role             string `json:"role"`
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}

type nousChoice struct {
	Message nousMessageResp `json:"message"`
}

type nousChatResponse struct {
	Choices []nousChoice `json:"choices"`
}

type nousErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// StatusError is returned when the gateway responds with a non-200 status. It
// exposes the HTTP code so a FallbackChain can decide whether another model
// could recover (availability/quota) or whether the request is hopeless (4xx).
type StatusError struct {
	Code    int
	Message string
}

func (e *StatusError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("nous: API error %d: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("nous: unexpected status %d", e.Code)
}

func (n *NousAdapter) Name() string {
	return fmt.Sprintf("Nous (%s)", n.Model)
}

func (n *NousAdapter) Polish(ctx context.Context, text, systemPrompt string) (string, error) {
	reqBody := nousChatRequest{
		Model: n.Model,
		Messages: []nousMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: text},
		},
		Temperature:        nousPolishTemperature,
		MaxTokens:          nousMaxTokens,
		ChatTemplateKwargs: map[string]any{"enable_thinking": false},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("nous: marshal request: %w", err)
	}

	baseURL := n.BaseURL
	if baseURL == "" {
		baseURL = nousDefaultBaseURL
	}
	// Accept the base URL either as the host root or with a trailing "/v1" (the
	// ecosystem's NAN_BASE_URL convention), so we never produce "/v1/v1/...".
	baseURL = strings.TrimSuffix(strings.TrimRight(baseURL, "/"), "/v1")
	url := baseURL + "/v1/chat/completions"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("nous: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+n.APIKey)

	resp, err := n.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("nous: request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp nousErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return "", &StatusError{Code: resp.StatusCode, Message: errResp.Error.Message}
	}

	var chatResp nousChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("nous: decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("nous: empty response choices")
	}

	// Only the answer — reasoning_content is intentionally discarded. A blank
	// answer is treated as a failure so the FallbackChain advances to the next
	// model rather than returning an empty polish.
	out := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	if out == "" {
		return "", fmt.Errorf("nous: empty content in response")
	}
	return out, nil
}

func (n *NousAdapter) Available() bool {
	return n.APIKey != ""
}
