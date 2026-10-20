# Phase 4 deferred items

## scripts/uat/18-backstops.js intermittent page error (found in 04-06 Task 3, 2026-10-20)

- The full-suite re-run failed once with `UAT FAIL: page error: A: CancelledRejectionError: Unhandled rejection in cancelled promise. Reason: RuntimeError: context canceled`. Three immediate re-runs of the same script passed (`UAT OK`).
- Where: section (d) of the script, page A (the panel). The script searches "dwight" on Finer Gifs and then switches to GIPHY with ⌘1. The likely cause is an in-flight search binding call that `cancel(...)` in `frontend/src/App.svelte` (first:/page: keys) cancels on the tab switch. Its rejection reaches the page as unhandled.
- Impact: none visible to the user, since an unhandled rejection only shows in the webview console. It makes the UAT flaky about 1 run in 4.
- Owner: the search/cancel path (Phase 2, App.svelte). Fix: swallow `CancelledRejectionError` where the search promise is cancelled. Do not loosen the UAT's page-error check.
  status: acknowledged
