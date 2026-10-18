---
phase: 04-settings-shell-polish-release
plan: 02
subsystem: settings
tags: [wails-v3, svelte5, giphy, klipy, keychain, go-keyring, search-params, cache, intl]

requires:
  - phase: 04-settings-shell-polish-release
    provides: "04-01 settings.Store, SetPref + settings:changed, Settings window, PANES, SettingRow, SettingsError"
  - phase: 02-providers-search-ui-first-run
    provides: "SaveKey verify-then-Keychain path, search.Service cache/429 budget, fakeapi fixture, smoke/UAT loop"
provides:
  - "provider.Params / DefaultParams / DefaultPageSize / PageSizes / Ratings / ParamsOf; provider.PageSize removed"
  - "Params func on giphy, klipy and finergifs clients, read per request (limit/per_page/size, rating, lang, locale)"
  - "search.Service.Flush (drops a budget's cached pages, keeps the 429 block); separate keyGen so a 429 after a Flush still blocks"
  - "giphy.Languages (32 GIPHY codes) and klipy.Countries (257 NSLocale codes)"
  - "platform.SystemLocale (gsSystemLocale cgo, stub, Fake Lang/Region)"
  - "Prefs giphyRating, pageSize, giphyLang, klipyCountry with Normalize; SetPref for all four with per-provider Flush"
  - "SettingsService.Params (Match System resolved per request), bound Choices, KeyStatus, TestKey, RemoveKey"
  - "SecretStore.Delete, Keys.Delete, Keys.MarkRejected/ClearRejected/Rejected; SettingsError keychain_delete_failed"
  - "GifService marks a key rejected on key_rejected from search (Clips -> GIPHY)"
  - "Settings Search pane (GIFs per page, Content rating, Language, Country) and Keys pane (KeyRow cards)"
  - "lib/prefs.ts (languageOptions, countryOptions, ratingLabel); errors.ts testKeyMessage, keySavedText, keyRemovedText, keyRemoveFailedText"
  - "panel: provider tabs marked stale on a provider-parameter change refetch their first page when shown"
  - "scripts/uat/16-settings-search.js, scripts/uat/16-settings-keys.js; smoke KeyStatus/TestKey/RemoveKey"
affects: [04-03, 04-04, 04-05, 04-06]

actuals:
  tokens: 21700    # chars/4 over the added lines of the realized diff (86804 chars, 2136 lines added, 46 files)
  tasks: 8
  commits: 15
plan_head_before: 9a80f80d5f1f156f2ad6ad6935a9c6666647c296
plan_head_after: cc5b99f5e53f09fc13a158085cd9d1b228c75131

tech-stack:
  added: []
  patterns:
    - "Provider parameters are a per-request func (Params func() provider.Params) on each client; nil means DefaultParams (tests)"
    - "Cache generation (gen) guards caching and is bumped by Flush and Invalidate; key generation (keyGen) guards 429 recording and is bumped only by Invalidate"
    - "A pref equal to its default is stored as empty (\"\" / 0), so choosing the default when nothing is stored is a no-op"
    - "Match System is stored as \"\" and resolved at request time from platform.SystemLocale, with en/us fallbacks"
    - "The panel keeps the last snapshot's provider params; a change marks provider tabs stale, refetched by show() on panel:show or a tab switch"
    - "Key UI only ever receives KeyStatus {provider, status, last4}"

key-files:
  created:
    - internal/provider/giphy/languages.go
    - internal/provider/klipy/countries.go
    - frontend/src/lib/prefs.ts
    - frontend/src/lib/prefs.test.ts
    - frontend/src/settings/SearchPane.svelte
    - frontend/src/settings/KeysPane.svelte
    - frontend/src/settings/KeyRow.svelte
    - scripts/uat/16-settings-search.js
    - scripts/uat/16-settings-keys.js
  modified:
    - internal/provider/provider.go
    - internal/provider/giphy/giphy.go
    - internal/provider/klipy/klipy.go
    - internal/provider/finergifs/finergifs.go
    - internal/provider/fakeapi/fakeapi.go
    - internal/search/search.go
    - internal/settings/settings.go
    - internal/services/settings.go
    - internal/services/gifs.go
    - internal/services/errors.go
    - internal/platform/platform.go
    - internal/platform/platform_darwin.go
    - internal/platform/platform_darwin.h
    - internal/platform/platform_darwin.m
    - internal/platform/platform_stub.go
    - internal/platform/fake.go
    - main.go
    - frontend/src/App.svelte
    - frontend/src/main.ts
    - frontend/src/lib/api.ts
    - frontend/src/lib/errors.ts
    - frontend/src/settings/SettingsApp.svelte
    - scripts/smoke-browser-loop.sh
    - scripts/uat/15-settings-general.js

key-decisions:
  - "settings imports the giphy/klipy lists directly for Normalize: go list shows no import cycle"
  - "search.Service gets a second counter (keyGen): Flush bumps only the cache generation, so a 429 that lands after a parameter change still blocks the budget"
  - "GIPHY search always sends lang (default en); Klipy always sends locale (default us); trending never sends lang"
  - "Default rating and page size are stored as empty, so picking PG-13 or 24 when nothing is stored writes nothing"
  - "Keys.ClearRejected added: a passing Test clears Rejected without rewriting the Keychain"
  - "The Keys card message line wraps instead of the UI-SPEC one-line ellipsis: K2, K3 and K-T2 are wider than 400px at 12px"
  - "The panel gets the initial Snapshot as a prop from main.ts, so it can tell when provider params change"

patterns-established:
  - "Settings panes that need no snapshot still declare the initial prop (PANES passes it to every pane)"
  - "Search-pane selects set selected per option: the options arrive from Choices() after mount"

requirements-completed: [SRCH-09, SETT-02, SETT-04]

coverage:
  - id: D1
    description: "Content rating G/PG/PG-13/R in the Search pane is stored, flushes GIPHY/Clips cached pages (not the 429 block) and is sent as rating on the next GIPHY and Clips request"
    requirement: SRCH-09
    verification:
      - kind: unit
        ref: "internal/services/settings_test.go#TestSetPrefGiphyRating, TestSetPrefGiphyRatingInvalid, TestParamsDefaultRating; internal/provider/giphy/giphy_test.go#TestParamsRating; internal/search/search_test.go#TestFlushKeepsRateLimitBlock, TestFlushKeepsOtherProviders, TestFlushInFlight429StillBlocks"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/16-settings-search.js"
        status: pass
    human_judgment: false
  - id: D2
    description: "GIFs per page 24/36/48: GIPHY limit min(n,50), Klipy per_page 8..50, Finer size n; more-results from the requested size; other values refused"
    requirement: SETT-04
    verification:
      - kind: unit
        ref: "internal/provider/...#TestDefaultParams, TestPageSizeLimit, TestPageSizeMore, TestPageSizeClamp; internal/services/settings_test.go#TestSetPrefPageSize"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/16-settings-search.js (36 -> 9-row first page after a tab switch)"
        status: pass
    human_judgment: false
  - id: D3
    description: "GIPHY language (search only) and Klipy country (search and trending) with Match System resolved from the macOS locale, English/United States fallbacks, lists sorted by English name"
    requirement: SETT-04
    verification:
      - kind: unit
        ref: "internal/provider/giphy/giphy_test.go#TestLangSearchOnly, TestLanguages; internal/provider/klipy/klipy_test.go#TestLocaleCountry, TestCountries; internal/services/settings_test.go#TestSetPrefLangCountry, TestParamsSystemLocale; internal/settings/settings_test.go#TestStoreNormalizeSearchPrefs; frontend/src/lib/prefs.test.ts"
        status: pass
      - kind: automated_ui
        ref: "scripts/uat/16-settings-search.js (Match System (English), German -> [lang=de], Match System (United States))"
        status: pass
    human_judgment: true
    rationale: "Match System on a real non-English macOS locale (gsSystemLocale) and real GIPHY/Klipy localisation only exist on the signed app with real keys"
  - id: D4
    description: "Panel tabs fetched with old provider params are marked stale and refetch their first page on panel:show or tab switch; library tabs untouched"
    requirement: SRCH-09
    verification:
      - kind: automated_ui
        ref: "scripts/uat/16-settings-search.js (Meta+4, Meta+1 refetch shows [rating=g] and the 36-item page)"
        status: pass
    human_judgment: false
  - id: D5
    description: "Keys in Go: KeyStatus (last 4 only), TestKey (uncached, inside the 429 budget, seeds, marks rejected), RemoveKey (Keychain, cache, Clips reset, event), Keychain failure keeps the key, search key_rejected marks the key"
    requirement: SETT-02
    verification:
      - kind: unit
        ref: "internal/services/settings_test.go#TestKeyStatus, TestTestKeyUncached, TestTestKeyRejected, TestTestKeyInsideBudget, TestTestKeyRefusals, TestRemoveKey, TestRemoveKeyKeychainFailure, TestSearchKeyRejectedMarks; internal/settings/settings_test.go#TestMemoryStoreDelete, TestKeyringStoreDelete, TestKeysDeleteAndRejected"
        status: pass
      - kind: integration
        ref: "bash scripts/smoke-browser-loop.sh (KeyStatus last4 -123, TestKey 200, RemoveKey 200, no key in settings.json or log)"
        status: pass
    human_judgment: false
  - id: D6
    description: "Keys pane: GIPHY then Klipy cards with Saved/Not set/Rejected/Testing status, Add/Change editor, Test, inline remove confirmation, UI-SPEC copy, Get a free key links"
    requirement: SETT-02
    verification:
      - kind: unit
        ref: "frontend/src/lib/errors.test.ts#testKeyMessage K-T1..K-T4, Keys pane texts"
        status: pass
      - kind: automated_ui
        ref: "bash scripts/smoke-browser-loop.sh --uat scripts/uat/16-settings-keys.js"
        status: pass
    human_judgment: true
    rationale: "Cmd+V into the key field (Edit menu) and the real Keychain delete surviving Quit and relaunch only exist on the signed install"

duration: 28min
completed: 2026-10-18
status: complete
---

# Phase 4 Plan 02: Search Parameters and Keys Pane Summary

**Content rating, GIFs per page, GIPHY language and Klipy country are now per-request settings in Go. Changing one flushes only the affected providers' cache and never lifts a 429 block. The Settings window gains a Search pane and a Keys pane that can add, test, change and remove keys, and the webview only ever sees the last 4 characters of a key.**

## Performance

- **Duration:** 28 min
- **Started:** 2026-10-18T21:15:44Z
- **Completed:** 2026-10-18T21:43:41Z
- **Tasks:** 8 (tracer + 7)
- **Files modified:** 46, including 5 regenerated binding files

## Accomplishments

- **Params on every request.** `provider.Params` replaces the old `PageSize` constant and the hard-coded `"pg-13"`. The GIPHY, Clips, Klipy and Finer Gifs clients read it on every request. Paging decides whether more results exist from the size actually requested.
- **Flush keeps the 429 block.** `search.Service.Flush` drops the cached pages for one budget. The 429 block stays, and so does any 429 that arrives after the flush, because a separate key generation now guards 429 recording. Only `Invalidate` (a key change) lifts a block.
- **Four new settings.** `SetPref` accepts `giphyRating`, `pageSize`, `giphyLang` and `klipyCountry`. Each value is checked against Go allow-lists and flushes only the providers it affects. Unknown values in settings.json are reset on load, with a log line. Match System is resolved for each request from `platform.SystemLocale`. Some language tags are mapped first: zh Hant/TW/HK/MO goes to zh-TW, other zh to zh-CN, nb/nn to no, fil to tl, iw to he. Anything left falls back to en, and an unknown country falls back to us.
- **Search pane.** It has GIFs per page (24/36/48 with helper S2), Content rating (G, PG, PG-13, R), Language (with helper S3) and Country. The language and country lists are named with `Intl.DisplayNames` and sorted A–Z, with Match System first. The panel marks provider tabs stale when a parameter changes. A stale tab refetches its first page on panel:show or when the user switches to it, and the results already on screen stay until then.
- **Keys in Go.** `KeyStatus` returns the last 4 characters only. `TestKey` makes one uncached trending request inside the 429 budget; on success it seeds the cache and clears Rejected, and on `key_rejected` it marks the key rejected. `RemoveKey` deletes the Keychain item, drops cached pages and the 429 block, resets Clips and emits `providers:changed`. All three share `SettingsService.mu` with `SaveKey`. A `key_rejected` from search also marks the key; for Clips it marks the GIPHY key.
- **Keys pane.** One card for GIPHY, then one for Klipy. Each shows its status, an Add/Change editor that uses the Phase 2 verify-then-save path, Test, an inline remove confirmation (focus on Cancel, Esc cancels) and the UI-SPEC copy.

## Task Commits

1. **Task 1 (tracer): content rating end to end.** `863acb1` (test, RED), `99898e8` (feat). The tracer gate re-ran the full `<verify>` (gofmt, vet, race tests, server build, node tests, svelte-check, smoke, and the 16-settings-search, 02-grid and 03-states UATs) and all passed before expansion continued.
2. **Task 2: page size from Params.** `169c1a6` (test, RED), `864cc2f` (feat)
3. **Task 3: GIPHY lang and Klipy locale.** `e56aa40` (test, RED), `5f6ebde` (feat)
4. **Task 4: platform.SystemLocale.** `912a2b4` (feat; glue only, no `<behavior>`)
5. **Task 5: page size, language and country settings.** `207e808` (test, RED), `a3c94ed` (feat)
6. **Task 6: Search pane rows and stale tabs.** `b1c1b7c` (test, RED), `3062f93` (feat)
7. **Task 7: Keys in Go.** `d054600` (test, RED), `abb826f` (feat)
8. **Task 8: Keys pane.** `c739ce8` (test, RED), `cc5b99f` (feat)

## TDD Gate Compliance

- Every TDD task has a `test(04-02)` RED commit before its `feat(04-02)` GREEN commit (7 pairs). Task 4 is glue with no `<behavior>`.
- **Frontend RED evidence (TAP).** `gsd-tools check tdd-red-evidence` returned `RED_EVIDENCE_OK` for "languageOptions puts Match System first, then languages by English name" (prefs.test.ts) and "testKeyMessage: passed (K-T1) and rejected (K-T2)" (errors.test.ts). Semantic check: each target failed on its planned assertion against a stub (`[]`, `''`), with no load errors.
- **Go RED evidence.** The classifier does not accept `go test` output, so this is a semantic assessment only. Every new Go target compiled against a skeleton and failed on its planned assertion: rating still pg-13, Flush a no-op, per_page 24, lang missing, empty lists, SetPref invalid_value, Params en/us, KeyStatus empty, TestKey and RemoveKey no-ops.
- **Tests that passed early, as expected:**
  - `TestPageSizeLimit`: Task 1 had already moved `limit` onto Params.
  - `TestDefaultParams`: added in Task 1.
  - `TestSetPrefGiphyRatingInvalid`: unknown names were already refused.
  - `TestParamsDefaultRating`'s Params half.
- **Task 3 commit order.** GREEN code was written before the RED commit. The RED failure output was captured first. The RED commit then stores the tests with empty `Languages`/`Countries` stubs, and the populated lists land in the GREEN commit.
- **One test expectation fixed in GREEN.** `TestTestKeyInsideBudget` expected Status 429 on `rate_limited`; the provider error carries Retry-After and Status 0.

## Verification (plan level)

- `gofmt -l` is clean. `go vet . ./internal/...` passes. `go test -race ./internal/... -count=1` passes in every package with no data race. `golangci-lint run . ./internal/...` reports 0 issues.
- `CGO_ENABLED=0 go build -tags server .` succeeds. `node --test` reports 76 pass, 0 fail. svelte-check reports 0 errors and 0 warnings.
- `scripts/smoke-browser-loop.sh` ends with SMOKE OK. It checks KeyStatus (`"last4":"-123"`, no full key), TestKey 200 and RemoveKey 200, and finds no fixture key in settings.json or the log.
- All 22 `scripts/uat/*.js` print UAT OK, including the new 16-settings-search and 16-settings-keys.
- go.sum is unchanged. Bindings were regenerated with `wails3 generate bindings -clean=true -ts -i`.
- Sources:
  - The GIPHY language list was fetched from https://developers.giphy.com/docs/optional-settings/ on 2026-10-18. It has 32 codes.
  - `klipy.Countries` comes from `NSLocale.ISOCountryCodes`: 257 codes.

## Decisions Made

See key-decisions in the frontmatter. The main ones:

- `settings` imports the provider lists directly.
- `keyGen` is kept apart from the cache generation.
- Defaults are stored as empty.
- `ClearRejected` was added.
- The key message line wraps.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] A 429 that lands after a Flush would have been dropped**
- **Found during:** Task 1
- **Issue:** Flush has to bump the generation so that an in-flight page fetched with the old parameters is not cached. That same generation also gated 429 recording, so a real 429 answered after a settings change would not block the budget, and the next request would go out into the 429 again.
- **Fix:** A separate `keyGen`, bumped only by `Invalidate`, now guards `noteRateLimit`. `gen` still guards caching. Test: `TestFlushInFlight429StillBlocks`.
- **Files:** internal/search/search.go, internal/search/search_test.go. **Commit:** 99898e8

**2. [Rule 1 - Bug] The Keys message line would have cut off errors**
- **Found during:** Task 8. The UI-SPEC backstop check in UAT caught it.
- **Issue:** UI-SPEC asks for one line with an ellipsis at a 400px width. K2 is 437px at 12px, and K3 and K-T2 are longer, so the "what to do next" part would be hidden.
- **Fix:** The message line wraps (`min-height: 16px`). UAT asserts no horizontal overflow and at most two lines.
- **Files:** frontend/src/settings/KeyRow.svelte, scripts/uat/16-settings-keys.js. **Commit:** cc5b99f

**3. [Rule 3 - Blocking] 15-settings-general.js assumed General was the only pane**
- **Found during:** Task 6
- **Issue:** Its ↑/↓ check expected focus never to leave General.
- **Fix:** The check now walks the panes, clamps at both ends and returns to General.
- **Files:** scripts/uat/15-settings-general.js. **Commit:** 3062f93

**4. [Rule 3 - Blocking] Existing provider tests banned lang/locale**
- **Found during:** Task 3
- **Issue:** `TestRequestParams` (giphy) and `TestRequest` (klipy) asserted that no lang/locale was sent.
- **Fix:** They now assert the new contract: `lang=en` on search only, and `locale=us` on Klipy search and trending.
- **Files:** internal/provider/giphy/giphy_test.go, internal/provider/klipy/klipy_test.go. **Commit:** e56aa40

### Plan adjustments (no behaviour lost)

- `provider.ParamsOf(f)` is a shared nil-to-defaults helper used by all three clients (three real callers), not a `params()` method per client.
- `Keys.ClearRejected(p)` was added beside MarkRejected and Rejected, so that a passing Test can clear the flag without a Keychain write.
- `frontend/src/main.ts` now passes the initial Snapshot to App as a prop, so the panel knows the starting provider params. main.ts was not in the plan's file list.
- The Search pane sets `selected` on each `<option>` because the lists arrive from `Choices()` after mount. The UATs wait for the options before reading values.
- S3 sits under the Language row, the last row of the GIPHY section.
- Add Key… and Change Key… keep their visible text as their accessible name. The card is a `role="group"` labelled by the provider name. Test and Remove carry `Test {Provider} key` and `Remove {Provider} key` as the plan asks. This keeps the visible text inside the accessible name for Add/Change.
- The SetPref failure copy on the Search pane reuses 04-01's `Couldn't save this setting. Try again.`.

**Total deviations:** 4 auto-fixed (2 Rule 1, 2 Rule 3) and 7 minor adjustments. **Impact:** none on scope. Both Rule 1 fixes make the plan's own truths hold: the 429 block is kept, and errors show in full.

## Issues Encountered

- Another agent edited 04-05-PLAN.md, 04-06-PLAN.md, 04-07-PLAN.md and 04-UI-SPEC.md while this plan ran. None of them were staged or touched.

## Known Stubs

None.

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: bound-surface | internal/services/settings.go | `SettingsService.Params()` is exported, so Wails binds it. It returns only the non-secret provider parameters (page size, rating, lang, country), which the Snapshot already exposes. No key. |
| threat_flag: country-list | internal/provider/klipy/countries.go | NSLocale.ISOCountryCodes includes 8 codes that are not ISO 3166-1 alpha-2: ac, cp, cq, dg, ea, ic, ta, xk. They are offered as Klipy locales. Per the plan's assumption, an unsupported locale just returns Klipy's default results. |

## Human-check items deferred to end of phase (native, user-verified on the signed install)

1. Open Settings → Keys → GIPHY → Change Key…, paste your real key with ⌘V and press Save Key. Expected: "Saved · ends in {last 4}". Then press Test. Expected: "GIPHY accepted the key."
2. Remove the Klipy key, quit Gifsmith from the menu bar and relaunch it. Expected: Klipy shows "Not set" and the panel's Klipy tab shows its setup card, which means the Keychain item is gone.
3. Set Content rating to G and then R. Each time, reopen the panel. The GIPHY and Clips trending results should change after the panel opens, not before.
4. If you can switch macOS to another language or region (for example German / Germany), check two things. Settings → Search should show "Match System (German)" and "Match System (Germany)". A GIPHY search and Klipy trending should come back localised. On an unsupported language, the label should read "Match System (English)".
5. Set GIFs per page to 48. Check that GIPHY and Klipy pages fill a longer first page and that scrolling still loads more.

## Next Phase Readiness

Plans 04-03..04-06 can build on these:

- `SetPref` for provider parameters, with `Svc.Flush`.
- `Choices`.
- The KeyStatus/TestKey/RemoveKey bound methods.
- The Search pane sections, where 04-04 adds Grid size above Results.
- The stale-tab mechanism in App.svelte.

---
*Phase: 04-settings-shell-polish-release*
*Completed: 2026-10-18*

## Self-Check: PASSED

- All created files exist.
- All 15 plan commits (863acb1..cc5b99f) are ancestors of HEAD; `git rev-list --count 9a80f80..HEAD` = 15.
- The acceptance-criteria greps were re-run, as was the plan-level verification above: all green.
