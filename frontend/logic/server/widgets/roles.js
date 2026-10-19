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
  (logic.dir || []).forEach((p) => { if (p.role) memberRoles[p.name] = p.role; });
  if (srv.me.role) memberRoles[srv.me.name] = srv.me.role;
  srv.rightsCatalog = cat;
  // reporting lines between roles (and, where chosen, the person a role reports to)
  const reportsTo = {}, reportsPerson = {};
  roles.forEach((r) => {
    if (r.reports_to_role) reportsTo[r.id] = r.reports_to_role;
    if (r.reports_to_person) reportsPerson[r.id] = r.reports_to_person === srv.me.person ? srv.me.name : srv.nameOf.get(r.reports_to_person) || '';
  });
  srv.reportSnap = JSON.stringify({ reportsTo, reportsPerson });
  logic.setState((st) => ({ reportsTo, reportsPerson, roleRights, roleCounts, memberRoles: { ...st.memberRoles, ...memberRoles }, actingRole: srv.me.role || st.actingRole, pendingRights: {} }));
}

const key = (role, mod, right) => role + '|' + mod + '|' + right;

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
  if (applied) { logic.setState((st) => ({ memberRoles: { ...st.memberRoles, [personName]: role } })); await load(logic, srv); } else if (srv.reload.approvals) srv.reload.approvals();
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
