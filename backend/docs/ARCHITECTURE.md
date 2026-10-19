# Architecture

## Processes
| Binary | Job | Scales by |
|---|---|---|
| `cmd/api` | HTTP + WebSocket API, all widgets | more nodes behind a load balancer; each node holds one Redis subscription |
| `cmd/calls` | voice/video signaling and room control, no database | more instances; a room is pinned to one by hashing its id |
| `cmd/worker` | ticket TAT sweeps, call history from the calls service | any number; every job is lock-safe |
| `cmd/migrate` | applies SQL with the owner role | run once per release |
| `cmd/integrity` | read-only audit of every sealed row and ledger | scheduled |

## Three kinds of data
1. **End-to-end encrypted (the server cannot read it):** direct messages, forums, polls, the community wall, moments,
   notes, ticket threads, project plans/tasks, calendar entry details, announcements/memos, file contents and names,
   secure-share keys. Devices hold X25519+Ed25519 keys; each conversation has a random key per *epoch*, wrapped separately
   for every member device. Messages are AES-256-GCM bound (AAD) to org, conversation, sender, id and epoch, then signed.
   Removing a member rotates the key, so they cannot read anything sent afterwards.
2. **Encrypted at rest for the server (it needs to read it):** phone, date of birth, gender, email of a person; organization
   settings and billing; award citations; alert notes; forum names. AES-256-GCM with a per-organization key wrapped by a KEK
   (KMS in production); every ciphertext is bound to its table, column, row and organization.
3. **Plain operational facts** the server must act on: ticket category/priority/step/due time, project status, roles and
   rights, approval state, who is in which conversation. No content.

## Tamper resistance ("data only from the system")
* The database accepts writes only from connections that present the **guard token** (`hub.assert_app`). Scripts and humans
  without it are refused, and the DBA role has no data access at all.
* Every mutable row is **sealed**: an HMAC (key outside the database) over the row's canonical JSON, its table and version.
  Writes move a row through *pending → sealed* in one transaction; the database enforces the protocol (`hub.require_mac`).
* Ledger, audit log and key tables are **append-only** (privileges revoked and triggers).
* Each organization has a **hash-chained, signed ledger**; deleting or editing history breaks every later hash, and
  removing the tail is caught by the head record.
* **Row-level security** on every tenant table: one organization can never see or write another's rows.
* `cmd/integrity` re-verifies all of it with the read-only role and exits non-zero on any mismatch.
* Migrations are checksummed; editing an applied file is refused.

A database *superuser* can still disable triggers; the point is that such an edit cannot produce a valid MAC or ledger entry
(the keys are not in the database), so it is detected, not prevented. Prevent it with the standard controls: no superuser
access in production, `pgaudit`, WAL archiving to write-once storage.

## Realtime
Every change is published once to Redis; each API node routes events to its own sockets. Sockets open with a 30-second
single-use ticket, are origin-checked, join org/account channels automatically and conversation channels after a membership
check. Slow clients are dropped (bounded queues). Signing a person out or blocking them closes their sockets.

## Approvals
`approval_flows` say how many levels each action needs and who approves at each (a title, a role, "Line manager"; locked roles
can stand in). A request carries the sealed change; nothing is applied until the last level approves. Modules register
what to do on approval (`person.update`, `rights.role`, `project.activate`, `comm.publish`, `award.grant`, `share.activate`).
