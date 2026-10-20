---
phase: 01-native-spike-signed-dev-loop
verified: 2026-10-20T19:21:04Z
status: gaps_found
score: 15/16 must-haves verified (1 failed: SC5 per-app paste matrix breadth; 0 behavior-unverified, 0 uncertain, 0 deferred)
covered_files:
  - ".planning/phases/01-native-spike-signed-dev-loop/01-01-PLAN.md"
  - ".planning/phases/01-native-spike-signed-dev-loop/01-01-SUMMARY.md"
  - ".planning/phases/01-native-spike-signed-dev-loop/01-02-PLAN.md"
  - ".planning/phases/01-native-spike-signed-dev-loop/01-02-SUMMARY.md"
  - ".planning/phases/01-native-spike-signed-dev-loop/01-03-PLAN.md"
  - ".planning/phases/01-native-spike-signed-dev-loop/01-03-SUMMARY.md"
  - ".planning/phases/01-native-spike-signed-dev-loop/01-04-PLAN.md"
  - ".planning/phases/01-native-spike-signed-dev-loop/01-04-SUMMARY.md"
  - ".planning/phases/01-native-spike-signed-dev-loop/01-PASTE-MATRIX.md"
  - "Taskfile.yml"
  - "build/config.yml"
  - "build/darwin/Info.dev.plist"
  - "build/darwin/Info.plist"
  - "build/darwin/Taskfile.yml"
  - "frontend/package.json"
  - "frontend/src/App.svelte"
  - "frontend/src/components/Footer.svelte"
  - "frontend/src/lib/api.ts"
  - "frontend/vite.config.ts"
  - "go.mod"
  - "internal/platform/fake.go"
  - "internal/platform/platform.go"
  - "internal/platform/platform_darwin.go"
  - "internal/platform/platform_darwin.h"
  - "internal/platform/platform_darwin.m"
  - "internal/platform/platform_darwin_test.go"
  - "internal/platform/platform_stub.go"
  - "internal/services/actions.go"
  - "internal/services/actions_test.go"
  - "internal/services/hotkeys.go"
  - "internal/services/panel.go"
  - "internal/settings/settings.go"
  - "internal/shell/position.go"
  - "internal/shell/position_test.go"
  - "internal/shell/shell.go"
  - "main.go"
  - "scripts/notarize.sh"
  - "scripts/smoke-browser-loop.sh"
  - "scripts/verify-release.sh"
  - "wire_desktop.go"
covered_digest: "v3:sha256:fafa3b4069af346673e555f0a7dc466a07d32dfde120bf304ce494a3a7ec123d"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: "12/16 (1 deferred to Phase 4, 2 present but behavior-unverified, 1 uncertain)"
  reason: "Flagged stale: Phase 4 changed covered code (shell.go, actions.go, App.svelte, Taskfile, main.go, platform). Re-verified 2026-10-20 after Phase 4."
  gaps_closed:
    - "SC5 notarized half: notarized + stapled builds exist (8428b30, then the published v1.0.0); /Applications/Gifsmith.app spctl 'source=Notarized Developer ID', stapler validate OK; designated requirement identical to the dev build (N15 PASS; scripts/verify-release.sh:88-92 enforces it)"
    - "01-04 artifact build/darwin/Taskfile.yml --keychain-profile: now present via NOTARY_PROFILE + scripts/notarize.sh (verify.artifacts 01-04 3/3)"
    - "N13 Accessibility request: same ActionService.RequestAccessibility observed through Settings 'Grant Access…' (04-UAT test 4, checklist A6 PASS); panel button wired to the same api call"
    - "P5 Settings…: superseded by the real Settings window (shell.OpenSettings), observed in 04-UAT checklist A1"
    - "P6 Quit Gifsmith: observed in Phase 3 native UAT (03-VERIFICATION.md:51 'app process exited (Quit observed)')"
    - "P11 smoke loop: SMOKE OK recorded by the Phase 4 verifier (04-VERIFICATION.md:263) on code identical to HEAD (last non-planning commit 9408093 precedes it)"
    - "Final-build regression (old human item 1): superseded by clean-VM install of notarized 1.0.0 (N1, N4, N7, N11, N12 PASS) and host hands-on UAT on the rebuilt 1.0.0 (04-UAT tests 4-8)"
  gaps_remaining:
    - "SC5 matrix breadth: Slack, Discord, Telegram, Safari composer, Chrome composer rows still PENDING (previously deferred to Phase 4; Phase 4 completed without filling them)"
  regressions: []
gaps:
  - truth: "ROADMAP SC5: a committed paste matrix records the result per target app (Slack, Messages, Mail, Discord, Notes, Telegram, Safari/Chrome composers)"
    status: partial
    reason: "Deferred to Phase 4 by the 2026-10-17 GO WITH FALLBACKS decision; Phase 4 (the last phase of v1.0) completed and v1.0.0 shipped without these rows being tested, so the deferral no longer has a target phase. 21 PENDING cells remain. No code defect: the file-URL payload is verified for TextEdit, Notes, Mail and Messages only."
    artifacts:
      - path: ".planning/phases/01-native-spike-signed-dev-loop/01-PASTE-MATRIX.md"
        issue: "Lines 46-50: Slack, Discord, Telegram, Safari composer, Chrome composer are PENDING in all four columns; line 52 still says they stay PENDING 'for the 01-04 notarized UAT'; notarized ⌘Enter column and N2/N3/N5/N6/N8/N9/N10/N13/N14 notarized cells PENDING"
    missing:
      - "Run Copy→⌘V (and ⌘Enter with Accessibility on) from the installed v1.0.0 into Slack, Discord, Telegram, a Safari composer and a Chrome composer and record PASS/STATIC/FAIL/n/a, OR record a user override accepting these as v1.0.1 tech debt (already listed in .planning/v1.0-MILESTONE-AUDIT.md tech_debt)"
---

# Phase 1: Native Spike & Signed Dev Loop Verification Report

**Phase Goal:** The three unproven native behaviors work on this Mac (typing into a non-activating panel, full-screen overlay, and copy/paste of a GIF file into the previous app), from a stable signed identity, so the architecture is confirmed before any real UI is built.
**Verified:** 2026-10-20T19:21:04Z
**Status:** gaps_found (one documentation/evidence gap; no code gap)
**Re-verification:** Yes. **Re-verified 2026-10-20 after Phase 4**, against branch `gsd/phase-04-settings-shell-polish-release` HEAD 7874dc6 (same code as main). The previous report (2026-10-17, `passed`) was flagged stale because Phase 4 changed code it covers.

## Re-verification note (2026-10-20, after Phase 4)

What Phase 4 (and Phases 2-3) changed in Phase 1's covered code, and what it means for Phase 1:

- **`internal/shell/shell.go`**: the panel options that carry the native behaviors are unchanged (`MacWindowClassPanel`, `NonActivating: true`, `CanJoinAllSpaces|FullScreenAuxiliary`, `MacWindowLevelFloating`, `HideOnFocusLost`, `show()` = SetPosition + Show + Focus + `FocusWebView`). Changes: `HideOnEscape` is now `false` and the page owns Esc (clear the query first, then `PanelService.Hide`, App.svelte `case 'escape'`); a `WindowClosing` hook keeps the panel from being destroyed; a real Settings window was added; the tray gained the Open accelerator and an Update item. The hotkey moved from `shell.Hotkey` to `services.Hotkeys` (`settings.DefaultOpenHotkey = "Cmd+Ctrl+Option+G"`, rebindable), registered at `ApplicationStarted` via `app.GlobalShortcut` (Carbon, no Accessibility).
- **`internal/services/actions.go`**: `Copy` now stages a real provider GIF (the spike's bundled `sample.gif` was removed in Phase 2) and no longer hides the panel itself (the page confirms, then hides, `scheduleHide`). `Paste` still runs the Phase 1 sequence through `paste()` unchanged in order: copy, `AccessibilityTrusted(false)` gate, hide, delay, reactivate the recorded app or copy only, one `PasteShortcut`.
- **`internal/platform/platform_darwin.m`**: additions only (+65, no deletions since 7111c39). `gsCopyFile` (writeObjects file URL), `gsFocusWebView`, `gsCursorScreenVisibleFrame`, `gsAXTrusted`, `gsPostCmdV` (private event source) are unchanged.
- **Release**: notarization, previously deferred to Phase 4, is done. `release`/`release:sign` tasks use `NOTARY_PROFILE` and `scripts/notarize.sh`; `scripts/verify-release.sh` checks `source=Notarized Developer ID`, stapler, and that the release designated requirement equals the dev build's. v1.0.0 was notarized and published 2026-10-20 (GitHub release v1.0.0, not draft, published 2026-10-20T19:03:44Z) and passed a clean-VM install (04-UAT test 3).
- **Superseded spike placeholders**: the "Settings window ships in Phase 4" status line (P5) is gone in behavior; Settings… now opens the real window. The `panel:status` listener remains in App.svelte:293 with no emitter (dead code, already listed in the v1.0 audit tech debt).

## Mode discrepancy (MVP)

Unchanged from the first report: ROADMAP still marks Phase 1 `**Mode:** mvp`, but the goal is not a user story. As before, this report uses the standard goal-backward method and has no User Flow Coverage section. User decision still open: reformat with `/gsd-mvp-phase 1` or drop `mode: mvp` for Phase 1.

## Goal Achievement

The phase goal holds in today's code and is now proven on more builds than at the original gate. The three native behaviors (non-activating panel that takes typing, full-screen overlay, file-URL copy and Accessibility-gated paste into the previous app) are produced by the same native code as at the gate, and the Go sequencing is unit-tested and passing. They were observed on the Developer ID dev build (e9f1230), during later native UAT on Phase 3 and Phase 4 builds, and on the notarized v1.0.0 on a clean VM. The signed identity is stable: the installed `/Applications/Gifsmith.app` (1.0.0) has designated requirement `identifier "com.kenscott.gifsmith" and anchor apple generic … leaf[subject.OU] = S2K7P43927`, with no cdhash.

The one remaining shortfall is SC5's per-app matrix. Slack, Discord, Telegram and the Safari/Chrome composers were never tested. They were deferred to Phase 4, and Phase 4 is complete. That is an evidence gap, not a code gap, and it is already recorded as v1.0 tech debt.

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| SC1 | From any app, incl. full-screen, ⌃⌥⌘G shows a floating panel that takes typing at once; the previous app keeps focus; hotkey again / Esc / click-away hides | ✓ VERIFIED | Matrix N1-N6 PASS (signed dev); notarized v1.0.0 clean VM: N1, N4 PASS; 04-UAT test 6 (hotkeys from other apps toggle the panel, notarized 1.0.0 on host). Code: shell.go:51-66 panel options; `Toggle` (shell.go:131) bound as `"open"` in main.go:161; `show()` focuses the webview (shell.go:147-161). Esc: App.svelte `case 'escape'` → `hide()` → `PanelService.Hide` → `sh.Hide` (main.go:170). Default ⌃⌥⌘G: settings.go:224 |
| SC2 | Menu bar icon only (no Dock, not in ⌘Tab); menu offers Open, Settings, Quit; Open shows the panel when the hotkey is unavailable | ✓ VERIFIED | N7 PASS (dev + notarized); `ActivationPolicyAccessory` main.go:138; Info.plist `LSUIElement=true`. Tray built in `shell.New` (shell.go:106-116) independent of hotkeys, which register later in `ApplicationStarted` (main.go:177-185) and only log on failure (`TestStartFailureIsNotBound` PASS). Settings… now opens the real window (Phase 4, as SC2 anticipated) |
| SC3 | Enter puts a GIF on the clipboard as a file; ⌘V pastes it animated into TextEdit, Accessibility NOT granted | ✓ VERIFIED (sample GIF superseded by real results) | N11 PASS dev and notarized (clean VM, GIPHY result, animated). `TestCopyFileWritesFileURL` (public.file-url, one URL) and `TestCopyStagesAndDoesNotHide` PASS; `Copy` makes no AX call (actions.go:113-130). Enter → `primaryAction` default `copy` (settings.go:212) → `run('copy')` → `ActionService.Copy` (api.ts:124) |
| SC4 | With Accessibility, paste drops the GIF into the previous app with no manual ⌘V; without it, copy + tell the user to press ⌘V and how to grant | ✓ VERIFIED | N14 PASS (dev, ×5 + Notes/Mail/Messages); N12 PASS dev and notarized ("Copied — press ⌘V to paste" + "Enable Automatic Paste…", Footer.svelte:62-64). Paste path exercised again in Phase 3 native UAT (Paste Square uses the same `paste()`). Grant flow: `RequestAccessibility` (actions.go:392) observed via Settings Grant Access… (04-UAT A6 PASS). `paste()` ordering tests all PASS |
| SC5 | Same behaviors from a signed dev build AND a notarized build; committed per-app paste matrix; explicit go/no-go | ✗ FAILED (partial: matrix breadth only) | **Met now:** notarized build exists and was used (clean VM 04-UAT test 3; host UAT 04 tests 4-8 on the rebuilt 1.0.0); `spctl` on /Applications/Gifsmith.app → `source=Notarized Developer ID`, `stapler validate` OK; DR identical to dev (N15 PASS, verify-release.sh:88-92). Go/no-go `Decision: GO WITH FALLBACKS` committed (01-PASTE-MATRIX.md §4). **Not met:** 01-PASTE-MATRIX.md:46-50 Slack, Discord, Telegram, Safari composer, Chrome composer have no result in any column (21 PENDING cells overall). Deferral target Phase 4 is complete, so this is no longer deferrable |
| P1 | Locked identity: com.kenscott.gifsmith, LSUIElement, LSMinimumSystemVersion 13.0, arm64-only, Wails v3.0.0-beta.30 pinned, pnpm lockfile | ✓ VERIFIED | Info.plist id/LSUIElement/13.0 (version now 1.0.0); `DMG_NAME … -arm64.dmg`; go.mod:8 `v3.0.0-beta.30`; package.json `"@wailsio/runtime": "3.0.0-beta.30"`; `frontend/pnpm-lock.yaml` tracked |
| P2 | Enter twice leaves exactly one file URL; the staged file stays while it is on the clipboard | ✓ VERIFIED (staging semantics refined) | `TestCopyFileWritesFileURL` double write, `TestCopyTwiceKeepsOnlyLatest` PASS. Phase 2-4 refinement: staging is cleared at launch and quit (main.go:192, `ServiceShutdown`), and Klipy pastes are deleted after a grace period (D-17/D-19) |
| P3 | A failed pasteboard write shows an error and the panel stays open | ✓ VERIFIED | `TestCopyErrorKeepsPanelOpen` PASS; App.svelte `actionFailed` → error status, no `scheduleHide` on the error path |
| P4 | Hotkey registration outcome is logged | ✓ VERIFIED | hotkeys.go:96/113/121 (`hotkey register error`, `not registered`, `hotkey registered`) |
| P5 | Settings… opens the panel with "Settings window ships in Phase 4" | ✓ VERIFIED (superseded) | Replaced by the real Settings window: tray `Settings…` → `s.OpenSettings()` (shell.go:112, 223-249). Observed: 04-UAT test 4, checklist A1 |
| P6 | Quit Gifsmith quits and the icon disappears | ✓ VERIFIED | `app.Quit()` shell.go:115; observed in Phase 3 native UAT ("app process exited (Quit observed)", 03-VERIFICATION.md:51) |
| P7 | Panel opens on the cursor's display, centered, top at 1/6 of visible height | ✓ VERIFIED | N9 PASS (dev); `TestPanelOrigin` PASS; shell.go:154-156 |
| P8 | `wails3 task darwin:install:dev` builds, signs Developer ID + hardened runtime, installs to ~/Applications, launches | ✓ VERIFIED | build/darwin/Taskfile.yml:219-238 unchanged in substance (sign `--identity "{{.SIGN_IDENTITY}}" --hardened-runtime`, path guard, ditto, open). Not re-run (it launches the app) |
| P9 | Designated requirement is identifier + team, not a per-build hash | ✓ VERIFIED | `codesign -d -r-` on /Applications/Gifsmith.app (notarized 1.0.0): identifier + Developer ID anchor + leaf OU S2K7P43927; verify-release.sh enforces release DR == dev DR |
| P10 | Paste ordering invariant: copy → AX gate (no prompt) → hide → delay → reactivate or copy only → exactly one ⌘V; one action at a time | ✓ VERIFIED (behavioral, tests ran) | `TestPasteTrustedOrder`, `…ReactivatesPrevious`, `…DeclinedReactivationCopiesOnly`, `…UnknownTargetCopiesOnly`, `…FrontmostErrorCopiesOnly`, `…CopyError`, `…PostError`, `…WithoutAccessibility`, `TestActionsRefuseWhileBusy`: all PASS this run. `PasteShortcut()` has one call site (actions.go:216) |
| P11 | `scripts/smoke-browser-loop.sh` round-trips through the Vite /wails proxy and stops only its own processes | ✓ VERIFIED (by the Phase 4 verifier's run) | 04-VERIFICATION.md:263 "SMOKE OK; 35/35 UAT OK", run after the last code commit (9408093, 16:54Z, before 04-VERIFICATION at 17:04Z). Not re-run here: ports 8089/9245 are shared with the concurrent Phase 2/3 re-verifiers |

**Score:** 15/16 truths verified (0 present but behavior-unverified, 1 failed: SC5 matrix breadth)

### Deferred Items

None. All three items deferred to Phase 4 in the first report have been resolved: the notarized build and N13 are now verified (see `re_verification.gaps_closed`), and the matrix breadth is now an open gap. There is no later phase in v1.0 to defer to.

### Advisory (New Scope, Unevidenced)

| # | Finding | Category | Why Advisory |
|---|---------|----------|--------------|
| 1 | `Events.On('panel:status', …)` at frontend/src/App.svelte:293 has no emitter now that the Phase 1 Settings placeholder is gone | other | Dead listener, harmless; already in v1.0 audit tech debt |
| 2 | 01-PASTE-MATRIX.md:52 and §3 still say untested apps "stay PENDING for the 01-04 notarized UAT" | other | Stale wording; resolves with the gap |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/platform/platform.go` | Platform interface, ErrUnsupported | ✓ VERIFIED | Extended (appearance, open URL, reveal, activate) without removing Phase 1 methods |
| `internal/platform/platform_darwin.m` | writeObjects, first responder, visible frame, AX, private-source ⌘V | ✓ VERIFIED | Lines 31-36, 96-100, 106-116, 125-129, 160-172; additive changes only |
| `internal/platform/platform_stub.go` | CGO-free stub | ✓ VERIFIED | Present; server build used by the smoke/UAT harness |
| `internal/services/actions.go` | Copy, Paste, RequestAccessibility, PasteDelay | ✓ VERIFIED | Bound via `application.NewService(actions)` (main.go:124) |
| `internal/shell/shell.go` | Panel, tray, placement, Previous | ✓ VERIFIED | `verify.artifacts` reports "Missing pattern: Ctrl+Option+Cmd+G": the hotkey moved to settings.go:224 (`Cmd+Ctrl+Option+G`, canonical form) and `services.Hotkeys`. Relocation, not loss |
| `internal/shell/position.go` / `position_test.go` | panelOrigin + table test | ✓ VERIFIED | PASS |
| `frontend/src/App.svelte` | Input, status; Enter → Copy, ⌘Enter → Paste, button → RequestAccessibility | ✓ VERIFIED | Calls go through `frontend/src/lib/api.ts` (124/132/140), so the plan's `ActionService\.Copy` grep in App.svelte no longer matches |
| `frontend/public/sample.gif` | Animated sample | ✓ SUPERSEDED | Removed in Phase 2 when real search replaced the spike sample (02-VERIFICATION.md:167). Copy now stages provider GIFs |
| `build/darwin/Taskfile.yml` | install:dev (01-03) | ✓ VERIFIED | Lines 219-238 |
| `build/darwin/Taskfile.yml` | keychain profile for notarization (01-04) | ✓ VERIFIED (was deferred) | `NOTARY_PROFILE: "gifsmith-notary"`; `scripts/notarize.sh` uses `--keychain-profile` |
| `scripts/smoke-browser-loop.sh` | Browser UAT smoke | ✓ VERIFIED | See P11 |
| `01-PASTE-MATRIX.md` | Matrix, payload conclusion, Decision line | ⚠️ PARTIAL | Decision and notarized N-rows filled (04-06); five app rows empty (gap) |
| `.planning/PROJECT.md` | Key Decisions updated | ✓ VERIFIED | Unchanged since the first report |

`gsd-tools verify.artifacts` (today): 01-01 5/7 (shell.go hotkey string relocated; sample.gif superseded), 01-02 4/4, 01-03 5/5, 01-04 3/3.

### Key Link Verification

| From | To | Via | Status |
|------|----|-----|--------|
| App.svelte | actions.go | `run()` → `copy`/`paste` (api.ts:124/132) → `ActionService.Copy/Paste`; `enableAutoPaste` → `requestAccessibility` (api.ts:140) | ✓ WIRED (via api.ts; tool regex misses the indirection) |
| actions.go | platform_darwin.m | `CopyFile` → `gsCopyFile`; `AccessibilityTrusted(false)` → `PasteShortcut` → `gsPostCmdV` | ✓ WIRED |
| shell.go | platform_darwin.m | `FocusWebView`, `CursorScreenWorkArea`, `FrontmostApp`, `ActivateApp` | ✓ WIRED |
| shell.go | position.go | `panelOrigin(wa, PanelWidth, PanelHeight)` | ✓ WIRED |
| main.go | shell.go / actions.go | `shell.New(app, plat, trayIcon)`; `actions.Hide = sh.Hide`; `actions.Previous = sh.Previous`; hotkey `"open": sh.Toggle` | ✓ WIRED |
| Taskfile release | notarytool profile | `scripts/notarize.sh "{{.NOTARY_PROFILE}}" …` | ✓ WIRED (was deferred) |
| 01-PASTE-MATRIX.md | PROJECT.md | Decision copied | ✓ WIRED |

`gsd-tools verify.key-links` (today): 01-01 3/4, 01-02 3/3, 01-03 3/4, 01-04 2/2. Both misses are the App.svelte → actions.go regex, now satisfied through `frontend/src/lib/api.ts`.

### Data-Flow Trace (Level 4)

| Artifact | Data | Source | Real Data | Status |
|----------|------|--------|-----------|--------|
| Pasteboard write | staged path | `ActionService.stage` → `staging.Store` download of the selected provider GIF (allowlisted) | Yes (the spike's fixed sample is gone) | ✓ FLOWING |
| Paste target | `prev` | `FrontmostApp()` at `show()` → `Previous()` | Yes | ✓ FLOWING |
| Status line | `status` | `Result.Message` from bound calls | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Go suite with race detector | `go test -race -count=1 ./internal/...` (single full run) | all 14 packages ok | ✓ PASS |
| Phase 1 named tests | `go test -v -run '<17 Phase 1 tests>' ./internal/platform ./internal/services ./internal/shell` | 17/17 PASS | ✓ PASS |
| Frontend unit tests | `node --test src/lib/*.test.ts` | 103 pass, 0 fail | ✓ PASS |
| Installed app identity | `plutil`, `codesign -d -r-` on /Applications/Gifsmith.app | 1.0.0; identifier + team DR | ✓ PASS |
| Notarization | `spctl -a -vv -t exec`, `xcrun stapler validate` | `source=Notarized Developer ID`; "The validate action worked!" | ✓ PASS (was deferred) |
| Published release | `gh release view v1.0.0` | not draft, published 2026-10-20T19:03:44Z, DMG + .sha256 | ✓ PASS |
| Browser smoke | `bash scripts/smoke-browser-loop.sh` | not run (port conflict with concurrent verifiers); SMOKE OK per 04-VERIFICATION.md:263 | ? SKIP (covered by prior run) |

No `go build ./...`, app launch, synthetic input or Application Support / Downloads access was performed.

### Probe Execution

No `scripts/*/tests/probe-*.sh` exist, and no PLAN declares a probe. Not applicable.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|-------------|--------|----------|
| SHEL-01 | 01-01, 01-02, 01-04 | Floating panel on a global hotkey (⌃⌥⌘G) | ✓ SATISFIED | SC1; now rebindable (Phase 4) with ⌃⌥⌘G default |
| SHEL-02 | 01-01, 01-03, 01-04 | Over full-screen apps, takes typing, doesn't steal focus | ✓ SATISFIED | N2-N4 dev; N4 notarized; panel options unchanged |
| SHEL-03 | 01-01, 01-04 | Closes on Esc and click-away | ✓ SATISFIED | N5, N6; Esc now page-owned (clears the query first, then hides) |
| SHEL-04 | 01-02, 01-04 | Menu bar, no Dock; menu has Open, Settings, Quit | ✓ SATISFIED | N7, N8; Settings real window (A1); Quit observed (Phase 3) |
| ACTN-01 | 01-01, 01-03, 01-04 | Copy GIF as a file, Enter, no Accessibility | ✓ SATISFIED | SC3; N11 dev + notarized |
| ACTN-02 | 01-03, 01-04 | Paste into previous app; without AX copy + instructions | ✓ SATISFIED | SC4; N12 dev + notarized, N14 dev |

All 6 IDs are in plan frontmatter and marked Complete in REQUIREMENTS.md; none orphaned. RELS-01/RELS-02, which SC5 de-risked, are now Complete under Phase 4 (REQUIREMENTS.md:151-152).

### Prohibitions (judgment-tier, re-checked against current code)

The 10 judgment-tier prohibitions from the first report were resolved at the 2026-10-17 human checkpoint (status set to passed). Re-checked against today's code, there is still no violation: `AccessibilityTrusted(true)` is called only in `RequestAccessibility` (actions.go:397), which runs only from explicit buttons (panel "Enable Automatic Paste…", Settings "Grant Access…"); `Copy` makes no AX call; `PasteShortcut()` has one call site, after hide; `git grep` finds no app-specific password or notary secret outside docs (only the profile name `gifsmith-notary`); the Taskfile only runs `pkill -x gifsmith`.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| Phase 1 covered Go/Taskfile files | | TBD/FIXME/XXX | none found | No debt-marker blockers |
| frontend/src/App.svelte | 293 | `panel:status` listener with no emitter | ℹ️ Info | Leftover of the superseded P5 placeholder |
| internal/shell/position.go | | `shortcut:` no clamping on narrow displays | ℹ️ Info | Documented shortcut, unchanged |
| 01-PASTE-MATRIX.md | 46-52 | PENDING rows | ⚠️ Warning | The SC5 gap |

### Human Verification Required

None. Every human item in the first report has been resolved by later evidence (see `re_verification.gaps_closed`). The remaining item is a gap that needs a user decision (test or override), not a verification step.

### Gaps Summary

**One gap, no code defect.** ROADMAP SC5 asks for a committed paste matrix with a result per target app. Five targets (Slack, Discord, Telegram, Safari composer, Chrome composer) have never been tested (01-PASTE-MATRIX.md:46-50). They were deferred to Phase 4 under GO WITH FALLBACKS. Phase 4 finished and v1.0.0 shipped without them, and the v1.0 milestone audit lists them as tech debt for 1.0.1. With no later phase to absorb them, they count as a gap here, not a deferral.

Everything else Phase 1 set out to prove still holds after Phase 4, and is better evidenced than at the gate. The notarized half of SC5, N13, N15, the tray Settings…/Quit behaviors and the final-build regression are all now covered.

**This looks intentional.** If you accept shipping without these rows (as the milestone audit implies), add this to the frontmatter and the status becomes `passed`:

```yaml
overrides:
  - must_have: "ROADMAP SC5: a committed paste matrix records the result per target app (Slack, Messages, Mail, Discord, Notes, Telegram, Safari/Chrome composers)"
    reason: "GO WITH FALLBACKS (2026-10-17) limit 2; v1.0.0 shipped with file-URL paste verified in TextEdit, Notes, Mail, Messages; remaining apps tracked as 1.0.1 tech debt in v1.0-MILESTONE-AUDIT.md"
    accepted_by: "{name}"
    accepted_at: "{ISO timestamp}"
```

---

_Verified: 2026-10-20T19:21:04Z (re-verification after Phase 4; first verified 2026-10-17T17:00:11Z)_
_Verifier: Claude (gsd-verifier)_
