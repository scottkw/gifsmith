---
phase: 03-action-pipeline-library
verified: 2026-10-18T18:05:00Z
status: human_needed
score: 15/19 must-haves verified
covered_files: [".planning/phases/03-action-pipeline-library/03-01-PLAN.md",".planning/phases/03-action-pipeline-library/03-01-SUMMARY.md",".planning/phases/03-action-pipeline-library/03-02-PLAN.md",".planning/phases/03-action-pipeline-library/03-02-SUMMARY.md",".planning/phases/03-action-pipeline-library/03-03-PLAN.md",".planning/phases/03-action-pipeline-library/03-03-SUMMARY.md",".planning/phases/03-action-pipeline-library/03-04-PLAN.md",".planning/phases/03-action-pipeline-library/03-04-SUMMARY.md",".planning/phases/03-action-pipeline-library/03-05-PLAN.md",".planning/phases/03-action-pipeline-library/03-05-SUMMARY.md",".planning/phases/03-action-pipeline-library/03-06-PLAN.md",".planning/phases/03-action-pipeline-library/03-06-SUMMARY.md",".planning/phases/03-action-pipeline-library/03-07-PLAN.md",".planning/phases/03-action-pipeline-library/03-07-SUMMARY.md",".planning/phases/03-action-pipeline-library/03-08-PLAN.md",".planning/phases/03-action-pipeline-library/03-08-SUMMARY.md","build/darwin/Info.dev.plist","build/darwin/Info.plist","frontend/bindings/gifsmith/internal/library/index.ts","frontend/bindings/gifsmith/internal/library/models.ts","frontend/bindings/gifsmith/internal/provider/models.ts","frontend/bindings/gifsmith/internal/services/actionservice.ts","frontend/bindings/gifsmith/internal/services/index.ts","frontend/bindings/gifsmith/internal/services/libraryservice.ts","frontend/bindings/gifsmith/internal/services/models.ts","frontend/src/App.svelte","frontend/src/app.css","frontend/src/components/ActionMenu.svelte","frontend/src/components/DetailsPane.svelte","frontend/src/components/Footer.svelte","frontend/src/components/LibraryGrid.svelte","frontend/src/components/ProviderTabs.svelte","frontend/src/components/RecentSearches.svelte","frontend/src/components/ResultGrid.svelte","frontend/src/components/SetupCard.svelte","frontend/src/components/Tile.svelte","frontend/src/lib/actions.test.ts","frontend/src/lib/actions.ts","frontend/src/lib/api.ts","frontend/src/lib/errors.test.ts","frontend/src/lib/errors.ts","frontend/src/lib/format.test.ts","frontend/src/lib/format.ts","frontend/src/lib/grid.test.ts","frontend/src/lib/grid.ts","frontend/src/lib/library.test.ts","frontend/src/lib/library.ts","frontend/src/lib/popover.test.ts","frontend/src/lib/popover.ts","frontend/src/lib/providers.ts","frontend/src/lib/searches.test.ts","frontend/src/lib/searches.ts","internal/imaging/square.go","internal/imaging/square_test.go","internal/library/store.go","internal/library/store_test.go","internal/platform/fake.go","internal/platform/platform.go","internal/platform/platform_darwin.go","internal/platform/platform_darwin.h","internal/platform/platform_darwin.m","internal/platform/platform_darwin_test.go","internal/platform/platform_stub.go","internal/provider/fakeapi/fakeapi.go","internal/provider/finergifs/finergifs.go","internal/provider/finergifs/finergifs_test.go","internal/provider/giphy/giphy.go","internal/provider/giphy/giphy_test.go","internal/provider/httpx/httpx.go","internal/provider/klipy/klipy.go","internal/provider/klipy/klipy_test.go","internal/provider/provider.go","internal/provider/provider_test.go","internal/services/actions.go","internal/services/actions_test.go","internal/services/errors.go","internal/services/library.go","internal/services/library_test.go","internal/services/settings_test.go","internal/settings/settings.go","internal/settings/settings_test.go","internal/staging/staging.go","internal/staging/staging_test.go","main.go","scripts/smoke-browser-loop.sh","scripts/uat/01-tracer.js","scripts/uat/02-keyboard.js","scripts/uat/05-tabs.js","scripts/uat/06-clips.js","scripts/uat/07-text-actions.js","scripts/uat/08-action-menu.js","scripts/uat/10-download.js","scripts/uat/11-square.js","scripts/uat/12-library-unreadable.js","scripts/uat/12-library.js","scripts/uat/13-chips.js","scripts/uat/13-details.js","scripts/uat/14-backstops.js","wire_desktop.go","wire_server.go"]
covered_digest: "v3:sha256:1f55bf98995a490a6cc0cdeb92a715d41a82abc68bd5e9941f63f49a657f32f8"
behavior_unverified: 2
overrides_applied: 0
deferred:
  - truth: "User can download a GIF to a CHOSEN folder (SC2 / ACTN-05): Phase 3 ships a fixed, injected destination (~/Downloads) with no picker"
    addressed_in: "Phase 4"
    evidence: "Phase 4 SC3: 'The user can set ... download folder, hide filename ...' (SETT-04); 03-CONTEXT Deferred Ideas: 'download folder picker and hide-filename toggle (SETT-04) — Phase 4'"
  - truth: "Optional hide-original-filename setting (ACTN-11): FileStem(g, hide=true) implements the gifsmith-<id> name, but Download always passes hide=false and no toggle exists"
    addressed_in: "Phase 4"
    evidence: "Phase 4 SC3 / SETT-04 'hide filename'"
behavior_unverified_items:
  - truth: "A panel:show cancels a pending hide-after-copy timer, so re-showing within 400 ms of a copy does not hide the panel again (03-01, 02 IN-07)"
    test: "Select a GIF, press Enter, and within about 0.4 s click outside the panel (it hides), then press the hotkey to show it again"
    expected: "The panel reappears and stays open; it is not hidden again by the copy's 400 ms timer"
    why_human: "App.svelte:242 clears hideTimer in the panel:show handler, but no unit or UAT test emits panel:show inside the 400 ms window; the server build cannot show/hide a native panel"
  - truth: "Staging is wiped at quit (ActionService.ServiceShutdown) on the installed app (03-03, ACTN-10, D-17)"
    test: "After a copy, run: ls ~/Library/Caches/com.kenscott.gifsmith/staging (one hash directory). Quit Gifsmith from the menu bar (and separately with Cmd+Q while the panel has focus), then run the ls again"
    expected: "The staging directory is empty or absent after each quit"
    why_human: "TestServiceShutdownClears proves the method clears staging and Wails beta.30 application.go:821 calls ServiceShutdown during cleanup, but no test drives a real Quit/Cmd+Q, and 03-08 recorded that the installed process was never quit"
human_verification:
  - test: "Fresh install of the current HEAD: run `wails3 task darwin:install:dev`, then double-click ~/Applications/Gifsmith.app a second time (WR-01 single instance)"
    expected: "bin/install-dev.log shows 'tray menu ready' and 'hotkey registered'; only one menu bar icon exists; the second launch shows the existing panel and exits; `git diff --quiet -- go.sum` still passes (restore objx lines if added)"
    why_human: "Native launch, SingleInstance and codesigning cannot run headless; the installed build predates the review fixes (WR-01..03, CR-01)"
  - test: "Text actions in real targets: open TextEdit, show the panel, search 'cat', press Cmd+Shift+C, then Cmd+V in TextEdit; Cmd+K -> Paste Link (Accessibility granted); Cmd+Shift+M then Cmd+V; Cmd+O on a GIPHY GIF; Cmd+O on a Klipy GIF"
    expected: "The link is a giphy.com media URL with no '?' query; Paste Link drops the URL into TextEdit without a manual Cmd+V; Markdown is ![title](url); Cmd+O opens the giphy.com page in the default browser; Cmd+O on Klipy does nothing"
    why_human: "Headless UAT runs platform.Fake: the real NSPasteboard string write, CGEvent Cmd+V and NSWorkspace open are only exercised natively"
  - test: "Paste Square GIF: select a wide GIPHY GIF, Cmd+K -> type 'square' -> Paste Square GIF into a Messages conversation"
    expected: "A square, still-animated GIF appears in the message field"
    why_human: "The crop is golden-tested in Go, but how Messages renders the staged square file needs a real paste"
  - test: "Klipy 30 s grace delete: on the Klipy tab press Cmd+Enter into Messages; immediately run ls ~/Library/Caches/com.kenscott.gifsmith/staging; run it again after 35 s. Separately, Cmd+Enter into Slack and send about 40 s later"
    expected: "One staged directory right after the paste, none after 35 s; the Slack message still sends the animated GIF"
    why_human: "TestKlipyGraceDeletesAfterPaste/GenerationGuard inject the timer; real timing and whether Slack has read the file before deletion are native-only (D-19, A2)"
  - test: "Clips MP4 paste (only if the GIPHY key is Clips-approved): Cmd+2, Enter on a clip, Cmd+V into Messages"
    expected: "An MP4 attachment (720p) pastes; footer said 'Copied MP4 — press ⌘V to paste'. If the key is not approved, record 'not verifiable natively (fixture-proven only)'"
    why_human: "Requires an approved Clips key; 03-08 step 8 was skipped (A1)"
  - test: "Library persistence and file mode: copy two GIFs from different providers, Cmd+Shift+F one, run a search and wait 2 s. Run ls -l ~/Library/Application\\ Support/com.kenscott.gifsmith/. Quit from the menu bar, relaunch, press Cmd+6 and Cmd+5, then clear the box"
    expected: "favorites.json, recents.json and searches.json exist with mode -rw-------; after relaunch Recents shows both GIFs grouped by provider (newest first), Favorites shows the starred one, and the recent-search chip appears in the empty box"
    why_human: "TestRecentsPersistAcrossStores and 12-library.js prove the round-trip in a temp dir; 03-08 found no library files on the installed app, so the real Application Support path and Quit/relaunch were never observed"
  - test: "WKWebView key handling: on a provider tab type 'hello' then Cmd+Backspace and Cmd+Z; on Recents select a tile and press Cmd+Backspace then Cmd+Z; with no download toast press Cmd+R; press Cmd+I then Esc; press Cmd+S then Cmd+R within 5 s"
    expected: "Provider tab: text deletes and the undo restores it natively. Recents: the item is removed with the 'Removed from Recents' toast and Cmd+Z restores it in place. Cmd+R does not reload (query and selection kept). Cmd+I opens Details, Esc closes it with the selection kept. Cmd+S saves to Downloads and Cmd+R reveals the file in Finder"
    why_human: "Headless Chromium proves the JS routing; WKWebView key-equivalent handling against the native menu can differ (A4)"
  - test: "Process prohibitions (03-01, 03-08, judgment tier): confirm no executor sent synthetic desktop input, killed processes it did not start, or read the Keychain / ~/Downloads during the phase"
    expected: "No such event observed by the user"
    why_human: "unverified-prohibition — human review recommended. Non-authoritative LLM-judge verdict: no evidence of violation (all UAT ran headless against -tags server with platform.Fake and bin/uat-* dirs; 03-08 SUMMARY records read-only checks only)"
---

# Phase 3: Action Pipeline & Library Verification Report

**Phase Goal:** Every action the Raycast extension offered works on the selected GIF, and favorites, recents, details, and recent searches put the GIFs a user reuses one keystroke away, even offline or rate-limited.
**Verified:** 2026-10-18T18:05:00Z
**Status:** human_needed
**Re-verification:** No (initial verification)

## Goal Achievement

The code delivers the goal. Every action, the ⌘K menu, download, square crop, Clips MP4, staging cleanup, Favorites/Recents, Details and recent-search chips exist, are wired Go to webview, and pass the unit tests and headless UAT that I re-ran myself on HEAD 9824e93. CR-01 is closed. What remains is native evidence: the installed build predates the review fixes, and its logs show only download and reveal activity. The items below need a fresh `install:dev` of HEAD.

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1: ⌘K lists every action for the selected GIF; copy/paste link, copy/paste Markdown, copy page link, Open in Browser | ✓ VERIFIED | `lib/actions.ts` ROWS (locked order, omit-not-grey); `App.svelte:684` `runAction` is the single dispatcher; Go builds every string (`actions.go` `textFor`/`linkFor`/`markdownFor`/`pageFor`); tests TestCopyTextLink/Markdown/Page/ClipLink, TestPasteText*, TestOpenPage; UAT 07-text-actions, 08-action-menu OK (my run) |
| 2 | SC2: download without overwrite; non-GIF/MP4 rejected; square crop not offered for Clips; Clips labelled MP4, 720p default | ✓ VERIFIED | `staging.Save` (CreateTemp + `os.Link` loop, never MkdirAll, 0644); `media()` checks GIF87a/89a or `ftyp` before any file exists; `giphy.go:147` 720p→480p→1080p→360p; square rows `applies: !isClip`, Go refuses `IsClip`; UAT 10-download, 11-square, 06-clips OK. Folder **picker** deferred to Phase 4 (see Deferred) |
| 3 | SC3: copy/paste/download temp files removed promptly; Klipy eagerly | ✓ VERIFIED | `stage.Clear()` before `app.Run()` (main.go); `Sweep(path)` after each staged copy; `putText` clears staging; Klipy `expire` with generation guard; downloads bypass staging; tests TestCopyTwiceKeepsOnlyLatest, TestCopyTextClearsStaging, TestKlipyGrace* (injected After), TestServiceShutdownClears |
| 4 | SC4: add/remove favorite; every copy/paste in Recents (newest first, capped); grouped by provider, never blended; remove from Recents; render from saved metadata without API calls | ✓ VERIFIED | `library/store.go` (RecentsMax 50, dedupe by provider:id, undo slot, corrupt-file refusal, WriteJSONAtomic 0600); `ActionService.recent` after every Copied result except page; `library` has no net/http in its deps (`go list -deps`); UAT 12-library (incl. rate-limited GIPHY), 12-library-unreadable (file bytes unchanged) OK |
| 5 | SC5: Details view (title, size, dimensions, tags, source links); empty box shows recent searches that rerun | ✓ VERIFIED | `DetailsPane.svelte` renders only carried fields, links open through `openPage`/`openSource` (Go re-validates); `searches.ts` chipsVisible/shouldSave; `SaveSearch` dedupe/cap 20; UAT 13-details, 13-chips OK |
| 6 | 03-01: link/Markdown/page rules (GIPHY query stripped, Klipy/Finer byte-for-byte, Clips link = GIF rendition), Go-only URL opening, Fake platform in server build, F1/F3 fixed copy, optional metadata fields | ✓ VERIFIED | `linkFor` strips query only for GIPHY/Clips; `wire_server.go` `Platform: &platform.Fake{}`; MarshalError composed in main.go; giphy/klipy mapper tests |
| 7 | 03-01: panel:show cancels a pending hide-after-copy timer | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED | `App.svelte:242` `clearTimeout(hideTimer)`; no test exercises it (cancellation invariant) |
| 8 | 03-02: ⌘K popover (filter, ↑↓, ↵, Esc/⌘K close, outside click, shortcuts while open, "No matching actions", ≥8 px placement backstop) | ✓ VERIFIED | `ActionMenu.svelte`, `popover.ts`; UAT 08-action-menu asserts 8 px inside 720x480 for tile 0 and bottom-row columns 1 and 4 (`08-action-menu.js:133-138`); no transition/animation in ActionMenu CSS |
| 9 | 03-03: Clips MP4 copy/paste, ftyp validation, MP4 badge, E8 gone; staging lifecycle; Klipy 30 s grace with generation guard | ✓ VERIFIED | `Tile.svelte` MP4 badge + aria-label; no "arrives in the next update" in `frontend/src`; tests TestCopyClipMP4, TestPasteClipMP4, TestCopyClipFailingMediaKinds, TestKlipyGraceGenerationGuard, TestKlipyGraceOnlyAfterKlipyPaste |
| 10 | 03-03: staging wiped at quit on the installed app | ⚠️ PRESENT_BEHAVIOR_UNVERIFIED | Method tested (TestServiceShutdownClears); Wails hook confirmed in `wails/v3@v3.0.0-beta.30/pkg/application/application.go:821`; no real Quit observed (03-08) |
| 11 | 03-04: ⌘S download naming, (2)/(3) no-overwrite, F2/F3/F4, 0644, path containment, Show in Finder via Go, ⌘R captured, not in Recents | ✓ VERIFIED | `Save`, `FileStem`; `Download` has no `recent()` call; `ShowDownload` takes no path; Info.plist + Info.dev.plist NSDownloadsFolderUsageDescription; tests TestDownload*, TestShowDownload*; UAT 10-download |
| 12 | 03-05: lossless square crop; F6/F7; 300 ms busy threshold; staging keeps only the square; decode budget closes CR-01 | ✓ VERIFIED | `imaging.Square` (SubImage, no re-quantize; golden tests); `framePixels` + `frameOverhead`; TestCopySquareDecodeBudget, …Extensions, …EmptyFrames pass; my adversarial overlay test (below) |
| 13 | 03-06: library tabs ⌘5/⌘6, sections in slot order with "{label} · {count}", local title filter, ⌘⇧F star, ⌘⌫ remove + 5 s Undo, Clear Recents, F5, unreadable-file refusal, empty/zero-one-many states, ⌘⌫/⌘Z native on provider tabs | ✓ VERIFIED | `LibraryGrid.svelte:59`; StateMessage copy in `App.svelte:944-957`; `onKeydown` returns before preventDefault for ⌘⌫ off-library and ⌘Z without an Undo toast; store tests TestUndoRestoresPosition, TestUnreadableFileNeverOverwritten, TestUnknownVersionNeverOverwritten; UAT 12-library |
| 14 | 03-07: Details pane behavior, chips (≤8, ↑/↓/←/→, ↵ rerun, ⌘⌫ silent remove), Phase 2 polish, IN-04 Clips error reset, long-text / contrast / keyboard backstops | ✓ VERIFIED | `format.ts`, `RecentSearches.svelte`; UAT 13-details, 13-chips, 14-backstops (P8 widest 492.5 px; contrast ≥ 4.5 in light and dark; aria-activedescendant checks) |
| 15 | 03-08: whole suite green on one tree | ✓ VERIFIED | Re-run on HEAD 9824e93 (see spot-checks) |
| 16 | 03-08: signed dev build installs; log shows readiness; go.sum untouched | ✓ VERIFIED | `bin/install-dev.log` has both lines (2 matches), no key leak; go.sum clean before and after my runs. Note: this install predates the review fixes |
| 17 | 03-08: native, user-verified UAT on the installed app (13 steps) | ? UNCERTAIN | User attested ("approved"; "Already did these before"), but 03-08 found only download and reveal in the log, no library files, and no Quit. Routed to human verification on a fresh install of HEAD |
| 18 | 03-08: Clips MP4 paste user-verified, or recorded as not verifiable with a reason | ✓ VERIFIED | 03-08 SUMMARY step 8: skipped, key not Clips-approved, fixture-proven only (D-15, A1) |
| 19 | 03-08: post-session checks (staging clean, library files 0600, no key in log) | ? UNCERTAIN | Log leak grep PASS (I re-ran it); the staging dir was absent; the library files did not exist, so their 0600 mode was never observed natively (the unit test covers WriteJSONAtomic `Chmod(0o600)`) |

**Score:** 15/19 truths verified (2 present, behavior-unverified; 2 uncertain, routed to human)

### CR-01 closure (square-crop decode budget)

- **Fix present:** `actions.go:443` `frameOverhead = 4096 + 256`, added per image descriptor at `actions.go:594`. Budget check before `DecodeAll` at `actions.go:521`.
- **Tests:** `go test -run TestCopySquareDecodeBudget -v` passes all three:
  - Full frames: 335,631,360 is over the 268,435,456 budget.
  - Plain Text, empty Application and unknown-label prefixes are counted or refused.
  - Empty transparent frames: 268,474,880 is over the budget, refused with `crop_failed` and no platform calls.
- **Parser parity:** I compared `framePixels` with Go's `image/gif` reader:
  - Plain Text: 13 raw bytes, then sub-blocks.
  - Graphic Control: 6 bytes, no sub-blocks.
  - Application: size byte, that many bytes, then sub-blocks.
  - Unknown labels: refused.
  - Frame bounds: limited by the logical screen, which is ≤ 4096².
  - Go accepts at most one extra sub-block after the image data. Where Go errors, it stops allocating, so walking every sub-block is harmless.
  - A missing trailer makes `framePixels` fail before `DecodeAll` runs.
- **Bypass attempt** (go test `-overlay`, no repo files changed). I built the worst-case files that stay just under the budget:

| Variant | File | Frames | Live heap after DecodeAll | Peak heap during CopySquare | Time |
|---------|------|--------|---------------------------|-----------------------------|------|
| 0x0 frames, out-of-range transparent index (Go `make(Palette,256)` per frame) | 1.2 MB | 61,630 | +292 MB | 758 MB | 0.86 s |
| 0x0 transparent frames, 256-entry global table (reviewer's shape) | 1.2 MB | 61,630 | +292 MB | 638 MB | 0.87 s |
| 0x0 frames with 256-entry local tables, 15 MB file | 15.7 MB | 19,958 | +114 MB | 215 MB | 0.38 s |

The 3.7 GB live / 9 GB process reproduction is now refused. Memory is bounded at roughly the intended 256 MB for decode and under 1 GB peak, with each action finishing in under 1 s. I found no bypass. One residual note: the "about 256 MB" comment understates the peak during `EncodeAll` by about 3x. This is Info, not a blocker.

### Deferred Items

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | Chosen download folder (picker); Phase 3 uses an injected `DownloadsDir` (~/Downloads) | Phase 4 | SC3 "download folder" (SETT-04); 03-CONTEXT Deferred Ideas |
| 2 | Hide-original-filename toggle (naming implemented in `FileStem(g, true)`, `Download` passes `false`) | Phase 4 | SC3 "hide filename" (SETT-04) |

### Required Artifacts

`gsd-tools verify.artifacts` reports 30/30 across 03-01 to 03-08: all exist and are substantive. I checked them manually:

| Artifact | Status | Details |
|----------|--------|---------|
| `internal/services/actions.go` | ✓ VERIFIED | CopyText/PasteText/OpenPage/OpenSource/Download/ShowDownload/CopySquare/PasteSquare/ServiceShutdown, all bound (ActionService registered in main.go) |
| `internal/staging/staging.go` | ✓ VERIFIED | magic (GIF/ftyp), Save via os.Link, Clear/Sweep/Remove |
| `internal/imaging/square.go` | ✓ VERIFIED | stdlib only (`go list -deps`: no wails/cgo/png/jpeg) |
| `internal/library/store.go` + `internal/services/library.go` | ✓ VERIFIED | LibraryService registered in main.go; `actions.Recents = lib` |
| `frontend/src/components/{ActionMenu,DetailsPane,LibraryGrid,RecentSearches}.svelte` | ✓ VERIFIED | all imported and rendered in App.svelte (lines 920, 929, 961, 1067) |
| `build/darwin/Info.plist` | ✓ VERIFIED | NSDownloadsFolderUsageDescription in both plists |
| `scripts/uat/07,08,10,11,12*,13*,14` | ✓ VERIFIED | all UAT OK on my run |

### Key Link Verification

`gsd-tools verify.key-links`: 18/20 automated. I checked the 2 misses by hand:

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| DetailsPane.svelte | actions.go | openPage/openSource | WIRED | The pattern missed because of prop casing (`onopenpage`). `App.svelte:922-923` passes `openItem(detail, openPage/openSource)` → `api.ts:81-82` → `ActionService.OpenPage/OpenSource` |
| Taskfile install:dev | ~/Applications/Gifsmith.app | build/sign/install | WIRED (historical) | `build/darwin/Taskfile.yml:212` `install:dev`; log evidence in bin/install-dev.log |
| actions.go | library/store.go | `Recents.Touch` | WIRED | main.go `actions.Recents = lib` |
| main.go | staging | `stage.Clear()` before `app.Run()` | WIRED | |
| actions.go | Wails shutdown | `ServiceShutdown` | WIRED | `application.go:821` calls it for registered services |

### Data-Flow Trace (Level 4)

| Artifact | Data | Source | Real data | Status |
|----------|------|--------|-----------|--------|
| LibraryGrid / Favorites / Recents | `library` snapshot | `LibraryService.Load` → `Store.Snapshot` → favorites.json / recents.json | Yes (file-backed, Touch on every copy) | ✓ FLOWING |
| RecentSearches chips | `library.searches[active]` | `SaveSearch` → searches.json | Yes | ✓ FLOWING |
| DetailsPane | `detail` (selected Gif) | provider page result or library snapshot | Yes (mapper fields; no request) | ✓ FLOWING |
| Download toast | `DownloadResult{name, folder}` | `Staging.Save` | Yes | ✓ FLOWING |

### Behavioral Spot-Checks (all run by me on HEAD 9824e93)

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Go vet | `go vet . ./internal/...` | clean | ✓ PASS |
| Go tests with race | `go test -race -count=1 ./internal/...` | 13 packages ok | ✓ PASS |
| CR-01 regressions | `go test -run TestCopySquareDecodeBudget -v ./internal/services/` | 3/3 PASS | ✓ PASS |
| Lint / fmt | `golangci-lint run . ./internal/...`; `gofmt -l` | 0 issues; clean | ✓ PASS |
| Frontend unit | `pnpm --dir frontend run test` | 63 pass / 0 fail | ✓ PASS |
| svelte-check | `pnpm --dir frontend run check` | 0 errors / 0 warnings | ✓ PASS |
| Smoke | `bash scripts/smoke-browser-loop.sh` | SMOKE OK | ✓ PASS |
| Headless UAT ×18 | `bash scripts/smoke-browser-loop.sh --uat scripts/uat/<x>.js` (env defaulted from each script's Run-via line) | 18/18 UAT OK, including 12-library-unreadable and FAKE_CLIPS=denied scripts | ✓ PASS |
| go.sum untouched | `git diff --quiet -- go.sum` | clean | ✓ PASS |

Cleanup: ports 8089 and 9245 were free afterwards, no `gifsmith-server` was left running, and no new headless Chrome was started. The pre-existing `gifsmith-uat` browser was reused and left as it was. ~/Applications/Gifsmith.app was not touched.

### Probe Execution

No `scripts/*/tests/probe-*.sh` exist and the plans declare no probes. SKIPPED.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| ACTN-03 | 03-01, 03-02, 03-08 | Copy/paste link and Markdown | ✓ SATISFIED (native paste → human) | Truths 1, 6 |
| ACTN-04 | 03-01, 03-02, 03-08 | Copy page link, open page in browser | ✓ SATISFIED | Truths 1, 6 |
| ACTN-05 | 03-04, 03-08 | Download to a chosen folder (default Downloads) | ✓ SATISFIED for default folder; picker deferred to Phase 4 | Truth 11, Deferred 1 |
| ACTN-06 | 03-05, 03-08 | Square crop, not for Clips | ✓ SATISFIED | Truth 12 |
| ACTN-07 | 03-02, 03-08 | ⌘K action menu | ✓ SATISFIED | Truth 8 |
| ACTN-09 | 03-03, 03-08 | Clips as MP4, labelled, 720p | ✓ SATISFIED (native paste needs approved key) | Truth 9 |
| ACTN-10 | 03-03, 03-08 | Temp files deleted promptly; Klipy eagerly | ✓ SATISFIED (quit path → human) | Truths 3, 9, 10 |
| ACTN-11 | 03-04, 03-08 | Validated, no overwrite; optional hide filename | ✓ SATISFIED; toggle deferred to Phase 4 | Truth 11, Deferred 2 |
| LIBR-01 | 03-06, 03-08 | Add/remove favorites | ✓ SATISFIED | Truth 13 |
| LIBR-02 | 03-06, 03-08 | Recents on every copy/paste, capped | ✓ SATISFIED | Truths 4, 13; TestRecentsRecordedOnEverySuccessfulCopy |
| LIBR-03 | 03-06, 03-08 | Browse grouped by provider, remove from Recents | ✓ SATISFIED | Truth 13 |
| LIBR-04 | 03-06, 03-08 | Render from saved metadata, offline/rate-limited | ✓ SATISFIED (native persistence → human) | Truth 4; 12-library rate-limited case |
| LIBR-05 | 03-07, 03-08 | Details view | ✓ SATISFIED | Truth 14 |
| SRCH-08 | 03-07, 03-08 | Recent searches when box empty, rerun | ✓ SATISFIED | Truth 14 |

**Orphaned requirements:** none. REQUIREMENTS.md maps exactly these 14 IDs to Phase 3, and every one appears in a plan's `requirements` field.

### Prohibitions

| Plan | Prohibition | Tier | Result |
|------|-------------|------|--------|
| 03-01 | Unit tests never write the general pasteboard or open URLs | judgment | Holds: platform_darwin_test uses a private named pasteboard |
| 03-01/02/07 | Webview never composes clipboard text or URLs; menu rows go through the same dispatcher | judgment | Holds: `api.ts` passes `(Gif, kind)`; Go builds and validates everything |
| 03-02 | No popover fade/slide | judgment | Holds: no transition/animation in ActionMenu.svelte |
| 03-03 | ServiceShutdown takes no busy lock / onMain; tests inject After | judgment | Holds: `actions.go:383-385`; tests use `After` |
| 03-04 | Download never lists, creates or overwrites; no path from JS | judgment | Holds: CreateTemp + Link only; DownloadResult is name + folder |
| 03-05 | imaging is stdlib image/gif only, no re-quantize | judgment | Holds: `go list -deps` |
| 03-06 | library has no Wails / provider HTTP; metadata only; tests use TempDir | judgment | Holds: no net/http in deps; Entry = Gif metadata |
| 03-01/03-08 | Executor conduct (no synthetic input, no foreign kills, no Keychain/~/Downloads access) | judgment | Flagged: unverified-prohibition, human review recommended (non-authoritative: no evidence of violation) |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (84 phase-modified files) | - | TBD/FIXME/XXX/TODO/HACK/placeholder | none found | - |
| `.planning/phases/03-action-pipeline-library/03-REVIEW-DISPOSITION.md` | 8 | CR-01 recorded `open`, but the fix landed in 6928379/39d5c97 | ℹ️ Info | Stale record: update the disposition to `fixed` |
| `internal/services/actions.go` | 442 | The "about 256 MB" comment; measured peak during CopySquare is up to 758 MB | ℹ️ Info | Bounded; optionally tighten the comment or the budget |
| App.svelte / errors.ts / staging.go / store.go / provider.go | per 03-REVIEW | IN-01..IN-07 (library refresh ordering, ⌘⌫ in the menu filter, error copy for text/`invalid_item`, `.part` left in ~/Downloads on a crash, IN-05 comment, stale Undo toast, ClampID prefix merge) | ℹ️ Info | None blocks a must-have; all were carried forward by the reviewer as Info |

### Human Verification Required

Run all of these on a fresh `wails3 task darwin:install:dev` of the current HEAD. The installed build predates WR-01..03 and CR-01.

1. **Install + single instance.** Run install:dev, then double-click the app again. Expected: one tray icon, the panel shows, and the log has its readiness lines.
2. **Text actions in real apps.** In TextEdit try ⌘⇧C, Paste Link, ⌘⇧M, then ⌘O on GIPHY and on Klipy. Expected: the link has no query, Markdown is `![title](url)`, the GIPHY page opens, and Klipy does nothing.
3. **Paste Square GIF into Messages.** Expected: a square GIF that is still animated.
4. **Klipy 30 s delete.** Paste with ⌘↵ into Messages and run `ls` on the staging dir immediately and again after 35 s. Then paste into Slack and send about 40 s later. Expected: the file is gone after 35 s and the Slack send still works.
5. **Clips MP4** (approved key only; otherwise record it as not verifiable).
6. **Library persistence + mode.** Copy two GIFs, star one, and search. Expected: `ls -l` on Application Support shows the three JSON files at 0600, and after Quit and relaunch ⌘6, ⌘5 and the chip still show them.
7. **Quit cleanup.** After a copy, quit from the menu bar and also with ⌘Q, running `ls` on the staging dir each time. Expected: empty or absent.
8. **WKWebView keys.** Check ⌘⌫ and ⌘Z on a provider tab (native editing) and on Recents (remove and Undo), ⌘R with no toast (no reload), ⌘I then Esc, and ⌘S then ⌘R (Finder reveal).
9. **Hide-timer cancel.** Press ↵, click away within about 0.4 s, then press the hotkey. Expected: the panel stays open.
10. **Process prohibitions.** Confirm no executor input, kills or Keychain access happened (judgment tier).

### Gaps Summary

No blocking gaps. All 14 requirement IDs have code, wiring and automated evidence. Code review CR-01 is closed, and my adversarial test found no further decode-budget bypass. Status is `human_needed` for three reasons:
- Two truths are present and wired, but no test drives them: the panel:show hide-timer cancel and staging cleanup on a real Quit.
- The 03-08 native UAT was user-attested on a build older than the review fixes, without log or file evidence for library persistence, ⌘Q cleanup, Klipy grace, Clips paste, WKWebView key handling or paste targets.
- One judgment-tier process prohibition is flagged for confirmation.

The download folder picker and the hide-filename toggle are deferred to Phase 4 (SETT-04).

---

_Verified: 2026-10-18T18:05:00Z_
_Verifier: Claude (gsd-verifier)_
