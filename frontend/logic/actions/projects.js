import { milestoneTimeline, todayIso as isoToday } from '~/logic/projects.js';
// HubLogic actions: projects. Mixed into HubLogic.prototype, so `this` is the HubLogic instance.
export const projectsActions = {
  // Changing a task's status always needs a comment saying why (and, going back, what is still missing). This opens the task
  // with the comment box; confirmTaskMove() moves it once the comment is written.
  askTaskMove(pid, ti, dir) {
    this.setState({ taskOpen: { pid, ti }, taskMove: { pid, ti, dir, comment: '', err: '' } });
  },
  confirmTaskMove() {
    const m = this.state.taskMove;
    if (!m) return;
    const comment = (m.comment || '').trim();
    if (comment.length < 8) { this.setState({ taskMove: { ...m, err: 'Add a comment of at least a few words. It stays with the task.' } }); return; }
    this.moveTask(m.pid, m.ti, m.dir, comment);
    this.setState({ taskMove: null });
  },

  moveTask(pid, ti, dir, comment) {
    const names = ['To do', 'In progress', 'Review', 'Done'];
    const today = isoToday(Date.now());
    const readyIds = () => milestoneTimeline(this.state.projects.find((p) => p.id === pid), today).filter((m) => m.ready).map((m) => m.id);
    const before = readyIds();
    this.moveTaskRaw(pid, ti, dir, names, comment);
    const proj = this.state.projects.find((p) => p.id === pid);
    milestoneTimeline(proj, today).filter((m) => m.ready && before.indexOf(m.id) === -1)
      .forEach((m) => this.notify('Milestone "' + m.name + '" in ' + proj.name + ' has all its tasks done. Close it with a comment.', '#B8F36B'));
  },

  moveTaskRaw(pid, ti, dir, names, comment) {
    const who = this.srv ? this.srv.me.name : 'You';
    this.setState((st) => ({ projects: st.projects.map((p) => {
      if (p.id !== pid) return p;
      const t = p.tasks[ti];
      const nc = Math.max(0, Math.min(3, t.col + dir));
      if (nc === t.col) return p;
      const note = { from: names[t.col], to: names[nc], comment: comment || '', by: who, at: new Date().toISOString() };
      return { ...p, tasks: p.tasks.map((x, j) => (j === ti ? { ...x, col: nc, history: (x.history || []).concat([note]) } : x)), activity: [{ text: 'You moved "' + t.title + '" ' + (dir > 0 ? 'forward' : 'back') + ' to ' + names[nc] + (comment ? ': ' + comment : ''), time: 'Just now' }].concat(p.activity) };
    }) }));
  }
};
