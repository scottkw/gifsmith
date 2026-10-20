---
phase: 02-providers-search-ui-first-run
plan: 01
subsystem: providers, copy pipeline, panel UI
tags: [go, wails-v3, svelte5, httptest, finergifs, staging, dev-browser]

requires:
  - phase: 01-native-spike-signed-dev-loop
    provides: platform.CopyFile/Fake, ActionService busy try-lock + Phase 1 paste rules, shell panel/Hide/Previous, -tags server browser loop
provides:
  - provider contract (ID, Gif, Page, Provider, Kind, URL-free Error, MarshalError, PageSize) wired as Options.MarshalError
  - httpx.GetJSON choke point (UA, classification, Retry-After on 429 only, 8 MB cap, no wrapping) + Remap for per-vendor bad-key statuses
  - finergifs client (results[].fields mapper, hits.found paging, fileid regex, no trending)
  - fakeapi fixture server (GET /finer/search, GET /media/<name>.gif) shared by go tests and the server build
  - staging.Store (allowlist, 45 s, 200 + GIF magic, 50 MB, hashed dir + slug, .part+rename, Sweep) and VendorAllow
  - GifService.Search/Trending, PanelService.Hide, ActionService.Copy/Paste(ctx, gif) with ErrClipsCopy
  - wire_desktop.go / wire_server.go deps split; every NewService stays in untagged main.go
  - app.css token set, lib/api.ts facade, App.svelte search -> 4-column grid -> copy flow
  - scripts/smoke-browser-loop.sh default/--serve/--uat modes and scripts/uat/01-tracer.js
affects: [02-02, 02-03, 02-04, 02-05, 02-06]

actuals:
  tokens: 20900   # chars/4 over the final content of the 30 changed files (83,731 chars)
  tasks: 3
  commits: 7
plan_head_before: 3723c39ac5c99912ba420f9d40218bb0ec80ebdf
plan_head_after: aa90d54e466208d0cf6daf1c8a6845a6f6f13dcf

tech-stack:
  added: []
  patterns:
    - "Typed provider.Error is the only error a provider returns; its Error() and JSON hold provider/kind/status/retryAfterSec only"
    - "Transport errors are classified into fresh values, never wrapped (httpx and staging)"
    - "Fixture server speaks the vendor wire format; tests and the -tags server build use the same handler"
    - "Downloads use their own http.Client bounded by Fetch's 45 s context, not the 15 s API client"
    - "Frontend hides the panel after a copy confirmation (PanelService.Hide); Go never hides on Copy"

key-files:
  created:
    - internal/provider/provider.go
    - internal/provider/provider_test.go
    - internal/provider/httpx/httpx.go
    - internal/provider/httpx/httpx_test.go
    - internal/provider/finergifs/finergifs.go
    - internal/provider/finergifs/finergifs_test.go
    - internal/provider/finergifs/testdata/search.json
    - internal/provider/fakeapi/fakeapi.go
    - internal/staging/staging.go
    - internal/staging/staging_test.go
    - internal/services/gifs.go
    - internal/services/panel.go
    - wire_desktop.go
    - wire_server.go
    - frontend/src/app.css
    - frontend/src/lib/api.ts
    - frontend/bindings/gifsmith/internal/services/gifservice.ts
    - frontend/bindings/gifsmith/internal/services/panelservice.ts
    - frontend/bindings/gifsmith/internal/provider/index.ts
    - frontend/bindings/gifsmith/internal/provider/models.ts
    - scripts/uat/01-tracer.js
  modified:
    - main.go
    - internal/services/actions.go
    - internal/services/actions_test.go
    - frontend/src/App.svelte
    - frontend/src/main.ts
    - frontend/bindings/gifsmith/internal/services/actionservice.ts
    - frontend/bindings/gifsmith/internal/services/index.ts
    - scripts/smoke-browser-loop.sh
    - .gitignore
  deleted:
    - frontend/public/sample.gif

key-decisions:
  - "Malformed Finer responses (missing results, bad fileid) carry the response's HTTP status, matching httpx's invalid-JSON classification; a bad cursor stays Status 0 because no request was made"
  - "Staging downloads use a dedicated http.Client with no client timeout; staging.Fetch's 45 s context is the bound (the 15 s API client would cut off multi-MB originals)"
  - "VendorAllow also rejects explicit ports, so only https on the vendor's default port is fetchable"
  - "On a CopyFile failure the newly staged file stays and nothing is swept; the next successful copy sweeps everything else"
  - "Any Copy/Paste failure other than ErrBusy shows E7 in the footer; ErrBusy is silently ignored"

patterns-established:
  - "UAT scripts: scripts/uat/<plan>-<name>.js run through `smoke-browser-loop.sh --uat`, use browser.newPage(), throw on failure, print UAT OK"
  - "Bound-call smoke checks use method_id <bindings-file> <Func> and call <id> <json-args>"

requirements-completed: [SRCH-01, SRCH-07, ONBD-03, ONBD-05]

coverage:
  - id: D1
    description: "Finer Gifs search runs in Go and returns fixture items with media URLs through the Wails binding (Vite proxy)"
    requirement: SRCH-01
    verification:
      - kind: unit
        ref: "internal/provider/finergifs/finergifs_test.go#TestSearchPaging"
        status: pass
      - kind: integration
        ref: "bash scripts/smoke-browser-loop.sh (GifService.Search check)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Every provider failure is a URL-free typed error (classification table + key redaction) and crosses Wails as err.cause"
    requirement: ONBD-03
    verification:
      - kind: unit
        ref: "internal/provider/httpx/httpx_test.go#TestRedaction"
        status: pass
      - kind: unit
        ref: "internal/provider/httpx/httpx_test.go#TestGetJSONClassification"
        status: pass
      - kind: unit
        ref: "internal/provider/provider_test.go#TestErrorString"
        status: pass
    human_judgment: false
  - id: D3
    description: "Per-provider network/timeout/HTTP errors and Finer-down (530) classification; Finer mapper never reorders or filters and rejects malformed responses"
    requirement: SRCH-07
    verification:
      - kind: unit
        ref: "internal/provider/finergifs/finergifs_test.go#TestFinerDown"
        status: pass
      - kind: unit
        ref: "internal/provider/finergifs/finergifs_test.go#TestMapperFixture"
        status: pass
      - kind: unit
        ref: "internal/provider/httpx/httpx_test.go#TestGetJSONTimeout"
        status: pass
    human_judgment: false
  - id: D4
    description: "Results are kept in returned order with duplicates intact at the mapper (dedupe policy is the search service's, plan 02-03)"
    requirement: ONBD-05
    verification:
      - kind: unit
        ref: "internal/provider/finergifs/finergifs_test.go#TestMapperFixture"
        status: pass
    human_judgment: false
  - id: D5
    description: "Copy/Paste stage the selected GIF (allowlist, 200, magic bytes, 50 MB, traversal-safe path, .part+rename), CopyFile the staged path, sweep only after success; Copy never hides in Go; Clips refused"
    verification:
      - kind: unit
        ref: "internal/services/actions_test.go#TestCopyStagesAndDoesNotHide"
        status: pass
      - kind: unit
        ref: "internal/staging/staging_test.go#TestFetchDisallowedHost"
        status: pass
      - kind: integration
        ref: "bash scripts/smoke-browser-loop.sh (Copy(gif) stages under bin/uat-cache, disallowed URL refused)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Panel UI in the server build: 2+ chars, 450 ms debounce, animated 4-column 166x144 grid, attribution, Enter shows E7 when the pasteboard stub rejects"
    verification:
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/01-tracer.js (screenshots ~/.dev-browser/tmp/gifsmith-01-grid.png, gifsmith-01-copy-error.png)"
        status: pass
    human_judgment: false
  - id: D7
    description: "Native copy of a real GIF (desktop build): file URL on the pasteboard, ⌘V pastes animated; success footer + 400 ms hide on the real panel"
    verification: []
    human_judgment: true
    rationale: "Needs the signed desktop build, the real pasteboard and the user's own apps; scheduled for the 02-06 native checkpoint (Finer Gifs may still be down, so GIPHY/Klipy keys cover it)"

duration: 13min
completed: 2026-10-17
status: complete
---

# Phase 2 Plan 01: Finer Gifs Tracer (search -> grid -> copy) Summary

**Typing in the panel now searches The Finer Gifs Club from Go (typed, URL-free provider errors through one httpx choke point), renders an animated 4-column grid, and Enter downloads, validates and stages the selected GIF before handing its file URL to the pasteboard; the bundled sample is gone and the smoke script is now the browser UAT harness.**

## Performance

- **Duration:** 13 min (wall clock between the recorded start and end)
- **Started:** 2026-10-17T18:38:23Z
- **Completed:** 2026-10-17T18:52:09Z
- **Tasks:** 3
- **Files modified:** 31 (30 changed + 1 deleted)

## Accomplishments

- Provider contract and `httpx.GetJSON`: timeout/network/cancel/429/non-2xx/empty-or-invalid-200 classification, Retry-After read only on 429, nothing wrapped. A secret in the URL path or query never shows up in `Error()` or the `MarshalError` JSON (tested for a refused connection and for a timeout).
- Finer Gifs client on the real wire format (`results[].fields`, `hits.found/start`). It checks `fileid` against `^[A-Za-z0-9._-]+$` before building a URL, rejects malformed responses with a typed error that carries the HTTP status, and makes no request for trending.
- Shared fixture server (`fakeapi`) with 60 Office items and generated animated GIF media. The server build starts it on 127.0.0.1 and points the client at it.
- Copy pipeline: `staging.Fetch` (https plus per-provider host allowlist, 45 s, HTTP 200, GIF87a/89a, 50 MB cap, `<sha1[:16]>/<slug>.gif` via `.part` and rename) → `platform.CopyFile` → `Sweep`. `Copy` no longer hides in Go. `Paste` keeps every Phase 1 rule. Clips are refused before any download.
- Panel UI per UI-SPEC tokens: 48 px search bar, 166×144 tiles with `object-fit: contain`, footer attribution, title/status precedence, copy spinner overlay, success check followed by a 400 ms hide, E7/E8 messages, double-click to copy.
- `smoke-browser-loop.sh` now has default, `--serve` and `--uat <script>` modes. `scripts/uat/01-tracer.js` checks geometry, attribution and the Enter → E7 path in headless Chromium.

## Task Commits

1. **Task 1: tracer, search Finer Gifs and see an animated grid**: `98b8819` (feat)
2. **Task 2: lock the error contract**: `be61336` (test, RED), then `7530d45` (feat, GREEN)
3. **Task 3: Enter copies the real GIF, plus the UAT harness**: `6ac0b46` (test, RED), `ebe185d` (feat, GREEN Go side), `19fe2e3` (feat, frontend + smoke + UAT), then `aa90d54` (style, golangci-lint errcheck)

**Plan metadata:** recorded in the docs commit that follows this summary

## TDD Gate Compliance

- **Task 2:** RED `be61336` → GREEN `7530d45`. The RED evidence is `bin/red-02-01-t2-record.json` (TAP from `go test -json`). The classifier returned `RED_EVIDENCE_OK` with target `TestMapperMissingResults`.
  - **Semantic assessment:** both target tests (`TestMapperMissingResults`, `TestMapperRejectsBadFileid`) compiled, ran, and failed only on the planned assertion: the typed error had Status 0 instead of the response's 200. No load or fixture faults.
  - **Unexpected GREEN:** the other Task 2 tests passed on first run because Task 1's tracer spec already required those behaviors (classification, redaction, Remap). To prove these tests can fail, I mutated `httpx.go` temporarily (restored before committing):
    - Wrapping the transport error with `%w` failed `TestRedaction` and `TestGetJSONNetwork`. The SECRETKEY123 URL appeared in the message.
    - Reading Retry-After on non-429 responses failed `TestGetJSONClassification/401_ignores_retry-after`.
- **Task 3:** RED `6ac0b46` → GREEN `ebe185d`. The RED commit includes a staging API skeleton and the new `Copy/Paste(ctx, gif)` signatures so the tests compile and fail on assertions, not on the build. Evidence is `bin/red-02-01-t3-record.json`: `RED_EVIDENCE_OK`, target `TestCopyStagesAndDoesNotHide`.
  - **Semantic assessment:** the target ran and failed on the planned assertions. The message was still Phase 1's. The calls were `CopyFile:<sample>` plus `hide`, not the staged path. The older staged copy was not swept. 16 of 25 tests failed this way, and none failed to compile.
- No REFACTOR commits. The `style` commit only adds explicit error discards and test-setup checks.

## Files Created/Modified

- `internal/provider/provider.go`: contract types, `Kind`, URL-free `Error`, `MarshalError`
- `internal/provider/httpx/httpx.go`: `GetJSON`, `retryAfter`, `Remap`
- `internal/provider/finergifs/finergifs.go`: Finer client and mapper
- `internal/provider/fakeapi/fakeapi.go`: fixture `/finer/search` and `/media/<name>.gif`
- `internal/staging/staging.go`: `Store.Fetch`/`Sweep`, `VendorAllow`, `slug`, the four sentinel errors
- `internal/services/gifs.go`, `internal/services/panel.go`: `GifService`, `PanelService`
- `internal/services/actions.go`: `Copy`/`Paste(ctx, gif)` through `stage()`; `SamplePath` removed
- `main.go`, `wire_desktop.go`, `wire_server.go`: deps split, MarshalError, three services registered untagged
- `frontend/src/app.css`, `frontend/src/lib/api.ts`, `frontend/src/App.svelte`, `frontend/src/main.ts`: tokens, facade, panel
- `scripts/smoke-browser-loop.sh`, `scripts/uat/01-tracer.js`: harness and tracer UAT
- `.gitignore`: `/gifsmith` (the plan's `go build -tags server .` verify drops this binary in the repo root)

## Decisions Made

See `key-decisions` in the frontmatter. In short: malformed responses keep their HTTP status, downloads get their own client, explicit ports are rejected, and a failed CopyFile sweeps nothing.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] The staging download used the 15 s API client**
- **Found during:** Task 3 (wiring `staging.Store` in `main.go`)
- **Issue:** The plan wires `&staging.Store{HTTP: d.HTTP}`. `d.HTTP` has a 15 s total timeout (3 s in the server build), so the planned 45 s download deadline (T-02-05) could never take effect, and a 4 MB GIPHY or Klipy original on a slow link would fail.
- **Fix:** Downloads use a dedicated `&http.Client{}`. `Fetch`'s `context.WithTimeout(45 s)` bounds them.
- **Files modified:** `main.go`
- **Verification:** smoke Copy(gif) stages successfully; staging tests pass
- **Committed in:** `ebe185d`

**2. [Rule 2 - Missing critical] VendorAllow rejects explicit ports**
- **Found during:** Task 3
- **Issue:** An allowlist that checks only the hostname also accepts `https://media2.giphy.com:8443/…`.
- **Fix:** `u.Port() != ""` is rejected. Tested in `TestVendorAllow`.
- **Committed in:** `6ac0b46` (test) / `ebe185d` (impl)

**3. [Rule 3 - Blocking] `.gitignore` gains `/gifsmith`**
- **Found during:** Task 1 verify
- **Issue:** The plan's `CGO_ENABLED=0 go build -tags server .` writes an untracked `gifsmith` binary at the repo root on every run.
- **Fix:** Ignore it.
- **Committed in:** `98b8819`

**4. [Rule 2 - Quality gate] golangci-lint errcheck**
- **Found during:** final checks. The Phase 1 baseline had 0 lint issues; the new packages introduced 14.
- **Fix:** Explicit discards for Close, cleanup Removes and fixture writes. Test setup errors now fail the test.
- **Committed in:** `aa90d54`

---

**Total deviations:** 4 auto-fixed (1 bug, 2 missing-critical/quality, 1 blocking)
**Impact on plan:** Each fix is required for correctness or tooling hygiene. No scope added.

## Issues Encountered

- Task 2 RED was only partial, because the tracer already implemented most of the contract. This is handled as described under TDD Gate Compliance: a genuine RED on malformed-response status, plus mutation checks for the rest.

## Known Gaps (planned for later plans, not stubs)

- A search failure shows only the generic footer line "Couldn't search The Finer Gifs Club." Per-kind states (E1–E5f, S2–S12), the search-bar spinner and superseded-request cancellation arrive in 02-03. Today a sequence counter drops stale responses.
- There are no arrow keys or paging yet; only the first page (24) renders (02-02). The grid is not virtualized (02-02).
- `GifService` has no direct unit test, so services coverage is 67.5%. The critical paths it fronts (httpx 82.5%, finergifs 100%, staging 90.6%, actions) are covered, and the smoke check calls `GifService.Search` end to end.

## User Setup Required

None. No external service configuration required.

## Next Phase Readiness

- Ready for 02-02: grid virtualization and keyboard on top of `App.svelte` and the `--uat` harness. Its package install (`@tanstack/svelte-virtual`) needs the human package gate.
- `provider.Gif`/`Page`/`Error` and the generated TS models are now the wire contract for GIPHY and Klipy (02-04).
- Finer Gifs live status was not probed in this plan. Native copy of a real GIF is a 02-06 user check.

## Self-Check: PASSED

- Files: all `key-files.created` entries exist; `frontend/public/sample.gif` is absent
- Commits on this branch: 98b8819, be61336, 7530d45, 6ac0b46, ebe185d, 19fe2e3, aa90d54
- Plan verification: gofmt clean; `go vet . ./internal/...` pass; `go test -race ./internal/...` 7/7 packages ok; `CGO_ENABLED=0 go build -tags server .` ok; desktop `go build .` ok; `pnpm check` 0 errors / 0 warnings; smoke `SMOKE OK`; `--uat scripts/uat/01-tracer.js` `UAT OK`; golangci-lint 0 issues
- Acceptance criteria (Tasks 1–3): all PASS (re-run after the final commit)

---
*Phase: 02-providers-search-ui-first-run*
*Completed: 2026-10-17*
