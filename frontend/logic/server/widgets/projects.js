// Projects: the Projects widget and page. A project is a conversation (its plan is sealed for the team: milestones,
// sprints, tasks, risks, activity are records) plus plain facts the server acts on (code, status, dates, budget, owner).
// The page edits state.projects directly; this module watches those edits and writes the differences to the server.
import { newConversationKey, wrap, newId, b64, unb64 } from '../e2ee.js';
import { normalizeProject } from '../../projects.js';

export const name = 'projects';

const COLS = ['todo', 'doing', 'review', 'done'];
const isUuid = (x) => typeof x === 'string' && /^[0-9a-f]{8}-[0-9a-f]{4}-7/.test(x);
const ago = (iso) => { const m = Math.round((Date.now() - new Date(iso)) / 60000); return m < 60 ? Math.max(1, m) + ' min ago' : m < 1440 ? Math.round(m / 60) + ' h ago' : Math.round(m / 1440) === 1 ? 'Yesterday' : Math.round(m / 1440) + ' days ago'; };
const nameOf = (srv, pid) => (pid === srv.me.person ? srv.me.name : srv.nameOf.get(pid) || '');

async function mapProject(srv, row) {
  const conv = await srv.vault.conversation(row.id, true);
  const { records } = await srv.request('GET', `/containers/${row.id}/records?limit=1000`);
  const live = records.filter((r) => !r.deleted);
  const open = await Promise.all(live.map(async (r) => ({ r, pl: (await srv.vault.openRecord(r)) || {} })));
  const of = (c) => open.filter((x) => x.r.collection === c);
  const milestones = of('milestone').map(({ r, pl }) => ({ id: r.id, name: pl.name || '', iso: r.meta.due || '', desc: pl.desc || '', closed: !!r.meta.closed, closedOn: pl.closedOn || '', closedBy: pl.closedBy || '', closedComment: pl.closedComment || '', history: pl.history || [] }));
  const sprints = of('sprint').map(({ r, pl }) => ({ id: r.id, name: pl.name || '', start: r.meta.start || '', end: r.meta.end || '', milestone: r.meta.milestone || '' }));
  const sprintByNo = (s) => { const m = /-s(\d+)$/.exec(s || ''); return m ? sprints.find((x) => x.name === 'Sprint ' + m[1]) : null; };
  const tasks = of('task').sort((a, b) => (a.r.id < b.r.id ? -1 : 1)).map(({ r, pl }) => {
    const sp = sprints.find((x) => x.id === (pl.sprint || r.meta.sprint)) || sprintByNo(pl.sprint || r.meta.sprint);
    return { id: r.id, title: pl.title || '', col: Math.max(0, COLS.indexOf(r.meta.column)), primary: pl.primary || '', secondary: pl.secondary || '', start: pl.start || '', due: r.meta.due || pl.due || '', desc: pl.desc || '', links: pl.links || [], objective: pl.objective || '', sprint: sp ? sp.id : '' };
  });
  const risks = of('risk').map(({ r, pl }) => ({ id: r.id, text: pl.text || '', level: r.meta.level || 'Medium' }));
  const activity = of('activity').sort((a, b) => (a.r.created_at < b.r.created_at ? 1 : -1)).map(({ r, pl }) => ({ id: r.id, text: pl.text || '', time: ago(r.created_at) }));
  const team = (conv.members || []).map((m) => (m.person === srv.me.person ? srv.me.name : m.name));
  const p = {
    id: row.id, name: conv.name || row.code, code: row.code, color: row.color, owner: row.owner_name || nameOf(srv, row.owner), team, status: row.status, goal: conv.about || '',
    start: row.start || '', due: row.due || '', budget: row.budget_pct || 0, sync: row.sync || 'Not synced', tickets: row.links || [],
    milestones, sprints, tasks, risks, activity
  };
  if (!milestones.length) p.milestones = [{ id: row.id + '-m1', name: 'Project delivery', iso: p.due }];
  return normalizeProject(p);
}

export async function load(logic, srv) {
  const { projects } = await srv.request('GET', '/projects');
  const list = await Promise.all(projects.filter((p) => p.status !== 'Pending approval').map((p) => mapProject(srv, p).catch(() => null)));
  srv.projSnap = new Map();
  const ok = list.filter(Boolean);
  ok.forEach((p) => srv.projSnap.set(p.id, JSON.parse(JSON.stringify(p))));
  logic.setState({ projects: ok });
  watch(logic, srv);
}

async function keysFor(srv, people) {
  const { devices } = await srv.request('POST', '/e2ee/people/devices', { people: [...new Set([srv.me.person, ...people])] });
  const key = newConversationKey();
  const keys = await Promise.all(devices.map(async (d) => ({ epoch: 1, device: d.id, wrapped_key: b64(await wrap(key, unb64(d.identity_key))) })));
  return keys;
}
const pidOf = (srv, n) => (n === srv.me.name || n === 'Sam' ? srv.me.person : srv.personOf.get(n));

// form: { name, goal, owner, team (names), start, due, budget }
export async function create(logic, srv, f, color) {
  const team = f.team.map((n) => pidOf(srv, n)).filter((x) => x && x !== srv.me.person);
  const conv = await srv.request('POST', '/conversations', { kind: 'project', name: f.name, about: f.goal, color, members: team, keys: await keysFor(srv, team) });
  const res = await srv.request('POST', '/projects', { conversation: conv.id, color, start: f.start, due: f.due, budget_pct: Math.min(100, Math.max(0, Number(f.budget) || 0)), sync: 'Not synced', links: [] });
  const ms = await srv.vault.seal(conv.id, { name: 'Project delivery', desc: '', history: [] });
  await srv.request('POST', `/containers/${conv.id}/records`, { id: ms.id, collection: 'milestone', epoch: ms.epoch, device: ms.device, envelope: ms.envelope, meta: { due: f.due, closed: false } });
  const act = await srv.vault.seal(conv.id, { text: 'Project created by ' + srv.me.name });
  await srv.request('POST', `/containers/${conv.id}/records`, { id: act.id, collection: 'activity', epoch: act.epoch, device: act.device, envelope: act.envelope, meta: {} });
  if (res.project && res.project.status !== 'Pending approval') await reload(logic, srv, conv.id);
  return res;
}

export async function reload(logic, srv, id) {
  const { projects } = await srv.request('GET', '/projects');
  const row = projects.find((p) => p.id === id);
  if (!row || row.status === 'Pending approval') return;
  const p = await mapProject(srv, row);
  srv.projSnap.set(p.id, JSON.parse(JSON.stringify(p)));
  srv.applyingProjects = true;
  logic.setState((st) => ({ projects: st.projects.some((x) => x.id === id) ? st.projects.map((x) => (x.id === id ? p : x)) : [p].concat(st.projects) }));
  srv.applyingProjects = false;
}

// ---- writing edits back ----
const SPEC = {
  milestones: { collection: 'milestone', payload: (m) => ({ name: m.name, desc: m.desc || '', closedOn: m.closedOn || '', closedBy: m.closedBy || '', closedComment: m.closedComment || '', history: m.history || [] }), meta: (m) => ({ due: m.iso, closed: !!m.closed }) },
  sprints: { collection: 'sprint', payload: (s) => ({ name: s.name }), meta: (s, tr) => ({ start: s.start, end: s.end, ...(tr(s.milestone) ? { milestone: tr(s.milestone) } : {}) }) },
  tasks: { collection: 'task', payload: (t, tr) => ({ title: t.title, desc: t.desc || '', primary: t.primary || '', secondary: t.secondary || '', start: t.start || '', due: t.due || '', links: t.links || [], objective: t.objective || '', sprint: tr(t.sprint) || t.sprint || '' }), meta: (t, tr, srv) => ({ column: COLS[Math.min(3, Math.max(0, t.col))], ...(t.due ? { due: t.due } : {}), ...(pidOf(srv, t.primary) ? { assignee: pidOf(srv, t.primary) } : {}), ...(t.sprint ? { sprint: String(tr(t.sprint) || t.sprint) } : {}) }) },
  risks: { collection: 'risk', payload: (r) => ({ text: r.text }), meta: (r) => ({ level: ['Low', 'Medium', 'High'].includes(r.level) ? r.level : 'Medium' }) },
  activity: { collection: 'activity', payload: (a) => ({ text: a.text }), meta: () => ({}) }
};

async function syncProject(logic, srv, prev, next) {
  const id = next.id;
  if (!srv.projIds) srv.projIds = new Map();
  const tr = (x) => (isUuid(x) ? x : srv.projIds.get(x));
  // plain facts
  const patch = {};
  if (prev.status !== next.status) patch.status = next.status;
  if (prev.due !== next.due) patch.due = next.due;
  if (prev.budget !== next.budget) patch.budget_pct = Math.min(100, Math.max(0, Number(next.budget) || 0));
  if (prev.sync !== next.sync) patch.sync = next.sync;
  if (prev.owner !== next.owner && pidOf(srv, next.owner)) patch.owner = pidOf(srv, next.owner);
  // team: give keys to newcomers before anything references them
  const added = next.team.filter((n) => prev.team.indexOf(n) === -1).map((n) => pidOf(srv, n)).filter(Boolean);
  if (added.length) {
    const conv = await srv.vault.conversation(id, true);
    const key = await srv.vault.keyFor(id, conv.epoch);
    const { devices } = await srv.request('POST', '/e2ee/people/devices', { people: added });
    await srv.request('POST', `/conversations/${id}/members`, { people: added, device: srv.vault.deviceId, keys: await srv.vault.wrapFor(id, conv.epoch, key, devices) });
  }
  for (const n of prev.team.filter((x) => next.team.indexOf(x) === -1)) {
    const pid = pidOf(srv, n);
    if (!pid) continue;
    // removing someone needs a fresh key for everyone who stays
    const conv = await srv.vault.conversation(id, true);
    const stay = (conv.members || []).filter((m) => m.person !== pid).map((m) => m.person);
    const { devices } = await srv.request('POST', '/e2ee/people/devices', { people: stay });
    const key = newConversationKey();
    const keys = await srv.vault.wrapFor(id, conv.epoch + 1, key, devices);
    await srv.request('DELETE', `/conversations/${id}/members/${pid}`, { device: srv.vault.deviceId, keys });
    srv.vault.keys.delete(id); srv.vault.convs.delete(id);
  }
  if (Object.keys(patch).length) await srv.request('PATCH', `/projects/${id}`, patch);
  // records, parents first so children can point at them
  const created = {};
  for (const k of ['milestones', 'sprints', 'tasks', 'risks', 'activity']) {
    const spec = SPEC[k];
    const before = new Map((prev[k] || []).filter((x) => x.id).map((x) => [x.id, JSON.stringify(x)]));
    const ids = new Set();
    for (const item of next[k] || []) {
      const known = item.id && (isUuid(item.id) || srv.projIds.has(item.id)) ? tr(item.id) : null;
      if (known) {
        ids.add(item.id);
        if (before.get(item.id) === JSON.stringify(item)) continue;
        const sealed = await srv.vault.seal(id, spec.payload(item, tr), known);
        await srv.request('PUT', `/containers/${id}/records/${known}`, { collection: spec.collection, epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope, meta: spec.meta(item, tr, srv) });
      } else {
        const sealed = await srv.vault.seal(id, spec.payload(item, tr));
        await srv.request('POST', `/containers/${id}/records`, { id: sealed.id, collection: spec.collection, epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope, meta: spec.meta(item, tr, srv) });
        if (item.id) { srv.projIds.set(item.id, sealed.id); ids.add(item.id); } else created[k] = (created[k] || []).concat([sealed.id]);
      }
    }
    for (const [oldId] of before) if (!ids.has(oldId) && tr(oldId)) await srv.request('DELETE', `/containers/${id}/records/${tr(oldId)}`);
  }
}

function watch(logic, srv) {
  let busy = false, again = false, timer = null;
  const run = async () => {
    if (busy) { again = true; return; }
    busy = true;
    try {
      for (const p of logic.state.projects || []) {
        if (!isUuid(p.id)) continue;
        const prev = srv.projSnap.get(p.id);
        const cur = JSON.parse(JSON.stringify(p));
        if (!prev || JSON.stringify(prev) === JSON.stringify(cur)) continue;
        srv.projSnap.set(p.id, cur);
        try { await syncProject(logic, srv, prev, cur); await reload(logic, srv, p.id); } catch (e) { logic.notify(e.message || 'Could not save the project change.', '#FF9B8A'); await reload(logic, srv, p.id).catch(() => {}); }
      }
    } finally { busy = false; if (again) { again = false; run(); } }
  };
  const orig = logic.setState.bind(logic);
  logic.setState = (update, done) => { orig(update, done); if (!srv.applyingProjects) { clearTimeout(timer); timer = setTimeout(run, 500); } };
}
