---
status: testing
phase: 04-settings-shell-polish-release
source: [04-VERIFICATION.md]
started: 2026-10-20T17:00:00Z
updated: 2026-10-20T17:00:00Z
---

## Current Test

number: 5
name: Keys & Search — checklist B7–B9
expected: |
  See 04-HANDSON-CHECKLIST.md group B.
awaiting: user response

## Tests

### 1. Release workflow dry run on main (SC7, RELS-07)
expected: split build/sign workflow passes end to end on Actions after the reviewer approves; no release or tag created.
result: pass — run 39473181660 on main ae3d0cd: build ok, reviewer approved, both notarizations Accepted, RELEASE VERIFY OK (notarized), TestRealDMG PASS, "dry run: no release created"; releases still only the v1.0.0 draft. CI run 39473179525 green.

### 2. Rebuild 1.0.0 from HEAD and replace the draft's assets (SC4, RELS-01)
expected: `wails3 task darwin:release` → RELEASE VERIFY OK (notarized); draft DMG + .sha256 replaced; TestRealDMG and TestReleaseJSON pass against the new draft.
result: pass — local build from ae3d0cd tree: 2× Accepted, RELEASE VERIFY OK (notarized); draft assets replaced, DMG sha256 e2a8b510f7fb0b1bfb0b184d6b1dc180d91ff42ad745a51ce1dc3d3544d67429; TestRealDMG + TestReleaseJSON PASS against the draft's JSON.

### 3. Clean-VM install of the rebuilt DMG (RELS-02, RELS-03)
expected: same 6 steps as the 2026-10-20 VM test (Gatekeeper prompt only, menu bar/no Dock, hotkey, copy/paste animated GIF, ⌘↵ copy-only + Enable Automatic Paste…, About 1.0.0 + up to date).
result: pass — fresh macOS 15.7.7 Tart VM, Gatekeeper enabled, draft DMG e2a8b510… downloaded via gh with quarantine, Finder drag to /Applications (not translocated): (1) only the "downloaded from the Internet… Apple checked it" prompt, spctl source=Notarized Developer ID; (2) menu bar icon, no Dock icon; (3) ⌃⌥⌘G opens panel, first-run key setup, real key saved, trending loads; (4) Enter puts a public.file-url .gif on the pasteboard, ⌘V into TextEdit pastes it animated; (5) ⌘↵ without Accessibility → "Copied — press ⌘V to paste" + Enable Automatic Paste…; (6) Settings → "Gifsmith 1.0.0 is up to date", About shows Version 1.0.0 with credits. VM deleted after.

### 4. Settings window — checklist A1–A6
expected: activation and typing, ⌘W/red button hide and restore the previous app, theme switch without flash, login item on/off, Login Items change reflected on reopen, Accessibility grant shows ✓ On.
result: pass — user on host (~/Applications copy of rebuilt 1.0.0). A1/A3/A6 yes. A2: user accepts current behavior (panel hides when Settings activates, by design). A4: macOS "background item added" notice on enable, none on disable (normal). A5: unchecked after close+reopen as specified; does not refresh live while Settings is open → deferred to 1.0.1 (user decision).

### 5. Keys & Search — checklist B7–B9
expected: real GIPHY key saves and tests OK; content rating changes trending; 48 per page fills and paginates.
result: [pending]

### 6. Hotkeys — checklist C10–C14
expected: rebind works across apps and in the tray; Favorites/Recents hotkeys; recording captures without firing and Esc/Tab/Shift+Tab cancel cleanly; hotkeys work after leaving a recording; clearing Open clears the tray shortcut and Reset to Default restores ⌃⌥⌘G.
result: [pending]

### 7. Downloads & grid — checklist D15–D18
expected: folder sheet + privacy prompt, ⌘S saves there; hidden names gifsmith-{id}.gif then (2); Large/Small grids; right-click menu placement; Klipy ⌘S refusal.
result: [pending]

### 8. Motion, About, Updates — checklist E19–E21
expected: Reduce motion stills; About links and GIPHY mark in both themes; Updates section in both themes.
result: [pending]

## Summary

total: 8
passed: 4
issues: 0
pending: 4
skipped: 0
blocked: 0

## Gaps

Deferred to 1.0.1 (user, 2026-10-20): General pane login-item checkbox should re-read status when Settings regains focus, not only on reopen.

Checklist detail: 04-HANDSON-CHECKLIST.md.
Phase 1 carry-over (not blocking v1.0.0): 21 PENDING cells in 01-PASTE-MATRIX.md (notarized column; Slack/Discord/Telegram/browser rows).
