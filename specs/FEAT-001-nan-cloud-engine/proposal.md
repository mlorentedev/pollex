---
id: "FEAT-001-nan-cloud-engine"
type: spec
status: draft # draft | implementing | verifying | archived
created: "2026-06-05"
tags: [spec, proposal]
template_version: "1.0"
---

# FEAT-001: NaN cloud inference engine

> **Naming**: file lives at `pollex/specs/FEAT-001-nan-cloud-engine/proposal.md`.

## Why

<!-- from 10_projects/pollex/11-tasks.md: FEAT-001 — add a nan.builders OpenAI-compatible adapter exposed as "Nous Cloud (auto)" alongside the Jetson llama.cpp engine, backed by a 3-model fallback chain (mimo-v2.5 → qwen3.6 → gemma4). -->

Pollex's inference depends on a single Jetson Nano 4GB. When that node is unavailable (OOM kill, power loss, `llama-server` crash, GPU contention) polishing fails completely — there is no second engine. Adding a cloud inference engine via **`nan.builders`** (a community OpenAI-compatible gateway already used in this ecosystem by the Hermes agent, key already managed as the `nan.api-key` age-secret) lets the user pick a cloud engine per request, decoupling service availability from a single edge device and offering a faster/higher-quality path when desired.

## What

Concrete behavior after this PR:

1. A new **`NousAdapter`** (`internal/adapter/nous.go`) implementing `LLMAdapter`, talking to `https://api.nan.builders/v1/chat/completions` with `Authorization: Bearer` auth. It reads the answer from `choices[0].message.content` and **ignores** the non-standard `reasoning_content` field; it sends `chat_template_kwargs:{enable_thinking:false}` to suppress reasoning latency/quota burn on reasoning models.
2. A new **`FallbackChain`** adapter (`internal/adapter/fallback.go`) wrapping an *ordered* list of adapters; `Polish()` tries each in turn and returns the first success, falling through on availability/quota/5xx errors. Configured order: `mimo-v2.5` → `qwen3.6` → `gemma4` (fixed, **not** user-selectable).
3. `GET /api/models` lists a **single** `"Nous Cloud (auto)"` entry (model id e.g. `nous-cloud`) when the NaN key is configured, coexisting with the Jetson llama.cpp model(s). The extension dropdown shows it; the user chooses the engine via the existing `model_id` request mechanism.
4. The NaN key is sourced from the existing `nan.api-key` age-secret, deployed to `/etc/pollex/secrets.env` on the Jetson via `make deploy-secrets`, surfaced as `POLLEX_NAN_API_KEY`.
5. A **`Throttle`** decorator (also an `LLMAdapter`) wraps the chain and bounds concurrent NaN calls (`POLLEX_NAN_MAX_CONCURRENT`, default 3) so Pollex stays under the gateway's account-wide 5-concurrent cap and leaves headroom for the user's other tooling (Hermes, `qq`).

## Out of scope

- Per-model user selection inside the cloud chain (order is fixed; the dropdown shows one auto entry).
- Automatic Jetson→cloud failover / a high-level "engine toggle" (cloud is an *explicit* dropdown choice here; auto-failover is a possible future PR).
- Non-chat NaN models (`qwen3-embedding`, `kokoro` TTS, `whisper` STT) and streaming responses.
- Per-user quota/billing tracking or cost dashboards.

## Risks / open questions

- **[RESOLVED 2026-06-05]** `mimo-v2.5` availability on `nan.builders` — **confirmed present** via live `GET /v1/models` (the `dotfiles/ai/nan/README.md` catalog was stale). Functional smoke test: all three chain models (`mimo-v2.5`, `qwen3.6`, `gemma4`) returned correct polish (HTTP 200). The chain order is valid as specified.
- **Latency: mimo is the slowest first hop.** Live smoke (identical 1-line input, `enable_thinking:false`): mimo `3.17s` vs qwen `1.37s` vs gemma `1.08s`, with **identical output** on simple text. mimo's quality edge only materializes on harder input; for the polishing endpoint a 1500-char input through mimo could reach 5–8s. **Action:** confirm the extension/client timeout budget tolerates the mimo first-hop latency (the chain only advances on *errors*, not slowness). Default-mimo is a quality-ceiling-over-latency choice — accepted by the user.
- **Account-wide rate limit**: nan.builders allows 100 RPM / 5 concurrent **per account**, shared with the user's `qq`/TUI/Hermes. Pollex is semi-public (Cloudflare Tunnel + API key); an extension traffic burst could starve the user's interactive tooling (429). Mitigation: cloud path stays API-key gated; consider backoff.
- **Reasoning-model cost/latency**: reasoning models run 3–10× slower and burn a silent monthly token quota; `enable_thinking:false` mitigates but the gateway must honor it — verify.
- **No SLA / no deprecation policy** on the community gateway; model slugs may rotate. Keep the chain config-driven, not hardcoded.
- **Secret hygiene**: ensure the NaN key reaches `/etc/pollex/secrets.env` and is never logged.

## Acceptance criteria

Observable outcomes. Each must be testable.

- [ ] `NousAdapter.Polish()` returns polished text from `nan.builders` and the request carries the `Authorization: Bearer` header + `enable_thinking:false` (unit test, `httptest`).
- [ ] `NousAdapter` reads `choices[0].message.content` and the polished output **never** contains `reasoning_content` (unit test with a reasoning-shaped response body).
- [ ] **All three models work individually** against the live gateway: `mimo-v2.5`, `qwen3.6`, `gemma4` each return a non-empty polish (integration test, skipped when `NAN_API_KEY` unset). *(Pre-confirmed manually 2026-06-05 — all 3 HTTP 200 with correct output; to be codified as an automated integration test.)*
- [ ] `FallbackChain` returns the qwen result when mimo errors, the gemma result when mimo+qwen error, and a wrapped error when all three fail (unit test, `httptest`).
- [ ] `GET /api/models` lists exactly one `"Nous Cloud (auto)"` entry when the NaN key is set, and none when it is unset.
- [ ] Concurrent NaN calls are bounded to `NanMaxConcurrent` (unit test); a request blocked on a full semaphore respects context cancellation.
- [ ] **Final cross-browser test**: the extension loads and successfully polishes text via the cloud engine against the deployed Jetson API in **Chrome, Edge, and Brave** (≥2 Chromium browsers; Firefox status documented). [AGENT-SUGGESTION — confirm exact browser list before archive]

## References

- Vault: `10_projects/pollex/11-tasks.md` (FEAT-001 backlog entry)
- NaN gateway docs: `dotfiles/ai/nan/README.md`; debug script `dotfiles/scripts/nan-debug.sh`
- Secret: `dotfiles/sensitive/nan.api-key.secret.age`, mapping `NAN_API_KEY=nan.api-key`
- Existing patterns to mirror: `internal/adapter/llamacpp.go` (OpenAI-compat shape), `internal/adapter/claude.go` (Bearer-style cloud + `Available()`)
- Related patterns: `00_meta/patterns/pattern-architecture.md`, `00_meta/patterns/pattern-config-defaults.md`
