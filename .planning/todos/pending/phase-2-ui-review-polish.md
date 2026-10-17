---
title: Phase 2 UI review polish (focus ring, footer button, spec copy)
created: 2026-10-17
resolves_phase: 3
source: .planning/phases/02-providers-search-ui-first-run/02-UI-REVIEW.md
---

Advisory fixes from the Phase 2 UI review (19/24). Fold into Phase 3's panel work.

1. Style the footer "Enable Automatic Paste…" button (`frontend/src/components/Footer.svelte`) with the spec's secondary-button style.
2. Move the accent focus ring into `frontend/src/app.css` for `button, a, input` so StateMessage, the ResultGrid Try Again link and the footer button get it (today only SetupCard has it).
3. Update 02-UI-SPEC.md C2 copy to "Email clips@giphy.com ↗" (deliberate 02-06 decision).
4. Minor: active-but-unavailable Clips tab looks disabled (`button.off:not(.active)`); setup card 8px vs 16px block gap; paging Try Again needs hover/focus; disabled Save Key style; long key errors truncate at 560px; add the Keychain-save and Accessibility-request error strings to the copy contract.
