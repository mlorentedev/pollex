---
id: adr-010-git-branch-workflow
type: adr
status: active
tags: [git, workflow, branch-protection]
---

# ADR-010: Git Branch Workflow and Branch Protection

**Date:** 2026-06-06
**Status:** Active
**Context:** Master was unprotected, allowing direct pushes. Other projects in the ecosystem use PR-only workflows.

## Decision

All production changes to `master` must go through pull requests. No direct pushes to `master`.

### Branch Protection Rules

- **Required PR reviews**: 1 approving review
- **Dismiss stale reviews**: On new push
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
6. At least 1 review/approval
7. Squash merge to master

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
