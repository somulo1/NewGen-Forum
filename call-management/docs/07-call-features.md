# 7. Call features in detail

For each feature: what the user sees, what actually happens underneath, and the rules.

## 7.1 Mute
- **User:** Mute button; the REC chip and timer keep running.
- **Underneath:** the browser stops sending microphone audio (SDK `mute`). No SIP message; the PBX still records the customer's side.
- **Rule:** muting does not pause recording.

## 7.2 Hold
- **User:** Hold → the customer hears the care tune; the agent sees "On hold · customer hears Calm care" and can press **Hear tune**.
- **Underneath:** SDK `hold` sends a re-INVITE with `a=sendonly`; the PBX plays its **music-on-hold** class to the customer. The PBX emits a hold event; the service records hold start/stop.
- **Rules:** hold time is measured per call (reports: average hold time). After 60 s on hold the agent gets a reminder; after 3 minutes the supervisor sees it on the wallboard.

## 7.3 Keypad (DTMF)
- Sends RFC 4733 events through the PBX (SDK `dtmf`). Used when calling another IVR (bank, M-Pesa support).
- Digits sent are stored on the call **only if** the target is not a customer (customers might type PINs if confused; we never store digits sent on customer calls).

## 7.4 Transfer
| Type | What the agent does | What happens |
|---|---|---|
| **Blind** | Transfer now → enters 1004 / a queue / an external number | SDK `blindTransfer`; the PBX sends the customer to the target. The agent's leg ends. If the target does not answer, the call goes to the target's voicemail or back to the queue (PBX setting). |
| **Attended** | Talk first → target answers → agent briefs them → Complete transfer | SDK `call(target, …, transferId)` creates a consult leg while the customer is on hold; `attendedTransfer` completes it. "Back to the customer" hangs up the consult leg and resumes. |

- Transfers to external numbers are allowed only to an allow-list (e.g. partner bank lines), set by an admin.
- The customer card, verification state and notes **travel with the call**: the receiving agent sees "Transferred from Daisy (1004) · verified ✔ · note: wants to restructure".
- One call record spans the transfer; each agent's talk time is stored separately.

## 7.5 Add someone (conference)
- Up to 5 parties in total. The PBX mixes the audio (a conference bridge).
- Use cases: a supervisor or a colleague from loans joins; an interpreter.
- The customer hears a short join tone. Everyone added is listed on the call and in the record.

## 7.6 Park
- Puts the call on a numbered parking slot (e.g. 701); anyone in the team can pick it up from the "Parked" chip or by dialling the slot.
- If nobody picks up within 2 minutes, it rings back the agent who parked it.

## 7.7 Recording (mandatory)
- **Every call is recorded from answer to hang-up and cannot be paused or stopped.** Recording is configured on the PBX per trunk/queue/extension as "always"; the Hub UI has no pause control.
- The customer hears the notice in the greeting ("this call may be recorded"); for outbound calls the agent states it in the opening line, which is in the call script.
- **Archive flow:** call ends → CDR event names the file → archiver downloads it from the PBX → computes SHA-256 → encrypts with the service's key (AES-256-GCM, keys in a KMS/vault) → uploads to object storage → deletes it from the PBX disk after verifying the upload → stores metadata (call ID, duration, checksum, storage key, retention date).
- **Playback:** only with `calls.recordings.play`; streamed through the service (never a public link); every play and download is written to the audit log with who, when, why (a reason field is required for downloads).
- **Retention:** default 1 year (setting), then deleted automatically unless under legal hold (a dispute, complaint or court order). Retention is a compliance decision (09, 11).

## 7.8 Voicemail
- Boxes: one per agent extension and one per queue (Customer care, Collections, …).
- The PBX records the message and emits an event; the archiver stores it like a recording.
- **Transcription:** optional speech-to-text (phase 3) so agents can read messages; transcripts are stored with the message.
- A queue voicemail is a **task**: assigned to the queue, someone must call back and mark it handled. Unhandled for 2 working hours → wallboard alert.

## 7.9 Agent status and queue membership
| Status | Queue sends calls? | Set by |
|---|---|---|
| Available | Yes | Agent |
| On a call / Ringing | No | PBX events |
| Wrap-up | No (for N seconds) | System after each call |
| Away, Lunch, Busy | No (queue pause with reason) | Agent |
| Do not disturb | No, and direct calls go to voicemail | Agent |
| Offline | No (browser not registered) | System |
| Logged out of queue X | Not for that queue | Agent switch on Monitoring, or supervisor |

Every status change is stored with time, so reports can show time spent in each status (occupancy, adherence).

## 7.10 Supervision (listen, whisper, barge)
| Mode | Supervisor hears | Agent hears supervisor | Customer hears supervisor |
|---|---|---|---|
| Listen | Both | No | No |
| Whisper | Both | Yes | No |
| Barge in | Both | Yes | Yes |

- Requires `calls.supervise`, and only for agents in the supervisor's teams/queues.
- Started from Monitoring; underneath, the Open API's monitoring function (verify name in 03) bridges the supervisor's extension into the call.
- Every session is logged (who, which call, mode, start, end) in the compliance log.

## 7.11 Messaging (SMS and WhatsApp)
- **SMS:** through an aggregator (e.g. Africa's Talking) with our registered **sender ID** (alphanumeric, e.g. "OURFINTECH") approved per network. Delivery reports update the message status (Sent → Delivered / Failed). Replies come in through a shortcode or long number and land in the customer's thread.
- **WhatsApp:** through the WhatsApp Business Platform (Meta, or a provider). Rules: we can only start a conversation with an **approved template**; free-form replies are allowed for 24 hours after the customer's last message. Templates (payment reminder, KYC request) are approved by Meta in advance.
- Messages are stored per customer thread with who sent them; templates are managed by admins.
- Opt-out words ("STOP") add the number to the SMS opt-out list automatically.

## 7.12 Wallboard and alerts (Monitoring)
Live, pushed every second from Redis: per queue waiting callers, longest wait, answered/abandoned today, service level (% answered within the target, e.g. 20 s), average wait and talk; per agent status, time in status, calls today, talk time.

Alerts (thresholds in Settings): longest wait over X s; service level under Y%; agent in wrap-up/away longer than Z min; trunk channels over 80% used; an agent's line quality poor for 2 minutes. Alerts can be dismissed; they come back if the problem clears and recurs.

## 7.13 Reports
All from the `calls` table (and status history), never from the PBX directly, so they cover transfers, outcomes and LMS data:

| Report | Key figures |
|---|---|
| Calls by day/hour | inbound, outbound, answered, abandoned, missed |
| Queue performance | service level, average speed of answer, abandon rate, longest wait |
| Agent performance | calls, talk time, average handle time (talk + hold + wrap-up), missed, outcomes |
| Outcomes | count per disposition; promise-to-pay amount and kept rate (joined from the LMS) |
| Campaigns | reached %, tries per reach, outcomes |
| First contact resolution | % of customers who did not call again within 7 days on the same issue |

Exports to CSV and scheduled emailed reports (phase 3).
