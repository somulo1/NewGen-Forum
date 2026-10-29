// Which home widgets this person sees. The rule is simple: a widget exists for an organization only when the organization has switched it
// on (Widget settings -> "Switched on for the organization"), and then only for the people it lets see it ("Everyone" or specific
// groups). Everyone in the organization sees the same switched-on set; nobody gets one the organization never switched on, whoever they
// are. An older setting that was saved before the switch existed counts as switched on. Until the organization's settings have arrived
// nothing is shown.
export const switchedOn = (wc, id) => { const c = (wc || {})[id]; return !!c && c.enabled !== false; };

export function widgetGate(logic, s, can) {
  const srv = !!logic.srv;
  const wc = s.wcfg || {};
  const loading = srv && !s.widgetsReady;
  const inGroup = (key) => {
    const [kind, ref] = String(key).split(':');
    if (kind === 'forum') return (s.forumsLive || []).some((f) => f.id === ref);
    return (((logic.org || {}).mine) || {})[kind] === ref;
  };
  const allows = (id) => {
    if (!srv) return true;   // not signed in to an organization: the built-in samples
    if (loading) return false;
    if (!switchedOn(wc, id)) return false;
    const c = wc[id];
    if (c.see === 'Specific groups') return (c.groups || []).some(inGroup);
    return true;
  };
  return { loading, allows, manage: can('home', 'Manage widget settings') };
}
