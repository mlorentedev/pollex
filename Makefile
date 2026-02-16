# Pollex — Makefile
# Configurable variables
JETSON_HOST ?= nvidia
JETSON_USER ?= manu
API_PORT    ?= 8090

# ─── Development ────────────────────────────────────────────
.PHONY: dev test lint

dev: ## Start API with mock adapter on :$(API_PORT)
	go run ./cmd/pollex --mock --port $(API_PORT)

test: ## Run all tests with race detector
	go test -v -race ./...

lint: ## Run go vet + check formatting
	go vet ./... && gofmt -l internal/ cmd/

# ─── Build ──────────────────────────────────────────────────
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  = -ldflags "-X main.version=$(VERSION)"

.PHONY: build build-arm64 ext-zip

build: ## Build binary for current platform
	go build $(LDFLAGS) -o dist/pollex ./cmd/pollex

build-arm64: ## Cross-compile for ARM64 (Jetson Nano)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/pollex-arm64 ./cmd/pollex

ext-zip: ## Package extension into dist/pollex-ext.zip
	cd extension && zip -r ../dist/pollex-ext.zip . -x '*.gitkeep'

# ─── Benchmark ──────────────────────────────────────────────
.PHONY: bench bench-jetson quality quality-jetson

bench: ## Run performance benchmark against local API
	go run ./cmd/benchmark --url http://localhost:$(API_PORT)

bench-jetson: ## Run benchmark against Jetson (via Cloudflare Tunnel)
	go run ./cmd/benchmark --url https://pollex.mlorente.dev --api-key $$POLLEX_API_KEY

quality: ## Run quality test against local API (shows input/output)
	go run ./cmd/benchmark --quality --url http://localhost:$(API_PORT)

quality-jetson: ## Run quality test against Jetson (via Cloudflare Tunnel)
	go run ./cmd/benchmark --quality --url https://pollex.mlorente.dev --api-key $$POLLEX_API_KEY

# ─── Deploy (Jetson) ────────────────────────────────────────
.PHONY: deploy deploy-init deploy-secrets deploy-llamacpp deploy-tunnel

deploy-init: ## First-time Jetson setup (packages, CUDA, dirs, systemd)
	scp deploy/systemd/pollex-api.service $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-api.service
	scp deploy/systemd/llama-server.service $(JETSON_USER)@$(JETSON_HOST):/tmp/llama-server.service
	scp deploy/systemd/jetson-clocks.service $(JETSON_USER)@$(JETSON_HOST):/tmp/jetson-clocks.service
	ssh $(JETSON_USER)@$(JETSON_HOST) 'bash -s' < deploy/scripts/init.sh

deploy: build-arm64 ## Build + deploy binary, config, prompt, and service to Jetson
	scp dist/pollex-arm64 $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex
	scp deploy/config.yaml $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-config.yaml
	scp prompts/polish.txt $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-polish.txt
	scp deploy/systemd/pollex-api.service $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-api.service
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo mv /tmp/pollex /usr/local/bin/pollex && sudo chmod +x /usr/local/bin/pollex && sudo mv /tmp/pollex-config.yaml /etc/pollex/config.yaml && sudo mv /tmp/pollex-polish.txt /etc/pollex/polish.txt && sudo cp /tmp/pollex-api.service /etc/systemd/system/pollex-api.service && sudo systemctl daemon-reload && sudo systemctl restart pollex-api'

deploy-secrets: ## Deploy API key from dotfiles to Jetson
	@test -n "$$POLLEX_API_KEY" || (echo "Error: POLLEX_API_KEY not set. Run: secrets_refresh" && exit 1)
	@echo "Deploying secrets to Jetson..."
	@ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo mkdir -p /etc/pollex && echo "POLLEX_API_KEY='"$$POLLEX_API_KEY"'" | sudo tee /etc/pollex/secrets.env > /dev/null && sudo chmod 600 /etc/pollex/secrets.env'
	@echo "Secrets deployed. Restarting pollex-api..."
	@ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo systemctl restart pollex-api'
	@echo "Done."

deploy-llamacpp: ## Build llama.cpp with CUDA on Jetson (~85 min)
	scp deploy/scripts/build-llamacpp.sh $(JETSON_USER)@$(JETSON_HOST):/tmp/build-llamacpp.sh
	scp deploy/systemd/llama-server.service $(JETSON_USER)@$(JETSON_HOST):/tmp/llama-server.service
	ssh $(JETSON_USER)@$(JETSON_HOST) 'bash /tmp/build-llamacpp.sh'

deploy-tunnel: ## Setup Cloudflare Tunnel on Jetson (interactive)
	scp deploy/scripts/setup-cloudflared.sh $(JETSON_USER)@$(JETSON_HOST):/tmp/setup-cloudflared.sh
	scp deploy/systemd/cloudflared.service $(JETSON_USER)@$(JETSON_HOST):/tmp/cloudflared.service
	ssh -t $(JETSON_USER)@$(JETSON_HOST) 'bash /tmp/setup-cloudflared.sh'

# ─── Jetson Remote ──────────────────────────────────────────
.PHONY: jetson-ssh jetson-logs jetson-status jetson-test jetson-tunnel-start jetson-tunnel-status jetson-tunnel-logs

jetson-ssh: ## SSH into Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST)

jetson-logs: ## Tail pollex-api service logs on Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo journalctl -u pollex-api -f'

jetson-status: ## Remote health check (via SSH)
	@ssh $(JETSON_USER)@$(JETSON_HOST) 'curl -s localhost:$(API_PORT)/api/health' | python3 -m json.tool

jetson-test: ## Test polish request on Jetson (end-to-end)
	@ssh $(JETSON_USER)@$(JETSON_HOST) 'curl -s -X POST localhost:$(API_PORT)/api/polish -H "Content-Type: application/json" -d '"'"'{"text":"This is a test to see if pollex works end to end on the jetson nano.","model_id":"qwen2.5-1.5b-gpu"}'"'"'' | python3 -m json.tool

jetson-tunnel-start: ## Start Cloudflare Tunnel on Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo systemctl start cloudflared && sudo systemctl status cloudflared'

jetson-tunnel-status: ## Check Cloudflare Tunnel status
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo systemctl status cloudflared'

jetson-tunnel-logs: ## Tail Cloudflare Tunnel logs on Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo journalctl -u cloudflared -f'

# ─── Docker ────────────────────────────────────────────────
.PHONY: docker-build docker-dev docker-down

docker-build: ## Build pollex Docker image
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

loadtest-jetson: ## Run k6 load test against Jetson (single-user, via Cloudflare Tunnel)
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
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'
