---
phase: 02-providers-search-ui-first-run
fixed_at: 2026-10-17T21:57:42Z
review_path: .planning/phases/02-providers-search-ui-first-run/02-REVIEW.md
iteration: 2
findings_in_scope: 3
fixed: 3
skipped: 0
status: all_fixed
---

# Phase 2: Code Review Fix Report (iteration 2)

**Fixed at:** 2026-10-17T21:57:42Z
**Source review:** .planning/phases/02-providers-search-ui-first-run/02-REVIEW.md (iteration 2)
**Iteration:** 2
**Scope:** critical_warning (WR-01 to WR-03). IN-01 to IN-09 are out of scope and were not changed.

**Summary:**
- Findings in scope: 3
- Fixed: 3 (WR-02 and WR-03 need human verification, see below)
- Skipped: 0

`workflow.use_worktrees` is `false`, so all edits and commits were made in the main checkout on `gsd/phase-02-providers-search-ui-first-run`. No worktree was created. While this ran, another session committed 83dca8c (`docs(requirements)`, `.planning/REQUIREMENTS.md` and `ROADMAP.md`) on the same branch, between bc4bc08 and fc0906e. None of the fix commits include its files.

## Fixed Issues

### WR-01: "Go to GIPHY" from the Clips rejected-key message usually lands on a GIPHY grid with no key field

**Status:** fixed
**Files modified:** `frontend/src/App.svelte`, `scripts/uat/06-clips.js`
**Commit:** bc4bc08
**Applied fix:** The button now clears GIPHY's results and sets GIPHY's state to `key_rejected` before it calls `switchTo('giphy')`. The setup card therefore shows even when GIPHY still holds cached trending results for the key. In `06-clips.js`, the CR-02 step now clears the search box after the rejection message appears and before it clicks "Go to GIPHY". It then waits for `.card input`. This is the real revoked-key case, where nothing forces a new GIPHY request. With the App.svelte change reverted, the UAT fails at that wait. With the change, it passes.

### WR-02: A Clips re-probe before the launch's first GIPHY success leaves Clips unavailable for the rest of the launch

**Status:** fixed: requires human verification (state-machine logic, native relaunch path)
**Files modified:** `internal/search/clips.go`, `internal/search/clips_test.go`, `frontend/src/App.svelte`
**Commit:** fc0906e
**Applied fix:**
- `clips.go`: three flags replace `probed`:
  - `triggered`: the GIFs-success probe has run.
  - `proven`: `Trigger` has been called, which means a GIFs call succeeded with this key.
  - `rechecked`: the one re-probe is spent.
- A Recheck before any GIFs success spends neither probe.
- `Trigger` is skipped only while a probe is running. Even then it sets `proven`, so the in-flight probe's answer counts.
- `classify` returns `disabled` for 401/403/404 only when `proven` is set. Before that, the result stays `unknown` and the caller gets the `key_rejected` error.
- `App.svelte`:
  - Choosing Clips while its result is unknown (`switchTo`) and Try Again on Clips (`retry`) now call `bootClips()`.
  - `bootClips()` now calls `trending('giphy', '')` directly. It no longer goes through `runFirst`, so it does not change the GIPHY tab's grid. The call is usually served from the 10-minute cache. A failure is shown on the Clips tab, and a rejected key gets the Go to GIPHY message. On success, `bootClips()` calls `recheckClips()`, which also clears an error copied there earlier.
- **Deviation from the review's sketch:** the review routed only `retry()` through GIPHY, and only when GIPHY was in error. I routed `switchTo` as well, every time. Consider a launch on the Klipy tab where GIPHY is never fetched, followed by ⌘2. Without the change, that path would probe before any GIFs success. With the new `classify`, a valid key that GIPHY has not approved for Clips would then show "GIPHY rejected your saved key".
- Tests: `TestClipsRecheckBeforeGifsSuccess` covers three cases. Two Rechecks before success time out, and a later `Trigger` still probes and enables Clips. A 401 before success stays `unknown`, and after `Trigger` it is `disabled`. A `Trigger` that arrives during an in-flight Recheck causes no second probe, and the 401 then counts as `disabled`. Against the old `clips.go` the test fails at the second Recheck.
- **Not covered by browser UAT:** the server build keeps keys in memory, and `SaveKey` always runs `Trigger`. The browser therefore cannot reach a launch where Clips is unknown before any GIPHY success. Confirm it on the native app: go offline, quit while on Clips, relaunch, press Try Again once back online, and check that Clips loads (or shows C2 for an unapproved key).

### WR-03: After the re-probe is spent, Try Again replays a stale rate-limit error and restarts its countdown

**Status:** fixed: requires human verification (state-machine logic)
**Files modified:** `internal/search/clips.go`, `internal/search/clips_test.go`, `internal/search/search.go`, `main.go`, `internal/services/settings_test.go`
**Commit:** e360050
**Applied fix:**
- `finish` clears `rechecked` when the probe returned `rate_limited`. A 429 says nothing about Clips access, and the 429 block already spaces out the next probe. After the block ends, Try Again probes again, so the stale countdown is never replayed.
- Fast path: `Clips.Blocked` is now `func() error` and is wired to a new `search.Service.BlockedErr(provider.Clips)`. `Blocked(p)` now calls `BlockedErr`. A Recheck inside the block returns the service's error with the real seconds left, instead of an error with no countdown.
- Tests: `TestClipsRateLimitedRecheckKeepsReprobe` uses a fake clock and the real `Guard`. A re-probe that gets a 429 is followed by a Recheck 30 s later, which must report 90 s left with no probe. After the block ends, a Recheck must probe and enable Clips. Both checks fail with the fix reverted. `TestClipsRecheck` used a 429 to exercise the re-probe cap, so it now uses a timeout. The cap still applies to every error except a 429.

## Verification

All gates ran in the **main checkout** (`/Users/ken/dev/gif-search`, branch `gsd/phase-02-providers-search-ui-first-run`, HEAD `e360050`). No isolated worktree was used, so these results can be reproduced from this tree.

| Gate | Result |
|------|--------|
| `gofmt -l .` (excluding the pre-existing `build/ios` template files) | clean |
| `go vet . ./internal/...` and `go vet -tags server . ./internal/...` | clean |
| `go test -race -count=1 ./internal/...` | 11 packages ok (fakeapi has no tests) |
| `golangci-lint run . ./internal/...` | 0 issues |
| `pnpm --dir frontend run check` | 0 errors, 0 warnings |
| `pnpm --dir frontend run test` | pass 23, fail 0 |
| `bash scripts/smoke-browser-loop.sh` | SMOKE OK |
| `--uat` 01-tracer, 02-grid, 02-keyboard, 03-states, 05-setup, 05-tabs, 06-clips | UAT OK each |
| `FAKE_CLIPS=denied --uat` 06-clips-denied, 06-backstops | UAT OK each |

`go.mod` and `go.sum` were not touched. Only fixture keys were used (`uat-key-123`, `uat-key-456`). The server build uses the in-memory secret store, so the login Keychain was never accessed. No synthetic desktop input was sent. The only processes stopped were the ones the smoke script started, through its own cleanup, and ports 8089 and 9245 were free afterwards.

## Iteration 1 (history)

Iteration 1 fixed all 6 in-scope findings from the first review:

| Finding | Commit(s) |
|---------|-----------|
| CR-01 shared GIPHY/Clips 429 budget | 814e7e4, f8d86c5 (errcheck follow-up) |
| CR-02 blank Clips area on a rejected key | e8be357 |
| WR-01 in-flight write after Invalidate | 08e0ff5 |
| WR-02 persisted Clips tab unreachable | e3501b7 |
| WR-03 staging redirects bypass the allowlist | 3f076fa |
| WR-04 Klipy ad filtering and empty-page fetch storm | 50824e1 |

CR-01, WR-02 and WR-04 were marked "requires human verification". WR-04's ad policy is still an open product decision, carried in the review as IN-09.

---

_Fixed: 2026-10-17T21:57:42Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 2_
