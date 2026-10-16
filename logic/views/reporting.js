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
  // The role a person reports to under a map. If the mapped role is not in use here, it climbs to that role's own reporting role.
  const targetRoleWith = (map) => (name) => {
    let rid = ridOf(name);
    const seen = {};
    for (let hop = 0; hop < 8; hop++) {
      const t = map[rid];
      if (!t || seen[t]) return '';
      seen[t] = true;
      if (vis(t)) return t;
      rid = t;
    }
    return '';
  };
  // By default a person reports to a role, and the profile shows the role name. If someone picked a specific person in that
  // role for them, the profile shows that exact person instead.
  const reportOfWith = (map) => (name) => {
    const t = targetRoleWith(map)(name);
    if (!t) return { name: '', roleName: '', display: 'No one', sub: 'Top of the organization', specific: false, top: true };
    const ov = s.personReports[name];
    const pick = ov && ov !== name && personOf(ov) && ridOf(ov) === t ? personOf(ov) : null;
    if (pick) return { name: pick.name, roleName: roleName(t), display: pick.name, sub: roleName(t) + (pick.dept ? ' · ' + pick.dept : ''), specific: true, top: false };
    return { name: '', roleName: roleName(t), display: roleName(t), sub: 'Role', specific: false, top: false };
  };
  const reportOf = reportOfWith(reportsMap);
  const targetRoleOf = targetRoleWith(reportsMap);
  return { reportOf, reportOfWith, reportsMap, ridOf, roleHolders: holdersOf, targetRoleOf, reportRoleName: roleName };
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
      C.t(first ? (to && !to.top ? first.name + ' → ' + to.display : first.name + ' → nobody (top level)') : 'No one has this role yet', { sub: to && !to.top && to.specific ? to.sub : 'Shown on their profile as the role. Pick a specific person when editing someone.' })] };
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
