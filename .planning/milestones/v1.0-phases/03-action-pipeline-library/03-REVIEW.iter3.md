---
phase: 03-action-pipeline-library
reviewed: 2026-10-18T17:32:23Z
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

# Phase 3: Code Review Report (iteration 2)

**Reviewed:** 2026-10-18T17:32:23Z
**Depth:** standard
**Files Reviewed:** 73
**Status:** issues_found

## Summary

This is a re-review after the iteration-1 fixes (1974e91 CR-01, 70b2c1e WR-01, eb1e74a WR-02, 195ff62 WR-03, 34436b8 IN-06). I checked each fix against the code and against the Wails v3.0.0-beta.30 and Go 1.27 `image/gif` sources. `go vet . ./internal/...` is clean, and `go test ./internal/...` passes.

**Status of the earlier findings**

| Prior | Status | Evidence |
|-------|--------|----------|
| CR-01 decode bomb | **Not fixed (bypassable)** | `framePixels` reads the Plain Text extension (0x01) differently from Go's decoder. A 16-byte prefix hides every frame from the budget. Reproduced: `framePixels` = 0, `gif.DecodeAll` decodes 20 frames of 4096x4096 (335 Mpx), and `CopySquare` succeeds. See CR-01. |
| WR-01 single instance | Fixed | `newSingleInstanceManager` runs inside `application.New`, before `stage.Clear()` (`main.go:124`), so a second process exits without wiping staging. `startSingleInstanceListener` runs only in `macosApp.run()`, after `sh` is assigned (`main.go:117`), so `sh` is never nil when the callback runs. The callback runs on a Wails goroutine, the same as the hotkey's `s.Toggle`. `Shell.Show` takes `s.mu`, and `FocusWebView` and `CursorScreenWorkArea` go through `onMain`. The tray and hotkey paths are unchanged. |
| WR-02 clamp | Fixed | `ClampID` backs up to `utf8.RuneStart`, so it never splits a rune. JSON-decoded strings are valid UTF-8, so the loop cannot run down to an empty id. Keyed `{#each}` blocks use the index, not the id, so a clamped duplicate cannot raise Svelte's duplicate-key error. There is a theoretical collision for ids longer than 200 bytes (IN-07). |
| WR-03 Undo slot | Fixed, with a narrow regression | A no-op removal clears the slot, and `canUndo` gates the toast. If a no-op removal happens while an earlier Undo toast is still on screen, that toast stays and its Undo now does nothing (IN-06). |
| IN-06 UAT defaults | Fixed | `: "${UAT_SEED_FAVORITES:={not json}"` gives `{not json` on macOS bash 3.2 (checked). Line 2 of each script names the fixture env it needs. |
| IN-01 to IN-05 | Still present | Carried forward unchanged. |

## Critical Issues

### CR-01: `framePixels` walks the Plain Text extension differently from `gif.DecodeAll`, so a 16-byte prefix hides every frame from the decode budget

**File:** `internal/services/actions.go:557-560` (the `case 0x21` branch of `framePixels`), `internal/services/actions.go:516-526`

**Issue:** `framePixels` treats every extension as a label byte followed by data sub-blocks. Go's decoder does not do this for the Plain Text extension. In `image/gif/reader.go:316-336`, `readExtension` handles `eText` (0x01) by reading **13 raw bytes** with `readFull(d.r, d.tmp[:13])`, whatever their values, and only then drains sub-blocks. The two parsers can disagree about where the extension ends.

Suppose the bytes after `0x21 0x01` are `0x00 0x3B` followed by 11 filler bytes and `0x00`:
- `framePixels` reads `0x00` as the block terminator, then reads `0x3B` as the trailer. It returns `px = 0, nil`.
- Go consumes the 13 bytes and the terminator, then decodes every image descriptor that follows.

Reproduction (overlay test, so no repo file changed): `bombGIF(t, 20)` with this 16-byte block inserted after the screen descriptor. The file is 232,150 bytes.

```
framePixels = 0, <nil> (budget 268435456)
DecodeAll decoded 20 frames of (0,0)-(4096,4096) = 335544320 px
CopySquare err = <nil>   (staged and copied)
```

A file at the 15 MB `SquareMaxBytes` cap therefore again allows about 1,400 full-canvas frames, roughly 22 GB once decoded. That is the out-of-memory kill the first review described, and the panel dies mid-action without running the quit-time staging cleanup.

The other extension types match Go:
- Graphic Control: Go reads a fixed 6 bytes and errors unless `[0]==4` and `[5]==0`.
- Application: a size byte plus that many bytes, which is the same as one sub-block.
- Comment: plain sub-blocks.
- Unknown labels: Go fails the decode.

Image data also matches: `blockReader.close` either drains to the terminator or errors. So 0x01 is the only gap, but it defeats the fix completely.

**Fix:** Make the extension branch follow Go's `readExtension` exactly, and reject labels that Go rejects:
```go
case 0x21:
	label, err := br.ReadByte()
	if err != nil {
		return 0, err
	}
	switch label {
	case 0x01: // Plain Text: Go reads 13 raw bytes, then sub-blocks
		if _, err := br.Discard(13); err != nil {
			return 0, err
		}
	case 0xF9: // Graphic Control: Go reads exactly 6 bytes and needs [0]==4, [5]==0; no sub-blocks follow
		var gc [6]byte
		if _, err := io.ReadFull(br, gc[:]); err != nil {
			return 0, err
		}
		if gc[0] != 4 || gc[5] != 0 {
			return 0, fmt.Errorf("gif: bad graphic control")
		}
		continue // the next byte is a block introducer
	case 0xFE, 0xFF: // Comment / Application: sub-blocks (Application's size byte is the first sub-block's length)
	default:
		return 0, fmt.Errorf("gif: unknown extension 0x%02x", label) // Go fails the decode too
	}
```
Add a regression test that inserts `21 01 00 3B` + 11 bytes + `00` before the frames of `bombGIF`. It must return `crop_failed` with no platform calls. A second, stronger guard would make the parser mismatch harmless: cap the frame count independently. Count `0x2C` introducers and refuse when `count × cfg.Width × cfg.Height` exceeds the budget. Because every frame must fit inside the logical screen, the canvas area is an upper bound on each frame.

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

### IN-03: Error copy names the wrong failure for text copies, oversize or disallowed downloads, and `invalid_item` (carried forward, extended)

**File:** `frontend/src/lib/errors.ts:103-121`
**Issue:**
- Text-action failures (`no_link`, `copy_failed`, `not_allowed`) fall through to "The download failed."
- Download `too_large` and `not_allowed` say "Check your connection".
- WR-02 removed the long-title cause of `invalid_item`, but the error can still occur, and it still shows F5 "Try again", which can never succeed. Two cases: a Klipy item with an empty `slug` (empty `ID`), and a Klipy `md` URL on a host outside `VendorAllow`.
**Fix:** Add rows for these kinds, and give `invalid_item` copy that does not suggest retrying.

### IN-04: An interrupted download leaves a hidden `.gifsmith-*.part` file in ~/Downloads (carried forward)

**File:** `internal/staging/staging.go:120-125`; `internal/services/actions.go:383-385`
**Issue:** `defer os.Remove(tmp)` does not run if the process quits or crashes mid-download. `ServiceShutdown` clears only staging.
**Fix:** Record the in-flight temp path under a mutex and remove it in `ServiceShutdown`. Or sweep stale `.gifsmith-*.part` files at the start of `Download`.

### IN-05: `refreshLibrary('')` does not select the first item, despite the comment (carried forward)

**File:** `frontend/src/App.svelte:438-449, 288-294`
**Issue:** When ⌘5 or ⌘6 is pressed on the tab that is already open, `selected` is not reset. `applyLibrary` with `want = ''` then clamps the current index, which contradicts the comment "'' = ... the first item".
**Fix:** In `applyLibrary`, set `selected = 0` when `want === ''` and the list is not empty, or correct the comment.

### IN-06: A no-op removal clears Go's Undo slot but leaves an earlier Undo toast on screen, so ⌘Z then restores nothing (new, from the WR-03 fix)

**File:** `internal/library/store.go:339-342`; `frontend/src/App.svelte:333-338, 744, 752, 756`
**Issue:** `remove` with nothing to remove now sets `s.undo = nil`. On that path the frontend calls no `setStatus` (`lib.canUndo` is false), and an Undo toast survives keystrokes (`undoVisible()` at line 335). Scenario:
1. The user unfavorites X, and the Undo toast appears.
2. Within 5 s, ⌘⇧F on a tile with a stale star (IN-01), or ⌘⌫ on a stale Recents entry, sends a no-op removal.
3. Go clears the slot, but X's Undo toast is still on screen.
4. ⌘Z calls `Undo`, which returns nil with nothing restored. X is gone even though the user pressed Undo.

This needs stale webview state to trigger, hence Info.
**Fix:** When a removal returns `canUndo === false`, also clear a visible Undo toast (`if (undoVisible()) setStatus('')`). Better still, have a removal return a token that `Undo(token)` must match.

### IN-07: `ClampID` can merge two distinct vendor ids that share a 200-byte prefix (new, from the WR-02 fix)

**File:** `internal/provider/provider.go:46-56`; effects at `internal/library/store.go:84`, `internal/search/search.go:170-183`, `internal/staging/staging.go:65`
**Issue:** Truncating keeps only the prefix. Klipy slugs put their unique suffix at the end. Two items whose ids exceed 200 bytes and share the first 200 would then:
- share one library key (favoriting one stars both, and Touch replaces the other in Recents)
- be merged by GIPHY's `dedupe`
- share a staging key

Real GIPHY ids are about 18 bytes, and Klipy slugs this long are unlikely. Before the fix such items could not be saved at all, so this is a net improvement.
**Fix:** If it matters, keep the uniqueness when clamping: take a 180-byte prefix (rune-safe), then `"-"` and the first 16 hex characters of `sha256(original)`.

---

_Reviewed: 2026-10-18T17:32:23Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
