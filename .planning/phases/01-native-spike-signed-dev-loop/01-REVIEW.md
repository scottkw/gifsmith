---
phase: 01-native-spike-signed-dev-loop
reviewed: 2026-10-17T00:00:00Z
depth: standard
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
  critical: 1
  warning: 6
  info: 12
  total: 19
status: issues_found
---

# Phase 1: Code Review Report

**Reviewed:** 2026-10-17
**Depth:** standard
**Files Reviewed:** 23
**Status:** issues_found

## Summary

I reviewed the Go, cgo/Objective-C, Svelte, build and script files for the native spike. To check threading and coordinate assumptions I read the Wails v3.0.0-beta.30 source in the module cache (`mainthread_darwin.go`, `webview_window.go`, `webview_window_darwin.go`, `global_shortcut_manager.go`, `application.go`, `events.go`).

Things I checked that are correct:
- **cgo memory handling.** Every `C.CString` and `strdup` result is freed exactly once. `CFRelease` is paired with each `TISCopy…` and `CGEvent…Create`. `TISGetInputSourceProperty` returns a value the caller does not own, and the code correctly does not release it. Every Go pointer handed to C is written only while the C call is blocked.
- **onMain.** `CFRunLoopPerformBlock` copies the block, so the captured semaphore is retained, and `dispatch_release` under manual reference counting is valid.
- **TIS calls (the e9f1230 fix).** They now run on the main thread.
- **Panel position.** The coordinate maths in `gsCursorScreenVisibleFrame` matches Wails `windowSetPosition` (top-left origin measured from the primary screen's height).
- **Hide timing.** `WebviewWindow.Hide()` uses `InvokeSync`, so the panel really is ordered out before the 80 ms paste delay begins.

The main problems:
- **The ⌘V path can type into the wrong app.** When the frontmost app is not the remembered one, the code still posts ⌘V after an activation request that macOS 14+ routinely declines. It also does this when the remembered app is stale.
- **Shared state is written and read without locking.** Wails runs each hotkey trigger, menu click and bound-method call on its own goroutine (`global_shortcut_manager.go:290`, `application.go:736`, `messageprocessor_call.go:139`), and nothing synchronises `Shell.prev` or the Paste flow.
- **A few failure paths crash or fail silently.**

## Critical Issues

### CR-01: ⌘V is posted into whatever app is frontmost when reactivation is declined or the remembered app is stale

**File:** `internal/services/actions.go:58-66`, `internal/shell/shell.go:84-88`
**Issue:** When `front.PID != prev.PID`, `Paste` calls `ActivateApp(prev.PID)`, sleeps 80 ms, then calls `PasteShortcut()` without checking anything. The code comment ("best effort; post anyway") and `platform_darwin.m:168` both admit that since macOS 14 activation is cooperative. A request from a background accessory app (Gifsmith is not active after `Hide`) is usually ignored. `activateWithOptions:` returns YES when the request is made, not when it succeeds, so the error branch almost never runs.

As a result, the one case this branch exists for (the frontmost app has changed) is the case where the synthetic ⌘V most likely lands in an app the user did not pick. That could be a chat window, a terminal, or a document.

Separately, `Shell.Show` only updates `s.prev` when `FrontmostApp` succeeds (shell.go:84-88). On failure it keeps the app from an earlier show, so `Paste` can try to reactivate, and paste into, an app from an earlier session.
**Fix:** Check the target again after the activation delay, and if it still does not match, fall back to copy-only. Also clear `prev` when the lookup fails:
```go
// actions.go
if prev := a.Previous(); prev.PID != 0 {
    if front, err := a.Platform.FrontmostApp(); err == nil && front.PID != prev.PID {
        if err := a.Platform.ActivateApp(prev.PID); err != nil {
            log.Printf("reactivate %s: %v", prev.Name, err)
        }
        sleep(PasteDelay)
        if now, err := a.Platform.FrontmostApp(); err != nil || now.PID != prev.PID {
            return Result{Copied: true, Message: "Copied — press ⌘V to paste (could not return to " + prev.Name + ")"}, nil
        }
    }
}

// shell.go Show()
app, err := s.plat.FrontmostApp()
if err != nil {
    log.Printf("frontmost app: %v", err)
    app = platform.App{} // never paste into a stale target
}
s.setPrev(app)
```
Add a Fake test in which `ActivateApp` does not change `Frontmost`, and assert that `PasteShortcut` is not called.

## Warnings

### WR-01: Data race on `Shell.prev` (a value made of several words, read and written on different goroutines)

**File:** `internal/shell/shell.go:24, 87, 100`, `internal/services/actions.go:58, 69`
**Issue:**
- `Show()` writes `s.prev`, a `platform.App` holding two string headers and an int. It runs on a new goroutine for each hotkey press (Wails `global_shortcut_manager.go:290`) or menu click (`application.go:736`).
- `Previous()` reads it from the bound-method goroutine. There is no mutex or atomic.
- The Go memory model treats a racing read of a value made of several words as undefined: a torn string header can point at freed or unrelated memory.
- The window is real. Pressing the hotkey during the 80-160 ms paste delay calls `Show()` while `Paste` reads `Previous()`.
- `Paste` also calls `a.Previous()` twice (lines 58 and 69), so the log line can name a different app from the one that was targeted.
- `Toggle()` checks `IsVisible()` and then acts on the result without holding anything, so two quick hotkey presses (two goroutines) can both run `Show` or interleave a Show with a Hide.

`go test -race` does not catch any of this because no test exercises `Shell`.
**Fix:** Guard `prev` with a `sync.Mutex`, or store it in an `atomic.Pointer[platform.App]`. Serialise `Toggle`/`Show`/`Hide` with a mutex held by `Shell`. In `Paste`, read `Previous()` once into a local and reuse it for the log line.

### WR-02: `[NSScreen screens][0]` raises NSRangeException when there are no screens, which aborts the app

**File:** `internal/platform/platform_darwin.m:91`
**Issue:** `[NSScreen screens]` can be empty: during display reconfiguration, with the lid closed and no external display, or on a headless machine. Indexing an empty NSArray throws an Objective-C exception. It is thrown inside a `CFRunLoopPerformBlock` block that nothing catches, so the process aborts on the next hotkey press. `[NSScreen mainScreen]` (line 86) can also be nil. Wails' own `windowSetPosition` guards against this with `firstObject` plus a fallback.
**Fix:**
```objc
NSScreen *primary = [[NSScreen screens] firstObject] ?: [NSScreen mainScreen];
if (screen == nil || primary == nil) { *x = *y = *w = *h = 0; return; }
CGFloat primaryHeight = primary.frame.size.height;
```
The Go side can then treat a zero-size rect as "keep the current position".

### WR-03: A ⌘V failure after Hide is never shown to the user

**File:** `internal/services/actions.go:56, 66-68`, `frontend/src/App.svelte:34-36`
**Issue:** `Paste` hides the panel at line 56. If `PasteShortcut()` then fails, it returns `(Result{Copied: true}, err)`. Wails throws away the result value when an error is returned. The frontend writes the error into `status` in a panel that is already hidden, and the next `panel:show` clears `status` (App.svelte:13). So the user gets no paste and no message, even though the GIF is on the clipboard.
**Fix:** Return a non-error result with an explanation, for example `return Result{Copied: true, Message: "Copied — automatic paste failed, press ⌘V"}, nil`, and log the error. Alternatively, emit a `panel:status` event that persists and is shown on the next show.

### WR-04: Enter during IME composition triggers Copy and hides the panel

**File:** `frontend/src/App.svelte:23-25`
**Issue:** `onKeydown` acts on any `e.key === 'Enter'`. With Japanese, Chinese or Korean input methods, Enter commits the composition and arrives as a keydown with `isComposing === true` (WebKit reports `keyCode === 229`). The user's first Enter while typing a query therefore replaces the clipboard and hides the panel. `e.repeat` is not checked either.
**Fix:**
```ts
if (e.key !== 'Enter' || e.isComposing || e.keyCode === 229 || e.repeat) return;
```

### WR-05: Paste and Copy can run at the same time: two ⌘V events and racing pasteboard writes

**File:** `internal/services/actions.go:33-71`, `frontend/src/App.svelte:23-37`
**Issue:**
- Each bound call runs on its own goroutine, and neither the service nor the frontend has an in-flight guard.
- Two Enter or ⌘Enter keydowns that arrive before `Hide` takes effect, or a Copy that overlaps a Paste, run `CopyFile` at the same time. Each one does `clearContents` + `writeObjects` on the shared `generalPasteboard` instance from a different thread.
- Two Paste calls each post a ⌘V, so the GIF is pasted twice into the target app.
**Fix:** Add a `sync.Mutex` (or a try-lock that returns early when busy) to `ActionService`, held for the whole of `Copy`, `Paste` and `RequestAccessibility`. In the frontend, also keep a `busy` flag that ignores Enter while a call is pending.

### WR-06: `install:dev` replaces the bundle and runs `open` while the old instance may still be shutting down

**File:** `build/darwin/Taskfile.yml:218-223`
**Issue:** `pkill -x gifsmith` sends SIGTERM and returns immediately. The task then runs `rm -rf` and `ditto` on the bundle and calls `open` on it. If the old process has not exited yet, LaunchServices can activate the running instance (same bundle id) instead of launching the new binary. The developer then tests a stale build that has the old code identity, which is exactly what the D-10 signed loop exists to prevent. `pkill -x gifsmith` also kills a running `wails3 dev` instance, because the executable name is the same.
**Fix:**
```yaml
- pkill -x {{.APP_NAME}} || true
- for i in $(seq 1 50); do pgrep -x {{.APP_NAME}} >/dev/null || exit 0; sleep 0.1; done; echo "gifsmith did not exit"; exit 1
```
Optionally target only the installed bundle, using `pkill -f "$HOME/Applications/Gifsmith.app/"`.

## Info

### IN-01: `kUCKeyTranslateNoDeadKeysBit` is passed where a mask is expected

**File:** `internal/platform/platform_darwin.m:124`
**Issue:** `UCKeyTranslate`'s `keyTranslateOptions` takes a bit mask. `kUCKeyTranslateNoDeadKeysBit` is 0, so no options are actually set. The correct constant is `kUCKeyTranslateNoDeadKeysMask`. It does no harm here because `dead` is reset on every iteration, but the code does not do what it claims. The Wails code it was copied from has the same mistake (`global_shortcut_darwin.go:68`).
**Fix:** Use `kUCKeyTranslateNoDeadKeysMask`.

### IN-02: `onMain` waits forever

**File:** `internal/platform/platform_darwin.m:13-21`
**Issue:** `DISPATCH_TIME_FOREVER` means any caller blocks permanently if the main run loop is not running. That happens under `go test` (the comment at lines 24-25 already notes it), before `app.Run`, or after the run loop exits during shutdown. A future test that calls `FrontmostApp`, `CursorScreenWorkArea` or `PasteShortcut` will hang rather than fail.
**Fix:** Document this in the `Platform` interface comments. Optionally use a bounded wait (for example 2 s) that returns a failure code, so the hang becomes an error.

### IN-03: `clearContents` runs before a write that may fail

**File:** `internal/platform/platform_darwin.m:35-36`
**Issue:** If `writeObjects:` returns NO, the user's previous clipboard contents are already gone. This is acceptable for a spike, but it is data loss on the failure path.
**Fix:** Accept and document it, or check that `fileURLWithPath:` produced a URL before clearing.

### IN-04: `[NSString stringWithUTF8String:]` can return nil, and the next call throws

**File:** `internal/platform/platform_darwin.m:28, 36`
**Issue:** Invalid UTF-8 makes `stringWithUTF8String:` return nil, and `fileURLWithPath:nil` then throws NSInvalidArgumentException, which aborts the process. This cannot happen with today's fixed `SamplePath`, but it will matter once arbitrary cache paths are used.
**Fix:** Return 0 when the converted string is nil.

### IN-05: The not-trusted message tells the user to press ⌘V while Gifsmith's own input still has focus

**File:** `internal/services/actions.go:50`
**Issue:** The panel stays open and the search input is focused, so pressing ⌘V straight away pastes into Gifsmith's input field, not the target app.
**Fix:** Change the wording to "press Esc, then ⌘V", or hide the panel on that path.

### IN-06: `enableAutoPaste` ignores the result of `RequestAccessibility`

**File:** `frontend/src/App.svelte:39-44`
**Issue:** When it returns `true` (already trusted), the UI still says "After turning Gifsmith on…".
**Fix:** Branch on the returned boolean.

### IN-07: The Fake cannot simulate a `FrontmostApp` failure or a declined activation

**File:** `internal/platform/fake.go:38-46`
**Issue:** `FrontmostApp` always returns nil and `ActivateApp` never changes `Frontmost`, so the paths behind CR-01 cannot be tested.
**Fix:** Add `FrontmostErr error`, and an optional `ActivateSetsFrontmost bool`.

### IN-08: The cache path repeats the bundle id, and dev and installed builds share it

**File:** `main.go:31`
**Issue:**
- `"com.kenscott.gifsmith"` is hard-coded, and `.dev.app` (`com.kenscott.gifsmith.dev`) writes to the same staging file.
- There is no `SingleInstance` option, so two instances can run at once. Both register the same Carbon hotkey (duplicates are allowed), so both panels toggle together.
- `log.Fatal` here kills an app that has no Dock icon (LSUIElement), with no visible signal to the user.
**Fix:** Add a constant for the identifier, and consider `application.Options.SingleInstance`.

### IN-09: The server-mode endpoint has no authentication

**File:** `main.go:53`
**Issue:** `-tags server` builds expose every bound method on `localhost:8089` with no authentication. Phase 1 is harmless because the stub returns `ErrUnsupported`. Once real service methods (keys, settings) are bound, any local process, or a browser page reaching it through DNS rebinding, could call them.
**Fix:** Keep server builds to UAT only. Add a check of the Host and Origin headers before binding methods that have side effects.

### IN-10: The smoke script's port checks are racy and use inconsistent hosts

**File:** `scripts/smoke-browser-loop.sh:14-20, 40, 52`
**Issue:**
- The `lsof` check and the later bind are separate steps, so another process can take the port in between.
- If `lsof` is missing, both ports are treated as free.
- The health check targets `127.0.0.1`, but the Vite proxy target is `localhost` (Node may try `::1` first).
- It works today. If the setup changes, the failure modes become confusing.
**Fix:** Use `127.0.0.1` for both `GIFSMITH_BACKEND` and the health check. Fail with a clear message if `command -v lsof` fails.

### IN-11: Template placeholders remain in the bundle metadata

**File:** `build/darwin/Info.plist:7-8, 27-28`, `build/darwin/Info.dev.plist:7-8, 32-33`, `build/config.yml:13, 85-87`, `frontend/index.html:5`
**Issue:** These still carry template values:
- `CFBundleGetInfoString` = "This is a comment" in both plists.
- `NSHighResolutionCapable` is a `<string>`, where `<true/>` is expected.
- `comments: "Some Product Comments"` and `other: My Other Data` in `build/config.yml`.
- `index.html` declares `type="image/svg+xml"` for a `.png` icon.

The bundle id is already correct, but these values will ship in the signed bundle.
**Fix:** Replace or remove them before the Phase 4 notarized build.

### IN-12: Bound methods ignore `ctx`

**File:** `internal/services/actions.go:33, 44, 75`
**Issue:** `Paste` sleeps up to 160 ms and cannot be cancelled. The project's Go convention asks for context-aware functions.
**Fix:** Drop the unused parameter, or have `sleep` honour `ctx.Done()`.

---

_Reviewed: 2026-10-17_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
