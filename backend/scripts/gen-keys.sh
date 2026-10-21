#!/usr/bin/env bash
# Generates every secret the backend needs for a new deployment, each from the system's secure random source.
#
#   scripts/gen-keys.sh                 print them
#   scripts/gen-keys.sh -o FILE         write them to FILE (mode 0600; refuses to overwrite an existing file)
#
# Put the result in your platform's environment (or /etc/hub/hub.env) and KEEP A COPY somewhere safe: the six keys cannot be
# changed afterwards, because data already stored with one key cannot be opened with another. Never reuse a development key in
# production, and never commit the output.
set -euo pipefail
command -v openssl >/dev/null || { echo "openssl is needed" >&2; exit 1; }

b64()  { openssl rand -base64 32; }                                    # 32 random bytes, base64 (the six keys)
word() { openssl rand -base64 64 | tr -dc 'A-Za-z0-9' | head -c "$1"; } # letters and digits only: safe inside a database URL
hex()  { openssl rand -hex 32; }

emit() {
cat <<OUT
# Generated $(date -u +%Y-%m-%dT%H:%M:%SZ) by scripts/gen-keys.sh. Keep a private copy; the six keys must never change.

# ---- the six keys: each 32 random bytes in base64, all different ----
HUB_KEK=$(b64)
HUB_MAC_KEY=$(b64)
HUB_BLIND_KEY=$(b64)
HUB_TOKEN_KEY=$(b64)
HUB_LEDGER_KEY=$(b64)
HUB_CALLS_SECRET=$(b64)

# ---- proves to the database that a write comes from Hub itself ----
HUB_DB_GUARD=$(hex)

# ---- database role passwords (used by scripts/bootstrap_remote.sh and inside the three database URLs) ----
HUB_OWNER_PW=$(word 32)
HUB_APP_PW=$(word 32)
HUB_AUDIT_PW=$(word 32)
HUB_DBA_PW=$(word 32)

# ---- monitoring and the media server (LiveKit) ----
HUB_METRICS_TOKEN=$(word 40)
HUB_SFU_KEY=API$(word 14)
HUB_SFU_SECRET=$(word 48)
OUT
}

if [ "${1:-}" = "-o" ]; then
  out="${2:?usage: gen-keys.sh -o FILE}"
  [ -e "$out" ] && { echo "$out already exists; not overwriting it" >&2; exit 1; }
  umask 077
  emit > "$out"
  echo "written to $out (mode 0600)"
else
  emit
fi
