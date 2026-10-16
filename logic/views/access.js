// renderVals step: access. Receives the names built by earlier steps and returns the ones later steps need.
export function buildAccess(ctx) {
  const s = this.state;
  const set = (o) => () => this.setState(o);
  const openPersonName = (n) => {
    if (n === 'Sam' || n === 'Samwel Okoth Omulo') { this.setState({ section: 'profile', personOpen: null }); return; }
    const d = this.dir.find((x) => x.name === n);
    if (d) this.setState({ personOpen: d.id });
  };

  // ---- Roles & rights ----
  const rightsDef = {
    home: ['Customize home', 'Manage widget settings', 'Open widgets page', 'Add & remove widgets'],
    announcements: ['View', 'Post', 'Edit', 'Delete', 'See who read'],
    wall: ['View', 'Post', 'Comment', 'Share outside Hub', 'Moderate posts'],
    forums: ['View', 'Create forum', 'Invite members', 'Start calls', 'Moderate'],
    tickets: ['View', 'Raise', 'Assign & hand over', 'Solve', 'Close', 'Manage TAT settings', 'Set up ticket categories'],
    projects: ['View', 'Create project', 'Manage milestones', 'Edit tasks', 'Move tasks', 'Manage team'],
    files: ['View', 'Upload', 'Delete', 'Share outside Hub'],
    secureshare: ['Create link', 'View link history'],
    sessions: ['Join', 'Start meeting', 'Record'],
    awards: ['View', 'Nominate', 'Approve awards'],
    people: ['View directory', 'See contact details', 'View full profiles', 'Onboard people', 'Edit profiles'],
    security: ['View alerts', 'View logs', 'Resolve alerts', 'Manage rules', 'Export logs'],
    roles: ['View roles', 'Manage roles & rights', 'Configure approvals'],
    calendar: ['View', 'Add events', 'Schedule meetings'],
    memos: ['Read', 'Send memos'],
    newsletter: ['Read', 'Publish newsletters'],
    departments: ['View dashboard', 'Manage department', 'See performance'],
    sectors: ['View dashboard', 'Manage sector', 'See performance'],
    branches: ['View dashboard', 'Manage branch', 'See performance'],
    units: ['View dashboard', 'Manage unit', 'See performance'],
    performance: ['View', 'Set targets'],
    chat: ['Send private messages', 'Private calls']
  };
  const W = s.widgets;
  const has = {
    departments: W.indexOf('departments') !== -1, sectors: W.indexOf('sectors') !== -1, branches: W.indexOf('branches') !== -1, units: W.indexOf('units') !== -1,
    chat: W.indexOf('chat') !== -1, calendar: W.indexOf('calendar') !== -1, memos: W.indexOf('memos') !== -1, newsletter: W.indexOf('newsletter') !== -1, announcements: W.indexOf('announcements') !== -1
  };
  has.structure = has.departments || has.sectors || has.branches || has.units;
  has.comms = has.memos || has.newsletter || has.announcements;
  ['tickets', 'projects', 'files', 'secureshare', 'sessions', 'awards'].forEach((k) => { has[k] = W.indexOf(k) !== -1; });
  if (!has.sessions) { rightsDef.calendar = ['View', 'Add events']; rightsDef.forums = rightsDef.forums.filter((x) => x !== 'Start calls'); }
  const optionalMods = ['calendar', 'memos', 'newsletter', 'announcements', 'departments', 'sectors', 'branches', 'units', 'performance', 'tickets', 'projects', 'files', 'secureshare', 'sessions', 'awards', 'chat'];
  const modEnabled = (m) => optionalMods.indexOf(m) === -1 || (m === 'performance' ? has.structure : W.indexOf(m) !== -1);
  const modInfo = {
    home: ['Home & widgets', '#B8F36B'], announcements: ['Announcements', '#B8F36B'], wall: ['Community wall', '#F49AC1'], forums: ['Forums', '#7CC4FF'],
    tickets: ['Tickets', '#FF9B8A'], projects: ['Projects', '#7CC4FF'], files: ['Files', '#C9B6FF'], secureshare: ['Secure share', '#6EE7C8'],
    sessions: ['Meetings & standups', '#6EE7C8'], awards: ['Performance awards', '#FFD66B'], people: ['People', '#C9B6FF'],
    security: ['Security & logs', '#FF9B8A'], roles: ['Roles & rights', '#FFB86B'],
    calendar: ['Calendar', '#7CC4FF'], memos: ['Memos', '#FFD66B'], newsletter: ['Newsletter', '#F49AC1'],
    departments: ['Departments', '#B8F36B'], sectors: ['Sectors', '#FFB86B'], branches: ['Branches', '#7CC4FF'], units: ['Units', '#C9B6FF'], performance: ['Performance', '#6EE7C8'], chat: ['Private messages', '#C9B6FF']
  };
  const genericRights = ['View', 'Edit content', 'Manage settings'];
  const roleDefs = [
    { id: 'superadmin', name: 'Super admin', desc: 'Full control of Hub. This role cannot be limited.', count: 2, locked: true },
    { id: 'admin', name: 'Admin', desc: 'Runs Hub day to day and manages widgets.', count: 5 },
    { id: 'hod', name: 'Head of department', desc: 'Leads a department and approves its requests.', count: 11 },
    { id: 'it', name: 'IT support', desc: 'Handles tickets and watches security.', count: 9 },
    { id: 'finance', name: 'Finance approver', desc: 'Approves spending and finance documents.', count: 4 },
    { id: 'staff', name: 'Staff', desc: 'Everyone working in the organization.', count: 212 },
    { id: 'guest', name: 'Guest', desc: 'People from outside. View only.', count: 5 },
    { id: 'sector', name: 'Sector manager', desc: 'Runs a sector and every branch in it.', count: 3, needs: 'sectors' },
    { id: 'branch', name: 'Branch manager', desc: 'Runs a branch and its units.', count: 4, needs: 'branches' },
    { id: 'unit', name: 'Unit lead', desc: 'Leads a unit inside a branch.', count: 6, needs: 'units' }
  ];
  const roleScope = { superadmin: 'Whole organization', admin: 'Whole organization', hod: 'Each department', it: 'Whole organization', finance: 'Whole organization', staff: 'Their own work', guest: 'Invited spaces only', sector: 'Each sector', branch: 'Each branch', unit: 'Each unit' };
  const visibleRoles = roleDefs.filter((r) => (r.needs ? has[r.needs] : true) && (r.id !== 'hod' || has.departments || !has.structure));
  const staffAllow = { home: ['Customize home'], announcements: ['View'], wall: ['View', 'Post', 'Comment', 'Share outside Hub'], forums: ['View', 'Start calls'],
    tickets: ['View', 'Raise'], projects: ['View', 'Edit tasks', 'Move tasks'], files: ['View', 'Upload'], secureshare: ['Create link'], sessions: ['Join', 'Start meeting'],
    awards: ['View', 'Nominate'], people: ['View directory', 'See contact details'], security: [], roles: [],
    calendar: ['View', 'Add events', 'Schedule meetings'], memos: ['Read'], newsletter: ['Read'], departments: [], sectors: [], branches: [], units: [], performance: ['View'], chat: ['Send private messages', 'Private calls'] };
  const extraAllow = {
    hod: { announcements: ['Post', 'Edit', 'See who read'], wall: ['Moderate posts'], forums: ['Create forum', 'Invite members', 'Moderate'], tickets: ['Assign & hand over', 'Solve', 'Close', 'Set up ticket categories'], projects: ['Create project', 'Manage milestones', 'Manage team'], files: ['Share outside Hub'], sessions: ['Record'], awards: ['Approve awards'], roles: ['View roles'] },
    it: { tickets: ['Assign & hand over', 'Solve', 'Close', 'Manage TAT settings', 'Set up ticket categories'], security: ['View alerts', 'View logs', 'Resolve alerts'], files: ['Delete'], forums: ['Create forum', 'Invite members'], secureshare: ['View link history'] },
    finance: { files: ['Share outside Hub'], projects: ['Create project'] },
    sector: { sectors: ['View dashboard', 'Manage sector', 'See performance'], memos: ['Send memos'], performance: ['Set targets'], announcements: ['Post'], tickets: ['Assign & hand over', 'Set up ticket categories'] },
    branch: { branches: ['View dashboard', 'Manage branch', 'See performance'], memos: ['Send memos'], performance: ['Set targets'], tickets: ['Assign & hand over', 'Solve', 'Set up ticket categories'] },
    unit: { units: ['View dashboard', 'Manage unit', 'See performance'], memos: ['Send memos'], tickets: ['Solve', 'Set up ticket categories'] }
  };
  extraAllow.hod.people = ['View full profiles', 'Onboard people']; extraAllow.hod.home = ['Open widgets page'];
  extraAllow.hod.memos = ['Send memos']; extraAllow.hod.departments = ['View dashboard', 'Manage department', 'See performance']; extraAllow.hod.performance = ['Set targets'];
  const defaultRight = (role, mod, right) => {
    if (role === 'superadmin' || role === 'admin') return true;
    if (!rightsDef[mod]) return (role === 'hod' || role === 'sector' || role === 'branch') ? right !== 'Manage settings' : right === 'View';
    if (role === 'guest') return ['announcements', 'wall', 'forums', 'sessions'].indexOf(mod) !== -1 && (right === 'View' || right === 'Join');
    const base = (staffAllow[mod] || []).indexOf(right) !== -1;
    const extra = ((extraAllow[role] || {})[mod] || []).indexOf(right) !== -1;
    return base || extra;
  };
  const rKey = (role, mod, right) => role + '|' + mod + '|' + right;
  const rightOf = (role, mod, right) => { const k = rKey(role, mod, right); return Object.prototype.hasOwnProperty.call(s.rightsOv, k) ? s.rightsOv[k] : defaultRight(role, mod, right); };
  const can = (mod, right) => rightOf(s.actingRole, mod, right);
  this._canFullProfile = can('people', 'View full profiles');
  const flowById = (id) => s.flows.find((f) => f.id === id);
  const nextReqId = () => 'AP-' + (34 + s.requests.length);
  const makeRequest = (flowId, title, extra) => {
    const f = flowById(flowId);
    const req = Object.assign({ id: nextReqId(), flowId: flowId, title: title, by: 'Samwel Okoth Omulo', when: 'Just now', level: 1, status: 'Pending', history: [] }, extra || {});
    this.setState((st) => ({ requests: [req].concat(st.requests) }));
    this.notify('Sent for approval: ' + title + ' (' + (f.levels === 1 ? 'direct approval' : f.levels + ' levels') + ').', '#FFB86B');
  };
  const toggleRight = (role, mod, right) => {
    if (role === 'superadmin') return;
    const k = rKey(role, mod, right);
    const nv = !rightOf(role, mod, right);
    const f = flowById('f8');
    const roleName = (roleDefs.find((r) => r.id === role) || {}).name;
    const logIt = (txt) => this.setState((st) => ({ extraLogs: [{ id: 'x' + Date.now(), ago: 0, user: 'Samwel Okoth Omulo', action: txt, module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' }].concat(st.extraLogs) }));
    const label = roleName + ' · ' + (modInfo[mod] ? modInfo[mod][0] : mod) + ' · ' + right + (nv ? ' (allow)' : ' (remove)');
    if (f && f.levels > 0) {
      if (s.pendingRights[k] !== undefined) return;
      this.setState((st) => ({ pendingRights: { ...st.pendingRights, [k]: nv } }));
      makeRequest('f8', 'Change rights: ' + label, { kind: 'rights', key: k, val: nv });
      logIt('Requested rights change: ' + label);
    } else {
      this.setState((st) => ({ rightsOv: { ...st.rightsOv, [k]: nv } }));
      logIt('Changed rights: ' + label);
      this.notify('Rights updated: ' + label + '. Logged for review.', '#FFB86B');
    }
  };
  return { can, flowById, genericRights, has, makeRequest, modEnabled, modInfo, openPersonName, rKey, rightOf, rightsDef, roleDefs, roleScope, s, set, toggleRight, visibleRoles };
}
