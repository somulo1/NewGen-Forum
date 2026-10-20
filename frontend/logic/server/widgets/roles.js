// Roles & rights: what each role may do (the rights matrix), who holds which role, and the role you act as. Rights changes
// and role changes follow their approval flows (f8, f11) on the server; until the last level approves they show as pending.
export const name = 'roles';

export async function load(logic, srv) {
  const [{ roles }, cat] = await Promise.all([srv.request('GET', '/roles'), srv.request('GET', '/access/catalog')]);
  const details = await Promise.all(roles.map((r) => srv.request('GET', '/roles/' + r.id).catch(() => null)));
  const roleRights = {};
  details.filter(Boolean).forEach((d) => { roleRights[d.id] = d.rights || {}; });
  const roleCounts = Object.fromEntries(roles.map((r) => [r.id, r.count]));
  // who holds which role, from the directory
  const memberRoles = {};
  (logic.dir || []).forEach((p) => { memberRoles[p.name] = p.role || ''; });
  memberRoles[srv.me.name] = srv.me.role || '';
  srv.rightsCatalog = cat;
  const preview = logic.state.actingRole && roles.some((r) => r.id === logic.state.actingRole) && logic.state.myRole && logic.state.actingRole !== logic.state.myRole;
  // reporting lines between roles (and, where chosen, the person a role reports to)
  const reportsTo = {}, reportsPerson = {};
  roles.forEach((r) => {
    if (r.reports_to_role) reportsTo[r.id] = r.reports_to_role;
    if (r.reports_to_person) reportsPerson[r.id] = r.reports_to_person === srv.me.person ? srv.me.name : srv.nameOf.get(r.reports_to_person) || '';
  });
  srv.reportSnap = JSON.stringify({ reportsTo, reportsPerson });
  logic.setState((st) => ({ reportsTo, reportsPerson, roleRights, roleCounts, serverRoles: roles, catalog: cat, memberRoles: { ...st.memberRoles, ...memberRoles }, myRole: srv.me.role || '', actingRole: preview ? st.actingRole : (srv.me.role || ''), roleSel: roles.some((r) => r.id === st.roleSel) ? st.roleSel : '', pendingRights: {} }));
  await loadGroups(logic, srv);
}

const key = (role, mod, right) => role + '|' + mod + '|' + right;

// ---- user groups: sets of roles that are given rights (and widget access) together ----

// Reads the groups and what each is allowed. A person without the right to see roles simply has none.
export async function loadGroups(logic, srv) {
  let groups = [];
  try { groups = (await srv.request('GET', '/groups')).groups; } catch (e) { groups = []; }
  const details = await Promise.all(groups.map((g) => srv.request('GET', '/groups/' + g.id).catch(() => null)));
  const groupRights = {};
  details.filter(Boolean).forEach((d) => { groupRights[d.id] = d.rights || {}; });
  logic.setState({ serverGroups: groups, groupRights });
}

// Allows or removes one right for a group. Returns 'applied' or 'pending' (a rights change may wait for approval).
export async function setGroupRight(logic, srv, group, mod, right, allowed) {
  const res = await srv.request('PUT', `/groups/${group}/rights`, { module: mod, right, allowed });
  if (res.applied) {
    logic.setState((st) => ({ groupRights: { ...st.groupRights, [group]: { ...(st.groupRights[group] || {}), [mod]: { ...((st.groupRights[group] || {})[mod] || {}), [right]: allowed } } } }));
    return 'applied';
  }
  logic.setState((st) => ({ pendingRights: { ...st.pendingRights, ['g:' + key(group, mod, right)]: allowed } }));
  if (srv.reload.approvals) srv.reload.approvals();
  return 'pending';
}
// g = { name, desc, roles: [role ids] }
export async function createGroup(logic, srv, g) {
  const res = await srv.request('POST', '/groups', { name: g.name, desc: g.desc || '', roles: g.roles || [] });
  await loadGroups(logic, srv);
  return res.id;
}
// patch = { name?, desc?, roles? }
export async function updateGroup(logic, srv, id, patch) {
  await srv.request('PATCH', '/groups/' + id, patch);
  await loadGroups(logic, srv);
}
export async function deleteGroup(logic, srv, id) {
  await srv.request('DELETE', '/groups/' + id);
  await loadGroups(logic, srv);
}

// Allows or removes one right for a role. Returns 'applied' or 'pending'.
export async function setRight(logic, srv, role, mod, right, allowed) {
  const res = await srv.request('PUT', `/roles/${role}/rights`, { module: mod, right, allowed });
  if (res.applied) {
    logic.setState((st) => ({ roleRights: { ...st.roleRights, [role]: { ...(st.roleRights[role] || {}), [mod]: { ...((st.roleRights[role] || {})[mod] || {}), [right]: allowed } } } }));
    return 'applied';
  }
  logic.setState((st) => ({ pendingRights: { ...st.pendingRights, [key(role, mod, right)]: allowed } }));
  if (srv.reload.approvals) srv.reload.approvals();
  return 'pending';
}

// Gives a person a role (flow f11 may hold it for approval).
export async function setRole(logic, srv, personName, role) {
  const pid = personName === srv.me.name ? srv.me.person : srv.personOf.get(personName);
  if (!pid) throw new Error('Choose someone from the directory.');
  const res = await srv.request('PATCH', '/people/' + pid, { role });
  const applied = !res || res.applied !== false;
  if (applied) {
    const d = (logic.dir || []).find((x) => x.name === personName);
    if (d) d.role = role;
    if (personName === srv.me.name) srv.me.role = role;
    logic.setState((st) => ({ memberRoles: { ...st.memberRoles, [personName]: role } }));
    await load(logic, srv);
  } else if (srv.reload.approvals) srv.reload.approvals();
  return applied ? 'applied' : 'pending';
}

// Saves the reporting lines that differ from what the server has.
export async function saveReporting(logic, srv) {
  const before = JSON.parse(srv.reportSnap || '{"reportsTo":{},"reportsPerson":{}}');
  const { reportsTo, reportsPerson } = logic.state;
  const roles = new Set([...Object.keys(reportsTo || {}), ...Object.keys(before.reportsTo), ...Object.keys(reportsPerson || {}), ...Object.keys(before.reportsPerson)]);
  for (const r of roles) {
    const to = (reportsTo || {})[r] || null, person = (reportsPerson || {})[r] || '';
    if (to === (before.reportsTo[r] || null) && person === (before.reportsPerson[r] || '')) continue;
    const pid = person ? (person === srv.me.name ? srv.me.person : srv.personOf.get(person)) : null;
    await srv.request('PUT', '/roles/reporting', { role: r, reports_to_role: to, reports_to_person: pid || null });
  }
  await load(logic, srv);
}

// ---- making and changing roles and the rights catalog (all data on the server) ----

const reload = (logic, srv) => load(logic, srv);

// r = { name, desc, copy_from }. Returns the new role's id.
export async function createRole(logic, srv, r) {
  const res = await srv.request('POST', '/roles', { name: r.name, desc: r.desc || '', copy_from: r.copy_from || '' });
  await reload(logic, srv);
  return res.id;
}
// patch = { name?, desc?, scope?, default? }
export async function updateRole(logic, srv, id, patch) {
  await srv.request('PATCH', '/roles/' + id, patch);
  await reload(logic, srv);
}
export async function deleteRole(logic, srv, id) {
  await srv.request('DELETE', '/roles/' + id);
  await reload(logic, srv);
}
// r = { module, name, kind }
export async function addRight(logic, srv, r) {
  await srv.request('POST', '/access/rights', { module: r.module, name: r.name, kind: r.kind || '' });
  await reload(logic, srv);
}
export async function removeRight(logic, srv, module, name) {
  await srv.request('DELETE', `/access/rights?module=${encodeURIComponent(module)}&name=${encodeURIComponent(name)}`);
  await reload(logic, srv);
}
export async function addModule(logic, srv, label) {
  await srv.request('POST', '/access/modules', { module: label, label });
  await reload(logic, srv);
}
