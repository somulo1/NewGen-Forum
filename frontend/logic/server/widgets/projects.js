// Projects: the Projects widget and page. A project is a conversation (its plan is sealed for the team: milestones,
// sprints, tasks, risks, activity are records) plus plain facts the server acts on (code, status, dates, budget, owner).
// The page edits state.projects directly; this module watches those edits and writes the differences to the server.
import { newConversationKey, wrap, newId, b64, unb64 } from '../e2ee.js';
import { normalizeProject } from '../../projects.js';
import * as files from '../files.js';
import { joinPage } from '../../pageCollab.js';

export const name = 'projects';

const COLS = ['todo', 'doing', 'review', 'done'];
// a whole id: a new item's temporary id starts with the project's id, so the start alone is not enough
const isUuid = (x) => typeof x === 'string' && /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[0-9a-f]{4}-[0-9a-f]{12}$/.test(x);
const ago = (iso) => { const m = Math.round((Date.now() - new Date(iso)) / 60000); return m < 60 ? Math.max(1, m) + ' min ago' : m < 1440 ? Math.round(m / 60) + ' h ago' : Math.round(m / 1440) === 1 ? 'Yesterday' : Math.round(m / 1440) + ' days ago'; };
const pidOf = (srv, n) => (n === srv.me.name || n === 'Sam' ? srv.me.person : srv.personOf.get(n));
const nameOf = (srv, pid) => (pid === srv.me.person ? srv.me.name : srv.nameOf.get(pid) || '');

// Older projects: their plan is still sealed in encrypted records. A member's device opens it once and hands it to the server as plain rows.
async function mapLegacy(srv, row) {
  const conv = await srv.vault.conversation(row.id, true);
  const { records } = await srv.request('GET', `/containers/${row.id}/records?limit=1000`);
  const live = records.filter((r) => !r.deleted);
  const open = await Promise.all(live.map(async (r) => { const pl = await srv.vault.openRecord(r); return { r, pl: pl || {}, unread: !pl }; }));
  const of = (c) => open.filter((x) => x.r.collection === c);
  const milestones = of('milestone').map(({ r, pl }) => ({ id: r.id, name: pl.name || '', iso: r.meta.due || '', desc: pl.desc || '', closed: !!r.meta.closed, closedOn: pl.closedOn || '', closedBy: pl.closedBy || '', closedComment: pl.closedComment || '', history: pl.history || [] }));
  const sprints = of('sprint').map(({ r, pl }) => ({ id: r.id, name: pl.name || '', start: r.meta.start || '', end: r.meta.end || '', milestone: r.meta.milestone || '' }));
  const sprintByNo = (s) => { const m = /-s(\d+)$/.exec(s || ''); return m ? sprints.find((x) => x.name === 'Sprint ' + m[1]) : null; };
  const tasks = of('task').sort((a, b) => (a.r.id < b.r.id ? -1 : 1)).map(({ r, pl }) => {
    const sp = sprints.find((x) => x.id === (pl.sprint || r.meta.sprint)) || sprintByNo(pl.sprint || r.meta.sprint);
    return { id: r.id, title: pl.title || '', col: Math.max(0, COLS.indexOf(r.meta.column)), primary: pl.primary || '', secondary: pl.secondary || '', start: pl.start || '', due: r.meta.due || pl.due || '', desc: pl.desc || '', links: pl.links || [], tickets: pl.tickets || [], objective: pl.objective || '', history: pl.history || [], sprint: sp ? sp.id : '' };
  });
  const risks = of('risk').map(({ r, pl }) => ({ id: r.id, text: pl.text || '', level: r.meta.level || 'Medium' }));
  const activity = of('activity').sort((a, b) => (a.r.created_at < b.r.created_at ? 1 : -1)).map(({ r, pl }) => ({ id: r.id, text: pl.text || '', time: ago(r.created_at) }));
  const team = (conv.members || []).map((m) => (m.person === srv.me.person ? srv.me.name : m.name));
  const p = {
    id: row.id, name: conv.name || row.code, code: row.code, color: row.color, owner: row.owner_name || nameOf(srv, row.owner), team, status: row.status, goal: conv.about || '',
    start: row.start || '', due: row.due || '', budget: row.budget_pct || 0, sync: row.sync || 'Not synced', tickets: row.links || [],
    milestones, sprints, tasks, risks, activity
  };
  p.unreadable = open.some((x) => x.unread && ['task', 'milestone', 'sprint', 'risk', 'activity'].includes(x.r.collection));
  p.times = new Map(live.map((r) => [r.id, r.created_at]));
  return p;
}

const plainOf = (x) => { const o = { ...x }; ['id', 'ref', 'key', 'changes', 'time', 'createdAt', 'updatedAt'].forEach((k) => delete o[k]); return o; };

async function importLegacy(srv, row) {
  const old = await mapLegacy(srv, row);
  if (old.unreadable) return false;   // this device cannot read all of it: leave it for one that can, losing nothing
  const items = [];
  const add = (collection, list, data) => (list || []).forEach((x) => { if (isUuid(x.id)) items.push({ id: x.id, collection, created_at: old.times.get(x.id), data: data(x) }); });
  add('milestone', old.milestones, plainOf);
  add('sprint', old.sprints, plainOf);
  add('task', old.tasks, (t) => ({ ...plainOf(t), kind: 'task', deps: [], status: COLS[t.col] || 'todo', assignee: pidOf(srv, t.primary) || '', secondary_id: pidOf(srv, t.secondary) || '' }));
  add('risk', old.risks, plainOf);
  add('activity', old.activity, plainOf);
  await srv.request('POST', `/projects/${row.id}/import`, { items });
  return true;
}

async function mapProject(srv, row) {
  if (!row.plan_imported && !(await importLegacy(srv, row))) {
    const old = await mapLegacy(srv, row);
    if (!old.milestones.length) old.milestones = [{ id: row.id + '-m1', name: 'Project delivery', iso: old.due }];
    return normalizeProject({ ...old, method: row.method, workflow: row.workflow, unreadable: undefined, times: undefined, pages: [] });
  }
  if (srv.rt) srv.rt.sub(row.id);   // live changes and co-editing arrive on the project's channel
  const conv = await srv.vault.conversation(row.id, true);
  const { items, changes } = await srv.request('GET', `/projects/${row.id}/plan`);
  const changesOf = new Map();
  (changes || []).forEach((c) => { if (!changesOf.has(c.item)) changesOf.set(c.item, []); changesOf.get(c.item).push(c); });
  const of = (c) => items.filter((x) => x.collection === c);
  const base = (x) => ({ ...x.data, id: x.id, createdAt: x.created_at, updatedAt: x.updated_at, changes: changesOf.get(x.id) || [] });
  const tasks = of('task').map((x) => {
    const t = base(x);
    return { ...t, key: row.code + '-' + x.ref, primary: (t.assignee && nameOf(srv, t.assignee)) || t.primary || '', secondary: (t.secondary_id && nameOf(srv, t.secondary_id)) || t.secondary || '', start: t.start || '', due: t.due || row.due || '', desc: t.desc || '', links: t.links || [], tickets: t.tickets || [], history: t.history || [] };
  });
  const p = {
    id: row.id, name: conv.name || row.code, code: row.code, color: row.color, owner: row.owner_name || nameOf(srv, row.owner), team: (conv.members || []).map((m) => (m.person === srv.me.person ? srv.me.name : m.name)), members: (conv.members || []).map((m) => ({ person: m.person, name: m.person === srv.me.person ? srv.me.name : m.name })), status: row.status, goal: conv.about || '',
    start: row.start || '', due: row.due || '', budget: row.budget_pct || 0, sync: row.sync || 'Not synced', tickets: row.links || [], method: row.method || 'agile', workflow: row.workflow,
    milestones: of('milestone').map(base), sprints: of('sprint').map(base), tasks, risks: of('risk').map(base), pages: of('page').map(base), comments: of('page_comment').map(base),
    activity: of('activity').reverse().map((x) => ({ ...base(x), time: ago(x.created_at) }))
  };
  if (!p.milestones.length) p.milestones = [{ id: row.id + '-m1', name: 'Project delivery', iso: p.due }];
  return normalizeProject(p);
}

export async function load(logic, srv) {
  const { projects } = await srv.request('GET', '/projects');
  const list = await Promise.all(projects.filter((p) => p.status !== 'Pending approval').map((p) => mapProject(srv, p).catch(() => null)));
  srv.projSnap = new Map();
  const ok = list.filter(Boolean);
  ok.forEach((p) => srv.projSnap.set(p.id, JSON.parse(JSON.stringify(p))));
  srv.applyingProjects = true;
  try { logic.setState((st) => ({ projects: ok, projectOpen: ok.some((p) => p.id === st.projectOpen) ? st.projectOpen : null })); } finally { srv.applyingProjects = false; }
  watch(logic, srv);
}

async function keysFor(srv, people) {
  const { devices } = await srv.request('POST', '/e2ee/people/devices', { people: [...new Set([srv.me.person, ...people])] });
  const key = newConversationKey();
  const keys = await Promise.all(devices.map(async (d) => ({ epoch: 1, device: d.id, wrapped_key: b64(await wrap(key, unb64(d.identity_key))) })));
  return keys;
}

// form: { name, goal, method, owner, team (names), start, due, budget }
export async function create(logic, srv, f, color) {
  const team = f.team.map((n) => pidOf(srv, n)).filter((x) => x && x !== srv.me.person);
  const conv = await srv.request('POST', '/conversations', { kind: 'project', name: f.name, about: f.goal, color, members: team, keys: await keysFor(srv, team) });
  const res = await srv.request('POST', '/projects', { conversation: conv.id, color, method: f.method || 'agile', start: f.start, due: f.due, budget_pct: Math.min(100, Math.max(0, Number(f.budget) || 0)), sync: 'Not synced', links: [] });
  await srv.request('PUT', `/projects/${conv.id}/items/${newId()}`, { collection: 'milestone', data: { name: f.method === 'waterfall' ? 'Delivery phase' : 'Project delivery', desc: '', iso: f.due, closed: false, history: [] } });
  await srv.request('PUT', `/projects/${conv.id}/items/${newId()}`, { collection: 'activity', data: { text: 'Project created by ' + srv.me.name } });
  if (res.project && res.project.status !== 'Pending approval') await reload(logic, srv, conv.id);
  return res;
}

export async function closeSprint(logic, srv, pid, sprint) {
  await srv.request('POST', `/projects/${pid}/sprints/${sprint}/close`, {});
  await reload(logic, srv, pid);
}

export async function passGate(logic, srv, pid, phase, comment) {
  const r = await srv.request('POST', `/projects/${pid}/gates/${phase}`, { comment });
  await reload(logic, srv, pid);
  return r;
}

// Saves pages from a Confluence export one by one: one that fails is reported, and the rest still arrive.
export async function importPages(logic, srv, pid, pages, taken, onProgress) {
  const ids = new Map();
  const failed = [];
  let done = 0;
  for (const g of pages) {
    const id = newId();
    let title = g.title;
    for (let n = 2; taken.has(title.toLowerCase()); n++) title = g.title + ' (' + n + ')';
    taken.add(title.toLowerCase());
    const uploaded = [];
    for (const a of g.attachments) {
      try { uploaded.push(await attach(srv, pid, new File([await a.read()], a.name))); } catch (e) { failed.push(title + ': attachment ' + a.name); }
    }
    try {
      await srv.request('PUT', `/projects/${pid}/items/${id}`, { collection: 'page', data: { title, body: g.body, parent: ids.get(g.parentFile) || '', labels: ['imported'], author: srv.me.name, watchers: [srv.me.person], files: uploaded } });
      ids.set(g.file, id);
    } catch (e) { failed.push(title + ': ' + (e.message || 'could not be saved')); }
    onProgress(++done);
  }
  await reload(logic, srv, pid);
  return failed;
}

export const joinLive = (srv, pid, page, onPeers) => joinPage(srv, pid, page, onPeers);

export async function attach(srv, pid, file) {
  const ref = await files.upload(srv, pid, file);
  return { ref, name: file.name, size: file.size };
}
export const openAttachment = async (srv, ref) => { window.open(await files.openUrl(srv, ref), '_blank', 'noopener'); };

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
  milestones: { collection: 'milestone', data: (m) => plainOf(m) },
  sprints: { collection: 'sprint', data: (x, tr) => ({ ...plainOf(x), milestone: tr(x.milestone) || '' }) },
  tasks: { collection: 'task', data: (t, tr, srv) => ({ ...plainOf(t), sprint: tr(t.sprint) || '', milestone: tr(t.milestone) || '', parent: tr(t.parent) || '', deps: (t.deps || []).map(tr).filter(Boolean), assignee: pidOf(srv, t.primary) || '', secondary_id: pidOf(srv, t.secondary) || '' }) },
  risks: { collection: 'risk', data: (r) => ({ ...plainOf(r), level: ['Low', 'Medium', 'High'].includes(r.level) ? r.level : 'Medium' }) },
  activity: { collection: 'activity', data: (a) => plainOf(a) },
  pages: { collection: 'page', data: (x, tr) => { const o = { ...plainOf(x), parent: tr(x.parent) || '', baseAt: x.liveSave ? '' : x.updatedAt || '' }; delete o.liveSave; return o; } },
  comments: { collection: 'page_comment', data: (x, tr) => ({ ...plainOf(x), page: tr(x.page) || '' }) }
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
  if (prev.method !== next.method) patch.method = next.method;
  if (JSON.stringify(prev.workflow) !== JSON.stringify(next.workflow)) patch.workflow = next.workflow;
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
  // items, parents first so children can point at them
  const save = (itemId, spec, item) => srv.request('PUT', `/projects/${id}/items/${itemId}`, { collection: spec.collection, data: spec.data(item, tr, srv) });
  for (const k of ['milestones', 'sprints', 'tasks', 'risks', 'activity', 'pages', 'comments']) {
    const spec = SPEC[k];
    const before = new Map((prev[k] || []).filter((x) => x.id).map((x) => [x.id, JSON.stringify(x)]));
    const ids = new Set();
    for (const item of next[k] || []) {
      const known = item.id && (isUuid(item.id) || srv.projIds.has(item.id)) ? tr(item.id) : null;
      if (known) {
        ids.add(item.id);
        if (before.get(item.id) !== JSON.stringify(item)) await save(known, spec, item);
      } else {
        const fresh = newId();
        await save(fresh, spec, item);
        if (item.id) { srv.projIds.set(item.id, fresh); ids.add(item.id); }
      }
    }
    for (const [oldId] of before) if (!ids.has(oldId) && tr(oldId)) await srv.request('DELETE', `/projects/${id}/items/${tr(oldId)}`);
  }
}

function watch(logic, srv) {
  // someone else on the team changed the plan
  if (srv.rt && !srv.projLive) {
    srv.projLive = true;
    try { srv.rt.on('project.plan', (ev) => { const d = ev.data || {}; if (d.by !== srv.me.person && d.project && srv.projSnap.has(d.project)) reload(logic, srv, d.project).catch(() => {}); }); } catch (e) { /* no live connection */ }
  }
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
