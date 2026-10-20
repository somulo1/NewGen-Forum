#!/usr/bin/env bash
# Runs the tamper check (cmd/integrity) and writes the result for monitoring. Exit code 2 means altered records or a broken
# audit chain: the alert LedgerTampered fires. Schedule hourly.
set -uo pipefail
cd "$(dirname "$0")/.."
./bin/hub-integrity "$@"; rc=$?
ok=1; [ $rc -eq 0 ] || ok=0
if [ -n "${HUB_METRICS_DIR:-}" ]; then
  printf '# TYPE hub_integrity_check_ok gauge\nhub_integrity_check_ok %s\n' "$ok" > "$HUB_METRICS_DIR/hub_integrity.prom.$$" && mv "$HUB_METRICS_DIR/hub_integrity.prom.$$" "$HUB_METRICS_DIR/hub_integrity.prom"
fi
exit $rc
