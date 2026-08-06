# Pollex — Makefile
JETSON_HOST     ?= jet1
JETSON_FALLBACK ?= jet1-lan
JETSON_USER     ?= manu
API_PORT        ?= 8090
EFFECTIVE_HOST   = $(JETSON_HOST)

# ─── Development ────────────────────────────────────────────
.PHONY: dev test lint

dev: ## Start API with mock adapter on :$(API_PORT)
	go run ./cmd/pollex --mock --port $(API_PORT)

test: ## Run all tests with race detector
	go test -v -race ./...

test-extension: ## Run extension tests (Vitest unit + Playwright e2e against local mock API)
	cd extension && npm run test:all

test-extension-unit: ## Run extension unit tests (Vitest, no browser needed)
	cd extension && npm test

test-extension-e2e: ## Run extension e2e tests (Playwright + Chromium, boots mock API on :8099)
	cd extension && npm run test:e2e

extension-deps: ## Install extension test tooling (npm install + Playwright Chromium)
	cd extension && npm install && npx playwright install chromium

lint: ## Run go vet + check formatting
	go vet ./... && gofmt -l internal/ cmd/

# ─── Build ──────────────────────────────────────────────────
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  = -ldflags "-X main.version=$(VERSION)"

.PHONY: build build-arm64

build: ## Build binary for current platform
	go build $(LDFLAGS) -o dist/pollex ./cmd/pollex

build-arm64: ## Cross-compile for ARM64 (Jetson Nano)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/pollex-arm64 ./cmd/pollex

# ─── Benchmark ──────────────────────────────────────────────
.PHONY: bench bench-jetson

bench: ## Run benchmark against local API (add --quality for input/output view)
	go run ./cmd/benchmark --url http://localhost:$(API_PORT) $(BENCH_ARGS)

bench-jetson: ## Run benchmark against Jetson (via Cloudflare Tunnel)
	go run ./cmd/benchmark --url https://pollex.mlorente.dev --api-key $$POLLEX_API_KEY $(BENCH_ARGS)

# ─── Deploy (Jetson) ────────────────────────────────────────
.PHONY: deploy deploy-init deploy-llamacpp deploy-tunnel deploy-secrets _resolve-jetson

# Internal: probe JETSON_HOST, fall back to JETSON_FALLBACK with warning
_resolve-jetson:
	$(eval EFFECTIVE_HOST := $(shell \
	  if ssh -o ConnectTimeout=3 -o BatchMode=yes $(JETSON_HOST) true 2>/dev/null; then \
	    echo $(JETSON_HOST); \
	  else \
	    printf "\033[33m⚠  $(JETSON_HOST) unreachable, using $(JETSON_FALLBACK)\033[0m\n" >&2; \
	    echo $(JETSON_FALLBACK); \
	  fi))

deploy-init: _resolve-jetson ## First-time Jetson setup (packages, CUDA, dirs, systemd)
	rsync -Pz deploy/systemd/pollex-api.service $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/pollex-api.service
	rsync -Pz deploy/systemd/llama-server.service $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/llama-server.service
	rsync -Pz deploy/systemd/jetson-clocks.service $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/jetson-clocks.service
	ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'bash -s' < deploy/scripts/init.sh

deploy: _resolve-jetson build-arm64 ## Build + deploy binary, config, prompts, and restart to Jetson
	rsync -Pz dist/pollex-arm64 $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/pollex
	rsync -Pz deploy/config.yaml $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/pollex-config.yaml
	rsync -Pz prompts/polish.txt $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/pollex-polish.txt
	rsync -Pz prompts/polish-cloud.txt $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/pollex-polish-cloud.txt
	rsync -Pz deploy/systemd/pollex-api.service $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/pollex-api.service
	ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'bash -s' < scripts/jetson-install.sh
	@echo "Restarting pollex-api..."
	@ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'sudo systemctl restart pollex-api'
	@echo "Done."

deploy-secrets: _resolve-jetson ## Deploy API keys from env to Jetson secrets.env
	@test -n "$$POLLEX_API_KEY" || (echo "POLLEX_API_KEY not set" && exit 1)
	@ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'sudo mkdir -p /etc/pollex && \
	  echo "POLLEX_API_KEY='"$$POLLEX_API_KEY"'" | sudo tee /etc/pollex/secrets.env > /dev/null && \
	  sudo chmod 600 /etc/pollex/secrets.env'
	@test -n "$$NAN_API_KEY" && \
	  ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'echo "NAN_API_KEY='"$$NAN_API_KEY"'" | sudo tee -a /etc/pollex/secrets.env > /dev/null' || true
	@echo "Secrets deployed."

deploy-llamacpp: _resolve-jetson ## Build llama.cpp with CUDA on Jetson (~85 min)
	rsync -Pz deploy/scripts/build-llamacpp.sh $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/build-llamacpp.sh
	rsync -Pz deploy/systemd/llama-server.service $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/llama-server.service
	ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'bash /tmp/build-llamacpp.sh'

deploy-tunnel: _resolve-jetson ## Setup Cloudflare Tunnel on Jetson (interactive)
	rsync -Pz deploy/scripts/setup-cloudflared.sh $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/setup-cloudflared.sh
	rsync -Pz deploy/systemd/cloudflared.service $(JETSON_USER)@$(EFFECTIVE_HOST):/tmp/cloudflared.service
	ssh -t $(JETSON_USER)@$(EFFECTIVE_HOST) 'bash /tmp/setup-cloudflared.sh'

# ─── Jetson Remote ──────────────────────────────────────────
.PHONY: jetson-ssh jetson-logs jetson-status jetson-test jetson-tunnel-status jetson-tunnel-logs

jetson-ssh: _resolve-jetson ## SSH into Jetson
	ssh $(JETSON_USER)@$(EFFECTIVE_HOST)

jetson-logs: _resolve-jetson ## Tail pollex-api service logs on Jetson
	ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'sudo journalctl -u pollex-api -f'

jetson-status: _resolve-jetson ## Health check via SSH (per-adapter status)
	@ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'curl -s localhost:$(API_PORT)/api/health' | python3 -m json.tool

jetson-test: _resolve-jetson ## End-to-end polish test on Jetson (needs POLLEX_API_KEY)
	@ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'curl -s -X POST localhost:$(API_PORT)/api/polish \
	  -H "Content-Type: application/json" \
	  -H "X-API-Key: '"$$POLLEX_API_KEY"'" \
	  -d '"'"'{"text":"This is a test to see if pollex works end to end on the jetson nano.","model_id":"qwen2.5-1.5b-gpu"}'"'"'' | python3 -m json.tool

jetson-tunnel-status: _resolve-jetson ## Check Cloudflare Tunnel status
	ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'sudo systemctl status cloudflared'

jetson-tunnel-logs: _resolve-jetson ## Tail Cloudflare Tunnel logs on Jetson
	ssh $(JETSON_USER)@$(EFFECTIVE_HOST) 'sudo journalctl -u cloudflared -f'

# ─── Docker ────────────────────────────────────────────────
.PHONY: docker-build docker-dev docker-down

docker-build: ## Build pollex Docker image (alpine:3.21, 24.7MB)
	docker build \
		--build-arg VERSION=$$(git describe --tags --always 2>/dev/null || echo dev) \
		--build-arg VCS_REF=$$(git rev-parse --short HEAD) \
		-t pollex:latest .

docker-dev: ## Start pollex in Docker (mock mode) on :8090
	docker compose up -d --build

docker-down: ## Stop pollex Docker container
	docker compose down

# ─── Monitoring ────────────────────────────────────────────
.PHONY: monitoring-up monitoring-down monitoring-validate

monitoring-up: ## Start Prometheus + Alertmanager + Grafana (needs `make dev` running)
	docker compose -f docker-compose.monitoring.yml up -d

monitoring-down: ## Stop monitoring stack
	docker compose -f docker-compose.monitoring.yml down

monitoring-validate: ## Validate Prometheus rules and config syntax
	docker run --rm --entrypoint promtool -v $(PWD)/deploy/prometheus:/p prom/prometheus check rules /p/alerts.yml
	docker run --rm --entrypoint promtool -v $(PWD)/deploy/prometheus:/p prom/prometheus check config /p/prometheus-local.yml

# ─── Load Testing ─────────────────────────────────────────
.PHONY: loadtest loadtest-jetson loadtest-soak

loadtest: ## Run k6 load test against local API (normal + burst)
	k6 run -e API_KEY=$$POLLEX_API_KEY deploy/loadtest/pollex.js

loadtest-jetson: ## Run k6 load test against Jetson (via Cloudflare Tunnel)
	k6 run -e SCENARIO=jetson -e BASE_URL=https://pollex.mlorente.dev -e API_KEY=$$POLLEX_API_KEY deploy/loadtest/pollex.js

loadtest-soak: ## Run 30-min soak test against Jetson
	k6 run -e SCENARIO=soak -e BASE_URL=https://pollex.mlorente.dev -e API_KEY=$$POLLEX_API_KEY deploy/loadtest/pollex.js

# ─── Utilities ──────────────────────────────────────────────
.PHONY: clean help

clean: ## Remove dist/ directory
	rm -rf dist/

# ─── Help ───────────────────────────────────────────────────
.DEFAULT_GOAL := help
help: ## Show this help
	@grep -E '^[a-zA-Z_][a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'
