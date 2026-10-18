---
phase: 04-settings-shell-polish-release
plan: 01
subsystem: settings
tags: [wails-v3, svelte5, settings-window, nsappearance, smappservice, autostart, prefs, theme]

requires:
  - phase: 01-native-spike
    provides: shell panel, tray, platform cgo seam (onMain, FrontmostApp, ActivateApp, AccessibilityTrusted)
  - phase: 02-providers-search-ui-first-run
    provides: SettingsService (SaveKey, LastProvider), WriteJSONAtomic, setup card, smoke/UAT loop
  - phase: 03-action-pipeline-library
    provides: ActionService.RequestAccessibility, ActionError/MarshalError, footer GIPHY mark
provides:
  - settings.Store (single settings.json owner; OpenStore/Get/Update, Prefs.Normalize; Theme, LaunchAtLogin)
  - SettingsService Snapshot / SetPref / SetLaunchAtLogin / AccessibilityStatus / ApplyStartup; settings:changed event with the full Snapshot
  - services.SettingsError (invalid_value, prefs_write_failed, autostart_failed, not_in_applications) through MarshalError
  - services.Autostarter seam (app.Autostart in the app, memAutostart in the server build, FAKE_AUTOSTART=fail)
  - platform.SetAppearance (gsSetAppearance via onMain; stub; Fake)
  - shell Settings window (hidden at launch, 640x520), OpenSettings, HideSettings, ApplyTheme; settings:show event
  - PanelService.OpenSettings and the Cmd+, route in the panel
  - frontend settings.html second Vite entry (rolldownOptions), SettingsApp (PANES), SettingRow, GeneralPane
  - lib/theme.ts (resolveTheme, applyTheme, onTheme); data-theme tokens; --control-border, --destructive-fill; shared .btn-primary/.btn-secondary/.btn-destructive
  - scripts/uat/15-settings-general.js, scripts/uat/15-settings-startup-fail.js
  - 04-KLIPY-EMAIL.md (Klipy clearance draft + Apple agreement steps)
affects: [04-02, 04-03, 04-04, 04-05, 04-06]

actuals:
  tokens: 15700    # chars/4 over the added lines of the realized diff (62982 chars); full content of the 42 changed files is 212509 chars (~53100)
  tasks: 7
  commits: 13
plan_head_before: a2444daeb71700a7d222188e57be832ff5859d27
plan_head_after: e1663a86677b303e0851b090c8b2696f16c813e4

tech-stack:
  added: []
  patterns:
    - "Prefs owner: Store.Update(copy -> mutate -> Normalize -> WriteJSONAtomic -> swap) under one mutex; SetPref validates through Normalize inside Update and treats an equal value as a no-op"
    - "Every settings write emits settings:changed with the full Snapshot; both pages apply it"
    - "Snapshot.launchAtLogin is the OS truth (Autostart.IsEnabled), not the pref"
    - "<html data-theme> holds the resolved theme before mount; data-theme is the only CSS switch"
    - "Shell window callbacks never take s.mu; a separate vmu guards settingsPrev/theme and is never held across a window call"

key-files:
  created:
    - internal/settings/settings.go (Store section)
    - frontend/settings.html
    - frontend/src/settings/main.ts
    - frontend/src/settings/SettingsApp.svelte
    - frontend/src/settings/SettingRow.svelte
    - frontend/src/settings/GeneralPane.svelte
    - frontend/src/lib/theme.ts
    - frontend/src/lib/theme.test.ts
    - scripts/uat/15-settings-general.js
    - scripts/uat/15-settings-startup-fail.js
    - .planning/phases/04-settings-shell-polish-release/04-KLIPY-EMAIL.md
  modified:
    - internal/services/settings.go
    - internal/services/errors.go
    - internal/services/panel.go
    - internal/platform/platform.go
    - internal/platform/platform_darwin.go
    - internal/platform/platform_darwin.m
    - internal/platform/platform_darwin.h
    - internal/platform/platform_stub.go
    - internal/platform/fake.go
    - internal/shell/shell.go
    - internal/shell/position.go
    - main.go
    - wire_server.go
    - scripts/smoke-browser-loop.sh
    - frontend/vite.config.ts
    - frontend/src/main.ts
    - frontend/src/App.svelte
    - frontend/src/app.css
    - frontend/src/components/Footer.svelte
    - frontend/src/components/SetupCard.svelte
    - frontend/src/lib/api.ts
    - frontend/src/lib/grid.ts

key-decisions:
  - "Vite 8 build.rolldownOptions.input accepted the second entry (no rollupOptions fallback needed)"
  - "LoadPrefs/SavePrefs removed and SettingsService.Dir dropped: their only callers moved to settings.Store"
  - "SetLastProvider write failures now return the fixed prefs_write_failed error instead of a raw os error (no path reaches the webview)"
  - "GeneralPane listens for settings:show itself instead of SettingsApp re-dispatching it (one listener, same effect)"
  - "ApplyStartup follows the plan rule (pref on and OS off -> Enable); the requires-approval case (user switched it off in System Settings) is flagged as risk A4 for native check"

patterns-established:
  - "Settings panes: PANES table in SettingsApp, each pane a stack of .settings-section blocks of SettingRow rows"
  - "Bound settings inputs are allow-listed names and enum strings validated in Go; JS only ever sees Snapshot fields"

requirements-completed: [SETT-01, SETT-05, SETT-06, SHEL-05]

coverage:
  - id: D1
    description: "settings.json has one owner (settings.Store); tab switches no longer erase other settings; concurrent writers never lose fields; unknown enum values reset with a log line"
    requirement: SETT-06
    verification:
      - kind: unit
        ref: "internal/settings/settings_test.go#TestStoreUpdateKeepsOtherFields, TestStoreConcurrentUpdates, TestStoreUnknownEnumResets, TestStoreWriteFailureKeepsPrefs; internal/services/settings_test.go#TestSetLastProviderKeepsTheme, TestSetPrefConcurrentWithLastProvider (go test -race)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Theme System/Light/Dark set in the Settings page persists and reaches both pages via settings:changed; dark tokens via data-theme; GIPHY mark follows the resolved theme"
    requirement: SETT-06
    verification:
      - kind: unit
        ref: "internal/services/settings_test.go#TestSetPrefTheme, TestSetPrefRejectsInvalid; frontend/src/lib/theme.test.ts"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/15-settings-general.js"
        status: pass
    human_judgment: false
  - id: D3
    description: "Native appearance and window backgrounds follow the theme (NSApp.appearance), System follows macOS live, no white flash"
    requirement: SETT-06
    verification:
      - kind: unit
        ref: "internal/platform/platform_darwin_test.go#TestAppearanceCode, TestFakeSetAppearanceRecords"
        status: pass
    human_judgment: true
    rationale: "NSApp.appearance, WKWebView prefers-color-scheme and window background colour only exist in the signed native app (Assumption A3)"
  - id: D4
    description: "Settings window opens from the tray Settings... item and Cmd+, in the panel (panel hides first), activates the app, keeps its position and pane; Cmd+W/red button hide it and return focus"
    requirement: SETT-01
    verification:
      - kind: unit
        ref: "internal/shell/position_test.go#TestSettingsOrigin; frontend/src/lib/grid.test.ts#routeKey maps Cmd+, to settings"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/15-settings-general.js (sidebar keyboard, pane scroll, Cmd+, in the panel page)"
        status: pass
    human_judgment: true
    rationale: "Accessory-app activation, key window, Cmd+W and focus return are AppKit behaviour the server build cannot exercise (A1, A2, A6)"
  - id: D5
    description: "Open Gifsmith at login: off by default, OS state shown, translocation/DMG guard (G3), failures revert with G2, startup re-apply"
    requirement: SHEL-05
    verification:
      - kind: unit
        ref: "internal/services/settings_test.go#TestSetLaunchAtLogin, TestSetLaunchAtLoginNotInApplications, TestSetLaunchAtLoginFailures, TestApplyStartup, TestSnapshotLaunchAtLoginIsOSState"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/15-settings-general.js; FAKE_AUTOSTART=fail scripts/uat/15-settings-startup-fail.js"
        status: pass
    human_judgment: true
    rationale: "SMAppService registration and Login Items visibility need the signed bundle in /Applications (A4, A5)"
  - id: D6
    description: "Automatic paste status On/Off with G4/G5, Grant Access... reuses RequestAccessibility, re-read on open and every 2 s while Off"
    requirement: SETT-05
    verification:
      - kind: unit
        ref: "internal/services/settings_test.go#TestAccessibilityStatus"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/15-settings-general.js (Off + Grant Access... + G5 on platform.Fake)"
        status: pass
    human_judgment: true
    rationale: "The TCC prompt and the Off->On flip after granting only happen on a real signed bundle"
  - id: D7
    description: "Klipy clearance email draft and Apple Developer agreement steps for the user"
    verification: []
    human_judgment: true
    rationale: "The user must send the email and accept the agreement; the executor never sends anything"

duration: 22min
completed: 2026-10-18
status: complete
---

# Phase 4 Plan 01: Settings Window, General Pane and settings.Store Summary

**A real Settings window (tray Settings… and Cmd+,) whose General pane sets Theme (NSApp.appearance + data-theme on both pages), Open at login (Wails Autostart with a translocation guard) and shows Automatic paste status, all on a new single-owner settings.Store that fixes the tab-switch overwrite bug.**

## Performance

- **Duration:** 22 min
- **Started:** 2026-10-18T20:48:26Z
- **Completed:** 2026-10-18T21:11:10Z
- **Tasks:** 7 (tracer + 6)
- **Files modified:** 42 (incl. 4 regenerated binding files)

## Accomplishments

- `settings.Store` owns settings.json: copy, mutate, normalise, atomic write, swap, all under one mutex. `SetLastProvider` no longer erases other fields. Unknown values reset with a log line naming the field.
- `SetPref("theme", …)` validates in Go, does nothing on an equal value, calls `ApplyTheme` and emits `settings:changed` with the full Snapshot. Both webviews apply it through `lib/theme.ts` before mount and live, so neither flashes. The footer GIPHY mark is chosen in script.
- Native side: `platform.SetAppearance` sets NSApp.appearance via `onMain`. The shell creates a hidden 640×520 Settings window. The red button and ⌘W hide it and give focus back to the app that was frontmost when it opened. The first show of a launch is placed on the display under the cursor. `ThemeChanged` refreshes the window backgrounds while the theme is System.
- ⌘, works from every panel context (search, setup card key field, action menu) through `PanelService.OpenSettings`. The tray placeholder toast is gone.
- Open at login: `Autostarter` seam, `/AppTranslocation/` and `/Volumes/` guard (G3), failure reverts with G2, the checkbox shows the OS state, and `ApplyStartup` runs at ApplicationStarted. The server build uses an in-memory fake (`FAKE_AUTOSTART=fail`).
- Automatic paste row: On/Off, `Grant Access…`, re-read on open and polled every 2 s while Off.
- Shared Settings scaffolding for later plans: `SettingsApp` (PANES, roving sidebar, clamped ↑/↓), `SettingRow`, `.settings-section`, shared buttons, `--control-border`, `--destructive-fill`.

## Task Commits

1. **Task 1 (tracer): Theme → Store → settings:changed.** `922945d` (test, RED), `f1bd8b5` (feat). Tracer verified end to end, then expansion continued.
2. **Task 2: resolved theme on both pages.** `c7d408a` (test, RED), `f71a649` (feat)
3. **Task 3: native Settings window and appearance.** `379a42a` (test, RED), `475a572` (feat)
4. **Task 4: ⌘, opens Settings.** `68dabca` (test, RED), `c033a85` (feat)
5. **Task 5: Mac settings-window behaviour and shared controls.** `90dbb88` (test, RED), `979e556` (feat)
6. **Task 6: Open at login and AX status in Go.** `96de1aa` (test, RED), `8a4590b` (feat)
7. **Task 7: Startup and Automatic paste rows, Klipy/Apple reminders.** `e1663a8` (feat; no `<behavior>`, UI and doc only)

## TDD Gate Compliance

- Every TDD task has a `test(04-01)` RED commit before its `feat(04-01)` GREEN commit (6 pairs). Task 7 is not TDD.
- RED evidence, frontend (TAP): `gsd-tools check tdd-red-evidence` returned `RED_EVIDENCE_OK` for "System follows the OS appearance" (theme.test.ts) and "routeKey maps Cmd+, to settings, not plain or shifted comma" (grid.test.ts). Semantic check: each target failed on its planned assertion against a skeleton (`'light'`, `'none'`), with no load errors.
- RED evidence, Go: `go test` output is not a format the classifier accepts (TAP, JUnit, swift-testing, unittest). Semantic assessment only. Every new Go target compiled against skeletons and failed on its planned assertion: Store not writing, SetPref returning the zero Snapshot, appearanceCode returning -1, settingsOrigin returning (0,0), SetLaunchAtLogin not calling Enable.
- Two RED cases passed early, both expected: `TestStoreLaunchAtLoginRoundTrip` (adding the `LaunchAtLogin` field is all it needs, and the test must compile) and `TestSnapshotLaunchAtLoginIsOSState` (the skeleton snapshot never reported true). Both remain as regression guards.

## Verification (plan level)

- `gofmt -l` is clean. `go vet . ./internal/...` passes. `go test -race ./internal/... -count=1` passes in all 13 packages with no data race.
- `CGO_ENABLED=0 go build -tags server .` succeeds. `node --test` reports 67 pass, 0 fail. svelte-check reports 0 errors and 0 warnings. `vite build` produces `dist/settings.html`.
- `golangci-lint run . ./internal/...` reports 0 issues.
- `scripts/smoke-browser-loop.sh` ends with SMOKE OK. All 20 `scripts/uat/*.js` print UAT OK, including 15-settings-general, 15-settings-startup-fail, 01-tracer, 05-setup and 06-backstops.
- go.sum is unchanged. Bindings were regenerated with `wails3 generate bindings -clean=true -ts -i`.

## Decisions Made

See key-decisions in the frontmatter. In short: Vite 8 accepted `rolldownOptions`; `LoadPrefs`, `SavePrefs` and `SettingsService.Dir` were removed because they had no other callers; prefs write errors are fixed-text SettingsErrors.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Security, T-04-03] SetLastProvider returns a fixed error**
- **Found during:** Task 1
- **Issue:** The old code returned the raw `os` error, which can carry a path, to the webview.
- **Fix:** It now logs `settings write failed: lastProvider` and returns `prefs_write_failed`.
- **Files:** internal/services/settings.go. **Commit:** f1bd8b5

### Plan adjustments (no behaviour lost)

- `gsSetAppearance` is `void` rather than `int`. It cannot fail once the mode is mapped, and the Go wrapper rejects an unknown mode with a fixed error before any cgo call.
- `settings:show` is handled by GeneralPane's own listener instead of being re-dispatched by SettingsApp. SettingsApp tracks `settings:changed` so a re-mounted pane gets the latest snapshot.
- The shared focus ring gains only `select:focus-visible`. `input:focus-visible` already covered checkboxes and radios.
- The sidebar tablist has `tabindex="-1"`, which svelte-check's a11y rule requires. It is not a tab stop; the selected tab is.
- The theme-save failure copy `Couldn't save this setting. Try again.` is the planner default noted in the plan; UI-SPEC has no row for it.

**Total deviations:** 1 auto-fixed (Rule 2), 5 minor adjustments. **Impact:** none on scope.

## Issues Encountered

- `.planning/REQUIREMENTS.md` (RELS-05/06 added) and `04-CONTEXT.md` were changed by someone else while this plan ran. They are not part of this plan, and this executor did not stage them.

## Known Stubs

None that block the plan's goal. On purpose, `PANES` in SettingsApp.svelte lists only General; plans 04-02..04-04 add Keys, Hotkeys, Search, Downloads and About.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: bound-surface | internal/services/settings.go | `ApplyStartup` is exported, so Wails binds it and the webview can call it. It is idempotent: it can only re-register a login item the user already turned on, never disable one or register one that is off. |

## Human-check items deferred to end of phase (native, user-verified on the signed install)

1. Click the tray icon, then Settings…. Separately, open the panel with ⌃⌥⌘G and press ⌘,. Each time, Settings should have an active (coloured) title bar and take typing; ⌘, should hide the panel first (A1).
2. ⌘W and the red button hide Settings, and the app you were in comes back to the front (A2). Esc does not close Settings.
3. Theme Dark, Light and System switch the panel, Settings, their native backgrounds and form controls together with no white flash. System follows a live macOS appearance change. Watch the log for a ThemeChanged → ApplyTheme("") loop or flicker (A3).
4. Check that 640×520 is the content size and that the first show is placed on the display under the cursor. Placement reuses panelOrigin: top edge one sixth down, not vertically centred (A6).
5. Check "Open Gifsmith at login" and confirm Gifsmith appears in System Settings → General → Login Items; uncheck it and confirm it is removed. From the DMG, G3 should show (A5).
6. A4 risk: turn Gifsmith off in Login Items while the pref is on, then relaunch. `ApplyStartup` will call Enable. Confirm whether macOS re-enables it or leaves it needing approval.
7. With Accessibility off, Grant Access… raises the system prompt or the Accessibility list. After you grant access, the row should show "✓ On" within about 2 s.
8. Send the Klipy email and accept the Apple agreement (`04-KLIPY-EMAIL.md`).

## Not done (outside this plan's tasks)

- The tray menu still shows no shortcut text on its items. UI-SPEC wants "⌘," on Settings… and the Open hotkey on Open Gifsmith; no task in this plan covers that, so it is left for a later 04 plan.

## Next Phase Readiness

Plans 04-02..04-06 can build on settings.Store, `SetPref` + `settings:changed`, the Settings window, PANES, SettingRow, shared buttons and SettingsError kinds.

---
*Phase: 04-settings-shell-polish-release*
*Completed: 2026-10-18*
