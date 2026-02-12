package main

import (
	"context"
	"testing"
)

func TestGenerateRequestID(t *testing.T) {
	id := generateRequestID()
	if len(id) != 32 {
		t.Errorf("length: got %d, want 32", len(id))
	}

	// Must be unique
	id2 := generateRequestID()
	if id == id2 {
		t.Error("two generated IDs should not be equal")
	}
}

func TestGenerateRequestIDIsHex(t *testing.T) {
	id := generateRequestID()
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("non-hex character %q in request ID %q", c, id)
		}
	}
}

func TestRequestIDContext(t *testing.T) {
	ctx := context.Background()

	// Empty context returns empty string
	if got := requestIDFromContext(ctx); got != "" {
		t.Errorf("empty context: got %q, want %q", got, "")
	}

	// Stored ID is retrievable
	ctx = contextWithRequestID(ctx, "abc123")
	if got := requestIDFromContext(ctx); got != "abc123" {
		t.Errorf("stored ID: got %q, want %q", got, "abc123")
	}
}
