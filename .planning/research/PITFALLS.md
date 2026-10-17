# Pitfalls Research

**Domain:** Go + Wails macOS menu-bar GIF search launcher (Gifsmith) — port of the Raycast "GIF Search" extension
**Researched:** 2026-10-17
**Confidence:** MEDIUM overall. Provider ToS, Wails source and Raycast extension findings are HIGH (read primary sources). Per-target-app paste/animation behavior and several macOS behaviors are MEDIUM or LOW because no authoritative documentation exists. Those are marked "verify in spike".

**Confidence tags used below:** [HIGH] read in primary source or code. [MEDIUM] consistent secondary sources or established macOS behavior. [LOW] inference. Must be proven by the native spike or a test.

**Phase names used for mapping** (the roadmapper may rename them, but the ordering is what matters):

| Tag | Phase |
|-----|-------|
| P1 | Native spike: hotkey, panel, copy GIF file, paste into TextEdit, signed `.app` |
| P2 | Provider clients and search core: GIPHY, Clips, Klipy, Finer Gifs, debounce, paging |
| P3 | Grid UI and panel UX |
| P4 | Actions and storage: copy/paste variants, favorites, recents, staging, download, square crop |
| P5 | Settings, permissions UX, keys |
| P6 | Signing, notarization, release |

---

## Critical Pitfalls

### Pitfall 1: Porting the extension's provider models verbatim (they target Raycast's proxy, not the real APIs)

**What goes wrong:**
`models/giphy.ts` and `models/klipy.ts` call `gif-search.raycast.com/api/{giphy,klipy}`, a Raycast-hosted proxy. The proxy returns shapes that differ from the real APIs:
- Klipy's mapper expects Tenor-v2-shaped JSON (`results[]`, `media_formats.{gif,nanogif,tinygif}`, cursor `next`, `pos`, `media_filter`). The real KLIPY API is `GET api/v1/{app_key}/gifs/search`. It uses `page`/`per_page` (min 8, max 50), returns `data.data[]` with `id`, `slug`, `title`, `file.{hd,md,sm,xs}.{gif,webp,jpg,mp4,webm}.{url,width,height,size}`, and `has_next`. Items are fetched by `slugs`, not ids. The key is in the URL path. [HIGH, docs.klipy.com]
- The extension's `type=videos` GIPHY Clips parameter is a proxy convention. The real endpoints are `/v1/clips/search` and `/v1/clips/trending`, and they are **"only available upon approval"** (email clips@giphy.com). A self-service beta key from a Gifsmith user will very likely not be able to call Clips at all. [HIGH, developers.giphy.com/docs/clips]
- The extension's Klipy changelog entries (2026-07-01) show that Klipy was moved onto Raycast's proxy specifically to avoid per-user keys. Our "bring your own key" model is the older, pre-proxy situation.

**Why it happens:**
The "portable" TypeScript looks self-contained, so the mapping code gets transliterated to Go. The proxy contract is invisible in the code.

**How to avoid:**
- Write each provider client against the vendor's documented API and add a provider-neutral `Gif` domain type (`provider`, `id`, `title`, `previewURL`, `fullURL`, `mp4URL?`, `pageURL`, `w`, `h`, `bytes`). Mappers produce it. Nothing else knows vendor shapes.
- Record real JSON fixtures per provider (live key, one-time) and test the mappers against them. Do not hand-write fixtures from the extension code.
- Treat **GIPHY Clips as "requires an approved key"**. Ship the provider hidden or disabled behind a capability probe: call `/v1/clips/trending?limit=1` on key save, and show the provider only on 200. Do not make it a launch blocker.
- Use the Klipy `slug` as the stable id (the `id` is numeric and the items endpoint takes slugs).

**Warning signs:**
Klipy results are empty or the grid shows broken images. Clips returns 401/403 with a valid GIPHY key. Mapper code references `nanogif`, `tinygif`, `next`, `pos`, `media_filter`.

**Phase to address:** P2. Run a 30-minute live-API probe for all four providers before any UI work.

---

### Pitfall 2: Building features the provider terms forbid (caching, mixing providers, URL rewriting, missing attribution)

**What goes wrong:**
Both GIPHY and KLIPY publish "prohibited practices" that collide with features copied from the extension:

| Rule (verbatim sources) | Gifsmith feature at risk |
|---|---|
| GIPHY: "Do not cache media URLs or copies of GIPHY media assets unless … explicitly approved" [HIGH]. KLIPY: "Do not store, mirror, re-host, rewrite, or retain copies of KLIPY media" [HIGH] | Favorites byte cache (`cached-gifs/`), persisted `download_url`s in favorites/recents, disk cache of thumbnails, response caching of trending |
| GIPHY/KLIPY: do not mix provider content "in the same grid" / result set [HIGH] | The extension's aggregate **Favorites** and **Recents** views interleave all providers |
| GIPHY: "Do not strip or modify URLs … or query params while rendering"; KLIPY: do not alter/reconstruct URL parameters [HIGH] | `stripQParams` is fine for the *copied link text*, but must never touch URLs used for rendering or download |
| GIPHY: "Do not reorder content or filter out" Search/Trending results; KLIPY: do not reorder/insert/remove/suppress [HIGH] | `dedupe()` across pages is a mild violation; any local "frecency" sort of live results is a real one |
| GIPHY: "Powered By GIPHY" must be conspicuous wherever the API is used [HIGH]. KLIPY: **"Search KLIPY" must be the search placeholder (REQUIRED)**; "Powered by KLIPY" optional [HIGH] | Generic "Search GIFs…" placeholder; no GIPHY mark in the panel |
| GIPHY: "Do not proxy requests"; calls must be made from the client [HIGH] | Fine for a desktop app that calls the API itself. A hosted key proxy (already out of scope) would violate it |
| GIPHY: call the Action Register / analytics pingback on view, click and send ("highly recommend") [HIGH]. KLIPY: Share Trigger endpoint with a stable anonymous `customer_id` [HIGH] | Not implemented, and production-key review may look for it |

Note that a user's own beta key is what carries the risk (revocation), not a Gifsmith server. But a public release that ships features contrary to the terms invites key revocation for every user.

**Why it happens:**
The Raycast extension used a partner key through the proxy and was effectively covered by Raycast's agreement. A copy-to-clipboard tool must physically download the media, so the terms look unenforceable and get ignored.

**How to avoid:**
- Make an explicit **compliance posture decision in P1/P2** and record it in PROJECT.md Key Decisions. Recommended posture:
  1. Grid thumbnails load straight from provider URLs in the WebView (no disk cache, no rewrite).
  2. The only copy of media on disk is the **transient staging file** needed to put a file on the pasteboard or save a download (Pitfall 4). It is short-TTL and never reused as a "cache".
  3. Favorites and recents persist `{provider, id/slug, title}` plus a short-lived URL hint, and re-resolve through the by-ID endpoint when displayed. Finer Gifs is exempt because it is a hobby API with deterministic URLs.
  4. Favorites and recents are **per-provider tabs/sections, never one interleaved grid**.
  5. Show "Powered by GIPHY" when the GIPHY provider is selected and use "Search KLIPY" as the Klipy placeholder.
  6. Keep response de-dup minimal (same id on the same page only) and never re-rank live results.
- If the author later wants a persistent favorites byte cache for offline/instant paste, make it an opt-in setting with a disclosure, not the default.
- Re-check the terms before release: the Best Practices page changed within the last year (it now has an "approved caching" section). [MEDIUM]

**Warning signs:**
A `cache/` directory that grows without bound. A "Favorites" view that renders GIPHY and Klipy tiles together. A placeholder string of "Search GIFs".

**Phase to address:** P2 (data model and policy decision), P3 (attribution UI), P4 (favorites storage).

---

### Pitfall 3: Choosing the wrong pasteboard representation, so GIFs paste as a still frame

**What goes wrong:**
macOS pasteboards can carry a file URL (`public.file-url`), raw GIF data (`com.compuserve.gif`), and image representations (TIFF/PNG). Receiving apps pick whichever type they prefer. Image-oriented paths frequently flatten to **the first frame**. Reports note this for Slack, WhatsApp and Imgur. [LOW/MEDIUM, secondary sources]

What is known for certain from the extension's own issue tracker:
- Copying a **file URL** and pasting into Slack produces a file upload that **shows the filename above the GIF** (raycast/extensions #14413). The author's workaround was a "Hide Filename" preference that renames to `gif.gif`. This confirms Slack consumes the file URL path and uploads the actual `.gif`. [HIGH for Slack on file-URL]
- `Clipboard.copy({file})` is the single mechanism the extension relies on for every target.

What is **not** documented anywhere and must be tested: Messages, Discord, Mail, Safari/Chrome web composers (Gmail, GitHub comment box), Notes, Telegram, Teams. Likely (inference, [LOW]): Messages and Mail accept a file URL as an attachment and animate; Chromium/Electron web apps expose it via `paste` event `clipboardData.files`; apps that only read image types lose animation. If you add extra representations (GIF data, PNG/TIFF) for compatibility, some apps will prefer the image representation over the file URL and **break the Slack/Discord path that works today**. Do not add representations speculatively.

**Why it happens:**
"Copy a GIF" sounds like one operation. Autogenerated TIFF/PNG fallbacks (NSPasteboard promises/conversions) are easy to trigger accidentally with `writeObjects:[NSImage]`.

**How to avoid:**
- Default: write **only** `NSURL` file URL(s) via `writeObjects:@[fileURL]` (the same primitive Raycast uses). Do not pass an `NSImage`.
- In the P1 spike, build a **paste matrix** (app × representation × result: animated / still / attachment / failed). Test file-URL only, file-URL + `com.compuserve.gif`, GIF-data only. Apps: Slack, Discord, Messages, Mail, Notes, Telegram, Safari + Chrome (Gmail compose, GitHub comment). Commit the matrix to the repo as a doc.
- Keep the **link and Markdown actions** as first-class fallbacks (they already exist in the extension). When pasting into an app that mangles files, a pasted media link unfurls in Slack/Discord.
- GIPHY Clips copy an `.mp4` file. That uploads as video, not a looping GIF, in most chat apps. Surface this in the action label ("Copy clip (MP4)").
- Verify `changeCount` after writing, and verify that `declaredTypes` contains only what you intended (`pbpaste`-style inspection in the spike).

**Warning signs:**
Pasted GIF is static, is labeled `image.png`/`Pasted Graphic`, or arrives as a generic file icon with no inline preview.

**Phase to address:** P1 (matrix, decides default). P4 (implement chosen representation per action).

---

### Pitfall 4: Staged-file lifecycle bugs (ENOENT on paste, wrong GIF served, HTML saved as .gif)

**What goes wrong:**
The Raycast CHANGELOG records a cluster of bugs that all come from how files get staged and cached. [HIGH]
- 2026-09-17: "Paste GIF failing with a file-not-found error for every GIF" and later issue #31144 "ENOENT". Staged temp copies disappeared or were never created.
- "Copy, Paste and Download serving a previously cached GIF instead of the one selected when Hide Filename is on". The cache was keyed by display filename, so every GIF named `gif.gif` collided. They fixed it with `sha1(provider:id)` plus the original extension, and documented "display name must never be used as cache identity".
- "Saving the server's response as a GIF when the request failed" (HTML or JSON error body written to `x.gif` and put on the clipboard).
- "A download overwriting an existing file of the same name". They fixed it with exclusive-create (`O_EXCL`, `name (1).gif`).
- "A download that fails partway leaving the toast spinning". They added a pipeline and error propagation.
- Empty slug filenames: a title that slugifies to nothing gave a name of just `.gif`/`.mp4`. The extension needed `getFileExtension` special-casing because `path.extname(".mp4")` is `""`. In Go the equivalent trap is `filepath.Ext(".gif") == ".gif"` with an empty stem. The result is a hidden dotfile in Finder and a blank name in Slack.

Additional hazards specific to a pasteboard-with-file-URL design (the receiving app reads the file *after* ⌘V, sometimes seconds later when Slack/Discord upload):
- Deleting the staged file immediately after posting ⌘V breaks the upload. [MEDIUM]
- The extension's ASCII-only `slugify` strips CJK and emoji titles to empty. Long titles can exceed APFS's 255-byte filename limit.

**Why it happens:**
Display name, cache identity and temp path get conflated. Download success is not validated. Cleanup is eager.

**How to avoid:**
- Staging path: `…/staging/<sha1(provider:id)>/<display-slug>.<ext>`. A unique directory per GIF gives collision-free paths **and** a pretty filename for Slack. Hide-filename mode just uses `gif.gif` inside the unique directory, so there is no collision.
- Display name function: unicode-aware slug, fall back to `gif` if empty, cap at about 80 bytes, strip `/` and `:`. Unit-test with emoji-only, CJK, 300-char, and empty titles.
- Download validation: status == 200, `Content-Type` is `image/gif` or `video/mp4`, and magic bytes (`GIF87a`/`GIF89a`, or `ftyp` for MP4). Reject anything else before it touches the staging dir.
- Write atomically: stream to `*.part`, `fsync`, `rename`. Use `context.WithTimeout` on every download so the UI always receives success or error. Delete partial files on every error path.
- Serialize concurrent requests for the same key with `singleflight`.
- Staging TTL: sweep on app launch and on a timer, deleting entries older than about 24 hours, never "after paste".
- User downloads: `os.OpenFile(path, O_CREATE|O_EXCL)` with `name (n).ext` retry.

**Warning signs:**
Same GIF pasted for different selections. Slack shows a broken attachment or a 0-byte file. Files named `.gif`. A staging directory that never shrinks (and, conversely, intermittent ENOENT).

**Phase to address:** P4.

---

### Pitfall 5: Synthesized ⌘V that fails silently, fires the wrong shortcut, or races focus

**What goes wrong:**
Several independent failure modes sit behind the "Paste GIF" action:

1. **No Accessibility permission, no error.** `CGEventPost` without trust drops events silently. [MEDIUM] The only signal is `AXIsProcessTrusted()`.
2. **Keycode is layout-dependent.** Hardcoding `kVK_ANSI_V` (9) is physical-key based. On Dvorak, V is keycode 47, and key 9 yields a different letter, so ⌘V can become ⌘K or similar in the target app. Libraries such as Clipy/Sauce exist solely for this. [MEDIUM, Sauce README: v = 9 QWERTY / 47 Dvorak]
3. **Held modifiers leak.** If the user triggers the action with ⌘⏎/⇧⏎/⌥⏎ (the extension has secondary actions on modifier combos), the physical modifier is still down when ⌘V is posted. The target sees ⌘⇧V (paste-and-match-style or plain paste) or ⌘⌥V. [MEDIUM, long-documented Cocoa behavior]
4. **Focus/ordering race.** If the panel is still key or the previous app has not regained key status, ⌘V goes to Gifsmith itself. Raycast's own Paste is not immune: issue #30630 "Pasting markdown GIF breaks" is a `Keyboard.Shortcut.Virtualization.Error`, and #31004 "Enter only copies instead of pasting" is a paste silently not happening. [HIGH that these exist]
5. **Raycast bug history** also shows that separating copy from paste mattered: the 2026-09-18 change made Paste resolve the file and paste "directly, without copying first or closing the window before paste" (ordering hazards between close and paste).
6. **Secure Input** (password fields, Terminal "Secure Keyboard Entry") may suppress synthetic input. [LOW, verify]

**Why it happens:**
It works on the author's machine in TextEdit with QWERTY and no modifiers.

**How to avoid:**
- Use a **non-activating panel** (Pitfall 7). The previously frontmost app never loses key status, which eliminates the "restore focus, then wait" race. If you ever must activate Gifsmith, record `NSWorkspace.frontmostApplication` before showing and re-activate it, then wait for `NSWorkspaceDidActivateApplicationNotification` before posting. Do not use a fixed sleep as the only guard.
- Post ⌘V as: key-down and key-up for the keycode that produces "v" in the *current* input source (resolve via `TISCopyCurrentKeyboardLayoutInputSource` plus `UCKeyTranslate`, or fall back to 9 with a QWERTY-⌘ layout check), `flags = kCGEventFlagMaskCommand` explicitly, created from a **private** `CGEventSource` (`kCGEventSourceStatePrivate`) so hardware modifier state does not merge. Alternatively, poll `CGEventSourceFlagsState` until modifiers are released (cap about 300 ms).
- Gate on `AXIsProcessTrusted()` **before** posting. If untrusted or if the post path errors, leave the file on the pasteboard and show "Copied — press ⌘V" rather than failing the whole action. Copy must always work without Accessibility (already a PROJECT constraint).
- Do the pasteboard write and the ⌘V post in this order: write, verify `changeCount` advanced, (hide panel if activating), post ⌘V. Never clear or restore the pasteboard within less than about 1 s of posting, or the target may paste the old content. [MEDIUM]
- Avoid AppleScript/`osascript` "System Events" paste. It triggers a separate Automation (Apple Events) TCC prompt and needs an extra entitlement and `NSAppleEventsUsageDescription`. The extension moved *away* from AppleScript (CHANGELOG 2025-04-14).
- Test matrix for paste: QWERTY and Dvorak layouts; triggered via plain ⏎ and via ⌘⏎/⇧⏎; target apps from the Pitfall 3 matrix; a password field.

**Warning signs:**
"Paste does nothing" while "Copy" works. Pasted content is plain text or a different command fired. The feature works from `wails3 dev` but not from the bundled app.

**Phase to address:** P1 (prove paste into TextEdit with modifiers and Dvorak), P4 (final action set and fallbacks), P5 (permission UX).

---

### Pitfall 6: Accessibility (TCC) grant is tied to code signature, so ad-hoc/dev rebuilds silently revoke it

**What goes wrong:**
TCC records the grant against the app's code identity. Ad-hoc-signed or unsigned rebuilds mint a new identity (cdhash) each time. System Settings still shows the toggle **on**, but `AXIsProcessTrusted()` returns false. [MEDIUM, multiple independent sources; no Apple primary doc]
- The first call of `AXIsProcessTrustedWithOptions({prompt: true})` shows the system dialog; later calls are silent. Without an in-app deep link and polling, users see nothing happen.
- Running via `go run` / `wails3 dev` is an unbundled binary. Permission attaches to the terminal or IDE that launched it (the "responsible process"), not to Gifsmith. Granting Accessibility to Terminal "fixes" dev but masks every real problem. [MEDIUM]
- Keychain items (API keys) show the same pattern: a changed signature triggers repeated "Gifsmith wants to use your confidential information" prompts. [MEDIUM]
- Stale entries after re-signing require removing the row with "−", re-adding, and sometimes `tccutil reset Accessibility <bundle-id>` or a reboot.

**Why it happens:**
Developers test paste once, grant permission, then rebuild dozens of times with different signatures.

**How to avoid:**
- Sign **every** dev build with the same stable identity (Apple Development or Developer ID cert) and a fixed bundle identifier (`wails3 task darwin:sign` is available). Never ad-hoc for anything that tests paste or Keychain.
- Test paste only from the bundled `.app` in `/Applications` (or a fixed path), never from `wails3 dev`. Dev mode is for frontend work (dev-browser) only.
- Permission UX: (a) check `AXIsProcessTrusted()` at paste time, not at launch; (b) first time, call `AXIsProcessTrustedWithOptions` with prompt, then show an in-panel banner with a button that opens `x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility`; (c) poll `AXIsProcessTrusted()` every second while the banner is visible, and auto-dismiss when it flips; (d) add a "Fix stuck permission" help line (remove row, re-add, `tccutil reset`).
- Do not request Input Monitoring. Carbon hotkeys (Pitfall 10) and `CGEventPost` do not need it.

**Warning signs:**
Settings shows Gifsmith enabled but paste fails. The permission prompt appears repeatedly or not at all after a rebuild. Developers report it works for them after granting Terminal.

**Phase to address:** P1 (establish signed dev loop on day one), P5 (permission UX), P6 (re-verify on a clean user account with a notarized build).

---

### Pitfall 7: Panel window misbehaves over full-screen Spaces, on multiple displays, or steals focus

**What goes wrong:**
A plain Wails/NSWindow launcher will (a) open on the wrong Space, or switch Spaces, when a full-screen app is active; (b) steal key focus from the app the user was typing in (activating Gifsmith), breaking paste and causing a visible menu-bar flicker; (c) appear on the primary display instead of where the user is; (d) not dismiss on click-away. Known facts:
- Wails v3 now exposes the right primitives: `Mac.WindowClass = MacWindowClassPanel` with `PanelPreferences{FloatingPanel, NonActivating, BecomesKeyOnlyIfNeeded, UtilityWindow}`, `Mac.WindowLevel` (`floating` … `screenSaver`), `Mac.CollectionBehavior` (`CanJoinAllSpaces | FullScreenAuxiliary`, bitwise OR), and `Mac.ActivationPolicy = Accessory`. On show/focus, non-activating panels use `orderFrontRegardless` and skip `activateIgnoringOtherApps`. [HIGH, read wails master source and issues #5023/#5359]
- That support is recent (issues closed 2026-02 and 2026-05). Behavior at the edges is lightly tested. [MEDIUM]
- An Apple developer-forum report says `CanJoinAllSpaces + FullScreenAuxiliary` alone did **not** show an `LSUIElement` app's window over a full-screen app. The reliable recipe reported is accessory policy + **non-activating NSPanel** + a level at or above `screenSaver`, on macOS 26. [MEDIUM]
- AppKit forbids combining `CanJoinAllSpaces` with `MoveToActiveSpace` (throws). [MEDIUM]
- In `LSUIElement`/accessory apps, a non-activating panel whose `canBecomeKeyWindow` is false cannot take typed input. Wails' base window class returns YES for `canBecomeKeyWindow`; confirm the panel subclass does too. [MEDIUM, verify in spike]
- Because Gifsmith never becomes the active app, `window focus lost` events may never fire, so click-away dismissal and `document.hasFocus()`-dependent web behavior (autofocus, `:focus-visible`, caret) need explicit handling. [LOW, verify]

**Why it happens:**
Spotlight-like windows are one of the hardest AppKit patterns, and the Wails v3 panel work is new.

**How to avoid:**
- P1 spike builds the panel exactly as: accessory policy, `MacWindowClassPanel`, `NonActivating: true`, `FloatingPanel: true`, level `status` or `screenSaver`, collection behavior `CanJoinAllSpaces | FullScreenAuxiliary` (and consider adding `Stationary`/`IgnoresCycle`).
- Position the panel natively on the screen under the **mouse cursor** (`NSScreen` containing `NSEvent.mouseLocation`), not via a Wails center helper. Re-position on every show.
- Dismiss on Esc (handled by frontend), on selecting an action, and on outside click via a global mouse-down monitor (`addGlobalMonitorForEvents` for mouse events needs no extra permission). [MEDIUM]
- When the panel opens, drive focus from native (`makeKeyAndOrderFront` on the panel without activation), then have the frontend `focus()` the input on a "shown" event, not on mount.
- Test matrix to keep in the repo: Safari full-screen; Zoom/Keynote full-screen; Stage Manager on/off; Mission Control multiple Spaces; two displays with "Displays have separate Spaces" on and off; external display at a different scale; menu bar auto-hide.
- Fallback if non-activating text input fails: activate the app briefly, record and restore the previous app (Pitfall 5 sequence).

**Warning signs:**
Menu bar shows "Gifsmith" when the panel opens. Space switches. The panel is invisible over a full-screen app. Typing in the search box does nothing until clicked. The panel opens on the wrong display.

**Phase to address:** P1 (hard gate, because it determines the paste design), refined in P3.

---

### Pitfall 8: Betting the product on a fast-moving Wails v3 beta (and on macOS 27 regressions)

**What goes wrong:**
Wails v3 is **beta, not stable**. All GitHub releases are marked pre-release, and the latest are beta.21 (13 Sep) through beta.30 (15 Oct), about ten betas in a month. [HIGH] The docs say "API is stable but you may still encounter issues". Concrete risks seen in the issue tracker:
- macOS 27 breakages: tray left-click dead (#5752, #6147, fixed in beta.24 by `systemTrayCoerceEventType`); `mac:WindowDidUpdate` forwarded ~120 times/s and, on macOS 27, each `evaluateJavaScript:` wipes the click's transient activation so `navigator.clipboard.writeText()` rejects (#6170; fixed in beta.30). The host used for this project reports Darwin 27. [HIGH]
- Open at time of research: `ExecJS` over-release crash when called from main thread (#6214), tray updates not delivered while menu open (#6154), Fn/Globe key swallowed (#6033), WKWebView capped at 60 FPS on ProMotion (#6056), and binding generation / Vite import problems (#6215). [HIGH]
- The built-in updater has an open issue "verification is fail-open, so unsigned releases install even when Config.PublicKey is set" (#6186). [HIGH]
- Wails v2 has no global-shortcut API and no panel window class, so falling back to v2 means re-implementing hotkey and panel natively in cgo. [MEDIUM]

**Why it happens:**
Pre-release frameworks change underneath you, and macOS 27 itself is new.

**How to avoid:**
- **Choose v3, pinned to an exact beta in `go.mod`** (at least beta.30, which contains the macOS 27 fixes and non-QWERTY global-shortcut fix from beta.29). Upgrade deliberately, one beta at a time, rerunning the P1 smoke checklist.
- Keep a **15-minute native smoke script** (hotkey opens panel over full-screen, type, Enter copies, ⌘V pastes into TextEdit, tray menu opens, settings window opens) and run it on both macOS 26 and 27 after each Wails bump.
- Keep all Gifsmith native glue in one small Go package (`internal/native`) so a v2 fallback or Wails change touches one place.
- Do not use the Wails updater in v1. Ship manual updates (or add Sparkle later), and watch #6186.
- Do not rely on `navigator.clipboard` from the web layer for anything. Do clipboard writes from Go (Pitfall 9).

**Warning signs:**
`go get` of a newer beta changes the generated binding output or window behavior. Tray or hotkey regressions after an OS update.

**Phase to address:** P1 (pin and smoke script), re-verified every phase.

---

### Pitfall 9: cgo + Cocoa main-thread rules (and Wails owning the run loop)

**What goes wrong:**
Wails owns the NSApplication main-thread run loop. Any custom Objective-C you add runs inside that world.
- AppKit UI calls (NSWindow, NSPanel properties, status item, NSScreen queries that touch UI) from a Go goroutine crash or corrupt state. Goroutines are not on the main thread. [HIGH, Cocoa rule]
- `dispatch_sync(dispatch_get_main_queue(), …)` from the main thread deadlocks. Wails' own `InvokeSync*` family is re-entrancy safe: `dispatchOnMainThread` runs inline if already on main, otherwise posts via `CFRunLoopPerformBlock` on common modes, so it also works while a modal or menu is up. Use it. [HIGH, read source]
- Libraries that need their own main loop (`golang.design/x/hotkey` via `mainthread.Init`, robotgo) conflict with Wails' loop. Using them would duplicate Carbon handlers or crash. Wails v3 ships its own `app.GlobalShortcut` (Carbon `RegisterEventHotKey`). [HIGH]
- Go callbacks exported to C run on arbitrary threads. A Go `panic` crossing into C aborts the process. [HIGH, cgo rule]
- `NSPasteboard` is usable off the main thread, but mixing AppKit view calls into the same helper is how people end up with "only on the main thread" assertions. [MEDIUM]
- Objective-C object lifetimes: with ARC off in cgo files, retain/release mistakes show up as crashes much later (cf. Wails #6214, an over-release in `ExecJS`). [HIGH]

**Why it happens:**
It "works" in a quick test because the call happens to occur on the main thread.

**How to avoid:**
- Rule for the repo: **no AppKit window/status-item/screen call outside `application.InvokeSync*`**. Put native code behind a Go interface (`Panel.Show/Hide`, `Pasteboard.WriteFile`, `Keys.PasteCmdV`) and wrap with `InvokeSync*` in the Go layer, not in C.
- Pasteboard writes and `CGEventPost` may run on a worker goroutine. Keep them in separate C functions from window code.
- Compile native files with `-fobjc-arc` where possible. Wrap exported Go callbacks with `defer recover()`.
- Do not add any second hotkey/event-loop library.
- Match Wails' `-mmacosx-version-min` in your cgo flags (Wails uses 10.13 in its own `#cgo CFLAGS`) so the linker does not warn or mix targets. Decide the real minimum macOS (suggest 13+) and set it consistently. [MEDIUM]

**Warning signs:**
Console message "should only be invalidated on the Main Thread". Intermittent crash on first show after sleep. Deadlock when invoking from a menu callback.

**Phase to address:** P1.

---

### Pitfall 10: Global hotkey conflicts, silent registration failure, and picking a bad default

**What goes wrong:**
- Carbon hotkeys are first-come: another app (Raycast, Alfred, **Vorssaint**, ChatGPT desktop, Spotlight ⌘Space, input-source ⌃Space) may already own the combination. Wails v3's `GlobalShortcut.Register` returns an error **only if the app is already running**. If you register before `App.Run`, the OS rejection is delivered to the app error handler and not returned. [HIGH, read source]
- macOS 15.0 made `RegisterEventHotKey` fail (-9868) for combinations with only Option or Option+Shift. Apple said 15.2 allows them again, but later reports (Nov 2025) describe regressions after screen unlock. Do not default to Option-only. [MEDIUM]
- Beta.29 fixed non-QWERTY layout handling. Earlier betas map letters to physical key positions. [HIGH]
- Hotkeys that collide with in-app shortcuts of other apps (⌘⇧G in Finder/Chrome, ⌘⇧P everywhere) shadow them globally with no warning.
- A settings "hotkey recorder" that unregisters the old binding before the new one succeeds leaves the user with **no** hotkey and no UI to recover.

**Why it happens:**
Carbon gives no visibility into who owns a combo, and the failure path is easy to miss.

**How to avoid:**
- Register after the app is running, check the returned error, and surface it in the menu-bar menu and Settings ("Hotkey ⌃⌥G is in use by another app").
- Default to a three-modifier combo that mainstream launchers do not use (for example ⌃⌥⌘G). Never ⌘Space, ⌥Space or Option-only. Keep the tray icon menu item "Open Gifsmith" as a hotkey-free fallback so a failed hotkey never bricks the app.
- Hotkey recorder: register new first, unregister old only on success, persist only on success. Validate that at least one of ⌘ or ⌃ is present.
- Test on Dvorak/AZERTY. Avoid CGEventTap-based hotkeys (they require Input Monitoring).

**Warning signs:**
Hotkey works on first run, then stops after another launcher starts. User says "nothing happens" on a fresh machine.

**Phase to address:** P1 (register and error path), P5 (recorder UI).

---

### Pitfall 11: GIPHY/Klipy rate limits (100 requests/hour) destroyed by naive search-as-you-type and paging

**What goes wrong:**
- GIPHY beta keys: **100 searches/API calls per hour**, search `limit` **max 50**, HTTP 429 when exceeded. [HIGH, developers.giphy.com] KLIPY testing keys: **100 requests per hour**. [HIGH, docs.klipy.com]
- Both quotas are per key across *all* calls: trending on every panel open, search per keystroke, each infinite-scroll page, by-ID favorites lookups, Clips probe. A user burns 100 in a few minutes of normal use.
- Cancelling an in-flight request (`ctx` cancel) saves nothing: the server already counted it.
- The extension's pagination `hasMore = results.length === limit` fails after de-duplication or when a provider returns fewer items (changelog 2026-07-02 and 2026-09-30 fixed "pagination for providers without a cursor" and "invalid Max Results stops scrolling"). Klipy's `per_page` minimum is 8, so "Max results = 5" would 400. GIPHY offset max is 4999 and `q` max length is 50 chars for Clips. [HIGH]
- Trending is the first thing every panel open requests.

**Why it happens:**
Raycast's partner key made quotas invisible.

**How to avoid:**
- Debounce at **400–500 ms**, minimum 2 characters, and search on Enter instead of keystroke as a setting. Cancel superseded in-flight requests.
- Short-lived in-memory cache keyed `(provider, q, page)` with about 5 min TTL for identical repeats. Cache trending for 15–30 min (response cache only; media stays uncached per Pitfall 2).
- Do not fetch trending on open if the user is mid-query. Fetch page 2 only on scroll-near-end, with a single in-flight guard.
- Compute `hasMore` from the provider's own pagination fields (`pagination.total_count/offset/count` for GIPHY, `has_next` for Klipy). Clamp requested page size to per-provider limits (GIPHY ≤50, Klipy 8–50). Validate numeric settings and fall back to a default (as the extension eventually did).
- Keep a local call counter per provider per rolling hour. At about 80% show a subtle "n/100 this hour" indicator. On 429 show a clear message ("GIPHY beta key limit reached; resets within the hour") and disable auto-paging. Do not retry in a loop.
- Document in onboarding that users should request a production key if they hit limits, and that the Finer Gifs provider needs no key.

**Warning signs:**
429s after a few minutes of use. Network log shows a request per keystroke. Scrolling stops at an arbitrary page.

**Phase to address:** P2.

---

### Pitfall 12: A grid of full-size animated GIFs in WKWebView burns memory and CPU

**What goes wrong:**
Every visible animated GIF is decoded and re-rendered per frame, so CPU/GPU stays busy while on screen. [MEDIUM] WKWebView runs in a separate `com.apple.WebKit.WebContent` process, so memory shows up there, not in your app. A 50-tile grid with `original`/`hd` renditions (Klipy hd GIF mean 3.9 MB, median 2.6 MB per item) is tens to hundreds of MB downloaded and decoded. [HIGH for sizes, from docs.klipy.com format sizes] Infinite scroll compounds it. A hidden window can keep decoded frames and timers alive.

**Why it happens:**
`<img src=original.gif>` is the obvious implementation, and Raycast hid this (CHANGELOG 2025-12-17 "Faster thumbnails in large grid items", 2022 "use original gif files for Large grid items").

**How to avoid:**
- Grid tiles use **small renditions only**: GIPHY `fixed_width_small`/`preview_gif`/`fixed_height_small`; Klipy `xs`/`sm`. Use the full-size file only for copy/paste/download.
- Prefer **MP4/WebM `<video autoplay muted loop playsinline>`** for tiles where offered. GIPHY's Best Practices explicitly recommend MP4 for GIFs; Klipy's mp4/webm are 5–25× smaller than gif. [HIGH]
- **Virtualize** the grid (mount only rows near the viewport). Use `loading="lazy"` and `decoding="async"`. Animate only the focused/hovered tile and a handful of visible ones, and show GIPHY's still renditions for the rest. [MEDIUM]
- On panel hide, unmount tiles and clear `src` to free decoded frames and cancel loads. On show, restore from the in-memory result list (no new API call).
- Cap page size (default 20–25), cap total mounted tiles (about 100), and drop pages beyond that.
- Set `prefers-reduced-motion` handling.
- Never pass GIF bytes through Wails bindings (base64 over JSON IPC). Pass URLs and IDs; Go downloads natively.
- Profile with Activity Monitor (WebContent process) and Instruments in P3. Define a budget (for example under 400 MB WebContent after 5 pages, under 10% CPU idle with the panel open).

**Warning signs:**
Fans spin when the panel is open. WebContent process grows by hundreds of MB. Scrolling stutters. Panel is slow to open after long idle.

**Phase to address:** P3.

---

### Pitfall 13: Square-crop that degrades quality or breaks animation

**What goes wrong:**
The extension's crop shells out to Swift/ImageIO: it composites each frame into an RGBA `CGContext`, then re-encodes with `CGImageDestination`. That full re-quantization is the quality-loss path (banding, bloated files, possible matte fringing on transparency). A Go port has additional traps:
- `image/gif.Encode`/`EncodeAll` on non-paletted images quantizes to the fixed Plan9 palette with Floyd–Steinberg. That is visibly bad for photographic GIFs. [MEDIUM, Go stdlib behavior]
- `gif.DecodeAll` returns delta frames with per-frame bounds, disposal methods, local palettes and transparency. A naive "decode, crop each paletted frame, re-encode" breaks frames that do not cover the full canvas. [HIGH]
- Delay quirks: browsers/ImageIO clamp very small delays (≤10 ms) to 100 ms; copying `Delay` straight can change speed. `LoopCount` semantics differ (0 = infinite, -1 = once). [MEDIUM]
- The extension's MP4 path: Clips are `.mp4`, so "square GIF" of a Clip needs video decoding. The ImageIO script cannot read an MP4. The extension's changelog shows only the *filename* being fixed to `.gif`. [MEDIUM]
- Decoding all frames of large GIFs into RGBA can use hundreds of MB.

**Why it happens:**
"Decode, crop, encode" is the obvious pipeline, and it re-quantizes.

**How to avoid:**
- Prefer a **palette-preserving, geometry-only crop** (what `gifsicle --crop` does): keep each frame's palette and indexed pixels, intersect each frame's rectangle with the crop window, translate offsets, and `SubImage`. Frames fully outside the window become a 1×1 transparent placeholder with the original delay. Leave disposal and delay untouched, and update the logical screen size. No compositing, no re-quantization, no loss. [LOW/MEDIUM, reasoned, verify visually]
- If compositing is unavoidable (for example frames of differing palettes plus transparency), re-encode with `Paletted` frames built by a proper quantizer (median-cut/`color/palette` with per-frame local palette), not Plan9.
- Disable square actions for MP4 Clips (or use the clip's GIF rendition if present).
- Golden tests: crop a fixed 3-GIF corpus (transparent, delta-frame, many-frame) and compare frame count, delays, loop count, and per-pixel diff against a reference render.
- Cap input size (for example 25 MB / 300 frames) and report a clear error beyond that.

**Warning signs:**
Cropped GIF is larger than the original, shows banding, flickers, or plays at a different speed. Missing frames at borders.

**Phase to address:** P4.

---

### Pitfall 14: Favorites/recents data model that only stores IDs (and trusts the network)

**What goes wrong:**
The extension stores `id` lists per service in LocalStorage and re-fetches full objects through the provider on every view. Its changelog and code show the consequences:
- IDs unique only within a provider; "Favorites tracked per service, so GIFs that share an ID across providers no longer affect each other". [HIGH]
- "A provider that can't be reached no longer empties Favorites and Recents for every other provider", and a code comment that "every populated provider failing looks identical to having saved nothing and the empty result would be cached over the last good one". [HIGH]
- `getServiceFromUrl` infers the provider from URL substrings (`giphy.com/gifs` vs `/clips`, `finergifs`). An unmatched URL returns `null` and silently skips tracking. Fragile. [HIGH]
- Finer Gifs lookup is `size=10`, so only the first 10 favorites load (the extension has an issue titled "Show all your favourite GIFs instead of just 10", #15808). [HIGH]
- Each Favorites view costs by-ID API calls against a 100/hr quota (Pitfall 11).
- Race: unfavoriting while a copy/download is in flight can recreate a cache entry that removal just evicted (comment in `resolveGifFile.ts`, "Membership is re-read rather than passed in").
- Recents order semantics: recents prepend and re-visit moves to top; favorites append (changelog 2026-04-23).

**How to avoid:**
- Store full records locally: `{provider, id, title, pageURL, tags, savedAt, lastUsedAt, kind}` in one SQLite or JSON-with-atomic-write file with a single owner goroutine (no concurrent read-modify-write). Primary key `(provider, id)`. Store `provider` explicitly, never infer from URLs.
- Refresh media URLs lazily and in the background. Treat network failure as "keep showing local records", never "replace with empty". Do not write through an empty result produced by an error.
- Finer Gifs: derive URLs from `fileid` and skip lookups (store the title).
- Check "still a favorite" under the same lock that writes any cache/staging artifact, and evict by digest prefix (all variants) on removal. (The extension learned that GIPHY can switch a Clip between MP4 and GIF.)
- Schema version field plus migration stub from day one.

**Warning signs:**
Favorites shows fewer items than saved. Offline, Favorites appears empty. Same-ID GIFs across providers affect each other.

**Phase to address:** P2 (schema), P4 (actions).

---

### Pitfall 15: Notarization/hardened runtime with cgo and Wails packaging

**What goes wrong:**
Locally built apps never undergo Gatekeeper assessment, so "it runs here" says nothing about downloaded builds. Gotchas:
- Notarization requires Developer ID signing with `--options runtime` (hardened runtime) and a secure timestamp on every Mach-O in the bundle, signed inside-out. Blanket `--deep` is discouraged. [MEDIUM]
- Wails v3 provides `wails3 task darwin:sign` / `darwin:sign:notarize`, `wails3 setup signing` (wraps `notarytool store-credentials`), and a **commented-out** `ENTITLEMENTS` var pointing at `build/darwin/entitlements.plist`. The docs do not describe what to put there. [HIGH on tooling, gap on content]
- A cgo Go binary plus WKWebView needs **no** JIT/unsigned-memory entitlements (WebKit runs out of process). Adding `allow-jit` or `disable-library-validation` "just in case" weakens the notarized app and can trigger review friction. Only add entitlements you can justify. [MEDIUM]
- Do not add Apple Events usage (Pitfall 5) unless you need it.
- `LSUIElement=true` must be in the generated `Info.plist` (and `ActivationPolicyAccessory` at runtime), or a Dock icon flashes on launch. Check `build/darwin/Info.plist`. [MEDIUM]
- Universal (arm64 + x86_64) cgo builds need two cgo builds and `lipo`. Cross-arch cgo needs the right `CC`/SDK. Decide arm64-only vs universal explicitly. [MEDIUM]
- Staple the ticket to the `.app`/`.dmg`, not a bare binary. Test the exact download artifact: quarantine attribute set, opened from `~/Downloads` (App Translocation changes the runtime path), on a **clean macOS user account** with no previous TCC state.
- Mac App Store remains out (sandbox forbids synthesized ⌘V). If a store build is ever wanted, Wails has an open issue about isolating private macOS APIs (#6060/#5955). [HIGH]

**How to avoid:**
Do a **throwaway signed + notarized "hello panel" build in P1** (not P6) to flush credentials, entitlements, and Info.plist issues early. Automate the pipeline as a script: build, sign, `codesign --verify --strict`, `notarytool submit --wait`, `stapler staple`, `spctl -a -vv`. Gate the release phase on a clean-account install test.

**Warning signs:**
"App is damaged" or "cannot verify developer" on a downloaded copy. `notarytool` returns Invalid with "not signed with hardened runtime" or "no secure timestamp". Dock icon bounce at launch.

**Phase to address:** P1 (early dry run), P6 (release gate).

---

### Pitfall 16: Leaking API keys through Go error strings, logs and the WebView

**What goes wrong:**
- Klipy puts the key **in the URL path** (`/api/v1/{app_key}/…`), and GIPHY puts `api_key` in the query. Go's `*url.Error` formats as `Get "https://api.klipy.com/api/v1/<KEY>/gifs/search?…": …`. Any `fmt.Errorf("%w")`, toast, log file, or crash report containing it leaks the key. The extension's `fetchProviderJson` interpolates raw error messages into user-visible text, which is the same hazard. [HIGH, Go behavior]
- Both vendors say keys are not secrets (Zulip docs: "KLIPY API keys are not secrets"; GIPHY expects client use), so the real risk is user key theft via shared logs/screenshots/bug reports, not secrecy of the binary. [MEDIUM]
- Keychain access prompts reappear on signature change (Pitfall 6). `go-keyring` on macOS has historically shelled out to `/usr/bin/security` with the secret on the command line, visible in `ps`. Verify the version you pick, or call Security.framework through cgo. [LOW, verify]

**How to avoid:**
Wrap every provider error with a redactor (`strings.ReplaceAll(msg, key, "•••")`, or unwrap `*url.Error` and rebuild with host and path only). Never log full URLs. Keep keys in Keychain. Do not send keys to the frontend except when the user is editing them. Never embed any key in the binary (already a PROJECT constraint). Do not call Raycast's proxy.

**Warning signs:**
A key string visible in a toast, log, or `ps` output.

**Phase to address:** P2 (redaction wrapper), P5 (Keychain).

---

## Moderate Pitfalls

### Pitfall 17: `navigator.clipboard` and paste shortcuts inside the WebView
**What goes wrong:** On macOS 27 with Wails ≤ beta.29, each `evaluateJavaScript:` consumed the click's transient activation, so `navigator.clipboard.writeText()` rejected (#6170, fixed beta.30). Separately, accessory apps have no Edit menu by default, so ⌘V/⌘C/⌘A in the **API-key text field** may do nothing (a past Wails v2 issue, #4918, was exactly this).
**Prevention:** Do all clipboard writes from Go. Test pasting a key into Settings with ⌘V. Ensure an Edit menu with standard selectors exists even though it is invisible (or handle `paste` natively). **Phase:** P3/P5.

### Pitfall 18: Settings window in an accessory app
**What goes wrong:** Windows of an accessory/`LSUIElement` app do not come to the front or take keyboard focus unless the app activates. A Wails tray menu click on macOS 27 was broken until beta.24.
**Prevention:** When opening Settings, temporarily switch to regular activation policy (or activate explicitly), then revert on close. Test with two displays. **Phase:** P5.

### Pitfall 19: Downloads folder and TCC "Files & Folders"
**What goes wrong:** A non-sandboxed Developer ID app writing to `~/Downloads`, `~/Desktop` or `~/Documents` triggers a consent prompt on first write, and a denial yields EPERM. [MEDIUM]
**Prevention:** Default download folder `~/Downloads` and handle EPERM with a "Choose folder…" `NSOpenPanel` fallback (which grants access). Surface the error rather than leaving a spinner (the extension's "toast spinning" bug). **Phase:** P4/P5.

### Pitfall 20: Corporate proxy/VPN divergence between Go and WKWebView
**What goes wrong:** Raycast issues #28546 (GlobalProtect VPN) and several "fetch failed" reports (#24313, #24571, #27216) show this class. Go's `http.ProxyFromEnvironment` reads env vars only, not macOS system proxy/PAC settings, and a Finder-launched `.app` has no shell env. API calls from Go can fail while images load fine in the WebView (CFNetwork honors system proxy). [MEDIUM]
**Prevention:** Keep the extension's friendly error ("Check your internet connection, VPN, proxy, or firewall"). Add a simple diagnostics line in Settings (provider reachable yes/no, status code). Decide in P2 whether API calls go through Go with `SCDynamicStoreCopyProxies` support, or from the WebView where system proxy is honored (GIPHY and Klipy both expect client-side calls). **Phase:** P2.

### Pitfall 21: Finer Gifs Club is an unofficial, flaky provider
**What goes wrong:** An undocumented CloudSearch-style endpoint (`q.parser=simple`, `size`, `start`). It has returned "Internal Server Error" before (raycast/extensions #18707). `fileid` is concatenated into a URL (`new URL(fileid + ".gif", base)`), so path-like IDs could redirect the URL. Trending is empty by design.
**Prevention:** Treat as best-effort: isolate failures per provider, validate `fileid` against `^[A-Za-z0-9._-]+$`, show an empty-state explaining that Finer Gifs has no trending, cap timeouts at 15 s as the extension does. **Phase:** P2.

### Pitfall 22: Settings read at construction time (stale default action)
**What goes wrong:** The 2026-09-18 changelog: "Read the default action preference when building the action panel so it does not retain a previous selection."
**Prevention:** Resolve default action and every setting at invocation time from the single settings store, never from closures captured at mount. Cover with a test that changes the setting between two invocations. **Phase:** P4/P5.

### Pitfall 23: Esc/back navigation order
**What goes wrong:** Raycast issue #28568: Esc from the details view closes the whole launcher instead of returning to the grid.
**Prevention:** Single key-handling stack: Esc pops details → clears query → closes panel. **Phase:** P3.

### Pitfall 24: Hidden-webview first-show latency
**What goes wrong:** Destroying and recreating the webview per invocation costs 1–2 s (Wails #6219 reports first paint regressions). Throttled hidden webviews can show a blank flash on first show after long idle. [LOW/MEDIUM]
**Prevention:** Create the panel once at launch (hidden) and keep it. Reset UI state on show instead of reloading. Measure hotkey-to-first-paint in P3 (target under 150 ms warm). **Phase:** P3.

### Pitfall 25: Clips large renditions and chat upload limits
**What goes wrong:** The extension picks 1080p MP4 first for Clips. Pasting a multi-MB MP4 or HD GIF may exceed target-app upload limits (limits vary by plan and change; verify per app) and fail inside the target app, where Gifsmith cannot report it.
**Prevention:** Default Clips to 720p, warn when the staged file exceeds a configurable threshold (suggest 10 MB), and offer the smaller rendition. **Phase:** P4.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Transliterate the extension's TS models to Go | Fast start | Proxy-shaped parsing; Clips and Klipy silently broken (Pitfall 1) | Never. Use as a behavior reference only |
| Persist favorites as ID lists only | Mirrors extension | Quota burn, empty lists offline, Finer Gifs 10-item cap (Pitfall 14) | Never |
| Persistent favorites byte cache on by default | Instant paste | ToS violation exposure (Pitfall 2) | Opt-in setting only, with disclosure |
| Fixed `sleep(150ms)` before ⌘V | Works on author's machine | Flaky paste across apps/loads | Only as a last-resort floor, behind a focus/modifier check |
| Hardcode keycode 9 for V | Two lines | Broken on Dvorak/Colemak-like layouts (Pitfall 5) | Never in release |
| Ad-hoc sign dev builds | No cert setup | Accessibility/Keychain re-prompt every rebuild (Pitfall 6) | Never for anything touching paste |
| Render grid with original GIFs | Looks great, trivial | Memory/CPU blowup (Pitfall 12) | Never beyond the details view |
| Skip automated mapper tests ("it's just JSON") | Time | Provider response drift goes unnoticed | Never. Keep recorded fixtures |
| Keep `dedupe` and local re-sorting on live results | Cleaner grid | Provider "no reorder/filter" rule (Pitfall 2) | Dedupe same-id-on-page only |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| GIPHY | Using `/v1/clips/*` with a beta key and assuming it works | Probe on key save; hide Clips unless it returns 200 (approval required) |
| GIPHY | Treating `limit` as unbounded or trusting `results.length === limit` for more pages | Clamp to 50, use `pagination` object; offset max 4999; `q` ≤ 50 chars for Clips |
| GIPHY | Never calling analytics pingback; stripping params from render URLs | Fire `analytics.onclick` / Action Register on copy/paste/send; keep URLs intact for rendering |
| GIPHY | No "Powered by GIPHY" mark | Show official mark in panel when provider = GIPHY |
| Klipy | Parsing Tenor-style `media_formats` (Raycast proxy shape) | Real shape: `file.{hd,md,sm,xs}.{gif,webp,mp4,webm,jpg}`, `page/per_page/has_next`, `slugs` for items |
| Klipy | Generic search placeholder; no `customer_id` | Placeholder **"Search KLIPY"** (required); stable random UUID `customer_id`; call Share Trigger on send |
| Klipy | Logging the URL (key in path) | Redact (Pitfall 16) |
| Finer Gifs | Treating its failure as a global failure; `size=10` lookups | Per-provider error isolation; derive URLs from `fileid` |
| NSPasteboard | `writeObjects:[NSImage]` for a GIF | Write the file `NSURL` (and test any extra types for animation) |
| CGEvent | Posting from the default event source with held modifiers | Private event source, explicit ⌘ flag, wait/clear modifiers |
| WKWebView | Fetching GIF bytes in JS and sending to Go | Send ids/URLs; Go downloads natively |
| Raycast proxy | Calling `gif-search.raycast.com` anywhere (even for testing) | Never (explicit PROJECT out-of-scope) |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Request per keystroke | 429 within minutes | 400–500 ms debounce, 2-char minimum, cancel stale, memory cache | Immediately (100/hr cap) |
| All tiles animating, large renditions | Fan noise, 500 MB+ WebContent | Small renditions, MP4/WebM tiles, virtualization, animate focused tile only | About 50+ mounted tiles |
| Decoding full GIF into RGBA for crop | Memory spike, UI stall | Palette-preserving crop; size/frame cap; run off the main thread | GIFs above about 10–20 MB or 200+ frames |
| Favorites view = N by-ID calls on every open | Quota burn, slow, empty offline | Local records, lazy URL refresh | Above about 20 favorites or frequent opens |
| Unbounded staging dir | Disk growth | TTL sweep on launch and timer | After weeks of use |
| Recreating the webview per hotkey | 1–2 s open latency | Create once, hide/show | Always |
| IPC with base64 GIF bytes | UI jank, memory | URLs only across IPC | Any GIF above 1 MB |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| API key in Go error strings, logs, toasts | Key leak via shared screenshots or bug reports | Redactor wrapper; never log URLs (Pitfall 16) |
| Unvalidated provider-supplied strings in file paths (`fileid`, titles, slugs) | Path traversal or odd filenames | Whitelist ids, slugify titles, always join under staging dir and check prefix |
| Writing downloaded bytes without validating type | HTML/JSON saved as `.gif` and pasted into chats | Status, `Content-Type` and magic-byte checks (Pitfall 4) |
| Over-broad entitlements "to make it work" | Weakens notarized app | Start from none; add only with a written reason |
| Using Wails updater with unverified payloads (#6186) | Unsigned update installs | No in-app updater in v1 |
| Loading remote content into the privileged WebView with Wails bindings exposed | Remote script can call Go methods | Serve UI from embedded assets only; remote origins only as `<img>/<video>` sources; set a CSP restricting `img-src`/`media-src` to provider CDNs |
| Synthesizing keystrokes from a long-lived loop | Phantom ⌘V incidents (Claude Code issue #26544 shows how harmful) | Post only in direct response to a user action; never from timers |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Silent failure of Paste when Accessibility is missing | "App is broken" | Fall back to copy plus "press ⌘V" toast, and banner to grant permission |
| Hotkey conflict not reported | "Nothing happens" | Status in tray menu plus tray "Open" fallback |
| Filename leak in Slack ("funny-reaction-gif-to-an-obvious-statement.gif" above the GIF, raycast/extensions #14413) | Kills the joke | Keep "Hide filename" option. Consider defaulting to on |
| Panel opens on the wrong display or Space | Feels broken | Position by cursor screen; test matrix (Pitfall 7) |
| No feedback while downloading large GIFs before copy | Feels frozen | Immediate "Copying…" state and progress, timeout error |
| 429 or offline errors look like "no results" | Confusion | Distinct error states per provider, including key-missing onboarding |
| Aggregate favorites mixing providers | ToS conflict and confusing | Per-provider tabs |
| Default action = Paste before permission is granted | First-run failure | Default Copy (already decided); offer Paste after permission setup |
| Esc closes everything from details view | Annoying | Layered Esc (Pitfall 23) |

## "Looks Done But Isn't" Checklist

- [ ] **Paste GIF:** Often works only in TextEdit. Verify across the Pitfall 3 matrix (Slack, Discord, Messages, Mail, browser composers), both layouts (QWERTY/Dvorak), and plain ⏎ vs ⌘⏎/⇧⏎ triggers.
- [ ] **Accessibility flow:** Verify on a *fresh* macOS user with a notarized build: prompt, deep link, live detection of grant, behavior after rebuilding with the same identity.
- [ ] **Panel over full-screen:** Verify Safari/Zoom full-screen, two displays, Stage Manager, and that the menu bar does not change to "Gifsmith".
- [ ] **Hotkey:** Verify conflict message when another app owns it, recovery via tray, Dvorak/AZERTY, and after screen lock/unlock.
- [ ] **Copy GIF:** Verify the pasted result is animated in each target, the filename is as intended, and the staged file still exists 60 s after paste.
- [ ] **Unfavorite:** Verify no staging/cache file remains, including when unfavoriting during a download.
- [ ] **Clips:** Verify behavior with a non-approved key (provider hidden, no 403 toast spam) and MP4 copy label.
- [ ] **Search:** Verify one request per settled query (watch network log), 429 handling, and paging end condition.
- [ ] **Square crop:** Verify frame count, delays, loop count, transparency and file size against the original on the fixed corpus.
- [ ] **Attribution:** "Powered by GIPHY" visible with GIPHY, "Search KLIPY" placeholder with Klipy.
- [ ] **Notarized build:** `spctl -a -vv` passes on the downloaded `.dmg`/`.app`, not just the local build.
- [ ] **Memory:** WebContent process under budget after 5 pages and after hide/show cycles.
- [ ] **Settings:** ⌘V into the API key field works; Settings window comes to front in an accessory app.
- [ ] **Secrets:** Grep logs/toasts after forcing a network error with a bad host. No key appears.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Wrong provider data model (1, 14) | HIGH | Introduce neutral `Gif` type, write migration for stored favorites, re-record fixtures |
| Paste representation wrong (3) | MEDIUM | Change writer function and default per action. Matrix tells you which to pick |
| TCC stuck after resigning (6) | LOW | Remove row in Settings, re-add; `tccutil reset Accessibility <bundle-id>`; restart if needed |
| Panel behavior wrong (7) | MEDIUM–HIGH | Switch panel config (level/collection), else fall back to activate-and-restore design |
| Wails beta regression (8) | MEDIUM | Revert to previously pinned beta; keep native glue isolated |
| ToS exposure discovered (2) | MEDIUM | Disable byte cache, switch favorites to ID-resolve, add attribution. Mostly config if designed in P2 |
| Key leaked in logs (16) | LOW | Rotate key in provider dashboard, add redactor, purge logs |
| Notarization rejection (15) | LOW–MEDIUM | Read `notarytool log`; fix signing flags/timestamps/nested binaries; resubmit |
| Memory blowup (12) | MEDIUM | Switch tiles to MP4/small renditions, add virtualization |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| 1. Proxy-shaped provider models; Clips approval | P2 | Live-API fixtures pass mapper tests; Clips probe hides provider on 401/403 |
| 2. ToS conflicts and attribution | P2 (decision), P3 (UI), P4 (storage) | Decision recorded in PROJECT.md; per-provider favorites tabs; attribution visible |
| 3. Pasteboard representation | P1 matrix, P4 implement | Committed paste matrix; animated in Slack and Messages at minimum |
| 4. Staging lifecycle and filenames | P4 | Tests for emoji-only/CJK/long titles; magic-byte rejection; file present 60 s post-paste |
| 5. Synthesized ⌘V | P1, P4 | Paste works with QWERTY and Dvorak, with held modifiers, with permission missing (fallback toast) |
| 6. TCC signature identity | P1 (signed dev loop), P5, P6 | Permission survives rebuild with same cert; clean-account test |
| 7. Panel/Spaces/displays | P1, P3 | Test matrix passes manually (needs user) |
| 8. Wails v3 beta pin | P1 and every phase | `go.mod` pinned to exact beta; smoke script green on macOS 26 and 27 |
| 9. cgo/main-thread | P1 | No AppKit call outside `InvokeSync*` (code review grep); no main-thread warnings in console |
| 10. Global hotkey | P1, P5 | Error surfaced; tray fallback; recorder preserves old binding on failure |
| 11. Rate limits and paging | P2 | Network log shows 1 request per settled query; 429 handled; counter shown |
| 12. Grid performance | P3 | Budget met in Activity Monitor; virtualization in place |
| 13. Square crop | P4 | Golden-corpus tests pass; visual inspection |
| 14. Favorites model | P2 schema, P4 | Offline Favorites works; unfavorite leaves no files; Finer Gifs shows >10 |
| 15. Notarization/cgo | P1 dry run, P6 | `spctl` passes on downloaded artifact; clean-account install |
| 16. Key leakage | P2, P5 | Forced-error log grep finds no key |
| 17–25 (moderate) | As labeled | Listed in the checklist |

## Phase Flags for Roadmap

- **P1 needs the most care.** It is a hard feasibility gate for: non-activating panel with typed input over full-screen apps, pasteboard representation per target app, ⌘V synthesis with Accessibility, signed dev loop, a throwaway notarized build, and the pinned Wails beta. A failure here changes the architecture (activation/restore design, or v2 plus custom cgo). Everything marked "verify in spike" belongs here, and native behaviors need manual testing by the user (per PROJECT.md).
- **P2 likely needs a short research spike of its own:** live-key probes of GIPHY (including Clips access), Klipy (path key, slugs, `customer_id`, share trigger) and Finer Gifs, plus the compliance-posture decision and system-proxy approach.
- **P3, P5:** standard patterns, except the performance budget (P3) and the accessory-app Settings/Edit-menu quirks (P5).
- **P4:** standard once P1's matrix exists. The square crop needs a short design note (palette-preserving vs re-encode) before coding.
- **P6:** standard but must be dry-run early (P1) so credentials and entitlements are not discovered at the end.

## Sources

**Primary / HIGH**
- GIPHY Best Practices (prohibited practices, caching, proxy, mixing, attribution): https://developers.giphy.com/docs/api/best-practices/ and https://developers.giphy.com/docs/api
- GIPHY quick start (beta key 100/hour, limit max 50, 429): https://developers.giphy.com/docs/api/quick-start-guide/
- GIPHY Clips (approval required, endpoints): https://developers.giphy.com/docs/clips
- KLIPY docs: Getting Started (100 req/hour testing key), Integration Requirements, Attribution ("Search KLIPY" required), Search/Items API shapes, Format Sizes: https://docs.klipy.com (`.md` variants: `getting-started.md`, `integration-requirements.md`, `attribution.md`, `gifs-api/gifs-search-api.md`, `gifs-api/gifs-items-api.md`, `llms.txt`)
- Wails v3 source (read via GitHub API): `v3/pkg/application/webview_window_options.go` (MacWindow, panel, level, collection behavior), `application_options.go` (MacOptions.ActivationPolicy), `global_shortcut_manager.go` (Register semantics), `mainthread.go` and `application.go` (`InvokeSync*`, `dispatchOnMainThread`), `webview_window_darwin.go` (non-activating panel show path)
- Wails releases (beta.21 to beta.30, all pre-release): https://github.com/wailsapp/wails/releases ; status page: https://docs-preview.wails.io/status/
- Wails issues: #5023, #5359 (panel), #5752/#5919/#6147 (macOS 27 tray), #6170 (transient activation/clipboard on 27), #5760 (macOS 27 AppKit prep), #6214, #6154, #6033, #6056, #6186 (updater), #6219, #6060/#5955 (private APIs), #4918 (Edit menu/clipboard shortcuts, v2)
- Raycast extension (local copy and GitHub issues): CHANGELOG.md, `resolveGifFile.ts`, `cachedGifs.ts`, `downloadFile.ts`, `localGifs.ts`, `useLocalGifs.ts`, `resolveSquareGifFile.ts`, `models/*`, `GifActions.tsx`. Issues raycast/extensions #14413 (Slack shows filename), #20631, #28546, #28568, #30630, #31004, #31144, #18707, #15808, #24313/#24571/#25591/#27216
- Wails v3 signing guide: https://v3.wails.io/guides/build/signing ; global shortcuts docs: https://v3.wails.io/features/keyboard/global-shortcuts/ (page itself returned 403 to fetch; semantics confirmed from source)

**Secondary / MEDIUM**
- RegisterEventHotKey Option-only regression and 15.2 fix: https://developer.apple.com/forums/thread/763878 and https://blog.eternalstorms.at/2024/09/23/keyboard-shortcuts-using-option-and-or-shift-modifiers-only-no-longer-allowed-on-macos-sequoia/
- Keycode layout dependence (V = 9 on QWERTY, 47 on Dvorak): https://github.com/Clipy/Sauce
- Ad-hoc signing and TCC/Accessibility grants resetting on rebuild: OpenClaw macOS permissions/signing docs (https://docs.openclaw.ai/platforms/mac/permissions.md); no Apple primary source found
- Full-screen overlay recipe (accessory plus non-activating NSPanel, high level): https://origin-devforums.apple.com/forums/thread/826308 ; collection-behavior conflict report: https://developer.apple.com/forums/thread/26677
- Zulip GIF-picker integration notes (keys are not secrets; test vs production keys): https://zulip.readthedocs.io/en/12.1/_sources/production/gif-picker-integrations.md.txt
- WKWebView memory accounting (separate process): https://developer.apple.com/forums/thread/21956

**Low confidence / needs spike verification**
- Per-app animated-paste behavior for Discord, Messages, Mail, browsers (no authoritative source found). Secondary hints only: https://www.xenforo.com/community/threads/odd-behaviour-cutting-and-pasting-gif-animations.136749/latest
- Secure Input effect on synthesized events; Go `image/gif` palette-preserving crop approach; hidden-webview throttling; `go-keyring` argv exposure; system-proxy divergence between Go and WKWebView.

**Not used:** the `research-plan` / `research-store` cache seam was not run (no `gsd_run` shim was needed for this one-off pass). Findings come from direct source reads, GitHub API, vendor docs and web search.

---
*Pitfalls research for: Go + Wails macOS menu-bar GIF search launcher (Gifsmith)*
*Researched: 2026-10-17*
