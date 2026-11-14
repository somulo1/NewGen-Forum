# Call management service — documentation

This folder documents the **call management service** before it is built: what it does, how a phone call actually moves through the system, how each feature works, and how we will build, secure and run it.

Nothing in this folder is code. The service does not exist yet. The Hub frontend already has the Calls management screens (running on simulated data in `frontend/logic/pbx/`); this documentation describes the real backend those screens will talk to.

## Start here

| Read | If you want to know |
|---|---|
| **`call-management-handbook.pdf`** | Everything at a glance, with diagrams. Read this first. |
| [01-voip-from-scratch.md](01-voip-from-scratch.md) | How a voice call travels over the internet: SIP, RTP, codecs, WebRTC, trunks, NAT. No prior knowledge assumed. |
| [02-architecture.md](02-architecture.md) | The pieces of the system and how they connect: Hub, the microservice, the Yeastar PBX, the loan system, the carriers. |
| [03-yeastar-integration.md](03-yeastar-integration.md) | Exactly how we use Yeastar: the Open API, the event stream, the Linkus WebRTC SDK. What is confirmed and what we must still verify. |
| [04-inbound-calls.md](04-inbound-calls.md) | A customer calls us: every step from the carrier to the agent's headset, with timings. |
| [05-outbound-calls.md](05-outbound-calls.md) | We call a customer: click-to-call, campaigns, caller ID, do-not-call, calling hours. |
| [06-caller-identification.md](06-caller-identification.md) | How a ringing number becomes a name using the loan system (LMS), and what happens when the caller is not in the LMS. |
| [07-call-features.md](07-call-features.md) | Each in-call and centre feature in detail: hold, transfer, conference, park, recording, voicemail, supervision, messaging, wallboard, reports. |
| [08-data-and-api.md](08-data-and-api.md) | The service's database tables, the API Hub calls, and the live events it pushes. |
| [09-security-and-compliance.md](09-security-and-compliance.md) | Kenyan data protection, call recording consent, access rights, audit, retention, fraud. |
| [10-operations.md](10-operations.md) | Deployment, capacity, monitoring, failure handling, disaster recovery. |
| [11-rollout-and-open-questions.md](11-rollout-and-open-questions.md) | The build plan in phases, and the decisions still to make (including things not yet asked about). |

## The goal in one paragraph

Agents at our fintech call customers on their personal phones and take customer calls, from inside Hub, without a desk phone. Every call is tied to the right customer from the loan system the moment it rings, is recorded for audit, is routed to the right team, and is measured. Supervisors see queues and agents live and can listen in. Everything a customer says or that is said to them is traceable, and only the people the organization allows can use any of it.

## Words used in these documents

| Word | Meaning |
|---|---|
| **PBX** | Private Branch Exchange: the phone system that owns our numbers, extensions, queues and routing. Ours is a **Yeastar P-Series** PBX. |
| **Extension** | An internal number for one agent (e.g. 1001). An agent's browser registers to it. |
| **Trunk** | The connection between our PBX and a phone carrier (Safaricom, Airtel, a SIP provider). Calls to and from real phone numbers go over trunks. |
| **DID** | Direct Inward Dialling number: a public number customers dial (e.g. 0709 000 200) that the carrier delivers on a trunk. |
| **SIP** | The signalling language phones and PBXs use to set up, change and end calls. |
| **RTP** | The stream that carries the actual voice, in small packets, while a call is up. |
| **WebRTC** | The browser technology that lets Hub be a phone (microphone, speaker, encryption) without a plugin. |
| **IVR** | Interactive Voice Response: the "press 1 for loans" menu. |
| **Queue** | A line of waiting callers served by a group of agents. |
| **CDR** | Call Detail Record: the PBX's record of one call (who, when, how long, result). |
| **LMS** | Our Loan Management System: the source of truth for customers, their numbers and accounts. |
| **Call service** | The microservice described here (`callsvc`). |
