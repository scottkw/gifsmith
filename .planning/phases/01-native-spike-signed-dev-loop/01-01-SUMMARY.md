---
phase: 01-native-spike-signed-dev-loop
plan: 01
subsystem: native-shell
tags: [wails-v3, svelte5, pnpm, cgo, objective-c, nspasteboard, nspanel, global-hotkey]

requires: []
provides:
  - Wails v3.0.0-beta.30 Svelte/TS scaffold at repo root with locked identity (com.kenscott.gifsmith, LSUIElement, macOS 13.0, arm64)
  - internal/platform Platform interface (CopyFile, FocusWebView), darwin cgo impl, CGO-free stub, Fake
  - internal/services ActionService.Copy bound to the frontend
  - internal/shell non-activating 720x480 panel toggled by Ctrl+Option+Cmd+G
  - Proven on this Mac: typing into the non-activating panel, full-screen overlay, Enter -> file-URL copy -> Cmd+V animated GIF in TextEdit without Accessibility
affects: [01-02 tray + cursor-screen placement, 01-03 paste into frontmost app + signing, 01-04 notarized build]

actuals:
  tokens: 89095   # chars/4 over git diff 44da990..HEAD excluding pnpm-lock.yaml (95144 including it)
  tasks: 3
  commits: 3
plan_head_before: 44da990688743a5e76cefbf7e467bcccdcd10e84
plan_head_after: 797032f40f700ed988b77eebd9f46783b9fbee76

tech-stack:
  added: [wails v3.0.0-beta.30 (Go module + wails3 CLI), "@wailsio/runtime 3.0.0-beta.30 (exact)", svelte 5.57.2, vite 8.3.4, "@sveltejs/vite-plugin-svelte 7.3.1", svelte-check 4.7.7, typescript 5.9.3]
  patterns:
    - "Platform interface in internal/platform; darwin && !server cgo impl, !darwin || server stub, Fake for service tests"
    - "Pasteboard calls run on the calling goroutine's thread inside an autorelease pool; only AppKit view work (FocusWebView) hops to main via CFRunLoopPerformBlock + semaphore"
    - "Show = Center + Show + Focus + FocusWebView + emit panel:show (Show alone never makes a non-activating panel key)"
    - "Darwin pasteboard tests use uniquely named private pasteboards released in t.Cleanup; never the general pasteboard"

key-files:
  created:
    - main.go
    - internal/platform/platform.go
    - internal/platform/platform_darwin.go
    - internal/platform/platform_darwin.h
    - internal/platform/platform_darwin.m
    - internal/platform/platform_stub.go
    - internal/platform/fake.go
    - internal/platform/platform_darwin_test.go
    - internal/services/actions.go
    - internal/services/actions_test.go
    - internal/shell/shell.go
    - frontend/src/App.svelte
    - frontend/public/sample.gif
    - frontend/bindings/gifsmith/internal/services/
  modified:
    - .gitignore
    - README.md
    - Taskfile.yml
    - build/config.yml
    - build/darwin/Info.plist
    - build/darwin/Info.dev.plist
    - build/darwin/Taskfile.yml

key-decisions:
  - "No pre-authorized fallback was needed: NonActivating panel + Focus + makeFirstResponder(WKWebView) takes typing; MacWindowLevelFloating + CanJoinAllSpaces|FullScreenAuxiliary shows over full-screen; HideOnFocusLost handles click-away (D-08 fallback, PopUpMenu level, and NSEvent mouse monitor all NOT applied)"
  - "Pasteboard writes stay off the main thread (autorelease pool on the calling thread); only FocusWebView dispatches to main"
  - "Frontend page <title> set to Gifsmith (replaces template title)"

patterns-established:
  - "Platform interface + build-tag split (darwin && !server / !darwin || server) keeps -tags server CGO-free"
  - "Bound service methods take no path arguments; Go computes the only copyable path (T-01-01)"

requirements-completed: [SHEL-01, SHEL-02, SHEL-03, ACTN-01]

coverage:
  - id: D1
    description: "Locked bundle identity: com.kenscott.gifsmith, LSUIElement true, LSMinimumSystemVersion 13.0, arm64-only, Wails beta.30 pinned, pnpm lockfile"
    verification:
      - kind: other
        ref: "Task 2 verify: wails3 task darwin:package + plutil/lipo/file checks"
        status: pass
    human_judgment: false
  - id: D2
    description: "Pasteboard file-URL write (single URL after double copy) and missing-path error"
    requirement: ACTN-01
    verification:
      - kind: unit
        ref: "internal/platform/platform_darwin_test.go#TestCopyFileWritesFileURL"
        status: pass
      - kind: unit
        ref: "internal/platform/platform_darwin_test.go#TestCopyFileMissingPath"
        status: pass
    human_judgment: false
  - id: D3
    description: "ActionService.Copy hides on success, keeps panel open and returns the error on failure"
    requirement: ACTN-01
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopyCopiesThenHides"
        status: pass
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopyErrorKeepsPanelOpen"
        status: pass
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopyTwiceUsesSamePath"
        status: pass
    human_judgment: false
  - id: D4
    description: "Hotkey toggles a non-activating panel that takes typing without a click, hides on Esc and click-away, appears over a full-screen app; previous app stays frontmost"
    requirement: SHEL-01
    verification:
      - kind: manual_procedural
        ref: "Task 3 tracer gate steps 1-7 and 9 (user approved 2026-10-17)"
        status: pass
    human_judgment: true
    rationale: "Key-window/first-responder behavior, full-screen overlay and click-away cannot be observed without synthetic input, which is forbidden on the user's machine"
  - id: D5
    description: "Enter copies the staged sample GIF; Cmd+V pastes it animated into TextEdit with Accessibility not granted"
    requirement: ACTN-01
    verification:
      - kind: manual_procedural
        ref: "Task 3 tracer gate step 8 (user approved) + osascript -e 'clipboard info' -> «class furl», 73"
        status: pass
    human_judgment: true
    rationale: "Real paste rendering in another app needs a human"

duration: 22min
completed: 2026-10-17
status: complete
---

# Phase 1 Plan 01: Native Tracer Summary

**Wails v3 beta.30 + Svelte 5 app at the repo root with the locked identity. ⌃⌥⌘G toggles a non-activating 720×480 NSPanel that takes typing with no click. Enter writes the staged sample GIF to NSPasteboard as a file URL, and ⌘V pastes it animated into TextEdit with Accessibility not granted. The user confirmed this at the tracer gate.**

## Performance

- **Duration:** about 22 min of execution (plan commit 14:02Z to this summary 14:24Z, which includes the wait at the user gate)
- **Started:** 2026-10-17T14:02:33Z
- **Completed:** 2026-10-17T14:24:05Z
- **Tasks:** 3/3
- **Files modified:** 111 (excluding node_modules), 116 in the git diff including generated bindings

## Accomplishments

- **Task 1, package gate:** the user replied "approved" for @wailsio/runtime@3.0.0-beta.30 (exact pin), svelte, vite, @sveltejs/vite-plugin-svelte and svelte-check. @tsconfig/svelte, typescript and tslib were already approved during research.
- **Task 2, scaffold:** the scaffold builds with pnpm into an arm64 bundle with LSUIElement, com.kenscott.gifsmith (dev plist: com.kenscott.gifsmith.dev) and minimum macOS 13.0. The deployment target is pinned in build:native. `/graft/` is kept in .gitignore.
- **Task 3, tracer:** platform, services and shell are wired end to end. The darwin pasteboard tests run on private pasteboards only. The `-tags server` build works without cgo.
- **Versions:** `wails3 version` prints `v3.0.0-beta.30`. The lockfile resolved @wailsio/runtime 3.0.0-beta.30, svelte 5.57.2, vite 8.3.4, @sveltejs/vite-plugin-svelte 7.3.1, svelte-check 4.7.7 and typescript 5.9.3.

## Tracer Gate Outcome (Task 3, user "approved")

| Step | Behavior | Result |
|------|----------|--------|
| 1-2 | ⌃⌥⌘G shows the panel; TextEdit stays the active app | pass |
| 3 | Typing "cats" without clicking lands in the input | pass |
| 4 | Esc hides the panel | pass |
| 5 | Pressing the hotkey twice shows, then hides | pass |
| 6 | Clicking the TextEdit window hides the panel | pass |
| 7 | Panel appears over a full-screen app | pass |
| 8 | Enter hides the panel; ⌘V pastes the animated GIF into TextEdit with no Accessibility | pass |
| 9 | With Vorssaint frontmost, the hotkey toggles every time and nothing else reacts | pass |

Read-only check after the gate: `osascript -e 'clipboard info'` returned `«class furl», 73`. The clipboard still held the GIF file URL.

**Pre-authorized fallbacks:** none applied. D-08 (activating panel plus reactivating the previous app), MacWindowLevelPopUpMenu plus Stationary, and the NSEvent global mouse monitor were all unnecessary because every gate step passed with the primary design. The matching RESEARCH.md Open Questions are resolved in favor of the primary design.

## Task Commits

1. **Task 1: Package legitimacy gate.** No commit (approval only).
2. **Task 2: Scaffold Wails v3 app with locked identity.** `c7d20f2` (chore)
3. **Task 3: Tracer: hotkey panel, typing, Enter copies sample GIF.** `f2e5005` (feat)
4. **Tidy: range-over-int in the double-copy test loops (gopls hint).** `797032f` (style)

**Plan metadata:** the docs commit that follows this summary.

## Files Created/Modified

- `main.go`: embeds frontend/dist, stages sample.gif to `~/Library/Caches/com.kenscott.gifsmith/staging/`, runs as an accessory app, wires shell and ActionService.
- `internal/platform/*`: the Platform interface, the cgo/ObjC NSPasteboard file-URL write, WKWebView first responder, the CGO-free stub and Fake.
- `internal/services/actions.go`: ActionService.Copy (copies, then hides; on error the panel stays open).
- `internal/shell/shell.go`: panel options, hotkey registration with logging, Toggle/Show/Hide.
- `frontend/src/App.svelte`: input, sample preview and status line; Enter calls ActionService.Copy.
- `build/config.yml`, `build/darwin/Info*.plist`, `build/darwin/Taskfile.yml`, `Taskfile.yml`: identity, macOS 13.0 target, pnpm.

## Decisions Made

- No fallback was needed (see above). The non-activating panel design is confirmed.
- Pasteboard calls run on the calling thread inside an `@autoreleasepool`. Only FocusWebView hops to the main run loop. NSPasteboard is thread-safe, and keeping it off main avoids blocking on the UI thread.
- The page `<title>` in frontend/index.html is "Gifsmith" (replaces the template title).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `wails3 version` prints to stderr**
- **Found during:** Task 2
- **Issue:** `wails3 version | grep` saw nothing because the version goes to stderr.
- **Fix:** verified with `wails3 version 2>&1 | grep -q v3.0.0-beta.30`.
- **Committed in:** n/a (verify-command change only)

**2. [Rule 1 - Bug] Pasteboard threading**
- **Found during:** Task 3
- **Issue:** the plan's `onMain` helper applied to every native call. For pasteboard work that is not needed and risks a deadlock or blocking when called from a binding goroutine.
- **Fix:** pasteboard calls run on the calling thread inside an autorelease pool. Only `gsFocusWebView` uses `onMain`.
- **Files modified:** internal/platform/platform_darwin.m
- **Committed in:** f2e5005

**3. [Rule 2 - Missing critical] FocusWebView nil-guard**
- **Found during:** Task 3
- **Issue:** `WebviewWindow.NativeWindow()` returns nil before the app runs; passing it through would hand cgo a nil NSWindow.
- **Fix:** the Go `FocusWebView` returns early on a nil pointer (internal/platform/platform_darwin.go).
- **Committed in:** f2e5005

**4. [Rule 2 - Missing critical] Double-copy covered at the darwin level**
- **Found during:** Task 3
- **Issue:** the plan asserted "exactly one URL after two copies" only through the Fake, which cannot see clearContents.
- **Fix:** `TestCopyFileWritesFileURL` writes twice to the private pasteboard and asserts exactly one URL.
- **Committed in:** f2e5005

**5. [Minor] Page title "Gifsmith".** When Task 3 cleaned up frontend/index.html, it also set `<title>Gifsmith</title>`. The plan only asked for the stylesheet link and bg div to be removed. Committed in f2e5005.

---

**Total deviations:** 5 (1 blocking, 1 bug, 2 missing critical, 1 minor). **Impact:** all are small and local, with no scope creep.

## Issues Encountered

- Unused template assets in frontend/public (bg-desktop.jpg, bg-mobile.jpg, Inter-Medium.ttf, style.css, svelte.svg, wails.png) were kept. They do no harm, and deleting them was out of scope for the tracer. Plan 01-02 or the UI phase can prune them.
- Harmless build and run noise: a `sandbox_extension_consume` log line at launch (WKWebView, not sandboxed) and a duplicate `-lobjc` linker warning from cgo plus Wails both adding it.

## Known Stubs

- `frontend/src/App.svelte` always shows and copies the fixed `sample.gif`, and the input does not search yet. This is intentional for the Phase 1 spike (D-03). Real search arrives in a later phase per ROADMAP.

## User Setup Required

None.

## Next Phase Readiness

- 01-02 (tray menu, cursor-screen placement) can build directly on `shell.Show` and its `Center` call.
- 01-03 extends Platform with FrontmostApp/ActivateApp/Accessibility and adds Developer ID signing. The ad-hoc build is not signed yet, so TCC grants made now would not carry over.
- The running app (bin/gifsmith.app, PID 84366) was left running.

---
*Phase: 01-native-spike-signed-dev-loop*
*Completed: 2026-10-17*

## Self-Check: PASSED

All 10 key files/dirs present; commits c7d20f2, f2e5005, 797032f are ancestors of HEAD; `go test ./internal/... -count=1` green (5/5) after the tidy; `gofmt -l main.go internal` clean.
