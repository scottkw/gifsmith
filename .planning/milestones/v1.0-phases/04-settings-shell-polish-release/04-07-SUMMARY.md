---
phase: 04-settings-shell-polish-release
plan: 07
subsystem: updates
tags: [github-releases, update-check, codesign, spctl, notarization, wails, svelte]

requires:
  - phase: 04-settings-shell-polish-release
    provides: "settings.Store / SetPref / SettingsError (04-01), General pane sections and SettingRow (04-01), tray menu with key equivalents (04-03), platform.OpenURL, fakeapi + smoke/UAT loop (04-04)"
provides:
  - "internal/update: Checker.Latest (Accept + User-Agent Gifsmith + If-None-Match only; 304 cache; 404 = none published), ParseVersion/Compare, DMGAssetName, GitHubAllow, Download, NewDownloadClient, Verify, Clean, Runner/ExecRunner"
  - "services.Version = \"1.0.0\", UpdateStatus, UpdateService (bound Status/Check/Download/Reveal), package func RunAutoUpdates"
  - "SettingsError kinds update_check_failed, update_download_failed, update_verify_failed, update_open_failed"
  - "Prefs.UpdateCheckOff (zero = automatic checks on), SetPref updateCheck, Snapshot.updateCheck"
  - "shell.SetUpdate / RevealUpdates; tray item \"Update to Gifsmith {latest}…\" (hidden until found); events update:status and settings:reveal"
  - "fakeapi /github routes (FAKE_UPDATE current | tamper), server build canned verifier, smoke FAKE_UPDATE pass-through"
  - "frontend lib/update.ts (U2-U7, U-E1-U-E4, T12), Settings → General → Updates section, panel toast, UAT 20-update / 20-update-fail / 20-update-current"
  - "env-gated TestRealDMG / TestReleaseJSON (internal/update/real_darwin_test.go) for plan 04-06"
affects: [04-05, 04-06]

actuals:
  tokens: 26500
  tasks: 4
  commits: 8
plan_head_before: f0fad44aed88907c5f392b38a00a551b3c185481
plan_head_after: c5221c0d355d8008fc603912adea66d66ef14fc0

tech-stack:
  added: []
  patterns:
    - "Update logic Wails-free in internal/update; the bound service holds state and emits update:status on every change"
    - "Automatic loop is a package func (RunAutoUpdates) so Wails never binds it to JS"
    - "Verification tools run through an injectable update.Runner with absolute paths; the server build uses a canned runner"
    - "A pending-reveal flag exported from GeneralPane's module script lets SettingsApp reveal a section before the pane mounts"

key-files:
  created:
    - internal/update/update.go
    - internal/update/update_test.go
    - internal/update/install.go
    - internal/update/install_test.go
    - internal/update/real_darwin_test.go
    - internal/services/updates.go
    - internal/services/updates_test.go
    - frontend/src/lib/update.ts
    - frontend/src/lib/update.test.ts
    - frontend/bindings/gifsmith/internal/services/updateservice.ts
    - scripts/uat/20-update.js
    - scripts/uat/20-update-fail.js
    - scripts/uat/20-update-current.js
  modified:
    - internal/services/errors.go
    - internal/services/settings.go
    - internal/services/settings_test.go
    - internal/settings/settings.go
    - internal/settings/settings_test.go
    - internal/platform/platform.go
    - internal/provider/fakeapi/fakeapi.go
    - internal/shell/shell.go
    - main.go
    - wire_desktop.go
    - wire_server.go
    - scripts/smoke-browser-loop.sh
    - frontend/src/lib/api.ts
    - frontend/src/settings/GeneralPane.svelte
    - frontend/src/settings/SettingsApp.svelte
    - frontend/src/App.svelte
    - frontend/bindings/gifsmith/internal/services/models.ts
    - frontend/bindings/gifsmith/internal/services/index.ts

key-decisions:
  - "Checker caches the raw 200 body with its ETag and re-parses it on 304, so drafts, bad tags and not-installable releases answer the same way after a 304"
  - "A release with no valid DMG asset counts as up to date (logged `update: release v… not installable`); only transport/status failures are update_check_failed"
  - "GitHubAllow also refuses userinfo and any `..` in the path, beyond the exact-host rule"
  - "UpdateService.Download is allowed from available or opened; it runs on its own 10-minute background context, and a failed open keeps the verified file for a retry"
  - "OnAvailable runs after the service mutex is released (the tray setters wait on the main thread)"
  - "The server build's update dir is made absolute (filepath.Abs(\"bin/uat-cache/updates\")) so the file:// URL is well formed"

patterns-established:
  - "Update copy lives only in frontend/src/lib/update.ts (pure, node:test covered)"
  - "FAKE_UPDATE on a UAT script's Run via line selects the fixture GitHub mode"

requirements-completed: [RELS-05, RELS-06]

coverage:
  - id: D1
    description: "GitHub release check: one GET with Accept, User-Agent Gifsmith and If-None-Match only; 304 reuse; 404 = up to date; drafts, prereleases and non-semver tags never offered; exact DMG asset rules; stdlib compare against services.Version"
    requirement: RELS-05
    verification:
      - kind: unit
        ref: "internal/update/update_test.go#TestLatestFound,TestLatestNotFoundAndErrors,TestLatestNotOffered,TestLatestNotInstallable,TestGitHubAllow,TestCompare,TestParseVersion"
        status: pass
      - kind: unit
        ref: "internal/services/updates_test.go#TestUpdateCheckAvailable,TestUpdateCheckCurrent,TestUpdateCheckFailed,TestUpdateCheckSingleFlight"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/20-update.js; scripts/uat/20-update-current.js"
        status: pass
    human_judgment: false
  - id: D2
    description: "Download and Open…: asset-only HTTPS download with redirect allow-list, size cap and SHA-256 = API digest; codesign --verify --strict, TeamIdentifier=S2K7P43927, spctl accepted + Notarized Developer ID, quarantine; open only after Verify; any failure deletes the file and opens nothing"
    requirement: RELS-06
    verification:
      - kind: unit
        ref: "internal/update/install_test.go#TestDownload*,TestVerify,TestVerifyFailures,TestClean"
        status: pass
      - kind: unit
        ref: "internal/services/updates_test.go#TestUpdateDownloadOpens,TestUpdateDownloadFailures,TestUpdateDownloadNotAvailable"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/20-update.js (progress, verifying, opened); scripts/uat/20-update-fail.js (tamper -> U-E3)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Automatic checks (about 10 s after launch, every 24 h, off via `Check for updates automatically` persisted as updateCheckOff); tray item and one-time panel toast that reveal Settings → General → Updates"
    requirement: RELS-05
    verification:
      - kind: unit
        ref: "internal/services/updates_test.go#TestRunAutoUpdates,TestRunAutoUpdatesOff,TestRunAutoUpdatesFailureIsQuiet,TestUpdateOnAvailable,TestUpdateReveal; internal/services/settings_test.go#TestSetPrefUpdateCheckPref; internal/settings/settings_test.go#TestStoreUpdateCheckOff"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/20-update.js (toast + View reveal + checkbox persistence)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Native: the signed install's launch check and Check for Updates… reach api.github.com; the tray item renders; the real notarized DMG passes the verifier and opens in Finder; Updates section in Light and Dark"
    requirement: RELS-06
    verification: []
    human_judgment: true
    rationale: "Only the signed desktop build talks to the real endpoint and owns the real tray; opening a disk image shows a Finder window on the user's desktop, which executors never do; no release newer than 1.0.0 exists yet"

duration: 19min
completed: 2026-10-20
status: complete
---

# Phase 4 Plan 07: GitHub Update Check Summary

**Settings → General → Updates asks GitHub Releases for scottkw/gifsmith (one plain GET with an ETag), alerts through the menu bar and a one-time panel toast, and on request downloads the DMG, checks its SHA-256 digest, Developer ID signature, Team ID and notarization, quarantines it and opens it for the usual drag to Applications. The app never replaces itself.**

## Performance

- **Duration:** about 19 min
- **Started:** 2026-10-20T02:26:05Z
- **Completed:** 2026-10-20T02:45Z
- **Tasks:** 4 of 4
- **Files modified:** 31 (2293 insertions)

## Accomplishments

- `internal/update` (Wails-free, stdlib only) has the check, the download and the verifier. The only request is `GET {api}/repos/scottkw/gifsmith/releases/latest` with `Accept`, `User-Agent: Gifsmith` and, after the first 200, `If-None-Match`. A test asserts that header set.
- `UpdateService` (bound: Status, Check, Download, Reveal) allows one check or download at a time and emits `update:status` on every change. `Download` calls `Open` only after `update.Verify` returns nil. Any failed check deletes the file and leaves the state at available.
- Automatic checks run about 10 s after launch and then every 24 h while `Check for updates automatically` is on. The setting is stored as `updateCheckOff`, where the zero value means on. The desktop build runs these checks; the UAT build never does.
- Alerts: a hidden tray item becomes `Update to Gifsmith {latest}…`, and the panel footer shows the T12 toast with `View` once per version per launch. Both open Settings at General with Updates in view.
- The headless UAT covers available → download progress → verifying → opened, the toast and View reveal, the checkbox persisting across reloads, a tampered download (U-E3) and the up-to-date case. All of it runs against the fixture; nothing contacts GitHub.

## Task Commits

1. **Task 1 (tracer): Check for Updates… end to end**: `3b93000` (test, RED), `55737a1` (feat). Tracer gate: the full `<verify>` was re-run before expansion and passed (gofmt, vet, race tests, server build, go.sum clean, node tests, svelte-check, SMOKE OK, 20-update UAT OK, 15-settings-general UAT OK).
2. **Task 2: Download and Open… with verification**: `aaea7ef` (test, RED), `f7b767b` (feat)
3. **Task 3: automatic checks, updateCheck pref, tray item, Reveal**: `cec8ca2` (test, RED), `275bb1f` (feat)
4. **Task 4: copy module, checkbox, reveal, toast, failure UATs, env-gated real tests**: `7db2f7f` (test, RED), `c5221c0` (feat)

### TDD notes

- Go RED commits ran the new tests against stub skeletons. Each target test failed on its behaviour assertion; for example, `TestLatestFound` failed with `Latest = false, <nil>` and `TestUpdateDownloadOpens` with an empty status and no OpenURL. None failed on compilation. One exception: the `TestStoreUpdateCheckOff` round trip needs the `Prefs.UpdateCheckOff` field just to compile, so that field went in with the RED skeleton, and the test serves as a regression guard. Its partner `TestSetPrefUpdateCheckPref` failed RED as planned.
- Frontend RED (`update.test.ts`): `gsd-tools check tdd-red-evidence` returned `RED_EVIDENCE_OK`. The target test `update status lines (U2-U7)` failed on its assertion (the stub returned `""`), with 3 of 3 failing in TAP. Semantic assessment: the failures were the intended ones.

## Verification (plan level)

- `gofmt -l` reported nothing. `go vet . ./internal/...` passed, `go test -race ./internal/... -count=1` passed in every package, and `golangci-lint run . ./internal/...` reported 0 issues.
- `TestRealDMG` and `TestReleaseJSON` print `--- SKIP` when their variables are unset.
- `CGO_ENABLED=0 go build -tags server .` passed. `git diff --quiet -- go.mod go.sum frontend/package.json` passed, so no dependency was added.
- `node --test frontend/src/lib/update.test.ts`: 3 pass. `pnpm run test`: 101 pass, 0 fail. svelte-check: 0 errors. vite build passed.
- `smoke-browser-loop.sh` printed SMOKE OK. `--uat` printed UAT OK for 20-update.js, 20-update-fail.js, 20-update-current.js, 15-settings-general.js and 01-tracer.js.

## Decisions Made

See `key-decisions` in the frontmatter. All of them stay within D-26..D-35. None changes the UI-SPEC.

## Deviations from Plan

**1. [Rule 2 - Security] GitHubAllow also refuses userinfo URLs and `..` path segments**
- **Found during:** Task 1
- **Issue:** A prefix check on `/scottkw/gifsmith/releases/download/` alone would pass `…/download/../../other/repo/x`, and `https://user@github.com/…` would also pass.
- **Fix:** Added two guards to `GitHubAllow`, both covered in `TestGitHubAllow`.
- **Files modified:** internal/update/update.go
- **Commit:** 55737a1

**2. [Rule 1 - Bug] The server build's update dir is absolute**
- **Found during:** Task 2
- **Issue:** `url.URL{Scheme: "file", Path: "bin/uat-cache/updates/…"}` turns `bin` into the URL host.
- **Fix:** `wire_server.go` resolves the dir with `filepath.Abs("bin/uat-cache/updates")`.
- **Files modified:** wire_server.go
- **Commit:** f7b767b

**3. [Rule 1 - Bug] OnAvailable no longer runs under the service mutex**
- **Found during:** Task 3
- **Issue:** The first GREEN placed a `defer` after `defer u.mu.Unlock()`. Defers run last-in, first-out, so the tray setter would have run while the lock was still held.
- **Fix:** A deferred closure registered before the lock now runs after the unlock.
- **Files modified:** internal/services/updates.go
- **Commit:** 275bb1f

**Total deviations:** 3 auto-fixed (1 security, 2 bugs). **Impact:** none on scope or the UI contract.

## Issues Encountered

- One early RED run of the services tests hung: `TestUpdateCheckSingleFlight` waited on a server the stub never called. It ended at Go's 10-minute test timeout, since killing the process was not permitted. The test now has a 5 s guard, and later runs use `-timeout`.

## Known Stubs

None. The server build's canned verifier (`cannedVerify` in wire_server.go) is the D-35 test seam. It never ships, because wire_server.go is `//go:build server`.

## Threat Flags

None beyond the plan's threat model (T-04-29..T-04-37). Each listed mitigation is implemented and has a unit test.

## Human-check items (deferred to the end of the phase)

1. On the signed install (plan 04-06): about 10 s after launch, run `grep 'update:' bin/install-dev.log`. Expect `update: no published release` while v1.0.0 is unpublished. Check for Updates… should then show `Gifsmith 1.0.0 is up to date.` Turn the box off, relaunch, and confirm no automatic check is logged. No `Update to Gifsmith…` tray item should appear.
2. After 04-06 Task 1 has produced the notarized DMG, run `UPDATE_REAL_DMG=bin/release/Gifsmith-1.0.0-arm64.dmg UPDATE_REAL_OPEN=1 go test ./internal/update -run TestRealDMG -count=1 -v`. Expect the test to pass and Finder to show the drag window; eject the image afterwards. 04-06 also runs `TestReleaseJSON` against the draft JSON and the DMG.
3. Check the Updates section on the signed install in Light and Dark.
4. When 1.0.1 ships: run Check for Updates… → Download and Open… on 1.0.0 against the real endpoint, and confirm the tray item renders.

## Notes for later plans

- 04-05's version_test should assert `services.Version` equals Info.plist `CFBundleShortVersionString` and `update.DMGAssetName(v)` equals `DMG_NAME`.
- If a newer release is withdrawn mid-launch, a later check that finds nothing newer does not hide the tray item again. The tests require that `OnAvailable` is called only for a newer release, so this is left as is.

## Self-Check: PASSED
