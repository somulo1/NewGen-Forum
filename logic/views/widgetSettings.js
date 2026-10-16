// renderVals step: widgetSettings. Receives the names built by earlier steps and returns the ones later steps need.
export function buildWidgetSettings(ctx) {
  const { O, catalog, chip, chipS, flowItem, forums, genericRights, has, rKey, rightOf, rightsDef, s, toggleRight, typeName, visibleRoles } = ctx;
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
      matrix: visibleRoles.map((r) => ({ role: r.name, cells: rn.map((x) => ({ label: r.name + ': ' + x, on: rightOf(r.id, wmod, x), locked: !!r.locked, pending: s.pendingRights[rKey(r.id, wmod, x)] !== undefined, toggle: () => toggleRight(r.id, wmod, x) })) })),
      flows: wflows, noFlows: wflows.length === 0,
      title: cfg.title, see: cfg.see, manage: cfg.manage, setTitle: setCfg('title'), setSee: setCfg('see'), setManage: setCfg('manage') };
    // Visibility (groups) and management (roles)
    const setW = (patch) => this.setState((st) => ({ wcfg: { ...st.wcfg, [sid]: Object.assign({ title: catalog[sid].title, see: 'Everyone', groups: [], managers: ['superadmin', 'admin'] }, st.wcfg[sid] || {}, typeof patch === 'function' ? patch(Object.assign({ groups: [], managers: ['superadmin', 'admin'] }, st.wcfg[sid] || {})) : patch) } }));
    const groups = cfg.groups || [];
    const managers = cfg.managers || ['superadmin', 'admin'];
    const isGroups = cfg.see === 'Specific groups';
    const groupSets = ['departments', 'sectors', 'branches', 'units'].filter((t) => has[t]).map((t) => ({ label: typeName[t] + (t === 'branches' ? 'es' : 's'), items: O[t].map((r) => ({ key: t + ':' + r.id, name: r.name })) }))
      .concat([{ label: 'Forums', items: forums.map((f) => ({ key: 'forum:' + f.id, name: f.name })) }])
      .map((gs) => ({ label: gs.label, items: gs.items.map((it) => { const on = groups.indexOf(it.key) !== -1; return { name: it.name, on: on, style: chipS(on), toggle: () => setW((c) => ({ groups: on ? c.groups.filter((x) => x !== it.key) : c.groups.concat([it.key]) })) }; }) }));
    Object.assign(sw, {
      seeOpts: ['Everyone', 'Specific groups'].map((v) => ({ label: v, on: cfg.see === v, style: chip(cfg.see === v), pick: () => setW({ see: v }) })),
      isGroups: isGroups, groupSets: groupSets,
      groupsHint: groups.length ? groups.length + (groups.length === 1 ? ' group selected. Only its members see this widget.' : ' groups selected. Only their members see this widget.') : 'Pick at least one group. You can\'t save until you do.',
      groupsHintColor: groups.length ? '#C9D6CF' : '#FFB86B', groupsBorder: groups.length ? '#24322C' : '#6A5426',
      managerRoles: visibleRoles.map((r) => { const on = r.id === 'superadmin' || managers.indexOf(r.id) !== -1; return { name: r.name, on: on, locked: r.id === 'superadmin', style: chipS(on), toggle: () => setW((c) => ({ managers: on ? c.managers.filter((x) => x !== r.id) : c.managers.concat([r.id]) })) }; })
    });
    // Communication rules for structure widgets
    if (['departments', 'sectors', 'branches', 'units'].indexOf(sid) !== -1) {
      const lvl = sid;
      const roleOpts = visibleRoles.filter((r) => r.id !== 'guest').map((r) => ({ v: r.id, l: r.name }));
      const setRule = (ty, patch) => this.setState((st) => ({ commRules: { ...st.commRules, [lvl]: { ...st.commRules[lvl], [ty]: Object.assign({}, st.commRules[lvl][ty], typeof patch === 'function' ? patch(st.commRules[lvl][ty]) : patch) } } }));
      const targetKeys = ['own', 'wall'].concat(['departments', 'sectors', 'branches', 'units'].filter((t) => has[t]));
      Object.assign(sw, {
        isLevel: true, levelSingular: typeName[lvl].toLowerCase(),
        commRules: ['Memo', 'Announcement'].filter((ty) => (ty === 'Memo' ? has.memos : has.announcements) || true).map((ty) => {
          const r = s.commRules[lvl][ty];
          return { label: typeName[lvl] + 's can send ' + ty.toLowerCase() + 's', on: r.on, toggle: () => setRule(ty, (x) => ({ on: !x.on })),
            targets: targetKeys.map((k) => { const on = r.to.indexOf(k) !== -1; return { label: k === 'own' ? 'Its own ' + typeName[lvl].toLowerCase() : k === 'wall' ? 'Community wall' : typeName[k] + (k === 'branches' ? 'es' : 's'), on: on, style: chipS(on),
              toggle: () => setRule(ty, (x) => ({ to: on ? x.to.filter((y) => y !== k) : x.to.concat([k]) })) }; }),
            sender: r.sender, approver: r.approver, roleOpts: roleOpts,
            setSender: (e) => { const v = e.target.value; setRule(ty, { sender: v }); },
            setApprover: (e) => { const v = e.target.value; setRule(ty, { approver: v }); } };
        })
      });
    }
  }
  return { sid, sw };
}
