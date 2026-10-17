---
phase: 01-native-spike-signed-dev-loop
fixed_at: 2026-10-17T00:00:00Z
review_path: .planning/phases/01-native-spike-signed-dev-loop/01-REVIEW.md
iteration: 1
findings_in_scope: 7
fixed: 7
skipped: 0
status: all_fixed
---

# Phase 1: Code Review Fix Report

**Fixed at:** 2026-10-17
**Source review:** .planning/phases/01-native-spike-signed-dev-loop/01-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 7 (CR-01, WR-01..WR-06; Info out of scope)
- Fixed: 7
- Skipped: 0

## Verification

All checks ran in the **main checkout** (`workflow.use_worktrees: false`, no worktree), on branch `gsd/phase-01-native-spike-signed-dev-loop`, after every fix and again at the end:
- `gofmt -l main.go internal`: no output
- `go vet . ./internal/...`: pass
- `go test ./internal/... -count=1`: pass (also `-race` after WR-05)
- `pnpm --dir frontend run check`: 0 errors, 0 warnings
- `build/darwin/Taskfile.yml`: YAML parse checked with Ruby; the wait loop was run in `sh` against a process name that does not exist (exits 0)

Not run, as instructed: `wails3 task darwin:install:dev`, any synthetic keystroke, any real ⌘V. The native behaviour of CR-01, WR-01 and WR-02 still needs a manual check on the installed build.

## Fixed Issues

### CR-01: ⌘V is posted into whatever app is frontmost when reactivation is declined or the remembered app is stale

**Files modified:** `internal/services/actions.go`, `internal/shell/shell.go`, `internal/platform/fake.go`, `internal/services/actions_test.go`
**Commit:** 91697c0
**Status:** fixed: requires human verification (logic change in the paste path)
**Applied fix:** After `ActivateApp` and the paste delay, `Paste` checks `FrontmostApp` again. If the lookup fails or the PID still differs, it returns `Result{Copied: true, Message: "Copied — press ⌘V to paste (could not return to <app>)"}` and posts no ⌘V. `Paste` now reads `Previous()` once. `Shell.Show` sets `prev` to an empty `platform.App{}` when the frontmost lookup fails. `platform.Fake` gained `ActivateSetsFrontmost` (when it is false, activation is declined). The new test `TestPasteDeclinedReactivationCopiesOnly` asserts that `PasteShortcut` is never called. `TestPasteReactivatesPrevious` now covers the successful-activation path, including the second `FrontmostApp` call.

### WR-01: Data race on `Shell.prev`

**Files modified:** `internal/shell/shell.go`
**Commit:** b360a87
**Status:** fixed: requires human verification (concurrency; `Shell` has no unit test because it needs a Wails app)
**Applied fix:** `Shell.mu sync.Mutex` serialises `Toggle`/`Show`/`Hide` (`Toggle` uses an unlocked `show()` helper so it does not lock twice) and guards `Previous()`. The single read of `Previous()` in `Paste` was done in CR-01.

### WR-02: `[NSScreen screens][0]` raises NSRangeException when there are no screens

**Files modified:** `internal/platform/platform_darwin.m`, `internal/shell/shell.go`, `internal/platform/platform.go`
**Commit:** f3fb856
**Applied fix:** Uses `[[NSScreen screens] firstObject] ?: [NSScreen mainScreen]`. If there is no screen or no primary screen, it returns a zero rect. `Shell.show` only calls `SetPosition` when the work area is non-zero, so the panel keeps its current position. The interface comment now documents the zero rect.

### WR-03: A ⌘V failure after Hide is never shown to the user

**Files modified:** `internal/services/actions.go`, `internal/services/actions_test.go`, `frontend/src/App.svelte`
**Commit:** 0c70162
**Applied fix:** A `PasteShortcut` error is now logged, and `Paste` returns `Result{Copied: true, Message: "Copied — automatic paste failed, press ⌘V"}, nil`. Returning the message alone would not have been enough: the panel is hidden and `panel:show` clears `status`. So `App.svelte` sets a `keepStatus` flag for copied-only results (not pasted, and not the accessibility case), and the next `panel:show` keeps the message. The same mechanism shows the CR-01 "could not return to" message. `TestPastePostError` was updated to expect a nil error and a message.

### WR-04: Enter during IME composition triggers Copy and hides the panel

**Files modified:** `frontend/src/App.svelte`
**Commit:** 2438080
**Applied fix:** `if (e.key !== 'Enter' || e.isComposing || e.keyCode === 229 || e.repeat) return;`

### WR-05: Paste and Copy can run at the same time

**Files modified:** `internal/services/actions.go`, `internal/services/actions_test.go`, `frontend/src/App.svelte`
**Commit:** 2a03f48
**Applied fix:** `ActionService` has an unexported `busy sync.Mutex`. `Copy`, `Paste` and `RequestAccessibility` each `TryLock` it and return the new `ErrBusy` if another call is in flight. A blocking lock would still let a queued second Paste post a second ⌘V, so `TryLock` was used instead. The frontend keeps a `busy` flag and ignores Enter while a call is pending (cleared in `finally`). The new test `TestActionsRefuseWhileBusy` covers all three methods.

### WR-06: `install:dev` replaces the bundle and runs `open` while the old instance may still be shutting down

**Files modified:** `build/darwin/Taskfile.yml`
**Commit:** 02df853
**Applied fix:** After `pkill -x {{.APP_NAME}} || true`, the task polls `pgrep -x` every 0.1 s for up to 5 s. If the old process is still running after that, the task fails before `rm -rf`/`ditto`/`open`. It still kills only the exact `gifsmith` process name. The optional `pkill -f` bundle-path targeting was not adopted, per the project constraint.

---

_Fixed: 2026-10-17_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
