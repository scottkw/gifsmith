# Requirements: Gifsmith

**Defined:** 2026-10-17
**Core Value:** Hotkey → type → Enter puts the right GIF on the clipboard (or into the frontmost app) in seconds, without Raycast.

## v1 Requirements

### Shell

- [x] **SHEL-01**: User can open and close a floating search panel with a global hotkey (default three-modifier combo, e.g. ⌃⌥⌘G)
- [x] **SHEL-02**: Panel appears over full-screen apps, takes typing immediately, and does not steal focus from the app the user was in
- [x] **SHEL-03**: Panel closes on Esc and on click-away
- [x] **SHEL-04**: App lives in the menu bar with no Dock icon; menu offers Open, Settings, Quit
- [x] **SHEL-05**: User can enable launch at login
- [x] **SHEL-06**: User can open Favorites or Recents directly with their own hotkeys

### Search

- [x] **SRCH-01**: User can search GIPHY, Klipy, and The Finer Gifs Club, switching provider from the panel; last provider is remembered
- [x] **SRCH-02**: GIPHY Clips appears as a provider only when the user's GIPHY key passes a Clips test call; otherwise a short explanation is shown
- [x] **SRCH-03**: With an empty query, user sees trending GIFs (providers that support trending)
- [x] **SRCH-04**: Results show as an animated grid that loads more as the user scrolls or arrows near the end
- [x] **SRCH-05**: User can navigate the grid entirely by keyboard (arrows, Enter, ⌘K, Esc)
- [x] **SRCH-06**: Searches are debounced so free keys (100 calls/hour) last; a rate-limit hit shows a clear message with when to retry
- [x] **SRCH-07**: Network, timeout, and provider errors show a per-provider message; one provider failing never blocks the others
- [x] **SRCH-08**: User sees recent searches when the box is empty and can rerun one
- [x] **SRCH-09**: User can set GIPHY content rating (default PG-13)

### Onboarding & Compliance

- [x] **ONBD-01**: First run opens on GIPHY with an inline setup card to paste a key, plus a "Try Finer Gifs (no key)" button
- [x] **ONBD-02**: Pasted key is verified with a test call before being saved
- [x] **ONBD-03**: API keys are stored in the macOS Keychain and never shown to the web UI or written to logs/errors
- [x] **ONBD-04**: "Powered by GIPHY" is shown on GIPHY/Clips results; Klipy search placeholder reads "Search KLIPY"
- [x] **ONBD-05**: Providers are never blended in one grid; results are never reordered or filtered (except same-page duplicates for non-Klipy)

### Actions

- [x] **ACTN-01**: User can copy the GIF as a file to the clipboard (default Enter action), without needing Accessibility permission
- [x] **ACTN-02**: User can paste the GIF directly into the app they were using; if Accessibility isn't granted, the GIF is copied and the user is told to press ⌘V and how to grant permission
- [x] **ACTN-03**: User can copy or paste the GIF link, and copy or paste GIF Markdown
- [x] **ACTN-04**: User can copy the page link and open the GIF page in the browser
- [x] **ACTN-05**: User can download the GIF to a chosen folder (default Downloads)
- [x] **ACTN-06**: User can copy or paste a center-cropped square version of a GIF (not offered for Clips)
- [x] **ACTN-07**: User can open an action menu (⌘K) listing all actions for the selected GIF
- [x] **ACTN-08**: User can choose the default Enter action and a second ⌘Enter action in settings
- [x] **ACTN-09**: Clips are handled as MP4 (labelled accordingly, 720p default)
- [x] **ACTN-10**: Temp files used for copy/paste/download are deleted promptly; Klipy temp files are deleted eagerly
- [x] **ACTN-11**: Downloaded files are validated (real GIF/MP4, not an error page) and never overwrite or collide; optional hide-original-filename setting

### Library

- [x] **LIBR-01**: User can add/remove a GIF to Favorites
- [x] **LIBR-02**: Every successful copy/paste records the GIF in Recents (most recent first, capped)
- [x] **LIBR-03**: User can browse Favorites and Recents, grouped by provider (never blended), and remove items from Recents
- [x] **LIBR-04**: Favorites and Recents render from saved metadata without API calls (work offline/rate-limited)
- [x] **LIBR-05**: User can open a details view for a GIF (title, size, dimensions, tags, source links)

### Settings

- [x] **SETT-01**: Settings open in their own window, where pasting into fields works
- [x] **SETT-02**: User can add, test, change (replace), and remove GIPHY and Klipy keys; each key row links to the provider's key sign-up page (GIPHY developer dashboard, Klipy Partner Panel)
- [x] **SETT-03**: User can rebind the global hotkeys; conflicts are reported and the old binding stays until the new one registers
- [x] **SETT-04**: User can set grid item size (search and trending), max results, download folder, hide filename, GIPHY and Klipy locales
- [x] **SETT-05**: Settings shows Accessibility permission status and a button to grant it
- [x] **SETT-06**: User can choose the window theme: Light, Dark, or System (default System, follows macOS appearance)

### Release

- [ ] **RELS-01**: App ships as a Developer ID–signed, notarized DMG for macOS 13+ on Apple Silicon
- [ ] **RELS-02**: App installs and runs on a clean user account (Gatekeeper, fresh permissions)
- [ ] **RELS-03**: About screen credits the original MIT-licensed Raycast extension and authors
- [ ] **RELS-04**: Klipy file actions are disabled (link/Markdown only) unless Klipy gives written clearance for temp file copies
- [ ] **RELS-05**: App checks GitHub Releases (scottkw/gifsmith) for a newer version and alerts the user when one is available; Settings has a "Check for Updates…" button
- [ ] **RELS-06**: If the user chooses, the app downloads the new DMG, verifies its code signature and notarization, and opens it (no in-place self-replacement)

## v2 Requirements

### Differentiators

- **DIFF-01**: Drag a GIF out of the panel into another app (v1.x, right after launch)
- **DIFF-02**: Category / suggestion chips and autocomplete
- **DIFF-03**: Request-budget indicator per provider
- **DIFF-04**: Opt-in offline media cache for favorites (only if vendor terms allow)
- **DIFF-05**: Quick full-size preview, menu bar recents submenu
- **DIFF-06**: Windows build

## Out of Scope

| Feature | Reason |
|---------|--------|
| Hosted API-key proxy | Users bring their own keys; no infra or GIPHY production approval needed |
| Raycast's proxy (gif-search.raycast.com) | Not ours; never called |
| Tenor | API shut down 2026-06-30 |
| Blended cross-provider grid / unified search | Forbidden by GIPHY and Klipy terms |
| Persistent GIF media cache | Vendor terms forbid storing media |
| Mac App Store build | Sandbox blocks synthesized ⌘V |
| Intel Macs | Apple Silicon only keeps the native build simple |
| GIF recording/editing, stickers, AI | Not the core workflow |
| Accounts, sync, telemetry, analytics pingbacks | Not needed; privacy |
| In-place self-replacing auto-install | Update flow downloads, verifies and opens the DMG instead (RELS-06); self-replacement revisited post-launch |
| AI tool / MCP server | Not the core workflow |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| SHEL-01 | Phase 1 | Complete |
| SHEL-02 | Phase 1 | Complete |
| SHEL-03 | Phase 1 | Complete |
| SHEL-04 | Phase 1 | Complete |
| SHEL-05 | Phase 4 | Complete |
| SHEL-06 | Phase 4 | Complete |
| SRCH-01 | Phase 2 | Complete |
| SRCH-02 | Phase 2 | Complete |
| SRCH-03 | Phase 2 | Complete |
| SRCH-04 | Phase 2 | Complete |
| SRCH-05 | Phase 2 | Complete |
| SRCH-06 | Phase 2 | Complete |
| SRCH-07 | Phase 2 | Complete |
| SRCH-08 | Phase 3 | Complete |
| SRCH-09 | Phase 4 | Complete |
| ONBD-01 | Phase 2 | Complete |
| ONBD-02 | Phase 2 | Complete |
| ONBD-03 | Phase 2 | Complete |
| ONBD-04 | Phase 2 | Complete |
| ONBD-05 | Phase 2 | Complete |
| ACTN-01 | Phase 1 | Complete |
| ACTN-02 | Phase 1 | Complete |
| ACTN-03 | Phase 3 | Complete |
| ACTN-04 | Phase 3 | Complete |
| ACTN-05 | Phase 3 | Complete |
| ACTN-06 | Phase 3 | Complete |
| ACTN-07 | Phase 3 | Complete |
| ACTN-08 | Phase 4 | Complete |
| ACTN-09 | Phase 3 | Complete |
| ACTN-10 | Phase 3 | Complete |
| ACTN-11 | Phase 3 | Complete |
| LIBR-01 | Phase 3 | Complete |
| LIBR-02 | Phase 3 | Complete |
| LIBR-03 | Phase 3 | Complete |
| LIBR-04 | Phase 3 | Complete |
| LIBR-05 | Phase 3 | Complete |
| SETT-01 | Phase 4 | Complete |
| SETT-02 | Phase 4 | Complete |
| SETT-03 | Phase 4 | Complete |
| SETT-04 | Phase 4 | Complete |
| SETT-05 | Phase 4 | Complete |
| SETT-06 | Phase 4 | Complete |
| RELS-01 | Phase 4 | Pending |
| RELS-02 | Phase 4 | Pending |
| RELS-03 | Phase 4 | Pending |
| RELS-04 | Phase 4 | Pending |
| RELS-05 | Phase 4 | Pending |
| RELS-06 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 48 total
- Mapped to phases: 48 (Phase 1: 6, Phase 2: 12, Phase 3: 14, Phase 4: 16)
- Unmapped: 0

---
*Requirements defined: 2026-10-17*
*Last updated: 2026-10-18 — added RELS-05/06 (GitHub update check + verified DMG download) at user request during Phase 4*
