---
phase: 03-action-pipeline-library
plan: 05
subsystem: actions
tags: [go, image-gif, square-crop, staging, wails-v3, svelte5, headless-uat]

requires:
  - phase: 03-action-pipeline-library
    plan: 03
    provides: stage()/Sweep, gen counter, expire() Klipy grace, shared paste(put, okMsg)
  - phase: 03-action-pipeline-library
    plan: 04
    provides: staging.media() request helper and spool(), App toast model, runAction dispatcher
provides:
  - package internal/imaging with Square (lossless center crop, stdlib image/gif only)
  - staging.Store.FetchMax (Fetch = FetchMax at the default cap) and Store.WriteDerived, sharing put()
  - services.SquareMaxBytes (15 MB), services.SquareMaxSide (4096), ActionService.CopySquare, ActionService.PasteSquare
  - ActionError kinds crop_failed and crop_too_large; F6/F7 copy in errors.ts
  - Copy Square GIF / Paste Square GIF menu rows (group 4, omitted for Clips); api.copySquare/pasteSquare
  - App overlay flag separate from copying; 300 ms Cropping… timer; Tile/ResultGrid overlay prop
  - fakeapi wide-*/tall-* media (48x32 / 32x48), Finer widemedia/tallmedia words, GIPHY hugesize word
  - smoke CopySquare check and scripts/uat/11-square.js
affects: [03-06, 03-07, 03-08]

actuals:
  tokens: 14567
  tasks: 2
  commits: 4
plan_head_before: af0d96ee85f345ad684597067366ae23cda01e8c
plan_head_after: 66c32bfd23d4328e9e1ca7945a44c95d760126bd

tech-stack:
  added: []
  patterns:
    - "Derived clipboard files are staged under their own hash dir (provider:id:variant), so Sweep(derived) also removes the downloaded original"
    - "Heavy decode work is gated twice: the claimed size before any request, then DecodeConfig before DecodeAll"
    - "Busy UI is a separate overlay flag from the action lock: the lock ignores repeats at once, the overlay can be delayed"

key-files:
  created:
    - internal/imaging/square.go
    - internal/imaging/square_test.go
    - scripts/uat/11-square.js
  modified:
    - internal/staging/staging.go
    - internal/staging/staging_test.go
    - internal/services/actions.go
    - internal/services/actions_test.go
    - internal/services/errors.go
    - internal/provider/fakeapi/fakeapi.go
    - frontend/bindings/gifsmith/internal/services/actionservice.ts
    - frontend/src/lib/api.ts
    - frontend/src/lib/actions.ts
    - frontend/src/lib/actions.test.ts
    - frontend/src/lib/errors.ts
    - frontend/src/lib/errors.test.ts
    - frontend/src/components/Tile.svelte
    - frontend/src/components/ResultGrid.svelte
    - frontend/src/App.svelte
    - scripts/smoke-browser-loop.sh
    - scripts/uat/08-action-menu.js

key-decisions:
  - "A frame entirely outside the crop window becomes a 1x1 transparent frame with disposal None, not its original disposal: with Background the placeholder clears the visible pixel (0,0). The golden test caught this in Research Code Example 1"
  - "Fetch and WriteDerived share one put() helper (hash dir, Rel containment check, .part + rename); media() and spool() take the byte cap as an argument instead of reading the Store"
  - "The square file is staged as <slug>-square.gif under sha1(provider:id:square), so the CopyFile + Sweep(squarePath) sequence of stage() removes the original"
  - "Text actions still show the overlay at once (unchanged); only square actions delay it"

patterns-established:
  - "stageSquare mirrors stage(): download, transform, WriteDerived, CopyFile, bump gen, Sweep; PasteSquare wraps it in paste() and calls expire() for Klipy"

requirements-completed: [ACTN-06]

coverage:
  - id: D1
    description: "Copy/Paste Square GIF put a center-cropped square GIF on the clipboard (file URL, staged like a normal copy) with frames, delays, loop count and disposal intact and every pixel in the window identical"
    requirement: ACTN-06
    verification:
      - kind: unit
        ref: "internal/imaging/square_test.go#TestSquareGolden, TestSquareWide, TestSquareTall, TestSquareAlreadySquareIsIdentical, TestSquareOutsideFramePlaceholder"
        status: pass
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopySquare, TestPasteSquareTrusted, TestPasteSquareWithoutAccessibility, TestPasteSquareKlipyGrace"
        status: pass
      - kind: integration
        ref: "bash scripts/smoke-browser-loop.sh (CopySquare -> HTTP 200, exactly one *-square.gif and no other staged file)"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/11-square.js"
        status: pass
    human_judgment: false
  - id: D2
    description: "Square rows are omitted for Clips and Go refuses a square crop for a Clip with no request"
    requirement: ACTN-06
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopySquareRefusesClips, TestPasteSquareRefusesClips; frontend/src/lib/actions.test.ts#Paste Square GIF follows Copy Square GIF; a Clip never lists a square row"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/11-square.js and 06-clips.js (Clips menu has no Square row)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Claimed size over 15 MB refused before any request (F7); a download over 15 MB ends in F7; a canvas over 4096 or an undecodable GIF ends in F6; HTML is F3; the panel stays open"
    requirement: ACTN-06
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopySquareRefusesClaimedSize, TestCopySquareDownloadOverCap, TestCopySquareCropFailures; frontend/src/lib/errors.test.ts#square crop errors show F6 and F7"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/11-square.js (GIPHY hugesize -> F7, grid and query kept)"
        status: pass
    human_judgment: false
  - id: D4
    description: "The busy overlay and Cropping… appear only if the crop is still running after 300 ms; repeats are ignored while it runs"
    requirement: ACTN-06
    verification:
      - kind: automated_ui
        ref: "scripts/uat/11-square.js (MutationObserver: a fast fixture crop never shows the overlay or Cropping…)"
        status: pass
    human_judgment: true
    rationale: "Only the negative side (fast crop shows nothing) is automated; the positive side needs a crop slower than 300 ms, which the fixtures do not produce. Check with a large real GIF at the plan 03-08 native checkpoint."
  - id: D5
    description: "Native: a pasted square GIF animates in Slack/Messages"
    verification: []
    human_judgment: true
    rationale: "Real pasteboard and synthetic ⌘V are forbidden in automated runs; belongs to the plan 03-08 native checkpoint."

duration: 12min
completed: 2026-10-18
status: complete
---

# Phase 3 Plan 05: Copy and Paste Square GIF Summary

**Copy Square GIF and Paste Square GIF in the ⌘K menu crop the selected GIF to a center square in Go with stdlib `image/gif`. Frames are cut, not re-encoded through a quantizer, so palettes, delays, loop count and disposal survive. A golden test renders every frame with disposal applied and finds 0 mismatched pixels inside the window. The square file is staged like any copy and the downloaded original is swept. Oversized GIFs are refused before any download (F7), giant canvases before the full decode (F6), and Clips never see the rows.**

## Performance

- **Duration:** about 12 min
- **Started:** 2026-10-18T01:05:31Z
- **Completed:** 2026-10-18T01:17:05Z
- **Tasks:** 2 (each RED, then GREEN)
- **Files modified:** 20

## Accomplishments
- `internal/imaging.Square`: the window is `min(w, h)` square and centered. Each frame is `SubImage` re-based with `Rect.Sub(win.Min)`. A frame that misses the window becomes a 1x1 transparent frame with disposal None and keeps its own delay. The package imports only `image`, `image/color` and `image/gif`; `go list -deps` shows no Wails.
- staging: `FetchMax(ctx, g, maxBytes)`, with `Fetch` delegating at the default cap. `WriteDerived(g, variant, b)` writes `<sha1(provider:id:variant)[:16]>/<slug>-<variant>.gif`. Both go through one `put()` (Rel containment check, `.part`, rename).
- services: `CopySquare` and `PasteSquare` hold the busy lock and do these steps in order:
  1. Refuse a Clip (`not_allowed`).
  2. Refuse `g.Size > SquareMaxBytes` (`crop_too_large`, no request).
  3. Fetch with the 15 MB cap.
  4. Run `DecodeConfig` and refuse either side over 4096 (`crop_failed`).
  5. `DecodeAll`, then `Square`, then `EncodeAll`.
  6. `WriteDerived`, `CopyFile`, bump `gen`, `Sweep(square)`.

  PasteSquare uses the shared `paste()` helper and calls `expire()` after a pasted Klipy square.
- Frontend:
  - Rows `copySquare` and `pasteSquare` sit in group 4 after Download, with no hint, and are omitted for Clips.
  - F6 and F7 copy.
  - App keeps `copying`, which ignores repeats at once, separate from a new `overlay` flag. Copy, paste and download set `overlay` at once. Square actions set it, along with `Cropping…`, only from a 300 ms timer that is cleared when the call settles.
  - Tile and ResultGrid take `overlay`.
- Fixtures: fakeapi `wide*` (48x32) and `tall*` (32x48) media, Finer `widemedia`/`tallmedia` words, and a GIPHY `hugesize` word (`original.size` "16000000").

## Task Commits

1. **Task 1 (tracer): Copy Square GIF end to end**: `dc94f4b` (test, RED), `b2ec7a9` (feat, GREEN)
2. **Task 2: Golden tests, limits, Paste Square, 300 ms rule**: `b2c988d` (test, RED), `66c32bf` (feat, GREEN)

## TDD Gate Compliance

- **Task 1 RED (TS):** the target "Copy Square GIF follows Download in group 4 and is omitted for Clips (D-15)" failed under `node --test --test-reporter=tap` (4 pass, 3 fail; the other 2 failures are existing row-list tests updated to the new row). `gsd-tools check tdd-red-evidence` returned `RED_EVIDENCE_OK` (`bin/red/0305-t1-actions.json`). Semantic assessment: the target ran and failed because there was no copySquare row. No load faults.
- **Task 1 RED (Go):** the tests compiled against skeletons: `Square` returned its input, and `FetchMax`, `WriteDerived` and `CopySquare` returned errors. TestSquareWide ("canvas 120x60, want 60x60"), TestSquareTall, TestFetchMaxCap, TestWriteDerived, TestCopySquare and TestCopySquareRefusesClips failed on their own assertions. TestSquareAlreadySquareIsIdentical passed in RED by design: it guards against a crop that changes square input. Output is in `bin/red/0305-t1-go.txt` (gitignored). Go `-v` output is not a classifier format.
- **Task 2 RED:** both TS targets ("Paste Square GIF follows Copy Square GIF; a Clip never lists a square row" and "square crop errors show F6 and F7") returned `RED_EVIDENCE_OK`. On the Go side, TestSquareGolden failed with 1 mismatch: a real bug in the Task 1 Square, found by the disposal-aware render (see Deviations). TestSquareOutsideFramePlaceholder also failed. TestCopySquareRefusesClaimedSize failed (1 request was made), TestCopySquareCropFailures failed (the giant and tall canvases were cropped), and the PasteSquare tests failed against the skeleton. TestCopySquareDownloadOverCap passed in RED: the 15 MB fetch cap was already Task 1 scope. Output is in `bin/red/0305-t2-imaging.txt` and `bin/red/0305-t2-go.txt`.
- Each GREEN commit follows its RED commit. No refactor commits.

## Verification (latest run)

- `gofmt -l` empty. `go vet . ./internal/...` ok. `go test -race ./internal/... -count=1` all ok. `golangci-lint run . ./internal/...` reports 0 issues. `go build .` (desktop, cgo) and `CGO_ENABLED=0 go build -tags server .` both build.
- `bin/test-03-05-t2.txt` has no FAIL, has `--- PASS: TestSquareGolden` with "0 mismatches" for `wide every disposal` (6 frames), `tall` (3) and `square` (2), and has the imaging `ok` line.
- `pnpm --dir frontend run test`: 43 pass, 0 fail. `svelte-check`: 0 errors, 0 warnings.
- `scripts/smoke-browser-loop.sh`: SMOKE OK. CopySquare returned HTTP 200 with "Copied square GIF — press ⌘V to paste". Staging holds exactly one file, `*/office-moment-1-square.gif`.
- `--uat` reports UAT OK for 11-square, 06-clips, 10-download, 08-action-menu, 01-tracer, 02-grid, 02-keyboard, 03-states, 05-setup, 05-tabs and 07-text-actions. With `FAKE_CLIPS=denied`, 06-clips-denied and 06-backstops also pass.
- Acceptance greps match:
  - `func Square(g *gif.GIF) *gif.GIF`
  - `SquareMaxBytes = 15 << 20`
  - `DecodeConfig`
  - `func (a *ActionService) CopySquare` and `PasteSquare`
  - `Cropping…` next to `setTimeout(busy, CROP_BUSY_MS)` (300)

  `go list -deps ./internal/imaging | grep -c wailsapp` prints 0. `square_test.go` is 218 lines.

## Decisions Made
See `key-decisions` in the frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] The research crop algorithm cleared a visible pixel**
- **Found during:** Task 2 RED (TestSquareGolden: 1 mismatch in the wide fixture)
- **Issue:** Research Code Example 1 copies each frame's disposal onto its 1x1 placeholder. If a frame outside the window has DisposalBackground, the placeholder clears pixel (0,0) of the square after it shows. The original frame never touched that pixel. The interfaces line "every Delay and Disposal copied" cannot hold for placeholders.
- **Fix:** placeholders get DisposalNone. Inside the window this matches the original for every disposal kind, because the original frame drew nothing there. The golden test and `checkSquare` expect None for frames outside the window. TestSquareOutsideFramePlaceholder asserts it.
- **Files modified:** internal/imaging/square.go, internal/imaging/square_test.go
- **Commit:** `66c32bf`

**2. [Rule 3 - Blocking] Existing tests hard-coded the action rows**
- **Found during:** Task 1 and Task 2 GREEN
- **Issue:** `frontend/src/lib/actions.test.ts` (row lists, Klipy length, the "paste" filter) and `scripts/uat/08-action-menu.js` (Finer rows, separators, GIPHY row count, clamp row) listed the menu without square rows. 08-action-menu.js is not in the plan's file list.
- **Fix:** both now expect the square rows: 9 Finer rows with 3 separators, 11 GIPHY rows, and the clamp lands on Paste Square GIF.
- **Commits:** `b2ec7a9`, `66c32bf`

**3. [Rule 1 - Lint] errcheck on deferred Close**
- `defer f.Close()` in `cropSquare` and in TestCopySquare became `defer func() { _ = f.Close() }()`. Committed in `66c32bf`.

**4. 11-square.js waits out the search debounce for "hugesize"**
- The GIPHY fixture ids are the same for trending and search, so waiting for tile 0 alone would act on the trending item. The script waits 700 ms and then for `aria-busy=false` before opening the menu.

---

**Total deviations:** 3 auto-fixed (1 correctness bug, 1 blocking, 1 lint) and 1 note. **Impact:** no scope creep. The crop is now lossless for frames outside the window as well.

## Issues Encountered
- If the crop fails after the download (F6), the downloaded original stays in staging until the next clipboard write or quit. It cannot be removed safely: the clipboard may point at the same hash directory from an earlier plain copy of the same GIF. This is the same retention D-17 accepts for a Klipy file that is copied but not pasted.
- Only the negative side of the 300 ms rule is automated: a fast crop shows no busy state. A crop slower than 300 ms needs a large real GIF. Check it at the plan 03-08 native checkpoint.
- Output size can grow, because Go's encoder does no inter-frame optimization (STACK item 6, -45% to +29%). The 15 MB input cap bounds it.

## Known Stubs
None.

## Threat Flags
None. T-03-18 to T-03-20 are mitigated as planned:
- **T-03-18:** the claimed-size refusal happens before any request, the 15 MB FetchMax cap applies whatever the claim, DecodeConfig refuses canvases over 4096 before DecodeAll, and the busy lock allows one crop at a time.
- **T-03-19:** WriteDerived uses its own hash directory with the Rel containment check, writes `.part` then renames, and Sweep keeps only the square file.
- **T-03-20:** the crop changes geometry only, and the golden per-pixel test checks it.

No package installs.

## User Setup Required
None.

## Next Phase Readiness
- Plan 03-06 can record Recents at the shared success point. CopySquare and PasteSquare return the same `Result` shape as Copy and Paste, so they need nothing special.
- WriteDerived is generic over `variant` if a later derived file appears.
- Native confirmation (a square paste animates in Slack/Messages, and the 300 ms positive path on a large GIF) waits for the plan 03-08 checkpoint.

---
*Phase: 03-action-pipeline-library*
*Completed: 2026-10-18*

## Self-Check: PASSED
