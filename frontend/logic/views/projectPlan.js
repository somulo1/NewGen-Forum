import { decorate, expandMacros, pageHtml, pageText, lineDiff } from '~/logic/pageMarkup.js';
import { boardDataUrl, boardSvg } from '~/logic/boardSvg.js';
import { parseExport } from '~/logic/confluenceImport.js';
import { CAT_COLORS, addDaysIso, criticalPath, doneAt, isoMs, openDeps, statusEntry, wfOf } from '~/logic/projects.js';

// The open project's plan views: backlog or work breakdown, timeline, reports, documents and settings. Edits go into state.projects, and the
// projects widget writes them to the server like every other project change.
const TEMPLATES = {
  blank: { name: 'Blank page', title: '', body: '' },
  board: { name: 'Whiteboard', title: 'Whiteboard', body: '{}', kind: 'board' },
  diagram: { name: 'Diagram', title: 'Diagram', body: '{}', kind: 'board' },
  requirements: { name: 'Requirements', title: 'Requirements', body: ['<h2>Background</h2><p></p>','<h2>Goals</h2><p></p>','<h2>What is in scope</h2><p></p>','<h2>What is out of scope</h2><p></p>','<h2>Acceptance criteria</h2><p></p>'].join('') },
  notes: { name: 'Meeting notes', title: 'Meeting notes', body: ['<h2>Date and attendees</h2><p></p>','<h2>Discussion</h2><p></p>','<h2>Decisions</h2><p></p>','<h2>Action items</h2><p></p>'].join('') },
  decision: { name: 'Decision record', title: 'Decision: ', body: ['<h2>Context</h2><p></p>','<h2>Options considered</h2><p></p>','<h2>Decision</h2><p></p>','<h2>Consequences</h2><p></p>'].join('') },
  runbook: { name: 'Runbook', title: 'Runbook: ', body: ['<h2>Purpose</h2><p></p>','<h2>Before you start</h2><p></p>','<h2>Steps</h2><p></p>','<h2>If something goes wrong</h2><p></p>','<h2>Who to contact</h2><p></p>'].join('') },
  release: { name: 'Release notes', title: 'Release notes', body: ['<h2>Version and date</h2><p></p>','<h2>What is new</h2><p></p>','<h2>Fixes</h2><p></p>','<h2>Known issues</h2><p></p>'].join('') }
};
const DAY = 86400000;

export function buildPlan(ctx) {
  const { s, can, me, today, cur_p: p, tools, shortDate, longDate, projectsW } = ctx;
  const pid = p.id;
  const waterfall = p.method === 'waterfall';
  const tab = s.planTab || 'overview';
  const isManager = p.owner === me;
  const may = (x) => can('projects', x);
  const canSettings = may('Manage workflow');
  const tasksOf = (fn) => p.tasks.map((t, ti) => ({ t, ti })).filter(({ t }) => fn(t));
  const patch = (fn) => this.setState((st) => ({ projects: st.projects.map((x) => (x.id === pid ? fn(x) : x)) }));
  const pts = (list) => list.reduce((a, t) => a + (Number(t.points) || 0), 0);
  const usesPoints = p.tasks.some((t) => Number(t.points) > 0);
  const size = (t) => (usesPoints ? Number(t.points) || 0 : 1);
  const unit = usesPoints ? 'points' : 'tasks';
  const row = ({ t, ti }) => {
    const st = statusEntry(p, t);
    return { id: t.id || 't' + ti, key: t.key || '', title: t.title, kind: t.kind || 'task', points: t.points === null || t.points === undefined ? '' : t.points, status: st.name, statusColor: CAT_COLORS[st.cat], due: t.due ? shortDate(t.due) : '', owner: t.primary || 'Nobody', blocked: t.col !== 3 && openDeps(p, t).length > 0, open: () => tools.openTask(ti), sprint: t.sprint || '',
      setSprint: (e) => { const v = e.target.value; const sp = p.sprints.find((x) => x.id === v); patch((x) => ({ ...x, tasks: x.tasks.map((y, j) => (j === ti ? { ...y, sprint: v, due: sp ? (y.due < sp.start ? sp.start : y.due > sp.end ? sp.end : y.due) : y.due } : y)) })); } };
  };

  // ---- backlog / work breakdown ----
  const epics = tasksOf((t) => t.kind === 'epic').map(({ t, ti }) => {
    const under = p.tasks.filter((x) => x.parent === t.id || p.tasks.some((y) => y.parent === t.id && x.parent === y.id));
    const done = under.filter((x) => x.col === 3);
    const pct = under.length ? Math.round((pts(done) || done.length) / (pts(under) || under.length) * 100) : 0;
    return { id: t.id || ti, key: t.key || '', title: t.title, count: under.length, done: done.length, points: pts(under), pct, bar: 'height:100%;width:' + Math.max(2, pct) + '%;background:#7CC4FF;border-radius:3px', open: () => tools.openTask(ti) };
  });
  const rankOf = (t, i) => (t.rank === undefined || t.rank === null ? i * 1000 : t.rank);
  const ordered0 = tasksOf((t) => !t.sprint && !t.milestone && t.col !== 3 && t.kind !== 'epic').sort((a, b) => rankOf(a.t, a.ti) - rankOf(b.t, b.ti));
  // the order is the priority: moving an item renumbers the list
  const moveRank = (from, to) => {
    if (from === to || to < 0 || to >= ordered0.length) return;
    const l = ordered0.slice();
    const [x] = l.splice(from, 1);
    l.splice(to, 0, x);
    const ranks = new Map(l.map((e, n) => [e.ti, n * 1000]));
    patch((pp) => ({ ...pp, tasks: pp.tasks.map((y, j) => (ranks.has(j) && y.rank !== ranks.get(j) ? { ...y, rank: ranks.get(j) } : y)) }));
  };
  const backlog = ordered0.map((e, n) => ({ ...row(e), n, first: n === 0, last: n === ordered0.length - 1, up: () => moveRank(n, n - 1), down: () => moveRank(n, n + 1), dropOn: (from) => moveRank(from, n) }));
  const phases = tools.tl.map((m) => ({ id: m.id, name: m.name, due: shortDate(m.iso), done: m.done, total: m.total, status: m.status, rows: tasksOf((t) => t.milestone === m.id || p.sprints.some((sp) => sp.id === t.sprint && sp.milestone === m.id)).map(row) }));

  // ---- timeline ----
  const dated = p.tasks.filter((t) => t.due);
  const allDates = [p.start, p.due].concat(dated.map((t) => t.start || t.due), dated.map((t) => t.due), p.milestones.map((m) => m.iso)).filter(Boolean).sort();
  const from = allDates[0] || today, to = allDates[allDates.length - 1] || today;
  const days = Math.max(1, Math.round((isoMs(to) - isoMs(from)) / DAY) + 1);
  const LABEL = 210, ROW = 30, TOP = 26;
  const dayW = Math.max(10, Math.min(28, 760 / days));
  const xOf = (iso) => LABEL + ((isoMs(iso) - isoMs(from)) / DAY) * dayW;
  const rows = [];
  const yOfTask = new Map();
  const addTask = ({ t, ti }) => {
    const start = t.start && t.start <= t.due ? t.start : t.due;
    const y = TOP + rows.length * ROW;
    yOfTask.set(t.id, { y, x1: xOf(start), x2: xOf(t.due) + dayW });
    const st = statusEntry(p, t);
    rows.push({ type: 'task', key: t.id || 'x' + ti, y, label: t.title.length > 28 ? t.title.slice(0, 27) + '…' : t.title, x: xOf(start), w: Math.max(dayW, xOf(t.due) + dayW - xOf(start)), color: t.col !== 3 && t.due < today ? '#FF9B8A' : CAT_COLORS[st.cat],
      critical: critical.has(t.id), baseX: t.baseDue ? xOf(t.baseStart || t.baseDue) : null, baseW: t.baseDue ? Math.max(dayW, xOf(t.baseDue) + dayW - xOf(t.baseStart || t.baseDue)) : 0, canDrag: may('Reschedule tasks'), shift: (px) => shiftTask(ti, Math.round(px / dayW)),
      tip: (t.key ? t.key + ' · ' : '') + t.title + ' · ' + st.name + ' · ' + shortDate(start) + ' – ' + shortDate(t.due) + (t.primary ? ' · ' + t.primary : ''), open: () => tools.openTask(ti) });
  };
  const critical = criticalPath(p);
  // moving a bar moves its dates; the work that depends on it follows when it would now overlap
  const shiftTask = (ti, days) => {
    if (!days) return;
    patch((pp) => {
      const tasks = pp.tasks.map((x) => ({ ...x }));
      const move = (t, d) => { if (t.start) t.start = addDaysIso(t.start, d); t.due = addDaysIso(t.due, d); };
      move(tasks[ti], days);
      const walk = (t, n) => { if (n > 50) return; tasks.filter((x) => (x.deps || []).includes(t.id) && x.col !== 3).forEach((x) => { const st = x.start || x.due; if (st <= t.due) { const gap = Math.round((isoMs(t.due) - isoMs(st)) / DAY) + 1; move(x, gap); walk(x, n + 1); } }); };
      if (tasks[ti].id) walk(tasks[ti], 0);
      return { ...pp, tasks };
    });
  };
  const seen = new Set();
  tools.tl.forEach((m) => {
    rows.push({ type: 'phase', key: 'm' + m.id, y: TOP + rows.length * ROW, label: m.name.length > 28 ? m.name.slice(0, 27) + '…' : m.name, x: xOf(m.iso), tip: m.name + ' · ' + (waterfall ? 'gate ' : 'due ') + shortDate(m.iso) + ' · ' + m.done + ' of ' + m.total + ' done' });
    tasksOf((t) => t.due && (t.milestone === m.id || p.sprints.some((sp) => sp.id === t.sprint && sp.milestone === m.id))).forEach((x) => { seen.add(x.t); addTask(x); });
  });
  tasksOf((t) => t.due && !seen.has(t)).forEach((x) => addTask(x));
  const links = [];
  p.tasks.forEach((t) => (t.deps || []).forEach((d) => {
    const a = yOfTask.get(d), b = yOfTask.get(t.id);
    if (!a || !b) return;
    const mid = Math.max(a.x2 + 6, b.x1 - 6);
    links.push({ key: d + t.id, d: 'M' + a.x2 + ' ' + (a.y + 15) + ' H' + mid + ' V' + (b.y + 15) + ' H' + b.x1 });
  }));
  const ticks = [];
  for (let i = 0; i < days; i += Math.max(1, Math.round(70 / dayW))) { const iso = addDaysIso(from, i); ticks.push({ x: xOf(iso), label: shortDate(iso) }); }
  const timeline = { hasCritical: critical.size > 0, hasBaseline: p.tasks.some((t) => t.baseDue), canBaseline: may('Save baseline'), saveBaseline: () => { patch((pp) => ({ ...pp, tasks: pp.tasks.map((t) => ({ ...t, baseStart: t.start || t.due, baseDue: t.due })) })); this.notify('Baseline saved. Later date changes show against it.', '#B8F36B'); }, rows, links, ticks, width: Math.ceil(LABEL + days * dayW + 20), height: TOP + rows.length * ROW + 8, todayX: today >= from && today <= to ? xOf(today) : null, label: LABEL, empty: rows.length === 0 };

  // ---- reports ----
  const sel = s.planRange || 'current';
  const curSprint = tools.spl.find((x) => x.status === 'current');
  const rangeId = sel === 'current' ? (curSprint ? curSprint.id : 'project') : sel;
  const sprint = tools.spl.find((x) => x.id === rangeId);
  const r0 = sprint ? sprint.start : p.start, r1 = sprint ? sprint.end : p.due;
  const scope = sprint ? p.tasks.filter((t) => t.sprint === sprint.id) : p.tasks;
  const total = scope.reduce((a, t) => a + size(t), 0);
  const rDays = r0 && r1 ? Math.max(1, Math.round((isoMs(r1) - isoMs(r0)) / DAY)) : 1;
  const W = 640, H = 220, L = 44, B = 26, T = 12, R = 12;
  const cx = (i) => L + (i / rDays) * (W - L - R);
  const cy = (v) => T + (1 - v / Math.max(1, total)) * (H - T - B);
  const points = [];
  for (let i = 0; i <= rDays; i++) {
    const iso = addDaysIso(r0, i);
    if (iso > today) break;
    const end = isoMs(iso) + DAY;
    const left = total - scope.reduce((a, t) => { const at = doneAt(t); return a + (at && new Date(at).getTime() < end ? size(t) : 0); }, 0);
    points.push({ x: cx(i), y: cy(left), tip: shortDate(iso) + ': ' + left + ' ' + unit + ' left', date: shortDate(iso), left });
  }
  const velocitySprints = tools.spl.filter((x) => x.closed || x.status === 'current');
  const velDone = velocitySprints.map((x) => { const list = p.tasks.filter((t) => t.sprint === x.id && t.col === 3); return { id: x.id, name: x.name, closed: !!x.closed, value: x.closed && usesPoints ? Number(x.completed) || 0 : x.closed ? Number(x.doneCount) || 0 : list.reduce((a, t) => a + size(t), 0) }; });
  const finished = velDone.filter((x) => x.closed);
  const maxV = Math.max(1, ...velDone.map((x) => x.value));
  const barW = Math.min(44, (W - L - R) / Math.max(1, velDone.length) - 8);
  const reports = {
    unit, rangeValue: sel, ranges: [{ v: 'current', l: 'Current sprint' }, { v: 'project', l: 'Whole project' }].concat(tools.spl.map((x) => ({ v: x.id, l: x.name }))),
    onRange: (e) => this.setState({ planRange: e.target.value }), title: sprint ? sprint.name : 'Whole project', W, H, L, B, T, R,
    empty: !total || !r0, line: points.map((q, i) => (i ? 'L' : 'M') + q.x + ' ' + q.y).join(' '), points, idealX1: cx(0), idealY1: cy(total), idealX2: cx(rDays), idealY2: cy(0),
    last: points[points.length - 1] || null, total, startLabel: r0 ? shortDate(r0) : '', endLabel: r1 ? shortDate(r1) : '',
    showVelocity: !waterfall, velEmpty: !velDone.length,
    bars: velDone.map((x, i) => { const h = (x.value / maxV) * (H - T - B); return { id: x.id, name: x.name, value: x.value, x: L + i * ((W - L - R) / velDone.length) + 4, y: H - B - h, w: barW, h: Math.max(1, h) }; }),
    average: finished.length ? Math.round(finished.reduce((a, x) => a + x.value, 0) / finished.length) : 0,
    stats: [
      { label: 'Finished', value: p.tasks.filter((t) => t.col === 3).reduce((a, t) => a + size(t), 0) + ' of ' + p.tasks.reduce((a, t) => a + size(t), 0) + ' ' + unit },
      { label: 'Blocked now', value: p.tasks.filter((t) => t.col !== 3 && openDeps(p, t).length).length + ' tasks' },
      { label: 'Overdue', value: p.tasks.filter((t) => t.col !== 3 && t.due && t.due < today).length + ' tasks' },
      { label: 'In the backlog', value: backlog.length + ' tasks' }
    ]
  };

  // ---- documents ----
  const query = (s.pageQuery || '').trim().toLowerCase();
  const labelOn = s.pageLabel || '';
  const draft = s.pageDraft && s.pageDraft.pid === pid ? s.pageDraft : null;
  const openId = s.pageOpen && s.pageOpen.pid === pid ? s.pageOpen.id : null;
  const pages = p.pages || [];
  const cur = pages.find((x) => x.id === openId) || null;
  const depth = (x, n = 0) => { const up = pages.find((y) => y.id === x.parent); return up && n < 5 ? depth(up, n + 1) : n; };
  const ordered = [];
  const place = (parent) => pages.filter((x) => (x.parent || '') === parent).sort((a, b) => (a.title < b.title ? -1 : 1)).forEach((x) => { ordered.push(x); if (x.id) place(x.id); });
  place('');
  pages.forEach((x) => { if (!ordered.includes(x)) ordered.push(x); });
  const setDraft = (k) => (e) => { const v = e.target.value; this.setState((st) => ({ pageDraft: { ...st.pageDraft, [k]: v, err: '' } })); };
  const fmt = (at) => new Date(at).toLocaleString('en-GB', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });
  const openPage = (x) => this.setState({ pageOpen: { pid, id: x.id }, pageDraft: null, pageDiff: null });
  const updPage = (id, fn) => patch((x) => ({ ...x, pages: x.pages.map((y) => (y.id === id ? fn(y) : y)) }));
  const bodyChanges = cur ? (cur.changes || []).filter((c) => c.field === 'body').slice().reverse() : [];
  const versions = bodyChanges.map((c, i) => ({ key: c.at, who: c.who, when: fmt(c.at),
    compare: () => this.setState({ pageDiff: { id: cur.id, at: c.at } }),
    restore: () => this.setState({ pageDraft: { pid, id: cur.id, kind: cur.kind || '', seed: cur.body || '', title: cur.title, parent: cur.parent || '', body: c.old || '', labels: (cur.labels || []).join(', '), err: '' }, pageDiff: null }) }));
  const shownDiff = cur && s.pageDiff && s.pageDiff.id === cur.id ? bodyChanges.find((c) => c.at === s.pageDiff.at) : null;
  const allLabels = [...new Set(pages.flatMap((x) => x.labels || []))].sort();
  const mePerson = this.srv ? this.srv.me.person : me;
  const comments = cur ? (p.comments || []).filter((c) => c.page === cur.id) : [];
  const recent = pages.filter((x) => x.updatedAt).slice().sort((a, b) => (a.updatedAt < b.updatedAt ? 1 : -1)).slice(0, 5).map((x) => ({ id: x.id, title: x.title, when: fmt(x.updatedAt), open: () => openPage(x) }));
  const download = (name, text, type) => { const a = document.createElement('a'); a.href = URL.createObjectURL(new Blob([text], { type })); a.download = name; a.click(); setTimeout(() => URL.revokeObjectURL(a.href), 2000); };
  const safeName = (t) => t.replace(/[^\w\- ]+/g, '').trim() || 'page';
  const docs = {
    query: s.pageQuery || '', onQuery: (e) => this.setState({ pageQuery: e.target.value }), recent,
    labels: allLabels.map((l) => ({ name: l, on: l === labelOn, pick: () => this.setState({ pageLabel: l === labelOn ? '' : l }) })),
    list: ordered.filter((x) => (!query || (x.title + ' ' + pageText(x.body) + ' ' + (x.labels || []).join(' ')).toLowerCase().includes(query)) && (!labelOn || (x.labels || []).includes(labelOn))).map((x) => ({ id: x.id, title: x.title, indent: query ? 0 : depth(x) * 14, on: x.id === openId, open: () => openPage(x) })),
    templates: Object.entries(TEMPLATES).map(([k, t]) => ({ key: k, name: t.name, pick: () => this.setState({ pageDraft: { pid, id: '', title: t.title, body: t.body, kind: t.kind || '', parent: cur ? cur.id : '', labels: '', err: '' }, pageDiff: null }) })),
    page: cur && !draft ? { id: cur.id, title: cur.title, isBoard: cur.kind === 'board', boardHtml: cur.kind === 'board' ? boardSvg(cur.body) : '',
      body: cur.kind === 'board' ? '' : decorate(expandMacros(pageHtml(cur.body || ''), { children: pages.filter((x) => x.parent === cur.id).map((x) => x.title), tasks: p.tasks.map((t) => ({ ...t, statusName: statusEntry(p, t).name })), boardUrl: (title) => { const x = pages.find((y) => y.kind === 'board' && y.title.toLowerCase() === String(title || '').toLowerCase()); return x ? boardDataUrl(x.body) : ''; } }), p.tasks.map((t) => t.key).filter(Boolean), comments),
      canRestrict: may('Restrict pages') && (p.members || []).length > 0,
      restrictLists: ['view', 'edit'].map((k) => ({ key: k, label: k === 'view' ? 'Who can view (nobody ticked: everyone on the team)' : 'Who can edit (nobody ticked: everyone who can edit pages)', people: (p.members || []).map((m) => ({ person: m.person, name: m.name, on: ((cur.restrict || {})[k] || []).includes(m.person), toggle: () => updPage(cur.id, (y) => { const r0 = y.restrict || {}; const l = r0[k] || []; return { ...y, restrict: { ...r0, [k]: l.includes(m.person) ? l.filter((v) => v !== m.person) : l.concat([m.person]) } }; }) })) })),
      restricted: ((cur.restrict || {}).view || []).length + ((cur.restrict || {}).edit || []).length > 0,
      exportPdf: () => {
        const clean = new DOMParser().parseFromString(cur.kind === 'board' ? boardSvg(cur.body) : expandMacros(pageHtml(cur.body || ''), { children: pages.filter((x) => x.parent === cur.id).map((x) => x.title), tasks: [], boardUrl: (title) => { const x = pages.find((y) => y.kind === 'board' && y.title.toLowerCase() === String(title || '').toLowerCase()); return x ? boardDataUrl(x.body) : ''; } }), 'text/html');
        clean.querySelectorAll('script,iframe,object,embed,link,meta,base').forEach((n) => n.remove());
        clean.querySelectorAll('*').forEach((n) => [...n.attributes].forEach((a) => { if (/^on/i.test(a.name)) n.removeAttribute(a.name); }));
        const f = document.createElement('iframe');
        f.setAttribute('sandbox', 'allow-same-origin allow-modals');
        f.style.cssText = 'position:fixed;width:0;height:0;border:0';
        f.srcdoc = '<!doctype html><meta charset="utf-8"><title>' + cur.title.replace(/</g, '&lt;') + '</title><style>body{font:14px/1.5 sans-serif;color:#111}img,svg{max-width:100%}table{border-collapse:collapse}td,th{border:1px solid #999;padding:4px 8px}[data-panel]{border-left:4px solid #1F6FB5;padding:6px 10px;margin:8px 0}[data-status]{font-weight:700}</style><h1>' + cur.title.replace(/</g, '&lt;') + '</h1>' + clean.body.innerHTML;
        f.onload = () => { f.contentWindow.focus(); f.contentWindow.print(); setTimeout(() => f.remove(), 60000); };
        document.body.appendChild(f);
      },
      followRef: (m) => { const [kind, ...rest] = m.split(':'); const v = rest.join(':'); if (kind === 'page') { const x = pages.find((y) => y.title.toLowerCase() === v.toLowerCase()); if (x) openPage(x); else this.notify('No page is called "' + v + '" yet.', '#FFB86B'); return; } const ti = p.tasks.findIndex((t) => t.key === v); if (ti !== -1) tools.openTask(ti); },
      commentQuote: s.pageQuote || '', setQuote: (q) => this.setState({ pageQuote: q }), clearQuote: () => this.setState({ pageQuote: '' }),
      updated: cur.updatedAt ? fmt(cur.updatedAt) : '', versions, labels: cur.labels || [], locked: !!cur.locked,
      canLock: may('Lock pages'), canDelete: may('Delete pages'), mayEdit: may('Edit pages') && (!cur.locked || may('Lock pages') || isManager || cur.author === me),
      backlinks: pages.filter((x) => x.id !== cur.id && pageText(x.body).toLowerCase().includes('[[' + cur.title.toLowerCase() + ']]')).map((x) => ({ id: x.id, title: x.title, open: () => openPage(x) })),
      crumbs: (() => { const l = []; let u = pages.find((y) => y.id === cur.parent); for (let n = 0; u && n < 6; n++) { const x = u; l.unshift({ id: x.id, title: x.title, open: () => openPage(x) }); u = pages.find((y) => y.id === x.parent); } return l; })(),
      children: pages.filter((x) => x.parent === cur.id).map((x) => ({ id: x.id, title: x.title, open: () => openPage(x) })),
      watching: (cur.watchers || []).includes(mePerson), toggleWatch: () => updPage(cur.id, (y) => ({ ...y, watchers: (y.watchers || []).includes(mePerson) ? y.watchers.filter((w) => w !== mePerson) : (y.watchers || []).concat([mePerson]) })),
      toggleLock: () => updPage(cur.id, (y) => ({ ...y, locked: !y.locked })),
      diff: shownDiff ? lineDiff(pageText(shownDiff.old), pageText(shownDiff.new)) : null, closeDiff: () => this.setState({ pageDiff: null }),
      files: (cur.files || []).map((f, i) => ({ key: i, name: f.name, open: () => (this.srv ? projectsW.openAttachment(this.srv, f.ref).catch((e) => this.notify(e.message || 'Could not open it.', '#FF9B8A')) : null), remove: () => updPage(cur.id, (y) => ({ ...y, files: y.files.filter((_, j) => j !== i) })) })),
      onFile: (e) => { const list = Array.from(e.target.files || []); e.target.value = ''; if (!this.srv) return; const id = cur.id; Promise.all(list.map((f) => projectsW.attach(this.srv, pid, f))).then((added) => { updPage(id, (y) => ({ ...y, files: (y.files || []).concat(added) })); this.notify(added.length + ' file' + (added.length === 1 ? '' : 's') + ' attached.', '#B8F36B'); }).catch((x) => this.notify(x.message || 'Could not attach the file.', '#FF9B8A')); },
      comments: comments.map((c) => ({ quote: c.quote || '', resolved: !!c.resolved, toggleResolved: () => patch((x) => ({ ...x, comments: x.comments.map((y) => (y.id === c.id ? { ...y, resolved: !y.resolved } : y)) })), id: c.id, who: c.who, text: c.text, when: c.createdAt ? fmt(c.createdAt) : 'Just now', mine: c.who === me, remove: () => patch((x) => ({ ...x, comments: x.comments.filter((y) => y.id !== c.id) })) })),
      comment: s.pageComment || '', onComment: (e) => this.setState({ pageComment: e.target.value }),
      addComment: () => { const text = (this.state.pageComment || '').trim(); if (!text) return; const quote = this.state.pageQuote || ''; this.setState({ pageComment: '', pageQuote: '' }); patch((x) => ({ ...x, comments: (x.comments || []).concat([{ page: cur.id, who: me, text, quote, resolved: false }]) })); },
      exportDoc: () => download(safeName(cur.title) + '.doc', '<html><head><meta charset="utf-8"></head><body><h1>' + cur.title.replace(/</g, '&lt;') + '</h1>' + pageHtml(cur.body || '') + '</body></html>', 'application/msword'),
      exportHtml: () => download(safeName(cur.title) + '.html', '<!doctype html><meta charset="utf-8"><title>' + cur.title.replace(/</g, '&lt;') + '</title><h1>' + cur.title.replace(/</g, '&lt;') + '</h1>' + pageHtml(cur.body || ''), 'text/html'),
      copy: () => this.setState({ pageDraft: { pid, id: '', kind: cur.kind || '', title: 'Copy of ' + cur.title, body: cur.body || '', parent: cur.parent || '', labels: (cur.labels || []).join(', '), err: '' } }),
      edit: () => this.setState({ pageDraft: { pid, id: cur.id, kind: cur.kind || '', seed: cur.body || '', title: cur.title, body: cur.body || '', parent: cur.parent || '', labels: (cur.labels || []).join(', '), err: '' }, pageDiff: null }),
      remove: () => { if (pages.some((x) => x.parent === cur.id)) { this.notify('Move or delete the pages under this one first.', '#FFB86B'); return; } this.setState({ pageOpen: null }); patch((x) => ({ ...x, pages: x.pages.filter((y) => y.id !== cur.id), comments: (x.comments || []).filter((y) => y.page !== cur.id) })); } } : null,
    canImport: may('Import pages') && !!this.srv, importing: s.pageImport || { busy: false, text: '', problems: [] },
    importFile: (e) => {
      const f = e.target.files && e.target.files[0];
      e.target.value = '';
      if (!f) return;
      const say = (text, problems = []) => this.setState({ pageImport: { busy: true, text, problems } });
      say('Reading the export…');
      parseExport(f, (n, of) => say('Reading page ' + n + ' of ' + of + '…')).then(async ({ pages: found, problems }) => {
        const taken = new Set(pages.map((x) => x.title.trim().toLowerCase()));
        const failed = await projectsW.importPages(this, this.srv, pid, found, taken, (n) => say('Saving page ' + n + ' of ' + found.length + '…'));
        this.setState({ pageImport: { busy: false, text: (found.length - failed.filter((x) => !x.includes(': attachment ')).length) + ' of ' + found.length + ' pages imported.', problems: problems.concat(failed) } });
      }).catch((x) => this.setState({ pageImport: { busy: false, text: x.message || 'Could not read that export.', problems: [] } }));
    },
    canCreate: may('Create pages'), canAttach: may('Attach to pages'), canComment: may('Comment on pages'), canExport: may('Export pages'),
    draft: draft ? { ...draft, isNew: !draft.id, onTitle: setDraft('title'), onParent: setDraft('parent'), onLabels: setDraft('labels'),
      liveOn: !!draft.id && !draft.kind && !!this.srv, joinLive: (onPeers) => projectsW.joinLive(this.srv, pid, draft.id, onPeers), isBoard: draft.kind === 'board', setBody: (v) => this.setState((st) => ({ pageDraft: { ...st.pageDraft, body: v, err: '' } })),
      parentOpts: pages.filter((x) => x.id && x.id !== draft.id).map((x) => ({ v: x.id, l: x.title })),
      cancel: () => this.setState({ pageDraft: null }),
      save: () => {
        const title = draft.title.trim();
        if (!title) { this.setState({ pageDraft: { ...draft, err: 'Give the page a title.' } }); return; }
        if (pages.some((x) => x.id !== draft.id && x.title.trim().toLowerCase() === title.toLowerCase())) { this.setState({ pageDraft: { ...draft, err: 'A page in this project already has that title.' } }); return; }
        const data = { title, body: draft.body, kind: draft.kind || '', liveSave: !!draft.id && !!this.srv, parent: draft.parent || '', labels: [...new Set((draft.labels || '').split(',').map((x) => x.trim()).filter(Boolean))] };
        this.setState({ pageDraft: null });
        patch((x) => ({ ...x, pages: draft.id ? x.pages.map((y) => (y.id === draft.id ? { ...y, ...data } : y)) : (x.pages || []).concat([{ ...data, author: me, watchers: [mePerson], files: [] }]) }));
        this.notify('Page saved.', '#B8F36B');
      } } : null
  };

  // ---- settings: project type and workflow ----
  const wd = s.wfDraft && s.wfDraft.pid === pid ? s.wfDraft : null;
  const list = wd ? wd.list : wfOf(p);
  const change = (fn) => this.setState((st) => { const d = st.wfDraft && st.wfDraft.pid === pid ? st.wfDraft : { pid, list: wfOf(p) }; return { wfDraft: { ...d, list: fn(d.list), err: '' } }; });
  const settings = {
    canEdit: canSettings, method: p.method,
    methods: [{ v: 'agile', l: 'Agile (sprints and a backlog)' }, { v: 'waterfall', l: 'Waterfall (phases and gates)' }, { v: 'hybrid', l: 'Hybrid (phases with sprints inside)' }],
    canMethod: may('Change project type'), onMethod: (e) => { const v = e.target.value; patch((x) => ({ ...x, method: v })); this.notify('Project type changed. Your tasks stay as they are.', '#B8F36B'); },
    cats: ['Not started', 'In progress', 'Being checked', 'Finished'].map((l, v) => ({ v, l })),
    dirty: !!wd, err: wd ? wd.err : '',
    rows: list.map((w, i) => ({ targets: list.filter((x) => x.key !== w.key).map((x) => ({ key: x.key, name: x.name, on: !w.to || !w.to.length || w.to.includes(x.key), toggle: () => change((l) => l.map((y, j) => { if (j !== i) return y; const all = l.filter((z) => z.key !== y.key).map((z) => z.key); const cur2 = y.to && y.to.length ? y.to : all; const nx = cur2.includes(x.key) ? cur2.filter((k) => k !== x.key) : cur2.concat([x.key]); return { ...y, to: nx.length === all.length ? [] : nx }; })) })), key: w.key, name: w.name, cat: w.cat, wip: w.wip || '', inUse: p.tasks.filter((t) => t.status === w.key).length, first: i === 0, last: i === list.length - 1,
      onName: (e) => { const v = e.target.value; change((l) => l.map((x, j) => (j === i ? { ...x, name: v } : x))); },
      onCat: (e) => { const v = Number(e.target.value); change((l) => l.map((x, j) => (j === i ? { ...x, cat: v } : x))); },
      onWip: (e) => { const v = Math.max(0, Number(e.target.value) || 0); change((l) => l.map((x, j) => (j === i ? { ...x, wip: v } : x))); },
      up: () => change((l) => { const n = l.slice(); [n[i - 1], n[i]] = [n[i], n[i - 1]]; return n; }),
      down: () => change((l) => { const n = l.slice(); [n[i + 1], n[i]] = [n[i], n[i + 1]]; return n; }),
      remove: () => change((l) => l.filter((_, j) => j !== i)) })),
    add: () => change((l) => { const n = l.slice(); n.splice(Math.max(0, n.length - 1), 0, { key: 's' + Date.now(), name: 'New status', cat: 1 }); return n; }),
    canDeleteProject: may('Delete project') && !!this.srv,
    deleteProject: () => {
      const typed = window.prompt('This deletes "' + p.name + '" with its tasks, sprints, milestones, documents and history, for good. Type the project code (' + p.code + ') to confirm.');
      if (typed === null) return;
      if (typed.trim() !== p.code) { this.notify('That code does not match, so nothing was deleted.', '#FFB86B'); return; }
      projectsW.removeProject(this, this.srv, pid).then(() => this.notify('Project deleted.', '#B8F36B')).catch((e) => this.notify(e.message || 'Could not delete the project.', '#FF9B8A'));
    },
    cancel: () => this.setState({ wfDraft: null }),
    save: () => {
      const l = wd.list;
      let err = '';
      if (l.length < 2) err = 'A workflow needs at least two statuses.';
      else if (l[0].cat === 3 || l[l.length - 1].cat !== 3) err = 'The first status must be unfinished and the last one finished.';
      else if (l.some((x) => !x.name.trim())) err = 'Every status needs a name.';
      else { const gone = p.tasks.find((t) => !l.some((x) => x.key === t.status)); if (gone) err = 'Move "' + gone.title + '" out of its status before removing that status.'; }
      if (err) { this.setState({ wfDraft: { ...wd, err } }); return; }
      const clean = l.map((x) => ({ ...x, name: x.name.trim(), wip: Number(x.wip) || 0 }));
      this.setState({ wfDraft: null });
      patch((x) => ({ ...x, workflow: clean, tasks: x.tasks.map((t) => { const w = clean.find((y) => y.key === t.status); return w ? { ...t, col: w.cat } : t; }) }));
      this.notify('Workflow saved.', '#B8F36B');
    }
  };

  const tabs = [['overview', 'Overview', null], ['board', 'Board', null], ['backlog', waterfall ? 'Phases' : 'Backlog & sprints', waterfall ? tools.tl.length : backlog.length], ['timeline', 'Timeline', null, 'View timeline'], ['reports', 'Reports', null, 'View reports'], ['docs', 'Documents', pages.length, 'View documents'], ['settings', 'Type & workflow', null, null, true]].filter(([, , , need, cfg]) => (!need || may(need)) && (!cfg || may('Manage workflow') || may('Change project type') || may('Delete project')));
  const stuck = p.stuck ? { unread: p.stuck.unread, move: () => projectsW.moveReadablePlan(this, this.srv, pid).then(() => this.notify('The plan was moved. It can be changed again.', '#B8F36B')).catch((e) => this.notify(e.message || 'Could not move the plan.', '#FF9B8A')) } : null;
  return {
    stuck,
    waterfall, methodLabel: { agile: 'Agile', waterfall: 'Waterfall', hybrid: 'Hybrid' }[p.method] || 'Agile',
    tabs: tabs.map(([k, l, n]) => ({ key: k, label: l, count: n, on: tab === k, go: () => this.setState({ planTab: k }) })), tab,
    isOverview: tab === 'overview' || !tabs.some(([k]) => k === tab), isBoard: tab === 'board', isBacklog: tab === 'backlog', isTimeline: tab === 'timeline' && may('View timeline'), isReports: tab === 'reports' && may('View reports'), isDocs: tab === 'docs' && may('View documents'), isSettings: tab === 'settings' && tabs.some(([k]) => k === 'settings'),
    epics, backlog, backlogPoints: pts(backlog.map((x) => p.tasks.find((t) => (t.id || '') === x.id) || {})), phases, canMs: can('projects', 'Manage milestones'), sprintOpts: [{ v: '', l: 'Backlog' }].concat(tools.spl.filter((x) => !x.milestoneClosed && !x.closed).map((x) => ({ v: x.id, l: x.name }))),
    addBacklog: tools.openTaskFormBacklog, canAdd: may('Edit tasks'), canRank: may('Rank backlog'), canPlan: may('Plan sprints'), timeline, reports, docs, settings,
    longStart: longDate(p.start)
  };
}
