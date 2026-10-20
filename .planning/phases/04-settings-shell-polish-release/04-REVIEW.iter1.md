---
phase: 04-settings-shell-polish-release
reviewed: 2026-10-20T15:32:29Z
depth: standard
files_reviewed: 136
files_reviewed_list:
  - .github/workflows/ci.yml
  - .github/workflows/release.yml
  - build/config.yml
  - build/darwin/Info.dev.plist
  - build/darwin/Info.plist
  - build/darwin/Taskfile.yml
  - docs/release-notes-1.0.0.md
  - docs/releasing.md
  - frontend/bindings/gifsmith/internal/provider/index.ts
  - frontend/bindings/gifsmith/internal/provider/models.ts
  - frontend/bindings/gifsmith/internal/services/index.ts
  - frontend/bindings/gifsmith/internal/services/models.ts
  - frontend/bindings/gifsmith/internal/services/panelservice.ts
  - frontend/bindings/gifsmith/internal/services/settingsservice.ts
  - frontend/bindings/gifsmith/internal/services/updateservice.ts
  - frontend/bindings/gifsmith/internal/settings/index.ts
  - frontend/bindings/gifsmith/internal/settings/models.ts
  - frontend/index.html
  - frontend/settings.html
  - frontend/src/app.css
  - frontend/src/App.svelte
  - frontend/src/components/ActionMenu.svelte
  - frontend/src/components/DetailsPane.svelte
  - frontend/src/components/Footer.svelte
  - frontend/src/components/LibraryGrid.svelte
  - frontend/src/components/ResultGrid.svelte
  - frontend/src/components/SetupCard.svelte
  - frontend/src/components/Tile.svelte
  - frontend/src/lib/actions.test.ts
  - frontend/src/lib/actions.ts
  - frontend/src/lib/api.ts
  - frontend/src/lib/errors.test.ts
  - frontend/src/lib/errors.ts
  - frontend/src/lib/grid.test.ts
  - frontend/src/lib/grid.ts
  - frontend/src/lib/hotkeys.test.ts
  - frontend/src/lib/hotkeys.ts
  - frontend/src/lib/library.test.ts
  - frontend/src/lib/library.ts
  - frontend/src/lib/motion.test.ts
  - frontend/src/lib/motion.ts
  - frontend/src/lib/popover.test.ts
  - frontend/src/lib/popover.ts
  - frontend/src/lib/prefs.test.ts
  - frontend/src/lib/prefs.ts
  - frontend/src/lib/providers.ts
  - frontend/src/lib/theme.test.ts
  - frontend/src/lib/theme.ts
  - frontend/src/lib/update.test.ts
  - frontend/src/lib/update.ts
  - frontend/src/main.ts
  - frontend/src/settings/AboutPane.svelte
  - frontend/src/settings/DownloadsPane.svelte
  - frontend/src/settings/GeneralPane.svelte
  - frontend/src/settings/HotkeyRecorder.svelte
  - frontend/src/settings/HotkeysPane.svelte
  - frontend/src/settings/KeyRow.svelte
  - frontend/src/settings/KeysPane.svelte
  - frontend/src/settings/main.ts
  - frontend/src/settings/SearchPane.svelte
  - frontend/src/settings/SettingRow.svelte
  - frontend/src/settings/SettingsApp.svelte
  - frontend/vite.config.ts
  - go.mod
  - internal/platform/fake.go
  - internal/platform/platform_darwin_test.go
  - internal/platform/platform_darwin.go
  - internal/platform/platform_darwin.h
  - internal/platform/platform_darwin.m
  - internal/platform/platform_stub.go
  - internal/platform/platform.go
  - internal/provider/fakeapi/fakeapi.go
  - internal/provider/finergifs/finergifs_test.go
  - internal/provider/finergifs/finergifs.go
  - internal/provider/giphy/giphy_test.go
  - internal/provider/giphy/giphy.go
  - internal/provider/giphy/languages.go
  - internal/provider/klipy/countries.go
  - internal/provider/klipy/klipy_test.go
  - internal/provider/klipy/klipy.go
  - internal/provider/provider_test.go
  - internal/provider/provider.go
  - internal/search/search_test.go
  - internal/search/search.go
  - internal/services/actions_test.go
  - internal/services/actions.go
  - internal/services/errors.go
  - internal/services/fsprobe_darwin.go
  - internal/services/fsprobe_other.go
  - internal/services/fsprobe_test.go
  - internal/services/fsprobe.go
  - internal/services/gifs.go
  - internal/services/hotkeys_test.go
  - internal/services/hotkeys.go
  - internal/services/panel.go
  - internal/services/settings_test.go
  - internal/services/settings.go
  - internal/services/updates_test.go
  - internal/services/updates.go
  - internal/services/version_test.go
  - internal/settings/settings_test.go
  - internal/settings/settings.go
  - internal/shell/position_test.go
  - internal/shell/position.go
  - internal/shell/shell.go
  - internal/update/install_test.go
  - internal/update/install.go
  - internal/update/real_darwin_test.go
  - internal/update/update_test.go
  - internal/update/update.go
  - main.go
  - README.md
  - scripts/dmg-background.swift
  - scripts/notarize.sh
  - scripts/release-version.sh
  - scripts/smoke-browser-loop.sh
  - scripts/stapler-retry.sh
  - scripts/uat/11-square.js
  - scripts/uat/15-settings-general.js
  - scripts/uat/15-settings-startup-fail.js
  - scripts/uat/16-settings-keys.js
  - scripts/uat/16-settings-search.js
  - scripts/uat/17-hotkeys-fail.js
  - scripts/uat/17-hotkeys.js
  - scripts/uat/17-polish.js
  - scripts/uat/18-backstops.js
  - scripts/uat/18-default-actions.js
  - scripts/uat/18-downloads.js
  - scripts/uat/18-grid-size.js
  - scripts/uat/19-about.js
  - scripts/uat/20-update-current.js
  - scripts/uat/20-update-fail.js
  - scripts/uat/20-update.js
  - scripts/verify-release.sh
  - wire_desktop.go
  - wire_server.go
findings:
  critical: 0
  warning: 17
  info: 20
  total: 37
status: issues_found
---

# Phase 4: Code Review Report

Reviewed in two parallel halves at standard depth: backend/release/CI (65 files, IDs CR-B/WR-B/IN-B) and frontend/UAT (71 files, IDs CR-F/WR-F/IN-F).


## Part (backend / release / CI half)

**Reviewed:** 2026-10-20T15:30:23Z
**Depth:** standard
**Files Reviewed:** 65
**Status:** issues_found

## Summary

This covers the Go backend, native glue, the update pipeline, the settings store, hotkeys, release scripts, the Taskfile, the GitHub Actions workflows and the docs, all diffed against `32eee422`. The frontend and UAT are reviewed separately.

**Update path (download, verify, open).** I traced `Checker.Latest`, `parse`, `Download`, `Verify`, `UpdateService.Download` and `OpenURL` and found no exploitable defect:
- The asset name is built from the validated tag, never read from the JSON.
- Size is capped at 200 MiB and the byte count must match exactly. The SHA-256 comes from the GitHub API over TLS.
- The first URL and every redirect pass `GitHubAllow`. Hosts are compared exactly, `..` is refused, and so are userinfo and port variants.
- The `.part` file is created with `O_EXCL`.
- The tools are called by absolute path with no shell.
- Opening the DMG is gated on `codesign --strict`, an exact `TeamIdentifier=` line, `spctl` reporting "Notarized Developer ID", and a quarantine flag being set.
- The window between Verify and Open is only reachable by a process running as the same user (IN-B08).

**Release workflow.** `${{ }}` values go through `env:`, never inline into scripts. The keychain is created per run and deleted in an `always()` cleanup. `bash -eo pipefail` covers the `tee` pipelines.

The defects found are about state consistency (hotkeys versus the store, a corrupt settings.json being overwritten), how much the signing identity is exposed in CI, and timing-dependent tests.

## Warnings

### WR-B01: A failed settings write leaves the registered hotkey different from the stored one

**File:** `internal/services/settings.go:542-563`
**Issue:** `SetHotkey` calls `s.Hotkeys.Bind(slot, accel)` first, which unregisters the old combo and registers the new one with Carbon. Only then does it run `s.Store.Update`. If that write fails (disk full, permissions, rename error), it returns `errPrefsWriteFailed` but never rolls back `Bind`:
- The new combo stays registered and is reported by `Hotkeys.Bindings()`, so the returned Snapshot shows it.
- `settings.json` still holds the old combo.
- On the next launch `Start` registers the old combo.

The struct comment says `hkMu` makes "the Store end with what is registered", and this path breaks that. The user also gets a "settings file could not be written" error while the new hotkey is in fact live.
**Fix:** Roll the binding back when the write fails:
```go
old := *slotOf(&before, slot) // capture s.Hotkeys.Bindings() before Bind
if err := s.Hotkeys.Bind(slot, accel); err != nil { return s.Snapshot(ctx), err }
...
case err != nil:
    if rbErr := s.Hotkeys.Bind(slot, old); rbErr != nil {
        log.Printf("hotkey rollback failed: %s", slot)
    }
    return s.snapshotOf(p), errPrefsWriteFailed
```
(For `open`, roll back to `cmp.Or(old, DefaultOpenHotkey)`.)

### WR-B02: At startup, a hotkey that failed to register is still recorded as bound

**File:** `internal/services/hotkeys.go:109-117`
**Issue:** `Start` writes `*slotOf(&h.bound, slot) = c` before calling `register`, and keeps it when `register` fails (it only skips the success log). Effects:
- `Bindings()`, and so the Settings Hotkeys pane and the tray key equivalent (`sh.SetOpenAccelerator(hk.Bindings().Open)` in `main.go:181`), show a shortcut that does nothing.
- No error reaches the UI, although the comment at `main.go:176` says registering in ApplicationStarted is done so that "an OS rejection comes back to Start".
- `ArmTest` accepts the slot, so Test arms a 5 s window that can never fire.
- A later `Bind` calls `Unregister(old)` on a combo that was never registered (it is logged and ignored).

This contradicts the CONTEXT decision "show an error when registration fails".
**Fix:** Record only bindings that registered, and remember the failures so the Hotkeys pane can show them:
```go
if err := h.register(slot, c); err != nil {
    h.failed[slot] = c // surface via Bindings/Snapshot (e.g. a per-slot error kind)
    continue
}
*slotOf(&h.bound, slot) = c
```
Alternatively, have `Start` return `map[slot]error` and emit `hotkey:error` / include it in `Snapshot`.

### WR-B03: A corrupt or unreadable settings.json is silently reset, then overwritten on the next write

**File:** `internal/settings/settings.go:379-393`
**Issue:** `OpenStore` treats any read or parse error other than ENOENT as a first run (`s.p = Prefs{}`). The next `Store.Update` replaces the file atomically with defaults. `SetLastProvider` writes on every provider-tab switch, so this happens within seconds of normal use. Before this phase the file only held `lastProvider`. It now holds hotkeys, the download folder, default actions, the theme, the launch-at-login choice and update preferences, so the user loses all of them with no recoverable copy.

The trigger is not exotic: a single wrong-typed field (for example a hand-edited `"pageSize": "36"`) fails `json.Unmarshal` and discards every field. EACCES and EIO are treated the same way. This is the "silent fallback, then corruption" pattern the project's CLAUDE.md forbids. It is also inconsistent with the library store, which by design never overwrites an unreadable file (see `UAT_SEED_FAVORITES` in `scripts/smoke-browser-loop.sh`).
**Fix:** If the file exists but cannot be read or parsed, keep a copy before running on defaults, and refuse to write over a file that cannot be read:
```go
if err != nil && !errors.Is(err, os.ErrNotExist) {
    bak := filepath.Join(dir, prefsFile+".corrupt-"+time.Now().Format("20060102T150405"))
    _ = os.Rename(filepath.Join(dir, prefsFile), bak) // preserve the user's data
    log.Printf("settings unreadable, moved aside to %s: %v", filepath.Base(bak), err)
    s.p = Prefs{}
}
```
For a non-ENOENT read error, set a `readOnly` flag that makes `Update` return an error instead of writing.

### WR-B04: The Developer ID identity is unlocked for the whole build, which runs third-party code

**File:** `.github/workflows/release.yml:82-125`
**Issue:** The "Signing keychain and notary profile" step does the following before any build:
- Imports the Developer ID private key with `-T /usr/bin/codesign` and `set-key-partition-list -S apple-tool:,apple:,codesign:`.
- Makes that keychain the default and sets it unlocked for 6 h.
- Stores the notary API key.

The next step, `wails3 task darwin:release`, runs:
- `go mod tidy` (`common:go:mod:tidy`)
- `pnpm install` (`build/Taskfile.yml:57`)
- binding generation
- `vite build`, which runs every Vite and Svelte plugin and their transitive dev dependencies in-process.

Any compromised locked dependency running in that step can call `/usr/bin/codesign --sign "Developer ID Application: Ken Scott (S2K7P43927)"` on arbitrary binaries, or call `notarytool` with the stored profile. pnpm 12 blocking lifecycle scripts reduces the risk but does not remove it, because Vite and plugin code runs during the build.
**Fix:** Build first, then unlock credentials only for the signing and notarization steps:
```yaml
- name: Build (no credentials)
  run: wails3 task darwin:package ARCH=arm64
- name: Signing keychain and notary profile
  ...
- name: Sign, notarize, verify
  run: wails3 task darwin:release   # make release:assemble skip the package dep when bin/gifsmith.app is fresh
```
Alternatively, split `release:assemble` into `release:build-app` and `release:sign` tasks so the keychain step can sit between them.

### WR-B05: `workflow_dispatch` skips the main-ancestry gate but still signs and notarizes

**File:** `.github/workflows/release.yml:11, 46-55`
**Issue:** The `push` path refuses a tag that is not on `origin/main`. The `workflow_dispatch` path ("dry run") skips that check, yet still runs the keychain, sign, notarize and staple steps with the production identity and notary key. The only thing stopping a dispatch from an arbitrary ref is the environment deployment-branch policy, which lives outside the repo ("main and `v*` tags" per `docs/releasing.md:49`).

Any `v*` tag passes that policy, including one that was rejected by the push gate because it is not on main. So `gh workflow run release.yml --ref v9.9.9` gets unreviewed code Developer ID-signed and notarized by Apple under the Team ID. `docs/releasing.md:32-33` says the job "refuses before touching any credential when the tag ... is not on `main`", which is only true for pushes.
**Fix:** Enforce the gate in the workflow for both events:
```bash
git merge-base --is-ancestor HEAD origin/main || { echo "::error::$GITHUB_REF is not on main"; exit 1; }
```
Move it above the `if`, or add `if: github.event_name == 'push' || github.ref == 'refs/heads/main'` at job level.

### WR-B06: Timing-based tests are likely to flake on loaded CI runners with `-race`

**File:** `internal/services/updates_test.go:377-405` (also `:428-430`, and `internal/services/settings_test.go:268, 752, 773, 789`)
**Issue:** `TestRunAutoUpdates` relies on wall-clock timing:
- It requires at least 2 HTTP checks within 150 ms.
- It assumes an in-flight check finishes within 40 ms after `on.Store(false)`, then requires zero new hits over the next 100 ms.
- It requires `RunAutoUpdates` to return within 100 ms of `cancel()`.

CI runs `go test -race` on shared `macos-15` runners (`ci.yml:61`), where an httptest round trip under the race detector can easily exceed 40 ms. A check that started just before `on.Store(false)` then lands after `n` is sampled and fails the test. The `settings_test` assertions of the form "sleep 50 ms then assert no probe happened" have the same problem in reverse: they pass vacuously if the goroutine is simply slow.
**Fix:**
- Inject the timer, or pass a channel-driven `tick` into `RunAutoUpdates`.
- Make the handler signal each hit on a channel and assert on received events instead of sleeping.
- For "no further checks", block the handler, flip `enabled`, release the handler, then advance a fake tick and assert no new request.

### WR-B07: A failed hotkey restore clears Open, reports the opposite, and leaves the tray stale

**File:** `internal/services/hotkeys.go:156-160`, `internal/services/errors.go:79`
**Issue:** When registering the new combo fails and re-registering the old one also fails, `Bind` sets `*cur = ""` and returns `errHotkeyFailed`, whose text is "macOS couldn't register the shortcut; **the previous one still works**". That is false in this branch, where neither shortcut works. Effects for the `open` slot:
- `Bindings().Open` becomes `""`, which breaks the Snapshot invariant "Open always set" (`settings.go:221`).
- `SetHotkey` returns early, so `settings.json` keeps the old combo.
- `OnOpenHotkey` is not called, so the tray keeps showing the dead accelerator.
- The user has no global way to open the panel until relaunch, and nothing tells them.
**Fix:** Return a distinct kind (for example `hotkey_register_failed_lost`, "couldn't register the shortcut, and the previous one was lost; choose another") when the restore fails. Have `SetHotkey` update the tray (`OnOpenHotkey("")`, which should hide the key equivalent) and emit `settings:changed` so the UI reflects the unbound state.

## Info

### IN-B01: `SetLaunchAtLogin` is not serialized

**File:** `internal/services/settings.go:479-506`
**Issue:** `Enable()`/`Disable()` and the later `Store.Update` are not under any service mutex. Two quick toggles (on, off) can interleave as Enable, Disable, Store(off), Store(on), leaving the pref `on` while the OS login item is off. `ApplyStartup` then re-enables it at the next launch against the user's last click.
**Fix:** Hold a dedicated mutex across `apply()` and `Store.Update`, as `hkMu` does for hotkeys.

### IN-B02: The tray update item is never cleared, and a newer non-installable release is reported as "current"

**File:** `internal/services/updates.go:98-123`, `internal/shell/shell.go:180-187`
**Issue:** `OnAvailable` is only ever called with a version, never with `""`. If a later check returns "current" or `ErrNotInstallable` (for example the release was pulled, or its DMG asset was deleted), Settings says up to date while the tray still offers "Update to Gifsmith X…". The `ErrNotInstallable` branch also tells the user "current" even when a newer version exists.
**Fix:** Call `u.OnAvailable("")` (after unlock) whenever the resulting state is not `available`. Consider a distinct `unavailable` state for a newer release with no installable DMG.

### IN-B03: The version gate does not check every file the docs say to bump

**File:** `scripts/release-version.sh:14-23`, `docs/releasing.md:15`
**Issue:** `docs/releasing.md` lists `build/config.yml` `info.version` as a required bump, but neither `release-version.sh` nor `version_test.go` checks it, and nobody checks `build/darwin/Info.dev.plist` either. They will drift silently on 1.0.1.
**Fix:** Add `grep -qxF "  version: \"$v\" # The application version" build/config.yml || fail ...`, or drop the line from the docs if the file is unused.

### IN-B04: The release job restores the shared Go cache

**File:** `.github/workflows/release.yml:58-60`
**Issue:** `actions/setup-go` caches by default, so the signing job restores a module and build cache written by other workflow runs on `main`. This is standard cache-poisoning hardening for jobs that hold signing secrets.
**Fix:** Add `cache: false` to `setup-go` in `release.yml`.

### IN-B05: Keychain and .p12 passwords are passed on argv

**File:** `.github/workflows/release.yml:93-100`
**Issue:** `security create-keychain -p`, `unlock-keychain -p`, `import -P "$MACOS_CERT_PASSWORD"` and `set-key-partition-list -k` all put passwords in the process arguments, which any process on the runner can read via `ps`. Combined with WR-B04, build-time code could read them.
**Fix:** Acceptable on an ephemeral runner once WR-B04 is fixed. Otherwise use `security import ... -P` from a here-string-fed variable inside a subshell that runs before any third-party code.

### IN-B06: The release build runs `go mod tidy` and a non-explicit `pnpm install`

**File:** `build/darwin/Taskfile.yml:40` (`common:go:mod:tidy`), `build/Taskfile.yml:57`
**Issue:** The tag build changes `go.mod`/`go.sum` before compiling. Memory notes it already adds `stretchr/objx` locally, so the shipped binary is not built from exactly the committed module graph. `pnpm install` is frozen only because pnpm detects `CI=true`.
**Fix:** In CI, skip tidy and run `pnpm install --frozen-lockfile` explicitly. Optionally add `git diff --exit-code go.mod go.sum` after the build.

### IN-B07: Nothing asserts that the DMG itself is notarized

**File:** `scripts/verify-release.sh:79-80`, `.github/workflows/release.yml:125`
**Issue:** `spd` (the DMG `spctl` result) is checked only for exit status. The workflow's `grep 'source=Notarized Developer ID'` is satisfied by the app's line (`sp`) alone. For the local 1.0.0 release path, no step asserts the DMG's notarization source; `stapler validate` covers this indirectly. CI's `TestRealDMG` does check it.
**Fix:** `grep -qF 'source=Notarized Developer ID' <<<"$spd" || fail "dmg is not notarized: $spd"`.

### IN-B08: There is a short race between Verify and Open on the update DMG

**File:** `internal/services/updates.go:173-178`
**Issue:** The DMG is verified by path and then opened by path. A process running as the same user could replace the file in `~/Library/Caches/com.kenscott.gifsmith/updates` between the two calls. Same-user code is outside the threat boundary, and the quarantine flag plus Gatekeeper at mount and launch still apply, so this is noted only for completeness.
**Fix:** None required. Optionally re-hash the file right before `Open`, or open it via a file descriptor.

### IN-B09: `SuspendHotkeys` is not covered by `hkMu`

**File:** `internal/services/settings.go:538-541, 573-574`
**Issue:** `SetHotkey` calls `Resume()` and then `Bind()` under `hkMu`, but `SuspendHotkeys` does not take `hkMu`. A recorder that restarts in between can leave `Bind` in its suspended branch. That branch stores the combo without registering it, so a later registration failure in `Resume` is only logged.
**Fix:** Take `s.hkMu` in `SuspendHotkeys`/`ResumeHotkeys` too.


_Reviewed: 2026-10-20T15:30:23Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_


## Part (frontend / UAT half)

**Reviewed:** 2026-10-20T15:32:29Z
**Depth:** standard
**Files Reviewed:** 71
**Status:** issues_found

## Summary

This covers the Svelte 5 frontend for Phase 4: the panel changes, the Settings window and its six panes, the hotkey recorder, the update UI, the lib modules and their tests, the generated bindings, and the Playwright UAT scripts. The Go backend, release scripts and CI are in the parallel review.

Checks run during the review:
- `svelte-check` reports 0 errors and 0 warnings.
- `node --test src/lib/*.test.ts` passes 101 of 101.
- The bindings were regenerated into a scratch directory with the Taskfile flags (`wails3 generate bindings -clean=true -ts -i`). The output is byte-identical to `frontend/bindings`, so the bindings are **not stale**.
- There is no `{@html}`, `innerHTML` or other HTML sink anywhere in `frontend/src`. Provider strings are only rendered as text.

No blockers were found. The most important problem comes from one fact the frontend ignores: **the Settings window is created hidden at app launch** (`internal/shell/shell.go:76-82`). Its page is never reloaded, so:
- every snapshot-derived field the panes show dates from launch (WR-F01);
- the General pane polls Accessibility every 2 s for the whole app lifetime, even when Settings has never been opened (WR-F02).

Other themes:
- Several bridge calls are fire-and-forget with no rejection handling. Their failure states fall back silently to wrong UI; for example, the Keys pane shows "Not set" for a saved key (WR-F03, WR-F04).
- The hotkey recorder has keyboard and accessibility gaps (WR-F05 to WR-F07).
- The update UI has two state bugs (WR-F08, WR-F09).
- Settings writes use a last-response-wins pattern (WR-F10).

Several UAT assertions are weaker than they look (IN-F01 to IN-F03).

I did not dig further into the known deferred item (the CancelledRejectionError from `cancel(...)` on tab switch) and do not re-report it.

## Warnings

### WR-F01: Settings panes render a launch-time snapshot, so D8 (missing folder) and the Startup checkbox go stale

**File:** `frontend/src/settings/main.ts:8`, `frontend/src/settings/SettingsApp.svelte:27,31`, `frontend/src/settings/DownloadsPane.svelte:12,46`, `frontend/src/settings/GeneralPane.svelte:31`
**Issue:** `settings/main.ts` reads `snapshot()` once, when the hidden Settings webview loads at app launch. After that, `SettingsApp.snap` only changes on `settings:changed`, which fires only on a settings write. Two snapshot fields are computed by Go from live OS state:
- `downloadDirMissing`, from a stat of the folder;
- `launchAtLogin`, which comes from `Autostart.IsEnabled()` and not from the pref (`settings.go` `snapshotOf`).

Neither is ever re-read when Settings is shown. So:
- If the download folder is deleted after launch, opening Downloads never shows D8, which UI-SPEC requires "when the pane opens". The user only finds out from an F4 failure in the panel.
- If the login item is switched off in System Settings, the Startup checkbox keeps showing it as on.

`settings:show` only re-reads Accessibility (GeneralPane) and key status (KeysPane).
**Fix:** Re-read the snapshot on every show and pass it down.
```ts
// SettingsApp.svelte onMount
const offShow = Events.On('settings:show', async () => {
  try { snap = await snapshot(); } catch { /* keep the last snapshot */ }
});
```
Then have DownloadsPane and GeneralPane follow the `initial` prop, using a `$effect` that copies `initial` into local state, rather than `untrack`-ing it once at mount.

### WR-F02: The Accessibility poll runs every 2 s for the app's lifetime in a hidden window, and can leak an orphan interval

**File:** `frontend/src/settings/GeneralPane.svelte:58-67,85`
**Issue:** UI-SPEC says to poll "every 2 s while the General pane is visible and the status is Off". In practice:
- General is the default pane, and the Settings page mounts hidden at launch.
- So `readAX()` starts `setInterval(readAX, 2000)` at launch for every user without Accessibility. Copy-only use is a supported mode, so that is a normal setup.
- The poll never stops while the window is hidden or was never opened. Nothing stops it on hide.

There is also a leak:
- The cleanup calls `clearInterval(axTimer)` but leaves `axTimer` set.
- Suppose a `readAX()` is in flight at unmount and `axTimer` was `undefined` (status was On, or the first read had not returned yet). If that read resolves `false`, `axTimer ??= setInterval(...)` creates an interval that is never cleared and keeps writing to a destroyed component.

`readAX` also has no `try/catch`, so a failing bridge raises an unhandled rejection every 2 s (see WR-F03).
**Fix:** Poll only while the window is visible, and guard against running after unmount.
```ts
let alive = true;
async function readAX() {
  let on: boolean;
  try { on = await accessibilityStatus(); } catch { return; }
  if (!alive) return;
  axOn = on;
  if (on || document.hidden) { clearInterval(axTimer); axTimer = undefined; }
  else axTimer ??= setInterval(readAX, 2000);
}
// onMount: also listen for visibilitychange (or a Go 'settings:hide' event) to stop it;
// cleanup: alive = false; clearInterval(axTimer); axTimer = undefined;
```

### WR-F03: Fire-and-forget bridge calls with no rejection handling leave silent wrong states (and unhandled rejections)

**File:**
- `frontend/src/settings/KeysPane.svelte:32-34`
- `frontend/src/settings/SearchPane.svelte:29-34`
- `frontend/src/settings/GeneralPane.svelte:59,68`
- `frontend/src/settings/HotkeyRecorder.svelte:33,40`
- `frontend/src/settings/KeyRow.svelte:168`
- `frontend/src/settings/AboutPane.svelte:22`
- `frontend/src/App.svelte:302,704`

**Issue:** These calls are never awaited inside a `try`, and none has a `.catch`:
- `keyStatus()`
- `choices()`
- `accessibilityStatus()` and `updateStatus()`
- `suspendHotkeys()` and `resumeHotkeys()`, which go through `inOrder()`; it returns the uncaught promise
- `openExternal(...)`
- `revealUpdates()` (the T12 `View` action)
- `openSettings()` (⌘,)

What each failure does:
- **`choices()`:** the four Search selects stay empty for good. Nothing tells the user why, and they cannot change rating, language, country or page size.
- **`keyStatus()`:** see WR-F04.
- **`suspendHotkeys()`:** the recorder believes Gifsmith's hotkeys are suspended when they are not, so pressing the current combo fires it instead of being recorded.

Every case also raises an unhandled rejection. The UAT harness counts those as `pageerror` failures, so this is the same class of flake as the deferred CancelledRejectionError.
**Fix:** Handle each call:
- Load paths (`keyStatus`, `choices`): add `try/catch` that sets an explicit error state and shows an inline error row ("Couldn't load these options. Try again.").
- Fire-and-forget calls: add `.catch(() => {})` or log.
- `suspendHotkeys`: show H-style feedback (or stop recording) if it rejects.

### WR-F04: The Keys pane shows "Not set" and "Add Key…" for a saved key while loading, and permanently if the read fails

**File:** `frontend/src/settings/KeysPane.svelte:30,49`, `frontend/src/settings/SettingsApp.svelte:68-75`
**Issue:** `statuses` starts as `{}` and each card falls back to `{ status: 'not_set' }`. Only the active pane is mounted (`{#if i === active}`), so KeysPane remounts on every visit. Each time, both cards render "Not set" and the primary `Add Key…` button until `KeyStatus()` returns. That call reads the Keychain once per provider through go-keyring, which runs `/usr/bin/security`, so the wrong state can be visible.

A user who clicks `Add Key…` during that window opens an editor for a provider that already has a key. If `KeyStatus()` rejects, the wrong state stays and nothing reports it. This is the silent fallback CLAUDE.md warns against: "`or {}` converts hard failures into silent corruption".
**Fix:** Model the unknown state and the failed state explicitly.
```ts
let statuses = $state<Record<string, KeyStatus> | null>(null);
let loadError = $state('');
async function load() {
  try { statuses = Object.fromEntries((await keyStatus()).map((s) => [s.provider, s])); loadError = ''; }
  catch { loadError = "Couldn't read your keys from the macOS Keychain. Try again."; }
}
// render cards only when statuses !== null; otherwise a neutral placeholder or loadError
```

### WR-F05: Cancelling a recording (Esc, blur, Tab) leaves a stale H1 error under the row

**File:** `frontend/src/settings/HotkeyRecorder.svelte:36-41,66,69,76,100`
**Issue:** UI-SPEC says cancel shows the "previous value, no error". Here is what happens instead:
1. Pressing a key without ⌘, ⌥ or ⌃ calls `onfail('hotkey_no_modifier')` (H1, `role="alert"`) and recording continues.
2. Pressing Esc, Tabbing away, or blurring calls `stop()`.
3. `stop()` never clears the message.

So the H1 error stays under a field that is no longer recording and shows its old, valid binding. The only code that clears messages runs at `start()` and after a successful save. UAT 17-hotkeys.js:40-51 clears H1 with a successful combo, so it never covers this path.
**Fix:**
```ts
function cancel() { if (!recording) return; onfail('', '', ''); stop(); }
// Esc, Tab, blur → cancel(); keep stop() (no clear) for the H2/H3/H4 paths that must keep their error.
```

### WR-F06: Shift+Tab while recording is swallowed and reported as H1

**File:** `frontend/src/settings/HotkeyRecorder.svelte:65-67,74-76`
**Issue:** `mods` includes `shiftKey`, so `e.key === 'Tab' && !mods` does not match Shift+Tab. The code then:
1. calls `e.preventDefault()`, so focus does not move back;
2. passes the event to `accelFromEvent`, which maps `code: 'Tab'` to `Tab` with no ⌘, ⌥ or ⌃;
3. gets `no_modifier`, so the recorder shows H1 and keeps recording.

A keyboard user moving backward cannot leave the field without knowing to press Esc. The spec lists Tab as a cancel key, and Settings is meant to be fully keyboard-navigable (backstop c).
**Fix:**
```ts
if (e.key === 'Tab' && !e.metaKey && !e.ctrlKey && !e.altKey) return cancel(); // Tab and Shift+Tab
```

### WR-F07: The recorder's accessible name hides the current shortcut and the recording state

**File:** `frontend/src/settings/HotkeyRecorder.svelte:93-102`
**Issue:** The element is `role="button"` with `aria-label="{label} shortcut"`. An `aria-label` replaces the element's content as its accessible name. VoiceOver therefore announces "Open Gifsmith shortcut, button", and never:
- the bound combo (`⌃⌥⌘G`, or `Not set`);
- the `Type shortcut…` prompt;
- the live modifiers being held.

A screen-reader user cannot find out what any hotkey is set to, and gets no confirmation that recording started. `aria-describedby` points only at the message line, which is empty when idle.
**Fix:** Keep the spec's name, but describe the value and state.
```svelte
<span id="{describedby}-value" class="sr-only">{recording ? 'Recording. Type a shortcut.' : value ? spoken(value) : 'Not set'}</span>
<div … aria-describedby="{describedby}-value {describedby}" aria-pressed={recording}>
```
`spoken('Cmd+Ctrl+Option+G')` should return "Control Option Command G". Keep the exact-name UAT selectors working.

### WR-F08: The update progress bar reads 100% for the whole download when the size is unknown

**File:** `frontend/src/settings/GeneralPane.svelte:242`
**Issue:** The markup is `<progress max={upd.total || undefined} value={upd.done}>`. When `total` is 0, Svelte drops `max`, but `value` is still set. Per the HTML spec, `max` then defaults to 1 and any `value` of 1 byte or more is clamped to it. The bar is full from the first byte and stays full.

U5 explicitly defines this case ("without ` {n}%` while the size is unknown"), so the UI does have the state, and it renders wrong. Go seeds `Total` from the release asset's `size`, which can be 0 or missing. UAT 20-update.js only exercises a known total.
**Fix:** Omit `value` so the bar is indeterminate.
```svelte
{#if upd.total > 0}
  <progress class="update-progress" max={upd.total} value={upd.done}></progress>
{:else}
  <progress class="update-progress"></progress>
{/if}
```

### WR-F09: An update error stays next to a later, contradicting status

**File:** `frontend/src/settings/GeneralPane.svelte:41,69,127-143,233`
**Issue:** `updError` is cleared only when the user clicks Check or Download again. The Settings page lives for the whole launch, and automatic checks run every 24 h, each emitting `update:status`.

So a U-E1 error ("Couldn't check for updates…") stays on screen after a later automatic check succeeds, right beside "Gifsmith 1.0.0 is up to date." or U4. The same applies to U-E2 to U-E4 after a later successful check.
**Fix:** Clear the error whenever a new attempt starts, including automatic ones. Do not clear on `available`: Go's failure path emits a restored `available` before the RPC rejects, so clearing then could race the fresh error.
```ts
const offUpdate = Events.On('update:status', (e) => {
  upd = e.data as UpdateStatus;
  if (upd.state === 'checking' || upd.state === 'downloading') updError = '';
});
```

### WR-F10: Last-response-wins snapshot writes race each other and `settings:changed`

**File:** `frontend/src/settings/GeneralPane.svelte:94,105,119-120,149`, `frontend/src/settings/SearchPane.svelte:43`, `frontend/src/settings/DownloadsPane.svelte:24,34`
**Issue:** Each handler awaits `setPref` (or a similar call) and then overwrites local state with the snapshot that call returned. SearchPane's `values = valuesOf(...)` and DownloadsPane's `snap = ...` replace every field, not just the one changed.

Wails serves each bound call as its own HTTP request, concurrently. So:
- responses can arrive out of order relative to each other;
- responses can also arrive after a newer `settings:changed` event.

Arrowing through the Theme radios, or changing two Search selects quickly, sends several calls. A late older response then overwrites newer state: the radio or select shows a value Go no longer stores. Because SearchPane replaces all fields, an older response for one field can also revert another field's control.
**Fix:** Treat `settings:changed` (ordered by Go) as the only source of truth for displayed values. Use the call's result only on failure, to revert. Or tag each write with a sequence number and ignore responses older than the last applied one.

## Info

### IN-F01: The ⌘, UAT assertion passes even if ⌘, does nothing

**File:** `scripts/uat/15-settings-general.js:99-105`
**Issue:** The only check is that the footer shows no error class after ⌘,. It would still pass if `routeKey`'s `'settings'` case or `onMainKeydown`'s `openSettings()` call were deleted.
**Fix:** Make the server build's shell fake record `OpenSettings` calls, for example in a log line or a fixture endpoint the script can query, and assert on that.

### IN-F02: Two other UAT assertions cannot fail

**File:** `scripts/uat/19-about.js:56`, `scripts/uat/18-grid-size.js:73`
**Issue:**
- `fits: pane.scrollHeight >= pane.clientHeight` is always true.
- `... > 1 && d3.x !== 8` accepts any menu clamped against the left inset, including one placed on the wrong side of the tile. The placement check therefore only holds when the left position is not clamped.

**Fix:** Drop `fits`, or replace it with a real overflow assertion. In 18-grid-size, assert the exact expected `left` (152 for the Large third tile) without the escape hatch.

### IN-F03: 18-backstops hardcodes copy and has an ordering race

**File:** `scripts/uat/18-backstops.js:58-72,193-198`
**Issue:**
- (b) measures a hand-copied list of K/K-T strings, not what `errors.ts` produces. If the copy in `errors.ts` changes, the backstop keeps passing on the old strings.
- (d) selects two actions in B and immediately loads page A. A reads `snapshot()` before it subscribes to `settings:changed`, so if either `SetPref` lands after A's snapshot, the hints assertion fails. This is a second possible flake source in this script, besides the deferred one.

**Fix:**
- (b): drive the real messages through the fixture (bad key, Test, and so on), or generate the list from `errors.ts` at build time.
- (d): wait for `primary.inputValue()` and `secondary.inputValue()` to settle and add a short wait for the write, or poll A's hints, before asserting.

### IN-F04: The `stale` flag is only cleared in `show()`, so fresh fetches still trigger a refetch on the next show

**File:** `frontend/src/App.svelte:126,470-472,511,564`
**Issue:**
- After a parameter change, `keySaved`, `retry()` and `onInput` (through `load`/`runFirst`) all fetch a first page with the new parameters.
- None of them clears `stale`.
- So the next `panel:show` runs `show(active)` and loads the same page again. That costs a request against keys capped at 100 per hour, unless Go's cache absorbs it.

**Fix:** Clear `s.stale = false` in `runFirst`.

### IN-F05: `hasModifier` is exported but used only by tests

**File:** `frontend/src/lib/hotkeys.ts:86`
**Issue:** No production caller uses it; it is dead code.
**Fix:** Remove it, or use it where the recorder checks for a modifier.

### IN-F06: Unmounting `HotkeyRecorder` while it is recording does not resume hotkeys

**File:** `frontend/src/settings/HotkeyRecorder.svelte`, `frontend/src/settings/SettingsApp.svelte:33-36`
**Issue:** The component has no teardown. Removing a focused element does not reliably fire `blur`. So if `settings:reveal` switches panes mid-recording, `resumeHotkeys()` is never sent. Go's `WindowLostFocus` hook resumes hotkeys in practice, so this is defence in depth.
**Fix:** `onDestroy(() => { if (recording) resumeHotkeys().catch(() => {}); });`

### IN-F07: A Test that fails to arm shows the "another app" message

**File:** `frontend/src/settings/HotkeysPane.svelte:67-71`
**Issue:** If the `testHotkey` RPC rejects (for example `errHotkeyInvalid`), the row shows H7: "Nothing happened. Another app may be using {combo}…". That blames a clash that was never tested.
**Fix:** Use the generic "Couldn't save this setting. Try again." (or a dedicated "Couldn't start the test") on RPC failure.

### IN-F08: In Small, the third column's action menu overlaps its tile

**File:** `frontend/src/lib/popover.ts:19-23`
**Issue:** In Small, the third tile (x≈294) has no room on the right, so the menu opens on the left at -34 and is clamped to 8. It then covers about 34px of the selected tile. UI-SPEC accepts overlap only for the Large middle column.
**Fix:** Accept it and update the spec, or pick the side with more room when neither side fits.

### IN-F09: Both entry points block mount on `snapshot()` with no error path

**File:** `frontend/src/main.ts:7`, `frontend/src/settings/main.ts:8`
**Issue:** Before this phase, the panel mounted without a snapshot. Now a bridge failure at load leaves a blank panel (the app's only UI) or a blank Settings window, and only an unhandled rejection records it. Go's `Snapshot` cannot itself fail, so the exposure is limited to transport failures.
**Fix:** `try`/`catch`: mount with defaults plus a visible error, or retry once.

### IN-F10: A text default action on a Clip with no `gifUrl` is advertised but unavailable

**File:** `frontend/src/lib/prefs.ts:43-47`, `frontend/src/lib/actions.ts:39`
**Issue:**
- The menu hides Copy Link and Copy Markdown for such Clips (`linkable`).
- `effectiveAction` still returns `copyLink` when the user chose it.
- So the footer shows "↵ Copy Link", and ↵ fails with F11.

**Fix:** In `effectiveAction`, fall back to the file action (or the other default) when the chosen text action does not apply to the item.

### IN-F11: Sidebar tabs point `aria-controls` at panels that are not in the DOM

**File:** `frontend/src/settings/SettingsApp.svelte:62,68-75`
**Issue:** Only the active pane is rendered, so five of the six `aria-controls="pane-…"` references point at missing ids.
**Fix:** Set `aria-controls` only on the selected tab, or keep every panel mounted with `hidden`.


_Reviewed: 2026-10-20T15:32:29Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_

