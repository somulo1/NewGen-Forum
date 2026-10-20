// renderVals step: nav. Receives the names built by earlier steps and returns the ones later steps need.
export function buildNav(ctx) {
  const { can, has, s, set } = ctx;
  const navBase ='display:flex;align-items:center;gap:12px;width:100%;min-height:44px;padding:0 12px;border:0;border-radius:10px;cursor:pointer;text-align:left;font-size:14px;';
  const sections = ['home', 'profile', 'feed', 'meetings', 'projects', 'approvals', 'security', 'roles', 'calendar', 'comms', 'structure', 'orgunit', 'tickets', 'files', 'share', 'people', 'orgedit', 'widgets', 'ticketdesk', 'settings', 'notes'];
  const nav = {}, go = {}, show = {};
  const secNeeds = { meetings: 'sessions', projects: 'projects', calendar: 'calendar', comms: 'comms', structure: 'structure', orgunit: 'structure', tickets: 'tickets', files: 'files', notes: 'notes', share: 'secureshare' };
  const ticketDeskOk = !!has.tickets && !has.structure && (can('tickets', 'Solve') || can('tickets', 'Assign & hand over'));
  const secOk = (k) => (k === 'ticketdesk' ? ticketDeskOk : k === 'orgedit' ? !!this.srv && can('settings', 'Edit organization profile') : k === 'widgets' ? can('home', 'Open widgets page') : k === 'security' ? (can('security', 'View alerts') || can('security', 'View logs')) : k === 'roles' ? can('roles', 'View roles') : k === 'settings' ? can('settings', 'View settings') : !secNeeds[k] || !!has[secNeeds[k]]);
  const pending = !!this._accessPending;   // rights are still arriving: do not call a page "no access" yet
  const curSec = pending || secOk(s.section) ? s.section : 'home';
  const labels = { meetings: 'Meetings & standups', projects: 'Projects', calendar: 'Calendar', comms: 'Communications', structure: 'Organization structure', orgedit: 'Manage organization structure', orgunit: 'Organization dashboards', tickets: 'My tickets', ticketdesk: 'Ticket desk', files: 'Files', share: 'Secure share', security: 'Security & logs', roles: 'Roles & rights', widgets: 'Widgets', settings: 'Organization settings', notes: 'Notes' };
  // Why a page can't open: its feature is not switched on (404-style), or the role lacks the right (403).
  const denied = pending || secOk(s.section) ? null : { label: labels[s.section] || 'this page', reason: secNeeds[s.section] && !has[secNeeds[s.section]] || (s.section === 'ticketdesk' && (!has.tickets || has.structure)) ? 'off' : 'rights' };
  sections.forEach((k) => {
    const on = curSec === k;
    nav[k] = navBase + (on ? 'background:var(--k-22302A,#22302A);color:var(--k-E8EFEA,#E8EFEA);font-weight:600' : 'background:transparent;color:var(--k-C9D6CF,#C9D6CF)');
    go[k] = () => this.setState((x) => ({ callMode: x.callMode === 'full' ? 'docked' : x.callMode, mode: x.mode === 'page' ? 'docked' : x.mode, section: k, ticketOpen: null, projectOpen: null, folder: null, notifOpen: false, statusMenu: false }));
    show[k] = on;
  });
  const tabs = ['discussions', 'polls', 'sessions', 'calls', 'members', 'files', 'about'];
  const tab = {}, tabTo = {}, ftab = {};
  tabs.forEach((k) => {
    const on = s.tab === k;
    tab[k] = 'min-height:44px;padding:0 16px;border:0;background:transparent;cursor:pointer;font-size:14px;white-space:nowrap;border-bottom:2px solid ' + (on ? '#B8F36B' : 'transparent') + ';color:' + (on ? 'var(--k-E8EFEA,#E8EFEA)' : 'var(--k-9DB0A6,#9DB0A6)') + ';font-weight:' + (on ? '600' : '500');
    tabTo[k] = set({ tab: k });
    ftab[k] = on && (k !== 'sessions' || has.sessions);
  });
  return { denied, curSec, ftab, go, nav, show, tab, tabTo, ticketDeskOk };
}
