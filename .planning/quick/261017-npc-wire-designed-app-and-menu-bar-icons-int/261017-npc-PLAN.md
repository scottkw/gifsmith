---
phase: quick-261017-npc
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - icons/README.md
  - icons/app/appicon.png
  - icons/app/Gifsmith-AppIcon.svg
  - icons/app/Gifsmith.iconset
  - icons/menubar/GifsmithTemplate.png
  - icons/menubar/GifsmithTemplate-2x.png
  - icons/menubar/GifsmithTemplate-3x.png
  - icons/menubar/GifsmithTemplate.svg
  - icons/menubar/GifsmithTemplate-tray-2x.png
  - main.go
  - internal/shell/shell.go
  - build/appicon.png
  - build/appicon.icon/icon.json
  - build/appicon.icon/Assets/gifsmith_glyph.svg
  - build/darwin/Assets.car
  - build/darwin/icons.icns
  - build/windows/icon.ico
autonomous: false
requirements: [quick-261017-npc]

estimate:
  tokens: 45000
  raw_tokens: 45000
  tasks: 3
  confidence: low

must_haves:
  truths:
    - "The menu bar shows the Gifsmith template glyph (two offset frames with a play triangle), not the Wails stock icon, and it adapts to light and dark menu bars"
    - "The installed ~/Applications/Gifsmith.app shows the designed Gifsmith app icon in Finder, not the Wails logo"
    - "The built bundle's Assets.car and icons.icns are compiled from the Gifsmith artwork; no Wails logo asset remains in Assets.car"
    - "The designer's icons/ source (README, SVGs, PNGs, iconset) is committed, with no .DS_Store files"
    - "The two deferred Phase 2 Clips native checks (quit-on-Clips relaunch, offline relaunch then Try Again) are confirmed by the user"
  artifacts:
    - path: "icons/menubar/GifsmithTemplate-tray-2x.png"
      provides: "44x44 RGBA template image: the 36x36 @2x glyph centred on a transparent 22pt@2x canvas"
    - path: "main.go"
      provides: "go:embed of the tray PNG, passed into shell.New"
    - path: "internal/shell/shell.go"
      provides: "shell.New(app, plat, trayIcon []byte) sets the tray template icon from the passed bytes"
    - path: "build/appicon.icon/Assets/gifsmith_glyph.svg"
      provides: "Gifsmith foreground artwork for Icon Composer (tile removed, viewBox cropped to the tile)"
    - path: "build/appicon.icon/icon.json"
      provides: "Icon Composer document: solid #F6F4EF fill, one flat layer"
    - path: "build/darwin/Assets.car"
      provides: "Compiled appicon asset catalog (CFBundleIconName=appicon) used on macOS 26+"
    - path: "build/darwin/icons.icns"
      provides: "CFBundleIconFile icon used on macOS 13-25"
  key_links:
    - from: "main.go"
      to: "internal/shell/shell.go"
      via: "shell.New(app, plat, trayIcon)"
      pattern: "shell\\.New\\(app, plat, trayIcon\\)"
    - from: "internal/shell/shell.go"
      to: "Wails SystemTray"
      via: "tray.SetTemplateIcon(trayIcon)"
      pattern: "SetTemplateIcon\\(trayIcon\\)"
    - from: "build/appicon.icon"
      to: "build/darwin/Assets.car + icons.icns"
      via: "wails3 generate icons -iconcomposerinput appicon.icon -macassetdir darwin (actool)"
      pattern: "iconcomposerinput appicon.icon"
    - from: "build/darwin/Info.plist"
      to: "Assets.car"
      via: "CFBundleIconName=appicon"
      pattern: "CFBundleIconName"
---

<objective>
Replace the Wails stock menu bar icon and the Wails-logo app icon with the user's designed Gifsmith icons from `icons/`, reinstall the Developer ID-signed dev build, and have the user eyeball both icons (plus two deferred Phase 2 Clips native checks).

Purpose: the app currently ships Wails' placeholder branding in the menu bar and in Finder.
Output: embedded tray template PNG, rebuilt Icon Composer source plus regenerated `Assets.car` / `icons.icns` / `icon.ico`, committed `icons/` source, reinstalled `~/Applications/Gifsmith.app`.
</objective>

<execution_context>
@~/.claude/gsd-core/workflows/execute-plan.md
@~/.claude/gsd-core/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@.claude/CLAUDE.md
@icons/README.md
@internal/shell/shell.go
@main.go
@build/Taskfile.yml
@build/darwin/Taskfile.yml
@build/appicon.icon/icon.json
@icons/app/Gifsmith-AppIcon.svg
@icons/menubar/GifsmithTemplate.svg

Facts established while planning (do not re-derive):
- Branch: `gsd/phase-02-providers-search-ui-first-run`. Stay on it. `workflow.use_worktrees=false`.
- `go:embed` cannot reach parent directories, but `main.go` sits at the repo root, so the main package can embed `icons/...` directly. `shell.New` has exactly one caller (`main.go:94`). `main.go` is untagged, so both the desktop and `-tags server` builds compile it.
- Wails darwin tray (`systemtray_darwin.m`, `systemTraySetIcon`) forces the NSImage size to `NSStatusBar thickness` (22pt square) regardless of the PNG. The designed glyph is 18pt (`GifsmithTemplate-2x.png` is 36x36 @2x with content edge to edge). Passed unpadded, it would be stretched to 22pt (1.22x, soft, oversized). Padding it to 44x44 (22pt@2x) keeps 1:1 pixels and the designed 18pt glyph size.
- `build/Taskfile.yml` `common:generate:icons` runs, from `build/`: `wails3 generate icons -input appicon.png -macfilename darwin/icons.icns -windowsfilename windows/icon.ico -iconcomposerinput appicon.icon -macassetdir darwin`. When the `.icon` compiles, Wails runs `/usr/bin/actool` (Xcode 27.1 present, needs 26+) and gets BOTH `darwin/Assets.car` and `darwin/icons.icns` from the `.icon`. `appicon.png` then only feeds `windows/icon.ico`; it feeds `icons.icns` only if the `.icon` path is removed. The darwin `build` task depends on this task, so `package` and `install:dev` re-run it whenever its sources change.
- Both `build/darwin/Info.plist` and `Info.dev.plist` set `CFBundleIconFile=icons` and `CFBundleIconName=appicon`. `create:app:bundle` and `run` copy `Assets.car` only if it exists.
- `icons/app/Gifsmith-AppIcon.svg` is a Big Sur-style icon: an 824px rounded tile (x/y 100, rx 185, fill #f6f4ef) on a 1024 canvas, plus the glyph group `translate(189 206) scale(34)`. Icon Composer's canvas IS the tile (the system applies the squircle mask), so the design's tile must be removed and the viewBox cropped to it. Otherwise you get a tile inside a tile.
- `APP_NAME` is `gifsmith`: the built bundle is `bin/gifsmith.app`, the process name is `gifsmith`, and the install path is `~/Applications/Gifsmith.app`.
- The current `Assets.car` (`assetutil --info build/darwin/Assets.car`) lists `appicon_Assets/...` entries named after the old Wails logo layer. After regeneration they must name `gifsmith_glyph`.
</context>

<tasks>

<task type="tracer">
  <name>Task 1: Tracer: designed menu bar glyph embedded and set on the tray, plus icons/ source committed</name>
  <files>icons/menubar/GifsmithTemplate-tray-2x.png, icons/README.md, main.go, internal/shell/shell.go, icons/ (designer source: README.md, app/appicon.png, app/Gifsmith-AppIcon.svg, app/Gifsmith.iconset/*, menubar/*.png, menubar/GifsmithTemplate.svg)</files>
  <read_first>internal/shell/shell.go (lines 1-80), main.go (lines 1-30 and 85-100), icons/README.md</read_first>
  <action>
1. Create `icons/menubar/GifsmithTemplate-tray-2x.png`: a 44x44 RGBA PNG, fully transparent, with `icons/menubar/GifsmithTemplate-2x.png` (36x36) copied pixel for pixel at offset (4,4), using draw.Src and no resampling. Generate it with a throwaway Go program (image/png + image/draw) in your session scratchpad, outside the repo, and do not commit the program. Reason: Wails forces the tray image to a 22pt square (see context), so the 18pt glyph needs 2pt of transparent margin per side to render at its designed size with 1:1 pixels.
2. In `icons/README.md`, under the Menu bar section, add one line: `menubar/GifsmithTemplate-tray-2x.png` is the 2x glyph centred on a 44x44 (22pt@2x) transparent canvas, because Wails sizes tray images to the 22pt menu bar thickness. This is the file the app embeds. Leave the rest of the README alone.
3. In `main.go`, next to the existing `//go:embed all:frontend/dist` var, add `//go:embed icons/menubar/GifsmithTemplate-tray-2x.png` on a new `var trayIcon []byte` (`embed` is already imported). Change the call at line 94 to `shell.New(app, plat, trayIcon)`. Embedding from the main package was chosen over copying the PNG into `internal/shell`: one source file, no duplicate asset, and a one-parameter signature change with a single caller.
4. In `internal/shell/shell.go`, change the signature to `func New(app *application.App, plat platform.Platform, trayIcon []byte) *Shell`. Replace the stock Wails template icon argument on line 69 with `tray.SetTemplateIcon(trayIcon)`, and drop the now-unused Wails icons package import. Update the `New` doc comment to mention the tray icon bytes come from the caller. Keep everything else unchanged.
5. Commit the designer source together with the code. Use `git add` with the explicit paths `icons/README.md icons/app icons/menubar main.go internal/shell/shell.go`. Never `git add -A` or `git add .`, because the repo root has untracked `.DS_Store`, `.ignore`, `.planning/milestone.lock`, `.planning/state.json` and REVIEW iteration files that must stay untracked. Check that no `.DS_Store` got staged. Message: `feat(quick-261017-npc): use designed Gifsmith menu bar icon`, ending with the Co-Authored-By trailer.

Safety: do not run `go mod tidy`. If any go command adds `stretchr/objx` lines to go.sum, restore go.sum with `git checkout go.sum`. `go build ./...` fails in `build/ios` because of a pre-existing template issue, so build `.` explicitly as shown in verify.
  </action>
  <verify>
    <automated>file icons/menubar/GifsmithTemplate-tray-2x.png | grep -q '44 x 44, 8-bit/color RGBA' && go vet . ./internal/shell/ && go test ./internal/shell/ && go build -o /dev/null . && go build -tags server -o /dev/null . && test "$(grep -c 'SystrayMacTemplate' internal/shell/shell.go)" = 0 && grep -q 'SetTemplateIcon(trayIcon)' internal/shell/shell.go && grep -q 'shell.New(app, plat, trayIcon)' main.go && test -z "$(git ls-files -- ':(glob)icons/**/.DS_Store')" && test -z "$(git status --porcelain go.sum)"</automated>
  </verify>
  <done>Tray PNG is 44x44 RGBA with the glyph centred. Both desktop and server builds compile. shell.go sets the tray template icon from the bytes passed in and no longer imports the Wails stock icons package. icons/ source plus the code change are committed in one commit with no .DS_Store.</done>
</task>

<task type="auto">
  <name>Task 2: Rebuild the Icon Composer app icon from the Gifsmith design, regenerate Assets.car/icons.icns/icon.ico, reinstall the signed dev build</name>
  <files>build/appicon.png, build/appicon.icon/icon.json, build/appicon.icon/Assets/gifsmith_glyph.svg, build/appicon.icon/Assets/ (old Wails logo SVG deleted), build/darwin/Assets.car, build/darwin/icons.icns, build/windows/icon.ico</files>
  <read_first>build/appicon.icon/icon.json, icons/app/Gifsmith-AppIcon.svg (ignore the base64 metadata block), icons/menubar/GifsmithTemplate.svg (mask technique), build/Taskfile.yml generate:icons, build/darwin/Taskfile.yml install:dev</read_first>
  <action>
1. Copy `icons/app/appicon.png` to `build/appicon.png`. It becomes the Windows `.ico` source and the `icons.icns` source if the fallback in step 6 is needed.
2. Delete the old Wails logo SVG from `build/appicon.icon/Assets/` with `git rm`. Create `build/appicon.icon/Assets/gifsmith_glyph.svg` from `icons/app/Gifsmith-AppIcon.svg`:
   - Drop the c2pa `<metadata>` block and the tile `<rect>` (x 100, y 100, 824x824, rx 185, fill #f6f4ef).
   - Set `width="1024" height="1024" viewBox="100 100 824 824"` so the design's tile area maps onto the full Icon Composer canvas.
   - Keep the `<g transform="translate(189 206) scale(34)">` group with the back frame rect (stroke #22201c), the orange card rect (#e0612f) and the light play-triangle path (#f6f4ef) exactly as designed.
   - Remove the tile-coloured knockout rect (x -0.5, y 3.5, 15x15, rx 4, fill #f6f4ef). Instead, put a `<mask>` on the back-frame rect: a white rect covering the group area plus a black rounded rect at x -0.5, y 3.5, 15x15, rx 4. This is the same technique `icons/menubar/GifsmithTemplate.svg` uses. Reason: a knockout painted in the tile colour only works on a tile-coloured background, and the system dark/tinted icon styles replace the background.
3. Rewrite `build/appicon.icon/icon.json`, keeping its existing structure and key style:
   - Root `fill` becomes `{"solid": "srgb:0.96471,0.95686,0.93725,1.00000"}` (#F6F4EF).
   - One group with one layer: `image-name` `gifsmith_glyph.svg`, `name` `gifsmith_glyph`, position scale 1.0, translation [0,0].
   - No layer `fill-specializations`. They flatten the whole layer to one colour, which suited the monochrome Wails glyph and would erase this design's three colours.
   - Group effects set flat to match the design: `specular` false, `translucency` `{"enabled": false, "value": 0.5}`, `shadow` `{"kind": "none", "opacity": 0.5}`. If actool rejects `"none"`, delete the `shadow` key instead.
   - Keep `supported-platforms` unchanged.
4. Regenerate from the repo root, running the exact Taskfile command so it happens regardless of task checksums: `(cd build && wails3 generate icons -input appicon.png -macfilename darwin/icons.icns -windowsfilename windows/icon.ico -iconcomposerinput appicon.icon -macassetdir darwin)`. It must exit 0. Afterwards, make sure no stray file such as `build/darwin/temp.plist` is left in `git status`.
5. Visual check, done by you, not the user: `iconutil -c iconset build/darwin/icons.icns -o <scratchpad>/check.iconset`, then use the Read tool to view `icon_512x512.png` and `icon_32x32.png`. Expected: the designed glyph (dark back frame, orange front card, light play triangle) filling a light squircle at roughly the design's proportions, with no tile inside a tile, no Wails logo, and no back-frame line drawn across the orange card. If the mask did not render (the back frame is missing or crosses the card), switch the mask to a `clipPath` with `clip-rule="evenodd"` (outer rect plus the rounded card rect as one path), regenerate, and re-check. If `xcrun ictool --help` offers an appearance/rendition export, optionally render the Dark appearance too and note the result in the SUMMARY. Dark icon style is an opt-in system setting and is not a gate.
6. Fallback, only if actool fails or the render is still wrong after one fix attempt: drop the Icon Composer path. In `build/Taskfile.yml` `generate:icons`, remove `-iconcomposerinput appicon.icon -macassetdir darwin` from the cmd and `"appicon.icon"` from `sources`. Run `git rm -r build/appicon.icon build/darwin/Assets.car`. Remove the `CFBundleIconName` key and its string from both `build/darwin/Info.plist` and `build/darwin/Info.dev.plist`. Rerun the generate command without the two removed flags, so `icons.icns` comes from `build/appicon.png`. Record in the SUMMARY which path shipped and why.
7. Install: `wails3 task darwin:install:dev`. It builds, Developer ID-signs with hardened runtime, runs `pkill -x gifsmith` itself, replaces `~/Applications/Gifsmith.app`, and opens it. Do not kill anything else. Never `pkill -f` vite, node or pnpm. Then refresh the LaunchServices icon cache for that bundle only: `/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister -f ~/Applications/Gifsmith.app`. No synthetic desktop input of any kind: no clicks, keystrokes, hotkey simulation or AppleScript UI scripting. This is the user's primary Mac.
8. Commit with explicit paths: `build/appicon.png build/appicon.icon build/darwin/Assets.car build/darwin/icons.icns build/windows/icon.ico`, plus `build/Taskfile.yml` and both plists if the fallback was taken. Message: `feat(quick-261017-npc): use designed Gifsmith app icon`, with the Co-Authored-By trailer. Commit after the install so the committed `Assets.car`/`icons.icns` are byte-identical to what the install's `package` step built.

Item (4) from the brief, DMG icons: leave `build/darwin/dmg-file-icon.*` untouched for the Phase 4 release work. The DMG volume icon already defaults to `build/darwin/icons.icns` (`DMG_VOLUME_ICON`), so it picks up the new icon automatically.
  </action>
  <verify>
    <automated>cmp -s icons/app/appicon.png build/appicon.png && assetutil --info build/darwin/Assets.car | grep -q 'gifsmith_glyph' && test "$(assetutil --info build/darwin/Assets.car | grep -c 'icon_vector')" = 0 && cmp -s build/darwin/Assets.car ~/Applications/Gifsmith.app/Contents/Resources/Assets.car && cmp -s build/darwin/icons.icns ~/Applications/Gifsmith.app/Contents/Resources/icons.icns && test "$(plutil -extract CFBundleIconName raw ~/Applications/Gifsmith.app/Contents/Info.plist)" = appicon && codesign --verify --deep --strict ~/Applications/Gifsmith.app && pgrep -x gifsmith >/dev/null && test -z "$(git status --porcelain build/)"</automated>
  </verify>
  <done>Primary path: the Assets.car asset names reference gifsmith_glyph and nothing Wails. The installed bundle's Assets.car and icons.icns match the committed build files, CFBundleIconName=appicon, and the signature verifies. The executor has looked at the extracted icns PNGs and they show the Gifsmith design in a single squircle. The app is running from ~/Applications/Gifsmith.app. Fallback path, if taken and recorded in the SUMMARY: `test ! -e ~/Applications/Gifsmith.app/Contents/Resources/Assets.car`, `plutil -extract CFBundleIconName raw` on the installed Info.plist fails, and icons.icns matches the build file. The rest of the checks are the same.</done>
</task>

<task type="checkpoint:human-verify" gate="blocking">
  <name>Task 3: User eyeballs the menu bar and app icons and runs the two deferred Clips native checks</name>
  <what-built>The designed Gifsmith menu bar template glyph (replacing Wails' stock icon) and the designed app icon (replacing the Wails logo), installed as the Developer ID-signed dev build at ~/Applications/Gifsmith.app, which is now running. The Phase 2 code-review fixes WR-02/WR-03 (Clips probe timing) are also in this build.</what-built>
  <how-to-verify>
    All steps are done by the user. The executor must not send any input to the desktop.
    1. Menu bar icon: find the Gifsmith icon in the menu bar. It should be two offset rounded frames with a play triangle, crisp, about the size of neighbouring icons, not the Wails logo. Switch System Settings, Appearance between Light and Dark (or use a wallpaper that gives a dark menu bar) and confirm the glyph inverts correctly in both. Click it and check the menu still shows Open Gifsmith / Settings… / Quit Gifsmith.
    2. App icon: open ~/Applications in Finder (icon view). Gifsmith should show the light rounded tile with the dark back frame, orange card and light play triangle. It should not be the Wails logo, a tile inside a tile, or a grey "jailed" square. If Finder still shows the old icon, give it a few seconds or reopen the window, since the icon cache was refreshed via lsregister.
    3. Deferred check (a): open the panel with ⌃⌥⌘G, switch to the Clips tab, then Quit Gifsmith from the menu bar. Relaunch ~/Applications/Gifsmith.app and press ⌃⌥⌘G. Expected: a brief loading state, then the Clips message (not a stale or empty state).
    4. Deferred check (b): with the Clips tab last selected, quit Gifsmith and turn Wi-Fi off. Relaunch, open the panel (offline state shown), turn Wi-Fi back on, then press Try Again. Expected: Clips loads, or shows the "not enabled" message for a non-approved key, rather than a stuck error.
  </how-to-verify>
  <resume-signal>Type "approved", or describe what is wrong for each item (1-4). For an app-icon problem (tile in tile, grey jail, wrong art), say "icon-fallback" and the executor will apply Task 2 step 6 (icns-only path), reinstall, and ask again.</resume-signal>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| repo assets → signed bundle | Image files in icons/ and build/ are embedded or copied into the Developer ID-signed app |
| install:dev → ~/Applications | The task deletes and replaces an installed bundle and kills a process by exact name |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-npc-01 | Tampering | ~/Applications/Gifsmith.app signature | medium | mitigate | Task 2 verify runs `codesign --verify --deep --strict` on the installed bundle after install:dev signs it, and cmp checks that the installed icon files match the committed build files |
| T-npc-02 | Denial of Service | Process kill during reinstall | medium | mitigate | Only install:dev's own `pkill -x gifsmith` (exact name) is used. pkill -f on vite/node/pnpm and any broad kill are forbidden in the action text |
| T-npc-03 | Information Disclosure | Committing untracked files | low | mitigate | Explicit-path git add only, and verify asserts no .DS_Store is tracked under icons/. Root .DS_Store, .ignore, milestone.lock, state.json and REVIEW iteration files stay untracked |
| T-npc-04 | Tampering | build/Taskfile.yml / plists (fallback only) | low | accept | Edits are limited to removing the Icon Composer flags and the CFBundleIconName key, and are reviewed in the commit diff |
| T-npc-SC | Tampering | npm/pip/cargo installs | high | accept | No package installs in this plan. The only tools used are wails3, actool, iconutil, assetutil, codesign and the stdlib Go image packages, all already present |
</threat_model>

<verification>
- `go vet . ./internal/shell/ && go test ./internal/shell/ && go build -o /dev/null . && go build -tags server -o /dev/null .`
- `assetutil --info ~/Applications/Gifsmith.app/Contents/Resources/Assets.car | grep -q gifsmith_glyph` (primary path)
- `codesign --verify --deep --strict ~/Applications/Gifsmith.app`
- `git status --porcelain` shows no tracked-file changes, and only the pre-existing untracked items remain (.DS_Store, .ignore, .planning/milestone.lock, .planning/state.json, REVIEW iteration files)
- User approval at Task 3
</verification>

<success_criteria>
- The menu bar shows the designed Gifsmith template glyph at its designed 18pt size, in both light and dark menu bars (user-confirmed)
- Finder shows the designed Gifsmith app icon for ~/Applications/Gifsmith.app (user-confirmed)
- No Wails logo asset remains in build/darwin/Assets.car (or Assets.car is removed under the recorded fallback)
- icons/ source committed with no .DS_Store
- Deferred Clips checks (a) and (b) confirmed by the user
- Still on branch gsd/phase-02-providers-search-ui-first-run, and go.sum unchanged
</success_criteria>

<output>
Create `.planning/quick/261017-npc-wire-designed-app-and-menu-bar-icons-int/261017-npc-SUMMARY.md` when done. Record which app-icon path shipped (Icon Composer or icns fallback), the result of the visual icns check, any optional Dark render, and the user's results for checks 1-4.
</output>
