import { widgetGate, switchedOn } from '~/logic/widgetAccess.js';
import * as rolesW from '~/logic/server/widgets/roles.js';
import { sayPermissions } from '~/logic/wording.js';
// renderVals step: access. Receives the names built by earlier steps and returns the ones later steps need.
export function buildAccess(ctx) {
  const s = this.state;
  const set = (o) => () => this.setState(o);
  const openPersonName = (n) => {
    if (n === 'Sam' || n === (this.srv ? this.srv.me.name : 'Samwel Okoth Omulo')) { this.setState({ section: 'profile', personOpen: null }); return; }
    const d = this.dir.find((x) => x.name === n);
    if (d) this.setState({ personOpen: d.id });
  };

  // ---- Roles & rights ----
  // The rights catalog and the roles are data from the server (Roles & rights): nothing about roles or rights is built in here.
  const catalog = (s.catalog && s.catalog.modules) || [];
  const rightsDef = Object.fromEntries(catalog.map((m) => [m.key, m.rights.map((r) => r.name)]));
  const rightMeta = Object.fromEntries(catalog.map((m) => [m.key, Object.fromEntries(m.rights.map((r) => [r.name, { ...r, label: sayPermissions(r.label || r.name) }]))]));
  // the right that decides who has a widget (its "see" right), as the server names it
  const seeOf = Object.fromEntries(catalog.map((m) => [m.key, m.see || '']));
  // Which features (and so which sidebar items and pages) exist for this person: the ones the organization switched on for them. It is
  // the organization's choice, not each person's own list: the person's own list only decides what is drawn on their Home. Without the
  // server (the built-in samples) the person's own list decides, as before.
  const FEATURES = ['pbx', 'tickets', 'projects', 'files', 'secureshare', 'sessions', 'awards', 'notes', 'chat', 'calendar', 'memos', 'newsletter', 'announcements', 'jobposts', 'performance', 'departments', 'sectors', 'branches', 'units'];
  const gate = widgetGate(this, s, () => false);
  const W = this.srv ? FEATURES.filter((id) => gate.allows(id)) : s.widgets;
  // A structure level exists when the organization has created at least one of it (People > Organization structure).
  // Without the server connected, the widgets still decide, as in the sample data.
  const lvl = (k) => (this.srv ? ((this.org && this.org[k]) || []).length > 0 : W.indexOf(k) !== -1);
  const has = {
    departments: lvl('departments'), sectors: lvl('sectors'), branches: lvl('branches'), units: lvl('units'),
    chat: W.indexOf('chat') !== -1, calendar: W.indexOf('calendar') !== -1, memos: W.indexOf('memos') !== -1, newsletter: W.indexOf('newsletter') !== -1, announcements: W.indexOf('announcements') !== -1
  };
  const orgBlocks = s.orgBlocks || {};
  has.privateCalls = has.chat;   // private calls can stay on when private chats are blocked
  if (orgBlocks.blockPrivateChats) has.chat = false;
  if (orgBlocks.blockWall) has.wall = false;
  has.jobposts = W.indexOf('jobposts') !== -1 || has.memos || has.newsletter || has.announcements;   // job posts come with communications
  has.structure = has.departments || has.sectors || has.branches || has.units;
  has.comms = has.memos || has.newsletter || has.announcements || has.jobposts;
  ['tickets', 'projects', 'files', 'secureshare', 'sessions', 'awards', 'notes', 'pbx'].forEach((k) => { has[k] = W.indexOf(k) !== -1; });
  // the phone app leaves out Calls management (the customer phone system); everything else is the same as on the web
  if (typeof window !== 'undefined' && window.HubNative && window.HubNative.app) has.pbx = false;
  const optionalMods = ['pbx', 'calendar', 'memos', 'newsletter', 'announcements', 'jobposts', 'departments', 'sectors', 'branches', 'units', 'performance', 'tickets', 'projects', 'files', 'secureshare', 'sessions', 'awards', 'chat'];
  // Rights can only be given for what the organization has switched on. A page that is not a widget (wall, forums, people, roles,
  // security, settings, the home page and its widget settings) always has its rights, so the first administrator can start giving
  // them. A widget counts when the ORGANIZATION switched it on, whoever can see it; structure levels count once the organization
  // has one. Without the server, the sample's own widget list decides.
  const orgOn = (id) => (this.srv ? switchedOn(s.wcfg, id) : s.widgets.indexOf(id) !== -1);
  const modEnabled = (m) => optionalMods.indexOf(m) === -1 || (m === 'jobposts' ? (orgOn('jobposts') || orgOn('memos') || orgOn('newsletter') || orgOn('announcements')) : m === 'performance' ? has.structure : ['departments', 'sectors', 'branches', 'units'].indexOf(m) !== -1 ? has[m] : orgOn(m));
  const modInfo = Object.fromEntries(catalog.map((m) => [m.key, [sayPermissions(m.label), m.tint]]));
  // the roles of the organization, as made on the Roles page (the count and the "needs" level come from the server)
  const roleDefs = (s.serverRoles || []).map((r) => ({ id: r.id, name: r.name, desc: r.desc || '', count: r.count, locked: !!r.locked, needs: r.needs || '', scope: r.scope || '', isDefault: !!r.default }));
  const roleScope = Object.fromEntries(roleDefs.map((r) => [r.id, r.scope || 'Whole organization']));
  const visibleRoles = roleDefs.filter((r) => (r.needs ? has[r.needs] : true));
  const rKey = (role, mod, right) => role + '|' + mod + '|' + right;
  const rightOf = (role, mod, right) => {
    const k = rKey(role, mod, right);
    if (Object.prototype.hasOwnProperty.call(s.rightsOv, k)) return s.rightsOv[k];
    const sr = s.roleRights && s.roleRights[role] && s.roleRights[role][mod];
    return !!(sr && sr[right]);               // the server's matrix: a right nobody gave the role is off (the owner's are on until switched off)
  };
  // What I may do comes from the server, which counts my role and my personal rights (the rights tables above are only readable by
  // people who manage roles). Previewing another role uses those tables.
  const previewing = !!(s.actingRole && s.myRole && s.actingRole !== s.myRole);
  const mine = this.srv && !previewing && this.srv.me && this.srv.me.rights ? this.srv.me.rights : null;
  const can = mine ? (mod, right) => (mine[mod] || []).indexOf(right) !== -1
    : (mod, right) => rightOf(s.actingRole, mod, right);
  // May I open the dashboards of this level (departments, sectors, branches, units)? The server's one rule (switched on, and the
  // role's "View dashboard"), so every page and the sidebar give the same answer.
  const canDash = (t) => (mine && Array.isArray(s.myWidgets) ? s.myWidgets.indexOf(t) !== -1 : can(t, 'View dashboard'));
  const refName = (ref) => (roleDefs.find((r) => r.id === ref) || {}).name || ref;
  const holdsRef = (ref) => s.actingRole === ref;
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
    const k = rKey(role, mod, right);
    const nv = !rightOf(role, mod, right);
    if (!nv && (roleDefs.find((r) => r.id === role) || {}).locked && ((rightMeta[mod] || {})[right] || {}).kept) return;   // the owner role always keeps it
    const f = flowById('f8');
    const roleName = (roleDefs.find((r) => r.id === role) || {}).name;
    const logIt = (txt) => this.setState((st) => ({ extraLogs: [{ id: 'x' + Date.now(), ago: 0, user: 'Samwel Okoth Omulo', action: txt, module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' }].concat(st.extraLogs) }));
    const label = roleName + ' · ' + (modInfo[mod] ? modInfo[mod][0] : mod) + ' · ' + right + (nv ? ' (allow)' : ' (remove)');
    if (this.srv) {
      if (s.pendingRights[k] !== undefined) return;
      // Rights changes that need approval are marked first and sent together as ONE approval (Roles > Rights: "Send for approval").
      if (f && f.levels > 0) {
        this.setState((st) => {
          const marks = { ...(st.rightsMarks || {}) };
          if (marks[k]) delete marks[k]; else marks[k] = { role, mod, right, allowed: nv };
          return { rightsMarks: marks };
        });
        return;
      }
      rolesW.setRight(this, this.srv, role, mod, right, nv)
        .then((r) => this.notify(r === 'applied' ? 'Permissions updated: ' + label + '. Logged for review.' : 'Sent for approval: Change permissions: ' + label + '.', r === 'applied' ? '#B8F36B' : '#FFB86B'))
        .catch((e) => this.notify(e.message || 'Could not change the permission.', '#FF9B8A'));
      return;
    }
    if (f && f.levels > 0) {
      if (s.pendingRights[k] !== undefined) return;
      this.setState((st) => ({ pendingRights: { ...st.pendingRights, [k]: nv } }));
      makeRequest('f8', 'Change permissions: ' + label, { kind: 'rights', key: k, val: nv });
      logIt('Requested permissions change: ' + label);
    } else {
      this.setState((st) => ({ rightsOv: { ...st.rightsOv, [k]: nv } }));
      logIt('Changed permissions: ' + label);
      this.notify('Permissions updated: ' + label + '. Logged for review.', '#FFB86B');
    }
  };
  return { can, canDash, refName, holdsRef, catalog, rightMeta, seeOf, flowById, has, makeRequest, modEnabled, modInfo, openPersonName, rKey, rightOf, rightsDef, roleDefs, roleScope, s, set, toggleRight, visibleRoles };
}
