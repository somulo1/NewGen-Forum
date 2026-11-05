import { mayMoveTask, milestoneTimeline, openDeps, statusEntry, stepOf, todayIso as isoToday } from '~/logic/projects.js';
// HubLogic actions: projects. Mixed into HubLogic.prototype, so `this` is the HubLogic instance.
export const projectsActions = {
  // Changing a task's status always needs a comment saying why (and, going back, what is still missing). This opens the task
  // with the comment box; confirmTaskMove() moves it once the comment is written.
  askTaskMove(pid, ti, dir) {
    if (!this.canMoveTask(pid, ti, dir)) return;
    this.setState({ taskOpen: { pid, ti }, taskMove: { pid, ti, dir, comment: '', err: '' } });
  },
  // The project manager moves tasks into and out of Done; the people a task is assigned to move it between the other columns.
  canMoveTask(pid, ti, dir) {
    const p = this.state.projects.find((x) => x.id === pid);
    const t = p && p.tasks[ti];
    const me = this.srv ? this.srv.me.name : 'Samwel Okoth Omulo';
    if (t && mayMoveTask(p, t, dir, me)) return true;
    this.notify(t && (t.col === 3 || (stepOf(p, t, dir) || {}).cat === 3) ? 'Only the project manager (' + ((p && p.owner) || 'nobody yet') + ') can move a task to Done or back from Done.' : 'Only the people this task is assigned to, or the project manager, can move it.', '#FFB86B');
    return false;
  },
  confirmTaskMove() {
    const m = this.state.taskMove;
    if (!m || !this.canMoveTask(m.pid, m.ti, m.dir)) { this.setState({ taskMove: null }); return; }
    const comment = (m.comment || '').trim();
    if (comment.length < 8) { this.setState({ taskMove: { ...m, err: 'Add a comment of at least a few words. It stays with the task.' } }); return; }
    const pr = this.state.projects.find((x) => x.id === m.pid);
    const tk = pr && pr.tasks[m.ti];
    const wait = tk && (stepOf(pr, tk, m.dir) || {}).cat === 3 ? openDeps(pr, tk) : [];
    if (wait.length) { this.setState({ taskMove: { ...m, err: 'Blocked: "' + wait[0].title + '" has to finish first.' } }); return; }
    this.moveTask(m.pid, m.ti, m.dir, comment);
    this.setState({ taskMove: null });
  },

  moveTask(pid, ti, dir, comment) {
    const today = isoToday(Date.now());
    const readyIds = () => milestoneTimeline(this.state.projects.find((p) => p.id === pid), today).filter((m) => m.ready).map((m) => m.id);
    const before = readyIds();
    this.moveTaskRaw(pid, ti, dir, comment);
    const proj = this.state.projects.find((p) => p.id === pid);
    milestoneTimeline(proj, today).filter((m) => m.ready && before.indexOf(m.id) === -1)
      .forEach((m) => this.notify('Milestone "' + m.name + '" in ' + proj.name + ' is reached: all its tasks are done.', '#B8F36B'));
  },

  moveTaskRaw(pid, ti, dir, comment) {
    const who = this.srv ? this.srv.me.name : 'You';
    this.setState((st) => ({ projects: st.projects.map((p) => {
      if (p.id !== pid) return p;
      const t = p.tasks[ti];
      const to = stepOf(p, t, dir);
      if (!to) return p;
      const note = { from: statusEntry(p, t).name, to: to.name, comment: comment || '', by: who, at: new Date().toISOString() };
      return { ...p, tasks: p.tasks.map((x, j) => (j === ti ? { ...x, col: to.cat, status: to.key, history: (x.history || []).concat([note]) } : x)), activity: [{ text: 'You moved "' + t.title + '" ' + (dir > 0 ? 'forward' : 'back') + ' to ' + to.name + (comment ? ': ' + comment : ''), time: 'Just now' }].concat(p.activity) };
    }) }));
  }
};
