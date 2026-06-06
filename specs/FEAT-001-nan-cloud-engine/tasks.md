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

- [x] Write failing unit test for `NousAdapter.Polish` (`httptest`): Bearer, `enable_thinking:false`, parse `content` — `495beb5`
- [x] Implement `internal/adapter/nous.go` (`NousAdapter`) — `495beb5`
- [x] Write failing unit test: response with `reasoning_content` → output excludes reasoning — `495beb5`
- [x] Harden parsing (read `content` only; ignore `reasoning_content`) — `495beb5`; base-URL `/v1` tolerance — `6fd9447`
- [x] Write failing unit test for `FallbackChain` fall-through — `054a293`
- [x] Implement `internal/adapter/fallback.go` (`FallbackChain`) + `shouldFallback` classifier — `054a293`
- [x] Add config `NanAPIKey`/`NanBaseURL`/`NanModels` + `POLLEX_NAN_*` overrides + tests — `0e59fd4`
- [x] Wire `buildAdapters`: chain registered as `nous-cloud` / "Nous Cloud (auto)" / provider `nan` — `0897c22`
- [x] Integration test for each of mimo/qwen/gemma (tag `integration`, skips without key) — `c9fe978`
- [x] `Throttle` decorator: concurrency semaphore over the chain + `nan_max_concurrent` config — `de148ef`
- [ ] **dotfiles (cross-repo handoff): add `POLLEX_NAN_API_KEY=nan.api-key` to `sensitive/env-mapping.conf`** so `secrets_refresh` exports it for `make deploy-secrets`
- [x] Extension renders `nous-cloud` from `/api/models` (already dynamic) + `nan` provider label — `484572f`
- [x] Docs: ADR-009 + `CLAUDE.md` + `README.md` — `877b3d2`

## Closing

- [x] Every acceptance criterion covered by ≥1 test (AC6 cross-browser is manual)
- [x] `features.json` emitted
- [x] `go test -race ./...` green; `make lint` clean
- [x] No feature scope creep (pre-existing gofmt + stale extension label fixed in their own commits)
- [ ] **Final cross-browser deployment test** (AC6): extension + cloud-engine polish in Chrome, Edge, Brave; document Firefox/MV3 status. Record in `verification.md`. — needs key deployed to Jetson
- [x] `verification.md` filled in
- [ ] PR opened referencing this spec folder

## Learning note — your call to make (FallbackChain error policy)

The one genuinely interesting decision in this feature is **which errors make the chain advance to the next model**. Options:

- **Advance on any error** — simplest, most resilient, but a `400 bad request` (malformed prompt) re-fails identically on every model = wasted calls + latency.
- **Advance only on availability/quota errors** (`404` unknown model, `429` rate/quota, `5xx`, network/timeout) — matches your "si no está o ha saturado" intent; a `400` fails fast without retrying.

Recommended: the second. You write the ~10-line classifier in `fallback.go` (`shouldFallback(err) bool` / a sentinel-error switch) — that's where your intent lives. The rest is mechanical.

## Machine-readable features

Emit a sibling `features.json` ([[pattern-feature-list-as-primitive]]): one entry per acceptance criterion with `id`, `behavior`, `verification` (executable command), `state` (`pending` until the harness proves `passing`), `evidence`.
