#!/usr/bin/env bash
# Browser UAT loop smoke: CGO-free -tags server build behind the Vite /wails proxy.
# Proves a bound Go method is reachable from a plain browser origin. Stops only the
# processes it started; refuses to run if either port is already taken.
set -euo pipefail

SERVER_PORT="${SERVER_PORT:-8089}"
VITE_PORT="${VITE_PORT:-9245}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

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

PIDS=()
cleanup() {
  for pid in "${PIDS[@]:-}"; do
    [ -n "$pid" ] && kill "$pid" 2>/dev/null || true
  done
  for pid in "${PIDS[@]:-}"; do
    [ -n "$pid" ] && wait "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT

WAILS_SERVER_PORT="$SERVER_PORT" bin/gifsmith-server >bin/smoke-server.log 2>&1 &
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

# method_id <bindings-file> <FuncName>: the numeric $Call.ByID of a generated binding.
method_id() {
  awk -v fn="export function $2(" 'index($0, fn) == 1 {f=1} f && /\$Call\.ByID\(/{match($0, /[0-9]+/); print substr($0, RSTART, RLENGTH); exit}' "$1"
}
# call <methodID> <json-args-array>: POST a bound call through the Vite proxy; body lands in $BODY, prints the HTTP code.
call() {
  curl -s -o "$BODY" -w '%{http_code}' -X POST "http://127.0.0.1:$VITE_PORT/wails/runtime" \
    -H 'Content-Type: application/json' -H 'x-wails-client-id: smoke' \
    -d "{\"object\":0,\"method\":0,\"args\":{\"call-id\":\"1\",\"methodID\":$1,\"args\":$2}}"
}
SVC=frontend/bindings/gifsmith/internal/services
BODY=$(mktemp)
trap 'rm -f "$BODY"; cleanup' EXIT

SEARCH_ID=$(method_id "$SVC/gifservice.ts" Search)
[ -n "$SEARCH_ID" ] || fail "could not find the GifService.Search method ID"
CODE=$(call "$SEARCH_ID" '["finergifs","dwight",""]')
if [ "$CODE" != "200" ] || ! grep -q '"items":\[' "$BODY" || ! grep -q '/media/' "$BODY"; then
  fail "GifService.Search via proxy: HTTP $CODE, body: $(head -c 300 "$BODY")"
fi
echo "GifService.Search(finergifs, dwight) via Vite proxy: HTTP 200, returned items with fixture /media/ URLs"

ID=$(method_id "$SVC/actionservice.ts" Copy)
[ -n "$ID" ] || fail "could not find the Copy method ID"
CODE=$(call "$ID" '[]')
if [ "$CODE" != "422" ] || ! grep -q 'not supported in this build' "$BODY"; then
  fail "ActionService.Copy via proxy: HTTP $CODE, body: $(cat "$BODY")"
fi
echo "ActionService.Copy via Vite proxy: HTTP 422 $(cat "$BODY")"

CODE=$(curl -s -o /dev/null -w '%{http_code}' "http://127.0.0.1:$VITE_PORT/sample.gif")
[ "$CODE" = "200" ] || fail "GET /sample.gif via vite: HTTP $CODE"
echo "GET /sample.gif via vite: 200"

echo "SMOKE OK"
