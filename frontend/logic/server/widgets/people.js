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
  all.forEach((p) => {
    srv.nameOf.set(p.id, p.name);
    if (p.id === me.person) return;   // the signed-in person is "Sam"/"me" in the UI, not a directory row
    srv.personOf.set(p.name, p.id);
    dir.push({
      id: dir.length, uid: p.id, account: p.account, name: p.name, email: p.email || '', initials: p.initials, dept: p.department ? p.department.name : '',
      title: p.title || '', loc: p.branch ? p.branch.name : '', sector: p.sector ? p.sector.name : '', unit: p.unit ? p.unit.name : '',
      color: p.color || '#7CC4FF', status: STATUS[p.availability] || 'Available'
    });
  });
  srv.dirByUid = new Map(dir.map((d) => [d.uid, d]));
  logic.dir = dir;
}
