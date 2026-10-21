#!/usr/bin/env bash
# Prints the secrets a production deployment needs, as lines for the backend environment. Run it ONCE per deployment and keep the
# output in your secret store. Never change a key afterwards: the data it protects can then no longer be opened.
#   HUB_KEK          wraps each organization's data key (account emails and every encrypted field depend on it)
#   HUB_BLIND_KEY    makes the searchable index of emails (changing it locks everyone out)
#   HUB_MAC_KEY      seals every protected row against edits made outside Hub
#   HUB_TOKEN_KEY    signs session and reset tokens
#   HUB_LEDGER_KEY   signs the audit ledger
#   HUB_CALLS_SECRET lets the API and the calls service trust each other
#   HUB_DB_GUARD     the token the application presents to the database
#   HUB_METRICS_TOKEN  lets monitoring read /metrics
set -euo pipefail
b64() { openssl rand -base64 32; }
echo "HUB_KEK=$(b64)"
echo "HUB_BLIND_KEY=$(b64)"
echo "HUB_MAC_KEY=$(b64)"
echo "HUB_TOKEN_KEY=$(b64)"
echo "HUB_LEDGER_KEY=$(b64)"
echo "HUB_CALLS_SECRET=$(b64)"
echo "HUB_DB_GUARD=$(openssl rand -hex 32)"
echo "HUB_METRICS_TOKEN=$(openssl rand -hex 24)"
