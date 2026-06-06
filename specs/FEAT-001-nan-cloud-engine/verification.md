---
tags: [spec, verification, templates]
created: "2026-06-05"
---

# Verification - FEAT-001-nan-cloud-engine

## Evidence

Map every acceptance criterion from `proposal.md` to concrete proof.

- [x] AC1 NousAdapter Bearer + enable_thinking:false -> `TestNousAdapterPolish` (commit `495beb5`); base-URL `/v1` handling `TestNousAdapterBaseURLWithV1` (`6fd9447`)
- [x] AC2 reasoning_content ignored -> `TestNousAdapterIgnoresReasoningContent` (`495beb5`)
- [x] AC3 all 3 models work live (mimo/qwen/gemma) -> `TestNousIntegrationModels` (`c9fe978`); live run 2026-06-05 PASS for all three (mimo 2.9s, qwen 1.0s, gemma 1.0s)
- [x] AC4 FallbackChain fall-through -> `TestFallbackChain_*` (advance on quota/network, fail-fast on 400, all-fail wrapped) (`054a293`)
- [x] AC5 /api/models lists single "NaN Cloud (auto)" gated on key -> `TestBuildAdaptersNanCloud` / `TestBuildAdaptersNoNanWithoutKey` (`0897c22`)
- [x] AC6 concurrent NaN calls bounded + ctx-cancel respected -> `TestThrottle_*` (`de148ef`)
- [ ] AC7 cross-browser polish via cloud engine -> **PENDING manual smoke** (Chrome/Edge/Brave) — requires deployed key on Jetson

## Pre-implementation smoke test (2026-06-05)

Live against `https://api.nan.builders/v1` (key from `nan.api-key` age-secret):

- `GET /v1/models` → chat models present: `deepseek-v4-flash`, `mimo-v2.5`, `qwen3.6`, `gemma4` (+ non-chat `kokoro`/`whisper`/`qwen3-embedding`). Confirms `mimo-v2.5` exists (stale README corrected).
- Polish smoke (input `"i has went to the store yesterday and buyed two breads."`, `enable_thinking:false`, temp 0.3):
  | Model | HTTP | wall | reasoning_tokens | output |
  |---|---|---|---|---|
  | mimo-v2.5 | 200 | 3.17s | 0 | "I went to the store yesterday and bought two loaves of bread." |
  | qwen3.6 | 200 | 1.37s | 0 | (identical) |
  | gemma4 | 200 | 1.08s | 0 | (identical) |
- Conclusion: chain `mimo→qwen→gemma` valid; `enable_thinking:false` honored (no reasoning tokens); mimo ~2-3× slower than qwen/gemma on simple text.

## Test status

- Unit suite: `source ~/.zshrc && go test -race ./...` -> all packages PASS (2026-06-05)
- Integration (live gateway): `NAN_API_KEY=… go test -tags integration -run TestNousIntegrationModels ./internal/adapter/` -> PASS (mimo/qwen/gemma all polished)
- Lint: `make lint` clean (also fixed pre-existing gofmt debt in `internal/server/integration_test.go`, commit `db2f6c0`)
- Manual smoke test: PENDING (AC6 cross-browser)
- No regressions in existing suite: yes

## Mutation testing (gremlins, 2026-06-05)

Tool: `gremlins` (`go install github.com/go-gremlins/gremlins/cmd/gremlins@latest`). Command: `gremlins unleash --timeout-coefficient 5 <pkg>`.

| Package | Mutator coverage | Survivors (Lived) |
|---|---|---|
| `internal/adapter` | 100% | 0 |
| `internal/config` | 100% | 0 |
| `cmd/pollex` | feature wiring covered; survivors 0 | 0 |

Findings acted on (commit `1ca8032`):

- 2 not-covered mutants in `FallbackChain.Name()` composed branch → added `TestFallbackChain_NameComposed` (adapter coverage 86.7% → 100%).
- Manual-review hardening: `NousAdapter` now errors on a 200 with blank content (chain advances); covered `Polish` empty-adapters path and the `buildAdapters` key-but-no-models guard.

The later `Throttle` decorator + config field were also mutation-tested: `internal/adapter` and `internal/config` stayed at 100% mutator coverage with 0 survivors.

Notes: `cmd/pollex` overall mutator coverage is low (~27%) because `main()`, the adapter-probe loop, and the pre-existing llama.cpp/Claude/Ollama branches are not unit-tested — **pre-existing debt, not introduced here**; the NaN wiring block is fully covered. The many gremlins "timed out" results in `internal/adapter` are an artifact of HTTP tests with real client timeouts/sleeps, not survivors.

## Decisions made during implementation

- **FallbackChain error policy** (`shouldFallback`): advance to the next model on availability/quota errors (HTTP 429, 404, 5xx, network/timeout); fail fast on client errors (400, 401) and caller cancellation (`context.Canceled`/`DeadlineExceeded`). Errors carry HTTP status via a typed `*StatusError`.
- **Integration test resilience:** live tests skip (not fail) on transient gateway errors (timeout/5xx/429) because nan.builders is best-effort with no SLA; they fail only on contract violations. Caught a real gemma4 transient 60s hang during the first run.
- **Base URL hardening:** the adapter strips a trailing `/v1` so it accepts both the host root and the ecosystem's `NAN_BASE_URL` form.
- **Default-mimo trade-off accepted:** mimo is ~2–3× slower than qwen/gemma even with thinking off; quality-ceiling-over-latency, with the unlimited tail absorbing quota/availability hits.

## Promotion candidates

- [ ] Lesson for repo `docs/lessons.md`? Candidate: NaN reasoning lives in `reasoning_content` (not `<think>`); `enable_thinking:false` zeroes reasoning_tokens; account-wide rate limit shared with Hermes/qq. (Decide at archive.)
- [x] ADR-worthy -> **done**: `docs/adr/009-cloud-inference-engine.md` (cloud engine + fixed fallback chain).
- [ ] New pattern for `00_meta/patterns/`? Maybe: "availability-aware fallback chain over an adapter interface" if it recurs (>1 project). Defer.

## Archive checklist

- [ ] `proposal.md` frontmatter `status: archived`
- [ ] Folder moved: `specs/FEAT-001-nan-cloud-engine/` -> `specs/archive/FEAT-001-nan-cloud-engine/`
- [ ] Backlog entry in vault `11-tasks.md` ticked with PR link
- [ ] Promotions above executed (if any)
