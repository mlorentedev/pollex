# Pollex

**Polish your English text** — fixes grammar, syntax, and coherence. The output sounds like a fluent non-native speaker: professional and clear, not AI-generated.

## Architecture

```mermaid
graph LR
    subgraph Your Machine
        EXT["Browser Extension<br/>(Manifest V3)"]
    end

    subgraph Jetson Nano 4GB
        API["Pollex API<br/>(Go · :8090)"]
        OLL["Ollama<br/>(:11434)"]
        MODEL["Qwen 2.5 1.5B<br/>(Q4 · ~1GB VRAM)"]
    end

    subgraph Cloud
        CLAUDE["Claude API<br/>(optional)"]
    end

    EXT -- "HTTP JSON<br/>LAN" --> API
    API -- "/api/chat" --> OLL
    OLL --> MODEL
    API -. "Messages API<br/>(comparison)" .-> CLAUDE

    style EXT fill:#4a90d9,stroke:#3a7bc8,color:#fff
    style API fill:#2ecc71,stroke:#27ae60,color:#fff
    style OLL fill:#e67e22,stroke:#d35400,color:#fff
    style MODEL fill:#f39c12,stroke:#e67e22,color:#fff
    style CLAUDE fill:#9b59b6,stroke:#8e44ad,color:#fff
```

**Three layers, zero complexity:**

| Layer | Tech | Role |
|-------|------|------|
| Extension | Manifest V3 | Pure UI — no API keys, no logic |
| API | Go 1.26, stdlib `net/http` | Routes text to LLM backends, returns polished result |
| LLM | Ollama + Qwen 2.5 1.5B | Local inference on Jetson Nano (Claude API optional) |

## How It Works

```mermaid
sequenceDiagram
    participant U as User
    participant E as Extension
    participant A as Pollex API
    participant O as Ollama
    participant M as Qwen 2.5

    U->>E: Paste text + click Polish
    E->>E: Show spinner (0s...)
    E->>A: POST /api/polish
    A->>O: POST /api/chat
    O->>M: Inference (~10-30s)
    M-->>O: Polished text
    O-->>A: Response
    A-->>E: {"polished":"...", "elapsed_ms":...}
    E->>E: Hide spinner, show result
    U->>E: Click Copy
    E->>E: Copy to clipboard
```

## Quick Start

### Development (Docker)

```sh
make ollama-up      # Start Ollama in Docker + pull model (~1GB)
make dev-ollama     # Start API connected to Ollama on :8090
```

Then load the extension in Chrome: `chrome://extensions` → Developer mode → Load unpacked → select `extension/`.

### Development (Mock)

```sh
make dev            # Start API with mock adapter (no LLM needed)
```

### Run Tests

```sh
make test           # 33 tests with race detector
make lint           # go vet + gofmt
```

## API

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/polish` | Polish text via selected model |
| `GET` | `/api/models` | List available models |
| `GET` | `/api/health` | Health check |

**Example:**

```sh
curl -X POST http://localhost:8090/api/polish \
  -H 'Content-Type: application/json' \
  -d '{"text":"i goes to store yesterday","model_id":"qwen2.5:1.5b"}'

# {"polished":"I went to the store yesterday.","model":"qwen2.5:1.5b","elapsed_ms":4830}
```

## Project Structure

```
pollex/
├── backend/              # Go API (package main, ~700 lines)
│   ├── main.go           # Entry point, wiring, --mock flag
│   ├── config.go         # YAML config + env var overrides
│   ├── adapter.go        # LLMAdapter interface
│   ├── adapter_mock.go   # Mock adapter for development
│   ├── adapter_ollama.go # Ollama (local LLM)
│   ├── adapter_claude.go # Claude API (optional)
│   ├── handler_*.go      # HTTP handlers
│   ├── middleware.go      # CORS, logging, timeout
│   └── *_test.go         # Table-driven tests
├── extension/            # Browser extension (Manifest V3)
│   ├── popup.*           # Main UI
│   ├── settings.*        # API URL configuration
│   └── api.js            # HTTP client
├── prompts/
│   └── polish.txt        # System prompt (9 rules)
├── deploy/               # systemd, install scripts
└── Makefile              # All targets
```

## Hardware Target

**Jetson Nano 4GB** — ARM64, CUDA 10.2, 128 Maxwell cores.

| Component | RAM |
|-----------|-----|
| JetPack OS (headless) | ~500MB |
| Ollama runtime | ~200MB |
| Qwen 2.5 1.5B (Q4) | ~1.0GB |
| Pollex API | ~15MB |
| **Free** | **~2.3GB** |

## Deploy to Jetson

```sh
make deploy-setup   # First time: install Ollama + systemd service
make deploy         # Build ARM64 binary + SCP + restart service
make jetson-status  # Remote health check
```

## Makefile Targets

```sh
make help
```

| Target | Description |
|--------|-------------|
| `dev` | Start API with mock adapter |
| `dev-ollama` | Start API with local Ollama |
| `test` | Run all tests with race detector |
| `lint` | go vet + gofmt |
| `ollama-up` | Start Ollama in Docker |
| `ollama-down` | Stop Ollama container |
| `build` | Build for current platform |
| `build-arm64` | Cross-compile for Jetson |
| `deploy` | Deploy to Jetson |
| `deploy-setup` | First-time Jetson setup |

## License

[MIT](LICENSE)
