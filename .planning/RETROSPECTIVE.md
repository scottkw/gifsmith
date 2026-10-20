# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.0 — Gifsmith MVP

**Shipped:** 2026-10-20 (v1.0.0, https://github.com/scottkw/gifsmith/releases/tag/v1.0.0)
**Phases:** 4 | **Plans:** 26 | **Timeline:** 2026-10-17 → 2026-10-20 (4 days, ~340 commits)

### What Was Built
- A non-activating hotkey panel that takes typing while the previous app keeps focus, puts the GIF on the pasteboard as a file URL, and pastes into the previous app with ⌘↵ when Accessibility is granted.
- GIPHY, Clips (behind a probe), Klipy and Finer Gifs search, with Keychain keys, caching, a 429 budget and vendor-compliant attribution.
- The full ⌘K action pipeline:
  - copy and paste of the GIF, its link or Markdown
  - lossless square crop
  - validated downloads
  - favorites, recents and recent searches
- A Settings window covering keys, search, hotkeys, downloads, theme, login item and updates.
- A signed, notarized DMG, an in-app update check, and CI with reviewer-gated, tag-driven notarized draft releases.

### What Worked
- **Risk-first Phase 1 spike.** The go/no-go gate on panel, pasteboard and paste found no framework blockers in Wails v3 beta.30 before any UI existed.
- **Browser UAT through the CGO-free server build plus platform.Fake.** The 35 UI scripts run headless, and native-only behavior is left for a short hands-on checklist.
- **A Tart VM driven over SSH and VNC for the clean-install test.** It proved Gatekeeper, notarization and first run without touching the author's Mac.
- **Code review with up to 3 fix iterations per phase.** It caught real data-loss and rollback bugs, such as overwriting an unreadable settings.json and hotkey rollback.

### What Was Inefficient
- **Apple's 403 (missing agreement) cost a day of notarization uncertainty.** It cleared on its own about 2 days after the agreement was signed.
- **Stapling on GitHub macOS runners fails about 40% of the time (CloudKit -1005).** It took 3 dry runs because the first retry fix covered `staple` but not `validate`.
- **Fix passes kept adding new warnings,** so the 1.0.0 rebuild and re-notarization happened after the first draft.
- **The first VM install used `cp` over SSH,** which triggered App Translocation. Only a Finder drag tests what users actually do.

### Patterns Established
- Pin Wails, the `wails3` CLI and `@wailsio/runtime` to one exact beta. Upgrade only between phases.
- Release signing: build in a job with no secrets, sign in an environment job with a fresh keychain, wrap every stapler call in a retry, and make the tag ruleset maintainer-only.
- Vendor rules are product requirements:
  - no blending
  - no reordering
  - attribution marks
  - Klipy file actions disabled until written clearance arrives
- Settings never overwrites an unreadable file. It refuses and offers a reset that writes a .bak first.

### Key Lessons
1. Test installs the way users do: a quarantined download, a Finder drag into /Applications, then a first open.
2. When a fix addresses one flaky external call, apply it to every call to that same service.
3. Re-verify earlier phases after a later phase touches their code, before closing the milestone.
4. A one-word user approval covers that action only. Ask again for each push, dispatch and asset swap.

### Cost Observations
- Model mix: mostly Opus 5.5 (orchestrator and GSD agents)
- Notable: multi-agent review and fix loops plus parallel re-verification kept the orchestrator's context lean

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Phases | Plans | Notes |
|-----------|--------|-------|-------|
| v1.0 | 4 | 26 | Risk-first spike, then vertical slices. 49/49 requirements, 8/8 release UAT |
