---
id: "pollex-adr-001-local-llm-jetson"
type: adr
status: accepted
tags: [adr, pollex]
created: "2026-02-10"
owner: manu
---

# ADR-001: Local LLM on Jetson Nano

**Status:** Accepted
**Date:** 2026-02-10

## Context

Pollex needs an LLM backend to polish English text. The two main options are:

1. **Cloud API only** (Claude, OpenAI) — high quality, simple setup, but ongoing cost per request and dependency on external services.
2. **Local inference on Jetson Nano** — zero cost per inference, full privacy, works offline, but limited to small models (1-3B params) and slower responses (10-30s).

The Jetson Nano 4GB is already available hardware. It has 4GB shared CPU/GPU memory, CUDA 10.2 with 128 Maxwell cores.

## Decision

Run **local LLM inference on the Jetson Nano** as the primary backend, using Ollama as the runtime and **Qwen2.5-1.5B** (Q4 quantized) as the default model. Keep Claude API as an optional secondary backend for quality comparison.

## Rationale

- **Privacy:** Text never leaves the local network. Important for drafts, personal messages, and work documents.
- **Cost:** Zero per-inference cost after hardware acquisition. Cloud APIs charge per token.
- **Latency:** LAN round-trip is predictable (~10-30s on Jetson). No dependency on internet quality or API rate limits.
- **Availability:** Works during internet outages. No API key management for the primary path.
- **Learning:** Practical experience running models on constrained hardware.

## Consequences

### Positive

- Complete data privacy (text stays on LAN)
- No recurring costs
- Works offline
- Predictable performance (no API throttling)

### Negative

- Limited to models that fit in ~1GB VRAM (1-3B params, quantized)
- Response quality inferior to Claude or GPT-4 class models
- Slower inference (10-30s vs 1-3s for cloud APIs)
- Hardware-specific deployment complexity (ARM64, CUDA 10.2, JetPack)

### Mitigations

- Claude API adapter available for quality comparison and fallback
- System prompt optimized for small model capabilities (shorter version if needed)
- Tiered timeout configuration (extension 70s → API 65s → adapter 60s)

## Model Selection

| Model | Size | VRAM | Quality | Speed |
|-------|------|------|---------|-------|
| Qwen2.5-0.5B | 0.4GB | ~0.5GB | Low | Fast (~5s) |
| **Qwen2.5-1.5B** | **1.0GB** | **~1.0GB** | **Good** | **Medium (~15s)** |
| Qwen2.5-3B | 2.0GB | ~2.0GB | Better | Slow (~30s+) |
| Phi-3-mini | 2.3GB | ~2.3GB | Good | Very slow |

Qwen2.5-1.5B chosen as best balance of quality vs resource usage on 4GB hardware.
