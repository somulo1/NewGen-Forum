import { HOME_ITEMS, homeItems, widgetGate } from '~/logic/widgetAccess.js';
// renderVals step: widgetSettings. A widget's settings are its switch, its name, the roles that manage it and how many entries it
// lists on Home. Who sees it and what each role may do in it are rights, set on Roles & rights > Rights.
export function buildWidgetSettings(ctx) {
  const { can, catalog, chipS, modInfo, rightOf, s, visibleRoles } = ctx;
  const sid = s.settingsFor;
  if (!sid || !catalog[sid]) return { sid, sw: {} };
  const gate = widgetGate(this, s, can);
  const base = catalog[sid].title;
  const cfg = Object.assign({ title: base, managers: [] }, s.wcfg[sid] || {});
  const setW = (patch) => this.setState((st) => ({ wcfg: { ...st.wcfg, [sid]: Object.assign({ title: base, enabled: false }, st.wcfg[sid] || {}, typeof patch === 'function' ? patch(Object.assign({ managers: [] }, st.wcfg[sid] || {})) : patch) } }));
  const isOn = cfg.enabled === true;
  const managers = cfg.managers || [];
  const roleName = (k) => (visibleRoles.find((r) => r.id === k) || {}).name || k;
  const canRights = can('roles', 'View roles') && !!modInfo[sid];
  const sw = {
    tint: catalog[sid].tint, baseTitle: base, title: cfg.title,
    setTitle: (e) => { const v = e.target.value.slice(0, 60); setW({ title: v }); },
    orgOn: isOn, toggleOrgOn: () => setW({ enabled: !isOn }),
    orgOnLabel: isOn ? 'Switched on for the organization' : 'Switched off: nobody in the organization has this widget',
    // only someone who manages every widget changes who manages this one; a widget's own managers see the list
    editManagers: gate.manage,
    managerRoles: gate.manage
      ? visibleRoles.map((r) => {
        const every = rightOf(r.id, 'home', 'Manage widget settings'), on = every || managers.indexOf(r.id) !== -1;
        return { name: r.name, on, locked: every, title: every ? 'Manages every widget (Home & widgets permission)' : '', style: chipS(on), toggle: () => setW((c) => ({ managers: on ? c.managers.filter((x) => x !== r.id) : c.managers.concat([r.id]) })) };
      })
      : [],
    managerNames: managers.map(roleName).join(', ') || 'Only those who manage widget settings',
    items: HOME_ITEMS[sid] ? { value: String(homeItems(s.wcfg, sid)), options: [3, 4, 5, 6, 8, 10, 15, 20].map(String), set: (e) => setW({ items: Number(e.target.value) }) } : null,
    seeHint: 'Who sees this widget and what each role may do in it are set on Roles & permissions › Permissions' + (modInfo[sid] ? ', under ' + modInfo[sid][0] + '.' : '.'),
    canRights, openRights: () => this.setState({ settingsFor: null, section: 'roles', rolesTab: 'rights', rightsMod: sid })
  };
  return { sid, sw };
}
