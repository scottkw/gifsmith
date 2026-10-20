#!/usr/bin/env bash
# Verifies a release directory built by `wails3 task darwin:release:build` (or
# `darwin:release` with --notarized): signatures, bundle contents, the DMG's
# contents (mounted read-only, never shown in Finder) and the checksum.
# Usage: scripts/verify-release.sh [--notarized] {release-dir}
set -euo pipefail

IDENTITY='Developer ID Application: Ken Scott (S2K7P43927)'
TEAM_ID='S2K7P43927'
VERSION='1.0.0'

fail() {
  echo "RELEASE VERIFY FAIL: $*" >&2
  exit 1
}

notarized=false
if [ "${1:-}" = "--notarized" ]; then
  notarized=true
  shift
fi
[ $# -eq 1 ] || fail "usage: verify-release.sh [--notarized] {release-dir}"
dir=$1
app="$dir/Gifsmith.app"
[ -d "$app" ] || fail "missing $app"

shopt -s nullglob
dmgs=("$dir"/Gifsmith-*-arm64.dmg)
shopt -u nullglob
[ ${#dmgs[@]} -eq 1 ] || fail "expected exactly one Gifsmith-*-arm64.dmg in $dir, found ${#dmgs[@]}"
dmg=${dmgs[0]}

# --- app bundle ---
codesign --verify --deep --strict -vv "$app" || fail "codesign --verify --deep --strict (app)"
info=$(codesign -dvv "$app" 2>&1) || fail "codesign -dvv (app)"
grep -qxF "Authority=$IDENTITY" <<<"$info" || fail "app not signed by $IDENTITY"
grep -qF 'flags=0x10000(runtime)' <<<"$info" || fail "app lacks the hardened runtime flag"
ents=$(codesign -d --entitlements - "$app" 2>/dev/null) || fail "codesign --entitlements (app)"
[ -z "$ents" ] || fail "app carries entitlements (none expected)"

plist="$app/Contents/Info.plist"
v=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleShortVersionString' "$plist") || fail "no CFBundleShortVersionString"
[ "$v" = "$VERSION" ] || fail "CFBundleShortVersionString is $v, expected $VERSION"
ui=$(/usr/libexec/PlistBuddy -c 'Print :LSUIElement' "$plist") || fail "no LSUIElement"
[ "$ui" = "true" ] || fail "LSUIElement is $ui, expected true"
[ -f "$app/Contents/Resources/LICENSE" ] || fail "LICENSE missing from Contents/Resources"
exe=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleExecutable' "$plist") || fail "no CFBundleExecutable"
archs=$(lipo -archs "$app/Contents/MacOS/$exe") || fail "lipo -archs"
[ "$archs" = "arm64" ] || fail "executable architectures are '$archs', expected arm64"

# --- DMG ---
codesign --verify -vv "$dmg" || fail "codesign --verify (dmg)"
dinfo=$(codesign -dvv "$dmg" 2>&1) || fail "codesign -dvv (dmg)"
# The in-app updater (04-07) refuses a DMG without this exact line.
grep -qxF "TeamIdentifier=$TEAM_ID" <<<"$dinfo" || fail "dmg TeamIdentifier is not $TEAM_ID"

mnt=$(mktemp -d)
attached=false
cleanup() {
  if $attached; then
    hdiutil detach "$mnt" -quiet || hdiutil detach "$mnt" -force -quiet || true
  fi
  rmdir "$mnt" 2>/dev/null || true
}
trap cleanup EXIT
hdiutil attach -readonly -nobrowse -noautoopen -mountpoint "$mnt" "$dmg" >/dev/null || fail "hdiutil attach"
attached=true
[ -d "$mnt/Gifsmith.app" ] || fail "dmg lacks Gifsmith.app"
[ -L "$mnt/Applications" ] && [ "$(readlink "$mnt/Applications")" = "/Applications" ] || fail "dmg Applications is not a symlink to /Applications"
[ -f "$mnt/LICENSE" ] || fail "dmg lacks LICENSE"
ls -A "$mnt" | grep -q '^\.background' || fail "dmg lacks a .background entry"

(cd "$dir" && shasum -a 256 -c "$(basename "$dmg").sha256") || fail "checksum"

# --- notarization ---
if $notarized; then
  sp=$(spctl -a -vv -t exec "$app" 2>&1) || fail "spctl rejected the app: $sp"
  grep -qF 'source=Notarized Developer ID' <<<"$sp" || fail "app is not 'source=Notarized Developer ID': $sp"
  spd=$(spctl -a -vv -t open --context context:primary-signature "$dmg" 2>&1) || fail "spctl rejected the dmg: $spd"
  xcrun stapler validate "$app" || fail "stapler validate (app)"
  xcrun stapler validate "$dmg" || fail "stapler validate (dmg)"
fi

# --- TCC grant continuity with the Developer ID dev install ---
dev="$HOME/Applications/Gifsmith.app"
if [ -d "$dev" ]; then
  rel_dr=$(codesign -d -r- "$app" 2>&1 | grep '^designated => ') || fail "no designated requirement (release)"
  dev_dr=$(codesign -d -r- "$dev" 2>&1 | grep '^designated => ') || fail "no designated requirement ($dev)"
  echo "release: $rel_dr"
  echo "dev:     $dev_dr"
  [ "$rel_dr" = "$dev_dr" ] || fail "designated requirement differs from $dev (Accessibility grants would not carry over)"
fi

if $notarized; then
  echo "RELEASE VERIFY OK (notarized)"
else
  echo "RELEASE VERIFY OK"
fi
