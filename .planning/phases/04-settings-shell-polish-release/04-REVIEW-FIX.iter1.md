---
phase: 04-settings-shell-polish-release
fixed_at: 2026-10-20T16:12:27Z
review_path: .planning/phases/04-settings-shell-polish-release/04-REVIEW.md
iteration: 1
findings_in_scope: 17
fixed: 17
skipped: 0
status: all_fixed
---

# Phase 4: Code Review Fix Report

**Fixed at:** 2026-10-20T16:12:27Z
**Source review:** .planning/phases/04-settings-shell-polish-release/04-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 17 (WR-B01..WR-B07, WR-F01..WR-F10; Info items out of scope)
- Fixed: 17
- Skipped: 0

`workflow.use_worktrees` is `false`, so all edits and commits were made in the main checkout on
`gsd/phase-04-settings-shell-polish-release`. Every gate below also ran in the main checkout, so
the results can be reproduced from this tree.

## Fixed Issues

### WR-B01: A failed settings write leaves the registered hotkey different from the stored one

**Files modified:** `internal/services/hotkeys.go`, `internal/services/settings.go`, `internal/services/hotkeys_test.go`
**Commit:** 6c0df9e
**Applied fix:** `Bind`'s registration logic moved into `rebind`. A new `Hotkeys.Revert(slot, old)` uses it. When the Store write fails, `SetHotkey` reverts to the binding captured before `Bind`, so what is registered matches what is stored again. Test: `TestSetHotkeyWriteFailureRollsBack`.
**Status:** fixed: requires human verification (state logic)

### WR-B02: At startup, a hotkey that failed to register is still recorded as bound

**Files modified:** `internal/services/hotkeys.go`, `internal/services/settings.go`, `internal/services/hotkeys_test.go`, `internal/shell/shell.go`, `frontend/bindings/gifsmith/internal/services/models.ts`, `frontend/src/settings/HotkeysPane.svelte`
**Commit:** 331de57
**Applied fix:**
- `Start` binds only combos that registered. Refused combos go into a new `failed` set.
- `Snapshot.hotkeysFailed` exposes that set (bindings regenerated), and the Hotkeys pane shows H4b under the row.
- A successful `Bind` clears the slot's failure. Re-recording the stored combo now updates the tray and emits `settings:changed`, even though the Store write is a no-op.
- `Shell.SetOpenAccelerator("")` clears the key equivalent with `RemoveAccelerator` plus `Menu.Update`. Wails' `SetAccelerator("")` would report an error instead.

Tests: `TestStartFailureIsNotBound`, `TestSetHotkeyStoredComboThatFailedAtLaunch`.
**Status:** fixed: requires human verification (the tray menu rebuild is desktop-only and was not run natively)

### WR-B07: A failed hotkey restore clears Open, reports the opposite, and leaves the tray stale

**Files modified:** `internal/services/hotkeys.go`, `internal/services/settings.go`, `internal/services/errors.go`, `internal/services/hotkeys_test.go`, `frontend/src/lib/errors.ts`, `frontend/src/lib/errors.test.ts`
**Commit:** 5b1446a
**Applied fix:**
- New error kind `hotkey_register_failed_lost`. The pane text is "macOS couldn't register {combo}, and your previous shortcut stopped working. Choose another one."
- The old combo is recorded as failed.
- `SetHotkey` calls `OnOpenHotkey("")` for Open and emits `settings:changed`, so the tray and Settings show the slot as unbound.

Test: `TestSetHotkeyRestoreFailsUnbindsTray`, plus the updated `TestBindRestoreFailsLogs`.
**Status:** fixed: requires human verification (state logic)

### WR-B03: A corrupt or unreadable settings.json is silently reset, then overwritten on the next write

**Files modified:** `internal/settings/settings.go`, `internal/settings/settings_test.go`
**Commit:** cdb14e6
**Applied fix:** This now follows the library store's rule:
- A file that exists but cannot be read or parsed leaves the store running on defaults with an `unreadable` flag set.
- `Update` refuses to write and returns `ErrUnreadable`, so the file is never overwritten.
- Each `Update` re-reads the file first, so a file that was fixed or removed works again.

Test: `TestStoreNeverOverwritesUnreadable`, which uses a wrong-typed `pageSize`, checks the file bytes are untouched, then recovers after the file is fixed.
**Status:** fixed: requires human verification (data-safety logic)

### WR-B04: The Developer ID identity is unlocked for the whole build, which runs third-party code

**Files modified:** `build/darwin/Taskfile.yml`, `.github/workflows/release.yml`, `docs/releasing.md`
**Commit:** 7b70ea0
**Applied fix:**
- The Taskfile now has `release:app`, which runs `package` for arm64 and covers dependency install, bindings, vite and the go build.
- It also has `release:sign`, which signs, notarizes, staples and verifies an existing `bin/gifsmith.app`. It checks that the app exists and that the notary profile works.
- Locally, `release` and `release:build` still run the build, then the signing steps.
- CI now runs "Build (no credentials)" before the keychain step, then "Sign, notarize, verify". That step deletes the keychain before `go test` compiles anything. `actionlint` v1.7.12 is clean, and `task --dry` confirms the build-then-sign order.
**Status:** fixed: requires human verification (not run on GitHub Actions; confirm with the next dry run)

### WR-B05: `workflow_dispatch` skips the main-ancestry gate but still signs and notarizes

**Files modified:** `.github/workflows/release.yml`, `docs/releasing.md`
**Commit:** 9fd09a5
**Applied fix:** `git merge-base --is-ancestor HEAD origin/main` now runs for both events, before anything touches credentials. The Dry run section of `docs/releasing.md` says so.
**Status:** fixed

### WR-B06: Timing-based tests are likely to flake on loaded CI runners with `-race`

**Files modified:** `internal/services/updates.go`, `internal/services/updates_test.go`, `internal/services/settings_test.go`
**Commit:** 5b5b25f
**Applied fix:**
- `RunAutoUpdates` wraps `runAutoUpdates`, which takes its timer as `after func(time.Duration) <-chan time.Time`.
- The tests drive that timer through a channel handshake: each tick returns only once the loop is waiting again. They now assert exact check counts and the delay/every sequence, with no sleeps.
- The three "sleep 50 ms, then assert no Clips probe" checks now assert state that `Trigger` sets synchronously.
- I left `slowProvider`'s 50 ms delay alone: it only makes overlap observable and cannot cause a false failure.

All tests ran 5 times under `-race` and passed.
**Status:** fixed

### WR-F01: Settings panes render a launch-time snapshot, so D8 and the Startup checkbox go stale

**Files modified:** `frontend/src/settings/SettingsApp.svelte`, `GeneralPane.svelte`, `DownloadsPane.svelte`, `HotkeysPane.svelte`, `SearchPane.svelte`
**Commit:** 032098c
**Applied fix:**
- `SettingsApp` re-reads `snapshot()` on `settings:show`.
- The panes now derive their values from that snapshot with writable `$derived` (Svelte 5.57), instead of copying it once with `untrack`.
- Their own `settings:changed` listeners are removed: `SettingsApp` already applies those events.
**Status:** fixed: requires human verification (the server/UAT build never emits `settings:show`, so the re-read was not exercised headless)

### WR-F02: The Accessibility poll runs every 2 s for the app's lifetime in a hidden window

**Files modified:** `frontend/src/settings/GeneralPane.svelte`, `internal/shell/shell.go`
**Commit:** b7a86dc
**Applied fix:**
- The poll starts only when the page is visible (`!document.hidden` at mount) or on `settings:show`.
- It stops on a new `settings:hide` event, which `HideSettings` now emits.
- An `alive` guard stops a read that resolves after unmount from creating an orphan interval.
- `readAX` catches rejections.
**Status:** fixed: requires human verification (whether WKWebView reports `document.hidden` for the hidden-at-launch window needs a native check)

### WR-F03: Fire-and-forget bridge calls with no rejection handling

**Files modified:** `frontend/src/App.svelte`, `settings/KeyRow.svelte`, `settings/AboutPane.svelte`, `settings/GeneralPane.svelte`, `settings/HotkeyRecorder.svelte`, `settings/SearchPane.svelte`, `lib/errors.ts`, `lib/errors.test.ts`
**Commit:** eaafe29
**Applied fix:**
- A failed `choices()` shows an inline error on the Search pane.
- A failed `suspendHotkeys()` stops the recording with "Couldn't start recording. Try again." (new `record_failed` text).
- `resumeHotkeys`, `updateStatus`, `openExternal`, `revealUpdates` and `openSettings` are caught.
- `keyStatus()` is handled under WR-F04.
**Status:** fixed

### WR-F04: The Keys pane shows "Not set" and "Add Key…" for a saved key while loading

**Files modified:** `frontend/src/settings/KeysPane.svelte`, `frontend/src/settings/KeyRow.svelte`
**Commit:** fb5f049
**Applied fix:**
- `statuses` is `null` until the first read. A card with an unknown status shows no state and no key buttons; the title, description and "Get a free key" still render, so the UAT's card order check holds.
- A failed `keyStatus()` shows an alert above the cards.
**Status:** fixed

### WR-F05: Cancelling a recording leaves a stale H1 error under the row

**Files modified:** `frontend/src/settings/HotkeyRecorder.svelte`, `scripts/uat/17-hotkeys.js`
**Commit:** 150517d
**Applied fix:**
- Esc, Tab and blur now call `cancel()`, which clears the message and then stops.
- The H2-H4 paths still call `stop()` and keep their error.
- The UAT covers Esc after an H1. I confirmed it fails without the fix.
**Status:** fixed

### WR-F06: Shift+Tab while recording is swallowed and reported as H1

**Files modified:** `frontend/src/settings/HotkeyRecorder.svelte`, `scripts/uat/17-hotkeys.js`
**Commit:** a8f175e
**Applied fix:** Tab cancels unless ⌘, ⌃ or ⌥ is held, so Shift+Tab now cancels and focus moves back. The UAT asserts the cancel, that there is no H1, and that focus left the field. I confirmed it fails without the fix.
**Status:** fixed

### WR-F07: The recorder's accessible name hides the current shortcut and the recording state

**Files modified:** `frontend/src/lib/hotkeys.ts`, `frontend/src/lib/hotkeys.test.ts`, `frontend/src/settings/HotkeyRecorder.svelte`, `scripts/uat/17-hotkeys.js`
**Commit:** 21868f6
**Applied fix:**
- The accessible name stays "{label} shortcut", so the UAT selectors still match.
- A visually hidden span, referenced first in `aria-describedby`, gives `spoken(value)` (for example "Control Option Command G"), or "Not set", or "Recording. Type a shortcut."
- `aria-pressed` follows `recording`.
- `spoken()` has a node test, and the UAT checks the description.
**Status:** fixed

### WR-F08: The update progress bar reads 100% for the whole download when the size is unknown

**Files modified:** `frontend/src/settings/GeneralPane.svelte`
**Commit:** d4e842e
**Applied fix:** When `total > 0`, the bar uses `max`/`value`. Otherwise it renders a `<progress>` with no value, which is indeterminate.
**Status:** fixed

### WR-F09: An update error stays next to a later, contradicting status

**Files modified:** `frontend/src/settings/GeneralPane.svelte`
**Commit:** 5b6fc99
**Applied fix:** `updError` is cleared when `update:status` reports `checking` or `downloading`. It is not cleared on `available`, because Go's failure path emits that before the call rejects.
**Status:** fixed

### WR-F10: Last-response-wins snapshot writes race each other and `settings:changed`

**Files modified:** `internal/services/settings.go`, `internal/services/settings_test.go`, `frontend/src/settings/GeneralPane.svelte`, `SearchPane.svelte`, `DownloadsPane.svelte`
**Commit:** 15c703a
**Applied fix:**
- **Panes:** they no longer apply their own call's result. They keep the change optimistically, show only the snapshots from `SettingsApp` (`settings:changed` and the re-read on show), and revert to the last snapshot on failure.
- **Go:** the review's fix assumed `settings:changed` is ordered, but Go did not guarantee that. A new `SettingsService.changed()` now takes the snapshot from the current store and emits it under one `emitMu`. All five emit sites use it, so events leave in write order and the last one matches the store.

Invariant test: `TestSettingsChangedEndsWithTheStore`.
**Status:** fixed: requires human verification (concurrency logic)

## Verification (all run in the main checkout)

- `gofmt -l main.go wire_desktop.go wire_server.go internal` (the CI scope): clean. A repo-wide `gofmt -l .` only lists the untouched `build/ios` template files, which already failed before this run.
- `go vet . ./internal/...`: OK.
- `go test -race ./internal/...`: every package passes.
- `golangci-lint run . ./internal/...` (the CI scope): 0 issues. A bare `golangci-lint run` reports one issue, `unused`, in the untouched `build/ios/app_options_default.go`.
- `CGO_ENABLED=0 go build -tags server .`: OK.
- `node --test src/lib/*.test.ts`: 102/102.
- `svelte-check --tsconfig ./tsconfig.json`: 0 errors, 0 warnings.
- `scripts/smoke-browser-loop.sh`: SMOKE OK.
- Full UAT set (`--uat` for all 33 `scripts/uat/*.js`): 33/33 pass.
- `actionlint` v1.7.12 on `.github/workflows/*.yml`: clean. It also runs shellcheck on the `run:` blocks.
- Standalone `shellcheck scripts/*.sh` reports only the existing SC2015 info findings in `smoke-browser-loop.sh` and SC2010 in `verify-release.sh`. I did not change those files.
- Bindings regenerated with the Taskfile flags: no diff. `go.sum` was restored after each `wails3` run.

Not verified natively, as no desktop input was allowed:
- the tray key-equivalent clear (`Menu.Update`);
- `settings:show` and `settings:hide` and `document.hidden` in WKWebView;
- the split release workflow on GitHub Actions.

---

_Fixed: 2026-10-20T16:12:27Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
