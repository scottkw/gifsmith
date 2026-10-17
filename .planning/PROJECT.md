# Gifsmith

## What This Is

Gifsmith is a standalone macOS menu bar app for finding and sharing animated GIFs: press a global hotkey, search GIPHY, GIPHY Clips, Klipy or The Finer Gifs Club, and copy or paste the GIF into whatever app you were using. It is a port of the Raycast "GIF Search" extension for people who don't use Raycast — starting with the author, who has replaced every other piece of Raycast with Vorssaint (which has no plugin system). Free, public release.

## Core Value

Hotkey → type → Enter puts the right GIF on the clipboard (or into the frontmost app) in seconds, without Raycast.

## Requirements

### Validated

- ✓ Global hotkey (⌃⌥⌘G) opens a non-activating floating panel that takes typing while the previous app keeps focus, shows over full-screen apps, hides on Esc/click-away; menu bar only, no Dock icon — Phase 1
- ✓ Copy GIF as a file (no Accessibility needed) and ⌘Enter auto-paste into the previous app (Accessibility-gated, copy fallback) — Phase 1 (TextEdit, Notes, Mail, Messages; other apps untested)
- ✓ Developer ID–signed dev build with a stable identity; Accessibility grant survives rebuilds — Phase 1 (notarization deferred to Phase 4)

### Active

- [ ] Search and trending across four providers with a provider switcher, mirroring the extension: GIPHY, GIPHY Clips, Klipy, The Finer Gifs Club
- [ ] Users enter their own GIPHY and Klipy API keys in settings; Finer Gifs Club works with no key
- [ ] Animated GIF grid with paging/infinite scroll and configurable item size
- [ ] Actions: copy GIF file, paste GIF into frontmost app, copy/paste GIF link, copy/paste GIF Markdown, copy page link, open in browser, download, copy/paste square-cropped GIF
- [ ] Configurable default (Enter) action; default is Copy GIF
- [ ] Favorites and recents per provider, browsable as their own views
- [ ] GIF details view (metadata, tags, source links)
- [ ] Settings: API keys, hotkey, grid sizes, max results, default action, download folder, hide filename, GIPHY/Klipy locales
- [ ] Signed and notarized macOS download (Developer ID), distributed outside the Mac App Store
- [ ] Accessibility onboarding deep-links to the permission list (macOS 27 has no browsable path) — emerged in Phase 1

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
| Spike first: hotkey panel → copy GIF file → paste into TextEdit | De-risk the native parts before building UI | ✓ Good (GO WITH FALLBACKS, 2026-10-17): native architecture holds on the Developer ID–signed dev build; limits recorded in 01-PASTE-MATRIX.md §4 |
| Wails v3 pinned at v3.0.0-beta.30 (no stable 3.0 yet) | v2 can't deliver tray, NSPanel, global hotkey or accessory policy without shims; v3 betas are nightly snapshots, so pin the exact version for Go module, `wails3` CLI and `@wailsio/runtime` (STACK.md) | ✓ Good: the Phase 1 spike ran on beta.30 with no framework blockers |
| Pasteboard payload: single file URL | D-12: file URL only unless the matrix proves otherwise | ✓ Sufficient for TextEdit, Notes, Mail, Messages (copy and ⌘Enter, animated); Slack, Discord, Telegram, Safari/Chrome composers unverified (Phase 3 design note) |
| Notarization deferred to Phase 4 (Apple agreement 403) | `notarytool store-credentials` returns 403 "A required agreement is missing or has expired"; the user chose to go on the signed dev build | ⚠ Open: the user must resolve the Apple Developer agreement before Phase 4; notarized dry run + N15 carry over |
| Phase 2 compliance posture (GIPHY, Klipy, Finer Gifs) | No persistent media cache: only the most recent staged copy file, removed on the next successful copy. API responses are cached in memory only (trending 10 min, search 5 min, no media bytes). Providers are never blended, and results are never reordered or filtered, except same-page duplicate removal for non-Klipy providers and the logged skip of Klipy items that have no GIF rendition (items with GIF renditions, ads included, stay in place — WR-04). Media loads directly from provider URLs. Attribution: the official GIPHY mark on GIPHY/Clips, "Search KLIPY" + "Powered by KLIPY" on Klipy, a courtesy credit on Finer Gifs. User-owned keys are used only from the user's machine. No analytics or pingbacks. Defaults: GIPHY rating pg-13, Klipy content_filter medium. Clips is probed at most once per launch, only after the key worked on GIPHY GIFs (D-20) | ✓ Recorded (Phase 2). ⚠ Klipy written clearance for temp-file copies still pending (Phase 4 gate) |
| API keys in the macOS Keychain via zalando/go-keyring v0.2.8; webview sees configured booleans only | Pure Go, survives ad-hoc dev rebuilds without prompts; no bound method returns a key and errors/logs never contain one | ✓ Built (Phase 2); real Keychain save and relaunch verified natively at the 02-06 Task 3 checkpoint |

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
*Last updated: 2026-10-17 after Phase 1*
