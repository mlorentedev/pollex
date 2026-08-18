---
id: lesson-001-starlight-css-overrides-must-be-scoped-to-acc
type: lesson
status: active
created: "2026-03-04"
owner: manu
tags: [pollex, lesson, starlight, astro, css, accessibility, docs-site]
---

# Starlight CSS overrides must be scoped to accent variables only

**Context:** Setting up the Starlight docs site for pollex with a cyan-700 brand palette. Wanted to align the overall greys with a Slate-inspired look.

**Problem:** Overriding `--sl-color-white`, `--sl-color-black`, and `--sl-color-gray-1..6` to swap in a Slate palette broke contrast in **both** light and dark modes — sidebar text, code-block frames, callouts, and body copy all became low-contrast or invisible. Starlight's components reference these gray vars contextually (e.g., on hover, inside cards), so flipping them globally cascades through every component.

**Solution:** In `site/src/styles/custom.css`, keep only the three accent vars: `--sl-color-accent-low` (backgrounds), `--sl-color-accent` (links/buttons), `--sl-color-accent-high` (text on accent). Delete every `--sl-color-white/black/gray-*` override and every component-level border tweak that referenced them.

**Why:** Starlight's gray scale is calibrated as a coordinated pair (light + dark mode). The components don't just use them for surfaces — they use them for borders, hover states, and text-on-surface combinations. There is no safe way to swap individual gray tokens without breaking that calibration.

**Tags:** `#starlight` `#astro` `#css` `#accessibility` `#docs-site`

---
