---
id: "pollex-runbook-cicd"
type: runbook
status: active
tags: [runbook, pollex]
created: "2026-02-22"
owner: manu
---

# CI/CD — Pollex

## Pipelines

### CI (`ci.yml`) — runs on every push/PR

Triggers: push to `master`, any PR.

Steps:
1. `go vet ./...` + `gofmt -l` (lint)
2. `go test -v -race ./...` (80+ tests with subtests)
3. `go build` for `linux/amd64` + `linux/arm64`

### Release (`release-please.yml`) — runs on merge to master

Two chained jobs:

**Job 1: release-please**
- Maintains a release PR with auto-generated changelog (conventional commits)
- On merge of the release PR: creates a GitHub tag + release
- Syncs extension version: `extra-files` with jsonpath on `extension/manifest.json`

**Job 2: goreleaser** (triggered by release-please tag creation)
- Builds binaries: `linux/amd64` + `linux/arm64`
- Creates GitHub release assets: `pollex_linux_amd64.tar.gz`, `pollex_linux_arm64.tar.gz`
- Packages extension zip: `pollex-ext-vX.Y.Z.zip`
- Changelog generated from conventional commits via `.goreleaser.yml`

## Conventional Commits

Used for automatic changelog generation:

| Prefix | Changelog section |
|--------|------------------|
| `feat:` | Features |
| `fix:` | Bug Fixes |
| `chore:` | Chores (no user-facing changelog) |
| `docs:` | Documentation |
| `refactor:` | Refactoring |

## Release Flow

```
Developer push → CI (lint + test + build)
    ↓
release-please PR auto-updated (changelog accumulates)
    ↓
Merge release PR → tag vX.Y.Z created + GitHub release draft
    ↓
goreleaser job triggered by tag → binaries + extension zip attached
```

## Versioning

- Managed by `release-please` (semver, conventional commits)
- Extension `manifest.json` version synced automatically via `extra-files` config
- Binary version injected at build: `-ldflags "-X main.version=vX.Y.Z"`
- Exposed in `/api/health` response (`version` field) + extension Settings page

## Key Files

```
.github/workflows/ci.yml          # Lint + test + build
.github/workflows/release-please.yml  # release-please + goreleaser chain
.goreleaser.yml                   # Binary + extension zip config
release-please-config.json        # release-please package config
.release-please-manifest.json     # Current version tracking
```

## Gotchas

- **goreleaser requires a tag event**: The goreleaser job is chained to release-please via
  `needs: release-please` and only runs when `steps.release.outputs.release_created == 'true'`.
  Direct tag pushes do NOT trigger it (GITHUB_TOKEN tag events don't trigger other workflows).

- **Extension version sync**: `extra-files` in `release-please-config.json` uses jsonpath
  to update `extension/manifest.json`. If the jsonpath is wrong, the extension version
  won't match the release tag.
