# Phase 1: Native Spike & Signed Dev Loop - Research

**Researched:** 2026-10-17
**Domain:** Wails v3 (v3.0.0-beta.30) macOS shell: NSPanel, Carbon hotkey, tray, cgo AppKit/CoreGraphics glue, Developer ID signing and notarization
**Confidence:** HIGH for APIs, scaffold and build (read at the tag and exercised in a scratch probe). MEDIUM for panel typing (one probe attempt was inconclusive, see Pitfall 1). Paste, full-screen overlay and notarization are not run yet. The user verifies those by hand.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Spike Scope
- Spike code is the foundation, not throwaway: the Wails v3 scaffold, `internal/shell` (panel/hotkey/tray) and `internal/platform` (cgo: file-URL pasteboard write, CGEvent ⌘V, Accessibility check, frontmost app) become the real packages used from Phase 2 on.
- Enter = copy GIF file to clipboard; ⌘Enter = paste into the previous app (matches v1 default + sibling actions).
- Uses one small sample GIF bundled/embedded in the app — no provider or API key in Phase 1.
- Panel UI is minimal: a plain text input plus one GIF preview, no styling. Real UI is Phase 2.

#### Hotkey & Panel
- Default global hotkey ⌃⌥⌘G via Wails v3 `app.GlobalShortcut` (Carbon), registered after the app is running so conflicts surface; user confirms no clash with Vorssaint.
- Panel opens on the display under the mouse cursor, horizontally centered, near the top (Spotlight-like).
- Fixed panel size 720×480.
- If a non-activating panel cannot take typed input without stealing focus, fall back to record-frontmost-app → activate panel → reactivate previous app before pasting (`platform.ActivateApp`), rather than redesigning.

#### Signing & Verification
- Bundle id `com.kenscott.gifsmith`.
- Signed dev builds are installed to and run from `~/Applications/Gifsmith.app`, signed with `Developer ID Application: Ken Scott (S2K7P43927)`, so the Accessibility (TCC) grant survives rebuilds. Paste is never tested from `wails3 dev`.
- Notarization uses a notarytool keychain profile named `gifsmith-notary`, created once by the user (`xcrun notarytool store-credentials gifsmith-notary`, app-specific password). This is a user action checkpoint; Claude uses the profile afterwards.
- Paste matrix is a markdown checklist in this phase directory; user tests the apps they have (Slack, Messages, Mail, Discord, Notes, Telegram, Safari/Chrome composers, TextEdit), marking missing apps n/a. File-URL-only pasteboard payload unless the matrix proves otherwise.

#### Carried from project decisions
- Wails v3 `v3.0.0-beta.30` pinned exactly (go.mod, `wails3` CLI, `@wailsio/runtime`); pnpm from the first command; Svelte 5 template.
- macOS 13+ minimum, arm64 only; `LSUIElement=true`; accessory activation policy; no Dock icon.
- Do NOT use `golang.design/x/hotkey` (needs Accessibility on macOS now).
- cgo isolated in `internal/platform` with `//go:build darwin && !server` plus a stub, so `-tags server` builds for browser UAT.
- Copy must work without Accessibility; paste gates on `AXIsProcessTrusted` and degrades to "copied — press ⌘V" with how-to-grant guidance.
- ⌘V synthesized from a private CGEventSource with explicit Command flag and layout-aware V keycode; hide panel and wait ~60–100 ms before posting (tune in spike).

### Claude's Discretion
- Exact package/file layout within the agreed `internal/shell` / `internal/platform` split.
- Paste delay tuning and smoke-script form.
- Whether to prove the `-tags server` + Vite proxy browser→bindings loop here (roadmap expects it; keep it small).

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| SHEL-01 | Open/close floating panel with global hotkey (⌃⌥⌘G) | `app.GlobalShortcut.Register("Ctrl+Option+Cmd+G", toggle)` inside an `ApplicationStarted` hook. Probe: the synthetic hotkey showed the panel, and a second press hid it (Pattern 2) |
| SHEL-02 | Over full-screen apps, takes typing immediately, doesn't steal focus | NSPanel + `NonActivating` + `CanJoinAllSpaces\|FullScreenAuxiliary` + `Floating`. Probe: the frontmost app stayed the same after show. **`Show()` alone does not make the panel key.** Call `Focus()`, then `platform.FocusWebView` (Pitfall 1) |
| SHEL-03 | Closes on Esc and click-away | `HideOnEscape` (a Wails key binding, handled before the webview sees the key) and `HideOnFocusLost` (fires on `WindowDidResignKey`) |
| SHEL-04 | Menu bar only; menu Open/Settings/Quit | `ActivationPolicyAccessory` + `LSUIElement` in **both** plists. `app.SystemTray.New()` + `SetTemplateIcon` + `SetMenu` (Pattern 3) |
| ACTN-01 | Copy GIF as a file, no Accessibility needed | cgo `[NSPasteboard writeObjects:@[NSURL fileURLWithPath:]]` on a staged copy of the embedded sample (Pattern 4) |
| ACTN-02 | Paste into previous app; degrade without AX | Gate on `AXIsProcessTrustedWithOptions`. Hide the panel, sleep about 80 ms, then post ⌘V from a private CGEventSource with a layout-aware keycode (Pattern 4) |
</phase_requirements>

## Project Constraints (from CLAUDE.md)

- Node: use **pnpm** (`/Users/ken/dev/CLAUDE.md` Tech Stack). Go: `go fmt`, `golangci-lint`, context-aware functions (`ctx context.Context` first param on bound methods).
- TS: camelCase, PascalCase components, TypeScript types. ESLint and Prettier are configured globally but not in the Wails template. Phase 1 can rely on `svelte-check`.
- 80%+ coverage in critical components. Test the pure Go parts (action sequencing with a fake platform, panel-position math). cgo and AppKit are verified by hand.
- "Silent Fallbacks: `or {}` … Let it crash." Platform calls return errors. They do not swallow them. Example: `CopyFile` returns an error when `writeObjects:` returns NO.
- Chesterton's Fence and RULE 0: if the native behavior surprises you, stop and report. Do not pile on workarounds.
- **Never `kill node`**. Kill only `gifsmith` by exact name (`pkill -x gifsmith`).
- UAT in the browser uses dev-browser (the `-tags server` + Vite loop). Native behaviors are manual checkpoints.
- Ponytail is active for executors: build the smallest code that proves each behavior. No speculative abstractions. A `Platform` interface is justified because the stub (server build) and the fake (unit tests) are two real implementations beyond darwin.
- GSD workflow: edits go through `/gsd-execute-phase`. `.planning/config.json` has `commit_docs: true` and `security_enforcement: true` (ASVS L1). `nyquist_validation: false`, so this file has no Validation Architecture section.

## Summary

Wails v3 beta.30 provides everything outside the four cgo primitives. I read each API at the `v3.0.0-beta.30` tag (commit `623b529`, 2026-10-15) and compiled it in a scratch probe on this Mac (Go 1.27.2, macOS 27.0.1). The probe scaffolded with `wails3 init -t svelte` and switched to pnpm. It added a `platform` package (darwin cgo plus a server stub), a `shell` package (NSPanel, tray, hotkey) and a bound `ActionService`. It built `bin/gifsmith.app` with `wails3 task darwin:package` in about 55 s, built a CGO-free `-tags server` binary, and called a bound Go method over HTTP, both directly and through a Vite proxy. A synthetic ⌃⌥⌘G showed the panel at floating level (CGWindow layer 3), sized 720×480, centered on the cursor's screen near the top. **The frontmost app stayed the same.** A second press hid it.

Three findings change the plan compared with the project research:
1. **`Show()` on a non-activating panel only calls `orderFrontRegardless`.** It does not make the panel key. You must call `Focus()`, which does `orderFrontRegardless` + `makeKeyWindow`. Also, the panel's `keyDown:` swallows keystrokes when the window rather than the WKWebView is first responder, and Wails never makes the webview first responder for panels. Its own notch window does (`makeFirstResponder:self.webView`). A probe keystroke did not reach the page. That attempt was confounded by the user switching apps, so it is inconclusive. Ship a 10-line `platform.FocusWebView(nativeWindow)` from the start.
2. **Signing does not read `SIGN_IDENTITY`/`KEYCHAIN_PROFILE` Taskfile vars** in beta.30. `darwin:sign` and `darwin:sign:notarize` run `wails3 tool sign --input … {{.CLI_ARGS}}`. The identity and profile come from flags, or from the global `~/.config/wails/defaults.yaml`, which does not exist on this machine. Add project Taskfile vars and pass them as flags (Pattern 5).
3. **`wails3 init` creates a `<name>/` subdirectory** and leaves `build/config.yml` at template defaults even when you pass `-productidentifier`. Scaffold, move the files to the repo root, and fix `config.yml` by hand. Otherwise a later `update:build-assets` reverts the bundle id.

**Primary recommendation:** Build on the probe structure in this document. Run every paste test from `~/Applications/Gifsmith.app`, Developer ID–signed with hardened runtime, using a project-local `darwin:install:dev` task. Give `wails3 dev` a separate bundle id (`com.kenscott.gifsmith.dev`) so the ad-hoc build never touches the signed app's TCC entry.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Global hotkey, panel show/hide/position, tray menu | Go `internal/shell` (Wails API) | `internal/platform` (cursor screen, frontmost app, webview focus) | Wails owns windows and the main thread. Shell is plain Go over Wails |
| Pasteboard file URL, ⌘V synthesis, AX trust, frontmost app, activate app | Go `internal/platform` (cgo/ObjC) | — | The only cgo package. darwin (`darwin && !server`) plus a stub (`!darwin \|\| server`) |
| Copy/Paste sequencing (copy → hide → delay → ⌘V, AX gate) | Go `internal/services` (bound `ActionService`) | `platform`, `shell.Hide` | The bound surface stays thin. Ordering is testable with a fake platform |
| Text input, Enter/⌘Enter handling, status message | Browser tier: Svelte in WKWebView | Wails KeyBindings (fallback for ⌘Enter) | UI only. It never sees file paths |
| Sample GIF storage | Build: `frontend/public/sample.gif` → embedded `frontend/dist` | Go stages it to `~/Library/Caches/com.kenscott.gifsmith/staging/sample.gif` | One source file. The pasteboard needs a real file on disk |
| Signing, notarization, install | Taskfile (`build/darwin/Taskfile.yml`) | User checkpoints (keychain profile, first codesign prompt) | Build tooling, not app code |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/wailsapp/wails/v3` | `v3.0.0-beta.30` [VERIFIED: probe go.mod `require github.com/wailsapp/wails/v3 v3.0.0-beta.30`, `go 1.25.0` after tidy] | Shell, panel, tray, hotkey, bindings, server mode | Locked decision |
| `wails3` CLI | `v3.0.0-beta.30` [VERIFIED: `go install …/cmd/wails3@v3.0.0-beta.30` → `wails3 version` printed `v3.0.0-beta.30`] | init, bindings, embedded Task runner, sign/notarize | Must match the library. **No separate `task` binary needed**: `wails3 task …` embeds Task (verified, `task` is not installed) |
| `@wailsio/runtime` | `3.0.0-beta.30` exact [VERIFIED: npm dist-tag `latest: '3.0.0-beta.30'`, published 2026-10-15T15:17:31Z] | JS runtime + Vite plugin (`@wailsio/runtime/plugins/vite`) | Template ships `"latest"`. Pin it exactly |
| svelte / vite / @sveltejs/vite-plugin-svelte / svelte-check / typescript | 5.57.2 / 8.3.4 / 7.3.1 / 4.7.6 / 5.9.3 [VERIFIED: resolved by `pnpm install` in probe, lockfile] | Template frontend | Comes from the official `svelte` (TS) template. Keep the template ranges and commit `pnpm-lock.yaml` |
| Apple frameworks via cgo | macOS SDK (Xcode 27.1) | Cocoa (NSPasteboard, NSWorkspace, NSScreen), ApplicationServices (AX, CGEvent), Carbon (TIS/UCKeyTranslate) | No third-party library needed |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| own cgo `CopyFile` | `golang.design/x/clipboard@v0.11.0` `WriteFiles` | Already rejected in STACK.md. Keep it as an escape hatch only |
| Wails `app.Clipboard` | — | Text-only. Cannot write file URLs |
| `golang.design/x/hotkey` | — | Forbidden (needs Accessibility) |

**Installation (exact commands, verified in probe):**
```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.30     # → ~/go/bin/wails3 (leave ~/go/bin/wails v2 alone)
SCR=$(mktemp -d) && wails3 init -n gifsmith -t svelte -d "$SCR" -mod gifsmith \
  -productname Gifsmith -productidentifier com.kenscott.gifsmith -productcompany "Ken Scott"
# init creates "$SCR/gifsmith/" — copy into the repo root, merging .gitignore:
cp -R "$SCR/gifsmith/." /Users/ken/dev/gif-search/   # then append the scaffold's .gitignore lines to the repo's
cd frontend && pnpm add @wailsio/runtime@3.0.0-beta.30 && pnpm install
```
- The template name is `svelte`, which is the TypeScript variant (`template.yaml`: `name: "Svelte + Vite (TypeScript)"`, `typescript: true`). **There is no `svelte-ts`.** `svelte-js` is the JS variant. [VERIFIED: internal/templates/svelte/template.yaml:2-8]
- `-mod` defaults to `changeme` when no `-git` is given. Pass `-mod gifsmith`, so the bindings import from `../bindings/gifsmith/...`. [VERIFIED: commands/init.go: `options.ModulePath = "changeme"`]
- Switch to pnpm **before the first build**. In `Taskfile.yml` line 6, change `PACKAGE_MANAGER: '{{.PACKAGE_MANAGER | default "npm"}}'` to `default "pnpm"`. The common Taskfile already has `install:frontend:deps:pnpm`, `frontend:run:pnpm` and `frontend:dev:pnpm` tasks. [VERIFIED: probe Taskfile.yml:6, build/Taskfile.yml pnpm tasks; build used `pnpm run build`]
- The scaffold includes `frontend/.npmrc` with `minimum-release-age=10080`. Keep it. Observed: an exact pin of the 2-day-old `@wailsio/runtime@3.0.0-beta.30` installed fine, and ranged deps resolved to older releases (svelte-check 4.7.6 from 2026-08-13 rather than the 10-16 release). [VERIFIED: probe pnpm install, clean store]

## Package Legitimacy Audit

| Package | Registry | Age (latest publish) | Downloads/wk | Source Repo | Verdict | Disposition |
|---------|----------|-----|-----------|-------------|---------|-------------|
| @wailsio/runtime | npm | latest 2026-10-15 (2 d) | 30,105 | github.com/wailsapp/wails | SUS (too-new) | Keep. Official Wails package, exact pin. Flag for checkpoint |
| svelte | npm | latest 2026-10-06 | 6.8M | github.com/sveltejs/svelte | SUS (too-new) | Keep. Template dep, lockfile-pinned |
| vite | npm | latest 2026-10-08 | 195M | github.com/vitejs/vite | SUS (too-new) | Keep. Template dep |
| @sveltejs/vite-plugin-svelte | npm | latest 2026-09-23 | 4.3M | github.com/sveltejs/vite-plugin-svelte | SUS (too-new) | Keep. Template dep |
| svelte-check | npm | latest 2026-10-16 | 3.1M | (official) | SUS (too-new) | Keep. Resolved to 4.7.6 (2026-08-13) |
| @tsconfig/svelte | npm | 2026-02-17 | 970K | — | OK | Approved |
| typescript | npm | 2026-07-08 | 307M | — | OK | Approved |
| tslib | npm | 2024-10-31 | 435M | — | OK | Approved |
| github.com/wailsapp/wails/v3 | Go proxy | tag 2026-10-15 | — | github.com/wailsapp/wails | not checked by seam (Go) | Approved (locked decision, source read at tag) |

**Packages removed due to [SLOP] verdict:** none.
**Packages flagged as suspicious [SUS]:** @wailsio/runtime, svelte, vite, @sveltejs/vite-plugin-svelte, svelte-check. Every flag is the "too-new" heuristic on each package's *latest* release. These are high-download official packages chosen by the official Wails template, and none has a postinstall script (`postinstall: null` for all). The planner should add **one** `checkpoint:human-verify` before the first `pnpm install` that covers this set and the committed lockfile. [VERIFIED: gsd-tools package-legitimacy check, 2026-10-17]

## Architecture Patterns

### System Architecture Diagram

```
 ⌃⌥⌘G (Carbon RegisterEventHotKey, no AX)        Tray click → menu: Open / Settings… (stub) / Quit
          │ callback on own goroutine                         │
          ▼                                                   ▼
   shell.Toggle() ── visible? ──yes──► panel.Hide()      shell.Show()
          │ no
          ▼
   platform.FrontmostApp()  (record prev app; log it)
   platform.CursorScreenWorkArea() → panel.SetPosition(centerX, top+h/6)
   panel.Show() → panel.Focus() → platform.FocusWebView(panel.NativeWindow())
   app.Event.Emit("panel:show")  ───────────────►  Svelte: input.focus(); input.select()
                                                         │ keydown Enter / ⌘Enter
                                                         ▼
                                     ActionService.Copy() / .Paste()   (bound Go, goroutine)
                                                         │
               ┌─────────────────────────────────────────┴───────────────────────┐
               ▼ Copy                                                             ▼ Paste
   platform.CopyFile(stagedSample)                                   platform.CopyFile(stagedSample)
   shell.Hide()  → Result{copied}                                    AccessibilityTrusted(false)?
                                                                       ├─ no → Result{copied, needsAccessibility}
                                                                       │       (panel stays open: "press ⌘V" + how to grant)
                                                                       └─ yes → shell.Hide(); sleep ~80ms;
                                                                                 [fallback: frontmost≠prev → ActivateApp(prev)]
                                                                                 platform.PasteShortcut() → Result{pasted}
   Esc → HideOnEscape binding (before webview) │ click-away → WindowDidResignKey → HideOnFocusLost
```

### Recommended Project Structure
```
/ (repo root = Wails project root)
├── main.go                    # wiring: stage sample, application.New, shell.New, services, Run
├── Taskfile.yml               # PACKAGE_MANAGER default "pnpm"
├── build/
│   ├── config.yml             # info: Gifsmith / com.kenscott.gifsmith (edit by hand!)
│   └── darwin/
│       ├── Info.plist         # + LSUIElement true, LSMinimumSystemVersion 13.0
│       ├── Info.dev.plist     # + LSUIElement true, bundle id com.kenscott.gifsmith.dev
│       └── Taskfile.yml       # 13.0 deployment target; SIGN_IDENTITY/KEYCHAIN_PROFILE vars; install:dev
├── internal/
│   ├── platform/
│   │   ├── platform.go        # Platform interface, App, Rect, ErrUnsupported (no build tag)
│   │   ├── platform_darwin.go # //go:build darwin && !server  (cgo, New() → darwin impl)
│   │   ├── platform_darwin.h / platform_darwin.m   # .m starts with //go:build darwin && !server
│   │   ├── platform_stub.go   # //go:build !darwin || server  (ErrUnsupported)
│   │   └── fake.go            # recording fake for unit tests (no build tag)
│   ├── shell/shell.go         # panel options, Toggle/Show/Hide, tray, hotkey registration
│   └── services/actions.go    # ActionService{Copy, Paste} + actions_test.go with fake
├── frontend/
│   ├── public/sample.gif      # generated, ~2 KB animated GIF (no license issue)
│   ├── bindings/              # generated by wails3 (commit it)
│   ├── src/App.svelte         # <input> + <img src="/sample.gif"> + status line
│   └── vite.config.ts         # + optional /wails proxy for browser mode
└── .planning/phases/01-…/01-PASTE-MATRIX.md
```

### Pattern 1: Panel window options (exact, compiled in probe)
```go
// Source: probe internal/shell/shell.go; API at webview_window_options.go:238-246, 615-693, 709-719, 742-768
panel := app.Window.NewWithOptions(application.WebviewWindowOptions{
    Name: "panel", Width: 720, Height: 480, URL: "/",
    Frameless: true, Hidden: true, DisableResize: true,
    HideOnEscape: true, HideOnFocusLost: true,
    Mac: application.MacWindow{
        WindowClass:      application.MacWindowClassPanel,
        PanelPreferences: application.MacPanelPreferences{NonActivating: true},
        CollectionBehavior: application.MacWindowCollectionBehaviorCanJoinAllSpaces |
            application.MacWindowCollectionBehaviorFullScreenAuxiliary,
        WindowLevel: application.MacWindowLevelFloating,
    },
})
```
Verbatim API facts [VERIFIED: v3/pkg/application/webview_window_options.go]:
- `:681-693`: `type MacPanelPreferences struct { FloatingPanel bool; BecomesKeyOnlyIfNeeded bool; NonActivating bool; UtilityWindow bool }`. The comment on NonActivating reads "applies NSWindowStyleMaskNonactivatingPanel. Showing or focusing the panel then leaves the currently active application active."
- `:712-719`: `MacWindowLevelNormal = "normal"`, `MacWindowLevelFloating = "floating"`, `MacWindowLevelTornOffMenu`, `MacWindowLevelModalPanel`, `MacWindowLevelMainMenu`, `MacWindowLevelStatus = "status"`, `MacWindowLevelPopUpMenu = "popUpMenu"`, `MacWindowLevelScreenSaver = "screenSaver"`.
- `:747`, `:755`, `:763`: `MacWindowCollectionBehaviorCanJoinAllSpaces … = 1 << 0`, `MacWindowCollectionBehaviorStationary … = 1 << 4`, `MacWindowCollectionBehaviorFullScreenAuxiliary … = 1 << 8`.
- If the full-screen overlay fails, the Wails docs' "Spotlight/menu-bar panel" recipe uses `WindowLevel: application.MacWindowLevelPopUpMenu` and adds `MacWindowCollectionBehaviorStationary` [CITED: docs/mpress/content/features/windows/options.md:869-879 at tag]. Try that before anything else.
- **Never `Close()` the panel.** Hide it. `WebviewPanel` sets `self.releasedWhenClosed = YES;` and `self.hidesOnDeactivate = NO;` [VERIFIED: webview_panel_darwin.m:27-28].
- **Do not copy the template's `ApplicationShouldTerminateAfterLastWindowClosed: true`** into the panel app (template `main.go.tmpl:47`, and the spotlight example sets it too). Leave it false for a tray app.

### Pattern 2: Show sequence and hotkey registration
```go
// Source: probe internal/shell/shell.go (compiled; smoke-tested for show/toggle/frontmost)
app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
    if err := app.GlobalShortcut.Register("Ctrl+Option+Cmd+G", s.Toggle); err != nil {
        log.Printf("hotkey: %v", err) // surfaced because we register after start
    }
})

func (s *Shell) Show() {
    if fa, err := s.plat.FrontmostApp(); err == nil { s.prev = fa }
    wa := s.plat.CursorScreenWorkArea()                       // Y-down, primary-top-left points
    s.panel.SetPosition(wa.X+(wa.Width-720)/2, wa.Y+wa.Height/6)
    s.panel.Show()
    s.panel.Focus()                                           // REQUIRED: Show() doesn't make a non-activating panel key
    s.plat.FocusWebView(s.panel.NativeWindow())               // WKWebView first responder → typing works without a click
    s.app.Event.Emit("panel:show", s.prev.Name)
}
```
- `Show()` for a non-activating panel runs `[nsWindow orderFrontRegardless]; return;`. `Focus()` runs `[nsWindow orderFrontRegardless]; [nsWindow makeKeyWindow]; return;` [VERIFIED: webview_window_darwin.go:942-952, 1099-1111]. Wails' own tray toggle uses `Window.Show().Focus()` [VERIFIED: systemtray.go:171-176].
- `Register` comment, verbatim: "Register may be called before [App.Run]; the OS binding is then performed when the application starts and any OS-level rejection is reported via the application's error handler rather than returned here." The callback "is invoked - on its own goroutine" [VERIFIED: global_shortcut_manager.go:79-100]. A Carbon `-9878 (eventHotKeyExistsErr)` is returned as an error [VERIFIED: global_shortcut_darwin.go:166-168]. Carbon only reports a clash with an *exclusive* registration, so a missing error does not prove there is no conflict. [ASSUMED] The user must confirm Vorssaint does not clash.
- Accelerator modifier names: `"cmd"`, `"ctrl"`, `"option"`/`"alt"`, `"shift"` [VERIFIED: keys.go:55-65]. Letter keys resolve through the active layout at registration time [VERIFIED: global_shortcut_darwin.go:205-222].
- `SetPosition(x, y)` is absolute, in points, Y-down from the top of the **primary** screen (`frame.origin.y = primaryHeight - frame.size.height - y`) [VERIFIED: webview_window_darwin.go:770-781]. Wails has **no cursor-position API on darwin**, so `CursorScreenWorkArea` lives in `platform` and returns `visibleFrame` flipped the same way. Probe log: `workarea={X:0 Y:38 Width:1710 Height:1074}`, panel placed at X=495 Y=217.

### Pattern 3: Tray menu (accessory app)
```go
// Source: probe; API per examples/systray-menu/main.go at tag
tray := app.SystemTray.New()
tray.SetTemplateIcon(icons.SystrayMacTemplate)          // github.com/wailsapp/wails/v3/pkg/icons
menu := app.NewMenu()
menu.Add("Open Gifsmith").OnClick(func(*application.Context) { s.Show() })
menu.Add("Settings…").OnClick(func(*application.Context) { /* Phase 4: real window. Phase 1: show panel with a "coming in a later build" line */ })
menu.AddSeparator()
menu.Add("Quit Gifsmith").OnClick(func(*application.Context) { app.Quit() })
tray.SetMenu(menu)   // do NOT AttachWindow: with a menu and no window, left-click opens the menu
```
App options: `Mac: application.MacOptions{ActivationPolicy: application.ActivationPolicyAccessory}` [VERIFIED: application_options.go:248-261]. Add `<key>LSUIElement</key><true/>` to **both** `build/darwin/Info.plist` and `Info.dev.plist`, because the generated plists have no such key [VERIFIED: probe Info.plist]. Keep "Settings…" enabled so the menu "offers" it. A disabled item may fail SHEL-04's wording.

### Pattern 4: cgo platform (compiled in probe; darwin + server stub both build)
```objc
// platform_darwin.m — first line: //go:build darwin && !server
// Main-thread hop: same rationale as Wails mainthread_darwin.go:13-23 (run loop, common modes,
// not the main dispatch queue, which stalls while a menu is tracking).
static void onMain(void (^block)(void)) {
    if ([NSThread isMainThread]) { block(); return; }
    dispatch_semaphore_t done = dispatch_semaphore_create(0);
    CFRunLoopRef loop = CFRunLoopGetMain();
    CFRunLoopPerformBlock(loop, kCFRunLoopCommonModes, ^{ block(); dispatch_semaphore_signal(done); });
    CFRunLoopWakeUp(loop);
    dispatch_semaphore_wait(done, DISPATCH_TIME_FOREVER);
    dispatch_release(done);
}
int gsCopyFile(const char *path) {               // ACTN-01
    __block BOOL ok = NO;
    onMain(^{ @autoreleasepool {
        NSPasteboard *pb = [NSPasteboard generalPasteboard];
        [pb clearContents];
        ok = [pb writeObjects:@[[NSURL fileURLWithPath:[NSString stringWithUTF8String:path]]]];
    }});
    return ok;
}
int gsAXTrusted(bool prompt) {
    NSDictionary *o = @{(__bridge NSString *)kAXTrustedCheckOptionPrompt: prompt ? @YES : @NO};
    return AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)o);
}
// Layout-aware V: translate WITH the command modifier so "Dvorak - QWERTY ⌘" picks the QWERTY key.
// Pattern copied from Wails keyCodeForChar (global_shortcut_darwin.go:58-108); fallback kVK_ANSI_V.
static CGKeyCode vKeyCode(void) { /* TISCopyCurrentKeyboardLayoutInputSource → UCKeyTranslate(code, kUCKeyActionDown,
                                     (cmdKey >> 8) & 0xFF, LMGetKbdType(), …) == 'v' */ }
int gsPostCmdV(void) {                           // ACTN-02 (call only after AX check + hide + delay)
    CGEventSourceRef src = CGEventSourceCreate(kCGEventSourceStatePrivate);
    CGKeyCode v = vKeyCode();
    CGEventRef d = CGEventCreateKeyboardEvent(src, v, true), u = CGEventCreateKeyboardEvent(src, v, false);
    CGEventSetFlags(d, kCGEventFlagMaskCommand); CGEventSetFlags(u, kCGEventFlagMaskCommand);
    CGEventPost(kCGSessionEventTap, d); CGEventPost(kCGSessionEventTap, u);
    CFRelease(d); CFRelease(u); CFRelease(src);
    return 1;
}
// FrontmostApp: [[NSWorkspace sharedWorkspace] frontmostApplication] (pid, bundleIdentifier, localizedName; strdup → Go frees)
// ActivateApp:  [[NSRunningApplication runningApplicationWithProcessIdentifier:pid] activateWithOptions:0] (best effort, macOS 14+ cooperative)
// CursorScreenWorkArea: screen whose frame contains [NSEvent mouseLocation]; visibleFrame flipped by primary screen height
// FocusWebView(void *nsWindow): find WKWebView in contentView subtree (NSClassFromString(@"WKWebView")); [w makeFirstResponder:wv]
```
```go
// platform_darwin.go header (probe, compiles; no ld warnings once the Taskfile target is 13.0)
//go:build darwin && !server
/*
#cgo CFLAGS: -x objective-c -mmacosx-version-min=13.0
#cgo LDFLAGS: -framework Cocoa -framework Carbon -framework ApplicationServices
#include <stdlib.h>
#include "platform_darwin.h"
*/
import "C"
```
- cgo here is manual reference counting, not ARC: `__bridge` casts are no-ops, so `CFRelease` CF objects and `free` strdup'd strings in Go. Wails' own `.m` files are MRC (`retain`/`release` in webview_panel_darwin.m:97-110).
- `FocusWebView` takes `WebviewWindow.NativeWindow()`, which returns the `NSWindow*` (`return w.nsWindow`) [VERIFIED: webview_window.go:1663-1669, webview_window_darwin.go:1781-1783]. It is nil before the app runs, so call it only from Show.
- Paste sequencing in Go (`ActionService.Paste`): `CopyFile`. If `!AccessibilityTrusted(false)`, return `{Copied:true, NeedsAccessibility:true, Message:"Copied — press ⌘V. To paste automatically, enable Gifsmith in System Settings › Privacy & Security › Accessibility."}` and keep the panel open. Otherwise `Hide()`, then `time.Sleep(80*time.Millisecond)`, then `PasteShortcut()`. Offer a button that calls `AccessibilityTrusted(true)` (system prompt) and opens `x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility` [CITED: STACK.md item 4].

### Pattern 5: Signing, install, notarize (project Taskfile, not global defaults)
`wails3 tool sign` reads flags first. `resolveSigningDefaults` fills only empty flags from `~/.config/wails/defaults.yaml` (`signing.darwin.identity`, `keychainProfile`, `entitlements`) [VERIFIED: internal/commands/sign.go:18-36; internal/defaults/defaults.go:52-56]. That file does not exist here [VERIFIED: `ls ~/.config/wails` → No such file]. Codesign args: `--force --deep --sign <identity>`, plus `--options runtime` when `--hardened-runtime` or `--notarize` is set. Then `codesign --verify --deep --strict`. Notarize does `ditto -c -k --keepParent` → `xcrun notarytool submit <zip> --keychain-profile <p> --wait` → `xcrun stapler staple` [VERIFIED: sign.go:122-134, 151, 202-244].

Add to `build/darwin/Taskfile.yml`:
```yaml
vars:
  CROSS_IMAGE: wails-cross
  SIGN_IDENTITY: "Developer ID Application: Ken Scott (S2K7P43927)"
  KEYCHAIN_PROFILE: gifsmith-notary
  INSTALL_PATH: '{{.HOME}}/Applications/Gifsmith.app'

tasks:
  # change existing build:native env: CGO_CFLAGS/CGO_LDFLAGS "-mmacosx-version-min=13.0", MACOSX_DEPLOYMENT_TARGET "13.0"
  install:dev:
    summary: Build, Developer ID-sign (hardened runtime) and install to ~/Applications/Gifsmith.app
    deps: [package]
    cmds:
      - wails3 tool sign --input "{{.BIN_DIR}}/{{.APP_NAME}}.app" --identity "{{.SIGN_IDENTITY}}" --hardened-runtime
      - pkill -x {{.APP_NAME}} || true
      - rm -rf "{{.INSTALL_PATH}}"
      - ditto "{{.BIN_DIR}}/{{.APP_NAME}}.app" "{{.INSTALL_PATH}}"
      - open "{{.INSTALL_PATH}}"
  sign:notarize:   # replace the template body
    deps: [package]
    cmds:
      - wails3 tool sign --input "{{.BIN_DIR}}/{{.APP_NAME}}.app" --identity "{{.SIGN_IDENTITY}}" --notarize --keychain-profile "{{.KEYCHAIN_PROFILE}}"
```
(`{{.HOME}}`: Task exposes env vars as template vars [ASSUMED]. If it does not, use `$HOME` inside the command string.) The app bundle is `bin/gifsmith.app` (`APP_NAME` = binary name `gifsmith`) [VERIFIED: probe]. Renaming it to `Gifsmith.app` on install does not affect the signature.

### Anti-Patterns to Avoid
- **Testing paste from `wails3 dev` or from `bin/`.** Ad-hoc signatures have `designated => cdhash H"…"`, which changes every build, so the TCC grant is lost each time [VERIFIED: probe `codesign -d -r-` on the ad-hoc app].
- **Running `wails3 task common:update:build-assets` casually.** It re-extracts the plists and merges your old keys *under* the template values. `LSUIElement` survives. `LSMinimumSystemVersion` goes back to `12.0.0`, and `CFBundleIdentifier` reverts to whatever `build/config.yml` says (template default `com.mycompany.myproduct`) [VERIFIED: build-assets.go:307-320, 430-431 "start with backup (user's content), apply new values on top"; config.yml:7-14 defaults].
- **Using `HideOnEscape` once the panel has nested UI (Phase 2+).** The binding consumes Esc before the webview sees it. Fine for Phase 1. Revisit when the ⌘K menu needs "Esc closes menu first".
- **Calling `dispatch_sync(dispatch_get_main_queue(), …)` from platform code.** It stalls while a tray menu is tracking. Use the run-loop helper above.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Global hotkey | CGEventTap / NSEvent monitor | `app.GlobalShortcut` (Carbon) | No Accessibility needed. Layout-aware letters |
| Floating, non-activating window | own NSPanel subclass via cgo | `MacWindowClassPanel` + `PanelPreferences` | Wails owns the window/webview lifecycle |
| Menu bar icon/menu | NSStatusItem cgo | `app.SystemTray` | Runs on Wails' main thread |
| Main-thread dispatch from Go | goroutine-locking tricks | `application.InvokeSync` (in shell) / run-loop helper (in platform .m) | Wails handles reentrancy (`dispatchOnMainThread` runs inline on main, application.go:981-986) |
| Signing/notarize/staple | shell scripts | `wails3 tool sign … --notarize` | Already does xattr clean, verify, ditto zip, `--wait`, staple |
| Sample GIF asset | a downloaded GIF of unclear license | generate with `image/gif` (3 frames, 120×120, about 1.8 KB, verified `GIF image data, version 89a`) | No licensing question. Reproducible |

## Common Pitfalls

### Pitfall 1: Panel shows but typing doesn't reach the input
**What goes wrong:** The panel appears, but keystrokes are lost or go to the previous app.
**Why:** `Show()` doesn't make a non-activating panel key (Pattern 2). Even when the panel is key, `WebviewPanel keyDown:` only forwards the key to Wails bindings and never calls super [VERIFIED: webview_panel_darwin.m:49-55]. So if the window itself is first responder, text is swallowed. Wails' notch window explicitly does `[self makeFirstResponder:self.webView]` "for immediate keyboard input" [VERIFIED: webview_notch_window_darwin.m:174-182], and the panel path does not. In the probe, one synthetic "a" after show was not seen by the page. The user was switching apps at the time, so treat that result as unproven rather than failed.
**How to avoid:** `Show()` → `Focus()` → `platform.FocusWebView(NativeWindow())`, and have JS call `input.focus()` on `panel:show`. If typing still fails, use the locked fallback: activate the panel, then `ActivateApp(prev)` before paste.
**Warning signs:** The input has no caret. Characters appear in the previous app. Esc does nothing (HideOnEscape needs key status).

### Pitfall 2: TCC grant keeps disappearing
**Why:** Ad-hoc dev builds use a cdhash designated requirement. `wails3 dev`'s `Info.dev.plist` uses the **same** bundle id as release (`com.kenscott.gifsmith`) [VERIFIED: probe Info.dev.plist:10-11]. An ad-hoc build asking for AX under that id may conflict with the signed app's TCC row [ASSUMED].
**How to avoid:** Set `Info.dev.plist` `CFBundleIdentifier` to `com.kenscott.gifsmith.dev`. Test paste only from `~/Applications/Gifsmith.app` installed by `darwin:install:dev`. Check with `codesign -d -r- ~/Applications/Gifsmith.app`: the designated requirement must contain `identifier "com.kenscott.gifsmith"` and `certificate leaf[subject.OU] = S2K7P43927`, not `cdhash` [ASSUMED format]. To test the no-AX path again, remove the app's row in System Settings or run `tccutil reset Accessibility com.kenscott.gifsmith` [ASSUMED: works without sudo].

### Pitfall 3: Synthetic ⌘V lands in Gifsmith or nowhere
**Why:** The panel is still key when the event posts, or AX is not granted. CGEventPost drops events silently without AX [ASSUMED, consistent with STACK.md].
**How to avoid:** Check AX first, then `Hide()` (a synchronous `orderOut:` via InvokeSync), sleep 60–100 ms (start at 80), then post. Log `FrontmostApp()` right before posting. If it is not `prev`, `ActivateApp(prev)` and wait again. Record the delay that works in the matrix.
**Warning signs:** Nothing pastes, or the GIF pastes into the Gifsmith input.

### Pitfall 4: Bundle id / min-OS drift
- `wails3 init -productidentifier …` fills the plists but leaves `build/config.yml` at `productIdentifier: "com.mycompany.myproduct"` [VERIFIED: probe]. Edit `config.yml` `info:` to match.
- The darwin Taskfile targets 12.0, which made the linker warn "built for newer 'macOS' version (13.0) than being linked (12.0)" once platform used `-mmacosx-version-min=13.0` [VERIFIED: probe build output]. Set the Taskfile env to 13.0 and `LSMinimumSystemVersion` to `13.0` in both plists.

### Pitfall 5: First Developer ID codesign blocks on a keychain dialog
**Why:** The first use of the private key from a new tool can show "codesign wants to sign using key…" [ASSUMED]. A non-interactive executor would hang.
**How to avoid:** Make the first `darwin:install:dev` a user checkpoint ("click Always Allow if prompted"). Also make `xcrun notarytool store-credentials gifsmith-notary --apple-id <id> --team-id S2K7P43927` a user checkpoint, and verify it with `xcrun notarytool history --keychain-profile gifsmith-notary`.

### Pitfall 6: Notarization rejects the signature
**Why:** `wails3 tool sign` does not pass `--timestamp`. `man codesign`: "If this option is not given at all, a system-specific default behavior is invoked. This may result in some but not all code signatures being timestamped." [VERIFIED: local man page]. Developer ID signatures are normally timestamped by default [ASSUMED].
**How to avoid:** Before submitting, `codesign -dvv bin/gifsmith.app 2>&1 | grep Timestamp`. On rejection, run `xcrun notarytool log <id> --keychain-profile gifsmith-notary`. Afterwards check `spctl -a -vvv -t exec <app>` (expect `source=Notarized Developer ID`) and `xcrun stapler validate <app>`. No entitlements file should be needed for a non-sandboxed WKWebView app without JIT [ASSUMED]. Notarization will tell you if one is.

### Pitfall 7: Browser-mode WebSocket events rejected through the Vite proxy
**Why:** Server mode enforces same-origin on WebSockets ("the default empty list enforces same-origin connections") [VERIFIED: application_options.go:163-172]. A proxied page's Origin (`http://127.0.0.1:9245`) differs from the backend host. Bound calls over HTTP POST worked through the proxy (verified). Events over WS are untested.
**How to avoid:** In Phase 1 only calls are needed. If events are needed in browser mode, set `Server: application.ServerOptions{…, WebSocketOriginPatterns: []string{"http://127.0.0.1:9245"}}`. **Never** set `WebSocketAllowAllOrigins`.

### Pitfall 8: Port 8089 already taken on this machine
A leftover research probe is listening on 127.0.0.1:8089: PID 27210, `/var/folders/…/go-build694470340/b001/exe/false`, started 07:32 today [VERIFIED: lsof/ps]. The user should kill it, or the plan should pick another port (the probe used 8090).

### Pitfall 9: Typing synthetic keys during smoke tests
Synthetic keystrokes go to whatever is focused. During the probe the user changed apps between steps, and a test "a" may have landed in their app. Automated smoke must send **only** the global hotkey, which Carbon consumes. Never send Enter or text unless the user is not using the machine (checkpoint first).

## Code Examples

### Frontend panel (Svelte 5, probe-compiled)
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { ActionService } from '../bindings/gifsmith/internal/services';
  let query = $state(''); let status = $state(''); let input: HTMLInputElement;
  onMount(() => { Events.On('panel:show', () => { status = ''; input.focus(); input.select(); }); });
  async function onKeydown(e: KeyboardEvent) {
    if (e.key !== 'Enter') return;
    e.preventDefault();
    try { status = (e.metaKey ? await ActionService.Paste() : await ActionService.Copy()).message; }
    catch (err) { status = String(err); }
  }
</script>
<input bind:this={input} bind:value={query} onkeydown={onKeydown} placeholder="Search GIFs" autocomplete="off" />
<img src="/sample.gif" alt="sample" width="120" height="120" />
<p>{status}</p>
```
The generated binding is `export function Copy(): $CancellablePromise<$models.Result>` in `frontend/bindings/gifsmith/internal/services/actionservice.ts` [VERIFIED: probe]. If ⌘Enter doesn't reach JS, add `KeyBindings: map[string]func(application.Window){"Cmd+Enter": …}` to the panel. Wails handles key equivalents before the webview, and only when a binding matches [VERIFIED: webview_window_darwin.m:66-74, webview_panel_darwin.m:37-47].

### Main wiring + sample staging
```go
//go:embed all:frontend/dist
var assets embed.FS
func stageSample() (string, error) {
    b, err := fs.ReadFile(assets, "frontend/dist/sample.gif") // Vite copies public/ → dist/
    if err != nil { return "", err }
    dir, err := os.UserCacheDir(); if err != nil { return "", err }
    p := filepath.Join(dir, "com.kenscott.gifsmith", "staging", "sample.gif")
    if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil { return "", err }
    return p, os.WriteFile(p, b, 0o644)   // must persist: the pasteboard references this path
}
app := application.New(application.Options{
    Name: "Gifsmith",
    Services: []application.Service{application.NewService(actions)},
    Assets:   application.AssetOptions{Handler: application.AssetFileServerFS(assets)},
    Mac:      application.MacOptions{ActivationPolicy: application.ActivationPolicyAccessory},
    Server:   application.ServerOptions{Host: "localhost", Port: 8089}, // used only by -tags server; env WAILS_SERVER_PORT overrides
})
```
Optionally register the event type for typed TS: `application.RegisterEvent[string]("panel:show")` in `init()` (template pattern, `main.go.tmpl:20-25`).

### Browser loop (`-tags server` + Vite proxy), verified
```ts
// frontend/vite.config.ts — keep template plugins; proxy only when GIFSMITH_BACKEND is set
const backend = process.env.GIFSMITH_BACKEND;
server: { host: "127.0.0.1", port: Number(process.env.WAILS_VITE_PORT) || 9245, strictPort: true,
          proxy: backend ? { "/wails": { target: backend, ws: true } } : undefined },
```
```bash
CGO_ENABLED=0 go build -tags server -o bin/gifsmith-server . && WAILS_SERVER_PORT=8089 ./bin/gifsmith-server &
cd frontend && GIFSMITH_BACKEND=http://localhost:8089 pnpm exec vite      # open http://127.0.0.1:9245 in dev-browser
# smoke: curl -X POST http://127.0.0.1:9245/wails/runtime -H 'Content-Type: application/json' -H 'x-wails-client-id: t' \
#   -d '{"object":0,"method":0,"args":{"call-id":"1","methodID":<id from actionservice.ts>,"args":[]}}'
# → HTTP 422 {"message":"not supported in this build","kind":"RuntimeError"}  (stub reached = loop works)
```
[VERIFIED: probe. The 422 response came through both the direct port and the Vite proxy. `CGO_ENABLED=0` proves the platform stub has no cgo. `/health` → 200 and `/sample.gif` → `image/gif` in server mode.] The template's `build:server`/`run:server` tasks also exist [VERIFIED: build/Taskfile.yml].

## What Claude Can Automate vs User Checkpoints

| Automatable (executor) | User checkpoint (manual) |
|---|---|
| Install `wails3`, scaffold, move to root, pnpm switch, pin runtime | Approve the npm install set (one human-verify, per audit) |
| `go vet ./...`, `go test ./internal/...` (ActionService with fake: copy without AX never calls PasteShortcut; paste hides **before** posting; position math) | `xcrun notarytool store-credentials gifsmith-notary` (app-specific password) |
| `CGO_ENABLED=0 go build -tags server .` + curl binding smoke through the Vite proxy | First Developer ID codesign keychain prompt ("Always Allow") |
| `pnpm -C frontend check` (svelte-check) | Confirm ⌃⌥⌘G doesn't clash with Vorssaint |
| `wails3 task darwin:install:dev`. Then `codesign --verify --deep --strict`, `codesign -d -r-` (no cdhash), `plutil -extract LSUIElement raw <app>/Contents/Info.plist` → `true`, `/usr/bin/pgrep -x gifsmith` | Typing in the panel, Esc, click-away, hotkey toggle from a **full-screen** app (SHEL-01/02/03) |
| Hotkey-only smoke (needs a terminal with AX, verified trusted): post ⌃⌥⌘G with a CGEvent helper, then check that `CGWindowListCopyWindowInfo` shows a Gifsmith window at layer 3, 720×480, and that `frontmostApplication` is unchanged; press again to hide. Probe harness: `/private/tmp/claude-501/-Users-ken-dev-gif-search/9f53d405-1914-4df1-8002-37e6a551a349/scratchpad/smoke/main.go` | Menu bar only: no Dock icon, not in ⌘Tab; menu Open/Settings/Quit (SHEL-04) |
| Pasteboard check after the user presses Enter: read `NSPasteboard` types and expect `public.file-url` → `…/staging/sample.gif` | Enter → ⌘V into TextEdit animates, with AX **not** granted (ACTN-01) |
| `wails3 task darwin:sign:notarize`, `spctl -a -vvv -t exec`, `xcrun stapler validate` | AX prompt/grant flow; ⌘Enter pastes into TextEdit; the no-AX message shows (ACTN-02) |
| Write the `01-PASTE-MATRIX.md` template | Fill the paste matrix (Slack, Messages, Mail, Discord, Notes, Telegram, Safari/Chrome composers, TextEdit; n/a if missing) from the signed dev build **and** the notarized build; explicit go/no-go |

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `SIGN_IDENTITY`/`KEYCHAIN_PROFILE` in Taskfile (STACK.md assumption) | `wails3 tool sign` flags, or global `~/.config/wails/defaults.yaml` via `wails3 setup` | at beta.30 (read at tag) | Add project vars yourself (Pattern 5) |
| `wails3 init -n x -d .` scaffolds into `.` (STACK.md) | Creates `./x/` and refuses a non-empty `x` | at beta.30 (probe) | Scaffold to temp and copy to root |
| `codesign --deep` | Deprecated for signing as of macOS 13 (man page) but still works | macOS 13 | Fine for a single-binary bundle. No nested code |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `FocusWebView` (makeFirstResponder on WKWebView) is what makes typing work without a click | Pattern 2, Pitfall 1 | Typing still fails, so use the locked activate-panel fallback |
| A2 | Carbon only reports a hotkey clash for exclusive registrations, so a missing error ≠ no conflict | Pattern 2 | User may hit double-firing with Vorssaint |
| A3 | Developer ID designated requirement = identifier + team OU (stable across rebuilds) | Pitfall 2 | TCC grant lost per rebuild. Verify with `codesign -d -r-` |
| A4 | Ad-hoc dev build with the same bundle id can disturb the signed app's TCC row | Pitfall 2 | Low. The `.dev` id is a cheap mitigation either way |
| A5 | `tccutil reset Accessibility com.kenscott.gifsmith` works without sudo | Pitfall 2 | Use System Settings to remove the row instead |
| A6 | CGEventPost silently drops without AX | Pitfall 3 | None for design, since we gate on AX first |
| A7 | First codesign may show a keychain prompt | Pitfall 5 | Executor hangs. Hence the checkpoint |
| A8 | Developer ID codesign timestamps by default, and no entitlements are needed for notarization | Pitfall 6 | Notarization rejected. Re-sign with `--timestamp` / add entitlements |
| A9 | Task exposes `{{.HOME}}` as a template var | Pattern 5 | Use `$HOME` in the command string |
| A10 | UCKeyTranslate with the command modifier resolves "Dvorak - QWERTY ⌘" to the QWERTY V | Pattern 4 | Wrong key on that layout only. Maccy special-cases the layout name |
| A11 | Left-mouse global monitors need no AX (fallback if HideOnFocusLost misses click-away) | Open Q2 | Fallback unavailable without AX |

## Open Questions (RESOLVED)

All four are resolved for planning: each has an explicit gate step in 01-01 Task 3 (tracer gate human-check, numbered steps below) and a pre-authorized fallback in that task's action, so no question is left open going into execution. The empirical answers are recorded in 01-01-SUMMARY.md.

1. **Does `FocusWebView` make the panel take typing immediately (and keep the frontmost app)?** Unverified. The probe attempt was confounded by user activity. Recommendation: include it in the first build, then run the user checkpoint. Use the locked fallback if needed.
   **RESOLVED** — Gate: 01-01 Task 3 human-check steps 2-3 (menu bar still shows TextEdit; "cats" typed with no click lands in the input). Fallback: locked decision D-08 (non-activating off, record frontmost app on Show, reactivate it on Hide; pulls `FrontmostApp`/`ActivateApp` forward from plan 01-03).
2. **Does click-away hide reliably when the user clicks a window of the already-active previous app?** `HideOnFocusLost` depends on the panel resigning key. During the probe the panel did hide after the user switched apps (weak evidence). Fallback: `[NSEvent addGlobalMonitorForEventsMatchingMask:NSEventMaskLeftMouseDown|NSEventMaskRightMouseDown]` → Hide (A11).
   **RESOLVED** — Gate: 01-01 Task 3 human-check step 6 (click on the TextEdit window hides the panel). Fallback: NSEvent global left/right mouse-down monitor in `internal/platform` that calls Hide (A11; needs no Accessibility).
3. **Synthetic hotkey flake:** in one of four probe launches the hotkey didn't show the panel on two tries. The cause is unknown: possibly Vorssaint active with its command bar or a keyboard feature, or user activity. Repeat the hotkey several times in the user checkpoint, with Vorssaint frontmost and with a full-screen app.
   **RESOLVED** — Gate: 01-01 Task 3 human-check steps 5, 7 and 9 (toggle twice; full-screen app; repeated presses with Vorssaint frontmost), plus the "hotkey registered" log grep in its verify; re-checked in the 01-03/01-04 matrix rows N1, N2, N10. Fallback: the tray "Open Gifsmith" item (plan 01-02) shows the panel without the hotkey (roadmap criterion 2); a reproducible miss is a RULE 0 stop-and-report at the gate rather than an improvised fix.
4. **Floating vs PopUpMenu level over full-screen apps on macOS 27.** Start with Floating. Switch to the docs' PopUpMenu + Stationary recipe if the overlay fails.
   **RESOLVED** — Gate: 01-01 Task 3 human-check step 7 (panel appears over a full-screen app). Fallback: `MacWindowLevelPopUpMenu` plus `MacWindowCollectionBehaviorStationary` (research Pattern 1 docs recipe).

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | build | ✓ | go1.27.2 darwin/arm64 | — |
| wails3 CLI | scaffold/build/sign | ✗ (only `~/go/bin/wails` v2) | install v3.0.0-beta.30 (verified installs and runs) | — |
| Task runner | Taskfiles | ✓ via `wails3 task` (embedded); standalone `task` ✗ | — | not needed |
| Node / pnpm | frontend | ✓ | node v24.14.1, pnpm 12.14.0 (ran as 12.6.0) | — |
| Xcode / notarytool | cgo, notarize | ✓ | Xcode 27.1, notarytool 1.1.3 | — |
| Developer ID identity | signing | ✓ | `Developer ID Application: Ken Scott (S2K7P43927)` (1 valid identity) | — |
| notarytool profile `gifsmith-notary` | notarize | ✗ | — | user creates (checkpoint) |
| `~/Applications` | dev install | ✓ | exists | — |
| Port 8089 | server mode | ✗ (held by stale probe PID 27210) | — | kill it, or use another port |
| Terminal Accessibility | hotkey smoke harness | ✓ (`AXIsProcessTrusted` = 1) | — | skip automated smoke |

**Missing, no fallback:** none blocking. The `wails3` install and the notary profile are plan steps.

## Security Domain

### Applicable ASVS Categories (L1)
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | — |
| V3 Session Management | no | — |
| V4 Access Control | yes (local) | Bound methods take **no path or URL from the webview**. The sample path is fixed in Go. Paste runs only on explicit Enter/⌘Enter |
| V5 Input Validation | minimal | No user input reaches native calls in Phase 1 (the query is unused) |
| V6 Cryptography | no | — (signing via Apple tooling) |
| V14 Configuration | yes | Server mode binds `localhost` only. Never set `WebSocketAllowAllOrigins`. Never ship `-tags server` builds. Notary credentials live only in the keychain profile. Nothing secret in the repo or Taskfile (the identity name is public) |

### Known Threat Patterns
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Any script in the webview invokes Paste (keystroke injection) | Elevation | Load local assets only, no remote content in Phase 1. Paste posts only ⌘V, never arbitrary keys |
| Localhost CSRF against the server-mode backend | Tampering | Dev-only binary. Stub platform returns ErrUnsupported. WS same-origin on by default |
| Hardened-runtime gaps | Tampering | Sign dev builds with `--hardened-runtime` so dev matches the notarized build |

## Sources

### Primary (HIGH)
- Wails repo at tag `v3.0.0-beta.30` (commit 623b529, cloned to scratchpad): `internal/templates/{svelte,_common}`, `internal/commands/{init.go,sign.go,build-assets.go,build_assets/darwin/Taskfile.yml,build_assets/Taskfile.tmpl.yml,updatable_build_assets/darwin/*.plist.tmpl}`, `internal/defaults/defaults.go`, `internal/flags/{init,sign}.go`, `pkg/application/{webview_window_options.go,webview_window.go,webview_window_darwin.go,webview_window_darwin.m,webview_panel_darwin.m,webview_notch_window_darwin.m,global_shortcut_*.go,systemtray.go,screen_darwin.go,mainthread_darwin.go,application_options.go,keys.go,events_*.go}`, `examples/{spotlight,systray-menu,server}`, `docs/mpress/content/features/windows/options.md`
- Scratch probe (this session): scaffold, pnpm install, `wails3 task darwin:package`, server build + HTTP/Vite-proxy binding call, hotkey/show/frontmost smoke. Code at `/private/tmp/claude-501/-Users-ken-dev-gif-search/9f53d405-1914-4df1-8002-37e6a551a349/scratchpad/initprobe/gifsmith`
- npm registry (`npm view`), gsd-tools package-legitimacy, local `man codesign`

### Secondary (MEDIUM)
- `.planning/research/STACK.md`, `ARCHITECTURE.md`, `PITFALLS.md` (Maccy paste routine, AX deep link, pasteboard type probe)

## Metadata

**Confidence breakdown:**
- Standard stack / scaffold / build: HIGH. Executed end to end on this Mac.
- Panel/hotkey/tray APIs: HIGH. Read at the tag and compiled. Show/toggle/frontmost smoke passed.
- Typing into the panel: MEDIUM-LOW until the user checkpoint (Pitfall 1).
- Signing/notarization: MEDIUM. Code read and identity present. Not executed (needs user checkpoints).
- Paste behavior across apps: LOW until the matrix is filled.

**Research date:** 2026-10-17
**Valid until:** while pinned to beta.30 (re-check on any Wails upgrade; nightly betas move every 2–5 days)
