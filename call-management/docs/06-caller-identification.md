# 6. Caller identification with the loan system (LMS)

How a ringing number becomes "Grace Wanjiku · LN-20431", and what we do when the number is not in the LMS.

## 6.1 The principle

**A phone number identifies a phone, not a person.** Numbers get recycled by carriers, shared in families, lent, ported, or spoofed. So the system does two separate things:

1. **Identify (suggest):** from the number, suggest who is probably calling. Fast and automatic.
2. **Verify (confirm):** before discussing any account detail, the agent confirms the caller's identity with questions. Only a verified call unlocks sensitive details.

## 6.2 The lookup, step by step

```
Number from PBX event: "0712 345 678" / "254712345678" / "+254712345678"
 │
 ▼ 1. Normalise to E.164 → +254712345678   (reject garbage, mark "Withheld" if empty/anonymous)
 ▼ 2. Redis cache  key lms:phone:+254712345678  (TTL 10 min, holds matches or "none")
 │       hit → done
 ▼ 3. LMS API  GET /customers?phone=+254712345678   (timeout 300 ms, 1 retry only in the background)
 │       returns 0, 1 or many customers
 ▼ 4. If 0 matches → secondary sources (6.4)
 ▼ 5. Build the caller card (only fields allowed for this agent's role) → push to Hub
 ▼ 6. Snapshot on the call row: customer_id(s), name, account number, match source
```

### What we need from the LMS API

The LMS team must provide (or we agree on) a read-only API with:

| Endpoint | Purpose | Response fields we need |
|---|---|---|
| `GET /customers?phone={e164}` | Who uses this number (primary, alternate, next-of-kin numbers flagged separately) | `customer_id`, `full_name`, `phone_type` (primary/alternate/next_of_kin/employer), `status` (active/closed/blacklisted), `do_not_contact`, `fraud_flag`, `kyc_status` |
| `GET /customers/{id}/summary` | The card shown to the agent | accounts: `account_no`, `product`, `balance`, `arrears_days`, `next_due_date`, `next_due_amount`, last payment date and amount, assigned collector/branch |
| `GET /customers/{id}/verification` | Data for identity questions (never shown in full) | date of birth, ID number (agent sees only last 3 digits to compare), registered branch, last payment amount |
| `POST /customers/{id}/interactions` | Write the call outcome back (promise to pay, notes) | — |
| `GET /customers/search?q=` | Agent searches by name, ID number, account number | list of matches |

Requirements on that API: HTTPS with mutual TLS or an API key restricted to the call service's IPs; p95 latency under 200 ms for phone lookup (an index on normalised phone numbers in the LMS database); numbers stored in the LMS normalised to E.164 (the LMS team may need a one-off clean-up: our analysis job lists unparseable numbers).

### Why a cache, and why only 10 minutes
Repeat callers and campaign calls hit the same numbers; caching avoids loading the LMS. Ten minutes keeps data fresh enough (a payment made by M-Pesa reflects within minutes) and limits how long personal data sits outside the LMS. The cache holds only the lookup result (IDs + name + flags), not balances; the summary is fetched when the card opens.

## 6.3 One number, several customers

When a number matches more than one customer (shared family phone, a guarantor's number, a recycled number), the card shows:

> **This number is used by 2 customers.** Grace Wanjiku (primary number) · Peter Wanjiku (next of kin). Verify who is calling.

No account details show until the agent picks one and verification passes. Matches are ordered: primary number first, then alternate, then next-of-kin/employer.

## 6.4 Not found in the LMS: identifying everyone else

A caller missing from the LMS is not necessarily a stranger. The resolver checks, in order:

| # | Source | What it finds | Card shows |
|---|---|---|---|
| 1 | **Hub staff directory** | A colleague calling from their own mobile | "Staff: Kevin Odenda (Collections)" |
| 2 | **Our own call history** | The number has called or been called before, with a name the agent saved | "Called before (3 times) · saved as 'Mary – loan applicant'" |
| 3 | **Unidentified-callers register** (ours) | Numbers agents have captured details for but who are not LMS customers yet | "Prospect: Mary Atieno, asked about personal loans" |
| 4 | **LMS loan applications / leads** (if the LMS exposes them) | People who applied but are not customers | "Applicant · application APP-8812 (pending)" |
| 5 | **Partner lists** (agents, merchants, guarantors) if available | Business contacts | "Merchant: Kamau Hardware" |
| — | Nothing matched | — | "Unknown caller · first call from this number" |

### Capturing an unknown caller
During or after the call the agent fills a short form: name, reason for calling (new loan enquiry, wrong number, someone else's account, complaint, other), and optionally the LMS customer they turned out to be (found by search on ID or account number after verification).

- If they turned out to be an existing customer calling from a **new number**, the agent cannot change the LMS phone directly. The call service files a **"number change request"** in the LMS (or a task for the back office), because phone changes are a classic fraud route (SIM swap, account takeover) and must follow the LMS's own verification process.
- Otherwise the number is saved in the **unidentified-callers register** so the next call shows the captured name (source 3).

### Avoiding confusion (the goal you described)
- The card always says **where the name came from** ("from LMS", "saved by an agent on 3 Nov", "staff directory"). A name typed by an agent is never shown as if it came from the LMS.
- A name from history or the register is shown in grey with "unverified".
- If the LMS and our register disagree for the same number, the LMS wins and the register entry is flagged for review.

## 6.5 Verification before sensitive details

Before showing balances, arrears or transaction history, the agent asks **two** questions from the verification set, chosen at random by the system (e.g. "your date of birth", "last 3 digits of your ID number", "amount of your last payment", "your registered branch"). The agent ticks pass/fail for each.

- Pass both → the card unlocks the full summary; the call row stores `verified = true`, which questions, and by whom.
- Fail → details stay hidden; the agent can only give general information; three failed verifications on one account within 24 h create a fraud alert.
- Inbound calls from the customer's registered primary number still need verification (spoofing is cheap). Outbound calls we made to the registered number need only one question, because we dialled the number ourselves.

## 6.6 What the agent sees, by role

| Field | Agent (before verification) | Agent (verified) | Supervisor | Fraud team |
|---|---|---|---|---|
| Name, account number(s), product | ✔ | ✔ | ✔ | ✔ |
| Status (current / in arrears X days / closed) | ✔ | ✔ | ✔ | ✔ |
| Balance, next due, last payment | — | ✔ | ✔ | ✔ |
| Transaction history | — | ✔ (last 10) | ✔ | ✔ |
| Full ID number, date of birth | — | — (compare only) | — | ✔ |
| Fraud flags and notes | tag only | tag only | ✔ | ✔ |
