---
phase: 01-native-spike-signed-dev-loop
plan: 03
subsystem: native-shell
tags: [wails-v3, codesign, developer-id, hardened-runtime, accessibility, cgevent, cgo, objective-c, vite-proxy, server-mode]

requires:
  - phase: 01-01
    provides: "Platform interface + onMain helper, pasteboard file-URL copy, ActionService.Copy, shell.Show/Hide, stub and Fake"
  - phase: 01-02
    provides: "Tray menu, CursorScreenWorkArea and panelOrigin placement in Show (kept unchanged)"
provides:
  - "wails3 task darwin:install:dev: Developer ID sign (hardened runtime), install to ~/Applications/Gifsmith.app, launch with stderr to bin/install-dev.log"
  - "Stable designated requirement (identifier + team S2K7P43927, no cdhash) so the Accessibility grant survives rebuilds"
  - "Platform: FrontmostApp, ActivateApp, AccessibilityTrusted, OpenAccessibilitySettings, PasteShortcut (layout-aware ⌘V from a private CGEventSource; key-code lookup on the main thread)"
  - "ActionService.Paste (copy → AX gate → hide → 80 ms → reactivate previous app if needed → one ⌘V) and RequestAccessibility (button only)"
  - "⌘Enter UI with an 'Enable automatic paste…' button when Accessibility is missing"
  - "01-PASTE-MATRIX.md with the signed-dev columns filled from the user's gate"
  - "Browser UAT loop: ServerOptions on localhost:8089, Vite /wails proxy behind GIFSMITH_BACKEND, scripts/smoke-browser-loop.sh"
affects: [01-04 notarized build + go/no-go, Phase 2 browser UAT via dev-browser, Phase 3 paste action, Phase 4 Settings/onboarding (Accessibility copy)]

actuals:
  tokens: 7973   # chars/4 over git diff 8068996..2f1cb7e
  tasks: 2       # Task 2 and Task 3; Task 1 deferred by the user
  commits: 6
plan_head_before: 80689965119e2c9b47ad958e6b48ac5f20121cab
plan_head_after: 2f1cb7e7a7f1b572a15eeeb0973c4c4f1d0fa97a

tech-stack:
  added: []
  patterns:
    - "Every TIS/HIToolbox call goes through onMain: Wails runs bound methods on goroutines, and TISCopyCurrentKeyboardLayoutInputSource traps off the main thread"
    - "Dev installs are Developer ID + hardened runtime from day one, so dev matches the notarized build and TCC grants key on identifier + team"
    - "Browser UAT = CGO-free -tags server binary + Vite proxy; the stub's ErrUnsupported (HTTP 422) through the proxy is the loop's health signal"

key-files:
  created:
    - scripts/smoke-browser-loop.sh
    - .planning/phases/01-native-spike-signed-dev-loop/01-PASTE-MATRIX.md
  modified:
    - build/darwin/Taskfile.yml
    - internal/platform/platform.go
    - internal/platform/platform_darwin.go
    - internal/platform/platform_darwin.h
    - internal/platform/platform_darwin.m
    - internal/platform/platform_stub.go
    - internal/platform/fake.go
    - internal/services/actions.go
    - internal/services/actions_test.go
    - internal/shell/shell.go
    - main.go
    - frontend/src/App.svelte
    - frontend/vite.config.ts
    - frontend/bindings/gifsmith/internal/services/actionservice.ts
    - frontend/bindings/gifsmith/internal/services/models.ts

key-decisions:
  - "PasteDelay stays 80 ms: 5/5 ⌘Enter pastes into TextEdit landed; no tuning needed"
  - "No Cmd+Enter KeyBindings fallback: ⌘Enter reaches the page and ActionService.Paste"
  - "The untrusted-paste message points at the 'Enable automatic paste…' button, not a System Settings path: macOS 27.0.1 shows no Accessibility row under Privacy & Security; the list is reachable only via x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"
  - "vKeyCode (TIS + UCKeyTranslate) runs on main via onMain; event creation and posting stay on the calling goroutine"
  - "Task 1 (gifsmith-notary keychain profile) deferred by the user after Apple's site rejected sign-in; it must exist before 01-04"

patterns-established:
  - "Smoke scripts record their own PIDs, refuse busy ports, and trap EXIT to kill only those PIDs (no pkill/killall)"

requirements-completed: [ACTN-02, ACTN-01, SHEL-02]

coverage:
  - id: D1
    description: "One command builds, Developer ID-signs with hardened runtime, installs to ~/Applications/Gifsmith.app and launches"
    verification:
      - kind: other
        ref: "wails3 task darwin:install:dev && codesign --verify --deep --strict ~/Applications/Gifsmith.app && codesign -d -r- (identifier com.kenscott.gifsmith, OU S2K7P43927, no cdhash) && codesign -dvv shows (runtime)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Paste sequencing: no keystroke without Accessibility or on copy failure; hide before ⌘V; reactivate previous app when frontmost changed; prompt only from RequestAccessibility"
    requirement: ACTN-02
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestPasteWithoutAccessibility,TestPasteTrustedOrder,TestPasteReactivatesPrevious,TestPasteCopyError,TestPastePostError,TestRequestAccessibilityOpensSettings"
        status: pass
    human_judgment: false
  - id: D3
    description: "⌘Enter with Accessibility ON pastes the GIF into the previous app with no manual ⌘V; with it OFF, copies and explains"
    requirement: ACTN-02
    verification:
      - kind: manual_procedural
        ref: "01-PASTE-MATRIX.md N12, N14, Section 2 ⌘Enter (dev)"
        status: pass
    human_judgment: true
    rationale: "Real keystroke delivery to another app and Accessibility prompts need the user's own input; executors may not synthesize input"
  - id: D4
    description: "'Enable automatic paste…' button raises the prompt / opens the Accessibility list"
    requirement: ACTN-02
    verification:
      - kind: manual_procedural
        ref: "01-PASTE-MATRIX.md N13"
        status: unknown
    human_judgment: true
    rationale: "Button click not exercised in the re-test; only the URL it opens was confirmed. Re-check at 01-04 UAT"
  - id: D5
    description: "Browser UAT loop: bound ActionService.Copy reachable through the Vite /wails proxy against the CGO-free server build; script cleans up only its own processes"
    verification:
      - kind: integration
        ref: "bash scripts/smoke-browser-loop.sh → SMOKE OK; no listeners left on 8089/9245"
        status: pass
    human_judgment: false

duration: 75min
completed: 2026-10-17
status: complete
---

# Phase 1 Plan 03: Signed Dev Loop and ⌘Enter Paste Summary

**Developer ID-signed, hardened-runtime Gifsmith installs to ~/Applications with one command, ⌘Enter pastes the GIF straight into the previous app when Accessibility is granted (5/5 in TextEdit, plus Notes, Mail, Messages), and a CGO-free server build answers bound calls through the Vite proxy for browser UAT.**

## Performance

- **Duration:** about 75 min wall clock, including two human gate rounds
- **Started:** 2026-10-17T14:32Z (after 01-02)
- **Completed:** 2026-10-17T15:47Z
- **Tasks:** 2 of 3 executed (Task 1 deferred by the user)
- **Files modified:** 17

## Accomplishments

- `wails3 task darwin:install:dev` signs with `Developer ID Application: Ken Scott (S2K7P43927)` and `--hardened-runtime`, installs to `~/Applications/Gifsmith.app` and launches it.
- Designated requirement (exact, from `codesign -d -r-`):
  `designated => identifier "com.kenscott.gifsmith" and anchor apple generic and certificate 1[field.1.2.840.113635.100.6.2.6] /* exists */ and certificate leaf[field.1.2.840.113635.100.6.1.13] /* exists */ and certificate leaf[subject.OU] = S2K7P43927`
  It contains no cdhash, so the Accessibility grant should survive rebuilds (N15 is checked in 01-04).
- ACTN-02 shown on the signed build: with Accessibility OFF, ⌘Enter copies, pastes nothing and shows guidance (N12 PASS). With it ON, ⌘Enter hides the panel and the GIF lands in TextEdit with no manual ⌘V, 5/5 times (N14 PASS). Notes, Mail and Messages also PASS.
- Browser UAT loop: `bash scripts/smoke-browser-loop.sh` builds `bin/gifsmith-server` with `CGO_ENABLED=0 -tags server`, starts it and Vite, calls `ActionService.Copy` through `/wails/runtime` on the Vite port (HTTP 422 "not supported in this build"), checks `/sample.gif` returns 200, then prints `SMOKE OK`. Afterwards nothing is listening on 8089 or 9245. If a port is already taken, it exits 1 and leaves the other process alone (tested against my own listener on 18089).

## Task Commits

1. **Task 1: Store gifsmith-notary keychain profile**: DEFERRED by the user (Apple's site rejected sign-in). Open item, must be done before 01-04.
2. **Task 2: Tracer, signed install + ⌘Enter paste**: `ce0c96d` (feat), `4a92f37` (fix: path logging + contract test), `e9f1230` (fix: main-thread TIS + message), `44139c0` and `9e87ce9` (docs: matrix results)
3. **Task 3: Browser UAT loop**: `2f1cb7e` (feat)

**Plan metadata:** recorded in the final docs commit

## Files Created/Modified

- `build/darwin/Taskfile.yml`: `SIGN_IDENTITY`, dynamic `INSTALL_PATH`, `install:dev` (path guard before `rm -rf`, `pkill -x gifsmith` only)
- `internal/platform/*`: AX trust, ⌘V via private event source with layout-aware key code, frontmost app, activate, open AX settings; stub and Fake to match
- `internal/services/actions.go` / `actions_test.go`: `Paste`, `RequestAccessibility`, `PasteDelay`, six sequencing tests
- `internal/shell/shell.go`: records the frontmost app at Show and exposes `Previous()`
- `frontend/src/App.svelte`: ⌘Enter → Paste; "Enable automatic paste…" button
- `main.go`: `actions.Previous = sh.Previous`; `Server: application.ServerOptions{Host: "localhost", Port: 8089}`
- `frontend/vite.config.ts`: `/wails` proxy only when `GIFSMITH_BACKEND` is set
- `scripts/smoke-browser-loop.sh`: browser loop smoke
- `01-PASTE-MATRIX.md`: native checklist + per-app matrix, dev columns from the gate

## Gate Story (Task 2 tracer)

1. **First round (ce0c96d):** the user reported N13 FAIL. The cause was using plain Enter, which runs Copy, not ⌘Enter, so the button never showed. `4a92f37` added logging of which path ran and a test that pins the `needsAccessibility` JSON key.
2. **Wrong Settings path in the message:** the untrusted message said "System Settings › Privacy & Security › Accessibility", but macOS 27.0.1 shows no such row. `e9f1230` changed it to point at the button, and a test pins the new text. The orchestrator opened the `Privacy_Accessibility` URL and the user granted access from that list.
3. **Crash on first trusted ⌘Enter:** SIGTRAP in `TISCopyCurrentKeyboardLayoutInputSource`, because Wails runs bound methods on a goroutine, not the main thread. `e9f1230` moved the key-code lookup into `onMain`.
4. **Re-test:** the user reported "It all worked": ⌃⌥⌘G then ⌘Enter ×5 in TextEdit pasted every time, no crash, and the optional ⌘Enter checks worked. Gate approved.

Plan output items: final **PasteDelay = 80 ms**. **Cmd+Enter KeyBindings fallback: not needed.** **Keychain prompt on the first Developer ID codesign:** not recorded in the gate notes (`install:dev` ran to completion). **Gate outcome:** approved.

## Decisions Made

See `key-decisions` in the frontmatter. The one with lasting impact: Phase 4 onboarding must send the user to the Accessibility list via the `x-apple.systempreferences:…?Privacy_Accessibility` URL (button), never by describing a Settings path.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] ⌘V key-code lookup crashed off the main thread**
- **Found during:** Task 2 gate
- **Issue:** `TISCopyCurrentKeyboardLayoutInputSource` traps when called from a Wails binding goroutine
- **Fix:** `vKeyCode` runs via `onMain`; event posting is unchanged
- **Files modified:** internal/platform/platform_darwin.m
- **Commit:** e9f1230

**2. [Rule 1 - Bug] Untrusted message named a Settings path that doesn't exist on macOS 27**
- **Found during:** Task 2 gate
- **Fix:** the message points at the "Enable automatic paste…" button; a test pins the text
- **Files modified:** internal/services/actions.go, internal/services/actions_test.go
- **Commit:** e9f1230

**3. [Rule 2 - Diagnostics] Logging of which path ran (Copy / Paste / RequestAccessibility) to install-dev.log**
- **Found during:** Task 2 gate round 1 (to tell plain Enter apart from ⌘Enter)
- **Commit:** 4a92f37

**4. Task 1 deferred (user decision, not an executor deviation):** there is no `gifsmith-notary` keychain profile yet, so plan success criterion "notary profile exists for plan 01-04" is **not met**.

### Plan criteria not fully met

- The plan wants no PENDING cells in the dev columns. Still PENDING: N10 (Vorssaint, not re-checked on the signed build), N13 (button click unverified), and Slack, Discord, Telegram, and the Safari and Chrome composers in both dev columns (not tested). The orchestrator approved the gate with these open; they carry into the 01-04 UAT.

## Known Stubs

None that block this plan's goal. The `-tags server` platform stub returns `ErrUnsupported` on purpose: it is the browser loop's health signal and is never shipped.

## Open Items / Blockers

- **BLOCKER for 01-04:** create the `gifsmith-notary` keychain profile (app-specific password + `xcrun notarytool store-credentials gifsmith-notary --apple-id … --team-id S2K7P43927`). Apple's sign-in page rejected the user on 2026-10-17.
- Re-check N10, N13 and the untested apps at the 01-04 UAT.

## Self-Check: PASSED

- FOUND: scripts/smoke-browser-loop.sh (executable), 01-PASTE-MATRIX.md, ~/Applications/Gifsmith.app
- FOUND commits: ce0c96d, 4a92f37, e9f1230, 44139c0, 9e87ce9, 2f1cb7e (all ancestors of HEAD)
- `go test ./internal/... -count=1`: platform, services, shell ok; `wails3 task darwin:package` succeeded after the Vite change
