---
phase: 02-providers-search-ui-first-run
verified: 2026-10-20T19:19:58Z
status: passed
score: 94/94 must-haves verified (5 roadmap SCs + 89 plan truths; 2 via accepted overrides; 6 hold as intentionally superseded by later-phase requirements; 0 failed, 0 behavior-unverified)
covered_files:
  - ".gitignore"
  - ".planning/phases/02-providers-search-ui-first-run/02-01-PLAN.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-01-SUMMARY.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-02-PLAN.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-02-SUMMARY.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-03-PLAN.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-03-SUMMARY.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-04-PLAN.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-04-SUMMARY.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-05-PLAN.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-05-SUMMARY.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-06-PLAN.md"
  - ".planning/phases/02-providers-search-ui-first-run/02-06-SUMMARY.md"
  - "frontend/bindings/gifsmith/internal/provider/index.ts"
  - "frontend/bindings/gifsmith/internal/provider/models.ts"
  - "frontend/bindings/gifsmith/internal/services/actionservice.ts"
  - "frontend/bindings/gifsmith/internal/services/gifservice.ts"
  - "frontend/bindings/gifsmith/internal/services/index.ts"
  - "frontend/bindings/gifsmith/internal/services/models.ts"
  - "frontend/bindings/gifsmith/internal/services/panelservice.ts"
  - "frontend/bindings/gifsmith/internal/services/settingsservice.ts"
  - "frontend/package.json"
  - "frontend/pnpm-lock.yaml"
  - "frontend/public/attribution/giphy-powered-dark.png"
  - "frontend/public/attribution/giphy-powered-light.png"
  - "frontend/src/App.svelte"
  - "frontend/src/app.css"
  - "frontend/src/components/Footer.svelte"
  - "frontend/src/components/ProviderTabs.svelte"
  - "frontend/src/components/ResultGrid.svelte"
  - "frontend/src/components/SetupCard.svelte"
  - "frontend/src/components/StateMessage.svelte"
  - "frontend/src/components/Tile.svelte"
  - "frontend/src/lib/actions.ts"
  - "frontend/src/lib/api.ts"
  - "frontend/src/lib/errors.test.ts"
  - "frontend/src/lib/errors.ts"
  - "frontend/src/lib/grid.test.ts"
  - "frontend/src/lib/grid.ts"
  - "frontend/src/lib/prefs.ts"
  - "frontend/src/lib/providers.ts"
  - "frontend/src/main.ts"
  - "frontend/tsconfig.json"
  - "go.mod"
  - "go.sum"
  - "internal/provider/fakeapi/fakeapi.go"
  - "internal/provider/finergifs/finergifs.go"
  - "internal/provider/finergifs/finergifs_test.go"
  - "internal/provider/finergifs/testdata/search.json"
  - "internal/provider/giphy/giphy.go"
  - "internal/provider/giphy/giphy_test.go"
  - "internal/provider/giphy/testdata/trending.json"
  - "internal/provider/giphy/testdata/unauthorized.json"
  - "internal/provider/httpx/httpx.go"
  - "internal/provider/httpx/httpx_test.go"
  - "internal/provider/klipy/klipy.go"
  - "internal/provider/klipy/klipy_test.go"
  - "internal/provider/klipy/testdata/invalid_key.json"
  - "internal/provider/klipy/testdata/search.json"
  - "internal/provider/provider.go"
  - "internal/provider/provider_test.go"
  - "internal/search/clips.go"
  - "internal/search/clips_test.go"
  - "internal/search/search.go"
  - "internal/search/search_test.go"
  - "internal/services/actions.go"
  - "internal/services/actions_test.go"
  - "internal/services/gifs.go"
  - "internal/services/panel.go"
  - "internal/services/settings.go"
  - "internal/services/settings_test.go"
  - "internal/settings/settings.go"
  - "internal/settings/settings_test.go"
  - "internal/shell/shell.go"
  - "internal/staging/staging.go"
  - "internal/staging/staging_test.go"
  - "main.go"
  - "scripts/smoke-browser-loop.sh"
  - "scripts/uat/01-tracer.js"
  - "scripts/uat/02-grid.js"
  - "scripts/uat/02-keyboard.js"
  - "scripts/uat/03-states.js"
  - "scripts/uat/05-setup.js"
  - "scripts/uat/05-tabs.js"
  - "scripts/uat/06-backstops.js"
  - "scripts/uat/06-clips-denied.js"
  - "scripts/uat/06-clips.js"
  - "wire_desktop.go"
  - "wire_server.go"
covered_digest: "v3:sha256:2e77c9020ef9bcea93e55a79c17bf5bf43e0285733781b62066a04425176fd71"
behavior_unverified: 0
overrides_applied: 2
overrides:
  - must_have: "Klipy results map to Gif{...}; only items with type 'gif' are mapped and any other items are counted in a log line"
    reason: "WR-04 fix 50824e1: Klipy's integration rules forbid filtering, so items with GIF renditions (including ads) stay in place; only items with no GIF rendition are skipped and logged. Ad labelling stays open as IN-09 pending developers@klipy.com."
    accepted_by: "Ken Scott"
    accepted_at: "2026-10-17T23:04:59Z"
  - must_have: "at launch the panel opens on the persisted provider, on GIPHY when none is persisted (first run), and on GIPHY when the persisted provider is Clips but Clips is unavailable"
    reason: "WR-02 fix e3501b7: a persisted Clips tab is restored when a GIPHY key exists and probes once behind the S2 skeleton (bootClips); without a GIPHY key it still opens on GIPHY. User observed 'Clips tab restored after relaunch' natively."
    accepted_by: "Ken Scott"
    accepted_at: "2026-10-17T23:04:59Z"
re_verification:
  previous_status: passed
  previous_score: "92/94 (2 uncertain, then accepted as overrides on 2026-10-17T23:04:59Z)"
  reason: "Flagged stale: Phase 3 and Phase 4 changed code this report covers (provider Params from Settings, App.svelte, SettingsService, staging, actions)."
  gaps_closed: []
  gaps_remaining: []
  regressions: []
  superseded_by_later_phases:
    - "⌘K bound with no effect -> opens the action menu (Phase 3, SRCH-05/ACTN)"
    - "Enter/⌘Enter on a Clip refused with E8 -> Clips copy/paste as MP4 (Phase 3, ACTN-09); TestCopyClipRefused replaced by TestCopyClipMP4/TestPasteClipMP4"
    - "Enter always copies the GIF file -> Enter runs the user's primary action (default copy; Phase 4 SETT default actions); on Klipy it falls back to Copy Link until Klipy clears temp-file copies (Phase 4, RELS-04, KlipyFileActionsCleared=false)"
    - "GIPHY rating fixed pg-13 / fixed page size -> per-request provider.Params from Settings, defaults pg-13 / en / us (Phase 4, SRCH-09, SETT-04); Clips probe keeps pg-13 fixed"
    - "Single staged copy file is the only media on disk -> plus user-initiated Downloads and square-crop staging with eager cleanup (Phase 3, ACTN-10)"
    - "Keys never reach the webview at all -> Settings › Keys shows the last 4 characters only (Phase 4, SETT-02; the full key still never reaches the webview)"
human_verification_resolution: "Carried forward from 2026-10-17T23:04:59Z (overrides, prohibitions P1-P11, MVP-mode decision). No new human items in this re-verification."
---

# Phase 2: Providers, Search UI & First Run Verification Report

**Phase Goal:** A stranger can open the panel, get set up with a key (or skip to the keyless provider), search or browse trending GIFs on any available provider in a fast keyboard-driven animated grid, and Enter copies the GIF, with vendor-required attribution and honest error handling.
**Verified:** 2026-10-20T19:19:58Z (code at HEAD 7874dc6 on `gsd/phase-04-settings-shell-polish-release`; no code diff from `main`)
**Status:** passed
**Re-verification:** Yes. **Re-verified 2026-10-20 after Phase 4.** The 2026-10-17 report (initial verification, code at 7c2c247) was flagged stale because Phases 3 and 4 changed files it covers. This pass re-checked every roadmap SC, plan truth group and requirement ID against today's code.

## Re-verified 2026-10-20 after Phase 4

What I did, and what I relied on:

- **Diffed** the Phase 2 code surface from `c9ffb49` (Phase 2 complete) to HEAD: 182 non-planning files changed. The Phase 2 core (`search.go`, `clips.go`, `httpx.go`, `gifs.go`, the three provider clients, `SaveKey`, `providers.ts`, `Footer.svelte`, `SetupCard.svelte`, `ProviderTabs.svelte`, the official GIPHY PNGs) was read at HEAD.
- **Ran myself:** `go test -race -count=1 ./internal/...` (14 packages ok, fakeapi has no tests; exit 0), `go vet . ./internal/...` (clean), `node --test src/lib/*.test.ts` (103 pass, 0 fail), `svelte-check` (191 files, 0 errors, 0 warnings), a debt-marker grep (none) and sha256 of the attribution PNGs (unchanged since Phase 2; no diff in `frontend/public/attribution`).
- **Not re-run, by instruction:** the browser smoke and headless UAT loop (ports are shared with concurrent verifiers). The most recent verifier-run evidence for those scripts is the Phase 4 verification (`04-VERIFICATION.md:263`, HEAD 7e8792a): SMOKE OK and 35/35 UAT OK, which includes all nine Phase 2 scripts (01-tracer, 02-grid, 02-keyboard, 03-states, 05-setup, 05-tabs, 06-backstops, 06-clips, 06-clips-denied). `git diff 7e8792a HEAD -- . ':!.planning'` is empty, so that run covers today's code exactly.
- **Known accepted debt (not a gap):** the 18-backstops `CancelledRejectionError` flake recorded in `.planning/phases/04-settings-shell-polish-release/deferred-items.md` (status acknowledged). Its owner is the Phase 2 search/cancel path in `App.svelte`, and it has no user-visible effect.

### What Phase 3 and Phase 4 changed in Phase 2's surface

| Phase 2 behavior | Today | Changed by | Verdict |
|------------------|-------|------------|---------|
| GIPHY `rating=pg-13`, fixed page size | `provider.Params` read per request from SettingsService (`main.go:71-78`), default `pg-13`/`en`/`us` (`provider.go:77-79`); `lang` sent on search only; Clips probe keeps `pg-13` (`giphy.go:118`). Param change calls `Search.Flush` (drops cache, keeps the 429 block) | Phase 4 SRCH-09, SETT-04 | Superseded intentionally; Phase 2 defaults preserved |
| Klipy `content_filter=medium`, key in path, 8..50 per page | Unchanged; `per_page` now `min(max(PageSize,8),50)`, `locale` from Params (`klipy.go:117-120`). No-filter/no-reorder mapping and WR-04 logged skip unchanged | Phase 4 SETT-04 | Holds |
| ⌘K bound, no effect | Opens the action menu (`App.svelte:652-653,715`) | Phase 3 | Superseded (SC2 already scoped the menu to Phase 3) |
| Clip copy refused with E8 | Clips copy/paste the 720p MP4 (`giphy.go:151-158` clipRendition, `TestCopyClipMP4`) | Phase 3 ACTN-09 | Superseded (was listed as deferred in the original report) |
| Enter copies the GIF file | Enter runs `effectiveAction(..., 'primary')` (`actions.ts:43-45`); default `copy` (`settings.go:212`). On Klipy it returns Copy Link because `KlipyFileActionsCleared = false` (`actions.go:45`), enforced again in Go (`actions.go:104`) | Phase 4 RELS-04, SETT default actions | Superseded intentionally for Klipy pending written clearance; GIPHY and Finer Gifs still copy the file by default |
| Keys never reach the webview | `KeyStatus` returns status plus `Last4` only (`services/settings.go:108-112`) | Phase 4 SETT-02 (04-CONTEXT.md:19) | Holds in substance; the full key never reaches the webview or logs |
| Hide-after-copy timer survives a re-show (IN-07) | `clearTimeout(hideTimer)` on `panel:show` (`App.svelte:277`) | Phase 3/4 fix | Info item closed |

## Mode discrepancy (MVP)

Resolved 2026-10-17: `mvp` mode was dropped from Phases 2-4 in ROADMAP.md. Standard goal-backward verification applies. The User Flow Coverage table is kept because the goal is phrased as a flow.

## User Flow Coverage

Flow, derived from the goal: «A stranger opens the panel → sets up a key or skips to the keyless provider → searches or browses trending on any available provider in a fast keyboard grid → Enter copies the GIF, with attribution and honest errors.»

| Step | Expected | Evidence (HEAD) | Status |
|------|----------|-----------------|--------|
| Open the panel | Global hotkey shows a non-activating panel | `shell.go:59-71` (NonActivating, `HideOnEscape:false`, WindowClosing hook); default `Cmd+Ctrl+Option+G` (`settings.go:224`), now rebindable (Phase 4 SETT-03). Native approval from 2026-10-17 plus Phase 4 UAT. | ✓ |
| First run: setup card | Opens on GIPHY with the S1 card, key field, Save Key, "Get a free key ↗", "Try Finer Gifs (no key)" | `App.svelte:310-316` (persisted slot, else GIPHY; Clips without a GIPHY key → GIPHY), `card` derivation `App.svelte:194-198`, `SetupCard.svelte:115,134,136` | ✓ |
| Paste key → verified → saved | Real trending call, then Keychain; bad key → K2 | `services/settings.go:71-105` (trim, `validKey`, `NewClient().Trending`, `Keys.Set`, `Invalidate`+`Seed`, Clips Reset/Trigger, `providers:changed`). `TestSaveKeyVerifiesAndSeeds`, `TestSaveKeyRejectedKeepsOldKey`, `TestSaveKeyConcurrent`, `TestSaveKeyNeverLogsOrReturnsKey` pass | ✓ |
| Or skip to the keyless provider | Try Finer switches to Finer Gifs | `tryFiner` `App.svelte:520-523` | ✓ |
| Browse trending / search | Empty box → trending; typing → 450 ms debounce; virtualized animated grid | `DEBOUNCE_MS = 450` (`App.svelte:67`), `wanted()` 2+ code points (`App.svelte:397-400`), trending at once on short query (`:415`), `ResultGrid.svelte` | ✓ |
| Any available provider | GIPHY, Clips (when approved), Klipy, Finer via ⌘1–⌘4 or pills; per-provider state | `providers.ts` SLOTS, `onMainKeydown` slot routing (`App.svelte:703-708`), `switchTo` persists via `setLastProvider` (`:476-488`), Clips pill `aria-disabled` (`ProviderTabs.svelte:23-29`) | ✓ |
| Keyboard-driven | Arrows, Enter, ⌘Enter, Esc clears then hides, ⌘K, Tab bound | `grid.ts` `routeKey` (node tests pass), `App.svelte` escape case (`:655-660`), `'tab'` bound no effect (`:666`) | ✓ |
| Enter copies the GIF | Download through the allowlist, stage, file URL on the pasteboard, hide 400 ms later | `effectiveAction` default `copy`; `HIDE_AFTER_COPY_MS` (`App.svelte:902-905`); `TestCopyTwiceKeepsOnlyLatest` passes. Klipy: Copy Link by default (RELS-04, see above) | ✓ (Klipy superseded by RELS-04) |
| Attribution | Official GIPHY mark on GIPHY/Clips, "Search KLIPY" + "Powered by KLIPY", Finer credit | `Footer.svelte:42-49` (mark now follows the resolved theme), `providers.ts` placeholders, PNG hashes unchanged | ✓ |
| Honest errors | Per-provider copy, no URL or key, one failing provider never blocks others | `errors.ts` (node tests pass), per-provider `states`, `MarshalError` (`main.go:131-135`) | ✓ |
| **Outcome** | A stranger goes from nothing to a copied GIF without help | Every step holds at HEAD | ✓ |

## Goal Achievement

### Observable Truths: Roadmap Success Criteria (contract)

| # | Truth | Status | Evidence (HEAD) |
|---|-------|--------|-----------------|
| SC1 | Trending with an empty box; debounced search; grid loads more on scroll or arrows; switch Finer/GIPHY/Klipy, last provider remembered; never blended, reordered or filtered | ✓ VERIFIED | `search.go` TTL cache and `dedupe` skips Klipy (`search.go:184-189`); `TestNoDedupeKlipy`, `TestDedupeNonKlipy` pass. Debounce `App.svelte:67,415-417`. `setLastProvider` in `switchTo`. Klipy WR-04 skip unchanged (override 1). Library tabs (Phase 3) group by provider, never blended (03-VERIFICATION SC4). |
| SC2 | Arrows move, Enter copies a file that pastes with ⌘V, Esc closes; ⌘K bound | ✓ VERIFIED | `grid.ts` + node tests; escape case `App.svelte:655-660`; Enter → primary action default `copy`. ⌘K now opens the Phase 3 menu (superset). Klipy Enter = Copy Link under RELS-04 (intentional). |
| SC3 | First run on GIPHY with the card; key verified before save; bad key rejected; Try Finer works keyless; keys in Keychain, never in UI/logs/errors | ✓ VERIFIED | `wire_desktop.go` KeyringStore; `SaveKey` returns error only; log lines in `services/settings.go` name provider and kind only (lines 81-95, 163-196); `TestSaveKeyNeverLogsOrReturnsKey` passes. Settings shows last 4 only (SETT-02). |
| SC4 | "Powered by GIPHY" on GIPHY and Clips; Klipy placeholder "Search KLIPY" | ✓ VERIFIED | `Footer.svelte:42-44`, `attribution` derived per slot (`App.svelte:188-190`), `providers.ts` Klipy placeholder `'Search KLIPY'`. |
| SC5 | Clips only after its test call passes, else C2; debounce saves requests; rate limit gives retry time; per-provider errors, others stay usable | ✓ VERIFIED | `clips.go` state machine (tests pass: `TestClipsRecheckBeforeGifsSuccess`, `TestClipsRateLimitedRecheckKeepsReprobe`); `gifs.go` refuses Clips unless enabled; `ProbeClips` 401/403 → key_rejected (`giphy.go:112-121`). 429: `TestBlockKnownRetry`, `TestClipsShareGiphyBudget`; Phase 4 split `gen`/`keyGen` so a Params flush keeps the 429 block (`search.go:82-110`). `TestCancelNotCachedNotBlocked`, `TestInvalidateDropsInFlightResult` pass. |

### Observable Truths: Plan must-haves (89)

| Plan | Truths | Verified | Notes at HEAD |
|------|--------|----------|---------------|
| 02-01 tracer | 15 | 15 | `MarshalError` still first in the chain (`main.go:131-135`). httpx classification and redaction tests pass; `LooseInt` added for display-only numbers. Staging allowlist/magic/cap/.part+rename and Sweep: `TestCopyTwiceKeepsOnlyLatest` passes. Service count grew to 6 (`main.go:124-129`, Library and Update added in Phases 3-4), all untagged. |
| 02-02 grid/keyboard | 16 | 16 | Virtualized `ResultGrid.svelte`; shell `HideOnEscape:false`, WindowClosing cancel+Hide, App/Edit menu (`shell.go:59-102`). ⌘K "no effect" superseded by the Phase 3 menu. |
| 02-03 states/cache | 17 | 17 | TTLs and `max(Retry-After, 60 s)` block intact; Phase 4 `Flush` added without loosening the block. `latest()`/cancel path unchanged in intent; the acknowledged 18-backstops `CancelledRejectionError` flake lives here (accepted debt). |
| 02-04 GIPHY/Klipy/keys | 16 | 16 (1 via override) | GIPHY q clamp 50 runes, `messaging_non_clips`, flexInt, offset caps 4999/499, 401/403 → key_rejected. Rating now `Params.Rating` default pg-13 (SRCH-09). Klipy path key + `PathEscape` + `badKey`, `content_filter=medium`, 401/403/404 → key_rejected. SaveKey serialized. Override 1 (WR-04) still matches `klipy.go:139-145`. |
| 02-05 tabs/setup | 14 | 14 (1 via override) | Placeholders, per-slot attribution, aria-disabled Clips with title, setup card, Try Finer, `providers:changed` refresh (`App.svelte:294`). Override 2 (WR-02) still matches `App.svelte:312-315`. |
| 02-06 Clips/native | 11 | 11 | Probe after a GIFs success, once per key, inside the GIPHY 429 budget. E8-on-Clip truth superseded by ACTN-09 (MP4 copy). Compliance row in PROJECT.md:79 still accurate. Backstop truths (long copy fits, ≥4.5:1 contrast) covered by `06-backstops.js` in the Phase 4 verifier's 35/35 run on identical code. |

**Score:** 94/94 truths verified (0 present-but-behavior-unverified, 0 uncertain, 0 failed; 2 via accepted overrides; 6 intentionally superseded by later-phase requirements, listed in frontmatter)

Behavior-dependent truths each have a named test that ran in my `go test -race` run: `TestCancelNotCachedNotBlocked`, `TestBlockKnownRetry`, `TestClipsShareGiphyBudget`, `TestClipsRecheckBeforeGifsSuccess`, `TestClipsRateLimitedRecheckKeepsReprobe`, `TestSaveKeyConcurrent`, `TestInvalidateDropsInFlightResult`.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/provider/provider.go` | IDs, Gif, Page, Error, MarshalError (+ Params since Phase 4) | ✓ VERIFIED | Wired in `main.go:131-135`. |
| `internal/provider/httpx/httpx.go` | GetJSON, Remap | ✓ VERIFIED | Used by all three clients. |
| `internal/provider/{finergifs,giphy,klipy}/*.go` | clients + mappers | ✓ VERIFIED | Registered in `search.New` (`main.go:74-79`) with `Params`. |
| `internal/provider/fakeapi/fakeapi.go` | fixture server | ✓ VERIFIED | Server build only. |
| `internal/staging/staging.go` | Fetch/Sweep/VendorAllow | ✓ VERIFIED | Wired through ActionService. |
| `internal/search/search.go`, `clips.go` | cache, 429 budget, dedupe, Clips probe | ✓ VERIFIED | Tests pass. |
| `internal/services/{gifs,settings,actions,panel}.go` | bound services | ✓ VERIFIED | `application.NewService` in `main.go:124-129`. |
| `internal/settings/settings.go` | KeyringStore, MemoryStore, Keys, Prefs | ✓ VERIFIED | Tests pass. |
| `internal/shell/shell.go` | panel hooks, App/Edit menu | ✓ VERIFIED | Lines 59-102. |
| `frontend/src/components/*.svelte` (Phase 2's six) | grid, tile, tabs, footer, setup card, state message | ✓ VERIFIED | All rendered by App.svelte. |
| `frontend/src/lib/{api,grid,errors,providers}.ts` | facade, key routing, copy rows, slots | ✓ VERIFIED | 103 node tests pass (suite grew from 23). |
| `frontend/public/attribution/giphy-powered-{light,dark}.png` | official marks | ✓ VERIFIED | sha256 `8e8619ee…`/`73508f69…`, unchanged since the 2026-10-17 official-zip match. |
| `scripts/smoke-browser-loop.sh`, `scripts/uat/*.js` (Phase 2's 9) | harness | ✓ VERIFIED | Not re-run here (instruction); Phase 4 verifier run 35/35 on identical code. |

### Key Link Verification

| From | To | Via | Status |
|------|----|-----|--------|
| App.svelte | GifService.Search/Trending | `api.ts` `search`/`trending` in `load` (`App.svelte:528`) | ✓ WIRED |
| gifs.go | search.Service | `s.Svc.Search/Trending` | ✓ WIRED |
| providers | httpx | `httpx.GetJSON` in all three clients | ✓ WIRED |
| main.go | provider.MarshalError | `MarshalError` chain | ✓ WIRED |
| main.go | SettingsService.Params | `params` closure into every client (`main.go:71-95`) | ✓ WIRED (new in Phase 4) |
| actions.go | staging | `Fetch` → CopyFile → Sweep | ✓ WIRED |
| App.svelte | PanelService.Hide | `hideTimer` after copy; Esc on empty query | ✓ WIRED |
| App.svelte | grid.ts | `routeKey` | ✓ WIRED |
| SettingsService.SaveKey | candidate client → Keys.Set → Invalidate/Seed → Clips Reset/Trigger | `services/settings.go:71-105` | ✓ WIRED |
| SetupCard | SettingsService.SaveKey | `saveKey(p, key)` (`SetupCard.svelte:52`) | ✓ WIRED |
| App.svelte | Providers / LastProvider / SetLastProvider / `providers:changed` | onMount and `switchTo` | ✓ WIRED |
| App.svelte | GifService.CheckClips | `recheckClips` via `bootClips` | ✓ WIRED |

### Data-Flow Trace (Level 4)

| Artifact | Data | Source | Real data | Status |
|----------|------|--------|-----------|--------|
| ResultGrid/Tile | `cur.items` | GifService → search.Service → vendor HTTP (desktop) / fakeapi (server) | Yes | ✓ FLOWING |
| Footer | attribution, status | slot or selected item's provider; action results | Yes | ✓ FLOWING |
| ProviderTabs/card | `statuses` | `GifService.Providers()` → Keys + Clips state | Yes | ✓ FLOWING |
| StateMessage | `errView` | `errorState(cur.err, …)` | Yes | ✓ FLOWING |

### Behavioral Spot-Checks (run by this verifier, 2026-10-20)

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Go tests | `go test -race -count=1 ./internal/...` (once) | 14 packages ok, exit 0 (linker version warnings only) | ✓ PASS |
| Go vet | `go vet . ./internal/...` | clean | ✓ PASS |
| Frontend unit | `node --test src/lib/*.test.ts` | 103 pass, 0 fail | ✓ PASS |
| Frontend types | `svelte-check --tsconfig ./tsconfig.json` | 191 files, 0 errors, 0 warnings | ✓ PASS |
| Phase 2 named tests exist | grep for the 17 tests cited in the original report | 16 present; `TestCopyClipRefused` replaced by `TestCopyClipMP4`/`TestPasteClipMP4` (ACTN-09) | ✓ PASS |
| Attribution PNGs | sha256 + `git diff c9ffb49 HEAD -- frontend/public/attribution` | unchanged | ✓ PASS |
| Smoke + headless UAT | not run (shared ports, by instruction) | Phase 4 verifier: SMOKE OK, 35/35 UAT OK at 7e8792a; no code diff to HEAD | ? SKIP (covered by prior verifier run) |

### Probe Execution

Step 7c: no `scripts/*/tests/probe-*.sh` exist and none are declared.

### Requirements Coverage

| Requirement | Description | Status | Evidence at HEAD |
|-------------|-------------|--------|------------------|
| SRCH-01 | Search GIPHY, Klipy, Finer; switch; last provider remembered | ✓ SATISFIED | SC1. |
| SRCH-02 | Clips only after the Clips test call, else an explanation | ✓ SATISFIED | SC5. |
| SRCH-03 | Trending with an empty query | ✓ SATISFIED | `wanted()==''` → trending; Finer S8. |
| SRCH-04 | Animated grid, loads more on scroll or arrows | ✓ SATISFIED | ResultGrid, `nearEnd`. |
| SRCH-05 | Keyboard navigation (arrows, Enter, ⌘K, Esc) | ✓ SATISFIED | ⌘K now opens the Phase 3 menu. |
| SRCH-06 | Debounced; rate limit shows retry time | ✓ SATISFIED | 450 ms; E1 countdown; 429 block tests. |
| SRCH-07 | Per-provider errors; one failing never blocks others | ✓ SATISFIED | per-slot `states`; errors.ts tests. |
| ONBD-01 | First run on GIPHY with card + Try Finer | ✓ SATISFIED | `App.svelte:310-316`, SetupCard. |
| ONBD-02 | Key verified with a test call before saving | ✓ SATISFIED | `SaveKey`. |
| ONBD-03 | Keychain; never in web UI, logs or errors | ✓ SATISFIED | Full key never leaves Go; Settings shows last 4 by design (SETT-02). |
| ONBD-04 | "Powered by GIPHY"; "Search KLIPY" | ✓ SATISFIED | SC4. |
| ONBD-05 | Never blended, reordered or filtered (except non-Klipy same-page dupes) | ✓ SATISFIED (override 1) | `dedupe` skips Klipy; library tabs grouped by provider. |

All 12 Phase 2 IDs in REQUIREMENTS.md (lines 19-35, 115-128) are covered; none orphaned. SRCH-08 and SRCH-09 map to Phases 3 and 4.

### Prohibitions (judgment-tier, accepted 2026-10-17; re-checked at HEAD)

| # | Prohibition | Verdict at HEAD | Note |
|---|-------------|-----------------|------|
| P1 | No synthetic desktop input outside human checkpoints | No violation found | I sent none; UAT not run. |
| P2 | Kill only processes you started | No violation found | I started and stopped none. |
| P3 | No API key in repo/logs/fixtures/tests | No violation found | Log lines name provider and kind only. |
| P4 | No TCC changes | No violation found | |
| P5 | No media on disk beyond the single staged copy | Superseded | Phase 3 added user-initiated Downloads and crop staging with eager cleanup (ACTN-10). |
| P6 | No Raycast proxy, analytics, pingbacks | No violation found | Phase 4 updater talks to GitHub Releases only (RELS-05), not a provider. |
| P7 | No unapproved npm packages | Out of Phase 2 scope now | Phase 3/4 packages were their phases' decisions. |
| P8 | No provider requests the user did not cause | No material violation (3 bounded exceptions as accepted) | Params changes flush the cache but do not fetch. |
| P9 | Response cache in memory, metadata only | No violation found | `search.go` holds `provider.Page` only. |
| P10 | Tests never touch the login Keychain | No violation found | Tests passed with MemoryStore / `keyring.MockInit()`. |
| P11 | UAT never clicks OpenURL controls | No violation found | |
| P12 | Official, unmodified GIPHY PNGs | Verified deterministically | Hashes unchanged. |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| Phase 2 code files at HEAD | n/a | TBD, FIXME, XXX, TODO, HACK | None found | Debt-marker gate passes. |
| `frontend/src/components/SetupCard.svelte` | 95 | Step 3 hardcodes "Reopen Gifsmith with ⌃⌥⌘G", but Phase 4 (SETT-03) made the open hotkey rebindable and unbindable | ℹ️ Info | Stale instruction only for users who rebound the hotkey before adding a key. Not a Phase 2 must-have; new since Phase 4, not tracked elsewhere. |
| `frontend/src/App.svelte` | search/cancel path | `CancelledRejectionError` unhandled on tab switch (18-backstops flake) | ℹ️ Info (accepted debt) | Recorded in Phase 4 deferred-items.md, acknowledged. |
| `internal/provider/klipy/klipy.go` | 139-145 | unlabelled ads (IN-09) | ℹ️ Info | Open policy question with Klipy; RELS-04 gate also pending. |
| `internal/provider/httpx/httpx.go` | 72 | not capped (IN-02) | ℹ️ Info | Unchanged from the original report. |

The IN-07 hide-timer item from the original report is fixed (`App.svelte:277`).

### Deferred / later-phase items

All four deferred items from the original report were delivered: Clips MP4 copy (Phase 3 ACTN-09), temp-file lifecycle (Phase 3 ACTN-10), ⌘K action menu (Phase 3), content rating control (Phase 4 SRCH-09). Klipy written clearance remains open and is now enforced as a gate (RELS-04).

### Human Verification Required

None new. The four items from 2026-10-17 (two overrides, prohibition review, MVP-mode decision) were resolved by the user on 2026-10-17T23:04:59Z and still apply: both overrides match the current code.

### Gaps Summary

There are no gaps. Every roadmap success criterion, plan truth group and Phase 2 requirement ID still holds at HEAD. Six Phase 2 details were intentionally changed by later-phase requirements (configurable GIPHY params, the ⌘K menu, Clips MP4 copy, configurable primary action with the Klipy file-action gate, extra user-initiated files on disk, a last-4 key display in Settings). Each traces to a recorded Phase 3 or 4 requirement or decision, and the Phase 2 goal survives each of them. I ran the Go race tests, go vet, the node tests and svelte-check myself. I did not re-run the browser smoke and UAT loop because of the shared ports. The Phase 4 verifier's 35/35 run covers it, on code identical to HEAD. One new Info item: the hardcoded hotkey in the setup card copy (`SetupCard.svelte:95`).

---

_Verified: 2026-10-20T19:19:58Z (re-verification after Phase 4; original 2026-10-17T22:41:41Z)_
_Verifier: Claude (gsd-verifier)_
