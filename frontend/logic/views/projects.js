import * as projectsW from '~/logic/server/widgets/projects.js';
import { buildPlan } from '~/logic/views/projectPlan.js';
import { CAT_COLORS, addDaysIso, boardTasks, openDeps, statusEntry, statusOf, stepOf, wfOf, defaultTaskDue, mayEditTask, mayMoveTask, milestoneTimeline, nextSprintName, normalizeProject, sprintDefaults, sprintTimeline, taskContext, todayIso as isoToday, validateMilestone, validateSprint, validateTask } from '~/logic/projects.js';

// renderVals step: projects. Project > milestone > sprint > task. Progress is always calculated from the tasks.
export function buildProjects(ctx) {
  const { badge, can: canBase, has, ini, pColor, s, set } = ctx;
  // someone looking at a project they are not on (View all projects) changes nothing in it
  const openedView = !!((s.projects || []).find((p) => p.id === s.projectOpen) || {}).viewOnly;
  const can = (m, r) => canBase(m, r) && (!openedView || (m === 'projects' && /^(View|Export pages)/.test(r)) || r === 'Create project');
  const nowMs = Date.now();
  const dateOf = (d) => new Date(d + 'T00:00:00+03:00').getTime();
  const shortDate = (d) => new Date(d + 'T00:00:00+03:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
  const longDate = (d) => new Date(d + 'T00:00:00+03:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
  const statusColor = { 'On track': '#B8F36B', 'At risk': '#FFB86B', 'Off track': '#FF9B8A' };
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
  const me = this.srv ? this.srv.me.name : 'Samwel Okoth Omulo';   // who is signed in: the project manager and a task's owners are known by name
  const today = isoToday(nowMs);
  const canMs = can('projects', 'Manage milestones');
  const canEditT = can('projects', 'Edit tasks');
  const initialsOf = (n) => n.split(' ').map((x) => x[0]).join('').slice(0, 2);
  const openProjectForm = () => {
    if (!has.projects || !can('projects', 'Create project')) return;
    if (this.srv) { this.setState({ projectEligible: null }); projectsW.eligible(this.srv).then((l) => this.setState({ projectEligible: l })).catch((e) => { this.setState({ projectEligible: [] }); this.notify(e.message || 'Could not load who can join.', '#FF9B8A'); }); }
    const start = today;
    this.setState({ projectForm: { name: '', goal: '', owner: this.srv ? this.srv.me.name : 'Samwel Okoth Omulo', method: 'agile', team: [this.srv ? this.srv.me.name : 'Samwel Okoth Omulo'], teamQuery: '', start, due: addDaysIso(start, 29), budget: '', err: '' } });
  };
  let projectForm = { open: false };
  if (s.projectForm && has.projects && can('projects', 'Create project')) {
    const f = s.projectForm;
    const allowed = this.srv && s.projectEligible ? new Set(s.projectEligible) : null;
    const ownerOptions = (this.srv ? (allowed && allowed.has(this.srv.me.person) ? [{ name: this.srv.me.name, color: (s.meInfo || {}).color || '#C2185B' }] : []) : [{ name: 'Samwel Okoth Omulo', color: '#C2185B' }]).concat(this.dir.filter((p) => !this.srv || (allowed && allowed.has(p.uid))).map((p) => ({ name: p.name, color: p.color })))
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
      open: true, ...f, owners, loadingPeople: !!this.srv && !s.projectEligible,
      selectedTeam: f.team.map((name) => {
        const member = ownerOptions.find((p) => p.name === name) || { color: '#9DB0A6' };
        return { name, initials: ini(name), color: member.color, removable: name !== f.owner };
      }),
      teamResults: ownerOptions.filter((p) => p.name.toLowerCase().includes((f.teamQuery || '').trim().toLowerCase())).slice(0, 40)
        .map((p) => ({ name: p.name, initials: ini(p.name), color: p.color, selected: f.team.includes(p.name), toggle: () => this.setState((st) => {
          const team = st.projectForm.team.includes(p.name) ? st.projectForm.team.filter((n) => n !== p.name) : st.projectForm.team.concat([p.name]);
          return { projectForm: { ...st.projectForm, team, err: '' } };
        }) })),
      onMethod: setField('method'), onName: setField('name'), onGoal: setField('goal'), onOwner: setField('owner'),
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

        const owner = ownerOptions.find((p) => p.name === f.owner) || { color: '#9DB0A6' };
        this.setState({ projectForm: null });
        projectsW.create(this, this.srv, { name, goal, method: f.method, owner: f.owner, team: Array.from(new Set([f.owner, ...f.team])), start: f.start, due: f.due, budget }, owner.color)
          .then((r) => this.notify(r.approval ? 'Project "' + name + '" sent for approval. It opens in Projects once approved.' : 'Project "' + name + '" created.', r.approval ? '#FFB86B' : '#B8F36B'))
          .catch((e) => this.notify(e.message || 'Could not create the project.', '#FF9B8A'));
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
      nextMilestone: c.next ? c.next.name : p.milestones.length ? 'All done' : 'None yet', dueLabel: longDate(p.due), dueShort: shortDate(p.due),
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
    const openSprints = spl.filter((sp) => !sp.milestoneClosed && !sp.closed);
    const sprintSel = s.sprintFilter;
    const openTask = (ti) => this.setState({ taskOpen: { pid: pid, ti: ti } });
    const openMsClose = (m) => this.setState({ msDialog: { mode: 'close', pid: pid, mid: m.id, comment: '', err: '' } });
    const openMsReopen = (m) => this.setState({ msDialog: { mode: 'reopen', pid: pid, mid: m.id, comment: '', err: '' } });
    const openSprint = (milestone) => {
      if (!cur_p.milestones.some((m) => !m.closed)) { this.notify('Add a milestone first. A sprint sits under a milestone.', '#FFB86B'); return; }
      const d = sprintDefaults(cur_p, milestone);
      this.setState({ msDialog: { mode: 'sprint', pid: pid, lockProject: true, milestone: milestone || '', name: nextSprintName(cur_p), start: d.start, end: d.end, err: '' } });
    };
    const openTaskForm = (sprintId) => {
      const pick = openSprints.find((sp) => sp.id === sprintId) || (sprintId === null ? null : openSprints.find((sp) => sp.status === 'current') || openSprints.find((sp) => sp.status !== 'done') || openSprints[0]);
      const phase = tl.find((m) => m.status !== 'done' && !m.closed);
      this.setState({ taskForm: { title: '', desc: '', objective: '', kind: 'task', points: '', parent: '', deps: [], sprint: pick && cur_p.method !== 'waterfall' ? pick.id : '', milestone: phase ? phase.id : '', owner: cur_p.owner, due: pick && cur_p.method !== 'waterfall' ? defaultTaskDue(pick, today) : cur_p.due, dueTouched: false, err: '' } });
    };
    tools = { tl, spl, openTask, openTaskFormBacklog: () => openTaskForm(null), openMsClose, openMsReopen, openSprint, openTaskForm, longDate, shortDate };
    pj = {
      ...cur_p, initials: initialsOf(cur_p.name),
      statusStyle: badge(c.sc), progress: c.progress, doneCount: c.done, taskCount: c.total,
      barStyle: 'height:100%;width:' + Math.max(3, c.progress) + '%;background:' + c.sc + ';border-radius:3px',
      daysLeft, startLabel: longDate(cur_p.start), dueLabel: longDate(cur_p.due),
      msCount: tl.length, msDone: tl.filter((m) => m.status === 'done').length, spCount: spl.length,
      canMilestones: canMs, tab: s.projTab, isMsTab: s.projTab === 'milestones', isSpTab: s.projTab === 'sprints',
      goMs: set({ projTab: 'milestones' }), goSp: set({ projTab: 'sprints' }),
      addMilestone: () => this.setState({ msDialog: { mode: 'new', pid: pid, lockProject: true, name: '', date: '', desc: '', err: '' } }),
      canSprints: can('projects', 'Plan sprints'), addSprint: () => openSprint(''),
      sprintOpts: [{ v: 'all', l: 'All open work' }].concat(spl.map((sp) => ({ v: sp.id, l: sp.name + ' · ' + shortDate(sp.start) + ' – ' + shortDate(sp.end) + (sp.status === 'current' ? ' (current)' : '') }))),
      sprintValue: sprintSel, setSprint: (e) => this.setState({ sprintFilter: e.target.value }),
      boardTitle: sprintSel === 'all' ? 'All open work' : (spl.find((sp) => sp.id === sprintSel) || { name: 'Sprint' }).name,
      boardNote: sprintSel === 'all' ? 'No sprint is chosen, so the board shows every task that is not closed yet. Pick a sprint in the Sprints tab to see only its tasks.' : 'Showing the tasks of this sprint. Clear the choice in the Sprints tab to see all open work.',
      boardAll: sprintSel === 'all', showAll: () => this.setState({ sprintFilter: 'all' }),
      manager: cur_p.owner, iAmManager: cur_p.owner === me,
      canNewTask: canEditT, openTaskForm: () => openTaskForm(''),
      columns: wfOf(cur_p).map((w) => {
        const name = w.name, color = CAT_COLORS[w.cat];
        const tasks = boardTasks(cur_p, sprintSel, tl.filter((m) => m.status === 'done').map((m) => m.id)).filter((x) => statusOf(cur_p, x.t) === w.key).map(({ t, ti }) => {
          const inf = taskInfo(t);
          const tcx = taskContext(cur_p, t);
          return {
            ctxLabel: tcx.sprint ? tcx.sprint.name + (tcx.milestone ? ' · ' + tcx.milestone.name : '') : 'No sprint', ctxTitle: t.objective ? 'Objective: ' + t.objective : 'No objective written',
            taskKey: t.key || '', kindLabel: t.kind && t.kind !== 'task' ? t.kind[0].toUpperCase() + t.kind.slice(1) : '', points: t.points || t.points === 0 ? t.points + ' pts' : '', blocked: t.col !== 3 && openDeps(cur_p, t).length > 0, blockedBy: openDeps(cur_p, t).map((x) => x.title).join(', '),
            title: t.title, primary: t.primary, secondary: t.secondary, initials: ini(t.primary), color: pColor[t.primary] || '#9DB0A6',
            hasSecondary: !!t.secondary, secInitials: t.secondary ? ini(t.secondary) : '', secColor: pColor[t.secondary] || '#9DB0A6',
            dueLabel: inf.label, dueStyle: tBadge(inf.color),
            hasLinks: has.tickets && (t.tickets || []).length > 0, linkLabel: (t.tickets || []).length + ' ticket' + ((t.tickets || []).length === 1 ? '' : 's'),
            cardStyle: 'padding:10px;border-radius:10px;background:var(--k-1D2A25,#1D2A25);border:1px solid ' + (inf.overdue ? '#FF9B8A' : '#2A3832') + ';display:flex;flex-direction:column;gap:8px',
            open: () => openTask(ti),
            canMove: canMoveT && mayMoveTask(cur_p, t, 1, me), moveLabel: stepOf(cur_p, t, 1) ? 'Move to ' + stepOf(cur_p, t, 1).name : '', move: () => this.askTaskMove(pid, ti, 1),
            canBack: canMoveT && mayMoveTask(cur_p, t, -1, me), backLabel: stepOf(cur_p, t, -1) ? 'Move back to ' + stepOf(cur_p, t, -1).name : '', back: () => this.askTaskMove(pid, ti, -1)
          };
        });
        return { name, color, tasks, count: tasks.length, wip: w.wip || 0, overWip: !!w.wip && tasks.length > w.wip };
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
          if (k === 'kind' && v === 'epic') nf.parent = '';
          if (k === 'sprint') { const nsp = openSprints.find((x) => x.id === v); if (nsp && !nf.dueTouched) nf.due = defaultTaskDue(nsp, today); }
          return { taskForm: nf };
        });
      };
      const waterfall = cur_p.method === 'waterfall';
      const canParent = cur_p.tasks.filter((x) => x.id);
      tform = { open: true, waterfall, kind: f.kind, points: f.points, parent: f.parent, deps: f.deps, onKind: setF('kind'), onPoints: setF('points'), onParent: setF('parent'), onMilestone: setF('milestone'), milestone: f.milestone,
        kindOpts: [['story', 'Story'], ['task', 'Task'], ['bug', 'Bug'], ['epic', 'Epic'], ['subtask', 'Subtask']].map(([v, l]) => ({ v, l })),
        parentOpts: canParent.filter((x) => x.kind !== 'subtask').map((x) => ({ v: x.id, l: (x.key ? x.key + ' · ' : '') + x.title })),
        phaseOpts: tl.filter((m) => !m.closed).map((m) => ({ v: m.id, l: m.name + ' · due ' + shortDate(m.iso) })),
        depOpts: canParent.map((x) => ({ v: x.id, l: (x.key ? x.key + ' · ' : '') + x.title })),
        onDeps: (e) => { const v = Array.from(e.target.selectedOptions, (o) => o.value); this.setState((st) => ({ taskForm: { ...st.taskForm, deps: v, err: '' } })); },
        title: f.title, desc: f.desc, objective: f.objective, sprint: f.sprint, owner: f.owner, due: f.due, err: f.err, hasErr: !!f.err,
        sprintOpts: openSprints.map((x) => ({ v: x.id, l: x.name + ' · ' + x.milestoneName + ' · ' + shortDate(x.start) + ' – ' + shortDate(x.end) })), noSprints: false,
        ownerOpts: cur_p.team.map((n) => ({ v: n, l: n })), minDue: sp ? sp.start : '', maxDue: sp ? sp.end : '',
        onTitle: setF('title'), onDesc: setF('desc'), onObjective: setF('objective'), onSprint: setF('sprint'), onOwner: setF('owner'), onDue: setF('due'),
        planLine: waterfall ? 'Sits in the chosen phase. The phase passes its gate once its tasks are finished.' : !sp ? 'No sprint: it waits in the backlog until it is planned into one.' : 'Tracked automatically: ' + sp.name + ' (' + sp.done + ' of ' + sp.total + ' done) › milestone "' + sp.milestoneName + '" › ' + cur_p.name,
        cancel: set({ taskForm: null }),
        addSprint: () => openSprint(''),
        submit: () => {
          const err = validateTask(cur_p, f);
          if (err) { this.setState({ taskForm: { ...f, err: waterfall && !cur_p.milestones.some((m) => !m.closed) ? 'Add a phase first. A task sits in a phase.' : err } }); return; }
          const title = f.title.trim();
          this.setState((st) => ({ taskForm: null, projects: st.projects.map((p) => (p.id === pid ? { ...p, tasks: p.tasks.concat([{ title: title, kind: f.kind || 'task', points: f.points === '' ? null : Number(f.points), parent: f.parent || '', deps: f.deps || [], milestone: waterfall ? f.milestone : '', status: wfOf(p)[0].key, desc: f.desc.trim(), objective: f.objective.trim(), sprint: waterfall ? '' : f.sprint, col: wfOf(p)[0].cat, primary: f.owner || p.owner, secondary: '', start: today, due: f.due, links: [] }]), activity: [{ text: 'You created "' + title + '"' + (sp ? ' in ' + sp.name + ' (' + sp.milestoneName + ')' : waterfall ? ' in its phase' : ' in the backlog'), time: 'Just now' }].concat(p.activity) } : p)) }));
          this.notify(sp ? 'Task "' + title + '" added to ' + sp.name + '. It now counts toward "' + sp.milestoneName + '".' : 'Task "' + title + '" added.', '#B8F36B');
        } };
    }
    pj.taskForm = tform;
    pj.viewOnly = !!cur_p.viewOnly;
    pj.plan = buildPlan.call(this, { s, can, me, today, cur_p, tools, shortDate, longDate, set, projectsW });
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
        if (mproj.method === 'waterfall' && this.srv) {
          this.setState({ msDialog: null });
          projectsW.passGate(this, this.srv, mproj.id, mtl.id, comment).then((r) => this.notify(r.approval ? 'Gate sign-off for "' + mtl.name + '" sent for approval.' : 'Phase "' + mtl.name + '" passed its gate.', '#B8F36B')).catch((e) => this.notify(e.message || 'Could not ask for sign-off.', '#FF9B8A'));
          return;
        }
        this.setState((st) => ({ msDialog: null, projects: st.projects.map((p) => (p.id === mproj.id ? logMs({ ...p, milestones: p.milestones.map((m) => (m.id === mtl.id ? { ...m, closed: true, closedOn: today, closedBy: 'Samwel Okoth Omulo', closedComment: comment, history: m.history.concat([{ when: today, who: 'Samwel Okoth Omulo', action: 'Reached', comment }]) } : m)) }, 'Milestone "' + mtl.name + '" reached: "' + comment + '"') : p)) }));
        this.notify('Milestone "' + mtl.name + '" marked as reached. Your comment was saved.', '#B8F36B');
      } else {
        this.setState((st) => ({ msDialog: null, projects: st.projects.map((p) => (p.id === mproj.id ? logMs({ ...p, milestones: p.milestones.map((m) => (m.id === mtl.id ? { ...m, closed: false, closedOn: '', closedBy: '', closedComment: '', history: m.history.concat([{ when: today, who: 'Samwel Okoth Omulo', action: 'Reopened', comment }]) } : m)) }, 'Milestone "' + mtl.name + '" reopened: "' + comment + '"') : p)) }));
        this.notify('Milestone "' + mtl.name + '" reopened.', '#FFB86B');
      }
    };
    const wfl = mproj && mproj.method === 'waterfall';
    const titles = { new: wfl ? 'New phase' : 'New milestone', sprint: 'New sprint', close: wfl ? 'Ask for gate sign-off' : 'Mark milestone as reached', reopen: wfl ? 'Reopen phase' : 'Reopen milestone' };
    const labels = { new: wfl ? 'Create phase' : 'Create milestone', sprint: 'Create sprint', close: wfl ? 'Send for sign-off' : 'Mark as reached', reopen: wfl ? 'Reopen phase' : 'Reopen milestone' };
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
    // Only the people the task is assigned to edit it: Edit opens the fields, Save writes them all at once, Cancel drops them.
    const mayEdit = mayEditTask(tt, me) || tp.owner === me || can('projects', 'Edit any task');
    const ed = mayEdit && s.taskEdit && s.taskEdit.pid === tp.id && s.taskEdit.ti === to.ti ? s.taskEdit : null;
    const setE = (k) => (e) => { const v = e.target.value; this.setState((st) => ({ taskEdit: { ...st.taskEdit, [k]: v, err: '' } })); };
    const draftSprint = ed ? tp.sprints.find((x) => x.id === ed.sprint) : null;
    const tcxView = ed && draftSprint ? { sprint: draftSprint, milestone: tp.milestones.find((m) => m.id === draftSprint.milestone) || null } : taskContext(tp, tt);
    const ticketOf = (ref) => s.tickets.find((x) => x.id === ref);
    const tickets = (ed ? ed.tickets : tt.tickets || []);
    // tickets this person is assigned to (or handles) that are not linked yet
    const myTickets = s.tickets.filter((x) => (x.assignees || []).indexOf(me) !== -1 || x.handler === me).filter((x) => tickets.indexOf(x.id) === -1);
    td = {
      project: tp.name, pcolor: tp.color, pinitials: initialsOf(tp.name),
      title: ed ? ed.title : tt.title, desc: ed ? ed.desc : tt.desc || 'No description yet.',
      canEditTask: mayEdit, editing: !!ed, hasEditErr: !!(ed && ed.err), editErr: ed ? ed.err : '',
      startEdit: () => this.setState({ taskEdit: { pid: tp.id, ti: to.ti, kind: tt.kind || 'task', points: tt.points === null || tt.points === undefined ? '' : tt.points, parent: tt.parent || '', deps: (tt.deps || []).slice(), milestone: tt.milestone || '', title: tt.title, desc: tt.desc || '', objective: tt.objective || '', sprint: tt.sprint, due: tt.due, secondary: tt.secondary || '', tickets: (tt.tickets || []).slice(), err: '' } }),
      cancelEdit: () => this.setState({ taskEdit: null }),
      onTitle: setE('title'), onDesc: setE('desc'),
      saveEdit: () => {
        const f = this.state.taskEdit;
        if (!f) return;
        const err = validateTask(tp, { title: f.title, objective: f.objective, sprint: f.sprint, due: f.due, kind: f.kind, parent: f.parent, points: f.points, milestone: f.milestone });
        if (err) { this.setState({ taskEdit: { ...f, err } }); return; }
        updT((x) => ({ ...x, kind: f.kind, points: f.points === '' ? null : Number(f.points), parent: f.parent, deps: f.deps, milestone: f.milestone, title: f.title.trim(), desc: f.desc.trim(), objective: f.objective.trim(), sprint: f.sprint, due: f.due, secondary: f.secondary, tickets: f.tickets }));
        this.setState((st) => ({ taskEdit: null, projects: st.projects.map((p) => (p.id === tp.id ? { ...p, activity: [{ text: 'You edited "' + f.title.trim() + '"', time: 'Just now' }].concat(p.activity) } : p)) }));
        this.notify('Task saved.', '#B8F36B');
      },
      canEdit: !!ed, objective: ed ? ed.objective : tt.objective, setObjective: setE('objective'),
      sprintLabel: tp.method === 'waterfall' ? 'Phase' : 'Sprint', taskKey: tt.key || '',
      sprintId: tp.method === 'waterfall' ? (ed ? ed.milestone : tt.milestone) : ed ? ed.sprint : tt.sprint,
      sprintOpts: tp.method === 'waterfall' ? tp.milestones.map((m) => ({ v: m.id, l: m.name + ' · due ' + shortDate(m.iso) })) : [{ v: '', l: 'Backlog (no sprint)' }].concat(spT.map((sp) => ({ v: sp.id, l: sp.name + ' · ' + sp.milestoneName + ' · ' + shortDate(sp.start) + ' – ' + shortDate(sp.end) }))),
      kind: ed ? ed.kind : tt.kind || 'task', points: ed ? ed.points : tt.points === null || tt.points === undefined ? '' : tt.points, parent: ed ? ed.parent : tt.parent || '', deps: ed ? ed.deps : tt.deps || [],
      kindOpts: [['story', 'Story'], ['task', 'Task'], ['bug', 'Bug'], ['epic', 'Epic'], ['subtask', 'Subtask']].map(([v, l]) => ({ v, l })),
      onKind: setE('kind'), onPoints: setE('points'), onParent: setE('parent'),
      onDeps: (e) => { const v = Array.from(e.target.selectedOptions, (o) => o.value); this.setState((st) => ({ taskEdit: { ...st.taskEdit, deps: v, err: '' } })); },
      taskOpts: tp.tasks.filter((x) => x.id && x.id !== tt.id).map((x) => ({ v: x.id, l: (x.key ? x.key + ' · ' : '') + x.title })),
      waitingOn: openDeps(tp, tt).map((x) => x.title), isBlocked: tt.col !== 3 && openDeps(tp, tt).length > 0,
      subtasks: tp.tasks.filter((x) => x.parent && x.parent === tt.id).map((x) => ({ title: x.title, key: x.key || '', status: statusEntry(tp, x).name, open: () => this.setState({ taskOpen: { pid: tp.id, ti: tp.tasks.indexOf(x) }, taskEdit: null }) })),
      canDeleteTask: can('projects', 'Delete tasks') && !!tt.id,
      deleteTask: () => {
        if (tp.tasks.some((x) => x.parent === tt.id || (x.deps || []).includes(tt.id))) { this.notify('Other tasks sit under this one or depend on it. Change them first.', '#FFB86B'); return; }
        if (!window.confirm('Delete "' + tt.title + '"?')) return;
        this.setState((st) => ({ taskOpen: null, taskEdit: null, taskMove: null, projects: st.projects.map((p) => (p.id === tp.id ? { ...p, tasks: p.tasks.filter((x) => x.id !== tt.id) } : p)) }));
      },
      showChanges: can('projects', 'View change history'),
      changes: (tt.changes || []).filter((c) => !['col', 'assignee', 'secondary_id', 'kind'].includes(c.field) || c.field === 'kind').slice().reverse().map((c) => {
        const names = { title: 'Title', desc: 'Description', objective: 'Objective', status: 'Status', sprint: 'Sprint', milestone: 'Phase', due: 'Due date', start: 'Start date', primary: 'Owner', secondary: 'Second owner', points: 'Story points', parent: 'Parent', deps: 'Depends on', kind: 'Type', tickets: 'Linked tickets', links: 'Links' };
        const nameOfId = (v) => { const x = tp.tasks.find((y) => y.id === v) || tp.sprints.find((y) => y.id === v) || tp.milestones.find((y) => y.id === v) || wfOf(tp).find((y) => y.key === v); return x ? x.title || x.name : v; };
        const show = (v) => { if (v === null || v === undefined || v === '') return 'empty'; if (c.field === 'desc' || c.field === 'objective') return v.length > 60 ? v.slice(0, 60) + '…' : v; if (c.field === 'deps') { try { return JSON.parse(v).map(nameOfId).join(', ') || 'none'; } catch (e) { return v; } } return nameOfId(v); };
        return { what: names[c.field] || c.field, line: show(c.old) + ' → ' + show(c.new), by: c.who, when: new Date(c.at).toLocaleString('en-GB', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }) };
      }),
      setSprint: (e) => {
        if (tp.method === 'waterfall') { const v = e.target.value; this.setState((st) => ({ taskEdit: { ...st.taskEdit, milestone: v, err: '' } })); return; }
        const sp = tp.sprints.find((x) => x.id === e.target.value);
        if (!sp) { this.setState((st) => ({ taskEdit: { ...st.taskEdit, sprint: '', err: '' } })); return; }
        this.setState((st) => ({ taskEdit: { ...st.taskEdit, sprint: sp.id, due: st.taskEdit.due < sp.start ? sp.start : st.taskEdit.due > sp.end ? sp.end : st.taskEdit.due, err: '' } }));
      },
      milestoneName: tcxView.milestone ? tcxView.milestone.name : 'None', milestoneDue: tcxView.milestone ? longDate(tcxView.milestone.iso) : '',
      milestoneState: tcxView.milestone ? ((milestoneTimeline(tp, today).find((m) => m.id === tcxView.milestone.id) || {}).status === 'done' ? 'Reached' : tcxView.milestone.iso < today ? 'Late' : 'Open') : '',
      sprintDates: tcxView.sprint ? shortDate(tcxView.sprint.start) + ' – ' + shortDate(tcxView.sprint.end) : '',
      dueISO: ed ? ed.due : tt.due, minDue: tcxView.sprint ? tcxView.sprint.start : tt.start, maxDue: tcxView.sprint ? tcxView.sprint.end : '',
      setDue: (e) => { const v = e.target.value; if (v) this.setState((st) => ({ taskEdit: { ...st.taskEdit, due: v, err: '' } })); },
      col: statusEntry(tp, tt).name, colStyle: 'height:36px;padding:0 14px;border-radius:18px;display:flex;align-items:center;font-size:13px;font-weight:700;background:' + CAT_COLORS[tt.col] + '22;color:' + CAT_COLORS[tt.col],
      canBack: canMoveT && mayMoveTask(tp, tt, -1, me), prevCol: stepOf(tp, tt, -1) ? stepOf(tp, tt, -1).name : '', back: () => this.askTaskMove(tp.id, to.ti, -1),
      canFwd: canMoveT && mayMoveTask(tp, tt, 1, me), nextCol: stepOf(tp, tt, 1) ? stepOf(tp, tt, 1).name : '', fwd: () => this.askTaskMove(tp.id, to.ti, 1),
      history: (tt.history || []).slice().reverse().map((h) => ({ line: h.from + ' → ' + h.to, by: h.by, when: new Date(h.at).toLocaleString('en-GB', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }), comment: h.comment })), hasHistory: (tt.history || []).length > 0,
      moving: !!(s.taskMove && s.taskMove.pid === tp.id && s.taskMove.ti === to.ti),
      moveTitle: s.taskMove ? 'Move to ' + ((typeof s.taskMove.dir === 'string' ? wfOf(tp).find((w) => w.key === s.taskMove.dir) : stepOf(tp, tt, s.taskMove.dir)) || {}).name : '',
      statusValue: statusOf(tp, tt), statusOpts: wfOf(tp).filter((w) => w.key === statusOf(tp, tt) || (canMoveT && mayMoveTask(tp, tt, w.key, me))).map((w) => ({ v: w.key, l: w.name })),
      onStatus: (e) => { const v = e.target.value; if (v !== statusOf(tp, tt)) this.askTaskMove(tp.id, to.ti, v); }, moveBack: !!(s.taskMove && (typeof s.taskMove.dir === 'number' ? s.taskMove.dir < 0 : (wfOf(tp).find((w) => w.key === s.taskMove.dir) || {}).cat < tt.col)),
      moveComment: s.taskMove ? s.taskMove.comment : '', moveErr: s.taskMove ? s.taskMove.err : '', hasMoveErr: !!(s.taskMove && s.taskMove.err),
      onMoveComment: (e) => this.setState((st) => ({ taskMove: { ...st.taskMove, comment: e.target.value, err: '' } })),
      confirmMove: () => this.confirmTaskMove(), cancelMove: () => this.setState({ taskMove: null }),
      overdue: inf.overdue, overDays: inf.overDays,
      reminderNote: 'An automatic reminder went to ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' today.',
      remind: () => { this.setState((st) => ({ taskReminders: { ...st.taskReminders, [rk]: (st.taskReminders[rk] || []).concat(['You reminded ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' just now']) } })); this.notify('Reminder sent to ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' about "' + tt.title + '".', '#FFB86B'); },
      startLabel: longDate(tt.start), dueLabel: longDate(tt.due), dueColor: inf.color,
      timeBar: 'height:100%;width:' + Math.max(3, tt.col === 3 ? 100 : inf.pct) + '%;background:' + inf.color + ';border-radius:4px',
      timeNote: tt.col === 3 ? 'Finished.' : inf.overdue ? inf.overDays + ' past the due date.' : inf.dd === 0 ? 'Due today.' : inf.dd + ' days left · ' + inf.pct + '% of the time used.',
      primary: tt.primary, pIni: ini(tt.primary), pColor: pColor[tt.primary] || '#9DB0A6',
      secondary: ed ? ed.secondary : tt.secondary || '',
      teamOpts: tp.team.filter((n) => n !== tt.primary).map((n) => ({ name: n })),
      setSecondary: (e) => { const v = e.target.value; this.setState((st) => ({ taskEdit: { ...st.taskEdit, secondary: v, err: '' } })); },
      // linked tickets only exist when the organization has the Tickets widget on
      showTickets: !!has.tickets,
      tickets: tickets.map((ref) => { const x = ticketOf(ref); return { ref, title: x ? x.title : 'A ticket you cannot see', meta: x ? ref + ' · ' + (x.cat || 'Ticket') : ref, known: !!x,
        open: () => { if (x) this.setState({ taskOpen: null, taskEdit: null, section: 'tickets', projectOpen: null, ticketOpen: ref, ticketFrom: null, tagOpen: false, tcomment: '' }); },
        unlink: () => this.setState((st) => ({ taskEdit: { ...st.taskEdit, tickets: st.taskEdit.tickets.filter((y) => y !== ref), err: '' } })) }; }),
      noTickets: tickets.length === 0,
      ticketOpts: myTickets.map((x) => ({ v: x.id, l: x.id + ' · ' + x.title })),
      noMyTickets: myTickets.length === 0,
      addTicket: (e) => { const v = e.target.value; if (v === '') return; this.setState((st) => ({ taskEdit: { ...st.taskEdit, tickets: st.taskEdit.tickets.concat([v]), err: '' } })); },
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
