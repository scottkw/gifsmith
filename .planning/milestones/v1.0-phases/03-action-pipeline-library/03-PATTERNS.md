# Phase 3: Action Pipeline & Library - Pattern Map

**Mapped:** 2026-10-17
**Files analyzed:** 30 (new + modified)
**Analogs found:** 28 / 30

All analog paths below are git-tracked source (verified via `git ls-files`).

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/services/actions.go` (MOD) | service (bound) | request-response + file-I/O | itself (lines 1-140) | exact |
| `internal/services/actions_test.go` (MOD) | test | — | itself (lines 1-50 fixture) | exact |
| `internal/services/library.go` (NEW) | service (bound) | CRUD | `internal/services/settings.go` | role-match |
| `internal/services/library_test.go` (NEW) | test | — | `internal/services/settings_test.go` | role-match |
| `internal/services/errors.go` (NEW, ActionError) | utility | transform | `internal/provider/provider.go` lines 51-89 (`Error`, `MarshalError`) | exact |
| `internal/library/store.go` (NEW) | store/model | file-I/O CRUD | `internal/settings/settings.go` lines 130-170 | role-match |
| `internal/library/store_test.go` (NEW) | test | — | `internal/settings/settings_test.go` | role-match |
| `internal/settings/settings.go` (MOD: `WriteJSONAtomic`) | utility | file-I/O | itself `SavePrefs` lines 142-170 | exact |
| `internal/staging/staging.go` (MOD: kind/MP4, Save, Clear, Remove) | service | file-I/O + HTTP | itself `Fetch`/`Sweep` | exact |
| `internal/staging/staging_test.go` (MOD) | test | — | itself | exact |
| `internal/imaging/square.go` (NEW) | utility | transform | none in repo — RESEARCH Code Example 1 | none |
| `internal/imaging/square_test.go` (NEW) | test | — | `internal/staging/staging_test.go` (table + t.TempDir style) | partial |
| `internal/platform/platform.go` (MOD: CopyText, RevealFile, OpenURL) | interface | — | itself lines 21-39 | exact |
| `internal/platform/platform_darwin.go` / `.m` / `.h` (MOD) | native glue | event/IPC | `CopyFile` / `gsCopyFile` / `gsOpenAXSettings` | exact |
| `internal/platform/platform_stub.go` (MOD) | stub | — | itself | exact |
| `internal/platform/fake.go` (MOD) | test double | — | itself | exact |
| `internal/platform/platform_darwin_test.go` (MOD) | test | — | itself (`pasteboardFileURLs` round-trip) | exact |
| `internal/provider/provider.go` (MOD: Gif +7 optional fields) | model | — | itself lines 33-40 | exact |
| `internal/provider/{giphy,klipy,finergifs}/*.go` + tests (MOD mappers) | adapter | transform | themselves | exact |
| `internal/provider/fakeapi/fakeapi.go` (MOD fixtures, mp4, wide GIFs) | test fixture server | request-response | itself | exact |
| `main.go`, `wire_desktop.go`, `wire_server.go` (MOD) | config/wiring | — | themselves | exact |
| `frontend/src/lib/actions.ts` (+test) (NEW) | utility (pure) | transform | `frontend/src/lib/grid.ts` / `providers.ts` | role-match |
| `frontend/src/lib/library.ts` (+test) (NEW) | utility (pure) | transform | `frontend/src/lib/grid.ts` (`move`) | exact |
| `frontend/src/lib/format.ts` (+test) (NEW) | utility (pure) | transform | `frontend/src/lib/errors.ts` (+`errors.test.ts`) | role-match |
| `frontend/src/lib/popover.ts` (+test) (NEW) | utility (pure) | transform | `frontend/src/lib/grid.ts` | role-match |
| `frontend/src/lib/grid.ts` / `grid.test.ts` (MOD) | utility | event routing | itself | exact |
| `frontend/src/lib/api.ts` (MOD: new bindings wrappers) | client | request-response | itself lines 9-65 | exact |
| `frontend/src/lib/providers.ts` (MOD: library tabs ⌘5/⌘6) | config | — | itself | exact |
| `frontend/src/components/ActionMenu.svelte` (NEW) | component | event-driven | `StateMessage.svelte` (props) + `ProviderTabs.svelte` | role-match |
| `frontend/src/components/DetailsPane.svelte` (NEW) | component | — | `StateMessage.svelte` | role-match |
| `frontend/src/components/RecentSearches.svelte` (NEW) | component | event-driven | `ProviderTabs.svelte` | role-match |
| `frontend/src/components/LibraryGrid.svelte` (NEW or ResultGrid ext) | component | virtualized list | `ResultGrid.svelte` | exact |
| `frontend/src/components/Footer.svelte`, `Tile.svelte`, `App.svelte` (MOD) | component | — | themselves | exact |
| `scripts/uat/07-*.js` (NEW) | UAT test | — | `scripts/uat/02-keyboard.js` | exact |

## Pattern Assignments

### `internal/services/actions.go` (MOD)

**Struct + injection seams** (lines 31-38) — add `Recents`, `After func(time.Duration, func())`, `DownloadsDir func() string`, `lastDownload string` the same way `Sleep` is injected:
```go
type ActionService struct {
	Platform platform.Platform
	Staging  *staging.Store
	Hide     func()
	Previous func() platform.App
	Sleep    func(time.Duration) // nil means time.Sleep
	busy     sync.Mutex          // one action at a time: no racing pasteboard writes, no double ⌘V
}
```
**Busy lock idiom** (every bound method, lines 60-64):
```go
	if !a.busy.TryLock() {
		return Result{}, ErrBusy
	}
	defer a.busy.Unlock()
```
**Stage-then-sweep** (lines 42-56): copy to pasteboard first, sweep after success. Remove the `g.IsClip` → `ErrClipsCopy` guard (lines 20-21, 44-46).
**Paste sequence to extract into `paste(ctx, put func() error, okMsg string)`** (lines 76-124): AccessibilityTrusted(false) → Hide → sleep(PasteDelay) → Previous PID check → FrontmostApp → ActivateApp fallback → PasteShortcut; every failure after hide returns `Result{Copied:true, Message:...}` not an error. Fix line 88 text to `Enable Automatic Paste…`.
**Logging rule:** `log.Printf` with provider/kind only, never URLs (line 49).

### `internal/services/library.go` (NEW)

**Analog:** `internal/services/settings.go` lines 1-30 — struct of injected deps + `mu sync.Mutex` serializing mutations, doc comment stating what is never returned to JS:
```go
// SettingsService saves verified keys and the last provider. No bound method
// returns a key; errors and log lines name the provider and kind only.
type SettingsService struct {
	Keys      *settings.Keys
	Dir       string // holds settings.json
	...
	mu sync.Mutex
}
```
Methods take `ctx context.Context` first (Wails binding convention) and return `(Snapshot, error)`. No Wails import.

### `internal/services/errors.go` (NEW ActionError)

**Analog:** `internal/provider/provider.go` lines 51-89. Copy the `Kind` string type + struct with json tags + URL-free `Error()` + `MarshalError(err) []byte` returning nil when `errors.As` fails. Compose in `main.go` `MarshalError:` (currently `provider.MarshalError`, main.go ~line 82): try provider first, then action.

### `internal/library/store.go` (NEW)

**Analog:** `internal/settings/settings.go` lines 130-170.
Load (missing file = empty, not error):
```go
	b, err := os.ReadFile(filepath.Join(dir, prefsFile))
	if errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
```
Atomic write — generalize into `WriteJSONAtomic(dir, name string, v any) error`, adding `f.Sync()` before Close/Rename:
```go
	if err := os.MkdirAll(dir, 0o700); err != nil { return err }
	f, err := os.CreateTemp(dir, prefsFile+".*.tmp")
	...
	_, err = f.Write(b)
	if err == nil { err = f.Chmod(0o600) }
	if cerr := f.Close(); err == nil { err = cerr }
	if err == nil { err = os.Rename(tmp, filepath.Join(dir, prefsFile)) }
	if err != nil { _ = os.Remove(tmp) }
```
Diverge from LoadPrefs on parse failure: do not silently reset; mark file read-only-until-reload (RESEARCH Pitfall 11).

### `internal/staging/staging.go` (MOD)

- Fetch structure lines 46-122: allow check → timeout ctx → `hc := *s.HTTP` with `CheckRedirect` re-checking allowlist (≤5 hops) → 6-byte magic check (line 85-88) → hash dir `sha1(provider:id)[:16]` → `.part` + Rename. Add kind param: MP4 checks bytes 4..8 == `ftyp` (read 8 bytes).
- Fixed URL-free error vars (lines 24-30) — add new errors in the same block.
- `Sweep` (lines 125-143) is the template for `Clear()` (remove all entries) and `Remove(dir)` (Klipy grace delete; validate with `filepath.Rel` + `..` check like line 126-129).
- `Save` (downloads): reuse `slug` (lines 162-186); use RESEARCH Code Example 3 (`os.Link` no-overwrite claim).

### `internal/platform/*` (MOD: CopyText, RevealFile, OpenURL)

- Interface: add to `platform.go` lines 21-39 with one-line doc comments.
- `.m`: `gsCopyText(pbName, text)` copies `gsCopyFile` (platform_darwin.m:31-39) — caller's thread, `pasteboard(pbName)`, `clearContents`, `writeObjects:@[NSString]`/`setString:forType:NSPasteboardTypeString`. `gsReveal`/`gsOpenURL` copy `gsOpenAXSettings` (lines 181-188): `__block int ok`, `onMain(^{ @autoreleasepool {...} })`.
- `platform_darwin.go`: Go wrapper like `OpenAccessibilitySettings` (lines 75-80) and `copyFileToPasteboard` (C.CString + defer C.free, `== 0` → errors.New). Add `pasteboardString(name)` test helper mirroring `pasteboardFileURLs` (lines 103+).
- `platform_stub.go`: one-line `func (stub) CopyText(string) error { return ErrUnsupported }` etc.
- `fake.go`: `func (f *Fake) CopyText(s string) error { f.Note("CopyText:" + s); return f.CopyErr }` pattern (lines 29-32).

### `internal/services/*_test.go`

**Analog:** `internal/services/actions_test.go` lines 25-50 — `newService(t, f *platform.Fake)` builds an `httptest.Server` around `fakeapi.Handler`, staging with `Allow` restricted to the test host, `Hide: func(){ f.Note("hide") }`; assertions compare `f.Calls` slices. Inject `Sleep`/`After` as no-op or recording funcs. Run with `-race`.

### `main.go` / `wire_*.go` (MOD)

- `deps` struct (main.go:33-41): add `DownloadsDir string`, `Platform` choice if server uses Fake.
- `wire_server.go` lines 33-46: add `DownloadsDir: "bin/uat-downloads"`; in server build use `&platform.Fake{Trusted:false}` (stub currently returns ErrUnsupported for CopyFile).
- `wire_desktop.go`: `filepath.Join(home, "Downloads")`.
- Services list (main.go ~line 76-81): append `application.NewService(library)`; call `stage.Clear()` before `app.Run()`; late-bound funcs set after `shell.New` like `actions.Hide = sh.Hide`.
- Bindings regenerated into `frontend/bindings/gifsmith/internal/services/` (do not hand-edit).

### `frontend/src/lib/{actions,library,format,popover}.ts` (NEW)

**Analog:** `frontend/src/lib/grid.ts` — header comment "No Svelte or DOM imports so node --test can load it"; exported union types + pure functions. `library.ts` `moveSectioned` mirrors `move()` (grid.ts:9-23) switch on Direction with clamping.
**Tests:** `frontend/src/lib/grid.test.ts` lines 1-30:
```ts
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { move, nearEnd, routeKey } from './grid.ts';
test('move up by a row, first row does nothing', () => { assert.equal(move(2, 'up', 10), 2); });
```

### `frontend/src/lib/grid.ts` (MOD)

Extend `KeyAction` union (line 5) and `routeKey` (lines 32-42); add `shiftKey` to the event param type; follow existing order (IME guard first, `e.repeat` suppression for Enter). Existing `'actions'` (⌘K) and slot pattern: `if (e.metaKey && ['1','2','3','4'].includes(e.key)) return \`slot${e.key}\``— extend to 5/6. Test both `"c"` and `"C"`.

### `frontend/src/lib/providers.ts` (MOD)

`SLOTS` readonly array of typed objects (lines 20-25); add library tab entries with `shortcut: '⌘5'/'⌘6'` or a sibling `LIBRARY_TABS` array; widen `SlotId`.

### `frontend/src/lib/api.ts` (MOD)

One-line arrow wrappers over generated bindings (lines 9-18, 62-65): `export const copy = (g: Gif) => ActionService.Copy(g);`. Add `copySquare`, `copyText`, `download`, `library*`. Errors: extend `providerError` (lines 56-61) style with `actionError(e)` reading `cause.kind`.

### `frontend/src/components/LibraryGrid.svelte` (NEW) / ResultGrid

**Analog:** `ResultGrid.svelte` lines 1-80. Copy: `$props()` typed destructure; `createVirtualizer` with per-index `estimateSize: (i) => ...` (line 46) — use 32 for header rows, 152 for tile rows; `untrack(() => $virt.setOptions({count}))` effect (lines 56-59); scrollToIndex effect (60-63); `role="listbox"`, rows absolutely positioned with `translateY({row.start}px)`; `<Tile gif={g} index={i} selected=... />`.

### `ActionMenu.svelte`, `DetailsPane.svelte`, `RecentSearches.svelte` (NEW)

**Analog:** `StateMessage.svelte` lines 1-30 — leading comment naming UI-SPEC IDs, `let { ... }: { ...types; onX?: () => void } = $props();`, callback props named `on<verb>` (lowercase), ARIA role on root. Tabs/chips row: follow `ProviderTabs.svelte`. Keyboard handling stays in `App.svelte` via `routeKey` (focus never leaves the search input — see 02-keyboard UAT).

### `Footer.svelte` (MOD)

Lines 1-40: prop-driven `text`/`tone`, existing `needsAccessibility` button; add optional `actionLabel` + `onaction` (Undo / Show in Finder) beside the center text, same `<button onclick>` style; attribution prop follows selected item's provider.

### `scripts/uat/07-*.js` (NEW)

**Analog:** `scripts/uat/02-keyboard.js` lines 1-30: header with run command `bash scripts/smoke-browser-loop.sh --uat ...`, `fail()` helper, `pageerror` collection, viewport 720×480, `goto http://127.0.0.1:9245/`, wait `section.content[aria-busy=false]`, `page.keyboard.press("Meta+4")`. Headless only — never desktop input. Update `scripts/uat/01-tracer.js` E7 and `scripts/smoke-browser-loop.sh` 422 check when server build uses `platform.Fake`.

## Shared Patterns

### URL-free errors to the webview
**Source:** `internal/staging/staging.go` lines 23-30; `internal/provider/provider.go` lines 74-89. **Apply to:** staging, services, library, imaging. Fixed `errors.New` sentinel text; never wrap transport errors (they contain URLs); typed kinds via MarshalError → `err.cause` in JS.

### Allowlist injection
**Source:** `staging.Store.Allow` (staging.go:41) + `wire_server.go` lines 42-45. **Apply to:** link/markdown text validation, library Gif validation, Save downloads — reuse the same `d.Allow`.

### Atomic 0600 JSON writes
**Source:** `settings.SavePrefs` (settings.go:142-170). **Apply to:** favorites.json, recents.json, searches.json.

### Injectable time/side effects for tests
**Source:** `ActionService.Sleep` (actions.go:36) and `Hide` closure in tests. **Apply to:** Klipy grace `After`, DownloadsDir, Recents recorder.

### Busy try-lock
**Source:** actions.go:60-63. **Apply to:** every new ActionService method incl. Download (not ServiceShutdown).

### Wails isolation
Only `main.go` and `internal/shell` import Wails; new `internal/library`, `internal/imaging` must not.

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| `internal/imaging/square.go` | utility | transform | No image processing in repo; use RESEARCH.md Code Example 1 (lines 470-498) and Pitfalls 1-2 |
| Popover placement in `ActionMenu.svelte` | component | — | No existing popover/overlay; use `lib/popover.ts` clamp per UI-SPEC |

## Metadata

**Analog search scope:** `internal/**`, `main.go`, `wire_*.go`, `frontend/src/**`, `scripts/uat/**`
**Files scanned:** ~35
**Pattern extraction date:** 2026-10-17
