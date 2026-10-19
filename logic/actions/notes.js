// Actions for notes opened from elsewhere (e.g. a note shared into a forum's Files tab).
export const noteActions = {
  openNote(id) {
    const n = this.state.notes.find((x) => x.id === id);
    if (!n) return;
    if (this.state.widgets.indexOf('notes') === -1) { this.notify('Notes is disconnected. Add the My notes widget to Home to open notes.', '#FFB86B'); return; }
    this.setState({ section: 'notes', noteShare: null, noteBook: null, noteEdit: { id: n.id, book: n.book, title: n.title, body: n.body, orig: { title: n.title, body: n.body, book: n.book }, err: '' } });
  }
};
