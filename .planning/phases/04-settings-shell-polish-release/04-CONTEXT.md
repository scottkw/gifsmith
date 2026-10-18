# Phase 4: Settings, Shell Polish & Release - Context

**Gathered:** 2026-10-18
**Status:** Ready for planning

<domain>
## Phase Boundary

Everything is configurable from a proper Settings window (keys, hotkeys, launch at login, theme, search/display/download options, default actions, Accessibility status), and a stranger can install a Developer ID–signed, notarized arm64 DMG on a clean Mac and use it within vendor (GIPHY, Klipy) and license (Raycast extension MIT) obligations. Requirements: SETT-01..06, SRCH-09, ACTN-08, SHEL-05, SHEL-06, RELS-01..04.

</domain>

<decisions>
## Implementation Decisions

### Settings Window
- A separate normal window (~640×520) with a sidebar: General · Keys · Hotkeys · Search · Downloads · About. Opened from the tray menu "Settings…" and ⌘, in the panel. The window activates the app so typing and ⌘V paste work (SETT-01).
- Changes apply immediately (no Save button), Mac-style. Exception: keys are verified with a test call before saving (Phase 2 SaveKey path).
- Keys tab: one row per provider (GIPHY, Klipy) with status (Saved ✓ / Not set / Rejected), last 4 characters of the key, and Change / Test / Remove buttons plus "Get a free key ↗" (GIPHY developer dashboard, Klipy Partner Panel). Remove asks for confirmation. The webview still never receives a full key (SETT-02).
- General tab: Theme (System / Light / Dark, default System — overrides both the webview colors and the native window appearance, SETT-06), Launch at login toggle (SHEL-05), Accessibility status with a Grant button (SETT-05), default Enter and ⌘Enter actions (ACTN-08; defaults Copy GIF / Paste GIF).

### Hotkeys & Launch at Login
- Click-to-record hotkey field: press a combination requiring at least one of ⌘/⌥/⌃; it registers immediately; if registration fails the old binding stays and an error shows; Esc cancels recording; "Reset to default".
- Three hotkeys: Open panel (default ⌃⌥⌘G), Open Favorites (no default), Open Recents (no default); Favorites/Recents may be left unset (SHEL-06).
- Conflict handling: macOS doesn't report clashes with other apps (Carbon RegisterEventHotKey succeeds anyway), so: block duplicates among Gifsmith's own hotkeys, show an error when registration fails, and offer a "Test" button that opens the panel to confirm (SETT-03).
- Launch at login off by default; toggle uses Wails `app.Autostart` (SMAppService on macOS 13+); the user's choice is persisted and re-applied (idempotently) at startup.

### Search, Display & Download Settings
- Grid item size Small / Medium / Large, separately for search and trending: Small = 5 columns, Medium = 4 (current), Large = 3; panel stays 720×480.
- Page size ("max results") 24 / 36 / 48, clamped to each provider's limits. GIPHY language and Klipy country as dropdowns defaulting to the system locale. GIPHY content rating G / PG / PG-13 / R, default PG-13 (SRCH-09).
- Download folder: "Choose…" opens the native folder picker, plus "Reset to Downloads"; stored by path. Picking an exFAT/FAT or network folder shows a clear error (the atomic os.Link claim needs a hard-link-capable volume). "Hide original filename" saves as `gifsmith-<id>` (existing `FileStem(g, hide)`) (SETT-04, ACTN-11).
- Reduce Motion: follow the macOS accessibility setting — tiles show a still first frame and only the selected tile animates. No separate toggle.

### Release & Compliance
- Build target: Apple Silicon (arm64) DMG for macOS 13+, drag-to-Applications layout, version 1.0.0 (RELS-01).
- Notarization: build and wire every release step now; notarize once the user resolves the Apple Developer agreement HTTP 403 (`notarytool store-credentials`). If still blocked at release time, the phase pauses at a blocking checkpoint asking the user to resolve it — never skipped silently, never shipped un-notarized as "done".
- Klipy file actions (Copy/Paste GIF, Copy/Paste Square, Download) are disabled until written clearance is recorded: on Klipy items they're hidden, Enter copies the link and ⌘Enter pastes the link; a setting records the clearance (RELS-04). Draft the email to developers@klipy.com (temp-file copies + ad labelling IN-09) for the user to send.
- Distribution: GitHub Release on scottkw/gifsmith with the DMG, a SHA-256 checksum and release notes; README covering install, getting keys, GIPHY/Klipy attribution and credits. No auto-updater in v1. About screen credits the original MIT-licensed Raycast GIF Search extension and its authors (RELS-03; LICENSE already carries the notice).
- Clean-account install test (RELS-02) is user-verified on a fresh macOS user account.

### Claude's Discretion
- Settings persistence schema (extend Phase 2 prefs JSON, atomic 0600), how settings changes propagate to the panel (Wails events), window component structure, hotkey recorder implementation details, how theme overrides NSAppearance, DMG tooling (Wails Taskfile `darwin:package:dmg`), and test strategy (Go -race, node:test, headless UAT; native behaviors user-verified).

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/shell` (panel, tray menu with Open/Settings/Quit placeholder, `GlobalShortcut` registration, SingleInstance from Phase 3), `internal/platform` (cgo: pasteboard, AX trust check + prompt, NSWorkspace via onMain, RevealFile, OpenURL), `internal/settings` (Keychain SecretStore, Prefs atomic JSON, `WriteJSONAtomic`), `internal/services` (GifService, SettingsService.SaveKey verify-before-save, ActionService with DownloadsDir seam + `FileStem(g, hide)`, LibraryService).
- Frontend: `App.svelte`, `SetupCard.svelte` (key entry + K2–K6 messages), `lib/actions.ts` (action table; default Enter/⌘Enter), `lib/providers.ts` (slots, key URLs), `app.css` tokens (light/dark via prefers-color-scheme).
- Build: `build/darwin/Taskfile.yml` (install:dev, package, dmg, sign:notarize), Developer ID identity present; `build/darwin/Info.plist` (LSUIElement, NSDownloadsFolderUsageDescription).

### Established Patterns
- Go owns HTTP, files and native calls; the webview gets typed results/booleans only, never keys or raw URLs in errors.
- TDD (RED then GREEN), headless UAT via `scripts/smoke-browser-loop.sh` + `scripts/uat/*.js` against the `-tags server` build with `platform.Fake` and fakeapi; native behaviors user-verified at a final checkpoint after `wails3 task darwin:install:dev`.
- Sequential plans on the main tree (use_worktrees=false); install:dev adds stretchr/objx to go.sum — restore it.

### Integration Points
- Tray "Settings…" item and ⌘, → open the Settings window (second Wails window, not the NSPanel).
- Hotkey changes → `shell` re-registers via Unregister/Register; Favorites/Recents hotkeys open the panel on ⌘5/⌘6.
- Settings changes (grid size, page size, locale, rating, theme, default actions, Klipy clearance) → Prefs → events to the panel; providers read rating/locale/page size per request.

</code_context>

<specifics>
## Specific Ideas

- Fold in `.planning/todos/pending/phase-3-ui-review-polish.md` (resolves_phase 4): destructive-row contrast, footer `aria-live`, shorter T5a message, right-click opens the action menu, Details "Esc Back" as a button, tertiary-text contrast, hide library hints on empty tabs.
- Consider the 7 open Phase 3 review Info items (03-REVIEW-DISPOSITION.md) where code is touched anyway, and surface the favorites cap (1000, oldest dropped silently).
- Settings window should feel native (system font, sidebar list), consistent with the panel's tokens.

</specifics>

<deferred>
## Deferred Ideas

- Universal (Intel) build, website distribution, Sparkle auto-updates — post-v1.
- Mac App Store — out of scope (sandbox forbids synthesized ⌘V).

</deferred>
