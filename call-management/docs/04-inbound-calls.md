# 4. Inbound calls: a customer calls us

Every step from the customer pressing "call" to an agent saying hello, with what each system does and how long it may take.

## 4.1 The journey

```
Customer dials 0709 000 200
 │
 ▼
[1] Carrier → SIP INVITE on our trunk to DID +254709000200          (0 s)
 │
 ▼
[2] PBX inbound route for that DID → time condition: are we open?
 │      ├── closed / holiday → after-hours message → voicemail box "Customer care"
 │      └── open ↓
 ▼
[3] Greeting + IVR menu (care tune underneath)                       (~8–15 s)
 │      "For loans press 1, wallet 2, fraud 3, an agent 0"
 │      no key after N s → repeat (R times) → default option
 ▼
[4] Queue (e.g. "Collections")                                        
 │      position announcement every X s, care tune, "press 1 for a call back"
 │      ├── press 1 → call-back request saved → caller hangs up
 │      ├── waited longer than max → voicemail box of the queue
 │      └── agent free ↓
 ▼
[5] Queue rings an agent's extension (strategy: longest idle / ring all / ...)
 │      PBX event → call service → LMS lookup → Hub ringing card       (target < 1 s)
 ▼
[6] Agent answers in Hub (WebRTC) → talk, recorded from the first second
 ▼
[7] Hang up → after-call work (outcome + notes) → agent ready again
 ▼
[8] PBX CDR + recording → call service joins them to the call row → archive
```

## 4.2 Step by step

### Step 1 — the call arrives
- The carrier sends `INVITE sip:+254709000200@pbx` with the caller's number in the `From`/`P-Asserted-Identity` header.
- **Caller ID can be withheld or spoofed.** A number shown is a hint, not proof of identity (see 06).
- If all trunk channels are busy, the carrier returns busy to the customer and we never see the call. The wallboard shows trunk usage so this is visible before it happens.

### Step 2 — are we open?
- Configured in Routing → "Are we open?": weekly hours, holidays, and where closed calls go.
- Kenyan public holidays are pre-loaded each year and checked by an admin.
- The fraud line (option 3) can be set to "always open" and routed to an on-call group.

### Step 3 — greeting and menu
- The greeting is a **recorded audio file** uploaded to the PBX (professional voice or our own recording). The browser's AI voice in Hub is only for previewing; the PBX plays real audio files to customers.
- Keys arrive as RFC 4733 DTMF events.
- Every key the caller pressed is stored on the call (`ivr_path`, e.g. `1`), so reports show why people called.

### Step 4 — waiting in the queue
- **Position announcement:** "You are caller number 3." Every X seconds (Routing setting).
- **Call-back offer:** if the caller presses 1, the PBX/IVR captures it, the call service creates a `callback_requests` row (number, queue, time, position), and the caller hears "we will call you back". The request appears in Routing → "While they wait" and on the wallboard, and is assigned to the next free agent in that queue as an outbound task.
- **Maximum wait:** after that many seconds the caller goes to the queue's voicemail box.
- **Abandoned:** a caller who hangs up while waiting is an *abandoned call*. It counts against service level and appears in the queue's missed list so someone can call back.

### Step 5 — ringing the agent (the critical second)

| Time | What happens |
|---|---|
| 0 ms | PBX starts ringing extension 1003 and emits the call event |
| ~50 ms | PBX connector receives it, normalises to `call.ringing` |
| ~60 ms | Caller ID resolver checks Redis cache for `+254712345678` |
| ≤ 300 ms | On a cache miss, LMS lookup (hard timeout 300 ms; on timeout we show the number and fill the name when the lookup completes) |
| ~350 ms | Event pushed to the agent's Hub tab; ringing card shows name, account, status, queue, wait time |
| same time | The browser SDK's `incoming` event fires on the WebRTC side; Hub plays the ring tone |

The ringing card and the SDK session are matched by the PBX's call ID. If the agent does not answer within the ring-group/queue ring time, the queue moves on to the next agent, and the call service records a "missed by 1003" event (reported as a missed call per agent).

**Who sees the ringing card:** only the agent whose extension rings. The card appears on any Hub page (it already does in the frontend).

### Step 6 — talking
- Recording starts the moment the call is answered and cannot be paused (07).
- The customer panel shows LMS data allowed for this agent's role (06, 09).
- The agent can hold, transfer, add people, park, send DTMF, take notes (07).

### Step 7 — after-call work (wrap-up)
- The agent has N seconds (setting) in "Wrap-up" status: the queue will not send them a new call.
- They pick an **outcome** (Resolved, Promise to pay, Payment made, Call back requested, Escalated, Fraud report, ...) and write notes. Outcomes are configurable per queue.
- "Promise to pay" opens two extra fields (amount, date) that are sent to the LMS through its API as a promise record (if the LMS supports it), so collections follow-up happens in the LMS.

### Step 8 — closing the record
- The PBX sends the CDR event; the call service matches it to the call row (by PBX call ID), and stores the official duration, disposition, trunk, and recording file name.
- The recording archiver downloads the file, encrypts it, stores it, and records its checksum (09).
- A nightly job compares the day's CDRs from the PBX with our call rows and repairs any gaps (missed events).

## 4.3 Edge cases and exactly what happens

| Situation | Behaviour |
|---|---|
| Caller withholds their number | Shown as "Withheld number". No LMS lookup. Agent identifies the caller by verification questions (06). |
| Number belongs to several LMS customers (shared phone) | Ringing card shows "2 customers use this number" and lists names; agent picks the right one after verifying. |
| Agent's browser closed / no internet | Extension is unregistered; the queue skips them. Their status shows "Offline". |
| Agent declines | Call returns to the queue (not hung up); logged as declined by 1003. |
| All agents busy for a long time | Wallboard alert (threshold in Settings); supervisor can log extra agents into the queue. |
| Call service down | **Calls still work** (the PBX routes and rings registered browsers). Hub shows the number without the LMS name and a banner "Customer details unavailable". Events are reconciled when the service is back. |
| LMS down or slow | Ringing card shows the number and "Customer lookup unavailable"; the agent can retry the lookup from the call panel. |
| Same customer calls 3 times in 10 minutes | Ringing card badge "Called 3 times today", to prioritise. |
| Call from a number on the fraud watch list (LMS flag) | Card shows a red "Fraud review" tag; account details hidden until verification. |
