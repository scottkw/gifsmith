---
phase: 03-action-pipeline-library
plan: 06
subsystem: library
tags: [go, json-files, favorites, recents, undo, svelte5, tanstack-virtual, headless-uat]

requires:
  - phase: 03-action-pipeline-library
    plan: 04
    provides: App toast model (message, tone, action, expiry), runAction dispatcher, lib/actions.ts ROWS
  - phase: 03-action-pipeline-library
    plan: 05
    provides: CopySquare/PasteSquare with the same Result shape as Copy/Paste
provides:
  - settings.WriteJSONAtomic (temp, Chmod 0600, Sync, rename; dir 0700); SavePrefs delegates to it
  - package internal/library (Store, Entry, Snapshot, Validate, RecentsMax 50, FavoritesMax 1000, SearchesMax 20, ErrUnreadable, ErrInvalid)
  - Store methods Snapshot, Touch, SetFavorite, RemoveRecent, ClearRecents, Undo, SaveSearch, RemoveSearch
  - services.LibraryService (Load, SetFavorite, RemoveRecent, ClearRecents, Undo, SaveSearch, RemoveSearch); ActionError kinds library_read, library_write, invalid_item
  - ActionService.Recents (Touch after every successful copy/paste except page links)
  - frontend LIBRARY_TABS/LibraryId, KeyAction slot5/slot6/favorite/remove/undo, lib/library.ts (sections, flatItems, rows, moveSectioned), LibraryGrid.svelte, Tile favorite badge, Footer library hints + empty attribution, api.ts library wrappers, menu rows favorite/removeRecent/clearRecents, errors.ts libraryErrorText (F5)
  - scripts/uat/12-library.js, scripts/uat/12-library-unreadable.js, smoke UAT_SEED_FAVORITES
affects: [03-07, 03-08]

actuals:
  tokens: 31932
  tasks: 3
  commits: 6
plan_head_before: c2419bd49ac4d63d28df41c4367bd43e04a8b836
plan_head_after: 56cdf9d0be8e4df460cf21be92ab546ea373956d

tech-stack:
  added: []
  patterns:
    - "Library mutations compute the new list, write it with WriteJSONAtomic, and only then swap it into memory"
    - "A library file that fails to parse, has another version, or holds an entry failing Validate is unreadable: re-read on every access, every write refused, never overwritten"
    - "Every LibraryService mutation returns the whole fresh Snapshot; the frontend never computes membership from stale state"
    - "Keys that double as native editing (⌘⌫, ⌘Z) are preventDefault-ed only when they act"

key-files:
  created:
    - internal/library/store.go
    - internal/library/store_test.go
    - internal/services/library.go
    - internal/services/library_test.go
    - frontend/bindings/gifsmith/internal/library/index.ts
    - frontend/bindings/gifsmith/internal/library/models.ts
    - frontend/bindings/gifsmith/internal/services/libraryservice.ts
    - frontend/src/lib/library.ts
    - frontend/src/lib/library.test.ts
    - frontend/src/components/LibraryGrid.svelte
    - scripts/uat/12-library.js
    - scripts/uat/12-library-unreadable.js
  modified:
    - internal/settings/settings.go
    - internal/settings/settings_test.go
    - internal/services/actions.go
    - internal/services/actions_test.go
    - internal/services/errors.go
    - main.go
    - frontend/bindings/gifsmith/internal/services/index.ts
    - frontend/src/lib/api.ts
    - frontend/src/lib/providers.ts
    - frontend/src/lib/grid.ts
    - frontend/src/lib/grid.test.ts
    - frontend/src/lib/actions.ts
    - frontend/src/lib/actions.test.ts
    - frontend/src/lib/errors.ts
    - frontend/src/lib/errors.test.ts
    - frontend/src/components/ProviderTabs.svelte
    - frontend/src/components/ResultGrid.svelte
    - frontend/src/components/Tile.svelte
    - frontend/src/components/Footer.svelte
    - frontend/src/App.svelte
    - scripts/smoke-browser-loop.sh
    - scripts/uat/05-tabs.js
    - scripts/uat/08-action-menu.js

key-decisions:
  - "internal/library does not import staging (staging pulls in provider/httpx, which the plan's go list -deps check forbids). The PageAllow rule is applied as 'GIPHY or Clips, and the injected media Allow passes', which is what staging.PageAllow does"
  - "Load never fails: a file that cannot be read is named in Snapshot.Unreadable and its tab shows the load error; library_read comes only from a mutation on such a file"
  - "Every library mutation error (library_read, library_write, invalid_item) shows F5 for the list that changed; UI-SPEC has no other copy for them"
  - "Switching to a library tab selects the first item of the freshly loaded list, not the item at the old index (UAT caught a stale selection)"
  - "Favorites past FavoritesMax (1000) drop the oldest, like Recents; it is a sanity cap"
  - "rowOfIndex was not added: LibraryGrid finds the selected tile row in its own rows array for scrollToIndex"

patterns-established:
  - "UAT_SEED_FAVORITES seeds a library file before a --uat run and fails the run if the app changed it"

requirements-completed: [LIBR-01, LIBR-02, LIBR-03, LIBR-04]

coverage:
  - id: D1
    description: "Every successful copy or paste of GIF/MP4, link, Markdown and square GIF records the item in Recents (newest first, cap 50, re-use moves to top, no duplicates); page link, open, download, failures and busy refusals do not"
    requirement: LIBR-02
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestRecentsRecordedOnEverySuccessfulCopy, TestRecentsNotRecorded, TestRecentsErrorDoesNotFailAction; internal/library/store_test.go#TestTouchMovesToTop, TestTouchCap50, TestRecentsPersistAcrossStores"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/12-library.js (steps 1-3)"
        status: pass
    human_judgment: false
  - id: D2
    description: "⌘6 Recents and ⌘5 Favorites render from saved snapshots, grouped by provider in slot order with '{label} · {count}' headers, filtered by title locally; they work while GIPHY is rate-limited"
    requirement: LIBR-03, LIBR-04
    verification:
      - kind: unit
        ref: "frontend/src/lib/library.test.ts (sections, filter, rows, moveSectioned)"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/12-library.js (steps 1-4: headers, order, rate-limited copy from Recents; step 6 filter, No matches, Esc)"
        status: pass
    human_judgment: false
  - id: D3
    description: "⌘⇧F / menu row toggles favorites with the 2 s toast and the star badge (not on the Favorites tab); ⌘⌫ removes on library tabs, Clear Recents in ⌘K, 5 s Undo toasts restore positions and reselect; ⌘⌫/⌘Z stay native on provider tabs"
    requirement: LIBR-01, LIBR-03
    verification:
      - kind: unit
        ref: "internal/library/store_test.go#TestSetFavorite, TestUndoRestoresPosition, TestUndoSkipsKeysPresentAgain; frontend/src/lib/grid.test.ts, actions.test.ts"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/12-library.js (steps 3, 5, 6, 7)"
        status: pass
    human_judgment: true
    rationale: "Headless Chromium has no native Edit menu. Whether WKWebView's Edit menu takes ⌘Z/⌘⌫ before the DOM keydown (Pitfall 10) can only be checked in the signed dev build at the plan 03-08 native checkpoint."
  - id: D4
    description: "Corrupt or unknown-version library files are never reset or overwritten: the tab shows Couldn't load with Try Again, writes are refused (F5), other lists keep working; a fixed file loads again"
    requirement: LIBR-04
    verification:
      - kind: unit
        ref: "internal/library/store_test.go#TestUnreadableFileNeverOverwritten, TestUnknownVersionNeverOverwritten, TestInvalidEntryOnDiskIsUnreadable, TestSearchesFileUnreadable; internal/services/library_test.go#TestLibraryServiceErrorKinds, TestLibraryServiceWriteFailureKeepsList"
        status: pass
      - kind: automated_ui
        ref: "UAT_SEED_FAVORITES='{not json' bash scripts/smoke-browser-loop.sh --uat scripts/uat/12-library-unreadable.js (file bytes checked unchanged after the run)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Library tabs show the S2 skeleton only if the read takes over 300 ms"
    verification: []
    human_judgment: true
    rationale: "Local reads answer in well under 300 ms, so the skeleton path is not exercised by any fixture."

duration: 20min
completed: 2026-10-18
status: complete
---

# Phase 3 Plan 06: Favorites and Recents Summary

**Every successful copy or paste now lands in Recents. ⌘⇧F toggles favorites. ⌘5 (Favorites) and ⌘6 (Recents) show the saved metadata grouped by provider, with removal, Clear Recents and a 5 s Undo. Nothing on these tabs calls a provider, so they work offline and while a provider is rate-limited. The lists live in `favorites.json`, `recents.json` and `searches.json` (version 1, mode 0600, written atomically). A file that fails to load is never overwritten.**

## Performance

- **Duration:** about 20 min
- **Started:** 2026-10-18T01:19:04Z
- **Completed:** 2026-10-18T01:39:09Z
- **Tasks:** 3, each RED then GREEN
- **Files modified:** 35

## Accomplishments
- `settings.WriteJSONAtomic` writes a temp file in the same dir, then Chmod 0600, Sync, Close and rename into a 0700 dir. On any error the temp file is removed and the old file stays. `SavePrefs` now delegates to it.
- `internal/library.Store`:
  - Holds one mutex and loads each file lazily.
  - `Validate` checks: known provider, id of 1 to 200 bytes, title up to 500 runes, Preview/Download (required) and Medium/Gif URLs through the injected `Allow`, a page URL only for GIPHY/Clips on an allowed host, an https source URL, and at most 50 tags of 100 runes each.
  - A file is unreadable if it fails to parse, has a version other than 1, holds an invalid or duplicate entry, or cannot be read. It refuses writes and is re-read on the next access.
  - One Undo slot holds the latest removal.
  - Searches are kept per provider.
  - No Wails import; `go list -deps` shows no wailsapp or httpx.
- `ActionService.Recents.Touch` runs once after each successful Copy, Paste (copy-only fallbacks included), CopyText link/markdown, PasteText, CopySquare and PasteSquare. Page links, Download, failures and busy refusals never call it. A Touch error is only logged.
- `LibraryService` is registered in main.go. It shares `d.ConfigDir`, so the server build uses `bin/uat-config`. Bindings were regenerated, including the new `internal/library` models.
- Frontend:
  - Pills: a divider, then Favorites (outline star, ⌘5) and Recents (⌘6). Neither is persisted as the last tab.
  - `LibraryGrid`: virtual rows of 32px headers and 152px tile rows, `paddingStart` 8. It is remounted when the row layout changes.
  - The query filters titles locally. Arrow keys use `moveSectioned`.
  - The footer attribution follows the selected item and is empty when nothing is selected. Footer hints read `↵ Copy ⌘K Actions ⌘⌫ Remove`.
  - Star badge on favorited tiles in provider grids and Recents.
  - Menu rows: favorite (group 5), plus Remove from Recents and Clear Recents (group 6, Recents tab only).
  - Toasts T7 to T10. Undo toasts survive keystrokes and selection changes.
  - States: empty Favorites, empty Recents, No matches, Couldn't load with Try Again, and F5.

## Task Commits

1. **Task 1 (tracer): copy then ⌘6 shows the item in Recents**: `b797eb1` (test, RED), `38cff48` (feat, GREEN). Tracer gate: the automated verify was re-run (Go, node tests, svelte-check, smoke, 12-library, 05-tabs) and passed.
2. **Task 2: complete store and bound service**: `afd5f85` (test, RED), `749176d` (feat, GREEN)
3. **Task 3: Favorites tab, keys, remove/Clear/Undo, states**: `296830d` (test, RED), `56cdf9d` (feat, GREEN)

## TDD Gate Compliance

- **Task 1 RED (TS):** "sections group by provider in slot order, keeping input order within a section" failed against the empty skeleton. `gsd-tools check tdd-red-evidence` returned `RED_EVIDENCE_OK` (`bin/red/0306-t1-library.json`). Semantic assessment: the target ran and failed on its deepEqual, with no load fault. The routeKey ⌘5/⌘6 test failed too.
- **Task 1 RED (Go):** skeletons compiled. All six library tests, TestRecentsRecordedOnEverySuccessfulCopy, TestRecentsErrorDoesNotFailAction and both WriteJSONAtomic tests failed on their assertions (`bin/red/0306-t1-go.txt`). TestRecentsNotRecorded passed in RED by design: it guards that nothing records, and the skeleton recorded nothing.
- **Task 2 RED (Go):** 13 store tests and 3 service tests failed against skeleton methods (`bin/red/0306-t2-go.txt`). TestInvalidEntryOnDiskIsUnreadable already passed, because Task 1's loader covers it.
- **Task 3 RED (TS):** the targets in library, actions, grid and errors each returned `RED_EVIDENCE_OK` (`bin/red/0306-t3-*.json`). The existing row-list tests were updated for the favorite row and failed with them.
- Each GREEN commit follows its RED commit. No refactor commits.

## Verification (latest run)

- `gofmt -l` is empty. `go vet . ./internal/...` passes. `go test -race ./internal/... -count=1` reports all ok. `golangci-lint run . ./internal/...` reports 0 issues.
- `go list -deps ./internal/library | grep -c -e wailsapp -e internal/provider/httpx` prints 0.
- `go build .` (desktop) and `CGO_ENABLED=0 go build -tags server .` both build.
- `bin/test-03-06-t2.txt` has no FAIL, has the TestUnreadableFileNeverOverwritten and TestUndoRestoresPosition PASS lines, and has the library `ok` line. `-run Search` shows 6 PASS.
- `pnpm --dir frontend run test`: 52 pass, 0 fail. svelte-check: 0 errors, 0 warnings.
- `scripts/smoke-browser-loop.sh`: SMOKE OK.
- `--uat` reports UAT OK for 12-library, 12-library-unreadable (with `UAT_SEED_FAVORITES`), 08-action-menu, 02-keyboard, 05-tabs, 07-text-actions, 01-tracer, 02-grid, 03-states, 05-setup, 06-clips, 10-download and 11-square. 06-clips-denied and 06-backstops pass with `FAKE_CLIPS=denied`.
- Every acceptance grep in all three tasks matches.

## Decisions Made
See `key-decisions` in the frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] internal/library cannot import staging**
- **Found during:** Task 1
- **Issue:** The plan says to validate PageURL with `staging.PageAllow`. But staging imports `internal/provider/httpx`, and the Task 1 verify requires `go list -deps ./internal/library` to show no httpx.
- **Fix:** Validate allows a page URL only for GIPHY/Clips, and only when the injected media `Allow` accepts it. That is the same rule as `PageAllow` (VendorAllow for GIPHY), without the import.
- **Commit:** `38cff48`

**2. [Rule 3 - Blocking] Existing UAT and test expectations listed the old tabs and rows**
- `05-tabs.js` counted exactly 4 `[role=tab]` and now expects six pills. `08-action-menu.js` now expects 10 Finer rows with 4 separators, 12 GIPHY rows, and the clamp landing on Add to Favorites. `actions.test.ts` row lists gained the favorite row.
- **Commits:** `38cff48`, `296830d`, `56cdf9d`

**3. [Rule 1 - Bug] Stale selection on a library tab switch**
- **Found during:** Task 1 UAT
- **Issue:** On ⌘6 the reselect kept the key of the item at the old index in the stale snapshot, so it landed on an older item.
- **Fix:** A tab switch reloads with `want = ''`, so the first item of the fresh list is selected.
- **Commit:** `38cff48`

**4. [Rule 2 - Coverage] UI check for the corrupt-file state**
- The must-have "Couldn't load … never overwritten" had no UI check. I added `UAT_SEED_FAVORITES` to the smoke script and `scripts/uat/12-library-unreadable.js`. The run fails if the seeded file's bytes change.
- **Commit:** `56cdf9d`

**5. rowOfIndex not added**
- LibraryGrid finds the selected tile row with `rows.findIndex`, so a second export was not needed. It is listed in the plan's Artifacts but not in must_haves.

---

**Total deviations:** 4 auto-fixed (2 blocking, 1 bug, 1 coverage) and 1 omission. **Impact:** no scope creep. The behavior matches UI-SPEC.

## Issues Encountered
- ⌘⌫ and ⌘Z only reach the DOM handler in headless Chromium. In WKWebView the native Edit menu may take ⌘Z first (Pitfall 10). Verify ⌘Z Undo and ⌘⌫ Remove in the signed dev build at the plan 03-08 checkpoint.
- Moving ↑ into a section's first tile row scrolls that row into view, but not its header. It looks fine at fixture sizes.
- A `searches.json` that fails to load is reported in `Snapshot.Unreadable`, but no UI shows it yet. Plan 03-07's chip row will.

## Known Stubs
None.

## Threat Flags
None beyond the plan's register:
- **T-03-21:** Validate runs on every write and every load.
- **T-03-22:** files that cannot be loaded are refused and never overwritten.
- **T-03-23:** files are 0600 in a 0700 dir and hold metadata only.
- **T-03-24:** caps of 50, 1000 and 20, plus the tag and title caps.
- **T-03-25:** Sync and rename, memory swapped only after a successful write, and a one-slot Undo.

No package installs.

## User Setup Required
None.

## Next Phase Readiness
- Plan 03-07 can use `api.saveSearch`/`removeSearch` and `Library.searches` (per provider id) for the chip row, and the `details` ActionId still has no row. ⌘⌫ with a chip selected must be added to App's `remove` capture condition.
- Plan 03-08 native checkpoint: ⌘Z/⌘⌫ in WKWebView, the favorite badge, and Recents after a real paste.

---
*Phase: 03-action-pipeline-library*
*Completed: 2026-10-18*

## Self-Check: PASSED
