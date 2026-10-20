---
phase: 04-settings-shell-polish-release
plan: 06
subsystem: release
tags: [notarytool, stapler, gatekeeper, spctl, dmg, github-release, tart, clean-install]

requires:
  - phase: 04-settings-shell-polish-release
    provides: "darwin:release task, scripts/verify-release.sh, docs/release-notes-1.0.0.md (04-05); internal/update TestRealDMG/TestReleaseJSON and DMGAssetName (04-07); every Settings/shell feature (04-01..04-04)"
provides:
  - "Notarized + stapled bin/release/Gifsmith.app and Gifsmith-1.0.0-arm64.dmg (SHA-256 a3129b96…395d4), RELEASE VERIFY OK (notarized)"
  - "Clean-machine install result (RELS-02): PASSED in a fresh macOS 15.7.7 VM from the quarantined DMG"
  - "Draft GitHub Release 'Gifsmith 1.0.0' (tag v1.0.0 on publish, target main) on scottkw/gifsmith with the DMG and its .sha256"
  - "01-PASTE-MATRIX.md notarized column and N15"
  - "Signed dev build installed in ~/Applications for the end-of-phase native UAT"
affects: [04-08, phase-04-verification, release]

actuals:
  tokens: 9500
  tasks: 3
  commits: 5
plan_head_before: 40ddab077cfe6c640c6bebc7f3b57a1c2cc4637b
plan_head_after: 1368868ca5f5b1c61d6bdf542403465ce29af984

tech-stack:
  added: []
  patterns:
    - "Clean-machine RELS-02 test in a throwaway Tart VM instead of a second user account on the primary Mac: stricter (fresh OS, no shared /Applications, app launched translocated) and leaves nothing behind"

key-files:
  created:
    - .planning/phases/04-settings-shell-polish-release/04-06-SUMMARY.md
  modified:
    - .planning/phases/01-native-spike-signed-dev-loop/01-PASTE-MATRIX.md
    - scripts/verify-release.sh
    - scripts/uat/18-backstops.js

key-decisions:
  - "RELS-02 was run in a fresh Tart VM (macOS 15.7.7 Sequoia vanilla, SIP on, Gatekeeper assessments re-enabled with spctl --master-enable) instead of a second account on the primary Mac; same quarantined DMG, stricter conditions, and nothing lands in the host's shared /Applications"
  - "The draft targets main although main (32eee42, Phase 3 merge) does not contain Phase 4 yet; a draft creates no tag, so this is safe only if the phase branch is merged to main before the user publishes"
  - "Finer Gifs was down during the test (api.thefinergifs.club HTTP 530 from VM and host); the paste check used a user-entered GIPHY key instead, and the app's 'seems to be down' state was itself verified"

patterns-established:
  - "Release verification evidence lives in bin/release.log (notarytool ids, spctl, stapler, DR comparison), bin/update-verify.log (TestRealDMG) and bin/release-json-verify.log (TestReleaseJSON on the draft JSON)"

requirements-completed: [RELS-01, RELS-02]

coverage:
  - id: D1
    description: "Whole phase green on one tree: gofmt, go vet, go test -race, golangci-lint, CGO-free server build, node tests, svelte-check, smoke and every scripts/uat/*.js"
    verification:
      - kind: integration
        ref: "full suite re-run on HEAD 1368868 (2026-10-20), source identical to the release build"
        status: pass
    human_judgment: false
  - id: D2
    description: "darwin:release notarizes and staples app and DMG; spctl source=Notarized Developer ID; stapler validate; SHA-256 matches; designated requirement equals the signed dev build's"
    requirement: RELS-01
    verification:
      - kind: integration
        ref: "bin/release.log: RELEASE VERIFY OK (notarized); notary ids a46eb52e… and f5f85406… Accepted"
        status: pass
    human_judgment: false
  - id: D3
    description: "The 04-07 updater accepts the shipped DMG and the draft release JSON"
    requirement: RELS-06
    verification:
      - kind: unit
        ref: "internal/update TestRealDMG (UPDATE_REAL_DMG) and TestReleaseJSON (UPDATE_RELEASE_JSON + UPDATE_DMG_PATH)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Clean-machine install from the quarantined DMG: Gatekeeper accepts, menu bar app with no Dock icon, first-run key screen, copy and paste an animated GIF into TextEdit, About and update check"
    requirement: RELS-02
    verification:
      - kind: manual
        ref: "user, fresh Tart VM macOS 15.7.7, 2026-10-20 (Task 2)"
        status: pass
    human_judgment: true
    rationale: "Gatekeeper first-launch and real desktop paste can only be checked by a person on a clean machine"
  - id: D5
    description: "Draft GitHub Release 'Gifsmith 1.0.0' with exactly the DMG and .sha256; no tag on the remote"
    verification:
      - kind: integration
        ref: "bin/release-draft.txt = true|Gifsmith 1.0.0|Gifsmith-1.0.0-arm64.dmg,Gifsmith-1.0.0-arm64.dmg.sha256; git ls-remote --tags origin v1.0.0 empty"
        status: pass
    human_judgment: false

duration: about 25 min of executor time across two sessions, plus the user's VM test
completed: 2026-10-20
status: complete
---

# Phase 4 Plan 06: Notarized 1.0.0, Clean-Machine Install and Draft Release Summary

**Gifsmith 1.0.0 is notarized and stapled (app and DMG both Accepted by Apple), passed a clean-machine install in a fresh macOS 15.7.7 VM from the quarantined DMG with only the standard "downloaded from the Internet / Apple checked it" prompt, and now waits as a draft GitHub Release with the DMG and its checksum. The 04-07 updater accepts both the DMG and the draft's release JSON.**

## Performance

- **Started:** 2026-10-20T03:02Z (Task 1); Task 3 resumed 2026-10-20T13:40Z
- **Completed:** 2026-10-20T13:53:57Z
- **Tasks:** 3 of 3 (Task 2 is the user's clean-machine checkpoint)
- **Files modified:** 3 (plus this summary)

## Accomplishments

- The whole phase is green on one tree (results below), and the notarized release was built from that tree.
- `wails3 task darwin:release` notarized and stapled the app and the DMG, and `verify-release.sh --notarized` passed.
- The release build's designated requirement equals the signed dev build's, so an Accessibility grant carries over (N15 PASS).
- `TestRealDMG` passes on a temp copy of the shipped DMG, so every 1.0.x updater will accept it. `TestReleaseJSON` passes on the draft release JSON: asset name, download URL prefix and digest match.
- The user installed the quarantined DMG in a fresh VM and it passed every step (results below).
- Draft release created: https://github.com/scottkw/gifsmith/releases/tag/untagged-f7666a5d9e82df445fe9

## Task Commits

1. **Task 1: full suite, notarized release, verification, signed dev install**
   - `8428b30` fix(04-04): keyboard backstop UAT walks the About pane added in 04-05
   - `3a8db9a` fix(04-05): verify-release prints the spctl assessments it checks
   - `5c63b84` docs(04-06): paste matrix records the notarized build and N15
2. **Task 2: clean-machine install (checkpoint):** no commit. The user's result is below.
3. **Task 3: draft release and outcome:** `1368868` docs(04-06): paste matrix notarized cells from the clean-VM install test, plus the final docs commit for this summary.

`d82cf66` (docs(04-08), the plan for GitHub Actions CI) landed between Task 1 and Task 3. It is counted in `commits:` because that number is measured from the plan's base ref, but it is not 04-06 work.

## Full-suite regression (re-run on HEAD 1368868, 2026-10-20)

The release was built at 8428b30. The source is unchanged since then: 3a8db9a only adds output lines to verify-release.sh, and the later commits are docs. Everything passed, with one flaky UAT run noted below the table:

| Check | Result |
|-------|--------|
| gofmt (main.go, wire_*.go, internal) | clean |
| go vet . ./internal/... | PASS |
| go test -race ./internal/... -count=1 | PASS (14 packages ok) |
| golangci-lint run . ./internal/... | 0 issues |
| CGO_ENABLED=0 go build -tags server . | PASS |
| pnpm --dir frontend run test | 101 tests, 101 pass, 0 fail |
| pnpm --dir frontend run check | 191 files, 0 errors, 0 warnings |
| bash scripts/smoke-browser-loop.sh | SMOKE OK |
| scripts/uat/*.js via smoke --uat (33 scripts) | 32 UAT OK on the first pass; 18-backstops failed once with a page error (see below), then passed on 3 of 3 re-runs |

UAT OK: 01-tracer, 02-grid, 02-keyboard, 03-states, 05-setup, 05-tabs, 06-backstops, 06-clips-denied, 06-clips, 07-text-actions, 08-action-menu, 10-download, 11-square, 12-library-unreadable, 12-library, 13-chips, 13-details, 14-backstops, 15-settings-general, 15-settings-startup-fail, 16-settings-keys, 16-settings-search, 17-hotkeys-fail, 17-hotkeys, 17-polish, 18-backstops (re-run), 18-default-actions, 18-downloads, 18-grid-size, 19-about, 20-update-current, 20-update-fail, 20-update.

**Flake:** on the first pass, 18-backstops failed with `page error: A: CancelledRejectionError: Unhandled rejection in cancelled promise. Reason: RuntimeError: context canceled`. Section (d) switches from a Finer Gifs search to GIPHY while a search binding call is still running. The cancelled call's rejection reaches the page unhandled, and the test's strict page-error check catches it. Users see nothing, because it only shows in the webview console. It is not caused by this plan and was not fixed here (scope boundary). It is logged in `deferred-items.md` for the search/cancel owner.

## Release verification (bin/release.log)

| Item | Value |
|------|-------|
| App notarization | submission `a46eb52e-c6e6-4ac2-8b28-05af458ca034`, status **Accepted**, stapled ("The staple and validate action worked!") |
| DMG notarization | submission `f5f85406-c0dc-45d8-bb46-df79e7293a1e`, status **Accepted**, stapled |
| spctl, app (exec) | `bin/release/Gifsmith.app: accepted`, `source=Notarized Developer ID`, `origin=Developer ID Application: Ken Scott (S2K7P43927)` |
| spctl, DMG (open, primary-signature) | `accepted`, `source=Notarized Developer ID`, same origin |
| stapler validate | "The validate action worked!" for both the app and the DMG (re-checked 2026-10-20) |
| Designated requirement | release == dev: `identifier "com.kenscott.gifsmith" and anchor apple generic and certificate 1[field.1.2.840.113635.100.6.2.6] /* exists */ and certificate leaf[field.1.2.840.113635.100.6.1.13] /* exists */ and certificate leaf[subject.OU] = S2K7P43927` |
| DMG | `Gifsmith-1.0.0-arm64.dmg`, 6,075,068 bytes |
| SHA-256 | `a3129b96641fc156859af9ad96bc8f76e6cc4428abe056aa8f23f355089395d4` (`shasum -c` OK; the /Users/Shared copy used for the VM test is identical) |
| Verdict | `RELEASE VERIFY OK (notarized)` |
| Updater on the real DMG | `--- PASS: TestRealDMG` (bin/update-verify.log) |
| Signed dev install | bin/install-dev.log: `tray menu ready`, `hotkey registered: Cmd+Ctrl+Option+G`; go.sum restored and unchanged |
| Leak check | no `api_key=`, no Klipy `/api/v1/<key>/` path, no Apple ID or app-specific password in bin/release.log or bin/install-dev.log |

## Clean-machine install (Task 2, RELS-02): PASSED, 2026-10-20

The user ran the test in a fresh Tart VM: macOS 15.7.7 Sequoia vanilla image, SIP on. The vanilla image ships with Gatekeeper assessments disabled, so `spctl --master-enable` turned them back on first. The VM was a deliberate change from the plan's second user account (see Deviations).

- The DMG was copied into the VM with the quarantine xattr (`0081;…;Safari;`). `shasum -a 256 -c` printed OK. `spctl -a -t open --context context:primary-signature` returned accepted with `source=Notarized Developer ID`, and `stapler validate` worked.
- The app was installed to /Applications from the mounted DMG with ditto over SSH, so it launched translocated, which is stricter than a Finder drag. `spctl -a -t exec` returned accepted, `source=Notarized Developer ID`, origin `Developer ID Application: Ken Scott (S2K7P43927)`. The hardened runtime flag is present, CFBundleShortVersionString is 1.0.0 and LSUIElement is true.
- First launch showed only the standard prompt: "“Gifsmith” is an app downloaded from the Internet… Apple checked it for malicious software and none was detected". After Open, the app ran. There was no "cannot be opened" or "unidentified developer" block.
- A menu bar icon appeared and there was no Dock icon. ⌃⌥⌘G opened the panel on the first-run GIPHY key screen.
- Finer Gifs was down (api.thefinergifs.club returned HTTP 530 from both the VM and the host), and the app correctly showed "The Finer Gifs Club seems to be down". The user entered a GIPHY key instead, and trending loaded.
- Searching "office" and pressing Enter copied the GIF (clipboard `«class furl»`), hid the panel and showed "Copied — press ⌘V to paste". ⌘V into TextEdit (rich text) pasted an animated GIF; three frame samples differed.
- ⌘Enter with no Accessibility grant copied only. The footer read "Copied — press ⌘V to paste" with the "Enable Automatic Paste…" button, and TextEdit stayed frontmost (non-activating panel).
- In Settings → General → Updates, Check for Updates… showed "✓ Gifsmith 1.0.0 is up to date." About shows Version 1.0.0, the credits to the Raycast GIF Search authors, and MIT.
- The VM was deleted afterwards. `/Applications/Gifsmith.app` does not exist on the host (checked 2026-10-20). `~/Applications/Gifsmith.app`, the dev install, was not touched.

Not covered by this test: the Finer Gifs path (service down), the notarized ⌘Enter auto-paste with Accessibility granted (no grant was made in the VM), and the optional step 8 (Slack/Messages). The DMG's Finder window look (background, arrow) was not reported, because the install went through ditto.

## Draft GitHub Release (Task 3)

- `gh release view v1.0.0` failed before creation (no release, no tag).
- `gh release create v1.0.0 --repo scottkw/gifsmith --draft --target main --title "Gifsmith 1.0.0" --notes-file docs/release-notes-1.0.0.md` created the draft with the DMG and its .sha256.
- Draft URL: https://github.com/scottkw/gifsmith/releases/tag/untagged-f7666a5d9e82df445fe9 (GitHub gives drafts an `untagged-…` URL; on publish it becomes `/releases/tag/v1.0.0`)
- `bin/release-draft.txt`: `true|Gifsmith 1.0.0|Gifsmith-1.0.0-arm64.dmg,Gifsmith-1.0.0-arm64.dmg.sha256`

| Asset | Size (bytes) | GitHub digest |
|-------|--------------|---------------|
| Gifsmith-1.0.0-arm64.dmg | 6,075,068 | `sha256:a3129b96641fc156859af9ad96bc8f76e6cc4428abe056aa8f23f355089395d4` (equals the local DMG) |
| Gifsmith-1.0.0-arm64.dmg.sha256 | 91 | `sha256:0a43383691fc6e42e88b19ef10fd1b8f04fdd6556400bfcc968759529a18b537` |

- `UPDATE_RELEASE_JSON=…/bin/release-draft.json UPDATE_DMG_PATH=…/bin/release/Gifsmith-1.0.0-arm64.dmg go test ./internal/update -run TestReleaseJSON -count=1 -v` prints `--- PASS: TestReleaseJSON` (bin/release-json-verify.log).
- `git ls-remote --tags origin v1.0.0` prints nothing, and there is no local v1.0.0 tag. Nothing was published or pushed.

## What the user must do before publishing

1. **Merge Phase 4 into main first.** The draft targets `main`, which is still at `32eee42` (the Phase 3 merge) and has none of Phase 4. Publishing creates tag `v1.0.0` on whatever main points to at that moment. If you publish before the merge, the tag lands on Phase 3 code while the DMG is Phase 4, and the release's source archives won't match the binary.
2. After the merge, publish the draft on GitHub (Releases → Gifsmith 1.0.0 → Edit → Publish release). Tag v1.0.0 is created on main at that point.
3. After publishing, Check for Updates… in 1.0.0 should show "up to date" through the real 200 response. The first full in-app download → verify → open happens when 1.0.1 is published with the same asset naming (`Gifsmith-{version}-arm64.dmg` plus its digest).
4. Still open, not blocking the draft: email developers@klipy.com about Klipy clearance (04-KLIPY-EMAIL.md). Klipy file actions stay off in 1.0.0 either way.

## Paste matrix (01-PASTE-MATRIX.md)

- Notarized build commit: 8428b30, with both notary ids. The test machine is recorded as the macOS 15.7.7 VM.
- N15: PASS (same designated requirement as the signed dev build).
- Notarized PASS: N1, N4, N7 (Dock; ⌘Tab not reported), N11, N12, and TextEdit Copy→⌘V.
- Everything else in the notarized column stays PENDING, marked "not tested in 04-06": N2, N3, N5, N6, N8, N10, N13, N14, TextEdit ⌘Enter, and every other app.

## Deviations from Plan

1. **[Rule 3 - Blocking/environment] Clean-machine test ran in a Tart VM, not a second macOS account.** The plan's Task 2 used a fresh user account on the primary Mac. The user ran it in a fresh macOS 15.7.7 VM instead, which keeps the same quarantined-DMG path and adds a fresh OS and translocated launch. One consequence: the clean-machine evidence is from macOS 15.7.7, not macOS 27 (the dev columns stay on macOS 27.0.1). Step 9's cleanup became deleting the VM.
2. **Finer Gifs down.** Plan step 5 used Finer Gifs. It returned HTTP 530, so the paste check used GIPHY with a user-entered key. The app's "seems to be down" state was verified instead, which the plan allowed for.
3. **Draft targets main before Phase 4 is merged.** The plan's command was run as written (`--target main`). Because main lacks Phase 4, publishing must wait for the merge (see "What the user must do").
4. **Full suite re-run in Task 3.** The Task 1 session did not leave a per-script record, so the whole suite was re-run on HEAD to give this summary real evidence. All results above come from that re-run.

## Known Stubs

None.

## Threat Flags

None. No new code surface. The only external write is the draft release under the user's gh login (T-04-27 mitigated: draft only, asserted isDraft true, no remote tag).

## Self-Check: PASSED

- Commits found in HEAD history: 8428b30, 3a8db9a, 5c63b84, 1368868
- Files found: bin/release/Gifsmith-1.0.0-arm64.dmg (+ .sha256), bin/release-draft.txt, bin/release-draft.json, bin/release-json-verify.log, bin/update-verify.log, deferred-items.md
- Draft release: isDraft true. `git ls-remote --tags origin v1.0.0` printed 0 lines.
- /Applications/Gifsmith.app is absent on the host. ~/Applications/Gifsmith.app (the dev install) is present and untouched.
- go.sum is unchanged (`git diff --quiet -- go.sum`).
