// Where a ticket lands depends on what the organization has set up.
//   - Any of department / unit / branch / sector: the ticket is received by one of them and shows on that dashboard.
//   - None of them: every ticket goes to the ticket desk, a dashboard of its own under the Tickets plugin.
const ORDER = ['departments', 'units', 'branches', 'sectors'];
export const LEVEL_LABEL = { departments: 'Department', units: 'Unit', branches: 'Branch', sectors: 'Sector' };
export const DESK = { type: 'desk', name: 'Ticket desk' };

export const levelsOf = (has) => ORDER.filter((t) => has[t]);

// A ticket that names its recipient keeps it, as long as that level still exists.
// Older tickets only know their department, so they are placed on the nearest level the organization has.
export function recipientOf(t, org, has) {
  const levels = levelsOf(has);
  if (!levels.length) return DESK;
  if (t.to && t.to.type !== 'desk' && has[t.to.type] && (org[t.to.type] || []).some((r) => r.name === t.to.name)) return { type: t.to.type, name: t.to.name };
  const dept = org.departments.find((d) => d.name === t.dept) || org.departments[0];
  const type = levels[0];
  let rec = null;
  if (type === 'departments') rec = dept;
  else if (type === 'units') rec = org.units.find((u) => dept && u.branch === dept.branch) || org.units[0];
  else if (type === 'branches') rec = org.branches.find((b) => dept && b.id === dept.branch) || org.branches[0];
  else rec = org.sectors.find((x) => dept && x.id === dept.sector) || org.sectors[0];
  return rec ? { type, name: rec.name } : DESK;
}

export const recipientLabel = (r) => (r.type === 'desk' ? DESK.name : LEVEL_LABEL[r.type] + ': ' + r.name);

// ---- Ticket categories per receiving unit ----
// A department, unit, branch, sector or the ticket desk can define its own categories, each with a TAT (hours), a priority
// and the members who always take it until they are delinked. Without its own list, the organization's default categories apply.
export const unitKey = (r) => r.type + ':' + r.name;

export function catsFor(unitCats, r, defaults) {
  const own = unitCats[unitKey(r)];
  if (own && own.length) return own.map((c) => ({ ...c, custom: true }));
  return defaults.map((c) => ({ id: 'g-' + c.name, name: c.name, tat: c.tat, prio: 'Medium', members: [], custom: false }));
}

// The member of a category with the fewest open tickets takes the new one.
export function pickHandler(cat, tickets) {
  if (!cat.members || !cat.members.length) return '';
  const load = (n) => tickets.filter((t) => t.step < 3 && t.handler === n).length;
  return cat.members.slice().sort((a, b) => load(a) - load(b))[0];
}
