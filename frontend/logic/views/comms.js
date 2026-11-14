import { askConfirm, confirmed } from '~/logic/confirm.js';
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
  const { C, O, can, chip, chipS, flowById, fmtDay, forums, has, holdsRef, makeRequest, recOf, roleNameOf, ruleSentence, s, set, todayISO, typeName } = ctx;
  // =============== Communications ===============
  const commTypes = [has.announcements && 'Announcement', has.memos && 'Memo', has.newsletter && 'Newsletter', has.jobposts && 'Job post'].filter(Boolean);
  const commColor = { Announcement: '#B8F36B', Memo: '#FFD66B', Newsletter: '#F49AC1', 'Job post': '#7CC4FF' };
  const openCm = (type) => this.setState({ cmOpen: true, cmd: { type: type, title: '', summary: '', body: '', fields: {}, files: [], wall: type !== 'Memo', forum: '', forumSel: {}, sel: {} } });
  const cmS = s.cmd;
  const setCm = (patch) => this.setState((x) => ({ cmd: Object.assign({}, x.cmd, patch) }));
  const f1 = flowById('f1');
  // Communication rules: the user's own level decides where memos/announcements may go
  const ruleLevel = ['departments', 'units', 'branches', 'sectors'].find((t) => has[t]);
  const ruleRec = ruleLevel ? recOf(ruleLevel, O.mine[ruleLevel]) : null;
  // a rule applies through the sender's own place in the structure; someone not placed in it has no level rule
  // Per-department sending rules are not in use for now (they are no longer set in the widget settings), so none applies here.
  const ruleObj = null;
  const ruleTo = ruleObj ? (ruleObj.on ? ruleObj.to : []) : null;
  const actingLocked = ((s.serverRoles || []).find((r) => r.id === s.actingRole) || {}).locked;
  const senderOk = !ruleObj || !ruleObj.sender || actingLocked || holdsRef(ruleObj.sender);
  const ruleNeeds = !!(ruleObj && ruleObj.on && ruleObj.approver && ruleObj.approver !== 'none');
  const wallAllowed = !ruleTo || ruleTo.indexOf('wall') !== -1;
  const cmNeedsApproval = ruleNeeds || (cmS.type === 'Announcement' && f1 && f1.levels > 0);
  const cm = {
    heading: 'New ' + cmS.type.toLowerCase(), title: cmS.title, body: cmS.body, wall: cmS.wall && wallAllowed, forum: cmS.forum, wallAllowed: wallAllowed,
    // any number of forums can receive it: tick each one
    forumList: forums.map((f) => { const on = !!(cmS.forumSel || {})[f.id]; return { id: f.id, name: f.name, on, style: chip(on), toggle: () => setCm({ forumSel: { ...(this.state.cmd.forumSel || {}), [f.id]: !on } }) }; }),
    forumCount: forums.filter((f) => (cmS.forumSel || {})[f.id]).length,
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
      const fos = forums.filter((f) => (c.forumSel || {})[f.id] || f.id === c.forum);
      const sel = Object.keys(c.sel).filter((k) => c.sel[k]);
      if (!(c.wall && wallAllowed) && !fos.length && !sel.length) { this.notify('Choose at least one place to send it.', '#FFB86B'); return; }
      this.setState({ cmOpen: false });
      commsW.publish(this, this.srv, { type: c.type, title: c.title.trim(), summary: (c.summary || '').trim(), body: c.body.trim(), fields: Object.fromEntries(Object.entries(c.fields || {}).filter(([, v]) => String(v).trim())), files: c.files || [], wall: c.wall && wallAllowed, forums: fos.map((fo) => ({ id: this.srv.forumConv && this.srv.forumConv[fo.id], name: fo.name })), sel: c.sel })
        .then((r) => this.notify(!r.pending ? c.type + ' published to ' + r.sent + (r.sent === 1 ? ' place.' : ' places.') : r.pending >= r.sent ? c.type + ' sent for approval. It reaches ' + (r.sent === 1 ? 'its place' : 'all ' + r.sent + ' places') + ' once approved.' : c.type + ' published to ' + (r.sent - r.pending) + ' of ' + r.sent + ' places. The other ' + r.pending + ' wait for approval.', r.pending ? '#FFB86B' : '#B8F36B'))
        .catch((e) => this.notify(e.message || 'Could not send it.', '#FF9B8A'));
    }
  };
  const commNew = commTypes.map((t, i) => ({ label: 'New ' + t.toLowerCase(), run: () => openCm(t), style: 'height:40px;padding:0 16px;border-radius:20px;font-weight:600;cursor:pointer;' + (i === 0 ? 'border:0;background:#B8F36B;color:#0F1714' : 'border:1px solid var(--k-2A3832,#2A3832);background:transparent;color:var(--k-E8EFEA,#E8EFEA)') }));
  const canInsights = (type) => !!this.srv && can(MODULE_OF[type], 'View insights');
  // Repost: post a published item again to the places it went to, for people given the page's "Repost" right
  const canRepost = (c) => !!this.srv && c.status === 'Published' && can(MODULE_OF[c.type], 'Repost');
  const repostNow = async (c) => {
    if (this._reposting) return;
    if (!await askConfirm({ title: 'Repost this?', body: '"' + c.title + '" goes to ' + ((c.to || []).join(', ') || 'the places it went to') + ' again, and everyone there gets it again.', action: 'Repost' })) return;
    if (this._reposting) return;
    this._reposting = true; this.forceUpdate();
    commsW.repost(this, this.srv, c)
      .then((r) => this.notify(r.pending >= r.sent ? c.type + ' reposted. It reaches ' + (r.sent === 1 ? 'its place' : 'all ' + r.sent + ' places') + ' once approved.' : c.type + ' reposted to ' + r.sent + (r.sent === 1 ? ' place' : ' places') + (r.skipped ? ' (' + r.skipped + ' you are not a member of were skipped)' : '') + '.', r.pending ? '#FFB86B' : '#B8F36B'))
      .catch((e) => this.notify((e && e.message) || 'Could not repost it.', '#FF9B8A'))
      .finally(() => { this._reposting = false; this.forceUpdate(); });
  };
  const tab = s.commTab && commTypes.includes(s.commTab) ? s.commTab : '';
  const commsTabs = [{ label: 'All', type: '' }].concat(commTypes.map((t) => ({ label: t === 'Job post' ? 'Job posts' : t + 's', type: t }))).map((t) => ({ ...t, count: s.comms.filter((c) => !t.type ? commTypes.includes(c.type) : c.type === t.type).length, on: tab === t.type, pick: () => this.setState({ commTab: t.type }),
    style: 'height:40px;padding:0 16px;border:0;border-bottom:3px solid ' + (tab === t.type ? '#B8F36B' : 'transparent') + ';background:transparent;color:' + (tab === t.type ? 'var(--k-E8EFEA,#E8EFEA)' : 'var(--k-9DB0A6,#9DB0A6)') + ';font-weight:600;font-size:14px;cursor:pointer' }));
  const openComm = (c) => { this.setState({ commOpen: c.uid, commView: 'read', commIns: null }); if (this.srv && c.status === 'Published' && !c.iRead) commsW.markRead(this, this.srv, c).catch(() => {}); if (canInsights(c.type) && c.status === 'Published') loadInsights(c); };
  const commsShown = s.comms.filter((c) => (tab ? c.type === tab : commTypes.indexOf(c.type) !== -1));
  const commsTable = { id: 'comms-' + (tab || 'all'), pageSize: 10, minWidth: '1000px', emptyText: 'Nothing sent yet.',
    cols: [{ label: 'Type', filter: 'select' }, { label: 'Title', filter: 'text' }, { label: 'Sent to', filter: 'text' }, { label: 'From', filter: 'select' }, { label: 'Date', filter: 'date' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
    rows: commsShown.map((c) => ({ cells: [C.b(c.type, commColor[c.type]), C.t(c.title, { strong: true, sub: c.summary || c.body }), C.t(c.to.length > 2 ? c.to.slice(0, 2).join(', ') + ' +' + (c.to.length - 2) + ' more' : c.to.join(', '), { fv: c.to.join(', ') }), C.t(c.by === 'Samwel Okoth Omulo' ? 'You' : c.by), C.t(fmtDay(c.date), { fv: c.date, nowrap: true }),
      C.b(c.status, c.status === 'Published' ? '#B8F36B' : '#FFB86B'),
      C.a([{ label: 'View', kind: 'primary', run: () => openComm(c) }].concat(canRepost(c) ? [{ label: 'Repost', run: () => repostNow(c) }] : []).concat(canInsights(c.type) && c.status === 'Published' ? [{ label: 'Insights', run: () => openComm(c) }] : []).concat(c.status === 'Published' ? [] : [{ label: 'See approval', run: () => this.setState({ section: 'approvals' }) }]))] })) };
  const loadInsights = (c) => {
    this.setState({ commIns: { loading: true, places: [], err: '' } });
    commsW.insights(this.srv, c).then(({ places, totals }) => this.setState({ commIns: { loading: false, places, totals, err: '' } })).catch((e) => this.setState({ commIns: { loading: false, places: [], err: e.message || 'Could not load.' } }));
  };
  // The page for reading one communication, with (for people allowed) a second page measuring how it was received.
  const oc = s.commOpen ? s.comms.find((c) => c.uid === s.commOpen) : null;
  const sum = (k) => ((s.commIns && s.commIns.places) || []).reduce((n, x) => n + (x[k] || 0), 0);
  // ---- measurability: did the communication reach its people, how fast, and who is still missing ----
  const TARGET = 80;   // a communication counts as having landed when 80% of its audience has read it
  const pctOf = (a, b) => (b ? Math.min(100, Math.round((a / b) * 100)) : 0);
  const fmtMins = (m) => (m == null ? '—' : m < 60 ? m + ' min' : m < 1440 ? Math.floor(m / 60) + ' h ' + (m % 60 ? (m % 60) + ' min' : '') : Math.round(m / 1440) + ' d');
  // the whole communication, each person counted once across every place (from the server's totals; never by adding places up)
  const measureOf = (tot) => {
    const T0 = tot || { audience: 0, read: 0, read_in_48h: 0, acknowledged: 0, median_read_mins: null, unread: [], by_department: [] };
    const audience = T0.audience, read = T0.read, read48 = T0.read_in_48h, acked = T0.acknowledged, unread = T0.unread || [];
    const byDept = (T0.by_department || []).map((d) => ({ department: d.department, audience: d.audience, read: d.read, acked: d.acknowledged, readPct: pctOf(d.read, d.audience), ackPct: pctOf(d.acknowledged, d.audience) })).sort((x, y) => x.readPct - y.readPct);
    return { audience, read, read48, acked, readPct: pctOf(read, audience), read48Pct: pctOf(read48, audience), ackPct: pctOf(acked, audience),
      medianMins: T0.median_read_mins == null ? null : T0.median_read_mins, median: fmtMins(T0.median_read_mins == null ? null : T0.median_read_mins), unread, unreadN: unread.length, byDept, landed: pctOf(read, audience) >= TARGET };
  };
  const commView = oc ? {
    open: true, ref: oc.id, status: oc.status, statusColor: oc.status === 'Published' ? '#B8F36B' : '#FFB86B', close: () => this.setState({ commOpen: null, commIns: null }), type: oc.type, typeColor: commColor[oc.type], title: oc.title, summary: oc.summary, hasSummary: !!oc.summary, body: oc.body,
    meta: (oc.by === 'Samwel Okoth Omulo' ? 'You' : oc.by) + (oc.byDept ? ' · ' + oc.byDept : '') + ' · ' + fmtDay(oc.date) + ' · ' + oc.status,
    to: oc.to.join(', '), toList: oc.to.slice(), hasTo: oc.to.length > 0, readNum: Number(oc.read) || 0, fromInitials: ((oc.by === 'Samwel Okoth Omulo' ? 'You' : oc.by) || '?').split(' ').map((x) => x[0]).join('').slice(0, 2).toUpperCase(),
    fields: (FORMS[oc.type] || []).filter((f) => (oc.fields || {})[f.key]).map((f) => ({ label: f.label, value: f.kind === 'date' ? fmtDay(oc.fields[f.key]) : oc.fields[f.key] })),
    files: (oc.files || []).map((f) => ({ name: f.name, meta: f.meta || '', open: () => commsW.openFile(this.srv, f).catch((e) => this.notify(e.message || 'Could not open it.', '#FF9B8A')) })),
    hasFiles: (oc.files || []).length > 0,
    canInsights: canInsights(oc.type) && oc.status === 'Published', showStats: canInsights(oc.type), canRepost: canRepost(oc), repost: () => repostNow(oc), reposting: !!this._reposting, isRead: s.commView !== 'insights', isInsights: s.commView === 'insights',
    showRead: () => this.setState({ commView: 'read' }), showInsights: () => { this.setState({ commView: 'insights' }); loadInsights(oc); },
    from: oc.by === 'Samwel Okoth Omulo' ? 'You' : oc.by, byDept: oc.byDept, hasDept: !!oc.byDept, dateLabel: fmtDay(oc.date), readPct: oc.read + '%',
    ins: s.commIns || { loading: false, places: [], err: '' }, hasPlaces: !!(s.commIns && s.commIns.places.length),
    places: ((s.commIns && s.commIns.places) || []).map((x) => ({ ...x, place: x.place || ((forums.find((f) => this.srv && this.srv.forumConv && this.srv.forumConv[f.id] === x.conversation) || {}).name ? (forums.find((f) => this.srv.forumConv[f.id] === x.conversation).name + ' forum') : 'Forum'), kindLabel: x.kind === 'wall' ? 'Community wall' : 'Forum', readPct: x.audience ? Math.round((x.read / x.audience) * 100) + '%' : '0%' })),
    // people counts come from the per-person totals; reactions and replies are message counts, so they add up per place
    totals: { audience: (s.commIns && s.commIns.totals) ? s.commIns.totals.audience : 0, read: (s.commIns && s.commIns.totals) ? s.commIns.totals.read : 0, likes: sum('likes'), replies: sum('replies'), acked: (s.commIns && s.commIns.totals) ? s.commIns.totals.acknowledged : 0 },
    measure: (() => {
      const m = measureOf(s.commIns && s.commIns.totals);
      const remindState = (s.commRemind && s.commRemind.uid === oc.uid) ? s.commRemind : {};
      return { ...m, target: TARGET, unreadShown: m.unread.slice(0, 40), moreUnread: Math.max(0, m.unread.length - 40), reminding: !!remindState.busy, remindNote: remindState.note || '',
        remind: () => {
          if (!m.unread.length || remindState.busy) return;
          this.setState({ commRemind: { uid: oc.uid, busy: true, note: '' } });
          commsW.remind(this.srv, oc).then((n) => { this.setState({ commRemind: { uid: oc.uid, busy: false, note: 'Reminder sent to ' + n + (n === 1 ? ' person' : ' people') + ' who had not read it.' } }); this.notify('Reminder sent to ' + n + ' people.', '#B8F36B'); })
            .catch((e) => this.setState({ commRemind: { uid: oc.uid, busy: false, note: (e && e.message) || 'Could not send the reminder.' } }));
        } };
    })(),
    // who acknowledged it, across every place it went to
    ackedBy: (s.commIns && s.commIns.totals && s.commIns.totals.acknowledged_by) || [],
    canAck: !!this.srv && oc.status === 'Published' && !oc.mine && !oc.iAcked, acked: !!oc.iAcked,
    acknowledge: () => commsW.acknowledge(this, this.srv, oc).then(() => this.notify('Acknowledged.', '#B8F36B')).catch((e) => this.notify(e.message || 'Could not acknowledge it.', '#FF9B8A')),
  } : { open: false };
  // ---- Engagement: how each published communication was received (its own page, for people who may see insights) ----
  const engTypes = commTypes.filter((t) => canInsights(t));
  const E = s.commEng || { loading: false, err: '', rows: [] };
  const loadEngagement = async () => {
    if (!this.srv || this.state.commEng.loading) return;
    this.setState({ commEng: { loading: true, err: '', rows: this.state.commEng.rows || [], at: 0 } });
    try {
      // one request measures everything; copies sent together (same batch) are one row
      const list = await commsW.engagement(this.srv);
      const groups = new Map();
      list.forEach((x) => groups.set(x.comm, [x]));   // the server returns one row per communication, people counted once
      const rows = [];
      groups.forEach((g) => {
        const c = this.state.comms.find((y) => y.uid === g[0].comm || (g[0].batch && y.batch === g[0].batch)) || { uid: g[0].comm, id: g[0].ref, type: g[0].type, title: g[0].ref, date: g[0].at, at: g[0].at };
        const x0 = g[0], audience = x0.audience, read = x0.read, read48 = x0.read_in_48h, acked = x0.acknowledged, medianMins = x0.median_read_mins == null ? null : x0.median_read_mins;
        rows.push({ c, at: g[0].at, places: g.length, audience, read, acked, likes: 0, replies: 0, ackedBy: [],
          m: { audience, read, read48, acked, readPct: pctOf(read, audience), read48Pct: pctOf(read48, audience), ackPct: pctOf(acked, audience), medianMins, median: fmtMins(medianMins), unreadN: Math.max(0, audience - read), landed: pctOf(read, audience) >= TARGET } });
      });
      rows.sort((a, b) => (a.at < b.at ? 1 : -1));
      this.setState({ commEng: { loading: false, err: '', rows, at: Date.now() } });
    } catch (e) { this.setState({ commEng: { loading: false, err: (e && e.message) || 'Could not load engagement.', rows: [], at: 0 } }); }
  };
  const pct = (r) => (r.audience ? Math.min(100, Math.round((r.read / r.audience) * 100)) : 0);
  const engSum = (k) => E.rows.reduce((n, r) => n + r[k], 0);
  const engagementTable = { id: 'engagement', pageSize: 25, minWidth: '1040px', emptyText: E.loading ? 'Measuring…' : 'Nothing published yet.',
    cols: [{ label: 'Type', filter: 'select' }, { label: 'Title', filter: 'text' }, { label: 'Date', filter: 'date' }, { label: 'Reached' }, { label: 'Read' }, { label: 'Within 48 h' }, { label: 'Acknowledged' }, { label: 'Time to read' }, { label: 'Still to read' }, { label: 'Action', align: 'right' }],
    rows: E.rows.filter((r) => r.m).map((r) => ({ cells: [C.b(r.c.type, commColor[r.c.type]), C.t(r.c.title, { strong: true, sub: r.places + (r.places === 1 ? ' place' : ' places') }), C.t(fmtDay(r.c.date), { fv: r.c.date, nowrap: true }),
      C.t(String(r.m.audience), { fv: r.m.audience }),
      C.bar(r.m.readPct, r.m.readPct >= 80 ? '#B8F36B' : r.m.readPct >= 40 ? '#FFB86B' : '#FF9B8A', r.m.read + ' (' + r.m.readPct + '%)'),
      C.t(r.m.read48Pct + '%', { fv: r.m.read48Pct }), C.t(r.m.ackPct + '%', { sub: r.m.acked + ' people', fv: r.m.ackPct }), C.t(r.m.median, { fv: r.m.medianMins == null ? 1e9 : r.m.medianMins }),
      C.t(String(r.m.unreadN), { sub: r.m.landed ? 'target met' : 'below 80% target', fv: r.m.unreadN }),
      C.a([{ label: 'View', kind: 'primary', run: () => { this.setState({ commPage: '' }); openComm(r.c); this.setState({ commView: 'insights' }); } }])] })) };
  const engagement = { open: s.commPage === 'engagement' && engTypes.length > 0, canOpen: engTypes.length > 0, loading: E.loading, err: E.err, hasErr: !!E.err,
    openIt: () => { this.setState({ commPage: 'engagement', commOpen: null }); loadEngagement(); }, close: () => this.setState({ commPage: '' }), refresh: loadEngagement,
    // the measures that answer "did our communication reach people?": what WhatsApp can never show
    kpis: (() => {
      const R = E.rows.filter((r) => r.m), aud = R.reduce((n, r) => n + r.m.audience, 0), avg = (f) => (R.length ? Math.round(R.reduce((n, r) => n + f(r), 0) / R.length) : 0);
      const meds = R.map((r) => r.m.medianMins).filter((x) => x != null).sort((a, b) => a - b);
      return [{ label: 'Published', n: R.length }, { label: 'Average read rate', n: avg((r) => r.m.readPct) + '%' }, { label: 'Read within 48 hours', n: avg((r) => r.m.read48Pct) + '%' },
        { label: 'Acknowledged', n: avg((r) => r.m.ackPct) + '%' }, { label: 'Reached the 80% target', n: R.filter((r) => r.m.landed).length + ' of ' + R.length },
        { label: 'Median time to read', n: fmtMins(meds.length ? meds[Math.floor(meds.length / 2)] : null) }, { label: 'People still to read', n: R.reduce((n, r) => n + r.m.unreadN, 0) + ' of ' + aud }];
    })(),
    avgRead: null,
    // week by week: of everyone reached by what was published that week, how many read it (the day-14 / day-90 proof)
    trend: (() => {
      const wk = (d) => { const x = new Date(d); x.setHours(0, 0, 0, 0); x.setDate(x.getDate() - ((x.getDay() + 6) % 7)); return x.getTime(); };
      const by = new Map();
      E.rows.filter((r) => r.m).forEach((r) => { const k = wk(r.at); const v = by.get(k) || { aud: 0, read: 0, n: 0 }; v.aud += r.m.audience; v.read += r.m.read; v.n++; by.set(k, v); });
      return [...by.entries()].sort((a, b) => a[0] - b[0]).slice(-12).map(([k, v]) => { const p = pctOf(v.read, v.aud); const d = new Date(k);
        return { label: d.getDate() + '/' + (d.getMonth() + 1), pct: p, n: v.n, color: p >= TARGET ? '#B8F36B' : p >= 40 ? '#FFB86B' : '#FF9B8A', title: 'Week of ' + d.toDateString() + ': ' + v.read + ' of ' + v.aud + ' read (' + p + '%), ' + v.n + ' published' }; });
    })() };
  return { cm, commNew, commsTable, commsTabs, commView, engagement, engagementTable, openCm };
}
