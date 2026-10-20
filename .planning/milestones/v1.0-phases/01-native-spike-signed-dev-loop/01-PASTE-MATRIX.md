# Phase 1 Paste Matrix

| Field | Value |
|-------|-------|
| macOS version | 27.0.1 |
| Dev build commit | e9f1230 (gate re-test build; first signed build ce0c96d) |
| Notarized build commit | 8428b30 (04-06 `wails3 task darwin:release`, 2026-10-19; app notary id a46eb52e-c6e6-4ac2-8b28-05af458ca034 Accepted, DMG notary id f5f85406-c0dc-45d8-bb46-df79e7293a1e Accepted, both stapled) |
| Notarized test machine | Clean Tart VM, macOS 15.7.7 Sequoia (vanilla image, SIP on, Gatekeeper assessments re-enabled), 2026-10-20 (04-06 Task 2); no Accessibility grant, so ⌘Enter (AX ON) cells stay PENDING |
| Paste delay (ms) | 80 |
| Date | 2026-10-17 |

**Legend:** PASS = pasted and animated · STATIC = pasted but not animated · FAIL = nothing pasted or wrong place · n/a = app not installed · PENDING = not tested yet

"Signed dev" = `~/Applications/Gifsmith.app` installed by `wails3 task darwin:install:dev` (Developer ID, hardened runtime, not notarized).

## 1. Native behavior checklist

| # | Behavior | Signed dev | Notarized | Notes |
|---|----------|------------|-----------|-------|
| N1 | ⌃⌥⌘G shows/hides the panel from a normal app | PASS | PASS | Notarized: opened the panel on the first-run GIPHY key screen (04-06 VM) |
| N2 | Same from a full-screen app | PASS | PENDING | Notarized: not tested in 04-06 |
| N3 | Typing reaches the input immediately, no click | PASS | PENDING | Notarized: not tested in 04-06 (search was typed, but click-free focus was not reported) |
| N4 | Previous app stays active while the panel is up | PASS | PASS | Notarized: TextEdit stayed frontmost through ⌘Enter (04-06 VM) |
| N5 | Esc hides | PASS | PENDING | Notarized: not tested in 04-06 |
| N6 | Click-away hides (including a click on the previous app's window) | PASS | PENDING | Notarized: not tested in 04-06 |
| N7 | No Dock icon, not in ⌘Tab | PASS | PASS | Notarized: menu bar icon, no Dock icon, LSUIElement true (04-06 VM); ⌘Tab not reported |
| N8 | Menu shows Open Gifsmith / Settings… / Quit Gifsmith, and Open shows the panel | PASS | PENDING | Notarized: not tested in 04-06 (Settings opened, menu contents not reported) |
| N9 | Panel on the cursor's display, centered near the top | PASS | PENDING | |
| N10 | No clash with Vorssaint | PENDING | PENDING | Passed in the 01-01 gate (unsigned dev build); not re-checked on the signed build; notarized: not tested in 04-06 |
| N11 | Enter then ⌘V in TextEdit inserts the animated GIF with Accessibility OFF | PASS | PASS | Notarized: GIPHY "office", Enter, panel hid, ⌘V in TextEdit pasted an animated GIF (three frame samples differed), 04-06 VM |
| N12 | ⌘Enter with Accessibility OFF shows the copied + how-to-grant message and pastes nothing | PASS | PASS | Notarized: copy only, "Copied — press ⌘V to paste" with "Enable Automatic Paste…" (04-06 VM) |
| N13 | "Enable automatic paste…" raises the prompt / opens the Accessibility pane | PENDING | PENDING | Notarized: not tested in 04-06. Earlier FAIL was plain Enter (Copy path). Button not clicked in the re-test; the user reached the permission list via the same `x-apple.systempreferences:…?Privacy_Accessibility` URL the button opens (run by the orchestrator), so the destination is confirmed but the click itself is unverified. Re-check at 01-04 UAT |
| N14 | ⌘Enter with Accessibility ON pastes into TextEdit with no manual ⌘V | PASS | PENDING | Notarized: not tested in 04-06 (no grant in the VM). First try crashed (TIS keyboard-layout lookup off the main thread); fixed in e9f1230. Re-test: ⌃⌥⌘G then ⌘Enter ×5 in TextEdit pasted each time, no crash |
| N15 | Accessibility grant carried over from the previous build without re-granting | n/a | PASS | Dev column n/a: first signed build. Notarized: same DR as the signed dev build; grant carries over (04-06: `codesign -d -r-` of bin/release/Gifsmith.app equals ~/Applications/Gifsmith.app, identifier "com.kenscott.gifsmith" + Developer ID anchor + leaf OU S2K7P43927) |

## 2. Per-app paste matrix

"Copy→⌘V" = Enter in the panel, then ⌘V yourself in the app. "⌘Enter" = ⌘Enter in the panel with Accessibility ON (Gifsmith posts ⌘V).

| App | Copy→⌘V (dev) | ⌘Enter (dev) | Copy→⌘V (notarized) | ⌘Enter (notarized) | Notes |
|-----|---------------|--------------|---------------------|--------------------|-------|
| TextEdit (rich text) | PASS | PASS | PASS | PENDING | Notarized ⌘Enter not tested in 04-06 (no Accessibility grant in the VM) |
| Notes | PASS | PASS | PENDING | PENDING | Notarized: not tested in 04-06 |
| Mail (new message) | PASS | PASS | PENDING | PENDING | Notarized: not tested in 04-06 |
| Messages | PASS | PASS | PENDING | PENDING | Notarized: not tested in 04-06 |
| Slack | PENDING | PENDING | PENDING | PENDING | Notarized: not tested in 04-06 |
| Discord | PENDING | PENDING | PENDING | PENDING | Notarized: not tested in 04-06 |
| Telegram | PENDING | PENDING | PENDING | PENDING | Notarized: not tested in 04-06 |
| Safari composer (e.g. a Gmail or GitHub comment box) | PENDING | PENDING | PENDING | PENDING | Notarized: not tested in 04-06 |
| Chrome composer (same site) | PENDING | PENDING | PENDING | PENDING | Notarized: not tested in 04-06 |

Slack, Discord, Telegram and the Safari/Chrome composers were not tested in the dev gate; they stay PENDING for the 01-04 notarized UAT.

**macOS 27 Settings finding (for Phase 4 Settings/onboarding):** on macOS 27.0.1 the user saw no "Accessibility" row under System Settings › Privacy & Security. The permission list is reachable only via the URL `x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility` (what "Enable automatic paste…" opens). Onboarding copy must not tell the user to navigate there by hand; use the button/URL.

**Notarized columns:** deferred to Phase 4 (Apple agreement 403). `xcrun notarytool store-credentials gifsmith-notary … --team-id S2K7P43927` fails with "HTTP status code: 403. A required agreement is missing or has expired", even after the user accepted the developer.apple.com agreement banner and retried about 30 minutes later. No keychain profile exists, so no notarized build was made. **Update (04-06, 2026-10-19):** resolved; the `gifsmith-notary` profile works and the first notarized build (8428b30) was accepted and stapled.

## 3. Payload conclusion

Derived from the user's results (signed dev columns only): a single file URL (`public.file-url`) on the pasteboard is sufficient for TextEdit, Notes, Mail and Messages. Copy→⌘V and ⌘Enter both PASS, animated, in all four. Slack, Discord, Telegram and the Safari/Chrome composers are unverified, so whether any of them need image data stays open for Phase 3's pasteboard design note (D-12: file URL only unless the matrix proves otherwise). Paste delay 80 ms. Not needed: the D-08 activation fallback, a PopUpMenu window level, a mouse monitor, or the Cmd+Enter KeyBindings fallback.

## 4. Go / No-Go

Decision: GO WITH FALLBACKS

User's decision (2026-10-17, commit aaab9a6): go, building on the Developer ID–signed dev build. Limits:
1. The notarized dry run (de-risking RELS-01/RELS-02; N15 and the notarized columns) is deferred to Phase 4. It is blocked on the Apple Developer account agreement 403, which the user must resolve (check membership expiry and App Store Connect agreements) before Phase 4 release work.
2. Paste in the untested apps (Slack, Discord, Telegram, Safari composer, Chrome composer), the "Enable automatic paste…" button click (N13), and N10 on the signed build are re-checked in later UAT (Phase 3 action pipeline / Phase 4).
3. On macOS 27.0.1 the Accessibility permission list can't be reached by browsing System Settings › Privacy & Security, only via `x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility`. Phase 4 onboarding and Settings must deep-link there and never cite a path.
4. Payload: a single file URL on the pasteboard is sufficient for TextEdit, Notes, Mail and Messages (Copy→⌘V and ⌘Enter both PASS, animated); other apps are unverified. PasteDelay 80 ms. None of these were needed: the D-08 activation fallback, a PopUpMenu level, a mouse monitor, the KeyBindings fallback.
