---
id: "pollex-adr-002-testing-strategy"
type: adr
status: accepted
tags: [adr, pollex]
created: "2026-02-11"
owner: manu
---

# ADR-002: Testing Strategy

**Status:** Accepted
**Date:** 2026-02-11

## Context

Pollex has 33 unit tests covering individual components (adapters, handlers, middleware, config) but lacks integration tests that exercise the full middleware stack. The backend uses a straightforward architecture: HTTP handlers behind a middleware chain, all in `package main`. We need a testing strategy that validates end-to-end behavior without external dependencies.

## Decision

Use **`httptest.NewServer`** with the full middleware stack and `MockAdapter` for integration testing. No external dependencies (no Docker, no running Ollama instance, no test databases).

### Key design choices:

1. **`setupMux()` extraction** — Extracted from `main()` to return a fully-wired `http.Handler` with all middleware. This is the unit under test for integration tests.

2. **`MockAdapter` as the LLM backend** — The existing `MockAdapter` (capitalizes first letter, configurable delay) provides deterministic behavior for assertions.

3. **`failingAdapter` for error propagation** — A test-only adapter that always returns an error, used to verify 502 responses propagate through the full stack.

4. **Real HTTP connections** — `httptest.NewServer` creates a real TCP listener, so tests exercise the actual HTTP transport (connection pooling, headers, body handling).

## Test Categories

| Category | Tool | What it validates |
|----------|------|-------------------|
| Unit tests | `httptest.NewRecorder` | Individual handler/middleware behavior in isolation |
| Integration tests | `httptest.NewServer` | Full middleware stack, CORS + request ID + logging + rate limit + maxBytes + timeout |

## Consequences

### Positive

- Zero external dependencies — tests run anywhere with `go test`
- Full middleware chain exercised (CORS, request ID, logging, rate limiting, body limits, timeout)
- Race condition detection via `-race` flag
- Fast execution (~1 second for full suite)

### Negative

- Does not test real LLM adapter behavior (network, parsing real responses)
- Does not test graceful shutdown (requires process lifecycle)
- MockAdapter behavior differs from real adapters (no actual text polishing)

### Mitigations

- Adapter-specific tests use `httptest` servers to simulate real API responses (Ollama, Claude)
- Rate limiter has dedicated unit tests with time-based window expiry
- Concurrent test validates thread safety under `-race`

## Test Count Progression

| Phase | Tests | What was added |
|-------|-------|----------------|
| Fase 1-3 | 33 | Unit tests for all components |
| Phase 6 | 43 | Integration tests (8) + rich health tests (2) |
| Phase 7 | 63+ | Request ID (3), middleware (4), rate limiter (7), text limits (2), integration hardening (3) |
