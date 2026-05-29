---
id: "pollex-troubleshoot-jetson-memory"
type: troubleshooting
status: active
tags: [troubleshooting, pollex]
created: "2026-02-10"
owner: manu
---

# Jetson Nano Memory Issues

The Jetson Nano 4GB shares RAM between CPU and GPU. Memory management is critical.

## Memory Budget

### With llama-server (Phase 9+, GPU inference)

| Component | RAM |
|-----------|-----|
| JetPack OS (headless) | ~500MB |
| llama-server + model (GPU layers) | ~1.2GB |
| Pollex Go API | ~15MB |
| **Total** | **~1.7GB** |
| **Free** | **~2.3GB** |

### With Ollama (fallback, CPU-only)

| Component | RAM |
|-----------|-----|
| JetPack OS (headless) | ~500MB |
| Ollama runtime | ~200MB |
| Active model (Qwen2.5-1.5B Q4) | ~1.0GB |
| Pollex Go API | ~15MB |
| **Total** | **~1.7GB** |
| **Free** | **~2.3GB** |

## Swap Configuration

The Jetson Nano does not have swap enabled by default. Adding swap provides a safety net for memory spikes during inference.

### Option A: Swap File (simple)

```sh
# Create 4GB swap file
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile

# Make persistent across reboots
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# Set swappiness low — prefer keeping data in RAM
echo 'vm.swappiness=10' | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

> **Note:** Swap on SD card will cause wear over time. Use `swappiness=10` to minimize swap usage — the kernel will only swap when truly necessary.

### Option B: ZRAM (better for SD card wear)

ZRAM creates a compressed swap device in RAM, reducing SD card writes:

```sh
# Install zram-config
sudo apt install -y zram-config

# Verify it's active
cat /proc/swaps
# Should show /dev/zram0 with ~2GB
```

ZRAM trades CPU cycles for memory compression. On the Jetson Nano, this is worthwhile because:
- SD card I/O is slow (~100MB/s sequential, much less random)
- The ARM CPU can compress/decompress faster than SD card I/O
- Reduces SD card wear (important for long-term headless operation)

### Verifying Swap Status

```sh
ssh jetson.local 'swapon --show && free -h'
```

## Common Issues

### OOM during inference

**Symptoms:** Ollama returns error, process killed, or Jetson becomes unresponsive.

**Possible causes:**
- Model too large for available memory
- Multiple models loaded simultaneously (Ollama keeps models in memory)
- OS services consuming unexpected memory

**Solutions:**
- Check loaded models: `curl http://jetson.local:11434/api/ps`
- Unload unused models: restart Ollama or use a smaller model
- Check system memory: `ssh jetson.local 'free -h'`
- Use Q4 quantization (not Q8) for all models

### Slow inference

**Symptoms:** Responses take 30s+ consistently.

**Possible causes:**
- GPU not being used (check CUDA availability)
- Swap thrashing
- Model too large, partially on CPU

**Solutions:**
- Verify CUDA: `ssh jetson.local 'nvcc --version'`
- Check swap usage: `ssh jetson.local 'swapon --show && free -h'`
- Downgrade to smaller model (Qwen2.5-0.5B) if necessary
