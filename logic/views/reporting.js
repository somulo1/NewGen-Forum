// Reporting lines: a role reports to another role, and a person reports to whoever holds that role in their own scope.
// Step 1 builds the lookup (used by profiles and the people drawer); step 2 builds the Roles page tab.
export const DEFAULT_REPORTS = { staff: 'hod', it: 'hod', finance: 'hod', hod: 'branch', unit: 'branch', branch: 'sector', sector: 'admin', guest: '' };
const ME = 'Samwel Okoth Omulo';
const SCOPE = { hod: 'dept', unit: 'unit', branch: 'loc', sector: 'sector' };
const DEFAULT_ROLE = { [ME]: 'admin', 'Brian Kiprop': 'it', 'David Otieno': 'staff', 'Amina Wanjiru': 'finance', 'Grace Njeri': 'hod', 'Faith Achieng': 'staff', 'Kevin Mutua': 'staff', 'Mercy Wambui': 'staff', 'John Kamau': 'staff', 'Lilian Chebet': 'staff' };

export function buildReporting(ctx) {
  const { has, s, visibleRoles } = ctx;
  const vis = (id) => visibleRoles.some((r) => r.id === id);
  const roleName = (id) => (visibleRoles.find((r) => r.id === id) || {}).name || id;
  const ridOf = (n) => {
    const d = this.dir.find((x) => x.name === n);
    const rid = s.memberRoles[n] || DEFAULT_ROLE[n] || (d && /Lead|Supervisor/.test(d.title) ? (has.branches ? 'branch' : 'hod') : 'staff');
    return vis(rid) ? rid : 'staff';
  };
  const meDir = { name: ME, title: 'Cybersecurity & Technology Risk Officer', dept: 'Risk & Compliance', loc: 'Head Office', sector: 'Corporate', unit: 'Service Desk' };
  const personOf = (n) => (n === ME ? meDir : this.dir.find((x) => x.name === n));
  const holdersOf = (rid) => this.dir.map((d) => d).concat([meDir]).filter((d, i, a) => a.findIndex((x) => x.name === d.name) === i).filter((d) => ridOf(d.name) === rid);
  const reportsMap = { ...DEFAULT_REPORTS, ...s.reportsTo };
  // Who a person reports to under a given map. If nobody holds the target role, it climbs to that role's own reporting role.
  const reportOfWith = (map) => (name) => {
    const d = personOf(name);
    let rid = ridOf(name);
    const seen = {};
    for (let hop = 0; hop < 8; hop++) {
      const t = map[rid];
      if (!t || seen[t]) break;
      seen[t] = true;
      if (vis(t)) {
        const list = holdersOf(t).filter((h) => h.name !== name);
        const f = SCOPE[t];
        const scoped = f && d ? list.filter((h) => h[f] && h[f] === d[f]) : list;
        const pick = scoped[0] || list[0];
        if (pick) return { name: pick.name, roleName: roleName(t), sub: roleName(t) + (pick.dept ? ' · ' + pick.dept : ''), top: false };
      }
      rid = t;
    }
    return { name: '', roleName: '', sub: 'Top of the organization', top: true };
  };
  const reportOf = reportOfWith(reportsMap);
  return { reportOf, reportOfWith, reportsMap, ridOf, roleHolders: holdersOf };
}

export function buildReportingTab(ctx) {
  const { C, can, reportOfWith, reportsMap, ridOf, roleHolders, s, visibleRoles } = ctx;
  const canEdit = can('roles', 'Manage roles & rights');
  const draft = s.reportDraft || {};
  const eff = { ...reportsMap, ...draft };
  const dirty = Object.keys(draft).some((k) => (draft[k] || '') !== (reportsMap[k] || ''));
  const roleName = (id) => (visibleRoles.find((r) => r.id === id) || {}).name || id;
  const effReport = reportOfWith(eff);
  const set = (rid, v) => this.setState((st) => ({ reportDraft: { ...(st.reportDraft || {}), [rid]: v } }));
  const rows = visibleRoles.map((r) => {
    const holders = roleHolders(r.id);
    const first = holders.find((h) => h.name !== 'Samwel Okoth Omulo') || holders[0];
    const to = first ? effReport(first.name) : null;
    const opts = [{ v: '', l: 'Nobody (top level)' }].concat(visibleRoles.filter((x) => x.id !== r.id).map((x) => ({ v: x.id, l: x.name })));
    return { cells: [C.t(r.name, { strong: true, sub: r.desc }), C.t(String(holders.length), { nowrap: true }),
      C.sel(eff[r.id] || '', opts, (e) => set(r.id, e.target.value), { label: r.name + ' reports to', minWidth: '200px', disabled: !canEdit }),
      C.t(first ? (to && !to.top ? first.name + ' → ' + to.name : first.name + ' → nobody (top level)') : 'No one has this role yet', { sub: to && !to.top ? to.sub : '' })] };
  });
  const reportingTable = { id: 'reporting', pageSize: 12, minWidth: '820px', cols: [{ label: 'Role', filter: 'text' }, { label: 'People' }, { label: 'Reports to' }, { label: 'Example' }], rows };
  // A role cannot end up reporting to itself through a chain.
  const cycleFor = () => { for (const r of visibleRoles) { const seen = {}; let id = r.id; while (eff[id]) { if (seen[eff[id]]) return roleName(r.id); seen[id] = true; id = eff[id]; if (id === r.id) return roleName(r.id); } } return ''; };
  const reportingBar = { canEdit, dirty, discard: () => this.setState({ reportDraft: null }),
    save: () => {
      const bad = cycleFor();
      if (bad) { this.notify('These lines go in a circle (' + bad + '). Change one so the chain ends at the top.', '#FFB86B'); return; }
      this.setState((st) => ({ reportsTo: { ...st.reportsTo, ...Object.keys(st.reportDraft || {}).reduce((m, k) => { m[k] = st.reportDraft[k]; return m; }, {}) }, reportDraft: null,
        extraLogs: [{ id: 'x' + Date.now(), ago: 0, user: 'Samwel Okoth Omulo', action: 'Updated reporting lines between roles', module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' }].concat(st.extraLogs) }));
      this.notify('Reporting lines saved. Profiles now show the new managers.', '#B8F36B');
    } };
  return { reportingTable, reportingBar };
}
