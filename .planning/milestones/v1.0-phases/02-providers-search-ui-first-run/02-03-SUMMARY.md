---
phase: 02-providers-search-ui-first-run
plan: 03
subsystem: search service (cache, 429 budget, dedupe), panel content states
tags: [go, search-cache, rate-limit, svelte5, wails-v3, cancellation, node-test, dev-browser]

requires:
  - phase: 02-providers-search-ui-first-run
    provides: "02-01 provider contract, httpx classification, fakeapi, lib/api.ts; 02-02 ResultGrid paging row, routeKey, smoke --uat harness"
provides:
  - "internal/search: Service with New/Search/Trending/Seed/Invalidate, TrendingTTL 10 min, SearchTTL 5 min, MinBlock 60 s"
  - "GifService{Svc *search.Service}: every webview call goes through the cache and the 429 block"
  - "fakeapi.Scenario: ratelimit, noretry, slow, boom, blank, pagelimit, pageboom (empty stays per route) for GIPHY/Klipy routes to reuse"
  - "lib/errors.ts: providerName, errorState (E1/E3/E4/E5/E5f with retryAt), pagingMessage (E2/E6)"
  - "lib/api.ts: latest(slot, promise) and cancel(slot); slots first:<provider> and page:<provider>"
  - "StateMessage.svelte; App state held per provider id (status idle|loading|refreshing|ready|empty|error)"
  - "scripts/uat/03-states.js"
affects: [02-04, 02-05, 02-06]

actuals:
  tokens: 15000   # chars/4 over the final content of the 14 changed files (59,944 chars)
  tasks: 2
  commits: 4
plan_head_before: 5962d87a0483d62e24f259c6e8399c4da892917a
plan_head_after: 8a1b7c52ace7271347960439e8b1c69bed1145ae

tech-stack:
  added: []
  patterns:
    - "search.Service owns the policy (cache, block, dedupe); providers stay dumb clients; one mutex, never held across a provider call"
    - "A blocked provider answers rate_limited from Go with the remaining seconds (0 stays unknown), so Try Again inside the window costs no request"
    - "Error copy is computed from (error, now, receivedAt): the countdown re-renders by moving `now`, retryAt stays fixed"
    - "Per-provider UI state record in App.svelte; the active provider is still the constant PROVIDER until the 02-05 tabs"

key-files:
  created:
    - internal/search/search.go
    - internal/search/search_test.go
    - frontend/src/lib/errors.ts
    - frontend/src/lib/errors.test.ts
    - frontend/src/components/StateMessage.svelte
    - scripts/uat/03-states.js
  modified:
    - internal/services/gifs.go
    - internal/provider/fakeapi/fakeapi.go
    - main.go
    - wire_server.go
    - frontend/src/lib/api.ts
    - frontend/src/components/ResultGrid.svelte
    - frontend/src/App.svelte
    - .gitignore

key-decisions:
  - "Copy discretion (a): a 2xx http error drops the '(HTTP n)' parenthetical (D-23); an unknown Retry-After reads 'Try again later this hour.' and shows Try Again"
  - "Copy discretion (b): n == 1 renders '1 minute'; every other n renders 'n minutes'; under 60 s 'in less than a minute' with no '(at …)'"
  - "Once a known retry window has passed, E1 drops its 'Try again …' sentence and shows Try Again; the paging row falls back to E6 'Couldn't load more.' with its Try Again link"
  - "errorState takes an optional third argument receivedAt (default now) so the minute countdown is re-evaluated against the time the error arrived"
  - "A first-page error with no provider cause (not a provider.Error) renders E5 without a status, never the caught text"
  - "enableAutoPaste failure shows a fixed string (\"Couldn't ask macOS for Accessibility access. Try again.\") instead of String(err)"
  - "fakeapi scenario words are case-insensitive and checked pageboom before boom; slow honours the request context so the fixture never holds a cancelled handler"

patterns-established:
  - "Rate-limit UAT ordering: the Go block is per provider, so noretry runs first, the script waits out MinBlock, then pagelimit starts a 120 s block that serves ratelimit"

requirements-completed: [SRCH-03, SRCH-06, SRCH-07, ONBD-05]

coverage:
  - id: D1
    description: "Trending cached 10 min and search 5 min per provider|kind|query|cursor in memory; repeat inside TTL makes zero provider calls"
    requirement: SRCH-03
    verification:
      - kind: unit
        ref: "internal/search/search_test.go#TestTrendingTTL"
        status: pass
      - kind: unit
        ref: "internal/search/search_test.go#TestSearchTTL"
        status: pass
      - kind: unit
        ref: "internal/search/search_test.go#TestSeparateKeys"
        status: pass
    human_judgment: false
  - id: D2
    description: "After a 429 the provider is blocked for max(Retry-After, 60 s) with no HTTP; remaining seconds only when known; Seed/Invalidate"
    requirement: SRCH-06
    verification:
      - kind: unit
        ref: "internal/search/search_test.go#TestBlockKnownRetry"
        status: pass
      - kind: unit
        ref: "internal/search/search_test.go#TestBlockUnknownRetry"
        status: pass
      - kind: unit
        ref: "internal/search/search_test.go#TestShortRetryAfterBlocksMinBlock"
        status: pass
      - kind: unit
        ref: "internal/search/search_test.go#TestSeedInvalidate"
        status: pass
    human_judgment: false
  - id: D3
    description: "Same-page dedupe for every provider except Klipy, which is returned exactly as received"
    requirement: ONBD-05
    verification:
      - kind: unit
        ref: "internal/search/search_test.go#TestDedupeNonKlipy"
        status: pass
      - kind: unit
        ref: "internal/search/search_test.go#TestNoDedupeKlipy"
        status: pass
    human_judgment: false
  - id: D4
    description: "Errors and cancellations are never cached or blocked; concurrent calls are race-free"
    requirement: SRCH-07
    verification:
      - kind: unit
        ref: "internal/search/search_test.go#TestNoCacheOnError"
        status: pass
      - kind: unit
        ref: "internal/search/search_test.go#TestCancelNotCachedNotBlocked"
        status: pass
      - kind: unit
        ref: "internal/search/search_test.go#TestConcurrent (-race)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Exact UI-SPEC copy for E1 (known/unknown/under a minute/passed), E3, E4, E5 (with and without status), E5f, E2, E6; no URL or 'undefined'"
    requirement: SRCH-06
    verification:
      - kind: unit
        ref: "frontend/src/lib/errors.test.ts (10 tests, # fail 0)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Browser: S8, S7, S2 skeleton + spinner, S11, E5f, 2xx E5, S4 keep-results, superseded search not rendered (Go log shows it cancelled), S12a, unknown-retry E1 + free Try Again, S9a E2 without Try Again, known-retry E1 without Try Again; no URL/host in page text at every step"
    requirement: SRCH-07
    verification:
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/03-states.js (screenshots ~/.dev-browser/tmp/gifsmith-03-ratelimit.png, gifsmith-03-ratelimit-unknown.png)"
        status: pass
    human_judgment: false
  - id: D7
    description: "One provider's error never changes another provider's state"
    requirement: SRCH-07
    verification: []
    human_judgment: true
    rationale: "State is held per provider id, but only Finer Gifs is wired until 02-04/02-05 add GIPHY, Klipy and the tabs; cross-tab isolation is exercised there and at the 02-06 checkpoint"

duration: 15min
completed: 2026-10-17
status: complete
---

# Phase 2 Plan 03: Search Service and Content States Summary

**Every provider call now goes through a Go search.Service that caches pages (trending 10 min, search 5 min), refuses to call a provider during a local 429 block, and dedupes non-Klipy pages. The panel cancels superseded searches through CancellablePromise and renders every UI-SPEC loading, empty, rate-limit and error state from fixed copy rows, never from error text.**

## Performance

- **Duration:** about 15 min (2026-10-17T20:09Z to 20:24Z)
- **Tasks:** 2 (both TDD)
- **Files modified:** 14 (6 created, 8 modified)

## Accomplishments

- `internal/search`: one `sync.Mutex` guards the cache and the block map, and it is released around the provider call. The cache key is `provider|kind|trimmed query|cursor`. A 429 blocks the provider for `max(Retry-After, 60 s)`. While blocked, every request returns `rate_limited` with the remaining seconds rounded up, or 0 when the vendor sent no Retry-After. Errors and cancellations are never cached. `Seed` and `Invalidate` are ready for 02-04 key verification. Nothing is written to disk.
- `GifService` delegates to `search.Service`. It still validates ids with `ParseID` and does not log `context.Canceled`.
- `fakeapi.Scenario` handles the scenario words for any route. In the server build, `SlowDelay` is 5 s against a 3 s client timeout.
- `errors.ts` holds the UI-SPEC rows. `latest()`/`cancel()` in `api.ts` provide one slot per kind and provider.
- `App.svelte` keeps state per provider and renders exactly one content state. Enter runs a visible Try Again when nothing is selected. A countdown refreshes once a minute and again exactly when the retry window ends. The Phase 1 "Couldn't search…" footer line and the `String(err)` status are gone.
- `03-states.js` checks each state in a headless browser and asserts that no URL or fixture host leaks at any step. The server log shows two timeouts rather than three, which proves the superseded slow search was cancelled in Go.

## Task Commits

1. **Task 1: search.Service**: `d2fb58a` (test, RED), then `8538428` (feat, GREEN)
2. **Task 2: states and cancellation**: `9bdd0c0` (test, RED), then `8a1b7c5` (feat, GREEN)

## TDD Gate Compliance

- **Task 1:** RED `d2fb58a` → GREEN `8538428`. The evidence is in `bin/red-02-03-t1-record.json` (TAP converted from `go test -json`). The classifier returned `RED_EVIDENCE_OK` with target `TestBlockKnownRetry`.
  - **Semantic assessment:** the package compiled against a pass-through skeleton and all 13 tests ran. The target failed on its first planned assertion (`err = <nil>, want rate_limited`). The other 6 RED failures were also assertion failures: TTL, unknown block, MinBlock floor, dedupe, and seed. The 6 tests that passed check behaviour a pass-through already has: separate keys, no error caching, cancel passthrough, Klipy untouched, unknown id, and race safety.
- **Task 2:** RED `9bdd0c0` → GREEN `8a1b7c5`. The evidence is in `bin/red-02-03-t2-record.json` (node --test TAP). The classifier returned `RED_EVIDENCE_OK` with target "rate limit with a known retry time".
  - **Semantic assessment:** the file loaded under Node type stripping and the target failed on `'E5' !== 'E1'`. 9 of the 10 errors tests failed on assertions. The leak guard passed against the stub's empty strings, which is the guard's expected property.
- No REFACTOR commits.

## Files Created/Modified

- `internal/search/search.go`, `search_test.go`: service and clock-injected tests
- `internal/services/gifs.go`, `main.go`: GifService on `search.New(...)`
- `internal/provider/fakeapi/fakeapi.go`, `wire_server.go`: scenario words and SlowDelay
- `frontend/src/lib/errors.ts`, `errors.test.ts`: copy rows and retry math
- `frontend/src/lib/api.ts`: `latest`, `cancel`
- `frontend/src/components/StateMessage.svelte`: centered message
- `frontend/src/components/ResultGrid.svelte`: `pagingNote` / `pagingRetry` props
- `frontend/src/App.svelte`: per-provider state, S2 skeleton, search-bar spinner, states, countdown
- `scripts/uat/03-states.js`: state UAT
- `.gitignore`: see Deviation 1

## Decisions Made

See `key-decisions` in the frontmatter. The two copy discretion rules the plan asked to record are (a) and (b) there.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `/gifsmith` was not actually ignored**
- **Found during:** Task 1 verify (`go build .` left an untracked `gifsmith` binary)
- **Issue:** 02-01 appended `/gifsmith` to a `.gitignore` with no trailing newline. The result was one line, `build/android/gen//gifsmith`, which ignored neither path.
- **Fix:** Split it back into `build/android/gen/` and `/gifsmith`.
- **Committed in:** `8538428`

**2. [Rule 1 - Plan ordering] UAT scenario order vs the 429 block**
- **Found during:** Task 2 (UAT design)
- **Issue:** The plan lists `ratelimit` before `noretry`, `boom`, `slow` and the rest. A correct per-provider block turns every later Finer request into `rate_limited` for 120 s, so the listed order cannot pass against a correct implementation.
- **Fix:** Non-429 scenarios run first. `noretry` comes next; its Try Again is pressed to show it costs nothing. The script then waits 61 s. `dwight pagelimit` follows (a fresh 429 with Retry-After 120), and then `ratelimit`, which the Go block serves. Every assertion the plan lists is kept. The run takes about 85 s, within the harness's 120 s limit.
- **Committed in:** `8a1b7c5`

**3. [Rule 2 - Missing critical] `String(err)` in `enableAutoPaste`**
- **Issue:** The status line rendered the caught error text (T-02-12).
- **Fix:** A fixed copy string.
- **Committed in:** `8a1b7c5`

**Total deviations:** 3 auto-fixed. **Impact:** no scope added and no packages added.

## Issues Encountered

- Blocked calls answered by Go are logged as `search finergifs: finergifs: rate_limited`, the same as a real 429. The log is URL-free; this is only noise.
- The generated `gifservice.ts` doc comment still carries the old `GifService` Go comment. It will refresh on the next bindings generation; signatures are unchanged.

## Not exercised by automation

- Cross-provider isolation (D7) needs the GIPHY/Klipy clients and the tabs (02-04/02-05).
- The reduced-motion skeleton and spinner glyph were checked by reading the CSS, not in an emulated browser.
- No native behaviour changed in this plan.

## User Setup Required

None.

## Next Phase Readiness

- 02-04 can call `Svc.Seed(p, "", page)` after key verification and `Svc.Invalidate(p)` on a key change. GIPHY and Klipy fixture routes can reuse `fakeapi.Scenario(w, r, q, firstPage, o)` and add their own `empty` bodies.
- 02-05 can make `PROVIDER` switchable. It should initialise `states` for all four ids and render `key_missing`/`key_rejected` (for which `errorState` returns null) as the setup card.

## Self-Check: PASSED

- Files: all 6 `key-files.created` exist
- Commits: d2fb58a, 9bdd0c0, 8538428, 8a1b7c5 are on HEAD (`git rev-list --count 5962d87..HEAD` = 4)
- Verification: gofmt clean; `go vet . ./internal/...` ok; `go test -race ./internal/... -count=1` 8 packages ok, no FAIL; `golangci-lint run . ./internal/...` 0 issues; server and desktop builds ok; `pnpm run test` # pass 17 / # fail 0; `pnpm run check` 0 errors / 0 warnings; smoke `SMOKE OK`; `--uat` 03-states, 02-grid, 02-keyboard and 01-tracer each print `UAT OK`
- Acceptance greps (Tasks 1–2): all match

---
*Phase: 02-providers-search-ui-first-run*
*Completed: 2026-10-17*
