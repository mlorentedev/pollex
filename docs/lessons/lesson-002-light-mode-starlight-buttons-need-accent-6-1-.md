---
id: lesson-002-light-mode-starlight-buttons-need-accent-6-1-
type: lesson
status: active
created: "2026-03-04"
owner: manu
tags: [pollex, lesson, starlight, wcag, contrast, design-system, docs-site]
---

# Light-mode Starlight buttons need accent ≥ 6:1 contrast vs white

**Context:** After fixing the gray-override bug, button text (white-on-accent) was still hard to read in light mode.

**Problem:** The brand color `#0e7490` (Tailwind cyan-700) has only ~4.5:1 contrast against white. WCAG AA says 4.5:1 is the minimum for *normal text*, but Starlight's button labels are visually small and the surrounding background is already light — the eye perceives this as borderline. Pure brand-color buttons looked "muddy".

**Solution:** Darken the light-mode accent to `#0b5e74` (~6:1 contrast vs white) and the accent-high to `#083344`. Use the original `#0e7490` only as the conceptual brand color (in copy, on the favicon gradient, etc.), not as the button background. In dark mode, *lighten* accent to `#2dd4e0` for the inverse reason — bright cyan reads cleanly on dark surfaces.

**Why:** Starlight composes `--sl-color-accent` as a solid button background with `--sl-color-white` text on top. The relevant contrast measurement is "white text on accent" (not accent-on-white). Anything below ~6:1 will look soft at the button's font weight, even though it passes raw AA. Dark mode flips the polarity — the accent becomes the *text* color, so it needs to be the lightest, not the deepest, end of the palette.

**Tags:** `#starlight` `#wcag` `#contrast` `#design-system` `#docs-site`

---
