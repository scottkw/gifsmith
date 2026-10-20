# Phase 4: Settings, Shell Polish & Release - Pattern Map

**Mapped:** 2026-10-18
**Files analyzed:** 30 (new + modified)
**Analogs found:** 27 / 30

All analog paths below are git-tracked source (checked with `git ls-files`).

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/settings/settings.go` (Prefs grows; `Store` single owner; `SecretStore.Delete`, `Keys.Delete`) | model/store | file-I/O, CRUD | itself (`Keys`, `WriteJSONAtomic`, `LoadPrefs`) | exact |
| `internal/settings/settings_test.go` | test | file-I/O | itself | exact |
| `internal/services/settings.go` (Snapshot, Set*, KeyStatus/TestKey/RemoveKey, ChooseFolder, AccessibilityStatus, `KlipyFileActions` const; fix `SetLastProvider`) | service (bound) | request-response | itself (`SaveKey`) | exact |
| `internal/services/settings_test.go` | test | request-response | itself (`newSettings` fixture) | exact |
| `internal/services/actions.go` (Klipy gate, T5a string, `DownloadsDir` from prefs, hide filename) | service | request-response | itself | exact |
| `internal/services/panel.go` (+ `OpenSettings`) | service (bound) | request-response | itself (`HideFunc`) | exact |
| `internal/services/fsprobe.go` or `internal/fsprobe/` (ValidateDownloadDir: Statfs + link probe) | utility | file-I/O | `internal/settings/settings.go` `WriteJSONAtomic` (temp-file/cleanup idiom) | partial |
| `internal/shell/shell.go` (settings window, tray "Settings…", theme apply) | shell/window | event-driven | itself (panel creation + WindowClosing hook) | exact |
| `internal/shell/hotkeys.go` (Registrar, Rebind/Suspend/Resume/Test) | shell | event-driven | `internal/shell/shell.go` lines 79-87 (Register on ApplicationStarted) | role-match |
| `internal/shell/hotkeys_test.go` | test | event-driven | `internal/shell/position_test.go` | role-match |
| `internal/shell/autostart.go` (interface + fake) | shell | request-response | `internal/platform/platform.go` + `fake.go` (interface/fake split) | role-match |
| `internal/platform/platform.go` (+ `SetAppearance`, `SystemLocale`) | interface | request-response | itself | exact |
| `internal/platform/platform_darwin.go` / `.h` / `.m` | native glue (cgo) | request-response | `gsOpenAXSettings` / `gsActivate` (.m 195-211) | exact |
| `internal/platform/platform_stub.go`, `fake.go` | stub/fake | — | themselves | exact |
| `internal/provider/giphy/giphy.go` (rating, lang, limit per request) | provider client | request-response | itself lines 170-172 | exact |
| `internal/provider/klipy/klipy.go` (per_page, locale) | provider client | request-response | itself lines 118-119 | exact |
| `internal/provider/provider.go` (`PageSize` const → param) | model | — | itself line 58-59 | exact |
| `internal/search/search.go` (+ `Flush` not resetting 429) | service/cache | CRUD | `Invalidate` lines 80-91 | exact |
| `internal/library/store.go` (surface FavoritesMax cap) | store | file-I/O | itself | exact |
| `main.go` (wire Store, settings window, new deps) | config/wiring | — | itself | exact |
| `wire_desktop.go` / `wire_server.go` (PickFolder, Autostart, Registrar fakes) | config | — | themselves (`Platform: &platform.Fake{}`) | exact |
| `frontend/settings.html`, `frontend/src/settings/main.ts` | entry | — | `frontend/index.html`, `frontend/src/main.ts` | exact |
| `frontend/vite.config.ts` (second rollup input) | config | — | itself | exact |
| `frontend/src/settings/SettingsApp.svelte`, `*Pane.svelte`, `KeyRow.svelte`, `HotkeyRecorder.svelte`, `SettingRow.svelte` | component | request-response | `frontend/src/components/SetupCard.svelte` | role-match |
| `frontend/src/lib/hotkeys.ts`, `theme.ts`, `prefs.ts` (+ `.test.ts`) | utility | transform | `frontend/src/lib/grid.ts` + `grid.test.ts` | role-match |
| `frontend/src/lib/api.ts` (new settings calls, EXTERNAL allow-list) | facade | request-response | itself | exact |
| `frontend/src/lib/actions.ts`, `App.svelte`, `Footer.svelte`, `app.css` (defaults, Klipy gate, data-theme) | component/utility | event-driven | themselves | exact |
| `scripts/uat/15-settings*.js` | test (UAT) | request-response | `scripts/uat/05-setup.js` | exact |
| `build/darwin/Taskfile.yml` (`release` task, LICENSE copy), `build/darwin/Info.plist` (1.0.0, usage strings) | config/build | batch | itself (`install:dev`, `create:app:bundle`) | exact |
| `build/darwin/dmg-background.png`, `README.md`, release notes | asset/docs | — | — | no analog |

## Pattern Assignments

### `internal/settings/settings.go` (store, file-I/O)

**Analog:** same file. Reuse `WriteJSONAtomic` (lines ~146-177) verbatim for the new `Store`; never write a `Prefs{...}` literal outside the Store.

Lock + write-then-swap pattern to copy for `Store.Update` (from `Keys.Set`):
```go
// Set writes the store first; the cache changes only when the write succeeded.
func (k *Keys) Set(p provider.ID, key string) error {
	if err := k.store.Set(string(p), key); err != nil {
		return err
	}
	k.mu.Lock()
	k.m[p] = key
	k.mu.Unlock()
	return nil
}
```
First-run semantics to preserve in `Store.Load` (from `LoadPrefs`):
```go
b, err := os.ReadFile(filepath.Join(dir, prefsFile))
if errors.Is(err, os.ErrNotExist) {
	return p, nil
}
```
Fixed-text Keychain errors for the new `Delete` (copy from `KeyringStore.Get`):
```go
v, err := keyring.Get(s.Service, account)
if errors.Is(err, keyring.ErrNotFound) {
	return "", ErrNoKey
}
if err != nil {
	return "", errors.New("keychain read failed")
}
```
`MemoryStore` needs a matching `Delete` (mutex-guarded `delete(s.m, account)`).

---

### `internal/services/settings.go` (bound service, request-response)

**Analog:** same file, `SaveKey`. New Set* methods follow: parse/validate → `s.mu.Lock()` → Store.Update → side effect (Flush / Rebind) → `s.Emit("settings:changed")`.
```go
func (s *SettingsService) SaveKey(ctx context.Context, p, key string) error {
	id, ok := provider.ParseID(p)
	if !ok {
		return errUnknownProvider
	}
	...
	s.mu.Lock()
	defer s.mu.Unlock()
	...
	if err := s.Keys.Set(id, key); err != nil {
		log.Printf("key not saved for %s: keychain write failed", id)
		return errors.New("couldn't save the key to the macOS Keychain")
	}
	s.Svc.Invalidate(id)
	s.Svc.Seed(id, "", page)
	...
	if s.Emit != nil {
		s.Emit("providers:changed")
	}
```
Bug to fix (lines 102-110) — erases every new field:
```go
return settings.SavePrefs(s.Dir, settings.Prefs{LastProvider: string(id)})
```
`TestKey` must call `s.NewClient`/the real client inside `s.Svc.Guard(p, …)` (see main.go clips Probe), not `Svc.Trending` (cache). Log lines name provider + kind only (`kindOf(err)`). `Emit` is `func(name string)`; a payload-carrying snapshot event needs a widened signature (`func(name string, data ...any)`) wired in main.go like `settingsSvc.Emit = func(n string) { app.Event.Emit(n) }`.

**Test analog:** `internal/services/settings_test.go` `newSettings(t, base)` fixture (fakeapi via httptest, `atomic.Int32` hit counter, events slice). Extend it with a temp `Dir` and Store; assert hit counts to prove TestKey bypasses cache.

---

### `internal/services/actions.go` (Klipy gate)

**Analog:** same file. Existing per-provider branch to mirror for the gate (line 124): `if r.Pasted && g.Provider == provider.Klipy {`. Add an early refusal in `Copy`, `Paste`, `Download`, `CopySquare`, `PasteSquare` when `g.Provider == provider.Klipy && !KlipyFileActions`. `DownloadsDir func() string` (line 50) already is a func — point it at `store.Get().DownloadDir` with default fallback. T5a string at line 161.

---

### `internal/services/panel.go` (+ OpenSettings)
```go
type PanelService struct {
	HideFunc func()
}
func (p *PanelService) Hide() {
	if p.HideFunc != nil {
		p.HideFunc()
	}
}
```
Add `OpenSettingsFunc func()` + `OpenSettings()` identically; wire `panel.OpenSettingsFunc = sh.OpenSettings` in main.go next to `panel.HideFunc = sh.Hide`.

---

### `internal/shell/shell.go` (window, event-driven)

**Analog:** same file lines 31-57 (panel creation and close hook) — copy for the settings window (normal window, not panel class; see RESEARCH Pattern 1):
```go
s.panel = app.Window.NewWithOptions(application.WebviewWindowOptions{
	Name:            "panel",
	Width:           PanelWidth,
	Height:          PanelHeight,
	URL:             "/",
	Frameless:       true,
	Hidden:          true,
	DisableResize:   true,
	...
})
s.panel.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
	e.Cancel()
	s.panel.Hide()
})
```
Tray placeholder to replace (lines 72-75):
```go
menu.Add("Settings…").OnClick(func(*application.Context) {
	s.Show()
	app.Event.Emit("panel:status", "Settings window ships in Phase 4")
})
```
Show/focus sequence + locking (lines 108-123): `show()` is the unlocked inner, public methods take `s.mu`; window callbacks must NOT take `s.mu`. Position via `s.plat.CursorScreenWorkArea()` + `panelOrigin` (`internal/shell/position.go`). Do not add File/Window menus (comment lines 59-64).

### `internal/shell/hotkeys.go`
**Analog:** shell.go lines 79-87:
```go
app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
	if err := app.GlobalShortcut.Register(Hotkey, s.Toggle); err != nil {
		log.Printf("hotkey register error: %v", err)
		return
	}
	log.Printf("hotkey registered: %s", Hotkey)
})
```
Replace `app.GlobalShortcut` with a `Registrar` interface field so tests and `-tags server` inject a fake. Keep `Hotkey` const as default. Theme application (`platform.SetAppearance`, `SetBackgroundColour`) also goes in this ApplicationStarted handler before any window shows.

**Test analog:** `internal/shell/position_test.go` (pure-logic tests, no Wails app) — keep Rebind/rollback logic in functions testable with the fake Registrar.

---

### `internal/platform/*` (+ SetAppearance, SystemLocale)

Add to the interface in `platform.go`, a stub line in `platform_stub.go` (`func (stub) SetAppearance(string) error { return ErrUnsupported }`), a recording method in `fake.go`:
```go
func (f *Fake) OpenURL(u string) error {
	f.Note("OpenURL:" + u)
	return f.OpenErr
}
```
Objective-C analog (`platform_darwin.m` ~203-211), always via `onMain` + `@autoreleasepool`:
```objc
int gsOpenAXSettings(void) {
    __block int ok = 0;
    onMain(^{ @autoreleasepool {
        NSURL *u = [NSURL URLWithString:@"x-apple.systempreferences:..."];
        ok = [[NSWorkspace sharedWorkspace] openURL:u] ? 1 : 0;
    }});
    return ok;
}
```
Go wrapper analog (`platform_darwin.go` `OpenURL`): `C.CString` + `defer C.free`, `0` → fixed-text error. Declare in `platform_darwin.h`. Strings returned from C use `strdup` and are freed by Go (see `gsFrontmostApp`). Build tags: `//go:build darwin && !server` / `!darwin || server`.

---

### Providers (`giphy.go`, `klipy.go`, `provider.go`)
Current hard-codes to parameterise (read prefs per request via an injected `Params func() provider.Params`, same shape as the existing `Key func() (string, bool)` field):
```go
// giphy.go 170-172
v.Set("limit", strconv.Itoa(provider.PageSize))
v.Set("offset", strconv.Itoa(offset))
v.Set("rating", "pg-13") // always sent: Clips search defaults to R
// klipy.go 118-119
v.Set("per_page", strconv.Itoa(provider.PageSize))
v.Set("content_filter", "medium")
```
Wiring analog in main.go: `Key: keys.Func(provider.GIPHY)`. Tests: existing `giphy_test.go` / `klipy_test.go` with `testdata/*.json`.

### `internal/search/search.go` (+ Flush)
Copy `Invalidate` (lines 80-91) minus `delete(s.blocked, …)`; keep the `s.gen[...]++` bump so late answers are dropped.

---

### `main.go` / `wire_*.go`
Add fields to `deps` (main.go lines 35-46) — e.g. `PickFolder`, `Autostart`, `Registrar` — and set real ones in `wire_desktop.go`, fakes in `wire_server.go` following:
```go
// Never the real pasteboard or browser; Trusted false, so paste reports needsAccessibility.
Platform: &platform.Fake{},
```
Services are registered untagged in `main.go` `Services: []application.Service{...}` so bindings generate without tags. Event emitters wired after `application.New` (pattern: `settingsSvc.Emit = func(n string) { app.Event.Emit(n) }`).

---

### Frontend entry: `frontend/settings.html`, `frontend/src/settings/main.ts`, `vite.config.ts`
```ts
// src/main.ts
import { mount } from 'svelte'
import './app.css'
import App from './App.svelte'
mount(App, { target: document.getElementById('app')! })
```
vite.config.ts keeps `server` + `wails("./bindings")`; add `build.rollupOptions.input: { main: 'index.html', settings: 'settings.html' }`.

### Settings components (`SettingsApp.svelte`, panes, `KeyRow`, `HotkeyRecorder`)
**Analog:** `frontend/src/components/SetupCard.svelte` — Svelte 5 runes (`$props`, `$state`, `$derived`, `untrack`), all backend calls via `../lib/api`, key value kept only while typed:
```svelte
import { hide, openExternal, providerError, saveKey } from '../lib/api';
import { keyMessage } from '../lib/errors';
let { slot, initialMessage = '', onsaved, ontryfiner }: { ... } = $props();
let value = $state('');
let verifying = $state(false);
```
Error copy goes through `lib/errors.ts` (`keyMessage`, `providerError`). Use `app.css` tokens.

### `frontend/src/lib/api.ts`
Thin facade; add `SettingsService.*` wrappers in the same one-liner style. Extend the allow-list only with fixed constants:
```ts
const EXTERNAL = new Set([GIPHY_KEY_URL, KLIPY_KEY_URL, CLIPS_ACCESS_MAILTO]);
export function openExternal(url: string) {
  if (!EXTERNAL.has(url)) throw new Error('refusing to open an unlisted URL');
  return Browser.OpenURL(url);
}
```

### `frontend/src/lib/hotkeys.ts`, `theme.ts`, `prefs.ts` + tests
**Analog:** `lib/actions.ts` header rule ("No Svelte or DOM imports so node --test can load it") and `lib/grid.test.ts`:
```ts
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { move, nearEnd, routeKey } from './grid.ts';
```
`package.json` `test` script globs `src/lib/*.test.ts` — keep new tests in `src/lib/`.

### Theme (`app.css`, `Footer.svelte`)
Replace `@media (prefers-color-scheme: dark) { :root {...} }` (app.css:27) with `:root[data-theme="dark"]`; Footer.svelte:32 `<source media="(prefers-color-scheme: dark)">` becomes a script-chosen src.

### `scripts/uat/15-settings*.js`
**Analog:** `scripts/uat/05-setup.js` — header comment with run command, `fail()` helper, `pageerror` collection, `page.goto("http://127.0.0.1:9245/")` (use `/settings.html`), `noKeyInPage()` check, never clicks external links. Run via `scripts/smoke-browser-loop.sh --uat`.

### `build/darwin/Taskfile.yml` / `Info.plist`
**Analog:** `install:dev` task (fail-hard steps, path guard `case ... esac`, `wails3 tool sign --identity "{{.SIGN_IDENTITY}}" --hardened-runtime`, `ditto`). Add `cp LICENSE ".../Contents/Resources"` in `create:app:bundle` next to `cp build/darwin/icons.icns`. Info.plist: bump `CFBundleShortVersionString`/`CFBundleVersion` 0.1.0 → 1.0.0; add usage strings beside `NSDownloadsFolderUsageDescription`. Edit by hand; never run `common:update:build-assets`.

## Shared Patterns

### Secrets never reach the webview / logs
**Source:** `internal/services/settings.go` comment + `kindOf`; `internal/settings/settings.go` fixed-text errors. Apply to KeyStatus (last4 only), TestKey, RemoveKey.

### Interface + stub + fake for anything native or Wails-only
**Source:** `internal/platform/platform.go`, `platform_stub.go`, `fake.go`; `wire_server.go`. Apply to Registrar, Autostart, PickFolder, SetAppearance — server build must never touch real OS state.

### Main-thread native calls
**Source:** `platform_darwin.m` `onMain` (CFRunLoopPerformBlock + semaphore, never `dispatch_sync`). Apply to SetAppearance.

### Event propagation
**Source:** `SettingsService.Emit` / `app.Event.Emit("providers:changed")`, `shell.show` emits `"panel:show"`. New: `settings:changed`, `settings:show`, `hotkey:test`, `panel:open-tab`.

### Atomic persistence
**Source:** `settings.WriteJSONAtomic` (0600 file / 0700 dir, temp+rename). All prefs writes through `Store.Update`.

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| download-dir validator (`Statfs` + `os.Link` probe) | utility | file-I/O | No filesystem-type probing exists; use RESEARCH Pattern 6 (only the temp-file cleanup idiom carries over) |
| `build/darwin/dmg-background.png` | asset | — | New artwork; geometry from RESEARCH Pattern 9 (540×380, icons at 151,164 / 388,164) |
| README / release notes / Klipy clearance email draft | docs | — | No prior docs beyond LICENSE |

## Metadata

**Analog search scope:** `internal/`, `frontend/src/`, `frontend/vite.config.ts`, `scripts/uat/`, `build/darwin/`, `main.go`, `wire_*.go`
**Files scanned:** ~35
**Pattern extraction date:** 2026-10-18
