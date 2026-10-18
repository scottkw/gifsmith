---
gsd_state_version: "1.0"
milestone: v1.0
milestone_name: Gifsmith MVP
current_phase: 3
current_phase_name: Action Pipeline & Library
status: executing
stopped_at: Completed 03-05-PLAN.md
last_updated: "2026-10-18T01:18:14.006Z"
last_activity: 2026-10-17
last_activity_desc: Phase 3 execution started
state_head: 66c32bfd23d4328e9e1ca7945a44c95d760126bd
progress:
  total_phases: 4
  completed_phases: 2
  total_plans: 18
  completed_plans: 15
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-10-17)

**Core value:** Hotkey → type → Enter puts the right GIF on the clipboard (or into the frontmost app) in seconds, without Raycast.
**Current focus:** Phase 3 — Action Pipeline & Library

## Current Position

Phase: 3 (Action Pipeline & Library) — EXECUTING
Plan: 6 of 8
Status: Ready to execute
Last activity: 2026-10-17 — Phase 3 execution started

Progress: [█████░░░░░] 50% (2/4 phases; 10/10 planned plans complete)

## Performance Metrics

**Velocity:**
- Total plans completed: 10
- Average duration: - min
- Total execution time: 0.0 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 1 | 4 | - | - |
| 2 | 6 | - | - |

**Recent Trend:**
- Last 5 plans: -
- Trend: -

*Updated after each plan completion*
**Per-Plan Metrics:**

| Plan | Duration | Tasks | Files |
|------|----------|-------|-------|
| Phase 1 P01 | 22min | 3 tasks | 111 files |
| Phase 01 P02 | 4min | 2 tasks | 9 files |
| Phase 01 P03 | 75min | 2 tasks | 17 files |
| Phase 01 P04 | 3min | 2 tasks | 3 files |
| Phase 2 P01 | 13min | 3 tasks | 31 files |
| Phase 2 P02 | 6min | 3 tasks | 12 files |
| Phase 2 P03 | 15min | 2 tasks | 14 files |
| Phase 02 P04 | 11min | 3 tasks | 26 files |
| Phase 02 P05 | 12min | 2 tasks | 22 files |
| Phase 02 P06 | 23min | 3 tasks | 22 files |
| Phase 03 P01 | 14min | 3 tasks | 34 files |
| Phase 3 P02 | 11min | 2 tasks | 10 files |
| Phase 03 P03 | 10min | 2 tasks | 16 files |
| Phase 03 P04 | 15min | 2 tasks | 31 files |
| Phase 03 P05 | 12min | 2 tasks | 20 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [Roadmap]: Coarse granularity — research's 7 phases merged to 4 (spike gate; providers + search UI + first run; actions + library; settings + release)
- [Roadmap]: Phase 1 is a hard go/no-go gate and owns ACTN-01/ACTN-02 (native copy/paste proof); Phase 2 wires them to real search results
- [Roadmap]: Wails v3 `v3.0.0-beta.30` pinned exactly; pnpm from the first command; Developer ID signed dev builds from day one
- [Roadmap]: No persistent media cache in v1 (favorites store metadata snapshots only); providers never blended in one grid
- [Phase 1]: No pre-authorized fallback needed: NonActivating panel + Focus + makeFirstResponder(WKWebView) takes typing; Floating level + FullScreenAuxiliary overlays full-screen; HideOnFocusLost handles click-away
- [Phase 1]: Pasteboard calls run on the calling thread in an autorelease pool; only FocusWebView hops to the main run loop
- [Phase 01]: 01-02: tray built in shell.New (independent of hotkey); App.svelte unchanged because Wails serializes frontend events
- [Phase 01]: 01-02: CursorScreenWorkArea returns Wails Y-down coords using screens[0] height (same reference as windowSetPosition)
- [Phase 01]: 01-03: PasteDelay stays 80 ms; no Cmd+Enter KeyBindings fallback needed
- [Phase 01]: 01-03: TIS/HIToolbox calls must run on main (onMain); Wails bound methods run on goroutines
- [Phase 01]: 01-03: macOS 27 has no Accessibility row in Privacy & Security; onboarding opens the Privacy_Accessibility URL (button), never describes a Settings path
- [Phase 01]: 01-03: Task 1 (gifsmith-notary profile) deferred by user; required before 01-04
- [Phase 01]: 01-04: Phase 1 GO WITH FALLBACKS (user) - build Phase 2 on the Developer ID-signed dev build; limits in 01-PASTE-MATRIX.md section 4
- [Phase 01]: 01-04: Notarization deferred to Phase 4 (Apple agreement 403); notarized dry run + N15 carry over
- [Phase 01]: 01-04: Pasteboard payload stays a single file URL; Phase 3 decides on image data only if untested apps fail
- [Phase 2]: 02-01: malformed provider responses keep their HTTP status in provider.Error (consistent with httpx invalid-JSON)
- [Phase 2]: 02-01: staging downloads use their own http.Client bounded by Fetch's 45 s context (API client's 15 s timeout would cut multi-MB originals)
- [Phase 2]: 02-01: VendorAllow = https + vendor host + default port only; server build adds only the fixture listener host
- [Phase 2]: 02-02: Keys inside an open IME composition route to none for every key, not only Enter
- [Phase 2]: 02-02: Paging drops stale pages via resultSeq (bumped on every result-set replacement); a failed page waits for Try Again
- [Phase 2]: 02-02: One main mousedown handler pins focus to the search input for clicks outside text fields
- [Phase 2]: 02-02: node tests excluded from svelte-check instead of adding @types/node
- [Phase 2]: search.Service: 429 blocks a provider locally for max(Retry-After, 60 s); blocked calls return rate_limited with remaining seconds (0 = unknown) and no HTTP
- [Phase 2]: Error copy is computed from (error, now, receivedAt) in lib/errors.ts; caught error text is never rendered
- [Phase 2]: UI state is held per provider id in App.svelte; latest() slots first:<p> and page:<p> cancel superseded calls
- [Phase 02]: GIPHY rendition width/height/size not decoded until Phase 3; flexInt decodes total_count
- [Phase 02]: LastProvider reads settings.json per call; invalid or unknown id reads as first run
- [Phase 02]: SetLastProvider does not share the SaveKey mutex (temp+rename keeps writes whole)
- [Phase 02]: fakeapi scenario words read from q plus the key, so SaveKey/trending can be driven into 429/500/slow in UAT
- [Phase 02]: Setup card width 560 (not 456): at 456 it overran the 364px content area by 21px and K2-K6 overflowed the one-line message slot
- [Phase 02]: After a key save the query is cleared so the box matches the trending grid (S1c); Try Finer Gifs clears it too (S8)
- [Phase 02]: A key_rejected provider keeps its setup card across tab switches until a new key is saved (S13)
- [Phase 02]: Keychain-failure inline copy: Couldn't save the key to the macOS Keychain. Try again.
- [Phase 02]: 02-06: Clips is offered only after one Clips probe per launch passes, made after a GIPHY GIFs success with the same key; 401/403/404 -> disabled, transient -> unknown with one user-initiated recheck; Go refuses Clips calls unless enabled
- [Phase 02]: 02-06: C2 primary action is an <a> 'Email clips@giphy.com ↗' opening mailto:clips@giphy.com with a subject (the only mailto in the openExternal allow-list)
- [Phase 02]: 02-06: Native UAT approved; the user's real GIPHY key is not approved for Clips (probe checking -> disabled, C2 shown); dev logs hold no key material
- [Phase 2 close]: Klipy items with GIF renditions (ads included) stay in place (WR-04); persisted Clips tab restored when a GIPHY key exists (WR-02) — both accepted verification overrides
- [Phase 2 close]: GIPHY and Clips share one 429 budget; Clips probed at most once per key after a GIPHY GIFs success; a rate-limited re-check does not spend the re-probe; staging downloads check the allowlist on every redirect
- [Phase 2 close]: MVP mode dropped for Phases 2–4; 11 Info-level review items left open in 02-REVIEW-DISPOSITION.md
- [Phase 03]: ActionError has no Unwrap: staging/transport errors are logged, then mapped to a fixed-text kind; JS reads err.cause.kind
- [Phase 03]: staging.PageAllow is the single page-URL rule (GIPHY/Clips, https giphy.com), used by the GIPHY mapper and ActionService
- [Phase 03]: Display-only vendor numbers decode through httpx.LooseInt (never fail a page)
- [Phase 03]: The -tags server UAT build wires platform.Fake (Trusted false); desktop wires platform.New()
- [Phase 3]: 03-02: every action (menu row or shortcut) runs through App runAction; later plans add ROWS entries only
- [Phase 3]: 03-02: any mousedown outside the action menu closes it; rows highlight on mousemove so a menu opening under the pointer keeps row 1
- [Phase 03]: 03-03: Klipy grace delete holds busy around the gen check (no re-stage race); ServiceShutdown takes no lock
- [Phase 03]: 03-03: text copies and pastes clear staging in putText; Clips assets reuse rendition (LooseInt)
- [Phase 03]: Download claims names with os.Link from a hidden temp file; never overwrites, never creates the folder
- [Phase 03]: ActionService.DownloadsDir is an injected func (Phase 4 picker replaces it); results carry only file and folder names
- [Phase 03]: Toast model: setStatus(text, tone, action, ms); a toast with an action survives panel:show inside its expiry
- [Phase 03]: Square crop: a 1x1 placeholder for a frame outside the window gets disposal None (Background cleared the visible pixel (0,0)); golden test proves 0 mismatches
- [Phase 03]: Derived clipboard files (square) are staged under sha1(provider:id:variant) so Sweep(derived) removes the downloaded original
- [Phase 03]: Square actions: claimed Size > 15 MB refused before any request; DecodeConfig refuses canvases over 4096 before DecodeAll

### Pending Todos

None yet.

### Blockers/Concerns

- [Phase 1→3/4]: N10, N13 and paste in Slack, Discord, Telegram, Safari/Chrome composers untested; re-check in later UAT (GO WITH FALLBACKS limit 2)
- [Phase 2→]: The user's GIPHY key is not approved for Clips (C2 shown); approval via clips@giphy.com is a user action
- [Phase 2→]: The Finer Gifs Club API was down on 2026-10-17; Phase 2 UAT accepted either results or the down message, so availability is still unconfirmed
- [Phase 4]: Klipy written clearance for temp file copies (developers@klipy.com) is a pre-release gate; request early, and ask about labelling Klipy ads (IN-09)
- [Phase 2→4]: Public repo github.com/scottkw/gifsmith holds only the MIT LICENSE; push code after the Phase 2 merge to main and add the Raycast GIF Search authors' MIT notice to LICENSE
- [Phase 4]: Apple Developer agreement 403 (notarytool store-credentials: 'A required agreement is missing or has expired') must be resolved by the user before Phase 4 notarization; notarized dry run + N15 carried over from Phase 1

### Quick Tasks Completed

| # | Description | Date | Commit | Directory |
|---|-------------|------|--------|-----------|
| 261017-npc | Wire designed app and menu bar icons into the build | 2026-10-17 | dbe648e | [261017-npc-wire-designed-app-and-menu-bar-icons-int](./quick/261017-npc-wire-designed-app-and-menu-bar-icons-int/) |

## Deferred Items

Items acknowledged and deferred at milestone close, most recent first:

| Category | Item | Status | Deferred At | Milestone |
|----------|------|--------|-------------|-----------|
| *(none)* | | | | |

## Session Continuity

Last session: 2026-10-18T01:18:13.859Z
Stopped at: Completed 03-05-PLAN.md
Resume file: None
