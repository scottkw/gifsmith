---
phase: 02-providers-search-ui-first-run
plan: 02
subsystem: panel UI (grid, keyboard), native panel shell
tags: [svelte5, tanstack-virtual, virtualization, keyboard, a11y, wails-v3, node-test, dev-browser]

requires:
  - phase: 02-providers-search-ui-first-run
    provides: "02-01 lib/api.ts facade, provider.Page cursor, App.svelte tracer grid, smoke-browser-loop.sh --uat harness"
provides:
  - "ResultGrid.svelte: row-virtualized 4-column grid (152px rows, overscan 2, 16px padding/scroll padding) with the S5/S6/S12a extra row and a last-row sentinel"
  - "Tile.svelte: selection ring, copy overlay, broken-preview icon, click selects / double-click copies"
  - "lib/grid.ts: COLS, move, nearEnd, routeKey, KeyAction (pure, node --test covered)"
  - "App paging: one page request at a time through a per-result-set fetch closure; resultSeq drops stale pages"
  - "Keyboard model: arrows, Enter/Cmd+Enter, Esc clear-then-hide, Cmd+K/Tab bound; focus pinned to the search input"
  - "shell.go: HideOnEscape false, WindowClosing cancel+hide hook, App+Edit-only application menu"
  - "frontend `pnpm run test` (node --test, TAP) and global .spinner in app.css"
  - "scripts/uat/02-grid.js, scripts/uat/02-keyboard.js"
affects: [02-03, 02-04, 02-05, 02-06]

actuals:
  tokens: 8300    # chars/4 over the final content of the 11 changed files, lockfile excluded (~33k chars)
  tasks: 3        # Task 1 = approved package gate (no commit), Tasks 2-3 executed
  commits: 3
plan_head_before: b0e13d33633d4b87fa1b01b7e118550292983621
plan_head_after: f2d897247a9ba0fb0f2ee29f1e288d1cbcaea8cd

tech-stack:
  added:
    - "@tanstack/svelte-virtual 3.13.47 (exact pin; brings @tanstack/virtual-core 3.19.6)"
  patterns:
    - "Virtualizer store calls (setOptions, scrollToIndex) run inside untrack; rows keyed by row index, tiles by column index"
    - "A new result set remounts ResultGrid via {#key resultSeq}, which resets scroll to the top"
    - "Pure UI logic lives in lib/*.ts with no Svelte/DOM imports and is tested by node --test (type stripping, explicit .ts imports)"
    - "Every key the page routes is preventDefault-ed; keys inside an IME composition are never routed"

key-files:
  created:
    - frontend/src/components/ResultGrid.svelte
    - frontend/src/components/Tile.svelte
    - frontend/src/lib/grid.ts
    - frontend/src/lib/grid.test.ts
    - scripts/uat/02-grid.js
    - scripts/uat/02-keyboard.js
  modified:
    - frontend/package.json
    - frontend/pnpm-lock.yaml
    - frontend/tsconfig.json
    - frontend/src/App.svelte
    - frontend/src/app.css
    - internal/shell/shell.go

key-decisions:
  - "Keys pressed during an open IME composition (isComposing or keyCode 229) route to 'none' for every key, not only Enter, so arrows/Esc reach the IME candidate window"
  - "One <main> mousedown handler keeps focus in the search input for any click outside a text field (replaces per-tile handlers; 02-05's key field is an input, so it stays clickable)"
  - "Paging uses resultSeq (bumped on every result-set replacement) instead of the search seq, so a page that lands after a new search is dropped rather than appended"
  - "A failed page stays 'error' until Try Again; neither the sentinel nor nearEnd retries it automatically"
  - "node tests are excluded from svelte-check instead of adding @types/node (the plan forbids new packages)"

patterns-established:
  - "Frontend unit tests: frontend/src/lib/*.test.ts, run with `pnpm --dir frontend run test` (TAP output, '# fail 0')"

requirements-completed: [SRCH-04, SRCH-05]

coverage:
  - id: D1
    description: "Row virtualization: with 60 loaded and the grid at the top, 20 of 60 tile images are mounted"
    requirement: SRCH-04
    verification:
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/02-grid.js"
        status: pass
    human_judgment: false
  - id: D2
    description: "Scroll paging reaches exactly 60 distinct tiles then 'End of results'; selection kept across pages; no each_key_duplicate"
    requirement: SRCH-04
    verification:
      - kind: automated_ui
        ref: "scripts/uat/02-grid.js (screenshot ~/.dev-browser/tmp/gifsmith-02-end.png)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Selection math and key routing (clamps, wrap, last full row, nearEnd threshold, Enter guards, IME)"
    requirement: SRCH-05
    verification:
      - kind: unit
        ref: "frontend/src/lib/grid.test.ts (7 tests, # fail 0)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Arrows move aria-activedescendant without moving the caret; click keeps input focus; arrow paging appends page 2 keeping selection; Cmd+K/Tab no-op; Esc clears then hides"
    requirement: SRCH-05
    verification:
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/02-keyboard.js"
        status: pass
    human_judgment: false
  - id: D5
    description: "Native panel: second Esc hides, Cmd+W hides (does not destroy) and the hotkey still reopens; Edit-menu Cmd+V reaches a text field in the non-activating panel"
    verification: []
    human_judgment: true
    rationale: "Needs real key input on the user's desktop; agents never send synthetic input. Scheduled for the 02-06 native checkpoint."

duration: 6min
completed: 2026-10-17
status: complete
---

# Phase 2 Plan 02: Virtualized Keyboard Grid Summary

**The result grid is now row-virtualized with @tanstack/svelte-virtual (only about five rows mounted), pages through the provider cursor to "End of results", and is fully keyboard-driven from the search input; the native panel no longer swallows Esc, cannot be destroyed by Cmd+W or cancelOperation, and has only App and Edit menus.**

## Performance

- **Duration:** 6 min of execution in this continuation (start 2026-10-17T20:01:53Z, end about 20:08Z)
- **Tasks:** 3 (Task 1 was the package gate, approved by the user before this run)
- **Files modified:** 12 (6 created, 6 modified)

## Accomplishments

- `ResultGrid.svelte` follows RESEARCH Pattern 8 and the UI-SPEC geometry. Rows are 152 px with a 48 px extra row. Overscan is 2, and padding and scroll padding are both 16. The extra row shows a spinner while loading (S5), "End of results" once 2 or more rows are loaded (S6), and "Couldn't load more." + Try Again on a page error (S12a). A sentinel fires when the last tile row is mounted. Selection changes call `scrollToIndex(row, { align: 'auto' })`.
- `Tile.svelte`: 166×144 tile with an accent ring at 2 px offset and a hover background. It shows the copy overlay spinner, and a 16 px alert-circle when the preview fails. The broken state is keyed by src, so a reused tile retries after a new search.
- `App.svelte` paging: at most one page in flight; appended pages keep the selection; a new result set resets to tile 0 and idle. Paging is triggered by the sentinel and by `nearEnd` after arrow moves.
- `lib/grid.ts` plus 7 node tests. `App.svelte` routes all keys through `routeKey`. The search debounce also starts from `compositionend`.
- `shell.go`: `HideOnEscape: false`, a `WindowClosing` hook that cancels the close and hides the panel, and `SetApplicationMenu` with only `AppMenu` and `EditMenu`. The tray and hotkey are unchanged.
- Shared `.spinner` moved to `app.css`. It takes its colour from `color`: white on the copy overlay, secondary text colour in the paging row.

## Task Commits

1. **Task 1: package legitimacy gate.** The user replied "approved" for `@tanstack/svelte-virtual@3.13.47` (exact pin, pnpm). No commit.
2. **Task 2: virtualized grid that pages to End of results**: `c3b48fd` (feat)
3. **Task 3: keyboard model and a panel that owns Esc**: `d54024a` (test, RED), then `f2d8972` (feat, GREEN)

## TDD Gate Compliance

- **Task 3:** RED `d54024a` → GREEN `f2d8972`. Evidence is in `bin/red-02-02-t3-record.json`, from the node --test TAP output. The classifier returned `RED_EVIDENCE_OK` with target "move down by a row, last full row falls to the last tile".
  - **Semantic assessment:** the file loaded under Node type stripping and all 7 tests were discovered. The target failed on its first `strictEqual` (`-1 !== 5`) against the RED skeleton. 6 of 7 tests failed on assertions. The IME-guard test passed against the skeleton because the expected value is the skeleton's `'none'`.
  - In the RED commit, `pnpm run check` failed because svelte-check could not resolve `node:test` types. The GREEN commit fixes this; see Deviation 1.
- No REFACTOR commit.

## Files Created/Modified

- `frontend/src/components/ResultGrid.svelte`: virtualizer, extra row, sentinel
- `frontend/src/components/Tile.svelte`: tile component
- `frontend/src/lib/grid.ts`, `frontend/src/lib/grid.test.ts`: selection math and key routing, plus tests
- `frontend/src/App.svelte`: paging state, ResultGrid, keyboard routing, focus pinning, compositionend
- `frontend/src/app.css`: global `.spinner`
- `frontend/package.json`, `frontend/pnpm-lock.yaml`: pinned dependency, `test` script
- `frontend/tsconfig.json`: `allowImportingTsExtensions`, `noEmit`, and an exclude for `*.test.ts`
- `internal/shell/shell.go`: Esc, close and menu changes
- `scripts/uat/02-grid.js`, `scripts/uat/02-keyboard.js`: headless UAT

## Decisions Made

See `key-decisions` in the frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] svelte-check could not type the node tests**
- **Found during:** Task 3 RED
- **Issue:** The repo has no `@types/node`, so `import { test } from 'node:test'` failed `pnpm run check`. The plan forbids adding packages.
- **Fix:** Added `"exclude": ["src/**/*.test.ts"]` to `frontend/tsconfig.json`. The tests still run under `node --test`, which strips types without checking them.
- **Committed in:** `f2d8972`

**2. [Rule 1 - Bug] IME composition keys other than Enter**
- **Found during:** Task 3
- **Issue:** The planned `routeKey` guarded only Enter. Arrows and Esc pressed while an IME candidate window is open would be preventDefault-ed and used to move the grid selection, which breaks Japanese and Chinese input.
- **Fix:** `routeKey` returns `'none'` for any key when `isComposing || keyCode === 229`. Two test assertions cover this.
- **Committed in:** `f2d8972`

**3. [Rule 2 - Missing critical] Stale page appended to a new result set**
- **Found during:** Task 2
- **Issue:** A page request started before a new search could land after the new results arrived and be appended under the same search `seq`.
- **Fix:** Added `resultSeq`, which is bumped by every `setResults`, and `loadMore` drops any page whose `resultSeq` no longer matches.
- **Committed in:** `c3b48fd`

**4. [Rule 2 - Plan wording] `test` script uses the TAP reporter**
- **Issue:** Node 24's default reporter is spec, but the acceptance check greps for `# fail 0`.
- **Fix:** `node --test --test-reporter=tap src/lib/*.test.ts`
- **Committed in:** `f2d8972`

**Total deviations:** 4 auto-fixed. **Impact:** no scope added, and only the approved package was installed.

## Issues Encountered

- `golangci-lint run ./...` reports one pre-existing `unused` finding in `build/ios/app_options_default.go`, from the Phase 1 template scaffold. This is out of scope. `golangci-lint run . ./internal/...` reports 0 issues.

## Not exercised by automation (no stubs)

- The S5 spinner row is only visible for a moment. The S12a row ("Couldn't load more." + Try Again) and the broken-preview icon are not exercised by UAT, because `fakeapi` has no scenario that fails page 2 only or serves a broken preview. Both pass svelte-check and are wired. 02-03 adds per-kind error states and is the natural place for a page-2 failure scenario.
- Native-only behaviour (Esc twice, Cmd+W, Edit-menu Cmd+V) is user-verified at the 02-06 checkpoint (D5).

## User Setup Required

None.

## Next Phase Readiness

- 02-03 can add the search-bar spinner (global `.spinner`), the per-kind states, and the paging error kinds (S9a/E2) in ResultGrid's extra row.
- 02-05 extends `routeKey` with ⌘1–⌘4. The `<main>` mousedown guard already lets clicks focus any `<input>`.

## Self-Check: PASSED

- Files: all 6 `key-files.created` exist
- Commits on this branch: c3b48fd, d54024a, f2d8972 (`git rev-list --count b0e13d3..HEAD` = 3)
- Verification: `pnpm run test` # pass 7 / # fail 0; `pnpm run check` 0 errors / 0 warnings; gofmt clean; `go vet . ./internal/...` ok; `go test -race ./internal/... -count=1` 7/7 packages ok; smoke `SMOKE OK`; `--uat` 02-grid, 02-keyboard and 01-tracer each print `UAT OK`; desktop and server builds ok
- Acceptance greps (Tasks 2–3): all match; FileMenu/ViewMenu/WindowMenu count 0; no `package-lock.json`

---
*Phase: 02-providers-search-ui-first-run*
*Completed: 2026-10-17*
