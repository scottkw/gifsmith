---
phase: 03-action-pipeline-library
reviewed: 2026-10-18T17:41:55Z
depth: standard
files_reviewed: 73
files_reviewed_list:
  - frontend/src/app.css
  - frontend/src/App.svelte
  - frontend/src/components/ActionMenu.svelte
  - frontend/src/components/DetailsPane.svelte
  - frontend/src/components/Footer.svelte
  - frontend/src/components/LibraryGrid.svelte
  - frontend/src/components/ProviderTabs.svelte
  - frontend/src/components/RecentSearches.svelte
  - frontend/src/components/ResultGrid.svelte
  - frontend/src/components/SetupCard.svelte
  - frontend/src/components/Tile.svelte
  - frontend/src/lib/actions.test.ts
  - frontend/src/lib/actions.ts
  - frontend/src/lib/api.ts
  - frontend/src/lib/errors.test.ts
  - frontend/src/lib/errors.ts
  - frontend/src/lib/format.test.ts
  - frontend/src/lib/format.ts
  - frontend/src/lib/grid.test.ts
  - frontend/src/lib/grid.ts
  - frontend/src/lib/library.test.ts
  - frontend/src/lib/library.ts
  - frontend/src/lib/popover.test.ts
  - frontend/src/lib/popover.ts
  - frontend/src/lib/providers.ts
  - frontend/src/lib/searches.test.ts
  - frontend/src/lib/searches.ts
  - internal/imaging/square_test.go
  - internal/imaging/square.go
  - internal/library/store_test.go
  - internal/library/store.go
  - internal/platform/fake.go
  - internal/platform/platform_darwin_test.go
  - internal/platform/platform_darwin.go
  - internal/platform/platform_darwin.h
  - internal/platform/platform_darwin.m
  - internal/platform/platform_stub.go
  - internal/platform/platform.go
  - internal/provider/fakeapi/fakeapi.go
  - internal/provider/finergifs/finergifs_test.go
  - internal/provider/giphy/giphy_test.go
  - internal/provider/giphy/giphy.go
  - internal/provider/httpx/httpx.go
  - internal/provider/klipy/klipy_test.go
  - internal/provider/klipy/klipy.go
  - internal/provider/provider.go
  - internal/services/actions_test.go
  - internal/services/actions.go
  - internal/services/errors.go
  - internal/services/library_test.go
  - internal/services/library.go
  - internal/services/settings_test.go
  - internal/settings/settings_test.go
  - internal/settings/settings.go
  - internal/staging/staging_test.go
  - internal/staging/staging.go
  - main.go
  - scripts/smoke-browser-loop.sh
  - scripts/uat/01-tracer.js
  - scripts/uat/02-keyboard.js
  - scripts/uat/05-tabs.js
  - scripts/uat/06-clips.js
  - scripts/uat/07-text-actions.js
  - scripts/uat/08-action-menu.js
  - scripts/uat/10-download.js
  - scripts/uat/11-square.js
  - scripts/uat/12-library-unreadable.js
  - scripts/uat/12-library.js
  - scripts/uat/13-chips.js
  - scripts/uat/13-details.js
  - scripts/uat/14-backstops.js
  - wire_desktop.go
  - wire_server.go
findings:
  critical: 1
  warning: 0
  info: 7
  total: 8
status: issues_found
---

# Phase 3: Code Review Report (iteration 3)

**Reviewed:** 2026-10-18T17:41:55Z
**Depth:** standard
**Files Reviewed:** 73
**Status:** issues_found

## Summary

This re-review covers the iteration-2 CR-01 fix: 4a0fd8c (the RED test) and 9ced557 (`framePixels` now walks extensions the way `readExtension` does). Since 34436b8, only `internal/services/actions.go` and `actions_test.go` have changed. `go vet . ./internal/...` is clean, apart from the usual ld macOS-version warnings, and `go test ./internal/...` passes.

**The parser mismatch is closed.** I compared `framePixels` with `$(go env GOROOT)/src/image/gif/reader.go` line by line:

- **Header:** 13 bytes. The global color table is `3 << (bits+1)`, which equals `3 * (1 << (1+bits))`.
- **Plain Text (0x01):** 13 raw bytes, then sub-blocks.
- **Graphic Control (0xF9):** 6 bytes, `[0]==4` and `[5]==0`, and no sub-blocks follow.
- **Comment (0xFE):** sub-blocks only.
- **Application (0xFF):** a size byte, that many bytes (0 is allowed), then sub-blocks. The NETSCAPE `readBlock` is just the first sub-block.
- **Unknown labels:** refused, as Go refuses them.
- **Image descriptor:** 9 bytes, an optional local table, the literal-width byte, then sub-blocks. When Go's `blockReader.close()` succeeds, it has consumed up to the first zero-length byte. That is the same place the sub-block loop stops.
- **Interlace:** this only permutes `Pix` and does not change the stream layout.

I also ran a differential fuzz (overlay-only) for 90 s, about 3.5M executions. It checked that whenever `gif.DecodeAll` succeeds, `framePixels` succeeds and returns exactly the sum of `Dx*Dy` over the frames. It found no divergence.

**The budget itself is still bypassable (CR-01 below).** The metric sums `w*h` over frames. It charges nothing for a frame's fixed decode cost, and Go spends about 4 KB of retained heap on each transparent frame even when the frame is 0x0. I reproduced a 15 MB GIF that scores 0 pixels, holds 3.7 GB live after `DecodeAll`, and goes through `CopySquare` successfully.

**Status of the earlier findings**

| Prior | Status | Evidence |
|-------|--------|----------|
| CR-01 (iter 2) extension mismatch | Fixed | Line-by-line comparison and the differential fuzz above. The RED test covers Plain Text, a zero-size Application block, and an unknown label. |
| CR-01 decode bomb (the threat) | **Still open, by another route** | Zero-pixel frames with a per-frame palette clone. See CR-01. |
| WR-01 single instance | Still fixed | `main.go` and `wire_*.go` are unchanged since the iteration-2 verification. |
| WR-02 clamp | Still fixed | `provider.go` and the mappers are unchanged. |
| WR-03 Undo slot | Still fixed | `store.go` and `App.svelte` are unchanged. |
| IN-06 (iter 1) UAT defaults | Still fixed | The `scripts/uat/*` files are unchanged. |
| IN-01 to IN-07 | Still present | Carried forward below. All line references still hold, because nothing above `actions.go:437` moved. |

## Critical Issues

### CR-01: The decode budget counts only `w*h`, so zero-pixel transparent frames still turn a 15 MB GIF into about 4 GB of live heap and 9 GB of process memory

**File:** `internal/services/actions.go:588-593` (the `case 0x2C` accounting), `actions.go:516-522` (the budget check), `actions.go:442`

**Issue:** `SquareMaxPixels` is meant to bound `DecodeAll` at "about 256 MB of `*image.Paletted`" (comment at line 442). That holds only if a frame's cost is its pixels. In `image/gif/reader.go`, `readImageDescriptor` also does three things for every frame, whatever its size:
- It allocates a fresh `*image.Paletted`.
- When the preceding Graphic Control extension sets the transparent flag, it **clones the palette**: `append(color.Palette(nil), d.globalColorTable...)`. That is 256 interface values, or 4 KB. If the index is out of range, `make(color.Palette, ti+1)` costs the same.
- It allocates a new `lzw.Reader` (about 20 KB of garbage).

A 0x0 frame is valid to Go: `readFull` on an empty `Pix` returns at once, the LZW read sees `ErrUnexpectedEOF`, which Go accepts, and the terminator closes the `blockReader`. `framePixels` adds `0*0 = 0` for it.

So each 20-byte unit `21 F9 04 01 00 00 00 00 | 2C 00..00 00 | 02 | 00` costs nothing against the budget and retains about 4.2 KB. `imaging.Square` then adds a 1x1 placeholder per frame, and `EncodeAll` writes all of them.

Reproduction (overlay test, no repo files changed). The file has a 16x16 canvas, a 256-entry global table, and 786,392 such frames:
```
file 15728622 bytes, frames 786392, framePixels = 0, <nil> (budget 268435456)
DecodeAll: 786392 frames in 9.19s; live heap after GC +3734 MB; TotalAlloc +20132 MB; Sys 7727 MB
copy square: staged and copied
CopySquare err = <nil>, calls ["CopyFile:.../raw-1-square.gif"], took 16.7s, Sys 9132 MB
```

The size is under the 15 MB `SquareMaxBytes` cap, and the canvas is far under `SquareMaxSide`. The budget reads zero, and the action succeeds. The live heap is about 14 times the intended bound, and the process footprint about 36 times. On an 8 GB Mac this means heavy swapping or a jetsam kill of the panel. A kill skips the quit-time staging cleanup, which is the failure mode CR-01 was opened for.

For the whole 17 s, `a.busy` is held and `DecodeAll` ignores `ctx`, so every other action returns `ErrBusy`. A hostile GIF only has to reach a provider's media URL, the same threat model as the original CR-01. The differential fuzz shows the parser now matches Go, so this is not a parsing bug. The metric itself leaves out a cost that Go pays.

**Fix:** Charge every frame a fixed overhead that covers the palette clone and the struct. That makes frame count part of the budget without rejecting real long animations:
```go
// frameOverhead is Go's retained per-frame cost in pixel-equivalents (bytes): a cloned
// 256-entry color.Palette (4 KB) plus the *image.Paletted and the delay/disposal entries.
const frameOverhead = 4096 + 256

case 0x2C:
	...
	px += int64(binary.LittleEndian.Uint16(d[4:]))*int64(binary.LittleEndian.Uint16(d[6:])) + frameOverhead
```
With a 256 Mi budget, this caps a file at about 60,000 frames, or about 260 MB retained in the worst case. A real 480x270 GIF with 200 frames uses about 27 M of the budget.

Add a regression test next to `TestCopySquareDecodeBudgetExtensions`. Build a GIF with a 256-entry global table and about 70,000 units of the transparent GCE plus a 0x0 frame. It must return `crop_failed` with no platform calls. You could also check `ctx.Err()` before `DecodeAll` so a cancelled action stops early, but that does not replace the budget.

## Info

### IN-01: Library mutation results are not ordered against an in-flight `refreshLibrary` (carried forward)

**File:** `frontend/src/App.svelte:270-283, 392-397, 767-776`
**Issue:** `refreshLibrary` discards stale answers using `libSeq`. `mutate()` and `removeChip()` assign `library` without bumping `libSeq`. A `Load` that took `Store.mu` before a ⌘⇧F or ⌘⌫ but answers after it overwrites the fresh snapshot: the star badge goes stale, or a removed item comes back. This is also one way to reach IN-06.
**Fix:** In `mutate` and `removeChip`, bump `libSeq` before the await and apply the result only if the sequence still matches.

### IN-02: ⌘⌫ and ⌘Z override text editing in the library filter and the action-menu filter (carried forward)

**File:** `frontend/src/App.svelte:549-550, 654`; `frontend/src/lib/grid.ts:80-81`
**Issue:**
- On a library tab with filter text typed, ⌘⌫ unfavorites or removes the selected item instead of deleting to the start of the line.
- `onMainKeydown` forwards every ⌘ key from the menu's filter to `onKeydown`, so ⌘⌫ there also removes the item.
- ⌘Z while an Undo toast is visible always runs the library Undo.
**Fix:** From the menu filter, forward only the shortcuts the menu lists. In the main input, take ⌘⌫ only when the input is empty or the caret is at position 0.

### IN-03: Error copy names the wrong failure for text copies, oversize or disallowed downloads, and `invalid_item` (carried forward)

**File:** `frontend/src/lib/errors.ts:103-121`
**Issue:**
- Text-action failures (`no_link`, `copy_failed`, `not_allowed`) fall through to "The download failed."
- Download `too_large` and `not_allowed` say "Check your connection".
- `invalid_item` still shows F5 "Try again", which can never succeed. It can happen when a Klipy item has an empty `slug`, or when a Klipy `md` URL is on a host outside `VendorAllow`.
**Fix:** Add rows for these kinds, and give `invalid_item` copy that does not suggest retrying.

### IN-04: An interrupted download leaves a hidden `.gifsmith-*.part` file in ~/Downloads (carried forward)

**File:** `internal/staging/staging.go:120-125`; `internal/services/actions.go:383-385`
**Issue:** `defer os.Remove(tmp)` does not run if the process quits or crashes mid-download. `ServiceShutdown` clears only staging.
**Fix:** Record the in-flight temp path under a mutex and remove it in `ServiceShutdown`. Or sweep stale `.gifsmith-*.part` files at the start of `Download`.

### IN-05: `refreshLibrary('')` does not select the first item, despite the comment (carried forward)

**File:** `frontend/src/App.svelte:438-449, 288-294`
**Issue:** When ⌘5 or ⌘6 is pressed on the tab that is already open, `selected` is not reset. `applyLibrary` with `want = ''` then clamps the current index, which contradicts the comment "'' = ... the first item".
**Fix:** In `applyLibrary`, set `selected = 0` when `want === ''` and the list is not empty, or correct the comment.

### IN-06: A no-op removal clears Go's Undo slot but leaves an earlier Undo toast on screen, so ⌘Z then restores nothing (carried forward)

**File:** `internal/library/store.go:339-342`; `frontend/src/App.svelte:333-338, 744, 752, 756`
**Issue:** `remove` with nothing to remove sets `s.undo = nil`. The frontend calls no `setStatus` on that path (`lib.canUndo` is false), so an earlier Undo toast stays on screen. Its ⌘Z then calls `Undo`, which restores nothing. Triggering this needs stale webview state (IN-01).
**Fix:** When a removal returns `canUndo === false`, also clear a visible Undo toast (`if (undoVisible()) setStatus('')`). Better still, have a removal return a token that `Undo(token)` must match.

### IN-07: `ClampID` can merge two distinct vendor ids that share a 200-byte prefix (carried forward)

**File:** `internal/provider/provider.go:46-56`; effects at `internal/library/store.go:84`, `internal/search/search.go:170-183`, `internal/staging/staging.go:65`
**Issue:** Truncating keeps only the prefix. Two items with ids over 200 bytes and the same first 200 bytes would:
- share one library key
- be merged by GIPHY's `dedupe`
- share a staging key

Real ids are much shorter.
**Fix:** If it matters, keep the uniqueness when clamping: take a 180-byte prefix (rune-safe), then `"-"` and the first 16 hex characters of `sha256(original)`.

---

_Reviewed: 2026-10-18T17:41:55Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
