---
phase: 02-providers-search-ui-first-run
fixed_at: 2026-10-17T23:00:00Z
review_path: .planning/phases/02-providers-search-ui-first-run/02-REVIEW.md
iteration: 1
findings_in_scope: 6
fixed: 6
skipped: 0
status: all_fixed
---

# Phase 2: Code Review Fix Report

**Fixed at:** 2026-10-17T23:00:00Z
**Source review:** .planning/phases/02-providers-search-ui-first-run/02-REVIEW.md
**Iteration:** 1
**Scope:** critical_warning (CR-01, CR-02, WR-01 to WR-04). IN-01 to IN-07 are out of scope and were not changed.

**Summary:**
- Findings in scope: 6
- Fixed: 6 (CR-01, WR-02 and WR-04 need human verification, see below)
- Skipped: 0

`workflow.use_worktrees` is `false`, so all edits and commits were made in the main checkout on `gsd/phase-02-providers-search-ui-first-run`. No worktree was created.

## Fixed Issues

### CR-01: GIPHY and Clips share one key and one rate limit, but the 429 block is per provider ID, and the Clips probe bypasses it

**Status:** fixed: requires human verification (state-machine logic)
**Files modified:** `internal/search/search.go`, `internal/search/clips.go`, `main.go`, `internal/search/search_test.go`, `internal/search/clips_test.go`, `internal/services/settings_test.go`
**Commits:** 814e7e4, f8d86c5 (errcheck follow-up in the new test)
**Applied fix:**
- `search.go`: `budget(p)` maps `clips` to `giphy`. The block read/write, `Blocked`, and `Invalidate` all use the budget, so `Invalidate(GIPHY)` also drops cached `clips|` pages. A new `Service.Guard(p, call)` runs an uncached call inside the budget: it refuses while blocked, and a 429 it returns blocks the budget. Blocked errors keep the requesting provider ID (`clips`), so the UI names the right tab.
- `main.go`: the Clips probe runs through `svc.Guard(provider.Clips, ...)`. The services test fixture is wired the same way.
- `clips.go`: `Recheck` re-probes at most once per key per launch (`rechecked`, cleared by `Reset`). After that it returns the last probe error (`lastErr`) without a request, so the panel shows an error rather than a skeleton.
- Tests: `TestClipsShareGiphyBudget` (a GIPHY 429 blocks Clips and the reverse, Guard refuses into a block, a Guard 429 blocks GIPHY, Invalidate(giphy) drops Clips pages), `TestClipsProbe429BlocksRecheck` (a probe 429 makes the next Recheck answer rate_limited with no probe), and an updated `TestClipsRecheck` (second recheck replays the error, Reset restores the re-probe). `TestCheckClipsUnknown` now changes the key the way `SaveKey` does (Invalidate + Reset).
- **Behavior to confirm:** once the single re-probe is spent, Try Again on a Clips error replays the stored error and makes no request until relaunch or a GIPHY key change. This follows RESEARCH ("re-probes once", "one call per launch at most"), but the Try Again button no longer does anything visible in that state.

### CR-02: A revoked GIPHY key on the Clips tab renders a blank content area with no recovery

**Status:** fixed
**Files modified:** `frontend/src/App.svelte`, `scripts/uat/06-clips.js`
**Commit:** e8be357
**Applied fix:** A new render branch handles a Clips `key_rejected`/`key_missing` error. It shows "GIPHY rejected your saved key" with the body "Clips use your GIPHY key. Paste a new key on the GIPHY tab to continue." and a primary "Go to GIPHY" button. `keySaved('giphy')` now also calls `setResults('clips', null, null)`. New UAT steps in `06-clips.js` search "revoked" on Clips, assert the message, go to GIPHY, save a new fixture key, and confirm ⌘2 loads Clips again. The UAT failed with the fix stashed (at the heading wait) and passes with it.

### WR-01: An in-flight request with the old key can land after `Invalidate` and restore the old cache entry or 429 block

**Status:** fixed
**Files modified:** `internal/search/search.go`, `internal/search/search_test.go`
**Commit:** 08e0ff5
**Applied fix:** Each budget has a generation counter that `Invalidate` increments. `get()` and `Guard()` capture the generation before the call. After the call, a 429 is recorded and a page is cached only if the generation has not changed. The caller still receives its answer. `TestInvalidateDropsInFlightResult` covers a late 429 and a late success. Both subtests failed with the fix stashed.

### WR-02: The persisted Clips tab can never be restored

**Status:** fixed: requires human verification (native relaunch)
**Files modified:** `frontend/src/App.svelte`, `scripts/uat/06-clips.js`
**Commit:** e3501b7
**Applied fix:** I took the review's first option. At boot, a persisted `clips` is restored whenever a GIPHY key is configured; without a key the panel still opens on GIPHY. If Clips is not yet enabled, `bootClips()` fetches GIPHY trending behind the S2 skeleton, and that success starts the probe. `refreshProviders` then shows Clips once the probe lands. If the GIPHY fetch fails, the panel shows its error on the Clips tab, and a rejected key gets the CR-02 "Go to GIPHY" message. If the fetch succeeds but no probe starts, the panel calls `recheckClips()` once. `06-clips.js` now reloads the page and asserts that it reopens on Clips with clip tiles. The browser UAT cannot reach the path where Clips is still unknown at boot, because the server build keeps keys in memory and loses them on restart. Confirm that path on the native app: quit while on Clips, then relaunch.

### WR-03: Staging downloads follow redirects without re-checking the host allowlist

**Status:** fixed
**Files modified:** `internal/staging/staging.go`, `internal/staging/staging_test.go`
**Commit:** 3f076fa
**Applied fix:** `Store.Fetch` copies its `http.Client` for each call and sets a `CheckRedirect` that applies the same allowlist to every hop (with `g.Provider`) and stops after 5 redirects. A refused redirect returns `ErrNotAllowed`, which is fixed text with no URL. `main.go` needed no change. `TestFetchRedirectAllowlist` covers three cases. A redirect to a host outside the allowlist returns ErrNotAllowed and the other host gets 0 hits. A same-host redirect succeeds. A redirect loop returns ErrNotAllowed. The test failed with the fix stashed.

### WR-04: The Klipy mapper drops ad items despite the no-filter rule, and a page of only ads keeps the grid fetching

**Status:** fixed: requires human verification (**needs a product call**)
**Files modified:** `internal/provider/klipy/klipy.go`, `internal/provider/klipy/klipy_test.go`, `internal/provider/klipy/testdata/search.json`
**Commit:** 50824e1
**Applied fix:**
- Mapping: a non-`gif` item that carries GIF renditions (an `sm` preview plus an `hd` or `md` download) is now mapped in its original position, like any other result. The tile is a plain `<img>`, so no ad HTML is ever rendered. Only items with nothing a GIF grid can show, such as an HTML `content` ad, are skipped and still logged (`klipy: skipped N non-gif items`). Copying an ad hosted outside `klipy.com` fails the staging allowlist with the normal "couldn't copy" message.
- Fetch storm: `Next` is set only when `has_next` is true and the page mapped at least one item, the same guard the GIPHY and Finer clients use. A page with nothing renderable now ends paging instead of starting back-to-back requests.
- Tests: the fixture adds an HTML-only ad. `TestMapping` expects the GIF-rendition ad to be kept in place and the HTML ad to be skipped and logged. `TestPaging` asserts that a page of only unrenderable items has no `Next`.
- **Product decision still open:** RESEARCH A4 expects ads never to reach a desktop user agent on a self-issued key. If they do arrive, ads now appear as tiles with no "Sponsored" label, and unrenderable ads are still dropped. Before release, confirm the ad policy (label, render, or omit) with developers@klipy.com, together with the copy-file clearance that is already pending. The PROJECT.md compliance row ("logged skip of Klipy non-gif items") still holds, but it now covers only unrenderable items.

## Verification

All gates ran in the **main checkout** (`/Users/ken/dev/gif-search`, branch `gsd/phase-02-providers-search-ui-first-run`, HEAD `f8d86c5`). No isolated worktree was used, so these results can be reproduced from this tree.

| Gate | Result |
|------|--------|
| `gofmt -l .` (excluding the pre-existing `build/ios` template files) | clean |
| `go vet . ./internal/...` and `go vet -tags server .` | clean |
| `go test -race -count=1 ./internal/...` | 11 packages ok |
| `golangci-lint run . ./internal/...` | 0 issues (after f8d86c5) |
| `pnpm --dir frontend run check` | 0 errors, 0 warnings |
| `pnpm --dir frontend run test` | pass 23, fail 0 |
| `bash scripts/smoke-browser-loop.sh` | SMOKE OK |
| `--uat` 01-tracer, 02-grid, 02-keyboard, 03-states, 05-setup, 05-tabs, 06-clips | UAT OK each |
| `FAKE_CLIPS=denied --uat` 06-clips-denied, 06-backstops | UAT OK each |

`go.sum` was not touched. Only fixture keys were used (`uat-key-123`, `uat-key-456`). The server build uses the in-memory secret store, so the login Keychain was never accessed. No synthetic desktop input was sent, and only the processes started by the smoke script were stopped, by that script's own cleanup.

---

_Fixed: 2026-10-17T23:00:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
