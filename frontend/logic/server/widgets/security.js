// Security & logs: alerts raised by the detection rules, the audit log, and the rules themselves (on/off and severity).
// The server writes every log entry and raises alerts as things happen; this reads them and records alert decisions.
export const name = 'security';

const minsAgo = (iso) => Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 60000));
// the frontend says "False positive", the server "False alarm"
const toServer = (st) => (st === 'False positive' ? 'False alarm' : st);
const fromServer = (st) => (st === 'False alarm' ? 'False positive' : st);

export async function load(logic, srv) {
  const can = (r) => ((srv.me.rights || {}).security || []).indexOf(r) !== -1;
  const [rules, alerts, logs] = await Promise.all([
    srv.request('GET', '/security/rules').catch(() => ({ rules: [] })),
    can('View alerts') ? srv.request('GET', '/security/alerts') : Promise.resolve({ alerts: [] }),
    can('View logs') ? srv.request('GET', '/security/logs?limit=500') : Promise.resolve({ logs: [] })
  ]);
  const rulesOn = {}, ruleSev = {};
  rules.rules.forEach((r) => { rulesOn[r.id] = !!r.enabled; ruleSev[r.id] = r.sev || r.severity; });
  const alertSt = {}, alertNotes = {};
  const alertsLive = alerts.alerts.map((a) => {
    alertSt[a.ref] = fromServer(a.status);
    alertNotes[a.ref] = (a.notes || []).map((n) => ({ status: fromServer(n.status), comment: n.comment, who: n.who, at: new Date(n.at).getTime() }));
    return { id: a.ref, uid: a.id, rule: a.rule, ago: minsAgo(a.created_at), title: a.title, detail: a.detail };
  });
  logic.baseLogs = logs.logs.map((l) => ({ id: l.id, ago: minsAgo(l.at), user: l.user, action: l.action, module: l.module, ip: l.ip, loc: l.location || '', result: l.result, rule: l.rule || null }));
  srv.ruleDefs = rules.rules.map((r) => ({ id: r.id, name: r.name, desc: r.desc, sev: r.severity }));
  logic.setState({ rulesOn, ruleSev, alertSt, alertNotes, alertsLive, secRuleDefs: srv.ruleDefs, extraLogs: [] });
  if (!srv.securityLive && srv.rt) {
    srv.securityLive = true;
    // a rule just fired (the server notifies the security team) or someone changed something: refresh
    srv.rt.on('notification', (ev) => { const k = (ev.data || {}).kind; if (k === 'alert' || k === 'security') load(logic, srv).catch(() => {}); });
  }
}

export async function setAlert(logic, srv, ref, status, comment) {
  const a = (logic.state.alertsLive || []).find((x) => x.id === ref);
  if (!a) throw new Error('That alert is no longer there.');
  await srv.request('POST', `/security/alerts/${a.uid}/status`, { status: toServer(status), comment });
  await load(logic, srv);
}

export async function setRule(logic, srv, id, enabled, severity) {
  await srv.request('PUT', `/security/rules/${id}`, { enabled, severity });
  await load(logic, srv);
}
