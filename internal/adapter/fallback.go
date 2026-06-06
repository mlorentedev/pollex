package adapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// FallbackChain is an LLMAdapter that wraps an ordered list of adapters. Polish
// tries each in turn and returns the first success; it advances to the next
// adapter only when shouldFallback reports the error is recoverable by another
// model (availability / quota), and fails fast otherwise. Because it satisfies
// LLMAdapter itself, it is registered like any single model.
type FallbackChain struct {
	Label    string // display name; if empty, composed from the children
	Adapters []LLMAdapter
}

func (c *FallbackChain) Name() string {
	if c.Label != "" {
		return c.Label
	}
	names := make([]string, len(c.Adapters))
	for i, a := range c.Adapters {
		names[i] = a.Name()
	}
	return "FallbackChain(" + strings.Join(names, " -> ") + ")"
}

func (c *FallbackChain) Polish(ctx context.Context, text, systemPrompt string) (string, error) {
	if len(c.Adapters) == 0 {
		return "", errors.New("fallback: no adapters configured")
	}
	var lastErr error
	for _, a := range c.Adapters {
		out, err := a.Polish(ctx, text, systemPrompt)
		if err == nil {
			return out, nil
		}
		lastErr = err
		if !shouldFallback(err) {
			return "", err // not recoverable by another model — fail fast
		}
	}
	return "", fmt.Errorf("fallback: all %d adapters failed: %w", len(c.Adapters), lastErr)
}

func (c *FallbackChain) Available() bool {
	for _, a := range c.Adapters {
		if a.Available() {
			return true
		}
	}
	return false
}

// shouldFallback reports whether an error from one adapter is worth retrying on
// the next adapter in the chain. It advances on availability/quota failures
// (network/timeout, HTTP 429, 404, 5xx) and stops on errors another model cannot
// fix (HTTP 400/401, which recur identically) or when the caller cancelled.
func shouldFallback(err error) bool {
	// The caller gave up (request cancelled or overall deadline hit) — trying
	// more models is pointless and would only burn the shared rate limit.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// The gateway answered with a status code: advance only when another model
	// could plausibly succeed (availability / quota), not on client errors that
	// would recur identically.
	var se *StatusError
	if errors.As(err, &se) {
		switch se.Code {
		case http.StatusTooManyRequests, // 429 — rate limit / quota saturated
			http.StatusNotFound,       // 404 — model absent from the catalog
			http.StatusRequestTimeout: // 408 — upstream timeout
			return true
		}
		return se.Code >= 500 // any 5xx is a server-side fault
	}
	// No status code => network / timeout / decode failure => availability issue.
	return true
}
