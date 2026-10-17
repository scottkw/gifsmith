---
schema_version: 1
open_count: 1
waived_count: 0
fixed_count: 0
total_count: 1
last_updated: 2026-10-17T14:25:46.210Z
---

# Broken Windows Ledger

> Cross-phase defect register. With `workflow.windows_enforce` enabled, `/gsd-ship` blocks while `open_count > 0`.
> Waive with `gsd-tools windows waive <id> "<reason>"` (reason required).
> Mark fixed with `gsd-tools windows fixed <id>`.

| id | phase | kind | file | line | description | status | reason | recorded_at | resolved_at |
|----|-------|------|------|------|-------------|--------|--------|-------------|-------------|
| 1 | 1 | stub | frontend/src/App.svelte |  | Panel always shows and copies the fixed sample.gif; input does not search yet (intentional Phase 1 spike, D-03) | open |  | 2026-10-17T14:25:46.210Z |  |

````json
[
  {
    "id": 1,
    "kind": "stub",
    "phase": "1",
    "file": "frontend/src/App.svelte",
    "line": null,
    "description": "Panel always shows and copies the fixed sample.gif; input does not search yet (intentional Phase 1 spike, D-03)",
    "status": "open",
    "reason": "",
    "recorded_at": "2026-10-17T14:25:46.210Z",
    "resolved_at": null,
    "milestone": "v1.0"
  }
]
````
