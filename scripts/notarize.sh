#!/usr/bin/env bash
# Submits a file to the notary service, requires status Accepted, then staples the ticket with
# retries. Used by build/darwin/Taskfile.yml release:assemble for the app (zip) and the DMG.
# Usage: scripts/notarize.sh {keychain-profile} {file-to-submit} {path-to-staple}
set -euo pipefail

[ $# -eq 3 ] || { echo "usage: notarize.sh {keychain-profile} {file-to-submit} {path-to-staple}" >&2; exit 2; }
profile=$1
submit=$2
target=$3

# notarytool --wait exits 0 even when the submission is Invalid, so check the status.
out=$(xcrun notarytool submit "$submit" --keychain-profile "$profile" --wait --output-format json)
echo "$out"
status=$(plutil -extract status raw -o - - <<<"$out")
if [ "$status" != Accepted ]; then
  echo "notarization of $submit: status '$status', not Accepted (xcrun notarytool log <id> --keychain-profile $profile)" >&2
  exit 1
fi

# On GitHub's macOS runners the ticket lookup sometimes drops (NSURLError -1005, stapler error 68).
for attempt in 1 2 3 4 5; do
  xcrun stapler staple "$target" && exit 0
  if [ "$attempt" -lt 5 ]; then
    echo "stapler failed (attempt $attempt of 5), retrying in 20 s" >&2
    sleep 20
  fi
done
echo "stapling $target failed after 5 attempts" >&2
exit 1
