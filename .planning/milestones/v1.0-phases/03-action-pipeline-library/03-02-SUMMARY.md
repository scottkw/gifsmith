---
phase: 03-action-pipeline-library
plan: 02
subsystem: frontend-actions
tags: [svelte5, action-menu, popover, a11y, headless-uat]

requires:
  - phase: 03-action-pipeline-library
    plan: 01
    provides: textAction/run/act paths, openItem, routeKey copyLink/copyMarkdown/open, Gif pageUrl/gifUrl
provides:
  - lib/actions.ts (ActionId, ActionCtx, ActionRow, ROWS, buildActions, filterActions)
  - lib/popover.ts (MENU_W, placeMenu)
  - components/ActionMenu.svelte (combobox filter + listbox popover; focus(), height())
  - App.svelte menuOpen, toggleMenu, closeMenu, runAction (single dispatcher)
  - app.css --shadow-popover; footer "⌘K Actions" hint
  - scripts/uat/08-action-menu.js
affects: [03-04, 03-05, 03-06, 03-07, 03-08]

actuals:
  tokens: 8237
  tasks: 2
  commits: 4
plan_head_before: cbd9b163708ac41da034c1adca04b84153e11f79
plan_head_after: 8b65d416772180fbc13aa8a087a4c4d6f3974aa6

tech-stack:
  added: []
  patterns:
    - "Action rows live in one ROWS table with applies()/label(ctx); later plans add rows, never touch the menu"
    - "Every action (menu row or shortcut) goes through App runAction(id); the menu closes before anything runs"
    - "Menu owns ↑ ↓ ↵ Esc Tab in its filter; ⌘ keys bubble to main and route through onKeydown"
    - "Popover is position: fixed, measured after tick() and placed with placeMenu before paint"

key-files:
  created:
    - frontend/src/lib/actions.ts
    - frontend/src/lib/actions.test.ts
    - frontend/src/lib/popover.ts
    - frontend/src/lib/popover.test.ts
    - frontend/src/components/ActionMenu.svelte
    - scripts/uat/08-action-menu.js
  modified:
    - frontend/src/App.svelte
    - frontend/src/components/Footer.svelte
    - frontend/src/app.css
    - scripts/uat/02-keyboard.js

key-decisions:
  - "Any mousedown outside the popover closes it, including one on the search input (UI-SPEC: a click outside the popover closes it)"
  - "Rows highlight on mousemove, not mouseenter, so a menu opening under a resting pointer keeps row 1 highlighted"
  - "placeMenu receives the real window size (innerWidth/innerHeight); 720x480 stays the default"
  - "⌘K is a no-op unless the selected tile element exists in the DOM, which covers empty, loading, message and setup-card states in one check"

requirements-completed: [ACTN-07, ACTN-03, ACTN-04]

coverage:
  - id: D1
    description: "⌘K opens the Actions dialog beside the selected tile, filter focused, first row highlighted, aria-activedescendant on the highlighted option; ⌘K again or Esc closes it with selection, query and focus restored"
    requirement: ACTN-07
    verification:
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/08-action-menu.js"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/02-keyboard.js"
        status: pass
    human_judgment: false
  - id: D2
    description: "Locked row order, groups, hints, MP4 labels for Clips, omission of page rows without pageUrl and of link rows for a Clip without gifUrl"
    requirement: ACTN-07
    verification:
      - kind: unit
        ref: "frontend/src/lib/actions.test.ts"
        status: pass
    human_judgment: false
  - id: D3
    description: "↓↓↵ runs Copy Link; click runs a row; ⌘⇧M, ⌘↵ and ⌘K work while the menu is open; Paste Link / Paste Markdown / Copy Page Link reachable from the menu"
    requirement: ACTN-03, ACTN-04
    verification:
      - kind: automated_ui
        ref: "scripts/uat/08-action-menu.js"
        status: pass
    human_judgment: false
  - id: D4
    description: "Filter: case-insensitive substring, highlight resets, separators hidden while filtering, single 'No matching actions' row where ↵ runs nothing"
    requirement: ACTN-07
    verification:
      - kind: unit
        ref: "frontend/src/lib/actions.test.ts#filterActions"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/08-action-menu.js"
        status: pass
    human_judgment: false
  - id: D5
    description: "Placement backstop: tile 0 and the bottom visible row in columns 1 and 4 keep the popover >= 8px inside 720x480; GIPHY's 8 rows scroll in the 224px list with the highlight in view"
    requirement: ACTN-07
    verification:
      - kind: unit
        ref: "frontend/src/lib/popover.test.ts"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/08-action-menu.js (gifsmith-08-menu-col1.png, gifsmith-08-menu-col4.png)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Native WKWebView: ⌘K / ⌘⇧ keys while the menu filter has focus in the real panel"
    verification: []
    human_judgment: true
    rationale: "Synthetic desktop input is forbidden on the user's machine; headless Chromium cannot prove WKWebView key equivalents. Belongs with the plan 08 checkpoint (same as 03-01 D7)."

duration: 11min
completed: 2026-10-18
status: complete
---

# Phase 3 Plan 02: ⌘K Action Menu Summary

**A Raycast-style ⌘K popover anchored to the selected tile. It lists the selected GIF's actions in the locked order, filters as you type, and runs rows through the same `runAction` dispatcher the shortcuts use. It stays at least 8px inside the window from any column.**

## Performance

- **Duration:** about 11 min
- **Started:** 2026-10-18T00:26:57Z
- **Completed:** 2026-10-18T00:37:18Z
- **Tasks:** 2 (each RED, then GREEN)
- **Files modified:** 10

## Accomplishments
- `lib/actions.ts`: a static `ROWS` table in D-02 order. Each row has an `applies(ctx)` predicate and a label that can depend on the item (Clips read Copy MP4 / Paste MP4). `buildActions` omits rows that do not apply, never greys them. `filterActions` matches a trimmed, case-insensitive label substring. The full `ActionId` union is declared now. Plans 03-04 to 03-07 only add rows.
- `lib/popover.ts`: `placeMenu`. Columns 1 and 2 open 8px right of the tile, columns 3 and 4 open 8px left of it, and the result is clamped 8px inside the window.
- `ActionMenu.svelte`: a 320px `role=dialog` "Actions" with a `role=combobox` filter (`aria-activedescendant`, `aria-controls`) and a `role=listbox` of options with `aria-keyshortcuts`. Hint glyphs are `aria-hidden`. The highlighted row uses `--surface-active`. Group separators show only while the filter is empty. A single "No matching actions" row appears when nothing matches. The list has a 224px max-height, scrolls, and keeps the highlight in view. Rows with `destructive: true` get `--destructive` labels. There is no motion.
- `App.svelte`: `runAction(id)` is a closed switch and the only dispatcher. ↵, ⌘↵, ⌘⇧C, ⌘⇧M, ⌘O and the menu rows all go through it, and it closes the menu before anything runs. `toggleMenu` measures the menu after `tick()` and places it before paint. The menu closes on Esc, on ⌘K, on a mousedown outside the menu, on `panel:show`, on a tab switch and on a new result set. With the menu open, ⌘ keys typed in its filter bubble to `main` and reach `onKeydown`.
- Footer hints now read `↵ Copy   ⌘↵ Paste   ⌘K Actions`. `--shadow-popover` was added to the light and dark themes.

## Task Commits

1. **Task 1 (tracer): ⌘K, ↓, ↵ runs Copy Link**: `15dbdde` (test, RED), `12c56c4` (feat, GREEN)
2. **Task 2: filter, mouse, overflow, shortcuts while open, placement backstop**: `0ca25e0` (test, RED), `8b65d41` (feat, GREEN)

## TDD Gate Compliance

- **Task 1 RED:** `node --test --test-reporter=tap src/lib/actions.test.ts src/lib/popover.test.ts` ran against skeleton stubs (`buildActions` returned `[]`, `placeMenu` returned `0,0`). All 9 tests failed, every one on a `deepEqual` or `equal` assertion. `gsd-tools check tdd-red-evidence` returned `RED_EVIDENCE_OK` for the target test "GIPHY item lists every row in the locked order with hints and groups". Semantic assessment: every target test ran and failed on the planned behavior, with no load or syntax faults.
- **Task 2 RED:** this task has no new pure logic. Its behavior is DOM-level, so RED is the extended `08-action-menu.js` UAT. The run printed `UAT FAIL: no-match row missing` after the filter steps that already passed ("mark" leaves 2 rows, separators hidden, first match highlighted). The dev-browser console output is not TAP, JUnit, swift-testing or unittest, so the machine classifier cannot read it. Semantic assessment: the script ran to the first unimplemented behavior and failed on its own assertion, not on a load, timeout or setup fault. The output is kept at `bin/red/0302-t2.txt`, which is gitignored.
- GREEN commits follow each RED commit. No refactor commits were needed.

## Verification (latest run)

- `pnpm --dir frontend run test`: 36 pass, 0 fail. `svelte-check`: 0 errors, 0 warnings.
- `go vet . ./internal/...` and `go test -race ./internal/... -count=1` are green (no Go changes).
- `--uat` runs print UAT OK for: 08-action-menu, 02-keyboard, 07-text-actions, 02-grid, 03-states, 01-tracer, 05-tabs, 05-setup and 06-clips.
- Screenshots `~/.dev-browser/tmp/gifsmith-08-menu-col1.png` and `gifsmith-08-menu-col4.png` show the popover beside the bottom-row tile in columns 1 and 4, fully inside the window.

## Decisions Made
- Any mousedown outside the popover closes it. The plan also excluded the search input, but UI-SPEC says "a click outside the popover closes it", and closing on that click too is simpler. Focus returns to the search input either way.
- Rows highlight on `mousemove` instead of `mouseenter`. A menu that opens under a resting pointer would otherwise move the highlight off row 1 without the user doing anything.
- `placeMenu` gets the real `innerWidth` and `innerHeight` from App. 720×480 remains the default, and the panel is 720×480.
- ⌘K checks for `#tile-{selected}` in the DOM. That one check covers every "no item" state: empty, loading, message, S8 and the setup card.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Correctness] Tab is captured in the menu filter**
- **Found during:** Task 1
- **Issue:** Tab in the filter would move DOM focus out of the popover. That breaks the rule that focus stays in the filter while the menu is open.
- **Fix:** `ActionMenu` calls `preventDefault` on Tab. Filter keys inside an IME composition are also left to the IME, matching `routeKey`.
- **Committed in:** `12c56c4`

Otherwise the plan was executed as written. Task 2's planned `actions.test.ts` assertions (filter keeps order, no duplicate ids) were already written in Task 1's RED commit, so Task 2 added no node test.

## Issues Encountered
- None blocking. Each `--uat` run rebuilds the server binary, so the full regression sweep takes a few minutes.

## Known Stubs
None. Rows for download, square, favorite, details and recents are deliberately absent until their handlers exist (plans 03-04 to 03-07). `ActionId` already declares them, and `ActionCtx.tab`, `inDetails` and `favorited` are wired with provider-tab defaults.

## Threat Flags
None. The filter text only matches static labels in the webview (T-03-06). `runAction` is a closed switch over `ActionId`, fed only from the static `ROWS` table (T-03-07). No packages were installed (T-03-SC).

## User Setup Required
None.

## Next Phase Readiness
- Later plans add a `ROWS` entry with `applies`/`label`/`destructive` and a `case` in `runAction`. ActionMenu does not change.
- For the Details pane, App should call `placeMenu` with the content area's top-right anchor instead of the tile. ActionMenu takes `left` and `top` only.
- Native check (D6) moves to the plan 08 checkpoint.

## Self-Check: PASSED
