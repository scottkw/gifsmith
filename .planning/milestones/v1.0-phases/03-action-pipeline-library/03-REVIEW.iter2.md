---
phase: 03-action-pipeline-library
reviewed: 2026-10-18T17:02:19Z
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
  warning: 3
  info: 6
  total: 10
status: issues_found
---

# Phase 3: Code Review Report

**Reviewed:** 2026-10-18T17:02:19Z
**Depth:** standard
**Files Reviewed:** 73
**Status:** issues_found

## Summary

I reviewed the Phase 3 diff (`a037917..HEAD`) for the action pipeline, staging, square crop, library store and service, platform cgo additions, provider mappers, and the Svelte keyboard, menu, details and library code. `go vet` is clean, and the library, staging, services and imaging tests pass.

**Desktop library persistence (the evidence-gap question).** The code path works. The missing files come from the session, not from a code bug.

- **Wiring.** `wire_desktop.go:37` sets `ConfigDir` to `os.UserConfigDir()/com.kenscott.gifsmith` (`~/Library/Application Support/com.kenscott.gifsmith`). `main.go:54` creates `library.Store{Dir: d.ConfigDir, Allow: d.Allow}`, and `main.go:55` sets `actions.Recents = lib`. `LibraryService` is registered at `main.go:91`. Writes go through `settings.WriteJSONAtomic`, which runs `MkdirAll` with mode 0700. The bindings `frontend/bindings/.../libraryservice.ts` exist. The build is not sandboxed: `Info.plist` has no sandbox entitlement and `install:dev` signs only with hardened runtime.
- **Probe.** I ran a probe test through a `go test -overlay`, so no repo file changed. It sent a realistic GIPHY payload through the real `giphy.Client` mapper, then called `SetFavorite`, `Touch` and `SaveSearch` on a `library.Store` that uses `staging.VendorAllow` (the desktop allowlist). All three files were written with mode 0600.
- **Frontend.** ⌘⇧F goes `routeKey → 'favorite' → runAction → toggleFavorite → mutate(setFavorite)`. A successful copy or paste goes `ActionService.Copy/Paste/CopyText/PasteText/CopySquare/PasteSquare → a.recent → Store.Touch`.
- **Log evidence.** `bin/install-dev.log` is the only stderr capture for the installed process. It has only `tray menu ready`, `hotkey registered`, `clips: …`, three `download giphy: saved` lines and one `show download: revealed`. Every copy or paste logs a line whether it succeeds (`copy: staged and copied`, `copy text: …`, `paste: …`) or fails (`stage …`, `copy text …`). A favorite either writes `favorites.json` or logs `library: …` / `library write: …`. There is no file and no such log line, so no copy, paste or favorite request reached this process.
- **Staging dir.** Nothing in the code removes the staging directory itself. `Clear`, `Sweep` and `Remove` only delete entries under it, and `Remove` refuses `rel == "."`. The first successful stage creates `~/Library/Caches/com.kenscott.gifsmith/staging` with `MkdirAll`, and it then stays. So its absence means no copy has ever succeeded at that path on this machine. Downloads correctly bypass staging (`Save` writes only into `DownloadsDir`).
- **Verifier hypothesis (not a finding).** UAT steps 9, 12 and 13 may have gone to a different Gifsmith process. There is no single-instance guard, and a `.dev` bundle and the installed bundle can both register `Ctrl+Option+Cmd+G` (see WR-01). The other possibility is that those steps were not run. Steps 9, 12 and 13 should be re-run against the installed build, with `pgrep -lf gifsmith` checked first.

**Other findings.**
- The square crop's limits do not bound decoded memory. A 15 MB GIF can make `gif.DecodeAll` allocate tens of GB (CR-01, measured).
- Library integrity has gaps: no single-instance guard (WR-01), items are rejected instead of truncated (WR-02), and the single Undo slot can come back stale (WR-03).

These areas held up when I traced them:
- **Downloads:** `os.Link` claim, no overwrite, the stem limited to ASCII-safe characters, and the `filepath.Dir` check.
- **URL handling:** link, page and source URLs go through Go allow-lists, and `openExternal` only accepts its fixed list.
- **Klipy grace delete:** the `gen` check runs under `busy`.
- **Main-thread calls:** `OpenURL` and `RevealFile` go through `onMain`.

## Critical Issues

### CR-01: The square-crop limits do not stop a decompression bomb, so `gif.DecodeAll` can exhaust memory and crash the app

**File:** `internal/services/actions.go:434-438, 495-520`
**Issue:** `cropSquare` enforces two limits: 15 MB compressed (`FetchMax`) and a logical screen of at most 4096x4096 (`DecodeConfig`). Then it calls `gif.DecodeAll`, which keeps every frame in memory as a full `*image.Paletted`. Go's decoder only requires each frame to fit inside the logical screen (`reader.go:517`), so a single frame can be 4096x4096 = 16 MB. Nothing limits the number of frames.

I measured this with Go's own encoder (which inserts clear codes): a uniform 4096x4096 frame compresses to about 11 KB. A 15 MB file can therefore hold about 1,400 such frames, and `DecodeAll` would allocate about 22 GB. An encoder that defers clear codes, which Go's LZW reader supports, packs about 2x more frames into the same file.

The comment at line 434 calls this a mitigation for "T-03-18 / Pitfall 2", but the size and side checks bound the input, not the decoded output. Any GIF on an allow-listed vendor CDN (for example a user upload on GIPHY or Klipy) can crash the process with an out-of-memory kill. That takes the hotkey panel down mid-action, and the staging cleanup at quit never runs.
**Fix:** Bound the total decoded pixel count before `DecodeAll`. Walk the GIF block structure once and sum frame areas (image descriptors are cheap to skip), or refuse when the frame count times the canvas size exceeds a budget:
```go
const squareMaxPixels = 256 << 20 // total decoded pixels across all frames (~256 MB of Paletted)

// framePixels walks the GIF blocks after the header/LSD/GCT and sums each image
// descriptor's w*h, skipping extension and LZW sub-blocks without decoding them.
func framePixels(r io.Reader) (int64, error) { /* ~40 lines: read 0x21/0x2C/0x3B introducers */ }

// in cropSquare, after DecodeConfig and before Seek/DecodeAll:
if px, err := framePixels(f); err != nil || px > squareMaxPixels {
	return nil, fmt.Errorf("decoded size over budget")
}
```
Add a test with a few hundred uniform 4096x4096 frames, built with `gif.EncodeAll` in the test, that must fail fast with `crop_failed` / `crop_too_large` and must not call `DecodeAll`.

## Warnings

### WR-01: No single-instance guard. Two Gifsmith processes share the library and staging files, which loses Favorites/Recents updates and wipes the other instance's clipboard file

**File:** `main.go:84-103`, `wire_desktop.go:37-38`, `internal/library/store.go:147-164, 186-196`
**Issue:** `application.Options` has no `SingleInstance`. `ConfigDir` and `StagingDir` are hard-coded to `com.kenscott.gifsmith` whatever the bundle id is, so `com.kenscott.gifsmith.dev` (from `wails3 dev`) and the installed `com.kenscott.gifsmith` use the same files. LaunchServices does not deduplicate across bundle ids or across `open -n`.

Phase 3 adds a new risk. `library.Store` reads each list once (`l.ok = true`) and never re-reads it, and every mutation writes the whole cached list. Suppose instance A has cached `[X]` and instance B adds Y, so the file holds `[Y, X]`. When A then adds Z, it writes `[Z, X]` and Y is silently lost. The same lost update applies to Recents (every copy is a `Touch`) and to searches.

`stage.Clear()` at launch (`main.go:116`) also deletes the file the other instance's clipboard points at. This was already noted in 03-03-SUMMARY, but the library now makes it a data-loss issue.

Both builds also register the same Carbon hotkey. macOS does not reject the duplicate, so the user's keystrokes may go to the wrong process. That could explain the missing files in the native evidence gap.
**Fix:** Set `SingleInstance: &application.SingleInstanceOptions{UniqueID: "com.kenscott.gifsmith", OnSecondInstanceLaunch: func(application.SecondInstanceData) { sh.Show() }}`. Separately, derive the dev build's directories from the bundle id, for example a `.dev` suffix under a dev build tag. If the store must tolerate other writers, re-read the file under `mu` before each mutation (compare mtime and size) instead of trusting the cache.

### WR-02: `Validate` rejects long titles and ids instead of truncating them, and the mappers never cap them. Copy works, but Recents silently skips the item and Favorite fails with a misleading "Try again"

**File:** `internal/library/store.go:90-95, 228-231, 242-246`; `internal/provider/finergifs/finergifs.go:74-79`; `internal/provider/giphy/giphy.go` (title is not capped); `internal/provider/klipy/klipy.go` (title is not capped)
**Issue:** `Validate` returns `ErrInvalid` when `Title` is longer than 500 runes or `ID` is longer than 200 bytes. No mapper caps `Title`. Finer Gifs uses the subtitle `text` field verbatim, and GIPHY and Klipy titles come from user uploads. For such an item:
- `Copy`/`Paste` succeed, `a.recent` logs `recents: copy: invalid library item: title`, and the item never shows up in Recents (D-07 breaks silently).
- ⌘⇧F returns `invalid_item`, and the UI shows F5 "Couldn't save your Favorites. Try again." (`errors.ts:118`). Retrying can never succeed.

The Klipy mapper already truncates tags to the same limits, so the title path is simply inconsistent.
**Fix:** Truncate at the source and keep `Validate` strict for tampered files:
```go
// provider package
func ClampTitle(s string) string {
	if r := []rune(s); len(r) > 500 { return string(r[:500]) }
	return s
}
```
Call it in all three mappers. Alternatively, have `Touch`/`SetFavorite` clamp `g.Title` before `Validate`. Also map `invalid_item` to its own message in `errors.ts`, not F5's "Try again".

### WR-03: A removal that changes nothing still reports success and shows an Undo toast, and ⌘Z then restores an earlier, unrelated removal

**File:** `internal/library/store.go:256-258, 277, 334-337`; `frontend/src/App.svelte:741-756`
**Issue:** `remove(l, nil)` returns `nil` and leaves `s.undo` alone. So `SetFavorite(g, false)` for an item that is not a favorite, `RemoveRecent` for a missing item, and `ClearRecents` on an empty list all succeed without replacing the slot. The frontend then shows "Removed from Favorites" / "Removed from Recents" / "Cleared Recents" with Undo (⌘Z). Undo restores whatever the last real removal was, possibly from the other list, minutes earlier, and deliberately removed.

The webview's `favKeys` can be stale (see IN-01, or another instance per WR-01). In that case ⌘⇧F sends `on=false` for an item Go does not hold, and the Undo brings back a different favorite. The slot never expires in Go, so a stale slot survives indefinitely.
**Fix:** In `remove`, clear the slot when nothing was removed (`s.undo = nil` before `return nil`), or report the no-op (for example a `Changed bool` in `Snapshot`) so the UI skips the toast. Also bind the Undo call to the removal it belongs to. One way: return a removal token from the mutation, pass it to `Undo(token)`, and have Go refuse a mismatched token.

## Info

### IN-01: Library mutation results are not ordered against an in-flight `refreshLibrary`

**File:** `frontend/src/App.svelte:270-283, 396, 765-773`
**Issue:** `refreshLibrary` drops stale answers with `libSeq`. `mutate()` and `removeChip()` assign `library` without bumping `libSeq`. Every `panel:show` and every `switchLib` starts a `Load`. If that `Load` took `Store.mu` before a quick ⌘⇧F or ⌘⌫ but its response arrives after the mutation's, the pre-mutation snapshot overwrites the fresh one: the star badge is wrong and the removed item reappears. This window is narrow, and it is a source of the stale `favKeys` in WR-03.
**Fix:** In `mutate`/`removeChip`, run `++libSeq` before awaiting and apply the result only if the sequence still matches, the same way `refreshLibrary` does.

### IN-02: ⌘⌫ and ⌘Z override text editing in the library filter and in the action-menu filter

**File:** `frontend/src/App.svelte:549-550, 654`; `frontend/src/lib/grid.ts:80-81`
**Issue:** On a library tab with a filter typed, ⌘⌫ (macOS "delete to line start") unfavorites or removes the selected item instead of editing the text. While ⌘K is open, `onMainKeydown` forwards every ⌘ key from the menu's filter field to `onKeydown`, so ⌘⌫ there also removes the library item. ⌘Z while an Undo toast is visible always runs the library Undo, even right after typing. Undo limits the damage, but the action is destructive.
**Fix:** In the menu filter, forward only the shortcuts the menu lists (⌘K, ⌘↵, ⌘⇧C/M/F, ⌘O, ⌘S, ⌘I), not ⌘⌫/⌘Z. For the main input, consider capturing ⌘⌫ only when the input is empty or the caret is at position 0.

### IN-03: Error copy names the wrong failure for text copies and oversize or disallowed downloads

**File:** `frontend/src/lib/errors.ts:103-115`
**Issue:** `textAction` failures (`no_link`, `copy_failed`, `not_allowed` from Copy Link or Markdown) fall through to "Couldn't copy this GIF. The download failed." No download happens for a text copy. A download that fails with `too_large` (over 50 MB) or `not_allowed` shows "Check your connection, then try again", which does not fix it.
**Fix:** Add rows for `no_link`, `copy_failed` and `not_allowed` on the copy path, and for `too_large` on the download path.

### IN-04: An interrupted download leaves a hidden `.gifsmith-*.part` file in ~/Downloads

**File:** `internal/staging/staging.go:120-125`
**Issue:** `Save` cleans up its temp file with `defer os.Remove(tmp)`, which never runs if the process quits (⌘Q mid-download, SIGTERM from `install:dev`'s `pkill`) or crashes. `ServiceShutdown` clears only staging. The result is a hidden, partial, 0600/0644 file left in the user's Downloads folder.
**Fix:** Have `ServiceShutdown` remove the in-flight temp file, by recording its path under a small mutex. Or sweep `.gifsmith-*.part` files older than an hour from `DownloadsDir` at the start of each `Download`.

### IN-05: `refreshLibrary('')` does not select "the first item" as the comment says

**File:** `frontend/src/App.svelte:443-448, 288-292`
**Issue:** When ⌘5/⌘6 is pressed again on the tab that is already open, `selected` is not reset (the `libTab !== id` branch is skipped). `applyLibrary` with `want = ''` then clamps the current index instead of choosing index 0, which contradicts the comment at line 448.
**Fix:** Set `selected = 0` when `want === ''` and `libItems.length > 0`, or fix the comment.

### IN-06: The default UAT loop cannot pass `12-library-unreadable.js`

**File:** `scripts/uat/12-library-unreadable.js`, `scripts/smoke-browser-loop.sh:50-53`
**Issue:** The script only passes with `UAT_SEED_FAVORITES='{not json}'`. A plain `for f in scripts/uat/*.js` loop reports `UAT FAIL`. This already caused a deviation in 03-08.
**Fix:** Have the script declare its seed in a header that the loop reads, or have `smoke-browser-loop.sh --uat` apply a default seed for `*-unreadable.js`.

---

_Reviewed: 2026-10-18T17:02:19Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
