# Hub backend

Go services for Hub: one API (`cmd/api`), a separate calls service (`cmd/calls`), a background worker (`cmd/worker`),
and the tools `cmd/migrate`, `cmd/bootstrap` (reference data: sounds, rights catalog, starter roles) and `cmd/integrity`. PostgreSQL 18 and Redis are the only dependencies.

## Run it locally

```sh
export PATH=/usr/lib/postgresql/18/bin:$PATH GOTMPDIR=$HOME/.cache/gotmp
bash scripts/bootstrap_db.sh          # once: roles and the empty database
go run ./cmd/migrate                  # schema (refuses to run if an applied migration file was edited)
go run ./cmd/bootstrap platform          # sounds (the API also installs these at start-up)
go run ./cmd/bootstrap org -slug <slug> -roles starter   # an organization's catalog and starter roles
go run ./cmd/api &                    # :8080
go run ./cmd/calls &                  # :8090
go run ./cmd/worker &                 # ticket timing + call history
go run ./cmd/integrity                # re-check the whole database against the MAC key and ledger
go test ./...
```

Sign in as `samwel.omulo@company.com` / `Hub!Demo2026`, organization `demo`. Every demo person uses the same password.
Development defaults (database 5433, Redis 6390, throwaway keys) are in `internal/config`; production refuses to start
without real keys.

Docs: [ARCHITECTURE](docs/ARCHITECTURE.md) · [ENCRYPTED PAYLOADS](docs/ENCRYPTED_PAYLOADS.md) · [API](docs/API.md) · [LIMITS AND ROADMAP](docs/ROADMAP.md)
