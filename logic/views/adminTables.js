import { LEVEL_LABEL, catsFor, levelsOf } from '~/logic/ticketRouting.js';

// renderVals step: adminTables. Receives the names built by earlier steps and returns the ones later steps need.
export function buildAdminTables(ctx) {
  const { C, O, can, agoLabel, defaultCats, deskTickets, myTickets, alertsAll, alertsRange, logsRange, allLogs, chip, corePeople, f8, flowItem, fmtDay, has, modEnabled, modInfo, nowMs, rKey, requestsShown, rightOf, rightsDef, roleScope, ruleDefs, ruleOn, ruleSev, s, sec, sevColor, tickets, titleOf, toggleRight, tz9, visibleRoles } = ctx;
  // =============== Tables for existing pages ===============
  const prioColor = { High: '#FF9B8A', Medium: '#FFB86B', Low: '#9DB0A6' };
  const ticketCols = [{ label: 'ID', filter: 'text' }, { label: 'Title', filter: 'text' }];
  const ticketCells = (t) => [C.t(t.id, { mono: true, nowrap: true }), C.t(t.title, { strong: true })];
  const tailCols = [{ label: 'Priority', filter: 'select' }, { label: 'Status', filter: 'select' }, { label: 'TAT', filter: 'select' }, { label: 'Handling', filter: 'select' }, { label: 'Raised', filter: 'date' }, { label: '', align: 'right' }];
  const tailCells = (t) => [C.b(t.prio, prioColor[t.prio]), C.t(t.statusLabel + (t.escalated ? ' · escalated' : ''), { fv: t.statusLabel }), C.b(t.slaLabel, t._c.color, { sub: t.remainingLabel }), C.t(t.handlerLabel), C.t(fmtDay(t.createdAt), { fv: t.createdAt, nowrap: true }), C.a([{ label: 'View', kind: 'primary', run: t.open }])];
  // My tickets: what you raised, and where each one is now.
  const ticketsTable = { id: 'tickets-mine', pageSize: 10, minWidth: '1060px', emptyText: 'You have not raised any tickets yet.',
    cols: ticketCols.concat([{ label: 'Sent to', filter: 'select' }, { label: 'Category', filter: 'select' }]).concat(tailCols),
    rows: myTickets.map((t) => ({ tint: t._c.sla === 'Overdue' ? '#FF9B8A' : '', cells: ticketCells(t).concat([C.t(t.toLabel, { fv: t.toLabel }), C.t(t.cat)]).concat(tailCells(t)) })) };
  // Ticket desk: every ticket received when the organization has no department, unit, branch or sector. Sorted by priority, then by how urgent the TAT is.
  const prioRank = { High: 0, Medium: 1, Low: 2 };
  const deskRows = deskTickets.slice().sort((a, b) => (a.step >= 3) - (b.step >= 3) || prioRank[a.prio] - prioRank[b.prio] || a._c.rem - b._c.rem);
  const deskTable = { id: 'ticket-desk', pageSize: 10, minWidth: '1100px', emptyText: 'No tickets have been received.',
    cols: ticketCols.concat([{ label: 'From', filter: 'select' }, { label: 'Category', filter: 'select' }]).concat(tailCols),
    rows: deskRows.map((t) => ({ tint: t._c.sla === 'Overdue' ? '#FF9B8A' : '', cells: ticketCells(t).concat([C.t(t.fromDept || t.raisedBy), C.t(t.cat)]).concat(tailCells(t)) })) };
  const deskCats = Array.from(new Set(deskTickets.map((t) => t.cat))).map((c) => { const list = deskTickets.filter((t) => t.cat === c); return { name: c, total: list.length, open: list.filter((t) => t.step < 3).length, overdue: list.filter((t) => t._c.sla === 'Overdue').length }; }).sort((a, b) => b.open - a.open);
  const statusColP = { 'Available': '#B8F36B', 'In a meeting': '#FFB86B', 'Focusing': '#7CC4FF', 'Away': '#9DB0A6' };
  const peopleTable = { id: 'people', pageSize: 15, minWidth: '880px',
    cols: [{ label: 'Person', filter: 'text' }].concat(has.departments ? [{ label: 'Department', filter: 'select' }] : []).concat(has.sectors ? [{ label: 'Sector', filter: 'select' }] : []).concat(has.branches ? [{ label: 'Branch', filter: 'select' }] : []).concat(has.units ? [{ label: 'Unit', filter: 'select' }] : []).concat([{ label: 'Status', filter: 'select' }, { label: 'Account', filter: 'select' }, { label: '', align: 'right' }]),
    rows: this.dir.filter((p) => !(s.peopleAdmin[p.id] || {}).deleted).map((p) => {
      const admin = s.peopleAdmin[p.id] || { active: true, blocked: false };
      return { cells: [C.p(p.name, { sub: p.title, dot: statusColP[p.status], fv: p.name })].concat(has.departments ? [C.t(p.dept)] : []).concat(has.sectors ? [C.t(p.sector)] : []).concat(has.branches ? [C.t(p.loc)] : []).concat(has.units ? [C.t(p.unit || '—', { fv: p.unit || 'No unit' })] : []).concat([
      C.b(p.status, statusColP[p.status]),
      C.b(admin.blocked ? 'Blocked' : admin.active ? 'Active' : 'Inactive', admin.blocked ? '#FF9B8A' : admin.active ? '#B8F36B' : '#9DB0A6'),
      C.a((has.chat ? [{ label: 'Message', run: () => this.openDm(p.name) }, { label: 'Call', run: () => this.startDmCall('voice', p.name, p.initials, p.color) }] : []).concat([{ label: 'Profile', kind: 'primary', run: () => this.setState({ personOpen: p.id }) }]).concat(can('people', 'Edit profiles') ? [{ label: 'Edit', run: () => this.openEditPerson(p.id) }, { label: 'Action', run: () => this.setState({ userActionOpen: p.id, userResetTemp: '' }) }] : []))]) };
    }) };
  const peopleSummary = this.dir.filter((p) => !(s.peopleAdmin[p.id] || {}).deleted).length + ' people' + (has.departments ? ' in ' + O.departments.length + ' departments' : '') + (has.branches ? ' across ' + O.branches.length + ' branches' : '') + ' · ' + this.dir.filter((p) => p.status === 'Available' && (s.peopleAdmin[p.id] || { active: true }).active && !(s.peopleAdmin[p.id] || {}).deleted).length + ' available now';
  const actionPerson = s.userActionOpen === null ? null : this.dir.find((p) => p.id === s.userActionOpen);
  const actionAdmin = actionPerson && (s.peopleAdmin[actionPerson.id] || { active: true, blocked: false, deleted: false });
  const userActions = {
    open: !!(actionPerson && can('people', 'Edit profiles')),
    name: actionPerson ? actionPerson.name : '',
    email: actionPerson ? actionPerson.email : '',
    active: !!(actionAdmin && actionAdmin.active),
    blocked: !!(actionAdmin && actionAdmin.blocked),
    tempPassword: s.userResetTemp,
    activate: () => actionPerson && this.togglePersonActive(actionPerson.id),
    toggleBlocked: () => actionPerson && this.togglePersonBlocked(actionPerson.id),
    remove: () => actionPerson && this.deletePerson(actionPerson.id),
    resetPassword: () => actionPerson && this.resetPersonPassword(actionPerson.id),
    close: () => this.setState({ userActionOpen: null, userResetTemp: '' })
  };
  const isoFromAgo = (m) => new Date(nowMs - m * 60000).toLocaleDateString('en-CA', tz9);
  const logsTable = { id: 'logs', pageSize: 12, minWidth: '1120px', emptyText: 'No log entries match.',
    cols: [{ label: 'When', filter: 'date' }, { label: 'User', filter: 'text' }, { label: 'Action', filter: 'text' }, { label: 'Module', filter: 'select' }, { label: 'IP · Location', filter: 'text' }, { label: 'Result', filter: 'select' }, { label: 'Flag', filter: 'select' }, { label: '', align: 'right' }],
    rows: logsRange.map((l) => {
      const fl = !!(l.rule && ruleOn(l.rule)); const r = fl ? ruleDefs.find((x) => x.id === l.rule) : null; const c = fl ? sevColor[ruleSev(r)] : '';
      return { tint: fl ? c : '', cells: [C.t(l.ago === 0 ? 'Just now' : agoLabel(l.ago), { fv: isoFromAgo(l.ago), nowrap: true, mono: true }), C.t(l.user, { nowrap: true }), C.t(l.action), C.t(l.module), C.t(l.ip + ' · ' + l.loc, { nowrap: true }),
        C.b(l.result, l.result === 'Success' ? '#9DB0A6' : l.result === 'Failed' ? '#FF9B8A' : '#FFB86B'), fl ? C.b(r.name, c, { fv: 'Flagged: ' + r.name }) : C.t('—', { fv: 'Not flagged', color: '#4A5C54' }),
        C.a(fl ? [{ label: 'View alert', kind: 'primary', run: () => this.setState({ secTab: 'alerts' }) }] : l.result !== 'Success' ? [{ label: 'Block IP', kind: 'danger', run: () => this.notify(l.ip + ' blocked. The block was logged.', '#FF9B8A') }] : [{ label: 'Details', run: () => this.notify(l.user + ' · ' + l.action + ' · ' + l.ip, '#7CC4FF') }])] };
    }) };
  const alertsTable = { id: 'alerts', pageSize: 8, minWidth: '1100px', emptyText: 'No alerts. Rules that are switched off do not raise alerts.',
    cols: [{ label: 'ID', filter: 'text' }, { label: 'Severity', filter: 'select' }, { label: 'Alert', filter: 'text' }, { label: 'Rule', filter: 'select' }, { label: 'Raised', filter: 'date' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
    rows: alertsRange.map((a) => ({ tint: a.isOpen ? a.color : '', cells: [C.t(a.id, { mono: true, nowrap: true }), C.b(a.sev, a.color), C.t(a.title, { strong: true, sub: a.detail + (a.lastNote ? ' · Latest comment: ' + a.lastNote.comment : '') }), C.t(a.ruleName), C.t(a.when, { fv: isoFromAgo(a.ago), nowrap: true }),
      C.b(a.status, a.status === 'Open' ? '#FF9B8A' : a.status === 'Investigating' ? '#FFB86B' : '#9DB0A6'),
      C.a((a.canInvestigate ? [{ label: 'Investigate', kind: 'primary', run: a.investigate }] : []).concat(a.canClose ? [{ label: 'Close', kind: 'primary', run: a.closeAlert }] : []).concat(a.canFalsePos ? [{ label: 'False positive', run: a.falsePos }] : []).concat([{ label: 'History', run: a.history }]))] })) };
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
  const rolesTabs = [['roles', 'Roles'], ['rights', 'Rights'], ['members', 'Members'], ['reporting', 'Reporting lines'], ['flows', 'Approval rules']].map(([k, l]) => ({ label: l, go: () => this.setState({ rolesTab: k }),
    style: 'min-height:44px;padding:0 16px;border:0;background:transparent;cursor:pointer;font-size:14px;font-family:inherit;border-bottom:2px solid ' + (s.rolesTab === k ? '#B8F36B' : 'transparent') + ';color:' + (s.rolesTab === k ? '#E8EFEA' : '#9DB0A6') + ';font-weight:' + (s.rolesTab === k ? '600' : '500') }));
  const rolesView = { isRoles: s.rolesTab === 'roles', isRights: s.rolesTab === 'rights', isMembers: s.rolesTab === 'members', isReporting: s.rolesTab === 'reporting', isFlows: s.rolesTab === 'flows' };
  const tLevels = levelsOf(has);
  const ntType = tLevels.indexOf(s.nt.toType) !== -1 ? s.nt.toType : (tLevels[0] || 'desk');
  const ntRecs = ntType === 'desk' ? [] : (O[ntType] || []);
  const ntName = ntRecs.some((r) => r.name === s.nt.toName) ? s.nt.toName : (ntRecs[0] ? ntRecs[0].name : '');
  const ntRecip = ntType === 'desk' ? { type: 'desk', name: 'Ticket desk' } : { type: ntType, name: ntName };
  const ntCatList = catsFor(s.unitCats, ntRecip, defaultCats);
  const ntCatSel = ntCatList.find((c) => c.name === s.nt.cat) || ntCatList[0];
  const ntCats = ntCatList.map((c) => ({ label: c.name + ' · ' + c.tat + 'h' + (c.custom ? ' · ' + c.prio + ' priority' : ''), pick: () => this.setState((x) => ({ nt: { ...x.nt, cat: c.name } })), style: chip(ntCatSel.name === c.name) }));
  const ntTargets = { hasLevels: tLevels.length > 0, multiple: tLevels.length > 1, type: ntType, name: ntName,
    types: tLevels.map((l) => ({ label: LEVEL_LABEL[l], pick: () => this.setState((x) => ({ nt: { ...x.nt, toType: l, toName: '' } })), style: chip(ntType === l) })),
    recs: ntRecs.map((r) => ({ label: r.name, pick: () => this.setState((x) => ({ nt: { ...x.nt, toType: ntType, toName: r.name } })), style: chip(ntName === r.name) })),
    recsLabel: ntType === 'desk' ? '' : 'Send to ' + LEVEL_LABEL[ntType].toLowerCase(),
    summary: ntType === 'desk' ? 'This organization has no departments, units, branches or sectors, so tickets go to the ticket desk.' : 'It will show on the ' + LEVEL_LABEL[ntType].toLowerCase() + ' dashboard of ' + ntName + '.' };
  return { ntCats, ntCatSel, deskCats, deskTable, addToRole, alertsTable, approvalsTable, flowsTable, logsTable, membersTable, ntTargets, peopleSummary, peopleTable, userActions, rightsHint, rightsTable, roleBar, rolesTable, rolesTabs, rolesView, rulesTable, setRoleSel, ticketsTable };
}
