---
tags: [spec, tasks, templates]
created: "2026-06-05"
---

# Tasks - FEAT-001-nan-cloud-engine

> TDD order. One task = one focused commit. Tick as you go. Reorder freely while spec is in `draft`; freeze once `implementing`.

## Setup

- [ ] Branch created from master: `feat/nan-cloud-engine`
- [ ] `proposal.md` complete and acceptance criteria testable
- [x] Open question resolved (2026-06-05): live `GET /v1/models` + functional smoke confirm **mimo-v2.5, qwen3.6, gemma4 all present and polishing correctly**; `enable_thinking:false` honored (reasoning_tokens=0). Chain order `mimo→qwen→gemma` valid.

## Implementation

> Small commits, TDD order. Mirror `internal/adapter/llamacpp.go` (OpenAI-compat) + `claude.go` (Bearer + `Available()`).

- [ ] Write failing unit test for `NousAdapter.Polish` (`httptest`): asserts `Authorization: Bearer`, `enable_thinking:false` in body, parses `choices[0].message.content`
- [ ] Implement `internal/adapter/nous.go` (`NousAdapter`: BaseURL/APIKey/Model/Client, `Name/Polish/Available`) to make it pass
- [ ] Write failing unit test: response containing `reasoning_content` → output excludes the reasoning text
- [ ] Harden `NousAdapter` parsing (read `content` only; ignore `reasoning_content`)
- [ ] Write failing unit test for `FallbackChain` fall-through (mimo fail→qwen; mimo+qwen fail→gemma; all fail→wrapped error) via `httptest`
- [ ] Implement `internal/adapter/fallback.go` (`FallbackChain{Adapters []LLMAdapter}`) — **business logic: which errors trigger fallback** (see learning note below)
- [ ] Add config: `NousAPIKey`, `NousBaseURL`, ordered `NousModels` (default `mimo-v2.5,qwen3.6,gemma4`) + `POLLEX_NAN_*` env overrides; extend `internal/config/config_test.go`
- [ ] Wire `cmd/pollex/main.go:buildAdapters`: when NaN key present, build the 3 `NousAdapter`s, wrap in `FallbackChain`, register under id `nous-cloud` with `ModelInfo{Name:"Nous Cloud (auto)", Provider:"nan"}`
- [ ] Integration test hitting `nan.builders` for **each** of mimo-v2.5 / qwen3.6 / gemma4 (build tag or `t.Skip` when `NAN_API_KEY` unset) — proves all three work
- [ ] dotfiles: add `POLLEX_NAN_API_KEY=nan.api-key` to `sensitive/env-mapping.conf`; confirm `make deploy-secrets` writes it to `/etc/pollex/secrets.env`
- [ ] Extension: ensure dropdown renders the `nous-cloud` option from `/api/models` (no hardcoded model list)
- [ ] Docs: ADR (`docs/adr/`) for the cloud engine + fixed fallback-chain decision; update `CLAUDE.md` + `README.md` + `docs/architecture/`

## Closing

- [ ] Every acceptance criterion covered by ≥1 test
- [ ] `features.json` emitted with non-vacuous verification commands
- [ ] `go test -race ./...` green; `make lint` clean
- [ ] No unrelated changes in the diff (no scope creep)
- [ ] **Final cross-browser deployment test**: load the extension + polish via the cloud engine against the deployed Jetson API in Chrome, Edge, Brave (Chromium); document Firefox/Manifest-V3 status. Record results in `verification.md`.
- [ ] `verification.md` filled in
- [ ] PR opened referencing this spec folder

## Learning note — your call to make (FallbackChain error policy)

The one genuinely interesting decision in this feature is **which errors make the chain advance to the next model**. Options:

- **Advance on any error** — simplest, most resilient, but a `400 bad request` (malformed prompt) re-fails identically on every model = wasted calls + latency.
- **Advance only on availability/quota errors** (`404` unknown model, `429` rate/quota, `5xx`, network/timeout) — matches your "si no está o ha saturado" intent; a `400` fails fast without retrying.

Recommended: the second. You write the ~10-line classifier in `fallback.go` (`shouldFallback(err) bool` / a sentinel-error switch) — that's where your intent lives. The rest is mechanical.

## Machine-readable features

Emit a sibling `features.json` ([[pattern-feature-list-as-primitive]]): one entry per acceptance criterion with `id`, `behavior`, `verification` (executable command), `state` (`pending` until the harness proves `passing`), `evidence`.
