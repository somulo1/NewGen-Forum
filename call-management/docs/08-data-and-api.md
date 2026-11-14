# 8. Data model and API

## 8.1 Main tables (PostgreSQL, the call service's own database)

```
agents            (person_id PK, org_id, extension, display_name, queues[], active, created_at)
calls             (id PK uuid, org_id, pbx_call_id UNIQUE, direction in|out, from_e164, to_e164,
                   did, trunk, queue, ivr_path, agent_person_id, extension,
                   customer_ids[], customer_name_snapshot, account_no_snapshot, match_source,
                   verified bool, verified_by, verification_questions[],
                   status, result, disposition, started_at, answered_at, ended_at,
                   wait_ms, ring_ms, talk_ms, hold_ms, wrap_ms, hung_up_by,
                   campaign_id, callback_id, transferred_from_call_id, quality_mos_avg, created_at)
call_legs         (call_id, seq, extension_or_number, joined_at, left_at, role agent|consult|conference|supervisor)
call_events       (id, call_id, type, at, payload jsonb)              -- normalised timeline
pbx_events_raw    (id, received_at, topic, payload jsonb)             -- everything the PBX sent (30 days)
call_notes        (id, call_id, author_person_id, text, created_at)
recordings        (id, call_id, storage_key, sha256, bytes, duration_ms, encrypted_key_id, retain_until, legal_hold bool)
voicemails        (id, box, from_e164, customer_ids[], storage_key, duration_ms, transcript, handled_by, handled_at)
agent_status_log  (person_id, status, reason, from_at, to_at)
callback_requests (id, from_e164, customer_ids[], queue, requested_at, due_at, assigned_to, tries, status)
campaigns         (id, name, queue, caller_id, window_start, window_end, max_tries, gap_min, mode, status, created_by)
campaign_targets  (campaign_id, customer_id, e164, tries, next_try_at, last_result, final bool, locked_by, locked_until)
dnc_numbers       (e164 PK, reason, added_by, added_at, source manual|sms_stop|lms_flag)
unidentified_callers (e164 PK, name, reason, notes, captured_by, captured_at, linked_customer_id)
messages          (id, channel sms|whatsapp, thread_customer_id, e164, direction, template, body, status, provider_id, sent_by, at)
routing_settings  (org_id, json)    -- mirror of PBX hours, IVR, groups for display/edit
audit_events      (id bigserial, at, actor_person_id, action, target, details jsonb, prev_hash, hash)
```

Indexes that matter: `calls(org_id, started_at desc)`, `calls(from_e164)`, `calls(to_e164)`, `calls(agent_person_id, started_at)`, `calls(pbx_call_id)`, `campaign_targets(campaign_id, next_try_at) where not final`.

Tables are partitioned by month (`calls`, `call_events`, `pbx_events_raw`, `audit_events`) because volume is high: 5,000 calls/day ≈ 1.8 million calls/year, with ~15 events each.

## 8.2 REST API for Hub (all under `/v1`, JSON, call-service token required)

| Method & path | Right | Purpose |
|---|---|---|
| `GET /me` | calls.use | My extension, queues, status, WebRTC config |
| `POST /me/webrtc-signature` | calls.use | Get the login signature for the browser SDK |
| `PUT /me/status` | calls.use | Set Available / Away / DND … (+ reason) |
| `PUT /me/queues/{queue}` | calls.use | Log in/out of a queue |
| `POST /calls` | calls.use | Ask to place an outbound call (runs the checks in 05) |
| `GET /calls?from&to&dir&agent&queue&q&page` | calls.use (own) / calls.supervise (team) | Call history, paginated |
| `GET /calls/{id}` | as above | One call with timeline, notes, recording link |
| `POST /calls/{id}/transfer` `{to, mode}` | calls.use | Server-side transfer (when not done in SDK) |
| `POST /calls/{id}/park` | calls.use | Park |
| `POST /calls/{id}/notes`, `PUT /calls/{id}/outcome` | calls.use | Wrap-up |
| `POST /calls/{id}/verify` `{questions, passed}` | calls.use | Record identity verification |
| `GET /lookup?number=` | calls.use | Caller card for a number |
| `GET /customers/search?q=` | calls.use | Search the LMS (proxied) |
| `POST /unidentified` | calls.use | Save an unknown caller's details |
| `GET /recordings/{id}/stream` | calls.recordings.play | Stream audio (audited) |
| `POST /recordings/{id}/download` `{reason}` | calls.recordings.download | Download (audited) |
| `GET /voicemails`, `PUT /voicemails/{id}/handled` | calls.use | Voicemail |
| `GET /callbacks`, `POST /callbacks/{id}/accept` | calls.use | Call-backs |
| `GET/POST /campaigns`, `POST /campaigns/{id}/start|pause` | calls.campaigns.manage / .run | Campaigns |
| `GET/POST/DELETE /dnc` | calls.dnc.manage | Do-not-call list |
| `GET /threads`, `POST /threads/{customer}/messages` | calls.messages | SMS/WhatsApp |
| `GET /monitor` | calls.supervise | Wallboard snapshot |
| `POST /monitor/{extension}` `{mode: listen|whisper|barge}` | calls.supervise | Supervision (audited) |
| `GET /reports/{name}?range=` | calls.reports | Reports, CSV with `Accept: text/csv` |
| `GET/PUT /routing` | calls.routing.manage | Hours, IVR, groups, waiting settings |
| `GET /audit?…` | calls.audit | Compliance log |

Errors use Hub's existing format (`{ "error": { "code", "message" } }`), and every list endpoint paginates (`page`, `size` ≤ 100, `total`).

## 8.3 Live events to Hub (WebSocket `/v1/live`)

The frontend opens one WebSocket with the token; the service subscribes it to channels the person may see: their own (`agent:{person}`), and for supervisors `queue:{name}` and `wallboard`.

| Event | Payload (main fields) | Who receives it |
|---|---|---|
| `call.ringing` | call_id, from, queue, waited_s, caller card | The ringing agent |
| `call.updated` | call_id, state, on_hold, legs, timer start | Agent on the call, supervisors |
| `call.ended` | call_id, result, wrap_until | Agent, supervisors |
| `agent.status` | person, status, since | Supervisors, wallboard |
| `queue.status` | queue, waiting, longest_s, sla | Supervisors, wallboard |
| `alert` | id, text, severity | Supervisors |
| `callback.new` | id, queue, number | Agents in the queue |
| `message.new` | thread, preview | Agents with the thread |

Messages carry a sequence number; on reconnect the frontend sends the last one it saw and gets what it missed (or a full snapshot).

## 8.4 How this maps to the existing Hub screens

The frontend's simulation (`frontend/logic/pbx/store.js`) already has one function per action marked `PBX:`. Each is replaced by a call to the endpoint above, and its state is filled from the live events, with no change to the screens.

| Frontend function | Becomes |
|---|---|
| `dial(number)` | `POST /v1/calls`, then SDK `call()` |
| `answer()`, `decline()`, `toggleHold()`, `sendDtmf()`, `transfer()`, `endCall()` | SDK functions; state from `call.updated` |
| `finishWrap()` | `PUT /v1/calls/{id}/outcome` + notes |
| `supervise()` | `POST /v1/monitor/{ext}` |
| `campaignDial()` | `POST /v1/campaigns/{id}/next` |
| history, voicemail, recordings, reports | the GET endpoints, paginated |
| `simulateIncoming()` | removed (real calls) |
