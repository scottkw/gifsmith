---
phase: 02
review: 02-REVIEW.md
titles: json
findings:
  - id: IN-01
    severity: info
    disposition: open
    title: "Raising the Wails log level to Debug would log every bound call's arguments, including the API key (carried forward)"
  - id: IN-02
    severity: info
    disposition: open
    title: "Retry-After is not capped (carried forward)"
  - id: IN-03
    severity: info
    disposition: open
    title: "Staged files persist across launches, against Klipy's no-retain rule (carried forward, deferred to Phase 3 as ACTN-10)"
  - id: IN-04
    severity: info
    disposition: open
    title: "A stale Clips error stays on screen after Clips becomes enabled (carried forward, one more path)"
  - id: IN-05
    severity: info
    disposition: open
    title: "SetupCard's error path dereferences `field` after the card may have unmounted (carried forward)"
  - id: IN-06
    severity: info
    disposition: open
    title: "A failed initial Providers() call leaves the panel blank for good (carried forward)"
  - id: IN-07
    severity: info
    disposition: open
    title: "The hide-after-copy timer is never cancelled (carried forward)"
  - id: IN-08
    severity: info
    disposition: open
    title: "Clips Try Again does nothing after the one re-probe (product choice, already surfaced)"
  - id: IN-09
    severity: info
    disposition: open
    title: "Klipy ads are shown without a label, and some of them cannot be copied (product choice, already surfaced)"
  - id: IN-10
    severity: info
    disposition: open
    title: "\"Not approved for Clips\" can still be inferred from a GIFs success that no longer proves anything (new)"
  - id: IN-11
    severity: info
    disposition: open
    title: "`search.Service.Blocked` is now used only by tests (new)"
  - id: WR-01
    severity: warning
    disposition: fixed
    title: "\"Go to GIPHY\" from the Clips rejected-key message usually lands on a GIPHY grid with no key field"
  - id: WR-02
    severity: warning
    disposition: fixed
    title: "A Clips re-probe before the launch's first GIPHY success leaves Clips unavailable for the rest of the launch"
  - id: WR-03
    severity: warning
    disposition: fixed
    title: "After the re-probe is spent, Try Again replays a stale rate-limit error and restarts its countdown"
  - id: CR-01
    severity: critical
    disposition: fixed
    title: "GIPHY and Clips share one key and one rate limit, but the 429 block is per provider ID, and the Clips probe bypasses it"
  - id: CR-02
    severity: critical
    disposition: fixed
    title: "A revoked GIPHY key on the Clips tab renders a blank content area with no recovery"
  - id: WR-04
    severity: warning
    disposition: fixed
    title: "The Klipy mapper drops ad items despite the no-filter rule, and a page of only ads keeps the grid fetching"
open: 11
total: 17
recorded: 2026-10-17T22:03:41.966Z
---

# Phase 02: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| IN-01 | info | open | - |
| IN-02 | info | open | - |
| IN-03 | info | open | - |
| IN-04 | info | open | - |
| IN-05 | info | open | - |
| IN-06 | info | open | - |
| IN-07 | info | open | - |
| IN-08 | info | open | - |
| IN-09 | info | open | - |
| IN-10 | info | open | - |
| IN-11 | info | open | - |
| WR-01 | warning | fixed | 02-REVIEW-FIX.md (not in the current review) |
| WR-02 | warning | fixed | 02-REVIEW-FIX.md (not in the current review) |
| WR-03 | warning | fixed | 02-REVIEW-FIX.md (not in the current review) |
| CR-01 | critical | fixed | 02-REVIEW-FIX.iter2.md (not in the current review) |
| CR-02 | critical | fixed | 02-REVIEW-FIX.iter2.md (not in the current review) |
| WR-04 | warning | fixed | 02-REVIEW-FIX.iter2.md (not in the current review) |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved. A `|` in the reason is kept as prose and escaped on the next run.
Re-running the gate keeps every row it can. A row the current review no longer reports is kept and its Source cell flagged, so a finding does not leave this record silently. ONE exception: when a finding id is REUSED by a different finding, the earlier decision cannot keep a row — the id is taken — and it is dropped. A RECORDED decision (anything but `open`) is named on the console when that happens; a row still at `open` is replaced silently, because `open` records no decision to lose.
