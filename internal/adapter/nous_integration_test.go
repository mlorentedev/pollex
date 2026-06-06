//go:build integration

package adapter

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// transientGatewayError reports whether err is the gateway being flaky (timeout,
// network, 5xx, 429) rather than a contract violation. nan.builders is a
// best-effort community gateway with no SLA, so an integration test must skip on
// availability hiccups and fail only on real regressions (404 model removed,
// 4xx contract errors, or a 200 with bad output).
func transientGatewayError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}
	var se *StatusError
	if errors.As(err, &se) {
		return se.Code == http.StatusTooManyRequests || se.Code >= 500
	}
	return true // network / decode failure
}

// TestNousIntegrationModels exercises every model in the NaN fallback chain
// against the live gateway, proving each one polishes text on its own
// (acceptance criterion AC3). Run with:
//
//	NAN_API_KEY=... go test -tags integration -run TestNousIntegrationModels ./internal/adapter/
//
// Excluded from the default build by the `integration` tag; skips if the key is
// absent or the gateway is transiently unavailable.
func TestNousIntegrationModels(t *testing.T) {
	key := os.Getenv("NAN_API_KEY")
	if key == "" {
		t.Skip("NAN_API_KEY not set; skipping live NaN integration test")
	}
	baseURL := os.Getenv("NAN_BASE_URL") // empty -> adapter default; /v1 suffix tolerated

	const systemPrompt = "You are an English text polisher. Return only the corrected text, no commentary."
	const input = "i has went to the store yesterday and buyed two breads."

	for _, model := range []string{"mimo-v2.5", "qwen3.6", "gemma4"} {
		t.Run(model, func(t *testing.T) {
			a := &NousAdapter{
				BaseURL: baseURL,
				APIKey:  key,
				Model:   model,
				Client:  &http.Client{Timeout: 30 * time.Second},
			}

			out, err := a.Polish(context.Background(), input, systemPrompt)
			if err != nil {
				if transientGatewayError(err) {
					t.Skipf("%s: transient gateway failure (best-effort, no SLA): %v", model, err)
				}
				t.Fatalf("%s: Polish failed (contract error): %v", model, err)
			}
			if strings.TrimSpace(out) == "" {
				t.Fatalf("%s: empty polish output", model)
			}
			if strings.EqualFold(strings.TrimSpace(out), input) {
				t.Errorf("%s: output unchanged from the (deliberately buggy) input: %q", model, out)
			}
			t.Logf("%s -> %q", model, out)
		})
	}
}
