---
id: "pollex-adr-007-slos-slis"
type: adr
status: accepted
tags: [adr, pollex]
created: "2026-02-15"
owner: manu
---

# ADR-007: SLOs and SLIs for Production Monitoring

**Status:** Accepted
**Date:** 2026-02-15
**Related:** [006-q4-0-quantization](006-q4-0-quantization.md), [jetson-nano-baseline](../benchmarks/jetson-nano-baseline.md)

## Context

Phase 13.1-13.2 added Prometheus metrics and structured JSON logging. The system now exposes four metric families:

- `pollex_requests_total` (counter, labels: method, path, status)
- `pollex_polish_duration_seconds` (histogram, label: model, buckets: 0.5–120s)
- `pollex_input_chars` (histogram)
- `pollex_adapter_available` (gauge, label: adapter)

With instrumentation in place, we need formal SLIs/SLOs to define "good enough" and create actionable error budgets. Pollex runs on a single Jetson Nano 4GB with no redundancy — SLOs must reflect that reality.

## Decision

Define three SLIs with corresponding SLOs measured over a **7-day rolling window**. Use error budgets to drive alerting decisions in Phase 13.4.

## SLI Definitions

### 1. Availability (composite)

**What it measures:** The system can accept and process polish requests end-to-end.

**Definition:** Both conditions must be true simultaneously:

| Component | Metric | Healthy |
|-----------|--------|---------|
| API process | `up{job="pollex"}` | `== 1` |
| Inference backend | `pollex_adapter_available{adapter="llamacpp"}` | `== 1` |

**PromQL (availability ratio):**

```promql
avg_over_time(
  (up{job="pollex"} == bool 1 and on() pollex_adapter_available{adapter="llamacpp"} == bool 1)
[7d:1m])
```

**Prerequisite:** The adapter gauge is only updated on `/api/health` hits (`internal/handler/health.go:27`). A synthetic probe must hit `/api/health` at the Prometheus scrape interval to keep the gauge fresh. Configure in Phase 13.4.

### 2. Latency (polish inference)

**What it measures:** How long polish requests take from the user's perspective.

**Definition:** Duration of successful POST `/api/polish` requests, measured by `pollex_polish_duration_seconds`.

**PromQL (p95):**

```promql
histogram_quantile(0.95,
  rate(pollex_polish_duration_seconds_bucket[7d])
)
```

**Benchmark baseline (Q4_0 + mlock, warm):**

| Input size | Chars | Observed latency | Estimated rate |
|------------|-------|-----------------|----------------|
| tiny | 103 | 5.7s | ~53 ms/char |
| short | 350 | 16.3s | ~46 ms/char |
| medium | 850 | 43.1s | ~51 ms/char |
| long | 1500 (limit) | ~79s (extrapolated) | ~53 ms/char |

Extension enforces 1500 char limit. Typical usage skews toward short-to-medium texts.

### 3. Error Rate (polish endpoint)

**What it measures:** Fraction of polish requests that fail with server errors.

**Definition:** 5xx responses on `/api/polish` as a proportion of total polish requests.

**PromQL:**

```promql
sum(rate(pollex_requests_total{path="/api/polish",status=~"5.."}[7d]))
/
sum(rate(pollex_requests_total{path="/api/polish"}[7d]))
```

**Note:** 4xx responses (400 bad input, 401 auth, 429 rate limit) are client errors — they do NOT count against the error SLO. Only 5xx (adapter failures, timeouts, panics) count.

## SLO Targets

| SLI | Target | Error Budget (7d) |
|-----|--------|-------------------|
| Availability | ≥ 99% | 100.8 minutes of downtime |
| Latency p50 | < 20s | — |
| Latency p95 | < 60s | 5% of requests can exceed 60s |
| Error rate | < 1% | 1% of polish requests can be 5xx |

### Rationale for Targets

**Availability 99%** — Single Jetson, no failover, no UPS. Reboots (kernel updates, headless mode switch, model changes) consume budget. 99.9% would be ~10 minutes/week — unrealistic without redundancy. 99% gives 1h40m for planned maintenance.

**Latency p50 < 20s** — Benchmark shows short text (350 chars, most common use case) at 16.3s warm. The p50 target validates that typical requests complete in reasonable time. Cold-start first request runs ~20% slower (KV cache warming).

**Latency p95 < 60s** — Accommodates medium text (850 chars, 43s) and some longer texts. At 53ms/char, 60s covers inputs up to ~1130 chars. The 1500 char extension limit means worst-case is ~79s, which falls in the allowed 5% above target.

**Error rate < 1%** — Transient failures (llama-server restart, Cloudflare hiccup) are expected but rare. With low request volume (personal tools), even 1-2 errors per week could approach 1% — target is tight enough to surface real problems.

## Error Budget Policy

| Budget state | Action |
|-------------|--------|
| Budget healthy (>50% remaining) | Normal operations, deploy freely |
| Budget warning (10-50% remaining) | Investigate. Avoid non-essential changes to Jetson |
| Budget exhausted (<10% remaining) | Freeze deployments. Focus on reliability |
| Budget violated (0% remaining) | Post-mortem. Identify root cause before resuming changes |

For a single-user tools, "freeze deployments" means: don't push model changes or system updates to Jetson until the SLO recovers in the next 7-day window.

## Histogram Bucket Assessment

Current `pollex_polish_duration_seconds` buckets: `0.5, 1, 2, 5, 10, 20, 30, 60, 120`.

These provide adequate resolution for the defined SLOs:

- p50 at ~16s falls in the 10–20 bucket (OK)
- p95 at ~43-60s falls in the 30–60 bucket (OK)
- p99 near 120s falls in the 60–120 bucket (OK)

No bucket changes needed.

## Consequences

### Positive

- Clear definition of "working" vs "degraded" for portfolio demonstration
- Error budgets make risk/reliability tradeoffs explicit
- PromQL queries ready for Phase 13.4 alerting rules
- Targets grounded in real benchmark data, not arbitrary percentages

### Negative

- Low request volume (personal tools) means noisy percentiles — a single slow request can swing p95 significantly
- Availability SLI requires a synthetic probe to keep adapter gauge fresh (implementation in Phase 13.4)

### Open Questions for Phase 13.4

- Probe interval: 30s vs 60s for health check (balance between gauge freshness and unnecessary load)
- Alert routing: where do alerts go? (Slack webhook, email, Grafana notification)
- Dashboard: which panels are essential vs nice-to-have for a single-node setup
