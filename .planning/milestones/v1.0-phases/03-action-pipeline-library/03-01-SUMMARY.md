---
phase: 03-action-pipeline-library
plan: 01
subsystem: actions
tags: [go, cgo, nspasteboard, nsworkspace, wails-v3, svelte5, headless-uat]

requires:
  - phase: 01-native-spike
    provides: platform seam (CopyFile, paste sequence), staging allowlist, Fake platform
  - phase: 02-providers-search-ui-first-run
    provides: GIPHY/Klipy/Finer clients, fakeapi fixture server, routeKey, footer toasts, smoke + UAT loop
provides:
  - Platform.CopyText (cgo text pasteboard) and Platform.OpenURL (onMain NSWorkspace), stub and Fake
  - services.ActionError kinds + MarshalError, composed after provider.MarshalError in main.go
  - ActionService.CopyText (link|markdown|page), PasteText (link|markdown), OpenPage, OpenSource, shared paste()
  - provider.Gif optional metadata (width, height, size, tags, pageUrl, sourceUrl, mediumUrl, gifUrl)
  - staging.PageAllow (the one page-URL rule) and Store.Allows
  - httpx.LooseInt for display-only vendor numbers
  - server (UAT) build runs platform.Fake; fakeapi broken*/html* media and brokenmedia/htmlmedia Finer words
  - frontend routes copyLink (Cmd+Shift+C), copyMarkdown (Cmd+Shift+M), open (Cmd+O); actionErrorText F1/F3/F8
  - scripts/uat/07-text-actions.js
affects: [03-02, 03-03, 03-04, 03-05, 03-06, 03-07, 03-08]

actuals:
  tokens: 21764
  tasks: 3
  commits: 6
plan_head_before: 3d9ec20871bf244048bbf6e9ee2ddc2cc4dc9cba
plan_head_after: 4a3002a7a140cd05ed2d446da5e1118cf742eb65

tech-stack:
  added: []
  patterns:
    - "Go builds every clipboard string and validates every URL; JS only names a kind"
    - "Typed ActionError {kind} with fixed text; frontend copy keyed by err.cause.kind"
    - "Display-only vendor fields decode tolerantly (LooseInt, any, RawMessage); action-critical URLs stay strict"
    - "Per-item URLs open only through Platform.OpenURL; UAT build wires platform.Fake"

key-files:
  created:
    - internal/services/errors.go
    - scripts/uat/07-text-actions.js
  modified:
    - internal/platform/platform.go
    - internal/platform/platform_darwin.go
    - internal/platform/platform_darwin.m
    - internal/platform/platform_darwin.h
    - internal/platform/platform_stub.go
    - internal/platform/fake.go
    - internal/services/actions.go
    - internal/provider/provider.go
    - internal/provider/giphy/giphy.go
    - internal/provider/klipy/klipy.go
    - internal/provider/httpx/httpx.go
    - internal/provider/fakeapi/fakeapi.go
    - internal/staging/staging.go
    - main.go
    - wire_desktop.go
    - wire_server.go
    - frontend/src/App.svelte
    - frontend/src/lib/api.ts
    - frontend/src/lib/errors.ts
    - frontend/src/lib/grid.ts
    - scripts/smoke-browser-loop.sh
    - scripts/uat/01-tracer.js

key-decisions:
  - "ActionError has no Unwrap: staging and transport errors are logged, then mapped to a kind; tests match on Kind"
  - "staging.PageAllow (GIPHY/Clips plus the GIPHY VendorAllow host rule) is the single page-URL rule, used by the GIPHY mapper and ActionService"
  - "Display-only vendor numbers decode through httpx.LooseInt (malformed -> 0, never fails a page); GIPHY url/source decode as any"
  - "The -tags server UAT build wires &platform.Fake{} (Trusted false); desktop wires platform.New()"
  - "Klipy tags over 100 runes are truncated, not dropped; at most 50 kept"

patterns-established:
  - "act(call, pasteIt): one App.svelte path for every copy/paste of the selected item, scheduleHide() cancelled on panel:show"
  - "paste(put, okMsg): one Go paste sequence for file and text pastes"

requirements-completed: [ACTN-03, ACTN-04]

coverage:
  - id: D1
    description: "Cmd+Shift+C copies the GIF link through Go; footer shows 'Copied link' with the check icon; plain Cmd+C stays native"
    requirement: ACTN-03
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopyTextLink"
        status: pass
      - kind: unit
        ref: "frontend/src/lib/grid.test.ts#routeKey maps Cmd+Shift+C to copyLink"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/07-text-actions.js"
        status: pass
    human_judgment: false
  - id: D2
    description: "Markdown copy escapes the alt text and falls back to 'GIF'; Clips links/Markdown use the GIF rendition; Klipy links pass unaltered"
    requirement: ACTN-03
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopyTextMarkdown, TestCopyTextClipLink, TestCopyTextLink"
        status: pass
    human_judgment: false
  - id: D3
    description: "Copy Page Link and Open in Browser only for GIPHY/Clips https giphy.com pages; Klipy/Finer refuse with no_page; Cmd+O is a no-op without a page URL"
    requirement: ACTN-04
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopyTextPage, TestOpenPage, TestOpenSource"
        status: pass
      - kind: unit
        ref: "internal/staging/staging_test.go#TestPageAllow"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/07-text-actions.js"
        status: pass
    human_judgment: false
  - id: D4
    description: "Paste Link / Paste Markdown share the Phase 1 paste sequence; without Accessibility they copy only with the capitalised 'Enable Automatic Paste…' message"
    requirement: ACTN-03
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestPasteTextTrusted, TestPasteTextWithoutAccessibility, TestPasteWithoutAccessibility"
        status: pass
    human_judgment: false
  - id: D5
    description: "Copy failures render F1 (500 media) and F3 (HTML media) from typed kinds"
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopyFailingMediaKinds; frontend/src/lib/errors.test.ts"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/01-tracer.js (brokenmedia E7), scripts/uat/07-text-actions.js (htmlmedia F3)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Provider results carry optional metadata only from vendor fields; malformed display-only fields never fail a page"
    verification:
      - kind: unit
        ref: "internal/provider/giphy/giphy_test.go#TestPageURL, TestSourceURL, TestDisplayFieldsTolerant, TestMediumURL, TestClipsGifURL; internal/provider/klipy/klipy_test.go#TestTagsTolerant, TestDimensionsFollowDownload"
        status: pass
    human_judgment: false
  - id: D7
    description: "Native: Cmd+Shift+C / Cmd+Shift+M / Cmd+O in the real WKWebView panel (key casing under Cmd+Shift, A5), real pasteboard text and real browser open"
    verification: []
    human_judgment: true
    rationale: "Synthetic desktop input is forbidden on the user's machine; headless Chromium cannot prove WKWebView key equivalents or NSWorkspace opening. Planned for the plan 08 checkpoint."

duration: 14min
completed: 2026-10-18
status: complete
---

# Phase 3 Plan 01: Text Actions (Link, Markdown, Page, Open) Summary

**Link, Markdown and page-link copy, link/Markdown paste and Open in Browser built and validated in Go (cgo NSPasteboard text, onMain NSWorkspace open), driven by Cmd+Shift+C / Cmd+Shift+M / Cmd+O, with typed ActionError kinds and a server UAT build that now runs platform.Fake.**

## Performance

- **Duration:** about 14 min
- **Started:** 2026-10-18T00:10:00Z
- **Completed:** 2026-10-18T00:23:38Z
- **Tasks:** 3 (each RED then GREEN)
- **Files modified:** 34

## Accomplishments
- Tracer: Cmd+Shift+C goes from routeKey through `ActionService.CopyText` and `Platform.CopyText` back to the footer toast. Headless UAT proves it on the `-tags server` build, which now uses `platform.Fake`, so it never touches the real pasteboard or opens a browser.
- `ActionError{kind}` (busy, download_failed, not_media, too_large, not_allowed, no_link, no_page, open_failed, copy_failed) with fixed, URL-free text. main.go's MarshalError tries provider errors first, then action errors. The frontend renders F1/F3/F8 from `err.cause.kind` and never shows the caught error's text.
- Go builds all clipboard text. GIPHY/Clips links lose their query string. Klipy and Finer links pass through byte for byte. A Clip's link is its GIF rendition (D-18). Markdown escapes `[`, `]` and backslash and collapses whitespace. Page actions go only through `staging.PageAllow`.
- `Paste` and `PasteText` share one `paste(put, okMsg)` sequence. The T5a message now names `Enable Automatic Paste…`.
- `provider.Gif` gains eight optional metadata fields, filled only from vendor fields. GIPHY: original dimensions and size, fixed_height medium, page and source URLs. Clips: GifURL. Klipy: dimensions and size from the download size key, md medium, and tolerant, capped tags.
- 02 IN-07: `panel:show` cancels a pending hide-after-copy.

## Task Commits

1. **Task 1: Cmd+Shift+C copies the GIF link (tracer)**: `5bc5f38` (test, RED), `5c7f3b4` (feat, GREEN)
2. **Task 2: Provider metadata fields**: `ae0ec85` (test, RED), `2f56deb` (feat, GREEN)
3. **Task 3: Markdown, page link, paste variants and Open in Browser**: `68db951` (test, RED), `4a3002a` (feat, GREEN)

## TDD Gate Compliance

Each task has a `test(...)` RED commit with tests compiling against skeletons and failing (`go test` FAIL lines or `# fail 2` in node --test), followed by a `feat(...)` GREEN commit where they pass. No refactor commits were needed.

## Verification (latest run)

- `gofmt -l` empty; `go vet . ./internal/...`; `go test -race ./internal/... -count=1` all ok; `golangci-lint run . ./internal/...`: 0 issues
- `CGO_ENABLED=0 go build -tags server .` ok; `pnpm --dir frontend run test`: 27 pass, 0 fail; `svelte-check`: 0 errors
- `scripts/smoke-browser-loop.sh`: SMOKE OK (Copy 200 `"copied":true` + staged .gif, CopyText 200 `Copied link`, disallowed URL 422 with `cause.kind not_allowed`, OpenPage 200 on Fake)
- `--uat` UAT OK: 01-tracer, 02-keyboard, 05-setup, 07-text-actions; regression runs also OK: 02-grid, 03-states, 05-tabs, 06-clips, 06-clips-denied and 06-backstops (both with `FAKE_CLIPS=denied`)

## Decisions Made
- `ActionError` has no `Unwrap`. Errors are logged with provider and kind only, then mapped. Three existing tests (`TestCopyErrorKeepsPanelOpen`, `TestCopyDisallowedURL`, `TestPasteCopyError`) now assert the kind instead of `errors.Is` on the staging or Fake error.
- `staging.PageAllow` is the single page rule. The GIPHY mapper imports `staging` to use it, which creates no import cycle.
- The busy check moved to `actionError(err)?.kind === 'busy'` in Task 1, together with the `run()` rewrite. The plan had scheduled it for Task 3.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `provider.Gif` is no longer comparable**
- **Found during:** Task 2
- **Issue:** `Tags []string` makes `Gif` non-comparable, which broke `==` and `slices.Equal` in the giphy, klipy and finergifs tests.
- **Fix:** Those comparisons now use `reflect.DeepEqual`. This touches `internal/provider/finergifs/finergifs_test.go`, which is not in the plan's file list. The finergifs test now also asserts that Finer items carry no metadata.
- **Committed in:** `ae0ec85`

**2. [Rule 2 - Correctness] Tolerant decoding for display-only numbers and URLs**
- **Found during:** Task 2
- **Issue:** GIPHY's `flexInt` returns an error on values like `"wide"`, and Klipy's int fields fail on strings or objects. Either would fail a whole page, which contradicts D-11.
- **Fix:** Added `httpx.LooseInt` (malformed values become 0, never an error) in `internal/provider/httpx/httpx.go`, which is not in the plan's file list, and used it for GIPHY and Klipy width, height and size. GIPHY `url` and `source` decode as `any`. Klipy `tags` decodes from `json.RawMessage`. `flexInt` stays strict for `total_count`.
- **Committed in:** `2f56deb`

**3. [Rule 2 - Correctness] NUL bytes rejected before the text pasteboard write**
- **Found during:** Task 1
- **Issue:** `C.CString` would silently cut text at an embedded NUL, for example one inside a vendor title in Markdown.
- **Fix:** `copyTextToPasteboard` rejects invalid UTF-8 and NUL before touching the pasteboard. Covered in `TestCopyTextWritesString`.
- **Committed in:** `5c7f3b4`

**4. [Rule 2 - Verification] Smoke also calls `ActionService.OpenPage` on a GIPHY item**
- Proves the server build sends opens to the Fake (HTTP 200, nothing opens). Committed in `4a3002a`.

---

**Total deviations:** 4 auto-fixed (1 blocking, 3 correctness/verification). **Impact:** no scope creep. All four are small and keep the plan's own truths true.

## Issues Encountered
- A8 (binding optional form): the generator emits quoted keys, e.g. `"pageUrl"?: string;` and `"tags"?: string[] | null;`. The properties are optional as intended, but the plan's literal grep `pageUrl?` does not match because of the closing quote. `grep -n '"pageUrl"?' frontend/bindings/gifsmith/internal/provider/models.ts` matches line 40.
- `06-backstops.js` fails without `FAKE_CLIPS=denied`. Its header documents that requirement, and it passes when run as documented.

## Known Stubs
None. Copy GIF for Clips still shows the Phase 2 "arrives in the next update" message and `ErrClipsCopy`. Removing them belongs to the Clips MP4 plan (D-15). The MP4 variants of F1/F3 are not implemented for the same reason.

## Threat Flags
None. The new surfaces (OpenURL, text pasteboard, vendor page and source URLs) are covered by T-03-01 to T-03-05 and mitigated as planned.

## User Setup Required
None.

## Next Phase Readiness
- Later plans can use `Platform.CopyText` and `Platform.OpenURL`, `ActionError` kinds and `MarshalError`, `act()`/`scheduleHide()` and `paste(put, okMsg)`, the Gif metadata fields, and the server build on `platform.Fake`.
- Native confirmation (D7) is pending at the plan 08 checkpoint: Cmd+Shift key casing in WKWebView (A5), real pasteboard text, and a real browser open.

## Self-Check: PASSED
