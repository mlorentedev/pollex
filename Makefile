# Pollex — Makefile
# Variables configurables
JETSON_HOST ?= jetson.local
JETSON_USER ?= manu
API_PORT    ?= 8090

# ─── Desarrollo ──────────────────────────────────────────────
.PHONY: dev dev-ollama test lint

dev: ## Arranca API en modo mock (sin Ollama) en :$(API_PORT)
	cd backend && go run . --mock --port $(API_PORT)

dev-ollama: ## Arranca API conectado a Ollama local
	cd backend && go run . --port $(API_PORT)

test: ## Corre todos los tests del backend
	cd backend && go test -v -race ./...

lint: ## go vet + formato
	cd backend && go vet ./... && gofmt -l .

# ─── Build ───────────────────────────────────────────────────
.PHONY: build build-jetson ext-pack

build: ## Compila binario para tu máquina
	cd backend && go build -o ../dist/pollex .

build-jetson: ## Cross-compila para ARM64 (Jetson Nano)
	cd backend && GOOS=linux GOARCH=arm64 go build -o ../dist/pollex-arm64 .

ext-pack: ## Empaqueta extensión en dist/pollex-ext.zip
	cd extension && zip -r ../dist/pollex-ext.zip . -x '*.gitkeep'

# ─── Deploy a Jetson ─────────────────────────────────────────
.PHONY: deploy deploy-setup deploy-models

deploy: build-jetson ## Cross-compila + scp binario + config + restart servicio
	scp dist/pollex-arm64 $(JETSON_USER)@$(JETSON_HOST):/usr/local/bin/pollex
	scp deploy/config.yaml $(JETSON_USER)@$(JETSON_HOST):/etc/pollex/config.yaml
	scp prompts/polish.txt $(JETSON_USER)@$(JETSON_HOST):/etc/pollex/polish.txt
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo systemctl restart pollex-api'

deploy-setup: ## Primera vez: instala Ollama + modelos + systemd en Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'bash -s' < deploy/install.sh

deploy-models: ## Pull/actualiza modelos en la Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'bash -s' < deploy/ollama-models.sh

# ─── Utilidades ──────────────────────────────────────────────
.PHONY: clean jetson-ssh jetson-logs jetson-status

clean: ## Limpia dist/
	rm -rf dist/

jetson-ssh: ## SSH directo a la Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST)

jetson-logs: ## Tail de logs del servicio en la Jetson
	ssh $(JETSON_USER)@$(JETSON_HOST) 'sudo journalctl -u pollex-api -f'

jetson-status: ## Healthcheck remoto
	curl -s http://$(JETSON_HOST):$(API_PORT)/api/health

# ─── Help ────────────────────────────────────────────────────
.DEFAULT_GOAL := help
help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
