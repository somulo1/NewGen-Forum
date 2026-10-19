// renderVals step: notes. Notebooks (like folders) hold notes; the widget on Home saves into "Quick notes".
export function buildNotes(ctx) {
  const { s } = ctx;
  const when = (t) => { const m = Math.round((Date.now() - t) / 60000); return m < 1 ? 'Just now' : m < 60 ? m + ' min ago' : m < 1440 ? Math.round(m / 60) + ' h ago' : new Date(t).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' }); };
  const firstLine = (b) => (b || '').split('\n').find((l) => l.trim()) || '';
  const bookOf = (id) => s.noteBooks.find((b) => b.id === id) || s.noteBooks[0];
  const set = (o) => this.setState(o);
  const flash = (msg) => { set({ noteFlash: msg }); clearTimeout(this._noteT); this._noteT = setTimeout(() => set({ noteFlash: '' }), 2500); };
  const q = (s.noteQ || '').trim().toLowerCase();
  const editOf = (n) => ({ id: n.id, book: n.book, title: n.title, body: n.body, orig: { title: n.title, body: n.body, book: n.book }, err: '' });
  const card = (n) => { const b = bookOf(n.book); return { id: n.id, body: n.body || '', kindTag: 'NOTE', meta: when(n.updated) + ' · ' + b.name, title: n.title || 'Untitled', preview: (n.body || '').split('\n').filter((l) => l.trim()).slice(0, 4).join('\n'), when: when(n.updated), pinned: n.pinned, color: b.color, bookName: b.name,
    open: () => set({ noteEdit: editOf(n) }),
    togglePin: () => this.setState((st) => ({ notes: st.notes.map((x) => (x.id === n.id ? { ...x, pinned: !x.pinned } : x)) })),
    remove: () => { this.setState((st) => ({ notes: st.notes.filter((x) => x.id !== n.id), noteEdit: null })); flash('Note deleted.'); } }; };
  const sorted = (list) => list.slice().sort((a, b) => (b.pinned - a.pinned) || (b.updated - a.updated));
  const books = s.noteBooks.map((b) => { const list = sorted(s.notes.filter((n) => n.book === b.id)); return { id: b.id, name: b.name, color: b.color, count: list.length, updated: list.length ? when(Math.max(...list.map((n) => n.updated))) : 'empty', samples: list.slice(0, 3).map((n, i) => ({ title: n.title || firstLine(n.body), kind: 'NOTE', color: b.color, h: ['88px', '72px', '80px'][i] })), open: () => set({ noteBook: b.id, noteQ: '' }) }; });
  const cur = s.noteBook ? s.noteBooks.find((b) => b.id === s.noteBook) : null;
  const inBook = cur ? sorted(s.notes.filter((n) => n.book === cur.id && (!q || (n.title + ' ' + n.body).toLowerCase().indexOf(q) !== -1))).map(card) : [];
  const allMatches = !cur && q ? sorted(s.notes.filter((n) => (n.title + ' ' + n.body).toLowerCase().indexOf(q) !== -1)).map(card) : [];
  const ed = s.noteEdit;
  const orig = ed && ed.orig ? ed.orig : { title: '', body: '', book: ed ? ed.book : '' };
  const cur0 = ed && ed.id ? s.notes.find((x) => x.id === ed.id) : null;
  const words = ed ? ((ed.body || '').trim() ? (ed.body || '').trim().split(/\s+/).length : 0) : 0;
  const editor = ed ? { open: true, isNew: !ed.id, title: ed.title, body: ed.body, book: ed.book, bookName: bookOf(ed.book).name, bookColor: bookOf(ed.book).color, err: ed.err || '', hasErr: !!ed.err, books: s.noteBooks.map((b) => ({ v: b.id, l: b.name })),
    dirty: !ed.id || ed.title !== orig.title || ed.body !== orig.body || ed.book !== orig.book, words: words + (words === 1 ? ' word' : ' words'), edited: cur0 ? 'Edited ' + when(cur0.updated) : 'Not saved yet', pinned: !!(cur0 && cur0.pinned), canPin: !!cur0,
    onTitle: (e) => set({ noteEdit: { ...ed, title: e.target.value, err: '' } }), onBody: (e) => set({ noteEdit: { ...ed, body: e.target.value, err: '' } }), onBook: (e) => set({ noteEdit: { ...ed, book: e.target.value } }),
    addEmoji: (ch) => set({ noteEdit: { ...ed, body: ed.body + ch } }),
    togglePin: () => { if (cur0) this.setState((st) => ({ notes: st.notes.map((x) => (x.id === cur0.id ? { ...x, pinned: !x.pinned } : x)) })); },
    close: () => set({ noteEdit: null }), cancel: () => set({ noteEdit: null }),
    save: () => {
      if (!(ed.title || '').trim() && !(ed.body || '').trim()) { set({ noteEdit: { ...ed, err: 'Write something to save.' } }); return; }
      const title = (ed.title || '').trim() || firstLine(ed.body).slice(0, 40);
      if (ed.id) this.setState((st) => ({ notes: st.notes.map((x) => (x.id === ed.id ? { ...x, title, body: ed.body, book: ed.book, updated: Date.now() } : x)), noteEdit: { ...ed, title, orig: { title, body: ed.body, book: ed.book }, err: '' } }));
      else { const id = 'n' + Date.now(); this.setState((st) => ({ notes: [{ id, book: ed.book, title, body: ed.body, updated: Date.now(), pinned: false }].concat(st.notes), noteEdit: { id, book: ed.book, title, body: ed.body, orig: { title, body: ed.body, book: ed.book }, err: '' } })); }
      flash('Saved.');
    },
    remove: ed.id ? () => { this.setState((st) => ({ notes: st.notes.filter((x) => x.id !== ed.id), noteEdit: null })); flash('Note deleted.'); } : null } : { open: false };
  const notesPage = { books, inBook: !!cur, book: cur ? { id: cur.id, name: cur.name, color: cur.color, count: inBook.length } : null, notes: inBook, hasNotes: inBook.length > 0, q: s.noteQ, onQ: (e) => set({ noteQ: e.target.value }), searching: !cur && !!q, matches: allMatches,
    back: () => set({ noteBook: null, noteQ: '' }), newNote: () => set({ noteEdit: { id: '', book: cur ? cur.id : s.noteBooks[0].id, title: '', body: '', orig: { title: '', body: '', book: '' }, err: '' } }),
    newBook: { open: s.noteNewBook !== null, name: s.noteNewBook || '', toggle: () => set({ noteNewBook: s.noteNewBook === null ? '' : null }), onName: (e) => set({ noteNewBook: e.target.value }),
      create: () => { const n = (s.noteNewBook || '').trim(); if (!n) return; const colors = ['#FFD66B', '#7CC4FF', '#C9B6FF', '#6EE7C8', '#F49AC1', '#FFB86B']; this.setState((st) => ({ noteBooks: st.noteBooks.concat([{ id: 'nb' + Date.now(), name: n, color: colors[st.noteBooks.length % colors.length] }]), noteNewBook: null })); flash('Notebook created.'); } },
    editor, flash: s.noteFlash, hasFlash: !!s.noteFlash, total: s.notes.length };
  // Home widget: a quick note with a Save button that files it under Quick notes.
  const noteWidget = { text: s.note, canSave: !!(s.note || '').trim(), recent: sorted(s.notes).slice(0, 3).map((n) => ({ id: n.id, title: n.title || firstLine(n.body), when: when(n.updated), color: bookOf(n.book).color, open: () => set({ section: 'notes', noteBook: n.book, noteEdit: editOf(n) }) })),
    save: () => { const body = (s.note || '').trim(); if (!body) return; this.setState((st) => ({ notes: [{ id: 'n' + Date.now(), book: 'nb1', title: firstLine(body).slice(0, 40), body, updated: Date.now(), pinned: false }].concat(st.notes), note: '' })); flash('Saved to Quick notes.'); },
    openAll: () => set({ section: 'notes', noteBook: null, noteQ: '' }), flash: s.noteFlash, count: s.notes.length };
  return { notesPage, noteWidget };
}
