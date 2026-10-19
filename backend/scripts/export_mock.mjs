// Dumps the frontend's mock data (read-only: the frontend is not modified) to JSON files the Go seeder loads.
// Usage: node --import ./scripts/register.mjs scripts/export_mock.mjs
import { writeFileSync, mkdirSync } from 'node:fs';
import path from 'node:path';
globalThis.window = { innerWidth: 1400, innerHeight: 900, matchMedia: () => ({ matches: false }), addEventListener() {}, removeEventListener() {}, location: { origin: 'http://localhost:3000' } };
globalThis.localStorage = { getItem: () => null, setItem() {}, removeItem() {} };
const ROOT = path.resolve(import.meta.dirname, '..', '..');
const OUT = path.resolve(import.meta.dirname, '..', 'seed', 'data');
mkdirSync(OUT, { recursive: true });
const { HubLogic } = await import(path.join(ROOT, 'logic/HubLogic.js'));
const L = new HubLogic({});
const clean = (v) => JSON.parse(JSON.stringify(v, (k, x) => (typeof x === 'function' ? undefined : x)));
let vals = {};
try { vals = L.renderVals() || {}; } catch (e) { console.error('renderVals failed (forum data will be partial):', e.message); }
const dump = (name, v) => { writeFileSync(path.join(OUT, name + '.json'), JSON.stringify(clean(v), null, 1)); console.log(name, JSON.stringify(clean(v)).length, 'bytes'); };
dump('state', L.state);
dump('dir', L.dir);
dump('org', L.org);
dump('baseLogs', L.baseLogs || []);
dump('forums', (vals.forums || []).map((f) => { const { open, rowStyle, chipStyle, ...rest } = f; return rest; }));

// Roles, rights catalog and the default rights of every role, computed by the frontend's own rules.
{
  const { buildAccess } = await import(path.join(ROOT, 'logic/views/access.js'));
  const all = ['departments', 'sectors', 'branches', 'units', 'chat', 'calendar', 'memos', 'newsletter', 'announcements', 'tickets', 'projects', 'files', 'secureshare', 'sessions', 'awards', 'notes'];
  const saved = L.state;
  L.state = { ...L.state, widgets: all, rightsOv: {}, actingRole: 'admin' };
  const a = buildAccess.call(L, {});
  L.state = saved;
  const modules = Object.keys(a.rightsDef).map((k) => ({ key: k, label: a.modInfo[k][0], tint: a.modInfo[k][1], rights: a.rightsDef[k] }));
  const defaults = {};
  for (const r of a.roleDefs) {
    defaults[r.id] = {};
    for (const m of modules) defaults[r.id][m.key] = m.rights.filter((x) => a.rightOf(r.id, m.key, x));
  }
  dump('access', { modules, roles: a.roleDefs.map((r) => ({ ...r, scope: a.roleScope[r.id] })), defaults, flows: saved.flows });
}
