import * as orgW from '~/logic/server/widgets/org.js';
import { catsFor } from '~/logic/ticketRouting.js';

// renderVals step: catMgr. Ticket categories for whatever is open: a unit dashboard, or the ticket desk.
export function buildCatMgr(ctx) {
  const { O, can, corePeople, defaultCats, ou, recOf, s, tickets } = ctx;
  const field = { departments: 'dept', branches: 'loc', sectors: 'sector', units: 'unit' };
  const right = { departments: 'Manage department', sectors: 'Manage sector', branches: 'Manage branch', units: 'Manage unit' };
  let target = null;
  if (s.section === 'ticketdesk') target = { type: 'desk', name: 'Ticket desk', canManage: can('tickets', 'Set up ticket categories'), pool: corePeople.slice() };
  else if (s.section === 'orgunit' && ou && ou.name && s.ou) {
    const parts = s.ou.split(':'), rec = recOf(parts[0], parts[1]);
    if (rec) target = { type: parts[0], name: rec.name, canManage: can('tickets', 'Set up ticket categories'), pool: this.dir.filter((p) => p[field[parts[0]]] === rec.name).map((p) => p.name).slice(0, 60) };
  }
  if (!target) return { catMgr: { active: false } };
  const key = target.type + ':' + target.name;
  const own = s.unitCats[key] || [];
  const working = !!(s.catWork && s.catWork.key === key);
  const list = working ? s.catWork.rows.map((c) => ({ ...c, custom: true })) : catsFor(s.unitCats, target, defaultCats);
  const dirty = working && JSON.stringify(s.catWork.rows) !== JSON.stringify(own);
  const d = s.catDraft;
  const person = (n) => { const p = this.dir.find((x) => x.name === n); return { name: n, initials: p ? p.initials : n.split(' ').map((x) => x[0]).join(''), color: p ? p.color : '#C9B6FF' }; };
  const ed = s.catEdit;
  const edCat = ed ? list.find((c) => c.id === ed.id) : null;
  const q = ed ? ed.q.trim().toLowerCase() : '';
  const same = ed && ed.members.length === ed.orig.length && ed.members.every((n) => ed.orig.indexOf(n) !== -1);
  const editor = edCat ? { open: true, name: edCat.name, q: ed.q, onQ: (e) => this.setCatEditQ(e.target.value),
    selected: ed.members.map((n) => ({ ...person(n), remove: () => this.toggleCatEdit(n) })), count: ed.members.length,
    results: target.pool.filter((n) => !q || n.toLowerCase().indexOf(q) !== -1).map((n) => ({ ...person(n), on: ed.members.indexOf(n) !== -1, toggle: () => this.toggleCatEdit(n) })),
    dirty: !same, save: () => this.applyCatEdit(key), cancel: () => this.cancelCatEdit() } : { open: false };
  // Performance threshold for this dashboard, edited here and saved with the button.
  const unitKey = s.section === 'orgunit' ? s.ou : null;
  const curThr = unitKey ? (s.targets[unitKey] || 95) : 95;
  const thr = s.ouThr && s.ouThr.key === unitKey ? s.ouThr.value : curThr;
  const thrNum = Math.max(50, Math.min(100, Number(thr) || 0));
  const threshold = { show: !!unitKey && target.type !== 'desk', current: curThr, value: thr, canEdit: target.canManage, dirty: String(thr) !== String(curThr) && !!thr,
    onInput: (e) => this.setState({ ouThr: { key: unitKey, value: e.target.value } }),
    save: () => { const v = Number(thr); if (!(v >= 50 && v <= 100)) { this.notify('Choose a threshold between 50 and 100%.', '#FFB86B'); return; } if (this.srv) { orgW.setTarget(this, this.srv, unitKey, Math.round(v)).then(() => { this.setState({ ouThr: null }); this.notify('Performance threshold saved at ' + Math.round(v) + '%.', '#B8F36B'); }).catch((e) => this.notify(e.message || 'Could not save the threshold.', '#FF9B8A')); return; } this.setState((st) => ({ targets: { ...st.targets, [unitKey]: Math.round(v) }, ouThr: null })); this.notify('Performance threshold saved at ' + Math.round(v) + '%.', '#B8F36B'); },
    discard: () => this.setState({ ouThr: null }), note: thrNum + '% of what ' + target.name + ' receives should be resolved within its time limit.' };
  return { catMgr: { active: true, threshold, editor, dirty, save: () => this.saveWork(), discard: () => this.discardWork(), noAccessNote: target.canManage ? '' : 'You can view these categories. Ask an admin to give your role the "Set up ticket categories" right to change them.', key, canManage: target.canManage, title: target.type === 'desk' ? 'Ticket desk categories' : 'Ticket categories for ' + target.name,
    usingDefaults: !own.length && !working, startOwn: () => this.startOwnCats(key, defaultCats),
    note: own.length ? 'Tickets sent here can only use these categories. Each has its own TAT and priority, and its members take new tickets automatically until you delink them.' : 'No categories of its own yet, so the organization defaults apply. Create your own to set TATs, priorities and who always takes them.',
    rows: list.map((c) => ({ id: c.id, name: c.name, tat: c.tat, prio: c.prio, custom: c.custom,       summary: c.members.slice(0, 3).map(person), extra: Math.max(0, c.members.length - 3), count: c.members.length, manage: () => this.openCatEdit(c.id, c.members),
      setTat: (e) => { const v = Number(e.target.value); if (v > 0) this.setCatField(key, c.id, { tat: v }); }, setPrio: (e) => this.setCatField(key, c.id, { prio: e.target.value }), remove: () => this.removeCat(key, c.id) })),
    draft: { name: d.name, tat: d.tat, prio: d.prio, err: d.err, hasErr: !!d.err, onName: (e) => this.setCatDraft({ name: e.target.value }), onTat: (e) => this.setCatDraft({ tat: e.target.value }), onPrio: (e) => this.setCatDraft({ prio: e.target.value }), add: () => this.addUnitCat(key) } } };
}
