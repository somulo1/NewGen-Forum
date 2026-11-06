import * as projectsW from '~/logic/server/widgets/projects.js';
// renderVals step: projectTables. Milestone and sprint tables for the open project, built from the project's own tasks.
export function buildProjectTables(ctx) {
  const { C, can, canEditT, canMs, cur_p, pj, projTools, s } = ctx;
  if (!cur_p) return {};
  const { tl, spl, openMsClose, openMsReopen, openSprint, openTaskForm, longDate, shortDate } = projTools;
  const msColor = { done: '#B8F36B', ready: '#B8F36B', late: '#FF9B8A', next: '#FFB86B', upcoming: '#9DB0A6' };
  const msLabel = { done: 'Reached', ready: 'Reached', late: 'Late', next: 'Next up', upcoming: 'Upcoming' };
  const spColor = { closed: '#6F8178', done: '#B8F36B', current: '#7CC4FF', late: '#FF9B8A', upcoming: '#9DB0A6', ended: '#6F8178' };
  const spLabel = { closed: 'Closed', done: 'Done', current: 'Current', late: 'Late', upcoming: 'Upcoming', ended: 'Ended' };
  const tasksLabel = (done, total) => (total ? done + ' / ' + total + ' tasks' : 'No tasks yet');

  pj.msTable = { id: 'ms-' + cur_p.id, pageSize: 6, minWidth: '760px', emptyText: cur_p.method === 'waterfall' ? 'No phases yet. Add one to start planning.' : 'No milestones yet. Add one to start planning.',
    cols: [{ label: cur_p.method === 'waterfall' ? 'Phase' : 'Milestone' }, { label: cur_p.method === 'waterfall' ? 'Gate' : 'Due' }, { label: 'Sprints' }, { label: 'Progress' }, { label: 'Status' }, { label: '', align: 'right' }],
    rows: tl.map((m) => {
      const key = m.status;   // reached by itself once every task is done, open again if one is not
      const sub = m.closed ? 'Reached ' + longDate(m.closedOn) + ' by ' + m.closedBy + ': "' + m.closedComment + '"'
        : m.ready ? 'Reached automatically: all tasks are done.'
        : m.total ? m.openTasks + (m.openTasks === 1 ? ' task is' : ' tasks are') + ' still open' : 'Add a sprint, then tasks, to track progress';
      const actions = [];
      if (canMs && m.closed) actions.push({ label: 'Reopen', run: () => openMsReopen(m) });
      if (can('projects', 'Plan sprints') && !m.closed && cur_p.method !== 'waterfall') actions.push({ label: 'Add sprint', run: () => openSprint(m.id) });
      if (!m.closed && cur_p.method === 'waterfall') can('projects', 'Request gate sign-off') && actions.push({ label: 'Gate sign-off', run: () => openMsClose(m) });
      if (canMs && !m.closed && !m.sprints.length && !cur_p.tasks.some((t) => t.milestone === m.id)) actions.push({ label: 'Delete', run: () => { if (window.confirm('Delete "' + m.name + '"?')) this.setState((st) => ({ projects: st.projects.map((x) => (x.id === cur_p.id ? { ...x, milestones: x.milestones.filter((y) => y.id !== m.id) } : x)) })); } });
      return { tint: m.status === 'done' ? '#B8F36B' : m.status === 'late' ? '#FF9B8A' : '', cells: [
        C.t(m.name, { strong: true, sub }), C.t(longDate(m.iso), { nowrap: true, fv: m.iso }),
        C.t(m.sprints.length ? m.sprints.map((sp) => sp.name).join(', ') : 'None yet'),
        C.bar(m.progress, msColor[key], tasksLabel(m.done, m.total)), C.b(msLabel[key], msColor[key], { fv: msLabel[key] }), C.a(actions)] };
    }) };

  pj.spTable = { id: 'sp-' + cur_p.id, pageSize: 6, minWidth: '760px', emptyText: 'No sprints yet. Add a sprint under a milestone.',
    cols: [{ label: 'Sprint' }, { label: 'Milestone' }, { label: 'Dates' }, { label: 'Progress' }, { label: 'Status' }, { label: '', align: 'right' }],
    rows: spl.map((sp) => ({ tint: s.sprintFilter === sp.id ? '#B8F36B' : sp.status === 'current' ? '#7CC4FF' : sp.status === 'late' ? '#FF9B8A' : '', cells: [
      C.t(sp.name, { strong: true, sub: s.sprintFilter === sp.id ? 'Shown on the task board' : sp.milestoneClosed ? 'Milestone reached' : '' }), C.t(sp.milestoneName),
      C.t(shortDate(sp.start) + ' – ' + shortDate(sp.end), { nowrap: true, fv: sp.start }),
      C.bar(sp.progress, spColor[sp.status], tasksLabel(sp.done, sp.total)), C.b(spLabel[sp.status], spColor[sp.status], { fv: spLabel[sp.status] }),
      C.a((can('projects', 'Plan sprints') && !sp.closed ? [{ label: 'Delete', run: () => { if (window.confirm('Delete ' + sp.name + '? Its tasks go back to the backlog.')) this.setState((st) => ({ sprintFilter: st.sprintFilter === sp.id ? 'all' : st.sprintFilter, projects: st.projects.map((x) => (x.id === cur_p.id ? { ...x, sprints: x.sprints.filter((y) => y.id !== sp.id), tasks: x.tasks.map((t) => (t.sprint === sp.id ? { ...t, sprint: '' } : t)) } : x)) })); } }] : []).concat(can('projects', 'Close sprints') && !sp.closed && this.srv ? [{ label: 'Close sprint', run: () => projectsW.closeSprint(this, this.srv, cur_p.id, sp.id).then(() => this.notify(sp.name + ' closed. Unfinished work moved on.', '#B8F36B')).catch((e) => this.notify(e.message || 'Could not close the sprint.', '#FF9B8A')) }] : []).concat([{ label: s.sprintFilter === sp.id ? 'Show all open work' : 'Show on board', kind: s.sprintFilter === sp.id ? 'primary' : undefined, run: () => this.setState({ sprintFilter: s.sprintFilter === sp.id ? 'all' : sp.id }) }].concat(canEditT && !sp.milestoneClosed && !sp.closed ? [{ label: 'Add task', run: () => openTaskForm(sp.id) }] : [])))] })) };
  return {};
}
