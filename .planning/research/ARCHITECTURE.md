# Architecture Research

**Domain:** macOS menu-bar launcher utility (global hotkey -> floating search panel -> copy/paste result into frontmost app), Go + Wails v3 + web frontend
**Project:** Gifsmith
**Researched:** 2026-10-17
**Confidence:** MEDIUM-HIGH. Wails v3 capabilities verified against Wails `master` source (HIGH that the APIs exist). Behavior of a non-activating NSPanel hosting a WKWebView text input is documented in a Wails example but not proven for our flow (MEDIUM, spike required). Klipy and GIPHY Clips API details are LOW (see Gaps).

## Headline Findings (read these first)

1. **Wails v3 already ships the whole "launcher shell".** Verified in source (v3.0.0-beta.30, published 2026-10-15; `go 1.25`): `MacWindowClassPanel` with `NonActivating`, `CollectionBehavior` (all Spaces + full-screen auxiliary), `WindowLevel`, `ActivationPolicyAccessory`, `app.GlobalShortcut.Register(...)` (Carbon `RegisterEventHotKey`, so **no Accessibility permission needed for the hotkey**), `SystemTray`, `HideOnFocusLost`, `HideOnEscape`, SMAppService autostart, and an official `v3/examples/spotlight`. **Our cgo layer shrinks to four things: file-URL pasteboard write, CGEvent Cmd+V, Accessibility check, frontmost-app query.** Panel, hotkey, tray, and launch-at-login are Wails calls, not cgo. Use v3 beta, pinned (`v3.0.0-beta.30`), not v2 (v2 has no panel class, no global shortcut API).
2. **Provider HTTP lives in Go. Confirmed as the right call, and now stronger:** Klipy puts the API key **in the URL path** (`/api/v1/<KEY>/gifs/search`) and GIPHY puts it in a query param. Go's `*url.Error` stringifies the full URL, so keys leak into toasts/logs unless errors are redacted at one choke point (`httpx`).
3. **The extension's provider contract does not map 1:1 onto the real APIs.** The Raycast proxy normalized everything (Tenor-shaped Klipy, `?ids=` lookup for every provider, offset-or-cursor `APIOpt`). The real APIs differ: Klipy is page-based (`page`, `per_page` 8-50, `has_next`), has no documented batch-by-ids endpoint (single GET by slug), and the GIPHY Clips API requires special approval with no documented by-ids route. Therefore: **opaque string cursor in the Go interface, and favorites/recents store a metadata snapshot, not just IDs.**
4. **Dev-browser testing is best served by Wails server mode** (`go build -tags server`, pure Go, no cgo, bindings + events over HTTP/WebSocket) combined with a **fake Platform**, not by pointing a browser at `wails3 dev`'s Vite port (which cannot reach the Go bindings by itself). Details in "Testing Strategy".

## Standard Architecture

### System Overview

```
┌──────────────────────────── WebView (WKWebView, one Vite bundle, 2 windows) ────────────────────────────┐
│  Panel window (/ route)                                   Settings window (#/settings route)             │
│  ┌──────────┐ ┌─────────────┐ ┌───────────┐ ┌──────────┐   ┌──────────────────────────────────────┐    │
│  │SearchBar │ │ProviderSwitch│ │ GifGrid   │ │ActionMenu│   │ Keys / Hotkey / Grid / Default action │    │
│  └────┬─────┘ └──────┬──────┘ └─────┬─────┘ └────┬─────┘   └──────────────────┬───────────────────┘    │
│       └──────────────┴──────┬───────┴────────────┘                            │                         │
│                     frontend/src/api  (facade over generated bindings; mockable)                         │
└─────────────────────────────────┬──────────────────────────────┬─────────────────────────────────────────┘
                 Wails bindings (JSON, cancellable, ctx-aware)     │  Wails events: panel:show, library:changed, settings:changed
┌─────────────────────────────────┴──────────────────────────────┴─────────────────────────────────────────┐
│ Go: Wails Services (THIN adapters; no business logic)                                                    │
│  ┌────────────┐ ┌─────────────┐ ┌───────────────┐ ┌─────────────────┐ ┌───────────┐                     │
│  │GifService  │ │ActionService│ │LibraryService │ │SettingsService  │ │PanelCtl   │                     │
│  └─────┬──────┘ └──────┬──────┘ └──────┬────────┘ └───────┬─────────┘ └─────┬─────┘                     │
├────────┴───────────────┴───────────────┴──────────────────┴─────────────────┴───────────────────────────┤
│ Go: domain packages (pure, unit-testable, no Wails import)                                               │
│  provider/{giphy,klipy,finergifs}+httpx   action (orchestrator)   library   settings+secrets   imaging   │
│  searchcache (TTL+singleflight)           gifcache (files)        gif (model, slugify, dedupe)           │
├──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Go: platform.Platform interface ── darwin (cgo, build tag darwin && !server)  |  stub/fake (everything else)│
│  CopyFile · CopyText · PasteShortcut · FrontmostApp · AccessibilityTrusted · OpenAccessibilitySettings · Reveal│
├──────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Storage: ~/Library/Application Support/Gifsmith/{settings.json,library.json}                             │
│          ~/Library/Caches/Gifsmith/{gifs/ (favorites cache), staging/ (clipboard files), trending.json}   │
│          macOS Keychain (API keys, via SecretStore)                                                      │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────┘
 Network egress (Go only): api.giphy.com · api.klipy.com · api.thefinergifs.club · GIF CDNs (downloads)
 Network egress (WebView only): GIF CDN preview <img> loads. Never an API call, never a key.
```

### Component Responsibilities

| Component | Responsibility | Talks to | Implementation |
|-----------|----------------|----------|----------------|
| `internal/gif` | `Gif` DTO (with `Provider` field), slugify, dedupe, strip-query, display-name/extension rules | everything | Pure Go; port of `models/gif.ts`, `dedupe.ts`, `stripQParams.ts`, name logic from `cachedGifs.ts` |
| `internal/provider` | `Provider` interface, `Registry`, per-provider clients + mappers, `httpx` (timeout, typed errors, key redaction) | `httpx` -> network; `settings` for keys (via `KeySource` func) | One package per provider; mappers are table-tested against recorded JSON fixtures |
| `internal/searchcache` | In-memory TTL cache + `singleflight` over Search/Trending (trending ~1h, search ~5m) | wraps `Registry` | Protects the low GIPHY free-key hourly cap; optional trending disk persistence |
| `internal/gifcache` | Content files: favorites cache (`sha1(provider:id)+ext`), staging dir with display names, download-to-folder with exclusive-create | `net/http`, `library` (membership re-check) | Port of `cachedGifs.ts`, `resolveGifFile.ts`, `downloadFile.ts` |
| `internal/imaging` | Square center-crop of animated GIF | `image/gif` | Pure Go replaces the extension's `/usr/bin/swift` script; keeps frame delays + loop count |
| `internal/library` | Favorites + recents per provider; snapshot store; usage stats for frecency | `library.json` | Mutex + atomic write (temp+rename); emits `library:changed` |
| `internal/settings` | Typed settings (JSON), `SecretStore` interface (Keychain impl) | `settings.json`, Keychain | `zalando/go-keyring` v0.2.8 behind `SecretStore` (Windows Credential Manager comes free later) |
| `internal/action` | The action orchestrator: resolve file -> platform op -> record usage -> hide panel -> result | `gifcache`, `imaging`, `platform`, `library`, `PanelCtl` | **All of `GifActions.tsx`'s behavior, minus presentation.** Single place that guarantees "failed op never records usage" |
| `internal/platform` | `Platform` interface + darwin cgo impl + stub/fake | AppKit/CoreGraphics/ApplicationServices | The *only* package with cgo |
| `internal/shell` (panel, hotkey, tray) | Window lifecycle, hotkey (re)registration, tray menu, settings window | Wails `application` API, `platform.FrontmostApp` | Plain Go over Wails APIs; no cgo |
| `internal/services` | Wails-bound adapters; DTO shaping; typed error marshaling | domain packages | ~20 lines per method; this is the only package whose exported methods the frontend sees |
| `frontend/src/api` | One facade over generated bindings; swappable for a mock via Vite alias | generated `bindings/` | Lets dev-browser run with fixtures or the real server-mode backend |
| `frontend/src/features/*` | search, library, details, settings views; keyboard model; grid | `api` | React + TS + Vite (see Frontend section) |

## Recommended Project Structure

```
gifsmith/
├── main.go                         # wiring only: build Platform, stores, providers, services, shell; app.Run()
├── build/                          # wails3 build assets, Info.plist (LSUIElement), entitlements, icons
├── internal/
│   ├── gif/                        # model.go, slug.go, dedupe.go, names.go (+ tests)
│   ├── provider/
│   │   ├── provider.go             # interface, Page, PageReq, Info, errors, Registry
│   │   ├── httpx/                  # GetJSON(ctx,u,meta) -> typed errors, redacts keys, 15s timeout
│   │   ├── giphy/                  # client.go (gifs+clips), map.go, testdata/*.json
│   │   ├── klipy/                  # client.go, map.go, testdata/*.json
│   │   └── finergifs/              # client.go, map.go, testdata/*.json
│   ├── searchcache/
│   ├── gifcache/                   # cache.go, resolve.go, download.go, staging.go
│   ├── imaging/                    # square.go
│   ├── library/                    # store.go, frecency.go
│   ├── settings/                   # settings.go, secrets.go (SecretStore + keyring impl + memory fake)
│   ├── action/                     # actions.go (enum + Perform), markdown.go
│   ├── platform/
│   │   ├── platform.go             # interface + types (no build tag)
│   │   ├── platform_darwin.go      # //go:build darwin && !server  (cgo)
│   │   ├── platform_darwin.m/.h    # ObjC bodies
│   │   ├── platform_stub.go        # //go:build !darwin || server  (returns ErrUnsupported)
│   │   └── fake.go                 # recording fake for tests + dev-browser (no build tag)
│   ├── shell/                      # panel.go, hotkey.go, tray.go, settingswin.go
│   └── services/                   # gif.go, action.go, library.go, settings.go (+ dev.go under `dev` tag)
├── frontend/
│   ├── bindings/                   # GENERATED by `wails3 generate bindings` (git-ignored or committed, pick one)
│   ├── src/
│   │   ├── api/                    # client.ts (facade), mock.ts (fixtures), events.ts
│   │   ├── features/{search,library,details,settings}/
│   │   ├── components/             # Grid, GridItem, ActionMenu, Toast, ProviderIcon
│   │   ├── keyboard/               # keymap, default-action ordering
│   │   └── main.tsx                # route: "/" panel, "#/settings" settings
│   └── vite.config.ts              # port 9245 strictPort; optional /wails proxy; VITE_MOCK alias
└── LICENSES/ or NOTICE             # attribution to original extension authors (MIT)
```

### Structure Rationale

- **`internal/*` domain packages never import Wails.** That is what makes them testable without a window and lets server mode run the same code. Only `shell/` and `main.go` import `wails/v3/pkg/application`.
- **`platform` is the only cgo package and has three implementations** (darwin, stub, fake). The fake is not just for unit tests: it is what makes dev-browser UAT able to assert "Enter put file X on the pasteboard" without a real pasteboard.
- **`services/` is deliberately boring.** Wails reflection-binds exported methods; keeping them as thin adapters means domain refactors never churn the generated TypeScript.
- **One frontend bundle, two windows** (panel + settings) selected by route. One Vite build, one set of bindings.

## Key Contracts

### Provider interface (Go)

```go
// internal/provider/provider.go
type ID string // "giphy" | "giphy-clips" | "klipy" | "finergifs"

type PageReq struct {
    Limit  int
    Cursor string // opaque; "" = first page. Provider encodes its own ("40" offset, "3" page, ...)
}
type Page struct {
    Items []gif.Gif
    Next  string // opaque; "" = no more results
}
type Info struct {
    ID         ID
    Title      string // "GIPHY GIFs", "GIPHY Clips", "Klipy", "Finer Gifs Club"
    NeedsKey   bool
    HasTrending bool  // Finer Gifs: false (extension returns [] for trending)
    CanLookup  bool   // batch lookup supported natively
}
type Provider interface {
    Info() Info
    Search(ctx context.Context, q string, r PageReq) (Page, error)
    Trending(ctx context.Context, r PageReq) (Page, error)
    Lookup(ctx context.Context, ids []string) ([]gif.Gif, error) // best-effort; may fan out per id
}
```

Decisions baked in (and why they differ from `IGifAPI` in `src/models/gif.ts`):

| Extension | Gifsmith | Reason |
|-----------|----------|--------|
| `APIOpt {offset, limit, next}` plus `useSearchAPI` computing `offset: page*limit` *and* passing `next` | Single opaque `Cursor`/`Next` string | The extension carries two paging models and sniffs `next === undefined` to guess. Real APIs are three different models (GIPHY offset+`pagination.total_count`, Klipy page+`has_next`, Finer CloudSearch `start`+`size`). The frontend must not know which. |
| `hasMore = next === undefined ? results.length === limit : next !== ""` | Provider sets `Next` (GIPHY: offset+len < total_count; Klipy: `has_next`; Finer: `len == limit` heuristic, since the response has no total) | Paging logic lives next to the API that defines it. |
| `getServiceFromUrl(gif)` URL-sniffing (`includes("giphy.com/gifs")`) | `Gif.Provider` field set by the mapper | The sniffing is fragile (it returns `null` for any unfamiliar CDN host, which silently disables caching and recents). Do not port it. |
| `gifs(ids)` required for all providers | `Lookup` best-effort + **snapshot in library** | Klipy has no documented batch lookup; GIPHY Clips has no documented by-ids route. Rendering favorites must not depend on it. |
| Relative-time strings baked into labels (`formatRelative`) | `CreatedAt` timestamp in DTO; frontend formats with `Intl.RelativeTimeFormat` | Presentation belongs in the UI; stale strings get cached. |
| Provider attribution = asset filename string | Frontend maps `Provider` -> logo | Same reason. |

`KeySource`: each keyed provider is constructed with `func() (string, error)` that reads the in-memory-cached secret, so a key saved in Settings applies on the next request with no restart and no provider rebuild. A missing key returns typed `ErrKeyMissing{Provider}`.

**Typed errors across the binding boundary.** Wails marshals returned `error` to a string by default; use `ServiceOptions.MarshalError` (verified in `services.go`) to emit JSON `{kind: "key_missing"|"rate_limited"|"network"|"timeout"|"http", provider, message}`. The frontend renders "Add your GIPHY key in Settings" vs. "Rate limited, retry in 12 min" differently. This replaces the extension's single `ProviderRequestError`.

**Ads.** Klipy results can contain ad items (the Kotlin client separates content from ad items). The Klipy mapper must filter non-GIF items. (LOW confidence on exact field; verify with a real response.)

### Platform interface (Go)

```go
// internal/platform/platform.go
type App struct{ BundleID, Name string; PID int }

type Platform interface {
    CopyFile(path string) error               // NSPasteboard: file URL (public.file-url)
    CopyText(s string) error                  // NSPasteboard string (own impl; do not depend on Wails Clipboard)
    PasteShortcut() error                     // CGEvent Cmd+V to HID tap. Requires Accessibility.
    FrontmostApp() (App, error)               // NSWorkspace.frontmostApplication
    AccessibilityTrusted(prompt bool) bool    // AXIsProcessTrustedWithOptions
    OpenAccessibilitySettings() error         // x-apple.systempreferences:...Privacy_Accessibility
    Reveal(path string) error                 // Finder reveal (can be `open -R`)
}
var ErrUnsupported = errors.New("not supported on this platform")
```

Notes:
- **Panel, hotkey, tray, launch-at-login are deliberately NOT in this interface**; Wails already abstracts them across macOS/Windows/Linux. Windows later = one new `platform_windows.go` (`CF_HDROP` clipboard, `SendInput` Ctrl+V, foreground window) plus nothing else.
- Wails' own `Clipboard` is text-only (`SetText`/`Text`), so file copy needs our cgo regardless. Use our `CopyText` too, for one consistent pasteboard path. (Wails beta.30 notes mention a fix for "activation-gated clipboard writes failing"; avoiding Wails' clipboard sidesteps that class of bug.)
- Paste = `CopyFile|CopyText` then `PasteShortcut`. The action layer owns sequencing; the platform layer stays dumb.
- **Square GIF stays out of `Platform`** (it is pure Go in `imaging`). The extension gated it on `process.platform === "darwin"` only because it shelled out to Swift.

### Wails services (bound surface, what the frontend sees)

```go
type GifService struct{ ... }      // Search(ctx, provider, q, cursor, limit) (Page, error)
                                   // Trending(ctx, provider, cursor, limit) (Page, error)
                                   // Providers() []ProviderInfo   (title, needsKey, configured, hasTrending)
type ActionService struct{ ... }   // Perform(ctx, action Action, g gif.Gif) (ActionResult, error)
                                   // Action: copyFile|pasteFile|copyLink|pasteLink|copyMarkdown|pasteMarkdown|
                                   //         copySquare|pasteSquare|copyPageLink|openInBrowser|download
type LibraryService struct{ ... }  // List(kind) ([]ProviderGroup, error)  // kind: favs|recent, partial failures reported per group
                                   // Favorite(g) / Unfavorite(g) / RemoveRecent(g) / Saved(provider) (ids)
type SettingsService struct{ ... } // Get() Settings   (keys appear only as {configured bool})
                                   // Update(patch) ; SetAPIKey(provider, key) ; ClearAPIKey(provider)
                                   // SetHotkey(accel) error ; AccessibilityStatus(prompt bool) bool ; OpenAccessibilitySettings()
type PanelService struct{ ... }    // Hide()   (everything else is Go-initiated)
```

- Methods take `ctx context.Context` first (verified: Wails binds `needsContext` when the first param is `context.Context`, and the frontend promise is cancellable). **That gives debounce-cancellation for free**: typing a new query cancels the in-flight HTTP request in Go.
- `ActionService.Perform` receives the `Gif` DTO from the frontend (it already has it) and **validates `DownloadURL`** (https scheme + host allowlist per provider) before fetching. Cheap, stateless, and the webview can never induce an arbitrary-URL fetch from Go.
- **Keys never cross to the webview.** `Settings.Get()` returns only `configured`. `SetAPIKey` is write-only.

### Events (Go -> frontend)

| Event | Payload | Frontend reaction |
|-------|---------|-------------------|
| `panel:show` | `{frontmostApp, accessibilityTrusted}` | Focus + select input, refresh library sections, resume or reset query by idle-time rule |
| `panel:hide` | none | Pause animations (drop `<img src>` or `display:none` grid); see Pitfalls |
| `library:changed` | `{provider}` | Invalidate library queries |
| `settings:changed` | Settings (no secrets) | Re-read grid size, default action, etc.; re-fetch if locale or key changed |

## Data Flow

### Search / trending (read path)

```
keystroke -> debounce 250ms (frontend) -> api.search(provider,q,cursor) [cancellable promise]
  -> GifService -> searchcache (TTL hit? singleflight) -> provider.Search(ctx) -> httpx.GetJSON -> network
  -> mapper -> []gif.Gif (+Provider) -> dedupe -> Page{Items,Next}
  -> frontend infinite query appends pages; <img> loads small_preview_gif_url directly from CDN
```
Empty `q` = Trending (cached ~1h). Provider switch changes the query key only.

### Action (write path), e.g. Enter on "Paste GIF"

```
Enter -> api.perform("pasteFile", gif)
  -> ActionService -> action.Perform:
       1. frontApp := platform.FrontmostApp()         (captured at panel:show, re-checked here)
       2. path := gifcache.Resolve(gif)               favorites cache hit -> stage copy named slug.gif
                                                      miss -> download -> stage; cache only if still favorite
       3. platform.CopyFile(path)
       4. panel.Hide()                                 (before the keystroke)
       5. sleep ~60-100ms; verify FrontmostApp == frontApp; platform.PasteShortcut()
       6. library.RecordUse(gif)                       ONLY on success of 1-5
  <- ActionResult{title, detail(path), revealable} -> frontend toast (or tray/notification if panel hidden)
```
**Copy** is steps 1-4 minus 5, and works without Accessibility. **Paste** first checks `AccessibilityTrusted(false)`; if false, it degrades to copy and returns `ActionResult{needsAccessibility:true}` so the UI can offer onboarding instead of failing silently.

### Hotkey -> panel lifecycle

```
startup:  ActivationPolicyAccessory; create panel window ONCE, Hidden:true (webview stays warm = instant show)
          GlobalShortcut.Register(settings.Hotkey, shell.Toggle)  (allowed before Run; deferred bind)
hotkey:   Toggle():
            visible?  -> Hide()
            else      -> app := platform.FrontmostApp(); center on screen under cursor;
                         panel.Show(); emit panel:show{app, trusted}
Esc / click-away: HideOnEscape / HideOnFocusLost (Wails options) -> emits panel:hide
action done: ActionService calls shell.Hide()
Tray:     "Open Gifsmith" (Toggle), "Settings..." (settings window), "Launch at login", "Quit"
Settings: separate regular NSWindow (NOT in the panel) - see Anti-Patterns
Hotkey change: SettingsService.SetHotkey -> Unregister(old); Register(new); on error re-Register(old) and return typed error
```

Panel window options (from the verified Wails API; mirror `examples/spotlight`):
`Frameless, DisableResize, Hidden, HideOnEscape, HideOnFocusLost, Mac{WindowClass: MacWindowClassPanel, PanelPreferences{NonActivating: true}, CollectionBehavior: CanJoinAllSpaces|FullScreenAuxiliary, WindowLevel: Floating (or screenSaver if floating fails over full-screen apps), Backdrop: Translucent}`. **Why NonActivating is the key decision:** the previously frontmost app *stays active*, so there is no "capture frontmost app and re-activate it" dance before pasting; the Cmd+V goes to the app that never lost focus. If the spike shows Wails' `Show()` still activates the app, fall back to capture + `NSRunningApplication.activate` in `Platform` (add `ActivateApp(App)`).

### Settings and secrets

```
Settings window -> SettingsService.SetAPIKey(p, key) -> SecretStore.Set(service="Gifsmith", user=p, key) [Keychain]
                -> in-memory key cache updated -> emit settings:changed (no secret)
Provider request -> KeySource() -> in-memory cache (Keychain read only at startup/miss)
Settings.json: everything non-secret (hotkey, grid sizes, maxResults, defaultAction, downloadPath,
               hideFilename, giphyLocale, klipyLocale, lastProvider, launchAtLogin)
```
`go-keyring` on macOS shells out to `/usr/bin/security` (secret via stdin on write; `-w` stdout on read; no cgo, no argv exposure on write). Acceptable and portable; note that Keychain ACL will name `security`, not Gifsmith. If a "Gifsmith wants to use your confidential information" prompt story matters later, swap `SecretStore` to a `SecItem*` cgo impl without touching callers. Keep the interface; do not pre-build the cgo version.

### Favorites / recents storage

```jsonc
// library.json  (version field for migrations)
{ "version": 1,
  "providers": {
    "giphy": {
      "favs":   [ { "id": "abc", "snap": { /* gif.Gif */ }, "added": "2026-10-17T..." } ],   // append
      "recent": [ { "id": "xyz", "snap": { ... }, "uses": 3, "last": "..." } ]                // prepend, cap ~200
    }, "klipy": {...}, "finergifs": {...}, "giphy-clips": {...} } }
```
- **Semantics preserved from the extension:** IDs unique only within a provider; recents prepend + move-to-front; favorites append; a GIF's membership is checked per provider.
- **Added:** `snap` (denormalized `Gif`) so Favorites/Recents render instantly and offline and never fail because `Lookup` is unavailable or a provider is down. `Lookup` becomes an optional background refresh. This is a deliberate deviation from "ID lists only"; it is justified by Klipy/Clips lookup gaps and by the extension's own bug class (its "provider outage empties favorites" fix in the 2026-09-17 changelog).
- **Fix an extension bug:** `save(..., "recent")` never truncates, so recents grow forever. Cap (e.g. 200) on write.
- **Frecency:** the extension used Raycast's `useFrecencySorting` for local sections. Rebuild minimally: `uses` + `last` per entry, score in `library/frecency.go`. Optional for MVP; plain recency/append order is acceptable first.
- Concurrency: one `sync.Mutex`, read-modify-write, atomic temp+rename. `IsSavedNow` semantic (read membership *inside* the lock at decision time, not a stale captured flag) is required by `gifcache` to avoid re-creating a cache entry after un-favoriting.

### GIF file cache and download

| Concern | Design | Extension origin |
|---------|--------|------------------|
| Cache key | `sha1(provider + ":" + id) + ext` (ext from `download_name`, `.gif` if empty; keep `.mp4` for Clips) | `getCacheKey` |
| What is cached | **Only favorites** (cache-on-favorite, evict on unfavorite incl. all extension variants by digest prefix) | `cacheIfStillFavorite`, `removeGifFromCache` |
| Staging | Per-action temp dir `Caches/Gifsmith/staging/<rand>/<displayName>`; clipboard file URL points here. Display name honors `hideFilename` (`gif.gif`) | `tempy.write({name})`, `getDisplayName` |
| Staging cleanup | On startup + delete entries older than ~24h. **Never delete right after paste**: the pasteboard URL must stay valid until the clipboard changes | new |
| Corrupt entry | Non-regular-file at cache path = miss, delete it | `getGifFromCache` |
| Cache vanishes mid-read | Fall back to a single download | test "cache file disappearing between lookup and read" |
| Download to folder | `os.OpenFile(O_CREATE|O_EXCL)` with `name (1).ext` counter; delete the claimed file on failed write | `writeToNewFile` |
| Previews in grid | **Not cached in Go.** WebView loads CDN URLs directly (HTTP cache handles repeats). Revisit only if offline favorites matter | new |
| Where | `~/Library/Caches/Gifsmith` (purgeable is fine; everything re-downloads) | `environment.supportPath` |

`resolveGifFile`'s invariant "Caching is an optimisation; a cache failure must never fail the user's action" is preserved.

## Architectural Patterns

### Pattern 1: Thin services over pure domain packages
**What:** Wails-bound structs only translate DTOs and errors; logic lives in `internal/*`.
**When:** Always. **Trade-off:** one extra layer of tiny methods; buys testability, server mode, and stable generated TS.

### Pattern 2: Orchestrator guarantees side-effect ordering
**What:** `action.Perform` is the only place that sequences resolve -> pasteboard -> hide -> keystroke -> record usage, with "failed op never records usage or reports success".
**When:** Every action. **Why:** the extension's `GifActions.tsx` (393 lines) interleaves toast, clipboard, usage tracking and window close in each handler, and its test file (`gifClipboard.test.mjs`) is mostly regression tests for that interleaving. Centralize once, test once.

### Pattern 3: Interface + build tags for native glue; fake for everything else
```go
//go:build darwin && !server
package platform
// #cgo LDFLAGS: -framework AppKit -framework ApplicationServices -framework CoreGraphics
// #include "platform_darwin.h"
import "C"
func New() Platform { return darwin{} }
```
```go
//go:build !darwin || server
package platform
func New() Platform { return stub{} }   // all methods return ErrUnsupported
```
`main.go` selects `platform.New()` or, when `GIFSMITH_PLATFORM=fake`, `platform.NewFake()`. Note that `-tags server` must exclude our cgo too (Wails' server mode is the no-cgo path).

### Pattern 4: Opaque cursor + capability flags
Frontend never branches on provider for paging; it branches on `Info.HasTrending` / `NeedsKey` only.

### Pattern 5: Frontend facade with mock seam
`frontend/src/api/client.ts` is the only importer of generated bindings. `VITE_MOCK=1` aliases it to `mock.ts` (fixtures), giving deterministic visual tests and a UI-first workflow before the Go side exists.

## Anti-Patterns to Avoid

| Anti-Pattern | Why bad | Instead |
|--------------|---------|---------|
| Provider calls from the frontend | Keys in webview memory/devtools; CORS/CSP exposure; GIPHY key rate cap hard to protect with a cache | Go `provider` + `searchcache` |
| Settings inside the panel | `HideOnFocusLost` kills the panel when the user switches to a password manager to copy an API key | Separate regular Settings window |
| cgo for panel/hotkey/tray | Wails v3 already does it; every cgo line is a Windows-port tax | Use Wails APIs; cgo only for pasteboard/CGEvent/AX/NSWorkspace |
| Keys in error strings or logs | `*url.Error` includes URL; Klipy key is in the path | Redact in `httpx` only; never format a raw URL into an error |
| Frontend decides when to record usage | Double-record, or record on failure | `action.Perform` records on success only |
| URL-sniffing to infer provider | Silent null on new CDN host | `Gif.Provider` |
| Keeping per-provider paging logic in the UI | Three paging models leak | Opaque cursor |
| Storing only IDs and resolving through the network | Offline/outage/unsupported-lookup breaks favorites | Snapshot + optional refresh |
| Porting the Swift crop script | Needs Xcode CLT at runtime; breaks on a clean Mac | Pure Go `image/gif` |
| Calling Wails' `Clipboard` for files | Text-only | Own `platform.CopyFile` |

## Extension -> Gifsmith Port Map

Legend: **PORT** = translate logic nearly 1:1; **REBUILD** = behavior kept, implementation new; **DROP** = do not port.

| Extension file (src/) | Lines | Lands in | Treatment / notes |
|-----------------------|-------|----------|-------------------|
| `models/gif.ts` (`IGif`, `slugify`, `renderGifMarkdownDetails`, `createGifLookupUrl`) | 60 | `internal/gif`, frontend `details/` | PORT `IGif` -> `Gif` (+`Provider`, `CreatedAt`); PORT `slugify`; markdown `<img>` goes to frontend; DROP `createGifLookupUrl` (proxy-specific `?ids=`) |
| `models/giphy.ts` | 103 | `internal/provider/giphy` | PORT `mapGiphyResponse` (Clips MP4 rendition selection 1080p->720p->360p->gif, `.mp4` name); REBUILD endpoints: `api.giphy.com/v1/gifs/search|trending`, `/v1/clips/search|trending`, `?ids=` lookup. Clips API needs approval (see Gaps) |
| `models/klipy.ts` | 112 | `internal/provider/klipy` | **REBUILD**: types are Tenor-shaped (`media_formats`, `next`) because the Raycast proxy normalized them. Native API: key in path, `page/per_page`, `has_next`, `files`. Re-derive from Klipy docs with a real key |
| `models/finergifs.ts` | 110 | `internal/provider/finergifs` | PORT nearly 1:1 (already calls the real API directly, no key). **Fix:** `gifs()` hardcodes `size=10`, truncating lookups >10 ids; set `size=len(ids)`. Season/episode label + TVDB link mapping port as-is |
| `preferences.ts` | 74 | `internal/settings`, `provider.Registry`, frontend constants | `getServices/getServiceTitle` -> `Provider.Info()`; `GRID_COLUMNS` -> frontend constants; prefs -> `settings.json`; `getPreferenceValues` -> `SettingsService.Get` |
| `hooks/useSearchAPI.ts` | 60 | `GifService` + frontend `useGifSearch` (infinite query) | `getAPIByServiceName` -> `Registry.Get`; cursor/offset duality REPLACED by opaque cursor; `dedupe` applied in Go |
| `hooks/useLocalGifs.ts` | 105 | `LibraryService.List` + frontend `useLibrary` | **Port the invariants** pinned by `tests/useLocalGifs.test.mjs`: all populated providers fail => error not empty; partial failure => show the rest + report failed names; empty providers make no lookup; first-2-rows slices for the in-search Favorites/Recent sections |
| `lib/localGifs.ts` | 61 | `internal/library` | PORT semantics (per-provider, prepend/append, `isSavedNow`, `isSaved` per provider); ADD snapshot + recents cap |
| `lib/cachedGifs.ts` | 126 | `internal/gifcache`, `internal/gif/names.go` | PORT key derivation, extension rule, display-name rule, variant eviction, corrupt-entry handling |
| `lib/resolveGifFile.ts` | 114 | `internal/gifcache/resolve.go` | PORT flow: cache -> stage; else download -> stage; `cacheIfStillFavorite`; absolute-path verify |
| `lib/downloadFile.ts` | 86 | `internal/gifcache/download.go` | PORT exclusive-create naming; failure cleanup; stream with `io.Copy` and ctx |
| `lib/resolveSquareGifFile.ts` | 156 | `internal/imaging/square.go` | REBUILD in pure Go (centered `min(w,h)` crop; preserve per-frame delay, loop count; handle sub-frame offsets/disposal). Always outputs `.gif` even for Clips MP4 inputs: **Clips cannot be square-cropped by `image/gif`**, so disable Square actions for Clips or fail clearly |
| `lib/fetchProviderJson.ts` | 54 | `internal/provider/httpx` | PORT (15s timeout, friendly messages) + ADD key redaction + typed error kinds + 429 `Retry-After` |
| `lib/dedupe.ts` | 12 | `internal/gif/dedupe.go` | PORT |
| `lib/getServiceFromUrl.ts` | 13 | none | **DROP** (`Gif.Provider`) |
| `lib/stripQParams.ts` | 5 | `internal/gif` | PORT (`u.RawQuery = ""`) |
| `lib/formatFileSize.ts` | 13 | frontend util | PORT |
| `search.tsx` | 154 | frontend `features/search`, `components/ProviderSwitch` | REBUILD UI: search bar, provider dropdown (+ Favorites/Recent entries), sections (Favorites, Recent, Trending), empty/error states; `lastProvider` persisted in settings (replaces `useCachedState("service")`); dedicated hotkeys for the "favorites/recents" commands become switcher entries |
| `favorites.tsx`, `recents.tsx`, `components/LocalGifsGrid.tsx` | 5+5+51 | frontend `features/library` | REBUILD: same grid grouped by provider title |
| `components/GifActions.tsx` | 393 | `internal/action` + frontend `ActionMenu`, `keyboard/` | **Split.** Behavior (copy/paste/link/markdown/square/download/page link/open, usage tracking) -> Go. Presentation (action list, sections, shortcuts, default-action-first reordering, toasts) -> frontend. Shortcut map ports as data (replace `cmd+opt+c` etc.; keep the `Cmd+Opt+C/P/L/D`, `Cmd+Shift+Enter/P/F/D/B` defaults) |
| `components/GifDetails.tsx` | 57 | frontend `features/details` | REBUILD: metadata, labels, links, tags, source + logo |
| `components/GifGridItem.tsx`, `GifGridSection.tsx` | 30+43 | frontend `components/Grid*` | REBUILD; `useFrecencySorting` -> `library/frecency.go` (later) |
| `tools/search-gifs.ts`, `ai.yaml` | 20 | none | **DROP** (out of scope: AI/MCP tool) |
| `tests/gifClipboard.test.mjs`, `useLocalGifs.test.mjs` | ~350 | `internal/action/*_test.go`, `internal/library/*_test.go` | **Port as Go table tests against `platform.Fake`.** Key cases: failed op never records usage; hideFilename keeps valid absolute file on cache hit and download; one cache entry + one recent ID after repeated copy/paste; missing cached file -> exactly one download; staging failure -> no second download; default action selects primary Enter despite missing optional actions |
| `CHANGELOG.md` (2026-09-17 entry) | - | test checklist | Treat each bullet as a regression test: per-service favorites, outage isolation, `.mp4` retention, download no-overwrite, hideFilename on all paths |

New with no extension equivalent: `platform/*`, `shell/*` (panel, hotkey, tray, settings window), `settings/secrets.go`, `searchcache`, Accessibility onboarding, staging cleanup.

## Frontend Notes

- **Stack assumption for this file:** Vite + TypeScript + React (the extension is React; its hook structure and `Grid`/`ActionPanel` concepts transliterate), `@wailsio/runtime`, generated bindings from `wails3 generate bindings`. A server-state library (TanStack Query `useInfiniteQuery`) replaces `useCachedPromise` pagination. Final framework/lib versions are decided in STACK.md.
- Vite config must listen on the port `wails3 dev` expects: `server.port = Number(process.env.WAILS_VITE_PORT) || 9245, strictPort: true` (verified in Wails docs/`dev.go`).
- **Keyboard model is the product.** Arrow-key grid navigation, Enter = default action, Cmd+K action menu, Cmd+Shift+F favorite, Esc. Build it as an explicit keymap module, not scattered handlers; the Wails `KeyBindings`/`HideOnEscape` coexist with it, so keep Esc ownership clear (HideOnEscape in Wails; frontend handles Esc only to close submenus first, via a Go `Hide()` call if needed).
- Grid performance: request the **small preview rendition** (`preview_gif`/`nanogif`-class) unless "large" item size is selected (matches `GifGridItem`); `loading="lazy"`; on `panel:hide` unmount or blank the `<img>` sources so hidden GIFs stop decoding.
- Needs a standard **Edit menu** (cut/copy/paste/select-all) available to the Settings window, or Cmd+V into the API key field may not work in an Accessory-policy app.

## Testing Strategy (dev-browser against Wails)

Three layers, in order of how much to trust each:

| Layer | What it proves | How |
|-------|----------------|-----|
| **L1: Go unit tests** | Providers (fixture JSON), httpx redaction, cache, library, action ordering | `go test ./...` with `platform.Fake`, `settings.MemorySecrets`, `httptest.Server` per provider |
| **L2: Browser UAT with real Go backend (server mode + fake platform)** | UI flows end-to-end: search, paging, provider switch, favorites, action menu, settings, error states | `GIFSMITH_PLATFORM=fake FRONTEND_DEVSERVER_URL=http://localhost:9245 go run -tags server .` (backend on `:8080`) + `pnpm dev` (Vite `:9245`); dev-browser opens `http://localhost:8080`. **Verify in the spike** that server mode proxies non-`/wails` paths to `FRONTEND_DEVSERVER_URL`; fallback: Vite `server.proxy` for `/wails` (+ ws) -> `:8080` and open `:9245` |
| **L3: Browser UAT with mock** | Pure visual/keyboard regression, no network, no keys | `VITE_MOCK=1 pnpm dev`; fixtures for all four providers incl. error/empty/rate-limit |
| **L4: Manual by user** | Everything native | Panel over full-screen apps, focus restore, Accessibility prompt, real paste into TextEdit/Slack/Messages, hotkey conflicts, Keychain |

Details that make L2 useful:
- `platform.Fake` records calls (`CopyFile(path)`, `PasteShortcut`) and a **dev-only `DevService`** (behind a `dev` build tag) exposes `PlatformLog()`, `TriggerPanelShow()`, `SetAccessibility(bool)`, so dev-browser can assert "Enter -> copyFile was called with `<staging>/slug.gif`, file exists, non-empty, starts with `GIF8`".
- Each browser tab is a Wails "window" (`browser-1`...); window-specific APIs (hide, position) no-op in server mode, so the frontend must tolerate `Hide()` doing nothing. Also use `httptest`-style stub providers selectable by env (`GIFSMITH_PROVIDER_BASE_URLS`) so L2 never needs real keys or burns the GIPHY hourly cap. Add recorded fixture mode for real-API shape drift checks run manually.
- L2 cannot exercise: NSPanel behavior, key-window/first-responder quirks of the non-activating panel, Cmd+V in panel inputs, global hotkey, TCC. Those are the **spike** and the **manual UAT list**.

## Scalability / Performance Considerations

| Concern | Typical use | Heavy use (100s of favorites, long sessions) | Mitigation |
|---------|-------------|---------------------------------------------|------------|
| Provider rate limits (GIPHY free key: low hourly cap) | 20 searches/day | Rapid typing; panel reopened often | 250ms debounce + ctx cancel + `searchcache` TTL + `singleflight`; typed `rate_limited` error with retry hint; trending cached ~1h (optionally persisted) |
| Webview memory with animated GIFs | 20 results | Infinite scroll to 200+ | Small previews, lazy loading, windowed rendering or cap loaded pages (e.g., 5), blank images on `panel:hide` |
| Library size | tens of entries | 100s of favorites | JSON + snapshot is fine to ~1k entries; cap recents at ~200; move to SQLite only if favorites exceed thousands |
| Warm webview resident memory | one WKWebView always alive | same | Accept (instant-show is the product); do not destroy/recreate per show |
| Staging dir growth | few files | many pastes | Startup + age-based sweep |

## Build Order (dependency-driven)

```
Phase 0  SPIKE  (de-risk native; mostly Wails + 4 cgo functions)
  Wails v3 beta.30 scaffold, ActivationPolicyAccessory, tray, hidden NonActivating panel,
  GlobalShortcut -> Toggle, <input autofocus> receives keys, Cmd+V/Cmd+A work in input,
  platform.CopyFile + PasteShortcut into TextEdit, AXIsProcessTrusted flow, over a full-screen app,
  `-tags server` + browser reaching bindings. GO/NO-GO gates for the whole design.
        |
Phase 1  Domain core (pure Go; no UI)         gif model -> httpx -> provider iface -> finergifs (keyless) first
        |                                      -> giphy -> klipy -> giphy-clips; settings + SecretStore (keys needed by giphy/klipy)
        |                                      -> searchcache.   Verified by L1 tests + fixture JSON.
Phase 2  Services + search UI                 GifService/SettingsService(keys) + frontend shell: search bar, provider
        |                                      switcher, grid, infinite scroll, typed errors, key-missing state.
        |                                      Verified by L2/L3 via dev-browser. (Can start UI on mock before Phase 1 ends.)
Phase 3  Actions + file pipeline              gifcache + action orchestrator + platform (darwin) + ActionMenu + default action;
        |                                      copy first, then paste (+ Accessibility onboarding), link/markdown, page link, open, download.
Phase 4  Library                              library store + snapshot -> favorites/recents UI -> in-search sections -> cache-on-favorite
        |                                      -> frecency (optional).  Depends on Phase 3 (usage recording, cache eviction).
Phase 5  Settings window + shell polish       Settings window, hotkey rebinder, grid sizes, locales, download folder, hideFilename,
        |                                      launch at login, idle-reset rule, details view, square crop (imaging).
Phase 6  Release                               Hardened runtime + Developer ID signing + notarization (network client entitlement,
                                               NO sandbox), stable signing identity for dev builds (TCC), LSUIElement, MIT attribution.
```

Ordering rationale:
- **Spike first** because the entire UX hinges on three unproven Wails behaviors (non-activating panel keyboard input, Show() not stealing activation, full-screen overlay). If any fails, the contingency is a small amount of cgo in `platform` (activate/restore), not a redesign, which is exactly why `platform` is an interface.
- **Finer Gifs first among providers**: keyless, real response shape already in the extension, so the whole search -> grid -> action path can be proven before any API key is involved.
- **Settings/SecretStore before GIPHY/Klipy**: keyed providers are untestable without it.
- **Actions before Library**: `RecordUse` and favorite-cache eviction are action-layer side effects.
- **Square crop late**: isolated, pure Go, independent of everything else; Clips caveat can be decided then.
- **Frontend can lead Go in Phase 2** thanks to the mock seam, but the typed-error and DTO shapes should be frozen first (generated bindings are the contract).

## Gaps and Research Flags

| Gap | Confidence | Impact | Resolve by |
|-----|------------|--------|------------|
| Klipy native endpoints for **lookup by ids**, `files` shape/keys, ads items, `customer_id` requirement | LOW (only third-party write-ups; docs site is JS-rendered) | Mapper rewrite, favorites refresh | Phase 1: call Klipy with a real key, save fixtures. Snapshot design makes this non-blocking for favorites |
| **GIPHY Clips API requires approval** (clips@giphy.com) and has no documented by-ids route | MEDIUM | "GIPHY Clips" provider may be unusable with a normal free key | Phase 1: test whether a standard beta key can call `/v1/clips/*`; if not, gate the provider on a capability probe and show a clear error. **Roadmap may need to mark Clips as conditional** |
| GIPHY terms on caching trending/search results and on client-side vs server calls (a community source claims trending must be client-side) | LOW | Could constrain `searchcache` | Read current GIPHY API ToS before Phase 1 ends; user-owned keys reduce risk |
| `wails3 dev` browser access to bindings | LOW | Testing approach | Spike; server mode (documented) is the fallback and recommended primary |
| Wails v3 is **beta** (beta.30, 2026-10-15; "API stable" but defects possible) | MEDIUM | Upgrade churn | Pin exact version; budget a spike week; keep `shell/` the only Wails-importing package so a swap is contained |
| Pasteboard payload: file URL only vs. also image data (`com.compuserve.gif`) for apps that ignore file URLs | LOW | Paste success in Slack/Discord/browsers | Phase 3 manual matrix; extension used file-only (`Clipboard.copy({file})`), start there |
| Paste timing (hide -> ~60-100ms -> Cmd+V) | MEDIUM | Flaky paste | Tune in Phase 3; verify frontmost app unchanged before sending keystroke |

## Sources

- Wails v3 source, `wailsapp/wails` `master` (read directly): `v3/pkg/application/webview_window_options.go` (MacWindowClassPanel, MacPanelPreferences.NonActivating, CollectionBehavior, WindowLevel, HideOnFocusLost, HideOnEscape), `application_options.go` (ActivationPolicyAccessory, Services, Assets), `global_shortcut_manager.go` / `global_shortcut_darwin.go` (Register/Unregister, Carbon RegisterEventHotKey, may register before Run), `services.go` (ServiceStartup/Shutdown, MarshalError, Route), `bindings.go` (context.Context first param), `systemtray.go`, `clipboard.go` (text only), `webview_panel_darwin.m` (canBecomeKeyWindow YES, key handling for non-activating panels), `v3/examples/spotlight/main.go`, `v3/examples/global-shortcuts/main.go`, `v3/internal/commands/dev.go` (default Vite port 9245). HIGH.
- Wails releases page (v3.0.0-beta.30, 15 Oct; beta.29 layout-aware global hotkeys; beta.30 macOS activation/clipboard fix) and Go module proxy listing (`v3.0.0-beta.27` to `beta.30`). HIGH.
- Wails docs `docs/mpress/content/guides/server-build.md` (server build tag, no CGO, bindings + WebSocket events, window APIs no-op, `:8080`, `/health`) and `guides/dev/frontend-frameworks.md` (Vite port 9245 strictPort). HIGH.
- `zalando/go-keyring` `keyring_darwin.go` (uses `/usr/bin/security`, secret via stdin on Set), latest release v0.2.8 (2026-03-23). HIGH.
- GIPHY developer docs: trending/search `limit` (max 100) + `offset`; Clips endpoints `/v1/clips/search|trending`, access only upon approval (developers.giphy.com/docs/api/endpoint/trending, /docs/clips/endpoint). MEDIUM (search-result excerpts).
- Klipy API, third-party: Zuplo/dev.to "Exploring the KLIPY API" (path-key URL, `page`/`per_page` 8-50, `data.has_next`, `files`), dansup/klipy-php, vyfor/klipy-kt (ads separated from content), docs.klipy.com nav (GIF API endpoint list: Trending, Search, Items, Recent, Categories). LOW; official endpoint bodies not retrievable.
- Apple Developer Forums thread on overlay panels above full-screen apps (accessory policy, non-activating panel, screen-saver level). LOW-MEDIUM (single forum report; informs the `WindowLevel` fallback).
- Local extension source (data): `src/models/*`, `src/lib/*`, `src/hooks/*`, `src/components/*`, `src/search.tsx`, `src/preferences.ts`, `tests/*`, `CHANGELOG.md`, `package.json` preferences. HIGH.
