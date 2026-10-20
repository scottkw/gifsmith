---
phase: 04-settings-shell-polish-release
reviewed: 2026-10-20T16:51:29Z
depth: standard
iteration: 3
files_reviewed: 24
files_reviewed_list:
  - .github/workflows/release.yml
  - build/darwin/Taskfile.yml
  - docs/releasing.md
  - frontend/bindings/gifsmith/internal/services/models.ts
  - frontend/bindings/gifsmith/internal/services/settingsservice.ts
  - frontend/src/lib/api.ts
  - frontend/src/lib/errors.test.ts
  - frontend/src/lib/errors.ts
  - frontend/src/settings/DownloadsPane.svelte
  - frontend/src/settings/GeneralPane.svelte
  - frontend/src/settings/HotkeyRecorder.svelte
  - frontend/src/settings/HotkeysPane.svelte
  - frontend/src/settings/SearchPane.svelte
  - internal/services/errors.go
  - internal/services/hotkeys.go
  - internal/services/hotkeys_test.go
  - internal/services/settings.go
  - internal/services/settings_test.go
  - internal/settings/settings.go
  - internal/settings/settings_test.go
  - scripts/release-stamp.sh
  - scripts/smoke-browser-loop.sh
  - scripts/uat/15-settings-unreadable.js
  - scripts/uat/17-hotkeys-open-unbound.js
findings:
  critical: 0
  warning: 2
  info: 30
  total: 32
status: issues_found
---

# Phase 4: Code Review Report (iteration 3, final)

**Reviewed:** 2026-10-20T16:51:29Z
**Depth:** standard (re-review of fix pass 2)
**Files Reviewed:** 24
**Status:** issues_found

## Summary

This re-review covers the 8 fix commits `7abb1b0..c97e49b` (`git diff e992c61..HEAD`). The iteration-2 report is kept in `04-REVIEW.iter2.md`.

Checks run:
- `go test -race -count=1 ./internal/services/ ./internal/settings/` passes.
- `node --test src/lib/*.test.ts` passes 103 of 103.
- `go test -run 'TestRealDMG|TestReleaseJSON' ./internal/update` compiles and runs against the committed `go.sum`, so the sign job's `go test` steps do not need the build job's `go mod tidy`.
- `shellcheck scripts/release-stamp.sh` is clean. `actionlint` is not installed here, so I checked `needs`, `outputs`, `environment` and job-level `env` by hand. They are correct.
- `gh api` (read only) confirms the repository settings that WR-02 depends on:
  - Ruleset 25489108 is `active` on `refs/tags/v*` with `creation`, `update` and `deletion` rules. Only RepositoryRole 5 (admin) can bypass it.
  - The `release` environment has `scottkw` as a required reviewer, and its deployment policies are `main` (branch) and `v*` (tag).
- `go mod tidy -diff` reports that `go.sum` is not tidy (WR-02 below).
- macOS `tar` was tested on xattr-bearing files. `tar -tf` merges the AppleDouble `._*` entries, so the unpack filter does not wrongly refuse a bundle that carries xattrs.

All seven iteration-2 warnings are resolved. Fix pass 2 introduced no critical issues and two new warnings:
- **WR-01:** the build job's token is now read-only, so the "already has a release" gate can no longer see the draft releases CI creates.
- **WR-02:** the new clean-tree precondition in `release:app` is broken by the build's own `go mod tidy`.

What I checked in the areas you asked about:
- **Two-job `release.yml`.**
  - *Permissions.* `build` has `contents: read`, no `environment` and no secrets. `sign` has `contents: write` and `environment: release`, so the reviewer gate applies to `sign` only.
  - *What runs on the sign runner.* Checkout of the same commit, `setup-go` with `cache: false`, the pinned and checksum-verified `wails3`, the artifact download, and then `release:sign`. `release:sign` runs `ditto`, `codesign`, `notarize.sh`, `tiffutil`, `wails3 tool package` (`cp -R` plus `hdiutil`) and `verify-release.sh`. None of these runs pnpm, Vite, `go mod tidy` or `go build`, and none executes the bundle.
  - *Keychain lifetime.* The keychain is deleted at the end of the sign step, before `go test` compiles anything. On every failure, cancellation or timeout, the `always()` cleanup deletes it.
  - *Write token.* `GH_TOKEN` is only in the draft step's `env`. TestReleaseJSON runs in a separate step.
  - *Unpack.* The `grep -vE` filter allows only `gifsmith.app`, `gifsmith.app/...` and `gifsmith.app.release`. bsdtar's defaults refuse `..` and writes through symlinks, and strip a leading `/`; the filter would already reject a leading `/`. If `tar -t` fails, `pipefail` keeps the `if` false and the following `tar -x` fails, so the step fails closed.
  - *Artifact integrity.* The artifact is still only as trustworthy as the build job, as `docs/releasing.md` and `release-stamp.sh` now say. Within the run, nothing else can replace it.
- **`scripts/release-stamp.sh`.**
  - The stamp is deterministic: `LC_ALL=C sort -z` over paths relative to the bundle's parent folder, so the build and sign runners compute the same digest.
  - It fails closed. A `shasum` error inside `$(digest)` makes the assignment fail under `set -e`/`pipefail`, and a missing or different stamp refuses.
  - It works on a shallow checkout, because it only needs `git rev-parse HEAD`.
- **Clean-tree check.**
  - It does not break the CI build job. A fresh checkout is clean, and `release:app` has no `deps:`, so its precondition runs before `package` installs anything.
  - It does not affect dev builds (`package`, `install:dev`, `dev`). However, those builds and `release:app` itself dirty `go.sum`, which then blocks the next local release (WR-02).
- **`prefs_unreadable` reset.**
  - Data safety: `Store.Reset` holds `Store.mu` for the whole check, link and remove, so it cannot interleave with `Update`.
  - `os.Link` never replaces an existing `.bak`. If `Remove` fails, both names remain and the store stays unreadable. Read errors other than parse errors fail closed: a directory cannot be hard-linked, and an unreadable folder refuses the link. No path loses the user's bytes.
  - A double-clicked Reset is idempotent (`!s.unreadable` returns early).
  - A failed reset surfaces as `prefs_reset_failed` in the banner.
- **`SetLaunchAtLogin` rollback.** Correct for the case it targets (WR-04 resolved). See IN-09 for two edges.
- **`SetHotkey` in-transaction check.**
  - Lock order is unchanged: `hkMu`, then `Hotkeys.mu` (released), then `Store.mu` (released), then `emitMu`, which takes `Store.mu` and `Hotkeys.mu` in turn.
  - The closure passed to `Store.Update` takes no other lock.
  - The comparison is sound: `Hotkeys.Canonical` and the store's `Normalize` both use `settings.CanonicalHotkey`.
  - No deadlock.
- **New UATs.**
  - `15-settings-unreadable.js` cannot pass vacuously. The theme is applied only from `settings:changed` (`settings/main.ts:10`), so `dataset.theme === 'dark'` proves that a save went through. The harness also requires exactly one `.bak` holding the seeded bytes.
  - `17-hotkeys-open-unbound.js` proves the WR-06 fix (the button is enabled). Its "a refused retry shows H4b" half passes even if the click does nothing (IN-11).

## Iteration-2 Warning Status

| ID | Status | Notes |
|----|--------|-------|
| WR-01 | Resolved | `build` and `sign` are separate jobs on separate runners. The sign runner never installs dependencies or builds, and its Go cache is off. The keychain exists only between the keychain step and the end of the sign step. The split introduced WR-01 below. |
| WR-02 | Resolved | `docs/releasing.md` now says the in-workflow gate guards against mistakes, not attacks. The tag ruleset and the required reviewer are applied; I confirmed both with `gh api`. |
| WR-03 | Resolved | `prefs_unreadable` is a distinct kind at all five `Store.Update` sites. `Snapshot.settingsUnreadable` drives a General banner with a confirmed Reset Settings… that keeps the file as `settings.json.<time>.bak`. The panel's tab switch still fails silently (IN-12). |
| WR-04 | Resolved | A failed write undoes the `Enable`/`Disable` unless the item was already in the requested state. The test uses a real write failure. Two edges remain (IN-09). |
| WR-05 | Resolved | `Bind` counts another slot's `failed` combo as taken. The Store transaction refuses a combo that another slot stores, with Open's `""` counted as the default, and reverts the registration. Both tests fail without the fix. |
| WR-06 | Resolved | `disabled={!!hk.open && canonical(hk.open) === DEFAULT_OPEN}` keeps Reset to Default available while Open is unbound. |
| WR-07 | Resolved | `hk` is a read-only `$derived(initial.hotkeys)`. `saved`/`onsaved` are gone, and neither `change()` nor the recorder applies a call result. |

Iteration-2 info items closed by this pass: **IN-02** (the write-failure tests now use `readOnlyDir`) and **IN-08** (`release:sign` checks the stamp against `HEAD` and the bundle).

## Warnings

### WR-01: The "already has a release" gate can no longer see a draft release, so a re-run signs, notarizes and creates a second draft (introduced by the job split)

**File:** `.github/workflows/release.yml:27-30, 47-63, 201-209`; `docs/releasing.md` ("The `build` job refuses ... or already has a release")
**Issue:** The gate now runs in `build`, whose token is `contents: read`. Before the split it ran in the single job with `contents: write`. GitHub returns draft releases only to callers with push access:
- `GET /releases/tags/{tag}` never returns a draft.
- When that fails, `gh release view` falls back to listing releases and matching the tag. With a read-only token, that list contains no drafts.

The only releases CI ever creates are drafts. So whenever a release already exists for the tag, `gh release view "$TAG"` now fails, and the gate passes. Two common ways to hit this:
1. **"Re-run failed jobs"** after the draft step succeeded but "Updater accepts the draft release" failed. Only `sign` re-runs (the gate is in `build`, which already succeeded). It approves, signs, notarizes and calls `gh release create "$TAG" --draft` again.
2. **"Re-run all jobs"**, or `gh run rerun`, of a tag run that already made its draft. The gate runs again but cannot see the draft.

GitHub accepts several drafts for one tag. The run then:
- leaves two drafts, the older being the one whose verification failed;
- burns a notarization round;
- breaks the documented promise that the build job "refuses, before any credential is touched, when [the tag] already has a release".

The maintainer can then publish the wrong draft from the Releases page.
**Fix:** Run the existing-release check where the token can see drafts, before any credential is touched: the first step of `sign`, before "Signing keychain and notary profile". Keep the published-release check in `build` if you like.
```yaml
  sign:
    steps:
      - name: No existing release (drafts included)
        if: github.event_name == 'push'
        env:
          GH_TOKEN: ${{ github.token }}   # contents: write sees drafts
        run: |
          n=$(gh api "repos/$REPO/releases" --paginate --jq "[.[] | select(.tag_name == env.TAG)] | length" | paste -sd+ - | bc)
          [ "$n" -eq 0 ] || { echo "::error::a release (or draft) for $TAG already exists; delete it before re-running"; exit 1; }
```
Do not use `gh release view` for this check: it stops at the first match, so it cannot report a count.

### WR-02: `release:app` refuses its own second run, because its build runs `go mod tidy` and the committed `go.sum` is not tidy (introduced by the clean-tree precondition)

**File:** `build/darwin/Taskfile.yml:263-266`, `build/darwin/Taskfile.yml:39-40` (`build:native` deps `common:go:mod:tidy`), `go.sum`
**Issue:** The new precondition `git diff --quiet HEAD --` refuses any tracked change. `go mod tidy -diff` on the committed tree reports two missing `go.sum` lines:
```
+github.com/stretchr/objx v0.5.2 h1:...
+github.com/stretchr/objx v0.5.2/go.mod h1:...
```
Every native build runs `go mod tidy` and adds them: `package`, `install:dev`, and `release:app` itself. The fix report says `go.sum` "was restored after each wails3 run", which hid this. So for the maintainer:
- After any `install:dev`, `wails3 task darwin:release` and `release:build` refuse with "Tracked files differ from HEAD".
- After one `wails3 task darwin:release` (for example, one that stops on the notarization 403 that is currently blocking), re-running it refuses for the same reason.

CI is not affected, because each job starts from a fresh checkout and the precondition runs before tidy. The local release path, which is how v1.0.0 shipped, breaks every time after the first build.
**Fix:** Commit the tidied `go.sum` (`go mod tidy`, then commit), so tidy is a no-op and the tree stays clean after a build. To stop it drifting again, add `go mod tidy -diff` to `ci.yml`, which fails when `go.mod`/`go.sum` are not tidy.

## Info

### Carried forward (unchanged; see `04-REVIEW.iter1.md` and `04-REVIEW.iter2.md`)

These 26 items are counted in `info`. Fix pass 2 did not change them:
- IN-B01 to IN-B09 and IN-F01 to IN-F11 (iteration 1)
- IN-01, IN-03, IN-04, IN-05, IN-06 and IN-07 (iteration 2)

IN-04 now has a second trigger: `ResetSettings` on a file that was fixed by hand loads that file without re-applying its theme, hotkeys or login item. The fix report acknowledges this.

### New in iteration 3

### IN-09: The login-item rollback undoes when the prior state is unknown, and it toggles the real login item while `settings.json` is unreadable

**File:** `internal/services/settings.go:511-540`
**Issue:**
- When `IsEnabled()` errors, `wasErr != nil` triggers `undo()`. If the item was already in the requested state, this flips it. For example, with `on=true` on an already-enabled item, the rollback disables a login item the user had on.
- While the store is unreadable, every toggle runs a real `SMAppService` register followed by an unregister, even though the write can never succeed. On macOS 13+ the register can post a "Login Item Added" notification for a change that is reported as failed.

**Fix:** Return `writeFailed(...)` before `apply()` when `s.Store.Unreadable()` is true. Skip the undo when `wasErr != nil`, and log instead.

### IN-10: `Revert` does not restore `failed`, so a reverted Bind drops the slot's H4b state

**File:** `internal/services/hotkeys.go:127-163`, `internal/services/settings.go:617-627`
**Issue:** A successful `Bind` clears `failed[slot]`, and `Revert` only calls `rebind`. Example: Open is refused at launch, so `failed.Open` is ⌃⌥⌘G. The user clicks Reset to Default while `settings.json` is unreadable. `Bind` registers the combo, `Update` returns `ErrUnreadable`, and `Revert("open", "")` unregisters it. `failed.Open` is now `""`, so `Snapshot.hotkeysFailed` no longer reports Open even though it is still unbound.

The Store check from WR-05 still prevents another slot from storing the combo, so this is a status display issue only.
**Fix:** Have `Revert` take the `failed` value from before the `Bind`, and restore it together with `bound`.

### IN-11: The second half of `17-hotkeys-open-unbound.js` passes even if the retry never runs

**File:** `scripts/uat/17-hotkeys-open-unbound.js:22-25`
**Issue:** The H4b text is already on screen from mount (`HotkeysPane.svelte:27-30`). After `reset.click()`, `getByText(...).waitFor()` can resolve before the click's `say(slot, '')` runs, and "Not set" is true either way. The script therefore does not show that clicking Reset sent `SetHotkey` or that the refusal was rendered again. The "enabled" assertion, which is the WR-06 fix, is real.
**Fix:** After the click, wait for the message to detach and then reappear. Alternatively, wait for the `SetHotkey` request with `B.waitForRequest(r => r.url().includes('/wails/') && r.postData()?.includes('SetHotkey'))` (or by its method ID).

### IN-12: The panel's `setLastProvider` rejection is unhandled on every tab switch while `settings.json` is unreadable

**File:** `frontend/src/App.svelte:480`
**Issue:** `setLastProvider(p)` is called without `await` or `.catch`. While the store is unreadable, each tab switch produces an unhandled promise rejection (`prefs_unreadable`). The panel has no banner, so the only place the cause appears is Settings › General. This is harmless to the user, but it adds console noise and would trip a `pageerror` listener in a panel UAT that runs on a seeded corrupt file.
**Fix:** `setLastProvider(p).catch(() => {})`, with a comment explaining that Settings › General reports `prefs_unreadable`.

---

_Reviewed: 2026-10-20T16:51:29Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
_Iteration: 3 (iteration 2 archived as 04-REVIEW.iter2.md)_
