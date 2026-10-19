import * as rolesW from '~/logic/server/widgets/roles.js';
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
  // The rights catalog and the roles are data from the server (Roles & rights): nothing about roles or rights is built in here.
  const catalog = (s.catalog && s.catalog.modules) || [];
  const rightsDef = Object.fromEntries(catalog.map((m) => [m.key, m.rights.map((r) => r.name)]));
  const rightMeta = Object.fromEntries(catalog.map((m) => [m.key, Object.fromEntries(m.rights.map((r) => [r.name, r]))]));
  const W = s.widgets;
  // A structure level exists when the organization has created at least one of it (People > Organization structure).
  // Without the server connected, the widgets still decide, as in the sample data.
  const lvl = (k) => (this.srv ? ((this.org && this.org[k]) || []).length > 0 : W.indexOf(k) !== -1);
  const has = {
    departments: lvl('departments'), sectors: lvl('sectors'), branches: lvl('branches'), units: lvl('units'),
    chat: W.indexOf('chat') !== -1, calendar: W.indexOf('calendar') !== -1, memos: W.indexOf('memos') !== -1, newsletter: W.indexOf('newsletter') !== -1, announcements: W.indexOf('announcements') !== -1
  };
  has.structure = has.departments || has.sectors || has.branches || has.units;
  has.comms = has.memos || has.newsletter || has.announcements;
  ['tickets', 'projects', 'files', 'secureshare', 'sessions', 'awards', 'notes'].forEach((k) => { has[k] = W.indexOf(k) !== -1; });
  const optionalMods = ['calendar', 'memos', 'newsletter', 'announcements', 'departments', 'sectors', 'branches', 'units', 'performance', 'tickets', 'projects', 'files', 'secureshare', 'sessions', 'awards', 'chat'];
  const modEnabled = (m) => optionalMods.indexOf(m) === -1 || (m === 'performance' ? has.structure : ['departments', 'sectors', 'branches', 'units'].indexOf(m) !== -1 ? has[m] : W.indexOf(m) !== -1);
  const modInfo = Object.fromEntries(catalog.map((m) => [m.key, [m.label, m.tint]]));
  const genericRights = ['View'];
  // the roles of the organization, as made on the Roles page (the count and the "needs" level come from the server)
  const roleDefs = (s.serverRoles || []).map((r) => ({ id: r.id, name: r.name, desc: r.desc || '', count: r.count, locked: !!r.locked, needs: r.needs || '', scope: r.scope || '', isDefault: !!r.default }));
  const roleScope = Object.fromEntries(roleDefs.map((r) => [r.id, r.scope || 'Whole organization']));
  const visibleRoles = roleDefs.filter((r) => (r.needs ? has[r.needs] : true));
  const rKey = (role, mod, right) => role + '|' + mod + '|' + right;
  const rightOf = (role, mod, right) => {
    const k = rKey(role, mod, right);
    if (Object.prototype.hasOwnProperty.call(s.rightsOv, k)) return s.rightsOv[k];
    const def = roleDefs.find((r) => r.id === role);
    if (def && def.locked) return true;      // a locked (owner) role can do everything
    const sr = s.roleRights && s.roleRights[role] && s.roleRights[role][mod];
    return !!(sr && sr[right]);               // a right nobody gave the role is off
  };
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
  this._makeRequest = makeRequest; this._flowById = flowById;
  const toggleRight = (role, mod, right) => {
    if ((roleDefs.find((r) => r.id === role) || {}).locked) return;
    const k = rKey(role, mod, right);
    const nv = !rightOf(role, mod, right);
    const f = flowById('f8');
    const roleName = (roleDefs.find((r) => r.id === role) || {}).name;
    const logIt = (txt) => this.setState((st) => ({ extraLogs: [{ id: 'x' + Date.now(), ago: 0, user: 'Samwel Okoth Omulo', action: txt, module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' }].concat(st.extraLogs) }));
    const label = roleName + ' · ' + (modInfo[mod] ? modInfo[mod][0] : mod) + ' · ' + right + (nv ? ' (allow)' : ' (remove)');
    if (this.srv) {
      if (s.pendingRights[k] !== undefined) return;
      rolesW.setRight(this, this.srv, role, mod, right, nv)
        .then((r) => this.notify(r === 'applied' ? 'Rights updated: ' + label + '. Logged for review.' : 'Sent for approval: Change rights: ' + label + '.', r === 'applied' ? '#B8F36B' : '#FFB86B'))
        .catch((e) => this.notify(e.message || 'Could not change the right.', '#FF9B8A'));
      return;
    }
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
  return { can, catalog, rightMeta, flowById, genericRights, has, makeRequest, modEnabled, modInfo, openPersonName, rKey, rightOf, rightsDef, roleDefs, roleScope, s, set, toggleRight, visibleRoles };
}
