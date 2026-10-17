# Gifsmith

## What This Is

Gifsmith is a standalone macOS menu bar app for finding and sharing animated GIFs: press a global hotkey, search GIPHY, GIPHY Clips, Klipy or The Finer Gifs Club, and copy or paste the GIF into whatever app you were using. It is a port of the Raycast "GIF Search" extension for people who don't use Raycast — starting with the author, who has replaced every other piece of Raycast with Vorssaint (which has no plugin system). Free, public release.

## Core Value

Hotkey → type → Enter puts the right GIF on the clipboard (or into the frontmost app) in seconds, without Raycast.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Global hotkey opens a floating search panel; app lives in the menu bar with no Dock icon
- [ ] Search and trending across four providers with a provider switcher, mirroring the extension: GIPHY, GIPHY Clips, Klipy, The Finer Gifs Club
- [ ] Users enter their own GIPHY and Klipy API keys in settings; Finer Gifs Club works with no key
- [ ] Animated GIF grid with paging/infinite scroll and configurable item size
- [ ] Actions: copy GIF file, paste GIF into frontmost app, copy/paste GIF link, copy/paste GIF Markdown, copy page link, open in browser, download, copy/paste square-cropped GIF
- [ ] Configurable default (Enter) action; default is Copy GIF
- [ ] Favorites and recents per provider, browsable as their own views
- [ ] GIF details view (metadata, tags, source links)
- [ ] Settings: API keys, hotkey, grid sizes, max results, default action, download folder, hide filename, GIPHY/Klipy locales
- [ ] Signed and notarized macOS download (Developer ID), distributed outside the Mac App Store

### Out of Scope

- Hosted key proxy — users bring their own keys; avoids running servers and needing GIPHY production approval. Revisit if adoption makes key setup a real barrier.
- Raycast's proxy (gif-search.raycast.com) — not ours; must never be called.
- Mac App Store — sandbox forbids synthesizing ⌘V into other apps, which "paste into frontmost app" requires.
- Windows — v1 is macOS only; keep Go code portable so Windows native glue can be added later.
- AI tool / MCP server (the extension's `search-gifs` tool) — not part of the core workflow.

## Context

- Source being ported: `raycast/extensions/extensions/gif-search` (MIT, ~2k lines TypeScript). Local copy: `/private/tmp/claude-501/-Users-ken-dev-gif-search/9f53d405-1914-4df1-8002-37e6a551a349/scratchpad/rc/extensions/gif-search` (scratchpad; re-clone with sparse checkout if gone).
- Portable from the extension: provider clients and response mapping (`models/*`), fetch-with-timeout errors (`lib/fetchProviderJson.ts`), dedupe, slugify, file cache naming, favorites/recents semantics (IDs unique only within a provider; recents prepend, favorites append).
- Raycast-provided pieces that must be rebuilt: grid UI, search bar + dropdown, action menu, toasts, details view, preferences UI, LocalStorage, frecency sorting, window/hotkey handling, `Clipboard.copy({file})`, `Clipboard.paste`.
- The extension's square crop shells out to `/usr/bin/swift`; Gifsmith should do it natively (Go `image/gif` or cgo ImageIO).
- GIPHY free (beta) keys have a low hourly request cap — search must be debounced and trending cached.
- Claude Code does the implementation. Frontend can be exercised with dev-browser against the `wails dev` server; native behaviors (paste, focus restore, Accessibility prompt, panel over full-screen apps) need manual testing by the user.

## Constraints

- **Tech stack**: Go + Wails (prefer v3 if stable at build time, else v2) with a web frontend — author's existing toolchain.
- **Platform**: macOS only for v1; native glue via cgo/Objective-C (NSPasteboard file URLs, CGEvent ⌘V, NSWorkspace frontmost app, NSPanel-style window behavior).
- **Permissions**: Paste requires Accessibility permission; copy must work without it.
- **Distribution**: Developer ID signing + notarization; no App Store.
- **API keys**: Never embedded in the binary; stored locally per user (Keychain preferred).
- **License**: Port of MIT code — retain attribution to the original extension authors.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go + Wails over native Swift / Electron | Author's existing stack; web view animates GIF grids well; Claude can verify UI via dev-browser | — Pending |
| User-entered API keys, no proxy | No infra, no GIPHY production approval; Finer Gifs works keyless out of the box | — Pending |
| Mirror all four extension providers + switcher | Feature parity with the extension being replaced | — Pending |
| Menu bar + hotkey, no Dock icon | Launcher-style utility | — Pending |
| Copy GIF is default action; paste is opt-in | Copy needs no Accessibility permission | — Pending |
| macOS only for v1 | Native glue is per-platform; ship one well | — Pending |
| Spike first: hotkey panel → copy GIF file → paste into TextEdit | De-risk the native parts before building UI | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-10-17 after initialization*
