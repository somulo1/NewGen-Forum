// renderVals step: org. Receives the names built by earlier steps and returns the ones later steps need.
export function buildOrg(ctx) {
  const { C, badge, cur_t, fmtDay, has, isoAdd, projects, roleDefs, s, tagNames, tatText, tickets, todayISO, ts } = ctx;
  // =============== Organization structure ===============
  const O = this.org;
  const typeName = { sectors: 'Sector', branches: 'Branch', units: 'Unit', departments: 'Department' };
  const typeTint = { sectors: '#FFB86B', branches: '#7CC4FF', units: '#C9B6FF', departments: '#B8F36B' };
  const mgrRole = { departments: 'Head of department', sectors: 'Sector manager', branches: 'Branch manager', units: 'Unit lead' };
  const levels = ['sectors', 'branches', 'units', 'departments'].filter((t) => has[t]);
  const parentFor = (type, r) => {
    if (type === 'sectors') return null;
    if (type === 'branches') return has.sectors ? ['sectors', r.sector] : null;
    if (type === 'units') return has.branches ? ['branches', r.branch] : null;
    return has.branches ? ['branches', r.branch] : has.sectors ? ['sectors', r.sector] : null;
  };
  const recOf = (type, id) => (O[type] || []).find((x) => x.id === id);
  const pathOf = (type, r) => { const out = []; let p = parentFor(type, r); while (p) { const pr = recOf(p[0], p[1]); if (!pr) break; out.unshift(pr.name); p = parentFor(p[0], pr); } return ['[YOUR ORGANIZATION]'].concat(out).join(' › '); };
  const membersOf = (type, name) => this.dir.filter((p) => (type === 'departments' ? p.dept : type === 'branches' ? p.loc : type === 'sectors' ? p.sector : p.unit) === name);
  const shortOf = (n) => n.split(' ').map((x) => x[0]).join('').slice(0, 2).toUpperCase();
  // ---- Ticket escalation ----
  const escSeed = (str) => { let h = 11; for (let i = 0; i < str.length; i++) h = (h * 33 + str.charCodeAt(i)) >>> 0; return h; };
  const deptCats = { 'IT Operations': ['Network & Wi-Fi', 'Hardware', 'Access request'], 'Software Development': ['Software bug', 'Change request'], 'HR': ['HR request', 'Payroll query'], 'Finance': ['Payment query', 'Budget approval'], 'Risk & Compliance': ['Security incident', 'Compliance query'], 'Customer Care': ['Customer complaint'], 'Legal': ['Contract review'] };
  const escLevels = ['departments', 'units', 'branches', 'sectors'].filter((t) => has[t]);
  const escTypes = escLevels.length ? escLevels : ['desk'];
  const escRecs = (type) => type === 'desk' ? [{ id: 'desk', name: 'Ticket team' }] : (O[type] || []).filter((r) => !(cur_t && type === 'departments' && r.name === cur_t.dept));
  const escCats = (type, r) => type === 'desk' ? ts.cats.map((c) => c.name) : type === 'departments' ? (deptCats[r.name] || [r.name + ' request', 'General enquiry']) : ['Escalated tickets', 'Service complaint', 'Approval & sign-off'];
  const escPeople = (type, r, cat) => {
    const pool = type === 'desk' ? this.dir.filter((p) => tagNames.indexOf(p.name) !== -1) : membersOf(type, r.name);
    if (!pool.length) return [];
    const named = pool.filter((p) => tagNames.indexOf(p.name) !== -1);
    const h = escSeed(r.name + cat), n = Math.min(pool.length, 3);
    const out = named.slice(0, 1);
    for (let i = 0; out.length < n && i < pool.length; i++) { const p = pool[(h + i * 7) % pool.length]; if (out.indexOf(p) === -1) out.push(p); }
    return out;
  };
  const ES = s.esc;
  const escType = escTypes.indexOf(ES.type) !== -1 ? ES.type : escTypes[0];
  const escRecList = escRecs(escType);
  const escRec = escRecList.find((r) => r.id === ES.id) || (escType === 'desk' ? escRecList[0] : null);
  const escCatList = escRec ? escCats(escType, escRec) : [];
  const escCat = escCatList.indexOf(ES.cat) !== -1 ? ES.cat : '';
  const escPpl = escRec && escCat ? escPeople(escType, escRec, escCat) : [];
  const escPerson = escPpl.find((p) => p.name === ES.person) ? ES.person : '';
  const setEsc = (patch) => this.setState((st) => ({ esc: Object.assign({}, st.esc, patch) }));
  const escLabel = (t) => (t === 'desk' ? 'Ticket team' : typeName[t]);
  const escReady = !!(escRec && escCat && escPerson);
  const esc = {
    open: s.escOpen, btnLabel: s.escOpen ? 'Cancel' : 'Escalate ticket', canEscalate: !!cur_t && cur_t.step < 3,
    toggle: () => this.setState((st) => ({ escOpen: !st.escOpen, esc: { type: '', id: '', cat: '', person: '', why: '' } })),
    types: escTypes.map((t) => ({ label: escLabel(t), pick: () => setEsc({ type: t, id: '', cat: '', person: '' }), style: 'height:34px;padding:0 12px;border-radius:17px;cursor:pointer;font-size:13px;' + (t === escType ? 'border:1px solid #FF9B8A;background:#3A2420;color:#FFD3CA;font-weight:600' : 'border:1px solid #2A3832;background:#121B17;color:#C9D6CF') })),
    typeLabel: escLabel(escType), showRec: escType !== 'desk',
    recs: escRecList.map((r) => ({ id: r.id, name: r.name })), recValue: escRec ? escRec.id : '',
    onRec: (e) => setEsc({ id: e.target.value, cat: '', person: '' }),
    hasRec: !!escRec, cats: escCatList.map((c) => ({ name: c, tat: tatText(c) })), catValue: escCat,
    onCat: (e) => setEsc({ cat: e.target.value, person: '' }),
    hasCat: !!escCat, people: escPpl.map((p) => ({ name: p.name, title: p.title })), personValue: escPerson,
    noPeople: !!escCat && escPpl.length === 0,
    onPerson: (e) => setEsc({ person: e.target.value }),
    why: ES.why, onWhy: (e) => setEsc({ why: e.target.value }),
    btnStyle: 'height:42px;border-radius:10px;border:0;font-weight:600;cursor:pointer;font-size:14px;' + (escReady ? 'background:#FF9B8A;color:#0F1714' : 'background:#2A3832;color:#9DB0A6'),
    submit: () => {
      if (!cur_t || !escReady) { this.notify('Pick where to escalate, the category and the person.', '#FFB86B'); return; }
      const dest = escType === 'desk' ? 'the ticket team' : typeName[escType] + ' ' + escRec.name;
      const why = (this.state.esc.why || '').trim();
      this.updTicket(cur_t.id, (t) => Object.assign({}, t, {
        escalated: true, escalatedTo: (escType === 'desk' ? 'Ticket team' : typeName[escType] + ': ' + escRec.name) + ' · ' + escCat,
        dept: escType === 'departments' ? escRec.name : t.dept, cat: escCat, handler: escPerson, handlerSince: 'Just now',
        assignees: t.assignees.indexOf(escPerson) === -1 ? t.assignees.concat([escPerson]) : t.assignees, step: Math.max(1, Math.min(t.step, 2)), hold: false,
        activity: this.act(t, { text: 'Escalated to ' + dest + ' (' + escCat + '), assigned to ' + escPerson + (why ? '. Reason: ' + why : '') })
      }));
      this.setState({ escOpen: false, esc: { type: '', id: '', cat: '', person: '', why: '' } });
      this.notify(cur_t.id + ' escalated to ' + dest + '. ' + escPerson + ' has been notified.', '#FF9B8A', cur_t.id);
    }
  };
  const seedOf = (str) => { let h = 7; for (let i = 0; i < str.length; i++) h = (h * 31 + str.charCodeAt(i)) >>> 0; return h; };
  const unitData = (type, r) => {
    let h = seedOf(type + r.id);
    const rnd = () => { h = (h * 1664525 + 1013904223) >>> 0; return h / 4294967296; };
    const pick = (arr) => arr[Math.floor(rnd() * arr.length)];
    const pOn = 0.86 + (seedOf(r.id + 'p') % 14) / 100;
    const others = levels.reduce((acc, t) => acc.concat(O[t].filter((x) => x.name !== r.name).map((x) => x.name)), []);
    const cps = others.length ? others : ['Head Office', 'Finance', 'IT Operations'];
    const comms = Array.from({ length: 26 }, (_, i) => {
      const dirn = rnd() < 0.6 ? 'Received' : 'Sent';
      const ok = rnd() < pOn;
      return { date: isoAdd(todayISO, -Math.floor(rnd() * 30)), dir: dirn, type: pick(['Memo', 'Announcement', 'Newsletter', 'Email']), subject: pick(['Updated leave policy', 'Quarter-end close dates', 'System maintenance window', 'New password rules', 'Branch opening hours', 'Audit evidence request', 'Customer complaint trends', 'Training schedule', 'Budget submission reminder', 'Holiday rota']), cp: pick(cps),
        status: dirn === 'Received' ? (ok ? 'Acknowledged on time' : pick(['Acknowledged late', 'Not acknowledged'])) : pick(['Delivered', 'Read by all', 'Partly read']), ok: ok };
    });
    const reqs = Array.from({ length: 22 }, (_, i) => {
      const x = rnd(); const dirn = x < 0.55 ? 'Received' : x < 0.85 ? 'Sent' : 'Escalated';
      const ok = rnd() < pOn;
      return { ref: 'BR-' + (1000 + (seedOf(r.id) % 800) + i), date: isoAdd(todayISO, -Math.floor(rnd() * 30)), dir: dirn, subject: pick(['Approve purchase of laptops', 'Customer refund review', 'Access to finance reports', 'Change to branch rota', 'New vendor onboarding', 'Loan document check', 'Office supplies order', 'Fix branch signage', 'Data correction request', 'Budget reallocation']), cp: pick(cps),
        status: dirn === 'Sent' ? pick(['Resolved on time', 'In progress', 'Waiting on them']) : (ok ? 'Resolved on time' : pick(['Resolved late', 'Overdue', 'In progress'])), ok: ok };
    });
    let tks = Array.from({ length: 14 }, (_, i) => {
      const x = rnd(); const dirn = x < 0.55 ? 'Received' : x < 0.85 ? 'Sent' : 'Escalated';
      const ok = rnd() < pOn;
      return { id: 'TK-' + (500 + (seedOf(r.id + 't') % 400) + i), date: isoAdd(todayISO, -Math.floor(rnd() * 30)), dir: dirn, title: pick(['Printer not working', 'Cannot log in to email', 'Slow network', 'Request new monitor', 'Report not loading', 'Password reset', 'Update contact list', 'Phone line down']), cp: pick(cps),
        status: dirn === 'Sent' ? pick(['Solved on time', 'In progress', 'Assigned']) : (ok ? 'Solved on time' : pick(['Solved late', 'Overdue', 'In progress'])), ok: ok, real: null };
    });
    if (type === 'departments') {
      const real = tickets.filter((t) => t.dept === r.name || t.fromDept === r.name).map((t) => {
        const dirn = t.escalated && t.dept === r.name ? 'Escalated' : t.dept === r.name ? 'Received' : 'Sent';
        return { id: t.id, date: t.createdAt, dir: dirn, title: t.title, cp: dirn === 'Sent' ? t.dept : t.fromDept, status: t.slaLabel === 'Overdue' ? 'Overdue' : t.statusLabel, ok: t._c.sla !== 'Overdue' && t._c.sla !== 'Solved late', real: t };
      });
      tks = real.concat(tks);
    }
    const recv = (arr) => arr.filter((x) => x.dir !== 'Sent');
    const pct = (arr) => { const rr = recv(arr); return rr.length ? Math.round(rr.filter((x) => x.ok).length / rr.length * 100) : 100; };
    const all = recv(comms).concat(recv(reqs)).concat(recv(tks));
    const perf = all.length ? Math.round(all.filter((x) => x.ok).length / all.length * 100) : 100;
    return { comms, reqs, tks, perf, pc: pct(comms), pr: pct(reqs), pt: pct(tks) };
  };
  const targetOf = (key) => s.targets[key] || 95;
  const perfColor = (v, tg) => (v >= tg ? '#B8F36B' : v >= tg - 10 ? '#FFB86B' : '#FF9B8A');
  const openOu = (type, id) => () => this.setState({ section: 'orgunit', ou: type + ':' + id, ouTab: 'comms', ticketOpen: null, projectOpen: null });
  const orgRows = [{ depth: 0, type: 'Organization', name: '[YOUR ORGANIZATION]', sub: this.dir.length + ' people · ' + levels.map((t) => O[t].length + ' ' + typeName[t].toLowerCase() + (O[t].length === 1 ? '' : (t === 'branches' ? 'es' : 's'))).join(' · '), color: '#E8EFEA', canOpen: false }];
  const walk = (ptype, pid, depth) => {
    levels.forEach((t) => O[t].forEach((r) => {
      const p = parentFor(t, r);
      const match = ptype === null ? p === null : (p && p[0] === ptype && p[1] === pid);
      if (!match) return;
      orgRows.push({ depth: depth, type: typeName[t], name: r.name, sub: mgrRole[t] + ' · ' + membersOf(t, r.name).length + ' people', color: typeTint[t], canOpen: true, open: openOu(t, r.id) });
      walk(t, r.id, depth + 1);
    }));
  };
  if (has.structure) walk(null, null, 1);
  const org = {
    chain: ['Organization'].concat(levels.filter((t) => t !== 'departments').map((t) => typeName[t] + (t === 'branches' ? 'es' : 's'))).map((l, i, a) => ({ label: l, arrow: i < a.length - 1, style: 'padding:6px 12px;border-radius:16px;background:#16201C;border:1px solid #2A3832;font-size:13px' })),
    deptNote: has.departments, deptNoteText: 'Departments sit under ' + (has.branches ? 'their branch' : has.sectors ? 'their sector' : 'the organization') + '.',
    rows: orgRows.map((o, i) => ({ ...o, rowStyle: 'display:flex;flex-wrap:wrap;align-items:center;gap:12px;padding:12px 16px;' + (i ? 'border-top:1px solid #24322C' : 'background:#121B17'),
      indent: 'width:' + (o.depth * 26) + 'px;flex:none;align-self:stretch;' + (o.depth ? 'border-right:2px solid #2A3832' : ''), badge: badge(o.color) }))
  };
  const allUnits = levels.reduce((acc, t) => acc.concat(O[t].map((r) => ({ type: t, r: r }))), []);
  const unitPerf = {};
  allUnits.forEach((u) => { unitPerf[u.type + ':' + u.r.id] = unitData(u.type, u.r); });
  const myUnits = ['departments', 'branches', 'sectors', 'units'].filter((t) => has[t]).map((t) => {
    const r = recOf(t, O.mine[t]);
    const key = t + ':' + r.id;
    const on = s.section === 'orgunit' && s.ou === key;
    return { name: r.name, type: typeName[t], typeLower: typeName[t].toLowerCase(), short: shortOf(r.name), color: typeTint[t], key: key, perf: unitPerf[key].perf, open: openOu(t, r.id),
      style: 'display:flex;align-items:center;gap:10px;width:100%;min-height:46px;padding:4px 10px;border:0;border-radius:10px;cursor:pointer;color:#E8EFEA;font-family:inherit;background:' + (on ? '#22302A' : 'transparent') };
  });
  const roleNameOf = (id) => (roleDefs.find((r) => r.id === id) || {}).name || id;
  const targetLabel = (lvl, k) => (k === 'own' ? 'its own ' + typeName[lvl].toLowerCase() : k === 'wall' ? 'the Community wall' : 'other ' + (typeName[k] + (k === 'branches' ? 'es' : 's')).toLowerCase());
  const ruleSentence = (lvl, ty) => {
    const r = (s.commRules[lvl] || {})[ty];
    if (!r || !r.on) return ty + 's: switched off.';
    const to = r.to.filter((k) => k === 'own' || k === 'wall' || has[k]).map((k) => targetLabel(lvl, k));
    return ty + 's go to ' + (to.length ? to.join(', ') : 'nowhere yet') + '. Sent by ' + roleNameOf(r.sender) + '. ' + (r.approver === 'none' ? 'No approval.' : 'Approved by ' + roleNameOf(r.approver) + '.');
  };
  let ou = {}, ouTable = { id: 'none', cols: [], rows: [] };
  const ouTab = (s.ouTab === 'tks' && !has.tickets) || (s.ouTab === 'projects' && !has.projects) ? 'comms' : s.ouTab;
  const ouParts = s.ou ? s.ou.split(':') : null;
  const ouRec = ouParts && has[ouParts[0]] ? recOf(ouParts[0], ouParts[1]) : null;
  if (ouRec) {
    const t = ouParts[0], key = s.ou, D = unitPerf[key] || unitData(t, ouRec), tg = targetOf(key);
    const mem = membersOf(t, ouRec.name);
    const lead = mem.find((p) => /Lead|Supervisor|Manager/.test(p.title)) || mem[0];
    const dirC = { Received: '#7CC4FF', Sent: '#C9B6FF', Escalated: '#FF9B8A' };
    const stC = (st) => (/on time|Delivered|Read by all|Solved on time/.test(st) ? '#B8F36B' : /late|Partly|Waiting|In progress|Assigned/.test(st) ? '#FFB86B' : /Overdue|Not ack/.test(st) ? '#FF9B8A' : '#9DB0A6');
    const cnt = (arr, d) => arr.filter((x) => x.dir === d).length;
    const deptProj = { 'IT Operations': ['network'], 'Software Development': ['mobile'], 'HR': ['onboard'], 'Risk & Compliance': ['audit'], 'Finance': ['audit'], 'Operations': ['network'], 'Customer Care': ['mobile'] };
    const pids = t === 'departments' ? (deptProj[ouRec.name] || []) : s.projects.filter((p, i) => (seedOf(ouRec.id) + i) % 2 === 0).map((p) => p.id);
    const ps = projects.filter((p) => pids.indexOf(p.id) !== -1);
    const mk = (label, pc, detail) => ({ label: label, pct: pc, detail: detail, color: perfColor(pc, tg), bar: 'height:100%;width:' + pc + '%;background:' + perfColor(pc, tg) + ';border-radius:5px', marker: 'position:absolute;top:-3px;left:calc(' + tg + '% - 1px);width:2px;height:16px;background:#E8EFEA;border-radius:1px' });
    ou = {
      name: ouRec.name, type: typeName[t], short: shortOf(ouRec.name), color: typeTint[t], badge: badge(typeTint[t]), path: pathOf(t, ouRec),
      managerRole: mgrRole[t], manager: lead ? lead.name : 'Not set', members: mem.length,
      typePlural: typeName[t] + (t === 'branches' ? 'es' : 's'),
      commRules: ['Memo', 'Announcement'].map((ty) => ({ text: ruleSentence(t, ty) })),
      sendMemo: () => this.setState({ cmOpen: true, cmd: { type: has.memos ? 'Memo' : 'Announcement', title: '', body: '', wall: false, forum: '', sel: { [t + ':' + ouRec.name]: true } } }),
      kpis: [
        { label: 'Communications', parts: [{ n: cnt(D.comms, 'Received'), l: 'Received', color: '#7CC4FF' }, { n: cnt(D.comms, 'Sent'), l: 'Sent', color: '#C9B6FF' }] },
        { label: 'Business requests', parts: [{ n: cnt(D.reqs, 'Received'), l: 'Received', color: '#7CC4FF' }, { n: cnt(D.reqs, 'Sent'), l: 'Sent', color: '#C9B6FF' }, { n: cnt(D.reqs, 'Escalated'), l: 'Escalated', color: '#FF9B8A' }] },
        { label: 'Tickets', parts: [{ n: cnt(D.tks, 'Received'), l: 'Received', color: '#7CC4FF' }, { n: cnt(D.tks, 'Sent'), l: 'Sent', color: '#C9B6FF' }, { n: cnt(D.tks, 'Escalated'), l: 'Escalated', color: '#FF9B8A' }] },
        { label: 'Projects', parts: [{ n: ps.length, l: 'Active', color: '#E8EFEA' }, { n: ps.filter((p) => p.status !== 'On track').length, l: 'Need attention', color: '#FFB86B' }] }
      ],
      target: tg, setTarget: (e) => { const v = Math.max(50, Math.min(100, Number(e.target.value) || 95)); this.setState((x) => ({ targets: { ...x.targets, [key]: v } })); },
      perf: D.perf, perfColor: perfColor(D.perf, tg), perfNote: D.perf >= tg ? 'Target met' : (tg - D.perf) + ' points below target',
      breakdown: [mk('Communications acknowledged on time', D.pc, D.comms.filter((x) => x.dir !== 'Sent').length + ' received'), mk('Business requests resolved on time', D.pr, D.reqs.filter((x) => x.dir !== 'Sent').length + ' received'), mk('Tickets resolved on time', D.pt, D.tks.filter((x) => x.dir !== 'Sent').length + ' received')],
      tabs: [['comms', 'Communications', D.comms.length], ['reqs', 'Business requests', D.reqs.length], ['tks', 'Tickets', D.tks.length], ['projects', 'Projects', ps.length]].map(([k, l, n]) => ({ label: l, count: n, go: () => this.setState({ ouTab: k }),
        style: 'min-height:44px;padding:0 16px;border:0;background:transparent;cursor:pointer;font-size:14px;white-space:nowrap;font-family:inherit;border-bottom:2px solid ' + (ouTab === k ? '#B8F36B' : 'transparent') + ';color:' + (ouTab === k ? '#E8EFEA' : '#9DB0A6') + ';font-weight:' + (ouTab === k ? '600' : '500') }))
    };
    const ouOff = (l) => (/Ticket/.test(l) && !has.tickets) || (/Project/.test(l) && !has.projects);
    ou.kpis = ou.kpis.filter((k) => !ouOff(k.label)); ou.tabs = ou.tabs.filter((k) => !ouOff(k.label)); ou.breakdown = ou.breakdown.filter((k) => !ouOff(k.label));
    const dirCell = (d) => C.b(d, dirC[d]);
    if (ouTab === 'comms') ouTable = { id: 'ou-c-' + key, pageSize: 10, minWidth: '900px', cols: [{ label: 'Date', filter: 'date' }, { label: 'Direction', filter: 'select' }, { label: 'Type', filter: 'select' }, { label: 'Subject', filter: 'text' }, { label: 'From / to', filter: 'text' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
      rows: D.comms.slice().sort((a, b) => b.date.localeCompare(a.date)).map((c) => ({ cells: [C.t(fmtDay(c.date), { fv: c.date, nowrap: true }), dirCell(c.dir), C.t(c.type), C.t(c.subject, { strong: true }), C.t(c.cp), C.b(c.status, stC(c.status)), C.a([{ label: 'Open', run: () => this.notify('Opening "' + c.subject + '".', '#7CC4FF') }])] })) };
    if (ouTab === 'reqs') ouTable = { id: 'ou-r-' + key, pageSize: 10, minWidth: '940px', cols: [{ label: 'Ref', filter: 'text' }, { label: 'Date', filter: 'date' }, { label: 'Direction', filter: 'select' }, { label: 'Request', filter: 'text' }, { label: 'With', filter: 'select' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
      rows: D.reqs.slice().sort((a, b) => b.date.localeCompare(a.date)).map((c) => ({ tint: c.status === 'Overdue' ? '#FF9B8A' : '', cells: [C.t(c.ref, { mono: true }), C.t(fmtDay(c.date), { fv: c.date, nowrap: true }), dirCell(c.dir), C.t(c.subject, { strong: true }), C.t(c.cp), C.b(c.status, stC(c.status)),
        C.a([{ label: 'View', run: () => this.notify('Opening ' + c.ref + '.', '#7CC4FF') }].concat(c.dir === 'Received' && c.status !== 'Resolved on time' && c.status !== 'Resolved late' ? [{ label: 'Escalate', kind: 'danger', run: () => this.notify(c.ref + ' escalated to the ' + (has.branches ? 'branch manager' : 'head of department') + '.', '#FF9B8A') }] : []))] })) };
    if (ouTab === 'tks') ouTable = { id: 'ou-t-' + key, pageSize: 10, minWidth: '900px', cols: [{ label: 'Ticket', filter: 'text' }, { label: 'Date', filter: 'date' }, { label: 'Direction', filter: 'select' }, { label: 'Title', filter: 'text' }, { label: 'With', filter: 'select' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
      rows: D.tks.slice().sort((a, b) => b.date.localeCompare(a.date)).map((c) => ({ tint: c.status === 'Overdue' ? '#FF9B8A' : '', cells: [C.t(c.id, { mono: true }), C.t(fmtDay(c.date), { fv: c.date, nowrap: true }), dirCell(c.dir), C.t(c.title, { strong: true }), C.t(c.cp), C.b(c.status, stC(c.status)),
        C.a([{ label: 'View', run: c.real ? () => this.setState({ section: 'tickets', ticketOpen: c.id }) : () => this.notify('Opening ' + c.id + '.', '#7CC4FF') }])] })) };
    if (ouTab === 'projects') ouTable = { id: 'ou-p-' + key, pageSize: 10, minWidth: '760px', emptyText: 'No projects for this ' + typeName[t].toLowerCase() + ' yet.', cols: [{ label: 'Project', filter: 'text' }, { label: 'Status', filter: 'select' }, { label: 'Progress' }, { label: 'Due', filter: 'date' }, { label: '', align: 'right' }],
      rows: ps.map((p) => ({ cells: [C.t(p.name, { strong: true, sub: 'Led by ' + p.owner }), C.b(p.status, p.statusColor), C.bar(p.progress, p.statusColor), C.t(p.dueLabel, { fv: p.due, nowrap: true }), C.a([{ label: 'View', kind: 'primary', run: p.open }])] })) };
  }
  const orgWidgetItems = {};
  ['departments', 'units'].forEach((t) => { if (!has[t]) return; orgWidgetItems[t] = O[t].map((r) => { const k = t + ':' + r.id; const pf = unitPerf[k].perf; const pp = parentFor(t, r); return { name: r.name, sub: (pp ? recOf(pp[0], pp[1]).name + ' · ' : '') + membersOf(t, r.name).length + ' people', perf: pf, perfColor: perfColor(pf, targetOf(k)), open: openOu(t, r.id) }; }); });
  if (has.branches) orgWidgetItems.branches = O.branches.map((b) => ({ name: b.name, initials: shortOf(b.name), place: has.sectors ? recOf('sectors', b.sector).name + ' sector' : 'Branch', head: (membersOf('branches', b.name).find((p) => /Supervisor|Lead|Manager/.test(p.title)) || {}).name || 'Not set', staff: membersOf('branches', b.name).length, statusColor: '#B8F36B', statusLabel: unitPerf['branches:' + b.id].perf + '% on time', open: openOu('branches', b.id) }));
  if (has.sectors) orgWidgetItems.sectors = O.sectors.map((sc, i) => ({ name: sc.name, color: ['#FFB86B', '#7CC4FF', '#F49AC1'][i % 3], depts: (has.branches ? O.branches.filter((b) => b.sector === sc.id) : O.departments.filter((d) => d.sector === sc.id)).map((x) => x.name).join(' · '), staff: membersOf('sectors', sc.name).length, open: openOu('sectors', sc.id) }));
  orgWidgetItems.performance = myUnits.map((u) => { const tg = targetOf(u.key); return { name: u.name, type: u.type, perf: u.perf, color: perfColor(u.perf, tg), open: u.open, bar: 'height:100%;width:' + u.perf + '%;background:' + perfColor(u.perf, tg) + ';border-radius:4px', marker: 'position:absolute;top:-3px;left:calc(' + tg + '% - 1px);width:2px;height:14px;background:#E8EFEA' }; });
  return { O, esc, myUnits, org, orgWidgetItems, ou, ouTable, recOf, roleNameOf, ruleSentence, typeName };
}
