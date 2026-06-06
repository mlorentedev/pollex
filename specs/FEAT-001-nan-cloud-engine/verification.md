---
tags: [spec, verification, templates]
created: "2026-06-05"
---

# Verification - FEAT-001-nan-cloud-engine

## Evidence

Map every acceptance criterion from `proposal.md` to concrete proof (commit hash, test name, or observed behavior). Fill during implementation.

- [ ] AC1 NousAdapter Bearer + enable_thinking:false -> test `<name>` / commit `<hash>`
- [ ] AC2 reasoning_content ignored -> test `<name>` / commit `<hash>`
- [ ] AC3 all 3 models work live (mimo/qwen/gemma) -> integration test `<name>` (+ raw curl output)
- [ ] AC4 FallbackChain fall-through -> test `<name>` / commit `<hash>`
- [ ] AC5 /api/models lists single "Nous Cloud (auto)" gated on key -> test `<name>`
- [ ] AC6 cross-browser polish via cloud engine -> manual smoke (Chrome/Edge/Brave) screenshots/notes

## Pre-implementation smoke test (2026-06-05)

Live against `https://api.nan.builders/v1` (key from `nan.api-key` age-secret):

- `GET /v1/models` → chat models present: `deepseek-v4-flash`, `mimo-v2.5`, `qwen3.6`, `gemma4` (+ non-chat `kokoro`/`whisper`/`qwen3-embedding`). Confirms `mimo-v2.5` exists (stale README corrected).
- Polish smoke (input `"i has went to the store yesterday and buyed two breads."`, `enable_thinking:false`, temp 0.3):
  | Model | HTTP | wall | reasoning_tokens | output |
  |---|---|---|---|---|
  | mimo-v2.5 | 200 | 3.17s | 0 | "I went to the store yesterday and bought two loaves of bread." |
  | qwen3.6 | 200 | 1.37s | 0 | (identical) |
  | gemma4 | 200 | 1.08s | 0 | (identical) |
- Conclusion: chain `mimo→qwen→gemma` valid; `enable_thinking:false` honored (no reasoning tokens); mimo ~2-3× slower than qwen/gemma on simple text.

## Test status

- Test suite: `source ~/.zshrc && go test -race ./... -> <output>`
- Integration (live gateway): `NAN_API_KEY=… go test -race -run TestNous_Integration ./internal/adapter/ -> <output>`
- Manual smoke test: extension cloud-engine polish in <browsers>, what was observed
- No regressions in existing suite: yes / no

## Decisions made during implementation

- (FallbackChain error policy — record the final `shouldFallback` classifier choice here)
-

## Promotion candidates

- [ ] Lesson for repo `docs/lessons.md`? <yes / no - one line> (e.g. NaN reasoning_content quirk, rate-limit sharing with Hermes)
- [ ] ADR-worthy for repo `docs/adr/adr-XXX.md`? <likely YES — "cloud inference engine + fixed fallback chain"> 
- [ ] New pattern for `00_meta/patterns/`? Only if it recurs in >1 project. <yes / no - one line>

## Archive checklist

- [ ] `proposal.md` frontmatter `status: archived`
- [ ] Folder moved: `specs/FEAT-001-nan-cloud-engine/` -> `specs/archive/FEAT-001-nan-cloud-engine/`
- [ ] Backlog entry in vault `11-tasks.md` ticked with PR link
- [ ] Promotions above executed (if any)
