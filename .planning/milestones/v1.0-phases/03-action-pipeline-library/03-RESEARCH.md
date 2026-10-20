# Phase 3: Action Pipeline & Library - Research

**Researched:** 2026-10-17
**Domain:** Go action pipeline (clipboard text/file, download, square crop, MP4), local JSON library store, Svelte 5 keyboard-driven overlays, Wails v3 beta.30 lifecycle
**Confidence:** HIGH for the Go/code-level findings (read in source, crop and link behavior probed). MEDIUM for native app behaviors (MP4 paste targets, TCC Downloads prompt, WebKit key equivalents), which only the user can verify.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Action Menu (⌘K)
- ⌘K opens a Raycast-style popover list over the grid, anchored to the selected tile, with a filter-as-you-type field; Esc closes it back to the grid with the same selection.
- Action order (extension order, defaults first): Copy GIF (↵) · Paste GIF (⌘↵) · Copy Link · Paste Link · Copy Markdown · Paste Markdown · Copy Page Link · Open in Browser · Download · Copy Square · Paste Square · Add/Remove Favorite · Show Details. Actions that don't apply (square for Clips) are omitted, not greyed.
- Direct shortcuts for common actions, shown in the menu: ⌘⇧C copy link, ⌘⇧M copy Markdown, ⌘O open in browser, ⌘S download, ⌘⇧F favorite toggle, ⌘I details. Everything else via ⌘K.
- Feedback mirrors Phase 2 copy: toast ("Copied link", "Downloaded to Downloads"), copy/paste actions hide the panel after 400 ms; Download, Favorite and Details keep the panel open.

#### Favorites & Recents
- Two extra tabs after the providers: ★ Favorites (⌘5) and Recents (⌘6), same grid and keyboard model; the search box filters them by title (local, no API).
- Grouped by provider in sections with small headers (GIPHY / Clips / Klipy / Finer Gifs), never blended within a section; newest first per section.
- Recents: last 50, recorded on every successful copy/paste of GIF, link, Markdown or square; re-using an item moves it to the top (no duplicates).
- ⌘⌫ removes the selected item (Recents) or unfavorites it (Favorites); also in ⌘K, plus a "Clear Recents" action. No confirmation dialog — a 5 s "Undo" toast instead.
- Both render from saved metadata snapshots only (no API calls, no persistent media cache — previews load from provider URLs as today).

#### Details View & Recent Searches
- Details (⌘I) is a pane over the grid: large animated preview left, metadata right (title, provider, dimensions, file size when provided, tags, page link, source link). Esc returns to the grid with selection kept.
- Metadata comes only from fields the search result already carries — no extra API call; missing fields are hidden.
- Empty search box shows up to 8 recent-search chips under the box, above trending; ↑ from the first grid row moves into the chips; Enter reruns one. Kept per provider.
- A search is saved when a GIF from it is copied/pasted, or after 1.5 s of no typing (avoids half-typed queries); last 20 per provider, no duplicates; ⌘⌫ removes a chip.

#### Files, Clips MP4 & Square Crop
- Download: `<title-slug>-<id>.gif` (or `.mp4`) into ~/Downloads; ` (2)`, ` (3)` suffix on collision, never overwrite; validated as real GIF/MP4 before saving (error pages rejected). Phase 4's hide-filename setting switches to `gifsmith-<id>`. Toast "Downloaded to Downloads" with Show in Finder.
- Clips: the 720p MP4 file goes on the clipboard via the same file-URL path; labelled "MP4" on the tile and in menus; no square crop; paste works the same.
- Square crop: center crop in Go (`image/gif`, lossless — proven in research), on demand at copy time, staged like a normal copy; busy indicator after 300 ms; GIFs over 15 MB show "too large to crop".
- Staging holds only the latest file; wiped at launch and at quit; Klipy staged files deleted right after the paste completes or on the next copy. Downloads never go through staging.

### Claude's Discretion
- Storage format/location for favorites, recents and recent searches (JSON in the app's settings dir, atomic writes like Phase 2 prefs) and the metadata snapshot shape.
- Exact popover component structure, shortcut-conflict handling with existing keys, toast/Undo implementation, and how the Clips 720p rendition is chosen when 720p is missing (nearest available).
- How tests/UAT cover clipboard file types without synthetic input (headless UAT + Go tests; native paste of MP4 user-verified).

### Deferred Ideas (OUT OF SCOPE)
- Settings window: choosing default Enter/⌘Enter actions (ACTN-08), download folder picker and hide-filename toggle (SETT-04), Favorites/Recents hotkeys (SHEL-06), theme Light/Dark/System (SETT-06), key change + sign-up links (SETT-02) — all Phase 4.
- Klipy file-action gating on written clearance and ad labelling (IN-09) — Phase 4 release gate.

### Also folded in (CONTEXT "Specific Ideas", not deferred)
- Phase 2 UI-review polish todo (`.planning/todos/pending/phase-2-ui-review-polish.md`): shared accent focus ring in app.css, styled footer paste button, Clips-tab active styling, UI-SPEC C2 copy update, minor spacing/states items (03-UI-SPEC P1–P11).
- Phase 2 review Info items: IN-07 (hide-after-copy timer never cancelled), IN-04 (stale Clips error until Try Again), IN-03 (staged files not deleted at launch/quit, now ACTN-10).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ACTN-03 | Copy or paste the GIF link, copy or paste GIF Markdown | New `Platform.CopyText` (cgo, private-pasteboard testable); text built in Go from `Gif` fields (Pattern 3); GIPHY link = query-stripped URL, Klipy URL unaltered (Pitfall 6); paste reuses the Phase 1 paste sequence |
| ACTN-04 | Copy the page link, open the GIF page in the browser | New `PageURL` field: GIPHY `url` only; Klipy native v1 and Finer Gifs carry none, so those rows are omitted (Provider Field Map); open via Go `Platform.OpenURL` with an https + vendor-host check, never the webview's `Browser.OpenURL` (Pitfall 9) |
| ACTN-05 | Download the GIF to a chosen folder (default Downloads) | `staging.Save` writes a hidden temp in the destination, validates magic bytes, claims the name with `os.Link` (probe-verified no-overwrite), `DownloadDir` func injected for Phase 4 |
| ACTN-06 | Copy or paste a center-cropped square GIF (not Clips) | Geometry-only crop with `image/gif` re-verified today: 0 pixel mismatches on 6 real GIFs + a synthetic GIF that has a frame fully outside the window (Code Example 1) |
| ACTN-07 | ⌘K action menu listing all actions | Pure `lib/actions.ts` (list builder, filter, shortcut map) + `ActionMenu.svelte`; routing additions in `lib/grid.ts` (Pattern 7) |
| ACTN-09 | Clips handled as MP4 (labelled, 720p default) | `clips_grid_picker` bundle carries `video.assets.{360p,480p,720p,1080p}`; 360p always present; rendition order 720p → 480p → 1080p → 360p; MP4 validated by the `ftyp` box |
| ACTN-10 | Temp files deleted promptly; Klipy eagerly | `Store.Clear()` at launch (before `app.Run`) and in `ActionService.ServiceShutdown`; sweep after every clipboard write (file or text); Klipy grace timer after ⌘V (Pitfall 4) |
| ACTN-11 | Downloads validated, never overwrite or collide; hide-filename option | Magic-byte check before any byte reaches the destination; `os.Link` claim loop ` (2)`, ` (3)`; filename stem function takes a `hide bool` (Phase 4 wires the setting) |
| LIBR-01 | Add/remove favorites | `library.Store` (JSON, atomic 0600) + bound `LibraryService.SetFavorite` |
| LIBR-02 | Every successful copy/paste records the GIF in Recents | `ActionService` calls `Recents.Touch(g)` on every `Result.Copied` for file/square/link/Markdown; cap 50, move-to-top |
| LIBR-03 | Browse Favorites/Recents grouped by provider; remove from Recents | Pure `lib/library.ts` sections + virtual rows + cross-section navigation; Go-side one-slot Undo |
| LIBR-04 | Render from saved metadata, no API calls | Snapshot = the full extended `provider.Gif` + timestamp; previews load from stored provider URLs |
| LIBR-05 | Details view (title, size, dimensions, tags, source links) | New optional `Gif` fields: `width`, `height`, `size`, `tags`, `pageUrl`, `sourceUrl`, `mediumUrl` (Provider Field Map) |
| SRCH-08 | Recent searches when the box is empty, rerunnable | `searches.json` (20 per provider, dedupe, newest first) via `LibraryService.SaveSearch/RemoveSearch`; 1.5 s idle timer + save-on-copy in App.svelte |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

From `/Users/ken/dev/CLAUDE.md` and `/Users/ken/dev/gif-search/.claude/CLAUDE.md`:

- **Stack:** Go + Wails v3 `v3.0.0-beta.30` pinned (Go module, `wails3` CLI, `@wailsio/runtime`); never `@latest`. Svelte 5 + TS frontend. pnpm only (`PACKAGE_MANAGER` default is pnpm in `Taskfile.yml`).
- **Wails isolation:** Wails-specific code stays in `internal/shell` (and `main.go` wiring); domain packages (providers, staging, crop, library store) have no Wails import.
- **cgo:** only in `internal/platform` with `//go:build darwin && !server` plus the stub; AppKit work hops to main via `onMain`; pasteboard calls stay on the caller's thread; never call `onMain` from a test-exercised path (no main run loop under `go test`).
- **Secrets:** API keys never in the binary, JSON prefs, logs, errors or the webview. Library files are non-secret.
- **Go conventions:** `go fmt`, `golangci-lint`, context-aware bound methods (`ctx context.Context`).
- **JS/TS:** camelCase, PascalCase components, TypeScript types.
- **Testing:** 80%+ coverage on critical components; integration tests for bound services.
- **Silent fallbacks forbidden:** `or {}`-style recovery that hides a hard failure is not allowed. A corrupt library file must surface an error, never be silently reset (Pitfall 11).
- **Chesterton's Fence:** before removing `ErrClipsCopy`, the server-build stub, or `CLIPS_COPY`, state why they existed (all three are Phase 2 placeholders for "Clips copy ships in Phase 3" and "the CGO-free build cannot touch NSPasteboard").
- **Premature abstraction:** no generic "action framework"; a flat set of bound methods is enough (3 real variants before abstracting).
- **Safety (primary Mac):** agents never send synthetic desktop input, never kill processes they did not start, never read the Keychain, never use real API keys. Native behaviors are user-verified at checkpoints.
- **UAT:** browser UAT uses the dev-browser skill against the `-tags server` build; never ask the user to do what dev-browser can do.
- **Ponytail** is active on executor agents: smallest change that solves the task.
- **GSD:** edits only through GSD workflows.

## Summary

Phase 3 is mostly plumbing on top of a codebase that already has the hard parts: a file-URL pasteboard writer, an Accessibility-gated paste sequence, an allowlisted staging downloader, atomic JSON prefs, a virtualized grid and a pure key router. Nothing new needs to be installed. The work splits cleanly into Go (provider fields, text clipboard, download, square crop, MP4, staging lifecycle, library store) and Svelte (action menu, details pane, library tabs, chips, toasts with actions, polish).

The three questions with real answers behind them:
1. **Metadata is uneven by provider.** GIPHY gives a page URL (`url`), a source URL (`source`), dimensions and byte size (strings) but **no tags** in its documented GIF object. Klipy native v1 gives dimensions, size and `tags` but **no page URL**. Finer Gifs gives only a file id and text. Clips give `video.assets` with width/height/url but **no size**. So "Copy Page Link" and "Open in Browser" exist only for GIPHY and Clips, and every Details field must be optional.
2. **The square crop is solved.** The geometry-only `image/gif` crop (SubImage + re-based rectangle, original palette/delay/disposal, 1×1 transparent placeholder for frames fully outside the window) produced 0 pixel mismatches against a disposal-aware reference render on 6 real GIFs (up to 182 frames) and a synthetic GIF with every disposal type. Re-verified today.
3. **"Delete Klipy right after the paste completes" has no observable completion.** `⌘V` is posted and the target app reads the file later. An immediate delete can break the paste (Slack is the likely victim). Use a short grace timer after the ⌘V and have the user verify it in Messages and Slack.

**Primary recommendation:** Extend `provider.Gif` with seven optional fields filled by the three mappers, add `CopyText`/`RevealFile`/`OpenURL` to `platform.Platform`, grow `ActionService` with flat bound methods (CopyText/PasteText, CopySquare/PasteSquare, Download, ShowDownload, OpenPage, OpenSource), add `internal/imaging` (square crop) and `internal/library` (JSON store), and keep all UI logic that can be pure (action list, section layout, navigation, formatting, popover placement) in `frontend/src/lib/*.ts` with node tests.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Extract page/source URL, dimensions, size, tags, medium preview, Clips MP4 rendition | Go provider mappers (`internal/provider/*`) | — | Only Go sees vendor JSON; the webview gets typed `Gif` |
| Build link / Markdown / page-link text | Go `ActionService` | — | Format rules (GIPHY query strip, Klipy unaltered, Markdown escaping) unit-testable in Go; webview never composes clipboard payloads |
| Text and file clipboard writes, ⌘V, Show in Finder, open URL | Go `internal/platform` (cgo) | `ActionService` (sequencing, busy lock) | AppKit-only APIs; main-thread rules live in one file |
| Download fetch + validation + no-overwrite naming | Go `internal/staging` (`Save`) | `ActionService` | Shares the allowlisted, redirect-checked fetch with staging |
| Square crop | Go `internal/imaging` (new, pure stdlib) | `staging` (input fetch with 15 MB cap) | CPU/memory work, no Wails import |
| Staging lifecycle (launch/quit wipe, Klipy grace delete) | Go `ActionService` + `staging.Store` | `main.go` (launch call), Wails `ServiceShutdown` | Process lifecycle is Go's |
| Favorites / Recents / recent-search persistence | Go `internal/library` (new) | bound `LibraryService` | Atomic file writes, caps, undo, validation in one place; survives webview reloads |
| Recents recording | Go `ActionService` success paths | — | Records exactly the successful copies, independent of UI state |
| Action menu, details pane, chips, library tabs, toasts, keyboard routing | Svelte frontend | pure `lib/*.ts` | Presentation and interaction; pure logic node-tested |
| Library title filter, section grouping, cross-section navigation | Frontend `lib/library.ts` | — | Local, per keystroke, no request |

## Standard Stack

### Core (all already in the repo; nothing to install)

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib `image/gif` | Go 1.27.2 (go.mod `go 1.25.0`) | Square crop: `DecodeAll`/`EncodeAll`, `*image.Paletted.SubImage` | Lossless geometry crop, verified today `[VERIFIED: scratchpad probe, see Code Example 1]` |
| Go stdlib `os` (`CreateTemp`, `Link`, `Rename`) | Go 1.27.2 | Atomic library writes; no-overwrite download claim | `os.Link` fails with `fs.ErrExist` and never overwrites `[VERIFIED: scratchpad probe on APFS]` |
| Wails v3 | `v3.0.0-beta.30` | Bindings, events, `ServiceShutdown` lifecycle | Pinned project stack `[VERIFIED: go.mod, wails3 version]` |
| `@tanstack/svelte-virtual` | `3.13.47` | Library grid with header rows + tile rows (variable `estimateSize`) | Already drives `ResultGrid.svelte` `[VERIFIED: frontend/package.json]` |
| Svelte 5 | `^5.46.4` | New components | Existing |
| node:test (Node v24.14.1) | built-in | Pure TS logic tests (`src/lib/*.test.ts`) | Existing `pnpm test` script `[VERIFIED: frontend/package.json "test": "node --test --test-reporter=tap src/lib/*.test.ts"]` |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `zalando/go-keyring` | v0.2.8 | Unchanged (keys) | Not touched this phase |
| dev-browser CLI | at `/opt/homebrew/bin/dev-browser` | Headless UAT scripts | Every UI plan's verification |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Own cgo `CopyText` | Wails `app.Clipboard.SetText` | Wails' version always writes the general pasteboard (no private-pasteboard test hook), and its server build is not the `Fake`. Own cgo keeps one clipboard seam and is testable with a named pasteboard |
| `os.Link` claim | `O_CREATE|O_EXCL` claim then stream (what the extension does) | Simpler, but a partially written file is visible in Downloads during the transfer and on crash. `os.Link` publishes a complete, validated file atomically. Both never overwrite |
| `image/gif` crop | cgo ImageIO (`CGImageDestination`) | Re-quantizes and flattens disposal (the extension's Swift script). Rejected in milestone research |
| JSON files | SQLite / webview `localStorage` | Tiny data; `localStorage` on a custom-scheme WKWebView origin is fragile and the project keeps non-secret prefs in Go JSON |
| Go-side Undo slot | Frontend keeps the removed list and writes it back | Would send whole snapshot lists from JS to disk; Go slot keeps writes server-side and validated |

**Installation:** none. Phase 3 adds no Go module and no npm package.

## Package Legitimacy Audit

No external packages are installed in this phase (Go stdlib and existing pinned dependencies only). The legitimacy gate does not apply.

| Package | Registry | Age | Downloads | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| (none) | — | — | — | — | — | — |

**Packages removed due to [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Provider Field Map (the core data question)

Current model `[VERIFIED: internal/provider/provider.go:33-40]`:
```go
type Gif struct {
	Provider    ID     `json:"provider"`
	ID          string `json:"id"` // GIPHY id, Klipy slug, Finer fileid
	Title       string `json:"title"`
	PreviewURL  string `json:"previewUrl"`  // grid only
	DownloadURL string `json:"downloadUrl"` // copy only
	IsClip      bool   `json:"isClip"`
}
```
Provider IDs `[VERIFIED: internal/provider/provider.go:14-19]`: `GIPHY ID = "giphy"`, `Clips ID = "clips"`, `Klipy ID = "klipy"`, `FinerGifs ID = "finergifs"`.

**Recommended additions (all `omitempty`, so missing fields stay absent in JSON and optional in generated TS):**

```go
	Width     int      `json:"width,omitempty"`     // of the file Copy puts on the clipboard
	Height    int      `json:"height,omitempty"`
	Size      int64    `json:"size,omitempty"`      // bytes of that file; 0 = unknown
	Tags      []string `json:"tags,omitempty"`
	PageURL   string   `json:"pageUrl,omitempty"`   // https vendor page; "" = no page actions
	SourceURL string   `json:"sourceUrl,omitempty"` // https only
	MediumURL string   `json:"mediumUrl,omitempty"` // Details preview; "" = use PreviewURL
```

| Field | GIPHY (`messaging_non_clips`) | Clips (`clips_grid_picker`) | Klipy native v1 | Finer Gifs |
|-------|------|-------|-------|------------|
| DownloadURL | `images.original.url` (existing) | chosen `video.assets[r].url` (new) | `file.hd.gif.url` else `md` (existing) | `<media>/<fileid>.gif` (existing) |
| Width/Height | `images.original.width/height` (strings → existing `flexInt`) | chosen asset `width/height` | chosen `file.{hd|md}.gif.width/height` (numbers) | none |
| Size | `images.original.size` (string) | **none** (assets carry no size) | chosen `file.{hd|md}.gif.size` | none |
| Tags | **none** (GIF Object has no `tags` field) | none | `tags` (array; element type undocumented) | none |
| PageURL | `url` (`https://giphy.com/gifs/<slug>`) | `url` ("GIPHY – Be Animated page for the clip") | **none** | **none** |
| SourceURL | `source` (any third-party site; keep only https) | none documented | none | none (the extension's TVDB link is optional, see Open Question 4) |
| MediumURL | `images.fixed_height.url` (200px tall; in the bundle) | none → PreviewURL (`fixed_width`) | `file.md.gif.url` | none → PreviewURL (already the full GIF) |

Evidence:
- `messaging_non_clips` includes exactly `original, fixed_height, fixed_height_downsampled, fixed_height_small, fixed_width, fixed_width_downsampled, fixed_width_small`; `clips_grid_picker` includes `fixed_width, original, video.assets.360p/720p/480p/1080p/source` `[CITED: developers.giphy.com/docs/optional-settings/]`. So `downsized*` is **not** available for the Details preview; use `fixed_height`.
- GIF Object fields: `type, id, slug, url, bitly_url, embed_url, username, source, rating, content_url, user, source_tld, source_post_url, *_datetime, images, title, alt_text`; **no `tags`**. Rendition `width`, `height`, `size` are strings `[CITED: developers.giphy.com/docs/api/schema/]`.
- Clips: "The 360p rendition is always present by default. All other renditions will only be present if they are available." Each rendition has `width, height, url`, no size `[CITED: developers.giphy.com/docs/clips/]`.
- Klipy item: `id, slug, title, file.{hd,md,sm,xs}.{gif,webp,jpg,mp4,webm}.{url,width,height,size}, tags, type, blur_preview`; **no page/share URL field**; `tags` is empty in the only example and its element type is not specified `[CITED: docs.klipy.com/gifs-api/gifs-search-api.md]`.
- Probing `https://klipy.com/gifs/<slug>` returned 403 for both a real and a nonsense slug (bot protection), so a constructed Klipy page URL cannot be validated. Do not construct one `[VERIFIED: curl probe 2026-10-17]`.
- Repo fixture `internal/provider/giphy/testdata/trending.json` already carries `"url": "https://giphy.com/gifs/cat-3o7aD2saalBwwftBIY"` and `original` `width/height/size` strings `[VERIFIED: fixture read this session]`.

**Mapper rules:**
- **Clips rendition choice:** first present of `720p`, `480p`, `1080p`, `360p` (nearest to 720 by height; tie → higher quality). Never `4k`/`source` (size unknown, 50 MB staging cap). A clip with none of these is a malformed item → `c.bad(status)` (the existing "never a silently shorter page" rule at giphy.go:144-146).
- **Clips decode:** add `Video struct{ Assets map[string]asset }` where `asset` has `URL string; Width, Height flexInt`. Use `flexInt` because GIPHY sends numerics as strings elsewhere and the fixture uses strings.
- **GIPHY `url` scheme:** the docs example shows `http://giphy.com/...`; the live-recorded fixture shows `https://`. Accept `https`; upgrade `http` to `https` only when the host is `giphy.com` or a subdomain; otherwise drop the field.
- **`source`:** keep only when it parses as `https` with a non-empty host. It can point anywhere.
- **Klipy `tags`:** decode as `[]json.RawMessage` (or `[]any`) and keep string elements only. Decoding into `[]string` would fail the whole page if Klipy ever sends tag objects (Pitfall 7). Cap 50 tags × 100 runes.
- **Klipy dimensions/size** come from the same size key (`hd`, else `md`) as `DownloadURL`. Add `Width/Height/Size` to the existing `size.GIF` struct.
- Generated TS: regenerate bindings with `wails3 generate bindings -f '{{.BUILD_FLAGS}}' -clean=true -ts -i` `[VERIFIED: build/Taskfile.yml:191]` (or `wails3 task common:generate:bindings`). The current output is interface-style (`export interface Gif {...}`), and `omitempty` fields become optional properties `[ASSUMED: generator behavior for omitempty]`.

## Raycast Extension Semantics to Port (not code)

Read from `raycast/extensions/extensions/gif-search` @ `4f1b962` (MIT) this session `[VERIFIED: sparse clone]`.

| Action | Extension behavior | Gifsmith port |
|--------|-------------------|---------------|
| Copy/Paste GIF Link | `stripQParams(gif_url)` (drops the whole query) | GIPHY: strip the query (same). **Klipy: do not alter** (Klipy rule: "do not alter URL params"). Finer: no query anyway |
| Copy/Paste Markdown | `` `![${item.title}](${stripQParams(gif_url)})` `` | `![{title}]({link})`, alt `GIF` when the title is empty (UI-SPEC); escape `[`, `]`, `\` in the title and collapse newlines (the extension does not; one-line fix) |
| Clips link | `gif_url` = `images.original.url` (the **GIF**), even for Clips | UI-SPEC says the **MP4 URL** for Clips. Conflict → Open Question 1 |
| Copy Page Link / Open in Browser | Only when `url` is set | Same: omitted for Klipy and Finer Gifs |
| Download naming | `<slug>.<ext>`; collision `<stem> (1).<ext>` via exclusive create; hide filename → `gif.<ext>` | CONTEXT locks `<title-slug>-<id>.<ext>`, suffix ` (2)`, ` (3)`; hide → `gifsmith-<id>` (Phase 4) |
| Download validation | None (only HTTP status) | Magic bytes required (ACTN-11) |
| Square crop | Downloads and runs a Swift ImageIO script (re-encodes); not offered off macOS | Pure Go geometry crop; omitted for Clips |
| Clips download | `1080p ?? 720p ?? 360p ?? gif_url`, `.mp4` name | 720p first (CONTEXT), then nearest |
| Usage tracking (Recents) | `trackUsage` on copy/paste GIF, square, link, Markdown, **page link, open in browser, view details** | CONTEXT narrows it to copy/paste of GIF, link, Markdown, square. Do **not** record page link, open, download or details |
| Favorites/Recents storage | `LocalStorage["<service>-favs"|"<service>-recent"]` = **IDs only**, then re-fetched by API `gifs(ids)` | Snapshot of the full `Gif` (LIBR-04: no API calls). Recents prepend; favorites append (extension) → Gifsmith shows newest first per section (CONTEXT) |
| Recents cap | None in storage (display slices to 2 rows) | 50 (CONTEXT) |
| Recent searches | **Not in the extension** (no search history anywhere in `src/`) | New for Gifsmith (SRCH-08) |
| Details | Title, Width px, Height px, Size (SI, `formatFileSize`), labels (Created, User), links (Source with `source_tld` text), tags, provider | UI-SPEC field set; size format `{n.n} MB` / `{n} KB` |

## Architecture Patterns

### System Architecture Diagram

```
 keydown / click (webview)                                     Go (bound services, own goroutines)
 ───────────────────────────                                   ─────────────────────────────────────
 search input ─► routeKey() ─┬─ grid nav / tabs / chips          
                             ├─ ⌘K ─► ActionMenu ─► run(action)─┐
 ⌘⇧C ⌘⇧M ⌘O ⌘S ⌘⇧F ⌘I ⌘⌫ ⌘Z ─┘                                  │
                                                                 ▼
   run(action, gif) ──────────────► ActionService ── busy.TryLock ──┐
     copy/paste file ─────────────►   Copy/Paste ──► staging.Fetch(kind gif|mp4) ─► magic check ─► <staging>/<h>/<slug>.<ext>
     copy/paste square ───────────►   CopySquare ──► staging.Fetch(≤15 MB) ─► imaging.Square ─► <staging>/<h'>/<slug>.gif
     copy/paste link|md|page ─────►   CopyText ───► buildText(g, kind)
                                          │
                                          ├─► Platform.CopyFile / CopyText ─► NSPasteboard (general)
                                          ├─► staging.Sweep(keep) | Clear()        (only the clipboard's file survives)
                                          ├─► Recents.Touch(g)  ──────────────────► library.Store ─► recents.json
                                          ├─► [paste] Hide ─► 80 ms ─► frontmost check ─► Platform.PasteShortcut (⌘V)
                                          └─► [Klipy + ⌘V posted] time.AfterFunc(grace) ─► delete that staged dir
     download ────────────────────►   Download ──► staging.Save(dir) ─► temp in dir ─► magic ─► os.Link claim ─► Downloads/<name>
     Show in Finder (⌘R) ─────────►   ShowDownload ─► Platform.RevealFile(lastDownload)
     open page / source ──────────►   OpenPage/OpenSource ─► https+host check ─► Platform.OpenURL
   favorites / recents / searches ─► LibraryService ─► library.Store (mutex, in-memory + atomic JSON, one Undo slot)

 Lifecycle: main() ─► stage.Clear() ─► app.Run()  …  Quit/SIGTERM ─► Wails cleanup ─► ActionService.ServiceShutdown ─► stage.Clear()
 Errors: fixed-text Go errors → MarshalError → err.cause.kind → UI copy rows F1–F8 (text never rendered)
```

### Recommended Project Structure (additions only)

```
internal/
├── provider/            # Gif gains 7 optional fields; giphy/klipy mappers fill them
│   └── fakeapi/         # + page/source/tags/dims fixtures, 720p MP4 assets, /media/*.mp4, wide GIFs
├── platform/            # + CopyText, RevealFile, OpenURL (darwin cgo, stub, Fake)
├── staging/             # Fetch gains a kind (gif|mp4) + size cap param; + Save (downloads), Clear, Remove(dir)
├── imaging/             # NEW: square.go (pure image/gif), square_test.go (golden: frames, delays, loop, pixels)
├── library/             # NEW: store.go (favorites, recents, searches, undo), store_test.go
└── services/            # actions.go grows; library.go NEW (bound LibraryService); errors.go (ActionError kinds)
frontend/src/
├── lib/
│   ├── actions.ts(+test)   # action list for (gif, context), filter, shortcut table
│   ├── library.ts(+test)   # sections, virtual rows, flat index ↔ (section, col), cross-section move
│   ├── format.ts(+test)    # fileSize, linkText, tags summary
│   ├── popover.ts(+test)   # placement clamp inside 720×480
│   └── grid.ts(+test)      # routeKey: new KeyActions (shift-aware)
└── components/
    ├── ActionMenu.svelte  DetailsPane.svelte  RecentSearches.svelte
    ├── LibraryGrid.svelte (or ResultGrid with header rows)
    └── Footer.svelte (toast action + hint), Tile.svelte (MP4 + favorite badges)
scripts/uat/            # 07-*.js … new UAT scripts (action menu, library, details, chips, download)
```

### Pattern 1: One clipboard seam with three new primitives

Current interface `[VERIFIED: internal/platform/platform.go:21-39]` has `CopyFile`, `FocusWebView`, `CursorScreenWorkArea`, `FrontmostApp`, `ActivateApp`, `AccessibilityTrusted`, `OpenAccessibilitySettings`, `PasteShortcut`. Add:

```go
	// CopyText puts a plain string on the general pasteboard (public.utf8-plain-text).
	CopyText(text string) error
	// RevealFile selects path in a Finder window (activates Finder).
	RevealFile(path string) error
	// OpenURL opens an https URL in the default browser.
	OpenURL(u string) error
```

- darwin: `gsCopyText(pbName, text)` on the caller's thread (same rule as `gsCopyFile`: "Pasteboard calls run on the caller's thread"), plus a test helper `pasteboardString(name)` reading it back from a private named pasteboard. `gsReveal` and `gsOpenURL` use `onMain` like `gsOpenAXSettings` and are **not** unit-tested (no main run loop under `go test`).
- stub: `ErrUnsupported`. Fake: records `CopyText:<text>`, `RevealFile:<path>`, `OpenURL:<url>`.

### Pattern 2: Flat bound methods on ActionService (no action framework)

```go
func (a *ActionService) Copy(ctx, g) (Result, error)          // existing: GIF or (now) MP4 file
func (a *ActionService) Paste(ctx, g) (Result, error)         // existing
func (a *ActionService) CopySquare(ctx, g) (Result, error)
func (a *ActionService) PasteSquare(ctx, g) (Result, error)
func (a *ActionService) CopyText(ctx, g, kind string) (Result, error)  // kind: link | markdown | page
func (a *ActionService) PasteText(ctx, g, kind string) (Result, error) // link | markdown
func (a *ActionService) Download(ctx, g) (DownloadResult, error)       // {name string} display only, never the path
func (a *ActionService) ShowDownload(ctx) error                        // reveals the last download Go remembers
func (a *ActionService) OpenPage(ctx, g) error
func (a *ActionService) OpenSource(ctx, g) error
func (a *ActionService) ServiceShutdown() error                        // Wails lifecycle hook: stage.Clear()
```
- Refactor `Paste` into `paste(ctx, put func() error, okMsg string)` so file, square and text pastes share the Hide → `PasteDelay` → frontmost check → ⌘V sequence (actions.go:94-124). `PasteDelay = 80 * time.Millisecond` stays `[VERIFIED: internal/services/actions.go:24]`.
- Every method takes `busy.TryLock` (one pasteboard writer, no double ⌘V; Download shares it so the busy overlay rule "repeat triggers are ignored while one runs" holds).
- Delete `ErrClipsCopy` and its `IsClip` guard: it was the Phase 2 placeholder "copying Clips arrives in a later update" `[VERIFIED: internal/services/actions.go:20-21]`. Also retire `CLIPS_COPY` in App.svelte `[VERIFIED: frontend/src/App.svelte:37]`.
- Fix T5a copy: the Go string at actions.go:88 says `Enable automatic paste…`; UI-SPEC T5a requires `Enable Automatic Paste…`.
- `ShowDownload` takes no path from JS: Go stores the last successful download path in the service (consistent with "They take no paths from JS", actions.go:1-2).
- Recents: `ActionService.Recents interface{ Touch(provider.Gif) error }`, called after every result with `Copied == true` from Copy/Paste/CopySquare/PasteSquare/CopyText(link|markdown)/PasteText. Not for `page`. A Touch failure is logged and does not fail the copy (the copy succeeded; the extension's `trackUsage` makes the same choice).
- Typed errors: add `type ActionError struct{ Kind string }` (kinds: `busy`, `download_failed`, `not_media`, `too_large`, `crop_failed`, `crop_too_large`, `folder_unwritable`, `open_failed`, `no_page`) and compose `MarshalError` in main.go (`provider.MarshalError` first, then the action one). The frontend maps `cause.kind` to F1–F8; it already compares `BUSY` by message text `[VERIFIED: frontend/src/App.svelte:39]` and can move to `cause.kind === 'busy'`.

### Pattern 3: Clipboard text built in Go

```go
func linkFor(g provider.Gif) string // GIPHY/Clips: DownloadURL (or GIF URL, per OQ1) with RawQuery="" ; Klipy/Finer: unchanged
func markdownFor(g provider.Gif) string {
	alt := strings.Join(strings.Fields(g.Title), " ") // newlines/tabs → single spaces
	if alt == "" { alt = "GIF" }
	alt = strings.NewReplacer(`\`, `\\`, `[`, `\[`, `]`, `\]`).Replace(alt)
	return "![" + alt + "](" + linkFor(g) + ")"
}
```
Validate before writing: the link URL must pass the injected staging allowlist (same `Allow` func as downloads); the page URL must be https on a vendor host.

### Pattern 4: Staging lifecycle

- `staging.Store` today `[VERIFIED: internal/staging/staging.go:38-43]`: `Dir`, `HTTP`, `Allow` (nil = `VendorAllow`), `maxBytes` (0 = 50 MB). Errors `[VERIFIED: staging.go:25-30]`: `ErrNotAllowed`, `ErrDownload`, `ErrNotGIF`, `ErrTooLarge`.
- Generalize `Fetch` to take the expected kind and a byte cap: GIF → `GIF87a`/`GIF89a` (existing check at staging.go:85-88); MP4 → bytes 4..8 == `ftyp`. Rename `ErrNotGIF` usage to a kind-neutral "not media" error and keep the old text for GIFs, or add `ErrNotMP4`. Extension `.mp4` for Clips.
- Square output goes in its own hash dir (`sha1(provider:id:square)`) so `Sweep(keep)` removes the downloaded original after the crop.
- `Clear()` removes every entry under `Dir` (used at launch, at quit, after a text copy).
- **Launch:** call `stage.Clear()` in `main()` before `app.Run()` (simple, no Wails dependency). **Quit:** implement `ServiceShutdown() error` on `ActionService`; Wails calls it from `cleanup()` on `Quit`, ⌘Q via `applicationShouldTerminate`, and SIGINT/SIGTERM (default signal handler calls `a.Quit`) `[VERIFIED: wails v3.0.0-beta.30 application.go:920-967, application_darwin_delegate.m:60-68, signal_handler_desktop.go]`. `Options.OnShutdown` is equivalent; `ShouldQuit` is for vetoing, not cleanup `[VERIFIED: application_options.go:78-94]`. Shutdown runs on the main thread inside `InvokeSync`; keep it to `os.RemoveAll` (no `onMain`, no busy lock).
- **Klipy:** after `Paste`/`PasteSquare` of a Klipy item with ⌘V posted, `time.AfterFunc(KlipyPasteGrace, …)` removes that staged dir, guarded by a generation counter so it never deletes a newer copy of the same GIF. `Copy` (no ⌘V) keeps the file until the next clipboard write or quit. Inject the timer (`After func(time.Duration, func())`) for tests, like `Sleep`.

### Pattern 5: Download with atomic no-overwrite claim

```go
// Save downloads g (validated) into dir as <stem>.<ext>, " (2)", " (3)"…; never overwrites.
func (s *Store) Save(ctx context.Context, g provider.Gif, dir, stem, ext string) (string, error)
```
1. Fetch via the shared allowlisted, redirect-checked transport; read the magic bytes **before** creating anything.
2. `f := os.CreateTemp(dir, ".gifsmith-*.part")`, stream the body (cap 50 MB), `f.Sync()`, `f.Chmod(0o644)` (CreateTemp files are 0600), close.
3. For n := 1..999: `os.Link(tmp, candidate)`; `errors.Is(err, fs.ErrExist)` → next suffix; other error → `folder_unwritable`.
4. `os.Remove(tmp)` always (deferred).
Name stem: `slug(title)` (existing, ≤80 bytes, "gif" when empty) + `-` + sanitized id (`[A-Za-z0-9_-]` only, ≤64). With `hide` (Phase 4): `gifsmith-<id>`.
Downloads dir: `filepath.Join(os.UserHomeDir(), "Downloads")` on desktop; `bin/uat-downloads` in the server build. Inject as `func() string` on `ActionService` so Phase 4's folder picker plugs in.

### Pattern 6: Library store

- Files in `ConfigDir` (`~/Library/Application Support/com.kenscott.gifsmith`, wire_desktop.go:32): `favorites.json`, `recents.json`, `searches.json`, each `{"version":1,…}`, mode 0600, dir 0700. Generalize `settings.SavePrefs`' temp+rename `[VERIFIED: internal/settings/settings.go:142-170]` into `WriteJSONAtomic(dir, name string, v any) error` and add `f.Sync()` before `Rename`.
- Entry: `type Entry struct { Gif provider.Gif `json:"gif"`; At int64 `json:"at"` }` (explicit field, not embedded, so the binding generator and JSON shape are obvious). Key = `provider + ":" + id`.
- Ordering: newest first in the file. Recents: `Touch` removes the existing key and prepends; truncate at 50. Favorites: prepend on add; no cap needed (consider a sanity cap of 1000). Searches: map provider → `[]string`, dedupe case-insensitively (keep the latest casing), prepend, cap 20, only for the four provider IDs, trimmed, ≤100 runes, ≥2 code points.
- Load once (lazily), keep in memory under a mutex, write-through on each mutation. If a file fails to parse or has `version > 1`: report a `library_read` error, refuse writes to that file until a successful reload (Try Again), and never overwrite it (Pitfall 11).
- Undo: one slot `{list, entries, indices}` replaced by each removal (`RemoveFavorite`, `RemoveRecent`, `ClearRecents`). `Undo()` re-inserts at the recorded indices (clamped), skipping keys already present, then re-applies caps. Reselection is the frontend's job (it knows the restored key).
- Validation of a `Gif` coming from JS before persisting (and before any action): known provider ID; non-empty ID ≤200 bytes; title ≤500 runes; `PreviewURL`/`DownloadURL`/`MediumURL` pass the injected staging `Allow`; `PageURL` https on the vendor host; `SourceURL` https; tags capped. Use the same injected `Allow` as staging, because the server build's fixture URLs are `http://127.0.0.1:<port>` `[VERIFIED: wire_server.go:42-45]`.
- Bound `LibraryService`: `Load(ctx) (Snapshot, error)`, `SetFavorite(ctx, g, on bool) (Snapshot, error)`, `RemoveRecent(ctx, p, id)`, `ClearRecents(ctx)`, `Undo(ctx)`, `SaveSearch(ctx, p, q) error`, `RemoveSearch(ctx, p, q) error`. Return the whole snapshot (two lists ≤ a few hundred entries) so the frontend never computes membership from stale state. Reload on `panel:show` and on switching to a library tab (Recents change while the panel is hidden after a paste). No new event is needed.

### Pattern 7: Keyboard routing additions (`lib/grid.ts`)

Current `KeyAction` `[VERIFIED: frontend/src/lib/grid.ts:5]`: `Direction | 'copy' | 'paste' | 'escape' | 'actions' | 'tab' | 'slot1' | 'slot2' | 'slot3' | 'slot4' | 'none'`. `routeKey` receives `{ key, metaKey, isComposing, keyCode, repeat }` (grid.ts:32) and has no `shiftKey`.

Add `shiftKey` to the event type and new actions: `slot5`, `slot6`, `copyLink` (⌘⇧C), `copyMarkdown` (⌘⇧M), `open` (⌘O), `download` (⌘S), `favorite` (⌘⇧F), `details` (⌘I), `remove` (⌘⌫ = `metaKey && key === 'Backspace'`), `undo` (⌘Z, not shift), `reveal` (⌘R). Compare letters with `e.key.toLowerCase()` **and** check `shiftKey` explicitly, so ⌘C/⌘⇧C and ⌘Z/⌘⇧Z never collide regardless of whether WebKit reports `"c"` or `"C"` under ⌘⇧ (Pitfall 10).

The handler must stop calling `preventDefault()` unconditionally for every routed key (App.svelte:345-347 does today): `remove` and `undo` are prevented **only** when they act (library tab selection or a selected chip; a visible Undo toast). Otherwise native ⌘⌫ (delete to line start) and ⌘Z (Edit menu undo) must reach the input.

Mode precedence for one `<main>`-level dispatcher: action menu open → Details open → chip selected → grid. The menu's filter input owns ↑ ↓ ↵ Esc and typing; every listed shortcut still works while it is open (UI-SPEC).

### Pattern 8: Library grid = virtual rows of two heights

`lib/library.ts`: `sections(entries, filter)` → ordered `[{provider, items}]` in slot order GIPHY, Clips, Klipy, Finer Gifs, empty sections dropped; `rows(sections)` → `[{kind:'header', provider, count} | {kind:'tiles', section, start}]` with heights 32 and 152; `flatIndex ↔ (section, offset)`; `moveSectioned(index, dir, sizes)` implementing the UI-SPEC cross-section rule (↓ from a section's last row → same column of the next section's first row, clamped). The existing `ResultGrid` already passes a per-index `estimateSize` to `createVirtualizer` (ResultGrid.svelte:46), so header rows fit the same API.

### Anti-Patterns to Avoid
- **JS-built URLs or clipboard strings:** the webview must never compose link/Markdown/page text or open arbitrary URLs. Go builds and validates them.
- **`Browser.OpenURL` from JS for page links:** the existing `openExternal` allow-list holds three fixed URLs (api.ts:21); page URLs are per-item. Opening through JS in the server build would also launch a real browser on the user's Mac during headless UAT (Pitfall 9).
- **Deleting a Klipy file the instant ⌘V is posted** (Pitfall 4).
- **Silently resetting a corrupt library file** (Pitfall 11).
- **Listing `~/Downloads`** to find a free name: unnecessary (the link claim is per path) and widens TCC exposure.
- **Re-quantizing or compositing frames** for the square crop.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Atomic, never-overwrite publish of a download | exists-check then write | `os.Link(tmp, candidate)` loop | Atomic, `EEXIST` on collision, no TOCTOU, never a partial visible file |
| Atomic JSON write | open/truncate/write | temp in same dir + `Sync` + `Rename` (existing SavePrefs pattern) | Crash leaves the old file whole |
| GIF crop | per-frame RGBA compositing + quantizer | `(*image.Paletted).SubImage` + `Rect.Sub(win.Min)` + `gif.EncodeAll` | Lossless; disposal/palette/delay preserved; verified |
| MP4 detection | full MP4 parser | `ftyp` at bytes 4..8 | ISO BMFF requires the file-type box first `[ASSUMED: GIPHY MP4s begin with ftyp]` |
| Virtualized sectioned grid | custom scroll math | `@tanstack/svelte-virtual` with variable `estimateSize` | Already in use |
| Popover a11y | custom focus trap library | `role=combobox` input + `aria-activedescendant` (UI-SPEC) | Same pattern as the search input |
| Redirect-safe download | new HTTP client | staging's existing `CheckRedirect` allowlist (staging.go:67-73) | Already reviewed (Phase 2 fix) |

**Key insight:** the dangerous parts here (file naming, URL trust, pasteboard timing) each already have one reviewed implementation in the repo. Extend those; do not add a second path.

## Common Pitfalls

### Pitfall 1: Square crop frame bounds
**What goes wrong:** `gif.EncodeAll` fails with `gif: image block is out of bounds` `[VERIFIED: $GOROOT/src/image/gif/writer.go:245-247]` if a cropped frame keeps its original coordinates.
**How to avoid:** `p := fr.SubImage(r).(*image.Paletted); p.Rect = p.Rect.Sub(win.Min)` and set `Config.Width/Height = side`, keeping `Config.ColorModel` (global palette), `BackgroundIndex`, `LoopCount`, every `Delay` and `Disposal`. A frame whose rectangle misses the window becomes a 1×1 frame with palette `{color.RGBA{}}`; the encoder marks the first alpha-0 entry transparent `[VERIFIED: writer.go:250-258]`. Do not fold the delay into the previous frame (that changes when the previous frame's disposal runs).
**Warning signs:** encode error; frame count or delays differ; flicker at the edges.

### Pitfall 2: Crop memory and the 15 MB rule
**What goes wrong:** `gif.DecodeAll` holds every frame in memory (1 byte/pixel per frame rect). The probe peaked around 250 MB while also rendering RGBA reference frames for 7 GIFs; production crop is lower but a highly compressible 15 MB GIF can expand far more.
**How to avoid:** refuse before work when `g.Size > 15 MB` (UI-SPEC "refused before any work"); otherwise fetch with a 15 MB byte cap (`ErrTooLarge` → `crop_too_large`, F7). Call `gif.DecodeConfig` first and refuse canvases over 4096×4096 (`crop_failed`). Accept the residual decompression risk from vendor CDNs.

### Pitfall 3: TCC "Downloads folder" consent
**What goes wrong:** macOS gates `~/Downloads` for non-sandboxed apps; the first write may raise a consent prompt. While it is up, the write blocks, and the alert can take focus so `HideOnFocusLost` hides the panel (shell.go:43). A denial returns EPERM.
**How to avoid:** add `NSDownloadsFolderUsageDescription` to both `build/darwin/Info.plist` and `Info.dev.plist` (Apple: optional but highly recommended, macOS 10.15+) `[CITED: developer.apple.com/documentation/bundleresources/information-property-list/nsdownloadsfolderusagedescription]`. Map EPERM/EACCES to `folder_unwritable` (F4). Keep the download toast visible across a hide (set the keep-status flag when the result lands while hidden). Grants are per bundle id: the dev build (`com.kenscott.gifsmith.dev`) and the release build prompt separately. Whether *creating* a file prompts at all is not settled: Apple says an app can access a file it created in Downloads without consent → user-verified checkpoint `[ASSUMED]`.

### Pitfall 4: Deleting Klipy files before the target app reads them
**What goes wrong:** `PasteShortcut` returns as soon as the ⌘V events are posted. The target app reads the pasteboard and the file afterwards; Slack/Electron may read it again on send. Deleting immediately can produce a broken or empty paste.
**How to avoid:** delete after a grace period (recommend `KlipyPasteGrace = 30 * time.Second`), on the next clipboard write, or at quit, whichever comes first. User-verify Klipy paste into Messages and Slack, including sending after a pause. After deletion, the clipboard still holds a file URL to a missing file; a later manual ⌘V pastes nothing. That is the accepted cost of the Klipy rule.

### Pitfall 5: Quitting clears the clipboard's file too
**What goes wrong:** ACTN-10 wipes staging at quit, so a GIF copied just before quitting can no longer be pasted.
**How to avoid:** nothing to fix; it follows from the locked decision. Note it in the user-facing verification so it is not reported as a bug.

### Pitfall 6: Klipy forbids altering URLs
**What goes wrong:** porting the extension's `stripQParams` to every provider alters Klipy URLs. Klipy's rules: "load media directly from returned URLs, do not alter URL params" (STACK.md item 10, from docs.klipy.com integration-requirements).
**How to avoid:** strip the query only for GIPHY/Clips links.

### Pitfall 7: Strict decoding of loosely documented fields
**What goes wrong:** a new `Tags []string` field fails the whole Klipy page if an element is an object; GIPHY sends numerics as strings (`width`, `size`) and the existing `flexInt` exists for that reason (giphy.go:28-43).
**How to avoid:** tolerant decoding for every *display-only* field (`[]json.RawMessage` for tags; `flexInt` for GIPHY numbers); keep the existing strict rule only for fields an action needs (preview/download URLs).

### Pitfall 8: Unsafe names from provider ids and titles
**What goes wrong:** Klipy's id is its slug (unvalidated text from the API); Finer's regex `^[A-Za-z0-9._-]+$` admits `..`. A download name built from them could contain `/` or be `..`.
**How to avoid:** the stem function whitelists `[A-Za-z0-9_-]` for the id part and reuses `slug()` for the title; assert in a test that the joined path's `filepath.Dir` equals the destination dir.

### Pitfall 9: Opening URLs in the server (UAT) build
**What goes wrong:** Wails' `Browser.OpenURL` (JS runtime or `app.Browser.OpenURL` in Go) execs the system opener in every build: `exec.Command("open", target)` on darwin, with no server-build exclusion `[VERIFIED: wails v3.0.0-beta.30 internal/browser/browser_darwin.go:7-8]`. A headless UAT that triggers "Open in Browser" would open a real browser window on the user's primary Mac (forbidden).
**How to avoid:** route `OpenPage`/`OpenSource`/`ShowDownload` through `platform.Platform`, and give the server build a non-native platform (see Pattern 10 / Validation Architecture).

### Pitfall 10: Key equivalents in WKWebView vs headless Chromium
**What goes wrong:** UAT runs in Chromium; the app runs WKWebView with a native Edit menu (⌘Z, ⌘C, ⌘V, ⌘X, ⌘A) and the App menu (⌘Q, ⌘H) (shell.go:61-64). Whether `preventDefault` on a DOM keydown suppresses the Edit menu's ⌘Z only shows up in the real app. Shifted letters under ⌘ may arrive as `"C"` or `"c"`.
**How to avoid:** match `toLowerCase()` + `shiftKey`; capture ⌘Z only while the Undo toast is visible; user-verify ⌘Z undo, ⌘⌫ remove and ⌘R reveal in the signed dev build. ⌘R is safe to capture: there is no View menu (Wails' default menu with ⌘R reload was replaced, shell.go:59-64) `[VERIFIED: shell.go]`. Native WebKit/AppKit dispatch order is `[ASSUMED]`.

### Pitfall 11: Corrupt library files
**What goes wrong:** a parse error treated as "empty" and then a write wipes the user's favorites.
**How to avoid:** surface `library_read` → UI error state `Couldn't load {Favorites|Recents}` with Try Again; refuse writes for that file until a clean load; never auto-reset.

### Pitfall 12: The hide-after-copy timer (02 IN-07)
**What goes wrong:** `setTimeout(hide, HIDE_AFTER_COPY_MS)` (App.svelte:397) is never cancelled; re-showing within 400 ms hides the panel again.
**How to avoid:** keep the handle; `clearTimeout` on `panel:show`.

### Pitfall 13: `CreateTemp` permissions leak into downloads
**What goes wrong:** `os.CreateTemp` creates mode 0600; a hard link shares that inode, so the downloaded file is 0600.
**How to avoid:** `Chmod(0o644)` before linking (matches browsers' downloads).

### Pitfall 14: Stale UAT expectations once the server build can copy
**What goes wrong:** `scripts/smoke-browser-loop.sh` expects `ActionService.Copy` to return HTTP 422 `not supported in this build`, and `scripts/uat/01-tracer.js` expects E7 after Enter. Both break if the server build gets a working fake platform.
**How to avoid:** update both in the same plan; drive E7/F1 with a fixture media name that returns 500 (new fakeapi scenario) instead of the stub.

## Code Examples

### 1. Square crop (verified)

```go
// Source: probe in session scratchpad (cropprobe/main.go), run 2026-10-17 on Go 1.27.2:
//   synthetic-120x60 (all disposal kinds, one frame fully outside): 0 mismatches
//   6 real GIFs (16–182 frames, 327x270 … 499x300): frames, delays, loop equal; 0 pixel mismatches
func Square(g *gif.GIF) *gif.GIF {
	w, h := g.Config.Width, g.Config.Height
	s := min(w, h)
	win := image.Rect((w-s)/2, (h-s)/2, (w-s)/2+s, (h-s)/2+s)
	out := &gif.GIF{LoopCount: g.LoopCount, BackgroundIndex: g.BackgroundIndex,
		Config: image.Config{ColorModel: g.Config.ColorModel, Width: s, Height: s}}
	for i, fr := range g.Image {
		var p *image.Paletted
		if r := fr.Rect.Intersect(win); r.Empty() {
			// Fully outside: a 1x1 transparent frame keeps the delay and the disposal sequence.
			p = image.NewPaletted(image.Rect(0, 0, 1, 1), color.Palette{color.RGBA{}})
		} else {
			p = fr.SubImage(r).(*image.Paletted)
			p.Rect = p.Rect.Sub(win.Min)
		}
		out.Image = append(out.Image, p)
		out.Delay = append(out.Delay, g.Delay[i])
		out.Disposal = append(out.Disposal, g.Disposal[i]) // DecodeAll always fills Disposal
	}
	return out
}
```
Golden test: decode → `Square` → `EncodeAll` → `DecodeAll`, then compare frame count, delays, loop count and a disposal-aware RGBA render of every frame inside the window (the probe's `render` function is ~25 lines and belongs in `square_test.go`). Fixtures: generate small GIFs in the test (a wide one, a tall one, one with a frame fully outside, one already square); do not commit the user's Downloads GIFs.

### 2. MP4 / GIF magic check

```go
// head holds the first 12 bytes of the response body.
func validMedia(head []byte, mp4 bool) bool {
	if mp4 {
		return len(head) >= 8 && string(head[4:8]) == "ftyp"
	}
	return len(head) >= 6 && (string(head[:6]) == "GIF87a" || string(head[:6]) == "GIF89a")
}
```
Fixture MP4 for fakeapi (24 bytes, valid `ftyp` box): `00 00 00 18 'f' 't' 'y' 'p' 'i' 's' 'o' 'm' 00 00 02 00 'i' 's' 'o' 'm' 'm' 'p' '4' '1'`.

### 3. No-overwrite claim (verified)

```go
// Probe 2026-10-17 (APFS): Link to an existing name → "file exists", errors.Is(err, fs.ErrExist) == true;
// next candidate linked; the existing file's bytes were untouched.
for n := 1; n < 1000; n++ {
	name := stem + ext
	if n > 1 {
		name = fmt.Sprintf("%s (%d)%s", stem, n, ext)
	}
	err := os.Link(tmp, filepath.Join(dir, name))
	if err == nil {
		return filepath.Join(dir, name), nil
	}
	if !errors.Is(err, fs.ErrExist) {
		return "", errFolder // fixed text: *LinkError's message contains both paths
	}
}
```

### 4. cgo text pasteboard (pattern from gsCopyFile, platform_darwin.m:31-39)

```objc
int gsCopyText(const char *pbName, const char *text) {
    BOOL ok = NO;
    @autoreleasepool {
        NSPasteboard *pb = pasteboard(pbName);
        [pb clearContents];
        ok = [pb setString:[NSString stringWithUTF8String:text] forType:NSPasteboardTypeString];
    }
    return ok ? 1 : 0;
}

int gsReveal(const char *path) {          // onMain: AppKit/NSWorkspace UI
    __block int ok = 0;
    NSString *p = [NSString stringWithUTF8String:path];
    onMain(^{ @autoreleasepool {
        [[NSWorkspace sharedWorkspace] activateFileViewerSelectingURLs:@[[NSURL fileURLWithPath:p]]];
        ok = 1;
    }});
    return ok;
}
```
`stringWithUTF8String:` returns nil for invalid UTF-8; validate with `utf8.ValidString` in Go first. Copy the C string into an `NSString` before the block (the Go side frees it after the call returns, and `onMain` waits, so this is belt-and-braces).

### 5. Klipy grace delete

```go
if g.Provider == provider.Klipy && pasted {
	dir, gen := filepath.Dir(path), a.gen.Add(1)
	a.after(KlipyPasteGrace, func() {
		if a.gen.Load() == gen { // a newer clipboard write already swept or re-staged it
			_ = os.RemoveAll(dir)
		}
	})
}
```
Every clipboard write bumps `gen`, so a later copy of the same Klipy GIF (same hash dir) is never deleted by an older timer.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Extension stores favorite IDs and re-fetches by ID | Store full metadata snapshots | This project (LIBR-04) | Works offline and rate-limited; no `gifs?ids=` endpoint needed for Klipy/Clips |
| Extension square crop via Swift + ImageIO | Pure Go geometry crop | Milestone research, re-verified today | Lossless, no `/usr/bin/swift` dependency |
| Phase 2 `ErrClipsCopy` / `CLIPS_COPY` placeholders | MP4 copy/paste | This phase | Retire both strings and UI-SPEC E8 |

**Deprecated/outdated:**
- GIPHY `tags` on the GIF object: present in `@giphy/js-types`, absent from the documented API schema. Do not rely on it.
- Wails default menu ⌘R reload: already replaced in Phase 1.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Messages and Slack accept a pasted `.mp4` file URL as a video (same `CopyFile` path as GIFs) | ACTN-09 | Clips paste fails in those apps; would need a different payload. User-verified checkpoint |
| A2 | A 30 s grace is long enough for Messages/Slack to read a pasted Klipy file | Pitfall 4 | Broken Klipy pastes; raise the grace constant |
| A3 | Writing to `~/Downloads` from the signed dev build may raise a TCC prompt; `NSDownloadsFolderUsageDescription` sets its text | Pitfall 3 | First download blocks or fails; F4 copy shows; user grants once |
| A4 | In WKWebView, `preventDefault` on a DOM keydown for ⌘Z stops the Edit menu's Undo, and unprevented ⌘Z still undoes text | Pitfall 10 | ⌘Z double-acts or never reaches the input. User-verified |
| A5 | WebKit reports `e.key` as either case for ⌘⇧letter; matching `toLowerCase()` + `shiftKey` covers both | Pattern 7 | Shortcuts don't fire; trivial fix once observed |
| A6 | GIPHY MP4 assets begin with an `ftyp` box | Code Example 2 | Valid clips rejected as `not_media`; fallback: also accept `http.DetectContentType == "video/mp4"` |
| A7 | GIPHY Clip `url` is an https giphy.com page | Provider Field Map | Page actions hidden for Clips if the host check fails; harmless |
| A8 | Wails' binding generator emits optional TS properties for `omitempty` fields and `string[] \| null` for slices | Provider Field Map | Type-only friction in TS; adjust guards |
| A9 | Clips media hosts are under `giphy.com` (existing `VendorAllow` covers them) | Pattern 4 | MP4 fetch refused with `ErrNotAllowed`; extend the allowlist after seeing a real URL |
| A10 | Klipy `tags` elements are strings when present | Pitfall 7 | Mitigated by tolerant decoding either way |

## Open Questions (RESOLVED)

> All five resolved by the user on 2026-10-17 — see 03-CONTEXT.md "Post-research decisions" (plans index them as D-18 Clips link = GIF rendition, Klipy 30 s grace, chip only with ≥1 result, Finer Source omitted, server build uses platform.Fake).

1. **Clips Link/Markdown payload: MP4 URL (UI-SPEC) or the GIF rendition (extension, FEATURES.md)?**
   - What we know: UI-SPEC "Payload formats" says the MP4 URL for Clips. The extension uses `images.original.url` (a GIF) for link and Markdown, and `clips_grid_picker` includes `original`. `![t](x.mp4)` does not render as an image in most Markdown renderers.
   - Recommendation: keep the UI-SPEC default (MP4, no extra field) unless the user prefers parity; switching costs one optional `GifURL` field filled from `images.original.url`. Flag at plan review.
2. **Klipy paste grace length** (A2). Recommendation: 30 s constant, verified natively; adjust after the Slack check.
3. **Recent-search save rule details.** Save after 1.5 s idle even when the search returned 0 results or failed? Recommendation: save only after a first page arrives with ≥1 item (a typo with no results is not worth a chip), plus always on copy/paste. Planner can decide; it is UI-only.
4. **Finer Gifs "Source" link** (the extension shows a TheTVDB season link derived from the file id). Recommendation: omit in v1 (Go-constructed third-party URL, low value).
5. **Server-build platform.** Recommendation: the `-tags server` build uses `platform.Fake` (copy succeeds, Accessibility untrusted) so UAT can exercise success toasts, Recents recording and downloads; failures are driven by fixture scenarios. Requires updating the smoke script and 01-tracer.js (Pitfall 14).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | all Go work, `-race` tests | ✓ | go1.27.2 darwin/arm64 | — |
| wails3 CLI | bindings regeneration | ✓ | v3.0.0-beta.30 | — |
| Node | node:test, Vite | ✓ | v24.14.1 | — |
| pnpm | frontend scripts | ✓ | 12.14.0 | — |
| dev-browser | headless UAT | ✓ | `/opt/homebrew/bin/dev-browser` | — |
| Xcode toolchain / cgo | darwin platform build, signed dev build | ✓ (Phase 1–2 builds) | macOS 27 SDK (ld warns "built for newer macOS 27.0 than being linked 11.0"; harmless, pre-existing) | — |
| Real Slack/Messages for MP4 + Klipy paste | ACTN-09, ACTN-10 native verification | user's apps | — | user-verified checkpoint only |
| Live API keys | none (fixtures only) | ✗ (by policy) | — | fakeapi fixtures |

Current suites green at research time: `go test -race ./...` all packages ok; `pnpm test` 23 pass `[VERIFIED: run this session]`.

**Missing dependencies with no fallback:** none.
**Missing dependencies with fallback:** native-app verification → user checkpoint at the end of the phase (`human_verify_mode: end-of-phase`).

## Validation Architecture

(`workflow.nyquist_validation` is `false` in `.planning/config.json`; this section is included because the orchestrator asked for it.)

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go `testing` (+ `-race`); node:test (Node 24, native TS); dev-browser headless (Chromium) against the `-tags server` build + fakeapi |
| Config file | none for Go; `frontend/package.json` `test` script; `scripts/smoke-browser-loop.sh` |
| Quick run command | `go test -race ./internal/... && (cd frontend && pnpm test)` |
| Full suite command | `go test -race ./... && (cd frontend && pnpm test && pnpm check) && bash scripts/smoke-browser-loop.sh && for f in scripts/uat/*.js; do bash scripts/smoke-browser-loop.sh --uat "$f" \|\| exit 1; done` |

### Phase Requirements → Test Map

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ACTN-03 | Link/Markdown text per provider (GIPHY query stripped, Klipy unaltered, Markdown escaping, empty title → `GIF`); PasteText runs the paste sequence | unit (Fake platform) | `go test -race ./internal/services -run 'Text\|Markdown\|Link'` | ❌ Wave 0 |
| ACTN-03 | CopyText writes `public.utf8-plain-text` on a private pasteboard | unit (darwin cgo) | `go test ./internal/platform -run CopyText` | ❌ Wave 0 |
| ACTN-04 | PageURL mapping; page actions omitted without it; OpenPage refuses non-https/non-vendor | unit | `go test -race ./internal/provider/... ./internal/services -run 'Page'` | ❌ Wave 0 |
| ACTN-04 | Menu omits Copy Page Link / Open in Browser for Klipy/Finer; ⌘O no-op | unit (node) + UAT | `cd frontend && pnpm test`; `scripts/uat/07-actions.js` | ❌ Wave 0 |
| ACTN-05/11 | Save: magic check before any write, `(2)`/`(3)` suffix, never overwrite, temp removed, mode 0644, sanitized names | unit (httptest + TempDir) | `go test -race ./internal/staging -run Save` | ❌ Wave 0 |
| ACTN-05 | ⌘S toast `Downloaded to Downloads`, file under `bin/uat-downloads` | UAT + smoke | `bash scripts/smoke-browser-loop.sh --uat scripts/uat/08-download.js` | ❌ Wave 0 |
| ACTN-06 | Square crop golden (frames, delays, loop, per-pixel), outside-frame placeholder, >15 MB refused, Clips refused | unit | `go test -race ./internal/imaging ./internal/services -run Square` | ❌ Wave 0 |
| ACTN-07 | Action list order/omissions/filter/shortcuts; popover placement clamp; ⌘K/Esc/↑↓↵ | unit (node) + UAT | `pnpm test`; `scripts/uat/07-actions.js` | ❌ Wave 0 |
| ACTN-09 | Clips rendition order 720→480→1080→360; MP4 `ftyp` validation; `.mp4` staging name; MP4 badge + labels | unit + UAT | `go test -race ./internal/provider/giphy ./internal/staging -run 'Clip\|MP4'` | ❌ Wave 0 |
| ACTN-10 | Clear at launch/shutdown; sweep after text copy; Klipy timer with generation guard (injected `after`) | unit | `go test -race ./internal/services ./internal/staging -run 'Clear\|Klipy\|Shutdown'` | ❌ Wave 0 |
| LIBR-01..03 | Store: add/remove, Touch move-to-top + cap 50, undo at original index, ClearRecents undo, corrupt file → error and no write, 0600 | unit | `go test -race ./internal/library` | ❌ Wave 0 |
| LIBR-02 | Recents recorded only on successful copy/paste of file/square/link/Markdown | unit (Fake + fake recorder) | `go test -race ./internal/services -run Recent` | ❌ Wave 0 |
| LIBR-03/04 | Sections in slot order, filter by title, cross-section navigation; tabs ⌘5/⌘6 render with no provider request | unit (node) + UAT | `pnpm test`; `scripts/uat/09-library.js` (assert fakeapi saw no request, e.g. via server log) | ❌ Wave 0 |
| LIBR-05 | Details fields shown/omitted, size format, link text | unit (node) + UAT | `pnpm test`; `scripts/uat/10-details.js` | ❌ Wave 0 |
| SRCH-08 | Save rules (1.5 s idle, on copy), dedupe, cap 20, chips ↑/↓/←/→/↵/⌘⌫ | unit (Go store + node) + UAT | `go test ./internal/library -run Search`; `scripts/uat/11-chips.js` | ❌ Wave 0 |
| Native | MP4 paste into Messages/Slack; Klipy paste after grace; Show in Finder; Downloads TCC prompt; ⌘Z/⌘⌫/⌘R in WKWebView; square GIF animates when pasted | manual (user) | — | end-of-phase checkpoint |

### Sampling Rate
- **Per task commit:** quick run command.
- **Per wave merge:** full Go + node suites, `pnpm check`, smoke script, the UAT scripts touched by the wave.
- **Phase gate:** full suite green, then the native checkpoint.

### Wave 0 Gaps
- [ ] `internal/provider/fakeapi`: GIPHY items gain `url` (`https://giphy.com/gifs/<slug>`), `source`, `fixed_height`; Klipy items gain `md` and `tags`; Clips gain `original`, `720p`/`480p` assets (and one scenario word without 720p); `/media/{name}.mp4` serves the 24-byte `ftyp` fixture; a `-wide`/`-tall` media name serves a non-square GIF; a media name that answers 500 and one that answers HTML 200 (F1/F3 drivers).
- [ ] `wire_server.go`: `DownloadsDir: "bin/uat-downloads"`, platform = `&platform.Fake{}`; smoke script `rm -rf bin/uat-downloads`; update the 422/E7 expectations.
- [ ] `internal/imaging/square_test.go` with generated fixtures and the disposal-aware render helper.
- [ ] `internal/library/store_test.go`, `internal/staging` Save tests, services tests with `platform.Fake` + fake recorder + injected `after`.
- [ ] `frontend/src/lib/{actions,library,format,popover}.test.ts`; extend `grid.test.ts` for the new key actions (both `"c"` and `"C"` with shift).
- [ ] New UAT scripts `scripts/uat/07-…11-*.js` (names are the planner's choice).

## Security Domain

`security_enforcement: true`, ASVS level 1.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | — (no accounts) |
| V3 Session Management | no | — |
| V4 Access Control | no | single local user |
| V5 Input Validation | yes | Go validates every `Gif` from JS (provider ID, lengths, URL allowlists); filenames whitelisted; Markdown alt escaped; library JSON version-checked |
| V6 Cryptography | no | (sha1 only for staging dir names, not security) |
| V8 Data Protection | yes | Library files 0600 in a 0700 dir; no secrets in library files; no media persisted (vendor terms) |
| V10 Malicious Code / V12 Files | yes | Download/crop size caps, magic validation, no-overwrite claim, no paths from JS, staging confined to its dir (existing `filepath.Rel` check) |
| V14 Configuration | yes | Server-build fakes never shipped (`-tags server` only) |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal via Klipy slug / Finer `..` / title in a download name | Tampering | Whitelist `[A-Za-z0-9_-]` for the id part, `slug()` for the title, assert result stays in the destination dir |
| Webview asks Go to open `file://`, `smb://`, custom schemes or phishing hosts | Elevation / Spoofing | `OpenPage`: https + vendor host; `OpenSource`: https only; never pass JS strings to `openURL:` unchecked |
| SSRF / arbitrary fetch via `downloadUrl`, `mediumUrl` or a redirect | Tampering / Info disclosure | Existing `Allow` func on the URL and on every redirect hop (staging.go:67-73), applied to Save and crop fetches too |
| Error page or HTML saved as `.gif`/`.mp4` | Tampering | Magic bytes before anything is written |
| Decompression bomb in square crop | DoS | 15 MB input cap, `DecodeConfig` canvas cap, busy lock (one crop at a time) |
| Overwriting a user file in Downloads | Tampering | `os.Link` claim, never `O_TRUNC` |
| Paths/URLs leaking into UI or logs via `*os.LinkError`/`*url.Error` text | Info disclosure | Fixed-text errors and typed kinds (existing pattern, actions.go:1-2, staging.go:24) |
| Library file tampering by another local process | Tampering | Re-validate snapshots on load (same `Gif` validation); out of scope beyond that (same-user threat) |
| Klipy media retained against vendor rules | Repudiation / compliance | Grace delete, sweep on next write, launch/quit wipe; Phase 4 clearance gate |

## Sources

### Primary (HIGH confidence)
- Repo source read this session: `internal/provider/{provider.go, giphy/giphy.go, klipy/klipy.go, finergifs/finergifs.go, fakeapi/fakeapi.go}`, `internal/services/{actions.go, gifs.go, settings.go, panel.go}`, `internal/staging/staging.go`, `internal/platform/{platform.go, platform_darwin.go, platform_darwin.m, platform_darwin.h, fake.go, platform_stub.go, platform_darwin_test.go}`, `internal/settings/settings.go`, `internal/shell/shell.go`, `main.go`, `wire_desktop.go`, `wire_server.go`, `frontend/src/{App.svelte, lib/grid.ts, lib/api.ts, lib/providers.ts, components/ResultGrid.svelte, components/Footer.svelte}`, `frontend/package.json`, `build/Taskfile.yml`, `build/darwin/Info*.plist`, `scripts/smoke-browser-loop.sh`, `scripts/uat/01-tracer.js`.
- Wails v3.0.0-beta.30 module source: `pkg/application/{application.go, application_options.go, application_darwin.go, application_darwin_delegate.m, application_server.go, services.go, browser_manager.go, clipboard_darwin.go, signal_handler_desktop.go}`.
- Go 1.27.2 `src/image/gif/writer.go` (bounds check, transparent index, EncodeAll config).
- Probes this session: square crop (cropprobe), `os.Link` claim (linkprobe), `go test -race ./...`, `pnpm test`, Klipy page URL curl.
- Raycast extension source (MIT) @ `4f1b962`: `src/models/{gif,giphy,klipy,finergifs}.ts`, `src/components/{GifActions,GifDetails}.tsx`, `src/lib/{localGifs,stripQParams,downloadFile,resolveSquareGifFile,resolveGifFile,cachedGifs,formatFileSize,getServiceFromUrl}.ts`.

### Secondary (MEDIUM confidence)
- [GIPHY optional settings (bundles)](https://developers.giphy.com/docs/optional-settings/)
- [GIPHY API schema (GIF object, renditions)](https://developers.giphy.com/docs/api/schema/)
- [GIPHY Clips (video.assets)](https://developers.giphy.com/docs/clips/)
- [Klipy GIFs search API](https://docs.klipy.com/gifs-api/gifs-search-api.md)
- [Apple: NSDownloadsFolderUsageDescription](https://developer.apple.com/documentation/bundleresources/information-property-list/nsdownloadsfolderusagedescription.md)
- [Apple forums: permission to open Downloads](https://developer.apple.com/forums/thread/731972)
- Milestone research `.planning/research/{STACK,PITFALLS,FEATURES,ARCHITECTURE,SUMMARY}.md`; Phase 2 `02-RESEARCH.md`; Phase 1 `01-PASTE-MATRIX.md` §4.

### Tertiary (LOW confidence)
- WebKit key-equivalent ordering and `e.key` casing under ⌘⇧ (training knowledge; A4, A5).

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH (no new dependencies; everything already in the repo and tested).
- Provider fields: HIGH for GIPHY/Klipy docs and fixtures; MEDIUM for Clips `url` shape and Klipy tag element type.
- Architecture: HIGH (follows existing seams; Wails lifecycle read in source).
- Square crop: HIGH (re-verified, 0 mismatches).
- Pitfalls: HIGH for Go-level ones; MEDIUM for TCC, Klipy grace, WebKit keys (native, user-verified).

**Research date:** 2026-10-17
**Valid until:** 2026-11-16 (stable stdlib/code findings); re-check GIPHY/Klipy docs and Wails beta pin if upgraded.
