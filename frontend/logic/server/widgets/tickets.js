// Tickets: the Tickets widget, My tickets, the Ticket desk and each ticket's page. Tickets are not end-to-end encrypted: the
// title, description, comments and attachments are kept by the server and shown only to the people on the ticket, next to the
// operational facts it acts on (category, priority, step, due time). Opening a ticket is one request with nothing to decrypt.
// The ticket's conversation only lists the people on it. Very old tickets whose text or files were sealed in that conversation
// are still opened from there, and only when the server has nothing for them. The clock (TAT) runs on the server.
import * as files from '../files.js';
import { ensure as ensureDashboard } from './org.js';

export const name = 'tickets';

const TYPE = { desk: 'desk', department: 'departments', unit: 'units', branch: 'branches', sector: 'sectors' };
const TYPE_BACK = { desk: 'desk', departments: 'department', units: 'unit', branches: 'branch', sectors: 'sector' };

export function when(iso) {
  const d = new Date(iso);
  const tz = { timeZone: 'Africa/Nairobi' };
  const day = (x) => x.toLocaleDateString('en-CA', tz);
  const hm = d.toLocaleTimeString('en-GB', { ...tz, hour: '2-digit', minute: '2-digit', hour12: false });
  const today = day(new Date());
  const that = day(d);
  if (that === today) return 'Today, ' + hm;
  if (that === day(new Date(Date.now() - 86400000))) return 'Yesterday, ' + hm;
  const days = Math.round((new Date(today) - new Date(that)) / 86400000);
  if (days < 7) return d.toLocaleDateString('en-GB', { ...tz, weekday: 'long' }) + ', ' + hm;
  return d.toLocaleDateString('en-GB', { ...tz, day: 'numeric', month: 'short' }) + ', ' + hm;
}

const me = (srv, name) => (name === srv.me.name ? 'Sam' : name);

// tickets raised before encryption was dropped: their text and files were sealed in the conversation
const LEGACY_BEFORE = '2026-11-15';
async function legacyMessages(srv, t, detail) {
  if (t.participant === false || (detail.title && t.created_at >= LEGACY_BEFORE)) return [];
  try {
    const msgs = (await srv.request('GET', `/conversations/${t.id}/messages?limit=100`)).messages;
    return Promise.all(msgs.map(async (m) => ({ m, pl: await srv.vault.openMessage(m) })));
  } catch (e) { return []; }
}

async function mapTicket(srv, t, preloaded) {
  let detail = preloaded || t;
  if (!preloaded) try { detail = await srv.request('GET', `/tickets/${t.id}`); } catch (e) { /* keep the list row */ }
  const opened = await legacyMessages(srv, t, detail);
  opened.sort((a, b) => (a.m.id < b.m.id ? -1 : 1));
  const sealedHead = opened.find(({ pl }) => pl && pl.title !== undefined);
  const first = detail.title ? { title: detail.title, desc: detail.description || '' } : sealedHead ? sealedHead.pl : null;
  const activity = [];
  (detail.events || []).forEach((e) => activity.push({ who: me(srv, e.who), kind: e.kind || 'event', text: e.text, time: when(e.at), at: e.at }));
  (detail.comments || []).forEach((c) => activity.push({ who: me(srv, c.who), kind: 'comment', text: c.text, time: when(c.at), at: c.at, mid: c.id }));
  opened.forEach(({ m, pl }) => {
    if (m.deleted || !pl || pl.title !== undefined) return;
    if (pl.text !== undefined) activity.push({ who: me(srv, m.sender_name), kind: 'comment', text: pl.text, time: when(m.created_at), at: m.created_at, mid: m.id });
    else if (pl.name) activity.push({ who: me(srv, m.sender_name), kind: 'file', text: pl.name, meta: pl.meta || '', isImg: !!pl.isImg, url: '', ref: pl.fileId ? { fileId: pl.fileId, key: pl.key, type: pl.type || '', isImg: !!pl.isImg, isVideo: !!pl.isVideo } : null, time: when(m.created_at), at: m.created_at, mid: m.id });
  });
  (detail.files || []).forEach((f) => activity.push({ who: me(srv, f.by), kind: 'file', text: f.name, meta: fmtSize(f.size), isImg: /^image\//.test(f.type), url: '', plain: { ticket: t.id, id: f.id, type: f.type }, time: when(f.created_at), at: f.created_at, mid: f.id }));
  activity.sort((a, b) => (a.at < b.at ? -1 : 1));
  const toType = TYPE[t.to_type] || 'desk';
  const toName = t.to_ref ? srv.unitName.get(t.to_ref) || '' : '';
  const raiser = srv.dirByUid && srv.dirByUid.get(t.raised_by);
  const names = (t.assignees || []).map((id) => srv.nameOf.get(id)).filter(Boolean);
  // the heads of the dashboard it went to and of the one it came from: the dashboard it was sent from, else the raiser's own place
  // (the deepest of their sector, branch, unit, department), at any level
  const headsOf = (ref) => (ref ? (((srv.unitList || []).find((u) => u.id === ref) || {}).heads || []).map((h) => me(srv, srv.nameOf.get(h.id) || h.name || '')).filter(Boolean) : []);
  const fromRef = detail.origin_ref || t.from_ref || null;   // the server works out the sending dashboard (sector, branch, unit or department)
  const toHeads = headsOf(t.to_ref), fromHeads = headsOf(fromRef);
  return {
    id: t.ref, uid: t.id, title: first ? first.title : t.category + ' ticket', desc: first ? first.desc || '' : '(hidden: you are not a participant)', cat: t.category, prio: t.priority,
    step: t.step, hold: !!t.hold, late: !!t.late, raisedBy: me(srv, t.raised_by_name), assignees: names, handler: t.handler_name || '', handlerSince: t.handler_since ? when(t.handler_since) : '',
    rating: t.rating || 0, autoClosed: !!t.auto_closed, closedAt: t.closed_at ? when(t.closed_at) : '', canAssign: !!detail.can_assign, canEscalate: !!detail.can_escalate, canWork: !!detail.can_work, kind: t.kind || 'ticket', toHeads, fromHeads, canAck: !!detail.can_acknowledge, canChase: !!detail.can_chase, acks: (detail.acknowledgements || t.acknowledgements || []).map((a) => ({ name: a.name, at: when(a.at) })), created: when(t.created_at), elapsed: detail.elapsed_hours || 0, tatH: t.tat_hours, sync: t.sync || 'Not synced', activity, dept: toType === 'departments' ? toName : (raiser ? raiser.dept : ''), fromDept: raiser ? raiser.dept : '',
    createdAt: t.created_at.slice(0, 10), escalated: !!t.escalated, to: toType === 'desk' ? { type: 'desk', name: 'Ticket desk' } : { type: toType, name: toName }, dueAt: t.due_at || '', unsealed: !first
  };
}

// The organization's timing policy and categories: desk categories are the defaults, the rest belong to one unit.
async function loadSettings(logic, srv) {
  const cfg = await srv.request('GET', '/tickets/settings');
  const { categories } = await srv.request('GET', '/tickets/categories?to_type=all');
  const dayH = Math.max(1, (hm(cfg.end) - hm(cfg.start)) / 60);
  const cats = categories.filter((c) => c.unit_type === 'desk').map((c) => (c.tat_hours % dayH === 0 ? { name: c.name, tat: c.tat_hours / dayH, unit: 'days' } : { name: c.name, tat: c.tat_hours, unit: 'hours' }));
  const unitCats = {};
  // a department's business request turnaround is kept with its categories but is not a ticket category
  const requestTat = {}, commTat = {};
  categories.filter((c) => c.unit_type !== 'desk').forEach((c) => {
    const key = TYPE[c.unit_type] + ':' + (srv.unitName.get(c.unit_ref) || '');
    if (c.name === 'Business request') { requestTat[key] = c.tat_hours; return; }
    if (c.name === 'Communication') { commTat[key] = c.tat_hours; return; }
    (unitCats[key] = unitCats[key] || []).push({ id: c.id, name: c.name, tat: c.tat_hours, prio: c.priority, members: (c.members || []).map((id) => srv.nameOf.get(id)).filter(Boolean) });
  });
  srv.ticketCats = categories;
  // places that take tickets only by escalation from places they chose: { [unit id]: [chosen place ids] }
  let intake = {};
  try { (await srv.request('GET', '/tickets/intake')).intake.forEach((x) => { (intake[x.unit] = intake[x.unit] || []).push(x.only_from); }); } catch (e) { intake = {}; }
  logic.setState((st) => ({ ticketIntake: intake, tset: { ...st.tset, cats, start: cfg.start, end: cfg.end, pause: cfg.pause, notify: cfg.notify, nearPct: cfg.nearPct, escalate: cfg.escalate }, unitCats, requestTat, commTat }));
}
const hm = (v) => { const p = String(v || '0:0').split(':'); return (Number(p[0]) || 0) * 60 + (Number(p[1]) || 0); };

const fmtSize = (n) => (n >= 1048576 ? (n / 1048576).toFixed(1) + ' MB' : Math.max(1, Math.round(n / 1024)) + ' KB');

// an attachment: downloaded once per session and shown from a local address
const plainUrls = new Map();
function plainUrl(srv, f) {
  if (!plainUrls.has(f.id)) {
    plainUrls.set(f.id, srv.bytes('GET', `/tickets/${f.ticket}/files/${f.id}`).then((b) => URL.createObjectURL(new Blob([b], { type: f.type || 'application/octet-stream' })))
      .catch((e) => { plainUrls.delete(f.id); throw e; }));
  }
  return plainUrls.get(f.id);
}

// Fetches each attachment, then shows it on its ticket (old sealed ones are opened as before).
function showFiles(logic, srv, rows) {
  return rows.flatMap((t) => (t.activity || []).filter((a) => (a.ref || a.plain) && !a.url).map((a) => {
    return (a.plain ? plainUrl(srv, a.plain) : files.openUrl(srv, a.ref)).then((url) => {
      logic.setState((st) => ({ tickets: st.tickets.map((x) => (x.uid !== t.uid ? x : { ...x, activity: x.activity.map((y) => (y.mid === a.mid ? { ...y, url } : y)) })) }));
    }).catch(() => { /* the file is gone or sealed for someone else */ });
  }));
}

export async function load(logic, srv) {
  await loadSettings(logic, srv);
  const { tickets } = await srv.request('GET', '/tickets?limit=100');
  const rows = [];
  for (let i = 0; i < tickets.length; i += 10) rows.push(...await Promise.all(tickets.slice(i, i + 10).map((t) => mapTicket(srv, t))));
  logic.setState({ tickets: rows });
  routeUnassigned(logic, srv, tickets).catch(() => {});   // quietly, after the list is shown
  await files.settle(showFiles(logic, srv, rows));   // the attachments of the newest tickets are in before tickets count as loaded
  watchSettings(logic, srv);
}

// After a ticket or business request is sent or changes: the dashboards show it at once. The one open now is fetched again
// straight away; every other one is fetched again the next time it is opened.
function refreshDashboards(logic, srv) {
  if (srv.ouAsked) srv.ouAsked.clear();
  const open = logic.state.section === 'orgunit' && logic.state.ou;
  if (open) try { ensureDashboard(logic, srv, open, true); } catch (e) { /* the dashboard catches up on its own */ }
}

export async function reload(logic, srv, uid) {
  const t = await srv.request('GET', `/tickets/${uid}`);
  const row = await mapTicket(srv, t);
  logic.setState((st) => ({ tickets: st.tickets.some((x) => x.uid === uid) ? st.tickets.map((x) => (x.uid === uid ? row : x)) : [row].concat(st.tickets) }));
  refreshDashboards(logic, srv);
  showFiles(logic, srv, [row]);
  return row;
}

// ---- actions ----

export async function act(logic, srv, t, body) {
  // an email about this action may say what the ticket is about (never the files); the server writes it into the email and does not keep it
  const readable = t.title && t.desc && !t.unsealed;
  const details = srv.emailDetails !== false && readable ? { title: t.title, description: t.desc, files: (t.activity || []).filter((x) => x.kind === 'file').length } : undefined;
  await srv.request('POST', `/tickets/${t.uid}/act`, details ? { ...body, details } : body);
  return reload(logic, srv, t.uid);
}
// resolving needs a comment on what was done; closing may carry 1 to 5 stars and a remark
export const setStep = (logic, srv, t, step, extra) => act(logic, srv, t, { do: 'step', step, ...(extra || {}) });
export const hold = (logic, srv, t) => act(logic, srv, t, { do: 'hold' });
export const resume = (logic, srv, t) => act(logic, srv, t, { do: 'resume' });

export async function comment(logic, srv, t, text) {
  await srv.request('POST', `/tickets/${t.uid}/comments`, { text });
  return reload(logic, srv, t.uid);
}

// Puts a person on the ticket if they are not on it yet (no keys: tickets are not end-to-end encrypted).
async function ensureMember(srv, t, pid) {
  const { members } = await srv.request('GET', `/conversations/${t.uid}`).catch(() => ({ members: null }));
  if (members && members.some((m) => m.person === pid)) return;
  await srv.request('POST', `/conversations/${t.uid}/members`, { people: [pid] });
}
// The categories a place sets on its dashboard decide who handles a ticket. A ticket raised before its category had people (or a
// business request or communication, which goes to the place's heads) is given to them here: someone on it adds them (they hold
// its key) and the server picks the one with the fewest open tickets. Each ticket is tried once per visit.
const TB2 = { department: 'department', unit: 'unit', branch: 'branch', sector: 'sector', desk: 'desk' };
async function routeUnassigned(logic, srv, list) {
  srv.routed = srv.routed || new Set();
  const cats = srv.ticketCats || [];
  for (const t of list) {
    if (t.handler || t.step >= 3 || t.participant === false || srv.routed.has(t.id)) continue;
    srv.routed.add(t.id);
    const kind = t.kind || 'ticket';
    let people = [];
    if (kind === 'ticket') {
      const c = cats.find((x) => x.name === t.category && x.unit_type === TB2[t.to_type] && (x.unit_ref || null) === (t.to_ref || null)) || cats.find((x) => x.name === t.category && x.unit_type === 'desk');
      people = (c && c.members) || [];
    }
    if (!people.length && t.to_ref) people = (((srv.unitList || []).find((u) => u.id === t.to_ref) || {}).heads || []).map((h) => h.id);
    people = [...new Set(people)].filter((pid) => pid !== t.raised_by);
    if (!people.length) continue;
    try {
      for (const pid of people) await ensureMember(srv, { uid: t.id }, pid);
      const out = await srv.request('POST', `/tickets/${t.id}/route`);
      if (out && out.handler) await reload(logic, srv, t.id);
    } catch (e) { /* not allowed to add people here, or nobody suitable: it stays for a head to assign */ }
  }
}

const personId = (srv, name) => (name === 'Sam' || name === srv.me.name ? srv.me.person : srv.personOf.get(name));

// Gives a person a key to the ticket's conversation (if they are not on it yet), then makes them the handler.
export async function assign(logic, srv, t, personName) {
  const pid = personId(srv, personName);
  if (!pid) throw new Error('Choose someone from the directory.');
  await ensureMember(srv, t, pid);
  return act(logic, srv, t, { do: 'assign', handler: pid });
}

// Moves a ticket to another sector, branch, unit, department or the desk, under one of that place's categories, and hands it
// to a named person there. e = { type: 'departments' | ... | 'desk', name, cat, person, why }
// someone at the receiving place says they received a business request or communication
export async function acknowledge(logic, srv, t) {
  await srv.request('POST', `/tickets/${t.uid}/acknowledge`);
  return reload(logic, srv, t.uid);
}

// An overdue ticket goes, with a note, to the heads of the place it was sent to. It stays where it is.
export async function chase(logic, srv, t, note) {
  await srv.request('POST', `/tickets/${t.uid}/chase`, { note });
  return reload(logic, srv, t.uid);
}

export async function escalate(logic, srv, t, e) {
  const toType = TYPE_BACK[e.type] || 'desk';
  const toRef = toType === 'desk' ? undefined : srv.structure[e.type].get(e.name);
  const pid = e.person ? personId(srv, e.person) : undefined;
  if (e.person && !pid) throw new Error('Choose someone from the directory.');
  if (pid) await ensureMember(srv, t, pid);
  return act(logic, srv, t, { do: 'escalate', to_type: toType, to_ref: toRef, person: pid, category: e.cat || undefined, why: e.why || undefined });
}

// Attaches files (evidence, screenshots) to a ticket: sent as they are, for the people on it.
async function upload(srv, uid, f) {
  const q = new URLSearchParams({ name: f.name, type: f.type || '' });
  await srv.bytes('PUT', `/tickets/${uid}/files?${q}`, new Uint8Array(await f.arrayBuffer()));
}
export async function attach(logic, srv, t, list) {
  for (const f of list) await upload(srv, t.uid, f);
  return reload(logic, srv, t.uid);
}

// nt = { title, desc }, cat = { name, prio }, target = { type: 'desk' | 'departments' ... , name }
export async function raise(logic, srv, nt, cat, target) {
  const toType = TYPE_BACK[target.type] || 'desk';
  const toRef = toType === 'desk' ? undefined : srv.structure[target.type].get(target.name);
  const cats = srv.ticketCats || [];
  const server = cats.find((c) => c.name === cat.name && ((c.unit_type === toType && c.unit_ref === toRef) || c.unit_type === 'desk'));
  // the people who can read it: the category's people and, for a
  // business request or communication, the heads of the place it goes to, so they can open it and assign it
  const heads = toRef && (target.kind === 'request' || target.kind === 'communication') ? (((srv.unitList || []).find((u) => u.id === toRef) || {}).heads || []).map((h) => h.id) : [];
  const members = [...new Set(((server && server.members) || []).concat(heads))];
  // the conversation only lists the people on the ticket (no keys)
  const conv = await srv.request('POST', '/conversations', { kind: 'ticket', members: members.filter((m) => m !== srv.me.person) });
  await srv.request('POST', '/tickets', { conversation: conv.id, title: nt.title, description: nt.desc || '', category: cat.name, priority: cat.prio || nt.prio, to_type: toType, to_ref: toRef, kind: target.kind === 'request' || target.kind === 'communication' ? target.kind : 'ticket',
    // sent from a place's dashboard: on that place's behalf
    ...(target.from ? { from_ref: srv.structure[target.from.level].get(target.from.name) } : {}),
    // emails about this ticket may show these (only sent when the organization allows details in emails; files are never included)
    ...(srv.emailDetails !== false ? { details: { title: nt.title, description: nt.desc || '', files: (nt.files || []).length } } : {}) });
  for (const f of (nt.files || []).map((x) => x.file).filter(Boolean)) await upload(srv, conv.id, f);
  return reload(logic, srv, conv.id);
}

// A place chooses to take tickets only by escalation from the places listed (their ids), or from anyone (an empty list).
export async function setIntake(logic, srv, unitId, onlyFrom) {
  const list = [...new Set(onlyFrom || [])];
  await srv.request('PUT', '/tickets/intake', { unit: unitId, only_from: list });
  logic.setState((st) => { const m = { ...(st.ticketIntake || {}) }; if (list.length) m[unitId] = list; else delete m[unitId]; return { ticketIntake: m }; });
}

// ---- settings: a unit's own categories (the "Ticket categories" table) ----
// how long a department takes to deal with business requests (its head sets it), 1 hour to 3 months
export async function setRequestTat(logic, srv, deptName, hours) { return setKindTat(logic, srv, 'request', 'departments', deptName, hours); }
// kind: 'request' (departments) or 'communication' (any place); level: departments | units | branches | sectors
export async function setKindTat(logic, srv, kind, level, name, hours) {
  await srv.request('PUT', '/tickets/request-tat', { department: srv.structure[level].get(name), hours, kind });
  await load(logic, srv);
}

export async function saveUnitCats(logic, srv, key, rows) {
  const [level, ...rest] = key.split(':');
  const unitName = rest.join(':');
  const unitType = TYPE_BACK[level] || 'desk';
  const unitRef = unitType === 'desk' ? null : srv.structure[level].get(unitName);
  const existing = (srv.ticketCats || []).filter((c) => c.unit_type === unitType && (c.unit_ref || null) === unitRef && c.name !== 'Business request' && c.name !== 'Communication');
  const keep = new Set();
  for (const r of rows) {
    const body = { unit_type: unitType, unit_ref: unitRef || undefined, name: r.name, tat_hours: Math.max(1, Math.round(Number(r.tat) || 1)), priority: r.prio || 'Medium', members: (r.members || []).map((n) => (n === 'Sam' ? srv.me.person : srv.personOf.get(n))).filter(Boolean) };
    const hit = existing.find((c) => c.id === r.id) || existing.find((c) => c.name === r.name);
    if (hit) { keep.add(hit.id); await srv.request('PUT', `/tickets/categories/${hit.id}`, body); } else await srv.request('POST', '/tickets/categories', body);
  }
  for (const c of existing) if (!keep.has(c.id)) await srv.request('DELETE', `/tickets/categories/${c.id}`);
  await loadSettings(logic, srv);
}

// ---- settings: the timing policy and default categories (the TAT settings page saves as you edit) ----
export function watchSettings(logic, srv) {
  if (srv.watchingTset) return;
  srv.watchingTset = true;
  const pick = (t) => JSON.stringify({ start: t.start, end: t.end, pause: t.pause, notify: t.notify, nearPct: t.nearPct, escalate: t.escalate });
  const cats = (t) => JSON.stringify(t.cats);
  let saved = { cfg: pick(logic.state.tset), cats: cats(logic.state.tset) };
  let timer = null;
  const flush = async () => {
    const t = logic.state.tset;
    if (pick(t) !== saved.cfg) {
      saved.cfg = pick(t);
      try {
        const cur = await srv.request('GET', '/tickets/settings');
        await srv.request('PUT', '/tickets/settings', { ...cur, start: t.start, end: t.end, pause: t.pause, notify: t.notify, nearPct: t.nearPct, escalate: t.escalate });
      } catch (e) { logic.notify(e.message || 'Could not save the ticket settings.', '#FF9B8A'); }
    }
    if (cats(t) !== saved.cats) {
      saved.cats = cats(t);
      const dayH = Math.max(1, (hm(t.end) - hm(t.start)) / 60);
      try {
        await saveUnitCats(logic, srv, 'desk:Ticket desk', t.cats.map((c) => ({ name: c.name, tat: c.unit === 'days' ? c.tat * dayH : c.tat, prio: 'Medium', members: [] })));
        saved.cats = cats(logic.state.tset);
      } catch (e) { logic.notify(e.message || 'Could not save the categories.', '#FF9B8A'); }
    }
  };
  const orig = logic.setState.bind(logic);
  logic.setState = (update, done) => { orig(update, done); clearTimeout(timer); timer = setTimeout(flush, 900); };
}
