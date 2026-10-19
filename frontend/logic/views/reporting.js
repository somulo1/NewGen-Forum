import * as rolesW from '~/logic/server/widgets/roles.js';
// Reporting lines: a role reports to another role, and a person reports to whoever holds that role in their own scope.
// Step 1 builds the lookup (used by profiles and the people drawer); step 2 builds the Roles page tab.
// Nothing here is built in: which roles exist, who holds them and who reports to whom are all data (Roles & rights).
export function buildReporting(ctx) {
  const { has, s, visibleRoles } = ctx;
  const vis = (id) => visibleRoles.some((r) => r.id === id);
  const roleName = (id) => (visibleRoles.find((r) => r.id === id) || {}).name || id;
  const ridOf = (n) => { const rid = s.memberRoles[n] || ''; return vis(rid) ? rid : ''; };
  const mi = s.meInfo || {};
  const ME = (this.srv && this.srv.me && this.srv.me.name) || mi.name || '';
  const meDir = { name: ME, title: mi.title || '', dept: mi.dept || '', loc: mi.branch || '', sector: mi.sector || '', unit: mi.unit || '' };
  const personOf = (n) => (n === ME ? meDir : this.dir.find((x) => x.name === n));
  const holdersOf = (rid) => this.dir.map((d) => d).concat([meDir]).filter((d, i, a) => a.findIndex((x) => x.name === d.name) === i).filter((d) => ridOf(d.name) === rid);
  const reportsMap = { ...s.reportsTo };
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
    // A person chosen for the whole role (set on the Reporting lines tab).
    const rp = s.reportsPerson[ridOf(name)];
    const rpick = rp && rp !== name && personOf(rp) && ridOf(rp) === t ? personOf(rp) : null;
    if (rpick) return { name: rpick.name, roleName: roleName(t), display: rpick.name, sub: roleName(t) + (rpick.dept ? ' · ' + rpick.dept : ''), specific: true, top: false };
    return { name: '', roleName: roleName(t), display: roleName(t), sub: 'Role', specific: false, top: false };
  };
  const reportOf = reportOfWith(reportsMap);
  const targetRoleOf = targetRoleWith(reportsMap);
  return { reportOf, reportOfWith, reportsMap, ridOf, roleHolders: holdersOf, targetRoleOf, reportRoleName: roleName };
}

export function buildReportingTab(ctx) {
  const { C, can, reportsMap, roleHolders, s, visibleRoles } = ctx;
  const canEdit = can('roles', 'Manage roles & rights');
  const draft = s.reportDraft || {};
  const pdraft = s.reportPDraft || {};
  const eff = { ...reportsMap, ...draft };
  const vis = (id) => visibleRoles.some((r) => r.id === id);
  const roleName = (id) => (visibleRoles.find((r) => r.id === id) || {}).name || id;
  const effPerson = (rid) => (pdraft[rid] !== undefined ? pdraft[rid] : s.reportsPerson[rid] || '');
  const dirty = Object.keys(draft).some((k) => (draft[k] || '') !== (reportsMap[k] || '')) || Object.keys(pdraft).some((k) => (pdraft[k] || '') !== (s.reportsPerson[k] || ''));
  const setRole = (rid, v) => this.setState((st) => ({ reportDraft: { ...(st.reportDraft || {}), [rid]: v }, reportPDraft: { ...(st.reportPDraft || {}), [rid]: '' } }));
  const setPerson = (rid, v) => this.setState((st) => ({ reportPDraft: { ...(st.reportPDraft || {}), [rid]: v } }));
  const rows = visibleRoles.map((r) => {
    const target = eff[r.id] && vis(eff[r.id]) ? eff[r.id] : '';
    const holders = target ? roleHolders(target).slice(0, 80) : [];
    const roleOpts = [{ v: '', l: 'Nobody (top level)' }].concat(visibleRoles.filter((x) => x.id !== r.id).map((x) => ({ v: x.id, l: x.name })));
    const cur = effPerson(r.id);
    const valid = cur && holders.some((h) => h.name === cur) ? cur : '';
    const personOpts = [{ v: '', l: target ? 'Not chosen (profiles show "' + roleName(target) + '")' : '—' }].concat(holders.map((h) => ({ v: h.name, l: h.name + (h.dept ? ' · ' + h.dept : '') })));
    return { cells: [C.t(r.name, { strong: true, sub: r.desc }), C.t(String(roleHolders(r.id).length), { nowrap: true }),
      C.sel(eff[r.id] || '', roleOpts, (e) => setRole(r.id, e.target.value), { label: r.name + ' reports to', minWidth: '190px', disabled: !canEdit }),
      C.sel(valid, personOpts, (e) => setPerson(r.id, e.target.value), { label: 'Person that ' + r.name + ' reports to', minWidth: '230px', disabled: !canEdit || !target })] };
  });
  const reportingTable = { id: 'reporting', pageSize: 12, minWidth: '860px', cols: [{ label: 'Role', filter: 'text' }, { label: 'People' }, { label: 'Reports to (role)' }, { label: 'Reports to (person)' }], rows };
  // A role cannot end up reporting to itself through a chain.
  const cycleFor = () => { for (const r of visibleRoles) { const seen = {}; let id = r.id; while (eff[id]) { if (seen[eff[id]]) return roleName(r.id); seen[id] = true; id = eff[id]; if (id === r.id) return roleName(r.id); } } return ''; };
  const reportingBar = { canEdit, dirty, discard: () => this.setState({ reportDraft: null, reportPDraft: null }),
    save: () => {
      const bad = cycleFor();
      if (bad) { this.notify('These lines go in a circle (' + bad + '). Change one so the chain ends at the top.', '#FFB86B'); return; }
      this.setState((st) => {
        const roles = { ...reportsMap, ...st.reportsTo, ...(st.reportDraft || {}) };
        const people = { ...st.reportsPerson, ...(st.reportPDraft || {}) };
        // A person only stays chosen while they still hold the role this role reports to.
        Object.keys(people).forEach((rid) => { const t = roles[rid]; if (!people[rid] || !t || !roleHolders(t).some((h) => h.name === people[rid])) delete people[rid]; });
        return { reportsTo: roles, reportsPerson: people, reportDraft: null, reportPDraft: null,
          extraLogs: [{ id: 'x' + Date.now(), ago: 0, user: 'Samwel Okoth Omulo', action: 'Updated reporting lines between roles', module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' }].concat(st.extraLogs) };
      });
      if (this.srv) { rolesW.saveReporting(this, this.srv).then(() => this.notify('Reporting lines saved. Profiles now show the new managers.', '#B8F36B')).catch((e) => this.notify(e.message || 'Could not save the reporting lines.', '#FF9B8A')); return; }
      this.notify('Reporting lines saved. Profiles now show the new managers.', '#B8F36B');
    } };
  return { reportingTable, reportingBar };
}
