#!/usr/bin/env bash
# Runs `xcrun stapler {staple|validate} {path}` up to 5 times, 20 s apart. On GitHub's macOS
# runners the ticket lookup sometimes drops (NSURLError -1005, stapler error 68).
# Used by scripts/notarize.sh and scripts/verify-release.sh.
set -euo pipefail

case "$#:${1:-}" in
  2:staple | 2:validate) ;;
  *) echo "usage: stapler-retry.sh {staple|validate} {path}" >&2; exit 2 ;;
esac

for attempt in 1 2 3 4 5; do
  xcrun stapler "$1" "$2" && exit 0
  if [ "$attempt" -lt 5 ]; then
    echo "stapler $1 failed (attempt $attempt of 5), retrying in 20 s" >&2
    sleep 20
  fi
done
echo "stapler $1 $2 failed after 5 attempts" >&2
exit 1
