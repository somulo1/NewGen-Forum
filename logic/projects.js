// Projects domain rules: milestone > objective > task, with sprints derived from dates.
// Nothing here stores progress or sprint numbers. They are always calculated from the tasks.
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

// Gives every milestone an id and a real date, creates an objective per milestone when none exist,
// and files each task under the objective of the first milestone due on or after the task.
export function normalizeProject(p) {
  const year = (p.start || '2026').slice(0, 4);
  const milestones = p.milestones.map((m, i) => {
    const closed = m.closed !== undefined ? m.closed : m.state === 'done';
    const iso = m.iso || shortToIso(m.date, year);
    return { id: m.id || p.id + '-m' + (i + 1), name: m.name, iso, desc: m.desc || '', closed,
      closedOn: closed ? (m.closedOn || iso) : '', closedBy: closed ? (m.closedBy || p.owner) : '', closedComment: closed ? (m.closedComment || 'Reached as planned.') : '',
      history: m.history || [] };
  });
  const objectives = p.objectives && p.objectives.length ? p.objectives : milestones.map((m, i) => ({ id: p.id + '-o' + (i + 1), title: m.name + ' deliverables', milestone: m.id }));
  const ordered = milestones.slice().sort(byDate);
  const tasks = p.tasks.map((t) => {
    if (t.objective) return t;
    const ms = ordered.find((m) => m.iso >= t.due) || ordered[ordered.length - 1];
    const obj = objectives.find((o) => o.milestone === ms.id);
    return { ...t, objective: obj ? obj.id : '' };
  });
  return { ...p, milestones, objectives, tasks };
}

export function sprintOf(project, iso) {
  const n = Math.max(1, Math.floor((isoMs(iso) - isoMs(project.start)) / (SPRINT_DAYS * DAY)) + 1);
  const start = addDaysIso(project.start, (n - 1) * SPRINT_DAYS);
  return { n, start, end: addDaysIso(start, SPRINT_DAYS - 1), label: 'Sprint ' + n };
}

export function sprintList(project) {
  const last = sprintOf(project, project.due).n;
  const out = [];
  for (let n = 1; n <= last; n++) out.push(sprintOf(project, addDaysIso(project.start, (n - 1) * SPRINT_DAYS)));
  return out;
}

export function objectiveStats(project, o) {
  const tasks = project.tasks.filter((t) => t.objective === o.id);
  const done = tasks.filter((t) => t.col === 3).length;
  return { ...o, total: tasks.length, done, progress: tasks.length ? Math.round((done / tasks.length) * 100) : 0 };
}

// Progress is the share of finished tasks across the milestone's objectives.
// A milestone is "ready" when it has tasks and all of them are done; it only becomes "done" when someone closes it with a comment.
export function milestoneTimeline(project, today) {
  const sorted = project.milestones.slice().sort(byDate);
  const firstOpen = sorted.find((m) => !m.closed);
  return sorted.map((m) => {
    const objectives = project.objectives.filter((o) => o.milestone === m.id).map((o) => objectiveStats(project, o));
    const total = objectives.reduce((a, o) => a + o.total, 0);
    const done = objectives.reduce((a, o) => a + o.done, 0);
    const ready = !m.closed && total > 0 && done === total;
    const progress = m.closed ? 100 : total ? Math.round((done / total) * 100) : 0;
    const status = m.closed ? 'done' : m.iso < today ? 'late' : firstOpen && firstOpen.id === m.id ? 'next' : 'upcoming';
    return { ...m, objectives, total, done, progress, ready, status, openTasks: total - done };
  });
}

export function taskContext(project, task, today) {
  const objective = project.objectives.find((o) => o.id === task.objective) || null;
  const milestone = objective ? project.milestones.find((m) => m.id === objective.milestone) || null : null;
  return { objective, milestone, sprint: sprintOf(project, task.due), currentSprint: sprintOf(project, today) };
}

// New tasks default to the end of the current sprint, but never later than their milestone.
export function defaultDue(project, milestone, today) {
  const end = sprintOf(project, today).end;
  const capped = milestone && milestone.iso < end ? milestone.iso : end;
  return capped < today ? today : capped;
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

export function validateTask(project, f) {
  if (!f.title.trim()) return 'Give the task a title.';
  if (!f.objective) return 'Choose the objective this task serves.';
  const obj = project.objectives.find((o) => o.id === f.objective);
  const ms = obj && project.milestones.find((m) => m.id === obj.milestone);
  if (ms && ms.closed) return 'That milestone is already reached. Pick an objective from an open milestone.';
  if (!f.due) return 'Pick a due date.';
  return '';
}
