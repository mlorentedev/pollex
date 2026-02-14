# Pollex — Makefile
# Configurable variables
JETSON_HOST ?= nvidia
JETSON_USER ?= manu
API_PORT    ?= 8090

# ─── Development ────────────────────────────────────────────
.PHONY: dev dev-ollama test lint

dev: ## Start API with mock adapter on :$(API_PORT)
	go run ./cmd/pollex --mock --port $(API_PORT)

dev-ollama: ## Start API connected to local Ollama (Docker)
	go run ./cmd/pollex --port $(API_PORT)

test: ## Run all tests with race detector
	go test -v -race ./...

lint: ## Run go vet + check formatting
	go vet ./... && gofmt -l internal/ cmd/

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
	go build -o dist/pollex ./cmd/pollex

build-arm64: ## Cross-compile for ARM64 (Jetson Nano)
	GOOS=linux GOARCH=arm64 go build -o dist/pollex-arm64 ./cmd/pollex

ext-pack: ## Package extension into dist/pollex-ext.zip
	cd extension && zip -r ../dist/pollex-ext.zip . -x '*.gitkeep'

# ─── Deploy (Jetson) ────────────────────────────────────────
.PHONY: deploy deploy-setup deploy-secrets deploy-models deploy-llamacpp deploy-cloudflared

deploy: build-arm64 ## Build + deploy binary, config, prompt, and service to Jetson
	scp dist/pollex-arm64 $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex
	scp deploy/config.yaml $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-config.yaml
	scp prompts/polish.txt $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-polish.txt
	scp deploy/pollex-api.service $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-api.service
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo mv /tmp/pollex /usr/local/bin/pollex && sudo chmod +x /usr/local/bin/pollex && sudo mv /tmp/pollex-config.yaml /etc/pollex/config.yaml && sudo mv /tmp/pollex-polish.txt /etc/pollex/polish.txt && sudo cp /tmp/pollex-api.service /etc/systemd/system/pollex-api.service && sudo systemctl daemon-reload && sudo systemctl restart pollex-api'

deploy-secrets: ## Deploy API key from dotfiles to Jetson
	@test -n "$$POLLEX_API_KEY" || (echo "Error: POLLEX_API_KEY not set. Run: secrets_refresh" && exit 1)
	@echo "Deploying secrets to Jetson..."
	@ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo mkdir -p /etc/pollex && echo "POLLEX_API_KEY='"$$POLLEX_API_KEY"'" | sudo tee /etc/pollex/secrets.env > /dev/null && sudo chmod 600 /etc/pollex/secrets.env'
	@echo "Secrets deployed. Restarting pollex-api..."
	@ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo systemctl restart pollex-api'
	@echo "Done."

deploy-setup: ## First-time setup: install Ollama + models + systemd on Jetson
	scp deploy/pollex-api.service $(JETSON_USER)@$(JETSON_HOST):/tmp/pollex-api.service
	ssh $(JETSON_USER)@$(JETSON_HOST) 'bash -s' < deploy/install.sh

deploy-models: ## Pull/update models on Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'bash -s' < deploy/ollama-models.sh

deploy-llamacpp: ## Build llama.cpp with CUDA on Jetson (~85 min)
	scp deploy/build-llamacpp.sh $(JETSON_USER)@$(JETSON_HOST):/tmp/build-llamacpp.sh
	scp deploy/llama-server.service $(JETSON_USER)@$(JETSON_HOST):/tmp/llama-server.service
	ssh $(JETSON_USER)@$(JETSON_HOST) 'bash /tmp/build-llamacpp.sh'

deploy-cloudflared: ## Setup Cloudflare Tunnel on Jetson (interactive)
	scp deploy/setup-cloudflared.sh $(JETSON_USER)@$(JETSON_HOST):/tmp/setup-cloudflared.sh
	scp deploy/cloudflared.service $(JETSON_USER)@$(JETSON_HOST):/tmp/cloudflared.service
	ssh -t $(JETSON_USER)@$(JETSON_HOST) 'bash /tmp/setup-cloudflared.sh'

# ─── Utilities ──────────────────────────────────────────────
.PHONY: clean jetson-ssh jetson-logs jetson-status jetson-test tunnel-start tunnel-status tunnel-logs

clean: ## Remove dist/ directory
	rm -rf dist/

jetson-ssh: ## SSH into Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST)

jetson-logs: ## Tail pollex-api service logs on Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo journalctl -u pollex-api -f'

jetson-status: ## Remote health check (via SSH, works through jump host)
	@ssh $(JETSON_USER)@$(JETSON_HOST) 'curl -s localhost:$(API_PORT)/api/health' | python3 -m json.tool

jetson-test: ## Test polish request on Jetson (end-to-end)
	@ssh $(JETSON_USER)@$(JETSON_HOST) 'curl -s -X POST localhost:$(API_PORT)/api/polish -H "Content-Type: application/json" -d '"'"'{"text":"This is a test to see if pollex works end to end on the jetson nano.","model_id":"qwen2.5-1.5b-gpu"}'"'"'' | python3 -m json.tool

# ─── Tunnel ─────────────────────────────────────────────────
tunnel-start: ## Start Cloudflare Tunnel on Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo systemctl start cloudflared && sudo systemctl status cloudflared'

tunnel-status: ## Check Cloudflare Tunnel status
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo systemctl status cloudflared'

tunnel-logs: ## Tail Cloudflare Tunnel logs on Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo journalctl -u cloudflared -f'

# ─── Help ───────────────────────────────────────────────────
.DEFAULT_GOAL := help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
