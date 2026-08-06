package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mlorentedev/pollex/internal/adapter"
	"github.com/mlorentedev/pollex/internal/config"
	"github.com/mlorentedev/pollex/internal/server"
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

// TestMockModeDisablesAuth: mock mode must force auth off even when
// POLLEX_API_KEY is set in the environment (dotfiles exposes it on every new
// shell, which used to break `make dev` for the extension).
//
// Exercises the real production path: config.Load (with the leaked env var) →
// applyFlagOverrides → buildAdapters → server.SetupMux → an unauthenticated
// request that must succeed.
func TestMockModeDisablesAuth(t *testing.T) {
	t.Setenv("POLLEX_API_KEY", "leaked-from-shell")

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("config load failed: %v", err)
	}

	applyFlagOverrides(&cfg, 0, true) // the exact call main() makes for --mock

	adapters, models := buildAdapters(cfg, true)
	if _, ok := adapters["mock"]; !ok {
		t.Fatal("expected mock adapter to be registered in mock mode")
	}

	// Wire the real middleware chain with the (cleared) production key and
	// assert an unauthenticated request gets through — proving auth is off.
	h := server.SetupMux(adapters, models, func(string) string { return "" }, cfg.APIKey, "test")
	req := httptest.NewRequest(http.MethodGet, "/api/models", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("unauthenticated GET /api/models: got %d, want 200 (auth should be off in mock mode)", rr.Code)
	}
}
