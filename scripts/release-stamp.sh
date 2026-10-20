#!/usr/bin/env bash
# Ties a release bundle to the commit `darwin:release:app` built it from, so `darwin:release:sign`
# refuses a leftover bin/gifsmith.app (install:dev, package:universal, an older commit) or one
# changed after the build. It proves provenance, not trust: the stamp is written by the build,
# so a compromised build can forge it. CI's signing job therefore never runs the build's code.
# Usage: scripts/release-stamp.sh write|check {app}
set -euo pipefail

[ $# -eq 2 ] || { echo "usage: release-stamp.sh write|check {app}" >&2; exit 2; }
mode=$1
app=$2
stamp="$app.release"

# One SHA-256 over every file's path and contents, so an added, removed or edited file changes it.
digest() {
  (cd "$(dirname "$app")" && find "$(basename "$app")" \( -type f -o -type l \) -print0 |
    LC_ALL=C sort -z | xargs -0 shasum -a 256) | shasum -a 256 | cut -d' ' -f1
}

[ -d "$app" ] || { echo "release-stamp: no $app" >&2; exit 1; }
want="$(git rev-parse HEAD) $(digest)"
case $mode in
  write) printf '%s\n' "$want" > "$stamp" ;;
  check)
    if [ ! -f "$stamp" ] || [ "$(cat "$stamp")" != "$want" ]; then
      echo "refusing to sign: $app is not the darwin:release:app build of $(git rev-parse --short HEAD)" >&2
      exit 1
    fi
    ;;
  *) echo "usage: release-stamp.sh write|check {app}" >&2; exit 2 ;;
esac
