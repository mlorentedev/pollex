# Pollex — TODO

Phases 1–17 complete. Full task history → vault `_index.md` and `roadmap.md`.

## Phase 17 — Pending

### Fase 7 — Cambios en el código del repo
- [ ] 7.1 Parametrizar `deploy/scripts/setup-cloudflared.sh` (TUNNEL_NAME, TUNNEL_HOSTNAME, SSH_HOSTNAME, LOCAL_PORT)
- [ ] 7.2 Parametrizar `deploy/systemd/cloudflared.service` (ExecStart usa config file)
- [ ] 7.3 Añadir variables y targets al `Makefile` (TUNNEL_*, deploy-office, deploy-home)
- [ ] 7.4 Actualizar `deploy/prometheus/prometheus.yml` (host: jetson-office)
- [ ] 7.5 Actualizar `deploy/prometheus/alerts.yml` (host label en annotations)
- [ ] 7.6 Actualizar `deploy/grafana/pollex-dashboard.json` (variable template host + filtros)
- [ ] Verificar: `make test`, `make monitoring-validate`

### Fase 8 — Documentación (Knowledge Vault)
- [ ] Crear `runbooks/setup-wifi-jetson.md`
- [ ] Actualizar `runbooks/flash-jetson.md` (variante WiFi)
- [ ] Actualizar `runbooks/setup-cloudflare-tunnel.md` (parametrizado, multi-ingress)
- [ ] Actualizar `architecture.md` (diagrama actualizado: single-node + headscale)
