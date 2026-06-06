package adapter

import (
	"context"
	"log/slog"
	"strings"
	"unicode"
)

// AutoRouter routes Polish requests to the Local adapter for short texts
// and the Cloud adapter for longer texts, based on sentence count.
// It is itself an LLMAdapter and is exposed as the "auto" model.
type AutoRouter struct {
	Local     LLMAdapter
	Cloud     LLMAdapter
	Threshold int // route to cloud when sentence count >= Threshold
}

func (r *AutoRouter) Name() string { return "auto" }

func (r *AutoRouter) Available() bool {
	return r.Local.Available() || r.Cloud.Available()
}

func (r *AutoRouter) Polish(ctx context.Context, text, systemPrompt string) (string, error) {
	n := countSentences(text)
	if n < r.Threshold {
		slog.Info("routing to local", "sentences", n, "threshold", r.Threshold)
		return r.Local.Polish(ctx, text, systemPrompt)
	}
	slog.Info("routing to cloud", "sentences", n, "threshold", r.Threshold)
	return r.Cloud.Polish(ctx, text, systemPrompt)
}

// countSentences counts sentence-ending punctuation (. ! ?) followed by
// whitespace or end-of-string. Single punctuation without trailing space
// (e.g. abbreviations mid-sentence) is ignored. Returns at least 1 for
// any non-empty input.
func countSentences(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	runes := []rune(text)
	count := 0
	for i, r := range runes {
		if r == '.' || r == '!' || r == '?' {
			next := i + 1
			if next >= len(runes) || unicode.IsSpace(runes[next]) {
				count++
			}
		}
	}
	if count == 0 {
		return 1 // treat as one sentence when no terminal punctuation
	}
	return count
}
