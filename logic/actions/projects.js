import { milestoneTimeline, todayIso as isoToday } from '~/logic/projects.js';
// HubLogic actions: projects. Mixed into HubLogic.prototype, so `this` is the HubLogic instance.
export const projectsActions = {
  moveTask(pid, ti, dir) {
    const names = ['To do', 'In progress', 'Review', 'Done'];
    const today = isoToday(Date.now());
    const readyIds = () => milestoneTimeline(this.state.projects.find((p) => p.id === pid), today).filter((m) => m.ready).map((m) => m.id);
    const before = readyIds();
    this.moveTaskRaw(pid, ti, dir, names);
    const proj = this.state.projects.find((p) => p.id === pid);
    milestoneTimeline(proj, today).filter((m) => m.ready && before.indexOf(m.id) === -1)
      .forEach((m) => this.notify('Milestone "' + m.name + '" in ' + proj.name + ' has all its tasks done. Close it with a comment.', '#B8F36B'));
  },

  moveTaskRaw(pid, ti, dir, names) {
    this.setState((st) => ({ projects: st.projects.map((p) => {
      if (p.id !== pid) return p;
      const t = p.tasks[ti];
      const nc = Math.max(0, Math.min(3, t.col + dir));
      if (nc === t.col) return p;
      return { ...p, tasks: p.tasks.map((x, j) => (j === ti ? { ...x, col: nc } : x)), activity: [{ text: 'You moved "' + t.title + '" ' + (dir > 0 ? 'forward' : 'back') + ' to ' + names[nc], time: 'Just now' }].concat(p.activity) };
    }) }));
  }
};
