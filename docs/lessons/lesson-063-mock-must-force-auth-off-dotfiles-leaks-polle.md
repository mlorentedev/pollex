---
id: lesson-063-mock-must-force-auth-off-dotfiles-leaks-polle
type: lesson
status: active
created: "2026-08-05"
owner: manu
tags: [pollex, lesson, go, extension, dev-loop, auth, dotfiles]
---

# `--mock` must force auth off — dotfiles leaks POLLEX_API_KEY into every shell

**Context:** Pollex dev loop: `make dev` runs `go run ./cmd/pollex --mock`, and the extension defaults to `http://localhost:8090` with an empty API key. dotfiles exposes `POLLEX_API_KEY` as a shell env var (age secret `pollex.api-key` → `expose.env`), so every new shell has it.

**Problem:** `make dev` inherited `POLLEX_API_KEY`, `config.Load` applied the env override, auth came on, and the extension got `401 missing API key` on `/api/models` and `/api/polish` — popup showed "Cannot reach API — check Settings." The plugin looked broken out of the box in local dev.

**Solution:** In `cmd/pollex/main.go`, mock mode clears `cfg.APIKey` (`if *useMock { cfg.APIKey = "" }`). Mock = dev loop = no auth, always. Regression test in `cmd/pollex/main_test.go` (`TestMockModeDisablesAuth`) simulates the leaked env var and asserts auth stays off.

**Why:** The dev loop must work for the extension with zero configuration. Production (Jetson) runs without `--mock`, so real auth is unaffected. This also unblocks `docker-dev`, which uses `--mock` too.

**Tags:** `#go` `#extension` `#dev-loop` `#auth` `#dotfiles`

---
