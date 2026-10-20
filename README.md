# Gifsmith

[![CI](https://github.com/scottkw/gifsmith/actions/workflows/ci.yml/badge.svg)](https://github.com/scottkw/gifsmith/actions/workflows/ci.yml)
[![Latest release](https://img.shields.io/github/v/release/scottkw/gifsmith)](https://github.com/scottkw/gifsmith/releases/latest)

A macOS menu bar app for finding and sharing animated GIFs. Press a global hotkey (default **⌃⌥⌘G**), search, and copy or paste the GIF into the app you were using. Gifsmith is a port of the Raycast "GIF Search" extension for people who don't use Raycast.

It needs macOS 13 or later on Apple silicon. It is built with Go, [Wails v3](https://v3.wails.io) (pinned to `v3.0.0-beta.30`) and Svelte 5.

## Install

1. Download `Gifsmith-1.0.0-arm64.dmg` and `Gifsmith-1.0.0-arm64.dmg.sha256` from the [latest release](https://github.com/scottkw/gifsmith/releases/latest).
2. [Verify the download](#verify-the-download).
3. Open the DMG and drag **Gifsmith** onto **Applications**.
4. Open Gifsmith from `/Applications`. It is signed with a Developer ID and notarized by Apple, so the only warning is macOS's standard one-time "downloaded from the Internet" prompt. Click **Open**.
5. Gifsmith lives in the menu bar and has no Dock icon. Press **⌃⌥⌘G** anywhere to open the search panel, or click the menu bar icon and choose **Open Gifsmith**.

### Verify the download

Keep both files in the same folder, for example Downloads:

```bash
cd ~/Downloads
shasum -a 256 -c Gifsmith-1.0.0-arm64.dmg.sha256
```

It should print `Gifsmith-1.0.0-arm64.dmg: OK`.

After you install, run `spctl -a -vv /Applications/Gifsmith.app`. It should report `source=Notarized Developer ID` and `origin=Developer ID Application: Ken Scott (S2K7P43927)`.

## Get API keys

Gifsmith never ships with keys. You add your own free keys in **Settings → Keys**. They are stored in the macOS Keychain, never in a file.

- **GIPHY**: create an app at the [GIPHY developer dashboard](https://developers.giphy.com/dashboard/) and copy its API key.
- **Klipy**: sign up at the [Klipy Partner Panel](https://partner.klipy.com) and copy your app key.

New GIPHY and Klipy keys are test keys, limited to **100 requests per hour**. Gifsmith caches trending results and waits until you stop typing before it searches, so normal use stays well under that. If you hit the limit, the panel tells you, and search works again within the hour.

The Finer Gifs Club needs no key.

## Use

| Key | Action |
|-----|--------|
| **⌃⌥⌘G** | Open or hide the panel (rebindable in Settings → Hotkeys) |
| **↵** | Copy the selected GIF |
| **⌘↵** | Paste the selected GIF into the app you were using |
| **Arrow keys** | Move through the grid |
| **⌘K** | List every action: copy, paste, link, Markdown, square crop, download, favorite |
| **⌘1 – ⌘4** | Switch provider |
| **⌘5 / ⌘6** | Show Favorites / Recents |
| **⌘,** | Open Settings |
| **Esc** | Hide the panel |

## Settings

Open **Settings…** from the menu bar icon, or press **⌘,** in the panel.

- **General**: theme, open at login, automatic paste (Accessibility), the ↵ and ⌘↵ actions, and updates.
- **Keys**: add, test, change or remove the GIPHY and Klipy keys.
- **Hotkeys**: change the Open Gifsmith shortcut, and add shortcuts for Favorites and Recents.
- **Search**: grid size, GIFs per page, GIPHY content rating and language, and Klipy country.
- **Downloads**: the download folder, and whether downloaded files keep their original names.
- **About**: version, links and credits.

## Updating

About 10 seconds after launch, and then once a day, Gifsmith checks [GitHub Releases](https://github.com/scottkw/gifsmith/releases) for a newer version. Each check is one unauthenticated GET to `api.github.com` (`/repos/scottkw/gifsmith/releases/latest`). It sends no identifiers: no account, no device ID and no usage data.

To turn automatic checks off, clear **Check for updates automatically** in **Settings → General → Updates**. **Check for Updates…** in the same place still checks when you ask.

When a newer version exists, the menu bar shows **Update to Gifsmith _version_…**, and Settings offers **Download and Open…**. Gifsmith then:

1. Downloads the DMG from GitHub.
2. Checks its SHA-256 checksum, its Developer ID signature (Team ID S2K7P43927) and its Apple notarization.
3. Opens the DMG only if every check passes, so you can drag the new Gifsmith to Applications as usual.

Gifsmith never replaces itself. If any check fails, it deletes the download and opens nothing.

## Attribution

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="frontend/public/attribution/giphy-powered-dark.png">
  <img src="frontend/public/attribution/giphy-powered-light.png" alt="Powered by GIPHY" height="20">
</picture>

- **Powered By GIPHY.** GIPHY and GIPHY Clips search use the [GIPHY API](https://developers.giphy.com/). The panel shows the official mark whenever GIPHY results are on screen.
- **KLIPY.** The Klipy search box reads **Search KLIPY**, and Klipy results stay in their own tab, in the order Klipy returns them.
- **The Finer Gifs Club.** GIFs from [The Finer Gifs Club](https://thefinergifs.club) are credited in the panel footer.
- **The original extension.** Gifsmith is a port of the [GIF Search extension for Raycast](https://github.com/raycast/extensions/tree/main/extensions/gif-search) by Joseph Schmitt and contributors: tm.wrnr, j3lte, samuelkraft, ridemountainpig, spookywy, seanparkross, YourMCGeek, thomas, tim_gailey, clins1994, 0xdhrv, Olli0103 and alexi.build. It is used under the MIT License, © 2021 Raycast.

## Known limits

- **Klipy file actions are turned off** until Klipy gives written clearance for the temporary local copy that copying a GIF needs. On Klipy results, ↵ copies the GIF's link and ⌘↵ pastes the link. Copy GIF, Paste GIF, the square variants and Download are hidden.
- **GIPHY Clips needs an approved GIPHY key.** GIPHY grants Clips access only on request (clips@giphy.com). With an ordinary key, the Clips tab explains how to ask.
- **Pasting needs Accessibility access** (Settings → General → Automatic paste). Copying always works without it. Without the permission, paste actions copy the GIF and tell you to press ⌘V.
- **Paste is verified** in TextEdit, Notes, Mail and Messages. Slack, Discord, Telegram and browser text fields are not tested yet.
- **⌥ + letter hotkeys on non-QWERTY layouts** can bind a different key from the one you pressed, because macOS resolves letters through the active keyboard layout. Use **Test** in Settings → Hotkeys to check a shortcut.
- **The Finer Gifs Club** is a hobby service and is sometimes unavailable. The other providers keep working.
- **Apple silicon only.** There is no Intel build.

## License

MIT. See [LICENSE](LICENSE), which also carries the MIT notice of the Raycast GIF Search extension that Gifsmith is ported from. The LICENSE file ships inside the DMG and the app bundle.

## Build

You need Go 1.25 or later (`go.mod` pins the 1.27.2 toolchain), pnpm, the Xcode command line tools and the matching Wails CLI:

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-beta.30
wails3 task darwin:package         # bin/gifsmith.app, ad-hoc signed
wails3 task darwin:release:build   # bin/release: Developer ID-signed Gifsmith.app and DMG, verified
wails3 task darwin:release         # the same, notarized and stapled (needs the gifsmith-notary keychain profile)
```

**CI:** every push and pull request to `main` runs the test suite. Pushing a `vX.Y.Z` tag builds, notarizes and verifies the DMG on GitHub Actions and creates a draft release. A maintainer must approve the signing step. See [docs/releasing.md](docs/releasing.md).

**Project history:** the v1.0 plans, verification reports and the Phase 1 walking skeleton are archived in [.planning/milestones/](.planning/milestones/). The skeleton is at [v1.0-phases/01-native-spike-signed-dev-loop/SKELETON.md](.planning/milestones/v1.0-phases/01-native-spike-signed-dev-loop/SKELETON.md).
