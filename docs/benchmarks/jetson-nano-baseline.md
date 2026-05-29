---
id: "pollex-benchmark-jetson-nano"
type: reference
status: active
tags: [benchmark, pollex]
created: "2026-02-14"
owner: manu
---

# Jetson Nano 4GB — Benchmark Baseline

## Hardware
- NVIDIA Jetson Nano 4GB (Maxwell, 128 CUDA cores, compute 5.3)
- CUDA 10.2, LPDDR4 25.6 GB/s (shared CPU/GPU)
- jetson_clocks enabled, -c 1024 context

## Configuration History

### v1 — Q4_K_M (Feb 12, 2026)
```
-m qwen2.5-1.5b-instruct-q4_k_m.gguf -ngl 999 -c 1024 -t 4
--batch-size 1024 --ubatch-size 512
```

| Sample | Chars | Warm Avg (ms) | ~tok/s | Status |
|--------|-------|---------------|--------|--------|
| tiny   | 103   | 6,957         | ~4     | OK     |
| short  | 350   | 20,968        | ~4     | OK     |
| medium | 850   | 55,394        | ~4     | Borderline |
| long   | 1800  | —             | —      | Timeout (65s) |
| max    | 5500  | —             | —      | Timeout (65s) |

### v2 — Q4_0 + mlock (Feb 14, 2026)
```
-m qwen2.5-1.5b-instruct-q4_0.gguf -ngl 999 -c 1024 -t 4
--batch-size 1024 --ubatch-size 512 --mlock
LimitMEMLOCK=infinity (systemd)
```

| Sample | Chars | Warm Avg (ms) | vs v1  |
|--------|-------|---------------|--------|
| tiny   | 103   | 5,719         | **-18%** |
| short  | 350   | 16,261        | **-22%** |
| medium | 850   | 43,097        | **-22%** |
| long   | 1800  | 117,095       | Was timeout, now borderline |

Changes applied:
- Q4_0 quantization: faster compute, slightly lower quality (1017MB vs 1.1GB)
- `--mlock`: model locked in RAM via `LimitMEMLOCK=infinity` in systemd
- Headless mode: `systemctl set-default multi-user.target` (frees ~400MB RAM)
- Thread A/B test: `-t 2` vs `-t 4` showed no difference with full GPU offload

## Key Metrics
- Speed: ~68ms per input character (v1) → ~53ms/char (v2 short text estimate)
- Run 1 always ~20% slower (cold KV cache)
- Timeout: 120s (middleware), 125s (extension), 180s (benchmark client)
- Extension char limit: 1500 (based on 120s timeout budget)

## Optimization Research

### Applied
1. **Q4_0 quantization**: 23% faster on short text, confirmed
2. **--mlock + LimitMEMLOCK=infinity**: prevents model page faults
3. **Headless mode**: `multi-user.target`, frees ~400MB RAM
4. **Batch size** (--batch-size 1024 --ubatch-size 512): already in v1
5. **GGML_CUDA_ENABLE_UNIFIED_MEMORY=1**: already in v1

### Tested, no impact
- **Thread tuning** (-t 2 vs -t 4): identical with -ngl 999

### Deferred
- **Zram tuning**: only 29MB/2GB used, negligible overhead
- **KV cache quantization**: requires flash_attn, NOT supported on Maxwell

### NOT applicable on Maxwell
- Flash attention (requires compute 6.0+)
- GGML_CUDA_F16 (requires compute 6.0+)

### Realistic ceiling
5-7 tok/s with all optimizations. Hard limit: LPDDR4 memory bandwidth.

## Load Test Results (k6)

### Single-user — Feb 15, 2026

Scenario: `jetson` (1 VU, 5 min, sequential via Cloudflare Tunnel).
Config: v2 (Q4_0 + mlock + headless). k6 with `SCENARIO=jetson`.

| Metric | Value | SLO (ADR-007) | Status |
|--------|-------|---------------|--------|
| Iterations | 32 in 5 min | — | — |
| Error rate | 0.00% | < 1% | **Pass** |
| Polish p50 | 6.4s | < 20s | **Pass** |
| Polish p95 | 14.6s | < 60s | **Pass** |
| Polish min/max | 2.4s / 14.6s | — | — |
| Health p95 | 166ms | < 2s | **Pass** |
| HTTP failed | 0.00% | — | — |
| Checks passed | 128/128 (100%) | — | — |

Observations:
- Without GPU queueing (1 VU), latency tracks input length linearly (~53ms/char, consistent with v2 benchmark)
- Health endpoint fast when not contending with inference (~67ms avg)
- All SLOs pass comfortably under single-user load

### Multi-user burst — Feb 15, 2026 (informational)

Scenario: original `all` (5 VUs burst + 0.2 iter/s normal). Not representative of real usage but documents GPU queueing behavior.

| Metric | Value | Notes |
|--------|-------|-------|
| Polish p50 | 50.4s | Requests queue on single GPU |
| Polish p95 | 72s | Exceeds 60s SLO |
| Health p95 | 2.08s | llama-server contended |
| Burst completed | 13/25 (52%) | 22 iterations dropped |
| Errors | 0% | All completed requests succeeded |

Conclusion: Jetson Nano is a single-user device. Concurrent inference serializes on the GPU, causing p95 > 60s. The `jetson` scenario (1 VU) is the correct test profile.

## Related
- [001-local-llm-on-jetson-nano](../adr/001-local-llm-on-jetson-nano.md)
- [004-llamacpp-gpu-acceleration](../adr/004-llamacpp-gpu-acceleration.md)
- [jetson-memory](../troubleshooting/jetson-memory.md)
- [007-slos-and-slis](../adr/007-slos-and-slis.md)
