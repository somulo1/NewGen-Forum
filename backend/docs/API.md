# API reference

For anyone integrating with Hub, the API describes itself:

* `GET /v1/docs` (also `/api/v1/docs`) is the readable reference: every route, who may call it, what it takes (path, query, JSON body), what it returns and the errors it can give.
* `GET /v1/docs/openapi.json` is the OpenAPI 3 document for code generators and API tools.

It is built from the running router and written entries in `internal/server/apidocs/*.json`. A test fails if any route has no
entry, so a new route cannot ship undocumented: add its entry in the same change.

## Which environments serve it
`HUB_ENV` names the environment (`dev`, `staging` or `prod`) and `HUB_DOCS_ENVS` lists the ones that serve the reference
(default `dev,staging,prod`). Remove one from the list to hide the reference there; an empty list hides it everywhere.

## Other settings
`HUB_RATE_IP`, `HUB_RATE_AUTH_IP`, `HUB_RATE_RPS` set the request limits; `HUB_TRUSTED_PROXIES` lists the proxies whose
`X-Forwarded-For` header is believed (default: loopback). Behind a load balancer, list its address range.
