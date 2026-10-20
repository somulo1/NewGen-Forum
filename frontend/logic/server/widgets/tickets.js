// Tickets: the Tickets widget, My tickets, the Ticket desk and each ticket's page. A ticket is a conversation (its title,
// description and comments are sealed for the people on it) plus plain operational facts the server acts on (category,
// priority, step, due time). The clock (TAT) runs on the server; the frontend shows what it reports.
import { newConversationKey, wrap, b64, unb64 } from '../e2ee.js';
import * as files from '../files.js';

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

async function mapTicket(srv, t) {
  let detail = t, msgs = [];
  try { detail = await srv.request('GET', `/tickets/${t.id}`); } catch (e) { /* keep the list row */ }
  if (t.participant !== false) try { msgs = (await srv.request('GET', `/conversations/${t.id}/messages?limit=100`)).messages; } catch (e) { /* not a participant: the text stays sealed */ }
  const opened = await Promise.all(msgs.map(async (m) => ({ m, pl: await srv.vault.openMessage(m) })));
  opened.sort((a, b) => (a.m.id < b.m.id ? -1 : 1));
  const first = opened.length ? opened[0].pl : null;
  const activity = [];
  (detail.events || []).forEach((e) => activity.push({ who: me(srv, e.who), kind: e.kind || 'event', text: e.text, time: when(e.at), at: e.at }));
  opened.slice(1).forEach(({ m, pl }) => {
    if (m.deleted || !pl) return;
    if (pl.text !== undefined) activity.push({ who: me(srv, m.sender_name), kind: 'comment', text: pl.text, time: when(m.created_at), at: m.created_at, mid: m.id });
    else if (pl.name) activity.push({ who: me(srv, m.sender_name), kind: 'file', text: pl.name, meta: pl.meta || '', isImg: !!pl.isImg, url: '', ref: pl.fileId ? { fileId: pl.fileId, key: pl.key, type: pl.type || '', isImg: !!pl.isImg, isVideo: !!pl.isVideo } : null, time: when(m.created_at), at: m.created_at, mid: m.id });
  });
  activity.sort((a, b) => (a.at < b.at ? -1 : 1));
  const toType = TYPE[t.to_type] || 'desk';
  const toName = t.to_ref ? srv.unitName.get(t.to_ref) || '' : '';
  const raiser = srv.dirByUid && srv.dirByUid.get(t.raised_by);
  const names = (t.assignees || []).map((id) => srv.nameOf.get(id)).filter(Boolean);
  return {
    id: t.ref, uid: t.id, title: first ? first.title : t.category + ' ticket', desc: first ? first.desc || '' : '(sealed: you are not a participant)', cat: t.category, prio: t.priority,
    step: t.step, hold: !!t.hold, late: !!t.late, raisedBy: me(srv, t.raised_by_name), assignees: names, handler: t.handler_name || '', handlerSince: t.handler_since ? when(t.handler_since) : '',
    created: when(t.created_at), elapsed: detail.elapsed_hours || 0, tatH: t.tat_hours, sync: t.sync || 'Not synced', activity, dept: toType === 'departments' ? toName : (raiser ? raiser.dept : ''), fromDept: raiser ? raiser.dept : '',
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
  categories.filter((c) => c.unit_type !== 'desk').forEach((c) => {
    const key = TYPE[c.unit_type] + ':' + (srv.unitName.get(c.unit_ref) || '');
    (unitCats[key] = unitCats[key] || []).push({ id: c.id, name: c.name, tat: c.tat_hours, prio: c.priority, members: (c.members || []).map((id) => srv.nameOf.get(id)).filter(Boolean) });
  });
  srv.ticketCats = categories;
  logic.setState((st) => ({ tset: { ...st.tset, cats, start: cfg.start, end: cfg.end, pause: cfg.pause, notify: cfg.notify, nearPct: cfg.nearPct, escalate: cfg.escalate }, unitCats }));
}
const hm = (v) => { const p = String(v || '0:0').split(':'); return (Number(p[0]) || 0) * 60 + (Number(p[1]) || 0); };

// Attachments are encrypted files: fetch and open each one, then show it on its ticket.
function showFiles(logic, srv, rows) {
  rows.forEach((t) => (t.activity || []).filter((a) => a.ref && !a.url).forEach((a) => {
    files.openUrl(srv, a.ref).then((url) => {
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
  showFiles(logic, srv, rows);
  watchSettings(logic, srv);
}

export async function reload(logic, srv, uid) {
  const t = await srv.request('GET', `/tickets/${uid}`);
  const row = await mapTicket(srv, t);
  logic.setState((st) => ({ tickets: st.tickets.some((x) => x.uid === uid) ? st.tickets.map((x) => (x.uid === uid ? row : x)) : [row].concat(st.tickets) }));
  showFiles(logic, srv, [row]);
  return row;
}

// ---- actions ----

export async function act(logic, srv, t, body) {
  await srv.request('POST', `/tickets/${t.uid}/act`, body);
  return reload(logic, srv, t.uid);
}
export const setStep = (logic, srv, t, step) => act(logic, srv, t, { do: 'step', step });
export const hold = (logic, srv, t) => act(logic, srv, t, { do: 'hold' });
export const resume = (logic, srv, t) => act(logic, srv, t, { do: 'resume' });

export async function comment(logic, srv, t, text) {
  const sealed = await srv.vault.seal(t.uid, { text });
  await srv.request('POST', `/conversations/${t.uid}/messages`, { ...sealed, type: 'text' });
  return reload(logic, srv, t.uid);
}

// Gives a person a key to the ticket's conversation if they are not on it yet.
async function ensureMember(srv, t, pid) {
  const conv = await srv.vault.conversation(t.uid, true);
  if ((conv.members || []).some((m) => m.person === pid)) return;
  const key = await srv.vault.keyFor(t.uid, conv.epoch);
  const { devices } = await srv.request('POST', '/e2ee/people/devices', { people: [pid] });
  const keys = await srv.vault.wrapFor(t.uid, conv.epoch, key, devices);
  await srv.request('POST', `/conversations/${t.uid}/members`, { people: [pid], device: srv.vault.deviceId, keys });
  srv.vault.convs.delete(t.uid);
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
export async function escalate(logic, srv, t, e) {
  const toType = TYPE_BACK[e.type] || 'desk';
  const toRef = toType === 'desk' ? undefined : srv.structure[e.type].get(e.name);
  const pid = e.person ? personId(srv, e.person) : undefined;
  if (e.person && !pid) throw new Error('Choose someone from the directory.');
  if (pid) await ensureMember(srv, t, pid);
  return act(logic, srv, t, { do: 'escalate', to_type: toType, to_ref: toRef, person: pid, category: e.cat || undefined, why: e.why || undefined });
}

// Attaches files (evidence, screenshots) to a ticket: each is encrypted on this device and shared with the people on it.
export async function attach(logic, srv, t, list) {
  for (const f of list) {
    const up = await files.upload(srv, t.uid, f);
    const sealed = await srv.vault.seal(t.uid, { name: up.name, meta: up.meta, type: up.type, isImg: up.isImg, isVideo: up.isVideo, fileId: up.fileId, key: up.key });
    await srv.request('POST', `/conversations/${t.uid}/messages`, { ...sealed, type: 'text' });
  }
  return reload(logic, srv, t.uid);
}

// nt = { title, desc }, cat = { name, prio }, target = { type: 'desk' | 'departments' ... , name }
export async function raise(logic, srv, nt, cat, target) {
  const toType = TYPE_BACK[target.type] || 'desk';
  const toRef = toType === 'desk' ? undefined : srv.structure[target.type].get(target.name);
  const cats = srv.ticketCats || [];
  const server = cats.find((c) => c.name === cat.name && ((c.unit_type === toType && c.unit_ref === toRef) || c.unit_type === 'desk'));
  const members = [...new Set((server && server.members) || [])];
  const { devices } = await srv.request('POST', '/e2ee/people/devices', { people: [...new Set([srv.me.person, ...members])] });
  const key = newConversationKey();
  const keys = await Promise.all(devices.map(async (d) => ({ epoch: 1, device: d.id, wrapped_key: b64(await wrap(key, unb64(d.identity_key))) })));
  const conv = await srv.request('POST', '/conversations', { kind: 'ticket', members: members.filter((m) => m !== srv.me.person), keys });
  const sealed = await srv.vault.seal(conv.id, { title: nt.title, desc: nt.desc || '', category: cat.name });
  await srv.request('POST', `/conversations/${conv.id}/messages`, { ...sealed, type: 'text' });
  for (const f of (nt.files || []).map((x) => x.file).filter(Boolean)) {
    const up = await files.upload(srv, conv.id, f);
    const fs = await srv.vault.seal(conv.id, { name: up.name, meta: up.meta, type: up.type, isImg: up.isImg, isVideo: up.isVideo, fileId: up.fileId, key: up.key });
    await srv.request('POST', `/conversations/${conv.id}/messages`, { ...fs, type: 'text' });
  }
  await srv.request('POST', '/tickets', { conversation: conv.id, category: cat.name, priority: cat.prio || nt.prio, to_type: toType, to_ref: toRef });
  return reload(logic, srv, conv.id);
}

// ---- settings: a unit's own categories (the "Ticket categories" table) ----
export async function saveUnitCats(logic, srv, key, rows) {
  const [level, ...rest] = key.split(':');
  const unitName = rest.join(':');
  const unitType = TYPE_BACK[level] || 'desk';
  const unitRef = unitType === 'desk' ? null : srv.structure[level].get(unitName);
  const existing = (srv.ticketCats || []).filter((c) => c.unit_type === unitType && (c.unit_ref || null) === unitRef);
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
