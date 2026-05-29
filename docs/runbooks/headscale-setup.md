---
id: headscale-setup
type: runbook
status: active
created: "2026-05-11"
---

# Headscale Setup for KubeLab Mesh

> **Status:** Stub — content pending. Referenced from [deploy-jetson](deploy-jetson.md).

## Overview

`kubelab-jet1` (Jetson Nano) joins the KubeLab headscale mesh at `100.64.0.8`. SSH access from any peer uses `ssh jet1`.

## Setup

TODO: Document headscale node enrollment, key exchange, and `tailscale up` invocation.

## Incident: jet1-dns-boot

The Jetson runs `systemd-resolved`. If tailscale logs out, a DNS circular dependency prevents reconnection on boot.

**Fix (already applied on jet1):**

```sh
sudo tee /etc/systemd/resolved.conf.d/fallback-dns.conf <<'EOF'
[Resolve]
FallbackDNS=1.1.1.1
EOF

sudo systemctl restart systemd-resolved
```

After applying, `tailscale up` succeeds on reboot even when the mesh DNS is unreachable.

## Related

- [Deploy to Jetson](deploy-jetson.md)
- [Cloudflare SSH (fallback)](setup-cloudflare-ssh.md)
