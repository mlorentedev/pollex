package adapter

import (
	"context"
	"errors"
	"testing"
)

func TestCountSentences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"no punctuation", "hello world", 1},
		{"one sentence", "Hello world.", 1},
		{"two sentences", "Hello world. How are you?", 2},
		{"three sentences", "One. Two! Three?", 3},
		// Simple counter treats "Dr." as a sentence boundary; acceptable approximation for routing.
		{"abbreviation mid-sentence", "Dr. Smith went home. He was tired.", 3},
		{"trailing spaces preserved", "Hello world.  Goodbye.", 2},
		{"question only", "Are you sure?", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countSentences(tt.input); got != tt.want {
				t.Errorf("countSentences(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

type captureAdapter struct {
	name   string
	called bool
}

func (c *captureAdapter) Name() string  { return c.name }
func (c *captureAdapter) Available() bool { return true }
func (c *captureAdapter) Polish(_ context.Context, _, _ string) (string, error) {
	c.called = true
	return "polished by " + c.name, nil
}

func TestAutoRouter_RoutesShortToLocal(t *testing.T) {
	local := &captureAdapter{name: "local"}
	cloud := &captureAdapter{name: "cloud"}
	r := &AutoRouter{Local: local, Cloud: cloud, Threshold: 5}

	// 3 sentences — below threshold
	text := "One sentence. Two sentence. Three sentence."
	out, err := r.Polish(context.Background(), text, "")
	if err != nil {
		t.Fatalf("Polish: %v", err)
	}
	if !local.called {
		t.Error("expected local adapter to be called")
	}
	if cloud.called {
		t.Error("expected cloud adapter NOT to be called")
	}
	if out != "polished by local" {
		t.Errorf("got %q, want 'polished by local'", out)
	}
}

func TestAutoRouter_RoutesLongToCloud(t *testing.T) {
	local := &captureAdapter{name: "local"}
	cloud := &captureAdapter{name: "cloud"}
	r := &AutoRouter{Local: local, Cloud: cloud, Threshold: 3}

	// 4 sentences — at or above threshold
	text := "One. Two. Three. Four."
	out, err := r.Polish(context.Background(), text, "")
	if err != nil {
		t.Fatalf("Polish: %v", err)
	}
	if local.called {
		t.Error("expected local adapter NOT to be called")
	}
	if !cloud.called {
		t.Error("expected cloud adapter to be called")
	}
	if out != "polished by cloud" {
		t.Errorf("got %q, want 'polished by cloud'", out)
	}
}

func TestAutoRouter_AvailableOrOfAdapters(t *testing.T) {
	unavail := &unavailableAdapter{}
	avail := &captureAdapter{name: "avail"}

	r := &AutoRouter{Local: unavail, Cloud: avail, Threshold: 5}
	if !r.Available() {
		t.Error("AutoRouter.Available() should be true when cloud is available")
	}

	r2 := &AutoRouter{Local: avail, Cloud: unavail, Threshold: 5}
	if !r2.Available() {
		t.Error("AutoRouter.Available() should be true when local is available")
	}
}

func TestAutoRouter_Name(t *testing.T) {
	r := &AutoRouter{Local: &MockAdapter{}, Cloud: &MockAdapter{}, Threshold: 5}
	if r.Name() != "auto" {
		t.Errorf("Name() = %q, want %q", r.Name(), "auto")
	}
}

// unavailableAdapter is an adapter that is always unavailable.
type unavailableAdapter struct{}

func (u *unavailableAdapter) Name() string  { return "unavail" }
func (u *unavailableAdapter) Available() bool { return false }
func (u *unavailableAdapter) Polish(_ context.Context, _, _ string) (string, error) {
	return "", errors.New("unavailable")
}
