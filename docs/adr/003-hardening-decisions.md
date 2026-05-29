---
id: "pollex-adr-003-hardening"
type: adr
status: accepted
tags: [adr, pollex]
created: "2026-02-11"
owner: manu
---

# ADR-003: Production Hardening Decisions

**Status:** Accepted
**Date:** 2026-02-11

## Context

Pollex is a LAN-facing API server running on a Jetson Nano. While not internet-exposed, it still needs basic hardening: protection against accidental abuse, request tracing for debugging, and graceful shutdown for clean deployments.

## Decisions

### 1. Graceful Shutdown (10-second drain)

**What:** Replace `http.ListenAndServe` with `http.Server{}` + `signal.Notify(SIGINT, SIGTERM)` + `srv.Shutdown(ctx)` with 10-second timeout.

**Why:** systemd sends SIGTERM during `systemctl restart`. Without graceful shutdown, in-flight polish requests (which can take 10-30 seconds on Ollama) would be killed mid-response. The 10-second drain is a pragmatic balance — long enough for most requests to complete, short enough to not block deployments.

### 2. Request Body Limit (64KB)

**What:** `maxBytesMiddleware` wraps `r.Body` with `http.MaxBytesReader(w, r.Body, 64*1024)`. The polish handler detects `*http.MaxBytesError` and returns 413.

**Why:** Prevents accidental or malicious large payloads from consuming memory. 64KB is generous for text polishing (a 10,000-character text in JSON is ~10KB). The limit applies at the middleware level before the handler reads the body.

### 3. Text Length Limit (10,000 characters)

**What:** `handlePolish` rejects text over `maxTextLength = 10000` with 400.

**Why:** Protects Ollama from excessively long prompts that would cause OOM on the Jetson Nano's 4GB RAM. 10,000 characters covers any reasonable single-text polishing request.

### 4. Rate Limiting (10 requests/minute per IP)

**What:** In-memory sliding window rate limiter using `sync.Mutex` + `map[string][]time.Time`. Returns 429 when exceeded.

**Why:** Prevents a stuck browser extension or script from overwhelming the Jetson. 10 req/min is generous for interactive use (one polish every 6 seconds) but catches runaway loops. In-memory storage is appropriate — the server is single-instance and restarts reset the window, which is acceptable.

**Not chosen:** Token bucket (more complex, no benefit for this use case), Redis-backed (external dependency), stdlib rate.Limiter (per-token, harder to configure for req/min).

### 5. Request ID (crypto/rand, 32 hex chars)

**What:** `requestIDMiddleware` generates a unique ID per request via `crypto/rand`, sets `X-Request-ID` response header, stores in context. Logging middleware includes the ID in log output: `[req-id] METHOD PATH STATUS DURATION`.

**Why:** Enables correlating extension errors with server logs. When a user reports "polish failed", we can match the X-Request-ID from the browser network tab to the server log.

### 6. Rich Health Check

**What:** `/api/health` returns per-adapter availability status with human-readable reasons for unavailability (e.g., "no API key" for Claude, "ollama unreachable" for Ollama).

**Why:** Operational visibility during deployment. One `curl /api/health` shows which backends are functional without checking each adapter separately.

## Middleware Chain Order

```
CORS → requestID → logging → rateLimit → maxBytes → timeout → mux
```

**Rationale:**
1. **CORS** first — handles OPTIONS preflight before any processing
2. **requestID** — assigns tracing ID before logging sees it
3. **logging** — captures all requests including rate-limited/rejected ones
4. **rateLimit** — rejects before body parsing (saves resources)
5. **maxBytes** — limits body size before handler reads it
6. **timeout** — wraps handler execution (65s matching tiered timeout design)

## Constraints

- Go stdlib only — no external middleware libraries
- All state is in-memory (rate limiter, request IDs)
- Single-instance deployment (no distributed coordination needed)
