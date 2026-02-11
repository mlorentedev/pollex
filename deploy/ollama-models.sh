#!/bin/bash
set -euo pipefail

echo "=== Pulling/updating Ollama models ==="

# Primary model — best quality/resource balance for 4GB
ollama pull qwen2.5:1.5b

# Optional: smaller model for faster responses
# ollama pull qwen2.5:0.5b

echo "=== Models ready ==="
ollama list
