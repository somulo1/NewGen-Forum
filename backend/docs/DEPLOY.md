# Deploying the backend (one process)

The API binary sets itself up. On every start it applies any missing database migrations, installs the reference data (sounds),
creates the first owner and organization (only when the database has no accounts and the settings below are given), runs the
background jobs, and serves the API. There is nothing to run before it.

**Build:** `go build -tags netgo -ldflags '-s -w' -o app`  (Root Directory: `backend`)
**Start:**  `./app`

## Environment
Set the variables from `.env.example`. On a fresh platform you need at least:
- `HUB_ENV=prod`, `HUB_ALLOWED_ORIGINS` (your frontend address), the three database URLs, `HUB_REDIS_ADDR`/`HUB_REDIS_PASS`,
  and the secrets from `scripts/gen-keys.sh` (never change them afterwards).
- The port: the server uses `PORT` when `HUB_HTTP_ADDR` is not set (Render sets `PORT`).
- First owner, created once on an empty database: `HUB_OWNER_NAME`, `HUB_OWNER_EMAIL`, `HUB_OWNER_PASSWORD`, `HUB_ORG_NAME`,
  `HUB_ORG_SLUG` (the password must be changed at first sign-in unless `HUB_OWNER_KEEP_PASSWORD=1`).
- Email: the `HUB_SMTP_*` values, or reset codes cannot be delivered.

## Switches
`HUB_AUTO_MIGRATE=off` (run `cmd/migrate` yourself), `HUB_WITH_WORKER=off` (run `cmd/worker` elsewhere), `HUB_WITH_CALLS=on`
(serve the small-call signalling on this port under `/calls-signal`; set `HUB_CALLS_URL=wss://<this address>/calls-signal`).
Several nodes may start at once: migrations take a database lock, so one waits for the other.

## Still outside the binary
The database roles (`scripts/bootstrap_remote.sh`, once), Redis, S3 storage for files (`HUB_FILE_SERVERS`; a platform's disk is
wiped on deploy), and the media server for big calls (LiveKit; it needs UDP, so it cannot run on most web-service platforms).
`go run ./cmd/bootstrap protect` checks the protections after a deploy.
