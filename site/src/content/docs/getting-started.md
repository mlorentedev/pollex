---
title: Getting Started
description: Install and run Pollex — from source, Docker, or release binary.
---

## Prerequisites

- **Go 1.26+** (from source)
- **Docker** (container mode)
- **Chrome** (for the browser extension)

No GPU needed for development — the mock adapter returns canned responses.

## From Source

Clone the repository and start the dev server:

```sh
git clone https://github.com/mlorentedev/pollex.git
cd pollex
make dev    # Starts API with mock adapter on :8090
```

Verify it works:

```sh
curl -s http://localhost:8090/api/health | python3 -m json.tool
```

Expected output:

```json
{
  "status": "ok",
  "version": "dev",
  "adapters": {
    "mock": { "available": true }
  }
}
```

## With Docker

Build and run in mock mode — no Go toolchain needed:

```sh
make docker-dev    # Build image + start on :8090 (mock mode)
```

The image is 24.7MB (multi-stage Alpine build, non-root user).

To stop:

```sh
make docker-down
```

## From Release Binary

Download the latest binary from the [GitHub releases page](https://github.com/mlorentedev/pollex/releases):

```sh
# Linux amd64
curl -LO https://github.com/mlorentedev/pollex/releases/latest/download/pollex-linux-amd64
chmod +x pollex-linux-amd64
./pollex-linux-amd64 --mock --port 8090

# Linux arm64 (Jetson)
curl -LO https://github.com/mlorentedev/pollex/releases/latest/download/pollex-linux-arm64
chmod +x pollex-linux-arm64
./pollex-linux-arm64 --mock --port 8090
```

## Load the Browser Extension

1. Open `chrome://extensions` in Chrome
2. Enable **Developer mode** (top right toggle)
3. Click **Load unpacked**
4. Select the `extension/` directory from the repository
5. Click the Pollex icon in the toolbar
6. Open **Settings** and set the API URL to `http://localhost:8090`

## Polish Text via API

Send a polish request directly:

```sh
curl -X POST http://localhost:8090/api/polish \
  -H 'Content-Type: application/json' \
  -d '{"text":"i goes to store yesterday","model_id":"mock"}'
```

Response:

```json
{
  "polished": "I went to the store yesterday.",
  "model": "mock",
  "elapsed_ms": 502
}
```

With API key authentication (production):

```sh
curl -X POST https://pollex.mlorente.dev/api/polish \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: YOUR_KEY' \
  -d '{"text":"i goes to store yesterday","model_id":"qwen2.5-1.5b-gpu"}'
```

## List Available Models

```sh
curl -s http://localhost:8090/api/models | python3 -m json.tool
```

```json
[
  { "id": "mock", "name": "Mock (dev)", "provider": "mock" }
]
```

## Configuration

Pollex uses YAML config with environment variable overrides. All env vars use the `POLLEX_` prefix.

| Variable | Default | Description |
| --- | --- | --- |
| `POLLEX_PORT` | `8090` | Listen port |
| `POLLEX_API_KEY` | _(none)_ | API key for `X-API-Key` authentication |
| `POLLEX_LLAMACPP_URL` | _(none)_ | llama.cpp server URL (e.g. `http://localhost:8080`) |
| `POLLEX_LLAMACPP_MODEL` | `qwen2.5-1.5b-gpu` | Model ID for llama.cpp adapter |
| `POLLEX_CLAUDE_API_KEY` | _(none)_ | Anthropic API key (optional cloud fallback) |
| `POLLEX_CLAUDE_MODEL` | `claude-sonnet-4-5-20250929` | Claude model ID |
| `POLLEX_OLLAMA_URL` | _(none)_ | Ollama server URL (legacy) |
| `POLLEX_PROMPT_PATH` | `prompts/polish.txt` | Path to system prompt file |

Example with env overrides:

```sh
POLLEX_PORT=9090 POLLEX_API_KEY=my-secret-key ./pollex --mock
```

Or with a config file:

```sh
./pollex --config deploy/config.yaml
```

## Run Tests

```sh
make test    # All tests with race detector
make lint    # go vet + gofmt check
```

80+ tests across 11 test files covering adapters, handlers, middleware, config, and integration scenarios.

## Monitoring Stack

Start the full observability stack alongside the API:

```sh
make dev              # Start pollex (mock mode)
make monitoring-up    # Start Prometheus + Alertmanager + Grafana
```

- Prometheus: [localhost:9090](http://localhost:9090) — 6 SLO-based alerting rules
- Grafana: [localhost:3000](http://localhost:3000) — auto-provisioned dashboard
- Alertmanager: [localhost:9093](http://localhost:9093) — Slack webhook routing

```sh
make monitoring-down       # Stop monitoring stack
make monitoring-validate   # Validate Prometheus rules syntax
```

## Next Steps

- Browse the full [API reference](https://github.com/mlorentedev/pollex#api) in the README
- Check the [Makefile targets](https://github.com/mlorentedev/pollex#quick-start) — 35 targets for dev, build, bench, docker, monitoring, deploy, and load testing
- Deploy to a Jetson Nano with `make deploy-init` for first-time setup
