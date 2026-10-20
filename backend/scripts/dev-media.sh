#!/usr/bin/env bash
# Starts a local media server (LiveKit) for big calls and prints the setting the API needs.
#   scripts/dev-media.sh            then run the API with the printed HUB_SFU_NODES
# The binary comes from https://github.com/livekit/livekit/releases (livekit-server, Apache-2.0).
set -euo pipefail
cd "$(dirname "$0")/.."
BIN="${LIVEKIT_BIN:-$(command -v livekit-server || echo "$HOME/.cache/lk/livekit-server")}"
[ -x "$BIN" ] || { echo "livekit-server not found. Download it from the LiveKit releases page and put it on PATH or set LIVEKIT_BIN." >&2; exit 1; }
echo "export HUB_SFU_NODES='http://127.0.0.1:7880|hubkey|hubsecret-hubsecret-hubsecret-hubsecret'"
exec "$BIN" --config deploy/media/livekit.dev.yaml --node-ip "${HUB_MEDIA_IP:-$(hostname -I | cut -d" " -f1)}"
