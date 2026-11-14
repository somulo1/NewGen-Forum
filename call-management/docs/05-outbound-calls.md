# 5. Outbound calls: we call a customer

## 5.1 Three ways an outbound call starts

| Way | Who starts it | How the PBX is asked |
|---|---|---|
| **Manual dial / click-to-call** | Agent types a number or presses Call on a customer, history row, voicemail or call-back | Browser SDK `call(number)` from the agent's registered extension |
| **Call-back task** | System assigns a waiting call-back request to a free agent | Same as manual, started when the agent accepts the task |
| **Campaign** | Agent presses "Call the next customer" (preview), or the dialer starts calls automatically (progressive) | Preview: browser SDK. Progressive: Open API "dial" that rings the agent first, then the customer |

## 5.2 Checks before any number is dialled (in this order)

Every outbound attempt goes through the call service first (`POST /v1/calls` with the number and the reason). The service refuses with a clear message if any check fails, and logs the refusal:

1. **Right:** the agent has `calls.use` (and `calls.campaigns.run` for campaigns).
2. **Number valid:** normalised to E.164; Kenyan mobile/landline pattern or an allowed international prefix. International calling is off unless the admin enables it per queue.
3. **Do not call:** the number is not on our DNC list, and the LMS customer does not have a "do not contact" flag.
4. **Calling hours:** for collections calls, only within allowed hours (default Mon–Sat 08:00–18:00 EAT, configurable; no calls on Sundays/public holidays). Personal calls back requested by the customer at a specific time are allowed at that time.
5. **Contact frequency cap:** a debt-collection contact limit per customer (e.g. max 3 attempts per day, 1 connected call per day). Fintech debt collection in Kenya is under regulatory and public scrutiny; harassment by repeated calls is a known complaint. Exact limits are a policy decision for compliance (see 11).
6. **Not already in a call with this customer** (another agent): prevents two agents calling the same customer at once.
7. **Caller ID allowed:** the selected "Call out as" number belongs to us and is allowed for this queue.

Only then does Hub's softphone (or the dialer) place the call.

## 5.3 The call itself

```
Hub (agent 1001)                     Call service                    PBX                    Carrier → customer
 │ POST /v1/calls {to, reason, as}──►│ checks 1–7, creates call row      │                      │
 │◄──────── 201 {call_id, allowed} ──│ (status "dialling")               │                      │
 │ SDK call("+254712345678") ───────────────────────────────────────────►│ outbound route:      │
 │                                    │                                  │ caller ID = chosen DID│
 │                                    │◄──── call.created / ringing ─────│──INVITE────────────►│
 │◄──── live: ringing (ringback) ─────│                                  │◄──180 Ringing───────│
 │                                    │◄──── call.answered ──────────────│◄──200 OK────────────│
 │◄──── live: talking, timer starts ──│  recording starts                │                      │
```

- The **outbound route** on the PBX picks the trunk and sets the presented caller ID from the selected "Call out as" number (Customer care, Collections, ...).
- **Ringback** (what the agent hears while it rings) comes from the carrier as early media; Hub also plays its own ringback tone if no early media arrives.
- Results recorded on the call row: `answered`, `no_answer` (rang out, usually 30–45 s), `busy`, `rejected`, `unreachable` (phone off / out of coverage), `failed` (network or trunk error), `cancelled` (agent hung up first).

## 5.4 Campaigns

A campaign is a list of customers to call for one purpose (arrears follow-up, KYC reminders, onboarding welcome).

**Creating:** select customers from the LMS by filters (e.g. product = personal loan, days in arrears 7–30, branch), or upload a list. The call service snapshots the list (customer ID + number) at creation.

**Rules per campaign:** queue/team, caller ID, calling window (e.g. 09:00–17:00), maximum tries per customer, minimum gap between tries, and which outcomes end the customer's attempts (Promise to pay, Payment made, Wrong number, Do not call).

**Dialling modes**

| Mode | How it works | When to use |
|---|---|---|
| **Preview** | Agent sees the next customer's details, then presses Call. | Collections (agent prepares before calling). Our default. |
| **Progressive** | When an agent becomes free, the dialer immediately calls the next customer; the agent is connected when the customer answers. One call per free agent. | High-volume reminders. |
| ~~Predictive~~ | Dials more customers than free agents and drops extras. | **Not used.** It creates "silent calls" (customer answers, nobody is there), which annoys customers and harms trust. |

**Progressive dialling, exactly:**
1. The dialer picks the next eligible row (not reached, tries left, gap passed, inside the window, all checks in 5.2 pass).
2. It locks the row (so two dialers never take the same customer) and calls the Open API "dial": ring the agent's extension first (agent hears a short beep and the customer details appear), then the PBX calls the customer.
3. Result is written to the row; if not reached, the next try is scheduled after the gap.
4. When all rows are reached or out of tries, the campaign is "Finished".

## 5.5 Call-backs

A call-back request (from the queue "press 1" or from a missed/abandoned call) becomes a task:

- Assigned to the queue; shown to free agents in that queue; the first to accept owns it.
- Due time: as soon as possible, or a time the customer chose ("call me after 2 pm").
- Tries: up to 3, 15 minutes apart. After the last failed try, an SMS goes out: "We tried to call you back. Call 0709 000 200 any time."

## 5.6 What the customer sees

- Our number as caller ID (never a hidden number: customers ignore withheld calls, and some regulators consider it misleading).
- Optionally (phase 3), a **branded caller name** if the carrier supports it.
- For campaigns, an SMS can be sent before the call ("We will call you today about your loan from 0709 000 100"), which raises answer rates.
