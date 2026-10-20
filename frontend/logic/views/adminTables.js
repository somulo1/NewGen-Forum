import * as rolesW from '~/logic/server/widgets/roles.js';
import { LEVEL_LABEL, catsFor, levelsOf } from '~/logic/ticketRouting.js';

// renderVals step: adminTables. Receives the names built by earlier steps and returns the ones later steps need.
export function buildAdminTables(ctx) {
  const { C, O, can, agoLabel, rightMeta, defaultCats, deskTickets, myTickets, alertsAll, alertsRange, logsRange, allLogs, chip, corePeople, f8, flowItem, fmtDay, has, modEnabled, modInfo, nowMs, rKey, requestsShown, rightOf, rightsDef, roleScope, ruleDefs, ruleOn, ruleSev, s, sec, sevColor, tickets, titleOf, toggleRight, tz9, visibleRoles } = ctx;
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
  // Sign-ins that were stopped this week: people of the organization locked out after wrong passwords, and addresses that kept
  // trying (including emails nobody owns). Blocks lift by themselves after two minutes; the record stays.
  const blocksTable = { id: 'blocks', pageSize: 12, minWidth: '980px', emptyText: 'Nobody was blocked this week.',
    cols: [{ label: 'When', filter: 'date' }, { label: 'Who', filter: 'text' }, { label: 'Kind', filter: 'select' }, { label: 'Address', filter: 'text' }, { label: 'Reason', filter: 'text' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
    rows: (s.signBlocks || []).map((b) => {
      const at = new Date(b.at), until = new Date(b.until), lifted = until.getTime() <= Date.now();
      const iso = b.at.slice(0, 10);
      return { cells: [C.t(at.toLocaleDateString('en-GB', { day: 'numeric', month: 'short' }) + ' · ' + at.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' }), { fv: iso, nowrap: true }),
        C.t(b.who, { strong: true }), C.b(b.member ? 'Organization user' : 'Not a user', b.member ? '#7CC4FF' : '#FFB86B', { fv: b.member ? 'Organization user' : 'Not a user' }),
        C.t(b.ip || '—', { mono: true, nowrap: true }), C.t(b.reason),
        C.b(lifted ? 'Unlocked automatically' : 'Blocked until ' + until.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' }), lifted ? '#B8F36B' : '#FF9B8A', { fv: lifted ? 'Unlocked automatically' : 'Blocked' }),
        C.a([{ label: 'History', kind: 'primary', run: () => this.setState({ blockOpen: { who: b.who, member: b.member, ip: b.ip } }) }])] };
    }) };
  // The history of one blocked person (or address): every block this week, and for members their sign-in log entries.
  const bo = s.blockOpen;
  const fmtAt = (ms) => new Date(ms).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' }) + ' · ' + new Date(ms).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });
  const blockHistory = bo ? (() => {
    const mine = (s.signBlocks || []).filter((b) => (bo.member ? b.who === bo.who && b.member : b.who === bo.who || (bo.ip && b.ip === bo.ip)));
    const entries = mine.map((b) => ({ at: new Date(b.at).getTime(), text: 'Blocked: ' + b.reason, sub: 'From ' + (b.ip || 'unknown address') + ' · ' + (new Date(b.until).getTime() <= Date.now() ? 'unlocked automatically' : 'blocked until ' + new Date(b.until).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })), color: '#FF9B8A' }));
    const logs = bo.member ? (this.baseLogs || []).filter((l) => l.user === bo.who && l.module === 'Auth').map((l) => ({ at: Date.now() - l.ago * 60000, text: l.action, sub: (l.ip || '') + (l.loc ? ' · ' + l.loc : '') + ' · ' + l.result, color: l.result === 'Success' ? '#B8F36B' : '#FFB86B' })) : [];
    const all = entries.concat(logs).sort((a, b) => b.at - a.at).map((e) => ({ ...e, when: fmtAt(e.at) }));
    const ips = [...new Set(mine.map((b) => b.ip).filter(Boolean))];
    return { open: true, who: bo.who, kind: bo.member ? 'Organization user' : 'Not a user', summary: mine.length + (mine.length === 1 ? ' block' : ' blocks') + ' this week' + (ips.length ? ' · from ' + ips.join(', ') : ''), entries: all, none: all.length === 0, close: () => this.setState({ blockOpen: null }) };
  })() : { open: false };
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
  // Roles and rights are data (made on this page, stored on the server): nothing about them is built in.
  const roleCur = visibleRoles.find((r) => r.id === s.roleSel) || visibleRoles.find((r) => r.isDefault) || visibleRoles[0] || { id: '', name: 'No role selected', desc: 'Create a role first.', locked: false, count: 0, isDefault: false };
  const fail = (e) => this.notify((e && e.message) || 'That did not go through.', '#FF9B8A');
  const srv = this.srv;
  const rolesTable = { id: 'roles', pageSize: 10, minWidth: '900px', emptyText: 'No roles yet. Create the first one above.',
    cols: [{ label: 'Role', filter: 'text' }, { label: 'Scope', filter: 'select' }, { label: 'People' }, { label: 'Rights granted' }, { label: '', align: 'right' }],
    rows: visibleRoles.map((r) => { const mods = Object.keys(rightsDef).filter(modEnabled); let on = 0, tot = 0; mods.forEach((m) => rightsDef[m].forEach((x) => { tot++; if (rightOf(r.id, m, x)) on++; })); const pc = tot ? Math.round(on / tot * 100) : 0;
      return { tint: r.id === roleCur.id ? '#B8F36B' : '', cells: [C.t(r.name + (r.isDefault ? ' · default' : '') + (r.locked ? ' · owner' : ''), { strong: true, sub: r.desc, fv: r.name }), C.t(roleScope[r.id] || 'Whole organization'), C.t(String(r.count)), C.bar(pc, pc > 80 ? '#FFB86B' : '#B8F36B', on + ' / ' + tot),
        C.a([{ label: 'Edit rights', kind: 'primary', run: () => this.setState({ roleSel: r.id, rolesTab: 'rights' }) }, { label: 'Members', run: () => this.setState({ roleSel: r.id, rolesTab: 'members' }) }]
          .concat(!r.locked && !r.count ? [{ label: 'Delete', kind: 'danger', run: () => rolesW.deleteRole(this, srv, r.id).then(() => this.notify('Role deleted.', '#B8F36B')).catch(fail) }] : []))] }; }) };
  // Rights grouped by what they let a person do: see a page, create, edit, delete, export, approve a level, manage.
  const KIND_LABEL = { view: 'See / open', create: 'Create / send', edit: 'Edit / update', delete: 'Delete', export: 'Export', approve: 'Approve', manage: 'Manage', other: 'Other' };
  const kindOptions = ((s.catalog && s.catalog.kinds) || Object.keys(KIND_LABEL)).map((k) => ({ v: k, l: KIND_LABEL[k] || k }));
  const rightsRows = [];
  Object.keys(rightsDef).filter(modEnabled).forEach((m) => rightsDef[m].forEach((x) => {
    const pend = s.pendingRights[rKey(roleCur.id, m, x)] !== undefined;
    const meta = (rightMeta[m] || {})[x] || {};
    rightsRows.push({ cells: [C.t(modInfo[m][0], { fv: modInfo[m][0] }), C.t(meta.label || x, { strong: true }), C.t(KIND_LABEL[meta.kind] || 'Other', { fv: KIND_LABEL[meta.kind] || 'Other' }),
      C.chk(rightOf(roleCur.id, m, x), () => toggleRight(roleCur.id, m, x), { disabled: !!roleCur.locked || !roleCur.id, pending: pend, label: roleCur.name + ': ' + modInfo[m][0] + ' · ' + x }),
      C.a(meta.builtin === false ? [{ label: 'Remove right', kind: 'danger', run: () => rolesW.removeRight(this, srv, m, x).then(() => this.notify('Right removed.', '#B8F36B')).catch(fail) }] : [])] });
  }));
  const rightsTable = { id: 'rights-' + roleCur.id, pageSize: 12, minWidth: '680px', emptyText: 'The rights catalog is empty.',
    cols: [{ label: 'Page', filter: 'select' }, { label: 'Right', filter: 'text' }, { label: 'Kind', filter: 'select' }, { label: roleCur.name + ' can', align: 'center' }, { label: '', align: 'right' }], rows: rightsRows };
  const rightsHint = 'Setting rights for ' + roleCur.name + '. Tick a box to let this role do that. ' + (f8 && f8.levels > 0 ? 'Changes need approval (' + (f8.levels === 1 ? 'direct' : f8.levels + ' levels') + ') and show as pending until approved. ' : '') + (roleCur.locked ? 'The owner role can always do everything.' : '');
  const meName = srv ? srv.me.name : '';
  const memberList = [meName].concat(this.dir.map((d) => d.name)).filter((n, i, a) => n && a.indexOf(n) === i);
  const roleOpts = [{ v: '', l: '— none —' }].concat(visibleRoles.map((r) => ({ v: r.id, l: r.name })));
  const ridOf = (n) => { const rid = s.memberRoles[n] || ''; return visibleRoles.find((r) => r.id === rid) ? rid : ''; };
  const assignRole = (n, v) => {
    const rn = (visibleRoles.find((r) => r.id === v) || {}).name || 'no role';
    rolesW.setRole(this, srv, n, v).then((r) => this.notify(r === 'applied' ? n + ' now has ' + rn + '. Logged.' : 'Sent for approval: ' + n + ' as ' + rn + '.', r === 'applied' ? '#B8F36B' : '#FFB86B')).catch(fail);
  };
  const roleMembers = memberList.filter((n) => ridOf(n) === roleCur.id);
  const roleAddOpts = memberList.filter((n) => ridOf(n) !== roleCur.id && n !== meName).map((n) => ({ v: n, l: n }));
  const membersTable = { id: 'members-' + roleCur.id, pageSize: 12, minWidth: '760px', emptyText: 'Nobody has the ' + roleCur.name + ' role yet. Use "Add a person" above.',
    cols: [{ label: 'Person', filter: 'text' }, { label: 'Department', filter: 'select' }, { label: 'Move to another role' }, { label: '', align: 'right' }],
    rows: roleMembers.map((n) => {
      const d = this.dir.find((x) => x.name === n);
      const mine = n === meName;
      return { cells: [C.p(n, { sub: d ? d.title : (s.meInfo || {}).title || '', fv: n }),
        C.t(d ? d.dept : (s.meInfo || {}).dept || '', {}),
        C.sel(roleCur.id, roleOpts, (e) => assignRole(n, e.target.value), { fv: roleCur.name, label: 'Role for ' + n, minWidth: '170px', disabled: mine }),
        C.a([{ label: 'Remove from role', kind: 'danger', disabled: mine, run: () => assignRole(n, '') }])] };
    }) };
  const nr = s.newRole || { name: '', desc: '', err: '' };
  const setNr = (patch) => this.setState((st) => ({ newRole: { ...st.newRole, ...patch, err: '' } }));
  const roleBar = { name: roleCur.name, desc: roleCur.desc, scope: roleScope[roleCur.id] || 'Whole organization', count: roleMembers.length, options: visibleRoles.map((r) => ({ v: r.id, l: r.name })), value: roleCur.id, addOpts: roleAddOpts, locked: !!roleCur.locked,
    hasRole: !!roleCur.id, isDefault: !!roleCur.isDefault, canDelete: !!roleCur.id && !roleCur.locked && !roleCur.count,
    toggleDefault: () => rolesW.updateRole(this, srv, roleCur.id, { default: !roleCur.isDefault }).then(() => this.notify(roleCur.isDefault ? 'New people no longer get a role automatically.' : roleCur.name + ' is now the role new people get.', '#B8F36B')).catch(fail),
    remove: () => rolesW.deleteRole(this, srv, roleCur.id).then(() => this.notify('Role deleted.', '#B8F36B')).catch(fail),
    editName: nr.editing ? nr.editName : roleCur.name, editDesc: nr.editing ? nr.editDesc : roleCur.desc, editing: !!nr.editing,
    startEdit: () => this.setState((st) => ({ newRole: { ...st.newRole, editing: true, editName: roleCur.name, editDesc: roleCur.desc } })),
    onEditName: (e) => setNr({ editName: e.target.value }), onEditDesc: (e) => setNr({ editDesc: e.target.value }),
    cancelEdit: () => setNr({ editing: false }),
    saveEdit: () => rolesW.updateRole(this, srv, roleCur.id, { name: nr.editName, desc: nr.editDesc }).then(() => { setNr({ editing: false }); this.notify('Role saved.', '#B8F36B'); }).catch((e) => this.setState((st) => ({ newRole: { ...st.newRole, err: e.message } }))) };
  const newRoleForm = { name: nr.name, desc: nr.desc, err: nr.err, hasErr: !!nr.err, copyFrom: nr.copy || '', copyOpts: [{ v: '', l: 'Start with no rights' }].concat(visibleRoles.filter((r) => !r.locked).map((r) => ({ v: r.id, l: 'Copy the rights of ' + r.name }))),
    onName: (e) => setNr({ name: e.target.value }), onDesc: (e) => setNr({ desc: e.target.value }), onCopy: (e) => setNr({ copy: e.target.value }),
    create: () => rolesW.createRole(this, srv, { name: nr.name, desc: nr.desc, copy_from: nr.copy || '' })
      .then((id) => { this.setState({ newRole: { name: '', desc: '', err: '', copy: '' }, roleSel: id, rolesTab: 'rights' }); this.notify('Role created. Tick the rights it should have.', '#B8F36B'); })
      .catch((e) => this.setState((st) => ({ newRole: { ...st.newRole, err: e.message || 'Could not create the role.' } }))) };
  // Add your own right (a named permission) to a page, or your own page.
  const nrt = s.newRight || { module: '', name: '', kind: '', err: '' };
  const setNrt = (patch) => this.setState((st) => ({ newRight: { ...st.newRight, ...patch, err: '' } }));
  const moduleOptions = ((s.catalog && s.catalog.modules) || []).map((m) => ({ v: m.key, l: m.label }));
  const newRightForm = { module: nrt.module || (moduleOptions[0] || {}).v || '', name: nrt.name, kind: nrt.kind || '', err: nrt.err, hasErr: !!nrt.err, moduleOptions, kindOptions: [{ v: '', l: 'Kind: choose for me' }].concat(kindOptions),
    onModule: (e) => setNrt({ module: e.target.value }), onName: (e) => setNrt({ name: e.target.value }), onKind: (e) => setNrt({ kind: e.target.value }),
    add: () => rolesW.addRight(this, srv, { module: nrt.module || (moduleOptions[0] || {}).v, name: nrt.name, kind: nrt.kind })
      .then(() => { this.setState({ newRight: { module: nrt.module, name: '', kind: '', err: '' } }); this.notify('Right added. Give it to roles in the table below.', '#B8F36B'); })
      .catch((e) => this.setState((st) => ({ newRight: { ...st.newRight, err: e.message || 'Could not add the right.' } }))) };
  const nm = s.newModule || { name: '', err: '' };
  const newModuleForm = { name: nm.name, err: nm.err, hasErr: !!nm.err, onName: (e) => this.setState({ newModule: { name: e.target.value, err: '' } }),
    add: () => rolesW.addModule(this, srv, nm.name).then(() => { this.setState({ newModule: { name: '', err: '' } }); this.notify('Page added. Add rights to it.', '#B8F36B'); }).catch((e) => this.setState({ newModule: { ...nm, err: e.message || 'Could not add it.' } })) };
  const setRoleSel = (e) => this.setState({ roleSel: e.target.value });
  const addToRole = (e) => { const n = e.target.value; if (n) assignRole(n, roleCur.id); };
  const levelOpts = [{ v: '0', l: 'No approval' }, { v: '1', l: 'Direct (1 level)' }, { v: '2', l: '2 levels' }, { v: '3', l: '3 levels' }];
  // Who approves a level is not typed in: it is whoever holds the right "Approve level N" on that flow's page. The names shown
  // come from the server (the roles that hold it).
  const flowsTable = { id: 'flows', pageSize: 10, minWidth: '1100px',
    cols: [{ label: 'Action', filter: 'text' }, { label: 'Page', filter: 'select' }, { label: 'Approval', filter: 'select' }, { label: 'Level 1 approved by' }, { label: 'Level 2 approved by' }, { label: 'Level 3 approved by' }],
    rows: s.flows.filter((f) => modEnabled(f.mod)).map((f) => {
      const fi = flowItem(f);
      // Each level: search for a role, then choose the person in that role who decides. Without a chosen person, anyone who holds
      // "Approve level N" can decide it. Levels beyond the number set above have nothing to choose.
      const roleOpts = (s.serverRoles || []).map((r) => ({ v: r.id, l: r.name }));
      const lv = (i) => {
        if (i >= f.levels) return C.t('—');
        const d = (f.designees || []).find((x) => x.level === i + 1);
        const pk = 'fr:' + f.id + ':' + (i + 1);
        const roleId = (s.flowRole || {})[pk] !== undefined ? s.flowRole[pk] : (d ? d.role_id : '');
        const role = roleOpts.find((r) => r.v === roleId);
        const people = this.dir.filter((p) => roleId && p.role === roleId && p.uid).map((p) => ({ v: p.uid, l: p.name }));
        const save = (person) => { if (this.srv) this.srv.request('PUT', `/approvals/flows/${f.id}/approvers/${i + 1}`, { person: person || null }).then(() => this.srv.reload.approvals()).catch((e) => this.notify(e.message || 'Could not save the approver.', '#FF9B8A')); };
        return { kind: 'picker', fv: d ? d.name + ' ' + d.role : '', id: pk, roleText: role ? role.l : '', roles: roleOpts, people, person: d && d.role_id === roleId ? d.person : '', hint: d ? '' : 'Anyone with the right "Approve level ' + (i + 1) + '"' + (f.approvers[i] ? ': ' + f.approvers[i] : ''),
          onRole: (e) => { const m = roleOpts.find((r) => r.l.toLowerCase() === e.target.value.trim().toLowerCase()); if (m) this.setState((st) => ({ flowRole: { ...(st.flowRole || {}), [pk]: m.v } })); else if (!e.target.value.trim()) { this.setState((st) => ({ flowRole: { ...(st.flowRole || {}), [pk]: '' } })); save(null); } },
          onPerson: (e) => save(e.target.value) };
      };
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
  const rolesTabs = [['roles', 'Roles'], ['groups', 'User groups'], ['rights', 'Rights'], ['members', 'Members'], ['reporting', 'Reporting lines'], ['flows', 'Approval rules']].map(([k, l]) => ({ label: l, go: () => this.setState({ rolesTab: k }),
    style: 'min-height:44px;padding:0 16px;border:0;background:transparent;cursor:pointer;font-size:14px;font-family:inherit;border-bottom:2px solid ' + (s.rolesTab === k ? '#B8F36B' : 'transparent') + ';color:' + (s.rolesTab === k ? 'var(--k-E8EFEA,#E8EFEA)' : 'var(--k-9DB0A6,#9DB0A6)') + ';font-weight:' + (s.rolesTab === k ? '600' : '500') }));
  const rolesView = { isRoles: s.rolesTab === 'roles', isRights: s.rolesTab === 'rights', isMembers: s.rolesTab === 'members', isReporting: s.rolesTab === 'reporting', isFlows: s.rolesTab === 'flows', isGroups: s.rolesTab === 'groups' };
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
  return { newRoleForm, newRightForm, newModuleForm, ntCats, ntCatSel, deskCats, deskTable, addToRole, alertsTable, approvalsTable, flowsTable, logsTable, blocksTable, blockHistory, blocksTable, membersTable, ntTargets, peopleSummary, peopleTable, userActions, rightsHint, rightsTable, roleBar, rolesTable, rolesTabs, rolesView, rulesTable, setRoleSel, ticketsTable };
}
