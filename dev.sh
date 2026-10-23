#!/usr/bin/env bash
# Start everything for local development: Redis (if not already up), api (which migrates the database and runs the background
# jobs itself), calls and the Nuxt frontend.
#   ./dev.sh          run all
# Ctrl+C stops everything this script started.
set -euo pipefail
ROOT=$(cd "$(dirname "$0")" && pwd)
cd "$ROOT/backend"
mkdir -p bin "$HOME/.cache/gotmp"; export GOTMPDIR="$HOME/.cache/gotmp"   # /tmp is small on some machines
set -a; source .env; set +a
HUB_REDIS_ADDR=${HUB_REDIS_ADDR:-127.0.0.1:6390}   # the same default the server uses; .env only needs it to point elsewhere

PIDS=()  # (all children share this script's process group, so cleanup kills the group)
cleanup() { trap - EXIT INT TERM; echo; echo "stopping..."; kill 0; }
trap cleanup EXIT INT TERM

REDIS_OFF=0; [[ "${HUB_REDIS:-}" == off ]] && REDIS_OFF=1   # no Redis at all: the API keeps an in-memory one, and serves the calls signalling itself

if [[ $REDIS_OFF == 0 ]]; then
LOCAL_ADDR=${HUB_REDIS_ADDR#*://}; LOCAL_ADDR=${LOCAL_ADDR##*@}   # accept redis://user:pass@host:port too
REDIS_HOST=${LOCAL_ADDR%:*}; REDIS_PORT=${LOCAL_ADDR##*:}
if ! (exec 3<>"/dev/tcp/$REDIS_HOST/$REDIS_PORT") 2>/dev/null; then
  if [[ "$REDIS_HOST" == 127.0.0.1 || "$REDIS_HOST" == localhost ]] && command -v redis-server >/dev/null; then
    echo "starting redis on :$REDIS_PORT"
    redis-server --port "$REDIS_PORT" --save "" --appendonly no >/dev/null & PIDS+=($!)
    sleep 1
  else
    echo "Redis not reachable at $HUB_REDIS_ADDR and cannot be started locally. Install redis or fix HUB_REDIS_ADDR." >&2; exit 1
  fi
fi
fi


go build -o bin/hub-api . ; go build -o bin/hub-calls ./cmd/calls
prefix() { sed -u "s/^/[$1] /"; }
if [[ $REDIS_OFF == 1 ]]; then
  HUB_WITH_CALLS=on HUB_CALLS_URL="http://127.0.0.1:${HUB_HTTP_PORT:-8080}/calls-signal" bin/hub-api 2>&1 | prefix api & PIDS+=($!)
else
  bin/hub-api    2>&1 | prefix api    & PIDS+=($!)
  bin/hub-calls  2>&1 | prefix calls  & PIDS+=($!)
fi
(cd "$ROOT/frontend" && npm run dev 2>&1 | prefix web) & PIDS+=($!)

echo "api :8080  calls :8090  web :3000   (Ctrl+C to stop)"
wait
