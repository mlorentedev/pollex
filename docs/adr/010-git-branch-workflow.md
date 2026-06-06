---
id: 010-git-branch-workflow
type: adr
status: active
tags: [adr, git, workflow, branch-protection]
created: "2026-06-06"
owner: manu
---

# ADR-010: Git Branch Workflow and Branch Protection

**Date:** 2026-06-06
**Status:** Active
**Context:** Master was unprotected, allowing direct pushes. Other projects in the ecosystem use PR-only workflows.

## Decision

All production changes to `master` must go through pull requests. No direct pushes to `master`.

### Branch Protection Rules

- **Required PR reviews**: 0 (single maintainer — no self-approval on GitHub)
- **Dismiss stale reviews**: Disabled
- **Force pushes**: Blocked
- **Branch deletion**: Blocked
- **Admin enforcement**: Yes (even maintainer goes through PRs)
- **Conversation resolution**: Required before merge

### Workflow

1. Create feature branch: `git checkout -b feat/<feature-name>` or `fix/<bug-name>`
2. Make changes, commit with conventional commits
3. Push branch: `git push origin <branch>`
4. Create PR: `gh pr create --base master`
5. CI runs (lint + test)
6. Squash merge to master

### Rationale for 0 required reviews

This repo has a single maintainer. GitHub blocks self-approval (`Review Can not approve your own pull request`), so requiring `required_approving_review_count: 1` would make PRs unmergeable. Other single-maintainer repos in the ecosystem (e.g. `hive`) use 0 required reviews. Multi-collaborator repos (e.g. `ts-bridge`) use 1.

### Branch Naming

- `feat/<description>` - new features
- `fix/<description>` - bug fixes
- `chore/<description>` - maintenance
- `docs/<description>` - documentation
- `refactor/<description>` - code refactoring

### Exception

Release-please automation (`release-please--branches--master`) is the only automated branch that bypasses this workflow.

## Consequences

- Slower iteration for single maintainer (trade-off for safety)
- All changes reviewed before reaching master
- Git history is clean and traceable
- Consistent with other projects in the ecosystem
- Prevents accidental direct pushes
