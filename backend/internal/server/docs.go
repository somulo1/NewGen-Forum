package server

import (
	"embed"
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

// The API reference is built from the running router, so it lists every route that exists. What each route does, who may
// call it, what it takes and what it returns is written in apidocs/*.json (one file per area); a test fails if any route has
// no entry, so a new route cannot ship undocumented.

//go:embed apidocs/*.json
var notesFS embed.FS

var braceRe = regexp.MustCompile(`\{([^}/:]+)(:[^}]*)?\}`)

// note is the written description of one route. Request, Response and Query are example objects: each field's value says
// what it is ("uuid, required. The person to add"). Request null means the route takes no body.
type note struct {
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	Access      string            `json:"access"` // who may call it
	Status      int               `json:"status"` // success status, 200 when absent
	Query       map[string]string `json:"query"`
	Request     json.RawMessage   `json:"request"`
	Response    json.RawMessage   `json:"response"`
	Errors      map[string]string `json:"errors"` // status -> when it happens
}

var apiNotes = loadNotes()

func loadNotes() map[string]note {
	out := map[string]note{}
	files, _ := notesFS.ReadDir("apidocs")
	for _, f := range files {
		b, err := notesFS.ReadFile("apidocs/" + f.Name())
		if err != nil {
			panic(err)
		}
		var m map[string]note
		if err := json.Unmarshal(b, &m); err != nil {
			panic("apidocs/" + f.Name() + ": " + err.Error())
		}
		for k, v := range m {
			out[k] = v
		}
	}
	return out
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

var typeWords = map[string]string{"string": "string", "uuid": "string", "email": "string", "timestamp": "string", "date": "string", "base64": "string", "url": "string",
	"integer": "integer", "int": "integer", "number": "number", "boolean": "boolean", "bool": "boolean", "object": "object", "array": "array"}

// schemaOf turns an example object, where each value says what the field is, into a JSON schema. A value that starts with a
// type word ("uuid, required. The person") gets that type; the rest is its description.
func schemaOf(raw json.RawMessage) map[string]any {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return map[string]any{"type": "object"}
	}
	return schemaVal(v)
}

func schemaVal(v any) map[string]any {
	switch x := v.(type) {
	case map[string]any:
		props := map[string]any{}
		for k, c := range x {
			props[k] = schemaVal(c)
		}
		return map[string]any{"type": "object", "properties": props}
	case []any:
		if len(x) == 0 {
			return map[string]any{"type": "array", "items": map[string]any{}}
		}
		return map[string]any{"type": "array", "items": schemaVal(x[0])}
	case string:
		out := map[string]any{"description": x}
		if w := strings.FieldsFunc(x, func(r rune) bool { return r == ' ' || r == ',' || r == '.' || r == ':' || r == '(' }); len(w) > 0 {
			if t, ok := typeWords[strings.ToLower(w[0])]; ok {
				out["type"] = t
			}
		}
		return out
	case nil:
		return map[string]any{"nullable": true}
	}
	return map[string]any{"example": v}
}

func (d Deps) openAPI(r chi.Routes) map[string]any {
	type route struct {
		method, path string
		mws          int
	}
	var rs []route
	base := 1 << 30
	_ = chi.Walk(r, func(method, p string, _ http.Handler, mws ...func(http.Handler) http.Handler) error {
		if strings.HasPrefix(p, "/v1/docs") || strings.HasPrefix(p, "/api/v1/docs") {
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
	paths := map[string]any{}
	tags := map[string]bool{}
	generic := map[string]string{
		"400": "The request is not valid. Body: {\"error\":{\"code\",\"message\"}}.",
		"401": "Missing, expired or invalid token.",
		"403": "Signed in, but your role does not have the right to do this.",
		"404": "Not found, or not visible to you.",
	}
	for _, x := range rs {
		public := x.mws == base
		oapi := braceRe.ReplaceAllString(x.path, "{$1}")
		n, documented := apiNotes[x.method+" "+x.path]
		if n.Summary == "" {
			n.Summary = summaryOf(x.method, x.path)
		}
		tag := tagOf(x.path)
		tags[tag] = true
		desc := n.Description
		if n.Access != "" {
			if desc != "" {
				desc += "\n\n"
			}
			desc += "Who can call it: " + n.Access
		} else if public {
			desc += "\n\nWho can call it: anyone, no token needed."
		}
		status := n.Status
		if status == 0 {
			status = 200
		}
		ok := map[string]any{"description": "Success."}
		if len(n.Response) > 0 && string(n.Response) != "null" {
			ok["content"] = map[string]any{"application/json": map[string]any{"schema": schemaOf(n.Response)}}
		} else if status != 204 {
			ok["description"] = "Success. No body."
		}
		resp := map[string]any{itoa(status): ok}
		for code, msg := range generic {
			if code == "401" && public || code == "403" && public {
				continue
			}
			resp[code] = map[string]any{"description": msg}
		}
		for code, msg := range n.Errors {
			resp[code] = map[string]any{"description": msg}
		}
		op := map[string]any{"summary": n.Summary, "description": desc, "tags": []string{tag}, "x-documented": documented,
			"operationId": strings.ToLower(x.method) + strings.NewReplacer("/", "_", "{", "", "}", "").Replace(x.path), "responses": resp}
		if public {
			op["security"] = []any{}
		}
		var params []any
		for _, m := range braceRe.FindAllStringSubmatch(x.path, -1) {
			params = append(params, map[string]any{"name": m[1], "in": "path", "required": true, "schema": map[string]any{"type": "string"}})
		}
		qn := make([]string, 0, len(n.Query))
		for q := range n.Query {
			qn = append(qn, q)
		}
		sort.Strings(qn)
		for _, q := range qn {
			params = append(params, map[string]any{"name": q, "in": "query", "description": n.Query[q], "schema": map[string]any{"type": "string"}})
		}
		if len(params) > 0 {
			op["parameters"] = params
		}
		if len(n.Request) > 0 && string(n.Request) != "null" {
			op["requestBody"] = map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schemaOf(n.Request)}}}
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
			"description": "For anyone integrating with Hub. Every route is under /v1 and speaks JSON over HTTPS. Send `Authorization: Bearer <access token>` (get one from POST /v1/auth/login); routes that work inside an organization need a token issued for that organization (pass `org` at sign-in). Errors are `{\"error\":{\"code\",\"message\"}}` with the HTTP status. Ids are UUIDs and times are RFC 3339 UTC. Message, file, note and ticket-thread contents are end-to-end encrypted on the person's device: the API stores and returns sealed envelopes (base64) it cannot read, so an integration reads or writes them only with a device key, and works with the plain fields (ids, names, times, counts, statuses) otherwise."},
		"servers":    []any{map[string]any{"url": "/"}},
		"tags":       tl,
		"paths":      paths,
		"components": map[string]any{"securitySchemes": map[string]any{"bearerAuth": map[string]any{"type": "http", "scheme": "bearer", "bearerFormat": "Hub access token"}}},
		"security":   []any{map[string]any{"bearerAuth": []any{}}},
	}
}

func itoa(n int) string { b, _ := json.Marshal(n); return string(b) }

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
	// the same pages under /api/v1/docs, for people who expect the /api prefix
	r.Get("/api/v1/docs", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/v1/docs", http.StatusFound) })
	r.Get("/api/v1/docs/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/v1/docs/openapi.json", http.StatusFound)
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
*{box-sizing:border-box}body{margin:0;background:var(--bg);color:var(--fg);font:15px/1.55 system-ui,sans-serif}
header,main{max-width:1040px;margin:0 auto;padding:16px}h1{margin:8px 0 4px;font-size:26px}p.lead{color:var(--mut);margin:0 0 12px}
input{width:100%;height:40px;padding:0 12px;border-radius:10px;border:1px solid var(--line);background:var(--card);color:var(--fg);font:inherit}
h2{font-size:15px;text-transform:uppercase;letter-spacing:.06em;color:var(--mut);margin:24px 0 8px}
details{border:1px solid var(--line);border-radius:10px;background:var(--card);margin:6px 0}summary{cursor:pointer;padding:9px 12px;display:flex;gap:10px;align-items:center;flex-wrap:wrap}
.m{font:700 11px monospace;min-width:54px;text-align:center;padding:3px 6px;border-radius:6px;background:var(--acc);color:#0a100d}.m.GET{background:#5aa9e6}.m.DELETE{background:#e5484d;color:#fff}.m.PUT,.m.PATCH{background:#ffb86b}
code{font:13px/1.4 ui-monospace,monospace}.b{padding:0 14px 14px}.s{color:var(--mut);flex:1;min-width:140px}.pub{font-size:11px;color:var(--mut);border:1px solid var(--line);border-radius:10px;padding:1px 7px}
h3{font-size:13px;margin:14px 0 4px;text-transform:uppercase;letter-spacing:.05em;color:var(--mut)}
table{border-collapse:collapse;width:100%;font-size:13px}td{padding:5px 8px;border-top:1px solid var(--line);vertical-align:top}td:first-child{white-space:nowrap;font-family:ui-monospace,monospace}td.t{color:var(--acc);white-space:nowrap}
.d{white-space:pre-line}
</style></head><body><header><h1>Hub API reference</h1><p class="lead" id="lead"></p><input id="q" type="search" placeholder="Find a route, e.g. calls or /v1/tickets" aria-label="Find a route"><p><a id="raw" href="/v1/docs/openapi.json" style="color:var(--acc)">openapi.json</a> (OpenAPI 3) for code generators and tools</p></header><main id="out">Loading…</main>
<script>
const esc=s=>String(s==null?'':s).replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]));
const rows=(sc,ind,out)=>{if(!sc)return;
 if(sc.type==='array'){out.push('<tr><td style="padding-left:'+(8+ind*16)+'px">[ ]</td><td class="t">array</td><td>list of:</td></tr>');rows(sc.items,ind+1,out);return}
 for(const [k,v] of Object.entries(sc.properties||{})){const nested=v.type==='object'&&v.properties||v.type==='array';
  out.push('<tr><td style="padding-left:'+(8+ind*16)+'px">'+esc(k)+'</td><td class="t">'+esc(v.type||'')+'</td><td class="d">'+esc(nested&&!v.description?'':v.description||'')+'</td></tr>');
  if(nested)rows(v,ind+1,out)}};
const table=sc=>{const o=[];rows(sc,0,o);return o.length?'<table>'+o.join('')+'</table>':''};
fetch('/v1/docs/openapi.json').then(r=>r.json()).then(spec=>{
 document.getElementById('lead').textContent=spec.info.description;
 const by={};
 for(const [p,ms] of Object.entries(spec.paths))for(const [m,op] of Object.entries(ms))(by[op.tags[0]]=by[op.tags[0]]||[]).push({p,m:m.toUpperCase(),op});
 const out=document.getElementById('out');
 const draw=f=>{out.innerHTML=Object.keys(by).sort().map(t=>{const rs=by[t].filter(x=>!f||(x.m+' '+x.p+' '+x.op.summary).toLowerCase().includes(f));
  return rs.length?'<h2>'+esc(t)+'</h2>'+rs.map(x=>{const op=x.op,q=(op.parameters||[]).filter(a=>a.in==='query'),pp=(op.parameters||[]).filter(a=>a.in==='path');
   const rq=op.requestBody&&op.requestBody.content['application/json'].schema;
   const ok=Object.entries(op.responses).find(([c])=>c[0]==='2'),rsc=ok&&ok[1].content&&ok[1].content['application/json'].schema;
   const errs=Object.entries(op.responses).filter(([c])=>c[0]!=='2').map(([c,v])=>'<tr><td>'+c+'</td><td class="d" colspan="2">'+esc(v.description)+'</td></tr>').join('');
   return '<details><summary><span class="m '+x.m+'">'+x.m+'</span><code>'+esc(x.p)+'</code><span class="s">'+esc(op.summary)+'</span>'+(op.security&&!op.security.length?'<span class="pub">no token</span>':'')+'</summary><div class="b"><p class="d">'+esc(op.description)+'</p>'
    +(pp.length?'<h3>In the address</h3><table>'+pp.map(a=>'<tr><td>'+esc(a.name)+'</td><td class="t">string</td><td>'+esc(a.description||'The '+a.name+' from the route')+'</td></tr>').join('')+'</table>':'')
    +(q.length?'<h3>Query</h3><table>'+q.map(a=>'<tr><td>'+esc(a.name)+'</td><td class="t"></td><td class="d">'+esc(a.description)+'</td></tr>').join('')+'</table>':'')
    +(rq?'<h3>Request body (JSON)</h3>'+table(rq):'')
    +'<h3>Response '+(ok?ok[0]:'')+'</h3>'+(rsc?table(rsc):'<p>'+esc(ok?ok[1].description:'')+'</p>')
    +'<h3>Errors</h3><table>'+errs+'</table></div></details>'}).join(''):''}).join('')||'Nothing matches.'};
 draw('');document.getElementById('q').addEventListener('input',e=>draw(e.target.value.trim().toLowerCase()));
}).catch(()=>{document.getElementById('out').textContent='Could not load the reference.'});
</script></body></html>`
