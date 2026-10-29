// Which home widgets the organization lets this person see. The organization decides (Widget settings): once it has set any, a widget
// it never set is not shown to the others, and one limited to specific groups shows only to their members. People who manage
// widget settings see every widget so they can set them up. Until the organization's settings have arrived nothing is shown.
export function widgetGate(logic, s, can) {
  const srv = !!logic.srv;
  const wc = s.wcfg || {};
  const loading = srv && !s.widgetsReady;
  const orgSets = srv && Object.keys(wc).length > 0;
  const manage = can('home', 'Manage widget settings');
  const inGroup = (key) => {
    const [kind, ref] = String(key).split(':');
    if (kind === 'forum') return (s.forumsLive || []).some((f) => f.id === ref);
    return (((logic.org || {}).mine) || {})[kind] === ref;
  };
  const allows = (id) => {
    if (loading) return false;
    if (!orgSets || manage) return true;
    const c = wc[id];
    if (!c) return false;
    if (c.see === 'Specific groups') return (c.groups || []).some(inGroup);
    return true;
  };
  return { loading, allows };
}
