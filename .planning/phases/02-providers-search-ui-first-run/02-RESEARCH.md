# Phase 2: Providers, Search UI & First Run - Research

**Researched:** 2026-10-17
**Domain:** Go provider clients (GIPHY, GIPHY Clips, Klipy, Finer Gifs) behind Wails v3 bindings, Keychain secrets, and a Svelte 5 keyboard-driven virtualized grid
**Confidence:** MEDIUM-HIGH. Wails, go-keyring and runtime behavior come from source read this session (HIGH). Vendor shapes come from official docs plus live keyless probes (HIGH for error shapes, MEDIUM for success shapes because no live key exists). Native paste into the key field and Esc handling are MEDIUM until the user checks them natively.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Panel Layout & Look
- Provider switcher is a row of tabs across the top of the panel — GIPHY · Clips · Klipy · Finer Gifs — with ⌘1–⌘4 shortcuts. Clips tab only appears when the user's GIPHY key passes the Clips test call; otherwise a short explanation is shown where Clips would be.
- Grid of 4 columns with square-ish tiles, using each provider's small preview renditions (never originals in the grid); the selected tile is clearly highlighted.
- Attribution lives in a slim footer bar: "Powered by GIPHY" (official GIPHY asset, not the extension's PNG) on GIPHY and Clips; Klipy attribution on Klipy. The Klipy search placeholder reads exactly "Search KLIPY".
- Visual style follows the system light/dark appearance, system font, Spotlight-like rounded panel (opaque webview + CSS rounded corners, per research — no private_mac_apis translucency).

#### Keys & First Run
- First run opens on GIPHY with an inline setup card inside the panel: user copies their key, opens the panel, pastes (⌘V into the field), Save. The key is verified with a test call before saving; a bad key gets a clear message. Full key management (remove/replace) is Phase 4 Settings.
- The card has a "Get a free key" link that opens the provider's developer site in the default browser, plus a "Try Finer Gifs (no key)" button that switches to Finer Gifs.
- Switching to Klipy without a key shows the same inline card for Klipy.
- Keys are stored in the macOS Keychain (zalando/go-keyring behind a SecretStore interface); they never reach the webview (UI only learns `configured: bool`), logs, or error text (redaction in one HTTP choke point — Klipy's key is in the URL path).
- Live API testing: the user will create free GIPHY and Klipy keys and enter them through the app's own setup card at UAT. Planning research works from the vendors' documented API shapes and recorded fixtures; no keys are handed to research agents.

#### Search & Keyboard Behavior
- Search fires 450 ms after typing stops, minimum 2 characters; superseded in-flight requests are cancelled (Wails cancellable binding → ctx cancels the HTTP request).
- Page size 24 (6 rows × 4), clamped to provider limits (GIPHY ≤50, Klipy 8–50); next page loads when the selection is within 1–2 rows of the end and via a scroll sentinel.
- Keyboard: typing always goes to the search box; arrow keys move the grid selection; Enter copies the selected GIF and closes the panel; Esc clears the query first, then closes on a second press. ⌘1–⌘4 switch providers. ⌘K is bound (no-op placeholder until Phase 3).
- Trending results are cached for 10 minutes per provider so reopening the panel doesn't spend a request. Rate-limit (429) shows a clear message with when to retry (Retry-After); network/timeout/provider errors show per provider, and one failing provider never blocks the others (Finer Gifs was down at research time — must degrade cleanly).

#### Carried from earlier decisions
- Last-used provider is remembered across launches.
- Providers are never blended in one grid; results are never reordered or filtered (only same-page duplicate IDs removed, never for Klipy).
- Provider HTTP lives in Go (internal/provider/* + httpx), never in the webview.
- Existing Phase 1 primitives: `internal/platform` (CopyFile etc.), `internal/services.ActionService` (Copy/Paste with busy try-lock), `internal/shell` (panel/hotkey/tray), `frontend/src/App.svelte`; the copy path must now copy the selected result's GIF (download to staging, validated) instead of the bundled sample.
- Browser UAT via `scripts/smoke-browser-loop.sh` (-tags server + Vite proxy on 8089/9245); native behaviors are user-verified.
- Copy works for GIPHY, Klipy, Finer Gifs results; Clips copy waits for Phase 3 (MP4).
- Written compliance posture (no persistent media cache, no blending/reordering, attribution, eager staging cleanup) goes into PROJECT.md Key Decisions in this phase.

### Claude's Discretion
- Exact component breakdown of the Svelte UI, virtualization library choice per research (@tanstack/svelte-virtual), and visual spacing.
- Klipy API shape details (native v1 API per research), Clips capability-probe mechanics, and how the request budget/cache is structured internally.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.

Also out of this phase (from the CONTEXT boundary): action menu contents (⌘K opens it in Phase 3), Clips MP4 file handling (Phase 3, ACTN-09), temp-file lifecycle hardening (Phase 3, ACTN-10), favorites/recents (Phase 3), Settings window and key removal/replace (Phase 4), GIPHY rating control (Phase 4; PG-13 applied by default now).

**Approved UI contract:** `02-UI-SPEC.md` (status approved, user confirmations 2026-10-17) is binding for layout, copy, states S1–S13/C1–C2 and the keyboard table. This research does not restate it. It adds the backend contracts the UI needs and flags three small UI-SPEC conflicts found while verifying (see Open Questions 1–3).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SRCH-01 | Search GIPHY, Klipy, Finer Gifs; switch provider in panel; last provider remembered | Provider clients (Pattern 2), endpoint tables (Provider API Reference), `settings.json` lastProvider (Pattern 6) |
| SRCH-02 | Clips only when the key passes a Clips test call; else explanation | Clips probe state machine (Pattern 5); live probe: invalid key gets 401 on both `/v1/gifs` and `/v1/clips`, so the probe must run only after the GIFs check passes |
| SRCH-03 | Empty query shows trending (where supported) | Trending endpoints; Finer Gifs has none (state S8); 10-min trending cache (Pattern 4) |
| SRCH-04 | Animated grid with load-more on scroll or arrow near end | Opaque cursor per provider (Pattern 3), row virtualization (Pattern 8), Svelte duplicate-key pitfall (Pitfall 7) |
| SRCH-05 | Full keyboard navigation (arrows, Enter, ⌘K, Esc) | Keyboard module (Pattern 9); `HideOnEscape` must go false plus a `WindowClosing` hook (Pitfall 3) |
| SRCH-06 | Debounced searches; 429 shows when to retry | 450 ms debounce plus cancellation (Pattern 7), `blockedUntil` budget (Pattern 4), Retry-After parsing (GIPHY sends `retry-after: 0` on 401, Pitfall 4) |
| SRCH-07 | Per-provider network/timeout/provider errors; isolation | Typed `provider.Error` with `MarshalError` (Pattern 1); Finer Gifs live probe: HTTP 530 "error code: 1033" |
| ONBD-01 | First run opens on GIPHY setup card plus "Try Finer Gifs" | `Providers()` reports `configured`; frontend default provider rule (Pattern 6) |
| ONBD-02 | Key verified by a test call before saving | `SettingsService.SaveKey` verifies with a real first-page trending call and seeds the cache (Pattern 5); Klipy rejects bad keys with **404/403**, not 401 (live probe) |
| ONBD-03 | Keys in Keychain; never to UI, logs or errors | go-keyring v0.2.8 behind `SecretStore` (source read); Wails puts `err.Error()` in the call message, so `Error()` itself must never contain a URL (Pitfall 1) |
| ONBD-04 | "Powered by GIPHY" on GIPHY/Clips; "Search KLIPY" placeholder | UI-SPEC attribution contract (assets already chosen); placeholder is a per-provider constant |
| ONBD-05 | Never blended; never reordered or filtered except same-page dedupe for non-Klipy | One grid per provider; `dedupeSamePage` only when `provider != klipy`; Klipy ad objects only appear for mobile browser user agents (docs) |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

From `/Users/ken/dev/CLAUDE.md` and `./.claude/CLAUDE.md`:

- **Package managers:** Go uses `go mod`; Node uses **pnpm** (already set in `Taskfile.yml` `PACKAGE_MANAGER` default "pnpm"). Never mix with npm in `frontend/`.
- **Go conventions:** `go fmt`, `golangci-lint` (installed at `~/go/bin/golangci-lint`), context-aware functions (`ctx context.Context` first). Project pattern adds `go vet` and `go test -race`.
- **JS/TS conventions:** camelCase, PascalCase components, TypeScript types; `svelte-check` (`pnpm check`).
- **Testing:** 80%+ coverage in critical components (httpx, provider mappers, search service, key verification, copy pipeline are the critical components here). Integration tests for routes become "integration tests for bound services" in this Wails app.
- **Silent fallbacks forbidden:** `or {}` style defaults that turn failures into silent corruption are not allowed. Mappers that hit a malformed response must return a typed error, not an empty page. (Exception noted in Code Examples: tolerating a cosmetic preview width is fine; a missing download URL is not.)
- **Chesterton's Fence:** Phase 1 code (`HideOnEscape: true`, `SamplePath`, smoke-script checks) exists for recorded reasons. Each change below states why it must change.
- **Premature abstraction:** Three real examples before abstracting. The four providers justify the `Provider` interface. A generic rate limiter, a plugin registry, or a component library do not.
- **RULE 0:** Stop on unknown or catastrophic failures.
- **Never `kill node`** (Claude Code runs on Node). Smoke scripts already kill only their own PIDs; keep that.
- **LSP first** for code navigation (gopls, TypeScript).
- **UAT:** browser UAT uses the dev-browser skill. Native behaviors (Keychain, ⌘V into the key field, Esc in the real panel) are user-verified. **Never send synthetic input on the user's Mac.**
- **Ponytail** is active for executors: smallest change that fully solves the task. This research recommends exactly one new Go dependency and one new npm dependency.
- **GSD workflow:** commits on branch `gsd/phase-02-providers-search-ui-first-run`; `commit_docs: true`.

## Summary

Phase 2 turns the Phase 1 spike into the product: four provider clients in Go, a key setup flow that writes to the Keychain, a search service with caching and a 429 budget, typed errors that cross the Wails boundary, and a Svelte 5 panel with a virtualized 4-column grid and a hand-built keyboard model. The UI contract is already approved (`02-UI-SPEC.md`), so the open questions were all on the backend and integration side. This research answers them from vendor docs, live keyless probes run today, and the pinned Wails beta.30 and go-keyring v0.2.8 source.

Seven findings change the plan compared with the project-level research. (1) **Klipy rejects an invalid key with HTTP 404** (`{"result":false,"errors":{"message":["The provided API key is invalid."]}}`), and with 403 on some paths, never 401. Key-rejected handling must be per provider. (2) **GIPHY sends `retry-after: 0` on a 401**, so Retry-After may only be read on 429, and 0 means "unknown". (3) **GIPHY documents "200 with an empty body" as an error** and 414 for queries over 50 characters. (4) **Wails always sends `err.Error()` as the call's message, even when `MarshalError` is set.** Redaction cannot live only in the marshaler: the `Error()` string itself must never contain a URL. (5) **`HideOnEscape` is a Wails key binding that swallows Esc before the page sees it.** Removing it is right, but NSPanel's default `cancelOperation:` and the default menu's ⌘W can then close (and destroy) the panel, so a `WindowClosing` hook that cancels and hides is mandatory. (6) **Wails' default macOS application menu provides Edit > Paste (`paste:` to the first responder),** so ⌘V into the key field should work with no code. It still needs a native check, and the default View/File/Window menus bring ⌘W, ⌘R, ⌘+/⌘−, ⌘M, which a fixed panel does not want. Replace the menu with App + Edit only. (7) **Finer Gifs is still down** (HTTP 530 / Cloudflare 1033 at 17:36 UTC today). Its own website's bundle reveals the real response envelope (`results[].fields`, `hits.found`, `hits.start`), which lets paging use the server total instead of a length heuristic.

For testing without keys, use one fixture HTTP server (`internal/provider/fakeapi`) that speaks the real vendor wire formats. Go unit tests use it through `httptest`, and the `-tags server` build points all four clients at it for dev-browser UAT. That exercises the real mappers, httpx and error classification end to end, and magic query words can drive every UI state (429, timeout, empty, 5xx, bad key, Clips denied).

**Primary recommendation:** Build the Go domain first, contract first: `provider.Error` + `httpx` + the fixture server, then the providers, then `search.Service` + `SecretStore` + `SettingsService`. Then the Svelte panel against the server build, the copy pipeline swap, and the native user checks last. Add exactly two dependencies: `github.com/zalando/go-keyring v0.2.8` and `@tanstack/svelte-virtual` (exact pin, behind a human package-gate checkpoint).

## Architectural Responsibility Map

The "tiers" in this desktop app are the Go backend (domain packages), the Wails binding layer (services), the webview (Svelte), and native macOS (Keychain via `/usr/bin/security`, pasteboard via cgo).

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Provider HTTP (search, trending, key test, Clips probe) | Go domain (`internal/provider/*`, `httpx`) | — | Keys must never reach the webview; Klipy's key is in the URL path (locked decision) |
| Error classification and redaction | Go domain (`provider.Error`) | Wails `MarshalError` | `Error()` is sent verbatim; classification must happen before the boundary |
| Trending/search cache, 429 budget | Go domain (`internal/search`) | — | One choke point protects the 100/hr cap for all UI paths |
| Key storage | Native Keychain via go-keyring | Go in-memory cache | `security` CLI, no cgo; read once, cache in memory |
| Key entry | Webview (setup card) | Go `SettingsService.SaveKey` | The user types or pastes into the webview; Go verifies and stores; the key never comes back |
| Last provider, non-secret prefs | Go (`internal/settings`, JSON file) | — | Survives relaunch; webview storage would not be authoritative |
| Debounce, cancel superseded call | Webview | Go ctx | Frontend owns timing; `CancellablePromise.cancel()` cancels the Go ctx |
| Grid rendering, virtualization, selection, keyboard | Webview | — | Pure UI state |
| Media loading for tiles | Webview (`<img src>` from CDN) | — | GIPHY and Klipy require direct loads; no proxying |
| Copy selected GIF | Go (`ActionService`: download, validate, stage) | Native pasteboard (cgo) | Go fetches allowlisted URLs only and writes the file URL |
| Panel hide (Esc, after copy) | Go `PanelService.Hide` | Wails window | Esc no longer hides natively; the page decides |
| External links ("Get a free key ↗", dashboard) | Webview → `Browser.OpenURL` (Wails runtime) | — | Fixed URL constants in the frontend |

## Standard Stack

### Core (already in the repo, unchanged)
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Wails v3 | `v3.0.0-beta.30` (pinned) | Bindings, events, window, menu, browser open | [VERIFIED: go.mod line 5 `require github.com/wailsapp/wails/v3 v3.0.0-beta.30`] |
| `@wailsio/runtime` | `3.0.0-beta.30` (exact) | `CancellablePromise`, `CancelError`, `Events`, `Browser.OpenURL`, `Clipboard.Text` | [VERIFIED: frontend/package.json:14 `"@wailsio/runtime": "3.0.0-beta.30"`] |
| Svelte | 5.57.2 (installed) | UI with runes | [VERIFIED: frontend/node_modules/svelte/package.json `"version": "5.57.2"`] |
| Go stdlib `net/http`, `encoding/json`, `net/url`, `os` | Go 1.27.2 | HTTP client, JSON, files | No third-party HTTP client needed |

### New dependencies (exactly two)
| Library | Version | Purpose | Why |
|---------|---------|---------|-----|
| `github.com/zalando/go-keyring` | `v0.2.8` (2026-03-23) | Keychain behind `SecretStore` | Locked decision. Source read: darwin impl shells to `/usr/bin/security`, sends the secret on **stdin** (`security -i`, `add-generic-password -U`), and maps "could not be found" to `keyring.ErrNotFound`. No cgo, so `-tags server` stays CGO-free [VERIFIED: proxy.golang.org @latest + module source `keyring_darwin.go`] |
| `@tanstack/svelte-virtual` | `3.13.47` (pin exact; latest at research time) | Row virtualization | Named in UI-SPEC and CONTEXT. Peer `svelte: ^3.48.0 \|\| ^4.0.0 \|\| ^5.0.0`; depends on `@tanstack/virtual-core 3.19.6`; no postinstall. API: `createVirtualizer(opts)` returns a **Svelte store** of the virtualizer [VERIFIED: npm registry + tarball `dist/index.js`/`index.d.ts`] |

### Explicitly NOT added
| Candidate | Why not |
|-----------|---------|
| `golang.org/x/time/rate` (STACK.md "strongly recommended") | A token bucket does not model a vendor's opaque hourly window, and we cannot see the server's count anyway. A `blockedUntil` map set on 429 (Pattern 4) is the budget that matters: never retry into a 429. Debounce plus cache already minimize spend |
| `golang.org/x/sync/singleflight` | One in-flight page per provider is enforced by the UI; the cache absorbs repeats. Add only if duplicate concurrent trending fetches show up in logs |
| `vitest` | Node 24.14 runs `node --test` on `.ts` files natively (verified in scratch: 2/2 tests pass with type stripping). Keep pure TS logic (keyboard, error mapping) in `.ts` files and test with zero deps |
| `@giphy/js-types` | Types-only and stale; Go owns the shapes and Wails generates TS |
| Component libraries | UI-SPEC: none |

**Installation:**
```bash
go get github.com/zalando/go-keyring@v0.2.8
(cd frontend && pnpm add --save-exact @tanstack/svelte-virtual@3.13.47)
```

## Package Legitimacy Audit

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| `@tanstack/svelte-virtual` | npm | project since 2021; v3.13.47 published 2026-10-16 | 82,132/wk | github.com/TanStack/virtual | SUS (reason: `too-new`) | Flagged: planner adds `checkpoint:human-verify` before install (same gate pattern as Phase 1 Task 1) |
| `@tanstack/virtual-core` (transitive) | npm | v3.19.6 published 2026-10-16 | 32,664,852/wk | github.com/TanStack/virtual | SUS (reason: `too-new`) | Flagged with the above (one approval covers both) |
| `github.com/zalando/go-keyring` | Go proxy | v0.2.8 tagged 2026-03-23 | n/a (seam has no Go ecosystem) | github.com/zalando/go-keyring | not run by seam; source read this session | Approved by locked CONTEXT decision; source read confirms behavior |

`too-new` is structural, not a red flag: TanStack publishes automated patch releases every 1–2 days (3.13.40 on 2026-10-09 through 3.13.47 on 2026-10-16). `postinstall` is null for both npm packages [VERIFIED: `gsd-tools query package-legitimacy check`, `npm view ... time`].

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** `@tanstack/svelte-virtual`, `@tanstack/virtual-core`. Planner inserts one `checkpoint:human-verify` before `pnpm add`.

## Provider API Reference (plan-ready)

All requests: `GET`, `User-Agent: Gifsmith/<version>`, 15 s client timeout, page size 24.

### GIPHY GIFs [CITED: developers.giphy.com/docs/api/endpoint, /docs/api/schema, /docs/optional-settings]

| | Search | Trending |
|---|---|---|
| Path | `https://api.giphy.com/v1/gifs/search` | `https://api.giphy.com/v1/gifs/trending` |
| Params we send | `api_key`, `q` (**max 50 chars**, else 414), `limit=24` (beta max 50), `offset`, `rating=pg-13`, `bundle=messaging_non_clips` | `api_key`, `limit=24`, `offset`, `rating=pg-13`, `bundle=messaging_non_clips` |
| Offset cap | 4999 | **499** |
| Response | `{data: GIF[], pagination: {total_count, count, offset}, meta: {status, msg, response_id}}`. `total_count` "not returned on every endpoint" | same |
| Grid tile | `images.fixed_width.url` (200 px wide GIF; tile is 166 css px). Docs recommend fixed-width/height renditions for grids | same |
| Copy | `images.original.url` (extension parity) | same |
| Title | `title`, fallback `slug` | same |

- `bundle=messaging_non_clips` returns exactly: `original, fixed_height, fixed_height_downsampled, fixed_height_small, fixed_width, fixed_width_downsampled, fixed_width_small`. Both renditions we need are included, and payloads shrink [CITED: optional-settings].
- Do not send `lang` (Phase 4 locale) or `random_id` (not a documented param).
- **Wire quirk:** rendition `width`, `height`, `size` arrive as **strings** in real responses (`"width": "200"`) while `@giphy/js-types` declares numbers. Decode them with a tolerant int type [ASSUMED from community samples; type file read this session; confirm in the live-key test].
- Errors (docs): 400, 401 and 403 (key), 404, 414 (q too long), 429 (too many requests). Errors return `data: []` plus `meta`. A **200 with an empty body must be treated as an error** [CITED: developers.giphy.com/docs/api response codes].
- **Live probe today** (no key, and a bogus key): `HTTP/2 401`, `retry-after: 0`, body `{"data": [], "meta": {"status": 401, "msg": "Unauthorized", "response_id": ""}}` [VERIFIED: live probe 17:36 UTC].
- Paging: `next = offset + count`. More results if `total_count > 0 ? next < total_count : count == limit`, **and** `next <= cap` (4999 search, 499 trending).

### GIPHY Clips [CITED: developers.giphy.com/docs/clips/endpoint]

| | Search | Trending |
|---|---|---|
| Path | `/v1/clips/search` | `/v1/clips/trending` |
| Params | `api_key`, `q` (max 50), `limit`, `offset`, `rating` (**defaults to R: always send `rating=pg-13`**), `bundle=clips_grid_picker` | `api_key`, `limit`, `offset`, `rating` (defaults to PG; send `pg-13`), `bundle=clips_grid_picker` |
| Grid tile | `images.fixed_width.url` (`clips_grid_picker` = `fixed_width, original, video.assets.{360p,480p,720p,1080p,source}`) | same |
| Copy | Not in Phase 2 (footer E8) | |

- Approval: "Access to Clips endpoints is only available upon approval. Please reach out to clips@giphy.com." The status code for an unapproved key is **not documented** [CITED]. No observation was possible without a valid key.
- Live probe: an invalid key gets **401 on `/v1/clips/trending` as well**, identical to `/v1/gifs` [VERIFIED: live probe]. So a Clips 401 means nothing on its own; the probe is only meaningful after the GIFs call succeeded with the same key.

### Klipy native v1 [CITED: docs.klipy.com/gifs-api/gifs-search-api.md (fetched live today), gifs-trending-api.md, advertisements, integration-requirements]

| | Search | Trending |
|---|---|---|
| Path | `https://api.klipy.com/api/v1/{app_key}/gifs/search` | `https://api.klipy.com/api/v1/{app_key}/gifs/trending` |
| Params | `q`, `page` (min 1, default 1), `per_page` (**min 8**, max 50, default 24), optional `customer_id`, `locale`, `content_filter` (`off\|low\|medium\|high`), `format_filter` (`gif,webp,jpg,mp4,webm`) | `page`, `per_page` (min 1, max 50), same optional params |
| Response | `{result: true, data: {data: Item[], current_page, per_page, has_next}}`. No total count | same |
| Item | `id` (16-digit number), `slug`, `title`, `file.{hd,md,sm,xs}.{gif,webp,jpg,mp4,webm}.{url,width,height,size}`, `tags`, `type` (`"gif"`), `blur_preview` (data URI) | same |
| Grid tile | `file.sm.gif.url` (220×220, about 315 KB in the doc example). `sm.webp` is 4× smaller (80 KB) but see Assumption A3 | |
| Copy | `file.hd.gif.url` (about 4 MB in the doc example), fallback `md.gif` | |
| ID | **Use `slug`** (the items API takes slugs; `id` is near the float64 precision limit) | |

- Key in path: build with `url.PathEscape(key)`, and reject keys containing `/`, `?`, `#`, whitespace or control characters before any request.
- **Live probe, invalid key:** `HTTP/2 404`, `content-type: application/json`, body `{"result":false,"errors":{"message":["The provided API key is invalid."]}}` on both search and trending; `HTTP/2 403` with the same body on an unknown sub-path [VERIFIED: live probe 17:37 UTC]. Classify Klipy **401, 403 and 404** as `key_rejected` (our paths are constants, so 404 cannot mean "wrong route").
- 429 shape and Retry-After for Klipy: **not documented**; the Testing-mode limit is "100 API requests per hour" [CITED: getting-started]. Treat 429 generically.
- Ads: "When `type` is set to `"ad"`, it indicates that the object is an advertisement." Ads must be enabled per key in the Partner Panel, and "Advertisement delivery is only supported on mobile devices… you must include a browser-like User-Agent" [CITED: docs.klipy.com/advertisements]. With a `Gifsmith/x` UA and a self-issued key, ads should never arrive. Map only `type == "gif"` items and log a count of anything else (see Assumption A4).
- Rules that bind the implementation: preserve returned URLs exactly; load media directly from them; no media cache; requests from the end-user client; **do not reorder, insert, remove, suppress or filter results** (so no dedupe); dedicated grid [CITED: integration-requirements.md].

### The Finer Gifs Club [VERIFIED: live probe + site bundle read this session]

- **Status 2026-10-17 17:36 UTC:** `GET https://api.thefinergifs.club/search?q=dwight&q.parser=simple&size=2&start=0` returned `HTTP/2 530`, `content-type: text/plain`, body `error code: 1033`. `https://thefinergifs.club/` returned 200 (static site). `https://media.thefinergifs.club/` returned 404 `application/xml` (the media bucket host answers). Still down, the same as at the earlier probe today.
- Request (from the site's own Nuxt bundle `/_nuxt/app.3d29e032e6eefe5f040f.js`): `GET https://api.thefinergifs.club/search` with `q`, `q.parser=simple`, `start=(page-1)*size`, `size`, `sort="_score desc"` (or `"start_time asc"` with no query). Our client: `q`, `q.parser=simple`, `sort=_score desc`, `size=24`, `start=<cursor>`.
- Response envelope (the site reads `k.results` and `k.hits`, then `results.map(r => r.fields)`, `hits.start`, `hits.found`): `{results: [{fields: {name, text, fileid, ...}}], hits: {found, start}}`. The extension's model reads `results[].fields.{name,text,fileid}`.
- Media: `https://media.thefinergifs.club/<fileid>.gif`. That is the **only** rendition, so grid and copy use the same URL. The "never originals in the grid" rule cannot apply here, and the site itself does the same.
- Validate `fileid` against `^[A-Za-z0-9._-]+$` before building the URL (Pitfall 21 in project research).
- No trending: `Trending()` returns an empty page with `Next == ""` and no request (state S8 in the UI).
- Paging: `next = start + len(results)`; more results if `hits != nil ? next < hits.found : len(results) == size`.

## Architecture Patterns

### System Architecture Diagram

```
 keystroke ──► [Svelte SearchBar] ──450ms debounce, ≥2 chars──► api.search(provider,q,cursor)
                    │  ⌘1–4 / tab click (no debounce)               │  CancellablePromise
                    │  prev.cancel() ───────────────cancel call────►│
                    ▼                                               ▼
           [per-provider UI state]                  Wails binding (ctx cancelled on cancel)
                    ▲                                               │
                    │ Page{items,next} | err.cause{kind,...}        ▼
                    │                                    [services.GifService]
                    │                                               │
                    │                                               ▼
                    │                        [search.Service] ── blockedUntil[p] > now? ──► rate_limited (no network)
                    │                               │ cache hit (trending 10m / search 5m)? ──► cached Page
                    │                               ▼ miss
                    │                        [provider.{giphy,clips,klipy,finergifs}]
                    │                               │  key from SecretStore cache (missing → key_missing)
                    │                               ▼
                    │                        [httpx.GetJSON] ── 15s timeout, UA ──► vendor API (or fakeapi in -tags server)
                    │                               │  classify: canceled | timeout | network | 429(+Retry-After) | non-2xx | empty/invalid JSON
                    │                               ▼
                    │                        mapper → []Gif (+ same-page dedupe unless klipy) → Page{Next}
                    │
 Enter ──► api.copy(gif) ──► [services.ActionService] ─ busy try-lock ─► allowlist URL ─► download (size cap, ctx)
                                                                       ─► magic bytes GIF8 ─► staging/<hash>/<slug>.gif
                                                                       ─► platform.CopyFile ─► delete older staging dirs
 Frontend shows "Copied" → 400 ms → PanelService.Hide()

 Save Key ──► SettingsService.SaveKey(p,key) ─► sanitize ─► provider.Trending(ctx) with candidate key
                ├─ ok  → SecretStore.Set → cache seeded with that first page → (GIPHY) start Clips probe → emit providers:changed
                └─ err → typed error (key_rejected | rate_limited | network | timeout | http) → inline message K2–K4
```

### Recommended Project Structure
```
internal/
├── provider/
│   ├── provider.go        # ID consts, Gif, Page, Provider interface, Error/Kind, KeyFunc
│   ├── httpx/             # GetJSON (timeout, UA, classification, Retry-After), no URL in errors
│   ├── giphy/             # gifs + clips client (one package, two IDs), mapper, testdata/*.json
│   ├── klipy/             # client, mapper, testdata/*.json
│   ├── finergifs/         # client, mapper, testdata/*.json
│   └── fakeapi/           # http.Handler speaking all four wire formats + scenario words (tests + server build)
├── search/                # Service: cache, blockedUntil budget, dedupe policy, Clips probe state
├── settings/              # SecretStore (keyring impl + memory impl), prefs JSON (lastProvider)
├── staging/               # Fetch(ctx, url) → validated file under cache dir; Sweep(keep)
├── services/              # Wails adapters: GifService, SettingsService, ActionService (Copy/Paste take a Gif), PanelService
├── platform/, shell/      # Phase 1 (shell: HideOnEscape false + WindowClosing hook + App/Edit menu)
main.go                    # ALL application.NewService(...) calls stay here, untagged
wire_desktop.go            # //go:build !server : keyring store, real base URLs, real config dir
wire_server.go             # //go:build server  : memory store, fakeapi on 127.0.0.1, bin/uat-config dir
frontend/src/
├── app.css                # tokens (UI-SPEC)
├── App.svelte             # top-level state machine, event wiring
├── lib/api.ts             # facade over bindings; providerError(e); cancellation helper
├── lib/grid.ts            # pure selection math (node --test)
├── lib/errors.ts          # kind → UI-SPEC copy row (node --test)
└── components/            # SearchBar, ProviderTabs, ResultGrid, Tile, SetupCard, StateMessage, Footer
```

### Pattern 1: Typed provider errors that are safe by construction
**What:** One error type for everything the UI must distinguish. Its `Error()` string contains only the provider ID, kind and status, never a URL or body. `MarshalError` turns it into `err.cause` for JS.
**Why:** Wails builds the call error as `CallError{Message: errors.Join(errorOutputs...).Error(), Cause: marshalError(err), Kind: RuntimeError}` [VERIFIED: wails v3 beta.30 `pkg/application/bindings.go:424-436`]. The **message is always `err.Error()`**, so a wrapped `*url.Error` would put the Klipy key into the webview even with a perfect marshaler. `MarshalError` "may return a nil slice to fall back to the globally configured error handler" [VERIFIED: `services.go:31-39`, `application_options.go:41-49`]. The JS runtime sets `err.cause = json.cause` (already parsed JSON) [VERIFIED: `@wailsio/runtime/dist/runtime.js:136-146`].
```go
// internal/provider/provider.go
type Kind string

const (
	KeyMissing  Kind = "key_missing"
	KeyRejected Kind = "key_rejected"
	RateLimited Kind = "rate_limited"
	Network     Kind = "network"
	Timeout     Kind = "timeout"
	HTTP        Kind = "http" // non-2xx other than the above, or an empty/invalid body (Status = HTTP status)
)

type Error struct {
	Kind          Kind `json:"kind"`
	Provider      ID   `json:"provider"`
	Status        int  `json:"status,omitempty"`
	RetryAfterSec int  `json:"retryAfterSec,omitempty"` // 0 = unknown
}

// Error never includes a URL, query, key or response body: Wails sends this string to JS verbatim.
func (e *Error) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("%s: %s (HTTP %d)", e.Provider, e.Kind, e.Status)
	}
	return fmt.Sprintf("%s: %s", e.Provider, e.Kind)
}

// main.go
MarshalError: func(err error) []byte {
	var pe *provider.Error
	if errors.As(err, &pe) {
		b, _ := json.Marshal(pe) // cannot fail for this struct
		return b
	}
	return nil // default marshaling for everything else
},
```

### Pattern 2: Provider clients with an injectable base URL and key function
**What:** Each client is a struct with `BaseURL string`, `Key func() (string, bool)` (nil for Finer Gifs), and `HTTP *http.Client`. Production uses vendor URLs; tests and the server build use the fixture server.
```go
type Provider interface {
	Search(ctx context.Context, q, cursor string) (Page, error)
	Trending(ctx context.Context, cursor string) (Page, error) // Finer Gifs: empty page, no request
}

type Gif struct {
	Provider    ID     `json:"provider"`
	ID          string `json:"id"`          // GIPHY id, Klipy slug, Finer fileid
	Title       string `json:"title"`
	PreviewURL  string `json:"previewUrl"`  // grid only
	DownloadURL string `json:"downloadUrl"` // copy only (empty for Clips in Phase 2)
	IsClip      bool   `json:"isClip"`
}

type Page struct {
	Items []Gif  `json:"items"`
	Next  string `json:"next"` // opaque; "" = end of results
}
```
Keep `Gif` minimal. Phase 3 adds page URL, dimensions and metadata when the details and action menu need them.

### Pattern 3: Opaque cursor owned by each provider
GIPHY cursor = decimal offset; Klipy = page number; Finer = `start` offset. The frontend passes `next` back unchanged and never parses it. An unparsable cursor is a programmer error: return `&Error{Kind: HTTP}`, never silently restart at page 1.

### Pattern 4: `search.Service`: cache plus a 429 budget, one mutex
**What:** Wraps the four providers. Before any network call: if `blockedUntil[p] > now`, return `RateLimited` with the remaining seconds (no request). Cache key `p|kind|q|cursor`, TTL 10 min for trending (locked) and 5 min for search (discretion; it makes backspace-and-retype free). On a 429, set `blockedUntil[p] = now + max(RetryAfter, 60s)`, so a "Try Again" click inside the window costs nothing. On a key change, drop that provider's cache entries and block. Same-page dedupe runs here only when `p != klipy`.
**Why not a token bucket:** the vendor window and count are invisible to us (keys may also be used elsewhere). The thing that burns the quota is retrying into a 429, which this prevents.
**Compliance note:** GIPHY says "Do not cache media URLs" for partner integrations. A 10-minute in-memory cache of API responses is the user's locked decision. It never touches disk and holds no media bytes. Record it in the PROJECT.md compliance posture.

### Pattern 5: Key verification and the Clips probe
- `SettingsService.SaveKey(ctx, p, key)`: trim, then sanitize (length 8–128, no whitespace, control characters, `/`, `?`, `#`). Build a client with the **candidate** key and call `Trending(ctx, "")` with the real page size. If it succeeds: `SecretStore.Set`, update the in-memory key cache, seed the search cache with that page (verification doubles as the first trending fetch, so it costs no extra request), and emit `providers:changed`. If it fails: return the typed error, which the card maps to K2 (`key_rejected`), K3 (`network`/`timeout`) or K4 (`rate_limited`).
- **Clips probe** (GIPHY only): `GET /v1/clips/trending?limit=1&rating=pg-13` with the stored GIPHY key, only after the GIFs check has succeeded at least once this process (at SaveKey, or the first GIFs success after launch). Then 200 → `enabled`; 401/403/404 → `disabled` (state C2); 429/network/timeout → `unknown` (choosing Clips re-probes once and shows the skeleton; a repeat failure shows that error for "GIPHY Clips"). Hold the state in memory per process (one call per launch at most) and emit `providers:changed` on change.
- `GifService.Providers()` returns `[{id, configured, clips: "no_key"|"checking"|"enabled"|"disabled"|"unknown"}]` for the four fixed slots. Titles and placeholders stay frontend constants per UI-SPEC.
- A stored key rejected during search (S13) is just `key_rejected` from `Search`/`Trending`. The UI re-shows the card, and SaveKey overwrites the Keychain item (`add-generic-password -U` updates in place [VERIFIED: keyring_darwin.go Set]).

### Pattern 6: SecretStore and settings
```go
type SecretStore interface {
	Get(account string) (string, error) // settings.ErrNoKey when absent
	Set(account, secret string) error
}
// keyringStore: service "com.kenscott.gifsmith", account "giphy" | "klipy".
// Map keyring.ErrNotFound → ErrNoKey. Never include the secret in any error or log.
// memoryStore: map + mutex (unit tests and -tags server).
```
- Read both keys once at startup (each Get spawns `/usr/bin/security`) and cache them behind a mutex. `configured` comes from the cache.
- Prefs: `os.UserConfigDir()/com.kenscott.gifsmith/settings.json` = `{"lastProvider": "giphy"}`, written atomically (temp file + rename). `SettingsService.SetLastProvider(id)` is called on each tab switch. Startup rule (UI-SPEC): persisted provider, but Clips falls back to GIPHY if unavailable; no file means GIPHY (first run shows S1).
- Server build: config dir `bin/uat-config/` and the memory store, so UAT never touches the user's real Keychain or settings.

### Pattern 7: Cancellable search from Svelte
`$Call.ByID` returns a `CancellablePromise`; `.cancel()` sends a CancelCall that cancels the Go ctx [VERIFIED: `@wailsio/runtime/dist/calls.js:52-78`; Go side `messageprocessor_call.go:25-34,60` registers `runningCalls[callID] = cancel`]. The cancelled promise rejects with `CancelError` (exported from `@wailsio/runtime`).
```ts
// lib/api.ts
import { CancelError, type CancellablePromise } from '@wailsio/runtime';
import { GifService } from '../../bindings/gifsmith/internal/services';

let inflight: CancellablePromise<unknown> | null = null;

export function latest<T>(p: CancellablePromise<T>): Promise<T | undefined> {
  inflight?.cancel();
  inflight = p;
  return p.then(
    (v) => (inflight === p ? v : undefined),
    (e) => { if (e instanceof CancelError || inflight !== p) return undefined; throw e; },
  );
}

export type ProviderErr = { kind: string; provider: string; status?: number; retryAfterSec?: number };
export function providerError(e: unknown): ProviderErr | null {
  const c = (e as { cause?: unknown })?.cause as ProviderErr | undefined;
  return c && typeof c.kind === 'string' ? c : null;
}
```
Use separate "latest" slots for first-page queries and for page loads, and one slot per provider, so a background page load on one tab is not cancelled by typing on another. In Go, `context.Canceled` from httpx returns as is and is **not logged as an error**.

### Pattern 8: Row virtualization with `@tanstack/svelte-virtual` in Svelte 5
`createVirtualizer` returns a store and calls `virtualizerWritable.set(...)` on every `setOptions` [VERIFIED: tarball `dist/index.js`]. Reading `$virt` inside the `$effect` that calls `setOptions` makes the effect depend on the store it updates (a loop risk). Wrap the call in `untrack`.
```svelte
<script lang="ts">
  import { createVirtualizer } from '@tanstack/svelte-virtual';
  import { untrack } from 'svelte';
  import { get } from 'svelte/store';
  let { items, selected } = $props<{ items: Gif[]; selected: number }>();
  let scrollEl: HTMLDivElement;
  const COLS = 4, ROW = 152; // UI-SPEC: 144 tile + 8 gap
  const virt = createVirtualizer<HTMLDivElement, HTMLDivElement>({
    count: 0, getScrollElement: () => scrollEl, estimateSize: () => ROW, overscan: 2,
    paddingStart: 16, paddingEnd: 16, scrollPaddingStart: 16, scrollPaddingEnd: 16,
  });
  const rows = $derived(Math.ceil(items.length / COLS));
  $effect(() => { const n = rows; untrack(() => get(virt).setOptions({ count: n })); });
  $effect(() => { const r = Math.floor(selected / COLS); untrack(() => get(virt).scrollToIndex(r, { align: 'auto' })); });
</script>
<div bind:this={scrollEl} class="scroller" role="listbox" id="grid">
  <div style="height:{$virt.getTotalSize()}px; position:relative">
    {#each $virt.getVirtualItems() as row (row.index)}
      <div class="row" style="transform:translateY({row.start}px)">
        {#each items.slice(row.index * COLS, row.index * COLS + COLS) as g, c (c)}
          <!-- Tile: id="tile-{row.index*COLS+c}", role="option", aria-selected -->
        {/each}
      </div>
    {/each}
  </div>
</div>
```
Key rows by row index and tiles by column index, **never by GIF id** (Pitfall 7). Confidence: MEDIUM. The pattern follows the adapter source; confirm with `pnpm check` and browser UAT. Fallback if the store/runes interplay misbehaves: with fixed 152 px rows, a hand-rolled window (`start = floor(scrollTop/152) - 2`) is about 20 lines. Record a deviation if that path is used, because UI-SPEC names the library.

### Pattern 9: Keyboard model as pure functions
Put selection math and key routing in `lib/grid.ts` (no Svelte imports) so `node --test` covers it. The `onkeydown` on the search input calls `preventDefault()` for ←→↑↓ (the caret must not move), ↵, ⌘↵, ⌘1–4, ⌘K, Tab and **every** Esc (Pitfall 3). Ignore ↵ when `isComposing || keyCode === 229 || repeat || copyInFlight` (Phase 1 guards). Do not search during IME composition; run the debounce from `compositionend` too.

### Pattern 10: Copy pipeline swap (Phase 1 `SamplePath` → selected GIF)
`ActionService.Copy(ctx, g provider.Gif)` and `Paste(ctx, g provider.Gif)`:
1. Busy try-lock (existing `ErrBusy`).
2. `g.IsClip` → return a typed "clips copy unsupported" result (the UI shows E8 and never calls it, but Go enforces it too).
3. Validate `g.DownloadURL`: `https` and a host allowlist per `g.Provider` (GIPHY: `giphy.com` and `*.giphy.com`; Klipy: `*.klipy.com`; Finer: `media.thefinergifs.club`; server build adds the fakeapi host for its fake provider IDs only). The webview cannot make Go fetch an arbitrary URL.
4. `staging.Fetch(ctx, url, provider, id, title)`: 45 s ctx deadline, status 200, cap 50 MB (`io.LimitReader` + overflow check), first bytes `GIF87a`/`GIF89a`. Stream to `<UserCacheDir>/com.kenscott.gifsmith/staging/<sha1(provider:id)[:16]>/<slug>.gif.part`, then rename. The slug is Unicode-aware, falls back to `gif`, and is capped at 80 bytes.
5. `platform.CopyFile(path)`. Only after success, delete every other staging subdirectory (eager cleanup for Klipy compliance; the clipboard now points at the new file). Phase 3 hardens the lifecycle (ACTN-10).
6. **Copy no longer hides in Go.** UI-SPEC wants "Copied — press ⌘V to paste", then hide 400 ms later, so the frontend calls `PanelService.Hide()`. Paste keeps Phase 1's Go-side hide before ⌘V (required for the keystroke to land).
Chesterton's Fence: `SamplePath` and `stageSample()` existed to prove the pasteboard path without a provider [VERIFIED: main.go:20-35, 44]. Phase 2 supplies the real source, so both go. `TestCopyCopiesThenHides` changes to "copies, does not hide".

### Anti-Patterns to Avoid
- **Wrapping `*url.Error` with `%w`** anywhere on the provider path: its `Error()` prints the full URL, including the Klipy path key and the GIPHY `api_key`.
- **Putting `application.NewService(...)` in build-tagged files:** the binding generator finds services by statically analyzing `NewService` calls ("Delegate to NewService so that the static analyser may detect T" [VERIFIED: services.go]), and it runs without `-tags server`. Keep every `NewService` call in untagged `main.go`; tagged files only build dependencies.
- **Keying `{#each}` by GIF id.**
- **Reading Retry-After on non-429 responses** (GIPHY sends `retry-after: 0` on 401).
- **Treating a Clips 401 as "not approved" without a prior GIFs success.**
- **Dedupe or filtering on Klipy results.**
- **`{@html}` for titles:** provider titles are untrusted text; Svelte's `{title}` escapes them.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Keychain access | cgo SecItem calls | `zalando/go-keyring` v0.2.8 | Locked decision; stdin secret passing; survives ad-hoc rebuilds without prompts |
| Call cancellation plumbing | Request IDs and a Go-side cancel registry | Wails `CancellablePromise.cancel()` → ctx | Already wired end to end in beta.30 |
| Error transport to JS | String parsing of error messages | `Options.MarshalError` → `err.cause` | Structured, already parsed by the runtime |
| Edit menu / ⌘V in the key field | Custom paste handler | Wails `application.NewMenu()` + `AddRole(AppMenu)` + `AddRole(EditMenu)` | Paste maps to the `paste:` selector with no target, so it reaches WKWebView as first responder [VERIFIED: menuitem_roles.go:93-105, menuitem_selectors_darwin.go] |
| Hide instead of close | Subclassing NSPanel in cgo | `panel.RegisterHook(events.Common.WindowClosing, func(e){ e.Cancel(); panel.Hide() })` | Official pattern in `examples/hide-window/main.go:35-38` |
| Opening external links | cgo `NSWorkspace openURL` | `Browser.OpenURL(url)` in `@wailsio/runtime` | Exported (`dist/browser.js:18`) |
| Virtualization | Bespoke windowing | `@tanstack/svelte-virtual` | UI-SPEC contract (hand-roll is only the documented fallback) |
| HTTP-date parsing for Retry-After | Custom parser | `strconv.Atoi` then `http.ParseTime` | Handles both RFC forms |
| Frontend unit tests | A test framework dependency | `node --test file.test.ts` (Node 24 type stripping) | Zero deps; verified in scratch |

**Key insight:** every hard piece of this phase already exists in the pinned Wails runtime or the standard library. The real work is classification (which status means what for which vendor) and contracts (DTOs, error kinds), and those must be written and tested here.

## Common Pitfalls

### Pitfall 1: The API key leaks through `err.Error()` despite `MarshalError`
**What goes wrong:** A network failure returns `*url.Error` ("Get \"https://api.klipy.com/api/v1/<KEY>/gifs/search?...\": dial tcp..."), and Wails copies it into `CallError.Message`, which the UI or devtools shows.
**How to avoid:** httpx never returns or wraps the transport error. It classifies it (`errors.As(err, &netErr)` with `Timeout()`, or `ctx.Err()`) and returns a fresh `*provider.Error`. Logging uses `provider.Error.Error()` plus a path template (`/gifs/search`), never `req.URL`.
**Warning signs / test:** a unit test points a Klipy client at `http://127.0.0.1:1` (connection refused) with key `SECRETKEY123`, then asserts that neither `err.Error()` nor the `MarshalError` JSON contains `SECRETKEY123`. Repeat for GIPHY and for a timeout.

### Pitfall 2: Per-vendor "bad key" statuses
GIPHY: 401/403. Klipy: **404** (and 403) with `result:false` [VERIFIED: live probes]. A shared `401|403` check would show Klipy's bad key as "Klipy isn't responding (HTTP 404)". Classify in each client.

### Pitfall 3: Esc after removing `HideOnEscape`
**What goes wrong:** `HideOnEscape: true` registers an "escape" key binding that calls `window.Hide()` [VERIFIED: webview_window.go:381-385], and the panel's `sendEvent:` consumes bound keys before WKWebView [VERIFIED: webview_panel_darwin.m:34-46]. That is why the page never sees Esc today [VERIFIED: internal/shell/shell.go:42 `HideOnEscape:    true,`]. After setting it to false, an Esc the page does not `preventDefault` can travel up the responder chain to `-[WebviewPanel cancelOperation:]`, which calls `[super cancelOperation:]` [VERIFIED: webview_panel_darwin.m:84-90]. NSPanel's default is believed to close the panel [ASSUMED]. A user-initiated close emits `WindowClosing`, and Wails' own listener then destroys the window [VERIFIED: webview_window_darwin.m:479-494, webview_window.go:367-372]. The panel is created once and must never be destroyed.
**How to avoid:** (a) `preventDefault()` on every Esc keydown in the page; (b) `panel.RegisterHook(events.Common.WindowClosing, cancel + Hide)` in `shell.New`; (c) replace the default app menu so ⌘W (CloseWindow role) is gone.
**Warning signs:** after Esc or ⌘W the hotkey does nothing (window destroyed). This needs a native user check.

### Pitfall 4: Retry-After semantics
GIPHY includes `retry-after: 0` on 401 [VERIFIED: live probe]. Neither vendor documents Retry-After on 429 [CITED: GIPHY docs silent; Klipy docs silent]. Parse only on 429; treat ≤0 or garbage as unknown (`RetryAfterSec = 0`, UI copy "later this hour"); still block locally for 60 s.

### Pitfall 5: GIPHY query length and offset caps
`q` over 50 characters gets 414. Truncate to 50 **runes** before sending (GIPHY and Clips). Trending offset over 499 or search offset over 4999 must end paging (`Next = ""`), not error.

### Pitfall 6: GIPHY numeric fields as strings; Klipy 16-digit id
Decode GIPHY `width`/`height`/`size` with a tolerant type. Never decode Klipy `id` into `float64` or `interface{}`. Leave it out of the struct and use `slug`.

### Pitfall 7: Svelte 5 throws on duplicate keys
GIPHY pages often overlap across offsets, and same-page dedupe does not remove cross-page duplicates (and Klipy is never deduped). A keyed `{#each items as g (g.id)}` then throws `each_key_duplicate`. Key by position (row index, column index).

### Pitfall 8: Default macOS menu side effects in a fixed panel
Wails installs `DefaultApplicationMenu()` (App, File, Edit, View, Window, Help) when none is set [VERIFIED: application_darwin.go:303-312, menu_darwin.go:144-152]. That brings ⌘R reload (wipes UI state), ⌘+/⌘− zoom (breaks the fixed 720×480 layout), ⌘M minimize, and ⌘W close (Pitfall 3). Set `app.Menu.SetApplicationMenu(menu)` with only `AppMenu` (keeps ⌘Q, ⌘H) and `EditMenu` (⌘X/C/V/A/Z for the key field). It is invisible in an accessory app anyway.

### Pitfall 9: ⌘V into the key field is unproven natively
The mechanism is sound: Edit > Paste with `paste:` and a nil target goes to the first responder, and darwin `menuBindings` are not populated (Windows only), so nothing swallows ⌘V [VERIFIED: webview_window.go:1729-1755, menu_windows.go:146]. Whether AppKit consults `mainMenu` key equivalents for a key, **non-activating** panel of an inactive accessory app is MEDIUM: the user typed in Phase 1 but never pasted. **User check at native UAT.** Pre-planned contingency if it fails: in the key field's keydown, `metaKey && key === 'v'` → `preventDefault()` and insert `await Clipboard.Text()` from `@wailsio/runtime` (Go reads NSPasteboard). Do not ship both paths at once (double paste risk).

### Pitfall 10: Binding generator misses services wired under build tags
See Anti-Patterns. A symptom is generated bindings that silently lose a service, or server-mode calls failing with ReferenceError.

### Pitfall 11: The Phase 1 smoke script breaks when signatures change
`scripts/smoke-browser-loop.sh` calls `ActionService.Copy` with `args: []` and expects HTTP 422 "not supported in this build" [VERIFIED: smoke-browser-loop.sh:56-67], and checks `GET /sample.gif` returns 200 [VERIFIED: :69-71]. `Copy(ctx, gif)` turns the first into a TypeError, and dropping `public/sample.gif` breaks the second. Update the script in the same plan that changes the signature: call `GifService.Trending("finergifs","")` or a fake provider and assert JSON, and check a fakeapi media URL instead.

### Pitfall 12: Clips approval copy points at the wrong place
GIPHY docs say Clips access is requested by emailing **clips@giphy.com**; UI-SPEC C2 says "Request Clips access in the GIPHY developer dashboard". See Open Question 1.

### Pitfall 13: Server-mode window APIs are no-ops
In the browser, `PanelService.Hide()` does nothing and there is no `panel:show`. The UI must work without a show event (initial focus on mount), and must not depend on hide to reset state.

## Code Examples

### httpx classification (core)
```go
// internal/provider/httpx/httpx.go — Source: patterns from Go net/http docs; Wails message rule (bindings.go:424-436)
func GetJSON(ctx context.Context, c *http.Client, p provider.ID, u string, v any) (status int, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, &provider.Error{Kind: provider.HTTP, Provider: p} // malformed URL is a bug; never echo u
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return 0, ctx.Err() // superseded or caller deadline: not a provider failure
		}
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return 0, &provider.Error{Kind: provider.Timeout, Provider: p}
		}
		return 0, &provider.Error{Kind: provider.Network, Provider: p} // do NOT wrap err (*url.Error prints the URL)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return resp.StatusCode, &provider.Error{Kind: provider.Network, Provider: p}
	}
	switch {
	case resp.StatusCode == http.StatusTooManyRequests:
		return 429, &provider.Error{Kind: provider.RateLimited, Provider: p, RetryAfterSec: retryAfter(resp.Header.Get("Retry-After"), time.Now())}
	case resp.StatusCode < 200 || resp.StatusCode > 299:
		return resp.StatusCode, &provider.Error{Kind: provider.HTTP, Provider: p, Status: resp.StatusCode} // client remaps key statuses
	case len(bytes.TrimSpace(body)) == 0: // GIPHY: synthetic 200 with empty body is an error
		return resp.StatusCode, &provider.Error{Kind: provider.HTTP, Provider: p, Status: resp.StatusCode}
	}
	if err := json.Unmarshal(body, v); err != nil {
		return resp.StatusCode, &provider.Error{Kind: provider.HTTP, Provider: p, Status: resp.StatusCode}
	}
	return resp.StatusCode, nil
}

func retryAfter(h string, now time.Time) int {
	if s, err := strconv.Atoi(strings.TrimSpace(h)); err == nil {
		return max(s, 0)
	}
	if t, err := http.ParseTime(h); err == nil {
		return max(int(t.Sub(now).Seconds()+0.999), 0)
	}
	return 0
}
```
Each client then remaps: `if e.Kind == HTTP && slices.Contains(rejectStatuses, e.Status) { e.Kind = KeyRejected }` with GIPHY `{401, 403}` and Klipy `{401, 403, 404}`.

### Tolerant GIPHY int
```go
// Accepts "200", 200, "" (cosmetic fields only; DownloadURL/PreviewURL emptiness is a hard error in the mapper).
type flexInt int

func (f *flexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*f = flexInt(n)
	return nil
}
```

### Shell changes (Esc, close, menu)
```go
// internal/shell/shell.go (New) — Sources: webview_window.go:381-385, examples/hide-window/main.go:35-38, menu_darwin.go:144-152
s.panel = app.Window.NewWithOptions(application.WebviewWindowOptions{
	// ...unchanged...
	HideOnEscape:    false, // the page owns Esc: clear query first, then PanelService.Hide (UI-SPEC)
	HideOnFocusLost: true,
})
s.panel.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
	e.Cancel() // NSPanel Esc/cancelOperation or ⌘W must never destroy the only panel
	s.panel.Hide()
})
menu := application.NewMenu()
menu.AddRole(application.AppMenu)  // ⌘Q, ⌘H
menu.AddRole(application.EditMenu) // ⌘V/⌘C/⌘X/⌘A/⌘Z for the key field
app.Menu.SetApplicationMenu(menu)  // drops File(⌘W) View(⌘R, zoom) Window(⌘M)
```

### Example fixtures (use as `testdata/*.json` and fakeapi bodies)

**Klipy search page** [CITED: docs.klipy.com gifs-search-api.md response example; trimmed to sm and hd, other sizes have the same shape]
```json
{"result": true,
 "data": {"data": [{"id": 8041071659142944, "slug": "hello-hi-662", "title": "Hello",
   "file": {
     "hd": {"gif": {"url": "https://static.klipy.com/ii/935d7ab9d8c6202580a668421940ec81/14/af/um0L4dFH.gif", "width": 498, "height": 498, "size": 4001918},
            "webp": {"url": "https://static.klipy.com/ii/935d7ab9d8c6202580a668421940ec81/14/af/eUbp2uNc.webp", "width": 498, "height": 498, "size": 285228}},
     "sm": {"gif": {"url": "https://static.klipy.com/ii/935d7ab9d8c6202580a668421940ec81/14/af/y6iepZM7.gif", "width": 220, "height": 220, "size": 314884},
            "webp": {"url": "https://static.klipy.com/ii/935d7ab9d8c6202580a668421940ec81/14/af/SE72470w.webp", "width": 220, "height": 220, "size": 80118}}},
   "tags": [], "type": "gif", "blur_preview": "data:image/jpeg;base64,..."}],
  "current_page": 1, "per_page": 24, "has_next": true}}
```
**Klipy invalid key** (live, HTTP 404): `{"result":false,"errors":{"message":["The provided API key is invalid."]}}`

**GIPHY trending page** [constructed from schema + optional-settings bundle list + js-types; string numerics ASSUMED; replace with a live recording at UAT]
```json
{"data": [{"type": "gif", "id": "3o7aD2saalBwwftBIY", "slug": "cat-3o7aD2saalBwwftBIY", "title": "Cat GIF",
   "url": "https://giphy.com/gifs/cat-3o7aD2saalBwwftBIY", "rating": "g",
   "images": {
     "fixed_width": {"url": "https://media2.giphy.com/media/3o7aD2saalBwwftBIY/200w.gif?cid=abc&rid=200w.gif&ct=g", "width": "200", "height": "150", "size": "254321",
                     "webp": "https://media2.giphy.com/media/3o7aD2saalBwwftBIY/200w.webp?cid=abc&rid=200w.webp&ct=g", "webp_size": "98765"},
     "original": {"url": "https://media2.giphy.com/media/3o7aD2saalBwwftBIY/giphy.gif?cid=abc&rid=giphy.gif&ct=g", "width": "480", "height": "360", "size": "1543210", "frames": "40"}}}],
 "pagination": {"total_count": 5000, "count": 1, "offset": 0},
 "meta": {"status": 200, "msg": "OK", "response_id": "abc123"}}
```
**GIPHY error** (live, HTTP 401, header `retry-after: 0`): `{"data": [], "meta": {"status": 401, "msg": "Unauthorized", "response_id": ""}}`

**Finer Gifs page** [envelope from the site bundle; field values constructed]
```json
{"hits": {"found": 57, "start": 0},
 "results": [{"id": "1", "fields": {"name": "02x05-dwight-bears", "text": "Bears. Beets. Battlestar Galactica.", "fileid": "02x05-dwight-bears"}}]}
```
**Finer Gifs down** (live, HTTP 530, `text/plain`): `error code: 1033`

### Fixture server scenario words (fakeapi; same handler for `httptest` and the server build)
| Query contains | Response |
|---|---|
| `ratelimit` | 429 with `Retry-After: 120` (GIPHY/Klipy) |
| `noretry` | 429 with no Retry-After |
| `slow` | sleeps past the client timeout (server build sets the client timeout to 3 s via env so UAT stays fast) |
| `empty` | 200 with zero results |
| `boom` | 500 |
| `blank` | 200 with an empty body (GIPHY synthetic error) |
| (any other) | fixture page; `offset`/`page`/`start` honored so paging and "End of results" are reachable (for example 60 items total) |
| key starting `bad` | GIPHY 401 body above; Klipy 404 body above |
| env `FAKE_CLIPS=denied` | `/v1/clips/*` → 403 (state C2); default 200 |
| env `FAKE_FINER=down` | `/search` → 530 `error code: 1033` (state E5f); default fixtures |
Media URLs point at `http://127.0.0.1:<fakeport>/media/<n>.gif`, a handful of small valid GIFs from `testdata`, so the browser renders tiles and the copy pipeline's magic-byte check passes.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Extension calls the Raycast proxy (`gif-search.raycast.com`) with Tenor-shaped Klipy data | Native Klipy v1, key in path, `page/per_page/has_next`, `file.{hd,md,sm,xs}` | Klipy moved to the proxy 2026-07-01 (extension changelog) | Rebuild the mapper; never port `media_formats`/`next`/`pos` |
| Extension uses `preview_gif` for small tiles | `fixed_width` (200w). `preview_gif` is capped at 50 KB and shows only the first 1–2 seconds [CITED: GIPHY schema] | — | Tiles show the whole loop |
| Finer Gifs paging by `len == limit` | `hits.found`/`hits.start` from the response | Discovered in the site bundle today | Exact "End of results" |
| Wails `HideOnEscape` | Page-owned Esc plus a `WindowClosing` hook | UI-SPEC finding + this research | Layered Esc works and the panel can't be destroyed |

**Deprecated/outdated:** Tenor (shut down 2026-06-30); `golang.design/x/hotkey` (needs Accessibility; not used).

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | GIPHY rendition `width/height/size` are strings on the wire | Provider API Reference, Code Examples | Low: the tolerant decoder accepts both |
| A2 | `images.original.url` exists in GIPHY responses (the schema table omits `url` for `original`; js-types and the extension use it) | Provider API Reference | Medium: copy breaks for GIPHY. Mitigate with the live-key test at UAT; fallback `downsized_large` would need a bundle change |
| A3 | WKWebView on macOS 13+ animates WebP, so `sm.webp`/`fixed_width.webp` could replace GIF tiles to cut bandwidth about 4× | Provider API Reference | Low: Phase 2 recommends GIF tiles; WebP is an optional later optimization |
| A4 | Klipy never sends `type:"ad"` items to a desktop UA on a non-ad key, so dropping non-`gif` items never changes results in practice | Klipy section | Low-Medium: a policy grey area if ads appear; log the count so it is visible |
| A5 | NSPanel `cancelOperation:` closes the panel when Esc is not consumed | Pitfall 3 | Low: the `WindowClosing` hook + preventDefault make it moot either way |
| A6 | Main-menu key equivalents (Edit > Paste) fire for a key non-activating panel of an inactive accessory app | Pitfall 9 | Medium: ⌘V into the key field fails; contingency is documented; **user native check** |
| A7 | Download cap 50 MB and 45 s deadline suit GIPHY originals and Klipy hd | Pattern 10 | Low: tune after UAT |
| A8 | Klipy `content_filter` should be `medium` for parity with GIPHY PG-13 | Open Question 3 | Medium: unfiltered results in a public app if left unset |
| A9 | Klipy 429 uses status 429 (undocumented) | Klipy section | Low: other statuses fall to the generic HTTP error state |
| A10 | Search cache TTL 5 min (trending 10 min is locked) | Pattern 4 | Low |

## Open Questions (RESOLVED)

1. **UI-SPEC C2 copy vs GIPHY docs (Clips approval).**
   - Known: GIPHY says "Access to Clips endpoints is only available upon approval. Please reach out to clips@giphy.com to request permission." C2 says "Request Clips access in the GIPHY developer dashboard" and its primary button opens the dashboard.
   - Unclear: whether the dashboard has any Clips request flow.
   - Recommendation: change the C2 body to "…GIPHY has to approve it for Clips. Email clips@giphy.com to request access, then reopen Gifsmith." Keep `Open GIPHY Dashboard ↗` or swap it for a `mailto:` link. Needs a quick user confirmation because UI-SPEC is approved.
   - **RESOLVED (user, 2026-10-17):** C2 tells the user to email clips@giphy.com with a mailto button — see 02-CONTEXT.md Post-research decisions; implemented in 02-06.
2. **E5 copy for "HTTP 200" errors.** GIPHY's synthetic empty-200 and invalid JSON map to `kind: http` with status 200, which E5 would render as "returned an error (HTTP 200)". Recommendation: omit the "(HTTP n)" parenthetical when status is 2xx (a discretion-level copy tweak), or accept it as honest.
   - **RESOLVED:** omit "(HTTP n)" when status is 2xx — 02-CONTEXT.md Post-research decisions; 02-03.
3. **Klipy content filter.** CONTEXT fixes GIPHY at PG-13 but says nothing about Klipy. Recommendation: send `content_filter=medium` [ASSUMED] and confirm with the user; Phase 4 can expose it.
   - **RESOLVED (user):** content_filter=medium — 02-CONTEXT.md Post-research decisions; 02-04.
4. **K4 copy with unknown Retry-After.** K4 always contains `{n} minutes`. When unknown, use the E1 wording "later this hour". A trivial copy rule.
   - **RESOLVED:** "later this hour" when Retry-After is unknown — 02-CONTEXT.md Post-research decisions; 02-03.
5. **Will Finer Gifs return?** Unknown. The design treats it as optional; UAT can only show E5f against the real host unless it recovers. The fakeapi covers the success path.
   - **RESOLVED (by design):** provider optional; fakeapi covers the success path; UAT records the live status — 02-01/02-03.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | Backend | ✓ | go1.27.2 darwin/arm64 | — |
| Node | Vite, `node --test` | ✓ | v24.14.1 (TS type stripping on by default) | — |
| pnpm | Frontend deps | ✓ | 12.14.0 | — |
| wails3 CLI | Bindings, build | ✓ | v3.0.0-beta.30 | — |
| `/usr/bin/security` | go-keyring | ✓ | system | memory store (server build/tests) |
| dev-browser | Browser UAT | ✓ | `/opt/homebrew/bin/dev-browser` | — |
| lsof, curl | Smoke script | ✓ | system | — |
| golangci-lint | Lint (CLAUDE.md) | ✓ | `~/go/bin/golangci-lint` | — |
| api.giphy.com | GIPHY provider | ✓ reachable (401 keyless) | — | fakeapi |
| api.klipy.com | Klipy provider | ✓ reachable (404 bad key) | — | fakeapi |
| api.thefinergifs.club | Finer provider | ✗ HTTP 530 / 1033 | — | fakeapi; E5f state in real use |
| Live GIPHY / Klipy keys | Live verification | ✗ (by design) | — | user enters them at UAT via the setup card |

**Missing dependencies with no fallback:** none block execution.
**Missing with fallback:** live keys (fixture server plus user UAT), Finer Gifs API (fixture server; E5f state).

## Test Strategy

`workflow.nyquist_validation` is `false`, so this replaces the Validation Architecture section.

| Layer | What | How | Agent-runnable |
|-------|------|-----|----------------|
| Go unit | httpx classification (timeout, network, 429 + Retry-After forms, non-2xx, empty 200, bad JSON, cancel), **key-redaction tests** (Pitfall 1) | `httptest.NewServer(fakeapi.Handler(...))`, `go test -race ./internal/...` | ✓ |
| Go unit | Mappers for each provider from `testdata/*.json` (preview vs download rendition, title fallback, cursor/Next math incl. offset caps, Klipy slug id, Finer fileid validation, Klipy no-dedupe, non-Klipy same-page dedupe) | table tests | ✓ |
| Go unit | `search.Service`: TTLs (inject `now`), blockedUntil after 429, key-change invalidation | fake providers | ✓ |
| Go unit | `SettingsService.SaveKey`: K2/K3/K4 mapping per vendor (Klipy 404 → rejected), store written only on success, cache seeded, Clips probe state transitions | memory store + fakeapi | ✓ |
| Go unit | `ActionService.Copy`: allowlist rejection, non-GIF bytes rejected, size cap, CopyFile called with the staged path, older staging dirs removed only after success, no Go-side hide on Copy, busy lock | `platform.Fake` + fakeapi + `t.TempDir()` | ✓ |
| Go optional live | `TestLiveGIPHY` / `TestLiveKlipy`, skipped unless `GIFSMITH_LIVE_GIPHY_KEY` / `GIFSMITH_LIVE_KLIPY_KEY` is set; asserts the mapper yields non-empty Preview/Download URLs; never prints the key | the user runs it at UAT | user |
| TS unit | `lib/grid.ts` (arrows, clamps, ↓ from the last full row, load-more threshold), `lib/errors.ts` (kind → copy row) | `node --test frontend/src/lib/*.test.ts` | ✓ |
| Static | `go vet`, `gofmt -l`, `golangci-lint run`, `pnpm check`, `CGO_ENABLED=0 go build -tags server .` | | ✓ |
| Browser UAT | All UI states S1–S13, C1–C2 via scenario words; tab switching; paging to End of results; keyboard table; attribution and placeholder per tab; long-copy fit and contrast backstops | updated `scripts/smoke-browser-loop.sh` (add a `--serve` mode that keeps server + Vite + fakeapi running) + dev-browser at `http://127.0.0.1:9245` | ✓ |
| Native (user) | ⌘V into the key field (A6); Esc clears then hides; ⌘W does not kill the panel; Keychain save survives relaunch (`security find-generic-password -s com.kenscott.gifsmith -a giphy` shows the item exists, no `-w`); real GIPHY/Klipy keys; Clips probe outcome with a real key; Enter copies a real GIF and ⌘V pastes it animated; WebContent memory sane after paging | `wails3 task darwin:install:dev` | user |

Note tests need `allowImportingTsExtensions: true` + `noEmit: true` in `frontend/tsconfig.json` so `import { move } from './grid.ts'` passes `svelte-check` (Node requires the explicit `.ts` extension).

## Security Domain

`security_enforcement: true`, ASVS Level 1.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no (no user accounts; vendor keys are credentials, see V6/V8) | — |
| V3 Session Management | no | — |
| V4 Access Control | yes (bound surface) | Bound methods accept enum provider IDs and validated DTOs only; no path or URL parameters except the allowlisted `DownloadURL` |
| V5 Validation, Sanitization | yes | Key sanitizer (charset/length; blocks Klipy path injection); query truncation (GIPHY 50 runes); provider ID enum check; download URL scheme + host allowlist; Finer `fileid` regex; GIF magic bytes; response size caps (8 MB JSON, 50 MB media) |
| V6 Stored Cryptography | yes | macOS Keychain via go-keyring; no custom crypto |
| V7 Error Handling and Logging | yes | `provider.Error` contains no URL/body/key; transport errors never wrapped; cancellations not logged; redaction unit tests |
| V8 Data Protection | yes | Keys never returned to the webview (`configured: bool` only), never in prefs JSON, logs or errors; server build uses a memory store |
| V12 Files and Resources | yes | Staging path computed in Go under `UserCacheDir`; hashed directory + sanitized slug; `.part` + rename; eager deletion of older staged files |
| V14 Configuration | partial | Optional CSP meta (`img-src` limited to provider CDNs + `data:`) applied at build time only; core control is "no `{@html}`" |

### Known Threat Patterns

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Key disclosure via `*url.Error` text in UI/logs | Information disclosure | Pattern 1 + redaction tests |
| Webview-induced arbitrary fetch (SSRF-like) via a crafted `DownloadURL` | Tampering / Elevation | https + per-provider host allowlist in Go |
| Path traversal via `fileid` or title | Tampering | Regex-validated id; slugified title; join under staging dir and check prefix |
| HTML/JSON error page saved as `.gif` and pasted into chats | Tampering | Status 200 + magic bytes before rename |
| Oversized download / memory exhaustion | Denial of service | `io.LimitReader` caps; ctx deadline |
| XSS via provider titles | Tampering | Svelte text interpolation only; no `{@html}` |
| Key injection into the Klipy path (`/`, `?`) | Tampering | Sanitizer + `url.PathEscape` |
| Retry storms burning the shared hourly quota | Denial of service (self) | `blockedUntil` after 429; debounce; cache |

## Sources

### Primary (HIGH confidence)
- Live keyless probes 2026-10-17 17:36–17:38 UTC: `api.thefinergifs.club/search` (530, `error code: 1033`), `thefinergifs.club` (200), `media.thefinergifs.club` (404 XML), `api.giphy.com/v1/gifs/trending` and `/v1/clips/trending` without a key or with a bogus key (401, `retry-after: 0`, JSON meta), `api.klipy.com/api/v1/<bogus>/gifs/{search,trending}` (404 JSON), unknown sub-path (403 JSON).
- The Finer Gifs site bundle `/_nuxt/app.3d29e032e6eefe5f040f.js` (request params, `results`/`hits.found`/`hits.start`), `/_nuxt/pages/index...js` (media URL `https://media.thefinergifs.club/<fileid>.gif`).
- Wails v3 beta.30 module source: `pkg/application/bindings.go` (CallError, message = `err.Error()`), `services.go`, `application_options.go` (MarshalError), `messageprocessor_call.go` (ctx cancel), `webview_window.go` (HideOnEscape binding, processKeyBinding, WindowClosing destroy), `webview_panel_darwin.m` (sendEvent/cancelOperation), `webview_window_darwin.m` (windowShouldClose), `application_darwin.go` + `menu_darwin.go` (default menu), `menuitem_roles.go` + `menuitem_selectors_darwin.go` (Paste → `paste:`), `examples/hide-window/main.go`.
- `@wailsio/runtime` 3.0.0-beta.30 `dist/calls.js`, `cancellable.js`, `runtime.js`, `index.js`, `browser.js`, `clipboard.js`.
- `github.com/zalando/go-keyring@v0.2.8` `keyring.go`, `keyring_darwin.go`, `keyring_mock.go`; proxy.golang.org `@latest` for go-keyring, x/sync, x/time.
- `@tanstack/svelte-virtual@3.13.47` tarball `dist/index.js`, `index.d.ts`; npm registry metadata; gsd package-legitimacy seam.
- Repo files read this session: `main.go`, `go.mod`, `internal/shell/shell.go`, `internal/services/actions.go` + `actions_test.go`, `internal/platform/*.go`, `scripts/smoke-browser-loop.sh`, `frontend/{package.json,vite.config.ts,tsconfig.json,index.html,src/App.svelte}`, generated bindings.

### Secondary (MEDIUM confidence; official docs via fetch)
- developers.giphy.com: `/docs/api/` (response codes, beta 100/hr, empty-200 rule, client-side and caching rules), `/docs/api/endpoint/` (params, caps), `/docs/api/schema/` (renditions, pagination, meta), `/docs/optional-settings/` (bundles), `/docs/clips/endpoint/` (paths, rating defaults, approval).
- docs.klipy.com: `gifs-search-api.md` (fetched live today), plus the earlier-fetched `gifs-trending-api.md`, `llms-full.txt` (advertisements, browser-UA rule, Testing 100/hr, share trigger, items by slug), `integration-requirements.md`.

### Tertiary (LOW confidence)
- Web search: GIPHY numeric fields as strings (community samples); no source found for the Clips unapproved-key status code.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH. Versions and behavior read from source and registries.
- Provider contracts: HIGH for error shapes (live), MEDIUM for success shapes (docs + constructed fixtures; GIPHY needs the live-key test).
- Architecture: HIGH. Follows Phase 1 patterns and verified Wails mechanics.
- Native panel behaviors (⌘V, Esc): MEDIUM. Mechanism verified in source; needs the user's native check.
- Pitfalls: HIGH. Most are backed by a live probe or a source line.

**Research date:** 2026-10-17
**Valid until:** 2026-10-24 for Finer Gifs status and Wails/TanStack nightly-style versions; 2026-11-16 for vendor API shapes.
