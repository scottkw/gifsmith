---
phase: 03
review: 03-REVIEW.md
titles: json
findings:
  - id: CR-01
    severity: critical
    disposition: fixed
    title: "The decode budget counts only `w*h`, so zero-pixel transparent frames still turn a 15 MB GIF into about 4 GB of live heap and 9 GB of process memory"
  - id: IN-01
    severity: info
    disposition: open
    title: "Library mutation results are not ordered against an in-flight `refreshLibrary` (carried forward)"
  - id: IN-02
    severity: info
    disposition: open
    title: "⌘⌫ and ⌘Z override text editing in the library filter and the action-menu filter (carried forward)"
  - id: IN-03
    severity: info
    disposition: open
    title: "Error copy names the wrong failure for text copies, oversize or disallowed downloads, and `invalid_item` (carried forward)"
  - id: IN-04
    severity: info
    disposition: open
    title: "An interrupted download leaves a hidden `.gifsmith-*.part` file in ~/Downloads (carried forward)"
  - id: IN-05
    severity: info
    disposition: open
    title: "`refreshLibrary('')` does not select the first item, despite the comment (carried forward)"
  - id: IN-06
    severity: info
    disposition: open
    title: "A no-op removal clears Go's Undo slot but leaves an earlier Undo toast on screen, so ⌘Z then restores nothing (carried forward)"
  - id: IN-07
    severity: info
    disposition: open
    title: "`ClampID` can merge two distinct vendor ids that share a 200-byte prefix (carried forward)"
  - id: WR-01
    severity: warning
    disposition: fixed
    title: "No single-instance guard"
  - id: WR-02
    severity: warning
    disposition: fixed
    title: "`Validate` rejects long titles and ids instead of truncating them"
  - id: WR-03
    severity: warning
    disposition: fixed
    title: "A removal that changes nothing still shows an Undo toast, and ⌘Z then restores an earlier removal"
open: 8
total: 11
recorded: 2026-10-18T17:44:26.582Z
---

# Phase 03: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| CR-01 | critical | fixed | 6928379, 39d5c97 (frameOverhead per frame; verifier confirmed closed) |
| IN-01 | info | open | - |
| IN-02 | info | open | - |
| IN-03 | info | open | - |
| IN-04 | info | open | - |
| IN-05 | info | open | - |
| IN-06 | info | open | - |
| IN-07 | info | open | - |
| WR-01 | warning | fixed | 03-REVIEW-FIX.iter3.md (not in the current review) |
| WR-02 | warning | fixed | 03-REVIEW-FIX.iter3.md (not in the current review) |
| WR-03 | warning | fixed | 03-REVIEW-FIX.iter3.md (not in the current review) |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved. A `|` in the reason is kept as prose and escaped on the next run.
Re-running the gate keeps every row it can. A row the current review no longer reports is kept and its Source cell flagged, so a finding does not leave this record silently. ONE exception: when a finding id is REUSED by a different finding, the earlier decision cannot keep a row — the id is taken — and it is dropped. A RECORDED decision (anything but `open`) is named on the console when that happens; a row still at `open` is replaced silently, because `open` records no decision to lose.
