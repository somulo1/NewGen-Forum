# 3. Yeastar integration

How the call service and Hub use the Yeastar P-Series PBX.

## 3.1 What Yeastar gives us, and what is actually open source

| Tool | What it is | Licence | Used for |
|---|---|---|---|
| **P-Series PBX** (Appliance, Software or Cloud edition) | The phone system | Commercial | Everything telephony |
| **P-Series Open API** | REST API + WebSocket event stream on the PBX | Commercial feature, documented by Yeastar (must be enabled, may need a plan/licence) | Server-side control and events by the call service |
| **ys-webrtc-sdk-core** (github.com/Yeastar-PBX/ys-webrtc-sdk-core) | Browser WebRTC softphone toolkit | Public on GitHub, **no open-source licence stated** ("Copyright Xiamen Yeastar") | The softphone inside Hub |
| **ys-webrtc-sdk-ui** | Ready-made UI on top of the core SDK | Public on GitHub, no licence stated | Not used (Hub has its own UI) |
| **linkus-android-sdk** | Native Android softphone SDK | **MIT** | Only if we build a mobile app later |
| **linkus-iOS-MacOS-sdk** | Native iOS/macOS SDK | **MIT** | Only if we build a mobile app later |

**Important correction to "already open source":** only the two native mobile SDKs carry an open-source licence. The browser SDK we need is publicly published but its licence is not stated, and the PBX and its Open API are commercial. Before production we must get written confirmation from Yeastar (or our reseller) that we may use `ys-webrtc-sdk-core` in Hub, and which licence/plan enables the Open API and the WebRTC SDK on our PBX. Yeastar's Android SDK README states a prerequisite of a cooperation agreement with Yeastar; assume the same for the web SDK.

## 3.2 Two ways to control calls, and which we use for what

| | Browser SDK (in Hub) | Open API (in call service) |
|---|---|---|
| Runs in | The agent's browser | Our server |
| Controls | Only that agent's own calls | Any call on the PBX |
| Carries audio | **Yes** (it is the phone) | No |
| Good for | Answer, hang up, hold, mute, DTMF, blind & attended transfer, record start/pause | Click-to-call that rings the agent first, campaigns, supervisor listen/whisper/barge, park, queue agent login/pause, CDR, recordings download, configuration |

So: **the browser does what the agent does to their own call; the service does everything else** and is the single source of truth for state, because it sees the PBX's events for all calls.

## 3.3 Browser SDK: confirmed functions

From Yeastar's public documentation in `ys-webrtc-sdk-core` (`docs/Phone.md`, `docs/PBX.md`, `README.md`):

**Initialise / register**
```js
init({
  username: '1003',               // extension number
  secret: '<login signature>',    // created server-side, see below
  pbxURL: 'https://pbx.ourcompany.co.ke:8088',
  userAgent: 'WebClient',         // or 'WebPC'
  deviceIds: { cameraId, microphoneId },  // our chosen mic
  disableCallWaiting: false,
  enableLog: false,
});
```
The **login signature** is created by the server through the Open API endpoint **`/openapi/v1.0/sign/create`** (documented in `docs/CreateSign.md`). Our call service calls it and hands the signature to the agent's browser. The extension's SIP password never reaches the browser.

**Phone operator**
| Function | Use |
|---|---|
| `call(number, option?, transferId?)` | Dial out; with `transferId` it is the consult leg of an attended transfer |
| `answer(callId)`, `reject(callId)`, `hangup(callId)`, `terminate(callId, type)` | Answer / decline / end |
| `hold(callId)`, `unhold(callId)` | Hold |
| `mute(callId)`, `unmute(callId)` | Microphone |
| `dtmf(callId, digits)` | Send keys `0-9 * #` |
| `blindTransfer(callId, number)`, `attendedTransfer(callId, number)` | Transfers |
| `startRecord(callId)`, `pauseRecord(callId)` | Recording (we will **not** expose pause: all calls are recorded, see 07) |
| `setSessionStaticStatus(callId, {name, avatar, company})` | Show our LMS name on the SDK session |
| `isMaxCall` | The SDK allows **at most 2 concurrent calls** per browser session |

Events: `registered`, `registrationFailed`, `incoming` ({callId, session}), `newRTCSession`, `startSession`, `deleteSession`, `connected`, `disconnected`. PBX operator events: `cdrChange` ("query CDR again") and `runtimeError` with codes `-106 LINKUS_DISABLED`, `-107 LOGGED_IN_ELSEWHERE`, `-108 EXTENSION_DELETED`, `-109 RE_LOGIN`, `-110 SDK_PLAN_DISABLED`. Hub must handle each runtime error with a clear message (e.g. -107: "You signed in to your phone on another device").

The 2-call limit means: one active call + one on hold (or one consult leg). Hub's UI must not offer a third.

## 3.4 Open API: what we need from it

Yeastar's Open API documentation (help.yeastar.com, developer guide for P-Series) could not be retrieved automatically while writing this document. The capabilities below are what the design needs; **each line must be matched to its exact endpoint, fields and event ID from Yeastar's developer guide for our PBX edition and firmware before coding.** This verification is task 1 of phase 1 ([11-rollout-and-open-questions.md](11-rollout-and-open-questions.md)).

| Need | Expected Open API capability (verify) | Used by |
|---|---|---|
| Authenticate | Obtain an access token with the API client's credentials; refresh before expiry; revoke on shutdown. Typical pattern: short-lived access token (~30 min) + longer refresh token. API access is restricted to allow-listed IPs. | PBX connector token manager |
| Event stream | A WebSocket on the PBX; after connecting, send a subscription message listing event topics; send heartbeats; resubscribe after reconnect. | Event consumer |
| Call events | Call state changed (ringing, answered, hold, hang-up) with call/channel IDs, caller, callee, extension, trunk, queue | Live state, ringing card, wallboard |
| CDR event | New call detail record after each call | Joining the official record to our call row |
| Presence | Extension status changed (idle, ringing, busy, DND) | Agent status on wallboard |
| Queue | Agent login/logout/pause per queue; queue status (waiting callers, longest wait) | Queue switches, wallboard |
| Dial | Make a call between an extension and a number (ring agent, then dial customer) | Click-to-call, campaigns |
| Call control | Hang up, hold, transfer (blind/attended), park, query active calls | Supervisor and server-side actions |
| Monitoring | Listen / whisper / barge on an extension's call | Supervisor tools |
| Recordings | List and download recording files by call | Recording archiver |
| CDR search | Query CDR by time range, paginated | Nightly reconciliation, reports backfill |
| Voicemail | List and download voicemails per extension/queue box | Voicemail tab |
| Config | Read extensions, queues, ring groups, IVRs, inbound routes, time conditions | Routing tab (read first; write later, phase 4) |

If a capability is missing from the Open API on our firmware (for example writing IVR menus), the Routing tab shows that part read-only with a link to configure it on the PBX web portal. We do not scrape the PBX admin UI.

## 3.5 The PBX connector, in detail

```
              ┌─────────────── PBX connector ────────────────┐
 config ────► │ TokenManager   → keeps a valid access token  │
              │ HTTPClient     → retries, timeouts, rate cap │──► PBX Open API (REST)
              │ EventSocket    → connect, subscribe, ping    │◄── PBX event WebSocket
              │ EventNormaliser→ PBX JSON → our CallEvent    │──► internal event bus (Redis Streams)
              └──────────────────────────────────────────────┘
```

**TokenManager**
- Gets a token at start-up; refreshes at 80% of its lifetime; on a `401` refreshes once and retries the request once.
- A single instance holds the token (leader lock in Redis), so several service replicas do not fight over refreshes.

**HTTPClient**
- Timeouts: 3 s for call control, 30 s for downloads.
- Retries: only idempotent reads, max 2, with jitter. **Never retry "dial"** (it could call a customer twice); instead, query active calls to see if the first attempt worked.
- Rate cap: respect the PBX's documented API rate limit; queue excess requests.

**EventSocket**
- Connects on start-up; subscribes to the topics in 3.4.
- Heartbeat at the interval the PBX requires; if no message or pong for 2 intervals, reconnect.
- Reconnect with backoff 1 s, 2 s, 5 s, 10 s (cap). After reconnecting: resubscribe, then **reconcile** by querying active calls, so calls that changed while disconnected are corrected.
- Every raw event is appended to `pbx_events_raw` before processing (replayable).

**EventNormaliser**
Turns PBX-specific payloads into our own event types so the rest of the service is PBX-independent:

| Our event | Meaning |
|---|---|
| `call.created` | A new call leg exists (inbound from trunk, or outbound from extension) |
| `call.ringing` | Ringing an extension (who, which queue, how long the caller waited) |
| `call.answered` | Connected; talk time starts |
| `call.held` / `call.resumed` | Hold changes |
| `call.transferring` / `call.transferred` | Transfer started / finished (with target) |
| `call.joined` | A third party (conference, supervisor) joined |
| `call.ended` | Hung up, with who hung up and the reason |
| `call.cdr` | Official PBX record arrived (duration, disposition, recording file) |
| `agent.status` | Extension presence changed |
| `queue.status` | Waiting count / longest wait changed |
