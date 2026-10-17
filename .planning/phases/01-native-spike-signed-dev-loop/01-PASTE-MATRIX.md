# Phase 1 Paste Matrix

| Field | Value |
|-------|-------|
| macOS version | 27.0.1 |
| Dev build commit | PENDING (set to the 01-03 Task 2 commit when the dev columns are recorded) |
| Notarized build commit | PENDING (plan 01-04) |
| Paste delay (ms) | 80 |
| Date | 2026-10-17 |

**Legend:** PASS = pasted and animated · STATIC = pasted but not animated · FAIL = nothing pasted or wrong place · n/a = app not installed · PENDING = not tested yet

"Signed dev" = `~/Applications/Gifsmith.app` installed by `wails3 task darwin:install:dev` (Developer ID, hardened runtime, not notarized).

## 1. Native behavior checklist

| # | Behavior | Signed dev | Notarized | Notes |
|---|----------|------------|-----------|-------|
| N1 | ⌃⌥⌘G shows/hides the panel from a normal app | PENDING | PENDING | |
| N2 | Same from a full-screen app | PENDING | PENDING | |
| N3 | Typing reaches the input immediately, no click | PENDING | PENDING | |
| N4 | Previous app stays active while the panel is up | PENDING | PENDING | |
| N5 | Esc hides | PENDING | PENDING | |
| N6 | Click-away hides (including a click on the previous app's window) | PENDING | PENDING | |
| N7 | No Dock icon, not in ⌘Tab | PENDING | PENDING | |
| N8 | Menu shows Open Gifsmith / Settings… / Quit Gifsmith, and Open shows the panel | PENDING | PENDING | |
| N9 | Panel on the cursor's display, centered near the top | PENDING | PENDING | |
| N10 | No clash with Vorssaint | PENDING | PENDING | |
| N11 | Enter then ⌘V in TextEdit inserts the animated GIF with Accessibility OFF | PENDING | PENDING | |
| N12 | ⌘Enter with Accessibility OFF shows the copied + how-to-grant message and pastes nothing | PENDING | PENDING | |
| N13 | "Enable automatic paste…" raises the prompt / opens the Accessibility pane | PENDING | PENDING | |
| N14 | ⌘Enter with Accessibility ON pastes into TextEdit with no manual ⌘V | PENDING | PENDING | |
| N15 | Accessibility grant carried over from the previous build without re-granting | n/a | PENDING | Dev column n/a: first signed build |

## 2. Per-app paste matrix

"Copy→⌘V" = Enter in the panel, then ⌘V yourself in the app. "⌘Enter" = ⌘Enter in the panel with Accessibility ON (Gifsmith posts ⌘V).

| App | Copy→⌘V (dev) | ⌘Enter (dev) | Copy→⌘V (notarized) | ⌘Enter (notarized) | Notes |
|-----|---------------|--------------|---------------------|--------------------|-------|
| TextEdit (rich text) | PENDING | PENDING | PENDING | PENDING | |
| Notes | PENDING | PENDING | PENDING | PENDING | |
| Mail (new message) | PENDING | PENDING | PENDING | PENDING | |
| Messages | PENDING | PENDING | PENDING | PENDING | |
| Slack | PENDING | PENDING | PENDING | PENDING | |
| Discord | PENDING | PENDING | PENDING | PENDING | |
| Telegram | PENDING | PENDING | PENDING | PENDING | |
| Safari composer (e.g. a Gmail or GitHub comment box) | PENDING | PENDING | PENDING | PENDING | |
| Chrome composer (same site) | PENDING | PENDING | PENDING | PENDING | |

## 3. Payload conclusion

PENDING. Is a single file URL on the pasteboard enough? Which apps need more (image data, a URL string)?

## 4. Go / No-Go

PENDING (filled from plan 01-04's decision).
