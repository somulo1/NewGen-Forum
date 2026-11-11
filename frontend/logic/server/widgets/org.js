// Sectors, branches, units and departments: each one's dashboard (what it received and sent, how much it handled on time,
// its projects) and its performance threshold. The server builds all of it from real records; nothing is made up here.
import { loadStructure, load as loadPeople } from './people.js';

export const name = 'org';

const PLURAL = { sector: 'sectors', branch: 'branches', unit: 'units', department: 'departments' };
const KIND = { sectors: 'sector', branches: 'branch', units: 'unit', departments: 'department' };
const FRESH_MS = 60000;

// "departments:<id>" is how the pages name a unit
export const keyOf = (kind, id) => PLURAL[kind] + ':' + id;
export const idOf = (key) => key.slice(key.indexOf(':') + 1);

export async function load(logic, srv) {
  srv.ouData = srv.ouData || new Map();
  srv.ouAsked = srv.ouAsked || new Map();
  const { units } = await srv.request('GET', '/org/structure');
  const targets = {};
  units.forEach((u) => { if (u.target) targets[keyOf(u.kind, u.id)] = u.target; });
  logic.setState({ targets });
  await Promise.all([loadPerformance(logic, srv), loadLevels(logic, srv)]);
}

// The structure levels this organization uses: which of sector, branch, unit and department, their names, and which sits under which.
export async function loadLevels(logic, srv) {
  try { srv.levels = await srv.request('GET', '/org/levels'); } catch (e) { srv.levels = srv.levels || { levels: [], set: false }; }
  logic.forceUpdate();
}
export async function saveLevels(logic, srv, levels) {
  srv.levels = await srv.request('PUT', '/org/levels', { levels });
  logic.forceUpdate();
}
// Places people in a part of the organization: the part itself and every part above it, so their record stays consistent.
export async function addPeople(logic, srv, unitId, personIds) {
  const units = srv.unitList || [];
  const byId = new Map(units.map((u) => [u.id, u]));
  const body = {};
  for (let u = byId.get(unitId); u; u = byId.get(u.parent_id)) body[u.kind + '_id'] = u.id;
  const failed = [];
  for (const pid of personIds) {
    try { await srv.request('PATCH', '/people/' + pid, body); } catch (e) { failed.push((e && e.message) || 'That did not go through.'); }
  }
  const pw = await import('./people.js'); await pw.load(logic, srv);
  return failed;
}

// One short summary per unit this person may open: the percentages on the org page, the cards and the widgets.
export async function loadPerformance(logic, srv) {
  try {
    const { units } = await srv.request('GET', '/org/performance');
    srv.perf = new Map(units.map((u) => [keyOf(u.kind, u.id), u]));
  } catch (e) { srv.perf = new Map(); }
  srv.perfAt = Date.now();
  logic.forceUpdate();
}

// Loads (or refreshes) the dashboard of one unit; safe to call from a render, it asks the server at most once a minute.
export function ensure(logic, srv, key, force) {
  // the period the dashboard covers: this month, quarter or year, or the last 30 days
  const period = logic.state.ouPeriod == null ? 'month' : logic.state.ouPeriod;   // this month unless another period is picked
  const tag = key + '|' + period;
  const asked = srv.ouAsked.get(tag) || 0;
  if (!force && Date.now() - asked < FRESH_MS) return;
  srv.ouAsked.set(tag, Date.now());
  if (srv.ouPeriodOf && srv.ouPeriodOf.get(key) !== period) srv.ouData.delete(key);   // another period was shown: fetch, do not show the old numbers
  srv.request('GET', `/org/units/${idOf(key)}/dashboard` + (period ? '?period=' + period : '')).then((d) => {
    (srv.ouPeriodOf = srv.ouPeriodOf || new Map()).set(key, period);
    srv.ouData.set(key, d);
    logic.forceUpdate();
  }).catch(() => { (srv.ouFailed = srv.ouFailed || new Set()).add(key); logic.forceUpdate(); });   // no right to open it: the page shows nothing for it, not placeholders for ever
}

export async function setTarget(logic, srv, key, value) {
  await srv.request('PUT', `/org/units/${idOf(key)}/target`, { target: value });
  logic.setState((st) => ({ targets: { ...st.targets, [key]: value } }));
  ensure(logic, srv, key, true);
  loadPerformance(logic, srv);
}

export { KIND };

// ---- changing the structure (People > Organization structure) ----
// Levels exist once they have a unit, so adding the first sector makes sector pages, onboarding fields and CSV columns appear.
async function refresh(logic, srv) {
  await loadStructure(logic, srv);
  await load(logic, srv);
  srv.ouData.clear(); srv.ouAsked.clear();
  logic.forceUpdate();
}
export async function addUnit(logic, srv, u) {
  const first = !(srv.unitList || []).some((x) => x.kind === u.kind);
  await srv.request('POST', '/org/units', { kind: u.kind, name: u.name, parent_id: u.parent || undefined });
  await refresh(logic, srv);
  // the first sector, branch, unit or department makes that level's widget available; put it on this person's Home
  const id = PLURAL[u.kind];
  if (first && !(logic.state.widgets || []).includes(id)) logic.setState((st) => ({ widgets: st.widgets.concat(id) }));
}
export async function renameUnit(logic, srv, id, name) {
  await srv.request('PATCH', `/org/units/${id}`, { name });
  await refresh(logic, srv);
}
// Gives a unit its head role (a role key), or clears it (''): everyone who holds that role and is placed in the unit heads it.
export async function setHead(logic, srv, id, role) {
  await srv.request('PUT', `/org/units/${id}/head`, { role });
  await refresh(logic, srv);
}
export async function removeUnit(logic, srv, id) {
  await srv.request('DELETE', `/org/units/${id}`);
  await refresh(logic, srv);
}
// Moving or merging re-places people on the server, so the directory is read again too.
export async function moveUnit(logic, srv, id, parent) {
  await srv.request('POST', `/org/units/${id}/move`, { parent_id: parent || null });
  await refresh(logic, srv);
  await loadPeople(logic, srv);
}
export async function mergeUnits(logic, srv, into, from) {
  await srv.request('POST', '/org/units/merge', { into, from });
  await refresh(logic, srv);
  await loadPeople(logic, srv);
}
