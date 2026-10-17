# Project Research Summary

**Project:** Gifsmith
**Domain:** macOS menu-bar / global-hotkey GIF search launcher (standalone Go + Wails port of the Raycast "GIF Search" extension)
**Researched:** 2026-10-17
**Confidence:** MEDIUM-HIGH (stack, provider rules, and extension inventory verified against primary sources and probes; native panel/paste/per-app behavior unproven until the spike)

## Executive Summary

Gifsmith is a launcher-class utility: a hotkey reveals a warm, hidden, non-activating floating panel; the user types, sees an animated grid from one provider at a time, and presses Enter to put a GIF *file* on the pasteboard (or paste it into the app they were already in). The core loop only works if the native shell behaves — the panel must take typed input without activating Gifsmith, appear over full-screen apps, and leave the previous app focused. Wails v3 (`v3.0.0-beta.30`, pinned exactly) ships almost the entire launcher shell (NSPanel, non-activating, accessory policy, Carbon global shortcut, tray, autostart), shrinking our own cgo to four things: file-URL pasteboard write, CGEvent ⌘V, Accessibility check, and frontmost-app query. Wails v2 cannot do this product, so "prefer v3 if stable" resolves to "v3 beta, pinned, with rationale recorded as a Key Decision."

The recommended build is Go domain packages that never import Wails (providers, searchcache, library, gifcache, imaging, action orchestrator) behind thin Wails services and a tiny `platform` interface (darwin/stub/fake), with a Svelte 5 + Vite 8 + TS frontend rendering a virtualized, hand-built keyboard-first grid. All provider HTTP lives in Go so keys never reach the webview and errors can be redacted in one place. The extension's provider models target Raycast's proxy, not the real APIs — they are a behavior reference only; Klipy and GIPHY Clips need new mappers, and favorites/recents must store metadata snapshots rather than ID lists.

The dominant risks are (1) native feasibility — non-activating panel input, full-screen overlay, and ⌘V synthesis are unproven on macOS 27, hence the spike as a hard go/no-go gate; (2) provider terms and quotas — both vendors forbid storing media, blending providers, reordering results, and omitting attribution, and free keys are capped at 100 calls/hour; (3) dependency churn — Wails v3 is a nightly-snapshot beta; and (4) signing/permission identity — TCC grants and Keychain are bound to code signature, so dev builds must be signed with the Developer ID identity from day one. Mitigations are cheap if designed in early: explicit compliance posture, a request budgeter, a pinned beta, a signed dev loop, and a throwaway notarized build in the spike.

## Conflict Resolutions (where researchers disagreed)

| Topic | Disagreement | Resolution |
|-------|--------------|------------|
| Square crop | PITFALLS worried about Plan9-palette re-quantization / delta frames; STACK *probed* `SubImage` on paletted frames + re-based rects + own palettes + `EncodeAll` → 0 pixel mismatches on 5 real GIFs. | **Use the verified STACK approach: geometry-only crop on paletted frames, pure `image/gif`.** No RGBA conversion, no re-quantization. Handle the one trap STACK found: frames fully outside the window need a fully transparent placeholder (or folded delay). Keep PITFALLS' golden-corpus test (frame count, delays, loop count, per-pixel diff) and a size/frame cap. Clips (MP4) excluded. |
| Favorites media cache | ARCHITECTURE designed cache-on-favorite; FEATURES D-8 and PITFALLS 2 say both vendors forbid storing/mirroring media. | **No persistent byte cache in v1.** Temp staging files only for the pasteboard file URL / saved download, in `Caches/Gifsmith/staging/<sha1(provider:id)>/<name>`, swept on launch and by age (~24h, never immediately after paste); Klipy staging deleted eagerly (next copy / quit). Favorites persist **metadata snapshots only**. Offline-media cache is a post-v1 opt-in with disclosure. `gifcache` keeps resolve/staging/download, drops the favorites-cache branch. |
| Aggregate Favorites/Recents | Extension interleaves providers; vendor terms prohibit blending in one grid. | **Per-provider tabs/sections only**, each carrying its provider's attribution. |
| Global hotkey | `golang.design/x/hotkey` vs Wails built-in. | **Wails v3 `app.GlobalShortcut` (Carbon).** x/hotkey's macOS backend is a CGEventTap needing Accessibility — breaks "copy works without it." Register after the app is running so rejection errors are returned (PITFALLS); default to a three-modifier combo; tray "Open Gifsmith" is the hotkey-free fallback. |
| Frontend framework | ARCHITECTURE sketched React, deferred to STACK. | **Svelte 5** (official Wails template), `@tanstack/svelte-virtual`, small hand-rolled paging store (not TanStack Query). |
| Keyboard paging | FEATURES KB-4 vs STACK IntersectionObserver. | Both: fetch next page when the **selected index** is within 1–2 rows of the end, and keep the sentinel for mouse/scroll users. |
| Result dedupe | Extension dedupes across pages; vendors forbid filtering. | Dedupe only same-id-on-same-page, **never dedupe Klipy**, never re-rank live results; frecency applies only to the user's own Favorites. |
| Provider HTTP location | PITFALLS 20: Go ignores macOS system proxy/PAC. | **Keep Go** (keys out of webview, central rate limiting); add a Settings diagnostics line and decide on `SCDynamicStoreCopyProxies` support in the provider phase. |
| Keychain argv exposure | PITFALLS flagged a possible `ps` leak. | STACK read v0.2.8 source: secret goes via stdin. **Use `zalando/go-keyring` v0.2.8** behind a `SecretStore` interface. |
| Clips / Klipy lookup | No reliable by-id route for either. | Snapshot-first library; `Lookup` is an optional best-effort refresh. Klipy uses `slug` as the stable id. |

## Key Findings

### Recommended Stack

Go (≥ 1.25; 1.27.2 installed) with **Wails v3 `v3.0.0-beta.30` pinned exactly** in `go.mod`, the `wails3` CLI pinned (`go install ...@v3.0.0-beta.30`; the existing `~/go/bin/wails` is v2 — leave it), and `@wailsio/runtime@3.0.0-beta.30`. Never `@latest` (resolves to the newest nightly). Frontend Svelte 5.57.2 + TS 5.9 + Vite 8.3.4, **pnpm from the first command** (template Taskfile defaults to npm; mixing failed in the probe). Wails-specific code is confined to a thin `internal/shell` package.

**Core technologies:**
- Wails v3 beta.30 — shell, tray, NSPanel, global shortcut, autostart, bindings
- Own cgo/ObjC (one `internal/platform` package, build-tagged `darwin && !server`, with a stub) — `NSPasteboard writeObjects:@[fileURL]` (probe-verified), CGEvent ⌘V, `AXIsProcessTrustedWithOptions` / `CGPreflightPostEventAccess`, `NSWorkspace frontmostApplication`
- `image/gif` (stdlib) — lossless square crop (verified)
- `zalando/go-keyring` v0.2.8 — Keychain keys, no cgo, survives ad-hoc rebuilds
- JSON files with atomic write under `os.UserConfigDir()` — settings and library; no SQLite
- `golang.org/x/time/rate` — client-side budgeter for 100/hour caps
- Svelte 5 + `@tanstack/svelte-virtual` — virtualization is a correctness requirement
- Sign/notarize via `wails3 task darwin:sign:notarize` with the Developer ID identity already in the keychain (`Developer ID Application: Ken Scott (S2K7P43927)`)

Opaque webview + CSS rounded corners (no `private_mac_apis` tag); `LSUIElement=true` in `Info.plist`; never terminate-after-last-window-closed; browser-tab UAT requires `go run -tags server .` + a Vite proxy for `/wails`. `golang.design/x/clipboard` v0.11.0 `WriteFiles` is a documented fallback only.

### Expected Features

**Must have (table stakes):**
- Global hotkey → floating panel, Esc/click-away dismiss, focused input; menu-bar icon, no Dock (T-1, T-2)
- Search + Trending across four providers with a switcher that remembers the last provider; infinite scroll with page-size setting; Finer Gifs has no trending (T-3, T-4)
- Copy GIF as file (default action) → then Paste GIF with Accessibility gating and a "Copied — press ⌘V" fallback (T-5, T-6)
- Link / Markdown / Page Link / Open in Browser / Download actions; Square copy/paste (T-7, T-8)
- Action panel (⌘K) + configurable default Enter action (T-9, T-10)
- Favorites and Recents per provider, with frecency; Recents capped (100–200); details view (T-11–T-13)
- Settings (keys, hotkey, grid sizes, max results, default action, download folder, hide filename, locales) (T-14)
- **Keyless first run + guided key setup, vendor-compliant attribution, honest error/rate-limit states** (T-15–T-17) — moved early because they decide whether the public release works for strangers
- Keyboard-first grid navigation (T-19, KB-1–KB-9); signed + notarized build with MIT attribution (T-20)

**Should have (cheap differentiators for v1):**
- ⌘Enter sibling action (D-2); GIPHY content rating default PG-13 (D-5); recent searches (D-6)
- Inline key paste-and-verify (D-10); separate Favorites/Recents hotkeys (D-11)
- Metadata-snapshot favorites that work offline/without API calls (D-3) — a storage-format decision, so it is built in from the start

**Defer (v1.x / v2+):**
- Drag-out of the selected GIF (D-1) — best differentiator but needs a native drag source; plan after the file pipeline is stable
- Categories/suggestions/autocomplete (D-4), request-budget indicator (D-9; 429 handler + simple counter are v1)
- Persistent media cache (D-8), quick preview, frontmost-app display, menu-bar recents submenu (D-7, D-12, D-13)
- Analytics pingbacks / share triggers (off by default)

**Anti-features:** Tenor (API shut down 2026-06-30), hosted key proxy, any call to Raycast's proxy, unified cross-provider search or blended grid, re-ranking live results, GIF recording/editing, accounts/sync/telemetry, Mac App Store build, in-app Wails updater (fail-open issue #6186), unbounded Recents.

### Architecture Approach

Layered: pure Go domain packages (`gif`, `provider/*` + `httpx`, `searchcache`, `gifcache`, `imaging`, `library`, `settings`, `action`) under thin Wails services (`GifService`, `ActionService`, `LibraryService`, `SettingsService`, `PanelService`), plus a `platform.Platform` interface (darwin cgo / stub / fake) and a `shell` package (panel, hotkey, tray, settings window). The frontend talks only to a facade over generated bindings, swappable for a mock. Key contracts: **opaque cursor** in the Go provider interface (GIPHY offset, Klipy page+`has_next`, Finer `start`/`size`; the frontend branches only on `HasTrending`/`NeedsKey`); **`Gif.Provider` explicit field** (no URL-sniffing port); **typed errors** via `ServiceOptions.MarshalError` (`key_missing`, `rate_limited`, `network`, `timeout`, `http`); **`action.Perform` owns side-effect order** — resolve file → pasteboard write → hide panel → ~60–100 ms → verify frontmost → post ⌘V → record usage only on success; **Settings in a separate regular window** (HideOnFocusLost would kill an in-panel settings view; include an Edit menu so ⌘V works in the key field).

**Major components:**
1. `provider` + `httpx` — real-API clients/mappers with recorded fixtures; redacts keys; 15s timeout; honors 429 `Retry-After`
2. `searchcache` + rate budgeter — TTL + singleflight
3. `library` — per-provider favorites/recents, snapshot records, mutex + atomic write, schema version
4. `gifcache` (staging/resolve/download) — validated downloads (status, content-type, magic bytes), exclusive-create naming, TTL sweep; no favorites byte cache
5. `imaging` (pure-Go square crop), `action` (orchestrator), `platform` (only cgo)
6. `shell` — Wails panel/hotkey/tray/settings window; panel created once, hidden, kept warm

Testing layers: L1 Go unit with fakes/fixtures; L2 dev-browser UAT against `-tags server` + `platform.Fake`; L3 mock-mode visual; L4 manual user-run native matrix. L2 cannot exercise NSPanel, hotkey, TCC, or real paste.

### Critical Pitfalls

1. **Porting the extension's provider models verbatim** — they target Raycast's proxy. Write clients against vendor docs, map into a neutral `Gif`, record live fixtures; gate Clips behind a capability probe.
2. **Building features the provider terms forbid** — media caching, blended grids, URL rewriting for render/download, result filtering, missing attribution. "Powered by GIPHY" with GIPHY/Clips; Klipy placeholder exactly `Search KLIPY`; media loaded straight from returned URLs. Record the compliance posture in PROJECT.md Key Decisions.
3. **Panel/paste native behavior unproven** — accessory policy + non-activating panel so the prior app never loses focus; fall back to record-and-reactivate only if the spike fails. ⌘V from a private CGEventSource with an explicit Command flag and layout-aware V keycode; wait for held modifiers to clear; gate on Accessibility and degrade to "copied — press ⌘V."
4. **TCC/Keychain bound to code signature** — sign every dev build with the Developer ID identity and a fixed bundle id; test paste only from the bundled `.app` at a fixed path, never `wails3 dev`; check Accessibility at paste time and poll after the deep link.
5. **Rate limits (100/hour) and animated-grid cost** — debounce 400–500 ms with 2-char minimum, cancel superseded requests, cache; compute `hasMore` from provider fields and clamp page sizes (GIPHY ≤50, Klipy 8–50); small renditions, virtualization, unmount images on panel hide, never GIF bytes over Wails IPC.
6. **Wrong pasteboard representation** — file `NSURL` only (no `NSImage`); add types only if the spike matrix demands it; keep link/Markdown as first-class fallbacks; keep Hide Filename (Slack filename leak).
7. **API-key leakage** — Klipy key is in the URL path and Go's `*url.Error` prints full URLs; redact in `httpx` only, never log URLs.

Also folded into phases: staging-file lifecycle bugs, hotkey conflicts (register after `Run`; recorder registers new before unregistering old), notarization dry run, cgo main-thread rules (AppKit only via `InvokeSync*`), Clips MP4 handling (square disabled, label "Copy clip (MP4)", 720p default).

## Implications for Roadmap

Dependency-driven and risk-first, following ARCHITECTURE's build order, with FEATURES' first-run/attribution items pulled earlier and the compliance decision placed before storage.

### Phase 1: Native Spike & Signed Dev Loop (hard go/no-go gate)
**Rationale:** The UX hinges on three unproven behaviors (non-activating panel keyboard input, full-screen overlay, ⌘V into the previous app). Failure changes the architecture, so learn it first. The signing identity must be stable before any paste test means anything.
**Delivers:** Wails v3 beta.30 scaffold (pnpm, pinned CLI/runtime); accessory policy + `LSUIElement` + tray; hidden warm non-activating panel with autofocused input; `GlobalShortcut` toggle registered after `Run` with error path; `platform.CopyFile` / `PasteShortcut` / Accessibility check with paste into TextEdit; `-tags server` + Vite proxy proving browser→bindings; signed dev loop; throwaway signed + notarized build; committed paste matrix (Slack, Messages, Mail, Discord, Notes, Telegram, Safari/Chrome composers; file-URL only) and native smoke script; min-macOS and arch decision.
**Addresses:** T-1, T-2, T-5, T-6 feasibility proof
**Avoids:** Pitfalls 3, 5, 6, 7, 8, 9, 10, 15 (dry run)

### Phase 2: Provider Core & Compliance Posture
**Rationale:** Everything visible depends on real provider data, and the data model (neutral `Gif`, snapshot records, opaque cursor) is the most expensive thing to change later. The ToS posture and rate budget are design inputs here.
**Delivers:** `gif` model, `httpx` (timeout, typed errors, redaction, 429), `Provider` interface + registry; Finer Gifs (keyless) → GIPHY → Klipy (native v1 API, key in path, slugs) → Clips behind a capability probe; `settings` + `SecretStore` (Keychain); `searchcache` + rate budgeter; `fileid` validation; recorded live fixtures with mapper tests; written compliance decision in PROJECT.md; system-proxy decision.
**Uses:** Go, go-keyring, `x/time/rate`
**Avoids:** Pitfalls 1, 2 (decision), 11, 14 (schema), 16, 21

### Phase 3: Search UI, First Run & Attribution
**Rationale:** First-run and attribution decide whether the public release works for strangers, so they ship with the first UI. The grid performance budget is set here.
**Delivers:** Svelte shell, search bar, provider switcher (keyed providers visible with a "key needed" badge), virtualized grid, selection- and sentinel-driven infinite scroll, keyboard model with layered Esc, per-provider empty/error/429/key-missing states, inline setup card with paste-and-verify (D-10), keyless "Try Finer Gifs" path, "Powered by GIPHY" and `Search KLIPY` attribution, small renditions only, image unmount on panel hide; L2/L3 dev-browser tests.
**Addresses:** T-3, T-4, T-15, T-16, T-17, T-19, D-10
**Avoids:** Pitfalls 2 (UI), 12, 23, 24

### Phase 4: Action Pipeline
**Rationale:** Copy first, then paste, then variants. Usage recording lives in the action layer, so actions precede the library.
**Delivers:** `gifcache` staging + validated download + exclusive-create naming; `action.Perform` orchestrator; file-URL copy (default); paste with Accessibility onboarding and fallback toast; link/Markdown/page/open/download; ⌘K action panel; configurable default (resolved at invocation time); ⌘Enter sibling; square crop (pure Go, golden corpus); Clips handling (MP4 label, 720p default, square disabled); `gifClipboard` regression cases ported as Go table tests against `platform.Fake`.
**Addresses:** T-5–T-10, T-18, D-2
**Avoids:** Pitfalls 3, 4, 5, 13, 19, 22, 25

### Phase 5: Library (Favorites, Recents, Details)
**Rationale:** Needs usage recording (Phase 4) and the snapshot schema (Phase 2).
**Delivers:** versioned `library.json` with snapshots; per-provider sections (never blended); favorites append, recents prepend/move-to-front with cap; frecency decayed to now; offline rendering with optional background refresh; details view; recent searches; in-search Favorites/Recent sections (first two rows).
**Addresses:** T-11–T-13, D-3, D-6
**Avoids:** Pitfalls 2 (storage), 14

### Phase 6: Settings Window & Shell Polish
**Rationale:** Mostly standard once the shell is proven; accessory-app quirks apply.
**Delivers:** separate regular Settings window (temporary activation) + Edit menu so ⌘V works in the key field; hotkey recorder (register-new-first, persist on success); key status/test/remove; grid sizes, max results (clamped), download folder (EPERM fallback), hide filename (default on), locales, GIPHY rating (default PG-13); launch at login; Accessibility banner with live polling; multi-display/Space placement refinement; separate Favorites/Recents hotkeys.
**Addresses:** T-14, D-5, D-11
**Avoids:** Pitfalls 6 (UX), 10 (recorder), 17, 18

### Phase 7: Release
**Rationale:** The dry run already happened in Phase 1, so this is a gate, not a discovery.
**Delivers:** final bundle id and entitlements justified one by one (start from none); sign → notarize → staple pipeline script with `spctl` verification; DMG; clean-account install test (quarantine, translocation, fresh TCC); MIT attribution + About entry; re-read of vendor ToS and GIPHY attribution guide; Klipy clearance outcome.
**Addresses:** T-20
**Avoids:** Pitfalls 6, 15

**v1.x stretch:** drag-out (D-1), categories/suggestions (D-4), request-budget indicator (D-9), opt-in offline media cache if cleared.

### Phase Ordering Rationale

- Spike first because its failure is the only one that changes the architecture, and `platform` being an interface bounds the damage.
- Providers before UI, keyless Finer Gifs first: a keyless provider proves search → grid → action before any key exists; the UI can lead on the mock seam once DTO/error shapes are frozen.
- Data model and compliance posture before storage, to avoid the HIGH-cost recovery path in PITFALLS.
- Actions before library, since usage recording and staging are action-layer side effects.
- Square crop after the file pipeline works, since it is isolated and the Clips question is settled by then.
- Release last, but its riskiest parts (credentials, entitlements, Info.plist) are flushed in the spike.

### Research Flags

Needs `/gsd-plan-phase --research-phase <N>`:
- **Phase 1:** Highest care. Panel typing, Esc/click-away, full-screen overlay, ⌘V timing/keycode, paste matrix, and notarization entitlements all need manual user-run verification. Research is done; this is hands-on spike work.
- **Phase 2:** Short live-key probing session needed: Klipy request shapes are untested with a live key (none was available), plus `customer_id`, ad items, Clips approval behavior, the system-proxy choice, and a Finer Gifs re-probe (down on 2026-10-17).
- **Phase 4:** Short design note on pasteboard representation (informed by the Phase 1 matrix) and Clips handling before coding.

Standard patterns (skip research-phase):
- **Phase 3:** Well-trodden Svelte + virtualization, apart from the grid performance budget.
- **Phase 5:** JSON store and sections once the schema is fixed.
- **Phase 6:** Standard apart from known accessory-app Settings/Edit-menu quirks.
- **Phase 7:** Standard once dry-run credentials exist.

## Open Decisions for the User

| # | Decision | Research recommendation | Needed by |
|---|----------|------------------------|-----------|
| 1 | **GIPHY Clips handling** — endpoints need per-key approval (clips@giphy.com); a self-issued beta key will likely fail. | Capability-probed provider: shown only if `/v1/clips/trending?limit=1` returns 200 on key save, with a clear explanation otherwise; not a launch blocker. Option: defer Clips out of the first release. Square actions disabled for Clips. | Phase 2 |
| 2 | **Klipy copy-file clearance** — Klipy forbids retaining copies; Copy/Paste/Download need a local file. | Eager-delete staging, keep link/Markdown as the no-file path, email developers@klipy.com for written clearance; **pre-release gate**. Decide whether to disable file actions for Klipy if clearance is refused. | Before Phase 7 |
| 3 | **Default provider on first run** | GIPHY with inline setup card plus a "Try Finer Gifs (no key)" button. Alternative: Finer Gifs default with an "add a GIPHY key" banner. Caveat: Finer Gifs is an unofficial hobby API and was down at research time — weak first impression. | Phase 3 |
| 4 | **Recents sort order** — extension is inconsistent. | Strict most-recent-first for Recents; frecency (decayed to now) for Favorites only. | Phase 5 |
| 5 | **Minimum macOS version and arch** | macOS 13+ (SMAppService autostart requires it). Choose arm64-only (simpler) or universal (two cgo builds + `lipo`). Decide in the spike and set cgo flags consistently. | Phase 1 |
| 6 | **Persistent favorites media cache** | Off in v1 (see Conflict Resolutions); revisit only as opt-in with disclosure. | Settled unless the user objects |
| 7 | **Analytics pingbacks / share triggers** | Off by default; they send usage data and aren't clearly required for a personal key, though a production-key review may look for them. | Phase 2 |
| 8 | **Paste as default action** | Keep Copy default (no Accessibility needed); offer Paste after permission setup. | Settled |

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH (features) / MEDIUM (beta stability) | Wails source and docs read directly; pasteboard write, crop, server-mode bindings, and keychain mechanics probe-verified on macOS 27.0.1. Nightly beta cadence, TS 7, `@tanstack/svelte-virtual` fit, and end-to-end notarization untested. |
| Features | HIGH (extension inventory, Klipy rules) / MEDIUM (GIPHY wording, competitors) | Extension read in full from source; Klipy docs read as raw markdown; GIPHY via summarizer; competitor claims web-sourced; first-run flow is design inference. |
| Architecture | MEDIUM-HIGH | Wails APIs verified in source; non-activating panel hosting a typed webview input and the server-mode frontend proxy are unproven for our flow. |
| Pitfalls | MEDIUM | Provider ToS, Wails issues, and extension history HIGH; per-app paste behavior, TCC identity behavior (no Apple primary source), Secure Input, and full-screen edge cases MEDIUM/LOW and routed to the spike. |

**Overall confidence:** MEDIUM-HIGH — the design is well supported on paper; the real uncertainty is native behavior that only the user can verify by hand, which the plan isolates into Phase 1.

### Gaps to Address

- **Panel behavior on macOS 27** (typing, Esc, click-away, full-screen overlay, `canBecomeKey`, whether `Show()` activates) — Phase 1 manual spike; contingency `platform.ActivateApp` (record-and-reactivate) and higher window level (`screenSaver`).
- **Animated-paste behavior per target app** (Messages, Discord, Mail, browsers) — Phase 1 paste matrix; do not add pasteboard types speculatively.
- **Klipy live request/response shapes** (no key during research): `customer_id`, ad items, items-by-slug, share trigger — Phase 2 live probe + fixtures.
- **GIPHY Clips access with a standard beta key** — Phase 2 probe.
- **Klipy "no copies" policy vs the copy-file action** — developers@klipy.com answer before release.
- **GIPHY attribution wording/size rules** (read via summarizer) — re-read the current guide during Phase 3; use official logo assets, not the extension's bundled PNGs.
- **GIPHY dashboard key-creation steps** (Zulip says choose "SDK") — confirm against the live dashboard when writing setup-card copy.
- **Finer Gifs availability** (Cloudflare 1033 on 2026-10-17) — re-probe at implementation; per-provider failure isolation, never block startup, never the sole first-run path.
- **⌘V timing and layout handling** (hide → ~60–100 ms → verify frontmost; layout-derived V keycode; Dvorak-QWERTY-⌘ case) — tune in the spike, then Phase 4.
- **Default hotkey** — pick a three-modifier default not used by mainstream launchers (e.g. ⌃⌥⌘G) and confirm it doesn't conflict with Vorssaint.
- **Notarization with generated entitlements** — start from none, add only justified ones, validate in the Phase 1 dry run.
- **Wails v3 upgrade policy** — pin beta.30; upgrade deliberately between phases; rerun the smoke script on macOS 26 and 27.

## Sources

### Primary (HIGH confidence)
- Wails v3 source and docs (`wailsapp/wails` master, v3.0.0-beta.30 release notes) — panel options, `GlobalShortcut`, autostart, systray, services/`MarshalError`, server build, `examples/spotlight`
- GIPHY developer docs — API, endpoints, Clips (approval), best practices (caching, mixing, attribution, beta key limits)
- Klipy docs (`docs.klipy.com`, `llms-full.txt`, `.md` pages) — native v1 API, integration requirements, attribution, getting started
- Apple developer docs — `CGPreflightPostEventAccess`, `NSRunningApplication.activate`
- Source reads: `golang.design/x/hotkey` v0.6.4, `golang.design/x/clipboard` v0.11.0, `zalando/go-keyring` v0.2.8
- Local Raycast extension source, tests, and CHANGELOG (MIT); its GitHub issues
- Empirical probes on macOS 27.0.1 (2026-10-17): NSPasteboard types, GIF crop round-trip (0 mismatches / 5 files), AX/CG permission and frontmost-app via cgo, `-tags server` binding calls through a Vite proxy, live status of Finer Gifs (530/1033) and GIPHY envelope

### Secondary (MEDIUM confidence)
- Maccy (`Clipboard.swift`) as the reference ⌘V implementation; Clipy/Sauce keycode layout notes
- Apple Developer Forums on `RegisterEventHotKey` Option-only regressions and full-screen overlay panels
- Zulip GIF-picker integration docs (keys are not secrets)
- Competitor store pages/reviews (Quickgif, Gnome, Grabbie, GIFme, GIF Bar, Hard G, Alfred workflows); Tenor shutdown reports (9to5Google, Shacknews)

### Tertiary (LOW confidence — needs validation)
- Per-app animated paste behavior beyond Slack-on-file-URL; Secure Input effects on synthesized events; hidden-webview throttling
- Slack/Discord picker features and Raycast's default ⌘⏎/⌘P (general knowledge)
- Third-party Klipy write-ups/client libraries (superseded by official docs where they overlap)

Detailed findings: `.planning/research/STACK.md`, `FEATURES.md`, `ARCHITECTURE.md`, `PITFALLS.md`.

---
*Research completed: 2026-10-17*
*Ready for roadmap: yes*
