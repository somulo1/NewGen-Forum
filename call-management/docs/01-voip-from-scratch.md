# 1. VoIP from scratch

How a phone call travels over a data network. Read this before the other documents; everything later builds on these terms.

## 1.1 The two halves of every call

Every voice-over-IP call has two completely separate parts:

| Part | Protocol | What it carries | When |
|---|---|---|---|
| **Signalling** | SIP (Session Initiation Protocol) | "I want to call 0712 345 678", "it is ringing", "answered", "put on hold", "hang up" | Before, during (for changes) and at the end of the call |
| **Media** | RTP (Real-time Transport Protocol), encrypted as SRTP | The actual sound, cut into packets of 20 ms | Only while the call is up |

Signalling is like the envelope and the instructions; media is the conversation. They often take **different network paths** and use **different ports**, which is why "the call connects but nobody can hear anything" is the classic VoIP problem: signalling worked, media was blocked.

## 1.2 SIP: how a call is set up

SIP is text, similar to HTTP. The main messages ("methods") and replies:

| Message | Meaning |
|---|---|
| `REGISTER` | A phone (or Hub in a browser) tells the PBX "extension 1001 is reachable at this address". Repeated every few minutes. |
| `INVITE` | Start a call. Carries an SDP body describing what audio the caller can do (codecs, IP, port). |
| `100 Trying` | "Got it, working on it." Stops the caller re-sending. |
| `180 Ringing` | The far phone is ringing. The caller hears ringback. |
| `183 Session Progress` | Early media: the far side is sending sound before answering (e.g. a carrier announcement, or our hold tune). |
| `200 OK` | Answered. Carries the answering side's SDP. |
| `ACK` | Caller confirms. From here, RTP flows both ways. |
| `re-INVITE` / `UPDATE` | Change an ongoing call: hold (`a=sendonly`), resume, switch codec. |
| `REFER` | Transfer: "please call this other party instead". |
| `INFO` | Mid-call information (sometimes DTMF). |
| `BYE` | Hang up. Answered with `200 OK`. |
| `CANCEL` | Hang up before the other side answered. |
| `4xx/5xx/6xx` | Failures: `404` number not found, `480` unavailable, `486 Busy Here`, `487` cancelled, `603 Decline`. |

### A normal outbound call to a customer, message by message

```
Agent browser        Yeastar PBX            Carrier (SIP trunk)        Customer phone
     |--INVITE (1001→0712..)-->|                     |                        |
     |<---100 Trying-----------|--INVITE (+254712..)->|                        |
     |                         |<--100 Trying---------|--(mobile network)----->|
     |                         |<--180 Ringing--------|<-------ringing---------|
     |<---180 Ringing----------|                      |                        |
     |                         |<--200 OK (SDP)-------|<-------answered--------|
     |<---200 OK (SDP)---------|---ACK--------------->|                        |
     |---ACK------------------>|                      |                        |
     |<=====SRTP voice========>|<======RTP voice=====>|<=====voice============>|
     |---BYE------------------>|---BYE--------------->|                        |
     |<---200 OK---------------|<--200 OK-------------|                        |
```

The PBX sits in the middle of **both** legs. It is a "back-to-back user agent" (B2BUA): the agent's call and the customer's call are two separate SIP dialogs that the PBX joins. That is what lets the PBX record, transfer, put on hold and monitor: it owns both halves.

## 1.3 SDP: agreeing how to talk

Inside `INVITE` and `200 OK` is an SDP (Session Description Protocol) body. The key lines:

```
c=IN IP4 41.90.12.34          ← where to send my audio
m=audio 18234 RTP/SAVPF 111 8 0 101   ← my port, and codecs I accept, in order of preference
a=rtpmap:111 opus/48000/2
a=rtpmap:8 PCMA/8000           ← G.711 a-law
a=rtpmap:101 telephone-event/8000   ← DTMF keys as events
a=sendrecv                     ← (sendonly = I put you on hold)
```

Both sides pick the first codec they have in common.

## 1.4 Codecs: how voice becomes bytes

| Codec | Bandwidth per direction (with packet overhead) | Where we use it | Notes |
|---|---|---|---|
| **G.711 a-law (PCMA)** | ~87 kbps | PBX ↔ Kenyan carriers | The international standard outside North America. No compression, best compatibility, what carriers expect. |
| **G.711 µ-law (PCMU)** | ~87 kbps | Only if a carrier asks | North American variant. |
| **Opus** | 24–40 kbps typical | Browser (WebRTC) ↔ PBX | Modern, adapts to bad networks, wideband (sounds clearer). Mandatory in WebRTC. |
| **G.729** | ~31 kbps | Only on very thin links | Compressed, slightly robotic, may need a licence. |

When the browser speaks Opus and the carrier speaks G.711, the PBX **transcodes** between them. Transcoding uses PBX CPU; it is a sizing factor (see [10-operations.md](10-operations.md)).

**Rule of thumb for bandwidth:** plan **100 kbps up and down per concurrent call** at the PBX and per agent. Fifty agents on calls at once need ~5 Mbps each way at the office, with headroom ×2.

## 1.5 DTMF: the keys pressed during a call

When a caller presses "1" in our IVR, the key travels as **RFC 4733 telephone-events** (`telephone-event` in SDP), not as a beep in the audio. The PBX reads the event. Agents sending keys to another system (e.g. a bank IVR) do the same. We configure the trunk and the PBX to use RFC 4733 and keep "in-band" detection only as a fallback, because compressed audio garbles in-band tones.

## 1.6 WebRTC: Hub as the phone

Agents do not use desk phones. Hub runs in the browser, and the browser becomes the phone through **WebRTC**:

1. Hub asks the call service for a **login signature** for the agent's extension (the service gets it from the PBX; the agent never sees the extension password).
2. The **Yeastar WebRTC SDK** (`ys-webrtc-sdk-core`) in the page uses it to **register** the extension over **SIP over secure WebSocket (WSS)** to the PBX, e.g. `wss://pbx.ourcompany.co.ke:8088`.
3. Calls are set up with SIP over that WebSocket; audio flows as **SRTP** (encrypted RTP) using **DTLS** keys, peer-to-peer between browser and PBX.
4. The browser asks for microphone permission once (Hub already does this for its own calls).

WebRTC media is **always encrypted**. That is required by the standard, not an option.

## 1.7 NAT, STUN, TURN: why audio sometimes doesn't flow

Agents sit behind office routers or home Wi-Fi (NAT): their real IP is private (192.168.x.x). The SDP must contain an address the PBX can actually reach.

| Tool | What it does | Do we need it |
|---|---|---|
| **ICE** | The browser collects every possible address (local, public, relayed) and tests which works. | Always on in WebRTC. |
| **STUN** | A server that tells the browser its public address. | Yes, cheap. The PBX or a public STUN server. |
| **TURN** | A relay that forwards the audio when no direct path works (strict corporate firewalls, some mobile networks). | Yes for remote agents. We run our own **coturn** server. |

**Ports to open** (exact values are set on the PBX and confirmed at installation):

| Purpose | Port(s) | Direction |
|---|---|---|
| Hub/agents → PBX WebRTC signalling (WSS) | TCP 8088 (Yeastar web/WebRTC port, configurable) | Inbound to PBX |
| PBX ↔ carrier SIP signalling | UDP/TCP 5060, TLS 5061 | Between PBX and carrier IPs only |
| RTP/SRTP media | UDP 10000–12000 (PBX RTP range, configurable) | Both ways |
| TURN | UDP/TCP 3478, TLS 5349, relay UDP 49152–65535 | Inbound to TURN server |
| Call service → PBX Open API | TCP 443/8088 HTTPS + WSS | From the service only (IP allowlisted) |

## 1.8 Call quality: what "good" means

| Measure | Good | Poor | What causes it |
|---|---|---|---|
| **Latency** (one way) | < 150 ms | > 300 ms (people talk over each other) | Distance, Wi-Fi, VPNs, TURN relays far away |
| **Jitter** | < 30 ms | > 50 ms (choppy) | Busy networks; the jitter buffer smooths it at the cost of delay |
| **Packet loss** | < 1% | > 3% (words missing) | Wi-Fi, mobile data, congested links |
| **MOS** (Mean Opinion Score, 1–5) | ≥ 4.0 | < 3.5 | Combination of the above |

The browser reports these through WebRTC statistics every few seconds. The call service stores them per call, Hub shows the "Clear / Fair / Poor line" chip, and supervisors get an alert when an agent's line stays poor.

## 1.9 Trunks: reaching real phones in Kenya

The PBX reaches the public phone network through **SIP trunks** from a licensed carrier (Safaricom, Airtel, Jamii/Telkom, or a SIP provider). Each trunk gives:

- **DIDs**: our public numbers (e.g. 0709 000 200 Customer care, 0709 000 100 Collections). Inbound calls arrive on the trunk addressed to a DID.
- **Channels**: how many calls can be up at once (e.g. 30). Every inbound or outbound customer call uses one channel. Running out means callers hear busy, so trunk channels are a key capacity number.
- **Caller ID presentation**: which of our numbers the customer sees when we call. The carrier only lets us present numbers we own.

A second trunk from a **different carrier** is strongly recommended so that one carrier's outage does not stop all calls (see [10-operations.md](10-operations.md)).

Numbers are always stored in **E.164** format: `+254712345678`. Inputs like `0712 345 678`, `254712345678` or `712345678` are normalised to E.164 before any lookup, storage or comparison.
