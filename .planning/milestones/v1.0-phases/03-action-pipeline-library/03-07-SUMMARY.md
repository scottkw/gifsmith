---
phase: 03-action-pipeline-library
plan: 07
subsystem: frontend
tags: [svelte5, details, recent-searches, polish, a11y, headless-uat]

requires:
  - phase: 03-action-pipeline-library
    plan: 06
    provides: LibraryService saveSearch/removeSearch, Library.searches, libTab, toast model, Footer hint variants
  - phase: 03-action-pipeline-library
    plan: 01
    provides: Gif metadata fields (width/height/size/tags/pageUrl/sourceUrl/mediumUrl/gifUrl), OpenPage/OpenSource
provides:
  - lib/format.ts (fileSize, dimensions, linkText, tagsSummary, providerLine)
  - lib/searches.ts (chipsVisible, shouldSave, CHIP_MAX, SAVE_IDLE_MS)
  - lib/popover.ts placeInDetails; KeyAction 'details' (⌘I); actions.ts "Show Details" row (group 5, omitted in Details)
  - components/DetailsPane.svelte, components/RecentSearches.svelte; Footer 'details' hints
  - App detailsOpen, chip selection and keys, idle/copy save rules, IN-04 fix
  - Shared accent focus ring in app.css and the P2-P7 polish
  - scripts/uat/13-details.js, 13-chips.js, 14-backstops.js
affects: [03-08]

actuals:
  tokens: 22849
  tasks: 3
  commits: 6
plan_head_before: 7ec1c701209626deafda57eda1266dc4300b41b7
plan_head_after: 0066b6a16f2b96ff9bd89cedb3d87c7b14fa26a8

tech-stack:
  added: []
  patterns:
    - "Details is an absolutely positioned layer over the content area; the grid stays mounted below it, so ← → move the real grid selection and closing needs no remount or re-scroll"
    - "Chip keyboard selection is App state (chip index); the search input's aria-activedescendant points at chip-{i}, tiles are deselected (selected = -1) while a chip is selected"
    - "Scoped `input:focus-visible { outline: 0 }` opt-outs beat the global ring on specificity, not stylesheet order"

key-files:
  created:
    - frontend/src/lib/format.ts
    - frontend/src/lib/format.test.ts
    - frontend/src/lib/searches.ts
    - frontend/src/lib/searches.test.ts
    - frontend/src/components/DetailsPane.svelte
    - frontend/src/components/RecentSearches.svelte
    - scripts/uat/13-details.js
    - scripts/uat/13-chips.js
    - scripts/uat/14-backstops.js
  modified:
    - frontend/src/lib/popover.ts
    - frontend/src/lib/popover.test.ts
    - frontend/src/lib/actions.ts
    - frontend/src/lib/actions.test.ts
    - frontend/src/lib/grid.ts
    - frontend/src/lib/grid.test.ts
    - frontend/src/App.svelte
    - frontend/src/app.css
    - frontend/src/components/Footer.svelte
    - frontend/src/components/ProviderTabs.svelte
    - frontend/src/components/SetupCard.svelte
    - frontend/src/components/ResultGrid.svelte
    - frontend/src/components/ActionMenu.svelte
    - scripts/uat/08-action-menu.js
    - .planning/phases/02-providers-search-ui-first-run/02-UI-SPEC.md
    - .planning/todos/done/phase-2-ui-review-polish.md (moved from pending)

key-decisions:
  - "Details overlays the content area instead of replacing the grid in the #if chain: the grid keeps its scroll position and selection, and ⌘K placement in Details uses placeInDetails"
  - "While a chip is selected no tile is selected (selected = -1), so ↵ and ⌘⌫ unambiguously act on the chip and ⌘K / ⌘↵ are no-ops"
  - "The idle save fires at most once per provider:query until the next keystroke, from whichever of the 1.5 s timer or the first-page landing comes last; a chip rerun is a search and moves that chip to the front"
  - "IN-04: the stale Clips error is cleared only when the Clips state changes to enabled; clearing on 'unknown' would re-probe after every failed probe and loop while offline"
  - "P8 measures the string's own width (K6 = 492.5px, matching UI-SPEC's 493px); the 12px icon and 4px gap fit in the slot's remaining width, which scrollWidth <= clientWidth asserts"
  - "Chip row is 36px tall with -4px bottom margin (32px in layout) so the 2px+2px selection ring is not clipped by the row's horizontal overflow"

patterns-established:
  - "UAT scripts measure layout geometry (Details 720x364 at y 84, preview 332, chip row 32 + view 332, menu at 704/100) rather than only text"

requirements-completed: [LIBR-05, SRCH-08]

coverage:
  - id: D1
    description: "⌘I / Show Details: 332px preview (medium rendition, else preview), title (Untitled GIF), provider line, Dimensions/File size/Tags/Page/Source with missing fields omitted; no API call"
    requirement: LIBR-05
    verification:
      - kind: unit
        ref: "frontend/src/lib/format.test.ts; actions.test.ts#Show Details follows the favorite row; grid.test.ts#routeKey maps Cmd+I"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/13-details.js (steps 1, 2, 5)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Details keys: Esc/⌘I close with selection kept, ← → step (clamped), ↵/⌘↵ copy/paste, ⌘K top-right without Show Details, printable key closes and edits; Page/Source open through Go"
    requirement: LIBR-05
    verification:
      - kind: unit
        ref: "frontend/src/lib/popover.test.ts#placeInDetails"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/13-details.js (steps 1-6)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Recent-search chips: shown on an empty provider tab with saved searches (8 max, newest first, 32px row, grid 332px); never on library tabs; saved on 1.5 s idle with ≥1 result or on copy; ↑/↓/←/→/↵/⌘⌫/typing per UI-SPEC"
    requirement: SRCH-08
    verification:
      - kind: unit
        ref: "frontend/src/lib/searches.test.ts"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/13-chips.js"
        status: pass
    human_judgment: false
  - id: D4
    description: "Phase 2 polish P1-P11 and IN-04; contrast, keyboard-only and long-text backstops in light and dark"
    verification:
      - kind: automated_ui
        ref: "FAKE_CLIPS=denied bash scripts/smoke-browser-loop.sh --uat scripts/uat/14-backstops.js"
        status: pass
    human_judgment: true
    rationale: "IN-04 (Clips becoming enabled after a boot-time GIPHY error) has no fixture path to reproduce offline-then-online in the server build; the fix is by inspection. ⌘I and chip keys in WKWebView belong to the 03-08 native checkpoint."

duration: 33min
completed: 2026-10-18
status: complete
---

# Phase 3 Plan 07: Details and Recent Searches Summary

**⌘I opens a Details pane built only from the item's own metadata, with Page and Source links that open through Go. An empty search box now shows up to 8 recent-search chips per provider. A search is saved after 1.5 s idle with at least one result, or when a GIF from it is copied. Chips rerun, move and delete from the keyboard. The Phase 2 polish todo and IN-04 are closed, and 14-backstops.js checks contrast, keyboard-only use and long text in light and dark.**

## Performance

- **Duration:** about 33 min
- **Started:** 2026-10-18T01:41:00Z
- **Completed:** 2026-10-18T02:14:00Z
- **Tasks:** 3 (Tasks 1 and 2 RED then GREEN; Task 3 auto)
- **Files modified:** 25

## Accomplishments
- **Details (LIBR-05):**
  - `DetailsPane` sits in a layer over the 720x364 content area.
  - The preview box is 332x332. It shows `mediumUrl` (GIPHY `fixed_height`, Klipy `md`) or falls back to `previewUrl`, never the original. Clips get the MP4 badge.
  - The 340px metadata column has: title (3-line clamp, `title` attribute, `Untitled GIF`), the provider line, and a `<dl>` with the fixed order Dimensions, File size, Tags, Page, Source.
  - Links are `<button>`s showing the host and path, with the full URL as `title`. They call `openPage` / `openSource` (F8 on failure).
- **Details keys:** ⌘I and the menu row toggle Details; Esc closes it with the selection kept. ← → move the underlying grid selection (clamped, using the flat order on library tabs); ↑ ↓ do nothing. ↵ and ⌘↵ copy and paste. ⌘K opens the menu at `placeInDetails` (left 384, top 100) without Show Details. A printable key closes Details and goes to the input. Details closes on a tab switch, a new result set, panel show, and when a library list no longer has the item.
- **Chips (SRCH-08):**
  - `RecentSearches`: a listbox "Recent searches" with options `chip-{i}` labelled "Search again for {q}". Each chip has a clock icon and a 160px ellipsized label with `title`. The scrollbar is hidden and the selected chip is scrolled into view.
  - The grid area below shrinks to 332px. S8 and message states sit directly under the row.
- **Save rules:** the time of the last keystroke and a 1.5 s timer go through `maybeSave`, which runs both when the timer fires and when the first page lands. `shouldSave` needs at least 1 result and a query of 2+ code points. Every successful act() (copy, paste, text, square) from a searched provider grid calls `saveSearch` at once. Trending and library filters are never saved.
- **Chip keys:**
  - ↑ from the grid's first row, or ↓ with no grid, selects chip 0.
  - ← → move between chips, clamped.
  - ↓ goes back to tile 0, or does nothing when there is no grid.
  - ↵ reruns the chip at once, with no debounce.
  - ⌘⌫ removes the chip silently.
  - Typing clears the selection.
  - `aria-activedescendant` points at `chip-{i}`.
- **Polish:**
  - P1: one focus ring in app.css. The search input and menu filter opt out. SetupCard's copy is gone and the key field keeps the ring.
  - P2: secondary 24px "Enable Automatic Paste…" button.
  - P3: `button.off:not(.active)`.
  - P4: card spacing heading → 8 → body → 16 → steps.
  - P5: Try Again underlines on hover.
  - P6: disabled buttons use `--surface-active`/`--text-tertiary` with no opacity change.
  - P7: disabled search input text is `--text-tertiary`.
  - P9: C2 amended in four places in 02-UI-SPEC (body copy included).
  - P10: the F9/F10 strings match the code.
  - P11/IN-04: fixed (see the decisions).

## Task Commits

1. **Task 1 (tracer): ⌘I Details**: `3a028f6` (test, RED), `93eb9c2` (feat, GREEN). Tracer gate: the automated verify was re-run end to end (node tests, svelte-check, server build, 13-details, 08-action-menu, 12-library) and passed before Task 2.
2. **Task 2: recent-search chips**: `4a14400` (test, RED), `d9f593f` (feat, GREEN), `713cd25` (refactor: `saveSearch(` called directly from both paths)
3. **Task 3: polish and backstops**: `0066b6a` (feat)

## TDD Gate Compliance

- **Task 1 RED:** `bin/red/0307-t1-{format,actions,grid,popover}.json` each returned `RED_EVIDENCE_OK`. Semantic check: the skeletons loaded, and each target failed on its own assertion (empty strings, a missing row, `'none'`, `{0,0}`), with no load fault. The existing row-list tests in actions.test.ts were updated for the new row and failed with them.
- **Task 2 RED:** `bin/red/0307-t2-{chips,save}.json` returned `RED_EVIDENCE_OK`. The skeletons returned `false`, and the failures were on the deepEqual/equal for the true cases.
- Each GREEN commit follows its RED commit. Task 3 is `type="auto"`, so it has no RED.

## Verification (latest run)

- `pnpm --dir frontend run test`: 63 pass, 0 fail. svelte-check: 0 errors, 0 warnings.
- `go vet . ./internal/...` passes. `go test -race ./internal/... -count=1` is all ok (linker warnings only). `CGO_ENABLED=0 go build -tags server .` builds. `scripts/smoke-browser-loop.sh`: SMOKE OK.
- `--uat` prints UAT OK for: 01-tracer, 02-grid, 02-keyboard, 03-states, 05-setup, 05-tabs, 06-clips, 07-text-actions, 08-action-menu, 10-download, 11-square, 12-library, 13-details and 13-chips.
- With `FAKE_CLIPS=denied`: 06-clips-denied, 06-backstops and 14-backstops. With `UAT_SEED_FAVORITES`: 12-library-unreadable.
- 14-backstops contrast:
  - Light: white/scrim over white 5.74, over surface 6.20; text/surface-active 13.41; destructive/bg 5.38.
  - Dark: 5.74, 18.79, 10.42, 5.91.
  - P8 widest string is 492.5px.
- Every acceptance grep matches.
  - `saveSearch(` appears 2 times in App.svelte.
  - `Open GIPHY Dashboard` appears 0 times and `Email clips@giphy.com` 4 times in 02-UI-SPEC.
  - SetupCard has 0 `focus-visible` and 0 `opacity: 0.5`.
  - The todo is in `.planning/todos/done/`.

## Decisions Made
See `key-decisions` in the frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] 08-action-menu.js counted the old menu rows**
- **Issue:** The new Show Details row made Finer 11 rows and GIPHY 13. The clamp now lands on Show Details.
- **Fix:** I updated the expected list and counts. The assertions are just as strict.
- **Commit:** `93eb9c2`

**2. [Rule 1 - Bug] The chip selection ring would be clipped**
- **Issue:** The row's `overflow-x: auto` clips vertically too, so the 4px ring below a 24px chip in a 32px row was cut off.
- **Fix:** The row is 36px with a -4px bottom margin. It still takes 32px in the layout and the grid gets 332px. Checked by screenshot.
- **Commit:** `d9f593f`

**3. [Rule 1 - Bug] The global focus ring could override the input opt-outs**
- **Issue:** The global ring and the scoped `outline: 0` have equal specificity, so stylesheet order would decide.
- **Fix:** The search input and menu filter opt out with scoped `input:focus-visible { outline: 0 }`, which is more specific. 14-backstops asserts that the search input shows no outline.
- **Commit:** `0066b6a`

**4. Plan detail: where the P8 width is measured**
- The ≤ 500px limit is applied to the string's own width, as UI-SPEC defines it (K6 = 493px). It is not applied to the whole slot including the icon. The full slot is still asserted not to overflow (`scrollWidth <= clientWidth`).

**5. Files listed but not changed**
- No change was needed in `ResultGrid.svelte` for Task 2 or in `StateMessage.svelte`, because the shared ring in app.css covers its button and link. ActionMenu.svelte was changed for the P1 opt-out but is not in `files_modified`.

---

**Total deviations:** 3 auto-fixed (1 blocking, 2 bugs), plus 2 notes. **Impact:** no scope creep.

## Issues Encountered
- I could not reproduce IN-04 with the server-build fixtures (GIPHY offline at boot, then online). It is fixed by inspection: when Clips change to `enabled`, the stored Clips error is cleared, and `show('clips')` then loads. The review's second path (a `recheckClips` from an old key returning `key_rejected` after a new key was saved) was not part of P11 and is not addressed.
- Known gap carried forward: under `prefers-reduced-motion`, animated tiles and the Details preview keep animating. Still renditions are Phase 4.
- 03-08 native checkpoint: check ⌘I, the chip ↑/⌘⌫ keys and the Details link opening in the signed WKWebView build.

## Known Stubs
None.

## Threat Flags
None beyond the plan's register:
- **T-03-26:** Details uses text interpolation only, and its links call Go's OpenPage/OpenSource, which re-validate the URL.
- **T-03-28:** at most 8 chips are rendered.

No package installs.

## User Setup Required
None.

## Next Phase Readiness
- Plan 03-08 can run the native checkpoint. Phase 3 now has every Details, chip and polish item in the headless UAT suite.

---
*Phase: 03-action-pipeline-library*
*Completed: 2026-10-18*

## Self-Check: PASSED
