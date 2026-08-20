---
id: lesson-068-extension-e2e-in-ci-needs-a-headed-chromium-u
type: lesson
status: active
created: "2026-08-05"
owner: manu
tags: [pollex, lesson, ci, playwright, extension, e2e, github-actions]
---

# Extension e2e in CI needs a headed Chromium under `xvfb-run`

**Context:** Wiring the extension's Playwright e2e suite into `ci.yml` as a `test-extension` job (unit tests with Vitest, e2e against the mock API).

**Problem:** Playwright's default Chromium is the *headless shell* build, which cannot load browser extensions at all — `--load-extension` / `launchPersistentContext` silently gives you a context with no extension. Switching to a headed Chromium then fails on a clean GitHub runner: no display, and missing system libraries.

**Solution:** Three coupled steps in the job:
`npx playwright install chromium --with-deps` (installs the full Chromium *and* the system libs the runner lacks), then run the suite under `xvfb-run --auto-servernum npm run test:e2e` (virtual display; `--auto-servernum` picks a free display number so parallel jobs don't collide).

**Why:** Extension support is a full-browser feature — the headless shell is a stripped Chromium without the extension subsystem. Once headed is mandatory, a display becomes mandatory too, and CI runners have none. Any repo testing a browser extension end-to-end hits this same three-part fix.

**Tags:** `#ci` `#playwright` `#extension` `#e2e` `#github-actions`

---
