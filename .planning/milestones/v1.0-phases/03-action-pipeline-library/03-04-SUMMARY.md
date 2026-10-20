---
phase: 03-action-pipeline-library
plan: 04
subsystem: actions
tags: [go, staging, downloads, os-link, nsworkspace, wails-v3, svelte5, headless-uat]

requires:
  - phase: 03-action-pipeline-library
    plan: 01
    provides: ActionError kinds, actionErrorText, platform.Fake server build, onMain native pattern
  - phase: 03-action-pipeline-library
    plan: 02
    provides: lib/actions.ts ROWS and the runAction dispatcher
  - phase: 03-action-pipeline-library
    plan: 03
    provides: kind-aware Fetch (GIF magic / MP4 ftyp), busy lock, gen counter
provides:
  - staging.Store.Save (validated, temp file + os.Link claim, never overwrites, never creates the folder)
  - staging.FileStem (title slug + sanitized id, hide name) and staging.ErrFolder
  - one shared request helper (media) and spool used by both Fetch and Save
  - services.DownloadResult, ActionService.DownloadsDir, Download, ShowDownload, lastDownload
  - ActionError kinds folder_unwritable and nothing_to_show
  - Platform.RevealFile (darwin gsReveal via onMain, stub ErrUnsupported, Fake RevealErr)
  - deps.DownloadsDir (desktop ~/Downloads, server bin/uat-downloads)
  - NSDownloadsFolderUsageDescription in Info.plist and Info.dev.plist
  - KeyAction download (⌘S) and reveal (⌘R); Download GIF/MP4 menu row; F2-F4 copy
  - Footer toast action link + hint; App toast model (message, tone, action, expiry)
  - scripts/uat/10-download.js and smoke Download checks
affects: [03-05, 03-06, 03-07, 03-08, phase-4-settings]

actuals:
  tokens: 16477
  tasks: 2
  commits: 4
plan_head_before: 93aeeeb7e2a6959e1b8cfdd9bf049b62bacd8f84
plan_head_after: 56d569c4340dfcf35d10bb934f38264c76932331

tech-stack:
  added: []
  patterns:
    - "Remote media reaches disk only through staging.media: allowlist on URL and redirects, 200, magic bytes, then a capped reader"
    - "User-visible files are claimed with os.Link from a hidden temp file: fs.ErrExist means try the next ' (n)' name, so nothing is ever truncated or replaced"
    - "Toasts carry an optional {label, hint, run} action and an expiry; setStatus replaces message, action and timer together"

key-files:
  created:
    - scripts/uat/10-download.js
  modified:
    - internal/staging/staging.go
    - internal/staging/staging_test.go
    - internal/services/actions.go
    - internal/services/actions_test.go
    - internal/services/errors.go
    - internal/platform/platform.go
    - internal/platform/platform_darwin.go
    - internal/platform/platform_darwin.m
    - internal/platform/platform_darwin.h
    - internal/platform/platform_stub.go
    - internal/platform/fake.go
    - main.go
    - wire_desktop.go
    - wire_server.go
    - build/darwin/Info.plist
    - build/darwin/Info.dev.plist
    - frontend/bindings/gifsmith/internal/services/actionservice.ts
    - frontend/bindings/gifsmith/internal/services/models.ts
    - frontend/bindings/gifsmith/internal/services/index.ts
    - frontend/src/lib/api.ts
    - frontend/src/lib/grid.ts
    - frontend/src/lib/grid.test.ts
    - frontend/src/lib/actions.ts
    - frontend/src/lib/actions.test.ts
    - frontend/src/lib/errors.ts
    - frontend/src/lib/errors.test.ts
    - frontend/src/components/Footer.svelte
    - frontend/src/App.svelte
    - scripts/smoke-browser-loop.sh
    - scripts/uat/08-action-menu.js

key-decisions:
  - "Fetch and Save share media() (request, allowlist, status, magic) and spool() (copy with the 50 MB cap); Fetch behaviour is unchanged"
  - "Save maps every temp-file, Sync, Chmod, Close or non-ErrExist Link failure to ErrFolder; a LinkError (which holds both paths) is never wrapped"
  - "actionErrorText gained a 4th optional folder parameter (default 'Downloads'); App passes the folder name from the last successful download, so F4 follows the folder Go reports"
  - "On a download, not_allowed, too_large and untyped errors show F2 (UI-SPEC has no separate row); not_media shows F3 for both copy and download"
  - "Reveal failure has no UI-SPEC copy: Go logs it and the toast just clears (the frontend swallows the rejection)"
  - "A toast that has an action survives panel:show within its 5 s (Pitfall 3: a TCC prompt can hide the panel); its own timer clears it"
  - "⌘R runs the toast action only when the visible toast's hint is ⌘R; any other time it is captured (preventDefault) and ignored"

patterns-established:
  - "ActionService.DownloadsDir is a func: Phase 4's folder picker replaces the value with no change to Save, the toast or F4"
  - "Show-type actions take no argument from JS: Go remembers what it saved (lastDownload, guarded by busy) and reveals only that"

requirements-completed: [ACTN-05, ACTN-11]

coverage:
  - id: D1
    description: "⌘S or the Download GIF / Download MP4 row saves <title-slug>-<id>.gif (.mp4 for Clips) into the download folder; the panel stays open with 'Downloaded to {folder}' for 5 s"
    requirement: ACTN-05
    verification:
      - kind: unit
        ref: "internal/staging/staging_test.go#TestSaveGIF, TestSaveClipMP4; internal/services/actions_test.go#TestDownloadSavesWithoutStagingOrClipboard"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/10-download.js"
        status: pass
    human_judgment: false
  - id: D2
    description: "Never overwrites: a second download becomes '<stem> (2).gif', then ' (3)'; a pre-existing user file is untouched; no .part file is ever visible under a final name; files are 0644"
    requirement: ACTN-11
    verification:
      - kind: unit
        ref: "internal/staging/staging_test.go#TestSaveCollisionNeverOverwrites, TestSavePreexistingFileUntouched, TestSaveGIF"
        status: pass
      - kind: integration
        ref: "bash scripts/smoke-browser-loop.sh (Download twice -> name and name (2), both start GIF8)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Non-GIF/MP4 responses are rejected before any file exists (F3); HTTP/network failure shows F2; a missing or unwritable folder shows F4 and is never created"
    requirement: ACTN-11
    verification:
      - kind: unit
        ref: "internal/staging/staging_test.go#TestSaveNotMediaLeavesNothing, TestSaveFolderErrors; internal/services/actions_test.go#TestDownloadErrorKinds; frontend/src/lib/errors.test.ts#download errors show F2, F3 and F4"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/10-download.js (htmlmedia + ⌘S -> F3)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Names hold only the title slug plus [A-Za-z0-9_-] from the id (≤64 bytes); traversal ids stay inside the folder; the hide option gives gifsmith-<id>; any injected folder behaves the same"
    requirement: ACTN-11
    verification:
      - kind: unit
        ref: "internal/staging/staging_test.go#TestFileStem, TestSaveTraversalID, TestSaveAnyFolderHideName; internal/services/actions_test.go#TestDownloadFolderIsInjected"
        status: pass
    human_judgment: false
  - id: D5
    description: "Show in Finder (toast link or ⌘R while the toast is up) reveals the last saved file through Go; ⌘R at any other time is ignored and never reloads the webview"
    requirement: ACTN-05
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestShowDownloadRevealsLastDownload, TestShowDownloadNothingToShow, TestShowDownloadRevealError, TestShowDownloadRefusesWhileBusy"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/10-download.js (click and ⌘R paths, keystroke clear, 5 s expiry, window.__marker reload check; server log shows exactly 2 'show download: revealed')"
        status: pass
    human_judgment: false
  - id: D6
    description: "Native: first write to the real ~/Downloads from the signed dev build (TCC prompt text, grant/deny -> F4), Finder actually opens with the file selected, ⌘R not swallowed by WKWebView"
    verification: []
    human_judgment: true
    rationale: "Synthetic desktop input and real Finder/Downloads access are forbidden in automated runs; assumption A3 and Pitfall 10 belong to the plan 03-08 native checkpoint."

duration: 15min
completed: 2026-10-18
status: complete
---

# Phase 3 Plan 04: Download to a Folder and Show in Finder Summary

**⌘S (or the Download GIF / Download MP4 row) now saves the selected item into Downloads as `<title-slug>-<id>.gif` or `.mp4`. The bytes are validated before any file exists, the file is written to a hidden temp file and claimed with `os.Link` under the first free name (` (2)`, ` (3)`…), so nothing is ever overwritten. The 5 s toast offers Show in Finder, by click or ⌘R, which reveals the file Go itself saved.**

## Performance

- **Duration:** about 15 min
- **Started:** 2026-10-18T00:52:56Z
- **Completed:** 2026-10-18T01:02:18Z
- **Tasks:** 2 (each RED, then GREEN)
- **Files modified:** 31

## Accomplishments
- staging: `Fetch` now uses the same `media()` request helper as `Save`. `media()` applies the allowlist to the URL and every redirect, needs a 200, and checks the GIF or ftyp magic. `spool()` applies the 50 MB cap. `Save` checks the magic before calling `os.CreateTemp(dir, ".gifsmith-*.part")`, then calls Sync and `Chmod(0o644)`. It claims the name with `os.Link` for n = 1..999 and asserts `filepath.Dir(candidate) == filepath.Clean(dir)`. The temp file is always removed.
- `FileStem`: `slug(title)-<id chars>`, or `gifsmith-<id>` when hide is set. A traversal id such as `../../etc` becomes `etc`. An id with no safe characters leaves only the slug.
- services: `Download` holds the busy lock and refuses an empty folder with `folder_unwritable`. It returns `{name, folder}` (base names only) and records `lastDownload`. It never uses staging, the pasteboard, `gen` or Recents. `ShowDownload(ctx)` reveals only `lastDownload`.
- platform: `RevealFile` uses `gsReveal`, which runs `activateFileViewerSelectingURLs:` inside `onMain`. The stub returns `ErrUnsupported`, and the Fake records `RevealFile:<path>`.
- Both plists carry `NSDownloadsFolderUsageDescription`.
- Frontend:
  - ⌘S and ⌘R routing.
  - A Download row after Open in Browser (after Paste Markdown when the item has no page).
  - `Downloading…` with the busy overlay.
  - F2, F3 and F4 copy.
  - Footer action link and hint: the message ellipsizes, and the link and hint never shrink.
  - An App toast model with an action and an expiry.
- Server build: downloads go to `bin/uat-downloads`, which the smoke script recreates on every run. The user's `~/Downloads` is never touched.

## Task Commits

1. **Task 1 (tracer): ⌘S saves into Downloads**: `b3a0b20` (test, RED), `4ef9758` (feat, GREEN)
2. **Task 2: Show in Finder, toast action, usage description**: `babe15c` (test, RED), `56d569c` (feat, GREEN)

## TDD Gate Compliance

- **Task 1 RED (TS):** three targets ran under `node --test --test-reporter=tap` and failed (37 pass, 3 fail): "download errors show F2, F3 and F4", "Download comes right after Open in Browser, or after Paste Markdown without a page" and "routeKey maps Cmd+S to download and Cmd+R to reveal (no Shift)". `gsd-tools check tdd-red-evidence` returned `RED_EVIDENCE_OK` for each (records in `bin/red/0304-t1-{errors,actions,grid}.json`, gitignored). Semantic assessment: each failed on its planned assertion. The errors test got the copy wording (F1). The actions test found no download row. The grid test got `none`. There were no load faults.
- **Task 1 RED (Go):** the tests compiled against skeletons: `ErrFolder`; `FileStem` returning ""; `Save` returning ErrDownload; `Download` returning download_failed. 13 targets failed on their own assertions: 9 staging (Save/FileStem) and 4 services (Download). Go `-v` output is not a classifier format. Output is in `bin/red/0304-t1-go.txt`.
- **Task 2 RED (Go):** `RevealFile` was added to the interface, the Fake and the stub, with a darwin skeleton. A `ShowDownload` skeleton returned nil. TestShowDownloadRevealsLastDownload, TestShowDownloadNothingToShow, TestShowDownloadRevealError and TestShowDownloadRefusesWhileBusy all failed on assertions ("calls = []", "err = <nil>"). Output is in `bin/red/0304-t2-go.txt`. Task 2 has no TS unit: its frontend behaviour is proven by 10-download.js.
- Each GREEN commit follows its RED commit. No refactor commits.

## Verification (latest run)

- `gofmt -l` empty. `go vet . ./internal/...` and `go test -race ./internal/... -count=1` are all ok. `golangci-lint run . ./internal/...` reports 0 issues. `go build .` (desktop, cgo) and `CGO_ENABLED=0 go build -tags server .` both build.
- `go test -race ./internal/staging -run 'Save|FileStem' -v`: PASS for TestFileStem, TestSaveGIF, TestSaveCollisionNeverOverwrites, TestSavePreexistingFileUntouched, TestSaveClipMP4, TestSaveNotMediaLeavesNothing, TestSaveFolderErrors, TestSaveTraversalID and TestSaveAnyFolderHideName.
- `plutil -lint` reports OK for both plists.
- `pnpm --dir frontend run test`: 40 pass, 0 fail. `svelte-check`: 0 errors, 0 warnings.
- `scripts/smoke-browser-loop.sh`: SMOKE OK.
  - Download returned `office-moment-1-office-1.gif`, then `office-moment-1-office-1 (2).gif`. Both files start with GIF8, and no response body contains `uat-downloads/`.
  - The html media download returned 422 `not_media` and created no new file.
- `--uat` reports UAT OK for 10-download, 07-text-actions, 08-action-menu, 01-tracer, 02-keyboard, 02-grid, 03-states, 05-tabs, 05-setup and 06-clips.
- The server log after 10-download shows exactly 2 `show download: revealed` lines (the click and ⌘R paths) and none for the expired ⌘R.
- Acceptance greps: `os.Link`, `Chmod(0o644)`, `DownloadsDir` (wire_desktop, wire_server, main), `gsReveal` (inside `onMain`), the `ShowDownload(ctx context.Context) error` signature, `actionLabel`, `Staging.Save(`, `RevealFile(` and `showDownload(`. All match.

## Decisions Made
See `key-decisions` in the frontmatter. Discretion choices:
- **Reveal failure:** the toast simply clears, and Go logs `show download: <err>`.
- **Errors without a UI-SPEC row:** on a download, `too_large` and `not_allowed` show F2.
- **Download log lines:** Go logs `download <provider>: saved` on success and `show download: revealed` on a reveal. Neither includes a path.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Existing action-menu tests hard-coded the row list**
- **Found during:** Task 1 GREEN
- **Issue:** `frontend/src/lib/actions.test.ts` (ids, labels, hints, groups, Klipy length, bare Clip rows) and `scripts/uat/08-action-menu.js` (Finer rows, separators, GIPHY row count, clamp row) asserted the menu without Download. The second file is not in the plan's file list.
- **Fix:** both now expect the Download row: 7 Finer rows with 2 separators, 9 GIPHY rows, and the clamp lands on Download GIF.
- **Committed in:** `4ef9758`

**2. [Rule 2 - Correctness] F4 renders the folder Go reported**
- **Issue:** the interface listed `actionErrorText(kind, g, action)`, but step 7 and the assumption_delta decision require F4 to render the folder's display name.
- **Fix:** added an optional 4th `folder` parameter (default "Downloads"). App passes the name from the last successful download.
- **Committed in:** `4ef9758`

**3. [Rule 2 - Observability] `show download: revealed` log on success**
- **Issue:** the UAT could not prove that the click and ⌘R paths reached Go (the Fake's call list is not visible from the browser).
- **Fix:** ShowDownload now logs a path-free success line, consistent with `open: …` and `copy: …`.
- **Committed in:** `56d569c`

**4. Smoke html-media download:** the smoke script rewrites `ITEM`'s download URL to `html-1.gif` instead of running a second "htmlmedia" search. The fixture route and the behaviour checked are the same.

**5. Generated `frontend/bindings/.../index.ts`** gained the `DownloadResult` type export through regeneration. It is not in the plan's file list.

---

**Total deviations:** 3 auto-fixed (1 blocking, 2 correctness/observability) and 2 notes. **Impact:** no scope creep.

## Issues Encountered
- `os.Link` needs a filesystem that supports hard links. A download folder on FAT/exFAT or some network volumes would get F4 on every save. `~/Downloads` on APFS is fine. Phase 4's folder picker should either accept this or fall back to an `O_EXCL` create. This is not handled here.
- The hidden `.gifsmith-*.part` temp file exists in the destination folder while a download streams. It is removed on every path, including errors, but a crash mid-download would leave it behind.

## Known Stubs
- `internal/services/actions.go` `Download`: `hide` is hard-coded `false`. This is intentional. The Settings toggle is Phase 4 SETT-04 (CONTEXT Deferred Ideas), and `FileStem` already implements both names. It does not block this plan's goal.

## Threat Flags
None. Path traversal, overwrite, malicious content, path disclosure, reveal and DoS (T-03-12 to T-03-17) are mitigated as planned. Ids are whitelisted and the Dir assert runs before linking. `os.Link` never replaces an existing file. The magic check runs before any file exists, with the allowlist on every hop and the 50 MB cap. JS receives base names only, and the error texts are fixed. ShowDownload takes no arguments. The busy lock and the 45 s timeout are shared with the other actions.

## User Setup Required
None. On the first real download, the signed dev build may raise the Downloads TCC prompt with Gifsmith's text. This is to be confirmed at the plan 03-08 checkpoint.

## Next Phase Readiness
- Later plans can use the toast model through `setStatus(text, tone, {label, hint, run}, ms)` for the Undo toasts (⌘Z hint, 5 s). Undo toasts must not clear on keystrokes, so `clearTransient` will need a guard for them.
- Square crop (03-05) can reuse `staging.media()` for its download.
- Native confirmation (D6) waits for the plan 03-08 checkpoint: TCC prompt, Finder reveal, and ⌘R inside WKWebView.

---
*Phase: 03-action-pipeline-library*
*Completed: 2026-10-18*

## Self-Check: PASSED
