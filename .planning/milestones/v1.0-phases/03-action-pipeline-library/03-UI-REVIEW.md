# Phase 3 — UI Review

**Audited:** 2026-10-18
**Baseline:** 03-UI-SPEC.md (extends 02-UI-SPEC.md)
**Screenshots:** not captured fresh (no dev server on 3000/5173/8080, and none was started). Visual evidence comes from the existing headless-harness captures in `~/.dev-browser/tmp/`: `gifsmith-08-menu-col4.png`, `gifsmith-12-recents.png`, `gifsmith-13-chips.png`, `gifsmith-14-details-long-dark.png`, `gifsmith-14-toast-long-light.png`
**Interaction captures:** off (workflow.ui_interaction_capture not set). Interaction findings below come from reading the code.

Advisory, non-blocking. These accepted items were not scored against the phase: the 560px setup card, unlabeled Klipy ads, reduced-motion GIF animation (deferred to Phase 4), and the chip row drawn 4px taller to fit the focus ring.

---

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | 3/4 | Every contract string is present, but the T5a fallback is too long for the footer slot and gets cut off when the Enable button shows |
| 2. Visuals | 3/4 | Matches the spec's layout, but the action menu, Details and favorites can't be reached with the mouse, and Details has no visible way to close it |
| 3. Color | 3/4 | Fails the spec's contrast check: destructive labels on the highlighted menu row measure 4.29:1 (light) and 4.02:1 (dark), below 4.5 |
| 4. Typography | 4/4 | Only the 4 declared sizes and 2 weights are used; badge text uses a raw `font-size` instead of the `--label` token |
| 5. Spacing | 3/4 | Spacing stays on the 4/8/16 scale; the toast check icon has no explicit 4px gap, and the popover can cover the footer |
| 6. Experience Design | 3/4 | States and undo are complete, but toasts (Undo, errors) are not announced to screen readers and library footer hints show even when the list is empty |

**Overall: 19/24**

---

## Top 3 Priority Fixes

1. **WARNING: Destructive menu rows fail the spec's 4.5:1 contrast check when highlighted.** `--destructive` on `--surface-active` measures 4.29:1 (light, #D70015 on #E5E5EA) and 4.02:1 (dark, #FF6961 on #3A3A3C). The highlighted row is the one the user is about to run. Fix: in `ActionMenu.svelte`, add `.row.hi.destructive .label { color: var(--text) }` and keep the red only on rows that aren't highlighted. Or darken the token on the active surface (light about #B8000F, dark about #FF8A80) and measure again.
2. **WARNING: Footer toasts aren't announced to assistive tech.** `Footer.svelte:41` `.center` has no `aria-live`. `Removed from Recents  Undo ⌘Z`, the F1–F8 errors and `Downloaded to Downloads` change without being read out, so a VoiceOver user never hears about the 5 s undo window. Fix: add `role="status" aria-live="polite"` to `.center`, or put a visually hidden live region next to it.
3. **WARNING: The Accessibility fallback (T5a) is cut off.** `internal/services/actions.go:161` is about 600px at 12px. While `needsAccessibility` is set, `Enable Automatic Paste…` takes the right of the footer and the attribution takes the left, so the center shows about 150px (in `gifsmith-08-menu-col4.png` it renders as "Of…"). The instruction is lost and only the tooltip-less ellipsis remains. Fix: shorten the toast to the T1 text `Copied — press ⌘V to paste` and let the button carry the call to action, or put the full text in the button's `title`.

Further recommendations (minor):

4. **WARNING:** The action menu, Details and favorite toggle have no mouse entry point. There is no `contextmenu` handler and no visible button. The menu's rows accept clicks, but the menu can only be opened with ⌘K. Fix: open the menu on `contextmenu` over a tile, anchored with the same `placeMenu`.
5. **WARNING:** Details has no visible way to close it. The footer's `Esc Back` is a plain `<span>` you can't click (`Footer.svelte:57`). Fix: make the `Esc Back` hint a footer button, as `Show in Finder` is.
6. **WARNING:** `--text-tertiary` text measures 3.26:1 on `--bg`, in both themes. It's used for `No matching actions`, the ` · {count}` in section headers, the toast shortcut hints `⌘Z`/`⌘R`, and the footer hint words. These are small (12px) text, so they need 4.5:1. The tokens come from Phase 2 and the spec places these uses there, but Phase 3 put more text on the token. The chip clock icon (tertiary on `--surface`) measures 2.92:1 light and 2.75:1 dark, below the 3:1 minimum for non-text graphics.
7. **WARNING:** The library footer shows `↵ Copy  ⌘K Actions  ⌘⌫ Remove` even when the tab is in the empty or no-match state, where all three keys do nothing (`App.svelte` passes `hints='library'` whenever `libTab` is set). Fix: hide the hints when no item is selected.

---

## Detailed Findings

### Pillar 1: Copywriting (3/4)

- All copy-contract strings were found: the Favorites/Recents empty, no-match and load-error states (`App.svelte` library branch), T2–T10 and F1–F10 (split between `frontend/src/lib/errors.ts` and `internal/services/actions.go`), `Search actions`, `No matching actions`, `Untitled GIF`, and every action label in locked order (`lib/actions.ts:39-54`, with `media()` producing `Download GIF` / `Download MP4`).
- The three footer hint sets in `Footer.svelte:55-62` match the contract exactly.
- WARNING: T5a (`actions.go:161`) matches the spec word for word but doesn't fit its slot. See Fix 3.
- Minor: for Clips, the F3 string ends "isn't an MP4. … Try another GIF." (`errors.test.ts:176`). This follows the spec literally, but the noun is wrong. Suggest "Try another clip."
- Minor: the C2 card's secondary `Back to GIPHY` button isn't in the carried-forward C2 row (that cell is truncated in the spec). Confirm it was intended.

### Pillar 2: Visuals (3/4)

- The popover in `gifsmith-08-menu-col4.png` matches the spec: 320px, 12px radius, 40px filter row, 32px rows, group separators, a `--surface-active` highlight with no accent, and opening to the left of a column-4 tile.
- In the same capture, a bottom-row tile clamps the popover to 8px above the window edge, so it covers the footer attribution and center text. The clamp is within the contract, but while the menu is open the GIPHY/KLIPY attribution is hidden. Note this for the attribution compliance check.
- The Details pane (`gifsmith-14-details-long-dark.png`) has the 332px preview, a title clamped to 3 lines with an ellipsis, the provider line, and `<dl>` on aligned baselines. The tags row is correctly left out when there's no data.
- Badges match the spec (`Tile.svelte:101-129`). Only a background change marks tile hover (`Tile.svelte:81`), and an opaque GIF covers most of it, so mouse hover is barely visible. This is Phase 2 behavior, noted for completeness.
- WARNING: mouse users can't open the menu or Details and can't toggle favorites, and Details has no visible close control (Fixes 4 and 5).

### Pillar 3: Color (3/4)

- The accent is used only where the spec allows it: selection rings (`Tile.svelte:85`, `RecentSearches.svelte:78`), the Details links (`DetailsPane.svelte:141`), the toast action link (`Footer.svelte:111`), carets, and the shared focus ring (`app.css`). The highlighted menu row, chips, section headers and badges stay neutral. There are no hardcoded colors besides the spec's `#fff` on `--scrim` and the Phase 2 overlay `rgba(0,0,0,0.4)`.
- `--scrim` and `--shadow-popover` were added in both themes, as the spec requires.
- WARNING: the spec's contrast check fails for destructive labels on the highlighted row (4.29 and 4.02). See Fix 1.
- WARNING: tertiary text and icon contrast is too low (item 6).
- Contrasts that pass: white on scrim is at least 5.7:1, `--text-secondary` on a chip is 4.54 (light) and 5.42 (dark), the success toast is 5.41 (light), and the dark accent link is 6.35.

### Pillar 4: Typography (4/4)

- Font declarations in use: `var(--label)` ×13, `var(--body)` ×9, `var(--heading)` ×3, `var(--display)` ×1, `600 12px` ×1 (footer secondary button, P2), `600 14px` ×2 (buttons), and `400 14px` mono ×1 (key field). That is exactly sizes 12/14/18/22 and weights 400/600.
- Long-text rules are all in place: menu labels ellipsize while the hint has `flex: none`, chips cap at 160px, the Details title clamps to 3 lines, links stay on one line with an ellipsis, and the toast message ellipsizes while its action and hint stay whole (`gifsmith-14-toast-long-light.png`).
- Minor: the badges (`Tile.svelte:110-111`, `DetailsPane.svelte:92-93`) set `font-size: 12px; line-height: 16px` directly instead of `font: var(--label)`. The result looks the same, but it skips the token.

### Pillar 5: Spacing (3/4)

- A grep for padding, margin and gap values off the 4/8/16 scale found only 1px hairlines. The chip row has 8px top, 16px sides and an 8px gap. The menu list has 8px padding, row padding is 8px, the filter is 16px, and separators have 4px margins. Details has 16px padding and gap, 8px title-to-provider, 16px to the fields, and an 8px `<dl>` gap. Section headers are 8px/16px with `paddingStart: 8`. All of these match.
- Minor: the toast check icon (`Footer.svelte:44`) relies on a whitespace text node for its gap, about 3.5px, instead of the 4px icon-to-text gap the spec sets. In the long-toast capture the check sits right against "Downloaded". Fix: `.check { margin-right: 4px }` and drop the whitespace.
- Minor: `placeInDetails` (`lib/popover.ts`) ignores `menuH`. It's safe today (100 + 280 is less than 480), but it doesn't clamp if rows are ever added.
- The chip row being 36px with `margin-bottom: -4px` is accepted and wasn't scored.

### Pillar 6: Experience Design (3/4)

- State coverage is complete. The library has a skeleton that appears after 300ms (`LIBRARY_SKELETON_MS`), plus unreadable, empty, no-match and populated states. File actions show the busy overlay, and square crop shows it only after the 300ms delay (`CROP_BUSY_MS`). The 400ms hide timer is cancelled on show. Undo and download toasts last 5s and the favorite toast lasts 2s (`App.svelte:61-67`). Menu dismissal works by Esc, ⌘K, an outside mousedown (`App.svelte:883-884`) and panel hide.
- Removals are destructive and have no confirmation by design. Undo covers them, which is acceptable.
- Accessibility semantics: the menu has combobox, listbox and `aria-activedescendant` (`ActionMenu.svelte:74-77`). Chips have `Search again for {q}` labels, and tiles carry provider, MP4 and favorite in their labels. All good.
- WARNING: toasts don't use a live region (Fix 2).
- WARNING: there is no mouse path to the menu or Details (items 4 and 5).
- WARNING: library hints are shown on empty tabs (item 7). While `needsAccessibility` is set, all key hints are hidden, which makes the problem in Fix 3 worse.

Registry audit: skipped. `components.json` is absent and the spec declares no registries.

---

## Files Audited

- frontend/src/App.svelte
- frontend/src/app.css
- frontend/src/components/ActionMenu.svelte, DetailsPane.svelte, Footer.svelte, LibraryGrid.svelte, RecentSearches.svelte, Tile.svelte
- frontend/src/lib/actions.ts, popover.ts, errors.ts (via strings)
- internal/services/actions.go (toast strings)
- .planning/phases/03-action-pipeline-library/03-UI-SPEC.md
- Screenshots: ~/.dev-browser/tmp/gifsmith-08-menu-col4.png, gifsmith-12-recents.png, gifsmith-13-chips.png, gifsmith-14-details-long-dark.png, gifsmith-14-toast-long-light.png
