# 2. Architecture

## 2.1 The pieces

```
                        ┌──────────────────────────── Hub (browser) ─────────────────────────────┐
                        │  Calls management screens          Yeastar WebRTC SDK (the softphone)    │
                        └───────┬──────────────────────────────────────┬──────────────────────────┘
               HTTPS + WSS      │ (REST + live events)                  │ SIP over WSS + SRTP audio
                                ▼                                       ▼
  ┌──────────────── Call service (callsvc, Go) ────────────────┐   ┌──────── Yeastar P-Series PBX ────────┐
  │ API gateway  ─ auth via Hub session + rights              │   │ Extensions, queues, IVR, ring groups │
  │ PBX connector ─ Open API client + event stream consumer   │◄──┤ Recording, voicemail, CDR            │
  │ Call state store (Postgres + Redis)                       │──►│ Open API (REST + WebSocket events)   │
  │ Caller ID resolver ─ LMS adapter + cache                  │   └───────┬──────────────────────────────┘
  │ Recording archiver ─ encrypted object storage             │           │ SIP trunks (G.711)
  │ Campaign dialer                                           │           ▼
  │ Messaging adapter ─ SMS / WhatsApp providers              │   ┌──────── Carriers ────────┐
  │ Reports aggregator, Audit log                             │   │ Safaricom / Airtel / SIP │──► Customers' phones
  └───────┬───────────────────────────┬───────────────────────┘   └──────────────────────────┘
          │ HTTPS (read-only)          │ HTTPS
          ▼                            ▼
  ┌──── Loan Management System ───┐   ┌── SMS / WhatsApp providers ──┐
  │ Customers, phones, accounts   │   │ e.g. Africa's Talking, Meta  │
  └───────────────────────────────┘   └──────────────────────────────┘
```

## 2.2 Who does what

| Component | Responsibility | Never does |
|---|---|---|
| **Hub frontend** | Screens, the softphone in the browser (via Yeastar's WebRTC SDK), showing live state pushed by the call service. | Talk to the LMS or the PBX Open API directly; hold PBX credentials. |
| **Hub backend** (existing Go app) | Sign-in, sessions, roles and rights, widget switches ("Calls management" on/off and who can see it). | Call logic. It only answers "who is this user and what may they do". |
| **Call service** (`callsvc`, new) | Everything about calls: control, state, caller lookup, recordings, campaigns, messaging, reports, audit. Owns its own database. | Store LMS customer data permanently (it caches briefly). |
| **Yeastar PBX** | The phone system itself: registrations, routing, IVR, queues, mixing audio, recording, CDR, trunks. | Know who our customers are (the call service tells Hub). |
| **LMS** | Source of truth for customers, phone numbers, accounts, balances, KYC status. | Get written to by the call service (read-only integration). |
| **Carriers** | Deliver calls to and from real phone numbers. | — |

## 2.3 Why a separate microservice

- **Different load and failure profile.** Call events are continuous and bursty (thousands per hour); Hub's API is request/response. A slow call component must never slow Hub's sign-in, and vice versa.
- **Different secrets.** PBX API credentials and LMS credentials live only in the call service.
- **Different compliance scope.** Recordings and customer financial data have stricter retention and audit rules than the rest of Hub.
- **Replaceable PBX.** All Yeastar-specific code is in one adapter (the PBX connector). If we ever move to another PBX (e.g. Asterisk/FreePBX), only that adapter changes.

The service is written in **Go**, like the Hub backend, uses **PostgreSQL** (its own database, not Hub's) and **Redis** (live state, caching, rate limits), and is deployed as its own container.

## 2.4 How Hub and the call service trust each other

1. The agent is signed in to Hub as usual.
2. When the Calls management page opens, Hub's frontend asks Hub's backend for a **short-lived call-service token** (a signed JWT, 5 minutes, refreshed in the background). It contains: person ID, organization ID, and the person's calls-management rights (e.g. `calls.use`, `calls.supervise`, `calls.recordings.play`, `calls.campaigns.manage`).
3. The frontend sends that token to the call service on every request and when opening the live event WebSocket.
4. The call service verifies the signature with Hub's public key and **re-checks the right needed for each action**. A missing right is a `403`, logged.
5. The person-to-extension mapping (person 7c1… is extension 1003) is stored in the call service and managed by admins.

This means the access rule "only the people the organization chose" is enforced on the server, not only by hiding the page.

## 2.5 Live data flow (events)

```
PBX event (call ringing on ext 1003, caller +254712345678)
   → PBX connector receives it on the Open API event WebSocket
   → normalises it into our own event: call.ringing {call_id, ext 1003, from +254712345678, queue "Collections"}
   → Caller ID resolver: Redis cache? else LMS lookup (≤ 300 ms budget)
   → writes call state (Redis) + call row (Postgres)
   → publishes to the agent's live channel and to supervisors' wallboard channel
   → Hub shows the ringing card with "Grace Wanjiku · LN-20431 · In arrears (12 days)"
```

Every PBX event is stored raw (for replay and debugging) and as a normalised event, keyed by the PBX's call/channel IDs so duplicates and out-of-order delivery are harmless (idempotent processing).

## 2.6 Request flow (actions)

```
Agent clicks "Transfer to 1004 (attended)" in Hub
   → POST /v1/calls/{id}/transfer {to: "1004", mode: "attended"} with call-service token
   → service checks right calls.use, checks the call belongs to this agent
   → PBX connector calls the Yeastar API (or the browser SDK does it locally, see 03)
   → PBX emits events (consult leg ringing, answered, transfer complete)
   → service updates state, writes audit entry "1001 transferred call X to 1004"
   → Hub panels update from the live channel
```

Rule: **the browser never trusts its own view as the truth**. It sends commands and redraws from the events the service pushes, so every open screen (agent, supervisor, wallboard) agrees.

## 2.7 Where things are stored

| Data | Store | Kept |
|---|---|---|
| Live call state (ringing, talking, on hold, who, since when) | Redis | Until the call ends + 1 hour |
| Calls (one row per call, joined from PBX CDR + our events) | Postgres `calls` | 7 years (financial records; confirm with legal) |
| Notes, outcomes (dispositions) | Postgres `call_notes`, `calls.disposition` | With the call |
| Recordings | Encrypted object storage (S3-compatible), metadata in Postgres | Per retention setting (default 1 year, legal holds longer) |
| Customer name/account at time of call | Snapshot on the call row (name, account no.) | With the call. Not the full LMS record. |
| LMS lookup cache | Redis | 10 minutes |
| Audit log | Postgres `audit_events`, append-only, hash-chained | 7 years |
| Campaigns, DNC list, routing settings | Postgres | Until deleted |
| Messages (SMS/WhatsApp) | Postgres `messages` | 7 years |
