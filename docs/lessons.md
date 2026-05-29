---
id: "pollex-lessons"
type: lesson
status: active
owner: manu
created: "2026-03-28"
---

# pollex: Lessons


### [2026-03-04] Starlight CSS overrides must be scoped to accent variables only

**Context:** Setting up the Starlight docs site for pollex with a cyan-700 brand palette. Wanted to align the overall greys with a Slate-inspired look.

**Problem:** Overriding `--sl-color-white`, `--sl-color-black`, and `--sl-color-gray-1..6` to swap in a Slate palette broke contrast in **both** light and dark modes — sidebar text, code-block frames, callouts, and body copy all became low-contrast or invisible. Starlight's components reference these gray vars contextually (e.g., on hover, inside cards), so flipping them globally cascades through every component.

**Solution:** In `site/src/styles/custom.css`, keep only the three accent vars: `--sl-color-accent-low` (backgrounds), `--sl-color-accent` (links/buttons), `--sl-color-accent-high` (text on accent). Delete every `--sl-color-white/black/gray-*` override and every component-level border tweak that referenced them.

**Why:** Starlight's gray scale is calibrated as a coordinated pair (light + dark mode). The components don't just use them for surfaces — they use them for borders, hover states, and text-on-surface combinations. There is no safe way to swap individual gray tokens without breaking that calibration.

**Tags:** `#starlight` `#astro` `#css` `#accessibility` `#docs-site`

---

### [2026-03-04] Light-mode Starlight buttons need accent ≥ 6:1 contrast vs white

**Context:** After fixing the gray-override bug, button text (white-on-accent) was still hard to read in light mode.

**Problem:** The brand color `#0e7490` (Tailwind cyan-700) has only ~4.5:1 contrast against white. WCAG AA says 4.5:1 is the minimum for *normal text*, but Starlight's button labels are visually small and the surrounding background is already light — the eye perceives this as borderline. Pure brand-color buttons looked "muddy".

**Solution:** Darken the light-mode accent to `#0b5e74` (~6:1 contrast vs white) and the accent-high to `#083344`. Use the original `#0e7490` only as the conceptual brand color (in copy, on the favicon gradient, etc.), not as the button background. In dark mode, *lighten* accent to `#2dd4e0` for the inverse reason — bright cyan reads cleanly on dark surfaces.

**Why:** Starlight composes `--sl-color-accent` as a solid button background with `--sl-color-white` text on top. The relevant contrast measurement is "white text on accent" (not accent-on-white). Anything below ~6:1 will look soft at the button's font weight, even though it passes raw AA. Dark mode flips the polarity — the accent becomes the *text* color, so it needs to be the lightest, not the deepest, end of the palette.

**Tags:** `#starlight` `#wcag` `#contrast` `#design-system` `#docs-site`
