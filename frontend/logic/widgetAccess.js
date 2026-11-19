// Which home widgets this person sees. A widget exists for an organization only when the organization switched it on (Widget
// settings), and then only for the roles that hold its "see" right on Roles & rights; the server decides that (/me/widgets). Only
// the switch counts: a widget whose title was saved but never switched on is off. Until the settings have arrived nothing shows.
export const switchedOn = (wc, id) => { const c = (wc || {})[id]; return !!c && c.enabled === true; };

// How many entries a widget lists on Home unless its settings say otherwise (widgets not named here list no entries).
export const HOME_ITEMS = { announcements: 3, calendar: 4, chat: 4, files: 4, jobposts: 3, memos: 3, newsletter: 3, tickets: 3 };
export const homeItems = (wc, id) => { const n = Number(((wc || {})[id] || {}).items) || 0; return n > 0 ? Math.min(n, 20) : HOME_ITEMS[id] || 0; };

export function widgetGate(logic, s, can) {
  const srv = !!logic.srv;
  const wc = s.wcfg || {};
  const loading = srv && !s.widgetsReady;
  const manage = can('home', 'Manage widget settings');
  const myRole = srv && logic.srv.me ? logic.srv.me.role : '';
  const allows = (id) => {
    if (!srv) return true;   // not signed in to an organization: the built-in samples
    if (loading || !switchedOn(wc, id)) return false;
    return Array.isArray(s.myWidgets) ? s.myWidgets.indexOf(id) !== -1 : true;
  };
  // who may open a widget's settings: whoever manages widget settings, or a role named as that widget's manager
  const manages = (id) => manage || (!!myRole && ((wc[id] || {}).managers || []).indexOf(myRole) !== -1);
  return { loading, allows, manage, manages };
}
