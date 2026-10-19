# Limits, decisions and what comes next

Be clear-eyed about these before go-live.

## Scale
* Designed for many organizations and millions of users: tenant tables are keyed by `org_id` with row-level security,
  `messages` (32) and `records` (16) are hash-partitioned by organization, keyset pagination everywhere, one Redis
  subscription per node, stateless API nodes. **Load-test it** before promising numbers; connection pooling (PgBouncer),
  read replicas and moving hot tables to per-region clusters are the next steps for the very largest tenants.
* **End-to-end encryption and big audiences:** every member device needs its own wrapped key, so joining/leaving costs grow
  with audience size. An encrypted audience is capped at 5,000 people (`chat.MaxAudience`); larger organizations use one wall
  per branch/department. The scalable fix is MLS / TreeKEM (O(log n) group changes); the envelope format is versioned for it.
* **Key custody is client-driven.** When someone joins, leaves or adds a device, an online member's device must wrap or
  rotate keys (`GET /e2ee/todo`). Until then new members see ciphertext they cannot open. A client must run this loop.
* **Key directory trust:** the server serves public keys, so a malicious server could substitute one. Fingerprints are exposed
  for out-of-band comparison and device registration is bound to the account; key transparency (publishing the directory
  into the signed ledger) is the planned hardening.
* **Calls:** peer-to-peer mesh (DTLS-SRTP, end-to-end) limited to 8 people per room. Larger meetings need an SFU (pion) with
  insertable-stream E2EE; the signaling already relays opaque `key` messages for it.

## Not done / needs an external decision
* Real email/SMS delivery (dev mailer only), card-payment gateway (only last-4 is stored), KRA eTIMS fiscal numbers
  (invoice numbers are layout placeholders and flagged `integrated:false`), geo-IP for rule r2 (hook `App.Geo`),
  object storage for files (filesystem driver now, same interface for S3/GCS), a KMS/HSM for the KEK.
* Search over encrypted content is client-side by design.
* Frontend untouched, as instructed: it still needs the crypto client (device keys, wrap/unwrap, seal/open) and to call this API.

## Production checklist
Real `HUB_*` keys from a KMS; unique `HUB_DB_GUARD`; TLS everywhere; no superuser access; `pgaudit`; WAL archive to write-once
storage; run `cmd/integrity` on a schedule and alert on exit code 2; rate limits at the edge; run several `cmd/calls` instances
and list them in `HUB_CALLS_URL` (comma separated).
