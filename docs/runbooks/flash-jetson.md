---
id: "pollex-runbook-flash-jetson"
type: runbook
status: active
tags: [runbook, pollex]
created: "2026-02-10"
owner: manu
---

# Flash JetPack on Jetson Nano

## Overview

Complete from-zero setup of a Jetson Nano 4GB with JetPack OS, ready for Pollex deployment.

## What You Need

| Item | Notes |
|------|-------|
| Jetson Nano 4GB Developer Kit | B01 revision (2 CSI connectors) |
| MicroSD card | 64GB+ UHS-I recommended (slow cards = 1h+ first boot) |
| 5V 4A barrel jack power supply | **NOT** micro-USB (insufficient for inference load). **Requires J48 jumper** — see Step 3 |
| Ethernet cable OR WiFi dongle | Ethernet for KubeLab; WiFi dongle for office — see [setup-wifi-jetson](setup-wifi-jetson.md) |
| HDMI monitor + keyboard | Required for OEM setup (sshd not running until complete) |
| Host machine | Any OS with Raspberry Pi Imager, balenaEtcher, or `dd` |

## Step 1: Download JetPack Image

Download **JetPack 4.6.6** (the last supported version for Jetson Nano):

- Follow the official guide: [Get Started With Jetson Nano Developer Kit](https://developer.nvidia.com/embedded/learn/get-started-jetson-nano-devkit#intro)
- Or go directly to [NVIDIA JetPack Archive](https://developer.nvidia.com/embedded/jetpack-archive)
- Download the **SD Card Image** for Jetson Nano Developer Kit

> **Warning:** Do NOT use JetPack 5.x or 6.x — those require Jetson Orin or newer.

## Step 2: Flash the SD Card

### Option A: Raspberry Pi Imager (recommended)

1. Download [Raspberry Pi Imager](https://www.raspberrypi.com/software/)
2. Select "Use custom" (bajo "Choose OS") and pick the downloaded `.zip` file
3. Select the target SD card
4. Click "Write"

> **Note:** balenaEtcher used to be the recommended tools, but recent versions (v1.19+) have a known bug (`Error: h.requestMetadata is not a function`) that prevents flashing. Raspberry Pi Imager works reliably with any `.img` or `.zip` image, not just Raspberry Pi images.

### Option B: dd (Linux)

```sh
lsblk                           # Identify the SD card device
sudo umount /dev/sdX*           # Unmount if auto-mounted
unzip -p jetson-nano-jp466-sd-card-image.zip | sudo dd of=/dev/sdX bs=4M status=progress
sudo sync
```

## Step 3: First Boot

1. Insert the SD card into the Jetson Nano
2. **Place a jumper on J48 header pins** (next to the barrel jack). Without this jumper, the barrel jack is ignored and the board will not boot. A dupont female-female wire works if no jumper cap is available.
3. Connect Ethernet cable to the KubeLab switch (or WiFi dongle — see [setup-wifi-jetson](setup-wifi-jetson.md))
4. Connect HDMI monitor + keyboard
5. Connect the 5V barrel jack power supply
5. **Wait ~45 minutes** for the initial OEM installation to complete

> **Why so slow?** The first boot generates SSH keys, resizes the partition, and installs CUDA packages. SD card I/O is the bottleneck. This is normal — do not interrupt it.

### Complete OEM Setup (HDMI + keyboard required)

**Important:** sshd does NOT start until the OEM setup wizard is complete. You cannot skip this step.

When the wizard appears:

- Accept EULA
- Set locale: English, timezone as needed
- **Username:** `manu`
- **Hostname:** `jetson-home` (home/KubeLab) or `jetson-office` (office)
- Set password
- **Timezone:** as needed (`America/Denver` for office, `Europe/Madrid` for home)
- Let it finish (another 10-15 min of package configuration)

> **Note:** During config prompts like "Modified config file — keep or install new?", choose **Y** (install package maintainer's version).

## Step 4: Find the Jetson's IP

The Jetson is on the KubeLab LAN (192.168.2.0/24), behind the Proxmox router. Find its IP from Proxmox:

```sh
ssh pve 'cat /var/lib/misc/dnsmasq.leases'
```

The Jetson's MAC starts with `00:04:4b` (NVIDIA OUI). On first boot the hostname may show as `localhost` — this is normal.

## Step 5: Configure SSH Access

### 5.1 — Add to SSH config

Add to `~/.ssh/config` on your dev machine (inside the `CUBELAB-SSH` block):

```
Host jetson-home
  Hostname 192.168.2.59
  User manu
  ProxyJump pve
  IdentityFile ~/.ssh/id_ed25519
```

### 5.2 — Copy SSH key (removes password prompts)

```sh
ssh-copy-id jetson-home
```

### 5.3 — Verify passwordless access

```sh
ssh jetson-home 'echo ok'
```

### 5.4 — Configure passwordless sudo

This is required for `make deploy-setup` and `make deploy` to work non-interactively:

```sh
ssh jetson-home
sudo bash -c 'echo "manu ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/manu'
exit
```

## Step 6: System Updates

```sh
ssh jetson-home
sudo apt update && sudo apt upgrade -y
```

> **Critical:** Do NOT run `sudo apt dist-upgrade` or upgrade the kernel. This breaks CUDA drivers and JetPack compatibility.

## Step 7: CUDA Verification

CUDA is installed but not on PATH by default:

```sh
ssh jetson-home
echo 'export PATH=/usr/local/cuda/bin:$PATH' >> ~/.bashrc && source ~/.bashrc
nvcc --version
# Expected: Cuda compilation tools, release 10.2
```

> **Note:** The `install.sh` script also adds CUDA to PATH automatically.

## Step 8: Deploy Pollex

From your dev machine:

```sh
make deploy-setup    # First-time: installs curl, Ollama, pulls model, sets up systemd
make deploy          # Cross-compile + SCP binary + restart service
```

## Actual Setup Values

### Home Jetson (KubeLab, behind Proxmox bridge)

| Field | Value |
|-------|-------|
| Username | `manu` |
| Hostname | `jetson-home` |
| Timezone | Europe/Madrid |
| LAN IP (DHCP) | `192.168.2.59` |
| Jump host | `root@10.0.0.187` (pve-kubelab) |
| SSH shortcut | `ssh jetson-home` (via ProxyJump pve) |
| CUDA | 10.2, V10.2.300 |

### Office Jetson (direct WiFi, 24/7 production)

| Field | Value |
|-------|-------|
| Username | `manu` |
| Hostname | `jetson-office` |
| Timezone | America/Denver |
| WiFi dongle | TP-Link TL-WN725N (RTL8188EUS, auto-detected) |
| Network | Office WiFi, DHCP, client isolation active |
| SSH shortcut | `ssh jetson-office` (via Cloudflare Tunnel) |
| SSH DNS | `ssh-pollex.mlorente.dev` |
| Tunnel name | `pollex-office` |
| Tunnel protocol | `http2` (QUIC blocked by office firewall) |
| CUDA | 10.2, V10.2.300 |

## Post-Deploy Checklist

- [ ] SSH works: `ssh jetson-home`
- [ ] CUDA available: `ssh jetson-home 'nvcc --version'`
- [ ] Ollama running: `ssh jetson-home 'systemctl is-active ollama'`
- [ ] Model loaded: `ssh jetson-home 'curl -s localhost:11434/api/tags'`
- [ ] Pollex API running: `ssh jetson-home 'systemctl is-active pollex-api'`
- [ ] Health check passes: `make jetson-status`
- [ ] Polish works end-to-end: test from browser extension

## Lessons Learned

- **JetPack base image lacks `curl`** — `install.sh` installs it as a prerequisite
- **CUDA not on PATH by default** — must add `/usr/local/cuda/bin` to `~/.bashrc`
- **sshd not running on first boot** — OEM setup via HDMI+keyboard is mandatory
- **First boot takes ~45 min** on SD card — do not interrupt
- **WiFi dongle (TL-WN725N) worked out of the box** on JetPack 4.6.6 — no manual driver build needed. See [setup-wifi-jetson](setup-wifi-jetson.md)
- **Passwordless sudo required** — `make deploy-setup` runs `sudo` commands over SSH
- **`apt dist-upgrade` breaks CUDA** — only use `apt upgrade`
- **balenaEtcher v1.19+ broken** — use Raspberry Pi Imager instead (works with any image, not just RPi)
- **exFAT USB no letter on Windows** — if copying image via USB, Windows may not auto-assign a drive letter to exFAT-formatted drives. Fix: `diskmgmt.msc` > right-click > "Change Drive Letter"
- **J48 jumper required for barrel jack** — without the jumper the board ignores barrel jack power and won't boot. Most common "dead board" false alarm
- **CRLF breaks deploy from WSL** — files on Windows filesystem (`/mnt/c/`) have CRLF. Scripts fail with `invalid option namepefail`. Fix: `.gitattributes` forces LF, or `sed -i 's/\r$//'` manually
- **`llama-server.service` model filename** — both `build-llamacpp.sh` and `llama-server.service` now use `q4_0.gguf` (fixed Feb 2026). If upgrading from an older deployment, delete the old `q4_k_m.gguf` and re-run `make deploy-llamacpp`
