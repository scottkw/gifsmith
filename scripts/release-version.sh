#!/usr/bin/env bash
# Release version gate (D-41). Prints the version from internal/services/updates.go after
# checking that every other version source agrees and the release notes exist.
# Usage: scripts/release-version.sh [vX.Y.Z]   (with a tag, it must be exactly "v" + the version)
set -euo pipefail

cd "$(dirname "$0")/.."

fail() {
  echo "release-version: $*" >&2
  exit 1
}

v=$(sed -n 's/^const Version = "\(.*\)"$/\1/p' internal/services/updates.go)
[[ "$v" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]] || fail "internal/services/updates.go: const Version '$v' is not MAJOR.MINOR.PATCH"

grep -qxF "  VERSION: \"$v\"" build/darwin/Taskfile.yml || fail "build/darwin/Taskfile.yml: VERSION is not \"$v\""
grep -qxF "VERSION='$v'" scripts/verify-release.sh || fail "scripts/verify-release.sh: VERSION is not '$v'"
for key in CFBundleShortVersionString CFBundleVersion; do
  got=$(plutil -extract "$key" raw build/darwin/Info.plist) || fail "build/darwin/Info.plist: no $key"
  [ "$got" = "$v" ] || fail "build/darwin/Info.plist: $key is '$got', not '$v'"
done
[ -f "docs/release-notes-$v.md" ] || fail "docs/release-notes-$v.md is missing"

if [ $# -gt 1 ]; then
  fail "usage: release-version.sh [vX.Y.Z]"
fi
if [ $# -eq 1 ] && [ "$1" != "v$v" ]; then
  fail "tag '$1' does not match version $v (expected v$v)"
fi

echo "$v"
