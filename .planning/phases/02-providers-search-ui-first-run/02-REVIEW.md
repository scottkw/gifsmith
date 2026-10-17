---
phase: 02-providers-search-ui-first-run
reviewed: 2026-10-17T22:02:22Z
depth: standard
iteration: 3
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
  warning: 0
  info: 11
  total: 11
status: issues_found
---

# Phase 2: Code Review Report (iteration 3)

**Reviewed:** 2026-10-17T22:02:22Z
**Depth:** standard
**Files Reviewed:** 56
**Status:** issues_found (Info only)

## Summary

This re-review covers fix commits bc4bc08, fc0906e and e360050. I re-read `internal/search/clips.go`, `internal/search/search.go`, `internal/services/gifs.go`, `internal/services/settings.go`, `main.go`, `frontend/src/App.svelte`, `frontend/src/lib/api.ts`, `frontend/src/lib/errors.ts`, `frontend/src/components/ProviderTabs.svelte` and the updated tests and UAT. I traced the Clips state machine (`triggered`, `proven`, `rechecked`, `gen`) against the frontend's `bootClips`, `recheckClips`, `refreshProviders`, `switchTo` and `retry` paths. I did not rely on the fix report.

Gates, run read-only: `go vet . ./internal/...` and `go vet -tags server .` are clean (only linker version warnings). `go test -race -count=1 ./internal/...` passes in all 11 packages that have tests. `pnpm --dir frontend run check` reports 0 errors and 0 warnings. `pnpm --dir frontend run test` passes 23 of 23.

**Prior findings, checked against the code:**

| Prior | Verdict | Evidence |
|-------|---------|----------|
| WR-01: "Go to GIPHY" lands on a GIPHY grid with no key field | **Fixed** | `App.svelte:479-484` clears GIPHY's results and sets `err: key_rejected` before `switchTo('giphy')`. `show('giphy')` then returns at line 242, and `card` (lines 98-102) renders. The UAT step in `06-clips.js:54-60` clears the query, so GIPHY keeps its cached `ready` / `q: ''` state. That reproduces the real case. The step is deterministic: `show('clips')` returns at line 242 on `key_rejected`, so clearing the query starts no Clips reload that could remove the message before the click. |
| WR-02: a pre-success re-probe leaves Clips unavailable | **Fixed** | `start` sets `rechecked` only when `proven` (clips.go:122). `Trigger` is refused only when `triggered` is set or a probe is running (line 117), and it always sets `proven` (lines 57-59). `classify` returns `disabled` only when `proven` (line 155). On the frontend, `bootClips` (App.svelte:182-197) runs before every Clips probe the user starts. It calls `recheckClips` only after `trending('giphy')` succeeds, and `GifService.call` runs `Trigger` (and so sets `proven`) on that success before it returns (gifs.go:72-74). In the offline-at-login case, Try Again now retries GIPHY first and lets the Trigger probe run. |
| WR-03: Try Again replays a stale rate-limit countdown | **Fixed** | `finish` clears `rechecked` on `RateLimited` (clips.go:138-141), so a 429 re-probe is not counted. Once the block ends, a Recheck probes again rather than replaying `lastErr`. The fast path returns `svc.BlockedErr(provider.Clips)` (main.go:63, clips.go:75-79), and `blockedErr` returns the real seconds left when known. Retries cannot loop: an auto-recheck from `refreshProviders` is refused while `rechecking` is set, and the 429 block keeps probes at least 60 s apart. |

**Regression check on fc0906e and e360050:**
- `bootClips` uses an uncancellable `trending('giphy', '')` outside `latest()`. A late answer touches only `states.clips`, and only while Clips is still active. GIPHY's own tab state is not changed. That is correct.
- Delivery order of the `providers:changed` event (from `Trigger` → `start`) relative to the `trending` response does not matter. `Recheck` reads Go's state, gets `checking`, and returns `(checking, nil)`, so the skeleton stays until the probe finishes.
- `Trigger` refused during a Recheck leaves `triggered` false. The next GIPHY success can then run one more probe. This uses more probes but stays bounded.
- No new data races. `HasKey` and `Blocked` run without `c.mu` held. `Blocked` takes `svc.mu` alone, so the two locks are never held together.

There are no Critical or Warning findings. Info items IN-01 to IN-09 are still valid and are carried forward with updated line numbers. IN-04 gains one more path. IN-10 and IN-11 are new, low-impact items.

## Info

### IN-01: Raising the Wails log level to Debug would log every bound call's arguments, including the API key (carried forward)

**File:** `main.go:75-88`
**Issue:** Wails logs bound-call arguments at Debug level, and `SettingsService.SaveKey(p, key)` takes the key as an argument. This is safe while `LogLevel` is unset, but nothing guards it.
**Fix:** Add a comment that forbids Debug level, or set a `Logger` that drops the `args` attribute.

### IN-02: Retry-After is not capped (carried forward)

**File:** `internal/provider/httpx/httpx.go:73-81`, `internal/search/search.go:127`
**Issue:** A huge Retry-After blocks the shared GIPHY/Clips budget for the life of the process. A value above about 9.2e9 seconds overflows `time.Duration`, and the block falls back to 60 s. Since e360050 the fast path shows this value to the user as a countdown (`BlockedErr`), so a hostile or buggy value now shows up directly in the UI copy.
**Fix:** `return min(max(n, 0), 3600)` in `retryAfter`, with the same clamp on the HTTP-date branch.

### IN-03: Staged files persist across launches, against Klipy's no-retain rule (carried forward, deferred to Phase 3 as ACTN-10)

**File:** `main.go:45`, `internal/staging/staging.go:124-142`
**Issue:** `Sweep` runs only after a later successful copy (`actions.go:56`). A Klipy GIF copied just before quitting stays on disk.
**Fix:** Sweep at startup and in `ApplicationShouldTerminate`.

### IN-04: A stale Clips error stays on screen after Clips becomes enabled (carried forward, one more path)

**File:** `frontend/src/App.svelte:131`, `frontend/src/App.svelte:243`, `frontend/src/App.svelte:254`, `frontend/src/App.svelte:145-149`
**Issue:**
- **Still reproducible:** launch on Clips while offline. `bootClips` copies a GIPHY network or timeout error onto Clips (`q` stays `''`). The user fixes it on the GIPHY tab: GIPHY loads, `Trigger` probes, and Clips becomes `enabled`. Pressing ⌘2 calls `show('clips')`, which returns at line 243 (`status === 'error'`, `q === wanted()`). The old "Can't reach GIPHY" message stays until the user presses Try Again.
- **New path, made worse by the WR-01 fix:** a `recheckClips` call in flight for the old key can return a `key_rejected` error after `keySaved('giphy')` has cleared Clips. `show('clips')` then returns at line 242 for good. "Go to GIPHY" (lines 479-484) then marks the newly saved, valid GIPHY key as rejected and shows the setup card again. This needs a narrow timing window: the user switches tabs and saves a key while a Clips probe is still running.
**Fix:**
- In `switchTo` and `refreshProviders`: when `clipsEnabled` is true and `states.clips.status === 'error'`, call `setResults('clips', null, null)` and then `load('clips')`, not `show('clips')`.
- In `recheckClips`: record `states.clips.seq` (or a key generation) before the await. Drop the error if the value changed in the meantime.

### IN-05: SetupCard's error path dereferences `field` after the card may have unmounted (carried forward)

**File:** `frontend/src/components/SetupCard.svelte:59-60`
**Issue:** `field.focus()` and `field.select()` throw if the user switches tabs while `saveKey` is pending and the save then fails. Line 40 already uses `field?.`.
**Fix:** `field?.focus(); field?.select();`

### IN-06: A failed initial Providers() call leaves the panel blank for good (carried forward)

**File:** `frontend/src/App.svelte:128-129`, `frontend/src/App.svelte:170`
**Issue:** If `Promise.all([providers(), lastProvider()])` rejects, `statuses` stays `null` and the content area never renders. `refreshProviders` also lets a rejection escape from the event handler.
**Fix:** Wrap the boot in try/catch with a retryable message, and catch inside `refreshProviders`.

### IN-07: The hide-after-copy timer is never cancelled (carried forward)

**File:** `frontend/src/App.svelte:397`
**Issue:** If the panel is shown again within 400 ms of a copy, `setTimeout(hide, HIDE_AFTER_COPY_MS)` hides the newly shown panel.
**Fix:** Store the timer handle and clear it in the `panel:show` handler.

### IN-08: Clips Try Again does nothing after the one re-probe (product choice, already surfaced)

**File:** `internal/search/clips.go:80-90`, `frontend/src/App.svelte:312-315`
**Issue:**
- Once a non-429 re-probe is spent (for example, a timeout), each Try Again or ⌘2 runs `bootClips`. That makes a GIPHY trending call, which is a real request once the 10-minute cache expires. It then shows the skeleton and replays the same `lastErr`. No probe runs until relaunch or a GIPHY key change.
- After e360050 this applies only to non-429 errors, so no stale countdown is shown.
- When the Trigger probe times out while Clips is on screen, `refreshProviders` (line 132) spends the re-probe automatically. The user has not chosen to retry at that point.
**Fix:** If the button stays, change its copy for the spent state (for example "Reopen Gifsmith to check Clips again"), or hide it.

### IN-09: Klipy ads are shown without a label, and some of them cannot be copied (product choice, already surfaced)

**File:** `internal/provider/klipy/klipy.go:98-127`
**Issue:**
- Ad items with GIF renditions render as ordinary tiles, with no "Sponsored" label.
- Their previews load from whatever host the ad names.
- Copying an ad hosted outside `klipy.com` fails the staging allowlist and shows the generic "The download failed" message.
- A page that contains only HTML ads ends paging, even when `has_next` is true.
- The ad policy is still pending with developers@klipy.com.
**Fix:** Resolve the policy with Klipy before release. Consider an `IsAd` flag on `provider.Gif`.

### IN-10: "Not approved for Clips" can still be inferred from a GIFs success that no longer proves anything (new)

**File:** `internal/services/gifs.go:72-74`, `internal/search/clips.go:56-59`, `internal/search/clips.go:155`, `frontend/src/App.svelte:132`
**Issue:** `proven` stays set until `Reset`, and any GIPHY success sets it, including a cached page from `search.Service`:
- **Revocation within the cache window:** the key is revoked while the app runs, and Clips is `unknown` (an earlier probe timed out). Try Again runs `bootClips`. GIPHY trending is served from the 10-minute cache, `Trigger` keeps `proven = true`, and the re-probe's 401 is classified `disabled`. The user sees "Your GIPHY key works for GIFs, but GIPHY has to approve it for Clips" for the rest of the launch, when the real fix is a new key. The "Back to GIPHY" button shows the cached grid with no key field.
- **A probe that still runs without any GIFs success:** `refreshProviders` (line 132) calls `recheckClips` directly, without `bootClips`. It is reached through a `providers:changed` event, for example from `keySaved('klipy')` after the user saved a Klipy key and then pressed ⌘2 before GIPHY trending returned. In that case the probe can run before any GIFs success. A valid key that is not approved for Clips then gets "GIPHY rejected your saved key", and "Go to GIPHY" now (WR-01) forces the GIPHY setup card for a working key.

Both cases are unlikely. Before fc0906e every 401 was classified `disabled`, so this is narrower than it was, not a regression.
**Fix:**
- Call `Trigger` only on a network success. For example, have `search.Service.get` report whether it served from cache, or clear `proven` when a GIPHY call returns `KeyRejected`.
- Route the line-132 recheck through `bootClips()` so every probe the UI starts follows a GIPHY success.

### IN-11: `search.Service.Blocked` is now used only by tests (new)

**File:** `internal/search/search.go:186-189`
**Issue:** After e360050, production code calls only `BlockedErr` (main.go:63). `Blocked` is an exported wrapper used only by `search_test.go`, `clips_test.go` and `settings_test.go`.
**Fix:** Remove it and have the tests use `BlockedErr(p) != nil`, or keep it and note that it is a test helper.

---

_Reviewed: 2026-10-17T22:02:22Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
