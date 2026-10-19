// renderVals step: comms. Receives the names built by earlier steps and returns the ones later steps need.
export function buildComms(ctx) {
  const { C, O, chip, chipS, flowById, fmtDay, forums, has, makeRequest, recOf, roleNameOf, ruleSentence, s, set, todayISO, typeName } = ctx;
  // =============== Communications ===============
  const commTypes = [has.announcements && 'Announcement', has.memos && 'Memo', has.newsletter && 'Newsletter'].filter(Boolean);
  const commColor = { Announcement: '#B8F36B', Memo: '#FFD66B', Newsletter: '#F49AC1' };
  const openCm = (type) => this.setState({ cmOpen: true, cmd: { type: type, title: '', body: '', wall: type !== 'Memo', forum: '', sel: {} } });
  const cmS = s.cmd;
  const setCm = (patch) => this.setState((x) => ({ cmd: Object.assign({}, x.cmd, patch) }));
  const f1 = flowById('f1');
  // Communication rules: the user's own level decides where memos/announcements may go
  const ruleLevel = ['departments', 'units', 'branches', 'sectors'].find((t) => has[t]);
  const ruleObj = ruleLevel && (cmS.type === 'Memo' || cmS.type === 'Announcement') ? (s.commRules[ruleLevel] || {})[cmS.type] : null;
  const ruleRec = ruleLevel ? recOf(ruleLevel, O.mine[ruleLevel]) : null;
  const ruleTo = ruleObj ? (ruleObj.on ? ruleObj.to : []) : null;
  const senderOk = !ruleObj || ['admin', 'superadmin'].indexOf(s.actingRole) !== -1 || s.actingRole === ruleObj.sender;
  const ruleNeeds = !!(ruleObj && ruleObj.on && ruleObj.approver !== 'none');
  const wallAllowed = !ruleTo || ruleTo.indexOf('wall') !== -1;
  const cmNeedsApproval = ruleNeeds || (cmS.type === 'Announcement' && f1 && f1.levels > 0);
  const cm = {
    heading: 'New ' + cmS.type.toLowerCase(), title: cmS.title, body: cmS.body, wall: cmS.wall && wallAllowed, forum: cmS.forum, wallAllowed: wallAllowed,
    types: commTypes.map((t) => ({ label: t, pick: () => setCm({ type: t }), style: chip(cmS.type === t) })),
    setTitle: (e) => setCm({ title: e.target.value }), setBody: (e) => setCm({ body: e.target.value }),
    toggleWall: () => setCm({ wall: !this.state.cmd.wall }), setForum: (e) => setCm({ forum: e.target.value }),
    groups: ['departments', 'sectors', 'branches', 'units'].filter((t) => has[t]).map((t) => {
      let recs = O[t];
      if (ruleTo) recs = ruleTo.indexOf(t) !== -1 ? O[t] : (ruleTo.indexOf('own') !== -1 && t === ruleLevel ? [ruleRec] : []);
      return { label: typeName[t] + (t === 'branches' ? 'es' : 's'),
        items: recs.map((r) => { const k = t + ':' + r.name; const on = !!cmS.sel[k]; return { name: r.name, on: on, toggle: () => setCm({ sel: Object.assign({}, this.state.cmd.sel, { [k]: !on }) }), style: chipS(on) }; }) };
    }).filter((g) => g.items.length),
    hasRuleNote: !!ruleObj,
    ruleNote: ruleObj ? (ruleObj.on ? 'Rule for ' + ruleRec.name + ' (' + typeName[ruleLevel].toLowerCase() + '): ' + ruleSentence(ruleLevel, cmS.type) + (senderOk ? '' : ' Your role is not the sending role, so you cannot send this.') : cmS.type + 's from ' + typeName[ruleLevel].toLowerCase() + 's are switched off in the ' + typeName[ruleLevel] + (ruleLevel === 'branches' ? 'es' : 's') + ' widget settings.') : '',
    noStructure: !has.structure,
    needsApproval: cmNeedsApproval, approvalNote: ruleNeeds ? 'This needs approval by ' + roleNameOf(ruleObj.approver) + ' before it goes out.' : cmNeedsApproval ? 'Announcements need approval (' + (f1.levels === 1 ? 'direct' : f1.levels + ' levels') + ') before they go out.' : '',
    publishLabel: cmNeedsApproval ? 'Send for approval' : 'Publish',
    publishStyle: 'height:42px;padding:0 20px;border-radius:10px;border:0;font-weight:600;cursor:pointer;' + (cmS.title.trim() ? 'background:#B8F36B;color:#0F1714' : 'background:var(--k-2A3832,#2A3832);color:var(--k-9DB0A6,#9DB0A6)'),
    close: set({ cmOpen: false }),
    publish: () => {
      const c = this.state.cmd;
      if (!c.title.trim()) { this.notify('Add a title first.', '#FFB86B'); return; }
      if (ruleObj && !ruleObj.on) { this.notify(c.type + 's are switched off for your ' + typeName[ruleLevel].toLowerCase() + '.', '#FF9B8A'); return; }
      if (!senderOk) { this.notify('Only the ' + roleNameOf(ruleObj.sender) + ' role can send ' + c.type.toLowerCase() + 's here.', '#FF9B8A'); return; }
      const fo = forums.find((f) => f.id === c.forum);
      const to = (c.wall && wallAllowed ? ['Community wall'] : []).concat(fo ? [fo.name + ' forum'] : []).concat(Object.keys(c.sel).filter((k) => c.sel[k]).map((k) => k.split(':')[1]));
      if (!to.length) { this.notify('Choose at least one place to send it.', '#FFB86B'); return; }
      const id = 'C-' + (211 + this.state.comms.length);
      const needs = ruleNeeds || (c.type === 'Announcement' && f1 && f1.levels > 0);
      const item = { id: id, type: c.type, title: c.title.trim(), body: c.body.trim(), by: 'Samwel Okoth Omulo', date: todayISO, to: to, status: needs ? 'Awaiting approval' : 'Published', read: 0 };
      const patch = { cmOpen: false, comms: [item].concat(this.state.comms) };
      if (!needs && c.wall) patch.posts = [{ id: Date.now(), author: 'Sam', initials: 'S', color: '#F49AC1', meta: c.type + ' · Just now', text: item.title + (item.body ? '\n' + item.body : ''), likes: 0, extra: 0, shares: 0, comments: [] }].concat(this.state.posts);
      this.setState(patch);
      if (needs && ruleNeeds) {
        const apName = roleNameOf(ruleObj.approver);
        this.setState((x) => ({ flows: x.flows.map((fl) => (fl.id === 'f10' ? { ...fl, levels: 1, approvers: [apName, fl.approvers[1], fl.approvers[2]] } : fl)) }));
        makeRequest('f10', c.type + ' from ' + ruleRec.name + ': ' + item.title, { kind: 'comm', commId: id });
      } else if (needs) makeRequest('f1', c.type + ': ' + item.title, { kind: 'comm', commId: id });
      else { if (fo) this.addForumMsg(fo.id, c.type + ': ' + item.title); this.notify(c.type + ' published to ' + to.join(', ') + '.', '#B8F36B'); }
    }
  };
  const commNew = commTypes.map((t, i) => ({ label: 'New ' + t.toLowerCase(), run: () => openCm(t), style: 'height:40px;padding:0 16px;border-radius:20px;font-weight:600;cursor:pointer;' + (i === 0 ? 'border:0;background:#B8F36B;color:#0F1714' : 'border:1px solid var(--k-2A3832,#2A3832);background:transparent;color:var(--k-E8EFEA,#E8EFEA)') }));
  const commsShown = s.comms.filter((c) => commTypes.indexOf(c.type) !== -1);
  const commsTable = { id: 'comms', pageSize: 10, minWidth: '1000px', emptyText: 'Nothing sent yet.',
    cols: [{ label: 'Type', filter: 'select' }, { label: 'Title', filter: 'text' }, { label: 'Sent to', filter: 'text' }, { label: 'From', filter: 'select' }, { label: 'Date', filter: 'date' }, { label: 'Status', filter: 'select' }, { label: 'Read' }, { label: '', align: 'right' }],
    rows: commsShown.map((c) => ({ cells: [C.b(c.type, commColor[c.type]), C.t(c.title, { strong: true, sub: c.body }), C.t(c.to.join(', ')), C.t(c.by === 'Samwel Okoth Omulo' ? 'You' : c.by), C.t(fmtDay(c.date), { fv: c.date, nowrap: true }),
      C.b(c.status, c.status === 'Published' ? '#B8F36B' : '#FFB86B'), C.bar(c.read, c.read >= 80 ? '#B8F36B' : '#FFB86B'),
      C.a(c.status === 'Published' ? [{ label: 'Remind unread', run: () => this.notify('Reminder sent to everyone who has not read "' + c.title + '".', '#7CC4FF') }] : [{ label: 'See approval', kind: 'primary', run: () => this.setState({ section: 'approvals' }) }])] })) };
  return { cm, commNew, commsTable, openCm };
}
