# Gifsmith

A macOS menu bar app for finding and sharing animated GIFs: press a global hotkey (⌃⌥⌘G), search, and copy or paste the GIF into the app you were using. A port of the Raycast "GIF Search" extension for people who don't use Raycast.

Built with Go + [Wails v3](https://v3.wails.io) (pinned `v3.0.0-beta.30`) and Svelte 5. macOS 13+, Apple silicon only.

## Build

Requires Go 1.25+, pnpm, Xcode command line tools and the matching Wails CLI:

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.30
wails3 task darwin:package   # → bin/gifsmith.app (ad-hoc signed)
```

Architecture and the Phase 1 walking skeleton: [.planning/phases/01-native-spike-signed-dev-loop/SKELETON.md](.planning/phases/01-native-spike-signed-dev-loop/SKELETON.md).
