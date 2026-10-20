# Going live: runbook

Work through this in order on a **staging** deployment that is a copy of production, then repeat on production.

## 1. Media servers (big calls)
1. Provision media hosts per region (bare metal or a cheap-bandwidth provider; 16+ cores, 10 Gbps). Open the ports in `deploy/media/firewall.md`.
2. Put a certificate for the TURN domain at `/etc/hub/turn.crt` and `turn.key`. Point DNS for each region's address at its nodes.
3. On each host: copy `deploy/media/`, set `PUBLIC_IP`, `HUB_REDIS_ADDR/PASS`, `HUB_SFU_KEY/SECRET`, `HUB_API_URL`, `HUB_TURN_DOMAIN`, then `docker compose up -d`.
4. Set `HUB_SFU_NODES` on the API (one `url|key|secret` per region) and restart the API. Start a forum call and confirm `mode: "sfu"`.
5. From a machine outside the network, check TURN: join a call from a phone on mobile data with Wi-Fi off.
6. Load test with `lk load-test` (see `docs/CALLS_SCALE.md`) from other hosts, kill a node mid-test, and confirm people reconnect.

## 2. Database and data safety
1. Install pgaudit and wal-g (or pgBackRest). Include `deploy/ops/postgresql.hub.conf`. Create the roles with `scripts/bootstrap_db.sh`, then run `deploy/ops/hardening.sql` and confirm all four flags in its last query are `f`.
2. Create the backup bucket **with Object Lock in compliance mode** (35+ days). Create an `age` key pair; keep the private key offline (not on the backup host).
3. Install `deploy/ops/hub-backup.*`, `hub-restore-test.*`, `hub-integrity.*` to `/etc/systemd/system/`, copy `scripts/` to `/opt/hub/scripts/`, write `/etc/hub/backup.env`, then `systemctl enable --now` each timer.
4. Run `scripts/backup.sh` and `scripts/restore-test.sh` by hand once and read the output. **Do not go live until a restore has worked.**
5. Practise point-in-time recovery once (restore base backup + WAL to a chosen time) and write down how long it took.
6. Run `cmd/integrity` (build it to `bin/hub-integrity`) and confirm exit code 0. Exit code 2 means altered records or a broken audit chain: treat it as an incident.

## 3. Monitoring
1. The API serves `/metrics` (needs `HUB_METRICS_TOKEN`), the calls service `/metrics` (internal only), media servers port 6789. Install node, postgres and redis exporters. Use `deploy/monitoring/prometheus.yml` and `alerts.yml`; point Alertmanager at the on-call pager.
2. Backup and integrity results reach Prometheus through node_exporter's textfile collector (`HUB_METRICS_DIR`): alerts `BackupStale` and `LedgerTampered`.
3. Ship the JSON logs (stdout) to a log store. Send errors to an error tracker if you use one. Add an external uptime check on `/readyz`.
4. Trigger each alert once on staging (stop the API, stop a media node, fill a test disk) and confirm someone is paged.

## 4. Load test
Run against staging with demo data, from a separate machine:
`go run ./cmd/loadtest -url https://staging... -people 250 -duration 10m` (250 is the size of the demo directory; run several instances for more).
Gate: server errors < 0.5%, sign-in failures < 2%, no dropped sockets, p95 reads < 500 ms, database pool never above 90%.
Baseline on a laptop with the database and API on the same machine: 200 people, 195 requests/s, no errors, p95 under 50 ms (reads), sign-in p95 75 ms.
Then: calls (see item 1.6), a fresh organization signing up and onboarding 500 people by import, and a 30-minute soak.

## 5. First start
`migrate`, `bootstrap platform`, `bootstrap owner` (see docs/SEEDING.md), optionally `bootstrap org -slug <slug> -roles starter`, then `bootstrap protect`.

## 6. Before the first customer
Frontend built with `NUXT_PUBLIC_API_BASE=https://api.example.com`; `HUB_ALLOWED_ORIGINS` is that site; no demo seed was ever run on production; keys are in the KMS and a lost-key procedure is written; an on-call rota exists.
