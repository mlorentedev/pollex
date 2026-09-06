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

| Job | Does | Depends on |
|---|---|---|
| `lint` | `go vet ./...` + `gofmt -l` | — |
| `test` | `go test -v -race ./...` (80+ tests with subtests) | — |
| `test-extension` | Vitest unit tests + Playwright e2e against the mock API | `test` |
| `audit` | `npm audit --audit-level=high` on `site/` and `extension/` | — |
| `build` | `go build` for `linux/amd64` + `linux/arm64` | `lint`, `test` |
| `mutation` | `gremlins` on the `adapter` + `config` packages | `test` |

**`test-extension` gotcha:** the e2e suite needs a *headed* Chromium — Playwright's
headless shell cannot load browser extensions. The job therefore runs
`npx playwright install chromium --with-deps` (full browser + runner system libs)
and wraps the suite in `xvfb-run --auto-servernum`. See [lesson-068](../lessons/lesson-068-extension-e2e-in-ci-needs-a-headed-chromium-u.md).

**`audit` gate:** fails the PR on any high/critical advisory. Dependabot
(`.github/dependabot.yml`, weekly grouped updates) keeps it green; a red here means
a dependency needs upgrading before merge.

### Release (`release-please.yml`) — runs on merge to master

Two chained jobs:

Workflow-level `permissions: {}` — each job declares its own (least privilege).

**Job 1: release-please**
- Maintains a release PR with auto-generated changelog (conventional commits)
- On merge of the release PR: creates a GitHub tag + release
- Syncs extension version: `extra-files` with jsonpath on `extension/manifest.json`
- **Authenticates with `secrets.RELEASE_TOKEN`** — a repo-scoped fine-grained PAT
  (Contents + Pull requests: write), *not* the default `GITHUB_TOKEN`. Events made
  with `GITHUB_TOKEN` do not trigger other workflows, so CI would never run on the
  release PR and required status checks would block it forever. The PAT needs
  periodic rotation — see [lesson-069](../lessons/lesson-069-release-please-needs-a-pat-not-github-token-o.md).

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
