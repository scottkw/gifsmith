---
phase: 03-action-pipeline-library
fixed_at: 2026-10-18T19:40:00Z
review_path: .planning/phases/03-action-pipeline-library/03-REVIEW.md
iteration: 2
findings_in_scope: 1
fixed: 1
skipped: 0
status: all_fixed
---

# Phase 3: Code Review Fix Report

**Fixed at:** 2026-10-18T19:40:00Z
**Source review:** .planning/phases/03-action-pipeline-library/03-REVIEW.md (iteration 2)
**Iteration:** 2

**Summary:**
- Findings in scope: 1 (CR-01; no Warnings remain; IN-01 to IN-07 are out of scope)
- Fixed: 1
- Skipped: 0

Mode: `workflow.use_worktrees=false`. All edits and commits were made in the main checkout on `gsd/phase-03-action-pipeline-library`. No worktree was created.

## Fixed Issues

### CR-01: `framePixels` walks extensions differently from `gif.DecodeAll`, so a short prefix hides every frame from the decode budget

**Files modified:** `internal/services/actions.go`, `internal/services/actions_test.go`
**Commits:** 4a0fd8c (RED test), 9ced557 (GREEN fix)
**Applied fix:**
- The `0x21` branch of `framePixels` now follows Go 1.27's `readExtension` (`image/gif/reader.go`) one label at a time:
  - `0x01` Plain Text: discard 13 raw bytes, then data sub-blocks.
  - `0xF9` Graphic Control: read exactly 6 bytes and require `[0]==4` and `[5]==0`. No sub-blocks follow.
  - `0xFE` Comment: data sub-blocks only.
  - `0xFF` Application: read a size byte, discard that many bytes (even when the size is 0), then data sub-blocks.
  - Any other label: return an error. The crop is refused through the existing F6 `crop_failed` path.
- **A second bypass the review missed.** The review said Application blocks "match Go". They match only when the size byte is non-zero. Go reads the size byte, skips 0 bytes, and then reads sub-blocks. The old parser read that 0 as the block terminator. So `21 FF 00 3B …` hid every frame in the same way as the Plain Text prefix. Mirroring Go closes both.
- **No second, independent budget.** With the parser matching Go label by label, the reviewer's frame-count cap is not needed:
  - Image data already matched, because `blockReader` consumes whole sub-blocks and ends at the terminator or errors.
  - Wherever Go errors, it stops allocating frames, so a mismatch at that point is harmless.
- Regression test `TestCopySquareDecodeBudgetExtensions` builds each GIF in memory from `bombGIF` (20 full-canvas 4096² frames, about 335 Mpx), with no testdata files. It inserts a prefix after the screen descriptor:
  - the reviewer's Plain Text prefix: `21 01 00 3B` + 11 bytes + `00`
  - a zero-size Application block: `21 FF 00 3B` + 59 bytes + `00`
  - an unknown label: `21 02 00 3B`

  For the first two cases, `framePixels` must count all 20 frames. For the unknown label it must return an error. In every case `CopySquare` must return `crop_failed` with no platform calls.
- RED (4a0fd8c, before the fix): Plain Text and empty Application gave `framePixels = 0` and `CopySquare` reported "staged and copied". This shows that Go decodes both prefixes. The unknown label gave no error from `framePixels`. GREEN (9ced557): the test passes.

## Iteration 1

Fixed in iteration 1 and committed on the same branch:
- 1974e91 — CR-01 first pass: the `framePixels` decode budget. Iteration 2 found it could be bypassed; the bypass is now fixed above.
- 70b2c1e — WR-01 single-instance guard
- eb1e74a — WR-02 clamp vendor titles and ids in the mappers
- 195ff62 — WR-03 offer Undo only when a removal removed something
- 34436b8 — IN-06 UAT scripts default their fixture env

## Verification

All gates ran in the **main checkout** (`/Users/ken/dev/gif-search`, branch `gsd/phase-03-action-pipeline-library`), so the results can be reproduced from this tree.

- `gofmt -l main.go wire_*.go internal`: clean
- `go vet . ./internal/...`: clean
- `go test -race ./internal/...`: all packages ok
- `golangci-lint run . ./internal/...`: 0 issues
- `CGO_ENABLED=0 go build -tags server .` and `go build .`: both build. Binaries went to the scratchpad.
- `bash scripts/smoke-browser-loop.sh --uat scripts/uat/11-square.js`: UAT OK
- `~/Applications/Gifsmith.app` was not touched.

---

_Fixed: 2026-10-18T19:40:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 2_


## Iteration 3 (orchestrator, after max auto iterations)

The final re-review (03-REVIEW.md, iteration 3) found one more CR-01 route: Go retains ~4 KB per
frame (palette clone on transparent frames) regardless of frame size, so ~786K empty 0x0 transparent
frames in a <15 MB file counted 0 against the w*h budget and decoded to ~3.7 GB live heap. Fixed
with the reviewer's one-line remedy.

- 6928379 test(03): CR-01 RED empty transparent frames bypass the decode budget
- 39d5c97 fix(03): CR-01 charge a fixed per-frame cost (frameOverhead = 4096+256) in framePixels

Verification: gofmt clean; go vet clean; go test -race ./internal/... 13/13 ok; golangci-lint 0
issues; desktop and -tags server builds ok; UAT 11-square OK; go.sum unchanged. Not re-reviewed by
an agent (loop cap reached) — the phase verifier should confirm CR-01 closed.
