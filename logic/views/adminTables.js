// renderVals step: adminTables. Receives the names built by earlier steps and returns the ones later steps need.
export function buildAdminTables(ctx) {
  const { C, O, agoLabel, alertsAll, allLogs, chip, corePeople, f8, flowItem, fmtDay, has, modEnabled, modInfo, nowMs, rKey, requestsShown, rightOf, rightsDef, roleScope, ruleDefs, ruleOn, ruleSev, s, sec, sevColor, tickets, titleOf, toggleRight, tz9, visibleRoles } = ctx;
  // =============== Tables for existing pages ===============
  const ticketsTable = { id: 'tickets', pageSize: 10, minWidth: has.departments ? '1180px' : '1060px', emptyText: 'No tickets match.',
    cols: [{ label: 'ID', filter: 'text' }, { label: 'Title', filter: 'text' }].concat(has.departments ? [{ label: 'Department', filter: 'select' }, { label: 'From', filter: 'select' }] : []).concat([{ label: 'Category', filter: 'select' }, { label: 'Priority', filter: 'select' }, { label: 'Status', filter: 'select' }, { label: 'TAT', filter: 'select' }, { label: 'Handling', filter: 'select' }, { label: 'Raised', filter: 'date' }, { label: '', align: 'right' }]),
    rows: tickets.map((t) => ({ tint: t._c.sla === 'Overdue' ? '#FF9B8A' : '', cells: [C.t(t.id, { mono: true, nowrap: true }), C.t(t.title, { strong: true })].concat(has.departments ? [C.t(t.dept), C.t(t.fromDept)] : []).concat([
      C.t(t.cat), C.b(t.prio, { High: '#FF9B8A', Medium: '#FFB86B', Low: '#9DB0A6' }[t.prio]), C.t(t.statusLabel + (t.escalated ? ' · escalated' : ''), { fv: t.statusLabel }), C.b(t.slaLabel, t._c.color, { sub: t.remainingLabel }), C.t(t.handlerLabel), C.t(fmtDay(t.createdAt), { fv: t.createdAt, nowrap: true }),
      C.a([{ label: 'View', kind: 'primary', run: t.open }])]) })) };
  const statusColP = { 'Available': '#B8F36B', 'In a meeting': '#FFB86B', 'Focusing': '#7CC4FF', 'Away': '#9DB0A6' };
  const peopleTable = { id: 'people', pageSize: 15, minWidth: '880px',
    cols: [{ label: 'Person', filter: 'text' }].concat(has.departments ? [{ label: 'Department', filter: 'select' }] : []).concat(has.sectors ? [{ label: 'Sector', filter: 'select' }] : []).concat(has.branches ? [{ label: 'Branch', filter: 'select' }] : []).concat(has.units ? [{ label: 'Unit', filter: 'select' }] : []).concat([{ label: 'Status', filter: 'select' }, { label: '', align: 'right' }]),
    rows: this.dir.map((p) => ({ cells: [C.p(p.name, { sub: p.title, dot: statusColP[p.status], fv: p.name })].concat(has.departments ? [C.t(p.dept)] : []).concat(has.sectors ? [C.t(p.sector)] : []).concat(has.branches ? [C.t(p.loc)] : []).concat(has.units ? [C.t(p.unit || '—', { fv: p.unit || 'No unit' })] : []).concat([
      C.b(p.status, statusColP[p.status]),
      C.a((has.chat ? [{ label: 'Message', run: () => this.openDm(p.name) }, { label: 'Call', run: () => this.startDmCall('voice', p.name, p.initials, p.color) }] : []).concat([{ label: 'Profile', kind: 'primary', run: () => this.setState({ personOpen: p.id }) }]))]) })) };
  const peopleSummary = this.dir.length + ' people' + (has.departments ? ' in ' + O.departments.length + ' departments' : '') + (has.branches ? ' across ' + O.branches.length + ' branches' : '') + ' · ' + this.dir.filter((p) => p.status === 'Available').length + ' available now';
  const isoFromAgo = (m) => new Date(nowMs - m * 60000).toLocaleDateString('en-CA', tz9);
  const logsTable = { id: 'logs', pageSize: 12, minWidth: '1120px', emptyText: 'No log entries match.',
    cols: [{ label: 'When', filter: 'date' }, { label: 'User', filter: 'text' }, { label: 'Action', filter: 'text' }, { label: 'Module', filter: 'select' }, { label: 'IP · Location', filter: 'text' }, { label: 'Result', filter: 'select' }, { label: 'Flag', filter: 'select' }, { label: '', align: 'right' }],
    rows: allLogs.map((l) => {
      const fl = !!(l.rule && ruleOn(l.rule)); const r = fl ? ruleDefs.find((x) => x.id === l.rule) : null; const c = fl ? sevColor[ruleSev(r)] : '';
      return { tint: fl ? c : '', cells: [C.t(l.ago === 0 ? 'Just now' : agoLabel(l.ago), { fv: isoFromAgo(l.ago), nowrap: true, mono: true }), C.t(l.user, { nowrap: true }), C.t(l.action), C.t(l.module), C.t(l.ip + ' · ' + l.loc, { nowrap: true }),
        C.b(l.result, l.result === 'Success' ? '#9DB0A6' : l.result === 'Failed' ? '#FF9B8A' : '#FFB86B'), fl ? C.b(r.name, c, { fv: 'Flagged: ' + r.name }) : C.t('—', { fv: 'Not flagged', color: '#4A5C54' }),
        C.a(fl ? [{ label: 'View alert', kind: 'primary', run: () => this.setState({ secTab: 'alerts' }) }] : l.result !== 'Success' ? [{ label: 'Block IP', kind: 'danger', run: () => this.notify(l.ip + ' blocked. The block was logged.', '#FF9B8A') }] : [{ label: 'Details', run: () => this.notify(l.user + ' · ' + l.action + ' · ' + l.ip, '#7CC4FF') }])] };
    }) };
  const alertsTable = { id: 'alerts', pageSize: 8, minWidth: '1100px', emptyText: 'No alerts. Rules that are switched off do not raise alerts.',
    cols: [{ label: 'ID', filter: 'text' }, { label: 'Severity', filter: 'select' }, { label: 'Alert', filter: 'text' }, { label: 'Rule', filter: 'select' }, { label: 'Raised', filter: 'date' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
    rows: alertsAll.map((a) => ({ tint: a.isOpen ? a.color : '', cells: [C.t(a.id, { mono: true, nowrap: true }), C.b(a.sev, a.color), C.t(a.title, { strong: true, sub: a.detail }), C.t(a.ruleName), C.t(a.when, { fv: isoFromAgo(a.ago), nowrap: true }),
      C.b(a.status, a.status === 'Open' ? '#FF9B8A' : a.status === 'Investigating' ? '#FFB86B' : '#9DB0A6'),
      C.a([{ label: 'Investigate', run: a.investigate, disabled: a.status === 'Investigating' }, { label: 'Resolve', kind: 'primary', run: a.resolve, disabled: !a.isOpen }, { label: 'False positive', run: a.falsePos, disabled: !a.isOpen }])] })) };
  const sevOpts = [{ v: 'High', l: 'High' }, { v: 'Medium', l: 'Medium' }, { v: 'Low', l: 'Low' }];
  const rulesTable = { id: 'rules', pageSize: 10, minWidth: '860px',
    cols: [{ label: 'Rule', filter: 'text' }, { label: 'Severity', filter: 'select' }, { label: 'Matches today' }, { label: 'On', filter: 'select', align: 'center' }],
    rows: sec.rules.map((r) => ({ cells: [C.t(r.name, { strong: true, sub: r.desc }), C.sel(r.sev, sevOpts, r.setSev, { label: 'Severity for ' + r.name, minWidth: '110px' }), C.t(r.hits), C.chk(r.on, r.toggle, { label: 'Turn ' + r.name + (r.on ? ' off' : ' on') })] })) };
  const roleCur = visibleRoles.find((r) => r.id === s.roleSel) || visibleRoles.find((r) => r.id === 'staff') || visibleRoles[0];
  const rolesTable = { id: 'roles', pageSize: 10, minWidth: '900px',
    cols: [{ label: 'Role', filter: 'text' }, { label: 'Scope', filter: 'select' }, { label: 'People' }, { label: 'Rights granted' }, { label: '', align: 'right' }],
    rows: visibleRoles.map((r) => { const mods = Object.keys(rightsDef).filter(modEnabled); let on = 0, tot = 0; mods.forEach((m) => rightsDef[m].forEach((x) => { tot++; if (rightOf(r.id, m, x)) on++; })); const pc = Math.round(on / tot * 100);
      return { tint: r.id === roleCur.id ? '#B8F36B' : '', cells: [C.t(r.name, { strong: true, sub: r.desc }), C.t(roleScope[r.id] || 'Whole organization'), C.t(String(r.count)), C.bar(pc, pc > 80 ? '#FFB86B' : '#B8F36B', on + ' / ' + tot),
        C.a([{ label: 'Edit rights', kind: 'primary', run: () => this.setState({ roleSel: r.id, rolesTab: 'rights' }) }, { label: 'Members', run: () => this.setState({ roleSel: r.id, rolesTab: 'members' }) }])] }; }) };
  const rightsRows = [];
  Object.keys(rightsDef).filter(modEnabled).forEach((m) => rightsDef[m].forEach((x) => {
    const pend = s.pendingRights[rKey(roleCur.id, m, x)] !== undefined;
    rightsRows.push({ cells: [C.t(modInfo[m][0], { fv: modInfo[m][0] }), C.t(x, { strong: true }), C.chk(rightOf(roleCur.id, m, x), () => toggleRight(roleCur.id, m, x), { disabled: !!roleCur.locked, pending: pend, label: roleCur.name + ': ' + modInfo[m][0] + ' · ' + x })] });
  }));
  const rightsTable = { id: 'rights-' + roleCur.id, pageSize: 12, minWidth: '560px',
    cols: [{ label: 'Widget', filter: 'select' }, { label: 'Right', filter: 'text' }, { label: roleCur.name + ' can', align: 'center' }], rows: rightsRows };
  const rightsHint = 'Setting rights for ' + roleCur.name + '. Tick a box to let this role do that action. ' + (f8 && f8.levels > 0 ? 'Changes need approval (' + (f8.levels === 1 ? 'direct' : f8.levels + ' levels') + ') and show as PENDING until approved.' : 'Changes apply straight away and are logged.') + ' Widgets your organization does not use are hidden.';
  const defaultRole = { 'Samwel Okoth Omulo': 'admin', 'Brian Kiprop': 'it', 'David Otieno': 'staff', 'Amina Wanjiru': 'finance', 'Grace Njeri': 'hod', 'Faith Achieng': 'staff', 'Kevin Mutua': 'staff', 'Mercy Wambui': 'staff', 'John Kamau': 'staff', 'Lilian Chebet': 'staff' };
  const memberList = ['Samwel Okoth Omulo'].concat(corePeople).concat(this.dir.slice(0, 40).map((d) => d.name)).filter((n, i, a) => a.indexOf(n) === i);
  const roleOpts = visibleRoles.map((r) => ({ v: r.id, l: r.name }));
  const ridOf = (n) => {
    const d = this.dir.find((x) => x.name === n);
    const rid = s.memberRoles[n] || defaultRole[n] || (d && /Lead|Supervisor/.test(d.title) ? (has.branches ? 'branch' : 'hod') : 'staff');
    return visibleRoles.find((r) => r.id === rid) ? rid : 'staff';
  };
  const assignRole = (n, v) => {
    const rn = (visibleRoles.find((r) => r.id === v) || {}).name;
    this.setState((x) => ({ memberRoles: { ...x.memberRoles, [n]: v }, extraLogs: [{ id: 'x' + Date.now(), ago: 0, user: 'Samwel Okoth Omulo', action: 'Gave ' + n + ' the ' + rn + ' role', module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' }].concat(x.extraLogs) }));
    this.notify(n + ' is now ' + rn + '. Logged.', '#B8F36B');
  };
  const roleMembers = memberList.filter((n) => ridOf(n) === roleCur.id);
  const roleAddOpts = memberList.filter((n) => ridOf(n) !== roleCur.id && n !== 'Samwel Okoth Omulo').map((n) => ({ v: n, l: n }));
  const membersTable = { id: 'members-' + roleCur.id, pageSize: 12, minWidth: '760px', emptyText: 'Nobody has the ' + roleCur.name + ' role yet. Use "Add a person" above.',
    cols: [{ label: 'Person', filter: 'text' }, { label: 'Department', filter: 'select' }, { label: 'Move to another role' }, { label: '', align: 'right' }],
    rows: roleMembers.map((n) => {
      const d = this.dir.find((x) => x.name === n);
      const mine = n === 'Samwel Okoth Omulo';
      return { cells: [C.p(n, { sub: d ? d.title : titleOf(n), fv: n }),
        C.t(d ? d.dept : ({ 'Brian Kiprop': 'IT Operations', 'David Otieno': 'Software Development', 'Amina Wanjiru': 'Finance', 'Grace Njeri': 'HR', 'Faith Achieng': 'Customer Care', 'Kevin Mutua': 'Sales', 'Samwel Okoth Omulo': 'Risk & Compliance' }[n] || 'Operations')),
        C.sel(roleCur.id, roleOpts, (e) => assignRole(n, e.target.value), { fv: roleCur.name, label: 'Role for ' + n, minWidth: '170px', disabled: mine }),
        C.a([{ label: 'Remove from role', kind: 'danger', disabled: mine || roleCur.id === 'staff', run: () => { this.setState((x) => ({ memberRoles: { ...x.memberRoles, [n]: 'staff' } })); this.notify(n + ' moved back to Staff.', '#7CC4FF'); } }])] };
    }) };
  const roleBar = { name: roleCur.name, desc: roleCur.desc, scope: roleScope[roleCur.id] || 'Whole organization', count: roleMembers.length, options: visibleRoles.map((r) => ({ v: r.id, l: r.name })), value: roleCur.id, addOpts: roleAddOpts, locked: !!roleCur.locked };
  const setRoleSel = (e) => this.setState({ roleSel: e.target.value });
  const addToRole = (e) => { const n = e.target.value; if (n) assignRole(n, roleCur.id); };
  const approverOpts = ['Line manager', 'Head of department', 'Finance approver', 'HR Manager', 'Communications lead', 'Security officer', 'Head of IT Operations', 'Managing Director'].concat(has.sectors ? ['Sector manager'] : []).concat(has.branches ? ['Branch manager'] : []).concat(has.units ? ['Unit lead'] : []).map((x) => ({ v: x, l: x }));
  const levelOpts = [{ v: '0', l: 'No approval' }, { v: '1', l: 'Direct (1 level)' }, { v: '2', l: '2 levels' }, { v: '3', l: '3 levels' }];
  const flowsTable = { id: 'flows', pageSize: 10, minWidth: '1180px',
    cols: [{ label: 'Action', filter: 'text' }, { label: 'Widget', filter: 'select' }, { label: 'Approval', filter: 'select' }, { label: 'Level 1' }, { label: 'Level 2' }, { label: 'Level 3' }],
    rows: s.flows.filter((f) => modEnabled(f.mod)).map((f) => {
      const fi = flowItem(f);
      const lv = (i) => C.sel(f.approvers[i] || 'Line manager', approverOpts, (e) => { const v = e.target.value; this.setState((st) => ({ flows: st.flows.map((x) => (x.id === f.id ? { ...x, approvers: x.approvers.map((a, j) => (j === i ? v : a)) } : x)) })); }, { disabled: f.levels <= i, label: 'Level ' + (i + 1) + ' approver for ' + f.action, minWidth: '160px' });
      return { cells: [C.t(f.action, { strong: true }), C.t(fi.modName), C.sel(String(f.levels), levelOpts, fi.setLevels, { fv: levelOpts[f.levels].l, label: 'Approval levels for ' + f.action, minWidth: '150px' }), lv(0), lv(1), lv(2)] };
    }) };
  const approvalsTable = { id: 'approvals-' + s.apTab, pageSize: 8, minWidth: '1080px', emptyText: s.apTab === 'waiting' ? 'Nothing is waiting for approval.' : 'No requests yet.',
    cols: [{ label: 'ID', filter: 'text' }, { label: 'Request', filter: 'text' }, { label: 'Widget', filter: 'select' }, { label: 'Requested by', filter: 'text' }, { label: 'Approval path' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
    rows: requestsShown.map((rq) => {
      const path = rq.steps.map((st) => (st.n === '✓' ? '✓ ' : st.n + '. ') + st.who).join('  →  ');
      const cur = rq.steps.find((st) => st.state === 'Waiting now');
      const stFv = rq.status;
      return { tint: rq.canAct ? '#FFB86B' : '', cells: [C.t(rq.id, { mono: true, nowrap: true }), C.t(rq.title, { strong: true, sub: rq.hasNote ? rq.note : '' }), C.t(rq.modName),
        C.p(rq.by, { sub: rq.when, fv: rq.by }), C.t(path, { sub: rq.levelsLabel + (cur ? ' · waiting for ' + cur.who : '') }),
        C.b(rq.statusLabel, rq.status === 'Approved' ? '#B8F36B' : rq.status === 'Rejected' ? '#FF9B8A' : '#FFB86B', { fv: stFv }),
        C.a(rq.canAct ? [{ label: rq.approveLabel, kind: 'primary', run: rq.approve }, { label: 'Reject', kind: 'danger', run: rq.reject }] : [])] };
    }) };
  const rolesTabs = [['roles', 'Roles'], ['rights', 'Rights'], ['members', 'Members'], ['flows', 'Approval rules']].map(([k, l]) => ({ label: l, go: () => this.setState({ rolesTab: k }),
    style: 'min-height:44px;padding:0 16px;border:0;background:transparent;cursor:pointer;font-size:14px;font-family:inherit;border-bottom:2px solid ' + (s.rolesTab === k ? '#B8F36B' : 'transparent') + ';color:' + (s.rolesTab === k ? '#E8EFEA' : '#9DB0A6') + ';font-weight:' + (s.rolesTab === k ? '600' : '500') }));
  const rolesView = { isRoles: s.rolesTab === 'roles', isRights: s.rolesTab === 'rights', isMembers: s.rolesTab === 'members', isFlows: s.rolesTab === 'flows' };
  const ntDepts = has.departments ? O.departments.map((d) => ({ label: d.name, pick: () => this.setState((x) => ({ nt: { ...x.nt, dept: d.name } })), style: chip((s.nt.dept || 'IT Operations') === d.name) })) : [];
  return { addToRole, alertsTable, approvalsTable, flowsTable, logsTable, membersTable, ntDepts, peopleSummary, peopleTable, rightsHint, rightsTable, roleBar, rolesTable, rolesTabs, rolesView, rulesTable, setRoleSel, ticketsTable };
}
