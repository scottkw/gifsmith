---
phase: 01-native-spike-signed-dev-loop
reviewed: 2026-10-17T16:53:24Z
depth: standard
iteration: 2
files_reviewed: 23
files_reviewed_list:
  - main.go
  - internal/platform/platform.go
  - internal/platform/platform_darwin.go
  - internal/platform/platform_darwin.h
  - internal/platform/platform_darwin.m
  - internal/platform/platform_darwin_test.go
  - internal/platform/platform_stub.go
  - internal/platform/fake.go
  - internal/services/actions.go
  - internal/services/actions_test.go
  - internal/shell/shell.go
  - internal/shell/position.go
  - internal/shell/position_test.go
  - frontend/src/App.svelte
  - frontend/vite.config.ts
  - frontend/index.html
  - build/darwin/Taskfile.yml
  - build/darwin/Info.plist
  - build/darwin/Info.dev.plist
  - build/config.yml
  - Taskfile.yml
  - scripts/smoke-browser-loop.sh
  - .gitignore
findings:
  critical: 0
  warning: 1
  info: 15
  total: 16
status: issues_found
---

# Phase 1: Code Review Report (re-review, iteration 2)

**Reviewed:** 2026-10-17T16:53:24Z
**Depth:** standard
**Files Reviewed:** 23
**Status:** issues_found

## Summary

This is a re-review after fix commits 91697c0..02df853 (CR-01, WR-01..WR-06). The prior report is in `01-REVIEW.iter2.md`. Only 8 files changed (`git diff e212644..02df853`). I read all of them in full and re-checked the rest against the prior Info findings. `go vet ./...` and `go test -race -count=1 ./...` under `internal/` both pass.

### Prior Critical and Warning findings

| ID | Status | Evidence |
|----|--------|----------|
| CR-01 | **Resolved** for the case it describes (declined activation). | `actions.go:78-83` checks `FrontmostApp` again after `ActivateApp` and the delay, and returns copied-only on mismatch or error. `shell.go:96-101` clears `prev` when the lookup fails. `TestPasteDeclinedReactivationCopiesOnly` asserts that no `PasteShortcut` call is made. Two fail-open paths remain next to this fix; see WR-01. |
| WR-01 | **Resolved** | `Shell.mu` guards `prev` and serialises `Toggle`/`Show`/`Hide`. `Toggle` calls the unlocked `show()`, so it never locks twice. `Paste` reads `Previous()` once (`actions.go:72`) and reuses it for the log line (`:91`). |
| WR-02 | **Resolved** | `platform_darwin.m:91-92` uses `firstObject ?: mainScreen` and returns a zero rect when there is no screen. `shell.go:102` skips `SetPosition` for a zero-size rect. |
| WR-03 | **Resolved** (see IN-14 for a remaining edge case) | `actions.go:86-90` returns a non-error copied-only result. `App.svelte:38` sets `keepStatus`, so the message survives the next `panel:show`. |
| WR-04 | **Resolved** | `App.svelte:29` ignores `isComposing`, `keyCode 229` and `repeat`. |
| WR-05 | **Resolved** | `ActionService.busy.TryLock()` at the start of `Copy`, `Paste` and `RequestAccessibility`, plus the frontend `busy` flag. `TestActionsRefuseWhileBusy` covers all three methods. |
| WR-06 | **Resolved** (see IN-15 for a minor fail-open case) | `build/darwin/Taskfile.yml:220-226` polls `pgrep -x` for up to 5 s and fails before `rm -rf`/`ditto`/`open`. In Task, each `cmds` entry runs in its own shell, so `exit 0` ends only the wait step. |

### Possible regressions

**Deadlock risk from `Shell.mu` held across `panel.Show/Hide/Focus/SetPosition/IsVisible`, `onMain` and `Event.Emit`: none found.** I checked the Wails v3.0.0-beta.30 source:
- **Nothing that takes `s.mu` runs on the main thread.**
  - Hotkey callbacks run on their own goroutine (`global_shortcut_manager.go:290-293`).
  - Menu clicks run on their own goroutine (`application.go:736`).
  - Bound calls run on their own goroutine (`messageprocessor_call.go:139`).
  - `HideOnEscape` and `HideOnFocusLost` call `WebviewWindow.Hide()` directly, not `Shell`.
- **Nothing on the main thread waits on a goroutine that may hold `s.mu`.**
  - `InvokeSync` only queues work onto the main thread and waits for it (`mainthread.go:23-36`).
  - `Event.Emit` does not block: listeners run on goroutines and frontend delivery goes through an unbounded mailbox (`events.go:167-171`).
  - The native `windowEvents` channel is drained into per-event goroutines (`application.go:704-705`), so the main thread cannot stall on it.
- **The only lock order is `ActionService.busy` → `Shell.mu`** (`Paste`/`Copy` call `Hide`/`Previous`). `Shell` never takes `busy`, so the two locks cannot form a cycle.

**`TryLock` and `ErrBusy`: correct.**
- The frontend `busy` flag already prevents a second call from the same keyboard handler.
- `ErrBusy` can only be reached when the "Enable automatic paste…" button and ⌘Enter overlap. The frontend then shows it as `RuntimeError: another action is still running`, which is acceptable.
- `defer Unlock` runs before Wails sends the reply, so the frontend flag is cleared after the Go lock is released, never before it.

**`keepStatus`: correct for the sequential flow.** One ordering race remains; see IN-14.

**`install:dev` wait loop: correct.** One minor fail-open case remains; see IN-15.

### Open findings

- **One new Warning (WR-01).** Two paths still post ⌘V without confirming the target. They sit next to the CR-01 fix and contradict its comment, "never post ⌘V into an app the user did not pick".
- **All 12 Info findings from the prior review still apply.** They were out of scope for the fixer. IN-07 is partly addressed.
- **Three new Info findings (IN-13..IN-15).**

## Warnings

### WR-01: `Paste` still posts ⌘V without a confirmed target when the remembered app is unknown or the frontmost lookup fails

**File:** `internal/services/actions.go:72-86`, `internal/shell/shell.go:96-101`
**Issue:** The CR-01 fix only guards the branch where both lookups succeed and the PIDs differ. Two paths skip the check and go straight to `PasteShortcut()` at line 86:
1. **`prev.PID == 0`.** The fix made this more likely, not less. `Shell.show` now deliberately sets `prev = platform.App{}` when `FrontmostApp` fails, with the comment "never paste into a target left over from an earlier show". `Paste` then reads that as "no target, skip the check" (`if prev.PID != 0 { … }`) and posts ⌘V into whatever is frontmost. A stale target has become an unknown one, and ⌘V is still posted.
2. **`FrontmostApp` fails at paste time.** The condition `err == nil && front.PID != prev.PID` (line 74) is false when `err != nil`, so the block is skipped and ⌘V is posted. Line 80 correctly treats `err != nil` as "do not post", so the two checks disagree.

Both lookups fail only when `NSWorkspace.frontmostApplication` is nil, which happens briefly during app switches and at the lock or login window. The panel does not activate Gifsmith, so the frontmost app is usually still the user's app. Each path is therefore unlikely, but both are fail-open on the one action that can type into the wrong window. Neither is tested: `platform.Fake.FrontmostApp` cannot return an error (see IN-07).
**Fix:** Fail closed in both cases, using the same copied-only result:
```go
prev := a.Previous()
front, err := a.Platform.FrontmostApp()
if prev.PID == 0 || err != nil {
    log.Printf("paste: no confirmed target (prev=%+v, err=%v), copied only", prev, err)
    return Result{Copied: true, Message: "Copied — press ⌘V to paste"}, nil
}
if front.PID != prev.PID {
    if err := a.Platform.ActivateApp(prev.PID); err != nil {
        log.Printf("reactivate %s: %v", prev.Name, err)
    }
    sleep(PasteDelay)
    if now, err := a.Platform.FrontmostApp(); err != nil || now.PID != prev.PID {
        return Result{Copied: true, Message: "Copied — press ⌘V to paste (could not return to " + prev.Name + ")"}, nil
    }
}
```
Add `FrontmostErr error` to `platform.Fake`. Then add tests asserting that neither `Previous()` returning `App{}` nor a `FrontmostApp` error leads to a `PasteShortcut` call.

## Info

### IN-01: `kUCKeyTranslateNoDeadKeysBit` is passed where a mask is expected (carried over)

**File:** `internal/platform/platform_darwin.m:127`
**Issue:** `kUCKeyTranslateNoDeadKeysBit` is 0, so no option is set. It does no harm here because `dead` is reset on every iteration.
**Fix:** Use `kUCKeyTranslateNoDeadKeysMask`.

### IN-02: `onMain` waits forever (carried over)

**File:** `internal/platform/platform_darwin.m:13-21`
**Issue:** `DISPATCH_TIME_FOREVER` hangs any caller when the main run loop is not running (under `go test`, before `app.Run`, or during shutdown). `Shell.show` now calls `FrontmostApp`, `CursorScreenWorkArea` and `FocusWebView` while holding `Shell.mu`. A hang there would also block every later `Hide`/`Previous` call, and with it `Paste` (which holds `busy`). This is not reachable while the app runs normally.
**Fix:** Document this in the `Platform` comments, or use a bounded wait that returns a failure code.

### IN-03: `clearContents` runs before a write that may fail (carried over)

**File:** `internal/platform/platform_darwin.m:35-36`
**Issue:** If `writeObjects:` fails, the previous clipboard contents are already gone.
**Fix:** Accept and document it, or build the URL before clearing.

### IN-04: `[NSString stringWithUTF8String:]` can return nil, and the next call throws (carried over)

**File:** `internal/platform/platform_darwin.m:28, 36`
**Issue:** Invalid UTF-8 makes the conversion return nil, and `fileURLWithPath:nil` then throws and aborts the process. This cannot happen with today's fixed `SamplePath`.
**Fix:** Return 0 when the converted string is nil.

### IN-05: The not-trusted message tells the user to press ⌘V while Gifsmith's own input has focus (carried over)

**File:** `internal/services/actions.go:64`
**Issue:** The panel stays open and focused, so pressing ⌘V straight away pastes into the search field.
**Fix:** Change the wording to "press Esc, then ⌘V".

### IN-06: `enableAutoPaste` ignores the result of `RequestAccessibility` (carried over)

**File:** `frontend/src/App.svelte:51-54`
**Issue:** When it returns `true` (already trusted), the UI still says "After turning Gifsmith on…".
**Fix:** Branch on the returned boolean.

### IN-07: The Fake still cannot simulate a `FrontmostApp` failure (carried over, partly addressed)

**File:** `internal/platform/fake.go:40-51`
**Issue:**
- `ActivateSetsFrontmost` now covers declined activation.
- `FrontmostApp` still always returns `nil`, so the `err != nil` branch at `actions.go:80` and both WR-01 paths are untested.
- `ActivateApp` writes `f.Frontmost` without `f.mu`, while `f.mu` guards only `Calls`. This is harmless in today's single-goroutine tests.
**Fix:** Add `FrontmostErr error`. Either guard `Frontmost` with `f.mu`, or document that the Fake is not safe for concurrent use.

### IN-08: The cache path repeats the bundle id, dev and installed builds share it, and there is no single-instance guard (carried over)

**File:** `main.go:31, 47-54`
**Issue:**
- `"com.kenscott.gifsmith"` is hard-coded, and `.dev.app` shares the same staging file.
- There is no `SingleInstance` option.
- `log.Fatal` kills an LSUIElement app with no visible signal to the user.
**Fix:** Add a constant for the identifier, and consider `application.Options.SingleInstance`.

### IN-09: The server-mode endpoint has no authentication (carried over)

**File:** `main.go:53`
**Issue:** `-tags server` builds expose every bound method on `localhost:8089`. This is harmless while the stub returns `ErrUnsupported`.
**Fix:** Keep server builds to UAT only. Check the Host and Origin headers before binding methods that have side effects.

### IN-10: The smoke script's port checks are racy and use inconsistent hosts (carried over)

**File:** `scripts/smoke-browser-loop.sh:14-20, 40, 52`
**Issue:**
- The `lsof` check and the later bind are separate steps, so another process can take the port in between.
- A missing `lsof` makes both ports look free.
- The health check targets `127.0.0.1`, while the proxy target is `localhost`.
**Fix:** Use `127.0.0.1` in both places, and fail clearly when `lsof` is missing.

### IN-11: Template placeholders remain in the bundle metadata (carried over)

**File:** `build/darwin/Info.plist:7-8, 27-28`, `build/darwin/Info.dev.plist:7-8, 32-33`, `build/config.yml:13, 85-87`, `frontend/index.html:5`
**Issue:**
- `CFBundleGetInfoString` is still "This is a comment".
- `NSHighResolutionCapable` is a `<string>`.
- `build/config.yml` still has placeholder comments.
- `index.html` declares the wrong MIME type for the icon.
**Fix:** Clean these up before the Phase 4 notarized build.

### IN-12: Bound methods ignore `ctx` (carried over)

**File:** `internal/services/actions.go:39, 54, 97`
**Issue:** `Paste` sleeps for up to 160 ms and cannot be cancelled.
**Fix:** Drop the unused parameter, or have `sleep` honour `ctx.Done()`.

### IN-13: A hotkey press during the paste delay re-shows the panel, and ⌘V then lands in Gifsmith's own search field

**File:** `internal/services/actions.go:70-86`, `internal/shell/shell.go:79-87`
**Issue:**
- `Toggle` is not gated by `ActionService.busy`, and `Paste` does not check whether the panel is visible before posting ⌘V.
- If the user presses the hotkey in the 80-160 ms after ⌘Enter, `show()` makes the non-activating panel key again. The frontmost app is still the target, so both PID checks pass.
- The synthetic ⌘V then goes to the key window, which is Gifsmith's input field. Since WR-01's mutex, `show()` also refreshes `prev`, but that does not change where the keystroke goes.
- The window is narrow, and the result (a file reference pasted into the search box) does no harm.
**Fix:** Immediately before `PasteShortcut()`, skip the post when the panel is visible (copied-only), for example by passing an `IsVisible func() bool` into `ActionService`. Alternatively, have `Toggle` ignore presses while a paste is in flight.

### IN-14: The `keepStatus` message can reappear stale, because `panel:show` and the `Paste` reply are not ordered

**File:** `frontend/src/App.svelte:15-17, 35-38`
**Issue:**
- `panel:show` arrives through the event mailbox, and the `Paste` reply through the call-response path. Nothing orders the two.
- If the panel is re-shown while a Paste is in flight (as in IN-13), `panel:show` may run first, and then the reply sets `status` and `keepStatus = true` while the panel is already visible.
- The next hide and show then shows the copied-only message a second time.
- Separately, the message is kept for an unlimited time. If the user reopens Gifsmith much later, "Copied — press ⌘V to paste" may no longer be true, because the clipboard may have changed.
**Fix:** Store the time with the kept message and drop it after a short window (for example 30 s). Set `keepStatus` only when the panel is hidden, for example by tracking visibility through `panel:show` and a matching hide event.

### IN-15: The `install:dev` wait loop treats a `pgrep` error as "process exited"

**File:** `build/darwin/Taskfile.yml:220-226`
**Issue:**
- `pgrep -x … >/dev/null || exit 0` exits successfully on any non-zero status. That includes 2 and 3 (syntax or fatal error), not only 1 (no match).
- `pkill -x gifsmith` and `pgrep -x gifsmith` also match a running `wails3 dev` or `.dev.app` instance. That instance is killed and waited on, which was noted in the prior WR-06 and left as is by design.
**Fix:** Branch on `$?`, so that only exit status 1 counts as "gone":
```sh
pgrep -x {{.APP_NAME}} >/dev/null; rc=$?
[ $rc -eq 1 ] && exit 0
[ $rc -ne 0 ] && { echo "pgrep failed ($rc)"; exit 1; }
```

---

_Reviewed: 2026-10-17T16:53:24Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
