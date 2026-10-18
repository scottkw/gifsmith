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
