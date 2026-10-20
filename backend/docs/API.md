# API reference

The reference is served by the API itself and built from the running router, so it lists every route and cannot go stale:

* `GET /v1/docs` is the readable reference (search, per-route auth and rate limit, examples).
* `GET /v1/docs/openapi.json` is the OpenAPI 3 document for code generators and API tools.

`HUB_DOCS=off` hides both. `go run ./cmd/routes` prints the bare route list.

## Rate limits
Every answer carries `RateLimit-Limit` and `RateLimit-Remaining`; a refusal is `429` with `Retry-After` (seconds).

| Scope | Default | Setting |
|---|---|---|
| each address, all routes | 600 / minute | `HUB_RATE_IP` |
| each address, sign-in / sign-up / reset / invite routes | 30 / minute (fails closed if Redis is down) | `HUB_RATE_AUTH_IP` |
| each signed-in account | 500 / 10 seconds | `HUB_RATE_RPS` (x10) |
| expensive routes (start a call, create a conversation, add people, upload, ...) | per account, see `/v1/docs` | `Tiers` in `internal/server/limits.go` |
| wrong passwords, one address | 4 then blocked | |

Counts are shared by every API node through Redis (sliding window). `X-Forwarded-For` is believed only when the connection
comes from a proxy listed in `HUB_TRUSTED_PROXIES` (default: loopback); otherwise callers could choose their own address and
walk around the per-address limits. Behind a load balancer, list its address range.
