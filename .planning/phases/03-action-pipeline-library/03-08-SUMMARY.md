---
phase: 03-action-pipeline-library
plan: 08
subsystem: testing
tags: [regression, uat, dev-browser, codesign, install-dev, native-uat]

requires:
  - phase: 03-action-pipeline-library
    provides: "plans 03-01..03-07: action pipeline, ⌘K menu, download/reveal, square crop, Clips MP4, library, details, recent-search chips"
provides:
  - "Full Phase 3 regression record on one tree (Go, lint, server build, node, svelte-check, smoke, 18 UAT runs)"
  - "Developer ID-signed dev build installed at ~/Applications/Gifsmith.app (not notarized)"
  - "User's native UAT approval, with the native-evidence gap recorded for the phase verifier"
affects: [phase-03-verification, phase-04-release]

actuals:
  tokens: 2990
  tasks: 2
  commits: 0
plan_head_before: c814eff36232c1ffffd97bf4f869c548f891fb07
plan_head_after: c814eff36232c1ffffd97bf4f869c548f891fb07

tech-stack:
  added: []
  patterns: []

key-files:
  created:
    - .planning/phases/03-action-pipeline-library/03-08-SUMMARY.md
  modified: []

key-decisions:
  - "Native UAT accepted on the user's attestation ('approved'; 'Already did these before. We're good to go.'); steps without native log/file evidence on this install are flagged for the phase verifier, not re-run"
  - "Clips MP4 paste (step 8) skipped: GIPHY key not Clips-approved; stays fixture-proven only (D-15, A1)"

patterns-established:
  - "12-library-unreadable.js must run with UAT_SEED_FAVORITES='{not json}'; any full-suite UAT loop has to special-case it like the FAKE_CLIPS=denied scripts"

requirements-completed: [ACTN-03, ACTN-04, ACTN-05, ACTN-06, ACTN-07, ACTN-09, ACTN-10, ACTN-11, LIBR-01, LIBR-02, LIBR-03, LIBR-04, LIBR-05, SRCH-08]

coverage:
  - id: D1
    description: "Full automated suite green on one tree (c814eff)"
    verification:
      - kind: unit
        ref: "go test -race ./internal/... -count=1 (13 pkgs, 265 cases)"
        status: pass
      - kind: unit
        ref: "pnpm --dir frontend run test (63 pass / 0 fail)"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh (SMOKE OK) + 18 scripts/uat/*.js runs via --uat (UAT OK each)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Signed dev build installed and running; log shows tray menu ready and hotkey registered; no key material in bin/install-dev.log"
    verification:
      - kind: other
        ref: "wails3 task darwin:install:dev; grep 'tray menu ready|hotkey registered' bin/install-dev.log; leak grep"
        status: pass
    human_judgment: false
  - id: D3
    description: "Native behaviors on the installed app (steps 1-13: text actions, Paste Link, ⌘O, download/reveal, Paste Square, Klipy paste timing, library, native text editing keys, Details, chips, persistence across Quit)"
    verification:
      - kind: manual_procedural
        ref: "user reply 'approved' + 'Already did these before. We're good to go.'"
        status: pass
    human_judgment: true
    rationale: "Real pasteboard targets, Finder, browser, WebKit key equivalents and Quit/relaunch persistence cannot be exercised headlessly; only download/reveal has native log evidence on this install (see Native evidence gap)"

duration: ~10 min active (plus checkpoint wait 2026-10-17 21:19 to 2026-10-18)
completed: 2026-10-18
status: complete
---

# Phase 3 Plan 08: Phase Regression and Native UAT Summary

**Whole-phase suite green on c814eff (265 Go cases, 63 node tests, SMOKE OK, 18/18 UAT), Developer ID-signed dev build installed, and the user approved the 13-step native UAT; library persistence and ⌘Q staging cleanup lack native evidence on this install and are flagged for the verifier.**

## Performance

- **Duration:** ~10 min active for Task 1; the Task 2 checkpoint waited on the user from 2026-10-17 21:19 to 2026-10-18
- **Started:** 2026-10-18T02:15:14Z (2026-10-17 21:15 -05:00)
- **Completed:** 2026-10-18T16:48:49Z
- **Tasks:** 2/2
- **Files modified:** 1 (this SUMMARY)

## Accomplishments

- Ran every automated check together on one clean tree; all green.
- Installed the Developer ID-signed (hardened runtime) dev build; it launched and logged "tray menu ready" and "hotkey registered".
- The user approved the native UAT. The post-session read-only checks and the evidence gap are recorded below.

## Task 1: Full-suite regression and signed install (no commit, verification only)

| Check | Result |
|-------|--------|
| `gofmt -l main.go wire_desktop.go wire_server.go internal` | clean (no files listed) |
| `go vet . ./internal/...` | OK |
| `go test -race ./internal/... -count=1` | OK: 13 packages, 265 cases |
| `golangci-lint run . ./internal/...` | 0 issues |
| `CGO_ENABLED=0 go build -tags server .` | OK |
| `pnpm --dir frontend run test` | 63 pass / 0 fail |
| `pnpm --dir frontend run check` (svelte-check) | 0 errors / 0 warnings |
| `bash scripts/smoke-browser-loop.sh` | SMOKE OK |

UAT, each via `bash scripts/smoke-browser-loop.sh --uat <script>` (logs in the session scratchpad `uat-0308/`):

| Script | Env | Result |
|--------|-----|--------|
| 01-tracer.js | | UAT OK |
| 02-grid.js | | UAT OK (mounted 20 of 60, 60 distinct tiles seen) |
| 02-keyboard.js | | UAT OK |
| 03-states.js | | UAT OK |
| 05-setup.js | | UAT OK |
| 05-tabs.js | | UAT OK |
| 06-backstops.js | FAKE_CLIPS=denied | UAT OK (contrast ratios all >= 4.54) |
| 06-clips-denied.js | FAKE_CLIPS=denied | UAT OK |
| 06-clips.js | | UAT OK |
| 07-text-actions.js | | UAT OK |
| 08-action-menu.js | | UAT OK |
| 10-download.js | | UAT OK |
| 11-square.js | | UAT OK |
| 12-library.js | | UAT OK |
| 12-library-unreadable.js | UAT_SEED_FAVORITES='{not json}' | UAT OK (failed without the seed, see Deviations) |
| 13-chips.js | | UAT OK |
| 13-details.js | | UAT OK |
| 14-backstops.js | FAKE_CLIPS=denied | UAT OK (P8 widest 492.5px; contrast all >= 5.38) |

Install: `wails3 task darwin:install:dev` exited 0. The app was codesigned with "Developer ID Application: Ken Scott (S2K7P43927)" with the runtime flag, not notarized (Apple notarization blocker unchanged). `bin/install-dev.log` has 1 "tray menu ready" and 1 "hotkey registered". The go.sum objx lines added by `go mod tidy` were restored (`git diff --quiet -- go.sum` exits 0).

## Task 2: Native UAT (checkpoint:human-verify)

**User reply:** "approved". When the orchestrator pointed out the missing log evidence, the user added: "Already did these before. We're good to go."

| Step | Outcome | Evidence on this install |
|------|---------|--------------------------|
| 1 ⌘K + Copy Markdown, no query | user-attested | no native log/file evidence on this install |
| 2 ⌘⇧C Copy Link | user-attested | none |
| 3 Paste Link into TextEdit | user-attested | none |
| 4 ⌘O opens the GIPHY page | user-attested | none |
| 5 ⌘S twice, then ⌘R | passed | `bin/install-dev.log`: `download giphy: saved` x3 (2026-10-18 11:44:35, 11:44:39, 11:44:58) and `show download: revealed` (11:44:58). No Downloads permission prompt was reported. |
| 6 Paste Square GIF into Messages | user-attested | none (the log has no `copy: staged and copied` or paste lines) |
| 7 Klipy paste into Messages, then Slack after ~40 s | user-attested | none. The Klipy 30 s grace delete is fixture-proven only |
| 8 Clips MP4 into Messages | **skipped: GIPHY key not Clips-approved** | fixture-proven only (D-15, A1) |
| 9 Favorites / Recents / ⌘⌫ / ⌘Z | user-attested | none. No favorites.json or recents.json exists |
| 10 ⌘⌫ and ⌘Z edit text natively | user-attested | n/a (A4 not contradicted) |
| 11 ⌘I Details + page link | user-attested | none |
| 12 Recent-search chip reruns | user-attested | none. No searches.json exists |
| 13 Quit, relaunch, library persists | user-attested | none. The app process was not restarted (see below) |

The plan's how-to-verify has 13 numbered steps. The orchestrator's "16 steps A–E" refers to its own regrouping of the same checks. Every user-attested row means: user-attested, no native log/file evidence on this install.

### Post-session read-only checks (2026-10-18)

- `ls -A ~/Library/Caches/com.kenscott.gifsmith/staging`: no such directory (all of `~/Library/Caches/com.kenscott.gifsmith/` is absent). Nothing is staged, and no staging copy happened on this install.
- `stat -f %Lp` on the library files: `recents.json` and `favorites.json` **do not exist**. `~/Library/Application Support/com.kenscott.gifsmith/` holds only `settings.json` (mode 600). The 0600 mode of the library files is therefore only covered by unit tests on this install.
- Leak grep `! grep -Eq 'api_key=|api\.klipy\.com/api/v1/[^/]+/' bin/install-dev.log`: **PASS**.
- Gifsmith PID 92297 started at 2026-10-17 21:19:40 and was still running. No Quit or relaunch happened after this install, so ⌘Q staging cleanup (D-17) was not observed. The executor did not quit or relaunch the app.

## Native evidence gap (flag for the phase verifier)

The user's attestation is recorded as given. On this install, the only natively evidenced step is 5 (download x3, reveal). The verifier should treat these as unproven natively:

1. **Library persistence (LIBR-01..04, SRCH-08):** no favorites.json, recents.json or searches.json in Application Support. Steps 9, 12 and 13 left no files, so either they ran on an earlier install/config or they were not run against this one. Headless coverage: 12-library.js, 12-library-unreadable.js, 13-chips.js and the library unit tests.
2. **⌘Q staging cleanup (ACTN-10, D-17):** the installed process has been running since install, so Quit and relaunch were not observed. Headless/unit coverage: `stage.Clear()` at launch, `ServiceShutdown`.
3. **Klipy 30 s grace delete (ACTN-10, D-19, A2):** fixture-proven only. No Klipy copy or paste lines appear in the log.
4. **Clips MP4 paste (ACTN-09, A1):** fixture-proven only. Step 8 was skipped because the key is not Clips-approved.
5. **Downloads TCC prompt text (A3):** not reported either way.

## Task Commits

None. Task 1 was verification-only and Task 2 was a human checkpoint. `git rev-list --count c814eff..HEAD` = 0 at SUMMARY time. The plan metadata commit follows this file.

## Files Created/Modified

- `.planning/phases/03-action-pipeline-library/03-08-SUMMARY.md`: this record

## Decisions Made

- The native UAT was accepted on the user's attestation and not re-run. The evidence gap is recorded for the verifier and not treated as a failure.
- Step 8 (Clips) was skipped as the plan anticipated.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] The UAT loop needed the seeded-corrupt-favorites case**
- **Found during:** Task 1
- **Issue:** The plan's verify loop runs every `scripts/uat/*.js` with only the FAKE_CLIPS special case. `12-library-unreadable.js` requires `UAT_SEED_FAVORITES='{not json}'` (documented in the script header). Without it, the script timed out waiting for "Couldn't load Favorites" and printed `UAT FAIL`.
- **Fix:** Re-ran that script with the seed. Result: UAT OK, and the harness confirmed the file bytes were unchanged. No code change.
- **Files modified:** none
- **Verification:** `uat-0308/12-library-unreadable.seeded.log` ends `UAT OK`
- **Committed in:** n/a

---

**Total deviations:** 1 auto-fixed (Rule 3, harness invocation only)
**Impact on plan:** None on the product. Future full-suite loops must include the seed case.

## Issues Encountered

- `install:dev`'s `go mod tidy` added stretchr/objx lines to go.sum, as expected. They were restored before any commit.

## User Setup Required

None.

## Next Phase Readiness

- Phase 3 now has 8/8 plans summarized and is ready for `/gsd-verify-work` / phase verification. The verifier should weigh the Native evidence gap above, especially library persistence and ⌘Q cleanup.
- The Apple notarization blocker (403) is unchanged. The Klipy clearance gate remains a Phase 4 pre-release item.

---
*Phase: 03-action-pipeline-library*
*Completed: 2026-10-18*

## Self-Check: PASSED

- FOUND: .planning/phases/03-action-pipeline-library/03-08-SUMMARY.md
- FOUND: plan_head_before c814eff is HEAD (0 code commits, as expected for a verification plan)
- FOUND: bin/install-dev.log with tray menu ready and hotkey registered; leak grep PASS; go.sum clean
