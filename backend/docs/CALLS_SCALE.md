# Calls at scale

Two kinds of room, chosen by the API when a call starts (`mode` in the answer to `POST /calls`):

| | Who | Runs on | Limit |
|---|---|---|---|
| `mesh` | private (one-to-one) calls, or any call when no media server is configured | `cmd/calls` (signalling only; devices send media to each other) | 8 people |
| `sfu` | forum calls and meetings when `HUB_SFU_NODES` is set | a media server fleet ([LiveKit](https://github.com/livekit/livekit), Apache-2.0, Go) | `HUB_CALL_MAX_VOICE` / `HUB_CALL_MAX_VIDEO` (default 800) |

The API never touches media. It decides who may join, hands out a signed pass for one media server, enforces the limits, and
carries out host controls. Media servers report who joined and left through a signed webhook (`POST /v1/calls/sfu-hook`), which
feeds the same event list as `cmd/calls`, so call history works the same either way.

## How a big room works
* **Stage and audience.** Up to `HUB_CALL_OPEN_MIC` (25) people, everyone may speak. Beyond that, newcomers join as listeners.
  Hosts and moderators always speak, and can bring a listener on stage (`stage_on`) or send them back (`stage_off`). At most
  `HUB_CALL_STAGE` (25) people speak at once, so a listener receives a few streams, not hundreds. A listener's hand raise
  is a participant attribute the host sees in the people list.
* **Video quality.** Publishers send three layers (simulcast); the media server sends each viewer the layer that fits the size
  the tile is drawn at, and pauses video nobody sees (adaptive stream, dynacast). Audio uses DTX, so silent people send nothing.
* **Encryption.** Media is encrypted on the device before it leaves, with a key derived (HKDF) from the conversation's key, so
  the media server relays video it cannot watch. Calls with no conversation are encrypted to the media server only; the UI says so.
* **Limits.** The API counts people and rooms in Redis (shared by every API node) and the media server enforces the same
  per-room cap itself, so a wrong count cannot let a room grow past it. A full room answers `409 room_full`, too many rooms
  in one organization `429 too_many_calls`, a full deployment `503 calls_busy`. Nothing degrades for the people already in.
* **Controls** (`POST /calls/{id}/control`): `mute`, `cam_off`, `remove`, `mute_all`, `lock`, `unlock`, `end`, `stage_on`, `stage_off`.
  A moderator can only switch a mic or camera off. Only the host can end the call for everyone or remove the host.

## Settings (API)
```
HUB_SFU_NODES        https://media-eu.example.com|APIKEY|SECRET,https://media-af.example.com|APIKEY|SECRET
HUB_CALL_MAX_VOICE   800        people in one voice room
HUB_CALL_MAX_VIDEO   800        people in one video room
HUB_CALL_OPEN_MIC    25         rooms up to this size: everyone speaks
HUB_CALL_STAGE       25         most people speaking at once in a bigger room
HUB_CALL_ROOMS_ORG   100        rooms running at once in one organization
HUB_CALL_ROOMS_TOTAL 100000     rooms running at once on the deployment
```
A room is pinned to one node (a region's address) by rendezvous hashing, so every API node picks the same one. In development the
API uses `scripts/dev-media.sh`'s server on `127.0.0.1:7880` automatically when it is running.

## Running it
`deploy/media/livekit.prod.yaml` is a media server's settings: Redis so servers form one cluster (a room can span servers, and
servers can be added or drained without dropping calls), a UDP port range, TURN for people behind strict firewalls, load-based
refusal of new rooms, and `auto_create: false` so only the API opens rooms. Put each region's servers behind one address.
Use bare metal or a provider with cheap bandwidth: egress is the main cost.

## Sizing (estimates; measure before promising)
A listener receives about 3 active speakers' audio (~40 kbps each, ~120 kbps) plus the tiles they can see. For a room of N
listeners with a stage of 4 people on video (~1-2 Mbps to each viewer at 540p), egress is roughly `N × 2 Mbps`:
an 800-person room is about 1.6 Gbps, so 100 such rooms are 160 Gbps. A voice-only 800-person room is about 100 Mbps.
One well-provisioned node (16+ cores, 10 Gbps NIC) is realistically a few hundred to a few thousand subscribers depending on
video; plan a fleet with 30-40% headroom and scale on CPU and NIC use.

## Load testing
Run the generator on machines other than the media server, or the numbers are meaningless (on one laptop the generator
starves the server and shows packet loss that production would not have).
```
lk load-test --url wss://media-eu.example.com --api-key KEY --api-secret SECRET \
  --room loadtest-1 --video-publishers 4 --audio-publishers 20 --subscribers 800 --num-per-second 10 --duration 10m --simulate-speakers
```
Run many rooms at once (one command per room, several generator hosts), then kill a media node mid-test and check people
reconnect. The release gate is: no refused joins below the limits, p95 audio loss < 1%, no node over 80% CPU or NIC.

## Not done yet
* Rotating the call key when someone leaves mid-call (the key follows the conversation epoch at join).
* Cross-region cascading for one room is what a LiveKit cluster provides; it needs the multi-region deployment and a test.
* Server-side recording (impossible while media is end-to-end encrypted; would need an explicit, announced opt-out of encryption).
* A listener-count banner and "raise hand" queue ordering for very large rooms.
