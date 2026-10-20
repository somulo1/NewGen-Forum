import * as commsW from '~/logic/server/widgets/comms.js';
// What each kind of communication asks for besides the title, summary and message. Each field has a label and a hint so the form
// explains itself.
const FORMS = {
  Announcement: [
    { key: 'priority', label: 'Priority', kind: 'select', options: ['Normal', 'Important', 'Urgent'], help: 'Urgent ones are shown first.' },
    { key: 'effective', label: 'Takes effect on', kind: 'date', help: 'Leave empty if it applies straight away.' }
  ],
  Memo: [
    { key: 'ref', label: 'Reference number', kind: 'text', help: 'For example HR/2026/014. Optional.' },
    { key: 'actionBy', label: 'Action needed by', kind: 'date', help: 'The date people must act by, if they must.' },
    { key: 'actionFrom', label: 'Who must act', kind: 'text', help: 'For example "All line managers". Optional.' }
  ],
  Newsletter: [
    { key: 'issue', label: 'Issue or edition', kind: 'text', help: 'For example "October 2026" or "Issue 12".' }
  ],
  'Job post': [
    { key: 'team', label: 'Department or team', kind: 'text', help: 'Where the role sits.' },
    { key: 'location', label: 'Location', kind: 'text', help: 'Branch, city or "Remote".' },
    { key: 'employment', label: 'Employment type', kind: 'select', options: ['Full-time', 'Part-time', 'Contract', 'Internship'], help: '' },
    { key: 'closing', label: 'Closing date', kind: 'date', help: 'The last day to apply.' },
    { key: 'apply', label: 'How to apply', kind: 'area', help: 'Who to contact, or what to send.' }
  ]
};
const MODULE_OF = { Announcement: 'announcements', Memo: 'memos', Newsletter: 'newsletter', 'Job post': 'jobposts' };

// renderVals step: comms. Receives the names built by earlier steps and returns the ones later steps need.
export function buildComms(ctx) {
  const { C, O, can, chip, chipS, flowById, fmtDay, forums, has, makeRequest, recOf, roleNameOf, ruleSentence, s, set, todayISO, typeName } = ctx;
  // =============== Communications ===============
  const commTypes = [has.announcements && 'Announcement', has.memos && 'Memo', has.newsletter && 'Newsletter', has.jobposts && 'Job post'].filter(Boolean);
  const commColor = { Announcement: '#B8F36B', Memo: '#FFD66B', Newsletter: '#F49AC1', 'Job post': '#7CC4FF' };
  const openCm = (type) => this.setState({ cmOpen: true, cmd: { type: type, title: '', summary: '', body: '', fields: {}, files: [], wall: type !== 'Memo', forum: '', sel: {} } });
  const cmS = s.cmd;
  const setCm = (patch) => this.setState((x) => ({ cmd: Object.assign({}, x.cmd, patch) }));
  const f1 = flowById('f1');
  // Communication rules: the user's own level decides where memos/announcements may go
  const ruleLevel = ['departments', 'units', 'branches', 'sectors'].find((t) => has[t]);
  const ruleRec = ruleLevel ? recOf(ruleLevel, O.mine[ruleLevel]) : null;
  // a rule applies through the sender's own place in the structure; someone not placed in it has no level rule
  const ruleObj = ruleLevel && ruleRec && (cmS.type === 'Memo' || cmS.type === 'Announcement') ? (s.commRules[ruleLevel] || {})[cmS.type] : null;
  const ruleTo = ruleObj ? (ruleObj.on ? ruleObj.to : []) : null;
  const actingLocked = ((s.serverRoles || []).find((r) => r.id === s.actingRole) || {}).locked;
  const senderOk = !ruleObj || !ruleObj.sender || actingLocked || s.actingRole === ruleObj.sender;
  const ruleNeeds = !!(ruleObj && ruleObj.on && ruleObj.approver && ruleObj.approver !== 'none');
  const wallAllowed = !ruleTo || ruleTo.indexOf('wall') !== -1;
  const cmNeedsApproval = ruleNeeds || (cmS.type === 'Announcement' && f1 && f1.levels > 0);
  const cm = {
    heading: 'New ' + cmS.type.toLowerCase(), title: cmS.title, body: cmS.body, wall: cmS.wall && wallAllowed, forum: cmS.forum, wallAllowed: wallAllowed,
    types: commTypes.map((t) => ({ label: t, pick: () => setCm({ type: t }), style: chip(cmS.type === t) })),
    typeHelp: { Announcement: 'News for everyone it reaches. Staff mark it read.', Memo: 'An internal note, often asking for action.', Newsletter: 'A regular issue with news and updates.', 'Job post': 'An open role people can apply for.' }[cmS.type] || '',
    titleLabel: cmS.type === 'Job post' ? 'Job title' : 'Title', summary: cmS.summary || '', setSummary: (e) => setCm({ summary: e.target.value }),
    bodyLabel: cmS.type === 'Job post' ? 'Role description' : 'Message',
    fields: (FORMS[cmS.type] || []).map((f) => ({ ...f, id: 'cmf-' + f.key, isSelect: f.kind === 'select', isArea: f.kind === 'area', isInput: f.kind !== 'select' && f.kind !== 'area', type: f.kind === 'date' ? 'date' : 'text', value: (cmS.fields || {})[f.key] || '',
      onInput: (e) => setCm({ fields: Object.assign({}, this.state.cmd.fields, { [f.key]: e.target.value }) }) })),
    files: (cmS.files || []).map((f, i) => ({ name: f.name, size: this.fmtSize ? this.fmtSize(f.size) : Math.round(f.size / 1024) + ' KB', remove: () => setCm({ files: this.state.cmd.files.filter((_, j) => j !== i) }) })),
    hasFiles: (cmS.files || []).length > 0,
    fileRule: this.srv && this.srv.commFiles ? 'Up to ' + this.srv.commFiles.maxMB + ' MB each. Allowed: ' + (this.srv.commFiles.types.join(', ') || 'any type') + '. Your administrator sets this in Settings.' : '',
    accept: this.srv && this.srv.commFiles && this.srv.commFiles.types.length ? this.srv.commFiles.types.map((t) => '.' + t).join(',') : '',
    pickFiles: (e) => {
      const list = Array.from(e.target.files || []); e.target.value = '';
      const keep = [];
      list.forEach((f) => { const bad = this.srv ? commsW.fileProblem(this.srv, f) : ''; if (bad) this.notify(bad, '#FF9B8A'); else keep.push(f); });
      if (keep.length) setCm({ files: (this.state.cmd.files || []).concat(keep) });
    },
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
      const sel = Object.keys(c.sel).filter((k) => c.sel[k]);
      if (!(c.wall && wallAllowed) && !fo && !sel.length) { this.notify('Choose at least one place to send it.', '#FFB86B'); return; }
      this.setState({ cmOpen: false });
      commsW.publish(this, this.srv, { type: c.type, title: c.title.trim(), summary: (c.summary || '').trim(), body: c.body.trim(), fields: Object.fromEntries(Object.entries(c.fields || {}).filter(([, v]) => String(v).trim())), files: c.files || [], wall: c.wall && wallAllowed, forum: fo ? { id: this.srv.forumConv && this.srv.forumConv[fo.id], name: fo.name } : null, sel: c.sel })
        .then((r) => this.notify(r.pending ? c.type + ' sent for approval.' : c.type + ' published.', r.pending ? '#FFB86B' : '#B8F36B'))
        .catch((e) => this.notify(e.message || 'Could not send it.', '#FF9B8A'));
    }
  };
  const commNew = commTypes.map((t, i) => ({ label: 'New ' + t.toLowerCase(), run: () => openCm(t), style: 'height:40px;padding:0 16px;border-radius:20px;font-weight:600;cursor:pointer;' + (i === 0 ? 'border:0;background:#B8F36B;color:#0F1714' : 'border:1px solid var(--k-2A3832,#2A3832);background:transparent;color:var(--k-E8EFEA,#E8EFEA)') }));
  const canInsights = (type) => !!this.srv && can(MODULE_OF[type], 'View insights');
  const tab = s.commTab && commTypes.includes(s.commTab) ? s.commTab : '';
  const commsTabs = [{ label: 'All', type: '' }].concat(commTypes.map((t) => ({ label: t === 'Job post' ? 'Job posts' : t + 's', type: t }))).map((t) => ({ ...t, count: s.comms.filter((c) => !t.type ? commTypes.includes(c.type) : c.type === t.type).length, on: tab === t.type, pick: () => this.setState({ commTab: t.type }),
    style: 'height:40px;padding:0 16px;border:0;border-bottom:3px solid ' + (tab === t.type ? '#B8F36B' : 'transparent') + ';background:transparent;color:' + (tab === t.type ? 'var(--k-E8EFEA,#E8EFEA)' : 'var(--k-9DB0A6,#9DB0A6)') + ';font-weight:600;font-size:14px;cursor:pointer' }));
  const openComm = (c) => { this.setState({ commOpen: c.uid, commView: 'read', commIns: null }); if (this.srv && c.status === 'Published' && !c.iRead) commsW.markRead(this, this.srv, c).catch(() => {}); };
  const commsShown = s.comms.filter((c) => (tab ? c.type === tab : commTypes.indexOf(c.type) !== -1));
  const commsTable = { id: 'comms-' + (tab || 'all'), pageSize: 10, minWidth: '1000px', emptyText: 'Nothing sent yet.',
    cols: [{ label: 'Type', filter: 'select' }, { label: 'Title', filter: 'text' }, { label: 'Sent to', filter: 'text' }, { label: 'From', filter: 'select' }, { label: 'Date', filter: 'date' }, { label: 'Status', filter: 'select' }, { label: 'Read' }, { label: '', align: 'right' }],
    rows: commsShown.map((c) => ({ cells: [C.b(c.type, commColor[c.type]), C.t(c.title, { strong: true, sub: c.summary || c.body }), C.t(c.to.join(', ')), C.t(c.by === 'Samwel Okoth Omulo' ? 'You' : c.by), C.t(fmtDay(c.date), { fv: c.date, nowrap: true }),
      C.b(c.status, c.status === 'Published' ? '#B8F36B' : '#FFB86B'), C.bar(c.read, c.read >= 80 ? '#B8F36B' : '#FFB86B'),
      C.a([{ label: 'View', kind: 'primary', run: () => openComm(c) }].concat(canInsights(c.type) && c.status === 'Published' ? [{ label: 'Insights', run: () => { openComm(c); this.setState({ commView: 'insights' }); loadInsights(c); } }] : []).concat(c.status === 'Published' ? [] : [{ label: 'See approval', run: () => this.setState({ section: 'approvals' }) }]))] })) };
  const loadInsights = (c) => {
    this.setState({ commIns: { loading: true, places: [], err: '' } });
    commsW.insights(this.srv, c).then((places) => this.setState({ commIns: { loading: false, places, err: '' } })).catch((e) => this.setState({ commIns: { loading: false, places: [], err: e.message || 'Could not load.' } }));
  };
  // The page for reading one communication, with (for people allowed) a second page measuring how it was received.
  const oc = s.commOpen ? s.comms.find((c) => c.uid === s.commOpen) : null;
  const sum = (k) => ((s.commIns && s.commIns.places) || []).reduce((n, x) => n + x[k], 0);
  const commView = oc ? {
    open: true, close: () => this.setState({ commOpen: null, commIns: null }), type: oc.type, typeColor: commColor[oc.type], title: oc.title, summary: oc.summary, hasSummary: !!oc.summary, body: oc.body,
    meta: (oc.by === 'Samwel Okoth Omulo' ? 'You' : oc.by) + (oc.byDept ? ' · ' + oc.byDept : '') + ' · ' + fmtDay(oc.date) + ' · ' + oc.status,
    to: oc.to.join(', '), hasTo: oc.to.length > 0,
    fields: (FORMS[oc.type] || []).filter((f) => (oc.fields || {})[f.key]).map((f) => ({ label: f.label, value: f.kind === 'date' ? fmtDay(oc.fields[f.key]) : oc.fields[f.key] })),
    files: (oc.files || []).map((f) => ({ name: f.name, meta: f.meta || '', open: () => commsW.openFile(this.srv, f).catch((e) => this.notify(e.message || 'Could not open it.', '#FF9B8A')) })),
    hasFiles: (oc.files || []).length > 0,
    canInsights: canInsights(oc.type) && oc.status === 'Published', isRead: s.commView !== 'insights', isInsights: s.commView === 'insights',
    showRead: () => this.setState({ commView: 'read' }), showInsights: () => { this.setState({ commView: 'insights' }); loadInsights(oc); },
    ins: s.commIns || { loading: false, places: [], err: '' }, hasPlaces: !!(s.commIns && s.commIns.places.length),
    places: ((s.commIns && s.commIns.places) || []).map((x) => ({ ...x, place: x.place || ((forums.find((f) => this.srv && this.srv.forumConv && this.srv.forumConv[f.id] === x.conversation) || {}).name ? (forums.find((f) => this.srv.forumConv[f.id] === x.conversation).name + ' forum') : 'Forum'), kindLabel: x.kind === 'wall' ? 'Community wall' : 'Forum', readPct: x.audience ? Math.round((x.read / x.audience) * 100) + '%' : '0%' })),
    totals: { audience: sum('audience'), read: sum('read'), likes: sum('likes'), replies: sum('replies') }
  } : { open: false };
  return { cm, commNew, commsTable, commsTabs, commView, openCm };
}
