// renderVals step: projectTables. Milestone and sprint tables for the open project, built from the project's own tasks.
export function buildProjectTables(ctx) {
  const { C, canEditT, canMs, cur_p, pj, projTools } = ctx;
  if (!cur_p) return {};
  const { tl, spl, openMsClose, openMsReopen, openSprint, openTaskForm, longDate, shortDate } = projTools;
  const msColor = { done: '#B8F36B', ready: '#B8F36B', late: '#FF9B8A', next: '#FFB86B', upcoming: '#9DB0A6' };
  const msLabel = { done: 'Reached', ready: 'Ready to close', late: 'Late', next: 'Next up', upcoming: 'Upcoming' };
  const spColor = { done: '#B8F36B', current: '#7CC4FF', late: '#FF9B8A', upcoming: '#9DB0A6', ended: '#6F8178' };
  const spLabel = { done: 'Done', current: 'Current', late: 'Late', upcoming: 'Upcoming', ended: 'Ended' };
  const tasksLabel = (done, total) => (total ? done + ' / ' + total + ' tasks' : 'No tasks yet');

  pj.msTable = { id: 'ms-' + cur_p.id, pageSize: 6, minWidth: '760px', emptyText: 'No milestones yet. Add one to start planning.',
    cols: [{ label: 'Milestone' }, { label: 'Due' }, { label: 'Sprints' }, { label: 'Progress' }, { label: 'Status' }, { label: '', align: 'right' }],
    rows: tl.map((m) => {
      const key = m.ready ? 'ready' : m.status;
      const sub = m.closed ? 'Reached ' + longDate(m.closedOn) + ' by ' + m.closedBy + ': "' + m.closedComment + '"'
        : m.ready ? 'All tasks are done. Close it with a comment.'
        : m.total ? m.openTasks + (m.openTasks === 1 ? ' task is' : ' tasks are') + ' still open' : 'Add a sprint, then tasks, to track progress';
      const actions = [];
      if (canMs && m.ready) actions.push({ label: 'Mark as reached', kind: 'primary', run: () => openMsClose(m) });
      if (canMs && m.closed) actions.push({ label: 'Reopen', run: () => openMsReopen(m) });
      if (canMs && !m.closed) actions.push({ label: 'Add sprint', run: () => openSprint(m.id) });
      return { tint: m.ready ? '#B8F36B' : m.status === 'late' ? '#FF9B8A' : '', cells: [
        C.t(m.name, { strong: true, sub }), C.t(longDate(m.iso), { nowrap: true, fv: m.iso }),
        C.t(m.sprints.length ? m.sprints.map((sp) => sp.name).join(', ') : 'None yet'),
        C.bar(m.progress, msColor[key], tasksLabel(m.done, m.total)), C.b(msLabel[key], msColor[key], { fv: msLabel[key] }), C.a(actions)] };
    }) };

  pj.spTable = { id: 'sp-' + cur_p.id, pageSize: 6, minWidth: '760px', emptyText: 'No sprints yet. Add a sprint under a milestone.',
    cols: [{ label: 'Sprint' }, { label: 'Milestone' }, { label: 'Dates' }, { label: 'Progress' }, { label: 'Status' }, { label: '', align: 'right' }],
    rows: spl.map((sp) => ({ tint: sp.status === 'current' ? '#7CC4FF' : sp.status === 'late' ? '#FF9B8A' : '', cells: [
      C.t(sp.name, { strong: true, sub: sp.milestoneClosed ? 'Milestone reached' : '' }), C.t(sp.milestoneName),
      C.t(shortDate(sp.start) + ' – ' + shortDate(sp.end), { nowrap: true, fv: sp.start }),
      C.bar(sp.progress, spColor[sp.status], tasksLabel(sp.done, sp.total)), C.b(spLabel[sp.status], spColor[sp.status], { fv: spLabel[sp.status] }),
      C.a(canEditT && !sp.milestoneClosed ? [{ label: 'Add task', run: () => openTaskForm(sp.id) }] : [])] })) };
  return {};
}
