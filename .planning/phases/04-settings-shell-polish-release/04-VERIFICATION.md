---
phase: 04-settings-shell-polish-release
verified: 2026-10-20T17:04:43Z
status: human_needed
score: 4/7 roadmap success criteria verified by automated evidence (SC1, SC3, SC5, SC6); 3 await human/operational confirmation (SC2, SC4, SC7); 0 failed. 17/17 requirement IDs covered by code; all automatable plan truths verified
covered_files:
  - ".github/workflows/ci.yml"
  - ".github/workflows/release.yml"
  - ".planning/phases/04-settings-shell-polish-release/04-01-PLAN.md"
  - ".planning/phases/04-settings-shell-polish-release/04-01-SUMMARY.md"
  - ".planning/phases/04-settings-shell-polish-release/04-02-PLAN.md"
  - ".planning/phases/04-settings-shell-polish-release/04-02-SUMMARY.md"
  - ".planning/phases/04-settings-shell-polish-release/04-03-PLAN.md"
  - ".planning/phases/04-settings-shell-polish-release/04-03-SUMMARY.md"
  - ".planning/phases/04-settings-shell-polish-release/04-04-PLAN.md"
  - ".planning/phases/04-settings-shell-polish-release/04-04-SUMMARY.md"
  - ".planning/phases/04-settings-shell-polish-release/04-05-PLAN.md"
  - ".planning/phases/04-settings-shell-polish-release/04-05-SUMMARY.md"
  - ".planning/phases/04-settings-shell-polish-release/04-06-PLAN.md"
  - ".planning/phases/04-settings-shell-polish-release/04-06-SUMMARY.md"
  - ".planning/phases/04-settings-shell-polish-release/04-07-PLAN.md"
  - ".planning/phases/04-settings-shell-polish-release/04-07-SUMMARY.md"
  - ".planning/phases/04-settings-shell-polish-release/04-08-PLAN.md"
  - ".planning/phases/04-settings-shell-polish-release/04-08-SUMMARY.md"
  - "README.md"
  - "build/config.yml"
  - "build/darwin/Info.dev.plist"
  - "build/darwin/Info.plist"
  - "build/darwin/Taskfile.yml"
  - "docs/release-notes-1.0.0.md"
  - "docs/releasing.md"
  - "frontend/index.html"
  - "frontend/settings.html"
  - "frontend/src/App.svelte"
  - "frontend/src/app.css"
  - "frontend/src/components/ActionMenu.svelte"
  - "frontend/src/components/DetailsPane.svelte"
  - "frontend/src/components/Footer.svelte"
  - "frontend/src/components/LibraryGrid.svelte"
  - "frontend/src/components/ResultGrid.svelte"
  - "frontend/src/components/SetupCard.svelte"
  - "frontend/src/components/Tile.svelte"
  - "frontend/src/lib/actions.test.ts"
  - "frontend/src/lib/actions.ts"
  - "frontend/src/lib/api.ts"
  - "frontend/src/lib/errors.test.ts"
  - "frontend/src/lib/errors.ts"
  - "frontend/src/lib/grid.test.ts"
  - "frontend/src/lib/grid.ts"
  - "frontend/src/lib/hotkeys.test.ts"
  - "frontend/src/lib/hotkeys.ts"
  - "frontend/src/lib/library.test.ts"
  - "frontend/src/lib/library.ts"
  - "frontend/src/lib/motion.test.ts"
  - "frontend/src/lib/motion.ts"
  - "frontend/src/lib/popover.test.ts"
  - "frontend/src/lib/popover.ts"
  - "frontend/src/lib/prefs.test.ts"
  - "frontend/src/lib/prefs.ts"
  - "frontend/src/lib/providers.ts"
  - "frontend/src/lib/theme.test.ts"
  - "frontend/src/lib/theme.ts"
  - "frontend/src/lib/update.test.ts"
  - "frontend/src/lib/update.ts"
  - "frontend/src/main.ts"
  - "frontend/src/settings/AboutPane.svelte"
  - "frontend/src/settings/DownloadsPane.svelte"
  - "frontend/src/settings/GeneralPane.svelte"
  - "frontend/src/settings/HotkeyRecorder.svelte"
  - "frontend/src/settings/HotkeysPane.svelte"
  - "frontend/src/settings/KeyRow.svelte"
  - "frontend/src/settings/KeysPane.svelte"
  - "frontend/src/settings/SearchPane.svelte"
  - "frontend/src/settings/SettingRow.svelte"
  - "frontend/src/settings/SettingsApp.svelte"
  - "frontend/src/settings/main.ts"
  - "frontend/vite.config.ts"
  - "go.mod"
  - "go.sum"
  - "internal/platform/fake.go"
  - "internal/platform/platform.go"
  - "internal/platform/platform_darwin.go"
  - "internal/platform/platform_darwin.h"
  - "internal/platform/platform_darwin.m"
  - "internal/platform/platform_darwin_test.go"
  - "internal/platform/platform_stub.go"
  - "internal/provider/fakeapi/fakeapi.go"
  - "internal/provider/finergifs/finergifs.go"
  - "internal/provider/finergifs/finergifs_test.go"
  - "internal/provider/giphy/giphy.go"
  - "internal/provider/giphy/giphy_test.go"
  - "internal/provider/giphy/languages.go"
  - "internal/provider/klipy/countries.go"
  - "internal/provider/klipy/klipy.go"
  - "internal/provider/klipy/klipy_test.go"
  - "internal/provider/provider.go"
  - "internal/provider/provider_test.go"
  - "internal/search/search.go"
  - "internal/search/search_test.go"
  - "internal/services/actions.go"
  - "internal/services/actions_test.go"
  - "internal/services/errors.go"
  - "internal/services/fsprobe.go"
  - "internal/services/fsprobe_darwin.go"
  - "internal/services/fsprobe_other.go"
  - "internal/services/fsprobe_test.go"
  - "internal/services/gifs.go"
  - "internal/services/hotkeys.go"
  - "internal/services/hotkeys_test.go"
  - "internal/services/panel.go"
  - "internal/services/settings.go"
  - "internal/services/settings_test.go"
  - "internal/services/updates.go"
  - "internal/services/updates_test.go"
  - "internal/services/version_test.go"
  - "internal/settings/settings.go"
  - "internal/settings/settings_test.go"
  - "internal/shell/position.go"
  - "internal/shell/position_test.go"
  - "internal/shell/shell.go"
  - "internal/update/install.go"
  - "internal/update/install_test.go"
  - "internal/update/real_darwin_test.go"
  - "internal/update/update.go"
  - "internal/update/update_test.go"
  - "main.go"
  - "scripts/dmg-background.swift"
  - "scripts/notarize.sh"
  - "scripts/release-stamp.sh"
  - "scripts/release-version.sh"
  - "scripts/smoke-browser-loop.sh"
  - "scripts/stapler-retry.sh"
  - "scripts/uat/11-square.js"
  - "scripts/uat/15-settings-general.js"
  - "scripts/uat/15-settings-startup-fail.js"
  - "scripts/uat/15-settings-unreadable.js"
  - "scripts/uat/16-settings-keys.js"
  - "scripts/uat/16-settings-search.js"
  - "scripts/uat/17-hotkeys-fail.js"
  - "scripts/uat/17-hotkeys-open-unbound.js"
  - "scripts/uat/17-hotkeys.js"
  - "scripts/uat/17-polish.js"
  - "scripts/uat/18-backstops.js"
  - "scripts/uat/18-default-actions.js"
  - "scripts/uat/18-downloads.js"
  - "scripts/uat/18-grid-size.js"
  - "scripts/uat/19-about.js"
  - "scripts/uat/20-update-current.js"
  - "scripts/uat/20-update-fail.js"
  - "scripts/uat/20-update.js"
  - "scripts/verify-release.sh"
  - "wire_desktop.go"
  - "wire_server.go"
covered_digest: "v3:sha256:d91d00fcdd63278dfb992714f3530e67f60ce6868d99e50ab42c39cde106b5d4"
behavior_unverified: 0
overrides_applied: 0
human_verification:
  - test: "Rebuild, notarize and verify 1.0.0 from the current HEAD (wails3 task darwin:release), replace the v1.0.0 draft's DMG and .sha256 assets, then re-run TestRealDMG / TestReleaseJSON against the new draft and repeat the clean-VM install test from the quarantined DMG"
    expected: "RELEASE VERIFY OK (notarized); spctl source=Notarized Developer ID for app and DMG; draft asset digest equals the new local SHA-256; clean VM: Gatekeeper's standard 'downloaded from the Internet' prompt only, menu bar app with no Dock icon, fresh permissions, copy/paste works, About shows 1.0.0 credits"
    why_human: "The draft DMG (sha256 a3129b96…95d4) was built at 8428b30, before the 3 review-fix passes (00eac67..7e8792a). It verifies as notarized today but does not contain the shipped code. Notarization uses the user's credentials and the clean-machine install needs a person (RELS-01, RELS-02, SC4)"
  - test: "After merging the phase branch to main, run a workflow_dispatch dry run of release.yml (build job + sign job in the release environment)"
    expected: "Both jobs green: build uploads the app artifact; sign imports the .p12, notarizes, staples, prints 'RELEASE VERIFY OK (notarized)' and '--- PASS: TestRealDMG'; no release created; temporary keychain deleted"
    why_human: "The last green dry run (39446861711) predates the build/sign job split (7abb1b0) and the drafts-included release gate (f196276). SC7 requires a green dry run before the first tag (v1.0.1). Triggering a workflow is the user's call; agents must not run gh workflow run"
  - test: "Settings window (checklist A1-A6): tray Settings… and ⌘, from the panel; ⌘V paste into a Settings text field; ⌘W / red button hide; Esc does not close; theme Dark/Light/System with no white flash and live System follow; Open at login appears/disappears in System Settings > Login Items and the checkbox reflects an external change; Grant Access… prompt and '✓ On' within ~2 s"
    expected: "Settings activates the accessory app and takes typing and paste; theme flips native chrome and both webviews; login item tracks the toggle"
    why_human: "WKWebView activation, NSApp appearance, SMAppService and TCC prompts are native and cannot be driven headlessly (agents may not send desktop input)"
  - test: "Keys & Search (checklist B7-B9): paste a real GIPHY key via Change Key…, Save, Test; Content rating G then R; GIFs per page 48"
    expected: "'Saved · ends in XXXX'; Test reports accepted; trending changes with rating after reopening the panel; 48-item first page and paging continues"
    why_human: "Real Keychain and real vendor keys; UAT uses MemoryStore and fakeapi only"
  - test: "Hotkeys (checklist C10-C14): record a new Open shortcut and fire it from another app; set Favorites ⌃⌥⌘F / Recents ⌃⌥⌘R; recording captures the current combo without firing; click away while recording; clear Open and use Reset to Default; Test reports H7 for a combo another app owns"
    expected: "Rebound Carbon hotkey works system-wide; tray shows the current shortcut; Favorites/Recents open (and toggle) their tabs; hotkeys are never left suspended; Reset restores ⌃⌥⌘G"
    why_human: "Carbon RegisterEventHotKey behaviour across apps is only observable natively (SC2 is marked manual in the roadmap)"
  - test: "Downloads & grid (checklist D15-D18): Choose… folder sheet, Desktop TCC prompt, exFAT stick / network share refusal (D5/D6); Hide original file names; Search Large / Trending Small; right-click opens the action menu; ⌘S on a Klipy GIF"
    expected: "Folder saved only when valid; gifsmith-{id}.gif then gifsmith-{id} (2).gif; 3 vs 5 columns; T11 'Klipy GIFs can't be downloaded yet.'"
    why_human: "NSOpenPanel, TCC folder prompts and real volumes are native"
  - test: "Motion, About, Updates (checklist E19-E21): macOS Reduce Motion on; About in Light and Dark with each link; Updates section look"
    expected: "Only the selected tile animates; links open the right pages; GIPHY mark follows the theme"
    why_human: "Native accessibility setting and visual appearance"
  - test: "Notarized paste matrix carry-over from Phase 1 (01-PASTE-MATRIX.md): fill the notarized Copy→⌘V and ⌘Enter columns (and Slack, Discord, Telegram, Safari/Chrome composer rows) on the rebuilt notarized install with Accessibility granted"
    expected: "Every row has a PASS/FAIL result; 21 cells are still PENDING today"
    why_human: "Real third-party apps and Accessibility-granted ⌘V synthesis; ROADMAP Phase 4 lists 'paste-matrix N15 plus the notarized columns' as a Phase 1 carry-over (N15 is now PASS)"
---

# Phase 4: Settings, Shell Polish & Release Verification Report

**Phase Goal:** Everything is configurable from a proper Settings window with hotkeys that can be rebound and launch at login, and a stranger can install a signed, notarized build on a clean Mac and use it within vendor and license obligations, and it tells the user when a newer release is available on GitHub, offering a verified download of the new DMG.
**Verified:** 2026-10-20T17:04:43Z (HEAD 7e8792a, branch gsd/phase-04-settings-shell-polish-release)
**Status:** human_needed
**Re-verification:** No. This is the initial verification.

## Goal Achievement

### Observable Truths (ROADMAP success criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Settings from the menu bar opens a separate Settings window where paste works; add/test/change/remove GIPHY and Klipy keys with sign-up links; Accessibility status with a grant button | ✓ VERIFIED (headless). The native paste part is listed for human checks | `internal/shell/shell.go:76-92` creates a second window "Gifsmith Settings" 640×520, hidden, with ⌘W and close hooks that hide it. Tray `Settings…` → `OpenSettings` (Show+Focus). The app menu keeps the Edit role (`shell.go:98-100`), so ⌘V reaches fields. `KeyRow.svelte` has Add/Change Key…, Test, Remove Key and Get a free key ↗ (GIPHY_KEY_URL / KLIPY_KEY_URL in `lib/providers.ts:15-16`). `GeneralPane.svelte` has the Automatic paste row and Grant Access…. Tests: `TestKeyStatus`, `TestTestKeyUncached`, `TestRemoveKey`, `TestAccessibilityStatus`. UAT: 15-settings-general, 16-settings-keys OK |
| 2 | Rebindable hotkeys with conflict report and rollback; Favorites/Recents hotkeys; launch at login toggle | ? UNCERTAIN (manual SC). The code is verified | `internal/services/hotkeys.go` handles Bind with rollback and canonical duplicates. Tests: `TestBindRollback`, `TestBindDuplicate`, `TestBindRestoreFailsLogs`, `TestSetHotkeyWriteFailureRollsBack` and `TestStartFailureIsNotBound` pass under -race. `main.go:160-164` maps open/favorites/recents to Toggle/ShowTab. `shell.go:172` emits `panel:open-tab`, and `App.svelte:305` listens. `settingsSvc.Autostart = app.Autostart` (`main.go:144-147`). Tests: `TestSetLaunchAtLogin*`, `TestApplyStartup`. UAT: 17-hotkeys*, 15-settings-startup-fail OK. Firing Carbon hotkeys across apps and the real login item need the user |
| 3 | Grid sizes, max results, download folder, hide filename, locales, rating (PG-13 default), ↵/⌘↵ actions, theme take effect | ✓ VERIFIED | SearchPane rows: Search results, Trending, GIFs per page, Content rating, Language, Country. DownloadsPane: Save to, File names. GeneralPane: Theme, Primary/Secondary action. Data flow: giphy `v.Set("rating", prm.Rating)`; klipy `per_page` clamp and `locale`; `actions.DownloadsDir = settingsSvc.DownloadDir`; `HideFilenames` → `Staging.Save`; `colsFor(gridSize)` and `effectiveAction` in App.svelte. `DefaultParams` Rating "pg-13". Tests: `TestSetPref*`, `TestParamsDefaultRating`, `TestChooseDownloadDir*`, `TestValidateDownloadDir*`. UAT: 16-settings-search, 18-default-actions, 18-downloads, 18-grid-size OK |
| 4 | Developer ID–signed, notarized DMG installs and runs on a clean account (Gatekeeper, fresh permissions), macOS 13+ Apple Silicon | ? UNCERTAIN (manual SC; the artifact is stale) | I re-ran `scripts/verify-release.sh --notarized bin/release` and it printed `RELEASE VERIFY OK (notarized)`: spctl `source=Notarized Developer ID` for app and DMG, stapler validate OK, and the DR matches the dev build. `TestRealDMG` and `TestReleaseJSON` PASS against the draft's JSON (digest a3129b96…95d4). The user's clean-VM test (macOS 15.7.7) passed per 04-06-SUMMARY. **However**, that DMG was built at 8428b30, before the review fixes, so a rebuild, re-notarize and clean install are pending (known operational item) |
| 5 | About credits the MIT Raycast extension and authors; Klipy file actions disabled unless clearance recorded | ✓ VERIFIED | `AboutPane.svelte:41-43` credits Joseph Schmitt and contributors under MIT © 2021 Raycast, and `TestLicenseCredits` keeps it in step with LICENSE. `actions.go:45` has `const KlipyFileActionsCleared = false`. `gated()` is called in Copy, Paste, Download, CopySquare and PasteSquare (`actions.go:114,135,421,476,655`). The klipy_gated test is at `actions_test.go:1351`. `main.go:100-101` wires the const. UAT 19-about and 18-default-actions OK |
| 6 | Checks GitHub Releases, alerts, "Check for Updates…"; download → checksum → Developer ID signature → notarization → open; no self-replacement | ✓ VERIFIED | `updates.go:152-185`: Download (digest via `update.Download`), then `update.Verify` (codesign --verify --strict, TeamIdentifier=S2K7P43927, spctl `source=Notarized Developer ID`, quarantine), then `Open` (platform OpenURL) only when Verify returns nil. A failure deletes the file. `RunAutoUpdates` starts 10 s after launch and repeats every 24 h. The tray `Update to Gifsmith…` item is hidden until available. Tests: `TestUpdateDownloadFailures`, `TestDownloadDigestMismatch`, `TestVerifyFailures`, `TestRealDMG` (the real DMG passes and a byte-flipped copy is rejected). UAT 20-update, 20-update-current, 20-update-fail OK |
| 7 | CI on push/PR without secrets; vX.Y.Z tag → signed, notarized, verified draft release; dry run green before first tag | ? UNCERTAIN (operational) | `ci.yml` runs on push to main and on PRs to main, with `contents: read`, no `secrets.*`, and all checks including `go mod tidy -diff`. `release.yml` triggers on tags `v[0-9]+.[0-9]+.[0-9]+` and `!v1.0.0` plus workflow_dispatch, with no pull_request trigger. It has a build job (no environment) and a sign job (`environment: release`), and gates on `release-version.sh`, `RELEASE VERIFY OK (notarized)`, `TestRealDMG`, the drafts-included existing-release check and `TestReleaseJSON`. Every `uses:` is a SHA-pinned actions/*. Actions: CI 39446862985 green and Release dry run 39446861711 green at f6e3e40. **The current split workflow has not run on Actions**, because the review fixes are not on main |

**Score:** 4/7 roadmap SCs verified by automated evidence. 3 are UNCERTAIN (manual or operational), none FAILED, and 0 are present-but-behavior-unverified.

The plan-level must_haves (8 plans) were checked through tests and UAT. Every automatable truth is backed by a passing Go/node test or headless UAT. Every `verification: backstop` truth is native and is routed to Human Verification below.

### Required Artifacts

`verify.artifacts` passed every plan: 04-01 6/6, 04-02 6/6, 04-03 5/5, 04-04 5/5, 04-05 6/6, 04-06 2/2, 04-07 5/5, 04-08 4/4.

| Artifact | Status | Details |
|----------|--------|---------|
| internal/settings/settings.go (Store.Update) | ✓ VERIFIED | Single prefs owner. `TestSetPrefConcurrentWithLastProvider` passes |
| internal/services/settings.go | ✓ VERIFIED | SetPref, SetLaunchAtLogin, TestKey, RemoveKey, ChooseDownloadDir |
| internal/services/hotkeys.go | ✓ VERIFIED | Bind with rollback, Suspend/Resume, ArmTest |
| internal/services/fsprobe.go | ✓ VERIFIED | ValidateDownloadDir (FAT, network, unwritable tests) |
| internal/services/actions.go | ✓ VERIFIED | KlipyFileActionsCleared const gate |
| internal/update/update.go, install.go; internal/services/updates.go | ✓ VERIFIED | Checker.Latest, Download, Verify, Clean; `const Version = "1.0.0"` equals Info.plist (`TestVersionMatchesInfoPlist`) |
| frontend/src/settings/*.svelte (6 panes + recorder) | ✓ VERIFIED | Wired through lib/api.ts; svelte-check reports 0 errors |
| build/darwin/Taskfile.yml, scripts/verify-release.sh, notarize.sh, release-version.sh | ✓ VERIFIED | release = release:app + release:sign → assemble → notarize.sh → verify-release.sh |
| .github/workflows/ci.yml, release.yml, docs/releasing.md | ✓ VERIFIED (static) | Untested on Actions since the split (see SC7) |
| README.md, docs/release-notes-1.0.0.md | ✓ VERIFIED | Install, Verify, Get API keys, Settings, Updating, Attribution, Known limits, License sections |

### Key Link Verification

| From | To | Via | Status |
|------|----|-----|--------|
| GeneralPane | settings.go | setPref | ✓ WIRED |
| settings.go | settings.Store | Store.Update + settings:changed | ✓ WIRED |
| shell.go | platform | ApplyTheme → SetAppearance + SetBackgroundColour | ✓ WIRED |
| main.go | provider clients | `Params: params` (main.go:73-95) | ✓ WIRED |
| giphy.go | provider.Params | `provider.ParamsOf(c.Params)` (giphy.go:101,174) | ✓ WIRED. The tool's `params\(\)` regex missed it because the code calls ParamsOf |
| settings.go | search.Flush; gifs.go → MarkRejected | Flush on param change | ✓ WIRED |
| main.go | Hotkeys | `services.Hotkeys{Reg: app.GlobalShortcut …}` | ✓ WIRED |
| shell.go | Hotkeys.Resume | `sh.OnSettingsInactive = hk.Resume` | ✓ WIRED |
| App.svelte | prefs.effectiveAction | ↵ / ⌘↵ | ✓ WIRED |
| updates.go | update.Verify → OpenURL | `Open: plat.OpenURL` | ✓ WIRED |
| updates.go | shell | OnAvailable = SetUpdate, RevealFunc = RevealUpdates | ✓ WIRED |
| Taskfile release | verify-release.sh --notarized | Taskfile.yml:314 | ✓ WIRED. The tool could not resolve the "Taskfile.yml release" path, so I checked it by hand |
| bin/release DMG | GitHub draft v1.0.0 | gh release | ✓ EXISTS: a draft "Gifsmith 1.0.0", target main, DMG plus .sha256, state uploaded, digest equal to the local file |
| release.yml | release-version.sh, darwin:release:app/sign, TestReleaseJSON | | ✓ WIRED |

### Data-Flow Trace (Level 4)

| Setting | Source | Consumer | Status |
|---------|--------|----------|--------|
| Content rating / language / page size / country | settings.Store → SettingsService.Params | giphy/klipy/finergifs request params per call | ✓ FLOWING |
| Download folder / hide filenames | Store → DownloadDir() / HideFilenames() | ActionService.Download → Staging.Save (read at call time) | ✓ FLOWING |
| Grid size / default actions / theme | Store → Snapshot → `settings:changed` | App.svelte colsFor / effectiveAction / applyTheme; shell.ApplyTheme | ✓ FLOWING |
| Update status | GitHub releases/latest (fixture in UAT) | GeneralPane Updates row, tray item, panel toast | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Go formatting, vet, module tidiness | `gofmt -l …`, `go vet . ./internal/...`, `go mod tidy -diff` | clean / OK / OK | ✓ PASS |
| Go tests (race) | `go test -race -count=1 ./internal/...` | all 15 packages ok | ✓ PASS |
| Lint | `golangci-lint run . ./internal/...` (v2.14.0) | 0 issues | ✓ PASS |
| Server build | `CGO_ENABLED=0 go build -tags server .` | OK | ✓ PASS |
| Frontend tests / types | `pnpm test`, `pnpm check` | 103/103; 0 errors, 0 warnings (191 files) | ✓ PASS |
| Smoke + all headless UAT | `scripts/smoke-browser-loop.sh` and `--uat` for each of the 35 scripts in scripts/uat | SMOKE OK; 35/35 UAT OK (18-backstops did not flake this run) | ✓ PASS |
| Notarized artifact | `scripts/verify-release.sh --notarized bin/release` | RELEASE VERIFY OK (notarized) | ✓ PASS (stale build) |
| Updater accepts shipped artefacts | `UPDATE_REAL_DMG=… UPDATE_RELEASE_JSON=… go test ./internal/update -run 'TestRealDMG\|TestReleaseJSON'` | both PASS | ✓ PASS |

### Probe Execution

No `scripts/*/tests/probe-*.sh` exist and no plan declares one. Step 7c is not applicable. The release verification script was run directly (see above).

### Requirements Coverage

| Requirement | Source Plan | Status | Evidence |
|-------------|-------------|--------|----------|
| SETT-01 | 04-01 | ✓ SATISFIED (native paste: human) | Settings window, Edit menu role, UAT 15 |
| SETT-02 | 04-02 | ✓ SATISFIED | Key cards, KeyStatus last-4 only, TestKey/RemoveKey tests, UAT 16-settings-keys |
| SETT-03 | 04-03 | ✓ SATISFIED (native firing: human) | Bind rollback tests, UAT 17-hotkeys* |
| SETT-04 | 04-02, 04-04 | ✓ SATISFIED | Search/Downloads panes, fsprobe, grid sizes |
| SETT-05 | 04-01 | ✓ SATISFIED (TCC prompt: human) | AccessibilityStatus + Grant Access… |
| SETT-06 | 04-01 | ✓ SATISFIED (native appearance: human) | theme.ts, ApplyTheme, TestSetPrefTheme |
| SRCH-09 | 04-02 | ✓ SATISFIED | rating G/PG/PG-13/R, default pg-13, sent per request |
| ACTN-08 | 04-04 | ✓ SATISFIED | effectiveAction, swap rule, UAT 18-default-actions |
| SHEL-05 | 04-01 | ✓ SATISFIED (login item: human) | Autostart wiring and tests |
| SHEL-06 | 04-03 | ✓ SATISFIED (native: human) | favorites/recents → ShowTab → panel:open-tab |
| RELS-01 | 04-05, 04-06 | ? NEEDS HUMAN | Pipeline verified; the existing notarized DMG predates the review fixes, so a rebuild is pending |
| RELS-02 | 04-06 | ? NEEDS HUMAN | Clean-VM pass recorded for the stale DMG; repeat on the rebuild |
| RELS-03 | 04-05 | ✓ SATISFIED | About credits, TestLicenseCredits |
| RELS-04 | 04-04, 04-05 | ✓ SATISFIED | const gate + UI omission; README known limits |
| RELS-05 | 04-07 | ✓ SATISFIED | Checker.Latest, auto check, tray/toast alert, Check for Updates… |
| RELS-06 | 04-07 | ✓ SATISFIED | digest → codesign → TeamID → spctl → open; TestRealDMG |
| RELS-07 | 04-08 | ? NEEDS HUMAN | Workflows present and statically correct; dry run of the split workflow pending |

All 17 phase requirement IDs appear in plan frontmatter. REQUIREMENTS.md maps no other ID to Phase 4, so there are no orphans. REQUIREMENTS.md already ticks RELS-01, RELS-02 and RELS-07 as Complete. That claim rests on pre-review-fix evidence and should be re-confirmed by the items below.

### Anti-Patterns Found

I scanned the 131 phase-modified files (diff 711466f^..HEAD, excluding .planning and bindings).

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | - | TBD / FIXME / XXX | - | No debt markers |
| (none) | - | TODO / HACK / placeholder / not yet implemented | - | None found |
| frontend/src/App.svelte | search cancel path | Unhandled `CancelledRejectionError` on a tab switch mid-search (deferred-items.md) | ℹ️ Info | Console-only. It makes 18-backstops flaky about 1 run in 4, but it passed this run. Its owner is the Phase 2 search path |
| 04-REVIEW.md | - | 30 Info findings left open after 3 review iterations (IN-09 to IN-11 new) | ℹ️ Info | Non-blocking by the review's own classification. IN-09 (the login-item rollback while settings.json is unreadable) is worth a look |

### Human Verification Required

The hands-on checklist at `/private/tmp/claude-501/-Users-ken-dev-gif-search/f95b8e74-87dc-4816-8306-b8ba7d63e35c/scratchpad/phase4-handson.md` (items A1-E21) consolidates the planners' deferred `human-check` blocks for 04-01..04-07 and the untested review-fix items.

1. **Rebuild and re-ship 1.0.0.** Run `wails3 task darwin:release` on the current HEAD, replace the draft's DMG and .sha256, re-run TestRealDMG/TestReleaseJSON, and repeat the clean-VM install. Expected: notarized verify OK, the new digest on the draft, and a clean Gatekeeper first launch. Why human: notarization credentials and the clean machine.
2. **Release workflow dry run after merge.** Run `workflow_dispatch` on main with the split build/sign jobs. Expected: both jobs green and no release created. Why human: only the user triggers workflows. SC7 requires it before v1.0.1.
3. **Settings window, A1-A6.** Activation, ⌘V paste, hide/restore focus, theme, login item, Accessibility grant.
4. **Keys and Search, B7-B9.** A real key, the Keychain, rating and page size against live vendors.
5. **Hotkeys, C10-C14.** Rebinding, Favorites/Recents, recorder suspend, Reset to Default (SC2 is manual).
6. **Downloads and grid, D15-D18.** Folder sheet and TCC prompts, exFAT and network refusal, hidden names, grid sizes, Klipy ⌘S.
7. **Motion, About, Updates, E19-E21.** Reduce Motion, About links and marks in both themes.
8. **Phase 1 carry-over.** Fill the notarized paste matrix columns and the Slack, Discord, Telegram and browser-composer rows (21 PENDING cells in 01-PASTE-MATRIX.md).

### Gaps Summary

I found no code gaps. Every artifact exists, is substantive and is wired. Settings data flows to its real consumers. The full automated gate is green on HEAD 7e8792a: gofmt, vet, tidy, race tests, golangci-lint, the server build, 103 node tests, svelte-check, smoke and 35/35 headless UATs. The Klipy gate, the About credits and the verify-before-open updater each have direct tests.

The phase is not yet `passed` for three reasons, none of them code defects:
- **SC4 / RELS-01 / RELS-02.** The notarized DMG on the v1.0.0 draft is valid but stale, because it was built before the review fixes. Rebuild, re-notarize, replace the assets and repeat the clean install.
- **SC7 / RELS-07.** The split release workflow has not run on Actions. A dry run on merged main is required before v1.0.1.
- **SC2 and the native halves of SC1/SC3/SC5/SC6.** These are hands-on by design (Carbon hotkeys, login items, TCC, folder picker, WKWebView activation and paste).

Phase 4 is the last phase in the milestone, so none of these can be deferred to a later phase.

---

_Verified: 2026-10-20T17:04:43Z_
_Verifier: Claude (gsd-verifier)_
