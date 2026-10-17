# Phase 2 — UI Review

**Audited:** 2026-10-17
**Baseline:** 02-UI-SPEC.md (approved contract)
**Screenshots:** not captured fresh. The audit used the existing 02-06 harness captures in `~/.dev-browser/tmp/gifsmith-06-*.png` (E5f dark, K6 light, Clips C2 light), which were viewed. No servers were started.
**Interaction captures:** off (workflow.ui_interaction_capture not set). All interaction findings come from reading the code.

Advisory review. These accepted deviations are not scored against the phase: the setup card is 560px wide, Klipy ads carry no label, and Clips Try Again does nothing after the one re-probe.

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 3/4 | Strings match the contract word for word, except C2. C2 now offers an email CTA, and the spec was never updated to match |
| 2. Visuals | 3/4 | Each state has one clear focal point. A disabled Clips tab that is also active shows a dim label on the active pill, so it reads as half-selected |
| 3. Color | 3/4 | The tokens match the spec exactly. Accent focus rings are missing on state-message buttons, the paging link and the footer button |
| 4. Typography | 4/4 | 4 sizes and 2 weights, all set through tokens. No stray sizes |
| 5. Spacing | 3/4 | Every value is on the 4/8/16/24/32 scale. The setup card stacks heading, body and steps differently from the contract |
| 6. Experience Design | 3/4 | All states are handled. The footer's `Enable Automatic Paste…` button has no styling, and the Try Again text link has no hover or focus state |

**Overall: 19/24**

---

## Top 3 Priority Fixes

1. **WARNING: Footer `Enable Automatic Paste…` button is unstyled** (`components/Footer.svelte`, `button { font: var(--label); }` is its only rule). The button renders with the default WebKit bezel inside a 32px footer, the one native-looking control in the panel. That breaks the visual system right where the user has to act (granting Accessibility). Fix: give it the secondary-button look (`--surface-active` background, `--text`, radius 6px, no border), about 24px tall with 8px horizontal padding, plus the accent `:focus-visible` ring.
2. **WARNING: Accent `:focus-visible` ring is missing outside SetupCard** (spec accent item 4). `StateMessage.svelte` buttons and links (Try Again, Add GIPHY Key, Email…, Back to GIPHY), `ResultGrid.svelte .link` and the footer button all fall back to the WebKit default ring, or show none. Keyboard users get a different focus treatment on each card. Fix: move the `:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px }` rule into `app.css` for `button, a, input`, and delete the copy in SetupCard.
3. **WARNING: C2 copy and CTA differ from the contract, and the spec was not updated.** The screenshot shows `Email clips@giphy.com ↗` and the body says "Email clips@giphy.com to request access". The spec says `Open GIPHY Dashboard ↗` with dashboard copy. 02-06-SUMMARY records this as an intended change, but 02-UI-SPEC.md still has the old copy, so later audits and tests will flag it again. Fix: update the C2 row in 02-UI-SPEC.md, and its `Accent reserved for` item 2, to the email CTA.

---

## Detailed Findings

### Pillar 1: Copywriting (3/4)
- These strings match the spec exactly: N1 `No GIFs found`, N2 `Search The Office GIFs` (`App.svelte:504,533`), `End of results` and `Couldn't load more.` (`ResultGrid.svelte:13,83`), E7 `COPY_FAILED` (`App.svelte:36`), K3 (`lib/errors.ts:92`), E5f heading and body (screenshot), K6 (screenshot) and the Clips tooltip (`ProviderTabs.svelte:17`). There are no generic labels (Submit, OK, Click here).
- WARNING: C2 copy differs from the spec (see fix 3).
- WARNING: `lib/errors.ts:84` and `App.svelte:38` add strings the spec does not contain (Keychain save failure, Accessibility request failure). The copy is fine. List them in the copywriting table so the contract stays complete.
- Note: K6 just fits in the one-line inline slot at 560px. A longer provider name, or a localized string later, will be cut off with an ellipsis (`.message` has `nowrap; text-overflow: ellipsis`). Those lines are errors the user needs to read in full.

### Pillar 2: Visuals (3/4)
- Every captured state has one clear anchor. E5f: icon, 18/600 heading, single Try Again. K6: key field with a destructive border plus Save Key. C2: one accent-filled primary next to a secondary.
- WARNING: `ProviderTabs.svelte` has `button.off { color: var(--text-tertiary) }`. That rule wins even when the button is also `.active`, so the active Clips pill in the C2 screenshot shows a gray label on a gray pill and looks disabled rather than selected. The spec asks for a `--text` label on the active tab. Fix: scope the rule to `button.off:not(.active)`.
- Note: the K6 screenshot shows the search input holding the stale query "revoked" while it is disabled behind the setup card. This follows the spec (the query persists), but a disabled input showing old text reads oddly. Consider dimming it.
- Icons: the magnifier, alert, check and broken-tile icons are all `aria-hidden`. The spinners have `role=status aria-label="Loading"`. There are no icon-only buttons without labels.

### Pillar 3: Color (3/4)
- `app.css` tokens match the spec table hex for hex in both themes. Accent is used only for the tile ring, the primary fill, links, the caret and SetupCard focus, which is the reserved list. Tabs, spinners and attribution use no accent.
- Hardcoded colors: `#ffffff` on primary buttons (`SetupCard`, `StateMessage`), and `rgba(0,0,0,0.4)` with `#fff` for the copy overlay (`Tile.svelte`). All of these values come from the spec. Acceptable.
- WARNING: the accent focus ring is applied unevenly (fix 2).
- Note: disabled Save Key uses `opacity: 0.5` (K6 screenshot), which puts white text on a faded `#0068D9` at about 2:1. That is allowed for a disabled control, but the button looks like a weak primary rather than clearly unavailable. Consider the `--surface-active` / `--text-tertiary` pair for the disabled state.

### Pillar 4: Typography (4/4)
- Sizes in use: 12 (`--label`), 14 (`--body`, plus `600 14px/1` for buttons), 18 (`--heading`) and 22 (`--display`). Weights in use: 400 and 600. Tab labels stay at 400 when active, so tab width does not change. The footer hint glyphs force `font-weight: 400` on `<b>`.
- Minor: buttons use line-height `1` and the key field uses `400 14px/1 mono` instead of the spec's 1.5. This is harmless inside fixed 32px boxes.

### Pillar 5: Spacing (3/4)
- Search bar 48/16/8, tabs 36 / 8 padding / 4 gap / 28px pills with 8px padding, footer 32/16/16 gap, grid `repeat(4,166px)` × 144 with an 8px gap and 16px padding, state message padding-top 32 with an 8/8/16 stack. All match the spec.
- WARNING: in `SetupCard.svelte` the heading's bottom margin is 16px, the body is followed by `.steps` with `margin: 8px 0 16px`, and `.message` has `margin: 8px 0`. The spec says 16px between blocks and 4px between steps. The result: the body-to-steps gap is 8px while the heading-to-body gap is 16px, so the body reads as attached to the steps instead of to the heading. It sits on the scale but breaks the "gap between card blocks = 16" rule. This was probably done to fit the card above the footer. Record it next to the 560px deviation, or fix it.
- No 12px spacing values appear (the 12px values in the grep are icon sizes and fonts). No arbitrary px values sit off the scale.

### Pillar 6: Experience Design (3/4)
- State coverage in code: S1/S1a/S1b/S13 (SetupCard, `aria-invalid`, `role=alert` message), S2 skeleton with reduced-motion handling (`App.svelte` `.skeleton`), S4 busy spinner, S5/S6/S9a/S12a end row (`ResultGrid.svelte`), S7/S8/S9–S12 StateMessage, C1/C2, copy overlay, broken tile icon. Combobox/listbox ARIA is wired (`App.svelte:445-448`, `Tile.svelte:30-33`). Save Key is disabled while the field is empty.
- WARNING: the footer Accessibility button is unstyled (fix 1).
- WARNING: `ResultGrid.svelte .link` (Try Again in the paging row) has no hover underline and no focus ring. The spec says text links underline on hover, and SetupCard's `a:hover` does this.
- Note: the reduced-motion spinner swaps in `•••` but keeps `display:inline-block` with its size removed. This was not verified visually.
- Registry audit: no `components.json`, and the spec lists no third-party registries. Skipped.

---

## Files Audited
- frontend/src/app.css
- frontend/src/App.svelte
- frontend/src/components/Footer.svelte, ProviderTabs.svelte, ResultGrid.svelte, SetupCard.svelte, StateMessage.svelte, Tile.svelte
- frontend/src/lib/errors.ts (copy strings)
- ~/.dev-browser/tmp/gifsmith-06-backstop-E5f-dark.png, gifsmith-06-backstop-K6-light.png, gifsmith-06-clips-c2.png
