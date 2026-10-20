# Gifsmith 1.0.1

A maintenance release with no feature changes. It is the first release built, signed and notarized by GitHub Actions, and it lets you test the in-app updater: Gifsmith 1.0.0 offers this version from **Settings → General → Updates** and from the menu bar.

## Install or update

- **From 1.0.0:** choose **Update to Gifsmith 1.0.1…** in the menu bar, or **Download and Open…** in Settings → General → Updates. Gifsmith checks the download's checksum, its Developer ID signature and its notarization, then opens the DMG. Quit Gifsmith, drag the new Gifsmith onto Applications, and choose **Replace**.
- **New install:** download `Gifsmith-1.0.1-arm64.dmg` and `Gifsmith-1.0.1-arm64.dmg.sha256` below, then run `shasum -a 256 -c Gifsmith-1.0.1-arm64.dmg.sha256`. It should print `Gifsmith-1.0.1-arm64.dmg: OK`. Open the DMG and drag Gifsmith onto Applications.

## Requirements

- macOS 13 (Ventura) or later, on Apple silicon
- Your own free GIPHY and/or Klipy API key. The Finer Gifs Club needs none.

## Known limits

The limits are the same as in 1.0.0. Klipy file actions stay off until Klipy gives written clearance. Paste is verified in TextEdit, Notes, Mail and Messages.
