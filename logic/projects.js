// Projects domain rules: project > milestone > sprint > task.
// A task sits in one sprint, a sprint belongs to one milestone, a milestone belongs to one project.
// Progress, status and counts are never stored. They are always calculated from the tasks.
export const SPRINT_DAYS = 14;
const DAY = 86400000;
const MONTHS = { Jan: 0, Feb: 1, Mar: 2, Apr: 3, May: 4, Jun: 5, Jul: 6, Aug: 7, Sep: 8, Oct: 9, Nov: 10, Dec: 11 };
const tz = { timeZone: 'Africa/Nairobi' };

export const isoMs = (iso) => new Date(iso + 'T00:00:00+03:00').getTime();
export const addDaysIso = (iso, d) => new Date(isoMs(iso) + d * DAY).toLocaleDateString('en-CA', tz);
export const todayIso = (nowMs) => new Date(nowMs).toLocaleDateString('en-CA', tz);
export function shortToIso(short, year) {
  const m = String(short).trim().split(' ');
  const mo = MONTHS[m[1]];
  if (mo === undefined) return String(short);
  return year + '-' + String(mo + 1).padStart(2, '0') + '-' + String(Number(m[0])).padStart(2, '0');
}
const byDate = (a, b) => (a.iso < b.iso ? -1 : a.iso > b.iso ? 1 : 0);

// Sample projects only carry milestones and tasks. This gives every milestone an id and a real date,
// cuts the project into two-week sprints (each sprint belongs to the first milestone due on or after it),
// and puts each task in the sprint that contains its due date.
export function normalizeProject(p) {
  const year = (p.start || '2026').slice(0, 4);
  const milestones = p.milestones.map((m, i) => {
    const closed = m.closed !== undefined ? m.closed : m.state === 'done';
    const iso = m.iso || shortToIso(m.date, year);
    return { id: m.id || p.id + '-m' + (i + 1), name: m.name, iso, desc: m.desc || '', closed,
      closedOn: closed ? (m.closedOn || iso) : '', closedBy: closed ? (m.closedBy || p.owner) : '', closedComment: closed ? (m.closedComment || 'Reached as planned.') : '',
      history: m.history || [] };
  });
  const ordered = milestones.slice().sort(byDate);
  let sprints = p.sprints;
  if (!sprints || !sprints.length) {
    sprints = [];
    for (let n = 1, start = p.start; start <= p.due; n++, start = addDaysIso(start, SPRINT_DAYS)) {
      const end = addDaysIso(start, SPRINT_DAYS - 1) > p.due ? p.due : addDaysIso(start, SPRINT_DAYS - 1);
      const ms = ordered.find((m) => m.iso >= end) || ordered[ordered.length - 1];
      sprints.push({ id: p.id + '-s' + n, name: 'Sprint ' + n, start, end, milestone: ms.id });
    }
  }
  const tasks = p.tasks.map((t) => ({ ...t, objective: t.objective || '', sprint: t.sprint || sprintFor(sprints, t.due).id }));
  return { ...p, milestones, sprints, tasks };
}

// The sprint that contains a date. Dates outside every sprint fall back to the nearest one.
export function sprintFor(sprints, iso) {
  const sorted = sprints.slice().sort((a, b) => (a.start < b.start ? -1 : 1));
  return sorted.find((s) => s.start <= iso && iso <= s.end) || sorted.filter((s) => s.start <= iso).pop() || sorted[0];
}

export function sprintOf(project, iso) {
  const sp = sprintFor(project.sprints, iso);
  const n = project.sprints.findIndex((s) => s.id === sp.id) + 1;
  return { ...sp, n, label: sp.name };
}

export function sprintList(project) {
  return project.sprints.map((sp, i) => ({ ...sp, n: i + 1, label: sp.name }));
}

export function nextSprintName(project) {
  return 'Sprint ' + (project.sprints.length + 1);
}

// A new sprint starts the day after this milestone's last sprint, or the day after the previous milestone, and lasts two weeks at most.
export function sprintDefaults(project, milestoneId) {
  const ms = project.milestones.find((m) => m.id === milestoneId);
  if (!ms) return { start: '', end: '' };
  const own = project.sprints.filter((sp) => sp.milestone === ms.id).map((sp) => sp.end).sort().pop();
  const before = project.milestones.filter((m) => m.iso < ms.iso).map((m) => m.iso).sort().pop();
  let start = own ? addDaysIso(own, 1) : before ? addDaysIso(before, 1) : project.start;
  if (start > ms.iso) start = ms.iso;
  const end = addDaysIso(start, SPRINT_DAYS - 1);
  return { start, end: end > ms.iso ? ms.iso : end };
}

export function sprintTimeline(project, today) {
  return project.sprints.slice().sort((a, b) => (a.start < b.start ? -1 : 1)).map((sp) => {
    const tasks = project.tasks.filter((t) => t.sprint === sp.id);
    const done = tasks.filter((t) => t.col === 3).length;
    const total = tasks.length;
    const status = total > 0 && done === total ? 'done' : sp.end < today ? (total ? 'late' : 'ended') : sp.start <= today ? 'current' : 'upcoming';
    const ms = project.milestones.find((m) => m.id === sp.milestone) || null;
    return { ...sp, total, done, openTasks: total - done, progress: total ? Math.round((done / total) * 100) : 0, status, milestoneName: ms ? ms.name : '', milestoneClosed: !!(ms && ms.closed) };
  });
}

// Milestone progress is the share of finished tasks across its sprints.
// A milestone is "ready" when it has tasks and all of them are done; it only becomes "done" when someone closes it with a comment.
export function milestoneTimeline(project, today) {
  const sprints = sprintTimeline(project, today);
  const sorted = project.milestones.slice().sort(byDate);
  const firstOpen = sorted.find((m) => !m.closed);
  return sorted.map((m) => {
    const own = sprints.filter((sp) => sp.milestone === m.id);
    const total = own.reduce((a, sp) => a + sp.total, 0);
    const done = own.reduce((a, sp) => a + sp.done, 0);
    const ready = !m.closed && total > 0 && done === total;
    const progress = m.closed ? 100 : total ? Math.round((done / total) * 100) : 0;
    const status = m.closed ? 'done' : m.iso < today ? 'late' : firstOpen && firstOpen.id === m.id ? 'next' : 'upcoming';
    return { ...m, sprints: own, total, done, ready, progress, status, openTasks: total - done };
  });
}

export function taskContext(project, task) {
  const sprint = project.sprints.find((s) => s.id === task.sprint) || null;
  const milestone = sprint ? project.milestones.find((m) => m.id === sprint.milestone) || null : null;
  return { sprint, milestone };
}

// A new task is due at the end of its sprint, or today if the sprint is already running late.
export function defaultTaskDue(sprint, today) {
  if (!sprint) return '';
  return sprint.end < today ? today : sprint.end;
}

export function validateMilestone(project, f) {
  if (!project) return 'Choose the project this milestone belongs to.';
  if (!f.name.trim()) return 'Give the milestone a name.';
  if (project.milestones.some((m) => m.name.toLowerCase() === f.name.trim().toLowerCase())) return 'This project already has a milestone with that name.';
  if (!f.date) return 'Pick the date this milestone is due.';
  if (f.date < project.start) return 'The milestone cannot be due before the project starts (' + project.start + ').';
  if (f.date > project.due) return 'The milestone cannot be due after the project ends (' + project.due + ').';
  return '';
}

export function validateSprint(project, f) {
  const ms = project && project.milestones.find((m) => m.id === f.milestone);
  if (!ms) return 'Choose the milestone this sprint belongs to.';
  if (ms.closed) return 'That milestone is already reached. Pick an open milestone.';
  if (!f.name.trim()) return 'Give the sprint a name.';
  if (project.sprints.some((s) => s.name.toLowerCase() === f.name.trim().toLowerCase())) return 'This project already has a sprint with that name.';
  if (!f.start || !f.end) return 'Pick the start and end dates.';
  if (f.end < f.start) return 'The sprint cannot end before it starts.';
  if (f.start < project.start || f.end > project.due) return 'The sprint must fall inside the project dates (' + project.start + ' to ' + project.due + ').';
  if (f.end > ms.iso) return 'The sprint must end on or before its milestone date (' + ms.iso + ').';
  return '';
}

export function validateTask(project, f) {
  if (!f.title.trim()) return 'Give the task a title.';
  if (!f.objective.trim()) return 'Write the objective this task serves.';
  const sp = project.sprints.find((s) => s.id === f.sprint);
  if (!sp) return 'Choose the sprint this task belongs to.';
  const ms = project.milestones.find((m) => m.id === sp.milestone);
  if (ms && ms.closed) return 'That sprint belongs to a milestone that is already reached. Pick another sprint.';
  if (!f.due) return 'Pick a due date.';
  if (f.due < sp.start || f.due > sp.end) return 'The due date must fall inside the sprint (' + sp.start + ' to ' + sp.end + ').';
  return '';
}
