---
phase: 04-settings-shell-polish-release
plan: 03
subsystem: settings
tags: [wails-v3, svelte5, global-hotkeys, carbon, recorder, reduce-motion, a11y, polish]

requires:
  - phase: 04-settings-shell-polish-release
    provides: "04-01 settings.Store, SettingsService Snapshot/settings:changed, Settings window, PANES; 04-02 Keys pane, SettingsError kinds"
  - phase: 03-action-pipeline-library
    provides: "ActionService T5a path, library FavoritesMax, Undo toasts, Footer/ActionMenu/DetailsPane/Tile"
provides:
  - "services.Registrar seam and services.Hotkeys (Canonical, Start, Bind with rollback, Suspend/Resume, ArmTest, Bindings); DefaultOpenHotkey"
  - "settings.Hotkeys + Prefs.hotkeys, settings.CanonicalHotkey, Normalize for stored hotkeys"
  - "SettingsService SetHotkey/SuspendHotkeys/ResumeHotkeys/TestHotkey, OnOpenHotkey; Snapshot.hotkeys; SettingsError hotkey_* kinds"
  - "shell: hotkey registration removed; ShowTab (panel:open-tab), SetOpenAccelerator, OnSettingsInactive (hide, ⌘W, lost focus); tray key equivalents"
  - "server-build memRegistrar with FAKE_HOTKEY_FAIL; smoke pass-through"
  - "Hotkeys pane (Open Gifsmith, Open Favorites, Open Recents) with HotkeyRecorder, Test, Clear, Reset to Default"
  - "lib/hotkeys.ts (accelFromEvent, glyphs, canonical, DENYLIST, hasModifier, SLOT_LABELS); errors.ts hotkeyErrorText, F11, F12, Clips F3"
  - "lib/motion.ts animates; Reduce Motion canvas stills in Tile"
  - "Phase 3 polish P12-P14, P16-P18, P20-P25; polish todo closed"
  - "scripts/uat/17-hotkeys.js, 17-hotkeys-fail.js, 17-polish.js"
affects: [04-04, 04-05, 04-06, 04-07]

actuals:
  tokens: 19800    # chars/4 over the added lines of the realized diff (79169 chars, 1963 lines added, 42 files)
  tasks: 6
  commits: 13
plan_head_before: 17bf51646b39cf905626101b95b6201bf3db7089
plan_head_after: ad478014e0078e71a7161ba8564a8dd2318368ac

tech-stack:
  added: []
  patterns:
    - "Hotkeys are stored and compared in Wails' canonical form (modifiers sorted Cmd, Ctrl, Option, Shift, then the upper-cased key); the Open default is stored as \"\""
    - "Registrar seam: app.GlobalShortcut in the app, an in-memory fake in the -tags server build (registered before Run, since server mode never fires ApplicationStarted)"
    - "Bound hotkey calls from the webview run strictly in order (api.ts inOrder), so a late Suspend can never land after the Resume that ends a recording"
    - "Hotkeys resume on every way out: recorder end paths, SetHotkey, Settings hide/⌘W/lost focus"
    - "Reduce Motion: a one-time canvas first frame from the unaltered provider URL; only the selected tile keeps an <img>"

key-files:
  created:
    - internal/services/hotkeys.go
    - internal/services/hotkeys_test.go
    - frontend/src/settings/HotkeysPane.svelte
    - frontend/src/settings/HotkeyRecorder.svelte
    - frontend/src/lib/hotkeys.ts
    - frontend/src/lib/hotkeys.test.ts
    - frontend/src/lib/motion.ts
    - frontend/src/lib/motion.test.ts
    - frontend/bindings/gifsmith/internal/settings/models.ts
    - frontend/bindings/gifsmith/internal/settings/index.ts
    - scripts/uat/17-hotkeys.js
    - scripts/uat/17-hotkeys-fail.js
    - scripts/uat/17-polish.js
    - .planning/todos/done/phase-3-ui-review-polish.md (moved from pending)
  modified:
    - internal/services/settings.go
    - internal/services/settings_test.go
    - internal/services/errors.go
    - internal/services/actions.go
    - internal/services/actions_test.go
    - internal/settings/settings.go
    - internal/settings/settings_test.go
    - internal/shell/shell.go
    - main.go
    - wire_server.go
    - scripts/smoke-browser-loop.sh
    - frontend/bindings/gifsmith/internal/services/models.ts
    - frontend/bindings/gifsmith/internal/services/settingsservice.ts
    - frontend/src/App.svelte
    - frontend/src/app.css
    - frontend/src/lib/api.ts
    - frontend/src/lib/errors.ts
    - frontend/src/lib/errors.test.ts
    - frontend/src/settings/SettingsApp.svelte
    - frontend/src/components/ActionMenu.svelte
    - frontend/src/components/Footer.svelte
    - frontend/src/components/DetailsPane.svelte
    - frontend/src/components/Tile.svelte
    - frontend/src/components/ResultGrid.svelte
    - frontend/src/components/LibraryGrid.svelte
    - scripts/uat/11-square.js
    - scripts/uat/15-settings-general.js
    - scripts/uat/16-settings-keys.js

key-decisions:
  - "The hotkey parser lives in settings (settings.CanonicalHotkey), not services, so Normalize can validate stored hotkeys without an import cycle or an injected validator; services.Hotkeys.Canonical maps its errors to SettingsError kinds"
  - "Canonical keys are upper-cased exactly as Wails does (Ctrl+SPACE, Cmd+UP, Cmd+TAB); glyphs() renders them as Space and arrows"
  - "Open \"\" means ⌃⌥⌘G in settings.json; SetHotkey and Normalize store the default as \"\""
  - "A failed restore after a failed rebind clears the slot in memory (logged hotkey restore failed); the Store keeps the old binding, so the next launch registers it again"
  - "Resume failures are logged, not shown: ResumeHotkeys returns nothing to the webview"
  - "The recorder refuses the macOS denylist client-side (H3, no round trip); Go re-validates every accelerator and stays authoritative"
  - "The server build starts Hotkeys before app.Run because server mode never fires ApplicationStarted; the desktop app still registers at ApplicationStarted"
  - "SetOpenAccelerator updates the NSMenuItem through application.InvokeSync (AppKit main thread)"
  - "Reduce Motion uses the canvas path for every provider (results carry no still rendition, and building GIPHY _s.gif URLs would alter vendor URLs)"

patterns-established:
  - "Settings UAT role-name locators that can collide (Keys/Hotkeys, Open Gifsmith shortcut/Test Open Gifsmith shortcut) use exact: true"

requirements-completed: [SETT-03, SHEL-06]

coverage:
  - id: D1
    description: "Open Gifsmith hotkey rebinds from the recorder: Go canonicalises, refuses no-modifier/system/duplicate combos before any registration, unregisters the old binding, registers the new one, re-registers the old one on failure, stores the canonical form and registers stored bindings at startup"
    requirement: SETT-03
    verification:
      - kind: unit
        ref: "internal/services/hotkeys_test.go#TestCanonical, TestBindRefusals, TestBindDuplicate, TestBindRebind, TestBindRollback, TestBindRestoreFailsLogs, TestBindFailedNew, TestBindSameIsNoop, TestStartRegistersDefault, TestStartRegistersStored, TestSetHotkeyStores, TestSetHotkeyErrorKeepsStore"
        status: pass
      - kind: unit
        ref: "internal/settings/settings_test.go#TestStoreNormalizeHotkeys; frontend/src/lib/hotkeys.test.ts"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/17-hotkeys.js; FAKE_HOTKEY_FAIL=Cmd+Ctrl+Option+J scripts/uat/17-hotkeys-fail.js"
        status: pass
    human_judgment: true
    rationale: "Carbon registration, a rebound hotkey firing from other apps and the recorder capturing the current combo exist only natively (A10); automated runs use a fake Registrar"
  - id: D2
    description: "Recorder copy H1-H4b and H7; H1 keeps recording; Backspace clears Favorites/Recents and is ignored for Open"
    requirement: SETT-03
    verification:
      - kind: unit
        ref: "frontend/src/lib/errors.test.ts#hotkeyErrorText"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/17-hotkeys.js (H1, H2, H3, Backspace on Open); scripts/uat/17-hotkeys-fail.js (H4)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Open Favorites / Open Recents hotkeys show the panel on that tab and hide it when pressed again on that tab; Test listens 5 s (H5, H6, H7, Esc cancels); suspend while recording and resume on every exit path including Settings hide/lost focus"
    requirement: SHEL-06
    verification:
      - kind: unit
        ref: "internal/services/hotkeys_test.go#TestArmTestReportsWithinFiveSeconds, TestArmTestUnbound, TestSuspendResumeIdempotent, TestSuspendedComboIsCapturedNotFired, TestSetHotkeyResumesFirst, TestBindClear, TestTestHotkey"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/17-hotkeys.js (unset rows, Test timeout H7, Esc cancels listening, Clear); scripts/uat/12-library.js"
        status: pass
    human_judgment: true
    rationale: "panel:open-tab is driven by a real Carbon hotkey; the hide-on-same-tab path and Test success (H6) only happen natively"
  - id: D4
    description: "Tray menu shows the current Open hotkey on Open Gifsmith (updated on rebind), ⌘, on Settings… and ⌘Q on Quit Gifsmith"
    requirement: SETT-03
    verification:
      - kind: unit
        ref: "internal/services/hotkeys_test.go#TestSetHotkeyOpenUpdatesTray"
        status: pass
    human_judgment: true
    rationale: "NSMenuItem key equivalents only render in the native tray menu"
  - id: D5
    description: "Panel polish: short T5a with the Enable button title, polite live footer, destructive highlight contrast, --text-tertiary #737378/#98989D, Details Esc Back button, toast icon margin, badges on --label"
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestPasteWithoutAccessibility, TestPasteTextWithoutAccessibility"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/17-polish.js, 01-tracer.js, 08-action-menu.js, 11-square.js, 13-details.js"
        status: pass
    human_judgment: false
  - id: D6
    description: "F11/F12/Clips F3 copy, T7c at the 1,000-favorite cap, a no-op removal clears a visible Undo toast, library hints only with a selection"
    verification:
      - kind: unit
        ref: "frontend/src/lib/errors.test.ts#text copy/paste failures show F11 and invalid_item shows F12 (P25), Clips copy errors name the MP4"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/17-polish.js (empty Favorites shows no hints), 07-text-actions.js, 12-library.js, 14-backstops.js"
        status: pass
    human_judgment: true
    rationale: "T7c and the IN-06 Undo clearing are not exercised by any UAT (1,000 favorites and a no-op removal are not reachable in the fixture flows)"
  - id: D7
    description: "Reduce Motion: only the selected tile animates, others show a canvas first frame, followed live, hover never animates"
    verification:
      - kind: unit
        ref: "frontend/src/lib/motion.test.ts"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/17-polish.js (emulateMedia reducedMotion reduce / no-preference, ArrowRight moves the animation)"
        status: pass
    human_judgment: true
    rationale: "WebKit tracking the macOS Reduce Motion setting and cross-origin canvas drawing in WKWebView are native (A11)"

duration: "~33 min active (19 min on 2026-10-18, 14 min on 2026-10-19 after an API-limit pause)"
completed: 2026-10-19
status: complete
---

# Phase 4 Plan 03: Rebindable Hotkeys, Phase 3 Polish and Reduce Motion Summary

**The global hotkeys can now be rebound from Settings → Hotkeys, and Favorites and Recents get optional hotkeys of their own. Go canonicalises every combo, refuses no-modifier, macOS and duplicate combos, and re-registers the old binding when macOS refuses the new one. Recording suspends Gifsmith's hotkeys, and Test listens 5 s for a real press. The Phase 3 UI-review polish is folded in, and Reduce Motion users see still tiles, with only the selected one animating.**

## Performance

- **Duration:** about 33 min of work. It ran 2026-10-18 21:47Z–22:06Z, paused for an API limit, and finished 2026-10-20 01:33Z–01:47Z.
- **Tasks:** 6 (tracer + 5)
- **Files changed:** 42, including 4 regenerated binding files

## Accomplishments

- **Hotkeys manager in Go.** `services.Hotkeys` sits on a `Registrar` interface: `app.GlobalShortcut` in the app, an in-memory fake in the server build. It has three slots (open, favorites, recents).
  - `Bind` validates first, then unregisters the old combo and registers the new one. If the new one fails it re-registers the old one and returns H4, or H4b when nothing was bound before.
  - `Suspend`/`Resume` are idempotent.
  - `ArmTest` arms one report within 5 s. The slot callback emits `hotkey:test` and then does its normal action.
- **Stored and normalised.** `Prefs.hotkeys` holds canonical accelerators, with the Open default stored as `""`. `Normalize` canonicalises the stored hotkeys and resets invalid, macOS and duplicate ones, with a log line. `SetHotkey` resumes the hotkeys first, and the Store changes only after the Registrar accepted the new combo.
- **Shell.** It no longer registers a hotkey. It gains:
  - `ShowTab` (emits `panel:open-tab` {tab, wasVisible}).
  - `SetOpenAccelerator`.
  - `OnSettingsInactive`, called on Settings hide, ⌘W and WindowLostFocus, which resumes the hotkeys.
  - The tray now shows the Open hotkey, ⌘, and ⌘Q.
- **Hotkeys pane.** Three stacked rows, each with a 160×32 click-to-record field. Recording shows live modifier glyphs and an accent ring. The rows have Test, Reset to Default (Open) and Clear (Favorites/Recents), the H1–H8 copy and an aria-labelled button per row.
- **Polish.** P12–P14, P16–P18 and P20–P25 are done. The short T5a message and the button `title` are covered above. Also:
  - The footer centre is a live region.
  - A highlighted destructive row uses `--text` for its label.
  - Esc Back is a button.
  - The new tertiary text colours are in.
  - Text failures show F11; `invalid_item` shows F12.
  - The Clips "Try another clip." copy is in.
  - Adding at the 1,000-favorite cap shows T7c.
  - A no-op removal clears a stale Undo toast.
  - Library footer hints show only when an item is selected.
- **Reduce Motion.** `animates(index, selected, reduced)` decides which tiles animate. The other tiles draw one first frame to a `<canvas>` from the unaltered provider URL and then drop the image. The setting is followed live via `matchMedia`.

## Task Commits

1. **Task 1 (tracer): rebind Open end to end.** `08570c3` (test, RED), `f48ac8b` (feat). The tracer gate re-ran the full `<verify>`: gofmt, vet, race tests, server build, node tests, svelte-check, smoke and 17-hotkeys. It passed before expansion.
2. **Task 2: recorder rules, H1–H4b, Normalize, FAKE_HOTKEY_FAIL.** `a5f769f` (test, RED), `b71edd4` (feat)
3. **Task 3: Favorites/Recents, Test, tray key equivalents.** `d783fe8` (test), `fa4ecca` (feat)
4. **Task 4: panel polish P12–P14, P16, P17, P21, P22.** `c655a83` (test, RED), `343bc78` (feat)
5. **Task 5: F11/F12/Clips F3, T7c, IN-06, P18.** `a908eba` (test, RED), `da88143` (feat)
6. **Task 6: Reduce Motion, todo closed.** `94e865b` (test, RED), `0dd585f` (feat)
7. **Regression fix:** `ad47801` makes the Keys tab locator exact in 16-settings-keys (the new Hotkeys tab also matched "Keys").

## TDD Gate Compliance

- Every TDD task has a `test(04-03)` commit before its `feat(04-03)` commit (6 pairs).
- **Frontend RED evidence (TAP).** `gsd-tools check tdd-red-evidence` returned `RED_EVIDENCE_OK` for four tests:
  - "glyphs renders in macOS order ⌃⌥⇧⌘ then the key" (hotkeys.test.ts)
  - "hotkeyErrorText: H1, H2, H3, H4, H4b and H7…" (errors.test.ts)
  - "text copy/paste failures show F11 and invalid_item shows F12 (P25)" (errors.test.ts)
  - "animates: only the selected tile under Reduce Motion" (motion.test.ts)

  Semantic check: each test failed on its planned assertion against a stub (`''`, `false`) or the old copy, with no load errors.
- **Go RED evidence.** The classifier does not accept `go test` output, so this is a semantic assessment only. These targets compiled against skeletons and failed on their planned assertions:
  - The Task 1 Hotkeys tests: no calls, nil errors, empty Bindings.
  - `TestStoreNormalizeHotkeys`: nothing was reset.
  - The T5a tests: the old long message.
- **Task 3's Go tests passed early, as expected.** Plan step 1.1 had Task 1 build the full manager (Suspend/Resume/ArmTest/clear). The Task 3 tests (`d783fe8`) cover that behaviour and passed on first run, after a fix to the test's own event formatting. There was no Go GREEN step for Task 3. Its `feat` commit is the pane, shell and App wiring.

## Verification (plan level)

- `gofmt -l` is clean. `go vet . ./internal/...` passes. `go test -race ./internal/... -count=1` passes in every package with no data race. `golangci-lint run . ./internal/...` reports 0 issues.
- `CGO_ENABLED=0 go build -tags server .` succeeds. `node --test` reports 87 pass, 0 fail. svelte-check reports 0 errors and 0 warnings.
- `scripts/smoke-browser-loop.sh` ends with SMOKE OK. All 25 `scripts/uat/*.js` print UAT OK, including 17-hotkeys, 17-hotkeys-fail (with FAKE_HOTKEY_FAIL from its Run via line) and 17-polish.
- go.sum is unchanged. Bindings were regenerated with `wails3 generate bindings -clean=true -ts -i`.

## Decisions Made

See key-decisions in the frontmatter. The main ones:

- The parser lives in `settings`.
- Wails' upper-cased canonical keys are kept.
- The default Open hotkey is stored as `""`.
- The server build starts Hotkeys before `Run`.
- Reduce Motion uses the canvas path for every provider.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Server mode never fires ApplicationStarted**
- **Found during:** Task 1. The UAT showed Open as "Not set".
- **Issue:** Wails server mode emits no ApplicationStarted event, so `hk.Start` never ran in the UAT build. Calling `SetOpenAccelerator` before `Run` then crashed, because `InvokeSync` has no app impl yet.
- **Fix:** When `deps.Registrar` is the fake, main calls `hk.Start` before `app.Run`. It skips the tray update, since the server build has no tray. The desktop app still starts at ApplicationStarted.
- **Files:** main.go. **Commit:** f48ac8b

**2. [Rule 2 - Correctness, T-04-12] Hotkey calls from the webview could reorder**
- **Found during:** Task 1
- **Issue:** Each bound call runs on its own goroutine. A Suspend sent at recording start could land after the Resume that ends the recording, which would leave the hotkeys dead until Settings hides.
- **Fix:** `api.ts` chains `setHotkey`, `suspendHotkeys` and `resumeHotkeys` through one promise queue. `SettingsService.hkMu` also serialises `SetHotkey`, so the Store always ends in the registered state.
- **Files:** frontend/src/lib/api.ts, internal/services/settings.go. **Commit:** f48ac8b

**3. [Rule 3 - Blocking] The 15-settings-general pane walk assumed at most 3 panes**
- **Found during:** Task 1. Adding Hotkeys made 4 panes.
- **Fix:** The ↓ sequence is now generated as `min(i+1, last-1)`.
- **Files:** scripts/uat/15-settings-general.js. **Commit:** f48ac8b

**4. [Rule 3 - Blocking] Role-name locators matched the new controls**
- **Found during:** Task 3 and plan verification
- **Issue:** Playwright role names match substrings. "Open Gifsmith shortcut" also matched "Test Open Gifsmith shortcut", and "Keys" also matched "Hotkeys".
- **Fix:** The locators use `exact: true`.
- **Files:** scripts/uat/17-hotkeys.js, 17-hotkeys-fail.js, 16-settings-keys.js. **Commits:** fa4ecca, ad47801

### Plan adjustments (no behaviour lost)

- **Tray key equivalents** (orchestrator note): done in Task 3. Open Gifsmith shows the current Open hotkey and updates on rebind through `OnOpenHotkey`; Settings… shows `CmdOrCtrl+,` and Quit shows `CmdOrCtrl+Q`. Wails parsed all three, and the server log shows no "invalid accelerator".
- **Where the parser lives:** `settings.CanonicalHotkey` plus the `ErrHotkey*` sentinels, rather than an injected validator func. `DefaultOpenHotkey` is defined in settings and re-exported by services.
- **The Esc Back button is in `Footer.svelte`,** where the Details hint lives, not in `DetailsPane.svelte`. The acceptance grep for `Close Details (Esc)` therefore matches Footer.svelte. The P22 badge change is in DetailsPane.svelte and Tile.svelte.
- **The Test button arrived in Task 3** with its behaviour, not in Task 1, so Task 1 shipped no dead button.
- **`Snapshot.hotkeys` is typed `settings.Hotkeys`.** That generates a new bindings folder, `frontend/bindings/gifsmith/internal/settings/`.
- **`Prefs.hotkeys` uses `omitzero`,** so settings.json stays unchanged until a hotkey is set.
- **The P21 check icon** sits flush against the text: the `{#if}` block has no whitespace text node.
- **17-polish ordering.** The Details check runs before ⌘↵, because the footer shows Enable Automatic Paste… instead of the hints while `needsAccessibility` is set. The P21 gap is measured on the ↵ success toast, since T5a has no check icon.

**Total deviations:** 4 auto-fixed (1 Rule 2, 3 Rule 3) and 8 minor adjustments. **Impact:** none on scope.

## Issues Encountered

- The run paused for an API limit after the `d783fe8` commit and resumed from the working tree. No work was lost.

## Known Stubs

None.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: bound-surface | internal/services/settings.go | `SuspendHotkeys` and `ResumeHotkeys` are bound, so any page in either webview can suspend every Gifsmith hotkey. This is mitigated per T-04-12: Go resumes on Settings hide, ⌘W, lost focus and before every SetHotkey. The webviews load only the app's own assets. |
| threat_flag: bound-surface | internal/services/settings.go | `TestHotkey` is bound. It only arms a 5 s report for a slot that is already bound, and an unknown or unbound slot gets hotkey_invalid. |

## Human-check items deferred to end of phase (native, user-verified on the signed install)

1. **Rebinding (Task 3 human check).** Open Settings → Hotkeys and record a new Open shortcut. Switch to another app and press it: the panel should toggle.
2. **Favorites and Recents hotkeys.** Set Open Favorites to ⌃⌥⌘F and Open Recents to ⌃⌥⌘R. Press each from another app: the panel should open on that tab. Press ⌃⌥⌘F again while Favorites is showing: the panel should hide.
3. **Recording captures, not fires.** Click the Open recorder and press the current Open shortcut. It should be recorded, and the panel should not toggle.
4. **Test.** On a free combo, Test shows "It works." and opens the panel. On a combo another app uses (for example one of Vorssaint's), it shows "Nothing happened…", or the other app reacts. A10: check whether Carbon hotkeys fire while Settings is key.
5. **Tray menu.** It should show the Open shortcut next to Open Gifsmith, ⌘, next to Settings… and ⌘Q next to Quit Gifsmith, and the Open shortcut should update after a rebind.
6. **Hotkeys come back after leaving a recording.** Start recording, then click into another app. Without returning to Settings, press the Open shortcut: it should work, because WindowLostFocus resumes the hotkeys.
7. **Reduce Motion (Task 6 human check).** Turn on System Settings → Accessibility → Display → Reduce motion and arrow through the grid. Only the selected GIF should move, the others should show a still frame (no blank tiles), and turning it off should animate every tile again without reopening the panel. A11: if cross-origin canvas stills are blank in WKWebView, the fallback is removing `src`.
8. **Polish.** Clicking Esc Back in Details closes it. Hovering Enable Automatic Paste… shows its tooltip. With VoiceOver on, footer toasts are announced (P13).

## Next Phase Readiness

Plans 04-04..04-07 can build on these:

- The Hotkeys pane and `SLOT_LABELS`.
- `Shell.ShowTab`.
- The polish baseline. P15 (right-click menu) and P19 (menu bottom clamp) remain for 04-04.
- The tray now carries key equivalents. 04-07's "Update to Gifsmith…" item should keep the same `menu.Add(...)` pattern.

---
*Phase: 04-settings-shell-polish-release*
*Completed: 2026-10-19*

## Self-Check: PASSED

- All created files exist on disk.
- All 13 plan commits (08570c3..ad47801) are ancestors of HEAD, and `git rev-list --count 17bf516..HEAD` = 13. evaluation-scope resolves the plan's commits.
- The acceptance criteria were re-run: all pass. The one exception is `Close Details (Esc)`, which matches Footer.svelte instead of DetailsPane.svelte (see Plan adjustments). The plan-level verification above was re-run and is all green.
