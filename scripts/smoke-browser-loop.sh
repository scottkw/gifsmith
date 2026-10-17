#!/usr/bin/env bash
# Browser UAT loop: CGO-free -tags server build (with its in-process fixture
# server) behind the Vite /wails proxy. Stops only the processes it started;
# refuses to run if either port is already taken.
#
#   scripts/smoke-browser-loop.sh                 bound-call smoke checks, ends with SMOKE OK
#   scripts/smoke-browser-loop.sh --serve         start server + Vite and keep them up for dev-browser
#   scripts/smoke-browser-loop.sh --uat <file.js> start both, run the dev-browser script headless, need "UAT OK"
#
# FAKE_FINER=down (and FAKE_CLIPS, from plan 02-06) pass through to the fixture server.
set -euo pipefail

MODE=smoke
UAT_SCRIPT=""
case "${1:-}" in
  "") ;;
  --serve) MODE=serve ;;
  --uat)
    MODE=uat
    UAT_SCRIPT="${2:?usage: $0 --uat <script.js>}"
    ;;
  *) echo "usage: $0 [--serve | --uat <script.js>]" >&2; exit 2 ;;
esac

SERVER_PORT="${SERVER_PORT:-8089}"
VITE_PORT="${VITE_PORT:-9245}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
[ -z "$UAT_SCRIPT" ] || [ -f "$UAT_SCRIPT" ] || { echo "no such UAT script: $UAT_SCRIPT" >&2; exit 2; }

fail() { echo "SMOKE FAIL: $*" >&2; exit 1; }

for port in "$SERVER_PORT" "$VITE_PORT"; do
  if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "port $port already has a listener; not touching it:" >&2
    lsof -nP -iTCP:"$port" -sTCP:LISTEN >&2 || true
    exit 1
  fi
done

echo "building bin/gifsmith-server (CGO_ENABLED=0, -tags server)"
mkdir -p bin
CGO_ENABLED=0 go build -tags server -o bin/gifsmith-server .
# Fresh first-run state every run (UAT-only dirs; the real app uses the user's cache/config dirs).
rm -rf bin/uat-config bin/uat-cache

PIDS=()
BODY=$(mktemp)
cleanup() {
  rm -f "$BODY"
  for pid in "${PIDS[@]:-}"; do
    [ -n "$pid" ] && kill "$pid" 2>/dev/null || true
  done
  for pid in "${PIDS[@]:-}"; do
    [ -n "$pid" ] && wait "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT

WAILS_SERVER_PORT="$SERVER_PORT" FAKE_FINER="${FAKE_FINER:-}" FAKE_CLIPS="${FAKE_CLIPS:-}" \
  bin/gifsmith-server >bin/smoke-server.log 2>&1 &
PIDS+=("$!")
# exec so the recorded PID is vite itself, not a wrapper subshell.
(cd frontend && GIFSMITH_BACKEND="http://localhost:$SERVER_PORT" WAILS_VITE_PORT="$VITE_PORT" \
  exec node_modules/.bin/vite) >bin/smoke-vite.log 2>&1 &
PIDS+=("$!")

wait_for() {
  local url=$1
  for _ in $(seq 1 60); do
    curl -sf -o /dev/null "$url" && return 0
    sleep 0.5
  done
  return 1
}
wait_for "http://127.0.0.1:$SERVER_PORT/health" || fail "server /health not up after 30 s (see bin/smoke-server.log)"
wait_for "http://127.0.0.1:$VITE_PORT/" || fail "vite not up after 30 s (see bin/smoke-vite.log)"

if [ "$MODE" = serve ]; then
  echo "server: http://127.0.0.1:$SERVER_PORT  ui: http://127.0.0.1:$VITE_PORT  (Ctrl-C stops both)"
  wait
  exit 0
fi

if [ "$MODE" = uat ]; then
  OUT=$(dev-browser --headless --browser gifsmith-uat --timeout 120 run "$UAT_SCRIPT" 2>&1) || true
  echo "$OUT"
  grep -q "UAT OK" <<<"$OUT" || { echo "UAT FAIL: $UAT_SCRIPT" >&2; exit 1; }
  exit 0
fi

# method_id <bindings-file> <FuncName>: the numeric $Call.ByID of a generated binding.
method_id() {
  awk -v fn="export function $2(" 'index($0, fn) == 1 {f=1} f && /\$Call\.ByID\(/{match($0, /[0-9]+/); print substr($0, RSTART, RLENGTH); exit}' "$1"
}
# call <methodID> <json-args-array>: POST a bound call through the Vite proxy; body lands in $BODY, prints the HTTP code.
call() {
  curl -s -o "$BODY" -w '%{http_code}' -X POST "http://127.0.0.1:$VITE_PORT/wails/runtime" \
    -H 'Content-Type: application/json' -H 'x-wails-client-id: smoke' \
    --data-binary "{\"object\":0,\"method\":0,\"args\":{\"call-id\":\"1\",\"methodID\":$1,\"args\":$2}}"
}
SVC=frontend/bindings/gifsmith/internal/services

SEARCH_ID=$(method_id "$SVC/gifservice.ts" Search)
[ -n "$SEARCH_ID" ] || fail "could not find the GifService.Search method ID"
CODE=$(call "$SEARCH_ID" '["finergifs","dwight",""]')
if [ "$CODE" != "200" ] || ! grep -q '"items":\[' "$BODY" || ! grep -q '/media/' "$BODY"; then
  fail "GifService.Search via proxy: HTTP $CODE, body: $(head -c 300 "$BODY")"
fi
echo "GifService.Search(finergifs, dwight) via Vite proxy: HTTP 200, returned items with fixture /media/ URLs"

ITEM=$(node -e 'const p = JSON.parse(require("fs").readFileSync(0, "utf8")); process.stdout.write(JSON.stringify(p.items[0]))' <"$BODY")
PREVIEW=$(node -e 'process.stdout.write(JSON.parse(process.argv[1]).previewUrl)' "$ITEM")
MAGIC=$(curl -sf "$PREVIEW" | head -c 4) || fail "GET $PREVIEW failed"
[ "$MAGIC" = "GIF8" ] || fail "GET $PREVIEW: first bytes '$MAGIC', want GIF8"
echo "GET $PREVIEW: 200, image/gif bytes start GIF8"

COPY_ID=$(method_id "$SVC/actionservice.ts" Copy)
[ -n "$COPY_ID" ] || fail "could not find the ActionService.Copy method ID"
CODE=$(call "$COPY_ID" "[$ITEM]")
if [ "$CODE" != "422" ] || ! grep -q 'not supported in this build' "$BODY"; then
  fail "ActionService.Copy(gif) via proxy: HTTP $CODE, body: $(cat "$BODY")"
fi
STAGED=$(find bin/uat-cache/staging -name '*.gif' -type f 2>/dev/null | head -1)
[ -n "$STAGED" ] || fail "ActionService.Copy(gif) staged no .gif under bin/uat-cache/staging"
echo "ActionService.Copy(gif): staged $STAGED, then HTTP 422 $(cat "$BODY")"

EVIL=$(node -e 'const g = JSON.parse(process.argv[1]); g.downloadUrl = "https://evil.example.test/x.gif"; process.stdout.write(JSON.stringify(g))' "$ITEM")
CODE=$(call "$COPY_ID" "[$EVIL]")
if [ "$CODE" = "200" ] || grep -q 'evil.example' "$BODY"; then
  fail "ActionService.Copy(disallowed URL): HTTP $CODE, body: $(cat "$BODY")"
fi
echo "ActionService.Copy(disallowed URL): HTTP $CODE $(cat "$BODY")"

# Keys: fixture keys only (bad-key-123 is rejected by the fixture, uat-key-123 accepted). Never a real key.
SET="$SVC/settingsservice.ts"
SAVE_ID=$(method_id "$SET" SaveKey)
LAST_ID=$(method_id "$SET" SetLastProvider)
PROV_ID=$(method_id "$SVC/gifservice.ts" Providers)
TREND_ID=$(method_id "$SVC/gifservice.ts" Trending)
[ -n "$SAVE_ID" ] && [ -n "$LAST_ID" ] && [ -n "$PROV_ID" ] && [ -n "$TREND_ID" ] || fail "missing SettingsService/GifService method IDs"

for p in giphy klipy; do
  CODE=$(call "$SAVE_ID" "[\"$p\",\"bad-key-123\"]")
  if [ "$CODE" = "200" ] || ! grep -q 'key_rejected' "$BODY" || grep -q 'bad-key-123' "$BODY"; then
    fail "SaveKey($p, bad key): HTTP $CODE, body: $(cat "$BODY")"
  fi
  echo "SettingsService.SaveKey($p, bad key): HTTP $CODE $(cat "$BODY")"
done

CODE=$(call "$PROV_ID" '[]')
[ "$CODE" = "200" ] && grep -q '"id":"giphy","configured":false' "$BODY" || fail "Providers before save: HTTP $CODE, body: $(cat "$BODY")"
echo "GifService.Providers before save: $(cat "$BODY")"

CODE=$(call "$SAVE_ID" '["giphy","uat-key-123"]')
[ "$CODE" = "200" ] || fail "SaveKey(giphy, fixture key): HTTP $CODE, body: $(cat "$BODY")"
CODE=$(call "$PROV_ID" '[]')
[ "$CODE" = "200" ] && grep -q '"id":"giphy","configured":true' "$BODY" || fail "Providers after save: HTTP $CODE, body: $(cat "$BODY")"
echo "SaveKey(giphy, fixture key): 200; Providers now $(cat "$BODY")"

# The GIPHY save starts the launch's single Clips probe in the background.
for _ in $(seq 1 20); do
  CODE=$(call "$PROV_ID" '[]')
  grep -q '"id":"clips","configured":true,"clips":"enabled"' "$BODY" && break
  sleep 0.25
done
grep -q '"id":"clips","configured":true,"clips":"enabled"' "$BODY" || fail "Clips never reached enabled: HTTP $CODE, body: $(cat "$BODY")"
CODE=$(call "$TREND_ID" '["clips",""]')
[ "$CODE" = "200" ] && grep -q '"isClip":true' "$BODY" || fail "Trending(clips): HTTP $CODE, body: $(head -c 300 "$BODY")"
echo "Clips probe: enabled; GifService.Trending(clips) returned isClip items"

CODE=$(call "$TREND_ID" '["giphy",""]')
[ "$CODE" = "200" ] && grep -q '/media/giphy-' "$BODY" || fail "Trending(giphy): HTTP $CODE, body: $(head -c 300 "$BODY")"
echo "GifService.Trending(giphy): HTTP 200 with fixture /media/giphy- URLs"

CODE=$(call "$SAVE_ID" '["klipy","uat-key-123"]')
[ "$CODE" = "200" ] || fail "SaveKey(klipy, fixture key): HTTP $CODE, body: $(cat "$BODY")"
CODE=$(call "$TREND_ID" '["klipy",""]')
[ "$CODE" = "200" ] && grep -q '/media/klipy-' "$BODY" || fail "Trending(klipy): HTTP $CODE, body: $(head -c 300 "$BODY")"
DUPS=$(grep -o '"id":"klipy-dup"' "$BODY" | wc -l | tr -d ' ')
[ "$DUPS" = "2" ] || fail "Trending(klipy): klipy-dup appears $DUPS times, want 2 (Klipy pages are never deduped)"
echo "SaveKey(klipy, fixture key): 200; Trending(klipy) keeps both klipy-dup items"

CODE=$(call "$LAST_ID" '["giphy"]')
[ "$CODE" = "200" ] || fail "SetLastProvider(giphy): HTTP $CODE, body: $(cat "$BODY")"
grep -q '"lastProvider":"giphy"' bin/uat-config/settings.json || fail "settings.json: $(cat bin/uat-config/settings.json 2>&1)"
echo "SetLastProvider(giphy): bin/uat-config/settings.json = $(cat bin/uat-config/settings.json)"

for f in bin/uat-config/settings.json bin/smoke-server.log; do
  if grep -q -e 'uat-key-123' -e 'bad-key-123' "$f"; then
    fail "a key appears in $f"
  fi
done
echo "no fixture key in settings.json or the server log"

echo "SMOKE OK"
