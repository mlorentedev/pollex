# Pollex — TODO

Phases 1–17 complete. Full task history → vault `_index.md` and `roadmap.md`.

## Phase 17 — Remaining

### Fase 7 — Cambios en el código del repo
- [x] 7.1 setup-cloudflared.sh — ya parametrizado (TUNNEL_NAME, LOCAL_PORT). SSH_HOSTNAME WON'T DO (SSH via headscale) ✓ 2026-02-22
- [x] 7.2 cloudflared.service — ahora usa `--config` file, eliminado User=manu y hardening incompatibles con JetPack 4.6 ✓ 2026-02-22
- [~] 7.3 Makefile TUNNEL_*/deploy-office/deploy-home — WON'T DO: office Jetson retirado, single-node architecture
- [x] 7.4 prometheus.yml — host: jetson-nano → kubelab-jet1 ✓ 2026-02-22
- [~] 7.5 alerts.yml — no hardcoded host references, no changes needed
- [~] 7.6 grafana pollex-dashboard.json — DEFER: dashboard genérico, no host-specific

### Fase 8 — Documentación (Knowledge Vault)
- [~] Crear `runbooks/setup-wifi-jetson.md` — DEFER: solo relevante si se redespliega en oficina
- [~] Actualizar `runbooks/flash-jetson.md` — DEFER: procedimiento sigue siendo válido
- [~] Actualizar `runbooks/setup-cloudflare-tunnel.md` — DEFER
- [~] Actualizar `architecture.md` — diagrama sigue siendo válido (single-node, Cloudflare Tunnel)
