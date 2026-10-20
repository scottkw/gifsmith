---
phase: 04-settings-shell-polish-release
fixed_at: 2026-10-20T16:52:00Z
review_path: .planning/phases/04-settings-shell-polish-release/04-REVIEW.md
iteration: 2
findings_in_scope: 7
fixed: 7
skipped: 0
status: all_fixed
---

# Phase 4: Code Review Fix Report (iteration 2)

**Fixed at:** 2026-10-20T16:52:00Z
**Source review:** .planning/phases/04-settings-shell-polish-release/04-REVIEW.md
**Iteration:** 2 (the iteration-1 report is kept in `04-REVIEW-FIX.iter1.md`)

**Summary:**
- Findings in scope: 7 (WR-01..WR-07; Info items out of scope)
- Fixed: 7
- Skipped: 0

`workflow.use_worktrees` is `false`, so every edit and commit was made in the main checkout on
`gsd/phase-04-settings-shell-polish-release`. Every gate below also ran there, so the results can be
reproduced from this tree.

Two Info items were also covered, because the warnings' fixes required it:
- IN-08 is covered by WR-01.
- IN-02 is covered by WR-03.

## Fixed Issues

### WR-01: The signing step still runs on a runner the build step's code could tamper with

**Files modified:** `.github/workflows/release.yml`, `build/darwin/Taskfile.yml`, `docs/releasing.md`, `scripts/release-stamp.sh` (new)
**Commit:** 7abb1b0
**Applied fix:**
- **`build` job.** It has no environment, no secrets and `contents: read`. It runs the gate, pnpm, the pinned `wails3` and `darwin:release:app`. It then tars `gifsmith.app` plus its stamp (an artifact zip would drop the executable bit) and uploads the tar with `actions/upload-artifact@cf430e0… # v7.0.2`, with a retention of 1 day.
- **`sign` job.** It has `needs: build`, environment `release` and `contents: write`, and runs on a fresh runner. Its steps:
  - Checkout, then `setup-go` with `cache: false`, so a cache the build job saved cannot plant objects.
  - `go install` of the pinned `wails3` only. There is no setup-node, pnpm, Vite, `go mod tidy` or `go build`.
  - `actions/download-artifact@9000827… # v8.0.2`.
  - The unpack step refuses any tar entry other than `gifsmith.app/…` and `gifsmith.app.release`. bsdtar already refuses absolute paths, `..` and writes through symlinks.
  - Gatekeeper, the keychain step, then `darwin:release:sign`. The keychain is deleted before `go test` (TestRealDMG) compiles anything.
- **Draft release.** The draft-JSON `go test` (TestReleaseJSON) now runs in its own step without `GH_TOKEN`.
- **`release:app`.** It refuses a tree whose tracked files differ from `HEAD`. After the build it stamps the bundle: `bin/gifsmith.app.release` holds the commit plus a SHA-256 over every file's path and contents (`scripts/release-stamp.sh`).
- **`release:sign`.** It refuses unless the stamp matches `HEAD` and the bundle's current contents (IN-08). This rejects a leftover bundle from `install:dev`, from `package:universal` or from another commit.
- **Comments and docs.** These now describe the two jobs and say plainly what the stamp proves: provenance, not trust (a compromised build could forge it).

I tested the stamp script for a missing stamp, an added file, an edited file and a restored file, and the tar entry filter on good and bad names. `wails3 task --dry darwin:release:app` refuses a dirty tree. `actionlint` and `shellcheck` are clean.
**Status:** fixed: requires human verification. The workflow was not run on GitHub Actions; confirm it with the next dry run.

### WR-02: The main-ancestry gate is defined by the ref it gates, so it does not protect the signing secrets

**Files modified:** `.github/workflows/release.yml`, `docs/releasing.md`
**Commits:** c6c11ed, bc356dd
**Applied fix:**
- **`docs/releasing.md`** has a new "Protecting the signing secrets" section. It says the in-workflow gate guards against mistakes, not attacks, because both events run the triggering ref's own `release.yml`.
- **Repository settings.** The section documents the two controls that the orchestrator applied with the user's approval on 2026-10-20:
  - The tag ruleset "Release tags: maintainer only" (id 25489108) is active. Creating, updating and deleting `refs/tags/v*` is blocked for everyone except repository admins (bypass: always).
  - The `release` environment now requires `scottkw` as reviewer, on top of its existing `main` plus `v*` deployment policy.
- **Workflow behaviour.** Tag and dispatch runs now pause before the `sign` job until the reviewer approves them in the Actions UI. Release step 6, the Dry run section, the Secrets section and a comment on the gate step say so.

This fixer did not change any repository setting.
**Status:** fixed. The in-repo docs are done, and the repo settings were applied on 2026-10-20 by the orchestrator, with the user's approval.

### WR-03: An unreadable settings.json blocks every settings change, and the error never says why

**Files modified:**
- Go: `internal/settings/settings.go`, `internal/settings/settings_test.go`, `internal/services/errors.go`, `internal/services/settings.go`, `internal/services/settings_test.go`, `internal/services/hotkeys_test.go`
- Bindings: `frontend/bindings/gifsmith/internal/services/models.ts`, `frontend/bindings/gifsmith/internal/services/settingsservice.ts`
- Frontend: `frontend/src/lib/api.ts`, `frontend/src/lib/errors.ts`, `frontend/src/lib/errors.test.ts`, `frontend/src/settings/GeneralPane.svelte`, `SearchPane.svelte`, `DownloadsPane.svelte`
- UAT: `scripts/uat/15-settings-unreadable.js` (new), `scripts/smoke-browser-loop.sh`

**Commit:** 4eb4565
**Applied fix:**
- **Error kind.** A new `writeFailed` helper maps `settings.ErrUnreadable` to a new kind, `prefs_unreadable`, at all five `Store.Update` sites.
- **Row message.** Rows now say "Couldn't save. Gifsmith can't read its settings file. See General." through the new `saveErrorText`, which replaces the hard-coded strings in General, Search and Downloads and the fallbacks for Hotkeys and folders.
- **Snapshot flag.** `Snapshot.settingsUnreadable` comes from a new `Store.Unreadable()`.
- **Banner.** While the flag is set, General shows a banner: "Gifsmith couldn't read its settings file, so it's using the defaults and can't save changes." It has a **Reset Settings…** button. The button opens an inline confirm, which focuses Cancel, with **Cancel** and **Reset Settings** buttons.
- **`ResetSettings()` (bound).** It calls `Store.Reset`, which:
  - moves `settings.json` aside to `settings.json.<yyyymmdd-hhmmss>.bak` in the same folder, with a hard link and then a remove, so it never replaces an existing `.bak` and never deletes the bytes;
  - runs on defaults, then emits `settings:changed`.

  A file fixed by hand in the meantime is read instead of moved. A failed reset returns `prefs_reset_failed`.
- **Bindings.** Regenerated.
- **IN-02.** The two old "write failure" tests now make the settings directory read-only, so they cover the real write-failure path again.

Tests:
- Go: `TestStoreResetMovesUnreadableAside`, `TestStoreResetKeepsAFixedFile`, `TestUnreadableSettingsReset`, `TestResetSettingsFailure`.
- Node: a test of the copy.
- UAT: `15-settings-unreadable.js`. The harness gained `UAT_SEED_SETTINGS` and checks that the seeded bytes were kept as exactly one `.bak`.

**Status:** fixed: requires human verification. This is data-safety logic, and the banner was checked in headless Chromium only, not in WKWebView.

### WR-04: Open at login changes the login item, then reports that it failed

**Files modified:** `internal/services/settings.go`, `internal/services/settings_test.go`
**Commit:** 1b9e75f
**Applied fix:** `SetLaunchAtLogin` reads `IsEnabled()` first. When the Store write fails, it undoes the `Enable` or `Disable` it just made, so a reported failure leaves macOS unchanged. This is the same rollback pattern as the WR-B01 hotkey fix. If the item was already in the requested state, it does not flip it. A failed rollback is logged.

Test: `TestSetLaunchAtLoginWriteFailureRollsBack`, using a real write failure. It fails without the fix.
**Status:** fixed: requires human verification. This is state logic, and the real SMAppService was not exercised.

### WR-05: A combo refused at launch can be bound to another slot, and the store then silently drops that binding

**Files modified:** `internal/services/hotkeys.go`, `internal/services/settings.go`, `internal/services/hotkeys_test.go`
**Commit:** f256b97
**Applied fix:**
- **`Bind`.** Its duplicate check now also counts another slot's `failed` combo as taken.
- **`SetHotkey`.** Inside the Store transaction, it refuses a combo that another slot stores (Open's `""` counts as the default) with `hotkey_duplicate` and that slot's name, then reverts the registration.
- **Result.** The refused combo stays stored and stays reported as failed.

Tests (both fail without the fix):
- `TestSetHotkeyRefusesAComboFailedAtLaunch`: the default Open refused, a stored Favorites refused, and Open given another spelling of that Favorites combo. Before the fix, that last case made Normalize drop Favorites from `settings.json`.
- `TestSetHotkeyStoreDuplicateReverts`.

**Status:** fixed: requires human verification (state logic)

### WR-06: "Reset to Default" is disabled exactly when Open is unbound

**Files modified:** `frontend/src/settings/HotkeysPane.svelte`, `scripts/uat/17-hotkeys-open-unbound.js` (new)
**Commit:** b6e6607
**Applied fix:** The condition is now `disabled={!!hk.open && canonical(hk.open) === DEFAULT_OPEN}`, so an unbound Open always offers Reset to Default. The new UAT runs with `FAKE_HOTKEY_FAIL=Cmd+Ctrl+Option+G`, so the default is refused at launch. It checks that the button is enabled, and that a refused retry shows H4b while the button stays enabled. The UAT fails without the fix.
**Status:** fixed

### WR-07: The Hotkeys pane still applies its call's snapshot, so a late response can overwrite a newer `settings:changed`

**Files modified:** `frontend/src/settings/HotkeysPane.svelte`, `frontend/src/settings/HotkeyRecorder.svelte`
**Commit:** c97e49b
**Applied fix:**
- `hk` is now a read-only `$derived(initial.hotkeys)`.
- `saved` and `onsaved` are removed. `change()` and the recorder's `save()` await `setHotkey` and render only SettingsApp's snapshot.
- Every `SetHotkey` that changes registration emits `settings:changed`, and a failed call changes nothing.

All four `17-*` UATs pass.
**Status:** fixed

## Verification (all run in the main checkout)

- `gofmt -l main.go wire_desktop.go wire_server.go internal` (the CI scope): clean.
- `go vet . ./internal/...`: OK.
- `go test -race -count=1 ./internal/...`: every package passes.
- `golangci-lint run . ./internal/...` (the CI scope): 0 issues.
- `CGO_ENABLED=0 go build -tags server -o bin/gifsmith-server .`: OK.
- `node --test src/lib/*.test.ts`: 103/103.
- `svelte-check --tsconfig ./tsconfig.json`: 0 errors, 0 warnings.
- `scripts/smoke-browser-loop.sh`: SMOKE OK.
- Full UAT set, `--uat` for all 35 `scripts/uat/*.js` (33 existing plus 2 new): 35/35 pass.
- `actionlint` v1.7.12 on `.github/workflows/*.yml`: clean.
- `shellcheck`:
  - `scripts/release-stamp.sh`: clean.
  - `scripts/smoke-browser-loop.sh`: the same finding count as before the change.
- Bindings regenerated with the Taskfile's command (`wails3 generate bindings -f '' -clean=true -ts -i`): no diff. `go.sum` was restored after each `wails3` run.

Not verified natively, because no desktop input was allowed:
- the split release workflow and the reviewer pause on GitHub Actions;
- the General banner and Reset in WKWebView;
- the real login item rollback.

Residual limits worth knowing:
- `release:app`'s clean-tree check covers tracked files only (a `shortcut:` comment marks it).
- The stamp does not defend against a compromised build job. The job split is what defends against that.
- When Reset finds a file that was fixed by hand, it loads that file without re-applying its theme, hotkeys or login item until the next launch (IN-04, unchanged).

---

_Fixed: 2026-10-20T16:52:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 2_
