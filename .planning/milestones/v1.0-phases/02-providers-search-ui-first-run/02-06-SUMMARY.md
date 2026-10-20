---
phase: 02-providers-search-ui-first-run
plan: 06
subsystem: GIPHY Clips gating, compliance record, native phase UAT
tags: [go, wails-v3, svelte5, giphy-clips, probe, compliance, dev-browser, tdd, native-uat]

requires:
  - phase: 02-providers-search-ui-first-run
    provides: "02-04 SettingsService.SaveKey, GifService.Providers, providers:changed, search.Service rate-limit blocking, fakeapi GIPHY routes; 02-05 ProviderTabs, C1 + Clips S2 skeleton in App.svelte, openExternal allow-list, StateMessage"
provides:
  - "provider.ClipsDisabled error kind"
  - "giphy.Client Clips mode (/v1/clips/search|trending, rating pg-13, bundle clips_grid_picker, isClip, no download URL) and ProbeClips"
  - "search.Clips state machine: no_key | checking | enabled | disabled | unknown; one probe per launch, skipped while GIPHY is 429-blocked, Recheck only when unknown, Reset on GIPHY key change; search.Service.Blocked"
  - "GifService.CheckClips (bound); Providers() reports the Clips state; Clips Search/Trending refused (clips_disabled, no request) unless enabled; a GIPHY GIFs success triggers the probe"
  - "SettingsService: GIPHY SaveKey success -> Clips.Reset + Trigger"
  - "Panel: Clips pill follows the state; C2 card with the clips@giphy.com mailto and Back to GIPHY; Clips grid; E8 on Enter for a clip"
  - "PROJECT.md Key Decisions: Phase 2 compliance posture and the Keychain decision"
  - "fakeapi Clips routes + FAKE_CLIPS=denied; scripts/uat/06-clips.js, 06-clips-denied.js, 06-backstops.js"
affects: [phase-03, phase-04]

actuals:
  tokens: 33097   # chars/4 over the final content of the 22 changed files (132,389 chars)
  tasks: 3
  commits: 3
plan_head_before: 11801116f65b2273336a14e94d7ade0fa6e5273b
plan_head_after: 643274971fd69e769580547fbf1936ef1b251819

tech-stack:
  added: []
  patterns:
    - "Capability probe gated on a prior success with the same key: a Clips 401/403/404 means 'not approved' only after GIPHY GIFs already worked"
    - "Go refuses calls for a capability the UI shows as unavailable (typed error, zero requests), so the UI is not the only guard"
    - "A single in-flight guard (rechecking) prevents double re-probes when Clips is chosen repeatedly while unknown"

key-files:
  created:
    - internal/search/clips.go
    - internal/search/clips_test.go
    - scripts/uat/06-clips.js
    - scripts/uat/06-clips-denied.js
    - scripts/uat/06-backstops.js
  modified:
    - internal/provider/provider.go
    - internal/provider/giphy/giphy.go
    - internal/provider/giphy/giphy_test.go
    - internal/provider/fakeapi/fakeapi.go
    - internal/search/search.go
    - internal/services/gifs.go
    - internal/services/settings.go
    - internal/services/settings_test.go
    - main.go
    - wire_server.go
    - frontend/bindings/gifsmith/internal/services/gifservice.ts
    - frontend/src/App.svelte
    - frontend/src/components/StateMessage.svelte
    - frontend/src/lib/api.ts
    - frontend/src/lib/providers.ts
    - scripts/smoke-browser-loop.sh
    - .planning/PROJECT.md

key-decisions:
  - "C2 primary action is an <a> styled as the primary button, labelled 'Email clips@giphy.com ↗', opening mailto:clips@giphy.com with a subject (the only mailto in the openExternal allow-list)"
  - "Choosing Clips while unknown re-probes once through CheckClips, guarded by a 'rechecking' flag; Try Again on a Clips error state re-probes the same way"
  - "Recheck while GIPHY is rate-limit blocked makes no request and stays unknown"
  - "ProviderTabs.svelte needed no change: it already dims Clips whenever the status is not 'enabled'"
  - "A remembered Clips tab opens on GIPHY at launch until the probe enables Clips (02-05 rule kept)"
  - "Native UAT: the user's real GIPHY key is NOT approved for Clips; the probe went checking -> disabled and the panel shows C2"

patterns-established:
  - "Capability states that depend on a network probe log only the state name (main.go 'clips: <state>'), never the URL or key"

requirements-completed: [SRCH-01, SRCH-02, SRCH-05, ONBD-01, ONBD-02, ONBD-03, ONBD-04, ONBD-05]

coverage:
  - id: D1
    description: "giphy Clips mode mapping (clips endpoints, pg-13, clips_grid_picker, isClip, no download URL, fixed_width preview) and ProbeClips classification (200 nil; 401/403 key_rejected; 404 http; 429 rate_limited)"
    requirement: SRCH-02
    verification:
      - kind: unit
        ref: "internal/provider/giphy/giphy_test.go#TestClipsMode, #TestProbeClips"
        status: pass
    human_judgment: false
  - id: D2
    description: "Clips state machine: no_key without a key, checking -> enabled with OnChange per change, once per launch, outcomes (disabled for key_rejected/404, unknown for 429/network/timeout/other), blocked skip, Recheck only when unknown, Reset re-probes, stale probe dropped after Reset"
    requirement: SRCH-02
    verification:
      - kind: unit
        ref: "internal/search/clips_test.go#TestClipsNoKey, TestClipsTriggerOnce, TestClipsOutcomes, TestClipsBlockedSkip, TestClipsRecheck, TestClipsReset, TestClipsResetDropsStaleProbe"
        status: pass
    human_judgment: false
  - id: D3
    description: "Services: Clips refused unless enabled with zero requests, GIPHY success triggers one probe, GIPHY SaveKey resets and triggers, Providers() reports checking during a slow probe, denied fixture -> disabled, CheckClips returns the error while unknown"
    requirement: SRCH-02
    verification:
      - kind: unit
        ref: "internal/services/settings_test.go#TestProvidersClipsChecking, TestSaveKeyTriggersClips, TestGiphySuccessTriggersClipsOnce, TestClipsRefusedUnlessEnabled, TestClipsDeniedFixture, TestCheckClipsUnknown"
        status: pass
    human_judgment: false
  - id: D4
    description: "Browser: Clips pill enables after a GIPHY key, Clips grid with the GIPHY mark, Enter on a clip shows E8; denied fixture shows C2 with the email button and Back to GIPHY returns to the GIPHY grid"
    requirement: SRCH-02
    verification:
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/06-clips.js; FAKE_CLIPS=denied bash scripts/smoke-browser-loop.sh --uat scripts/uat/06-clips-denied.js"
        status: pass
    human_judgment: false
  - id: D5
    description: "Backstops: E5f, C2 and K6 fit the content area at 720x480, and text-secondary/accent/destructive meet 4.5:1, in light and dark"
    requirement: SRCH-05
    verification:
      - kind: automated_ui
        ref: "FAKE_CLIPS=denied bash scripts/smoke-browser-loop.sh --uat scripts/uat/06-backstops.js"
        status: pass
    human_judgment: false
  - id: D6
    description: "PROJECT.md Key Decisions holds the Phase 2 compliance posture and the Keychain decision"
    requirement: ONBD-04
    verification:
      - kind: other
        ref: "grep -n 'compliance posture' .planning/PROJECT.md"
        status: pass
    human_judgment: false
  - id: D7
    description: "Native UAT on the signed dev build with real keys: GIPHY key pasted with Cmd+V and saved, trending/search, Enter copies an animated GIF, Clips outcome, Klipy key and search with KLIPY attribution, Finer Gifs, Esc/Cmd+W hide, keys and last tab survive Quit and relaunch"
    requirement: ONBD-01
    verification:
      - kind: manual_procedural
        ref: "02-06 Task 3 checkpoint, user reply 'approved'; bin/install-dev.log 'clips: checking' -> 'clips: disabled'"
        status: pass
    human_judgment: true
    rationale: "Needs desktop input, real vendor keys, the Keychain and real paste targets; performed and approved by the user"
  - id: D8
    description: "No key material in the dev logs after the real-key session"
    verification:
      - kind: other
        ref: "grep -Eq 'api_key=|api\\.klipy\\.com/api/v1/[^/]+/' bin/install-dev.log bin/gifsmith-run.log (exit 1 for both)"
        status: pass
    human_judgment: false

duration: 23min
completed: 2026-10-17
status: complete
---

# Phase 2 Plan 06: GIPHY Clips Gating, Compliance Record and Native UAT Summary

**GIPHY Clips now appears only when the user's own GIPHY key passes one Clips test call per launch, made after a GIPHY GIFs call with the same key succeeded. Otherwise the C2 card explains how to ask clips@giphy.com for access, and Go refuses Clips calls itself. PROJECT.md records the Phase 2 compliance posture. The user approved the full native UAT on the signed dev build with real keys. Their GIPHY key is not approved for Clips: the probe went `checking -> disabled` and the panel showed C2.**

## Performance

- **Duration:** about 23 min of execution (2026-10-17T20:55Z to 21:18Z, through the signed install), plus the user's native UAT
- **Tasks:** 3 (Task 1 TDD, Task 2 auto, Task 3 human-verify approved)
- **Files modified:** 22 (5 created, 17 modified)

## Accomplishments

- `giphy.Client{Clips: true}` calls `/v1/clips/search|trending` with `rating=pg-13` (always sent, since Clips search defaults to R) and `bundle=clips_grid_picker`. Clip results carry `isClip: true` and no download URL. `ProbeClips` makes the single `GET /v1/clips/trending?limit=1` call.
- `search.Clips` owns the capability state and probes at most once per launch. It never probes while GIPHY is 429-blocked, and it re-probes only on a GIPHY key change or when the user picks Clips while the state is unknown. `OnChange` emits `providers:changed`, and main.go logs only the state name.
- `GifService` refuses Clips Search/Trending with `clips_disabled` unless Clips is enabled, and makes no request. `CheckClips` is bound for the UI.
- The panel: the Clips pill follows the state; checking shows the S2 skeleton; unknown re-probes once; disabled shows C2; enabled shows the clip grid with the GIPHY mark; Enter on a clip shows E8.
- PROJECT.md Key Decisions has the compliance posture and Keychain rows.
- Both UI backstops (long-copy fit and 4.5:1 contrast) pass in light and dark.

## Task Commits

1. **Task 1: GIPHY Clips mode and the once-per-launch Clips capability probe**: `2ab298a` (test, RED), then `1ec299b` (feat, GREEN)
2. **Task 2: Clips states in the panel, compliance posture, UI backstops, signed dev install**: `6432749` (feat)
3. **Task 3: Native UAT with real keys on the installed app**: checkpoint, user reply "approved" (no code commit)

## TDD Gate Compliance

- **Task 1:** RED `2ab298a` → GREEN `1ec299b`. The record is `bin/red-02-06-t1-record.json` (go test -json → TAP, 88 tests). The classifier returned `RED_EVIDENCE_OK` with target `TestClipsTriggerOnce`.
  - **Semantic assessment:** all three packages compiled against skeletons and every test ran without a panic. The target failed on its planned assertions (state unknown instead of checking/enabled, OnChange 0 instead of 2, probes 0 instead of 1). The other 23 new tests failed on stub assertions, and `TestProviders` failed on its updated Clips expectation. **`TestClipsResetDropsStaleProbe` passed early on the skeleton** because the stubs never run a probe, so no stale result can land. It guards against an in-flight probe finishing after `Reset`, and it means something only once probes run. It passes against the GREEN code for the right reason.
- No REFACTOR commit.

## Native UAT (Task 3)

- **Result:** the user replied "approved". All steps of the native checklist passed on the Developer ID-signed `~/Applications/Gifsmith.app`.
- **Clips outcome:** the user did not state it explicitly. The app log `bin/install-dev.log` shows `16:17:00 clips: checking` then `16:17:11 clips: disabled`. The probe therefore got 401, 403 or 404 after the real GIPHY key had already worked for GIFs, so the key is **not approved for Clips** and choosing Clips shows C2 ("Clips aren't enabled for your key" with "Email clips@giphy.com ↗"). The exact status code is not logged, by design.
- **Log locations:** the installed app's stdout/stderr go to `bin/install-dev.log` (written by `wails3 task darwin:install:dev`). `bin/gifsmith-run.log` is an older 93-byte log from 09:30 with no Clips lines.
- **Leak check (T-02-29):** `grep -Eq 'api_key=|api\.klipy\.com/api/v1/[^/]+/'` exited 1 (no match) for both `bin/install-dev.log` and `bin/gifsmith-run.log`. **PASS.**

## Decisions Made

See `key-decisions` in the frontmatter.

## Deviations from Plan

### Discretion choices and notes

**1. [Discretion] C2 primary action is an `<a>` mailto link.** It is labelled "Email clips@giphy.com ↗", styled as the primary button, and calls `openExternal(CLIPS_ACCESS_MAILTO)` (`mailto:clips@giphy.com?subject=GIPHY%20Clips%20API%20access%20request`). `StateMessage.svelte` gained an optional primary-link prop for it. The plan's `files_modified` did not list that file; this was a small, necessary addition. Committed in `6432749`.

**2. [Rule 2 - Correctness] Recheck guard.** A `rechecking` flag in App.svelte keeps repeated Clips selections while unknown from firing parallel `CheckClips` calls. Try Again on a Clips error state re-probes through the same path. Committed in `6432749`.

**3. [Rule 2 - Correctness] Rate-limited Recheck makes no request.** `Recheck` honours `Blocked()` like `Trigger` does, so a 429-blocked GIPHY never gets a Clips call. Committed in `1ec299b`.

**4. [Note] `ProviderTabs.svelte` unchanged.** It is in `files_modified`, but the 02-05 version already dims Clips for any status other than `enabled`, so no edit was needed.

**5. [Note] go.sum not committed.** The `go mod tidy` inside install:dev added `objx` lines to go.sum. They were restored and not committed, because no module dependency changed.

**Total deviations:** 2 correctness additions (Rule 2), 1 discretion choice, 2 notes. **Impact:** no new dependencies and no scope beyond the plan.

## Issues Encountered

- None blocking.
- **Coverage gap (known):** the UI path for a Clips re-probe that fails again with a transient error (unknown → CheckClips → still unknown → error state for "GIPHY Clips") is covered only by the Go tests (`TestCheckClipsUnknown`, `TestClipsRecheck`). No browser UAT drives it, because the fakeapi has no transient-Clips scenario.
- **By design:** a remembered Clips tab opens on GIPHY at launch, because Clips is not enabled until the probe finishes (02-05 rule). With the user's key Clips is `disabled`, so this always happens for them.

## Known Stubs

- None that block the plan goal. Clips copy (MP4) is intentionally E8 "Copying Clips (MP4) arrives in the next update." It is scheduled for Phase 3 (ACTN-09).

## User Setup Required

None. If the user wants Clips, they email clips@giphy.com from the C2 card and reopen Gifsmith after GIPHY approves the key.

## Next Phase Readiness

- Phase 2's plans are all complete, and the phase is ready for verification and code review.
- Phase 3 Clips copy (ACTN-09) can be tested natively only with a Clips-approved key. The user's current key is not approved, so native Clips copy needs GIPHY approval first or has to rely on the fakeapi.
- Still pending for Phase 4: Klipy written clearance for temp-file copies (recorded in the PROJECT.md compliance posture).

## Self-Check: PASSED

- Files: all 5 `key-files.created` exist
- Commits: 2ab298a, 1ec299b, 6432749 are on HEAD (`git rev-list --count 1180111..HEAD` = 3 before this SUMMARY)
- Clips outcome read from `bin/install-dev.log`; leak check exited 1 (no match) on both logs
- `bin/test-02-06-t1.txt` ends with `ok gifsmith/internal/staging` and contains no `--- FAIL`

---
*Phase: 02-providers-search-ui-first-run*
*Completed: 2026-10-17*
