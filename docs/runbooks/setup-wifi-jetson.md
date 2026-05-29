---
id: "pollex-runbook-wifi-jetson"
type: runbook
status: active
tags: [runbook, pollex]
created: "2026-02-17"
owner: manu
---

# Setup WiFi on Jetson Nano (TP-Link TL-WN725N)

## Overview

Configure WiFi connectivity on a Jetson Nano 4GB using a TP-Link TL-WN725N USB dongle (RTL8188EUS chipset). Tested on JetPack 4.6.6 (L4T R32, Ubuntu 18.04, kernel 4.9).

## Hardware

| Item | Detail |
|------|--------|
| Dongle | TP-Link TL-WN725N v3 |
| Chipset | Realtek RTL8188EUS |
| USB ID | `0bda:8179` |
| Interface | 802.11n, 2.4 GHz only |

## Driver Status

**JetPack 4.6.6 recognized the TL-WN725N out of the box** during the February 2026 office Jetson setup. The kernel's built-in `rtl8xxxu` staging driver was sufficient — no out-of-tree driver compilation was needed.

> **Note:** Prior research indicated the driver would NOT be included and that the `aircrack-ng/rtl8188eus` out-of-tree driver would be required. This turned out to be unnecessary for JetPack 4.6.6. If a future JetPack version drops the staging driver, see the "Manual Driver Build" section below.

### Verify Driver

```sh
lsusb | grep -i realtek
# Expected: Bus 001 Device 003: ID 0bda:8179 Realtek Semiconductor Corp. RTL8188EUS

ip link show
# Look for wlan0 or wlxXXXXXXXXXXXX

lsmod | grep rtl
# Expected: rtl8xxxu or 8188eu
```

## WiFi Configuration (NetworkManager)

JetPack 4.6 ships with NetworkManager. Use `nmcli` to connect:

```sh
# Scan for available networks
nmcli device wifi list

# Connect to WiFi
nmcli device wifi connect "YOUR_SSID" password "YOUR_PASSWORD"

# Verify
ip addr show wlan0
ping -c 3 8.8.8.8
```

### Auto-Connect on Boot

NetworkManager saves the connection profile with `autoconnect yes` by default. Verify and enforce:

```sh
# Check autoconnect
nmcli connection show "YOUR_SSID" | grep autoconnect

# Force if needed
nmcli connection modify "YOUR_SSID" connection.autoconnect yes
nmcli connection modify "YOUR_SSID" connection.autoconnect-priority 100
```

### Verify After Reboot

```sh
sudo reboot
# Wait 1-2 minutes for WiFi to reconnect
# Then from another machine on the same network (if reachable):
ping <jetson-ip>
```

## Manual Driver Build (Fallback)

Only needed if the built-in `rtl8xxxu` driver does NOT work (e.g., future JetPack version or different dongle revision).

### Prerequisites

Network access required (use USB tethering from phone or temporary Ethernet).

```sh
sudo apt update
sudo apt install -y build-essential dkms git bc
```

### Check Kernel Headers

```sh
ls /lib/modules/$(uname -r)/build
# If missing, prepare from L4T source (see below)
```

If headers are missing, download from NVIDIA:

```sh
# Check L4T version
head -1 /etc/nv_tegra_release

# Download matching source from:
# https://developer.nvidia.com/embedded/jetpack-archive
# Extract kernel_src.tbz2, then:
cd kernel/kernel-4.9
zcat /proc/config.gz > .config
make modules_prepare
sudo mkdir -p /lib/modules/$(uname -r)/build
sudo cp -a . /lib/modules/$(uname -r)/build/
```

### Blacklist Conflicting Drivers

```sh
sudo tee /etc/modprobe.d/blacklist-rtl8188.conf << 'EOF'
blacklist r8188eu
blacklist rtl8xxxu
EOF
```

### Build aircrack-ng/rtl8188eus

```sh
cd /tmp
git clone https://github.com/aircrack-ng/rtl8188eus.git
cd rtl8188eus

# Set ARM64 platform
sed -i 's/CONFIG_PLATFORM_I386_PC = y/CONFIG_PLATFORM_I386_PC = n/' Makefile
sed -i 's/CONFIG_PLATFORM_ARM64_RPI = n/CONFIG_PLATFORM_ARM64_RPI = y/' Makefile

# Install via DKMS
VER=$(grep 'PACKAGE_VERSION=' dkms.conf | cut -d= -f2 | tr -d '"')
sudo cp -r /tmp/rtl8188eus /usr/src/rtl8188eus-${VER}
sudo dkms add rtl8188eus/${VER}
sudo dkms build rtl8188eus/${VER}
sudo dkms install rtl8188eus/${VER}

# Load
sudo modprobe 8188eu
```

## Troubleshooting

### No wlan0 interface after plugging dongle

1. Check `lsusb` — if dongle not listed, try different USB port
2. Check `dmesg | tail -20` — look for driver loading errors
3. If `rtl8xxxu` loaded but interface unstable, try the manual driver build above

### WiFi connects but no internet

- Office firewalls may block traffic. Test: `curl -s https://cloudflare.com/cdn-cgi/trace`
- Check DNS: `ping 8.8.8.8` works but `ping google.com` fails → DNS issue
- Fix DNS: `nmcli connection modify "YOUR_SSID" ipv4.dns "8.8.8.8 1.1.1.1"`

### Client isolation (cannot SSH from same WiFi)

Corporate/office WiFi often enables AP isolation (blocks device-to-device traffic). This is expected — use [Cloudflare Tunnel for SSH](setup-cloudflare-ssh.md) instead of direct SSH.

### WiFi does not reconnect after reboot

1. Verify: `nmcli connection show` — profile should exist
2. Check: `nmcli connection show "YOUR_SSID" | grep autoconnect` — should be `yes`
3. Driver not loading? Check: `lsmod | grep rtl` and `dmesg | grep -i rtl`

## Lessons Learned

- **JetPack 4.6.6 includes RTL8188EUS support** via `rtl8xxxu` staging driver — no manual driver build needed (as of Feb 2026)
- **Do NOT assume the driver is missing** — plug in the dongle and check `ip link` before attempting manual driver compilation
- **Office WiFi client isolation** blocks SSH between devices on the same network — not a Jetson issue, use Cloudflare Tunnel for remote access
- **JetPack base image lacks `curl`** — install it: `sudo apt update && sudo apt install -y curl`
