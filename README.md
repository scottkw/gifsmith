# Gifsmith

A macOS menu bar app for finding and sharing animated GIFs: press a global hotkey (⌃⌥⌘G), search, and copy or paste the GIF into the app you were using. A port of the Raycast "GIF Search" extension for people who don't use Raycast.

Requires macOS 13 or later on Apple silicon. Built with Go + [Wails v3](https://v3.wails.io) (pinned `v3.0.0-beta.30`) and Svelte 5.

## Install

1. Download `Gifsmith-1.0.0-arm64.dmg` and `Gifsmith-1.0.0-arm64.dmg.sha256` from [GitHub Releases](https://github.com/scottkw/gifsmith/releases).
2. Verify the download (see below).
3. Open the DMG and drag **Gifsmith** onto **Applications**.
4. Open Gifsmith from Applications. It is signed with a Developer ID and notarized by Apple, so it opens without a warning.
5. Gifsmith lives in the menu bar (it has no Dock icon). Press **⌃⌥⌘G** anywhere to open the search panel, or click the menu bar icon and choose **Open Gifsmith**.

### Verify the download

With both files in the same folder (for example Downloads):

```bash
cd ~/Downloads
shasum -a 256 -c Gifsmith-1.0.0-arm64.dmg.sha256   # prints "Gifsmith-1.0.0-arm64.dmg: OK"
```

After installing, `spctl -a -vv /Applications/Gifsmith.app` should report `source=Notarized Developer ID` and `origin=Developer ID Application: Ken Scott (S2K7P43927)`.

## Get API keys

Gifsmith never ships with keys. Each user adds their own free keys in **Settings → Keys** (they are stored in the macOS Keychain, never in a file):

- **GIPHY**: create an app at the [GIPHY developer dashboard](https://developers.giphy.com/dashboard/) and copy its API key.
- **Klipy**: sign up at the [Klipy Partner Panel](https://partner.klipy.com) and copy your app key.

New GIPHY and Klipy keys are test keys limited to **100 requests per hour**. Gifsmith caches trending results and waits for you to stop typing before it searches, so normal use stays well inside that. If you do hit the limit, the panel says so and searching works again within the hour.

The Finer Gifs Club needs no key.

## Use

- Press ⌃⌥⌘G, type, and press **↵** to copy the selected GIF; **⌘↵** pastes it straight into the app you were using.
- Arrow keys move through the grid, **⌘K** lists every action (copy, paste, link, Markdown, square crop, download, favorite), **⌘1–⌘4** switch providers and **⌘5 / ⌘6** show Favorites and Recents.
- **Esc** hides the panel.

## Settings

Open **Settings…** from the menu bar icon (or press ⌘, in the panel):

- **General**: theme, open at login, automatic paste (Accessibility), the ↵ and ⌘↵ actions, and updates.
- **Keys**: add, test, change or remove the GIPHY and Klipy keys.
- **Hotkeys**: change the Open Gifsmith shortcut and add shortcuts for Favorites and Recents.
- **Search**: grid size, GIFs per page, GIPHY content rating and language, Klipy country.
- **Downloads**: the download folder and whether downloaded files keep their original names.
- **About**: version, links and credits.

## Updating

Gifsmith checks [GitHub Releases](https://github.com/scottkw/gifsmith/releases) for a newer version about 10 seconds after launch and then once a day. Each check is one unauthenticated GET to `api.github.com` (`/repos/scottkw/gifsmith/releases/latest`). It sends no identifiers: no account, no device ID, no usage data. To turn it off, clear **Check for updates automatically** in **Settings → General → Updates**. **Check for Updates…** there still checks on demand.

When a newer version exists, the menu bar shows **Update to Gifsmith {version}…** and Settings offers **Download and Open…**. That downloads the DMG from GitHub, checks its SHA-256 checksum, its Developer ID signature (Team ID S2K7P43927) and Apple notarization, and only then opens it, so you drag the new Gifsmith to Applications as usual. Gifsmith never replaces itself; if any check fails, the file is deleted and nothing is opened.

## Attribution

- <img src="frontend/public/attribution/giphy-powered-light.png" alt="Powered by GIPHY" height="16"> **Powered By GIPHY.** GIPHY and GIPHY Clips search use the [GIPHY API](https://developers.giphy.com/), and the panel shows the official mark whenever GIPHY results are on screen.
- **KLIPY.** The Klipy search box reads **Search KLIPY**, and Klipy results stay in their own tab, in the order Klipy returns them.
- **The Finer Gifs Club.** GIFs from [The Finer Gifs Club](https://thefinergifs.club) are credited in the panel footer.
- **The original extension.** Gifsmith is a port of the [GIF Search extension for Raycast](https://github.com/raycast/extensions/tree/main/extensions/gif-search) by Joseph Schmitt and contributors: tm.wrnr, j3lte, samuelkraft, ridemountainpig, spookywy, seanparkross, YourMCGeek, thomas, tim_gailey, clins1994, 0xdhrv, Olli0103 and alexi.build. Used under the MIT License, © 2021 Raycast.

## Known limits

- **Klipy file actions are disabled** until Klipy gives written clearance for the temporary local copy that copying a GIF needs. On Klipy results, ↵ copies the GIF's link and ⌘↵ pastes the link; Copy GIF, Paste GIF, the square variants and Download are hidden.
- **GIPHY Clips needs an approved GIPHY key.** GIPHY only grants Clips access on request (clips@giphy.com); with an ordinary key the Clips tab explains how to ask.
- **Pasting needs Accessibility access** (Settings → General → Automatic paste). Copying always works without it; without the permission, paste actions copy and tell you to press ⌘V.
- **⌥ + letter hotkeys on non-QWERTY layouts** may bind a different key than the one you pressed, because macOS resolves letters through the active keyboard layout. Use **Test** in Settings → Hotkeys to check a shortcut.
- The Finer Gifs Club is a hobby service and is sometimes unavailable; the other providers keep working.
- Apple silicon only; no Intel build.

## License

MIT. See [LICENSE](LICENSE), which also carries the MIT notice of the Raycast GIF Search extension that Gifsmith is ported from. The LICENSE file ships inside the DMG and the app bundle.

## Build

Requires Go 1.25+, pnpm, Xcode command line tools and the matching Wails CLI:

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.30
wails3 task darwin:package         # → bin/gifsmith.app (ad-hoc signed)
wails3 task darwin:release:build   # → bin/release: Developer ID-signed Gifsmith.app and DMG, verified
wails3 task darwin:release         # the same, notarized and stapled (needs the gifsmith-notary keychain profile)
```

Architecture and the Phase 1 walking skeleton: [.planning/phases/01-native-spike-signed-dev-loop/SKELETON.md](.planning/phases/01-native-spike-signed-dev-loop/SKELETON.md).
