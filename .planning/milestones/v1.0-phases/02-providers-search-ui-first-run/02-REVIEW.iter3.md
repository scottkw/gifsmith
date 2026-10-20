---
phase: 02-providers-search-ui-first-run
reviewed: 2026-10-17T21:47:28Z
depth: standard
iteration: 2
files_reviewed: 56
files_reviewed_list:
  - .gitignore
  - frontend/src/app.css
  - frontend/src/App.svelte
  - frontend/src/components/Footer.svelte
  - frontend/src/components/ProviderTabs.svelte
  - frontend/src/components/ResultGrid.svelte
  - frontend/src/components/SetupCard.svelte
  - frontend/src/components/StateMessage.svelte
  - frontend/src/components/Tile.svelte
  - frontend/src/lib/api.ts
  - frontend/src/lib/errors.test.ts
  - frontend/src/lib/errors.ts
  - frontend/src/lib/grid.test.ts
  - frontend/src/lib/grid.ts
  - frontend/src/lib/providers.ts
  - frontend/src/main.ts
  - go.mod
  - internal/provider/fakeapi/fakeapi.go
  - internal/provider/finergifs/finergifs_test.go
  - internal/provider/finergifs/finergifs.go
  - internal/provider/giphy/giphy_test.go
  - internal/provider/giphy/giphy.go
  - internal/provider/httpx/httpx_test.go
  - internal/provider/httpx/httpx.go
  - internal/provider/klipy/klipy_test.go
  - internal/provider/klipy/klipy.go
  - internal/provider/provider_test.go
  - internal/provider/provider.go
  - internal/search/clips_test.go
  - internal/search/clips.go
  - internal/search/search_test.go
  - internal/search/search.go
  - internal/services/actions_test.go
  - internal/services/actions.go
  - internal/services/gifs.go
  - internal/services/panel.go
  - internal/services/settings_test.go
  - internal/services/settings.go
  - internal/settings/settings_test.go
  - internal/settings/settings.go
  - internal/shell/shell.go
  - internal/staging/staging_test.go
  - internal/staging/staging.go
  - main.go
  - scripts/smoke-browser-loop.sh
  - scripts/uat/01-tracer.js
  - scripts/uat/02-grid.js
  - scripts/uat/02-keyboard.js
  - scripts/uat/03-states.js
  - scripts/uat/05-setup.js
  - scripts/uat/05-tabs.js
  - scripts/uat/06-backstops.js
  - scripts/uat/06-clips-denied.js
  - scripts/uat/06-clips.js
  - wire_desktop.go
  - wire_server.go
findings:
  critical: 0
  warning: 3
  info: 9
  total: 12
status: issues_found
---

# Phase 2: Code Review Report (iteration 2)

**Reviewed:** 2026-10-17T21:47:28Z
**Depth:** standard
**Files Reviewed:** 56
**Status:** issues_found

## Summary

This is a re-review after fix commits 814e7e4, f8d86c5, e8be357, 08e0ff5, e3501b7, 3f076fa and 50824e1. I re-read the code for each finding and did not rely on the fix report. `go vet . ./internal/...` and `go vet -tags server .` are clean, and `go test -race -count=1 ./internal/...` passes in all 11 packages.

**Prior findings, checked against the code:**

| Prior | Verdict | Evidence |
|-------|---------|----------|
| CR-01 shared GIPHY/Clips 429 budget | **Fixed** | `search.go` `budget()` maps clips to giphy for the block read and write, `Blocked`, `Invalidate` and `gen`. `main.go:59-61` routes the probe through `svc.Guard(provider.Clips, ...)`, so a probe 429 blocks GIPHY. `Invalidate(GIPHY)` drops `clips|` cache entries through `strings.Cut`. Re-probes are capped by `rechecked`. The cap causes the new WR-02 and WR-03 below. |
| CR-02 blank Clips area on a rejected key | **Partly fixed** | The blank area is gone: `App.svelte:464-472` renders a message, and `keySaved('giphy')` resets Clips (line 255). However, the message's "Go to GIPHY" button often lands on a GIPHY grid that has no key field. See WR-01. |
| WR-01 in-flight write after Invalidate | **Fixed** | `get()` and `Guard()` capture `gen[budget(p)]` before the call. The cache write and the `noteRateLimit` block are both gated on the same generation. Ordering against `SaveKey` (`Keys.Set`, then `Invalidate`, then `Seed`) is sound: any call that captured the old generation is dropped. |
| WR-02 persisted Clips tab unreachable | **Fixed** | `App.svelte:173` restores `clips` whenever a GIPHY key is configured, and `bootClips()` starts the probe through a GIPHY trending fetch. I traced the paths for enabled, error, still unknown and switching away. One remaining edge case is listed in WR-02 and IN-04. |
| WR-03 redirects bypass the allowlist | **Fixed** | `staging.go:67-77` copies the client for each call and sets `CheckRedirect` to run `allow(g.Provider, r.URL)` on every hop, with a limit of 5 hops. `errors.Is` sees `ErrNotAllowed` through the `*url.Error` wrapper. The body of the refused response is closed by `net/http`. |
| WR-04 Klipy ad filtering and empty-page fetch storm | **Fixed** (product call still open) | Items with GIF renditions are now kept in place, and `Next` is set only when at least one item was mapped (`klipy.go:123`). The open product questions are carried as IN-09. |

No new Critical issues. There are three Warnings, all in the Clips recovery paths that the fixes changed. The Info items carry forward IN-01 to IN-07 and add two product-level items: the Try Again no-op (IN-08) and unlabeled Klipy ads (IN-09).

## Warnings

### WR-01: "Go to GIPHY" from the Clips rejected-key message usually lands on a GIPHY grid with no key field

**File:** `frontend/src/App.svelte:464-472`, `frontend/src/App.svelte:98-102`, `frontend/src/App.svelte:232-237`
**Issue:** The CR-02 message tells the user "Paste a new key on the GIPHY tab to continue". The button calls `switchTo('giphy')`, then `show('giphy')`. The GIPHY tab shows a key field only when GIPHY's own state has `err.kind === 'key_rejected'` or no key is configured (`card`, lines 98-102). In the most common case, the GIPHY tab shows no key field:
- Clips trending fails with 401 after the key was revoked. `query` is `''`.
- `states.giphy` already holds trending results from earlier, with `status: 'ready'` and `q: ''`. On a `bootClips` launch it always does.
- `show('giphy')` returns early at line 235 because `s.status !== 'idle' && s.q === wanted()`. Even if it did reload, `search.Service` would serve GIPHY trending from its 10-minute cache.
- Result: the GIPHY grid renders with no setup card. The user has nowhere to paste the key the message asked for. Recovery requires typing a fresh GIPHY search that goes to the network and gets a 401.

The UAT step in `06-clips.js:51-58` passes only because the fixture's `"revoked"` keyword stays in `query` across the switch. That makes GIPHY run a new search, which the fake API also rejects. A real revoked key with an empty query does not take this path.
**Fix:** Have the button carry the rejection over to GIPHY's state, so the setup card renders at once:
```svelte
onbutton={() => {
  const g = states.giphy;
  setResults('giphy', null, null);
  Object.assign(g, { status: 'error', err: { kind: 'key_rejected', provider: 'giphy' }, errAt: Date.now() });
  switchTo('giphy');
}}
```
Then add a UAT step that triggers the rejection with an empty query, for example a fixture flag that rejects Clips trending, and asserts that `.card input` appears.

### WR-02: A Clips re-probe before the launch's first GIPHY success leaves Clips unavailable for the rest of the launch

**File:** `internal/search/clips.go:103-120`, `frontend/src/App.svelte:181-189`, `frontend/src/App.svelte:245`, `frontend/src/App.svelte:304-307`, `internal/provider/giphy/giphy.go:95-97`
**Issue:** `start(true)` (Recheck) sets both `probed = true` and `rechecked = true`. After that:
- `Trigger()`, which runs on the first GIPHY GIFs success, is a no-op, because `!recheck && c.probed`.
- `Recheck()` is a no-op, because `c.rechecked` is set.

Suppose the Recheck probe ends `unknown` because of a timeout, an offline network or a 5xx. Clips then stays `unknown` until relaunch or a GIPHY key change, even after GIPHY GIFs succeed again. For a menu-bar app that runs for days, that can be a long time. The iteration-1 fix introduced this: before the `rechecked` cap, a later Recheck could recover.

The new boot path makes this easy to reach. Gifsmith launches at login before Wi-Fi is up. `bootClips` gets a GIPHY timeout and copies it to the Clips tab. The user presses the only visible button, Try Again. `retry()` calls `recheckClips()` (line 305) instead of retrying GIPHY. That probes Clips while still offline and spends both the re-probe and the Trigger. Pressing ⌘2 before any GIPHY success has the same effect (line 245).

That same pre-success probe also breaks `ProbeClips`'s precondition: "callers probe only after a GIFs call succeeded with the same key" (giphy.go:95-97). If the key was revoked, the probe's 401 is classified as `disabled`. The user then sees "Clips aren't enabled for your key ... Email clips@giphy.com" instead of a rejected-key message, and that state persists for the launch.
**Fix:** Track the GIFs-success probe separately from the re-probe, so a Recheck cannot consume it. On the Clips tab, retry GIPHY while GIPHY is the reason Clips has not been probed:
```go
// clips.go
triggered bool // the GIFs-success probe ran since the last Reset
// start():
if (recheck && (!unknown || c.rechecked)) || (!recheck && (c.triggered || c.state == ClipsChecking)) { ... }
c.triggered = c.triggered || !recheck
c.rechecked = c.rechecked || recheck
```
```ts
// App.svelte retry()
if (active === 'clips' && !clipsEnabled && states.giphy.status === 'error') bootClips();
else if (active === 'clips' && !clipsEnabled) recheckClips();
else runFirst(active, cur.q);
```
Also consider classifying a pre-success 401 as `unknown` rather than `disabled`.

### WR-03: After the re-probe is spent, Try Again replays a stale rate-limit error and restarts its countdown

**File:** `internal/search/clips.go:73-83`, `frontend/src/App.svelte:146-149`, `frontend/src/lib/errors.ts:35-44`
**Issue:** If the one re-probe itself returns 429, `finish` stores it in `lastErr` with the vendor's `RetryAfterSec`, for example 3600. When the window ends, `errorState` shows Try Again. Pressing it calls `Recheck`. The fast path no longer applies because the block has expired, `start(true)` refuses because `rechecked` is set, and `lastErr` is replayed. `recheckClips` then sets `errAt = Date.now()`, so the panel again says "Try again in 60 minutes (at HH:MM)". That promise can never be kept: no probe will run until relaunch.

Before the fix, Try Again re-probed. Now it shows a false countdown, which is worse than a plain no-op. A 429 also says nothing about whether the key is approved for Clips, so spending the one re-probe on it gives up the capability check for no information.

Separately, the fast path at `clips.go:70-71` returns `RetryAfterSec: 0` even when `search.Service` knows the block's end, so that error shows no countdown at all.
**Fix:** Do not count a rate-limited re-probe against the cap. Reset `rechecked` in `finish` when the error is `RateLimited`, because the block already enforces the budget. Alternatively, strip `RetryAfterSec` from a replayed error and return a distinct kind, so the UI can say "Reopen Gifsmith to check Clips again". For the fast path, return `svc`'s `blockedErr`, through a `BlockedErr func() error` on `Clips`, instead of building a zero-retry error.

## Info

### IN-01: Raising the Wails log level to Debug would log every bound call's arguments, including the API key (carried forward)

**File:** `main.go:75-88`
**Issue:** Wails logs bound-call arguments at Debug level, and `SettingsService.SaveKey(p, key)` takes the key as an argument. This is safe today because `LogLevel` is unset, but nothing prevents someone from setting it.
**Fix:** Add a comment that forbids Debug level, or set a `Logger` that filters out the `args` attribute.

### IN-02: Retry-After is not capped (carried forward)

**File:** `internal/provider/httpx/httpx.go:73-82`, `internal/search/search.go:128`
**Issue:** A huge Retry-After blocks the budget for years in this process. A value above roughly 9.2e9 overflows `time.Duration` and falls back to 60 s. Because of CR-01, a huge value on a Clips probe now also blocks GIPHY GIFs.
**Fix:** `return min(max(n, 0), 3600)` in `retryAfter`, and apply the same clamp to the HTTP-date branch.

### IN-03: Staged files persist across launches, against Klipy's no-retain rule (carried forward, deferred to Phase 3 as ACTN-10)

**File:** `main.go:45`, `internal/staging/staging.go:124-142`
**Issue:** Staged files are swept only after a later successful copy, so a Klipy GIF copied just before quitting stays on disk.
**Fix:** Sweep at startup and in `ApplicationShouldTerminate`.

### IN-04: A stale Clips error stays on screen after Clips becomes enabled (carried forward; the WR-02 fix adds a path)

**File:** `frontend/src/App.svelte:131`, `frontend/src/App.svelte:187`, `frontend/src/App.svelte:232-237`
**Issue:** `show('clips')` returns early when `status === 'error'` and `q === wanted()`. There is a new way to hit this. `bootClips` copies a GIPHY error, such as a timeout, onto the Clips tab. The user then fixes it on the GIPHY tab: a successful GIPHY retry triggers the probe, and the probe enables Clips. Pressing ⌘2 then calls `show('clips')`, and the copied "GIPHY timed out" error stays until the user presses Try Again. A `recheckClips` call that was in flight for the old key and lands after `keySaved` reset Clips leaves the same stale error.
**Fix:** In `switchTo` and `refreshProviders`, when `clipsEnabled` is true and `states.clips.status === 'error'`, call `load('clips')` instead of `show('clips')`.

### IN-05: SetupCard's error path dereferences `field` after the card may have unmounted (carried forward)

**File:** `frontend/src/components/SetupCard.svelte:59-60`
**Issue:** `field.focus()` and `field.select()` throw if the user switches tabs while `saveKey` is pending and the save then fails. Line 40 already uses `field?.`.
**Fix:** `field?.focus(); field?.select();`

### IN-06: A failed initial Providers() call leaves the panel blank for good (carried forward)

**File:** `frontend/src/App.svelte:128-129`, `frontend/src/App.svelte:170`
**Issue:** If `Promise.all([providers(), lastProvider()])` rejects, `statuses` stays `null` and the content area never renders. `refreshProviders` also lets a rejection escape from an event handler.
**Fix:** Wrap the boot in try/catch with a retryable message, and catch errors inside `refreshProviders`.

### IN-07: The hide-after-copy timer is never cancelled (carried forward)

**File:** `frontend/src/App.svelte:389`
**Issue:** If the panel is re-shown within 400 ms of a copy, `setTimeout(hide, HIDE_AFTER_COPY_MS)` hides the freshly shown panel.
**Fix:** Store the timer handle and clear it in the `panel:show` handler.

### IN-08: Clips Try Again does nothing after the one re-probe (product choice, already surfaced)

**File:** `internal/search/clips.go:73-83`, `frontend/src/App.svelte:304-307`
**Issue:** This is the behavior chosen in the CR-01 fix: once the re-probe is spent, Try Again flashes the skeleton and replays the same error, and no request is made until relaunch or a GIPHY key change. It is recorded here as a known behavior. WR-02 and WR-03 above cover the cases where this cap gives wrong results, not just a no-op.
**Fix:** If the button stays, change its copy for the spent state, for example "Reopen Gifsmith to check Clips again", or hide it.

### IN-09: Klipy ads are shown without a label, and some of them cannot be copied (product choice, already surfaced)

**File:** `internal/provider/klipy/klipy.go:104-123`
**Issue:** Klipy ad items that have GIF renditions now render as ordinary tiles with no "Sponsored" label. Their previews load from whatever host the ad names, not only `klipy.com`. Copying an ad hosted outside `klipy.com` fails the staging allowlist and shows the generic "The download failed" message. A mid-stream page that contains only HTML ads now ends paging even though `has_next` is true, so any results after it cannot be reached. The ad policy, along with the copy-file clearance, is still pending with developers@klipy.com.
**Fix:** Resolve the policy with Klipy before release. Until then, consider an `IsAd` flag on `provider.Gif` so the tile can be labeled and copy can show an ad-specific message.

---

_Reviewed: 2026-10-17T21:47:28Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
