# Gifsmith 1.0.0

The first release of Gifsmith: press ⌃⌥⌘G anywhere, search GIPHY, GIPHY Clips, Klipy or The Finer Gifs Club, and copy or paste the GIF into the app you were using. A port of the Raycast "GIF Search" extension for people who don't use Raycast.

## Requirements

- macOS 13 (Ventura) or later
- Apple silicon (M1 or later); there is no Intel build
- Your own free GIPHY and/or Klipy API key (Settings → Keys); The Finer Gifs Club needs none

## Install

1. Download `Gifsmith-1.0.0-arm64.dmg` and `Gifsmith-1.0.0-arm64.dmg.sha256` below.
2. Verify the checksum (next section).
3. Open the DMG and drag Gifsmith onto Applications.
4. Open Gifsmith from Applications. It is Developer ID-signed and notarized by Apple.
5. Gifsmith runs in the menu bar. Press ⌃⌥⌘G to search, and open Settings… from the menu bar icon to add your keys.

## Verify the checksum

With both files in the same folder:

```bash
shasum -a 256 -c Gifsmith-1.0.0-arm64.dmg.sha256
```

It should print `Gifsmith-1.0.0-arm64.dmg: OK`.

## Providers and attribution

- Powered By GIPHY: GIPHY and GIPHY Clips search use the GIPHY API.
- Klipy: the Klipy search box reads "Search KLIPY", and Klipy results are shown in their own tab, unaltered.
- GIFs also from The Finer Gifs Club.
- Gifsmith is a port of the GIF Search extension for Raycast by Joseph Schmitt and contributors, used under the MIT License, © 2021 Raycast.

## Known limitation: Klipy

Klipy file actions are off in 1.0.0 while Gifsmith waits for Klipy's written clearance for the temporary local copy that copying a GIF needs. On Klipy results, ↵ copies the GIF's link and ⌘↵ pastes it; Copy GIF, Paste GIF, the square variants and Download are hidden.

## Updates

Gifsmith checks GitHub Releases for later versions (Settings → General → Updates); later versions are offered from GitHub Releases.
