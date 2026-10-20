# Phase 4: Settings, Shell Polish & Release - Research

**Researched:** 2026-10-18
**Domain:** Wails v3 beta.30 multi-window accessory app on macOS (Settings window, hotkey rebinding, login item, appearance), Go prefs and provider parameters, Developer ID signing, notarization and DMG release
**Confidence:** MEDIUM-HIGH. The Wails mechanics were read in the pinned module source. Native activation and paste behavior and the notarization run still have to be checked by hand on a real install.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Settings Window
- A separate normal window (~640×520) with a sidebar: General · Keys · Hotkeys · Search · Downloads · About. Opened from the tray menu "Settings…" and ⌘, in the panel. The window activates the app so typing and ⌘V paste work (SETT-01).
- Changes apply immediately (no Save button), Mac-style. Exception: keys are verified with a test call before saving (Phase 2 SaveKey path).
- Keys tab: one row per provider (GIPHY, Klipy) with status (Saved ✓ / Not set / Rejected), last 4 characters of the key, and Change / Test / Remove buttons plus "Get a free key ↗" (GIPHY developer dashboard, Klipy Partner Panel). Remove asks for confirmation. The webview still never receives a full key (SETT-02).
- General tab: Theme (System / Light / Dark, default System — overrides both the webview colors and the native window appearance, SETT-06), Launch at login toggle (SHEL-05), Accessibility status with a Grant button (SETT-05), default Enter and ⌘Enter actions (ACTN-08; defaults Copy GIF / Paste GIF).

#### Hotkeys & Launch at Login
- Click-to-record hotkey field: press a combination requiring at least one of ⌘/⌥/⌃; it registers immediately; if registration fails the old binding stays and an error shows; Esc cancels recording; "Reset to default".
- Three hotkeys: Open panel (default ⌃⌥⌘G), Open Favorites (no default), Open Recents (no default); Favorites/Recents may be left unset (SHEL-06).
- Conflict handling: macOS doesn't report clashes with other apps (Carbon RegisterEventHotKey succeeds anyway), so: block duplicates among Gifsmith's own hotkeys, show an error when registration fails, and offer a "Test" button that opens the panel to confirm (SETT-03).
- Launch at login off by default; toggle uses Wails `app.Autostart` (SMAppService on macOS 13+); the user's choice is persisted and re-applied (idempotently) at startup.

#### Search, Display & Download Settings
- Grid item size Small / Medium / Large, separately for search and trending: Small = 5 columns, Medium = 4 (current), Large = 3; panel stays 720×480.
- Page size ("max results") 24 / 36 / 48, clamped to each provider's limits. GIPHY language and Klipy country as dropdowns defaulting to the system locale. GIPHY content rating G / PG / PG-13 / R, default PG-13 (SRCH-09).
- Download folder: "Choose…" opens the native folder picker, plus "Reset to Downloads"; stored by path. Picking an exFAT/FAT or network folder shows a clear error (the atomic os.Link claim needs a hard-link-capable volume). "Hide original filename" saves as `gifsmith-<id>` (existing `FileStem(g, hide)`) (SETT-04, ACTN-11).
- Reduce Motion: follow the macOS accessibility setting — tiles show a still first frame and only the selected tile animates. No separate toggle.

#### Release & Compliance
- Build target: Apple Silicon (arm64) DMG for macOS 13+, drag-to-Applications layout, version 1.0.0 (RELS-01).
- Notarization: build and wire every release step now; notarize once the user resolves the Apple Developer agreement HTTP 403 (`notarytool store-credentials`). If still blocked at release time, the phase pauses at a blocking checkpoint asking the user to resolve it — never skipped silently, never shipped un-notarized as "done".
- Klipy file actions (Copy/Paste GIF, Copy/Paste Square, Download) are disabled until written clearance is recorded: on Klipy items they're hidden, Enter copies the link and ⌘Enter pastes the link; a setting records the clearance (RELS-04). Draft the email to developers@klipy.com (temp-file copies + ad labelling IN-09) for the user to send.
- Distribution: GitHub Release on scottkw/gifsmith with the DMG, a SHA-256 checksum and release notes; README covering install, getting keys, GIPHY/Klipy attribution and credits. No auto-updater in v1. About screen credits the original MIT-licensed Raycast GIF Search extension and its authors (RELS-03; LICENSE already carries the notice).
- Clean-account install test (RELS-02) is user-verified on a fresh macOS user account.

### Claude's Discretion
- Settings persistence schema (extend Phase 2 prefs JSON, atomic 0600), how settings changes propagate to the panel (Wails events), window component structure, hotkey recorder implementation details, how theme overrides NSAppearance, DMG tooling (Wails Taskfile `darwin:package:dmg`), and test strategy (Go -race, node:test, headless UAT; native behaviors user-verified).

### Specific Ideas (from CONTEXT.md)
- Fold in `.planning/todos/pending/phase-3-ui-review-polish.md` (resolves_phase 4): destructive-row contrast, footer `aria-live`, shorter T5a message, right-click opens the action menu, Details "Esc Back" as a button, tertiary-text contrast, hide library hints on empty tabs.
- Consider the 7 open Phase 3 review Info items (03-REVIEW-DISPOSITION.md) where code is touched anyway, and surface the favorites cap (1000, oldest dropped silently).
- Settings window should feel native (system font, sidebar list), consistent with the panel's tokens.

### Deferred Ideas (OUT OF SCOPE)
- Universal (Intel) build, website distribution, Sparkle auto-updates — post-v1.
- Mac App Store — out of scope (sandbox forbids synthesized ⌘V).

### Approved UI contract
`04-UI-SPEC.md` (approved 2026-10-18) is binding for layout, copy (G/K/H/S/D/A/T/F rows), tokens, the hotkey recorder states, the Klipy gate surfaces, Reduce Motion, theme and the release surfaces. Its Open Questions use stated defaults: Klipy clearance is **not** a user toggle; hotkey Test **listens 5 s for the real hotkey**; library tabs use the Search size; square actions are not offered as defaults; picking the same action for both keys **swaps** them; there is a small system-shortcut denylist.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SETT-01 | Settings open in their own window, where pasting into fields works | Second `WebviewWindow` (titled, hidden at start, close hook hides it); `Show()`+`Focus()` activate the app (`windowFocus` calls `activateIgnoringOtherApps:`); the existing app-menu Edit role gives ⌘V. See Pattern 1, Pitfalls 1-3 |
| SETT-02 | Add, test, change, remove GIPHY and Klipy keys; key sign-up links | Reuse `SaveKey`; add `SecretStore.Delete` (go-keyring `Delete` returns `ErrNotFound` when absent); `TestKey` must bypass the trending cache; Go-side "rejected" flag; `KeyStatus` returns last 4 only. Pattern 4, Pitfall 7 |
| SETT-03 | Rebind hotkeys; conflicts reported; old binding stays | `GlobalShortcutManager.Register/Unregister` semantics read in source (duplicate within the app → error, native failure → rollback). Shell-owned `Rebind` with re-register-old on failure; registrar seam for server-build UAT. Pattern 3, Pitfalls 4-6 |
| SETT-04 | Grid size, max results, download folder, hide filename, GIPHY/Klipy locales | Prefs schema; provider params read per request; `staging.Save(ctx,g,dir,hide)` already takes `hide`; folder validation via `syscall.Statfs` + a hard-link probe; native picker `app.Dialog.OpenFile()` with directory options. Patterns 2, 5, 6 |
| SETT-05 | Accessibility status + grant button | Existing `platform.AccessibilityTrusted(false)` and `ActionService.RequestAccessibility`; add a status getter; re-read on window focus / `settings:show`. |
| SETT-06 | Theme Light / Dark / System | No Wails runtime appearance setter exists. Add a cgo `[NSApp setAppearance:]` (nil = System); `app.Env.IsDarkMode()` + `events.Common.ThemeChanged` for native background colours; `data-theme` in CSS. Pattern 7 |
| SRCH-09 | GIPHY content rating (default PG-13) | `rating` is hard-coded `"pg-13"` today (giphy.go:172); make it a per-request param; GIPHY allows `g, pg, pg-13, r`. Clear the cache (not the 429 block) on change |
| ACTN-08 | Default Enter and ⌘Enter actions | Frontend mapping over `lib/actions.ts` ids; persisted in prefs; Clips and gated-Klipy remaps per UI-SPEC |
| SHEL-05 | Launch at login | `app.Autostart.Enable/Disable/IsEnabled` (SMAppService path for a bundled app on macOS 13+, LaunchAgent fallback); detect translocation / DMG for copy G3. Pattern 8, Pitfall 9 |
| SHEL-06 | Favorites / Recents hotkeys | Same hotkey manager; the callback emits an event, and the frontend decides whether to show the tab or hide the panel |
| RELS-01 | Developer ID-signed, notarized DMG, macOS 13+, arm64 | Release task: package → rename to `Gifsmith.app` → sign (hardened runtime) → notarize + staple app → `wails3 tool package --format dmg` → codesign DMG → notarize + staple DMG → verify → sha256. Pattern 9, Pitfalls 10-13 |
| RELS-02 | Clean-account install | User-verified checkpoint; list of exact checks in Validation section |
| RELS-03 | About credits the MIT extension and authors | About pane copy A2 (UI-SPEC); LICENSE copied into `Contents/Resources` (the template bundle task does not copy it today) |
| RELS-04 | Klipy file actions disabled unless cleared | Enforce in Go (ActionService refuses Klipy file actions) **and** in the UI (rows omitted, ↵/⌘↵ remapped). Gate is a compile-time constant, not a settings.json field |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

From `/Users/ken/dev/CLAUDE.md` and `/Users/ken/dev/gif-search/.claude/CLAUDE.md`:

- **Stack:** Go + Wails v3 `v3.0.0-beta.30` (pinned exactly; never `@latest`), Svelte 5 + TS, pnpm. `@wailsio/runtime` pinned to `3.0.0-beta.30`.
- **Go conventions:** `go fmt`, golangci-lint, context-aware functions (`ctx context.Context` first on bound methods).
- **JS/TS:** camelCase, PascalCase components, TypeScript types.
- **Testing:** aim for 80%+ coverage in critical components. Run Go tests with `-race`. Frontend unit tests use `node --test`. Headless UAT goes through `dev-browser` against the `-tags server` build.
- **API keys:** never embedded in the binary; Keychain only; the webview never receives a full key.
- **Permissions:** paste needs Accessibility, and copy must work without it.
- **Distribution:** Developer ID + notarization, no App Store.
- **License:** keep attribution to the original extension authors (MIT).
- **Keep Wails-specific code in `internal/shell`** (and main wiring); domain packages (providers, staging, settings, library) do not import Wails.
- **Silent fallbacks forbidden** (`or {}` style). Unknown or invalid states should fail loudly or reset to a declared default with a log line, never quietly corrupt state.
- **Chesterton's Fence** before deleting anything, including template leftovers.
- **RULE 0:** stop on unknown or catastrophic failures (for example a notarization failure you don't understand).
- **GSD workflow:** work only through GSD commands. Ponytail applies to executor agents: build the smallest correct thing.
- **Memory constraints (user's primary Mac):** agents never send synthetic keystrokes, mouse events or hotkeys (no CGEventPost, osascript keystroke, cliclick). Never broad `pkill -f`; kill only processes you started (the existing `install:dev` uses `pkill -x Gifsmith`, which is scoped project tooling). Never touch `~/Library/Application Support` or `~/Downloads`. Never run `go build ./...`.
- `install:dev` adds `stretchr/objx` to go.sum, which must be restored afterwards (Phase 3 note).

## Summary

Every native primitive this phase needs already exists in Wails v3 beta.30. This was checked in the pinned module source, not in docs. A second `WebviewWindow` gets standard key handling and Wails per-window `KeyBindings` (use `"Cmd+W"` → hide). `WebviewWindow.Focus()` activates an inactive app with `activateIgnoringOtherApps:` before `makeKeyAndOrderFront:`, which is what an accessory-policy app needs for a typable second window. The existing application menu already carries the Edit role that made ⌘V work in the panel's key field (Phase 2 native UAT), so it covers the Settings window too. `GlobalShortcutManager` has exactly the rebinding semantics the spec wants. A duplicate inside the app is rejected and the original kept. A native failure rolls the reservation back. Accelerators canonicalize to sorted modifiers (`Ctrl+Option+Cmd+G` → `Cmd+Ctrl+Option+G`). `app.Autostart` is SMAppService-based for a bundled app on 13+. `app.Dialog.OpenFile()` supports directory-only pickers, a cancel returns `""`, and `AttachToWindow` gives a sheet.

Three things are missing, and the plan must build them. (1) Wails has **no runtime appearance API** (only the create-time `Mac.Appearance` option), so theme override needs a ~10-line cgo `[NSApp setAppearance:]` in `internal/platform`. (2) The `-tags server` build stubs out global shortcuts, autostart and file dialogs (they always error), so headless UAT of the Settings window needs small injected seams (registrar, autostart, folder picker) with fakes in `wire_server.go`, the same pattern as `platform.Fake`. (3) The DMG tool positions icons only at fixed percentages (28% / 72% across, centre row, 96 px icons). The UI-SPEC's 135/405 at 128 px cannot be expressed through `wails3 tool package`. The cheapest correct move is to draw the background to Wails' layout (app ≈ (151,164), Applications ≈ (388,164)) rather than add a new DMG dependency.

The biggest code-level risk is in existing code. `SettingsService.SetLastProvider` writes `settings.Prefs{LastProvider: …}` wholesale (services/settings.go:109), so as soon as Prefs grows, every tab switch would wipe every other setting. Prefs needs one owner with load-modify-save under a mutex, defaults normalised on load. Release risk: the notarization keychain profile is not present on this Mac (probe below), and the Apple agreement 403 is a user action. Every signing step can be built and run (Developer ID signing already works in `install:dev`), but notarize/staple must sit behind a blocking user checkpoint.

**Primary recommendation:** Split into roughly six sequential plans. (1) Go core: one Prefs store, SettingsService API, provider params, Klipy gate, hotkey manager, autostart, folder validation, appearance and locale cgo, all TDD with seams and fakes. (2) Settings window shell and the multi-page frontend entry. (3) Settings panes. (4) Panel changes and the Phase 3 polish. (5) Release pipeline, README and Klipy email, with a blocking notarization checkpoint. (6) Native UAT and clean-account checkpoint. If the planner exceeds about six plans, split RELS-01..04 into a Phase 5 as the roadmap allows.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Settings window lifecycle (create hidden, show+activate, hide on close, ⌘W, position) | Go `internal/shell` (Wails) | — | Native window ops; must stay out of domain packages |
| Settings persistence (prefs JSON, defaults, validation) | Go `internal/settings` | Go `internal/services` (bound API) | Single owner prevents the SetLastProvider overwrite; never holds secrets |
| Key add/test/change/remove, last4, rejected flag | Go `internal/services` + `internal/settings` (Keychain) | Frontend (display only) | The webview never sees a key |
| Hotkey registration, rebind, suspend/resume, test-listen | Go `internal/shell` | Frontend recorder (captures combo, shows glyphs) | Carbon registration is native; the UI only proposes accelerator strings |
| Accelerator validation (modifier required, denylist, duplicates) | Go (authoritative) | Frontend (instant H1 feedback) | JS input is untrusted; Go re-validates |
| Launch at login | Go `internal/shell` (wraps `app.Autostart`) | — | SMAppService |
| Theme | Frontend (`data-theme`, tokens) | Go cgo `NSApp.appearance` + window background colours | Native chrome and webview must agree |
| Provider params (rating, lang, locale, page size) | Go provider clients (read per request) | `search.Service` cache flush | Keys and HTTP stay in Go |
| System locale resolution | Go cgo (`NSLocale`) | — | GUI apps have no `LANG`; resolve at request time |
| Download folder choice + validation | Go (native picker + Statfs + link probe) | — | JS never supplies a path |
| Klipy file-action gate | Go `ActionService` (enforce) | Frontend (omit rows, remap ↵/⌘↵) | Vendor rule binds the developer; enforce at the boundary |
| Grid size, default actions, Reduce Motion stills | Frontend | Go prefs (persist) | Pure presentation |
| Release (sign, notarize, staple, DMG, checksum, GitHub Release) | Build tooling (`build/darwin/Taskfile.yml`) | User checkpoints | Credentials and outward-facing publish are user actions |

## Standard Stack

No new libraries. Everything below is already pinned in `go.mod` / `frontend/package.json` or ships with Xcode/macOS.

### Core
| Library / tool | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/wailsapp/wails/v3` | `v3.0.0-beta.30` [VERIFIED: go.mod] | Second window, KeyBindings, GlobalShortcut, Autostart, Dialog, Env.IsDarkMode, events | Already the app shell; all APIs read in module source this session |
| `github.com/zalando/go-keyring` | `v0.2.8` [VERIFIED: go.mod] | Add `Delete` for Remove Key | `keyring.Delete(service,user)`; maps "could not be found" to `ErrNotFound` [VERIFIED: go-keyring@v0.2.8/keyring_darwin.go:104-114] |
| `wails3` CLI | `v3.0.0-beta.30` [VERIFIED: `wails3 version`] | `tool sign`, `tool package --format dmg`, `task` | Matches the library; the Taskfile already wraps it |
| Go stdlib `syscall.Statfs` | Go 1.27.2 [VERIFIED: `go version`] | Volume type + local check for the download folder | `Statfs_t` has `Flags uint32` and `Fstypename [16]int8` [VERIFIED: GOROOT/src/syscall/ztypes_darwin_arm64.go:85-100] |
| Xcode `notarytool` / `stapler` / `codesign` / `spctl` / `hdiutil` | notarytool 1.1.3 (42) [VERIFIED: `xcrun notarytool --version`] | Notarize, staple, verify | Apple's only supported path |
| `gh` CLI | 2.102.0, logged in as `scottkw` [VERIFIED: `gh auth status`] | GitHub Release | Repo `scottkw/gifsmith` is public [VERIFIED: `gh repo view`] |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `golang.org/x/sys/unix` | v0.46.0 (already indirect in go.mod) [VERIFIED: go.mod] | `unix.MNT_LOCAL` (= `0x1000`) | Optional. Stdlib `syscall` has no `MNT_LOCAL` constant (grep of GOROOT darwin syscall found none). Either promote x/sys to direct or declare `const mntLocal = 0x1000` with a comment citing x/sys zerrors_darwin_arm64.go:999 [VERIFIED] |
| Vite multi-page input | Vite 8.3.4 [VERIFIED: node_modules/vite/package.json] | `settings.html` second entry | `build.rolldownOptions.input` (Vite 8 deprecates `rollupOptions`) [VERIFIED: vite/dist/node/index.d.ts:2904-2907] |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `wails3 tool package --format dmg` (fixed icon layout) | Small `go run` program using `github.com/leaanthony/dmg/dmg` with `IconPositions` + `Icon.Size: 128` | Exact UI-SPEC coordinates, but it adds a new direct dependency (not in go.sum today) and needs a legitimacy gate. Its own `mkdmg` CLI cannot set icon positions (its JSON config has no positions field) [VERIFIED: leaanthony/dmg cmd/mkdmg/main.go]. Use only if the user rejects the layout amendment |
| cgo `NSApp.appearance` | Per-window `Mac.Appearance` option | Create-time only; no runtime setter in Wails [VERIFIED: webview_window_darwin.go:1716-1718 only applies it at init] |
| Settings created hidden at startup | Lazy create on first open | Lazy saves one WKWebView while unused, but adds a second code path (create vs show) and a close-destroy race. Hidden-at-start matches the panel pattern. Recommend hidden-at-start |
| Klipy gate as compile-time `const` | settings.json bool / UI toggle | A JSON flag is user-editable and bypasses a vendor rule that binds the developer. UI-SPEC OQ1 already says not user-visible |

**Installation:** none. No new npm or Go modules.

## Package Legitimacy Audit

This phase installs no external packages. Everything used is already in `go.mod` (`wails/v3 v3.0.0-beta.30`, `go-keyring v0.2.8`, indirect `golang.org/x/sys v0.46.0`) or `frontend/package.json`, or ships with macOS/Xcode.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| (none new) | — | — | — | — | — | — |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none
*If the planner chooses the custom-DMG alternative, `github.com/leaanthony/dmg` becomes a new direct dependency and needs a `checkpoint:human-verify` plus a legitimacy check first. It is Wails' own transitive dependency at `v0.0.0-20260731074841-5c28840cf819` [VERIFIED: wails v3 go.mod line 25], but it has no tagged release.*

## Architecture Patterns

### System Architecture Diagram

```
 Tray "Settings…" ──┐            ⌘, in panel (JS keydown) ──► PanelService.OpenSettings
                    ▼                                                │
              Shell.OpenSettings ◄────────────────────────────────────┘
                │ panel.Hide → settings.Show → settings.Focus (activates app)
                ▼
   ┌──────────────── Settings window (settings.html) ────────────────┐
   │ panes ─► SettingsService (bound) ─► settings.Store (prefs.json) │
   │   │            │   ├─► Keys (Keychain: Set/Delete), Test (no cache)
   │   │            │   ├─► Shell.Rebind/Suspend/Resume/TestHotkey ─► Registrar (app.GlobalShortcut | fake)
   │   │            │   ├─► Autostart (app.Autostart | fake)
   │   │            │   ├─► PickFolder (app.Dialog sheet | fake) ─► ValidateDir (Statfs + link probe)
   │   │            │   └─► platform.SetAppearance (cgo NSApp) + window bg colours
   │   │            └─► app.Event.Emit("settings:changed", snapshot)  ──────────────┐
   └────────────────────────────────────────────────────────────────┘             │
                                                                                   ▼
   Panel (index.html): data-theme, grid cols, ↵/⌘↵ mapping, Klipy gate, reset results on param change
        │ search/trending ─► GifService ─► search.Service (cache; Flush on param change) ─► provider clients
        │                                       providers read Params() per request: rating, lang, locale, pageSize
        └ actions ─► ActionService ─► refuses Klipy file actions while KlipyFileActions == false
   Global hotkey (Carbon) ─► callback goroutine ─► Shell.Toggle | emit "panel:open-tab" {favorites|recents}
```

### Recommended Project Structure (additions only)
```
internal/
├── settings/        # + Store (single owner of prefs.json: Load/Update, defaults, validation); SecretStore.Delete; Keys.Delete
├── services/        # SettingsService grows: Snapshot, Set* methods, KeyStatus/TestKey/RemoveKey, ChooseFolder, AccessibilityStatus; KlipyFileActions const
├── shell/           # + settings window; hotkeys.go (Registrar interface, Rebind/Suspend/Resume/Test); autostart wrapper; theme application
├── platform/        # + SetAppearance(mode), SystemLocale() (cgo; Fake + stub)
└── fsprobe/ (or inside services) # ValidateDownloadDir: Statfs + write + os.Link probe (pure Go, unit-testable on tmp dirs)
frontend/
├── index.html, settings.html          # two Vite entries
└── src/
    ├── settings/main.ts, SettingsApp.svelte, *Pane.svelte, KeyRow, HotkeyRecorder, SettingRow
    └── lib/ hotkeys.ts (event→accelerator, glyphs, validation), theme.ts, prefs.ts (+ .test.ts)
build/darwin/  # Info.plist 1.0.0, new dmg-background.png, Taskfile release tasks
```

### Pattern 1: Second window in an accessory app (SETT-01)
**What:** Create the Settings window once (hidden), cancel its close, show+focus to activate.
**Source:** [VERIFIED: wails v3 webview_window_darwin.go:942-953 `windowShow` → `makeKeyAndOrderFront:` for a normal window; :1100-1111 `windowFocus` → `if (![NSApp isActive]) [NSApp activateIgnoringOtherApps:YES]; [nsWindow makeKeyAndOrderFront:nil]; [nsWindow makeKeyWindow];`; webview_window.go:378-380 KeyBindings processed per window; webview_window_darwin.m:102-107 `performKeyEquivalent` dispatches window KeyBindings before Cocoa/menu handling]
```go
// internal/shell (sketch; field names quoted from webview_window_options.go:82-242)
s.settings = app.Window.NewWithOptions(application.WebviewWindowOptions{
    Name:                "settings",
    Title:               "Gifsmith Settings",
    Width:               640, Height: 520, // [ASSUMED] content size for titled windows — verify visually
    URL:                 "/settings.html",
    Hidden:              true,
    DisableResize:       true,
    MinimiseButtonState: application.ButtonDisabled, // ButtonEnabled=0, ButtonDisabled=1, ButtonHidden=2 (options.go:19-23)
    MaximiseButtonState: application.ButtonHidden,
    BackgroundColour:    bgFor(resolvedTheme), // application.NewRGB(...)
    KeyBindings: map[string]func(application.Window){
        "Cmd+W": func(w application.Window) { s.HideSettings() },
    },
})
s.settings.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
    e.Cancel()           // same pattern as the panel (shell.go:54-57): never destroy it
    s.settings.Hide()
    // resume hotkeys if a recorder was mid-recording; best-effort ActivateApp(prev) — see Pitfall 2
})

func (s *Shell) OpenSettings() {
    s.Hide()                          // panel hides first (UI-SPEC)
    if !s.settingsPlaced { center on CursorScreenWorkArea(); s.settingsPlaced = true }
    s.settings.Show()
    s.settings.Focus()                // activates the accessory app
    s.app.Event.Emit("settings:show") // page re-reads AX status, key status, autostart
}
```
The app menu already set in `shell.New` (AppMenu + EditMenu roles, shell.go:61-64) provides ⌘X/⌘C/⌘V/⌘A/⌘Z to whichever window is key. Do not add a File/Window menu (shell.go:59 comment: dropped deliberately). ⌘W comes from the window-level KeyBinding instead.

### Pattern 2: One prefs owner (fixes the overwrite bug)
**What:** `settings.Store` holds the decoded prefs in memory behind a mutex. `Update(func(*Prefs) error)` copies, mutates, validates, writes atomically with the existing `WriteJSONAtomic`, then swaps. Defaults are applied on load. Zero values mean "default" so old files and new fields both work.
**Why:** `SetLastProvider` currently does `settings.SavePrefs(s.Dir, settings.Prefs{LastProvider: string(id)})` [VERIFIED: internal/services/settings.go:102-110], which would erase every new field.
```go
// Prefs today [VERIFIED: internal/settings/settings.go:122-125]:
//   type Prefs struct { LastProvider string `json:"lastProvider"` }
type Prefs struct {
    LastProvider    string  `json:"lastProvider"`
    Theme           string  `json:"theme,omitempty"`           // "" = system | "light" | "dark"
    LaunchAtLogin   bool    `json:"launchAtLogin,omitempty"`
    PrimaryAction   string  `json:"primaryAction,omitempty"`   // "" = "copy"
    SecondaryAction string  `json:"secondaryAction,omitempty"` // "" = "paste"
    Hotkeys         Hotkeys `json:"hotkeys"`                    // Open "" = default; Favorites/Recents "" = unset
    GridSearch      string  `json:"gridSearch,omitempty"`      // "" = medium | small | large
    GridTrending    string  `json:"gridTrending,omitempty"`
    PageSize        int     `json:"pageSize,omitempty"`        // 0 = 24; allowed 24/36/48
    GiphyRating     string  `json:"giphyRating,omitempty"`     // "" = "pg-13"; g|pg|pg-13|r
    GiphyLang       string  `json:"giphyLang,omitempty"`       // "" = match system
    KlipyCountry    string  `json:"klipyCountry,omitempty"`    // "" = match system
    DownloadDir     string  `json:"downloadDir,omitempty"`     // "" = ~/Downloads
    HideFilenames   bool    `json:"hideFilenames,omitempty"`
}
```
Validate on load. An unknown enum value resets to the default and logs it, which is a declared default, not a silent fallback. An unreadable file keeps today's behaviour ("treat as first run", logged). Existing action ids to reuse for ↵/⌘↵: `'copy' | 'paste' | 'copyLink' | 'pasteLink' | 'copyMarkdown' | 'pasteMarkdown' | 'download'` [VERIFIED: frontend/src/lib/actions.ts:4-19 ActionId union includes these].

### Pattern 3: Hotkey rebind with rollback (SETT-03, SHEL-06)
**Source semantics** [VERIFIED: wails v3 global_shortcut_manager.go:100-148, 186-213, 230-246]: `Register` parses the accelerator and rejects a canonical name already registered by this app (`"global shortcut %q is already registered"`, original kept). After start it calls the native register on the main thread and on failure deletes the reservation. `Unregister` removes the maps **before** the native unregister. `UnregisterAll` clears everything. Callbacks run on their own goroutine. Canonical form = modifiers sorted alphabetically + upper-cased key [VERIFIED: keys.go:78-89], darwin names `Cmd`, `Ctrl`, `Option`, `Shift` [VERIFIED: keys.go:39-45], so the default `"Ctrl+Option+Cmd+G"` (shell.go:17) canonicalises to `Cmd+Ctrl+Option+G`. Compare canonical forms for H2 duplicates. Carbon status `-9878` is reported as "already registered (possibly by another application)" [VERIFIED: global_shortcut_darwin.go:165-171].
```go
type Registrar interface { // satisfied by *application.GlobalShortcutManager; fake in -tags server
    Register(accel string, cb func()) error
    Unregister(accel string) error
}
// Rebind: validate (≥1 of Cmd/Ctrl/Option; denylist; not equal to another slot's canonical) →
// if old != "" { Unregister(old) } → if new != "" { if err := Register(new, cb); err != nil {
//     if old != "" { _ = Register(old, cb) /* restore; log if that fails too */ }; return H4/H4b } }
// → persist via Store.Update → emit settings:changed. new == old canonical → no-op success.
// Suspend(): unregister all current (keep the slot map); Resume(): register all; idempotent; Resume also
// on settings window hide/lost-focus so a recorder can never leave hotkeys dead.
// TestHotkey(slot): arm a 5 s flag; the slot's callback, if armed, emits "hotkey:test" {slot} then does its normal action.
```
Key-name mapping for the recorder (JS `KeyboardEvent` → Wails key): named keys Wails accepts on macOS are `return/enter, tab, space, backspace, delete, escape, home, end, page up, page down, left, right, up, down, f1…f20`, digits, punctuation `= - ] [ ' ; \ , / . \``. Letters resolve through the active keyboard layout at registration time [VERIFIED: global_shortcut_darwin.go:223-274]. Never emit `+` as a key (it splits the accelerator; Wails uses `plus`), and never emit `super` (renders as `Cmd`, collides with `cmd`).

### Pattern 4: Key test must not hit the cache
`TestKey(p)` calls the provider client directly inside `svc.Guard(p, …)` (keeps the 429 budget), then `Seed`s the trending cache. Going through `svc.Trending` returns a cached page within `TrendingTTL` and reports "accepted" without contacting the vendor. On `KeyRejected` from either Test or a panel search, set an in-memory `rejected[p]` in Go (cleared on SaveKey/Remove) so both windows agree on the status.

### Pattern 5: Provider parameters per request + cache flush
GIPHY: `rating` from prefs (`g|pg|pg-13|r` [CITED: developers.giphy.com/docs/api/endpoint]), `lang` on **search only** (Trending does not accept `lang`) [CITED: same], `limit` = page size (beta keys capped at 50). Klipy: `per_page` (8-50), `locale` = ISO 3166 alpha-2 country [CITED: docs.klipy.com/gifs-api/gifs-search-api.md]. Today: `v.Set("rating", "pg-13")` and `v.Set("limit", strconv.Itoa(provider.PageSize))` [VERIFIED: giphy.go:170,172], `v.Set("per_page", strconv.Itoa(provider.PageSize))`, `v.Set("content_filter", "medium")` [VERIFIED: klipy.go:118-119], `const PageSize = 24` [VERIFIED: provider.go:59]. Inject `Params func() provider.Params` into the clients. On a param change call a new `search.Service.Flush(p)` that drops cached pages and bumps `gen`, but **keeps** the 429 block (`Invalidate` also clears the block, which is right only for a key change [VERIFIED: search.go:80-81 comment]).

### Pattern 6: Download folder (SETT-04)
Picker: `app.Dialog.OpenFile().CanChooseFiles(false).CanChooseDirectories(true).CanCreateDirectories(true).SetDirectory(cur).AttachToWindow(settingsWin).PromptForSingleSelection()`. A cancel closes the channel with no value, so the result is `""` with a nil error [VERIFIED: dialogs.go:262-275, dialogs_darwin.go:501-560]. In the server build it returns the error `"file dialogs not available in server mode"` [VERIFIED: application_server.go:383-395], so inject `PickFolder func(start string) (string, error)` with a fake. Validation (pure Go, testable on temp dirs): `os.Stat` is a dir (else D8). Then `syscall.Statfs`: `Fstypename` `msdos`/`exfat` → D5, and `Flags & 0x1000 (MNT_LOCAL) == 0` → D6. Then `os.CreateTemp(dir)` fails → D7. Then `os.Link(tmp, tmp+".l")` fails → D5. Remove both. Persist only on success. `Download` then uses `Store.DownloadDir()` and passes `HideFilenames` to `Staging.Save` (today `a.Staging.Save(ctx, g, dir, false)` [VERIFIED: actions.go:408]). JS never sends a path. "Reset to Downloads" takes no argument.

### Pattern 7: Theme (SETT-06)
- Frontend: `<html data-theme>` holds the resolved theme. System follows `matchMedia('(prefers-color-scheme: dark)')`. Replace the two `@media (prefers-color-scheme: dark)` blocks [VERIFIED: app.css:27 is one; Footer.svelte:31-34 `<picture>` media query must become script-chosen].
- Go: cgo `platform.SetAppearance(mode)` → on main: `NSApp.appearance = nil | [NSAppearance appearanceNamed:NSAppearanceNameAqua] | …DarkAqua`. Setting the app appearance also flips WKWebView's `prefers-color-scheme`, so System/Light/Dark stay consistent across native chrome, form controls and CSS [ASSUMED: WebKit derives prefers-color-scheme from the view's effectiveAppearance; verify natively].
- Window background colours: `win.SetBackgroundColour(application.NewRGB(…))` works at runtime [VERIFIED: webview_window.go:801-805]. Resolve System via `app.Env.IsDarkMode()` (reads `AppleInterfaceStyle`) [VERIFIED: application_darwin.go:73-85] and re-apply on `events.Common.ThemeChanged` (mapped from `Mac.ApplicationDidChangeTheme`) [VERIFIED: events_common_darwin.go:7-12].
- Apply the stored theme in `ApplicationStarted` before any window is shown.

### Pattern 8: Launch at login (SHEL-05)
`app.Autostart.Enable()` → SMAppService `mainAppService register` when running from a `.app` with a bundle id on macOS ≥ 13, else a LaunchAgent plist. `Disable()` tries both. `IsEnabled()` treats "requires approval" (user switched it off in System Settings) as not enabled [VERIFIED: autostart_darwin.go:26-111, autostart_darwin_smappservice.go:23-140]. The server build returns `ErrAutostartNotSupported` [VERIFIED: autostart_server.go], so inject an interface and a fake.
- The checkbox shows `IsEnabled()` (OS truth), not the pref.
- Startup re-apply (locked decision): only if `pref == true && !IsEnabled()` and the status is not "requires approval". Otherwise a user who turned it off in System Settings gets it forced back on every launch [ASSUMED behaviour of register in requiresApproval state].
- G3 (move to Applications): show it without calling `Enable` when the executable path contains `/AppTranslocation/` or starts with `/Volumes/` [ASSUMED heuristic; `runningFromAppBundle` only checks the `.app/Contents/MacOS` shape].

### Pattern 9: Release pipeline (RELS-01)
Existing facts: `APP_NAME: "gifsmith"` [VERIFIED: Taskfile.yml:4] → bundle `bin/gifsmith.app`. `create:app:bundle` ad-hoc signs and does not copy LICENSE [VERIFIED: build/darwin/Taskfile.yml:163-186]. `sign:notarize` depends on `package` and relies on `~/.config/wails/defaults.yaml` for the identity and profile, and that file does not exist here [VERIFIED: ls]. `wails3 tool sign` on a `.app` runs `xattr -cr`, `codesign --force --deep --sign ID [--options runtime]`, strict verify, then (with `--notarize`) `ditto -c -k --keepParent` → `xcrun notarytool submit … --keychain-profile P --wait` → `xcrun stapler staple` [VERIFIED: internal/commands/sign.go:104-247]. The tool has no `.dmg` branch, so sign and notarize the DMG with `codesign`/`notarytool` directly. Default codesign already adds a secure timestamp for the Developer ID identity: the installed dev build shows `Timestamp=Oct 18, 2026 at 12:59:20 PM`, `flags=0x10000(runtime)` and no entitlements [VERIFIED: `codesign -dvv ~/Applications/Gifsmith.app`].

Recommended `darwin:release` task (sequential, each step fails hard):
1. `wails3 task darwin:package` (production, arm64) → `bin/gifsmith.app`; copy `LICENSE` into `Contents/Resources` (add to `create:app:bundle`).
2. `rm -rf bin/release && mkdir bin/release && ditto bin/gifsmith.app bin/release/Gifsmith.app` (Finder-visible name).
3. `wails3 tool sign --input bin/release/Gifsmith.app --identity "{{.SIGN_IDENTITY}}" --hardened-runtime` (no entitlements).
4. **[checkpoint: notary profile exists]** `wails3 tool sign --input bin/release/Gifsmith.app --identity … --notarize --keychain-profile gifsmith-notary` (re-signs, notarizes, staples the app).
5. `wails3 tool package --format dmg --name Gifsmith --out bin/release --background build/darwin/dmg-background.png --volume-icon build/darwin/icons.icns --file-icon build/darwin/dmg-file-icon.icns` → `bin/release/Gifsmith.dmg` (the tool finds `<out>/<name>.app` [VERIFIED: tool_package.go:37-50]; volume name = `--name`). Rename to `Gifsmith-1.0.0-arm64.dmg`.
6. `codesign --force --sign "{{.SIGN_IDENTITY}}" --timestamp Gifsmith-1.0.0-arm64.dmg`
7. `xcrun notarytool submit Gifsmith-1.0.0-arm64.dmg --keychain-profile gifsmith-notary --wait` then `xcrun stapler staple` the DMG.
8. Verify: `codesign --verify --deep --strict -vv` (app), `spctl -a -vv -t exec Gifsmith.app` (expect `source=Notarized Developer ID`), `spctl -a -vv -t open --context context:primary-signature X.dmg`, `xcrun stapler validate` (both), `codesign -d -r-` designated requirement compared with the dev build's (TCC grant continuity, Phase 1 carry-over).
9. `shasum -a 256 Gifsmith-1.0.0-arm64.dmg > Gifsmith-1.0.0-arm64.dmg.sha256`.
10. **[checkpoint: user approves publish]** `gh release create v1.0.0 … --title "Gifsmith 1.0.0" --notes-file …`.

DMG layout: the Wails tool pins the app at `X = width*28/100`, Applications at `width*72/100`, `Y = height/2 - 26`, icon size 96 [VERIFIED: tool_package.go:104-125]. At 540×380 that is app (151,164) and Applications (388,164), 96 px. Draw `dmg-background.png` (540×380, flat `#F2F2F7`, arrow between those centres, "Drag Gifsmith to Applications" 48 px from the bottom) to that geometry. This amends UI-SPEC's (135,170)/(405,170)/128 px target; record it as a decision.

### Anti-Patterns to Avoid
- **Writing a fresh `Prefs{…}` literal anywhere.** Always `Store.Update`.
- **Calling `search.Service.Invalidate` for a rating, lang or page-size change.** It resets the 429 block. Use `Flush`.
- **Accepting a folder path, URL or accelerator from JS without Go validation.** Go validates accelerators. Folder paths come only from the native picker.
- **Running the Klipy gate only in Svelte.** `ActionService` must refuse too (`Copy`, `Paste`, `Download`, `CopySquare`, `PasteSquare` on `provider.Klipy`) while the gate is closed.
- **Leaving hotkeys suspended.** Resume on recorder end, Esc, blur, window hide and window lost-focus. Go owns the resume, not just JS.
- **Running `wails3 task common:update:build-assets`.** It regenerates `Info.plist` from `config.yml` [VERIFIED: build/Taskfile.yml:233-237] and would drop `LSUIElement` and `NSDownloadsFolderUsageDescription`. Edit Info.plist by hand.
- **Using the hardened-runtime ad-hoc `package` output for release.** It is ad-hoc signed. The release task re-signs with the Developer ID.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Global hotkeys | CGEventTap / `golang.design/x/hotkey` | `app.GlobalShortcut` (Carbon) | No Accessibility needed; rollback semantics built in |
| Login item | Writing LaunchAgent plists | `app.Autostart` | SMAppService + fallback + legacy cleanup already done |
| Folder picker | cgo NSOpenPanel | `app.Dialog.OpenFile()` directory mode | Sheet attach, cancel handling done |
| Atomic prefs write | New writer | Existing `settings.WriteJSONAtomic` | Already 0600/0700, temp+sync+rename, tested |
| Keychain delete | `security` exec | `keyring.Delete` | Maps not-found; secret never on argv |
| Notarize + staple app | Custom notarytool wrapper | `wails3 tool sign --notarize --keychain-profile` | Zip, submit --wait, staple in one step |
| DMG | hdiutil + AppleScript Finder layout | `wails3 tool package --format dmg` | Writes .DS_Store in pure Go (no Finder automation/TCC prompt) and uses `ditto` to preserve signatures [VERIFIED: leaanthony/dmg hdiutil_darwin.go:19-23,117,150-157] |
| Country names | Hand-written name table | `Intl.DisplayNames(['en'],{type:'region'})` over a static ISO 3166-1 alpha-2 code list | Names localise; only the code list is static |
| Dark-mode detection in Go | cgo | `app.Env.IsDarkMode()` + `events.Common.ThemeChanged` | Already exposed |

**Key insight:** The native layer is solved by Wails. The work is in the seams (fakes for server-build UAT), the single prefs owner, and correct failure and rollback paths.

## Common Pitfalls

### Pitfall 1: Settings window shows but is not key (typing goes nowhere)
**What goes wrong:** Under the accessory policy, `Show()` alone orders the window but does not activate the app.
**Why:** `windowShow` for a normal window only calls `makeKeyAndOrderFront:`. Activation happens in `windowFocus`.
**How to avoid:** Call `Show()` then `Focus()` (Pattern 1). On macOS 14+, `activateIgnoringOtherApps:` is cooperative. It is expected to succeed because the request follows a user action in our own UI (tray click, ⌘, in the key panel) [ASSUMED]. If the native check shows it fails, the documented fallback is that clicking the window activates an accessory app.
**Warning signs:** Native UAT: the title bar is inactive (grey traffic lights) and the key field ignores typing.

### Pitfall 2: Focus after closing Settings
**What goes wrong:** When the Settings window hides, the accessory app may stay active with no window, and keystrokes go nowhere until the user clicks another app.
**How to avoid:** Capture `plat.FrontmostApp()` when Settings opens. On hide (if the panel is not visible), call `plat.ActivateApp(prev.PID)`, which is best effort and the same mechanism Phase 1 already uses. Do **not** use `app.Hide()` (`[NSApp hide:]`): later `orderFrontRegardless` of the non-activating panel on a hidden app is unverified [ASSUMED risk]. Verify natively.

### Pitfall 3: Paste in the key field
**What goes wrong:** ⌘V does nothing in the Settings key field.
**Why:** WKWebView needs the Edit menu key equivalent. The app menu is shared, so it is present (shell.go:61-64), but it only reaches the **key** window.
**How to avoid:** Pitfall 1 fix. Phase 2 already proved ⌘V into a key field with this menu [VERIFIED: 02-05-SUMMARY.md native truth "Cmd+V pastes a copied key into the installed app's key field (Edit menu from 02-02)"].

### Pitfall 4: The hotkey recorder fires the hotkey instead of recording it
**Why:** Carbon intercepts registered combos before the webview sees them.
**How to avoid:** Suspend all Gifsmith hotkeys while any recorder is recording, and resume on every exit path, with Go guarding on window hide and lost focus (UI-SPEC "Suspend" row).

### Pitfall 5: Recorder key mapping (Option and layouts)
**What goes wrong:** With ⌥ held, `event.key` yields `©`, `˙` and so on. `event.code` is the QWERTY physical position, but Wails binds letters by the active layout's **label** [VERIFIED: global_shortcut_darwin.go:205-233].
**How to avoid:** For letters, use `event.key` when it is a single `a-z` character. Otherwise fall back to `event.code` (`KeyG` → `g`, `Digit5` → `5`, `F13` → `f13`, `Space` → `space`, `ArrowUp` → `up`). Accept that ⌥+letter on non-QWERTY layouts may bind a different key, and the Test button detects that. Document it in the README known-limits.
**Warning signs:** Test reports H7 on a combo that is free.

### Pitfall 6: Duplicate detection on raw strings
`"Ctrl+Option+Cmd+G"` and `"Cmd+Ctrl+Option+G"` are the same shortcut. Compare canonical forms. Either reproduce Wails' sort-and-uppercase in a small Go helper with tests, or rely on `Register`'s own duplicate error and map it to H2. Prefer explicit pre-validation so H2 names the other slot.

### Pitfall 7: Key test returns a cached success
See Pattern 4. Add a regression test that a second `TestKey` within the TTL still makes a request.

### Pitfall 8: Prefs overwrite and racing writers
See Pattern 2. Also note that the panel and Settings windows can both write: the prefs owner must serialise `Update`, and every write must emit `settings:changed` with the new snapshot so both windows converge.

### Pitfall 9: Launch at login from a translocated or dev copy
**What goes wrong:** Registering from a DMG or translocated path, or from `~/Applications/Gifsmith.app` (dev) and later `/Applications` (release), leaves the user unsure which copy launches.
**How to avoid:** G3 heuristic (Pattern 8). In UAT, turn the toggle off on the dev build before installing the release. On the user's main Mac, remove `~/Applications/Gifsmith.app` after installing the release (SingleInstance `UniqueID: "com.kenscott.gifsmith"` [VERIFIED: main.go:90-93] means two copies fight for one instance).

### Pitfall 10: Notarization blocked by the Apple agreement
**Status:** No notarytool keychain profile was found on this Mac. A `security find-generic-password -s com.apple.gke.notary.tool` probe returned "The specified item could not be found". The service name is [ASSUMED] to be notarytool's, so treat this as "no observation of a profile" and confirm with `xcrun notarytool history --keychain-profile gifsmith-notary` (a user-run step; it may raise a keychain access prompt).
**Fix path (user):** The Account Holder accepts the latest agreement at developer.apple.com/account and checks App Store Connect → Agreements, then runs `xcrun notarytool store-credentials gifsmith-notary --apple-id … --team-id S2K7P43927` [CITED: developer.apple.com/forums threads 744210, 821720, 839168 (LOW: forum)].
**How to plan:** All non-notary steps run unattended. Steps 4 and 7 sit behind `checkpoint:human-action` (blocking). Never mark RELS-01 done without `spctl … source=Notarized Developer ID`.

### Pitfall 11: Wrong bundle name and placeholders in the release
`APP_NAME` is lowercase `gifsmith` → `gifsmith.app` and volume `gifsmith`. Rename to `Gifsmith.app` before packaging. Info.plist still has `CFBundleShortVersionString` `0.1.0` [VERIFIED: build/darwin/Info.plist:19-20] and `CFBundleGetInfoString` `This is a comment`. config.yml `version: "0.1.0"` and `comments: "Some Product Comments"`. Template public assets remain (`frontend/public/wails.png`, `svelte.svg`, `bg-*.jpg`, `Inter-Medium.ttf`, `style.css`; `index.html` favicon `/wails.png`). Check usage before deleting (Chesterton).

### Pitfall 12: About version drifting from Info.plist
Keep a Go `const Version = "1.0.0"` exposed to the About pane, plus a unit test that reads `build/darwin/Info.plist` and asserts `CFBundleShortVersionString` matches.

### Pitfall 13: Folder access prompts (TCC)
Writing to Desktop, Documents or a removable volume from a non-sandboxed app triggers a macOS Files & Folders prompt. Info.plist has only `NSDownloadsFolderUsageDescription`. Add `NSDesktopFolderUsageDescription`, `NSDocumentsFolderUsageDescription` and `NSRemovableVolumesUsageDescription` with plain text [ASSUMED: without a purpose string the prompt is unexplained or access fails]. The validation probe (temp write) makes the prompt appear at Choose time, which is the right moment.

### Pitfall 14: Klipy pagination after a page-size change
The Klipy cursor is a page number, so changing `per_page` mid-list makes page N inconsistent, and Klipy forbids dedupe or filtering. On `settings:changed` with a param change, the panel drops each provider's result set and refetches the visible tab's first page on the next `panel:show` (one request).

### Pitfall 15: GIPHY stills are not in the trimmed bundle
`bundle=messaging_non_clips` returns only `original, fixed_height, fixed_height_downsampled, fixed_height_small, fixed_width, fixed_width_downsampled, fixed_width_small` [CITED: developers.giphy.com/docs/optional-settings], so no `*_still`. Constructing `_s.gif` URLs alters vendor URLs. Use the UI-SPEC canvas first-frame path for every provider (one code path). Draw the loaded `<img>` once to a `<canvas>`, then drop the animated `<img>`. Cross-origin images may be drawn, which only taints the canvas, and nothing reads pixels back.

## Code Examples

### Appearance via cgo (new, in `internal/platform/platform_darwin.m`)
```objc
// Source: AppKit NSApplication.appearance (macOS 10.14+) [ASSUMED API availability per Apple docs, not fetched this session]
void gsSetAppearance(int mode) { // 0 system, 1 light, 2 dark
    onMain(^{ @autoreleasepool {
        NSAppearance *a = nil;
        if (mode == 1) a = [NSAppearance appearanceNamed:NSAppearanceNameAqua];
        if (mode == 2) a = [NSAppearance appearanceNamed:NSAppearanceNameDarkAqua];
        NSApp.appearance = a; // nil = follow System
    }});
}
```
The `onMain` helper already exists [VERIFIED: internal/platform/platform_darwin.m:11-13 "schedule on the main run loop in common modes, never dispatch_sync to the main queue"].

### Vite second entry
```ts
// frontend/vite.config.ts (add); key name per Vite 8 BuildOptions [VERIFIED: index.d.ts:2904-2907]
build: { rolldownOptions: { input: { main: 'index.html', settings: 'settings.html' } } },
```
Production serves `/settings.html` from the embedded `frontend/dist` (`AssetFileServerFS`) [VERIFIED: main.go:25-26,107]. In `wails3 dev` and the server-build UAT, Vite serves it directly [ASSUMED for wails3 dev; UAT loads `http://127.0.0.1:9245/settings.html`].

### openExternal allow-list additions
`const EXTERNAL = new Set([GIPHY_KEY_URL, KLIPY_KEY_URL, CLIPS_ACCESS_MAILTO]);` [VERIFIED: frontend/src/lib/api.ts:31]. Add the four About URLs from UI-SPEC (`https://github.com/scottkw/gifsmith`, `…/issues`, `…/blob/main/LICENSE`, `https://github.com/raycast/extensions/tree/main/extensions/gif-search`).

### Strings to change (verbatim today)
- Tray placeholder: `app.Event.Emit("panel:status", "Settings window ships in Phase 4")` [VERIFIED: internal/shell/shell.go:74]. Replace with `OpenSettings`.
- T5a: `"Copied — press ⌘V to paste. To paste automatically, click Enable Automatic Paste… and turn on Gifsmith."` [VERIFIED: internal/services/actions.go:161] → `"Copied — press ⌘V to paste"` (UI-SPEC P14).
- Grid: `export const COLS = 4;` [VERIFIED: frontend/src/lib/grid.ts:2] → column count becomes a parameter.
- Favorites cap: `FavoritesMax = 1000 // sanity cap` [VERIFIED: internal/library/store.go:28] → surface T7c (P23).

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `SMLoginItemSetEnabled` helper apps | `SMAppService.mainAppService` | macOS 13 | Wails uses it automatically for bundled apps |
| `activateIgnoringOtherApps:` | Cooperative `activate` / `yieldToApplication:` | macOS 14 | Self-activation after a user action should work; test natively |
| `build.rollupOptions` | `build.rolldownOptions` | Vite 8 | Use the new key |
| `altool` notarization | `notarytool` | Xcode 13 (altool retired 2023) | `wails3 tool sign --notarize` already uses notarytool |

**Deprecated/outdated:** the Phase 1 tray placeholder toast. `@media (prefers-color-scheme)` theme blocks are replaced by `data-theme`.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `Show()+Focus()` reliably activates the accessory app on macOS 27 after a tray click or ⌘, | Pitfall 1 | Settings not typable; fallback is click-to-activate. Native UAT must check |
| A2 | `ActivateApp(prev)` on Settings hide returns focus acceptably | Pitfall 2 | Keystrokes go nowhere until the user clicks; cosmetic |
| A3 | Setting `NSApp.appearance` also drives WKWebView `prefers-color-scheme` | Pattern 7 | System/Light/Dark CSS could disagree with native chrome; mitigated because `data-theme` is authoritative for CSS |
| A4 | SMAppService `register` while the user disabled it ("requires approval") would re-enable or nag | Pattern 8 | Startup re-apply rule might be over-cautious; harmless |
| A5 | Translocation / `/Volumes/` heuristic is the right G3 trigger | Pattern 8 | G3 shown or missed in edge cases |
| A6 | Wails `Width/Height` for a titled window is the content size | Pattern 1 | Window 28 px off; adjust after first look |
| A7 | Missing folder purpose strings cause unexplained prompts or failures | Pitfall 13 | Add the strings anyway; low cost |
| A8 | `com.apple.gke.notary.tool` is notarytool's keychain service, so its absence means no profile | Pitfall 10 | Profile may exist under another name; the user's `notarytool history` settles it |
| A9 | `wails3 dev` serves `/settings.html` through its Vite proxy | Code Examples | Dev-loop only; production and UAT unaffected |
| A10 | Carbon hotkeys fire while Gifsmith itself is frontmost (Settings key) | Pattern 3 Test | Hotkey Test from Settings would report H7 falsely; verify natively |
| A11 | Canvas first-frame of cross-origin GIFs displays in WKWebView | Pitfall 15 | Reduce Motion stills blank; fallback is to pause by removing `src` |
| A12 | Accepting the agreement resolves the 403 | Pitfall 10 | Forum reports show lapsed memberships needing Apple support; still a user action |

## Open Questions (RESOLVED)

All five resolved by the orchestrator on 2026-10-18 with the recommended option (user standing preference: accept recommendations). Recorded as D-22 (DMG), D-23 (Klipy const) and the plan set in 04-01..04-06.

1. **DMG layout amendment.** RESOLVED: Wails fixed positions (151,164)/(388,164) at 96 px; no `leaanthony/dmg`. The Wails tool cannot place icons at UI-SPEC's (135,170)/(405,170) at 128 px.
   - Recommendation: amend the spec to Wails' (151,164)/(388,164) at 96 px and draw the background to match. Only if the user insists, add the custom `leaanthony/dmg` program behind a legitimacy checkpoint. LICENSE then ships in `Contents/Resources` (satisfies MIT) rather than as a loose DMG icon.
2. **Klipy clearance mechanism.** RESOLVED: compile-time const, default false. UI-SPEC says it is not user-visible. Recommendation: `const KlipyFileActions = false` in `internal/services`, exposed read-only to the frontend in the settings snapshot. Flipping it needs a rebuild and release, which matches "developer flips it once the written answer arrives."
3. **Phase size.** RESOLVED: 6 plans, no phase split. About 14 requirements plus about 14 polish items plus release. Recommendation: ≤6 plans as listed in the Summary. If the planner needs more, move RELS-01..04 (plans 5-6) to Phase 5 as the roadmap allows.
4. **Notarization timing.** RESOLVED: early reminder in 04-01; blocking checkpoint:human-action in 04-05. Is the 403 resolved yet? Unknown (no profile observed). Recommendation: the first plan carries an early non-blocking reminder, and the release plan's step 4 is a blocking `checkpoint:human-action`.
5. **Klipy email.** RESOLVED: draft only; user sends it; gate stays closed. Draft text is a deliverable (CONTEXT). Sending it is the user's action, and the gate stays closed regardless of timing.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | build/tests | ✓ | 1.27.2 | — |
| node / pnpm | frontend | ✓ | v24.14.1 / pnpm 12.14.0 (corepack-downloaded on first call) | — |
| wails3 CLI | build, sign, dmg, `wails3 task` | ✓ | v3.0.0-beta.30 | — |
| `task` binary | — | ✗ | — | Use `wails3 task …` (embedded Task) |
| Developer ID identity | signing | ✓ | 1 "Developer ID Application" identity (`security find-identity`) | — |
| notarytool / stapler | RELS-01 | ✓ | notarytool 1.1.3 (42) | — |
| Notary keychain profile `gifsmith-notary` | RELS-01 | ✗ (not observed) | — | **None. Blocking user action (Apple agreement 403)** |
| hdiutil, ditto, spctl, shasum, sips, iconutil, tiffutil | DMG/verify/checksum/background | ✓ | system | — |
| gh CLI (auth scottkw) | GitHub Release | ✓ | 2.102.0 | Manual upload via web |
| Fresh macOS user account | RELS-02 | user-provided | — | None. User-run |

**Missing dependencies with no fallback:** notarization credentials (user must resolve the Apple agreement and run `store-credentials`).
**Missing dependencies with fallback:** `task` → `wails3 task`.

## Validation Architecture

`workflow.nyquist_validation` is `false` in `.planning/config.json`, so this section is omitted per instructions. The test strategy stays the established one: `go test -race ./internal/...` (scoped packages, never `go build ./...`), `pnpm --dir frontend test` (node:test) and `pnpm --dir frontend check`, headless UAT via `scripts/smoke-browser-loop.sh --uat scripts/uat/<n>.js` (new scripts must load both `/` and `/settings.html`, and use `page.emulateMedia({ reducedMotion: 'reduce', colorScheme })` for the Reduce Motion and theme backstops), and native behaviors at a final user checkpoint. Server-build fakes needed: `Registrar` (with an env-selected failing combo, e.g. `FAKE_HOTKEY_FAIL`), `Autostart` (in-memory, with an env-selected failure), `PickFolder` (returns an env path under `bin/`), and `platform.Fake` gains `SetAppearance` / `SystemLocale`.

## Security Domain

`security_enforcement: true`, ASVS level 1, block on high.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | No user auth; vendor API keys are secrets (V8/V6) |
| V3 Session Management | no | — |
| V4 Access Control | yes (local) | Bound methods take no paths or URLs from JS; the external-URL allow-list; Klipy gate enforced in Go |
| V5 Input Validation | yes | Go re-validates every enum pref, page size, accelerator (allow-listed key names, modifier required, denylist), folder (picker-only + Statfs/link probe), key format (`validKey`) |
| V6 Cryptography | no | No crypto; Keychain via go-keyring |
| V8 Data Protection | yes | Keys only in the Keychain; `KeyStatus` returns last 4 only; logs name provider and kind only; prefs 0600 and never hold secrets |
| V10 Malicious Code / V14 Config | yes | Hardened runtime, no entitlements, Developer ID + notarization + staple, no new dependencies, SHA-256 published |
| V12 Files & Resources | yes | Download dir validated; `os.Link` claim never overwrites; LICENSE in bundle |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Webview supplies an arbitrary download path (write anywhere) | Tampering | No path-setting binding; path only from the native picker; validation before persist |
| Malformed accelerator / key name crashes or binds unexpected key | Tampering / DoS | Go-side parse via allow-list, map Wails parse errors to H4b, never log raw untrusted input beyond the canonical form |
| settings.json edited to enable Klipy file actions | Tampering (vendor policy) | Gate is a compile-time const, not a pref |
| Key fragments leak through status/errors | Information disclosure | last4 only; fixed-text Keychain errors (existing pattern settings.go:31-47) |
| Hotkeys left suspended (feature silently dead) | DoS (local) | Go resumes on window hide and lost focus; test for it |
| Tampered DMG on download | Tampering | Notarized + stapled app and DMG; published SHA-256; Gatekeeper `spctl` check in release verify |
| Notary credentials committed | Information disclosure | Keychain profile only; Taskfile references the profile name, never credentials |
| Opening unlisted URLs | Spoofing | Extend the fixed `EXTERNAL` set only with the four About URLs |

## Sources

### Primary (HIGH confidence, read this session)
- Wails v3 beta.30 module source (`~/go/pkg/mod/github.com/wailsapp/wails/v3@v3.0.0-beta.30`): `pkg/application/global_shortcut_manager.go`, `global_shortcut_darwin.go`, `global_shortcut_unsupported.go`, `keys.go`, `roles.go`, `webview_window.go`, `webview_window_options.go`, `webview_window_darwin.go`, `webview_window_darwin.m`, `webview_panel_darwin.m`, `dialogs.go`, `dialogs_darwin.go`, `application_server.go`, `autostart*.go`, `application_darwin.go`, `events_common_darwin.go`, `menuitem.go`; `internal/commands/sign.go`, `tool_package.go`; `pkg/events/events.go`
- `github.com/leaanthony/dmg@v0.0.0-20260731074841-5c28840cf819` (README, `cmd/mkdmg/main.go`, `dmg/options.go`, `dmg/hdiutil_darwin.go`)
- `github.com/zalando/go-keyring@v0.2.8/keyring_darwin.go`
- Project code: `main.go`, `wire_*.go`, `internal/{shell,settings,services,platform,provider,search,library}`, `frontend/{vite.config.ts,package.json,src/lib/{api,actions,grid}.ts}`, `build/darwin/{Taskfile.yml,Info.plist}`, `build/Taskfile.yml`, `Taskfile.yml`, `build/config.yml`
- Local tool probes: `wails3 version`, `wails3 tool package --help`, `wails3 tool sign --help`, `codesign -dvv ~/Applications/Gifsmith.app`, `xcrun notarytool --version`, `security find-identity`, `gh auth status`, `gh repo view scottkw/gifsmith`, `man codesign` (--timestamp default text)

### Secondary (MEDIUM, official docs via WebFetch)
- https://developers.giphy.com/docs/api/endpoint (search/trending params; `lang` search-only; rating values; beta limit 50)
- https://developers.giphy.com/docs/optional-settings/ (bundle renditions; language codes)
- https://docs.klipy.com/gifs-api/gifs-search-api.md (`per_page` 8-50, `locale` ISO 3166 alpha-2, `customer_id` optional)

### Tertiary (LOW, forums/web search)
- Apple Developer Forums on the notarytool 403 agreement error: [744210](https://developer.apple.com/forums/thread/744210), [821720](https://developer.apple.com/forums/thread/821720), [839168](https://developer.apple.com/forums/thread/839168), [744226](https://developer.apple.com/forums/thread/744226)
- Accessory-app activation on macOS 14: [Apple Forums 741701](https://developer.apple.com/forums/thread/741701)
- SMAppService errors: [Apple Forums 721598](https://developer.apple.com/forums/thread/721598), [SMAppService docs](https://developer.apple.com/documentation/servicemanagement/smappservice.md)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH. No new dependencies; every API read in pinned source.
- Architecture: HIGH for Go seams and prefs ownership. MEDIUM for native activation and focus return (needs native UAT).
- Pitfalls: HIGH for the prefs overwrite, cache, hotkey canonicalisation, DMG layout and release steps (source-verified). MEDIUM for the TCC folder prompts and the login-item edge cases.
- Release: MEDIUM. Signing is proven on this Mac; notarization has never run (blocked).

**Research date:** 2026-10-18
**Valid until:** 2026-11-01 (Wails pinned, but beta nightlies move fast; re-check if the pin changes)
