// People directory, current person and organization structure. Everything else names people through this.
export const name = 'people';

const STATUS = { Available: 'Available', Away: 'Away', 'In a meeting': 'In a meeting', Focusing: 'Focusing' };

export async function load(logic, srv) {
  const all = [];
  let next = '';
  do {
    const r = await srv.request('GET', '/people?limit=100' + (next ? '&cursor=' + encodeURIComponent(next) : ''));
    all.push(...r.people);
    next = r.next || '';
  } while (next);
  const me = srv.me;
  const dir = [];
  let selfRow = null;
  all.forEach((p) => {
    srv.nameOf.set(p.id, p.name);
    const row = (id) => ({
      id, uid: p.id, hasAccount: !!p.has_account, accStatus: p.status || 'active', name: p.name, email: p.email || '', initials: p.initials, dept: p.department ? p.department.name : '',
      title: p.title || '', loc: p.branch ? p.branch.name : '', sector: p.sector ? p.sector.name : '', unit: p.unit ? p.unit.name : '',
      color: p.color || '#7CC4FF', joined: p.joined_on || '', dob: p.date_of_birth || '', status: STATUS[p.availability] || 'Available', role: p.role || ''
    });
    // The signed-in person is "Sam"/"me" in the UI, so they are not in the pickers (assign, mention, message). They are in the
    // People page, where everyone is listed and searchable (logic.dirSelf).
    if (p.id === me.person) { selfRow = row(-1); return; }
    srv.personOf.set(p.name, p.id);
    dir.push(row(dir.length));
  });
  if (selfRow) selfRow.id = dir.length;
  logic.dirSelf = selfRow;
  srv.dirByUid = new Map(dir.map((d) => [d.uid, d]));
  logic.dir = dir;
  await loadStructure(logic, srv);
}

// Organization structure: the frontend names levels by their own words (sectors, branches, units, departments).
const LEVELS = { sector: 'sectors', branch: 'branches', unit: 'units', department: 'departments' };

export async function loadStructure(logic, srv) {
  const { units } = await srv.request('GET', '/org/structure');
  const byId = new Map(units.map((u) => [u.id, u]));
  const ancestor = (u, kind) => { let x = u; while (x) { if (x.kind === kind) return x; x = byId.get(x.parent_id); } return null; };
  const org = { sectors: [], branches: [], units: [], departments: [], mine: {} };
  srv.structure = { sectors: new Map(), branches: new Map(), units: new Map(), departments: new Map() };
  srv.unitName = new Map(units.map((u) => [u.id, u.name]));
  srv.unitList = units;   // as the server has them (with parents), for the Organization structure editor
  units.forEach((u) => {
    const level = LEVELS[u.kind];
    if (!level) return;
    const rec = { id: u.id, name: u.name };
    if (u.kind === 'branch') rec.sector = (ancestor(byId.get(u.parent_id), 'sector') || {}).id || '';
    if (u.kind === 'unit') rec.branch = (ancestor(byId.get(u.parent_id), 'branch') || {}).id || '';
    if (u.kind === 'department') { rec.branch = (ancestor(byId.get(u.parent_id), 'branch') || {}).id || ''; rec.sector = (ancestor(byId.get(u.parent_id), 'sector') || {}).id || ''; }
    org[level].push(rec);
    srv.structure[level].set(u.name, u.id);
  });
  const mine = await srv.request('GET', '/people/' + srv.me.person);
  const pick = (o) => (o && o.id) || '';
  org.mine = { departments: pick(mine.department), branches: pick(mine.branch), sectors: pick(mine.sector), units: pick(mine.unit) };
  logic.org = org;
}
