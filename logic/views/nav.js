// renderVals step: nav. Receives the names built by earlier steps and returns the ones later steps need.
export function buildNav(ctx) {
  const { has, s, set } = ctx;
  const navBase ='display:flex;align-items:center;gap:12px;width:100%;min-height:44px;padding:0 12px;border:0;border-radius:10px;cursor:pointer;text-align:left;font-size:14px;';
  const sections = ['home', 'profile', 'feed', 'meetings', 'projects', 'approvals', 'security', 'roles', 'calendar', 'comms', 'structure', 'orgunit', 'tickets', 'files', 'share', 'people', 'widgets'];
  const nav = {}, go = {}, show = {};
  const secNeeds = { meetings: 'sessions', projects: 'projects', calendar: 'calendar', comms: 'comms', structure: 'structure', orgunit: 'structure', tickets: 'tickets', files: 'files', share: 'secureshare' };
  const secOk = (k) => !secNeeds[k] || !!has[secNeeds[k]];
  const curSec = secOk(s.section) ? s.section : 'home';
  sections.forEach((k) => {
    const on = curSec === k;
    nav[k] = navBase + (on ? 'background:#22302A;color:#E8EFEA;font-weight:600' : 'background:transparent;color:#C9D6CF');
    go[k] = () => this.setState((x) => ({ callMode: x.callMode === 'full' ? 'docked' : x.callMode, section: k, ticketOpen: null, projectOpen: null, folder: null, notifOpen: false, statusMenu: false }));
    show[k] = on;
  });
  const tabs = ['discussions', 'polls', 'sessions', 'members', 'files', 'about'];
  const tab = {}, tabTo = {}, ftab = {};
  tabs.forEach((k) => {
    const on = s.tab === k;
    tab[k] = 'min-height:44px;padding:0 16px;border:0;background:transparent;cursor:pointer;font-size:14px;white-space:nowrap;border-bottom:2px solid ' + (on ? '#B8F36B' : 'transparent') + ';color:' + (on ? '#E8EFEA' : '#9DB0A6') + ';font-weight:' + (on ? '600' : '500');
    tabTo[k] = set({ tab: k });
    ftab[k] = on && (k !== 'sessions' || has.sessions);
  });
  return { curSec, ftab, go, nav, show, tab, tabTo };
}
