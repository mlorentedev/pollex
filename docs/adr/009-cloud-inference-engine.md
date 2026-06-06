---
id: 009-cloud-inference-engine
type: adr
status: active
created: "2026-06-05"
---

# ADR-009: Cloud Inference Engine (NaN) with Fixed Fallback Chain

> **Status:** Accepted — implemented in FEAT-001.

## Context

Pollex inference depended entirely on a single Jetson Nano 4GB. When that node is unavailable (OOM kill, power loss, `llama-server` crash, GPU contention) polishing fails with no alternative. We want a cloud engine the user can choose per request, decoupling availability from one edge device and offering a faster/higher-quality path.

The `nan.builders` gateway is already part of this ecosystem (consumed by the Hermes agent; key already managed as the `nan.api-key` age-secret). It is an **OpenAI-compatible** API (`https://api.nan.builders/v1`, `Authorization: Bearer`). Constraints discovered during scoping:

- **Community, best-effort:** no SLA, no deprecation policy, model slugs rotate.
- **Account-wide rate limit:** 100 RPM / 5 concurrent, **shared** with the user's `qq`/TUI/Hermes.
- **Reasoning fields:** reasoning models stream the chain-of-thought in a non-standard `reasoning_content` field; `enable_thinking:false` suppresses it (verified `reasoning_tokens=0`).
- **mimo-v2.5** is a reasoning model: ~2–3× slower than `qwen3.6`/`gemma4` and draws on a capped monthly token pool; `qwen3.6`/`gemma4` are unlimited.

## Decision

1. Add `NousAdapter` — an OpenAI-compatible adapter for `nan.builders` with Bearer auth. It reads `choices[0].message.content`, **ignores** `reasoning_content`, and sends `enable_thinking:false`. It accepts the base URL with or without a trailing `/v1`.
2. Expose the cloud option as a **single** `"NaN Cloud (auto)"` entry (`model_id` `nan-cloud`) backed by a `FallbackChain` — a composite `LLMAdapter` over an **ordered, non-user-selectable** list: **`mimo-v2.5` → `qwen3.6` → `gemma4`**. Unlimited models sit at the tail so the chain almost always returns something.
3. The cloud entry coexists with the Jetson `llama.cpp` model in `GET /api/models`; the user selects the engine through the existing `model_id` mechanism — **no new "engine" concept**.
4. **Fallback policy:** advance to the next model only on availability/quota errors (HTTP 429, 404, 5xx, network/timeout); fail fast on client errors (400, 401) and caller cancellation, which would recur identically.
5. The key is sourced from the existing `nan.api-key` age-secret, exported as `NAN_API_KEY` by dotfiles and deployed to `/etc/pollex/secrets.env`. The chain is config-driven via `POLLEX_NAN_MODELS` so rotating slugs need no code change.
6. A `Throttle` decorator wraps the cloud chain and bounds concurrent NaN calls (`POLLEX_NAN_MAX_CONCURRENT`, default 3) so Pollex stays under the gateway's account-wide 5-concurrent cap and leaves headroom for other consumers. It bounds *concurrency*, not request *rate* (the ~100 RPM cap would need a separate token-bucket).

## Consequences

**Positive**

- Service availability is no longer tied to a single edge device; cloud is a faster option for users who want it.
- `FallbackChain` is itself an `LLMAdapter`, so it registers like any model and the extension (which renders `/api/models` dynamically) needs no change beyond a provider label.
- The key never reaches the browser — cloud calls route through the Jetson-hosted Go API, keeping the secret server-side.

**Negative / trade-offs**

- "Cloud" still depends on the Go API host (the Jetson) being up; this mitigates GPU/`llama-server` failure, not total host loss. (Automatic Jetson→cloud failover is explicitly out of scope — see Alternatives.)
- Pollex shares the account-wide NaN rate limit with the user's interactive tooling; an extension traffic burst can contend (429). Mitigated for the *concurrency* cap by the `Throttle` semaphore (default 3); the ~100 RPM *rate* cap is not yet bounded (token-bucket = future work). The cloud path stays API-key gated.
- Default `mimo-v2.5` trades latency for a quality ceiling (~2–3× slower; capped pool). The chain absorbs quota/availability hits by falling to the unlimited tail.

## Alternatives considered

- **Engine toggle + automatic Jetson→cloud failover.** Rejected for now: adds orchestration; an explicit dropdown choice is simpler and sufficient. Viable future PR.
- **Extension calls NaN directly.** Rejected: would leak the API key to the client.
- **Reasoning-model default with thinking on.** Rejected: wasted latency and quota for a deterministic rewrite.

## Related

- [ADR-001: Local LLM on Jetson Nano](001-local-llm-on-jetson-nano.md)
- [ADR-004: llama.cpp GPU Acceleration](004-llamacpp-gpu-acceleration.md)
- Spec: `specs/FEAT-001-nan-cloud-engine/`
- Gateway reference: `dotfiles/ai/nan/README.md`
