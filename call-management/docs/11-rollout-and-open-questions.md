# 11. Rollout plan and open questions

## 11.1 Phases

| Phase | Weeks (indicative) | Delivers | Done when |
|---|---|---|---|
| **0. Decisions and access** | 1–2 | Answers to 11.2; Yeastar licences (Open API, WebRTC SDK) confirmed in writing; LMS API agreed; DPIA started | Signed-off decisions list |
| **1. Foundations** | 3–6 | Open API endpoints verified against Yeastar's guide (03 §3.4); call service skeleton; token manager; event stream; call state; Hub token exchange; agents↔extensions; softphone registration in Hub | An agent makes and receives a real call in Hub; events appear in the database |
| **2. Inbound with identity** | 7–10 | Caller lookup (LMS + secondary sources); ringing card; verification; recording archive; history; notes and outcomes; voicemail | A test customer calls, is identified, verified, recorded, and the record is complete |
| **3. Outbound and centre tools** | 11–15 | Outbound checks (DNC, hours, frequency); call-backs; campaigns (preview, then progressive); wallboard and alerts; supervision; SMS | Collections team runs a pilot campaign |
| **4. Messaging, routing, reports** | 16–20 | WhatsApp; routing edits from Hub (where the API allows); reports and exports; transcription | Full team live |
| **5. Hardening** | ongoing | Load test at 2× peak, failover drills, audit chain checks, security review/penetration test | Go-live sign-off |

Pilot with one team (e.g. 5 collections agents) for two weeks before everyone.

## 11.2 Decisions needed (including things not asked about yet)

**Telephony**
1. Yeastar edition: on-premises appliance/software vs Yeastar Cloud? (Affects data location, latency, maintenance.)
2. Which carriers and how many trunk channels? A second carrier for failover?
3. Which public numbers (DIDs) for which teams, and which caller IDs agents may use?
4. Does our Yeastar plan include the Open API and the WebRTC (Linkus) SDK? Written licence for `ys-webrtc-sdk-core` in a commercial web app?
5. Headsets: which model (USB, noise-cancelling) for agents? Audio quality depends heavily on it.
6. Will agents work from home? (Needs TURN, VPN policy, bandwidth checks.)

**Customer data (LMS)**
7. Who builds the LMS read API in 06 §6.2, and its timeline?
8. Are LMS phone numbers clean and in E.164? (A data clean-up may be needed first.)
9. Can the LMS accept interaction write-backs (promise to pay, notes)?
10. Are loan applicants/leads available for identifying non-customers?

**Compliance and policy**
11. Recording retention period (1 year? 7 years for loan disputes?).
12. What happens if a customer refuses recording?
13. Collection contact limits: calls per day/week per customer, allowed hours, no-call days.
14. Verification questions: which ones, and how many for inbound vs outbound?
15. Who may listen to recordings and supervise calls?
16. DPIA owner and data protection officer sign-off.

**Operations**
17. Opening hours, holidays, and after-hours handling (voicemail, mobile on-call, fraud line 24/7?).
18. Queues and teams (Customer care, Collections, Onboarding, Fraud?) and their service level targets.
19. Outcome (disposition) list per queue.
20. Professional voice recording for the greeting, IVR and on-hold messages (English and Swahili?). **Languages** for IVR: should callers choose English or Swahili first?
21. Who handles call-backs and queue voicemails, and within what time?

**Things you may have forgotten**
- **Swahili** IVR and greetings for many customers.
- **USSD/M-Pesa flows**: should the agent be able to send an M-Pesa STK push (payment prompt) to the customer during a call, so a "promise to pay" becomes a payment immediately? (Needs the M-Pesa Daraja API; high value for collections.)
- **Customer satisfaction survey** after the call (press 1–5, or an SMS survey).
- **Speech analytics** (keywords like "complaint", "fraud", "lawyer") on transcripts to flag calls for review.
- **Quality scoring**: supervisors score sample calls against a checklist; agents see their scores.
- **Workforce planning**: forecasting how many agents are needed per hour from call history.
- **Emergency/vulnerable customers**: a way to flag customers in hardship or bereavement so collections stop.
- **Complaints register** linked to calls (regulators may ask for complaint statistics).
- **Staff training mode**: new agents' calls auto-flagged for supervisor listening.
- **Number reputation**: register our numbers with carriers so they are not flagged as spam ("Likely scam") on customers' phones.
