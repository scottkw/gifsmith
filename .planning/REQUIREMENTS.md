# Requirements: Gifsmith

**Defined:** 2026-10-17
**Core Value:** Hotkey → type → Enter puts the right GIF on the clipboard (or into the frontmost app) in seconds, without Raycast.

## v1 Requirements

### Shell

- [x] **SHEL-01**: User can open and close a floating search panel with a global hotkey (default three-modifier combo, e.g. ⌃⌥⌘G)
- [x] **SHEL-02**: Panel appears over full-screen apps, takes typing immediately, and does not steal focus from the app the user was in
- [x] **SHEL-03**: Panel closes on Esc and on click-away
- [x] **SHEL-04**: App lives in the menu bar with no Dock icon; menu offers Open, Settings, Quit
- [ ] **SHEL-05**: User can enable launch at login
- [ ] **SHEL-06**: User can open Favorites or Recents directly with their own hotkeys

### Search

- [ ] **SRCH-01**: User can search GIPHY, Klipy, and The Finer Gifs Club, switching provider from the panel; last provider is remembered
- [ ] **SRCH-02**: GIPHY Clips appears as a provider only when the user's GIPHY key passes a Clips test call; otherwise a short explanation is shown
- [x] **SRCH-03**: With an empty query, user sees trending GIFs (providers that support trending)
- [x] **SRCH-04**: Results show as an animated grid that loads more as the user scrolls or arrows near the end
- [x] **SRCH-05**: User can navigate the grid entirely by keyboard (arrows, Enter, ⌘K, Esc)
- [x] **SRCH-06**: Searches are debounced so free keys (100 calls/hour) last; a rate-limit hit shows a clear message with when to retry
- [x] **SRCH-07**: Network, timeout, and provider errors show a per-provider message; one provider failing never blocks the others
- [ ] **SRCH-08**: User sees recent searches when the box is empty and can rerun one
- [ ] **SRCH-09**: User can set GIPHY content rating (default PG-13)

### Onboarding & Compliance

- [ ] **ONBD-01**: First run opens on GIPHY with an inline setup card to paste a key, plus a "Try Finer Gifs (no key)" button
- [ ] **ONBD-02**: Pasted key is verified with a test call before being saved
- [ ] **ONBD-03**: API keys are stored in the macOS Keychain and never shown to the web UI or written to logs/errors
- [ ] **ONBD-04**: "Powered by GIPHY" is shown on GIPHY/Clips results; Klipy search placeholder reads "Search KLIPY"
- [x] **ONBD-05**: Providers are never blended in one grid; results are never reordered or filtered (except same-page duplicates for non-Klipy)

### Actions

- [x] **ACTN-01**: User can copy the GIF as a file to the clipboard (default Enter action), without needing Accessibility permission
- [x] **ACTN-02**: User can paste the GIF directly into the app they were using; if Accessibility isn't granted, the GIF is copied and the user is told to press ⌘V and how to grant permission
- [ ] **ACTN-03**: User can copy or paste the GIF link, and copy or paste GIF Markdown
- [ ] **ACTN-04**: User can copy the page link and open the GIF page in the browser
- [ ] **ACTN-05**: User can download the GIF to a chosen folder (default Downloads)
- [ ] **ACTN-06**: User can copy or paste a center-cropped square version of a GIF (not offered for Clips)
- [ ] **ACTN-07**: User can open an action menu (⌘K) listing all actions for the selected GIF
- [ ] **ACTN-08**: User can choose the default Enter action and a second ⌘Enter action in settings
- [ ] **ACTN-09**: Clips are handled as MP4 (labelled accordingly, 720p default)
- [ ] **ACTN-10**: Temp files used for copy/paste/download are deleted promptly; Klipy temp files are deleted eagerly
- [ ] **ACTN-11**: Downloaded files are validated (real GIF/MP4, not an error page) and never overwrite or collide; optional hide-original-filename setting

### Library

- [ ] **LIBR-01**: User can add/remove a GIF to Favorites
- [ ] **LIBR-02**: Every successful copy/paste records the GIF in Recents (most recent first, capped)
- [ ] **LIBR-03**: User can browse Favorites and Recents, grouped by provider (never blended), and remove items from Recents
- [ ] **LIBR-04**: Favorites and Recents render from saved metadata without API calls (work offline/rate-limited)
- [ ] **LIBR-05**: User can open a details view for a GIF (title, size, dimensions, tags, source links)

### Settings

- [ ] **SETT-01**: Settings open in their own window, where pasting into fields works
- [ ] **SETT-02**: User can add, test, and remove GIPHY and Klipy keys
- [ ] **SETT-03**: User can rebind the global hotkeys; conflicts are reported and the old binding stays until the new one registers
- [ ] **SETT-04**: User can set grid item size (search and trending), max results, download folder, hide filename, GIPHY and Klipy locales
- [ ] **SETT-05**: Settings shows Accessibility permission status and a button to grant it

### Release

- [ ] **RELS-01**: App ships as a Developer ID–signed, notarized DMG for macOS 13+ on Apple Silicon
- [ ] **RELS-02**: App installs and runs on a clean user account (Gatekeeper, fresh permissions)
- [ ] **RELS-03**: About screen credits the original MIT-licensed Raycast extension and authors
- [ ] **RELS-04**: Klipy file actions are disabled (link/Markdown only) unless Klipy gives written clearance for temp file copies

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
| In-app auto-updater | Wails updater is immature; revisit post-launch |
| AI tool / MCP server | Not the core workflow |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| SHEL-01 | Phase 1 | Complete |
| SHEL-02 | Phase 1 | Complete |
| SHEL-03 | Phase 1 | Complete |
| SHEL-04 | Phase 1 | Complete |
| SHEL-05 | Phase 4 | Pending |
| SHEL-06 | Phase 4 | Pending |
| SRCH-01 | Phase 2 | Pending |
| SRCH-02 | Phase 2 | Pending |
| SRCH-03 | Phase 2 | Complete |
| SRCH-04 | Phase 2 | Complete |
| SRCH-05 | Phase 2 | Complete |
| SRCH-06 | Phase 2 | Complete |
| SRCH-07 | Phase 2 | Complete |
| SRCH-08 | Phase 3 | Pending |
| SRCH-09 | Phase 4 | Pending |
| ONBD-01 | Phase 2 | Pending |
| ONBD-02 | Phase 2 | Pending |
| ONBD-03 | Phase 2 | Pending |
| ONBD-04 | Phase 2 | Pending |
| ONBD-05 | Phase 2 | Complete |
| ACTN-01 | Phase 1 | Complete |
| ACTN-02 | Phase 1 | Complete |
| ACTN-03 | Phase 3 | Pending |
| ACTN-04 | Phase 3 | Pending |
| ACTN-05 | Phase 3 | Pending |
| ACTN-06 | Phase 3 | Pending |
| ACTN-07 | Phase 3 | Pending |
| ACTN-08 | Phase 4 | Pending |
| ACTN-09 | Phase 3 | Pending |
| ACTN-10 | Phase 3 | Pending |
| ACTN-11 | Phase 3 | Pending |
| LIBR-01 | Phase 3 | Pending |
| LIBR-02 | Phase 3 | Pending |
| LIBR-03 | Phase 3 | Pending |
| LIBR-04 | Phase 3 | Pending |
| LIBR-05 | Phase 3 | Pending |
| SETT-01 | Phase 4 | Pending |
| SETT-02 | Phase 4 | Pending |
| SETT-03 | Phase 4 | Pending |
| SETT-04 | Phase 4 | Pending |
| SETT-05 | Phase 4 | Pending |
| RELS-01 | Phase 4 | Pending |
| RELS-02 | Phase 4 | Pending |
| RELS-03 | Phase 4 | Pending |
| RELS-04 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 45 total
- Mapped to phases: 45 (Phase 1: 6, Phase 2: 12, Phase 3: 14, Phase 4: 13)
- Unmapped: 0

---
*Requirements defined: 2026-10-17*
*Last updated: 2026-10-17 after roadmap creation (traceability filled)*
