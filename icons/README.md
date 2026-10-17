# Gifsmith icons

## Menu bar (template image — pure black + alpha)
- menubar/GifsmithTemplate-2x.png — embed and pass to `systemTray.SetTemplateIcon(...)` (Wails v3)
- 1x / 3x PNG and SVG source included
- menubar/GifsmithTemplate-tray-2x.png — the 2x glyph centred on a 44x44 (22pt@2x) transparent canvas, because Wails sizes tray images to the 22pt menu bar thickness. This is the file the app embeds.

```go
//go:embed icons/GifsmithTemplate-2x.png
var trayIcon []byte
systemTray.SetTemplateIcon(trayIcon)
```

## App icon
- app/appicon.png (1024) — drop into `build/appicon.png`; Wails generates the .icns on build
- app/Gifsmith.iconset — before running `iconutil`, rename `-2x` files to `@2x` (e.g. `cd Gifsmith.iconset && for f in *-2x.png; do mv "$f" "${f/-2x/@2x}"; done`), then `iconutil -c icns Gifsmith.iconset`
- app/Gifsmith-AppIcon.svg — vector source

Colors: tile #F6F4EF, back frame #22201C, front frame #E0612F
