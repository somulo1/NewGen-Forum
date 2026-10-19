import * as ticketsW from '~/logic/server/widgets/tickets.js';
// HubLogic actions: ticket categories a unit (or the ticket desk) attends to. Mixed into HubLogic.prototype.
// Edits go into a working copy (state.catWork). Nothing reaches the saved list, or any ticket, until "Save changes".
export const ticketCatActions = {
  workRows(key) {
    const w = this.state.catWork;
    return w && w.key === key ? w.rows : (this.state.unitCats[key] || []);
  },
  editWork(key, fn) {
    this.setState((st) => ({ catWork: { key, rows: fn(st.catWork && st.catWork.key === key ? st.catWork.rows : (st.unitCats[key] || [])) } }));
  },
  saveWork() {
    const w = this.state.catWork;
    if (!w) return;
    this.setState({ catWork: null });
    ticketsW.saveUnitCats(this, this.srv, w.key, w.rows)
      .then(() => this.notify('Ticket categories saved. New tickets use them from now on.', '#B8F36B'))
      .catch((e) => this.notify(e.message || 'Could not save the categories.', '#FF9B8A'));
  },
  discardWork() { this.setState({ catWork: null }); },
  startOwnCats(key, defaults) {
    this.setState({ catWork: null });
    ticketsW.saveUnitCats(this, this.srv, key, defaults.map((c, i) => ({ id: 'c' + Date.now() + i, name: c.name, tat: c.tat, prio: 'Medium', members: [] })))
      .catch((e) => this.notify(e.message || 'Could not set up the categories.', '#FF9B8A'));
  },
  setCatDraft(patch) { this.setState((st) => ({ catDraft: { ...st.catDraft, ...patch, err: '' } })); },
  addUnitCat(key) {
    const d = this.state.catDraft;
    const name = d.name.trim(), tat = Number(d.tat);
    let err = '';
    if (!name) err = 'Give the category a name.';
    else if (this.workRows(key).some((c) => c.name.toLowerCase() === name.toLowerCase())) err = 'This unit already has a category with that name.';
    else if (!(tat > 0)) err = 'Set a TAT of at least 1 hour.';
    if (err) { this.setState({ catDraft: { ...d, err } }); return; }
    this.editWork(key, (l) => l.concat([{ id: 'c' + Date.now(), name, tat, prio: d.prio, members: [] }]));
    this.setState({ catDraft: { name: '', tat: '4', prio: 'Medium', err: '' } });
  },
  setCatField(key, id, patch) { this.editWork(key, (l) => l.map((c) => (c.id === id ? { ...c, ...patch } : c))); },
  removeCat(key, id) { this.editWork(key, (l) => l.filter((c) => c.id !== id)); },

  // The assign card: choose people, then "Apply". The table's "Save changes" makes it permanent.
  openCatEdit(id, members) { this.setState({ catEdit: { id, members: members.slice(), orig: members.slice(), q: '' } }); },
  setCatEditQ(q) { this.setState((st) => ({ catEdit: { ...st.catEdit, q } })); },
  toggleCatEdit(name) {
    this.setState((st) => { const m = st.catEdit.members; return { catEdit: { ...st.catEdit, members: m.indexOf(name) !== -1 ? m.filter((x) => x !== name) : m.concat([name]) } }; });
  },
  applyCatEdit(key) {
    const e = this.state.catEdit;
    if (!e) return;
    this.setCatField(key, e.id, { members: e.members });
    this.setState({ catEdit: null });
  },
  cancelCatEdit() { this.setState({ catEdit: null }); }
};
