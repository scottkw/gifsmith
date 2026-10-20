# Walking Skeleton — Gifsmith

**Phase:** 1
**Generated:** 2026-10-17

## Capability Proven End-to-End

From any app (including one in macOS full-screen), pressing ⌃⌥⌘G shows a floating panel that takes typing immediately while the previous app keeps focus; Enter puts a bundled sample GIF on the clipboard as a file, and ⌘Enter pastes it straight into the previous app, all from a Developer ID–signed build installed at `~/Applications/Gifsmith.app` (and from a throwaway notarized build).

## Architectural Decisions

| Decision | Choice | Rationale |
|---|---|---|
| App framework | Wails v3 pinned exactly at `v3.0.0-beta.30` (go.mod, `wails3` CLI from the same tag, `@wailsio/runtime` `3.0.0-beta.30`) | Only Go desktop shell with NSPanel, Carbon global shortcut, tray and accessory policy built in (STACK.md "v3 vs v2"). Beta, so pin exactly and upgrade only between phases with the spike checklist re-run |
| Frontend | Svelte 5 + TypeScript + Vite 8 from the official `svelte` template; pnpm only (`Taskfile.yml` `PACKAGE_MANAGER` default `pnpm`), lockfile committed, `frontend/.npmrc` `minimum-release-age=10080` kept | Smallest runtime for the later animated grid; one package manager from day one (mixing npm/pnpm broke `wails3 dev` in research) |
| Data layer | None in this project's Phase 1 (no database anywhere in Gifsmith). The only writes are a staged file at `~/Library/Caches/com.kenscott.gifsmith/staging/sample.gif` and the macOS pasteboard. Later phases use JSON files in `~/Library/Application Support` and the Keychain | Tiny data volume; vendor terms forbid a media cache |
| Native glue | `internal/platform`: the ONLY cgo package. `Platform` interface (`platform.go`, no build tag) with three implementations: darwin cgo (`//go:build darwin && !server`), stub (`//go:build !darwin \|\| server`, returns `ErrUnsupported` = "not supported in this build"), recording `Fake` for unit tests | Lets `-tags server` build CGO-free for browser UAT and lets action sequencing be unit-tested |
| Shell | `internal/shell`: Wails panel window (NSPanel, non-activating, all Spaces + full-screen auxiliary, floating level, 720×480, hidden at start, HideOnEscape, HideOnFocusLost), Carbon hotkey `Ctrl+Option+Cmd+G` registered in `ApplicationStarted`, tray menu (Open Gifsmith / Settings… / Quit Gifsmith), cursor-screen placement | Wails owns windows and the main thread; shell is plain Go over Wails |
| Bound surface | `internal/services.ActionService` — `Copy(ctx)`, `Paste(ctx)`, `RequestAccessibility(ctx)`; returns `Result{copied, pasted, needsAccessibility, message}`. Takes no path or URL from the webview | Thin, testable seam; webview cannot steer native calls |
| Events | Go → frontend `panel:show` (focus + select input), `panel:status` (status line text) | Smallest event surface for Phase 1 |
| Auth / permissions | No accounts. Copy needs no permission. Paste gates on `AXIsProcessTrustedWithOptions`; the system prompt appears only from an explicit button click | Locked decision: copy must work without Accessibility |
| Identity | Bundle id `com.kenscott.gifsmith` (release + signed dev), `com.kenscott.gifsmith.dev` for `wails3 dev` (`build/darwin/Info.dev.plist`); macOS 13.0 minimum; arm64 only; `LSUIElement=true` in both plists + `ActivationPolicyAccessory` | TCC grants key on bundle id + designated requirement; `.dev` id keeps ad-hoc builds away from the signed app's TCC row |
| Deployment (dev) | `wails3 task darwin:install:dev` → package → `wails3 tool sign --identity "Developer ID Application: Ken Scott (S2K7P43927)" --hardened-runtime` → `ditto` to `~/Applications/Gifsmith.app` → `open`. Paste is never tested from `wails3 dev` or `bin/` | Stable Developer ID designated requirement keeps the Accessibility grant across rebuilds |
| Deployment (notarized) | `wails3 task darwin:sign:notarize` → `wails3 tool sign … --notarize --keychain-profile gifsmith-notary` (profile created once by the user; credentials never in the repo) | Proves RELS-01/02 mechanics early; Phase 4 release becomes a gate, not a discovery |
| Browser UAT loop | `CGO_ENABLED=0 go build -tags server` backend on `localhost:8089` + Vite on `127.0.0.1:9245` with `/wails` proxy when `GIFSMITH_BACKEND` is set; `scripts/smoke-browser-loop.sh` proves a bound call round-trips | Phases 2–3 use dev-browser against this loop; native behaviors stay manual |
| Directory layout | Wails project at repo root: `main.go`, `Taskfile.yml`, `build/`, `frontend/` (`src/App.svelte`, `public/sample.gif`, generated `bindings/` committed), `internal/{platform,shell,services}`, `scripts/` | Matches research ARCHITECTURE.md; domain packages added in Phase 2 never import Wails |

## Decision Index (CONTEXT.md decisions, numbered in file order)

Plans cite these IDs. CONTEXT.md lists the decisions as bullets; the numbering below is positional.

| ID | Decision |
|---|---|
| D-01 | Spike code is the foundation: Wails scaffold, `internal/shell`, `internal/platform` are the real packages from Phase 2 on |
| D-02 | Enter = copy GIF file to clipboard; ⌘Enter = paste into the previous app |
| D-03 | One small sample GIF bundled/embedded; no provider or API key |
| D-04 | Panel UI: plain text input plus one GIF preview, no styling |
| D-05 | Default hotkey ⌃⌥⌘G via Wails `app.GlobalShortcut` (Carbon), registered after the app is running; user confirms no clash with Vorssaint |
| D-06 | Panel opens on the display under the mouse cursor, horizontally centered, near the top |
| D-07 | Fixed panel size 720×480 |
| D-08 | If a non-activating panel cannot take typing without stealing focus: record frontmost app → activate panel → reactivate previous app before pasting (`platform.ActivateApp`) |
| D-09 | Bundle id `com.kenscott.gifsmith` |
| D-10 | Signed dev builds installed to and run from `~/Applications/Gifsmith.app`, signed with `Developer ID Application: Ken Scott (S2K7P43927)`; paste never tested from `wails3 dev` |
| D-11 | Notarization via notarytool keychain profile `gifsmith-notary`, created once by the user (checkpoint) |
| D-12 | Paste matrix = markdown checklist in the phase dir; user tests apps they have, n/a for missing; file-URL-only payload unless the matrix proves otherwise |
| D-13 | Wails v3 `v3.0.0-beta.30` pinned exactly (go.mod, CLI, runtime); pnpm from the first command; Svelte 5 template |
| D-14 | macOS 13+, arm64 only; `LSUIElement=true`; accessory activation policy; no Dock icon |
| D-15 | No third-party hotkey module (it needs Accessibility on macOS) |
| D-16 | cgo isolated in `internal/platform` with `darwin && !server` tag plus a stub so `-tags server` builds |
| D-17 | Copy works without Accessibility; paste gates on `AXIsProcessTrusted` and degrades to "copied — press ⌘V" with how-to-grant guidance |
| D-18 | ⌘V from a private CGEventSource with explicit Command flag and layout-aware V keycode; hide panel and wait ~60–100 ms before posting (tune in spike) |

## Stack Touched in Phase 1

- [ ] Project scaffold — Wails v3 svelte template at repo root, pnpm, Taskfile, `go test`, `go vet`, `svelte-check` (plan 01)
- [ ] Routing — single `/` route served from the embedded `frontend/dist` into the panel window (plan 01)
- [ ] Data write/read — real file write (staged sample GIF) and real native write (pasteboard file URL, read back in a darwin test on a private named pasteboard) (plan 01); no database exists in this project
- [ ] UI — text input wired to bound `ActionService.Copy` (Enter) and `ActionService.Paste` (⌘Enter) (plans 01, 03)
- [ ] Deployment — Developer ID–signed, hardened-runtime install to `~/Applications/Gifsmith.app` via `wails3 task darwin:install:dev` (plan 03); notarized + stapled build via `wails3 task darwin:sign:notarize` (plan 04)
- [ ] Browser UAT loop — `scripts/smoke-browser-loop.sh` (plan 03)
- [ ] Menu-bar shell — tray menu (Open / Settings… / Quit) and cursor-screen placement (plan 02)

## Out of Scope (Deferred to Later Slices)

- Providers, search, trending, grid, provider switcher, key setup, attribution (Phase 2)
- Action menu (⌘K), link/Markdown/download/square crop, temp-file lifecycle, favorites/recents/details (Phase 3)
- Settings window (the Phase 1 "Settings…" menu item opens the panel with a status line), hotkey rebinding, launch at login, DMG, clean-account install, About/credits (Phase 4)
- Any pasteboard payload beyond a single file URL, unless the Phase 1 paste matrix proves an app needs more
- Panel styling, rounded corners, translucency (`private_mac_apis` tag stays off)
- Intel / universal builds

## Subsequent Slice Plan

Each later phase adds one vertical slice on top of this skeleton without altering its architectural decisions:

- Phase 2: Search a real provider (Finer Gifs keyless first, then GIPHY/Klipy with Keychain keys) → animated grid in the panel → Enter copies the selected GIF file through the same `platform.CopyFile` primitive
- Phase 3: Every extension action through one Go action orchestrator reusing the Phase 1 Accessibility-gated paste path; favorites/recents from metadata snapshots
- Phase 4: Settings window, rebindable hotkeys, launch at login, signed + notarized DMG installed on a clean account
