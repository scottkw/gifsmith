# Roadmap: Gifsmith

## Overview

Gifsmith ports the Raycast "GIF Search" extension to a standalone macOS menu bar app (Go + Wails v3). The journey is risk-first: a small, hard go/no-go spike proves the native behaviors the whole product depends on (non-activating hotkey panel, copy GIF as a file, paste into the previous app, signed and notarized builds). Then one vertical slice delivers real search end to end (providers, animated grid, first-run key setup, vendor-compliant attribution) so Hotkey → type → Enter works. Next the full action pipeline and the favorites/recents library complete the extension's feature set. Finally the Settings window, shell polish, and a signed, notarized DMG make it releasable to strangers.

## Milestones

- 🚧 **v1.0** Gifsmith MVP

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Native Spike & Signed Dev Loop** - Hard go/no-go gate: hotkey panel → copy GIF file → paste into TextEdit, from signed and notarized builds (completed 2026-10-17)
- [x] **Phase 2: Providers, Search UI & First Run** - Vertical slice: search/trending → animated grid → copy, across Finer Gifs, GIPHY, Klipy (and Clips), with key setup and attribution (completed 2026-10-17)
- [ ] **Phase 3: Action Pipeline & Library** - Every extension action (link, Markdown, download, square crop, ⌘K menu) plus favorites, recents, details, and recent searches
- [ ] **Phase 4: Settings, Shell Polish & Release** - Settings window, rebindable hotkeys, launch at login, then a signed, notarized DMG that installs on a clean Mac

## Phase Details

### Phase 1: Native Spike & Signed Dev Loop

**Goal**: The three unproven native behaviors work on this Mac (typing into a non-activating panel, full-screen overlay, and copy/paste of a GIF file into the previous app), from a stable signed identity, so the architecture is confirmed before any real UI is built.
**Mode:** mvp
**Depends on**: Nothing (first phase)
**Requirements**: SHEL-01, SHEL-02, SHEL-03, SHEL-04, ACTN-01, ACTN-02
**Success Criteria** (what must be TRUE):
  1. From any app, including one in macOS full-screen, pressing the global hotkey (default three-modifier combo, e.g. ⌃⌥⌘G) shows a floating panel whose input takes typing immediately, and the app the user came from keeps focus. Pressing the hotkey again, pressing Esc, or clicking away hides the panel. (manual, user-verified)
  2. Gifsmith appears only as a menu bar icon (no Dock icon, not in ⌘Tab); its menu offers Open, Settings, and Quit, and Open shows the panel when the hotkey is unavailable. The Settings entry is wired to the real window in Phase 4.
  3. Pressing Enter in the panel puts a bundled sample GIF on the clipboard as a file, and ⌘V pastes it as an animated GIF into TextEdit, with Accessibility permission NOT granted. (manual, user-verified)
  4. With Accessibility granted, the paste action drops the GIF straight into the app the user was in (TextEdit) with no manual ⌘V. Without it, the GIF is copied and the user is told to press ⌘V and how to grant permission. (manual, user-verified)
  5. The same behaviors hold from a signed dev build run from the bundled `.app` and from a throwaway Developer ID–signed, notarized build. A committed paste matrix records the result per target app (Slack, Messages, Mail, Discord, Notes, Telegram, Safari/Chrome composers), and the user records an explicit go/no-go. This de-risks RELS-01/RELS-02, which complete in Phase 4. (manual, user-verified)

**Plans:** 4/4 plans complete

Plans:
**Wave 1**
- [x] 01-01-PLAN.md — Walking skeleton: scaffold + identity, then ⌃⌥⌘G non-activating panel → Enter copies sample GIF as a file (tracer gate)

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 01-02-PLAN.md — Menu-bar shell: tray menu (Open / Settings… / Quit), panel on the cursor's display

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 01-03-PLAN.md — Signed dev loop (install:dev to ~/Applications) + ⌘Enter Accessibility-gated paste; paste matrix (dev column); browser UAT smoke loop

**Wave 4** *(blocked on Wave 3 completion)*
- [x] 01-04-PLAN.md — Notarized + stapled throwaway build, notarized matrix column, user go/no-go recorded and committed (GO WITH FALLBACKS; notarized build deferred to Phase 4, Apple agreement 403)

**Manual verification**: Required for every criterion above. Panel focus, full-screen overlay, Accessibility prompt, and real paste cannot be exercised by dev-browser. Phase scaffolding also pins Wails v3 `v3.0.0-beta.30` with pnpm, proves the `-tags server` browser→bindings loop used for UAT in Phases 2-3, and settles min-macOS/arch (macOS 13+, arm64).
**Research flag**: Highest care. Use `/gsd-plan-phase --research-phase 1`; research is done, the work is hands-on spike.

### Phase 2: Providers, Search UI & First Run

**Goal**: A stranger can open the panel, get set up with a key (or skip to the keyless provider), search or browse trending GIFs on any available provider in a fast keyboard-driven animated grid, and Enter copies the GIF, with vendor-required attribution and honest error handling.
**Depends on**: Phase 1 (go decision, signed dev loop, copy-GIF-as-file primitive, server-mode UAT loop)
**Requirements**: SRCH-01, SRCH-02, SRCH-03, SRCH-04, SRCH-05, SRCH-06, SRCH-07, ONBD-01, ONBD-02, ONBD-03, ONBD-04, ONBD-05
**Success Criteria** (what must be TRUE):
  1. With an empty box the panel shows trending GIFs for providers that support trending; typing searches the selected provider (debounced) and shows an animated grid that loads more as the user scrolls or arrows near the end. The user can switch between Finer Gifs, GIPHY, and Klipy from the panel, the last provider is remembered, and providers are never blended in one grid or reordered/filtered.
  2. The user can drive the whole panel from the keyboard: arrows move the selection, Enter copies the selected GIF as a file that pastes with ⌘V, Esc closes. (⌘K is bound here and opens the action menu in Phase 3.)
  3. First run opens on GIPHY with an inline setup card. Pasting a key is verified with a test call before it is saved (a bad key is rejected with a clear message), and "Try Finer Gifs (no key)" works without any key. Saved keys live in the macOS Keychain and never appear in the UI, logs, or error text.
  4. "Powered by GIPHY" shows on GIPHY and Clips results, and the Klipy search placeholder reads "Search KLIPY".
  5. GIPHY Clips appears as a provider only when the user's GIPHY key passes a Clips test call (otherwise a short explanation is shown). Rapid typing does not burn requests, a rate-limit hit shows a clear message with when to retry, and a network, timeout, or provider error shows a per-provider message while other providers stay usable.

**Plans:** 6/6 plans complete

Plans:
**Wave 1**
- [x] 02-01-PLAN.md — Tracer: Finer Gifs search in Go → animated grid → Enter copies the real GIF (typed URL-free errors, staging, smoke/UAT harness)

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 02-02-PLAN.md — Virtualized grid with paging to End of results, full keyboard model, panel owns Esc and survives ⌘W (npm package gate)

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 02-03-PLAN.md — search.Service cache + 429 budget + dedupe policy; every loading/empty/error state; cancellation of superseded searches

**Wave 4** *(blocked on Wave 3 completion)*
- [x] 02-04-PLAN.md — GIPHY and Klipy clients, Keychain-backed keys verified by a test call, provider status, last provider (Go, end to end via bound API)

**Wave 5** *(blocked on Wave 4 completion)*
- [x] 02-05-PLAN.md — Provider tabs (⌘1–⌘4, remembered), attribution and placeholders, first-run setup card, revoked-key re-entry

**Wave 6** *(blocked on Wave 5 completion)*
- [x] 02-06-PLAN.md — GIPHY Clips probe and states, compliance posture in PROJECT.md, UI backstops, signed install + native UAT with real keys

**UI hint**: yes

**Notes**: Copy works for Finer Gifs, GIPHY, and Klipy results here; Clips file handling (MP4) lands in Phase 3 (ACTN-09), and temp-file lifecycle hardening in Phase 3 (ACTN-10). Written compliance posture goes into PROJECT.md Key Decisions in this phase. Default PG-13 rating applies to GIPHY now; the user-facing control is in Phase 4.
**Research flag**: Needs `/gsd-plan-phase --research-phase 2`. Short live-key probing session for Klipy request shapes, Clips approval behavior, the system-proxy choice, and a Finer Gifs re-probe (it was down 2026-10-17).

### Phase 3: Action Pipeline & Library

**Goal**: Every action the Raycast extension offered works on the selected GIF, and favorites, recents, details, and recent searches put the GIFs a user reuses one keystroke away, even offline or rate-limited.
**Depends on**: Phase 2
**Requirements**: ACTN-03, ACTN-04, ACTN-05, ACTN-06, ACTN-07, ACTN-09, ACTN-10, ACTN-11, LIBR-01, LIBR-02, LIBR-03, LIBR-04, LIBR-05, SRCH-08
**Success Criteria** (what must be TRUE):
  1. ⌘K opens an action menu listing every action for the selected GIF. The user can copy or paste the GIF link, copy or paste GIF Markdown, copy the page link, and open the GIF page in the browser.
  2. The user can download a GIF to a chosen folder (default Downloads) without overwriting an existing file, and a response that is not a real GIF/MP4 (such as an error page) is rejected rather than saved. The user can copy or paste a center-cropped square GIF with animation intact; that option is not offered for Clips. Clips are labelled as MP4 and use 720p by default.
  3. Temp files used for copy, paste, and download are removed promptly, and Klipy temp files are removed eagerly, so the staging folder does not accumulate GIFs.
  4. The user can add or remove a favorite, and every successful copy or paste appears in Recents (most recent first, capped). Favorites and Recents are browsable grouped by provider (never blended), items can be removed from Recents, and both render from saved metadata with no API calls.
  5. The user can open a details view for a GIF (title, size, dimensions, tags, source links), and an empty search box shows recent searches that can be rerun.

**Plans:** 8/8 plans executed

Plans:
**Wave 1**
- [x] 03-01-PLAN.md — Tracer: ⌘⇧C copies the link through Go + Platform.CopyText (server build on platform.Fake); Gif metadata fields; Markdown, page link, paste variants, Open in Browser

**Wave 2** *(blocked on Wave 1 completion)*
- [x] 03-02-PLAN.md — ⌘K action menu: pure action table + placement, popover with filter/keyboard/mouse, one runAction dispatcher

**Wave 3** *(blocked on Wave 2 completion)*
- [x] 03-03-PLAN.md — Clips copy/paste as the 720p MP4 (badge, labels, E8 retired); staging lifecycle (launch/quit wipe, clear after text copy, Klipy 30 s grace delete)

**Wave 4** *(blocked on Wave 3 completion)*
- [x] 03-04-PLAN.md — Download (validated, atomic no-overwrite, injected folder) with Show in Finder (⌘R) and the Downloads usage description

**Wave 5** *(blocked on Wave 4 completion)*
- [x] 03-05-PLAN.md — Square GIF: lossless image/gif crop with golden tests, 15 MB/canvas limits, Copy/Paste Square, 300 ms busy rule

**Wave 6** *(blocked on Wave 5 completion)*
- [x] 03-06-PLAN.md — Favorites & Recents: library store + LibraryService, Recents recording, sectioned tabs ⌘5/⌘6, remove/Clear/Undo

**Wave 7** *(blocked on Wave 6 completion)*
- [x] 03-07-PLAN.md — Details pane (⌘I), recent-search chips, Phase 2 polish (P1–P11, IN-04) and UI backstops

**Wave 8** *(blocked on Wave 7 completion)*
- [x] 03-08-PLAN.md — Full-suite regression, signed dev install, native UAT checkpoint

**UI hint**: yes

**Notes**: Default Enter action is Copy GIF and ⌘Enter is a built-in sibling until Phase 4 exposes both in Settings (ACTN-08). Favorites persist metadata snapshots only, with no media cache (vendor terms). Paste variants reuse the Phase 1 Accessibility-gated paste path.
**Research flag**: Needs a short design note on pasteboard representation (informed by the Phase 1 paste matrix) and Clips handling before coding. Library and details are standard patterns.

### Phase 4: Settings, Shell Polish & Release

**Goal**: Everything is configurable from a proper Settings window with hotkeys that can be rebound and launch at login, and a stranger can install a signed, notarized build on a clean Mac and use it within vendor and license obligations.
**Depends on**: Phase 3
**Requirements**: SETT-01, SETT-02, SETT-03, SETT-04, SETT-05, SETT-06, SRCH-09, ACTN-08, SHEL-05, SHEL-06, RELS-01, RELS-02, RELS-03, RELS-04
**Success Criteria** (what must be TRUE):
  1. Choosing Settings from the menu bar menu opens a separate Settings window where pasting into fields works. There the user can add, test, change, and remove GIPHY and Klipy keys (each with a link to that provider's key sign-up page), and see Accessibility permission status with a button to grant it.
  2. The user can rebind the global hotkeys; a conflict is reported and the old binding stays until the new one registers. Favorites and Recents each open directly with their own hotkeys, and the user can turn launch at login on or off. (manual, user-verified)
  3. The user can set grid item size (search and trending), max results, download folder, hide filename, GIPHY and Klipy locales, GIPHY content rating (default PG-13), the default Enter and ⌘Enter actions, and the window theme (Light, Dark, or System), and changes take effect on next use.
  4. A Developer ID–signed, notarized DMG installs and runs on a clean user account (Gatekeeper accepts it, permissions start fresh) on macOS 13+ Apple Silicon. (manual, user-verified)
  5. The About screen credits the original MIT-licensed Raycast extension and its authors, and Klipy file actions stay disabled (link and Markdown only) unless written clearance from Klipy is recorded.

**Plans**: TBD
**UI hint**: yes

**Manual verification**: Criteria 2 and 4 need user-run native checks (global shortcut registration, login item, clean-account install). Klipy clearance (developers@klipy.com) is a pre-release gate and must be requested before this phase's release work.
**Carried over from Phase 1**: the notarized dry run (`darwin:sign:notarize`, staple, spctl/stapler, designated-requirement diff) and paste-matrix N15 plus the notarized columns; blocked on the Apple Developer agreement 403 until the user resolves it.
**Research flag**: Standard patterns apart from accessory-app Settings/Edit-menu quirks. The Phase 1 notarization dry run was deferred here, so notarization is still a discovery until it runs. If planning shows this phase is too large, release (RELS-01..04) can split into its own phase.

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Native Spike & Signed Dev Loop | 4/4 | Complete    | 2026-10-17 |
| 2. Providers, Search UI & First Run | 6/6 | Complete    | 2026-10-17 |
| 3. Action Pipeline & Library | 8/8 | In Progress | - |
| 4. Settings, Shell Polish & Release | 0/TBD | Not started | - |
