import { addDaysIso, defaultTaskDue, milestoneTimeline, nextSprintName, normalizeProject, sprintDefaults, sprintTimeline, taskContext, todayIso as isoToday, validateMilestone, validateSprint, validateTask } from '~/logic/projects.js';

// renderVals step: projects. Project > milestone > sprint > task. Progress is always calculated from the tasks.
export function buildProjects(ctx) {
  const { badge, can, has, ini, pColor, s, set } = ctx;
  const nowMs = Date.now();
  const dateOf = (d) => new Date(d + 'T00:00:00+03:00').getTime();
  const shortDate = (d) => new Date(d + 'T00:00:00+03:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
  const longDate = (d) => new Date(d + 'T00:00:00+03:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  const statusColor = { 'On track': '#B8F36B', 'At risk': '#FFB86B', 'Off track': '#FF9B8A' };
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
  const initialsOf = (n) => n.split(' ').map((x) => x[0]).join('').slice(0, 2);
  const openProjectForm = () => {
    if (!has.projects || !can('projects', 'Create project')) return;
    const start = today;
    this.setState({ projectForm: { name: '', goal: '', owner: 'Samwel Okoth Omulo', team: ['Samwel Okoth Omulo'], teamQuery: '', start, due: addDaysIso(start, 29), budget: '', err: '' } });
  };
  let projectForm = { open: false };
  if (s.projectForm && has.projects && can('projects', 'Create project')) {
    const f = s.projectForm;
    const ownerOptions = [{ name: 'Samwel Okoth Omulo', color: '#C2185B' }].concat(this.dir.map((p) => ({ name: p.name, color: p.color })))
      .filter((p, i, all) => all.findIndex((other) => other.name === p.name) === i);
    const owners = ownerOptions.map((p) => ({ v: p.name, l: p.name }));
    const setField = (key) => (e) => {
      const value = key === 'team' ? Array.from(e.target.selectedOptions, (option) => option.value) : e.target.value;
      this.setState((st) => {
        const next = { ...st.projectForm, [key]: value, err: '' };
        if (key === 'owner' && !next.team.includes(value)) next.team = next.team.concat([value]);
        return { projectForm: next };
      });
    };
    projectForm = {
      open: true, ...f, owners,
      selectedTeam: f.team.map((name) => {
        const member = ownerOptions.find((p) => p.name === name) || { color: '#9DB0A6' };
        return { name, initials: ini(name), color: member.color, removable: name !== f.owner };
      }),
      teamResults: ownerOptions.filter((p) => p.name.toLowerCase().includes((f.teamQuery || '').trim().toLowerCase())).slice(0, 40)
        .map((p) => ({ name: p.name, initials: ini(p.name), color: p.color, selected: f.team.includes(p.name), toggle: () => this.setState((st) => {
          const team = st.projectForm.team.includes(p.name) ? st.projectForm.team.filter((n) => n !== p.name) : st.projectForm.team.concat([p.name]);
          return { projectForm: { ...st.projectForm, team, err: '' } };
        }) })),
      onName: setField('name'), onGoal: setField('goal'), onOwner: setField('owner'),
      onTeamQuery: setField('teamQuery'), onStart: setField('start'), onDue: setField('due'),
      removeTeamMember: (name) => this.setState((st) => ({ projectForm: { ...st.projectForm, team: st.projectForm.team.filter((n) => n !== name), err: '' } })),
      onBudget: setField('budget'), cancel: () => this.setState({ projectForm: null }),
      submit: () => {
        const name = f.name.trim(), goal = f.goal.trim();
        const budget = f.budget === '' ? 0 : Number(f.budget);
        let err = '';
        if (!name) err = 'Enter a project name.';
        else if (s.projects.some((p) => p.name.trim().toLowerCase() === name.toLowerCase())) err = 'A project with that name already exists.';
        else if (!goal) err = 'Describe the project goal.';
        else if (!f.owner) err = 'Choose a project owner.';
        else if (!f.start || !f.due || f.due < f.start) err = 'Choose a due date on or after the start date.';
        else if (!Number.isFinite(budget) || budget < 0) err = 'Enter a valid non-negative budget.';
        if (err) { this.setState({ projectForm: { ...f, err } }); return; }

        const id = 'project-' + Date.now();
        const usedCodes = new Set(s.projects.map((p) => p.code));
        let codeNumber = Math.max(0, ...s.projects.map((p) => Number((p.code.match(/\d+/) || [0])[0])).filter(Number.isFinite)) + 1;
        let code = 'PRJ-' + String(codeNumber).padStart(2, '0');
        while (usedCodes.has(code)) { codeNumber++; code = 'PRJ-' + String(codeNumber).padStart(2, '0'); }
        const owner = ownerOptions.find((p) => p.name === f.owner) || { color: '#9DB0A6' };
        const project = normalizeProject({
          id, name, code, color: owner.color, owner: f.owner,
          team: Array.from(new Set([f.owner, ...f.team])),
          status: 'On track', goal, start: f.start, due: f.due, budget, sync: 'Not synced',
          tickets: [], milestones: [{ id: id + '-m1', name: 'Project delivery', iso: f.due }],
          sprints: [], tasks: [],
          activity: [{ text: 'Project created by Samwel Okoth Omulo', time: 'Just now' }]
        });
        const flow = this._flowById ? this._flowById('f2') : null;
        this.setState({ projectForm: null });
        if (flow && flow.levels > 0 && this._makeRequest) {
          this._makeRequest('f2', 'New project: ' + name, {
            kind: 'project', project,
            note: 'Goal: ' + goal + ' · Owner: ' + f.owner + ' · Dates: ' + f.start + ' to ' + f.due + ' · Budget: ' + budget
          });
        } else {
          this.setState((st) => ({ projects: [project].concat(st.projects) }));
          this.notify('Project "' + name + '" created.', '#B8F36B');
        }
      }
    };
  }
  const projCalc = (p) => {
    const total = p.tasks.length, done = p.tasks.filter((t) => t.col === 3).length;
    const progress = total ? Math.round((done / total) * 100) : 0;
    const next = milestoneTimeline(p, today).find((m) => m.status !== 'done');
    const cur = sprintTimeline(p, today).find((sp) => sp.start <= today && today <= sp.end);
    return { total, done, progress, sc: statusColor[p.status] || '#9DB0A6', next, cur };
  };
  const projects = s.projects.map((p) => {
    const c = projCalc(p);
    return { ...p, initials: initialsOf(p.name), progress: c.progress, statusStyle: badge(c.sc), statusColor: c.sc,
      barStyle: 'height:100%;width:' + Math.max(3, c.progress) + '%;background:' + c.sc + ';border-radius:4px',
      nextMilestone: c.next ? c.next.name : 'All done', dueLabel: longDate(p.due), dueShort: shortDate(p.due),
      sprintLabel: c.cur ? c.cur.name : 'No sprint now',
      team: p.team.slice(0, 4).map((n) => ({ name: n, initials: ini(n), color: pColor[n] || '#9DB0A6' })), teamSize: p.team.length,
      open: () => this.setState({ section: 'projects', projectOpen: p.id, ticketOpen: null, taskForm: null }) };
  });
  const projStats = [
    { label: 'Active projects', n: projects.length, color: '#E8EFEA' },
    { label: 'On track', n: projects.filter((p) => p.status === 'On track').length, color: '#B8F36B' },
    { label: 'At risk', n: projects.filter((p) => p.status === 'At risk').length, color: '#FFB86B' },
    { label: 'Off track', n: projects.filter((p) => p.status === 'Off track').length, color: '#FF9B8A' }
  ];
  const cur_p = has.projects ? s.projects.find((p) => p.id === s.projectOpen) : null;
  let pj = {};
  let tools = {};
  if (cur_p) {
    const c = projCalc(cur_p);
    const pid = cur_p.id;
    const dueMs = dateOf(cur_p.due);
    const daysLeft = Math.max(0, Math.ceil((dueMs - nowMs) / 86400000));
    const tl = milestoneTimeline(cur_p, today);
    const spl = sprintTimeline(cur_p, today);
    const openSprints = spl.filter((sp) => !sp.milestoneClosed);
    const sprintSel = s.sprintFilter;
    const openTask = (ti) => this.setState({ taskOpen: { pid: pid, ti: ti } });
    const openMsClose = (m) => this.setState({ msDialog: { mode: 'close', pid: pid, mid: m.id, comment: '', err: '' } });
    const openMsReopen = (m) => this.setState({ msDialog: { mode: 'reopen', pid: pid, mid: m.id, comment: '', err: '' } });
    const openSprint = (milestone) => {
      const d = sprintDefaults(cur_p, milestone);
      this.setState({ msDialog: { mode: 'sprint', pid: pid, lockProject: true, milestone: milestone || '', name: nextSprintName(cur_p), start: d.start, end: d.end, err: '' } });
    };
    const openTaskForm = (sprintId) => {
      const pick = openSprints.find((sp) => sp.id === sprintId) || openSprints.find((sp) => sp.status === 'current') || openSprints.find((sp) => sp.status !== 'done') || openSprints[0];
      this.setState({ taskForm: { title: '', desc: '', objective: '', sprint: pick ? pick.id : '', owner: cur_p.owner, due: pick ? defaultTaskDue(pick, today) : '', dueTouched: false, err: '' } });
    };
    tools = { tl, spl, openMsClose, openMsReopen, openSprint, openTaskForm, longDate, shortDate };
    pj = {
      ...cur_p, initials: initialsOf(cur_p.name),
      statusStyle: badge(c.sc), progress: c.progress, doneCount: c.done, taskCount: c.total,
      barStyle: 'height:100%;width:' + Math.max(3, c.progress) + '%;background:' + c.sc + ';border-radius:3px',
      daysLeft, startLabel: longDate(cur_p.start), dueLabel: longDate(cur_p.due),
      msCount: tl.length, msDone: tl.filter((m) => m.closed).length, spCount: spl.length,
      canMilestones: canMs, tab: s.projTab, isMsTab: s.projTab === 'milestones', isSpTab: s.projTab === 'sprints',
      goMs: set({ projTab: 'milestones' }), goSp: set({ projTab: 'sprints' }),
      addMilestone: () => this.setState({ msDialog: { mode: 'new', pid: pid, lockProject: true, name: '', date: '', desc: '', err: '' } }),
      addSprint: () => openSprint(''),
      sprintOpts: [{ v: 'all', l: 'All sprints' }].concat(spl.map((sp) => ({ v: sp.id, l: sp.name + ' · ' + shortDate(sp.start) + ' – ' + shortDate(sp.end) + (sp.status === 'current' ? ' (current)' : '') }))),
      sprintValue: sprintSel, setSprint: (e) => this.setState({ sprintFilter: e.target.value }),
      canNewTask: canEditT, openTaskForm: () => openTaskForm(''),
      columns: colNames.map(([name, color], ci) => {
        const tasks = cur_p.tasks.map((t, ti) => ({ t, ti })).filter((x) => x.t.col === ci && (sprintSel === 'all' || x.t.sprint === sprintSel)).map(({ t, ti }) => {
          const inf = taskInfo(t);
          const tcx = taskContext(cur_p, t);
          return {
            ctxLabel: tcx.sprint ? tcx.sprint.name + (tcx.milestone ? ' · ' + tcx.milestone.name : '') : 'No sprint', ctxTitle: t.objective ? 'Objective: ' + t.objective : 'No objective written',
            title: t.title, primary: t.primary, secondary: t.secondary, initials: ini(t.primary), color: pColor[t.primary] || '#9DB0A6',
            hasSecondary: !!t.secondary, secInitials: t.secondary ? ini(t.secondary) : '', secColor: pColor[t.secondary] || '#9DB0A6',
            dueLabel: inf.label, dueStyle: tBadge(inf.color),
            hasLinks: t.links.length > 0, linkLabel: t.links.length + ' linked',
            cardStyle: 'padding:10px;border-radius:10px;background:#1D2A25;border:1px solid ' + (inf.overdue ? '#FF9B8A' : '#2A3832') + ';display:flex;flex-direction:column;gap:8px',
            open: () => openTask(ti),
            canMove: canMoveT && ci < 3, moveLabel: ci < 3 ? 'Move to ' + colNames[ci + 1][0] : '', move: () => this.moveTask(pid, ti, 1),
            canBack: canMoveT && ci > 0, backLabel: ci > 0 ? 'Move back to ' + colNames[ci - 1][0] : '', back: () => this.moveTask(pid, ti, -1)
          };
        });
        return { name, color, tasks, count: tasks.length };
      }),
      teamCount: cur_p.team.length,
      team: cur_p.team.map((n) => ({ name: n, initials: ini(n), color: pColor[n] || '#9DB0A6', load: cur_p.tasks.filter((t) => (t.primary === n || t.secondary === n) && t.col < 3).length }))
    };

    // ---- New task: title, description, objective (typed) and the sprint it belongs to ----
    let tform = { open: false };
    if (s.taskForm) {
      const f = s.taskForm;
      const sp = openSprints.find((x) => x.id === f.sprint) || null;
      const setF = (k) => (e) => {
        const v = e.target.value;
        this.setState((st) => {
          const nf = { ...st.taskForm, [k]: v, err: '' };
          if (k === 'due') nf.dueTouched = true;
          if (k === 'sprint') { const nsp = openSprints.find((x) => x.id === v); if (nsp && !nf.dueTouched) nf.due = defaultTaskDue(nsp, today); }
          return { taskForm: nf };
        });
      };
      tform = { open: true, title: f.title, desc: f.desc, objective: f.objective, sprint: f.sprint, owner: f.owner, due: f.due, err: f.err, hasErr: !!f.err,
        sprintOpts: openSprints.map((x) => ({ v: x.id, l: x.name + ' · ' + x.milestoneName + ' · ' + shortDate(x.start) + ' – ' + shortDate(x.end) })), noSprints: openSprints.length === 0,
        ownerOpts: cur_p.team.map((n) => ({ v: n, l: n })), minDue: sp ? sp.start : '', maxDue: sp ? sp.end : '',
        onTitle: setF('title'), onDesc: setF('desc'), onObjective: setF('objective'), onSprint: setF('sprint'), onOwner: setF('owner'), onDue: setF('due'),
        planLine: sp ? 'Tracked automatically: ' + sp.name + ' (' + sp.done + ' of ' + sp.total + ' done) › milestone "' + sp.milestoneName + '" › ' + cur_p.name : 'Pick a sprint to see where this task is tracked.',
        cancel: set({ taskForm: null }),
        addSprint: () => openSprint(''),
        submit: () => {
          const err = validateTask(cur_p, f);
          if (err) { this.setState({ taskForm: { ...f, err } }); return; }
          const title = f.title.trim();
          this.setState((st) => ({ taskForm: null, projects: st.projects.map((p) => (p.id === pid ? { ...p, tasks: p.tasks.concat([{ title: title, desc: f.desc.trim(), objective: f.objective.trim(), sprint: f.sprint, col: 0, primary: f.owner || p.owner, secondary: '', start: today, due: f.due, links: [] }]), activity: [{ text: 'You created "' + title + '" in ' + sp.name + ' (' + sp.milestoneName + ')', time: 'Just now' }].concat(p.activity) } : p)) }));
          this.notify('Task "' + title + '" added to ' + sp.name + '. It now counts toward "' + sp.milestoneName + '".', '#B8F36B');
        } };
    }
    pj.taskForm = tform;
  } else {
    pj.taskForm = { open: false };
  }

  // ---- Milestone / sprint dialog: new milestone, new sprint, close or reopen a milestone with a comment ----
  let msd = { open: false };
  const md = s.msDialog;
  if (md) {
    const mproj = s.projects.find((p) => p.id === md.pid) || null;
    const mtl = mproj ? milestoneTimeline(mproj, today).find((m) => m.id === md.mid) : null;
    const setM = (k) => (e) => { const v = e.target.value; this.setState((st) => ({ msDialog: { ...st.msDialog, [k]: v, err: '' } })); };
    const fail = (err) => this.setState((st) => ({ msDialog: { ...st.msDialog, err } }));
    const logMs = (p, text) => ({ ...p, activity: [{ text, time: 'Just now' }].concat(p.activity) });
    const openMilestones = mproj ? mproj.milestones.filter((m) => !m.closed).sort((a, b) => (a.iso < b.iso ? -1 : 1)) : [];
    const submit = () => {
      const d = this.state.msDialog;
      if (d.mode === 'new') {
        const err = validateMilestone(mproj, { name: d.name || '', date: d.date || '' });
        if (err) { fail(err); return; }
        const name = d.name.trim();
        this.setState((st) => ({ msDialog: null, projects: st.projects.map((p) => (p.id === mproj.id ? logMs({ ...p,
          milestones: p.milestones.concat([{ id: mproj.id + '-m' + Date.now(), name: name, iso: d.date, desc: (d.desc || '').trim(), closed: false, closedOn: '', closedBy: '', closedComment: '', history: [] }]) }, 'You added milestone "' + name + '" due ' + shortDate(d.date)) : p)) }));
        this.notify('Milestone "' + name + '" added to ' + mproj.name + '. Add sprints under it, then tasks in those sprints.', '#B8F36B');
        return;
      }
      if (d.mode === 'sprint') {
        const err = validateSprint(mproj, { milestone: d.milestone, name: d.name || '', start: d.start || '', end: d.end || '' });
        if (err) { fail(err); return; }
        const ms = mproj.milestones.find((m) => m.id === d.milestone);
        const name = d.name.trim();
        this.setState((st) => ({ msDialog: null, projects: st.projects.map((p) => (p.id === mproj.id ? logMs({ ...p, sprints: p.sprints.concat([{ id: mproj.id + '-s' + Date.now(), name: name, start: d.start, end: d.end, milestone: ms.id }]) }, 'You added ' + name + ' to milestone "' + ms.name + '" (' + shortDate(d.start) + ' – ' + shortDate(d.end) + ')') : p)) }));
        this.notify(name + ' added under "' + ms.name + '". Tasks in it count toward that milestone.', '#B8F36B');
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
    const titles = { new: 'New milestone', sprint: 'New sprint', close: 'Mark milestone as reached', reopen: 'Reopen milestone' };
    const labels = { new: 'Create milestone', sprint: 'Create sprint', close: 'Mark as reached', reopen: 'Reopen milestone' };
    msd = { open: true, mode: md.mode, isNew: md.mode === 'new', isSprint: md.mode === 'sprint', isClose: md.mode === 'close', isReopen: md.mode === 'reopen',
      title: titles[md.mode], submitLabel: labels[md.mode], projectName: mproj ? mproj.name : '',
      name: md.name || '', date: md.date || '', desc: md.desc || '', comment: md.comment || '', err: md.err || '', hasErr: !!md.err,
      milestone: md.milestone || '', milestoneOpts: openMilestones.map((m) => ({ v: m.id, l: m.name + ' · due ' + shortDate(m.iso) })),
      start: md.start || '', end: md.end || '', minDate: mproj ? mproj.start : '', maxDate: mproj ? mproj.due : '',
      rangeNote: mproj ? 'Must fall between ' + shortDate(mproj.start) + ' and ' + shortDate(mproj.due) + ' (the project dates).' : '',
      msName: mtl ? mtl.name : '', msSummary: mtl ? (mtl.total ? mtl.done + ' of ' + mtl.total + ' tasks done · ' + mtl.progress + '%' : 'No tasks yet') : '',
      commentLabel: md.mode === 'close' ? 'What was achieved? (required)' : 'Why is it being reopened? (required)',
      onName: setM('name'), onDate: setM('date'), onDesc: setM('desc'), onComment: setM('comment'), onStart: setM('start'), onEnd: setM('end'),
      onMilestone: (e) => {
        const v = e.target.value; const d = sprintDefaults(mproj, v);
        this.setState((st) => ({ msDialog: { ...st.msDialog, milestone: v, start: d.start, end: d.end, err: '' } }));
      },
      cancel: set({ msDialog: null }), submit };
  }

  // ---- Task detail ----
  let td = {};
  const to = s.taskOpen;
  const tp = has.projects && to && s.projects.find((p) => p.id === to.pid);
  const tt = tp && tp.tasks[to.ti];
  if (tt) {
    const inf = taskInfo(tt);
    const updT = (fnT) => this.setState((st) => ({ projects: st.projects.map((p) => (p.id === tp.id ? { ...p, tasks: p.tasks.map((x, j) => (j === to.ti ? fnT(x) : x)) } : p)) }));
    const rk = tp.id + ':' + to.ti;
    const extraRem = s.taskReminders[rk] || [];
    const tcx = taskContext(tp, tt);
    const spT = sprintTimeline(tp, today).filter((sp) => !sp.milestoneClosed || sp.id === tt.sprint);
    td = {
      project: tp.name, pcolor: tp.color, pinitials: initialsOf(tp.name),
      title: tt.title, desc: tt.desc || 'No description yet.',
      canEdit: canEditT, objective: tt.objective, setObjective: (e) => { const v = e.target.value.trim(); updT((x) => ({ ...x, objective: v })); },
      sprintId: tt.sprint, sprintOpts: spT.map((sp) => ({ v: sp.id, l: sp.name + ' · ' + sp.milestoneName + ' · ' + shortDate(sp.start) + ' – ' + shortDate(sp.end) })),
      setSprint: (e) => {
        const sp = tp.sprints.find((x) => x.id === e.target.value);
        if (!sp) return;
        updT((x) => ({ ...x, sprint: sp.id, due: x.due < sp.start ? sp.start : x.due > sp.end ? sp.end : x.due }));
      },
      milestoneName: tcx.milestone ? tcx.milestone.name : 'None', milestoneDue: tcx.milestone ? longDate(tcx.milestone.iso) : '',
      milestoneState: tcx.milestone ? (tcx.milestone.closed ? 'Reached' : tcx.milestone.iso < today ? 'Late' : 'Open') : '',
      sprintDates: tcx.sprint ? shortDate(tcx.sprint.start) + ' – ' + shortDate(tcx.sprint.end) : '',
      dueISO: tt.due, minDue: tcx.sprint ? tcx.sprint.start : tt.start, maxDue: tcx.sprint ? tcx.sprint.end : '',
      setDue: (e) => { const v = e.target.value; if (v) updT((x) => ({ ...x, due: v })); },
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
  if (!has.projects) msd.open = false;
  return { canEditT, canMs, cur_p, msd, nowMs, openProjectForm, projectForm, pj, projStats, projects, projTools: tools, taskNotifs, td, today, tt };
}
