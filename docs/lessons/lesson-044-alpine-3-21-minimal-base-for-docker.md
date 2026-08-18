---
id: lesson-044-alpine-3-21-minimal-base-for-docker
type: lesson
status: active
created: "2026-05-01"
owner: manu
tags: [pollex, lesson, docker, alpine]
---

# `alpine:3.21` minimal base for Docker

**Context:** Containerizing the Go binary.

**Problem:** `scratch` is too minimal — lacks `curl` for health checks and `/etc/ssl/certs` for HTTPS.

**Solution:** Use `alpine:3.21` (24.7MB final image).

**Tags:** `#docker` `#alpine`

---
