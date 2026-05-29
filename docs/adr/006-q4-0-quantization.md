---
id: "pollex-adr-006-q4-quantization"
type: adr
status: accepted
tags: [adr, pollex]
created: "2026-02-14"
owner: manu
---

# ADR-006: Q4_0 Quantization for Inference Performance

**Status:** Accepted
**Date:** 2026-02-14
**Related:** [004-llamacpp-gpu-acceleration](004-llamacpp-gpu-acceleration.md), [jetson-nano-baseline](../benchmarks/jetson-nano-baseline.md)

## Context

After establishing a performance baseline with Qwen 2.5 1.5B using Q4_K_M quantization (ADR-004), benchmarks showed ~4 tok/s with ~68ms per input character. Medium-length texts (850 chars) took ~55s, borderline on the 120s timeout. Long texts (1800+ chars) consistently timed out.

The extension character limit was reduced to 1500 to stay within timeout budget, but even within that limit, processing times of 60-100s provide poor UX. Further optimization was needed without changing hardware.

## Decision

Switch from **Q4_K_M** to **Q4_0** quantization and enable **mlock** to lock the model in RAM. Additionally, configure the Jetson for **headless operation** (no GUI) to free ~400MB RAM.

## Alternatives Considered

| Option | Verdict | Reason |
|--------|---------|--------|
| **Q4_K_M (status quo)** | Rejected | ~68ms/char too slow for acceptable UX |
| **Q4_0 (chosen)** | Accepted | 18-22% faster, equal or better output quality in testing |
| **Q8_0** | Rejected | Higher quality but slower and larger file, wrong direction |
| **Q2_K** | Rejected | Faster but significant quality degradation expected |
| **3B model (Q4_K_M)** | Deferred | Want to maximize 1.5B first; 3B doubles memory pressure |

## Evidence

### Performance (via Cloudflare Tunnel, 3 runs per sample, steady-state)

| Sample | Chars | Q4_K_M | Q4_0 + mlock | Improvement |
|--------|-------|--------|--------------|-------------|
| tiny   | 103   | 6,957ms  | 5,719ms    | **-18%** |
| short  | 350   | 20,968ms | 16,261ms   | **-22%** |
| medium | 850   | 55,394ms | 43,097ms   | **-22%** |
| long   | 1800  | timeout (65s) | 117,095ms | Was impossible, now borderline |

### Quality (5 targeted samples, same input to both models)

| Sample | Type | Q4_0 | Q4_K_M | Verdict |
|--------|------|------|--------|---------|
| subtle (159 chars) | then/than, should of | Correct | Correct | Tie |
| technical (321 chars) | effecting/affecting, jargon | Correct | Correct | Tie |
| informal (410 chars) | run-on, tone elevation | Professional rewrite | Kept informal tone | **Q4_0 wins** |
| academic (541 chars) | wrong prepositions, typos | Correct | Correct | Tie |
| complex (898 chars) | verbose, multi-error | Good condensation | Good condensation | Tie |

Notable: Q4_K_M produced "I and my team" (grammatically wrong) where Q4_0 produced "My team" (correct) in an earlier test.

### Thread A/B Test

`-t 2` vs `-t 4` showed identical performance with full GPU offload (`-ngl 999`). CPU threads are irrelevant when all layers run on GPU. Keeping `-t 4`.

## Implementation

### Service file changes (`deploy/systemd/llama-server.service`)
- Model: `qwen2.5-1.5b-instruct-q4_0.gguf` (1017MB, vs 1.1GB for Q4_K_M)
- Added `--mlock` flag (locks model in RAM, prevents page faults)
- Added `LimitMEMLOCK=infinity` (systemd prerequisite for mlock)
- Previous attempt without `LimitMEMLOCK` caused service crash (ABRT signal)

### Jetson system changes
- `systemctl set-default multi-user.target` (headless mode, frees ~400MB RAM)

### Extension changes (`extension/popup.js`)
- `MS_PER_CHAR`: 68 → 36 (updated estimate for processing time warnings)
- `MAX_CHARS`: 10000 → 1500 (matches 120s timeout budget)
- Draft persistence added (`chrome.storage.local`)

### Benchmark tooling (`cmd/benchmark/`)
- Added 5 `QualitySamples` for reproducible quality comparison
- `--quality` flag shows input/output for each sample
- `make quality-jetson` target for quick quality regression testing

## Consequences

### Positive
- 18-22% faster inference across all text sizes
- Medium text (850 chars) drops from 55s to 43s — better UX
- Quality is equal or better than Q4_K_M (evidence-based)
- mlock eliminates potential page fault latency spikes
- Headless mode frees RAM headroom for future model upgrades

### Negative
- Q4_0 has slightly lower theoretical precision than Q4_K_M (but not measurable in practice for this use case)
- Long texts (1800+ chars) still borderline at 120s timeout — but extension enforces 1500 char limit
- Old Q4_K_M model file (1.1GB) remains on Jetson until manually cleaned

### Neutral
- Both model files coexist on Jetson (`/opt/llama-models/`); Q4_K_M can be restored by reverting service file
