#!/usr/bin/env bash
# Start everything for local development: Redis (if not already up), api, calls, worker and the Nuxt frontend.
#   ./dev.sh          run all
#   ./dev.sh seed     load demo data first (once), then run all
# Ctrl+C stops everything this script started.
set -euo pipefail
ROOT=$(cd "$(dirname "$0")" && pwd)
cd "$ROOT/backend"
set -a; source .env; set +a

PIDS=()  # (all children share this script's process group, so cleanup kills the group)
cleanup() { trap - EXIT INT TERM; echo; echo "stopping..."; kill 0; }
trap cleanup EXIT INT TERM

REDIS_HOST=${HUB_REDIS_ADDR%:*}; REDIS_PORT=${HUB_REDIS_ADDR##*:}
if ! (exec 3<>"/dev/tcp/$REDIS_HOST/$REDIS_PORT") 2>/dev/null; then
  if [[ "$REDIS_HOST" == 127.0.0.1 || "$REDIS_HOST" == localhost ]] && command -v redis-server >/dev/null; then
    echo "starting redis on :$REDIS_PORT"
    redis-server --port "$REDIS_PORT" --save "" --appendonly no >/dev/null & PIDS+=($!)
    sleep 1
  else
    echo "Redis not reachable at $HUB_REDIS_ADDR and cannot be started locally. Install redis or fix HUB_REDIS_ADDR." >&2; exit 1
  fi
fi

[[ "${1:-}" == seed ]] && go run ./cmd/seed

go build -o /tmp/hub-api ./cmd/api; go build -o /tmp/hub-calls ./cmd/calls; go build -o /tmp/hub-worker ./cmd/worker
prefix() { sed -u "s/^/[$1] /"; }
/tmp/hub-api    2>&1 | prefix api    & PIDS+=($!)
/tmp/hub-calls  2>&1 | prefix calls  & PIDS+=($!)
/tmp/hub-worker 2>&1 | prefix worker & PIDS+=($!)
(cd "$ROOT/frontend" && npm run dev 2>&1 | prefix web) & PIDS+=($!)

echo "api :8080  calls :8090  web :3000   (Ctrl+C to stop)"
wait
