<!-- GSD:project-start source:PROJECT.md -->

## Project

**Gifsmith**

Gifsmith is a standalone macOS menu bar app for finding and sharing animated GIFs: press a global hotkey, search GIPHY, GIPHY Clips, Klipy or The Finer Gifs Club, and copy or paste the GIF into whatever app you were using. It is a port of the Raycast "GIF Search" extension for people who don't use Raycast — starting with the author, who has replaced every other piece of Raycast with Vorssaint (which has no plugin system). Free, public release.

**Core Value:** Hotkey → type → Enter puts the right GIF on the clipboard (or into the frontmost app) in seconds, without Raycast.

### Constraints

- **Tech stack**: Go + Wails (prefer v3 if stable at build time, else v2) with a web frontend — author's existing toolchain.
- **Platform**: macOS only for v1; native glue via cgo/Objective-C (NSPasteboard file URLs, CGEvent ⌘V, NSWorkspace frontmost app, NSPanel-style window behavior).
- **Permissions**: Paste requires Accessibility permission; copy must work without it.
- **Distribution**: Developer ID signing + notarization; no App Store.
- **API keys**: Never embedded in the binary; stored locally per user (Keychain preferred).
- **License**: Port of MIT code — retain attribution to the original extension authors.

<!-- GSD:project-end -->

<!-- GSD:stack-start source:research/STACK.md -->

## Technology Stack

## Headline Decisions

## Recommended Stack

### Core Framework

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Go | >= 1.25 (installed: 1.27.2) | Backend, native glue host | Wails v3 requires Go 1.25+ for development (status page). Already installed. | HIGH |
| Wails v3 | `v3.0.0-beta.30` (2026-10-15) | App shell, webview, tray, panel, hotkey, autostart, bindings | Only Wails line with tray, NSPanel, global shortcuts, autostart, accessory policy. See "v3 vs v2". | HIGH (features) / MEDIUM (beta stability) |
| `wails3` CLI | `v3.0.0-beta.30` | Scaffold, dev, bindings, package, sign, notarize | Must match the library version. NOTE: `~/go/bin/wails` on this machine is **v2.16.0**; v3's binary is `wails3` and is not installed yet. | HIGH |
| `@wailsio/runtime` | `3.0.0-beta.30` | JS runtime and generated TS bindings client | Same release train as the Go module. Template uses `"latest"`; pin it to the exact version. | HIGH |
| cgo + Objective-C (own, small) | n/a | NSPasteboard file URL, CGEventPost Cmd+V, AX trust check, frontmost-app capture | Needed regardless; see item 3 and 4 below. Keep in one `internal/native` package with `//go:build darwin && !server` plus a stub. | HIGH |

### Frontend

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| Svelte 5 + TypeScript | svelte `5.57.2`, TS `~5.9` | UI (search, grid, details, settings) | Wails v3 ships a first-class `svelte` template (Svelte 5 + Vite 8). Smallest runtime of the framework options, fine-grained reactivity suits a 100+ animated tile grid. TS 7.0.2 exists on npm but was not tested with svelte-check or the Wails template; stay on 5.9 until the template moves. | HIGH (versions) / MEDIUM (TS 7 untested) |
| Vite | `8.3.4` | Dev server and bundler | Wails v3 expects Vite on port 9245 (`WAILS_VITE_PORT`) and its plugin (`@wailsio/runtime/plugins/vite`) generates bindings. | HIGH |
| `@sveltejs/vite-plugin-svelte` | `7.3.1` | Svelte compile | Matches Vite 8 in the template. | HIGH |
| `@tanstack/svelte-virtual` | `3.13.47` (core `3.19.6`) | Row-virtualized GIF grid | Peer range includes Svelte 5. Virtualize by **rows** (uniform square tiles, N columns computed from container width). Needed because every mounted animated `<img>` keeps decoding. | MEDIUM (not exercised in the probe) |

### Persistence and Secrets

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| `github.com/zalando/go-keyring` | `v0.2.8` (2026-03-23) | API keys in macOS Keychain | Pure Go (shells to `/usr/bin/security -i`, secret passed on stdin, not argv), no cgo, actively maintained (pushed 2026-10-16). Keeps working across ad-hoc dev rebuilds, which a cgo SecItem approach would not (ACL is bound to code identity, so every rebuild prompts). Wrap in a 3-method interface so it can be swapped later. | HIGH (behavior read from source) / MEDIUM (threat-model tradeoff) |
| JSON files via stdlib | n/a | Prefs, favorites, recents | `os.UserConfigDir()` resolves to `~/Library/Application Support`. One JSON file per concern, atomic write (temp + rename). SQLite is unjustified for this data volume. Non-secret prefs only. | HIGH |
| `os.UserCacheDir()` | stdlib | GIF file cache, trending cache | `~/Library/Caches/<bundle-id>`. Port the extension's cache-naming and hide-filename semantics. | HIGH |

### Imaging

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| `image/gif` (stdlib) | Go 1.27 | Square center-crop of animated GIFs | VERIFIED (probe): see item 6. Lossless on palettes, no re-quantization. | HIGH |

### Infrastructure (build, sign, ship)

| Technology | Version | Purpose | Why | Confidence |
|------------|---------|---------|-----|------------|
| `wails3 package` / `wails3 task darwin:package:dmg` / `darwin:package:universal` | beta.30 | `.app`, DMG, arm64+amd64 universal | Shipped in the template Taskfile (`build/darwin/Taskfile.yml`). | HIGH |
| `wails3 tool sign --input X.app --notarize [--keychain-profile P] [--identity ID] [--entitlements F]` | beta.30 | Developer ID sign (hardened runtime default) + notarytool | `wails3 task darwin:sign:notarize` wraps it. Credentials via `xcrun notarytool store-credentials`. | HIGH (CLI flags read from `--help`) / MEDIUM (end-to-end not run) |
| Developer ID cert | present | Signing identity | `Developer ID Application: Ken Scott (S2K7P43927)` is already in this machine's keychain. Xcode 27.1 provides `notarytool`. | HIGH |
| `app.Autostart` (Wails v3) | beta.30 | Launch at login | macOS 13+ bundled app uses `SMAppService.mainAppService` (no TCC prompt); falls back to a LaunchAgent plist. Takes effect on next login. | HIGH (docs + source) |

### Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `golang.org/x/time/rate` | latest | Client-side request budgeter | Strongly recommended: GIPHY and Klipy keys are capped at 100 calls/hour. Debounce, cache trending, and refuse before the 429. |
| `vitest` | `5.0.5` | Frontend unit tests | Provider mapping and state logic. |
| `@playwright/test` | `1.64.1` | Optional scripted browser tests | Only if you want CI-style tests; the project's `dev-browser` skill covers interactive verification against the same proxied URL. |
| `golang.design/x/clipboard` | `v0.11.0` (2026-09-26) | **Fallback only** for file-URL clipboard write | VERIFIED (probe): `WriteFiles` produces `public.file-url` on macOS 27. Purego-based, no cgo. Not primary: 4 releases in 4 months (v0.8 to v0.11), and on macOS it writes the deprecated `NSFilenamesPboardType` and relies on the OS bridging it. |
| `@giphy/js-types` | `5.1.0` (types only, last published 2024-05) | Reference for GIPHY response shape | **Not needed** if Go owns the client. Wails generates TS models from Go structs. Do not add. |

## v3 vs v2 (the actual decision)

- **v3**: `v3.0.0-beta.30` published 2026-10-15. The beta releases are **automated nightly snapshots from `master`** (release notes say so), roughly every 2 to 5 days (beta.22 on 2026-09-14 to beta.30 on 2026-10-15). Docs "status" page: "Current Status: Beta", goal is a stable 3.0, no date. The release notes call the API stable but warn of defects. Desktop (Windows, macOS, Linux) is the supported Beta contract.
- **v2**: tags `v2.15.0` and `v2.16.0` exist (2026-08-17, 2026-09-14); the GitHub "Latest" release page still shows `v2.14.0` (2026-08-10). Maintenance-pace line.

| Need | v3 beta.30 | v2.16 |
|------|-----------|-------|
| System tray / menu bar | `app.SystemTray.New()`, `SetTemplateIcon`, `SetMenu`, `OnClick`, `AttachWindow`, `WindowOffset` (docs `features/menus/systray`) | None in release (`v2/pkg/buildassets/onhold/tray` only). Would need `fyne.io/systray` or similar fighting Wails for the main thread. |
| Frameless, always on top | `Frameless`, `AlwaysOnTop`, `Mac.WindowLevel` (floating, status, popUpMenu, ...), `Mac.CornerRadius` | `Frameless`, `AlwaysOnTop` only |
| Hide/show | `window.Hide()/Show()/Focus()`, `Hidden: true`, `HideOnFocusLost`, `HideOnEscape` | `WindowHide/WindowShow` |
| No Dock icon | `Mac.ActivationPolicy: ActivationPolicyAccessory` (plus `LSUIElement` in Info.plist) | No option (v2 mac options list has none) |
| NSPanel, non-activating, over full-screen | `Mac.WindowClass: MacWindowClassPanel`, `PanelPreferences{NonActivating, FloatingPanel, BecomesKeyOnlyIfNeeded}`, `Mac.CollectionBehavior` (CanJoinAllSpaces, FullScreenAuxiliary). Official `examples/spotlight` does exactly this. | None |
| Global hotkey | `app.GlobalShortcut.Register("Cmd+Option+Space", fn)` (Carbon, no Accessibility needed) | None |
| Launch at login | `app.Autostart` | None |
| Single instance | `SingleInstance` option | Partial |
| Plain-browser bindings | Experimental `-tags server` HTTP mode (verified) | `wails dev` browser port (documented v2 feature, not re-verified here) |
- Pin `github.com/wailsapp/wails/v3 v3.0.0-beta.30` in `go.mod`, install the CLI with `go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.30`, pin `@wailsio/runtime` to the same string. Never use `@latest` (resolves to the newest nightly).
- Upgrade deliberately, between phases, with the spike checklist re-run.
- Keep all Wails-specific code behind a thin `internal/shell` package so the Go domain code (providers, cache, crop, store) has no Wails import.

## Capability-by-Capability Findings (the ten questions)

### 1. Wails v3 mechanics

- **Template:** `wails3 init -n gifsmith -t svelte` (templates: react, svelte, vue, vanilla, ios, plus `-js` variants). Generates `main.go`, `build/config.yml`, `build/<os>/Taskfile.yml`, `frontend/` with Vite 8 + Svelte 5.
- **Package manager:** the template Taskfile defaults `PACKAGE_MANAGER` to `npm` (`Taskfile.yml` line 6). Set it to `pnpm` (env or edit the default) to honor the repo's pnpm preference. VERIFIED (probe): running `pnpm install` first and then letting the npm task run produced `npm error Cannot read properties of null (reading 'matches')` and failed `wails3 dev`. Pick one manager from day one.
- **Cold bindings generation took ~44s** on first run (go package analysis); incremental runs are faster. Budget for it in dev-loop expectations.
- **Dev loop:** `wails3 dev` builds, ad-hoc codesigns a `*.dev.app`, starts Vite on 9245 and loads it through the in-process asset server (VERIFIED (probe): Go app opens no TCP listener).
- **Bindings in a plain browser tab:**
- **Panel/webview transparency needs a private API tag.** Translucent or transparent webviews require `-tags private_mac_apis` (private WebKit `drawsBackground`). Avoid it: use an opaque webview with CSS rounded content and `Mac.CornerRadius`, so the build uses public APIs only. `BackgroundColour` (native window colour) does not need the tag. `OpenDevTools()` programmatically also needs the tag; right-click Inspect works on macOS 13.3+ without it.
- **Quit semantics:** leave `ApplicationShouldTerminateAfterLastWindowClosed` unset/false. Hiding the only window with `orderOut:` counts as closed and would quit the app (documented caution on the global-shortcuts page).

### 2. Global hotkey

- **Use `app.GlobalShortcut.Register(accelerator, callback)`** (v3 beta.30, `global_shortcut_darwin.go`). Carbon `RegisterEventHotKey`; docs state it does not require Accessibility. Callback runs on its own goroutine; window methods are main-thread-safe internally; wrap custom main-thread work in `application.InvokeSync`. Can register before `app.Run()`. Re-register at runtime: `Unregister` then `Register`.
- Accelerator syntax: `"Cmd+Option+Space"`, `"CmdOrCtrl+Shift+G"`, function keys. Letters are layout-aware (translated through the active keyboard layout at registration time).
- **macOS does not reject duplicate hotkeys across apps** (docs): registration succeeds even if another app owns it, so you cannot detect conflicts via the error. Offer a configurable hotkey and a visible "test" affordance.
- Carbon hotkeys cannot express modifier-only triggers (double-tap Cmd, Fn). Do not promise them.
- **`golang.design/x/hotkey` v0.6.4 (2026-09-26): do not use.** Source (`hotkey_darwin.go`) states every hotkey is served by a CGEventTap requiring Accessibility (Input Monitoring) permission, and `Register` fails without it. It also requires the main thread for registration (fixed in 0.6.4 to not self-deadlock when called from `InvokeSync`, but the permission requirement remains). Adopting it would force the Accessibility prompt at first launch, breaking "copy works without it".

### 3. Writing a file URL to NSPasteboard

- Slack/Messages animate a pasted GIF when the pasteboard carries a file URL for a `.gif` file (not image data). Wails v3's `app.Clipboard` is text-only; `golang.design/x/clipboard` image support is PNG-only (animation lost).
- **Recommended: ~10 lines of cgo.** VERIFIED (probe) on macOS 27.0.1:
- **Fallback:** `golang.design/x/clipboard@v0.11.0` `clipboard.WriteFiles(ctx, paths)` (note: returns two values) also produced `public.file-url` on this machine. VERIFIED (probe). v0.11 added `FmtFiles`, `WriteAll`, `Sensitive`; v0.7-era knowledge (text + PNG only) is outdated. Keep it as a documented escape hatch, not a dependency.
- For link / Markdown / page-link actions use plain text (`app.Clipboard.SetText` or the same cgo file).
- Optionally also write `public.url`/string alongside the file with `writeObjects:` multiple items only if testing shows a target app prefers it; do not speculate in v1.
- Manual test matrix for the user (cannot be automated): paste into Slack, Messages, Mail, Notes, Discord, TextEdit (rich text), a browser text field.

### 4. Simulating Cmd+V, restoring focus, Accessibility

- **Post the keystroke** with CoreGraphics from cgo. Reference implementation (Maccy, 22k stars, pushed 2026-10-15; derived from Clipy): `CGEventSourceCreate(kCGEventSourceStateCombinedSessionState)`, `CGEventCreateKeyboardEvent(src, vCode, down)`, set flags to `kCGEventFlagMaskCommand | 0x08` (left-command device bit), post to `kCGSessionEventTap`, key-down then key-up. Maccy also (a) derives the V keycode from the active layout and (b) forces the QWERTY keycode when the layout is a "Dvorak - QWERTY Command" variant. A hard-coded keycode 9 is correct only for QWERTY-positioned V; handle at least the command-switches-to-QWERTY case or accept the limitation explicitly.
- **Hide the panel first, then wait ~50 to 100 ms, then post.** If the panel is still key, the synthetic Cmd+V goes to Gifsmith's own webview. (Timing value is typical practice, not measured; tune in the spike.)
- **Accessibility check:** `AXIsProcessTrustedWithOptions(@{kAXTrustedCheckOptionPrompt: @YES})` shows the system prompt; with `@NO` it silently reports. VERIFIED (probe): compiles and returns via cgo (returned 1 only because the parent terminal was trusted; a fresh app starts at 0). On macOS 10.15+ the narrower pair `CGPreflightPostEventAccess()` / `CGRequestPostEventAccess()` also exists (Apple docs, both macOS 10.15+; also returned true in the probe). Recommended: gate paste actions on `CGPreflightPostEventAccess()` and use `AXIsProcessTrustedWithOptions` to raise the prompt. Deep-link for the "Open Settings" button: `x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility`.
- **Never block copy on this permission.** Paste actions degrade to "copied, press Cmd+V" plus a toast.
- **TCC grants are bound to code identity.** `wails3 dev` ad-hoc signs the dev bundle (`codesign --sign -`), so the Accessibility grant is lost on every rebuild. For paste testing, sign dev builds with the Developer ID identity (stable designated requirement) or expect to re-grant. This is the single most likely time sink in the native spike.
- **Restoring focus:** with a non-activating panel (item 5) the frontmost app never changes, so there is nothing to restore. As a belt-and-braces fallback, capture `[[NSWorkspace sharedWorkspace] frontmostApplication]` (VERIFIED (probe): works from cgo) before showing, and after hiding call `activateWithOptions:` on it. Caveat from Apple docs: since macOS 14 activation is cooperative; `activate(from:options:)` "doesn't guarantee app activation" and `NSApplicationActivateIgnoringOtherApps` is ineffective. Treat the fallback as best-effort, not a guarantee. Confidence: LOW to MEDIUM for fallback reliability, HIGH that the primary path (don't activate) avoids the problem.

### 5. NSPanel-like behavior

- `NonActivating`: showing or focusing the panel leaves the current app active, "but the panel can still become key for controls and text input" (docs). `FullScreenAuxiliary` + `CanJoinAllSpaces` is what lets it float over full-screen apps.
- Panel `hidesOnDeactivate` is forced NO and it is released on close (intentional Wails overrides).
- Also set `LSUIElement` = `true` in `build/darwin/Info.plist` (the generated plist has no such key) so the Dock icon never flashes before the activation policy applies.
- Place the panel on the display containing the mouse cursor (Wails screens API) rather than always centering on the primary display.
- Confidence: HIGH that the options exist and are documented; **MEDIUM that typing, Esc-to-hide, click-away-to-hide, and full-screen overlay all behave on macOS 27**. No automated way to check; this is exactly what the planned spike must prove by hand.
- Settings: keep it a normal second window (activates the app, accepts a hotkey recorder) or a view inside the panel. Decide in the spike; a regular window under accessory policy needs explicit activation.

### 6. Square-cropping animated GIFs

- Output size changed from -45% to +29% relative to input (Go's encoder does no inter-frame optimization). Acceptable; mention in docs if size matters.
- **Pitfall found while writing the probe:** a frame that falls entirely outside the crop window cannot be dropped (it carries timing) and a 1x1 placeholder at index 0 may paint a visible pixel. Use a fully transparent palette entry, or fold its delay into the previous frame.
- **cgo ImageIO is the wrong tool.** `CGImageDestination` re-encodes through its own quantizer and flattens per-frame disposal into full-canvas frames (this is what the extension's Swift script does), with worse fidelity and more code. Only reconsider if you need scaling.
- Not covered: GIPHY Clips are `.mp4`. Square crop does not apply; hide or disable the square actions for Clips (AVFoundation re-encode is out of scope for v1).

### 7. Keychain

- No prompts, works on ad-hoc dev builds, no cgo.
- Items created through `security` are readable by other processes that invoke `security find-generic-password` as the same user (weaker than a per-app ACL). Acceptable for free-tier, rate-capped keys; revisit if keys become valuable.
- `99designs/keyring` v1.2.2: last push 2024-05-07. It is a multi-backend abstraction (file, pass, KWallet) whose macOS backend wraps keybase/go-keychain via cgo. Overkill for macOS-only; skip.
- Never put keys in `localStorage`, the JSON prefs file, env, or the binary. The frontend only ever sees "key present: yes/no" and a masked tail; saving goes through a bound `SetKey(provider, value)` method.

### 8. Frontend

- **Svelte 5 + Vite 8 + TS** via the official Wails `svelte` template. Rationale: smallest runtime and best-fit reactivity for a keyboard-driven grid; React adds ~40 KB baseline and reconciliation cost for no benefit here; vanilla TS would re-implement grid state, keyboard navigation, and the action menu by hand.
- **Virtualization is a correctness issue, not polish.** Each mounted animated `<img>` keeps decoding in WebKit; hundreds of them burn CPU and memory. Use `@tanstack/svelte-virtual` with row virtualization (column count = floor(containerWidth / cellSize), tiles are square so row height is fixed), overscan of 2 rows, and swap to the next page via an `IntersectionObserver` sentinel. `virtua` 0.53.5 (Svelte >= 5 peer) is a fine alternative if tanstack's store-style API proves awkward; both are MEDIUM (not exercised).
- **Preview media:** load small renditions in the grid (GIPHY `preview_gif`/`fixed_height_small`; Klipy `sm`/`xs` in `webp` or `gif` via `format_filter`), full `original`/`hd.gif` only on copy. Media must be loaded directly from provider URLs (a Klipy requirement).
- **GIPHY types:** `@giphy/js-types` 5.1.0 is types only and stale (2024-05). Because Go owns the HTTP client, define minimal Go response structs (the fields the extension's `mapGiphyResponse` reads) and let Wails generate the TS models. Skip the dependency.
- **Keyboard model** (arrow-key grid navigation, Enter default action, Cmd+K action menu, Esc hides) is hand-built; there is no off-the-shelf Raycast-style component worth the dependency.

### 9. Signing, notarization, launch at login

- Build: `wails3 package GOOS=darwin` produces `bin/Gifsmith.app`; universal via `wails3 task darwin:package:universal`; DMG via `wails3 task darwin:package:dmg`.
- Set in `build/darwin/Taskfile.yml`: `SIGN_IDENTITY: "Developer ID Application: Ken Scott (S2K7P43927)"`, `KEYCHAIN_PROFILE`; run `xcrun notarytool store-credentials` (or `wails3 setup signing`) once; then `wails3 task darwin:sign:notarize` (equivalent to `wails3 tool sign --input bin/Gifsmith.app --notarize`; hardened runtime on by default for notarization). `wails3 setup entitlements` can generate `entitlements.plist`; for a non-sandboxed WKWebView app the production file should be minimal (the docs say "just network access"), but treat "does notarization pass with the generated entitlements" as a Phase-N checklist item, not a given.
- Fix the placeholders in `build/config.yml` and `Info.plist` (template ships `com.example...`, "My Product"). Bundle id must be stable before first release because Autostart and TCC use it.
- Do not build inside iCloud/Dropbox-synced folders (Finder metadata invalidates codesign; Wails relocates output automatically and prints where).
- **Launch at login:** `app.Autostart.Enable()/Disable()/IsEnabled()`; SMAppService path on macOS 13+ for a bundled app, LaunchAgent plist otherwise. Persist the user's toggle and call `Enable()` on startup (documented as idempotent). The app should live in `/Applications` for a stable path.
- Mac App Store remains out of scope (sandbox forbids synthesized Cmd+V).

### 10. Provider APIs

- Base `https://api.giphy.com/v1/`, auth `api_key` query param. Search `GET /v1/gifs/search?q=&limit=&offset=&rating=&lang=&bundle=` (q max 50 chars, offset max 4999, limit max 50 for beta keys); trending `GET /v1/gifs/trending` (offset max 499); by IDs `GET /v1/gifs?ids=a,b,c` (up to 100). Response `{data[], pagination, meta}`; errors give empty `data` and a 4xx/5xx in `meta`.
- Rate limit: new keys are **beta, 100 calls/hour**, exceeding returns **429**. More requires a production application and a sales conversation. Search and trending "must be called client-side" (satisfied by a desktop app).
- `bundle=messaging_non_clips` trims renditions (smaller payloads); `clips_grid_picker` exists for Clips.
- Attribution: "Powered By GIPHY" must be conspicuously displayed wherever the API is used.
- **GIPHY Clips: `GET /v1/clips/search`, `GET /v1/clips/trending`, `GET /v1/content[/:id]` require approval ("Access to Clips endpoints is only available upon approval. Please reach out to clips@giphy.com").** The extension worked because Raycast's proxy holds an approved key. A user's self-issued beta key will likely be refused. Plan: ship Clips behind a "requires approved key" toggle, surface the 401/403 as an explanatory message, and consider deferring Clips out of the first release. Clip results carry `video.assets` (360p..4k mp4) and a GIF-style `images` subset.
- **Native v1 API:** `GET https://api.klipy.com/api/v1/{app_key}/gifs/search` and `.../gifs/trending`, `.../gifs/items?slugs=a,b`. The **key is a path segment**, not a query param. Query: `q`, `page` (1-based), `per_page` (search 8 to 50, default 24), `customer_id` (stable per-user id; generate and persist a UUID), `locale` (ISO 3166 alpha-2 country, e.g. `us`), `content_filter` (off|low|medium|high), `format_filter` (comma list of gif, webp, jpg, mp4, webm). Response: `{result, data: {data: [{id, slug, title, file: {hd|md|sm|xs: {gif|webp|jpg|mp4|webm: {url,width,height,size}}}, tags, type, blur_preview}], current_page, per_page, has_next}}`. Pagination is page-number plus `has_next`, not the Tenor `next` cursor the extension uses.
- Tenor-compat host `api.klipy.com` also exists (replace `tenor.googleapis.com`); prefer the native v1 API, it is the documented primary.
- Rate limit: **Test key 100 requests/hour; production unlimited** after a Partner Panel request (`partner.klipy.com`). Users get a key from the Partner Panel.
- **Integration rules that touch product design** (`integration-requirements`): load media directly from returned URLs, do not alter URL params; **do not store, mirror, re-host, or retain copies of KLIPY media** without written approval; requests from the end-user client only (desktop app qualifies); **do not reorder, insert, remove, or filter results** (so do not run the extension's dedupe on Klipy results); **keep Klipy in its own grid, never blended with other providers**. Attribution: search placeholder must read **"Search KLIPY"** (required); "Powered by KLIPY" mark optional.
- **Compliance risk:** "Copy GIF / Paste GIF / Download" necessarily writes a local file. Mitigate by writing Klipy files to the cache dir with eager cleanup (delete on next copy and at quit), keep Klipy links/Markdown actions as the no-file path, and email `developers@klipy.com` for clearance before public release. Flag for the roadmap as a pre-release gate.
- Search `GET https://api.thefinergifs.club/search?q=&q.parser=simple&sort=_score desc&size=N&start=OFFSET` (CloudSearch-style; lookup by `fileid` with `q.options={"fields":["fileid"]}`), media at `https://media.thefinergifs.club/<fileid>.gif`. No trending endpoint (the extension returns empty).
- **On 2026-10-17 the API returned HTTP 530 / Cloudflare error code 1033 on repeated requests** (the Cloudflare tunnel to its origin is down; the marketing site returned 200). The extension's changelog also records repeated "provider failed" reports for this provider. Treat it as a hobby service: never block startup or the other providers on it, show a distinct "service unavailable" state, and do not make it the default provider. This is observed state at one point in time and may recover.

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| App shell | Wails v3 beta.30 | Wails v2.16 | No tray, no NSPanel, no activation policy, no hotkey, no autostart. Third-party shims fight for the main thread. |
| App shell | Wails v3 | Native Swift (SwiftUI/AppKit) | Would be the lowest-risk path for panel/paste behaviors, but it discards the author's Go toolchain and the dev-browser UI verification loop (explicit project decision). |
| App shell | Wails v3 | Electron / Tauri | Heavier / different toolchain; Tauri would need Rust. Out of the author's stack. |
| Global hotkey | Wails `app.GlobalShortcut` (Carbon) | `golang.design/x/hotkey` v0.6.4 | CGEventTap backend requires Accessibility; extra main-thread coupling. |
| File-URL clipboard | Own cgo `writeObjects:` | `golang.design/x/clipboard` v0.11.0 `WriteFiles` | Works (verified) but fast-moving dependency writing a deprecated pasteboard type. Kept as fallback. |
| Key storage | go-keyring v0.2.8 | 99designs/keyring v1.2.2 | Stale (2024-05), multi-backend overkill, cgo. |
| Key storage | go-keyring | Own cgo SecItem | Per-app ACL is stronger, but every dev rebuild changes code identity and re-prompts. Revisit post-v1. |
| GIF crop | `image/gif` | cgo ImageIO | Re-quantizes, flattens disposal, more code, no benefit without scaling. |
| Frontend | Svelte 5 | React 19 / vanilla TS | React: unneeded weight. Vanilla: re-implements grid/keyboard state. Svelte has an official Wails template. |
| Virtualization | `@tanstack/svelte-virtual` rows | Render everything + `loading=lazy` | Animated images keep decoding once loaded; memory and CPU grow without bound. |
| Provider calls | Go | Webview `fetch` | Keys would reach the frontend, CORS issues, no Keychain access, harder to rate-limit centrally. |
| Persistence | JSON files | SQLite | Tiny data (favorites, recents, prefs). No query needs. |
| Klipy API | Native v1 | Tenor-compat endpoints | Native is the documented primary; compat layer is a migration aid. |

## Installation

# Toolchain (Go 1.27.2 and Xcode 27.1 already present)

# Scaffold (pnpm: set PACKAGE_MANAGER=pnpm in Taskfile.yml before first dev run)

# Frontend

# Dev

# Browser-testable backend (experimental server mode) + proxied Vite:

# vite.config.ts (browser profile): server.proxy = { "/wails": { target: "http://localhost:8089", ws: true } }

# Ship

## Open Items and Confidence Summary

| Topic | Confidence | What would raise it |
|-------|------------|---------------------|
| Wails v3 has every needed primitive | HIGH | Read in source and docs; spotlight example is the exact use case |
| v3-beta stability for this app | MEDIUM | Nightly snapshots; spike on tray + panel + hotkey + autostart before building features |
| Panel typing/Esc/click-away/full-screen on macOS 27 | MEDIUM | Manual spike (user-run); cannot be automated |
| Carbon hotkey needs no Accessibility | HIGH (docs and source) | Confirm on a clean TCC state in the spike |
| NSPasteboard file URL pastes as animated GIF in Slack/Messages | MEDIUM | Pasteboard types verified; actual app behavior needs the manual matrix |
| Cmd+V synthesis and keycode/layout handling | MEDIUM | Reference impl read (Maccy); post timing untested |
| Focus-restore fallback | LOW | Only needed if the non-activating panel misbehaves |
| Lossless square crop | HIGH | Verified on five real GIFs |
| Keychain via go-keyring | HIGH (mechanics) / MEDIUM (security tradeoff) | Decision-log entry |
| Vite-proxy browser testing | HIGH for bindings via HTTP; MEDIUM for WebSocket events | Exercise an event in the first UI phase |
| GIPHY limits and Clips gating | HIGH | Official docs fetched 2026-10-17 |
| Klipy v1 shape and rules | HIGH | Official llms-full.txt and md pages; request formats not exercised with a live key (no key available) |
| Klipy "no copies" vs copy-file feature | MEDIUM (policy ambiguity) | Written answer from developers@klipy.com |
| Finer Gifs Club availability | HIGH that it was down when probed; unknown going forward | Re-probe at implementation time |
| Signing/notarization end to end | MEDIUM | Not executed; identity and notarytool present, CLI flags confirmed |
| `@tanstack/svelte-virtual` fit, TS 7 compatibility | MEDIUM / LOW | Try in the grid phase |

## Sources

- Wails repo, releases and tags: https://github.com/wailsapp/wails (v3.0.0-beta.30 release notes, v2.14.0/2.15.0/2.16.0 tags)
- Wails v3 docs source (docs tree, `docs/mpress/content`): status, installation, `features/keyboard/global-shortcuts`, `features/autostart/basics`, `features/menus/systray`, `features/windows/options`, `guides/build/macos`, `guides/build/signing`, `guides/build/private-macos-apis`, `guides/dev/frontend-frameworks`, `guides/server-build`; published at https://v3.wails.io
- Wails v3 source read: `v3/pkg/application/global_shortcut_darwin.go`, `webview_window_options.go`, `webview_panel_darwin.m`, `permissions_darwin.go`, `application_server.go`; `v3/examples/spotlight`, `v3/examples/server`
- GIPHY: https://developers.giphy.com/docs/api/ , https://developers.giphy.com/docs/api/endpoint , https://developers.giphy.com/docs/clips , https://developers.giphy.com/docs/clips/endpoint/ , https://developers.giphy.com/docs/optional-settings/
- Klipy: https://docs.klipy.com/llms.txt , https://docs.klipy.com/llms-full.txt , https://docs.klipy.com/gifs-api/gifs-search-api.md , https://docs.klipy.com/gifs-api/gifs-trending-api.md , https://docs.klipy.com/gifs-api/gifs-items-api.md , https://docs.klipy.com/integration-requirements.md , https://docs.klipy.com/attribution.md , https://klipy.com/developers
- Apple: https://developer.apple.com/documentation/coregraphics/cgpreflightposteventaccess() , .../cgrequestposteventaccess() , https://developer.apple.com/documentation/appkit/nsrunningapplication/activate(from:options:)
- golang.design/x/hotkey v0.6.4 source and README; golang.design/x/clipboard v0.11.0 source and release notes; zalando/go-keyring v0.2.8 source; 99designs/keyring repo metadata
- npm registry versions for svelte, vite, @sveltejs/vite-plugin-svelte, @tanstack/svelte-virtual, virtua, @giphy/js-types, @wailsio/runtime, typescript, vitest, @playwright/test (queried 2026-10-17)

<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->

## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->

## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->

## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->

## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-fast` for a trivial task inline, with no subagents and no PLAN.md
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->

## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
