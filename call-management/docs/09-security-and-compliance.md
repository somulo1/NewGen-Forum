# 9. Security and compliance

We handle customers' financial data and their voices. This chapter lists what the law and good practice require and exactly how the service meets each point. **Legal requirements must be confirmed by our compliance/legal team; this is an engineering checklist, not legal advice.**

## 9.1 Kenyan law and regulators to check against

| Source | Why it matters here |
|---|---|
| **Data Protection Act, 2019** and its 2021 regulations (Office of the Data Protection Commissioner, ODPC) | Call recordings, phone numbers and loan data are personal data; financial data is sensitive. Requires a lawful basis, purpose limitation, minimisation, retention limits, security, data subject rights (access, correction, deletion where lawful), breach notification to the ODPC within 72 hours, registration as a data controller/processor, and a Data Protection Impact Assessment (DPIA) for high-risk processing such as call recording and monitoring. |
| **Central Bank of Kenya** (digital credit providers regulations) | Rules on fair debt collection, customer communication and data sharing for licensed digital credit providers; complaints about harassing collection calls have led to sanctions in the sector. |
| **Communications Authority of Kenya** | Rules on the use of numbers and caller ID, and on unsolicited marketing messages (SMS sender IDs, opt-out). |
| **Consumer Protection Act, 2012** | Fair treatment, no harassment, accurate information. |

Actions: run a DPIA before go-live; update our privacy notice to cover call recording and monitoring; register the processing with the ODPC; sign data processing agreements with Yeastar (if cloud-hosted), the SMS/WhatsApp providers and the hosting provider.

## 9.2 Recording consent and transparency
- Inbound: the greeting says the call may be recorded and why ("for your security and to improve our service").
- Outbound: the agent script's first sentence includes it.
- Customers who object: the call continues only for what can be handled without recording, or they are offered another channel (branch, email). This is a policy decision for legal (11).

## 9.3 Access control (server-enforced)

| Right | Lets a person |
|---|---|
| `calls.use` | Use the softphone, see their own calls and customers they are on a call with |
| `calls.supervise` | See their teams' calls, wallboard, listen/whisper/barge |
| `calls.recordings.play` / `.download` | Play / download recordings (download needs a reason) |
| `calls.campaigns.manage` / `.run` | Create / run campaigns |
| `calls.dnc.manage` | Edit the do-not-call list |
| `calls.routing.manage` | Change hours, IVR, groups |
| `calls.reports` | See reports |
| `calls.audit` | See the compliance log |

Defaults: agents get `calls.use` only. Rights are granted through Hub's existing roles and rights system, under the "Calls management" widget.

## 9.4 Audit log (tamper-evident)
Logged: sign-ins to the softphone, every call placed/answered/transferred, verifications, number reveals, recording plays and downloads (with reason), supervision sessions, DNC changes, routing changes, campaign starts, exports, and refused actions.

Each entry stores the hash of the previous entry (`prev_hash`, `hash = SHA-256(prev_hash + entry)`), so deleting or editing a row breaks the chain. A daily job verifies the chain and alerts on a break. The table is append-only for the service's database user (no UPDATE/DELETE grant).

## 9.5 Encryption
| Data | In transit | At rest |
|---|---|---|
| Agent ↔ PBX audio | SRTP (DTLS-SRTP, WebRTC) | — |
| Agent ↔ PBX signalling | SIP over WSS (TLS 1.2+) | — |
| PBX ↔ carrier | SIP over TLS + SRTP **if the carrier supports it**; otherwise a private link/VPN to the carrier | — |
| Hub ↔ call service | HTTPS / WSS (TLS 1.2+) | — |
| Service ↔ LMS | HTTPS with mutual TLS | — |
| Recordings, voicemails | — | AES-256-GCM per file, keys in a KMS/vault, rotated yearly |
| Database | — | Encrypted disk + encrypted backups |

## 9.6 Number masking and minimisation
- Customer numbers are masked in the UI by default (`+254 ••• ••• 678`); "Show" reveals and is audited.
- Exports mask numbers unless the exporter has a specific right.
- The service stores a snapshot (name, account number) on the call, never the full LMS record; LMS data in Redis expires in 10 minutes.

## 9.7 Fraud and abuse controls
- **SIM-swap / account takeover:** phone number changes are never done on a call; they go through the LMS process. Calls from a number recently changed in the LMS show a warning.
- **Social engineering:** verification questions are random and logged; repeated failures raise a fraud alert.
- **Toll fraud** (someone using our PBX to make expensive international calls): international dialling disabled by default; PBX SIP ports reachable only from carrier IPs; strong extension passwords never given to users (signature login); alerts on unusual call volume or destinations at night.
- **Insider misuse:** agents only see customers they are calling or who call them; supervisors only their teams; all lookups and reveals audited; monthly review of top lookups per agent.

## 9.8 Data subject requests
A customer may ask for their call recordings or ask what we hold. The service provides an export for one customer (all calls, notes, messages, recordings) for the data protection officer, and supports deletion where no legal retention obligation applies (loan records and their recordings usually must be kept; legal decides).

## 9.9 Retention (defaults, to confirm with legal)
| Data | Default | Then |
|---|---|---|
| Call records, notes, outcomes, messages | 7 years | Deleted or anonymised |
| Recordings and voicemails | 1 year | Deleted unless on legal hold |
| Raw PBX events | 30 days | Deleted |
| LMS cache | 10 minutes | Expires |
| Audit log | 7 years | Archived to cold storage |
