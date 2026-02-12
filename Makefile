# Pollex — Makefile
# Configurable variables
JETSON_HOST ?= nvidia
JETSON_USER ?= manu
API_PORT    ?= 8090

# ─── Development ────────────────────────────────────────────
.PHONY: dev dev-ollama test lint

dev: ## Start API with mock adapter on :$(API_PORT)
	cd backend && go run . --mock --port $(API_PORT)

dev-ollama: ## Start API connected to local Ollama (Docker)
	cd backend && go run . --port $(API_PORT)

test: ## Run all backend tests with race detector
	cd backend && go test -v -race ./...

lint: ## Run go vet + check formatting
	cd backend && go vet ./... && gofmt -l .

# ─── Local Ollama (Docker) ──────────────────────────────────
.PHONY: ollama-up ollama-down ollama-pull

ollama-up: ## Start Ollama in Docker (:11434) + pull model
	@docker inspect ollama >/dev/null 2>&1 && docker start ollama || \
		docker run -d --name ollama -p 11434:11434 -v ollama:/root/.ollama ollama/ollama
	@echo "Waiting for Ollama to start..."
	@until curl -sf http://localhost:11434/ >/dev/null 2>&1; do sleep 1; done
	@echo "Ollama ready. Checking model..."
	@docker exec ollama ollama list | grep -q 'qwen2.5:1.5b' || \
		docker exec ollama ollama pull qwen2.5:1.5b
	@echo "Ollama running on :11434 with qwen2.5:1.5b"

ollama-down: ## Stop and remove Ollama container
	@docker stop ollama 2>/dev/null; docker rm ollama 2>/dev/null; echo "Ollama stopped"

ollama-pull: ## Pull/update model in local Ollama (Docker)
	docker exec ollama ollama pull qwen2.5:1.5b

# ─── Build ──────────────────────────────────────────────────
.PHONY: build build-arm64 ext-pack

build: ## Build binary for current platform
	cd backend && go build -o ../dist/pollex .

build-arm64: ## Cross-compile for ARM64 (Jetson Nano)
	cd backend && GOOS=linux GOARCH=arm64 go build -o ../dist/pollex-arm64 .

ext-pack: ## Package extension into dist/pollex-ext.zip
	cd extension && zip -r ../dist/pollex-ext.zip . -x '*.gitkeep'

# ─── Deploy (Jetson) ────────────────────────────────────────
.PHONY: deploy deploy-setup deploy-models

deploy: build-arm64 ## Build + deploy binary, config, and prompt to Jetson
	scp dist/pollex-arm64 $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex
	scp deploy/config.yaml $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-config.yaml
	scp prompts/polish.txt $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-polish.txt
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo mv /tmp/pollex /usr/local/bin/pollex && sudo chmod +x /usr/local/bin/pollex && sudo mv /tmp/pollex-config.yaml /etc/pollex/config.yaml && sudo mv /tmp/pollex-polish.txt /etc/pollex/polish.txt && sudo systemctl restart pollex-api'

deploy-setup: ## First-time setup: install Ollama + models + systemd on Jetson
	scp deploy/pollex-api.service $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-api.service
	ssh $(JETSON_USER)@$(JETSON_HOST) 'bash -s' < deploy/install.sh

deploy-models: ## Pull/update models on Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'bash -s' < deploy/ollama-models.sh

# ─── Utilities ──────────────────────────────────────────────
.PHONY: clean jetson-ssh jetson-logs jetson-status jetson-test

clean: ## Remove dist/ directory
	rm -rf dist/

jetson-ssh: ## SSH into Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST)

jetson-logs: ## Tail pollex-api service logs on Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo journalctl -u pollex-api -f'

jetson-status: ## Remote health check (via SSH, works through jump host)
	@ssh $(JETSON_USER)@$(JETSON_HOST) 'curl -s localhost:$(API_PORT)/api/health' | python3 -m json.tool

jetson-test: ## Test polish request on Jetson (end-to-end)
	@ssh $(JETSON_USER)@$(JETSON_HOST) 'curl -s -X POST localhost:$(API_PORT)/api/polish -H "Content-Type: application/json" -d '"'"'{"text":"This is a test to see if pollex works end to end on the jetson nano.","model_id":"qwen2.5:1.5b"}'"'"'' | python3 -m json.tool

# ─── Help ───────────────────────────────────────────────────
.DEFAULT_GOAL := help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
