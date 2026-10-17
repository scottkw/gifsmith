---
quick_id: 261017-npc
slug: wire-designed-app-and-menu-bar-icons-int
status: complete
date: 2026-10-17
commits: [6197095, 95df3e4, dbe648e]
---

# Quick 261017-npc: Wire designed app and menu bar icons

The user's Claude Design icons (`icons/`) now ship in the app: a template menu bar glyph and a layered Icon Composer app icon that reads in Default, Dark, Tinted and Clear styles.

## Tasks

| Task | Commit | Result |
|------|--------|--------|
| 1. Menu bar icon | 6197095 | `main.go` embeds `icons/menubar/GifsmithTemplate-tray-2x.png` and passes it to `shell.New`; `shell.go` calls `tray.SetTemplateIcon(trayIcon)`. Wails' stock template icon removed. `icons/` source committed (no .DS_Store). |
| 2. App icon | 95df3e4 | `build/appicon.png` replaced; `build/appicon.icon` rebuilt with the Gifsmith artwork (design's own tile removed — Icon Composer supplies the squircle); `Assets.car`, `icons.icns`, `icon.ico` regenerated; signed dev build reinstalled. |
| 3. Human verify | dbe648e (fix) | Menu bar (light/dark), Clips relaunch restore, offline→Try Again: approved. App icon was hard to read in the Dark icon style → split into `back_frame` / `front_card` layers; back frame swaps to `back_frame_light.svg` (#E8E4DC) via `image-name-specializations` for dark and tinted. User approved ("Much better"). |

## Deviations

- Tray image padded to 44×44 (`GifsmithTemplate-tray-2x.png`): Wails scales the tray image to the 22pt bar height, so the raw 36×36 @2x would have been upscaled ~22%.
- Back-frame knockout uses an L-shaped `clipPath`; the planned `<mask>` rendered soft ends and an evenodd clip lost the gap.
- Per-appearance `fill-specializations` rendered correctly in ictool but actool flattened the stroke into a solid block in `icons.icns`; swapping images per appearance avoids it.
- `icons.icns` from actool holds 16/32/128/256 px only (larger sizes come from Assets.car) — same as the previous Wails-generated file.
- `go.sum` gains `stretchr/objx` lines from install:dev's `go mod tidy`; restored each time, not committed.
- DMG file icons left for Phase 4 (volume icon already uses `build/darwin/icons.icns`).

## Verification

- `go vet . ./internal/...`, `go test -race -count=1 ./internal/...` (11 packages), `go build .` and `-tags server` pass.
- `assetutil` shows back_frame, back_frame_light, front_card; installed Assets.car/icons.icns byte-identical to committed; `codesign --verify --deep --strict` passes (Developer ID, hardened runtime).
- ictool renders checked by eye for Default, Dark, TintedLight/Dark, ClearLight/Dark.
- User verified in Finder (`~/Applications`, Dark icon style) and the menu bar.

## Self-Check: PASSED
