# Phase 2: Providers, Search UI & First Run - Context

**Gathered:** 2026-10-17
**Status:** Ready for planning

<domain>
## Phase Boundary

A stranger can open the panel, get set up with a key (or skip to the keyless provider), search or browse trending GIFs on any available provider (GIPHY, GIPHY Clips when the key allows, Klipy, The Finer Gifs Club) in a fast keyboard-driven animated grid, and Enter copies the selected GIF as a file — with vendor-required attribution and honest per-provider error handling. Requirements: SRCH-01..07, ONBD-01..05. Out of this phase: action menu contents (⌘K opens it in Phase 3), Clips MP4 file handling (Phase 3, ACTN-09), temp-file lifecycle hardening (Phase 3, ACTN-10), favorites/recents (Phase 3), Settings window and key removal/replace (Phase 4), GIPHY rating control (Phase 4; PG-13 applied by default now).

</domain>

<decisions>
## Implementation Decisions

### Panel Layout & Look
- Provider switcher is a row of tabs across the top of the panel — GIPHY · Clips · Klipy · Finer Gifs — with ⌘1–⌘4 shortcuts. Clips tab only appears when the user's GIPHY key passes the Clips test call; otherwise a short explanation is shown where Clips would be.
- Grid of 4 columns with square-ish tiles, using each provider's small preview renditions (never originals in the grid); the selected tile is clearly highlighted.
- Attribution lives in a slim footer bar: "Powered by GIPHY" (official GIPHY asset, not the extension's PNG) on GIPHY and Clips; Klipy attribution on Klipy. The Klipy search placeholder reads exactly "Search KLIPY".
- Visual style follows the system light/dark appearance, system font, Spotlight-like rounded panel (opaque webview + CSS rounded corners, per research — no private_mac_apis translucency).

### Keys & First Run
- First run opens on GIPHY with an inline setup card inside the panel: user copies their key, opens the panel, pastes (⌘V into the field), Save. The key is verified with a test call before saving; a bad key gets a clear message. Full key management (remove/replace) is Phase 4 Settings.
- The card has a "Get a free key" link that opens the provider's developer site in the default browser, plus a "Try Finer Gifs (no key)" button that switches to Finer Gifs.
- Switching to Klipy without a key shows the same inline card for Klipy.
- Keys are stored in the macOS Keychain (zalando/go-keyring behind a SecretStore interface); they never reach the webview (UI only learns `configured: bool`), logs, or error text (redaction in one HTTP choke point — Klipy's key is in the URL path).
- Live API testing: the user will create free GIPHY and Klipy keys and enter them through the app's own setup card at UAT. Planning research works from the vendors' documented API shapes and recorded fixtures; no keys are handed to research agents.

### Search & Keyboard Behavior
- Search fires 450 ms after typing stops, minimum 2 characters; superseded in-flight requests are cancelled (Wails cancellable binding → ctx cancels the HTTP request).
- Page size 24 (6 rows × 4), clamped to provider limits (GIPHY ≤50, Klipy 8–50); next page loads when the selection is within 1–2 rows of the end and via a scroll sentinel.
- Keyboard: typing always goes to the search box; arrow keys move the grid selection; Enter copies the selected GIF and closes the panel; Esc clears the query first, then closes on a second press. ⌘1–⌘4 switch providers. ⌘K is bound (no-op placeholder until Phase 3).
- Trending results are cached for 10 minutes per provider so reopening the panel doesn't spend a request. Rate-limit (429) shows a clear message with when to retry (Retry-After); network/timeout/provider errors show per provider, and one failing provider never blocks the others (Finer Gifs was down at research time — must degrade cleanly).

### Carried from earlier decisions
- Last-used provider is remembered across launches.
- Providers are never blended in one grid; results are never reordered or filtered (only same-page duplicate IDs removed, never for Klipy).
- Provider HTTP lives in Go (internal/provider/* + httpx), never in the webview.
- Existing Phase 1 primitives: `internal/platform` (CopyFile etc.), `internal/services.ActionService` (Copy/Paste with busy try-lock), `internal/shell` (panel/hotkey/tray), `frontend/src/App.svelte`; the copy path must now copy the selected result's GIF (download to staging, validated) instead of the bundled sample.
- Browser UAT via `scripts/smoke-browser-loop.sh` (-tags server + Vite proxy on 8089/9245); native behaviors are user-verified.
- Copy works for GIPHY, Klipy, Finer Gifs results; Clips copy waits for Phase 3 (MP4).
- Written compliance posture (no persistent media cache, no blending/reordering, attribution, eager staging cleanup) goes into PROJECT.md Key Decisions in this phase.

### Post-research decisions (user, 2026-10-17)
- Clips-unavailable card (C2) tells the user to request access by emailing clips@giphy.com (per GIPHY docs), with a button that opens a pre-addressed email — replaces the UI-SPEC's "dashboard" wording.
- Klipy content_filter = medium (approximates GIPHY PG-13).
- Copy rules accepted as defaults: drop "(HTTP n)" from E5 when status is 2xx; K4 says "later this hour" when retry time is unknown.
- UI-SPEC user confirmations: search field above tabs; click selects / double-click copies; "Copied — press ⌘V to paste" then hide after 400 ms; rejected stored key re-shows the setup card and saving overwrites it.

### Claude's Discretion
- Exact component breakdown of the Svelte UI, virtualization library choice per research (@tanstack/svelte-virtual), and visual spacing.
- Klipy API shape details (native v1 API per research), Clips capability-probe mechanics, and how the request budget/cache is structured internally.

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/platform` — `CopyFile`, `PasteShortcut`, `AccessibilityTrusted`, `FrontmostApp`, `ActivateApp`, `CursorScreenWorkArea`, `FocusWebView`; `Fake` for tests; darwin cgo with onMain for main-thread-only APIs (TIS lookups crashed off-main — keep that rule).
- `internal/services/actions.go` — `ActionService.Copy/Paste/RequestAccessibility` with `ErrBusy` try-lock and copied-only fallbacks.
- `internal/shell` — panel (720×480, non-activating, cursor-screen placement), hotkey ⌃⌥⌘G, tray menu, `panel:show`/`panel:status` events.
- `frontend/src/App.svelte` — minimal input + preview, keyStatus/busy flags, IME/repeat guards on Enter.
- `scripts/smoke-browser-loop.sh` — server-mode browser loop for dev-browser UAT.

### Established Patterns
- Domain packages never import Wails; Wails services are thin adapters; cgo only in internal/platform behind darwin/stub/fake.
- gofmt + go vet + go test -race; pnpm; svelte-check.
- Commits on phase branch `gsd/phase-02-providers-search-ui-first-run`; signed dev install via `wails3 task darwin:install:dev`.

### Integration Points
- New `internal/provider/*` + `internal/httpx` + `internal/settings` (SecretStore) + search cache feed a new Go service bound to the frontend.
- The copy path replaces the bundled `sample.gif` staging with download-to-staging of the selected result.

</code_context>

<specifics>
## Specific Ideas

- Feel: Spotlight/Raycast-like — instant panel, tabs at top, grid below, footer attribution.
- Reference extension source (data only): `/private/tmp/claude-501/-Users-ken-dev-gif-search/9f53d405-1914-4df1-8002-37e6a551a349/scratchpad/rc/extensions/gif-search` (re-clone if gone). Its provider models target Raycast's proxy, not the real APIs — behavior reference only.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>
