# Phase 1: Native Spike & Signed Dev Loop - Context

**Gathered:** 2026-10-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Prove on this Mac (macOS 27, Apple Silicon) the three unproven native behaviors the product depends on — typing into a non-activating hotkey panel, the panel appearing over full-screen apps, and copy/paste of a GIF *file* into the previously focused app — from a stable Developer ID–signed identity, plus a throwaway notarized build and a paste matrix, ending in an explicit user go/no-go. Requirements: SHEL-01..04, ACTN-01, ACTN-02. No providers, no real grid, no settings window (Settings menu item is a stub until Phase 4).

</domain>

<decisions>
## Implementation Decisions

### Spike Scope
- Spike code is the foundation, not throwaway: the Wails v3 scaffold, `internal/shell` (panel/hotkey/tray) and `internal/platform` (cgo: file-URL pasteboard write, CGEvent ⌘V, Accessibility check, frontmost app) become the real packages used from Phase 2 on.
- Enter = copy GIF file to clipboard; ⌘Enter = paste into the previous app (matches v1 default + sibling actions).
- Uses one small sample GIF bundled/embedded in the app — no provider or API key in Phase 1.
- Panel UI is minimal: a plain text input plus one GIF preview, no styling. Real UI is Phase 2.

### Hotkey & Panel
- Default global hotkey ⌃⌥⌘G via Wails v3 `app.GlobalShortcut` (Carbon), registered after the app is running so conflicts surface; user confirms no clash with Vorssaint.
- Panel opens on the display under the mouse cursor, horizontally centered, near the top (Spotlight-like).
- Fixed panel size 720×480.
- If a non-activating panel cannot take typed input without stealing focus, fall back to record-frontmost-app → activate panel → reactivate previous app before pasting (`platform.ActivateApp`), rather than redesigning.

### Signing & Verification
- Bundle id `com.kenscott.gifsmith`.
- Signed dev builds are installed to and run from `~/Applications/Gifsmith.app`, signed with `Developer ID Application: Ken Scott (S2K7P43927)`, so the Accessibility (TCC) grant survives rebuilds. Paste is never tested from `wails3 dev`.
- Notarization uses a notarytool keychain profile named `gifsmith-notary`, created once by the user (`xcrun notarytool store-credentials gifsmith-notary`, app-specific password). This is a user action checkpoint; Claude uses the profile afterwards.
- Paste matrix is a markdown checklist in this phase directory; user tests the apps they have (Slack, Messages, Mail, Discord, Notes, Telegram, Safari/Chrome composers, TextEdit), marking missing apps n/a. File-URL-only pasteboard payload unless the matrix proves otherwise.

### Carried from project decisions
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

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- Greenfield repo — no application code yet. Only `.planning/` exists.
- Reference implementation (data only): Raycast extension at `/private/tmp/claude-501/-Users-ken-dev-gif-search/9f53d405-1914-4df1-8002-37e6a551a349/scratchpad/rc/extensions/gif-search` — `src/components/GifActions.tsx` (copy/paste sequencing, usage tracking only after success), `tests/gifClipboard.test.mjs`.
- Research probes confirmed on this machine: `[NSPasteboard writeObjects:@[NSURL]]` yields `public.file-url`; Wails v3 `examples/spotlight` shows the NSPanel setup.

### Established Patterns
- See `.planning/research/ARCHITECTURE.md` (component split, `platform.Platform` interface with darwin/stub/fake) and `.planning/research/STACK.md` (versions, signing tasks, server-mode recipe).

### Integration Points
- `internal/platform.CopyFile` / `PasteShortcut` / `AccessibilityTrusted` / `FrontmostApp` are the primitives Phase 2–3 build on.
- `wails3 task darwin:sign:notarize` for the notarized dry run.

</code_context>

<specifics>
## Specific Ideas

- Panel should feel like Spotlight: appears instantly on the hotkey, near the top of the active screen, disappears on Esc/click-away.
- Machine already has: Go 1.27.2, Xcode 27.1 notarytool, Developer ID identity; `wails3` CLI is NOT installed yet (`~/go/bin/wails` is v2 — leave it).

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>
