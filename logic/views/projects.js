import { defaultDue, milestoneTimeline, sprintList, sprintOf, taskContext, todayIso as isoToday, validateMilestone, validateTask } from '~/logic/projects.js';

// renderVals step: projects. Receives the names built by earlier steps and returns the ones later steps need.
export function buildProjects(ctx) {
  const { badge, can, has, ini, pColor, s, set, tickets } = ctx;
  // ---- Projects ----
  const nowMs = Date.now();
  const dateOf = (d) => new Date(d + 'T00:00:00+03:00').getTime();
  const shortDate = (d) => new Date(d + 'T00:00:00+03:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
  const longDate = (d) => new Date(d + 'T00:00:00+03:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  const statusColor = { 'On track': '#B8F36B', 'At risk': '#FFB86B', 'Off track': '#FF9B8A' };
  const levelColor = { High: '#FF9B8A', Medium: '#FFB86B', Low: '#9DB0A6' };
  const colNames = [['To do', '#9DB0A6'], ['In progress', '#7CC4FF'], ['Review', '#FFB86B'], ['Done', '#B8F36B']];
  const dayStart = new Date(new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }) + 'T00:00:00+03:00').getTime();
  const taskInfo = (t) => {
    const dueMs = dateOf(t.due), startMs = dateOf(t.start);
    const dd = Math.round((dueMs - dayStart) / 86400000);
    const overdue = t.col < 3 && dd < 0;
    const soon = t.col < 3 && dd >= 0 && dd <= 2;
    const color = t.col === 3 ? '#B8F36B' : overdue ? '#FF9B8A' : soon ? '#FFB86B' : '#9DB0A6';
    const label = t.col === 3 ? 'Done' : overdue ? 'Overdue ' + (-dd) + 'd' : dd === 0 ? 'Due today' : soon ? 'Due in ' + dd + 'd' : 'Due ' + shortDate(t.due);
    const pct = Math.min(100, Math.max(0, Math.round(((nowMs - startMs) / Math.max(1, dueMs + 86400000 - startMs)) * 100)));
    return { dd, overdue, soon, color, label, pct, overDays: (-dd) + (dd === -1 ? ' day' : ' days') };
  };
  const tBadge = (c) => 'display:inline-block;padding:2px 7px;border-radius:6px;font-size:10px;font-weight:700;background:' + c + '22;color:' + c;
  const canMoveT = can('projects', 'Move tasks');
  const today = isoToday(nowMs);
  const canMs = can('projects', 'Manage milestones');
  const canEditT = can('projects', 'Edit tasks');
  const projCalc = (p) => {
    const total = p.tasks.length, done = p.tasks.filter((t) => t.col === 3).length;
    const progress = total ? Math.round((done / total) * 100) : 0;
    const sc = statusColor[p.status] || '#9DB0A6';
    const next = milestoneTimeline(p, today).find((m) => m.status !== 'done');
    return { total, done, progress, sc, next };
  };
  const projects = s.projects.map((p) => {
    const c = projCalc(p);
    return { ...p, initials: p.name.split(' ').map((x) => x[0]).join('').slice(0, 2), progress: c.progress,
      statusStyle: badge(c.sc), statusColor: c.sc,
      barStyle: 'height:100%;width:' + Math.max(3, c.progress) + '%;background:' + c.sc + ';border-radius:4px',
      nextMilestone: c.next ? c.next.name : 'All done', dueLabel: longDate(p.due), dueShort: shortDate(p.due),
      riskCount: p.risks.length, riskColor: p.risks.some((r) => r.level === 'High') ? '#FF9B8A' : '#FFB86B',
      team: p.team.slice(0, 4).map((n) => ({ name: n, initials: ini(n), color: pColor[n] || '#9DB0A6' })), teamSize: p.team.length,
      open: () => this.setState({ section: 'projects', projectOpen: p.id, ticketOpen: null, newTask: '' }) };
  });
  const projStats = [
    { label: 'Active projects', n: projects.length, color: '#E8EFEA' },
    { label: 'On track', n: projects.filter((p) => p.status === 'On track').length, color: '#B8F36B' },
    { label: 'At risk', n: projects.filter((p) => p.status === 'At risk').length, color: '#FFB86B' },
    { label: 'Off track', n: projects.filter((p) => p.status === 'Off track').length, color: '#FF9B8A' }
  ];
  const cur_p = s.projects.find((p) => p.id === s.projectOpen);
  let pj = {};
  if (cur_p) {
    const c = projCalc(cur_p);
    const pid = cur_p.id;
    const startMs = dateOf(cur_p.start), dueMs = dateOf(cur_p.due);
    const daysLeft = Math.max(0, Math.ceil((dueMs - nowMs) / 86400000));
    const timeUsed = Math.min(100, Math.max(0, Math.round(((nowMs - startMs) / (dueMs - startMs)) * 100)));
    const openTask = (ti) => this.setState({ taskOpen: { pid: pid, ti: ti } });
    const msColor = { done: '#B8F36B', ready: '#B8F36B', late: '#FF9B8A', next: '#FFB86B', upcoming: '#9DB0A6' };
    const msLabel = { done: 'Reached', ready: 'Ready to close', late: 'Late', next: 'Next up', upcoming: 'Upcoming' };
    const tl = milestoneTimeline(cur_p, today);
    const curSprint = sprintOf(cur_p, today);
    const sprintSel = s.sprintFilter === 'all' ? 'all' : Number(s.sprintFilter);
    const pctBar = (pc, c) => 'height:100%;width:' + (pc > 0 ? Math.max(3, pc) : 0) + '%;background:' + c + ';border-radius:3px';
    const addObjective = (mid) => {
      const title = (this.state.objDraft[mid] || '').trim();
      if (!title) return;
      const ms = cur_p.milestones.find((m) => m.id === mid);
      this.setState((st) => ({ objDraft: { ...st.objDraft, [mid]: '' }, projects: st.projects.map((p) => (p.id === pid ? { ...p, objectives: p.objectives.concat([{ id: pid + '-o' + Date.now(), title: title, milestone: mid }]), activity: [{ text: 'You added objective "' + title + '" to milestone "' + ms.name + '"', time: 'Just now' }].concat(p.activity) } : p)) }));
    };
    const openMsClose = (m) => this.setState({ msDialog: { mode: 'close', pid: pid, mid: m.id, comment: '', err: '' } });
    const openMsReopen = (m) => this.setState({ msDialog: { mode: 'reopen', pid: pid, mid: m.id, comment: '', err: '' } });
    const openTaskForm = () => {
      const firstOpen = tl.find((m) => !m.closed && m.objectives.length);
      const obj = firstOpen ? firstOpen.objectives[0].id : '';
      this.setState({ taskForm: { title: '', desc: '', objective: obj, owner: cur_p.owner, due: defaultDue(cur_p, firstOpen, today), dueTouched: false, err: '' } });
    };
    pj = {
      ...cur_p, initials: cur_p.name.split(' ').map((x) => x[0]).join('').slice(0, 2),
      statusStyle: badge(c.sc), progress: c.progress, doneCount: c.done, taskCount: c.total,
      inFlight: cur_p.tasks.filter((t) => t.col === 1 || t.col === 2).length,
      barStyle: 'height:100%;width:' + Math.max(3, c.progress) + '%;background:' + c.sc + ';border-radius:3px',
      daysLeft, timeUsed, startLabel: longDate(cur_p.start), dueLabel: longDate(cur_p.due),
      budgetColor: cur_p.budget > timeUsed + 10 ? '#FF9B8A' : '#E8EFEA',
      budgetNote: cur_p.budget > timeUsed + 10 ? 'Spending ahead of the timeline' : 'In line with the timeline',
      msCount: tl.length, msDone: tl.filter((m) => m.closed).length,
      canMilestones: canMs, addMilestone: () => this.setState({ msDialog: { mode: 'new', pid: pid, lockProject: true, name: '', date: '', desc: '', err: '' } }),
      milestones: tl.map((m) => {
        const key = m.ready ? 'ready' : m.status;
        const c = msColor[key];
        return { id: m.id, name: m.name, desc: m.desc, color: c, stateLabel: msLabel[key], dateLabel: longDate(m.iso), progress: m.progress,
          countLabel: m.total ? m.done + ' of ' + m.total + ' tasks done' : 'No tasks yet',
          barStyle: pctBar(m.progress, c),
          cardStyle: 'padding:14px;border-radius:14px;background:#121B17;border:1px solid ' + (m.ready ? '#B8F36B' : m.status === 'late' ? '#6A3A34' : '#24322C') + ';display:flex;flex-direction:column;gap:12px',
          objectives: m.objectives.map((o) => ({ id: o.id, title: o.title, label: o.total ? o.done + ' / ' + o.total + ' tasks' : 'No tasks yet', progress: o.progress, barStyle: pctBar(o.progress, o.progress === 100 ? '#B8F36B' : '#7CC4FF') })),
          noObjectives: m.objectives.length === 0,
          closed: m.closed, closedLine: m.closed ? 'Reached ' + longDate(m.closedOn) + ' by ' + m.closedBy : '', closedComment: m.closedComment,
          canClose: canMs && m.ready, canReopen: canMs && m.closed,
          closeHint: m.closed || m.ready ? '' : m.total ? m.openTasks + (m.openTasks === 1 ? ' task is' : ' tasks are') + ' still open. Finish them to close this milestone.' : 'Add objectives and tasks to track progress toward this milestone.',
          hasHint: !m.closed && !m.ready,
          close: () => openMsClose(m), reopen: () => openMsReopen(m),
          objDraft: s.objDraft[m.id] || '', onObjDraft: (e) => this.setState((st) => ({ objDraft: { ...st.objDraft, [m.id]: e.target.value } })),
          objKey: (e) => { if (e.key === 'Enter') addObjective(m.id); }, addObj: () => addObjective(m.id),
          canAddObj: canMs && !m.closed,
          history: m.history.slice().reverse().map((h) => ({ text: h.action + ' by ' + h.who + ' · ' + longDate(h.when), comment: h.comment })) };
      }),
      sprintOpts: [{ v: 'all', l: 'All sprints' }].concat(sprintList(cur_p).map((sp) => ({ v: String(sp.n), l: sp.label + ' · ' + shortDate(sp.start) + ' – ' + shortDate(sp.end) + (sp.n === curSprint.n ? ' (current)' : '') }))),
      sprintValue: s.sprintFilter, setSprint: (e) => this.setState({ sprintFilter: e.target.value }),
      canNewTask: canEditT, openTaskForm,
      columns: colNames.map(([name, color], ci) => {
        const tasks = cur_p.tasks.map((t, ti) => ({ t, ti })).filter((x) => x.t.col === ci && (sprintSel === 'all' || sprintOf(cur_p, x.t.due).n === sprintSel)).map(({ t, ti }) => {
          const inf = taskInfo(t);
          const tcx = taskContext(cur_p, t, today);
          return {
            ctxLabel: 'S' + tcx.sprint.n + (tcx.milestone ? ' · ' + tcx.milestone.name : ''), ctxTitle: tcx.sprint.label + (tcx.objective ? ' · ' + tcx.objective.title : ''),
            title: t.title, primary: t.primary, secondary: t.secondary, initials: ini(t.primary), color: pColor[t.primary] || '#9DB0A6',
            hasSecondary: !!t.secondary, secInitials: t.secondary ? ini(t.secondary) : '', secColor: pColor[t.secondary] || '#9DB0A6',
            dueLabel: inf.label, dueStyle: tBadge(inf.color),
            hasLinks: t.links.length > 0, linkLabel: t.links.length + ' linked',
            cardStyle: 'padding:10px;border-radius:10px;background:#1D2A25;border:1px solid #2A3832;display:flex;flex-direction:column;gap:8px;' + (inf.overdue ? 'border-left:3px solid #FF9B8A' : ''),
            open: () => openTask(ti),
            canMove: canMoveT && ci < 3, moveLabel: ci < 3 ? 'Move to ' + colNames[ci + 1][0] : '', move: () => this.moveTask(pid, ti, 1),
            canBack: canMoveT && ci > 0, backLabel: ci > 0 ? 'Move back to ' + colNames[ci - 1][0] : '', back: () => this.moveTask(pid, ti, -1)
          };
        });
        return { name, color, tasks, count: tasks.length };
      }),
      teamCount: cur_p.team.length,
      team: cur_p.team.map((n) => ({ name: n, initials: ini(n), color: pColor[n] || '#9DB0A6', load: cur_p.tasks.filter((t) => (t.primary === n || t.secondary === n) && t.col < 3).length })),
      risks: cur_p.risks.map((r) => ({ ...r, style: 'flex:none;' + badge(levelColor[r.level]) })),
      linked: cur_p.tickets.map((id) => tickets.find((t) => t.id === id)).filter(Boolean).map((t) => ({ id: t.id, title: t.title, statusLabel: t.statusLabel, slaLabel: t.slaLabel, slaColor: t._c.color, open: () => this.setState({ section: 'tickets', ticketOpen: t.id, projectOpen: null }) })),
      noTickets: cur_p.tickets.length === 0,
      call: () => this.startCall('video', cur_p.name + ' standup', cur_p.name.split(' ').map((x) => x[0]).join('').slice(0, 2), cur_p.color)
    };
  }

  // ---- New task form: title, description and objective. Milestone and sprint follow automatically. ----
  let tform = { open: false };
  if (cur_p && s.taskForm) {
    const f = s.taskForm;
    const tlNow = milestoneTimeline(cur_p, today);
    const objOpts = [];
    tlNow.filter((m) => !m.closed).forEach((m) => m.objectives.forEach((o) => objOpts.push({ v: o.id, l: m.name + ' › ' + o.title })));
    const obj = cur_p.objectives.find((o) => o.id === f.objective);
    const ms = obj ? cur_p.milestones.find((m) => m.id === obj.milestone) : null;
    const sp = f.due ? sprintOf(cur_p, f.due) : null;
    const setF = (k) => (e) => {
      const v = e.target.value;
      this.setState((st) => {
        const nf = { ...st.taskForm, [k]: v, err: '' };
        if (k === 'due') nf.dueTouched = true;
        if (k === 'objective' && !nf.dueTouched) {
          const o2 = cur_p.objectives.find((o) => o.id === v);
          nf.due = defaultDue(cur_p, o2 ? cur_p.milestones.find((m) => m.id === o2.milestone) : null, today);
        }
        return { taskForm: nf };
      });
    };
    tform = { open: true, title: f.title, desc: f.desc, objective: f.objective, owner: f.owner, due: f.due, err: f.err, hasErr: !!f.err,
      objOpts, noObjectives: objOpts.length === 0, ownerOpts: cur_p.team.map((n) => ({ v: n, l: n })), minDue: cur_p.start, maxDue: cur_p.due,
      onTitle: setF('title'), onDesc: setF('desc'), onObjective: setF('objective'), onOwner: setF('owner'), onDue: setF('due'),
      planLine: ms && sp ? 'Counts toward milestone "' + ms.name + '" (due ' + shortDate(ms.iso) + ') · objective "' + obj.title + '" · ' + sp.label + ' (' + shortDate(sp.start) + ' – ' + shortDate(sp.end) + ')' : 'Pick an objective to see where this task fits.',
      lateWarn: !!(ms && f.due && f.due > ms.iso), warnLine: ms ? 'This is due after the milestone date, so the milestone will be late.' : '',
      cancel: set({ taskForm: null }),
      submit: () => {
        const err = validateTask(cur_p, f);
        if (err) { this.setState({ taskForm: { ...f, err } }); return; }
        const title = f.title.trim();
        this.setState((st) => ({ taskForm: null, projects: st.projects.map((p) => (p.id === cur_p.id ? { ...p, tasks: p.tasks.concat([{ title: title, desc: f.desc.trim(), objective: f.objective, col: 0, primary: f.owner || p.owner, secondary: '', start: today, due: f.due, links: [] }]), activity: [{ text: 'You created "' + title + '" under "' + obj.title + '" (' + ms.name + ', ' + sp.label + ')', time: 'Just now' }].concat(p.activity) } : p)) }));
        this.notify('Task "' + title + '" added to ' + sp.label + ' and counted toward "' + ms.name + '".', '#B8F36B');
      } };
  }
  pj.taskForm = tform;

  // ---- Milestone dialog: new milestone, close with comment, reopen with comment ----
  let msd = { open: false };
  const md = s.msDialog;
  if (md) {
    const mproj = s.projects.find((p) => p.id === md.pid) || null;
    const mtl = mproj ? milestoneTimeline(mproj, today).find((m) => m.id === md.mid) : null;
    const setM = (k) => (e) => { const v = e.target.value; this.setState((st) => ({ msDialog: { ...st.msDialog, [k]: v, err: '' } })); };
    const fail = (err) => this.setState((st) => ({ msDialog: { ...st.msDialog, err } }));
    const logMs = (p, text) => ({ ...p, activity: [{ text, time: 'Just now' }].concat(p.activity) });
    const submit = () => {
      const d = this.state.msDialog;
      if (d.mode === 'new') {
        const err = validateMilestone(mproj, { name: d.name || '', date: d.date || '' });
        if (err) { fail(err); return; }
        const mid = mproj.id + '-m' + Date.now();
        const name = d.name.trim();
        this.setState((st) => ({ msDialog: null, projects: st.projects.map((p) => (p.id === mproj.id ? logMs({ ...p,
          milestones: p.milestones.concat([{ id: mid, name: name, iso: d.date, desc: (d.desc || '').trim(), closed: false, closedOn: '', closedBy: '', closedComment: '', history: [] }]),
          objectives: p.objectives.concat([{ id: mproj.id + '-o' + Date.now(), title: name + ' deliverables', milestone: mid }]) }, 'You added milestone "' + name + '" due ' + shortDate(d.date)) : p)) }));
        this.notify('Milestone "' + name + '" added to ' + mproj.name + '. Add tasks under its objective to track progress.', '#B8F36B');
        return;
      }
      const comment = (d.comment || '').trim();
      if (comment.length < 8) { fail('Add a comment of at least a few words. It is kept with the milestone.'); return; }
      if (!mtl) { fail('This milestone no longer exists.'); return; }
      if (d.mode === 'close') {
        if (!mtl.ready) { fail(mtl.openTasks + ' task' + (mtl.openTasks === 1 ? ' is' : 's are') + ' still open. Finish them before closing the milestone.'); return; }
        this.setState((st) => ({ msDialog: null, projects: st.projects.map((p) => (p.id === mproj.id ? logMs({ ...p, milestones: p.milestones.map((m) => (m.id === mtl.id ? { ...m, closed: true, closedOn: today, closedBy: 'Samwel Okoth Omulo', closedComment: comment, history: m.history.concat([{ when: today, who: 'Samwel Okoth Omulo', action: 'Reached', comment }]) } : m)) }, 'Milestone "' + mtl.name + '" reached: "' + comment + '"') : p)) }));
        this.notify('Milestone "' + mtl.name + '" marked as reached. Your comment was saved.', '#B8F36B');
      } else {
        this.setState((st) => ({ msDialog: null, projects: st.projects.map((p) => (p.id === mproj.id ? logMs({ ...p, milestones: p.milestones.map((m) => (m.id === mtl.id ? { ...m, closed: false, closedOn: '', closedBy: '', closedComment: '', history: m.history.concat([{ when: today, who: 'Samwel Okoth Omulo', action: 'Reopened', comment }]) } : m)) }, 'Milestone "' + mtl.name + '" reopened: "' + comment + '"') : p)) }));
        this.notify('Milestone "' + mtl.name + '" reopened.', '#FFB86B');
      }
    };
    msd = { open: true, mode: md.mode, isNew: md.mode === 'new', isClose: md.mode === 'close', isReopen: md.mode === 'reopen',
      title: md.mode === 'new' ? 'New milestone' : md.mode === 'close' ? 'Mark milestone as reached' : 'Reopen milestone',
      submitLabel: md.mode === 'new' ? 'Create milestone' : md.mode === 'close' ? 'Mark as reached' : 'Reopen milestone',
      pid: md.pid || '', projectOpts: s.projects.map((p) => ({ v: p.id, l: p.name })), lockProject: !!md.lockProject, projectName: mproj ? mproj.name : '',
      name: md.name || '', date: md.date || '', desc: md.desc || '', comment: md.comment || '', err: md.err || '', hasErr: !!md.err,
      minDate: mproj ? mproj.start : '', maxDate: mproj ? mproj.due : '',
      rangeNote: mproj ? 'Must fall between ' + shortDate(mproj.start) + ' and ' + shortDate(mproj.due) + ' (the project dates).' : 'Choose a project first.',
      msName: mtl ? mtl.name : '', msSummary: mtl ? (mtl.total ? mtl.done + ' of ' + mtl.total + ' tasks done · ' + mtl.progress + '%' : 'No tasks yet') : '',
      commentLabel: md.mode === 'close' ? 'What was achieved? (required)' : 'Why is it being reopened? (required)',
      onProject: (e) => { const v = e.target.value; this.setState((st) => ({ msDialog: { ...st.msDialog, pid: v, err: '' } })); },
      onName: setM('name'), onDate: setM('date'), onDesc: setM('desc'), onComment: setM('comment'),
      cancel: set({ msDialog: null }), submit };
  }
  const newMilestone = () => this.setState({ msDialog: { mode: 'new', pid: '', name: '', date: '', desc: '', err: '' } });

  // ---- Task detail ----
  let td = {};
  const to = s.taskOpen;
  const tp = to && s.projects.find((p) => p.id === to.pid);
  const tt = tp && tp.tasks[to.ti];
  if (tt) {
    const inf = taskInfo(tt);
    const updT = (fnT) => this.setState((st) => ({ projects: st.projects.map((p) => (p.id === tp.id ? { ...p, tasks: p.tasks.map((x, j) => (j === to.ti ? fnT(x) : x)) } : p)) }));
    const rk = tp.id + ':' + to.ti;
    const extraRem = s.taskReminders[rk] || [];
    const tcx = taskContext(tp, tt, today);
    const tlT = milestoneTimeline(tp, today);
    const objOptsT = [];
    tlT.forEach((m) => m.objectives.forEach((o) => { if (!m.closed || o.id === tt.objective) objOptsT.push({ v: o.id, l: m.name + ' › ' + o.title }); }));
    td = {
      project: tp.name, pcolor: tp.color, pinitials: tp.name.split(' ').map((x) => x[0]).join('').slice(0, 2),
      title: tt.title, desc: tt.desc || 'No description yet.',
      canEdit: canEditT, objectiveId: tt.objective, objOpts: objOptsT, objectiveName: tcx.objective ? tcx.objective.title : 'None',
      setObjective: (e) => { const v = e.target.value; updT((x) => ({ ...x, objective: v })); },
      milestoneName: tcx.milestone ? tcx.milestone.name : 'None', milestoneDue: tcx.milestone ? longDate(tcx.milestone.iso) : '',
      milestoneState: tcx.milestone ? (tcx.milestone.closed ? 'Reached' : tcx.milestone.iso < today ? 'Late' : 'Open') : '',
      sprintLabel: tcx.sprint.label, sprintDates: shortDate(tcx.sprint.start) + ' – ' + shortDate(tcx.sprint.end), sprintCurrent: tcx.sprint.n === tcx.currentSprint.n,
      dueISO: tt.due, minDue: tt.start, setDue: (e) => { const v = e.target.value; if (v) updT((x) => ({ ...x, due: v })); },
      col: colNames[tt.col][0], colStyle: 'height:36px;padding:0 14px;border-radius:18px;display:flex;align-items:center;font-size:13px;font-weight:700;background:' + colNames[tt.col][1] + '22;color:' + colNames[tt.col][1],
      canBack: canMoveT && tt.col > 0, prevCol: tt.col > 0 ? colNames[tt.col - 1][0] : '', back: () => this.moveTask(tp.id, to.ti, -1),
      canFwd: canMoveT && tt.col < 3, nextCol: tt.col < 3 ? colNames[tt.col + 1][0] : '', fwd: () => this.moveTask(tp.id, to.ti, 1),
      overdue: inf.overdue, overDays: inf.overDays,
      reminderNote: 'An automatic reminder went to ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' today.',
      remind: () => { this.setState((st) => ({ taskReminders: { ...st.taskReminders, [rk]: (st.taskReminders[rk] || []).concat(['You reminded ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' just now']) } })); this.notify('Reminder sent to ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' about "' + tt.title + '".', '#FFB86B'); },
      startLabel: longDate(tt.start), dueLabel: longDate(tt.due), dueColor: inf.color,
      timeBar: 'height:100%;width:' + Math.max(3, tt.col === 3 ? 100 : inf.pct) + '%;background:' + inf.color + ';border-radius:4px',
      timeNote: tt.col === 3 ? 'Finished.' : inf.overdue ? inf.overDays + ' past the due date.' : inf.dd === 0 ? 'Due today.' : inf.dd + ' days left · ' + inf.pct + '% of the time used.',
      primary: tt.primary, pIni: ini(tt.primary), pColor: pColor[tt.primary] || '#9DB0A6',
      secondary: tt.secondary || '',
      teamOpts: tp.team.filter((n) => n !== tt.primary).map((n) => ({ name: n })),
      setSecondary: (e) => { const v = e.target.value; updT((x) => ({ ...x, secondary: v })); },
      links: tt.links.map((li) => ({ li, t: tp.tasks[li] })).filter((x) => x.t).map(({ li, t }) => ({ title: t.title, col: colNames[t.col][0], due: shortDate(t.due),
        open: () => this.setState({ taskOpen: { pid: tp.id, ti: li } }),
        unlink: () => updT((x) => ({ ...x, links: x.links.filter((y) => y !== li) })) })),
      noLinks: tt.links.length === 0,
      linkOpts: tp.tasks.map((t, i) => ({ i: String(i), title: t.title })).filter((o) => Number(o.i) !== to.ti && tt.links.indexOf(Number(o.i)) === -1),
      addLink: (e) => { const v = e.target.value; if (v === '') return; updT((x) => ({ ...x, links: x.links.concat([Number(v)]) })); },
      reminders: (inf.overdue ? ['Automatic overdue reminder sent to ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' today'] : inf.soon ? ['Due-soon reminder sent to ' + tt.primary + ' today'] : ['Reminders go out 2 days before the due date and every day once overdue']).concat(extraRem).map((x) => ({ text: x }))
    };
  }
  const taskNotifs = [];
  if (has.projects) s.projects.forEach((p) => p.tasks.forEach((t, i) => {
    const inf = taskInfo(t);
    if (inf.overdue) taskNotifs.push({ text: 'Task "' + t.title + '" (' + p.name + ') is ' + inf.overDays + ' overdue. Reminded ' + t.primary + (t.secondary ? ' and ' + t.secondary : '') + '.', color: '#FF9B8A', time: 'Today', ticket: null, task: { pid: p.id, ti: i } });
  }));
  return { canMs, cur_p, msd, newMilestone, nowMs, pj, projStats, projects, taskNotifs, td, tt };
}
