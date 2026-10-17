# Phase 1 Paste Matrix

| Field | Value |
|-------|-------|
| macOS version | 27.0.1 |
| Dev build commit | e9f1230 (gate re-test build; first signed build ce0c96d) |
| Notarized build commit | PENDING (plan 01-04) |
| Paste delay (ms) | 80 |
| Date | 2026-10-17 |

**Legend:** PASS = pasted and animated · STATIC = pasted but not animated · FAIL = nothing pasted or wrong place · n/a = app not installed · PENDING = not tested yet

"Signed dev" = `~/Applications/Gifsmith.app` installed by `wails3 task darwin:install:dev` (Developer ID, hardened runtime, not notarized).

## 1. Native behavior checklist

| # | Behavior | Signed dev | Notarized | Notes |
|---|----------|------------|-----------|-------|
| N1 | ⌃⌥⌘G shows/hides the panel from a normal app | PASS | PENDING | |
| N2 | Same from a full-screen app | PASS | PENDING | |
| N3 | Typing reaches the input immediately, no click | PASS | PENDING | |
| N4 | Previous app stays active while the panel is up | PASS | PENDING | |
| N5 | Esc hides | PASS | PENDING | |
| N6 | Click-away hides (including a click on the previous app's window) | PASS | PENDING | |
| N7 | No Dock icon, not in ⌘Tab | PASS | PENDING | |
| N8 | Menu shows Open Gifsmith / Settings… / Quit Gifsmith, and Open shows the panel | PASS | PENDING | |
| N9 | Panel on the cursor's display, centered near the top | PASS | PENDING | |
| N10 | No clash with Vorssaint | PENDING | PENDING | Passed in the 01-01 gate (unsigned dev build); not re-checked on the signed build |
| N11 | Enter then ⌘V in TextEdit inserts the animated GIF with Accessibility OFF | PASS | PENDING | |
| N12 | ⌘Enter with Accessibility OFF shows the copied + how-to-grant message and pastes nothing | PASS | PENDING | |
| N13 | "Enable automatic paste…" raises the prompt / opens the Accessibility pane | PENDING | PENDING | Earlier FAIL was plain Enter (Copy path). Button not clicked in the re-test; the user reached the permission list via the same `x-apple.systempreferences:…?Privacy_Accessibility` URL the button opens (run by the orchestrator), so the destination is confirmed but the click itself is unverified. Re-check at 01-04 UAT |
| N14 | ⌘Enter with Accessibility ON pastes into TextEdit with no manual ⌘V | PASS | PENDING | First try crashed (TIS keyboard-layout lookup off the main thread); fixed in e9f1230. Re-test: ⌃⌥⌘G then ⌘Enter ×5 in TextEdit pasted each time, no crash |
| N15 | Accessibility grant carried over from the previous build without re-granting | n/a | PENDING | Dev column n/a: first signed build |

## 2. Per-app paste matrix

"Copy→⌘V" = Enter in the panel, then ⌘V yourself in the app. "⌘Enter" = ⌘Enter in the panel with Accessibility ON (Gifsmith posts ⌘V).

| App | Copy→⌘V (dev) | ⌘Enter (dev) | Copy→⌘V (notarized) | ⌘Enter (notarized) | Notes |
|-----|---------------|--------------|---------------------|--------------------|-------|
| TextEdit (rich text) | PASS | PASS | PENDING | PENDING | |
| Notes | PASS | PASS | PENDING | PENDING | |
| Mail (new message) | PASS | PASS | PENDING | PENDING | |
| Messages | PASS | PASS | PENDING | PENDING | |
| Slack | PENDING | PENDING | PENDING | PENDING | |
| Discord | PENDING | PENDING | PENDING | PENDING | |
| Telegram | PENDING | PENDING | PENDING | PENDING | |
| Safari composer (e.g. a Gmail or GitHub comment box) | PENDING | PENDING | PENDING | PENDING | |
| Chrome composer (same site) | PENDING | PENDING | PENDING | PENDING | |

Slack, Discord, Telegram and the Safari/Chrome composers were not tested in the dev gate; they stay PENDING for the 01-04 notarized UAT.

**macOS 27 Settings finding (for Phase 4 Settings/onboarding):** on macOS 27.0.1 the user saw no "Accessibility" row under System Settings › Privacy & Security. The permission list is reachable only via the URL `x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility` (what "Enable automatic paste…" opens). Onboarding copy must not tell the user to navigate there by hand; use the button/URL.

## 3. Payload conclusion

PENDING. Is a single file URL on the pasteboard enough? Which apps need more (image data, a URL string)?

## 4. Go / No-Go

PENDING (filled from plan 01-04's decision).
