package main

import (
	"testing"

	"github.com/mlorentedev/pollex/internal/adapter"
	"github.com/mlorentedev/pollex/internal/config"
)

// TestBuildAdaptersNanCloud: with a NaN key configured, a single "NaN Cloud
// (auto)" entry is registered, backed by a fallback chain over every configured
// model. (Acceptance criterion AC5.)
func TestBuildAdaptersNanCloud(t *testing.T) {
	cfg := config.Config{
		NanAPIKey: "sk-test",
		NanModels: []string{"mimo-v2.5", "qwen3.6", "gemma4"},
	}

	adapters, models := buildAdapters(cfg, false)

	a, ok := adapters["nan-cloud"]
	if !ok {
		t.Fatal("expected nan-cloud adapter to be registered when NanAPIKey is set")
	}
	if a.Name() != "NaN Cloud (auto)" {
		t.Errorf("adapter name: got %q, want %q", a.Name(), "NaN Cloud (auto)")
	}

	th, ok := a.(*adapter.Throttle)
	if !ok {
		t.Fatalf("nan-cloud adapter should be *Throttle, got %T", a)
	}
	fc, ok := th.Adapter.(*adapter.FallbackChain)
	if !ok {
		t.Fatalf("throttle should wrap *FallbackChain, got %T", th.Adapter)
	}
	if len(fc.Adapters) != 3 {
		t.Errorf("chain length: got %d, want 3", len(fc.Adapters))
	}

	var count int
	for _, m := range models {
		if m.ID == "nan-cloud" {
			count++
			if m.Name != "NaN Cloud (auto)" || m.Provider != "nan" {
				t.Errorf("model info: got %+v", m)
			}
		}
	}
	if count != 1 {
		t.Errorf("expected exactly one nan-cloud model entry, got %d", count)
	}
}

// TestBuildAdaptersNoNanWithoutKey: no key => no cloud engine exposed.
func TestBuildAdaptersNoNanWithoutKey(t *testing.T) {
	cfg := config.Config{NanModels: []string{"mimo-v2.5", "qwen3.6", "gemma4"}} // no API key

	adapters, models := buildAdapters(cfg, false)

	if _, ok := adapters["nan-cloud"]; ok {
		t.Error("nan-cloud should NOT be registered without an API key")
	}
	for _, m := range models {
		if m.ID == "nan-cloud" {
			t.Error("nan-cloud should not appear in /api/models without a key")
		}
	}
}

// TestBuildAdaptersNanKeyButNoModels: a key with an empty model list must not
// register an empty (useless) cloud chain.
func TestBuildAdaptersNanKeyButNoModels(t *testing.T) {
	cfg := config.Config{NanAPIKey: "sk-test", NanModels: nil}

	adapters, _ := buildAdapters(cfg, false)

	if _, ok := adapters["nan-cloud"]; ok {
		t.Error("nan-cloud should NOT be registered with an empty model list")
	}
}
