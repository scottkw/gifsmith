#!/usr/bin/env bash
# Submits a file to the notary service, requires status Accepted, then staples the ticket with
# retries (scripts/stapler-retry.sh). Used by build/darwin/Taskfile.yml release:assemble for the
# app (zip) and the DMG.
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

bash "$(dirname "$0")/stapler-retry.sh" staple "$target"
