# Phase 3: Action Pipeline & Library - Context

**Gathered:** 2026-10-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Every action the Raycast extension offered works on the selected GIF (⌘K action menu: link/Markdown copy and paste, page link, open in browser, download, square crop, Clips as MP4), and Favorites, Recents, a details view and recent searches put reused GIFs one keystroke away — rendered from saved metadata, so they work offline or rate-limited. Temp/staged files are cleaned up promptly (Klipy eagerly). Requirements: ACTN-03..07, ACTN-09..11, LIBR-01..05, SRCH-08. Settings UI (default-action choice ACTN-08, download folder picker, hide-filename toggle, hotkeys for Favorites/Recents SHEL-06) is Phase 4 — this phase uses the defaults.

</domain>

<decisions>
## Implementation Decisions

### Action Menu (⌘K)
- ⌘K opens a Raycast-style popover list over the grid, anchored to the selected tile, with a filter-as-you-type field; Esc closes it back to the grid with the same selection.
- Action order (extension order, defaults first): Copy GIF (↵) · Paste GIF (⌘↵) · Copy Link · Paste Link · Copy Markdown · Paste Markdown · Copy Page Link · Open in Browser · Download · Copy Square · Paste Square · Add/Remove Favorite · Show Details. Actions that don't apply (square for Clips) are omitted, not greyed.
- Direct shortcuts for common actions, shown in the menu: ⌘⇧C copy link, ⌘⇧M copy Markdown, ⌘O open in browser, ⌘S download, ⌘⇧F favorite toggle, ⌘I details. Everything else via ⌘K.
- Feedback mirrors Phase 2 copy: toast ("Copied link", "Downloaded to Downloads"), copy/paste actions hide the panel after 400 ms; Download, Favorite and Details keep the panel open.

### Favorites & Recents
- Two extra tabs after the providers: ★ Favorites (⌘5) and Recents (⌘6), same grid and keyboard model; the search box filters them by title (local, no API).
- Grouped by provider in sections with small headers (GIPHY / Clips / Klipy / Finer Gifs), never blended within a section; newest first per section.
- Recents: last 50, recorded on every successful copy/paste of GIF, link, Markdown or square; re-using an item moves it to the top (no duplicates).
- ⌘⌫ removes the selected item (Recents) or unfavorites it (Favorites); also in ⌘K, plus a "Clear Recents" action. No confirmation dialog — a 5 s "Undo" toast instead.
- Both render from saved metadata snapshots only (no API calls, no persistent media cache — previews load from provider URLs as today).

### Details View & Recent Searches
- Details (⌘I) is a pane over the grid: large animated preview left, metadata right (title, provider, dimensions, file size when provided, tags, page link, source link). Esc returns to the grid with selection kept.
- Metadata comes only from fields the search result already carries — no extra API call; missing fields are hidden.
- Empty search box shows up to 8 recent-search chips under the box, above trending; ↑ from the first grid row moves into the chips; Enter reruns one. Kept per provider.
- A search is saved when a GIF from it is copied/pasted, or after 1.5 s of no typing (avoids half-typed queries); last 20 per provider, no duplicates; ⌘⌫ removes a chip.

### Files, Clips MP4 & Square Crop
- Download: `<title-slug>-<id>.gif` (or `.mp4`) into ~/Downloads; ` (2)`, ` (3)` suffix on collision, never overwrite; validated as real GIF/MP4 before saving (error pages rejected). Phase 4's hide-filename setting switches to `gifsmith-<id>`. Toast "Downloaded to Downloads" with Show in Finder.
- Clips: the 720p MP4 file goes on the clipboard via the same file-URL path; labelled "MP4" on the tile and in menus; no square crop; paste works the same.
- Square crop: center crop in Go (`image/gif`, lossless — proven in research), on demand at copy time, staged like a normal copy; busy indicator after 300 ms; GIFs over 15 MB show "too large to crop".
- Staging holds only the latest file; wiped at launch and at quit; Klipy staged files deleted right after the paste completes or on the next copy. Downloads never go through staging.

### Claude's Discretion
- Storage format/location for favorites, recents and recent searches (JSON in the app's settings dir, atomic writes like Phase 2 prefs) and the metadata snapshot shape.
- Exact popover component structure, shortcut-conflict handling with existing keys, toast/Undo implementation, and how the Clips 720p rendition is chosen when 720p is missing (nearest available).
- How tests/UAT cover clipboard file types without synthetic input (headless UAT + Go tests; native paste of MP4 user-verified).

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/services/actions.go` — ActionService Copy/Paste (busy try-lock, staging download + validate, file-URL pasteboard, ⌘V synthesis behind Accessibility check).
- `internal/staging/staging.go` — Fetch (allowlist checked on every redirect, ≤5 hops), Sweep, VendorAllow; extend for MP4, square-crop output and eager Klipy deletion.
- `internal/platform` — CopyFile (NSPasteboard file URL), FocusWebView, frontmost-app helpers (cgo, main-thread rules via onMain).
- `internal/settings/settings.go` — atomic 0600 JSON prefs + Keychain store; pattern for favorites/recents files.
- `internal/provider` Gif/Page models (GIPHY, Klipy, Finer Gifs mappers) — carry the metadata details/favorites need; Clips items carry `video.assets` renditions.
- Frontend: `App.svelte` (per-provider state, tabs, keyboard routing), `components/ResultGrid.svelte` + `Tile.svelte` (virtualized grid), `ProviderTabs.svelte`, `StateMessage.svelte`, `lib/grid.ts` (routeKey — ⌘K and Tab already captured as no-ops), `lib/api.ts` (`latest()`, openExternal allow-list), `lib/errors.ts`.

### Established Patterns
- Provider HTTP and file work live in Go; the webview only sees typed results and booleans (no keys, no URLs in errors).
- TDD with RED commits; Go `-race` tests; frontend node:test in `src/lib/*.test.ts`; headless UAT scripts in `scripts/uat/*.js` against the `-tags server` build + fakeapi; native behaviors user-verified at checkpoints.
- Toast "Copied — press ⌘V to paste" then hide after 400 ms; per-provider state map; never blend providers; Klipy never deduped/reordered.

### Integration Points
- New bound service(s) for library (favorites/recents/recent searches) wired in `main.go` + `wire_server.go`/`wire_desktop.go`; generated bindings under `frontend/bindings/`.
- ⌘K / ⌘5 / ⌘6 / shortcut routing in `lib/grid.ts` `routeKey`; tabs extended in `lib/providers.ts`.
- Recents recorded from ActionService success paths; recent searches from the search flow in App.svelte.

</code_context>

<specifics>
## Specific Ideas

- Keep it feeling like the Raycast extension: same action set and order, Raycast-style action panel.
- Fold in the Phase 2 UI-review polish todo (`.planning/todos/pending/phase-2-ui-review-polish.md`, resolves_phase 3): shared accent focus ring in app.css, styled footer paste button, Clips-tab active styling, UI-SPEC C2 copy update, and the minor spacing/states items.
- Phase 2 code-review Info items worth fixing while touching the same code: IN-07 (hide-after-copy timer never cancelled — re-show within 400 ms hides again), IN-04 (stale Clips error until Try Again), staged files not deleted at launch/quit (now ACTN-10).

</specifics>

<deferred>
## Deferred Ideas

- Settings window: choosing default Enter/⌘Enter actions (ACTN-08), download folder picker and hide-filename toggle (SETT-04), Favorites/Recents hotkeys (SHEL-06), theme Light/Dark/System (SETT-06), key change + sign-up links (SETT-02) — all Phase 4.
- Klipy file-action gating on written clearance and ad labelling (IN-09) — Phase 4 release gate.

</deferred>
