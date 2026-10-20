---
phase: 04-settings-shell-polish-release
reviewed: 2026-10-20T16:20:48Z
depth: standard
iteration: 2
files_reviewed: 29
files_reviewed_list:
  - .github/workflows/release.yml
  - build/darwin/Taskfile.yml
  - docs/releasing.md
  - frontend/bindings/gifsmith/internal/services/models.ts
  - frontend/src/App.svelte
  - frontend/src/lib/errors.test.ts
  - frontend/src/lib/errors.ts
  - frontend/src/lib/hotkeys.test.ts
  - frontend/src/lib/hotkeys.ts
  - frontend/src/settings/AboutPane.svelte
  - frontend/src/settings/DownloadsPane.svelte
  - frontend/src/settings/GeneralPane.svelte
  - frontend/src/settings/HotkeyRecorder.svelte
  - frontend/src/settings/HotkeysPane.svelte
  - frontend/src/settings/KeyRow.svelte
  - frontend/src/settings/KeysPane.svelte
  - frontend/src/settings/SearchPane.svelte
  - frontend/src/settings/SettingsApp.svelte
  - internal/services/errors.go
  - internal/services/hotkeys.go
  - internal/services/hotkeys_test.go
  - internal/services/settings.go
  - internal/services/settings_test.go
  - internal/services/updates.go
  - internal/services/updates_test.go
  - internal/settings/settings.go
  - internal/settings/settings_test.go
  - internal/shell/shell.go
  - scripts/uat/17-hotkeys.js
findings:
  critical: 0
  warning: 7
  info: 28
  total: 35
status: issues_found
---

# Phase 4: Code Review Report (iteration 2)

**Reviewed:** 2026-10-20T16:20:48Z
**Depth:** standard (re-review of the fix pass)
**Files Reviewed:** 29
**Status:** issues_found

## Summary

This re-review covers the 18 fix commits `6c0df9e..15c703a` (`git diff 00eac67..HEAD`). The iteration-1 report is kept in `04-REVIEW.iter1.md`.

Checks run:
- `go test -race -count=1 ./internal/services/ ./internal/settings/` passes.
- `node --test src/lib/*.test.ts` passes 102 of 102.
- One probe test was compiled in with `go test -overlay`; no file in the repo was changed. It confirms WR-05: Settings reports a hotkey as saved, but `settings.json` does not keep it.

Of the 17 iteration-1 warnings:
- 12 are resolved.
- 4 are partially resolved: WR-B03, WR-B04, WR-B05 and WR-F10.
- WR-B02 is resolved, but its fix introduced two new defects (WR-05, WR-06).

The fix pass introduced no critical issues. It left 7 warnings, listed below. Three of them were introduced by the fixes (WR-04 is an existing defect the fixes made easy to hit). The other four are what remains of the partially resolved items.

What I checked in the areas you asked about:
- **`SettingsService.changed()` and `emitMu`: no deadlock.**
  - Wails' `EventManager.Emit` runs the hooks, starts the Go listeners on their own goroutine, then adds the event to a FIFO mailbox (`internal/mailbox`) and returns at once. Holding `emitMu` across `Emit` never waits on a webview or the main thread.
  - Inside `changed()`, locks are taken in one order: `emitMu`, then `Store.mu` and `Hotkeys.mu`, each released quickly. No path takes them in the reverse order.
  - Only bound methods call `changed()`, and they run on goroutines, never on the main thread. No Go code listens for `settings:changed`.
  - `SetHotkey` keeps holding `hkMu` while `OnOpenHotkey` waits on the main thread (`InvokeSync` / `Menu.Update`). This is safe because the main thread never takes `hkMu`.
  - Events therefore leave in write order, and the last one matches the store.
- **Tray accelerator.** `SetOpenAccelerator("")` sets `accelerator = nil` and then calls `Menu.Update()`. That rebuilds every item on the main thread into the same `nsMenu` the tray has cached. Hidden items (the update item) and labels are kept. This is correct, but not verified natively.
- **release.yml keychain lifecycle.** Secrets reach only the keychain step's `env`.
  - On success, the sign step deletes the keychain.
  - On any failure or cancellation, the `always()` cleanup deletes it. `RELEASE_KEYCHAIN` is only set once the keychain step has started, so a failed build never runs a delete on a path that does not exist.
  - The weakness is isolation inside the runner, not keychain lifetime (WR-01).
- **Settings `ErrUnreadable`.** The user's file is now never overwritten, which was the data-loss part of WR-B03. However, the user can stay stuck with an error that cannot be fixed by retrying and says nothing about the cause (WR-03). It also makes WR-04 deterministic.

## Iteration-1 Warning Status

| ID | Status | Notes |
|----|--------|-------|
| WR-B01 | Resolved | `Revert` restores what is registered when the Store write fails (`settings.go:575-579`). A double failure, where re-registering the old combo is also refused, still leaves registered and stored different (IN-03). |
| WR-B02 | Resolved, with regressions | `Start` binds only combos that registered, and `hotkeysFailed` reaches the pane. Because a refused combo is no longer in `bound`, the duplicate check now lets another slot take it, and the store silently drops that binding (WR-05). An unbound Open also disables Reset to Default (WR-06). |
| WR-B03 | Partially resolved | The file is never overwritten, and the store recovers once the file is fixed. But writes fail forever with the generic "Couldn't save this setting" and nothing names the cause (WR-03). `SetLaunchAtLogin` now reliably changes the login item and then reports failure (WR-04). |
| WR-B04 | Partially resolved | The build runs before the keychain exists, but in the same job on the same runner. Code from the build step can still act while the identity is unlocked (WR-01). |
| WR-B05 | Partially resolved | The gate now runs for both events. It is defined in the workflow file of the ref being run, so a ref that carries its own `release.yml` skips it. The environment policy allows any `v*` tag (WR-02). |
| WR-B06 | Resolved | `runAutoUpdates` takes an injected `after`. The tests use a channel handshake with exact counts and wait sequences. The remaining sleeps are `slowProvider`'s 50 ms (it only makes overlap observable) and `waitClips`' 3 s polling deadline, and neither can cause a false failure. |
| WR-B07 | Resolved | There is a distinct `hotkey_register_failed_lost` kind. The old combo is recorded as failed. For Open, the tray is cleared with `OnOpenHotkey("")`, and `settings:changed` is emitted. |
| WR-F01 | Resolved (needs native check) | `SettingsApp` re-reads `snapshot()` on `settings:show`, and the panes use writable `$derived`. The re-read is a call result, so it can race `settings:changed` (IN-01). |
| WR-F02 | Resolved (needs native check) | The poll is gated on `document.hidden` at mount and on `settings:show` / `settings:hide`. An `alive` guard and `try/catch` were added. It still runs during ⌘H, and WKWebView's `document.hidden` for a window created hidden is unverified (IN-06). |
| WR-F03 | Resolved | Every listed call is caught. Failures of `choices()` and `suspendHotkeys()` now show a message. |
| WR-F04 | Resolved | `statuses` is `null` until the first read. A card shows no state and no key buttons until then. A failed read shows an alert. |
| WR-F05 | Resolved | `cancel()` clears the message on Esc, Tab and blur. The H2-H4 paths still keep theirs. The UAT covers it. |
| WR-F06 | Resolved | Tab with no ⌘, ⌃ or ⌥ (so Tab or Shift+Tab) cancels without `preventDefault`. The UAT asserts the cancel and that focus moved. |
| WR-F07 | Resolved | A hidden span, referenced first in `aria-describedby`, gives the spoken combo or the recording state. `aria-pressed` was added. |
| WR-F08 | Resolved | When `total` is 0 the bar has no value, so it is indeterminate. |
| WR-F09 | Resolved | The error is cleared on a `checking` or `downloading` event. A small ordering race remains (IN-07). |
| WR-F10 | Partially resolved | The Go ordering (`changed()`) is sound. General, Search and Downloads no longer apply their call's result. The Hotkeys pane still does, in two places (WR-07). |

## Warnings

### WR-01: The signing step still runs on a runner the build step's code could tamper with (WR-B04 residual)

**File:** `.github/workflows/release.yml:81-134`, `build/darwin/Taskfile.yml:259-276`, `docs/releasing.md:33-37`
**Issue:** "Build (no credentials)" and "Sign, notarize, verify" are steps in the same job, so they share the runner, the user account, `$HOME` and the workspace. Third-party code from `pnpm install`, Vite plugins and `go mod tidy` runs in the build step and can still reach the identity:
- **Background processes.** On GitHub-hosted runners, a process started in one step keeps running until the job ends; orphans are only cleaned up then. A process the build step starts can wait until `security find-identity` succeeds, then call `/usr/bin/codesign --sign "Developer ID Application: …"` or `notarytool --keychain-profile gifsmith-notary`.
- **Files the sign step runs.** The build step can rewrite files that the sign step runs while the keychain is unlocked: `build/darwin/Taskfile.yml`, `build/Taskfile.yml`, `scripts/notarize.sh`, `scripts/verify-release.sh`, `~/go/bin/wails3`, and the node and corepack shims on `PATH`.

The comments at `release.yml:81-82` and `Taskfile.yml:259-260` say their code "never runs while the Developer ID identity is unlocked". `docs/releasing.md` makes the same claim. None of these hold on a shared runner.
**Fix:** Split the work into two jobs so the signing runner never runs build-time code:
```yaml
jobs:
  build:            # no environment, no secrets, contents: read
    steps:
      - checkout, setup-go/node, pnpm, wails3
      - run: wails3 task darwin:release:app
      - uses: actions/upload-artifact@<sha>
        with: { name: app, path: bin/gifsmith.app }
  release:
    needs: build
    environment: release
    steps:
      - checkout (fresh), gate, install pinned wails3 only (no pnpm install / go mod)
      - uses: actions/download-artifact@<sha>
        with: { name: app, path: bin/gifsmith.app }
      - keychain step, then wails3 task darwin:release:sign, delete keychain
```
`codesign`, `notarytool`, `spctl` and `hdiutil` never run the bundle, so treating the downloaded `.app` as untrusted is enough. Until the split is done, correct the comments and docs.

### WR-02: The main-ancestry gate is defined by the ref it gates, so it does not protect the signing secrets (WR-B05 residual)

**File:** `.github/workflows/release.yml:46-47`, `docs/releasing.md:31-32, 41-42, 50-51`
**Issue:** Both `push` and `workflow_dispatch` run the `release.yml` from the triggering ref. A tag or branch whose own `release.yml` drops the `git merge-base` line passes the gate.

The only real control is the `release` environment's deployment policy, which the docs describe as "main and `v*` tags". Anyone who can push a `v*` tag, on any commit, gets the Developer ID and notary secrets. This is the exact scenario in WR-B05 (`gh workflow run release.yml --ref v9.9.9`): a tag pushed with a modified workflow skips the new gate too.

`docs/releasing.md` now promises that a dry run "refuses before touching any credential when the ref is not on `main`". That is true only when the ref's workflow file has not been edited.
**Fix:** Enforce the policy outside the workflow file:
- Add a tag ruleset so only the maintainer can create `v*` tags.
- Make "Required reviewers" mandatory on the `release` environment (step 6 currently says "if").
- Optionally, limit the environment's deployment refs to `main` plus protected tags.

Then state in `docs/releasing.md` that the in-workflow gate guards against mistakes, not attacks.

### WR-03: An unreadable settings.json blocks every settings change, and the error never says why (WR-B03 residual)

**File:** `internal/settings/settings.go:418-423`, `internal/services/settings.go:445-447, 481-484, 507-509, 571-576, 647-649`, `internal/services/errors.go:71`
**Issue:** `Store.Update` returns `ErrUnreadable` for as long as the file cannot be parsed. Every caller maps that to `errPrefsWriteFailed`, which the UI shows as "Couldn't save this setting. Try again." The consequences:
- Trying again can never work.
- The Snapshot has no flag for this state, so Settings cannot show a banner.
- Nothing names the file or the cause.
- The panel's `SetLastProvider` fails silently on every tab switch.

The trigger is the same one WR-B03 described: a single wrong-typed hand edit such as `"pageSize": "36"`. The user's only way out is to find and edit `~/Library/Application Support/…/settings.json`, and nothing in the app tells them to. Data is safe now, but the app is stuck on defaults with no in-app explanation or recovery. That breaks the project rule that a hard failure must stay informative.
**Fix:**
- Give it its own error kind, and expose the state:
  ```go
  case errors.Is(err, settings.ErrUnreadable):
      return s.snapshotOf(p), errPrefsUnreadable // kind "prefs_unreadable"
  ```
- Add `SettingsUnreadable bool` to `Snapshot`, set from a `Store.Unreadable()` accessor.
- Show a persistent banner in Settings, for example: "Gifsmith couldn't read its settings file, so changes aren't saved. Reset Settings…". The button calls a bound `ResetSettings()` that renames the file to `settings.json.corrupt-<timestamp>`, clears the flag, and emits `settings:changed`.

### WR-04: Open at login changes the login item, then reports that it failed (now on every attempt while settings.json is unreadable)

**File:** `internal/services/settings.go:500-511`, `frontend/src/settings/GeneralPane.svelte:123-135`
**Issue:** `SetLaunchAtLogin` calls `Autostart.Enable()` or `Disable()` first and writes the Store afterwards. If the write fails, it returns `errPrefsWriteFailed`, and the pane does two things:
- reverts the checkbox to `initial.launchAtLogin`;
- shows "Couldn't change Open at login. Try again."

But the login item has already changed. Before this phase the write rarely failed (disk full). With WR-B03's `ErrUnreadable` it fails on every attempt for as long as the file is unreadable. The user is told it failed and sees the old state, while macOS now has the new state.

The next re-read on show reports the OS state (`snapshotOf` reads `IsEnabled`), so the checkbox then flips on its own. At the next launch, `ApplyStartup` reads the default `LaunchAtLogin=false` and leaves the item enabled.
**Fix:** Undo the OS change when the store write fails, or write the store first:
```go
p, err := s.Store.Update(...)
if err != nil {
    undo := s.Autostart.Disable
    if !on { undo = s.Autostart.Enable }
    if uerr := undo(); uerr != nil { log.Printf("autostart: rollback failed") }
    return s.snapshotOf(p), errPrefsWriteFailed
}
```

### WR-05: A combo refused at launch can be bound to another slot, and the store then silently drops that binding (introduced by the WR-B02 fix)

**File:** `internal/services/hotkeys.go:140-144`, `internal/services/settings.go:560-571`, `internal/settings/settings.go:341-366`
**Issue:** `Bind`'s duplicate check compares only against `h.bound`. Since WR-B02, a stored combo that macOS refused at launch is kept in `h.failed`, not in `h.bound`, so another slot can take it. Then:
1. `Bind` registers the combo for the new slot.
2. `Store.Update` runs `Normalize()`, which resets the later slot as a duplicate of the stored one. Open's default counts as taken even when it is stored as `""`.
3. The write succeeds, so `SetHotkey` returns success and emits a `settings:changed` that shows the new binding.

The live registration and `settings.json` now disagree, which is the WR-B01 invariant again. The binding the user just saved disappears at the next launch, with no message.

Verified with an overlay test: default Open refused at launch, then `SetHotkey("favorites", "Cmd+Ctrl+Option+G")` gives
```
err=<nil> snap.Hotkeys={Open: Favorites:Cmd+Ctrl+Option+G}
bound={Open: Favorites:Cmd+Ctrl+Option+G} stored={Open: Favorites: Recents:}
```
**Fix:** Treat a stored but unbound combo as taken. Also refuse inside the Store transaction when `Normalize` would reset a slot, and revert:
```go
// Bind: duplicate check
for _, other := range hotkeySlots {
    if other == slot { continue }
    if *slotOf(&h.bound, other) == c || *slotOf(&h.failed, other) == c {
        return &SettingsError{Kind: "hotkey_duplicate", Slot: other}
    }
}
// SetHotkey's Update func, after setting the slot (the store's other fields are already normalized)
if n := *p; len(n.Normalize()) > 0 {
    return &SettingsError{Kind: "hotkey_duplicate"} // then SetHotkey reverts with s.Hotkeys.Revert
}
```
For Open stored as `""` that failed at launch, `failed.Open` already holds the canonical default, so the first check covers it.

### WR-06: "Reset to Default" is disabled exactly when Open is unbound (introduced by the WR-B02 / WR-B07 fixes)

**File:** `frontend/src/settings/HotkeysPane.svelte:109-114`
**Issue:** `disabled={canonical(hk.open || DEFAULT_OPEN) === DEFAULT_OPEN}` was written when the Snapshot guaranteed "Open always set". Now `hotkeys.open` is `""` whenever Open is not registered, in two cases:
- its combo was refused at launch (H4b);
- a restore failed (`hotkey_register_failed_lost`, whose text says "Choose another one").

In both cases `"" || DEFAULT_OPEN` equals the default, so the button is disabled. The obvious way to recover is blocked, and so is a retry of the default after a transient refusal. The user has no global hotkey and must know to record ⌃⌥⌘G by hand.
**Fix:**
```svelte
disabled={!!hk.open && canonical(hk.open) === DEFAULT_OPEN}
```
An unbound Open should always offer Reset. Go's `Bind` re-registers the default even when it is the stored value, through the `errUnchanged && b != was` path.

### WR-07: The Hotkeys pane still applies its call's snapshot, so a late response can overwrite a newer `settings:changed` (WR-F10 residual)

**File:** `frontend/src/settings/HotkeysPane.svelte:34-36, 49`, `frontend/src/settings/HotkeyRecorder.svelte:59, 105`
**Issue:** `change()` calls `saved(await setHotkey(...))`, and the recorder calls `onsaved(await setHotkey(...))`. Both assign `hk = {...s.hotkeys}` from the RPC result.

WR-F10's fix made `settings:changed` the only source of displayed values in General, Search and Downloads, because call responses are not ordered against events. The Hotkeys pane was missed. Example:
1. Clear Favorites, then quickly clear Recents.
2. Go serializes them and emits E1 (Favorites cleared), then E2 (both cleared).
3. Response 1 can arrive after E2. It sets `hk` back to the E1 state.
4. The Recents row now shows a shortcut that is no longer registered, and Test is enabled for it, until the next snapshot.
**Fix:** Drop `saved` and `onsaved`, and rely on `SettingsApp`'s snapshot. On success, every `SetHotkey` that changes registration emits `settings:changed`, including the `errUnchanged && b != was` path. On error, nothing changed, so there is nothing to revert.
```svelte
// HotkeysPane: await setHotkey(slot, accel);   (no saved(...))
// HotkeyRecorder: await setHotkey(slot, accel); onfail('', '', '');
```

## Info

### Carried forward from iteration 1 (unchanged; see `04-REVIEW.iter1.md`)

These 20 items are counted in `info`:
- IN-B01 to IN-B09
- IN-F01 to IN-F11

The fix pass did not target any of them. IN-F05 (`hasModifier` used only by tests) still applies, and the new `spoken` export is used in production.

### New in iteration 2

### IN-01: The show re-read is a call result, so it can overwrite a newer `settings:changed`

**File:** `frontend/src/settings/SettingsApp.svelte:34-40`
**Issue:** `snap = await snapshot()` is applied whenever it resolves. If a write from the panel webview lands between Go serving the snapshot and the webview applying it, the newer event is overwritten with the older snapshot. It stays that way until the next change or show. The window is narrow, because a Settings open is rarely accompanied by a write.
**Fix:** Count `settings:changed` events. Drop the re-read result if an event arrived after the re-read started.

### IN-02: The "write failure" tests now exercise `ErrUnreadable`, not a failed write

**File:** `internal/services/settings_test.go:513-528`, `internal/services/hotkeys_test.go:366-381`
**Issue:** `OpenStore(file)`, where `file` is a regular file, now fails the read with ENOTDIR. Since WR-B03 that marks the store unreadable. `TestSetPrefWriteFailure` and `TestSetHotkeyWriteFailureRollsBack` therefore pass through the `ErrUnreadable` early return. The `WriteJSONAtomic` failure branch in the services is no longer covered.
**Fix:** Open a valid store, then `chmod 0500` its directory (restore it in `t.Cleanup`) so that only the write fails. Keep one test per path.

### IN-03: A failed rollback leaves registered and stored different, with no event or tray update

**File:** `internal/services/settings.go:575-579`, `internal/services/hotkeys.go:157-162`
**Issue:** If `Revert` cannot re-register the old combo, it keeps or clears the new one, and the result is only logged:
- If it keeps the new combo, that combo stays live while `settings.json` holds the old one.
- If it clears it, `failed` is not updated, and the tray and webviews are not told.

This needs a write failure and a Carbon refusal at the same time, so it is rare.
**Fix:** If `Revert` returns an error, record `failed`, call `OnOpenHotkey(Bindings().Open)` for Open, and return `s.changed()` with the error.

### IN-04: A store that recovers mid-session swaps in the file's prefs without applying them

**File:** `internal/settings/settings.go:421`
**Issue:** When `load()` succeeds inside `Update`, `s.p` becomes the file's contents. The theme, hotkeys and login item that were applied from defaults at launch are not re-applied. Only the field being written takes effect, until relaunch.
**Fix:** Log that a relaunch is needed, or have the service re-apply its live state when the store reports that it recovered.

### IN-05: The H4b message is computed once at mount and lost when a recording starts

**File:** `frontend/src/settings/HotkeysPane.svelte:23-29`, `frontend/src/settings/HotkeyRecorder.svelte:32`
**Issue:**
- `hotkeysFailed` is read only from the mount-time `initial`. Later snapshots never add or clear the message.
- `start()` calls `onfail('')`, so clicking the field and pressing Esc removes H4b, although the slot is still unbound.
**Fix:** Derive the per-slot H4b text from the current snapshot's `hotkeysFailed`. Show it whenever no other message is set.

### IN-06: The Accessibility poll relies on `document.hidden` at mount and misses ⌘H

**File:** `frontend/src/settings/GeneralPane.svelte:62, 94-101`, `internal/shell/shell.go:263`
**Issue:**
- `visible = !document.hidden` is unverified for a WKWebView window created with `Hidden: true`. If it reads `false`, the launch-to-first-show poll that WR-F02 described still happens.
- ⌘H (the App menu's Hide) hides Settings without `HideSettings`, so the poll keeps running while the app is hidden.
**Fix:** Also listen for `visibilitychange`. Alternatively, start with `visible = false` and let `SettingsApp` pass down a `shown` flag that it tracks from `settings:show` and `settings:hide`.

### IN-07: A fast-failing manual check can lose its error to the `checking` event

**File:** `frontend/src/settings/GeneralPane.svelte:86-91, 148-155`
**Issue:** `check()` sets `updError` when the RPC rejects. The `checking` event that Go emitted before the failure travels a separate channel, the event mailbox, and is not ordered against the RPC response. If it is delivered after the rejection, it clears the fresh error. The fix's own comment names this race for `available`.
**Fix:** Clear `updError` on `checking` or `downloading` only when no manual call is in flight. Track that with an `inFlight` flag set by `check()` and `download()`.

### IN-08: `release:sign` signs whatever `bin/gifsmith.app` exists

**File:** `build/darwin/Taskfile.yml:265-276`
**Issue:** Run locally on its own, `release:sign` signs and notarizes any bundle found in `bin/`. That can be one left by `install:dev`, `package:universal`, or a build from an uncommitted tree. `verify-release.sh` checks the version, the architecture and `LSUIElement`, but not that the bundle came from `release:app` at the current commit.
**Fix:** Have `release:app` write `bin/gifsmith.app.release-sha` containing `git rev-parse HEAD`. Add a `release:sign` precondition that the file matches HEAD and the tree is clean.

---

_Reviewed: 2026-10-20T16:20:48Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
_Iteration: 2 (iteration 1 archived as 04-REVIEW.iter1.md)_
