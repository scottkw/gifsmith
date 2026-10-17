---
phase: 02-providers-search-ui-first-run
reviewed: 2026-10-17T21:30:34Z
depth: standard
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
  critical: 2
  warning: 4
  info: 7
  total: 13
status: issues_found
---

# Phase 2: Code Review Report

**Reviewed:** 2026-10-17T21:30:34Z
**Depth:** standard
**Files Reviewed:** 56
**Status:** issues_found

## Summary

I reviewed the Phase 2 diff (`c4f8403..HEAD`): provider clients (GIPHY, Clips, Klipy, Finer Gifs), the `httpx` choke point, `search.Service` (cache, 429 block, dedupe), the Clips probe state machine, the bound services (Gif, Settings, Action, Panel), staging, settings/Keychain, wiring, and the Svelte panel. `go vet ./...` and `go vet -tags server ./...` are clean, and `go test -race ./internal/...` passes.

Key handling holds up. `httpx` never formats a `*url.Error`. `provider.Error` carries no URL, query, or body. `SaveKey` logs only the provider and the error kind. Keychain errors are fixed text. No bound method returns a key. The Klipy key is validated and then `PathEscape`d. Klipy pages skip dedupe. The download allowlist is applied before the request. I found no XSS sinks: there is no `{@html}` or `innerHTML`.

There are two blocking defects:

1. The 429 budget is tracked per provider ID, but GIPHY GIFs and GIPHY Clips use the same key. The Clips probe also bypasses the block entirely. Together, these let the app send repeated requests into a known 429, which breaks the service's own "never retries into a 429" rule.
2. A revoked GIPHY key seen on the Clips tab leaves the content area blank, with no message and no way to recover.

The warnings cover a stale in-flight request that can undo `Invalidate`, a persisted Clips tab that can never be restored, staging redirects that skip the allowlist, and Klipy ad filtering.

## Critical Issues

### CR-01: GIPHY and Clips share one key and one rate limit, but the 429 block is per provider ID, and the Clips probe bypasses it

**File:** `internal/search/search.go:93,113`, `internal/search/clips.go:66-80,94`, `internal/provider/giphy/giphy.go:98-107`, `main.go:59`, `frontend/src/App.svelte:132,232`
**Issue:** `Service.blocked` is keyed by `provider.ID`. GIPHY GIFs (`giphy`) and GIPHY Clips (`clips`) are separate keys in that map, but they send the same `api_key` and draw from the same 100 calls/hour budget. This causes three failures:
- After a GIPHY 429 sets `blocked["giphy"]`, switching to the Clips tab calls `get(provider.Clips, ...)`. `blocked["clips"]` is not set, so the request goes out straight into the 429. A Clips 429 likewise does not block GIPHY GIFs.
- `ProbeClips` calls `httpx.GetJSON` directly rather than going through `search.Service`. A 429 on the probe therefore never sets a block. `classify` maps it to `unknown`. `Recheck` checks only `svc.Blocked(GIPHY)`, which is still clear, so it probes again.
- The frontend calls `recheckClips()` on every `switchTo('clips')` while the state is `unknown` (`App.svelte:232`), and also from `refreshProviders` (`App.svelte:132`). Every ⌘2 press re-probes, so this is not the "re-probes once" behavior specified in RESEARCH. While the key is rate-limited, each tab switch spends another request into the 429.

`SettingsService.SaveKey` also calls `Svc.Invalidate(provider.GIPHY)` only. Cached Clips pages and any `clips` block from the old key survive a GIPHY key change.
**Fix:** Collapse GIPHY and Clips into one budget, route the probe's 429 into that budget, and limit re-probes to one per launch:
```go
// search.go
func budget(p provider.ID) provider.ID {
	if p == provider.Clips {
		return provider.GIPHY // same api_key, same 100/h budget
	}
	return p
}
// in get(): s.blocked[budget(p)] for both the read and the write.
// Invalidate(GIPHY) should also drop "clips|" cache entries.

// Exported so the probe can record a 429:
func (s *Service) NoteRateLimit(p provider.ID, retryAfterSec int) { /* same block logic as get() */ }

// main.go
Probe: func(ctx context.Context) error {
	err := clipsClient.ProbeClips(ctx)
	var pe *provider.Error
	if errors.As(err, &pe) && pe.Kind == provider.RateLimited {
		svc.NoteRateLimit(provider.GIPHY, pe.RetryAfterSec)
	}
	return err
},
```
In `clips.go`, add a `rechecked bool` that `Reset` clears and that `start(true)` sets, so `Recheck` probes at most once per key per launch. Add a test that a GIPHY 429 blocks `Search(ctx, provider.Clips, ...)`, and a test that a probe 429 makes the next `Recheck` return `rate_limited` without calling `Probe`.

### CR-02: A revoked GIPHY key on the Clips tab renders a blank content area with no recovery

**File:** `frontend/src/App.svelte:98-102,221,442-508`, `frontend/src/lib/errors.ts:32-34`
**Issue:** The Clips slot has no `keyProvider`, so `card` is always false there. When a Clips search fails with `key_rejected`, for example because the GIPHY key was revoked after the probe passed, `runFirst` sets `status = 'error'` and `err.kind = 'key_rejected'`. Walking the render chain:
- `!giphyKey` is false because the key is still configured.
- `clipsState === 'disabled'` is false.
- `!clipsEnabled` is false.
- `finerIdle` is false and `items` is empty.
- `status` is neither `loading` nor `empty`.
- `errView` is `null`, because `errorState()` returns `null` for `key_rejected` and `key_missing`.

No branch matches, so the content area is empty: no message, no Try Again. `show('clips')` also returns early on `key_rejected` (line 221), so switching away and back cannot recover. Only a relaunch or saving a new GIPHY key on the GIPHY tab clears it. Even then, `keySaved('giphy')` never resets `states.clips`, so the `key_rejected` error and the early return in `show()` remain until the user relaunches.
**Fix:** Handle the Clips slot's key errors explicitly and reset Clips state when the GIPHY key changes:
```svelte
{:else if active === 'clips' && (cur.err?.kind === 'key_rejected' || cur.err?.kind === 'key_missing')}
  <StateMessage
    heading="GIPHY rejected your saved key"
    body="Clips use your GIPHY key. Paste a new key on the GIPHY tab to continue."
    button="Go to GIPHY" primary
    onbutton={() => switchTo('giphy')}
  />
```
In `keySaved(p)`, when `p === 'giphy'`, also call `setResults('clips', null, null)` so the next Clips view loads fresh.

## Warnings

### WR-01: An in-flight request with the old key can land after `Invalidate` and restore the old cache entry or 429 block

**File:** `internal/search/search.go:105-121`, `internal/services/settings.go:71-76`
**Issue:** `get()` reads the key inside `call(pr)` while holding no lock, and writes the cache or block when the call returns. If `SaveKey` runs `Keys.Set`, `Invalidate`, and `Seed` while a GIPHY or Klipy request using the old key is in flight, that request completes afterwards:
- A 429 for the old key re-blocks the provider for at least `MinBlock` (60 s) under the new key. This is the likely case, because users replace a key precisely when it hits its limit.
- A success overwrites the seeded `trending||` entry with the old key's page.

`SaveKey` serializes against other saves only, not against searches.
**Fix:** Give each provider a generation counter. Capture it before the call and drop the write if it has changed:
```go
type Service struct { /* ... */ gen map[provider.ID]int }
// get():
s.mu.Lock(); g := s.gen[p]; s.mu.Unlock()
pg, err := call(pr)
s.mu.Lock(); stale := s.gen[p] != g; s.mu.Unlock()
if stale { return pg, err } // never cache or block on behalf of a replaced key
// Invalidate(): s.gen[p]++
```

### WR-02: The persisted Clips tab can never be restored

**File:** `frontend/src/App.svelte:173`, `internal/services/gifs.go:72-74`, `internal/search/clips.go:45-47`
**Issue:** At startup, `onMount` restores `last === 'clips'` only when `st.clips?.clips === 'enabled'`. The Clips probe runs only after the first GIPHY GIFs success in this process (`gifs.go:72-74`), and that success is triggered by this same `onMount`'s `load(active)`. When `Providers()` answers at mount, the Clips state is always `""`, reported as `unknown`. The `last === 'clips'` branch is therefore unreachable: a user who quits on the Clips tab always reopens on GIPHY, although `switchTo` persists `clips` (line 231). The UAT note "last tab survives Quit and relaunch" only exercises the non-Clips slots.
**Fix:** If `last === 'clips'` and the state is `unknown` or `checking`, open on Clips (skeleton) and fire one GIPHY trending fetch in the background to start the probe. Or explicitly persist `giphy` when the user picks Clips, so the stored value does not promise something the boot path cannot honor.

### WR-03: Staging downloads follow redirects without re-checking the host allowlist

**File:** `main.go:44`, `internal/staging/staging.go:52-70`
**Issue:** `VendorAllow` is checked against `g.DownloadURL` only. The staging client is `&http.Client{}` with the default `CheckRedirect`, which follows up to 10 redirects to any host and any scheme, including a downgrade to `http://`. The `Gif` comes from the webview (`ActionService.Copy(g)`), and the allowlist exists to bound which URLs Go fetches. A vendor-host URL that redirects (an open redirect on a CDN or vendor domain, or a tampered `downloadUrl`) bypasses it. The magic-byte check limits what gets staged, not what gets requested.
**Fix:** Re-apply the allowlist on every hop:
```go
stage := &staging.Store{Dir: d.StagingDir, Allow: d.Allow, HTTP: &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 || !d.Allow(gifProviderFrom(via[0]), req.URL) {
			return staging.ErrNotAllowed
		}
		return nil
	},
}}
```
Alternatively, have `Store.Fetch` build a per-call client that captures `g.Provider`.

### WR-04: The Klipy mapper drops ad items despite the no-filter rule, and a page of only ads keeps the grid fetching

**File:** `internal/provider/klipy/klipy.go:99-103,121-127`, `frontend/src/components/ResultGrid.svelte:64-66`
**Issue:**
- `if d.Type != "gif" { skipped++; continue }` removes items from a Klipy page. The project constraint (CLAUDE.md, Klipy integration rules) says "do not reorder, insert, remove, or filter results". Logging the skip does not make it compliant. This needs the vendor's sign-off or a rendered ad tile.
- When every item on a page is skipped and `has_next` is true, the client returns `Items: []` with `Next` set. The GIPHY and Finer clients guard this case with `len(items) > 0`; Klipy does not. In the grid, appending zero items changes `paging` from `loading` back to `idle`. That changes `count`, which triggers `$virt.setOptions`, and the sentinel effect fires again while the last row is still mounted. The result is back-to-back page requests against a 100/hour key until a page contains a GIF.
**Fix:** Confirm the ad policy with developers@klipy.com, as already planned for the copy-file question. In the meantime, also guard the cursor:
```go
if r.Data.HasNext && len(items) > 0 { /* set p.Next */ }
```
Alternatively, have `loadMore` stop when a page adds no items: `if (!page.items?.length) { s.paging = 'end'; return; }`.

## Info

### IN-01: Raising the Wails log level to Debug would log every bound call's arguments, including the API key

**File:** `main.go:71-84` (via Wails `pkg/application/messageprocessor_call.go:126-131`)
**Issue:** Wails logs `"Binding call complete:" ... "args", string(jsonArgs)` at Debug for every bound call, so `SettingsService.SaveKey(p, key)` would be logged with the key. This is safe today: `LogLevel` is unset (Info), and the production logger discards output. But a single `LogLevel: slog.LevelDebug` added while debugging would leak keys to stderr and Console.
**Fix:** Add a comment next to `application.Options` forbidding Debug level, or set an explicit `Logger` that filters the `args` attribute. Optionally, add a test asserting that `Options.LogLevel` is above Debug.

### IN-02: Retry-After is not capped

**File:** `internal/provider/httpx/httpx.go:73-82`, `internal/search/search.go:111`
**Issue:** A Retry-After of `1000000000` blocks the provider for about 31 years in this process. The only recovery is changing the key, because `Invalidate` clears the block, or relaunching. Values above about 9.2e9 overflow `time.Duration` to a negative value and fall back to 60 s, which is inconsistent.
**Fix:** Clamp in `retryAfter`, for example `return min(max(n, 0), 3600)`, since both vendors' windows are hourly.

### IN-03: Staged files persist across launches (Klipy no-retain rule)

**File:** `main.go:44`, `internal/staging/staging.go:114-132`
**Issue:** Sweep runs only after a later successful copy. A Klipy GIF copied just before quitting stays in `~/Library/Caches/com.kenscott.gifsmith/staging` indefinitely. The project constraint asks for deletion "on next copy and at quit". This is deferred to Phase 3 (ACTN-10); it is recorded here so it is not lost.
**Fix:** In Phase 3, sweep at startup (for example `stage.Sweep("")` with a variant that keeps nothing) and on `ApplicationShouldTerminate`.

### IN-04: A stale Clips probe error stays on screen after Clips becomes enabled

**File:** `frontend/src/App.svelte:131,222`
**Issue:** A failed `recheckClips` sets `states.clips.status = 'error'` but leaves `q` at `''`. If Clips later becomes `enabled` (for example after a GIPHY key save), `refreshProviders` and `switchTo` call `show('clips')`. That returns early because `status !== 'idle' && s.q === wanted()`, so the old timeout or rate-limit error remains until the user clicks Try Again.
**Fix:** In `refreshProviders`, when Clips flips to enabled, call `load('clips')` instead of `show('clips')`, or reset `states.clips` first.

### IN-05: SetupCard's error path dereferences `field` after the card may have unmounted

**File:** `frontend/src/components/SetupCard.svelte:55-60`
**Issue:** If the user switches tabs (⌘1–⌘4) while `saveKey` is pending, the keyed `SetupCard` is destroyed and `bind:this` resets `field` to null. When the save then fails, `field.focus()` throws an unhandled TypeError. On success, `onsaved(p)` still runs `keySaved`, which is mostly harmless.
**Fix:** Use `field?.focus(); field?.select();`.

### IN-06: A failed initial Providers() call leaves a permanently blank panel

**File:** `frontend/src/App.svelte:128-129,170`
**Issue:** If `Promise.all([providers(), lastProvider()])` rejects (a transport error, or `lastProvider` failing), `statuses` stays `null` and the content area renders nothing forever. `refreshProviders` also lets a rejection escape from an event handler.
**Fix:** Wrap the boot in try/catch, fall back to `statuses = {}` with a retryable message, and catch inside `refreshProviders`.

### IN-07: The hide-after-copy timer is never cancelled

**File:** `frontend/src/App.svelte:375`
**Issue:** `setTimeout(hide, HIDE_AFTER_COPY_MS)` is not stored or cleared. If the panel is hidden and re-shown with the hotkey within 400 ms of a copy, the stale timer hides the freshly shown panel.
**Fix:** Store the handle and clear it in the `panel:show` handler.

---

_Reviewed: 2026-10-17T21:30:34Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
