---
title: Phase 3 UI review polish (contrast, live region, mouse paths)
created: 2026-10-18
resolves_phase: 4
source: .planning/phases/03-action-pipeline-library/03-UI-REVIEW.md
---

Advisory fixes from the Phase 3 UI review (19/24). Fold into Phase 4's panel/settings work.

1. Highlighted destructive menu rows are 4.29:1 / 4.02:1 — `.row.hi.destructive .label { color: var(--text) }` in `frontend/src/components/ActionMenu.svelte` (or adjust the token).
2. Footer toasts not announced — `role="status" aria-live="polite"` on `.center` in `frontend/src/components/Footer.svelte`.
3. T5a "copied, enable automatic paste" text truncates (~600px into ~150px) — shorten Go message at `internal/services/actions.go` (~161) to "Copied — press ⌘V to paste".
4. Mouse paths: right-click a tile opens the ⌘K menu; make the Details footer "Esc Back" a button.
5. Minor: `--text-tertiary` 3.26:1 (and chip clock icon ~2.8:1); hide library footer hints on empty tabs; menu over bottom row covers footer attribution; Clips F3 copy "Try another GIF."; toast icon 4px gap; badges use `--label`.

## Closed (plan 04-03)

Closed by plan 04-03 against the 04-UI-SPEC "Phase 3 polish" rows:

1. Highlighted destructive rows: P12 (`.row.hi.destructive .label { color: var(--text) }`).
2. Footer toasts announced: P13 (`role="status" aria-live="polite"` on the footer centre).
3. T5a too long: P14 (Go message is `Copied — press ⌘V to paste`; the Enable Automatic Paste… button carries the explanation in its `title`).
4. Mouse paths: P16 (the Details `Esc Back` hint is a button). Right-click a tile for the ⌘K menu is P15, owned by plan 04-04.
5. Minor items: P17 (`--text-tertiary` #737378 / #98989D), P18 (library hints only with a selection), P20 (Clips F3 ends "Try another clip."), P21 (toast icon 4px margin), P22 (badges use `--label`). The menu covering the footer attribution is P19, owned by plan 04-04.

Also folded in: P23 (T7c at the 1,000-favorite cap), P24 (IN-06 stale Undo toast cleared), P25 (IN-03: F11 for text copy/paste failures, F12 for `invalid_item`), and Reduce Motion stills (D-12).
