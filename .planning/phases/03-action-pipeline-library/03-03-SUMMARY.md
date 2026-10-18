---
phase: 03-action-pipeline-library
plan: 03
subsystem: actions
tags: [go, staging, mp4, giphy-clips, klipy, wails-v3, svelte5, headless-uat]

requires:
  - phase: 03-action-pipeline-library
    plan: 01
    provides: ActionError kinds, shared paste(put, okMsg), putText/CopyText, platform.Fake server build, Gif GifURL
  - phase: 03-action-pipeline-library
    plan: 02
    provides: action menu rows (Copy MP4 / Paste MP4 labels for Clips), runAction dispatcher
provides:
  - giphy Clips DownloadURL/Width/Height from video.assets (720p, then 480p, 1080p, 360p; none fails the page)
  - staging.ErrNotMP4, kind-aware Fetch (ftyp check, <slug>.mp4), Store.Clear, Store.Remove
  - services.KlipyPasteGrace, ActionService.After, ActionService.gen, ActionService.ServiceShutdown
  - stage.Clear() at launch in main.go
  - fakeapi Clips 720p/480p/360p assets, no720 word, /media/*.mp4 ftyp fixture (broken*/html* variants)
  - --scrim token, Tile MP4 badge and ", MP4" aria-label, Clips F1/F3 rows; E8 retired
affects: [03-04, 03-05, 03-06, 03-07, 03-08]

actuals:
  tokens: 7119
  tasks: 2
  commits: 4
plan_head_before: 6b11120522510ef02031eae7268188da5b22b061
plan_head_after: 1b87b1bb7bbcf2bb595333cdc936219440d2c2ad

tech-stack:
  added: []
  patterns:
    - "Every clipboard write bumps ActionService.gen; delayed cleanup acts only if gen is unchanged and holds busy while it removes"
    - "Staging holds only the clipboard's file: Sweep after a file copy, Clear after a text copy, at launch and at quit"
    - "Timers are injected (After) like Sleep; tests record and run them, never wait"

key-files:
  created: []
  modified:
    - internal/provider/giphy/giphy.go
    - internal/provider/giphy/giphy_test.go
    - internal/provider/fakeapi/fakeapi.go
    - internal/staging/staging.go
    - internal/staging/staging_test.go
    - internal/services/actions.go
    - internal/services/actions_test.go
    - internal/services/errors.go
    - internal/services/settings_test.go
    - main.go
    - frontend/src/App.svelte
    - frontend/src/components/Tile.svelte
    - frontend/src/app.css
    - frontend/src/lib/errors.ts
    - frontend/src/lib/errors.test.ts
    - scripts/uat/06-clips.js

key-decisions:
  - "Clips MP4 assets decode through the existing rendition type (LooseInt width/height), so a malformed display-only dimension never fails a page; plan text said flexInt"
  - "Text copies clear staging in putText, so Paste Link / Paste Markdown clear it too, not only CopyText"
  - "The Klipy grace callback takes the busy lock before checking gen: without it a copy could re-stage the same directory between the check and the removal"
  - "Store.Remove also refuses Dir itself (rel == \".\"), not only paths outside it"

patterns-established:
  - "stage() returns the staged path; Paste wraps paste(put, okMsg) and schedules post-paste work only when Result.Pasted"

requirements-completed: [ACTN-09, ACTN-10]

coverage:
  - id: D1
    description: "Enter on a Clips tile copies the Clip's MP4 as a file URL with 'Copied MP4 — press ⌘V to paste'; a trusted paste reports 'Pasted MP4'"
    requirement: ACTN-09
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopyClipMP4, TestPasteClipMP4"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/06-clips.js"
        status: pass
    human_judgment: false
  - id: D2
    description: "Clips rendition is the first of 720p, 480p, 1080p, 360p; 4k/source only or no assets makes the page a typed http error"
    requirement: ACTN-09
    verification:
      - kind: unit
        ref: "internal/provider/giphy/giphy_test.go#TestClipsRendition, TestClipsMode, TestClipsGifURL"
        status: pass
    human_judgment: false
  - id: D3
    description: "A staged Clip must start with an ftyp box and is named <slug>.mp4; non-MP4 -> not_media (F3 'isn't an MP4'), failed download -> F1 Clips variant, nothing left on disk"
    requirement: ACTN-09
    verification:
      - kind: unit
        ref: "internal/staging/staging_test.go#TestFetchStagesMP4, TestFetchNotMP4; internal/services/actions_test.go#TestCopyClipFailingMediaKinds; frontend/src/lib/errors.test.ts#Clips copy errors name the MP4"
        status: pass
    human_judgment: false
  - id: D4
    description: "Clips tiles show the aria-hidden MP4 badge on --scrim with aria-label '{title}, MP4'; the ⌘K menu lists Copy MP4 / Paste MP4 and no GIF rows; E8 is gone"
    requirement: ACTN-09
    verification:
      - kind: automated_ui
        ref: "scripts/uat/06-clips.js (badge text, aria-hidden, computed background rgba(0, 0, 0, 0.6), aria-label suffix, menu rows)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Staging is wiped at launch and at quit (ServiceShutdown, no busy lock), and a successful text copy clears it while a failed one keeps it"
    requirement: ACTN-10
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestServiceShutdownClears, TestCopyTextClearsStaging, TestCopyTextFailureKeepsStaging; internal/staging/staging_test.go#TestClear, TestRemove"
        status: pass
    human_judgment: false
  - id: D6
    description: "A pasted Klipy file is removed 30 s after ⌘V unless the clipboard was written since; Klipy copy, untrusted paste and non-Klipy paste schedule nothing"
    requirement: ACTN-10
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestKlipyGraceDeletesAfterPaste, TestKlipyGraceGenerationGuard, TestKlipyGraceOnlyAfterKlipyPaste"
        status: pass
    human_judgment: false
  - id: D7
    description: "Native: a real Clips MP4 pasted into Messages/Slack plays as video (A1), real GIPHY MP4s start with ftyp (A6), 30 s Klipy grace is long enough (A2), Wails calls ServiceShutdown on ⌘Q"
    verification: []
    human_judgment: true
    rationale: "The user's GIPHY key is not approved for Clips and synthetic desktop input is forbidden; only fixtures prove this plan. Belongs to the plan 03-08 native checkpoint."

duration: 10min
completed: 2026-10-18
status: complete
---

# Phase 3 Plan 03: Clips MP4 and Staging Lifecycle Summary

**GIPHY Clips now copy and paste their 720p MP4 (or the nearest of 480p, 1080p, 360p) through the same file-URL path as GIFs, with ftyp validation, an MP4 tile badge and MP4 wording. The staging folder now holds at most the clipboard's one file: it is wiped at launch and quit, cleared by text copies, and a pasted Klipy file is deleted 30 s after ⌘V behind a generation guard.**

## Performance

- **Duration:** about 10 min
- **Started:** 2026-10-18T00:38:56Z
- **Completed:** 2026-10-18T00:48:52Z
- **Tasks:** 2 (each RED, then GREEN)
- **Files modified:** 16

## Accomplishments
- giphy: Clips decode `video.assets` through the existing `rendition` type. `clipRendition` picks 720p, then 480p, 1080p, 360p, and never 4k or source. A clip with none of these fails the page through `c.bad(status)`. `GifURL` stays the link source (D-18).
- staging: `Fetch` checks the GIF magic or, when `g.IsClip`, an `ftyp` box at bytes 4..8. A Clip is written as `<slug>.mp4`. A non-MP4 returns `ErrNotMP4` (mapped to `not_media`) and leaves nothing on disk. The GIF path is unchanged.
- services: `ErrClipsCopy` and its `IsClip` guard are gone. Clips report `Copied MP4 — press ⌘V to paste` and `Pasted MP4`.
- Lifecycle: `Store.Clear` and `Store.Remove` (with a containment check). Every clipboard write bumps `gen`. A successful text copy or paste clears staging. `ServiceShutdown` clears staging without taking the busy lock. `main.go` clears staging before `app.Run()`. After a Klipy paste, `expire` schedules a removal `KlipyPasteGrace` later through the injectable `After`. The removal runs only if `gen` has not changed.
- Frontend: `--scrim` token, MP4 badge (top-left, 4px inset, 16px, aria-hidden) and the `, MP4` accessible name. Clips variants of F1 and F3. The E8 constant and its guard in `run()` are gone.
- fakeapi: Clips carry 720p, 480p and 360p assets, and the `no720` word drops 720p. `/media/*.mp4` serves the 24-byte ftyp fixture. Names starting `broken` or `html` fail the same way as the GIF route.

## Task Commits

1. **Task 1 (tracer): Enter on a Clip copies its 720p MP4**: `2e2eae6` (test, RED), `e36fa39` (feat, GREEN)
2. **Task 2: Staging lifecycle**: `0657d5f` (test, RED), `1b87b1b` (feat, GREEN)

## TDD Gate Compliance

- **Task 1 RED:** the TS target "Clips copy errors name the MP4 (F1 and F3 Clips variants)" ran under `node --test --test-reporter=tap` and failed (36 pass, 1 fail). `gsd-tools check tdd-red-evidence` returned `RED_EVIDENCE_OK`. The Go targets (TestFetchStagesMP4, TestFetchNotMP4, TestClipsMode, TestClipsGifURL, TestClipsRendition, TestCopyClipMP4, TestPasteClipMP4, TestCopyClipFailingMediaKinds) all compiled against an `ErrNotMP4` skeleton and failed on their own assertions: "the download is not a GIF", empty DownloadURL, and "copying Clips arrives in a later update". Go `-v` output is not a classifier format. Semantic assessment: every target ran and failed for the planned reason, with no load or setup faults. Output is in `bin/red/0303-t1-go.txt` (gitignored).
- **Task 2 RED:** compiled against skeletons (`Clear`/`Remove` returning nil, `ServiceShutdown` returning nil, `After` unused). TestClear, TestRemove, TestCopyTextClearsStaging, TestServiceShutdownClears, TestKlipyGraceDeletesAfterPaste and TestKlipyGraceGenerationGuard failed on assertions. TestKlipyGraceGenerationGuard failed on "After calls = [], want one", because the guard cannot run until a timer exists. TestCopyTextFailureKeepsStaging and TestKlipyGraceOnlyAfterKlipyPaste passed in RED by design: they guard against an over-eager implementation. Output is in `bin/red/0303-t2-go.txt`.
- Each GREEN commit follows its RED commit. No refactor commits.

## Verification (latest run)

- `gofmt -l` empty, `go vet . ./internal/...`, `go test -race ./internal/... -count=1` all ok, `golangci-lint run . ./internal/...` reports 0 issues. `go build .` (desktop, cgo) and `CGO_ENABLED=0 go build -tags server .` both ok.
- `bin/test-03-03-t2.txt`: no FAIL, with PASS lines for TestKlipyGraceGenerationGuard and TestServiceShutdownClears.
- `pnpm --dir frontend run test`: 37 pass, 0 fail. `svelte-check`: 0 errors, 0 warnings.
- `scripts/smoke-browser-loop.sh`: SMOKE OK.
- `--uat` UAT OK for 06-clips, 01-tracer, 02-keyboard, 02-grid, 03-states, 05-tabs, 05-setup, 07-text-actions (whose text copies now clear staging) and 08-action-menu. With `FAKE_CLIPS=denied`, 06-clips-denied and 06-backstops also pass. The C2 card is unchanged.
- Acceptance greps: `Copying Clips (MP4)` appears 0 times in App.svelte and 06-clips.js. `"ftyp"`, `720p`, `--scrim`, Tile `MP4`, `KlipyPasteGrace = 30 * time.Second`, `ServiceShutdown() error`, `Clear() error` and `Remove(dir string) error` are all present. `stage.Clear()` is at main.go:110, before `app.Run()` at line 113.

## Decisions Made
- Clips assets reuse `rendition` (LooseInt) instead of a new `asset{flexInt}` type. This is less code and keeps 03-01's rule that display-only numbers never fail a page.
- The clear-after-text-copy lives in `putText`, so it covers both CopyText and PasteText. Both leave the clipboard pointing at no file (D-17).
- The Klipy timer callback takes `busy` before it checks `gen`. Checking `gen` alone leaves a window: a re-copy of the same GIF could write into the same hash directory between the check and the `RemoveAll`. As a result, a keypress that lands in the microseconds the delete holds the lock is refused as busy. A `shortcut:` comment marks this.
- `Store.Remove` refuses `Dir` itself as well as anything outside it.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Phase 2 test asserted that Clips have no download URL**
- **Found during:** Task 1 GREEN
- **Issue:** `TestClipsRefusedUnlessEnabled` in `internal/services/settings_test.go` (not in the plan's file list) required `DownloadURL == ""` for Clips. That only held while Clips copy was deferred, which this plan ends.
- **Fix:** The test now expects the fixture's `-720p.mp4` URL.
- **Committed in:** `e36fa39`

**2. [Rule 1 - Correctness] Klipy grace delete holds the busy lock**
- **Found during:** Task 2
- **Issue:** Research Code Example 5 checks `gen` and then calls `RemoveAll` with no lock, which races with a concurrent re-stage of the same directory.
- **Fix:** The callback takes `busy` (blocking) around the check and the removal. ServiceShutdown still takes no lock, as the plan's prohibition requires.
- **Committed in:** `1b87b1b`

**3. [Rule 2 - Correctness] `internal/services/errors.go` comment**
- Dropped the "besides ErrClipsCopy" reference after deleting `ErrClipsCopy`. Committed in `e36fa39`.

---

**Total deviations:** 3 auto-fixed (1 blocking, 2 correctness). **Impact:** no scope creep.

## Issues Encountered
- The app does not set `SingleInstance`, so a second launch would run `stage.Clear()` and remove the first instance's current clipboard file. This pre-existing gap is not fixed here. Single-instance behaviour belongs to a later shell or packaging plan.
- Pitfall 5 is accepted by D-17: a file copied just before quitting can no longer be pasted after quit. The native UAT should note this so it is not reported as a bug.

## Known Stubs
None.

## Threat Flags
None. The MP4 download, Remove and Clear, and the Klipy timers are covered by T-03-08 to T-03-11 and mitigated as planned: the ftyp check runs before any byte is kept, Remove has a containment check with unit tests, and the generation guard is in place.

## User Setup Required
None.

## Next Phase Readiness
- Later plans can use `stage()` (which now returns the path), the `gen` and `expire` pattern for PasteSquare on Klipy (03-04/05 should call `expire` after a pasted Klipy square), `Store.Remove` and `Store.Clear`, and the MP4 badge and `--scrim` for the Favorites badge.
- Native confirmation (D7) waits for the plan 03-08 checkpoint. Clips MP4 paste can only be checked there if the user's GIPHY key is approved for Clips.

---
*Phase: 03-action-pipeline-library*
*Completed: 2026-10-18*

## Self-Check: PASSED
