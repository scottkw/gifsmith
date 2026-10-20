---
phase: 01-native-spike-signed-dev-loop
plan: 04
subsystem: native-shell
tags: [go-no-go, paste-matrix, notarization-deferred, developer-id, pasteboard, wails-v3]

requires:
  - phase: 01-03
    provides: "Developer ID-signed dev install, ⌘Enter paste, 01-PASTE-MATRIX.md signed-dev columns, PasteDelay 80 ms"
provides:
  - "Recorded Phase 1 decision: GO WITH FALLBACKS (user, 2026-10-17), with four limits in 01-PASTE-MATRIX.md §4"
  - "Derived payload conclusion: single public.file-url is sufficient for TextEdit, Notes, Mail, Messages; other apps unverified"
  - "PROJECT.md Key Decisions: spike outcome, Wails v3.0.0-beta.30 pin, file-URL payload, notarization deferral"
  - "ROADMAP Phase 4 note: notarized dry run + N15 carried over from Phase 1"
affects: [Phase 2 (go granted), Phase 3 pasteboard design note (untested apps), Phase 4 release (notarization, N15, Accessibility deep link)]

actuals:
  tokens: 1391   # chars/4 over git diff aaab9a6..81e84f9 (matrix + PROJECT.md); Task 1 not executed
  tasks: 2       # Task 2 resolved by the user, Task 3 executed; Task 1 deferred to Phase 4
  commits: 1
plan_head_before: aaab9a6e06195fd28745e2ed1fc64bee7be73a9e
plan_head_after: 81e84f993dfca0a6917ff4dab4afee0f00f77cbd

tech-stack:
  added: []
  patterns:
    - "Go/no-go is a committed line in the paste matrix plus a PROJECT.md Key Decisions outcome, not an impression"

key-files:
  created: []
  modified:
    - .planning/phases/01-native-spike-signed-dev-loop/01-PASTE-MATRIX.md
    - .planning/PROJECT.md
    - .planning/ROADMAP.md

key-decisions:
  - "Phase 1 GO WITH FALLBACKS (user): build Phase 2 on the Developer ID-signed dev build"
  - "Notarization deferred to Phase 4: notarytool store-credentials returns 403 'A required agreement is missing or has expired'; the user must fix the Apple Developer agreement first"
  - "Pasteboard payload stays a single file URL; Phase 3 decides on image data only if Slack/Discord/Telegram/browser composers fail"
  - "Phase 4 onboarding/Settings deep-links to Privacy_Accessibility; macOS 27.0.1 has no browsable Accessibility row"

patterns-established:
  - "Deferred external-account work is recorded in the matrix header, PROJECT.md Key Decisions, ROADMAP target phase and STATE blockers together"

requirements-completed: [SHEL-01, SHEL-02, SHEL-03, SHEL-04, ACTN-01, ACTN-02]

coverage:
  - id: D1
    description: "01-PASTE-MATRIX.md has exactly one Decision line (GO WITH FALLBACKS) with the user's limits, date and commit"
    verification:
      - kind: other
        ref: "grep -cE '^Decision: (GO|GO WITH FALLBACKS|NO-GO)$' 01-PASTE-MATRIX.md → 1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Payload conclusion derived from the user's signed-dev results and labelled as derived"
    verification:
      - kind: other
        ref: "grep -q 'Derived from the user' 01-PASTE-MATRIX.md"
        status: pass
    human_judgment: false
  - id: D3
    description: "PROJECT.md Key Decisions updated by scoped edit (spike outcome, Wails pin, payload, notarization deferral); all section headings kept"
    verification:
      - kind: other
        ref: "grep v3.0.0-beta.30 PROJECT.md; 7 section headings present; 'Spike first … — Pending' count 0"
        status: pass
    human_judgment: false
  - id: D4
    description: "Notarized build passes Gatekeeper and carries the Accessibility grant (N15, notarized columns)"
    verification: []
    human_judgment: true
    rationale: "Not executed: deferred to Phase 4 by the user (Apple agreement 403). Must be verified in Phase 4 release work"

duration: 3min
completed: 2026-10-17
status: complete
---

# Phase 1 Plan 04: Go/No-Go and Paste Matrix Close-out Summary

**The user recorded GO WITH FALLBACKS for Phase 1 on the Developer ID-signed dev build. The paste matrix now holds the decision, its four limits, and the derived conclusion that one file URL on the pasteboard pastes animated GIFs into TextEdit, Notes, Mail and Messages. The notarized dry run moves to Phase 4 because Apple's notary service returns a 403 for a missing agreement.**

## Performance

- **Duration:** about 3 min executor time (the decision and the notarization attempts happened earlier in the orchestrator conversation)
- **Started:** 2026-10-17T16:36:17Z
- **Completed:** 2026-10-17T16:37Z
- **Tasks:** 2 of 3 (Task 1 deferred, Task 2 resolved by the user, Task 3 executed)
- **Files modified:** 3 (matrix, PROJECT.md, ROADMAP.md) plus the STATE/REQUIREMENTS updates

## Accomplishments

- `01-PASTE-MATRIX.md` §4: `Decision: GO WITH FALLBACKS`, then the user's limits, date 2026-10-17 and commit aaab9a6.
- §3 Payload conclusion, derived from the user's results: a single `public.file-url` is enough for TextEdit, Notes, Mail and Messages (Copy→⌘V and ⌘Enter, animated). Slack, Discord, Telegram and the Safari/Chrome composers are still unverified. PasteDelay is 80 ms. The D-08 activation fallback, PopUpMenu level, mouse monitor and KeyBindings fallback were all unneeded.
- The header and a section note mark every Notarized cell as deferred to Phase 4 (Apple agreement 403).
- PROJECT.md Key Decisions: the "Spike first…" outcome is now ✓ Good (GO WITH FALLBACKS, 2026-10-17). New rows: Wails v3.0.0-beta.30 pin, single-file-URL payload, and "Notarization deferred to Phase 4 (Apple agreement 403)".
- ROADMAP Phase 4 has a "Carried over from Phase 1" line for the notarized dry run, N15 and the notarized columns.

## Decision (verbatim from the matrix)

```
Decision: GO WITH FALLBACKS
```
1. The notarized dry run (de-risking RELS-01/RELS-02; N15 and the notarized columns) is deferred to Phase 4. It is blocked on the Apple Developer account agreement 403, which the user must resolve (check membership expiry and App Store Connect agreements) before Phase 4 release work.
2. Paste in the untested apps (Slack, Discord, Telegram, Safari composer, Chrome composer), the "Enable automatic paste…" button click (N13), and N10 on the signed build are re-checked in later UAT (Phase 3 / Phase 4).
3. On macOS 27.0.1 the Accessibility list can't be reached by browsing System Settings › Privacy & Security, only via `x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility`. Phase 4 must deep-link there.
4. Payload: a single file URL is sufficient for TextEdit, Notes, Mail and Messages. Other apps are unverified. PasteDelay 80 ms, no fallbacks needed.

Execution did not stop: the decision is a GO variant, so Phase 2 may start.

## Task Commits

1. **Task 1: Tracer: notarize, staple, Gatekeeper-validate, install.** DEFERRED TO PHASE 4 by the user. `xcrun notarytool store-credentials gifsmith-notary --apple-id … --team-id S2K7P43927` fails with "HTTP status code: 403. A required agreement is missing or has expired". The user had accepted the developer.apple.com agreement banner and retried about 30 minutes later with the same result. `xcrun notarytool history --keychain-profile gifsmith-notary` fails, so the profile does not exist. Nothing was submitted to Apple, there is no submission ID and no spctl/stapler output, and `build/darwin/Taskfile.yml` was not changed.
2. **Task 2: Go/no-go checkpoint.** RESOLVED by the user: GO WITH FALLBACKS. No commit.
3. **Task 3: Record decision, payload conclusion, PROJECT.md.** `81e84f9` (docs)

**Plan metadata:** final docs commit (SUMMARY, STATE, ROADMAP, REQUIREMENTS)

## Files Created/Modified

- `.planning/phases/01-native-spike-signed-dev-loop/01-PASTE-MATRIX.md`: header deferral note, notarized-columns note, §3 payload conclusion, §4 Decision
- `.planning/PROJECT.md`: Key Decisions, one outcome updated and three rows added (scoped edit)
- `.planning/ROADMAP.md`: Phase 4 carry-over line. The Research flag sentence was corrected so it no longer claims the Phase 1 dry run happened

## Decisions Made

See `key-decisions`. The one with the most impact on later work: Phase 4 starts with notarization still unproven, and it cannot run until the user fixes the Apple Developer agreement.

## Deviations from Plan

**1. [User decision] Task 1 not executed.** The notarized build, the `KEYCHAIN_PROFILE` / `--keychain-profile` Taskfile change, `bin/dr-dev.txt` and the install of a notarized build were all skipped. The Taskfile change was left for Phase 4, together with the run that will prove it. Not met: the plan's must_have truths 1–4 (notarize/staple, spctl Notarized Developer ID, unchanged designated requirement, notarized build running with notarized columns filled) and success criterion "SHEL-01..04, ACTN-01, ACTN-02 re-verified on the notarized build".

**2. [User decision] Acceptance criterion adjusted: "no unfilled cells".** Task 3's verify expects `grep -c PENDING` = 0. As the user directed, 27 PENDING cells remain: all Notarized cells (deferred to Phase 4), N10 and N13 in the dev column, and the five untested apps in the dev columns. Each one has a reason recorded in the matrix. The D-12 completeness STOP was not triggered, because the user resolved the checkpoint with these cells open. Every other Task 3 check passes: exactly one `Decision:` line, the "Derived from the user" label, the beta.30 pin row in PROJECT.md, all seven section headings, and the "Spike first" row no longer "— Pending".

**3. [Rule 1 - Correctness] ROADMAP Phase 4 research flag.** It said "The Phase 1 notarization dry run means release is a gate, not a discovery", which is no longer true. It now says the dry run was deferred there.

**Total deviations:** 2 user-directed, 1 auto-fix (doc correctness). **Impact:** Phase 1's notarization de-risk (roadmap success criterion 5, notarized half) moves to Phase 4. The native architecture decision is recorded.

## Issues Encountered

- Apple notary service 403 "A required agreement is missing or has expired" (found before this executor ran). Not an executor issue. The user has to resolve it in the Apple Developer / App Store Connect account.

## Known Stubs

None. No code changed.

## User Setup Required

Before Phase 4 release work: resolve the Apple Developer agreement 403 (check membership expiry and pending App Store Connect agreements), then run `xcrun notarytool store-credentials gifsmith-notary --apple-id … --team-id S2K7P43927`.

## Next Phase Readiness

- Phase 1 plans complete. Decision: GO WITH FALLBACKS, so Phase 2 may start.
- Carry-overs: notarized dry run + N15 (Phase 4); untested-app paste, N10, N13 (later UAT); Accessibility deep link (Phase 4 onboarding).

---
*Phase: 01-native-spike-signed-dev-loop*
*Completed: 2026-10-17*

## Self-Check: PASSED

- FOUND: 01-PASTE-MATRIX.md, PROJECT.md, ROADMAP.md edits
- FOUND commit: 81e84f9 (ancestor of HEAD)
