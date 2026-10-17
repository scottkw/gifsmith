---
phase: 01-native-spike-signed-dev-loop
plan: 02
subsystem: native-shell
tags: [wails-v3, systray, nsscreen, cgo, objective-c, window-placement]

requires:
  - phase: 01-01
    provides: "shell.Show/Hide, the non-activating panel, the Platform interface with its onMain helper, Fake, stub"
provides:
  - Menu bar tray (template icon) with Open Gifsmith / Settings… / separator / Quit Gifsmith, built in shell.New so it works even if hotkey registration fails
  - Settings… shows the panel with the status line "Settings window ships in Phase 4"
  - platform.Rect and Platform.CursorScreenWorkArea (darwin visibleFrame of the screen under the mouse, in Wails Y-down coordinates; stub Rect{}; Fake.WorkArea)
  - panelOrigin + TestPanelOrigin; Shell.Show places the panel centered with its top edge at 1/6 of the cursor display's visible height
affects: [01-03 paste into frontmost app + signing, 01-04 notarized build, Phase 4 settings window]

actuals:
  tokens: 1825   # chars/4 over git diff df9ec50..eb75f1b
  tasks: 2
  commits: 3
plan_head_before: df9ec501180af8102832afcd0e52c270bacc2deb
plan_head_after: eb75f1bc382a4e77e911f673c9ee43010acceace

tech-stack:
  added: []
  patterns:
    - "Tray built in shell.New (not in the ApplicationStarted hotkey handler), menu only, no attached window"
    - "Screen geometry crosses the Platform boundary already converted to Wails window coordinates (Y down from the top of NSScreen.screens[0]), so Go only does arithmetic"

key-files:
  created:
    - internal/shell/position.go
    - internal/shell/position_test.go
  modified:
    - internal/shell/shell.go
    - internal/platform/platform.go
    - internal/platform/platform_darwin.go
    - internal/platform/platform_darwin.h
    - internal/platform/platform_darwin.m
    - internal/platform/platform_stub.go
    - internal/platform/fake.go

key-decisions:
  - "App.svelte left unchanged: Wails sends frontend events through one serialized mailbox (events.go frontendEvents), so panel:show (clears status) always arrives before the panel:status that Settings… emits"
  - "gsCursorScreenVisibleFrame runs on main via onMain, uses screens[0].frame.height as the primary height (the same reference Wails windowSetPosition uses), and falls back to mainScreen when the cursor is on no screen"

patterns-established:
  - "Go test output goes through a go test -json to TAP adapter (scratchpad, not committed) so the GSD RED-evidence classifier can read it"

requirements-completed: [SHEL-04, SHEL-01]

coverage:
  - id: D1
    description: "Tray menu Open Gifsmith / Settings… / separator / Quit Gifsmith; app runs as menu-bar-only (no Dock, no Cmd-Tab)"
    requirement: SHEL-04
    verification:
      - kind: other
        ref: "pgrep -x gifsmith + 'tray menu ready' and 'hotkey registered: Ctrl+Option+Cmd+G' in bin/gifsmith-run.log after relaunch"
        status: pass
    human_judgment: true
    rationale: "Dock/Cmd-Tab absence, tray menu rendering (incl. template icon in light/dark menu bars) and clicking the items need real input, which executors may not synthesize"
  - id: D2
    description: "Settings… opens the panel with 'Settings window ships in Phase 4'; Open Gifsmith works independently of the hotkey; Quit removes the icon"
    requirement: SHEL-04
    verification: []
    human_judgment: true
    rationale: "Requires clicking tray items; no synthetic input allowed"
  - id: D3
    description: "panelOrigin: centered, top at one sixth of the work area (primary, secondary with negative Y, narrower-than-panel overhang)"
    requirement: SHEL-01
    verification:
      - kind: unit
        ref: "internal/shell/position_test.go#TestPanelOrigin"
        status: pass
    human_judgment: false
  - id: D4
    description: "Panel opens on the display under the mouse cursor, centered near the top (hotkey and tray)"
    requirement: SHEL-01
    verification:
      - kind: unit
        ref: "internal/shell/position_test.go#TestPanelOrigin (arithmetic only)"
        status: pass
    human_judgment: true
    rationale: "The darwin coordinate conversion cannot be unit-tested (go test has no main run loop for onMain) and placement on real displays is visual"

duration: 4min
completed: 2026-10-17
status: complete
---

# Phase 1 Plan 02: Tray menu and cursor-screen placement Summary

**Gifsmith now has a menu bar tray (Open Gifsmith / Settings… / Quit Gifsmith) that works without the hotkey. The panel opens on the display under the mouse, centered, with its top edge at one sixth of that display's visible height.**

## Performance

- **Duration:** about 4 min
- **Started:** 2026-10-17T14:27:30Z
- **Completed:** 2026-10-17T14:31:00Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments

- Tray menu built in `shell.New`. It uses the Wails `SystrayMacTemplate` icon, has no attached window, and Settings… stays enabled. The log now prints `tray menu ready`.
- New `Platform.CursorScreenWorkArea` returns the visible frame of the screen under the cursor in Wails window coordinates. The stub and Fake implement it too, so the `-tags server` build stays CGO-free.
- `Shell.Show` calls `SetPosition(panelOrigin(...))` instead of `Center()`. Everything else 01-01 put in Show is unchanged: Show, Focus, FocusWebView and the `panel:show` emit.

## Task Commits

1. **Task 1: Tracer, tray menu end to end:** `8c0981c` (feat)
2. **Task 2: Cursor-screen placement (TDD):** RED `80d2723` (test), GREEN `eb75f1b` (feat). No refactor was needed.

**Plan metadata:** the docs commit that follows this summary.

## Files Created/Modified

- `internal/shell/shell.go`: the tray menu, and placement in Show.
- `internal/shell/position.go`: `panelOrigin`, with a shortcut comment noting it does not clamp.
- `internal/shell/position_test.go`: `TestPanelOrigin`, a table test with three cases.
- `internal/platform/platform.go`: `Rect` and the `CursorScreenWorkArea() Rect` interface method.
- `internal/platform/platform_darwin.{go,h,m}`: `gsCursorScreenVisibleFrame` and its Go wrapper.
- `internal/platform/platform_stub.go`, `fake.go`: the stub returns `Rect{}`; the Fake records the call and returns `WorkArea`.

## Decisions Made

- **App.svelte unchanged.** I read the Wails beta.30 event path: frontend delivery goes through one serialized mailbox. So `panel:status` emitted after `Show()` always arrives after `panel:show`, and the existing handlers (clear on show, set on status) already show the line in the right order.
- **Primary-screen reference.** The conversion uses `NSScreen.screens[0].frame.size.height`. That is the same reference `windowSetPosition` in Wails uses, so the Rect values go straight into `SetPosition`.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] RED evidence format for Go**
- **Found during:** Task 2 (RED)
- **Issue:** The GSD RED-evidence classifier does not accept `go test -v` output (`invalid_record: Unsupported report format`). The plan also says the test "must fail before position.go exists", but a compile failure counts as INVALID_RED.
- **Fix:** The RED commit has `platform.Rect` and a `panelOrigin` stub that returns `(0, 0)`, so the test compiles and fails on its coordinate assertions. The real `go test -json` output went through a small mechanical adapter to TAP (`/private/tmp/claude-501/-Users-ken-dev-gif-search/9f53d405-1914-4df1-8002-37e6a551a349/scratchpad/gojson2tap.py`, not committed). The classifier returned `RED_EVIDENCE_OK` (`target_test_failed`, 4/4 failing).
- **Files modified:** internal/shell/position.go, internal/platform/platform.go (in the RED commit)
- **Committed in:** 80d2723

---

**Total deviations:** 1 auto-fixed (Rule 3). **Impact on plan:** affects only how RED was evidenced. No scope change.

## TDD Gate Compliance

- **RED** `80d2723` test(01-02): `go test ./internal/shell/ -run TestPanelOrigin -count=1 -json`, converted to TAP, gave verdict `RED_EVIDENCE_OK`. **semanticAssessment:** the target `TestPanelOrigin` ran, and all three subtests failed on the planned coordinate assertion: got `(0, 0)`, wanted `(495, 217)`, `(2630, 35)` and `(-60, 150)`. There were no compile, setup or unrelated failures.
- **GREEN** `eb75f1b` feat(01-02): `--- PASS: TestPanelOrigin` with all three subtests, and `go test ./internal/... -count=1` passes.
- **REFACTOR:** none.

## Issues Encountered

- `gsCursorScreenVisibleFrame` has no darwin unit test. It hops to main through `onMain`, and `go test` runs no main run loop, so the hop would deadlock. Only the arithmetic is unit-tested. The conversion is covered by the Task 2 human check.

## Human-check outcomes

Deferred to end-of-phase UAT (`human_verify_mode: end-of-phase`). Nothing was observed visually and no synthetic input was sent. These are still unconfirmed:
- Task 1: no Dock icon and absent from Cmd-Tab; the tray menu shows the 4 entries; Open Gifsmith shows the panel; Settings… shows the panel with "Settings window ships in Phase 4".
- Whether the tray icon renders as a template icon in both light and dark menu bars.
- Task 2: hotkey placement on the cursor's display; placement on a second display, and any coordinate surprise there; Quit Gifsmith removes the icon.

Automated checks passed: package, `go vet`, svelte-check (0 errors), `-tags server` CGO-free build, `go test ./internal/...`, gofmt clean, and after relaunch `pgrep -x gifsmith` plus `tray menu ready` and `hotkey registered: Ctrl+Option+Cmd+G` in `bin/gifsmith-run.log`.

## User Setup Required

None.

## Next Phase Readiness

- 01-03 can add FrontmostApp/ActivateApp/Accessibility to the Platform interface the same way as `CursorScreenWorkArea`: interface method, darwin, stub and Fake.
- The rebuilt ad-hoc `bin/gifsmith.app` is running (PID 95572) for UAT.

---
*Phase: 01-native-spike-signed-dev-loop*
*Completed: 2026-10-17*

## Self-Check: PASSED

position.go and position_test.go present; commits 8c0981c, 80d2723, eb75f1b are ancestors of HEAD.
