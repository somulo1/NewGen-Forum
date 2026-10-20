package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

// The API reference is built from the running router, so it lists every route that exists and cannot fall behind. Each
// operation says whether it needs a token, which rate limit applies (read from the same table that enforces it), and the
// errors every route can return. Request and response examples come from apiNotes, which is added to as integrations need it.

var braceRe = regexp.MustCompile(`\{([^}/:]+)(:[^}]*)?\}`)

type note struct {
	Summary  string
	Detail   string
	Request  map[string]any
	Response map[string]any
}

var apiNotes = map[string]note{
	"POST /v1/auth/login": {Summary: "Sign in", Detail: "Returns a short-lived access token and a refresh token. Add `org` (the organization slug) to get a token for that organization, which the organization routes require. Five wrong passwords from one address block it for a while.",
		Request:  map[string]any{"email": "amani@example.com", "password": "…", "totp": "123456 (when two-step sign-in is on)", "org": "acme"},
		Response: map[string]any{"access_token": "…", "refresh_token": "…", "expires_in": 900, "account": "uuid", "org": "uuid"}},
	"POST /v1/auth/refresh": {Summary: "Get a new access token", Detail: "Refresh tokens are single use: each answer carries the next one. Presenting an old one signs the account out everywhere.",
		Request: map[string]any{"refresh_token": "…", "org": "acme"}},
	"POST /v1/auth/signup": {Summary: "Create an account", Request: map[string]any{"name": "Amani Kariuki", "email": "amani@example.com", "password": "at least 10 characters"}, Response: map[string]any{"account": "uuid"}},
	"POST /v1/auth/logout": {Summary: "Sign out this session"},
	"GET /v1/conversations": {Summary: "List conversations you are in", Detail: "Forums, private chats, ticket threads and wall posts. Message bodies are end-to-end encrypted; the server returns envelopes only.",
		Response: map[string]any{"conversations": []any{map[string]any{"id": "uuid", "kind": "forum", "epoch": 1, "my_role": "member", "last_message": "uuid"}}}},
	"POST /v1/conversations": {Summary: "Create a forum, private chat or other conversation", Detail: "The creator wraps the first conversation key for every member's devices and sends the wraps in `keys`.",
		Request: map[string]any{"kind": "forum", "name": "Risk & Compliance", "privacy": "invite", "members": []string{"person uuid"}, "keys": []any{map[string]any{"person": "uuid", "device": "uuid", "epoch": 1, "wrapped": "base64"}}}},
	"POST /v1/conversations/{id}/messages": {Summary: "Send an encrypted message", Detail: "The sender picks the message id (it is bound into the signature) and seals the text on the device. 120 messages a minute per account.",
		Request: map[string]any{"id": "uuid", "epoch": 1, "device": "uuid", "envelope": "base64", "type": "text", "reply_to": nil}},
	"POST /v1/calls": {Summary: "Start a voice or video call", Detail: "`mode` in the answer is `sfu` (a room on a media server: use `sfu.url` and `sfu.token` with a LiveKit client) or `mesh` (a small peer-to-peer room: use `url` and `token`). A full room is 409, too many rooms 429. See docs/CALLS_SCALE.md.",
		Request:  map[string]any{"kind": "video", "context": "forum", "conversation": "uuid"},
		Response: map[string]any{"mode": "sfu", "call": map[string]any{"id": "uuid"}, "sfu": map[string]any{"url": "wss://media.example.com", "token": "…", "can_speak": true, "limit": 800}}},
	"POST /v1/calls/{id}/join":    {Summary: "Join a running call", Detail: "Same answer as starting one. In a big room newcomers beyond the open-mic size join as listeners (`sfu.can_speak` false)."},
	"POST /v1/calls/{id}/control": {Summary: "Host and moderator controls for a big call", Request: map[string]any{"action": "mute | cam_off | remove | mute_all | lock | unlock | end | stage_on | stage_off", "identity": "participant id in the room"}},
	"POST /v1/calls/sfu-hook":     {Summary: "Notices from the media servers", Detail: "Called by the media servers, signed with their API key. Not for integrations."},
	"POST /v1/realtime/ticket":    {Summary: "Get a one-time ticket for the realtime socket", Detail: "Connect to `GET /v1/realtime?ticket=…` with it. Tickets work once and expire quickly."},
	"GET /healthz":                {Summary: "Is the process up"},
	"GET /readyz":                 {Summary: "Are the database and message bus reachable"},
}

func tagOf(route string) string {
	p := strings.Split(strings.TrimPrefix(route, "/"), "/")
	if len(p) >= 2 && p[0] == "v1" {
		return p[1]
	}
	return "system"
}

func summaryOf(method, route string) string {
	words := strings.ReplaceAll(strings.TrimPrefix(strings.TrimPrefix(route, "/v1/"), "/"), "/", " ")
	words = braceRe.ReplaceAllString(words, "")
	verb := map[string]string{"GET": "Get", "POST": "Create or run", "PUT": "Replace", "PATCH": "Change", "DELETE": "Delete"}[method]
	return strings.Join(strings.Fields(verb+" "+words), " ")
}

func (d Deps) openAPI(r chi.Routes) map[string]any {
	type route struct {
		method, path string
		mws          int
	}
	var rs []route
	base := 1 << 30
	_ = chi.Walk(r, func(method, p string, _ http.Handler, mws ...func(http.Handler) http.Handler) error {
		if strings.HasPrefix(p, "/v1/docs") {
			return nil
		}
		p = strings.TrimSuffix(p, "/")
		if p == "" {
			return nil
		}
		rs = append(rs, route{method, p, len(mws)})
		if len(mws) < base {
			base = len(mws)
		}
		return nil
	})
	sort.Slice(rs, func(i, j int) bool {
		if rs[i].path != rs[j].path {
			return rs[i].path < rs[j].path
		}
		return rs[i].method < rs[j].method
	})
	cfg := d.App.Cfg
	paths := map[string]any{}
	tags := map[string]bool{}
	errs := map[string]any{
		"400": map[string]any{"description": "The request is not valid. Body: {\"error\":{\"code\",\"message\"}}"},
		"401": map[string]any{"description": "Missing, expired or invalid token"},
		"403": map[string]any{"description": "Signed in, but your role does not have the right"},
		"404": map[string]any{"description": "Not found, or not visible to you"},
		"429": map[string]any{"description": "Rate limited. Wait `Retry-After` seconds."},
	}
	for _, x := range rs {
		public := x.mws == base
		oapi := braceRe.ReplaceAllString(x.path, "{$1}")
		key := x.method + " " + x.path
		n := apiNotes[key]
		if n.Summary == "" {
			n.Summary = summaryOf(x.method, x.path)
		}
		tag := tagOf(x.path)
		tags[tag] = true
		limit := "per account: " + itoa(cfg.RateRPS*10) + " requests per 10 s; per address: " + itoa(cfg.RateIP) + " per minute"
		if public {
			limit = "per address: " + itoa(cfg.RateIP) + " per minute"
			if publicPath(x.path) {
				limit += "; sign-in, sign-up, reset and invite routes " + itoa(cfg.RateAuthIP) + " per minute"
			}
		}
		if t, ok := Tiers[key]; ok {
			limit += "; this route " + itoa(t.Limit) + " per " + t.Window.String()
		}
		desc := n.Detail
		if desc != "" {
			desc += "\n\n"
		}
		desc += "Rate limit: " + limit + "."
		op := map[string]any{"summary": n.Summary, "description": desc, "tags": []string{tag}, "operationId": strings.ToLower(x.method) + strings.NewReplacer("/", "_", "{", "", "}", "").Replace(x.path),
			"x-rate-limit": limit, "responses": map[string]any{"200": map[string]any{"description": "Success. JSON body (201 when something was created, 204 when there is nothing to return)", "content": jsonExample(n.Response)}}}
		resp := op["responses"].(map[string]any)
		for k, v := range errs {
			resp[k] = v
		}
		if public {
			op["security"] = []any{}
		}
		var params []any
		for _, m := range braceRe.FindAllStringSubmatch(x.path, -1) {
			params = append(params, map[string]any{"name": m[1], "in": "path", "required": true, "schema": map[string]any{"type": "string"}})
		}
		if len(params) > 0 {
			op["parameters"] = params
		}
		if x.method == "POST" || x.method == "PUT" || x.method == "PATCH" {
			body := map[string]any{"type": "object"}
			if n.Request != nil {
				body["example"] = n.Request
			}
			op["requestBody"] = map[string]any{"content": map[string]any{"application/json": map[string]any{"schema": body}}}
		}
		pm, _ := paths[oapi].(map[string]any)
		if pm == nil {
			pm = map[string]any{}
			paths[oapi] = pm
		}
		pm[strings.ToLower(x.method)] = op
	}
	var tl []any
	names := make([]string, 0, len(tags))
	for t := range tags {
		names = append(names, t)
	}
	sort.Strings(names)
	for _, t := range names {
		tl = append(tl, map[string]any{"name": t})
	}
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{"title": "Hub API", "version": "v1",
			"description": "Everything under /v1. Send `Authorization: Bearer <access token>`; organization routes need a token issued for an organization (`org` at sign-in). Errors are `{\"error\":{\"code\",\"message\"}}`. Answers carry `RateLimit-Limit` and `RateLimit-Remaining`; a 429 carries `Retry-After`. Message, file and note contents are end-to-end encrypted on the device: the API stores and returns envelopes it cannot read. This reference lists every route the server has; request and response examples are filled in where integrations need them."},
		"servers":    []any{map[string]any{"url": "/"}},
		"tags":       tl,
		"paths":      paths,
		"components": map[string]any{"securitySchemes": map[string]any{"bearerAuth": map[string]any{"type": "http", "scheme": "bearer", "bearerFormat": "Hub access token"}}},
		"security":   []any{map[string]any{"bearerAuth": []any{}}},
	}
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

func jsonExample(v map[string]any) map[string]any {
	if v == nil {
		return map[string]any{"application/json": map[string]any{}}
	}
	return map[string]any{"application/json": map[string]any{"example": v}}
}

func (d Deps) mountDocs(r chi.Router) {
	var once sync.Once
	var spec []byte
	build := func() []byte {
		once.Do(func() { spec, _ = json.MarshalIndent(d.openAPI(r), "", " ") })
		return spec
	}
	r.Get("/v1/docs/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write(build())
	})
	r.Get("/v1/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; connect-src 'self'")
		_, _ = w.Write([]byte(docsPage))
	})
}

const docsPage = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Hub API reference</title>
<style>
:root{--bg:#fff;--fg:#14201a;--mut:#5d6f65;--card:#f5f8f6;--line:#dbe4df;--acc:#2f7d4f}
@media (prefers-color-scheme:dark){:root{--bg:#0a100d;--fg:#e8efea;--mut:#9db0a6;--card:#121b17;--line:#24322c;--acc:#b8f36b}}
*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--fg);font:15px/1.5 system-ui,sans-serif}
header,main{max-width:1000px;margin:0 auto;padding:16px}h1{margin:8px 0 4px;font-size:26px}p.lead{color:var(--mut);margin:0 0 12px;white-space:pre-line}
input{width:100%;height:40px;padding:0 12px;border-radius:10px;border:1px solid var(--line);background:var(--card);color:var(--fg);font:inherit}
h2{font-size:15px;text-transform:uppercase;letter-spacing:.06em;color:var(--mut);margin:24px 0 8px}
details{border:1px solid var(--line);border-radius:10px;background:var(--card);margin:6px 0}summary{cursor:pointer;padding:9px 12px;display:flex;gap:10px;align-items:center;flex-wrap:wrap}
.m{font:700 11px monospace;min-width:54px;text-align:center;padding:3px 6px;border-radius:6px;background:var(--acc);color:#0a100d}.m.GET{background:#5aa9e6}.m.DELETE{background:#e5484d;color:#fff}.m.PUT,.m.PATCH{background:#ffb86b}
code,pre{font:13px/1.4 ui-monospace,monospace}pre{background:var(--bg);border:1px solid var(--line);border-radius:8px;padding:10px;overflow:auto;margin:6px 0}
.b{padding:0 12px 12px}.s{color:var(--mut);flex:1;min-width:140px}.pub{font-size:11px;color:var(--mut);border:1px solid var(--line);border-radius:10px;padding:1px 7px}
</style></head><body><header><h1>Hub API reference</h1><p class="lead" id="lead"></p><input id="q" type="search" placeholder="Find a route, e.g. calls or /v1/tickets" aria-label="Find a route"><p><a id="raw" href="/v1/docs/openapi.json" style="color:var(--acc)">openapi.json</a> (OpenAPI 3) for code generators and tools</p></header><main id="out">Loading…</main>
<script>
const esc=s=>String(s).replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]));
fetch('/v1/docs/openapi.json').then(r=>r.json()).then(spec=>{
 document.getElementById('lead').textContent=spec.info.description;
 const by={};
 for(const [p,ms] of Object.entries(spec.paths))for(const [m,op] of Object.entries(ms))(by[op.tags[0]]=by[op.tags[0]]||[]).push({p,m:m.toUpperCase(),op});
 const out=document.getElementById('out');
 const draw=f=>{out.innerHTML=Object.keys(by).sort().map(t=>{const rows=by[t].filter(x=>!f||(x.m+' '+x.p+' '+x.op.summary).toLowerCase().includes(f));
  return rows.length?'<h2>'+esc(t)+'</h2>'+rows.map(x=>{const rq=x.op.requestBody&&x.op.requestBody.content['application/json'].schema.example,rs=x.op.responses['200'].content&&x.op.responses['200'].content['application/json'].example;
   return '<details><summary><span class="m '+x.m+'">'+x.m+'</span><code>'+esc(x.p)+'</code><span class="s">'+esc(x.op.summary)+'</span>'+(x.op.security&&!x.op.security.length?'<span class="pub">no token</span>':'')+'</summary><div class="b"><p style="white-space:pre-line">'+esc(x.op.description)+'</p>'+(rq?'<b>Request</b><pre>'+esc(JSON.stringify(rq,null,2))+'</pre>':'')+(rs?'<b>Response</b><pre>'+esc(JSON.stringify(rs,null,2))+'</pre>':'')+'</div></details>'}).join(''):''}).join('')||'Nothing matches.'};
 draw('');document.getElementById('q').addEventListener('input',e=>draw(e.target.value.trim().toLowerCase()));
}).catch(()=>{document.getElementById('out').textContent='Could not load the reference.'});
</script></body></html>`
