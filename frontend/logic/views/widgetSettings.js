// renderVals step: widgetSettings. Receives the names built by earlier steps and returns the ones later steps need.
export function buildWidgetSettings(ctx) {
  const { O, catalog, chip, chipS, flowItem, forums, genericRights, groupDefs, groupRightOf, has, rKey, rightOf, rightsDef, s, toggleGroupRight, toggleRight, typeName, viaGroups, visibleRoles } = ctx;
  // ---- Widget settings ----
  const sid = s.settingsFor;
  const special = ['tickets', 'announcements', 'awards', 'secureshare', 'sessions', 'clocks', 'branches', 'sectors'];
  let sw = { t: {} };
  if (sid && catalog[sid]) {
    const cfg = Object.assign({ title: catalog[sid].title, see: 'Everyone', manage: 'Admins' }, s.wcfg[sid] || {});
    const setCfg = (k) => (e) => { const v = e.target.value; this.setState((st) => ({ wcfg: { ...st.wcfg, [sid]: Object.assign({ title: catalog[sid].title, see: 'Everyone', manage: 'Admins' }, st.wcfg[sid] || {}, { [k]: v }) } })); };
    const wmod = sid;
    const rn = rightsDef[wmod] || genericRights;
    const wflows = s.flows.filter((f) => f.mod === wmod).map(flowItem);
    sw = { ...catalog[sid], baseTitle: catalog[sid].title, t: { [sid]: true }, genericOnly: special.indexOf(sid) === -1,
      rightNames: rn.map((x) => ({ name: x })),
      // a role's cell also says when a user group it is in gives the right (so a ticked group explains why a role can act)
      matrix: visibleRoles.map((r) => ({ role: r.name, cells: rn.map((x) => { const via = viaGroups(r.id, wmod, x).map((g) => g.name); return { label: r.name + ': ' + x, on: rightOf(r.id, wmod, x), locked: !!r.locked, pending: s.pendingRights[rKey(r.id, wmod, x)] !== undefined, via: via.length && !rightOf(r.id, wmod, x) && !r.locked ? 'via ' + via.join(', ') : '', toggle: () => toggleRight(r.id, wmod, x) }; }) })),
      // user groups: a right ticked for a group is held by every role in it
      groupMatrix: groupDefs.map((g) => ({ group: g.name, roles: g.roles.map((id) => (visibleRoles.find((r) => r.id === id) || {}).name).filter(Boolean).join(', ') || 'No roles yet', cells: rn.map((x) => ({ label: g.name + ' (user group): ' + x, on: groupRightOf(g.id, wmod, x), locked: false, pending: s.pendingRights['g:' + rKey(g.id, wmod, x)] !== undefined, toggle: () => toggleGroupRight(g.id, wmod, x) })) })),
      hasGroups: groupDefs.length > 0,
      flows: wflows, noFlows: wflows.length === 0,
      title: cfg.title, see: cfg.see, manage: cfg.manage, setTitle: setCfg('title'), setSee: setCfg('see'), setManage: setCfg('manage') };
    // Visibility (groups) and management (roles)
    const setW = (patch) => this.setState((st) => ({ wcfg: { ...st.wcfg, [sid]: Object.assign({ title: catalog[sid].title, see: 'Everyone', groups: [], managers: [] }, st.wcfg[sid] || {}, typeof patch === 'function' ? patch(Object.assign({ groups: [], managers: [] }, st.wcfg[sid] || {})) : patch) } }));
    const groups = cfg.groups || [];
    const managers = cfg.managers || [];
    const managerGroups = cfg.managerGroups || [];
    const isGroups = cfg.see === 'Specific groups';
    const groupSets = ['departments', 'sectors', 'branches', 'units'].filter((t) => has[t]).map((t) => ({ label: typeName[t] + (t === 'branches' ? 'es' : 's'), items: O[t].map((r) => ({ key: t + ':' + r.id, name: r.name })) }))
      .concat([{ label: 'Forums', items: forums.map((f) => ({ key: 'forum:' + f.id, name: f.name })) }])
      .map((gs) => ({ label: gs.label, items: gs.items.map((it) => { const on = groups.indexOf(it.key) !== -1; return { name: it.name, on: on, style: chipS(on), toggle: () => setW((c) => ({ groups: on ? c.groups.filter((x) => x !== it.key) : c.groups.concat([it.key]) })) }; }) }));
    Object.assign(sw, {
      seeOpts: ['Everyone', 'Specific groups'].map((v) => ({ label: v, on: cfg.see === v, style: chip(cfg.see === v), pick: () => setW({ see: v }) })),
      isGroups: isGroups, groupSets: groupSets,
      groupsHint: groups.length ? groups.length + (groups.length === 1 ? ' group selected. Only its members see this widget.' : ' groups selected. Only their members see this widget.') : 'Pick at least one group. You can\'t save until you do.',
      groupsHintColor: groups.length ? '#C9D6CF' : '#FFB86B', groupsBorder: groups.length ? '#24322C' : '#6A5426',
      managerRoles: visibleRoles.map((r) => { const on = r.locked || managers.indexOf(r.id) !== -1; return { name: r.name, on: on, locked: r.locked, style: chipS(on), toggle: () => setW((c) => ({ managers: on ? c.managers.filter((x) => x !== r.id) : c.managers.concat([r.id]) })) }; }),
      // a user group can manage the widget too: every role in it does
      managerGroups: groupDefs.map((g) => { const on = managerGroups.indexOf(g.id) !== -1; return { name: g.name, on: on, locked: false, style: chipS(on), toggle: () => setW((c) => ({ managerGroups: on ? (c.managerGroups || []).filter((x) => x !== g.id) : (c.managerGroups || []).concat([g.id]) })) }; })
    });
  }
  return { sid, sw };
}
