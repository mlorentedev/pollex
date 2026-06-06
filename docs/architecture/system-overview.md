---
id: "pollex-system-overview"
type: architecture
status: active
tags: [pollex, project]
created: "2026-02-15"
owner: manu
---

# Pollex — Architecture

## System Overview

```mermaid
graph TB
    subgraph Browser["Browser (User Machine)"]
        EXT["Chrome Extension<br/>Manifest V3<br/>popup.js · api.js"]
    end

    subgraph Cloudflare["Cloudflare (Cloud)"]
        DNS["DNS CNAME<br/>pollex.mlorente.dev"]
        EDGE["Cloudflare Edge<br/>TLS termination · DDoS"]
    end

    subgraph Jetson["Jetson Nano 4GB (ARM64, CUDA 10.2)"]
        CFD["cloudflared<br/>Tunnel daemon<br/>(outbound only)"]

        subgraph API["pollex-api · Go · :8090"]
            MW["Middleware Chain<br/>CORS → RequestID → Logging<br/>→ Metrics → APIKey → RateLimit<br/>→ MaxBytes(64KB) → Timeout(120s)"]
            HANDLERS["Handlers<br/>/api/polish · /api/health<br/>/api/models · /metrics"]
            ADAPTERS["Adapter Interface<br/>LlamaCpp · Claude · Ollama · Mock"]
            PROBE["Adapter Probe<br/>goroutine 30s<br/>updates availability gauge"]
        end

        subgraph LLM["llama-server · :8080"]
            GPU["128 Maxwell CUDA cores"]
            MODEL["Qwen 2.5 1.5B Q4_0<br/>1017MB · mlock'd"]
        end

        subgraph Config["Configuration"]
            YAML["/etc/pollex/config.yaml"]
            SECRET["/etc/pollex/secrets.env<br/>POLLEX_API_KEY"]
            PROMPT["/etc/pollex/polish.txt"]
        end
    end

    subgraph Monitoring["Monitoring Host (Docker)"]
        PROM["Prometheus · :9090<br/>scrape 30s · 6 alert rules"]
        AM["Alertmanager · :9093<br/>Slack webhook routing"]
        GRAF["Grafana · :3000<br/>Pollex SRE Overview<br/>11 panels"]
    end

    SLACK["Slack<br/>#pollex-alerts"]

    EXT -->|"HTTPS + X-API-Key"| DNS
    DNS --> EDGE
    EDGE <-->|"Tunnel (outbound)"| CFD
    CFD -->|"HTTP :8090"| MW
    MW --> HANDLERS
    HANDLERS --> ADAPTERS
    ADAPTERS -->|"POST /v1/chat/completions<br/>HTTP :8080"| GPU
    GPU --> MODEL
    API --- YAML
    API --- SECRET
    API --- PROMPT

    PROM -->|"GET /metrics<br/>HTTPS via tunnel"| EDGE
    PROM -->|"alerts"| AM
    AM -->|"webhook"| SLACK
    GRAF -->|"PromQL"| PROM
```

## Request Flow (Polish)

```mermaid
sequenceDiagram
    participant U as User
    participant E as Extension
    participant CF as Cloudflare Edge
    participant T as cloudflared
    participant MW as Middleware Chain
    participant H as Polish Handler
    participant A as LlamaCpp Adapter
    participant L as llama-server (GPU)

    U->>E: Paste text, click Polish
    E->>E: Validate (≤1500 chars)
    E->>E: Show spinner + timer (0.0s)

    E->>CF: POST /api/polish<br/>X-API-Key header
    CF->>T: Forward via tunnel
    T->>MW: HTTP localhost:8090

    Note over MW: CORS headers
    Note over MW: Generate X-Request-ID (crypto/rand 32 hex)
    Note over MW: Start logging timer
    Note over MW: Record pollex_requests_total
    Note over MW: Validate X-API-Key (constant-time compare)
    Note over MW: Check rate limit (10 req/min/IP)
    Note over MW: Check body ≤ 64KB
    Note over MW: Set 120s timeout context

    MW->>H: Request with context
    H->>H: Parse JSON, validate text ≤ 10000 chars
    H->>H: Record pollex_input_chars
    H->>A: Polish(ctx, text, systemPrompt)
    A->>L: POST /v1/chat/completions<br/>{"messages":[system, user]}
    L->>L: Tokenize → GPU inference → Detokenize
    Note over L: ~53ms per input char<br/>~3s tiny · ~16s short · ~43s medium
    L-->>A: {"choices":[{"message":{"content":"..."}}]}
    A-->>H: polished text
    H->>H: Record pollex_polish_duration_seconds
    H-->>MW: 200 {"polished":"...", "model":"...", "elapsed_ms":N}

    Note over MW: Log: request_id, method, path, status, duration_ms

    MW-->>T: HTTP response
    T-->>CF: Tunnel response
    CF-->>E: HTTPS response

    E->>E: Hide spinner, show polished text
    U->>E: Click Copy
    E->>E: Copy to clipboard
```

## Service Dependencies (Systemd)

```mermaid
graph LR
    subgraph Boot["Jetson Boot Sequence"]
        NET["network-online.target"]
        LLAMA["llama-server.service<br/>Loads model (~30s)<br/>GPU memory allocation"]
        POLLEX["pollex-api.service<br/>Reads config + secrets<br/>Starts adapter probe"]
        CFTUN["cloudflared.service<br/>Opens outbound tunnel<br/>(API + SSH ingress)"]
    end

    NET --> LLAMA
    NET --> CFTUN
    LLAMA -->|"After="| POLLEX

    subgraph Restart["Failure Recovery"]
        R1["Restart=on-failure<br/>RestartSec=5"]
        R2["Restart=on-failure<br/>RestartSec=5"]
        R3["Restart=on-failure<br/>RestartSec=10"]
    end

    LLAMA -.- R1
    POLLEX -.- R2
    CFTUN -.- R3
```

## Monitoring Architecture

```mermaid
graph TB
    subgraph Pollex["pollex-api (Jetson)"]
        METRICS["/metrics endpoint<br/>pollex_requests_total<br/>pollex_polish_duration_seconds<br/>pollex_input_chars<br/>pollex_adapter_available"]
        PROBE["Adapter Probe (30s)<br/>Updates availability gauge"]
    end

    subgraph SLOs["SLO Definitions (ADR-007)"]
        AVAIL["Availability ≥ 99%<br/>API up AND llamacpp available<br/>Budget: 100.8 min/week"]
        LAT["Latency<br/>p50 < 20s · p95 < 60s"]
        ERR["Error Rate < 1%<br/>5xx on /api/polish"]
    end

    subgraph Stack["Monitoring Stack (Docker)"]
        PROM["Prometheus<br/>Scrape 30s · 7d retention"]
        RULES["6 Alert Rules<br/>PollexDown (2m)<br/>LlamaCppDown (5m)<br/>HighLatencyP50 (10m)<br/>HighLatencyP95 (10m)<br/>HighErrorRate (5m)<br/>ErrorBudgetBurn (14.4x)"]
        AM["Alertmanager<br/>Group by alertname+severity<br/>Critical: repeat 1h<br/>Warning: repeat 4h"]
        GRAF["Grafana Dashboard<br/>SLO status row (5 panels)<br/>Traffic & errors (2 panels)<br/>Latency percentiles (2 panels)<br/>Infrastructure (2 panels)"]
    end

    SLACK["Slack #pollex-alerts"]

    PROBE --> METRICS
    METRICS -->|"scrape"| PROM
    PROM --> RULES
    SLOs -.-|"defines targets"| RULES
    RULES -->|"firing"| AM
    AM -->|"webhook"| SLACK
    PROM -->|"PromQL"| GRAF
```

## Secrets Flow

```mermaid
graph LR
    subgraph Dotfiles["~/Projects/dotfiles"]
        AGE["pollex.api-key.secret.age<br/>(age-encrypted)"]
    end

    subgraph Shell["Developer Shell"]
        ENV["$POLLEX_API_KEY<br/>(decrypted in memory)"]
    end

    subgraph Jetson["Jetson Nano"]
        SECRETS["/etc/pollex/secrets.env<br/>chmod 600"]
        SYSTEMD["pollex-api.service<br/>EnvironmentFile="]
        RUNTIME["pollex-api process<br/>cfg.APIKey in memory"]
    end

    subgraph Extension["Chrome Extension"]
        STORAGE["chrome.storage.local<br/>apiKey field"]
    end

    AGE -->|"age decrypt<br/>(shell profile)"| ENV
    ENV -->|"make deploy-secrets<br/>(SSH + tee)"| SECRETS
    SECRETS -->|"EnvironmentFile="| SYSTEMD
    SYSTEMD -->|"POLLEX_API_KEY env"| RUNTIME
    ENV -->|"manual paste<br/>in Settings"| STORAGE

    subgraph Rotation["Key Rotation"]
        ROT["secrets_rotate POLLEX_API_KEY<br/>→ make deploy-secrets<br/>→ update extension Settings"]
    end
```

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/polish` | `X-API-Key` | Polish text via selected model |
| GET | `/api/models` | `X-API-Key` | List available models |
| GET | `/api/health` | None | Rich health check with per-adapter availability |
| GET | `/metrics` | None | Prometheus metrics |

## Related

- [ADR-005: Cloudflare Tunnel](../adr/005-cloudflare-tunnel-public-access.md)
- [ADR-007: SLOs and SLIs](../adr/007-slos-and-slis.md)
- [ADR-008: Multi-Node Deployment](../adr/008-multi-node-deployment.md)
- [Benchmark Baseline](../benchmarks/jetson-nano-baseline.md)
