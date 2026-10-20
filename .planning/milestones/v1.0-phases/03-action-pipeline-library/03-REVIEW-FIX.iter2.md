---
phase: 03-action-pipeline-library
fixed_at: 2026-10-18T19:20:00Z
review_path: .planning/phases/03-action-pipeline-library/03-REVIEW.md
iteration: 1
findings_in_scope: 4
fixed: 4
skipped: 0
status: all_fixed
---

# Phase 3: Code Review Fix Report

**Fixed at:** 2026-10-18T19:20:00Z
**Source review:** .planning/phases/03-action-pipeline-library/03-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 4 (CR-01, WR-01, WR-02, WR-03; the orchestrator also asked for IN-06)
- Fixed: 4 in scope, plus IN-06
- Skipped: 0

Mode: `workflow.use_worktrees=false`. All edits and commits were made in the main checkout on `gsd/phase-03-action-pipeline-library`. No worktree was created.

## Fixed Issues

### CR-01: The square-crop limits do not stop a decompression bomb

**Files modified:** `internal/services/actions.go`, `internal/services/actions_test.go`
**Commit:** 1974e91
**Applied fix:**
- Added `SquareMaxPixels = 256 << 20`. This is the sum of w*h over all frames, about 256 MB of `*image.Paletted`.
- Added `framePixels(io.Reader)`. It walks the GIF block structure: header and screen descriptor, global and local color tables, extension sub-block chains, and image descriptors. It sums each frame's declared w*h without decoding any LZW data.
- `cropSquare` now runs `framePixels` after `DecodeConfig` and before `DecodeAll`. When the total is over budget, or the structure cannot be parsed, it refuses through the existing `crop_failed` (F6) path, the same path the canvas-size check uses.
- I chose F6 because F7's copy says "over 15 MB", which would be wrong for a small file.
- Regression test `TestCopySquareDecodeBudget`: an in-memory GIF of about 200 KB with a 4096x4096 canvas and 20 copies of one encoded full-canvas frame, which would decode to about 320 MB. The test checks:
  - the file is under the 15 MB cap
  - `framePixels` returns exactly 20 x 4096 x 4096
  - `CopySquare` returns `crop_failed` and makes no platform calls.

  No testdata file was added.

### WR-01: No single-instance guard

**Files modified:** `main.go`
**Commit:** 70b2c1e
**Applied fix:**
- Added `SingleInstance: &application.SingleInstanceOptions{UniqueID: "com.kenscott.gifsmith", OnSecondInstanceLaunch: func(application.SecondInstanceData) { sh.Show() }}`.
- `sh` is declared before `application.New` and assigned before `app.Run`.
- I checked this against the Wails v3.0.0-beta.30 source:
  - On darwin, an `flock` on `NSTemporaryDirectory()/<UniqueID>.lock`. The second process posts a distributed notification and exits.
  - In `-tags server` builds the lock is a no-op (`application_server.go` `serverLock`), so the browser UAT build is unaffected.
- Because the dev build and the installed build share one UniqueID, a `wails3 dev` build now exits and shows the installed app's panel while the installed app runs. The two can no longer race on the shared library and staging files or on the hotkey.
- **Not done:** separate config and staging folders for the dev bundle. `wire_desktop.go` has no build-time signal for the bundle id, so this is not trivial. It is also unnecessary now that the two processes cannot coexist. If you want to run dev alongside the installed app, that needs a dev build tag that also changes the UniqueID and the folders.
- **Not done:** re-reading the library file under `mu` before each mutation. With a single instance there are no other writers.

### WR-02: `Validate` rejects long titles and ids instead of truncating them

**Files modified:** `internal/provider/provider.go`, `internal/provider/provider_test.go`, `internal/provider/giphy/giphy.go`, `internal/provider/giphy/giphy_test.go`, `internal/provider/klipy/klipy.go`, `internal/provider/klipy/klipy_test.go`, `internal/provider/finergifs/finergifs.go`, `internal/provider/finergifs/finergifs_test.go`, `internal/library/store.go`
**Commit:** eb1e74a
**Applied fix:**
- Added `provider.MaxTitleRunes = 500` and `provider.MaxIDBytes = 200`.
- Added `ClampTitle` (truncates by rune) and `ClampID` (truncates by byte, cutting back to a rune start).
- All three mappers clamp `ID` and `Title` at the mapper boundary. The grid, the snapshot, the star-badge keys and the store therefore all see the same values.
- `library.Validate` keeps the same limits, now the shared constants, as the strict guard for JS input and tampered files.
- Tests:
  - `TestClamp`: rune safety and the limits.
  - One mapper test per provider: GIPHY long id and title, Klipy long slug that becomes both id and title, Finer long subtitle.
- **Not done:** a separate `invalid_item` message in `errors.ts`. After clamping, `invalid_item` comes only from a genuinely invalid item, such as a disallowed URL.

### WR-03: A removal that changes nothing still shows an Undo toast, and ⌘Z then restores an earlier removal

**Files modified:** `internal/library/store.go`, `internal/library/store_test.go`, `frontend/bindings/gifsmith/internal/library/models.ts`, `frontend/src/lib/api.ts`, `frontend/src/App.svelte`
**Commit:** 195ff62
**Applied fix:**
- In Go, `remove` with nothing to remove now empties the Undo slot (`s.undo = nil`).
- `Snapshot` gains `CanUndo bool` (json `canUndo`), set to `s.undo != nil`.
- Bindings were regenerated with `wails3 generate bindings -clean=true -ts -i` (beta.30). The only diff is the new field.
- In the frontend, `mutate()` now returns the mutation's own `Library` (or `null`) instead of a boolean. Unfavorite, Remove from Recents and Clear Recents show the Undo toast only when that snapshot has `canUndo`. ⌘Z is captured only while the Undo toast is visible, so Undo can no longer restore an unrelated removal.
- The slot effectively expires with the toast. The webview never calls Undo outside the toast window, and every new removal, including a no-op one, replaces or clears the slot.
- Test `TestNoOpRemovalClearsUndo` covers unfavoriting a missing item, removing a missing Recent, and clearing an empty Recents list. Each case checks that `CanUndo` is false and that a following `Undo` changes nothing.
- **Not done:** a removal token passed to `Undo(token)`.
- **Status: fixed, requires human verification.** This is state-handling logic. Confirm in the native app that ⌘⇧F on a stale star shows no Undo toast.

### IN-06 (out of scope, applied on orchestrator request): The default UAT loop cannot pass `12-library-unreadable.js`

**Files modified:** `scripts/smoke-browser-loop.sh`
**Commit:** 34436b8
**Applied fix:**
- `--uat` reads line 2 of the script (its "Run via:" line). When that line names them, it defaults `FAKE_CLIPS=denied` and `UAT_SEED_FAVORITES='{not json'`. A value set by the caller still wins.
- The same problem affected `06-backstops.js`, `06-clips-denied.js` and `14-backstops.js` (they need `FAKE_CLIPS=denied`). The first plain loop failed those three, and this fix covers them too.

## Verification

All gates ran in the **main checkout** (`/Users/ken/dev/gif-search`, branch `gsd/phase-03-action-pipeline-library`), so the results can be reproduced from this tree.

- `gofmt -l main.go wire_*.go internal`: clean
- `go vet . ./internal/...`: clean
- `go test -race ./internal/...`: all packages ok
- `golangci-lint run . ./internal/...`: 0 issues
  - `./...` reports 1 issue, an unused func in the `build/ios` template. It is unrelated to this change and is excluded per the executor rules.
- `go build .` (desktop, cgo) and `CGO_ENABLED=0 go build -tags server .`: both build. Binaries went to the scratchpad. The installed `~/Applications/Gifsmith.app` was not touched, quit or reinstalled.
- `pnpm check` (svelte-check): 0 errors, 0 warnings
- `pnpm test` (node:test): 63/63 pass
- `scripts/smoke-browser-loop.sh`: SMOKE OK
- Plain loop `for f in scripts/uat/*.js; do scripts/smoke-browser-loop.sh --uat "$f"; done` with no hand-set env: **18/18 PASS**. This includes `12-library-unreadable.js`, whose check that the seeded file is unchanged also passed.
- Not verified natively (agents send no desktop input):
  - a second launch exiting and focusing the existing panel (WR-01)
  - the absence of a stale Undo toast in the installed app (WR-03)

---

_Fixed: 2026-10-18T19:20:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
