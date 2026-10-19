// renderVals step: notes. Notebooks (like folders) hold notes; the widget on Home saves into "Quick notes".
// A note can be shared with a person or a forum, read only or editable. Shares are live links: receivers find the note under Files,
// and if the sender stops sharing (or removes the Notes widget from Home) they lose access until it is shared / added again.
import { noteText as plain, notesConnected, isMine, permLabel } from '~/logic/noteText.js';

export function buildNotes(ctx) {
  const { s, forumsData } = ctx;
  const when = (t) => { const m = Math.round((Date.now() - t) / 60000); return m < 1 ? 'Just now' : m < 60 ? m + ' min ago' : m < 1440 ? Math.round(m / 60) + ' h ago' : new Date(t).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' }); };
  const firstLine = (b) => plain(b).split('\n').find((l) => l.trim()) || '';
  const bookOf = (id) => s.noteBooks.find((b) => b.id === id) || s.noteBooks[0];
  const set = (o) => this.setState(o);
  const flash = (msg) => { set({ noteFlash: msg }); clearTimeout(this._noteT); this._noteT = setTimeout(() => set({ noteFlash: '' }), 2500); };
  const q = (s.noteQ || '').trim().toLowerCase();
  const connected = notesConnected(s);
  const mine = s.notes.filter(isMine);
  const editOf = (n) => ({ id: n.id, book: n.book, title: n.title, body: n.body, orig: { title: n.title, body: n.body, book: n.book }, err: '' });
  const card = (n) => { const b = bookOf(n.book); return { id: n.id, body: plain(n.body).slice(0, 400), kindTag: 'NOTE', meta: when(n.updated) + ' · ' + b.name, title: n.title || 'Untitled', preview: plain(n.body).split('\n').filter((l) => l.trim()).slice(0, 4).join('\n'), when: when(n.updated), pinned: n.pinned, color: b.color, bookName: b.name,
    open: () => set({ noteEdit: editOf(n) }),
    togglePin: () => this.setState((st) => ({ notes: st.notes.map((x) => (x.id === n.id ? { ...x, pinned: !x.pinned } : x)) })),
    remove: () => { this.setState((st) => ({ notes: st.notes.filter((x) => x.id !== n.id), noteEdit: null })); flash('Note deleted.'); } }; };
  const sorted = (list) => list.slice().sort((a, b) => (b.pinned - a.pinned) || (b.updated - a.updated));
  const books = s.noteBooks.map((b) => { const list = sorted(mine.filter((n) => n.book === b.id)); return { id: b.id, name: b.name, color: b.color, count: list.length, updated: list.length ? when(Math.max(...list.map((n) => n.updated))) : 'empty', samples: list.slice(0, 3).map((n, i) => ({ title: n.title || firstLine(n.body), kind: 'NOTE', color: b.color, h: ['88px', '72px', '80px'][i] })), open: () => set({ noteBook: b.id, noteQ: '' }) }; });
  const cur = s.noteBook ? s.noteBooks.find((b) => b.id === s.noteBook) : null;
  const inBook = cur ? sorted(mine.filter((n) => n.book === cur.id && (!q || (n.title + ' ' + plain(n.body)).toLowerCase().indexOf(q) !== -1))).map(card) : [];
  const allMatches = !cur && q ? sorted(mine.filter((n) => (n.title + ' ' + plain(n.body)).toLowerCase().indexOf(q) !== -1)).map(card) : [];

  // ---- Sharing ----
  const forumName = (id) => (forumsData.find((f) => f.id === id) || { name: id }).name;
  const shareLabel = (sh) => (sh.kind === 'forum' ? forumName(sh.to) + ' (forum Files)' : sh.toName);
  const addShare = (note, sh) => this.setState((st) => ({ notes: st.notes.map((x) => (x.id === note.id ? { ...x, shares: (x.shares || []).concat([sh]) } : x)) }));
  const stopShare = (note, sh) => {
    this.setState((st) => ({
      notes: st.notes.map((x) => (x.id === note.id ? { ...x, shares: (x.shares || []).filter((y) => y.id !== sh.id) } : x)),
      added: sh.kind === 'forum' ? { ...st.added, [sh.to]: (st.added[sh.to] || []).filter((m) => ![].concat(m.file || []).some((f) => f.shareId === sh.id)) } : st.added
    }));
    flash('Stopped sharing with ' + shareLabel(sh) + '.');
  };
  const setPerm = (note, sh, perm) => this.setState((st) => ({ notes: st.notes.map((x) => (x.id === note.id ? { ...x, shares: (x.shares || []).map((y) => (y.id === sh.id ? { ...y, perm } : y)) } : x)) }));
  const sp = s.noteShare;
  const people = this.dir.map((p) => p.name).filter((n) => n !== 'Sam').sort();
  const sendShare = (note) => {
    const perm = sp.perm || 'view';
    const id = 'sh' + Date.now();
    if (sp.kind === 'forum') {
      const f = forumsData.find((x) => x.id === sp.to);
      if (!f) { set({ noteShare: { ...sp, err: 'Choose a forum.' } }); return; }
      if ((note.shares || []).some((x) => x.kind === 'forum' && x.to === f.id)) { set({ noteShare: { ...sp, err: 'Already shared with ' + f.name + '.' } }); return; }
      addShare(note, { id, kind: 'forum', to: f.id, toName: f.name, perm, at: Date.now() });
      this.addForumMsg(f.id, '', [{ name: (note.title || 'Untitled') + '.note', meta: 'Note · ' + permLabel(perm), noteId: note.id, shareId: id }]);
      flash('Sent to ' + f.name + '. It is in the forum Files tab.');
    } else {
      const name = (sp.to || '').trim();
      if (!people.includes(name)) { set({ noteShare: { ...sp, err: 'Pick a person from the list.' } }); return; }
      if ((note.shares || []).some((x) => x.kind === 'person' && x.to === name)) { set({ noteShare: { ...sp, err: 'Already shared with ' + name + '.' } }); return; }
      addShare(note, { id, kind: 'person', to: name, toName: name, perm, at: Date.now() });
      flash('Shared with ' + name + '. They will find it under Files.');
    }
    set({ noteShare: { ...sp, to: '', err: '' } });
  };

  // ---- Editor ----
  const ed = s.noteEdit;
  const orig = ed && ed.orig ? ed.orig : { title: '', body: '', book: ed ? ed.book : '' };
  const cur0 = ed && ed.id ? s.notes.find((x) => x.id === ed.id) : null;
  const incoming = !!(cur0 && !isMine(cur0));
  const readOnly = incoming && cur0.access !== 'edit';
  const txt = ed ? plain(ed.body) : '';
  const words = txt.trim() ? txt.trim().split(/\s+/).length : 0;
  const shares = cur0 && !incoming ? (cur0.shares || []) : [];
  const editor = ed ? { open: true, isNew: !ed.id, incoming, readOnly, ownerName: incoming ? cur0.owner : '', accessLabel: incoming ? (readOnly ? 'Read only' : 'Can edit') : '',
    title: ed.title, body: ed.body, book: ed.book, bookName: incoming ? 'Shared by ' + cur0.owner : bookOf(ed.book).name, bookColor: incoming ? '#C9B6FF' : bookOf(ed.book).color, err: ed.err || '', hasErr: !!ed.err, books: s.noteBooks.map((b) => ({ v: b.id, l: b.name })),
    dirty: !readOnly && (!ed.id || ed.title !== orig.title || ed.body !== orig.body || ed.book !== orig.book), words: words + (words === 1 ? ' word' : ' words'), edited: cur0 ? 'Edited ' + when(cur0.updated) : 'Not saved yet', pinned: !!(cur0 && cur0.pinned), canPin: !!cur0 && !incoming,
    onTitle: (e) => set({ noteEdit: { ...ed, title: e.target.value, err: '' } }), onBody: (html) => set({ noteEdit: { ...ed, body: html, err: '' } }), onBook: (e) => set({ noteEdit: { ...ed, book: e.target.value } }),
    togglePin: () => { if (cur0) this.setState((st) => ({ notes: st.notes.map((x) => (x.id === cur0.id ? { ...x, pinned: !x.pinned } : x)) })); },
    close: () => set({ noteEdit: null, noteShare: null }), cancel: () => set({ noteEdit: null, noteShare: null }),
    save: () => {
      if (readOnly) return;
      if (!(ed.title || '').trim() && !plain(ed.body).trim()) { set({ noteEdit: { ...ed, err: 'Write something to save.' } }); return; }
      const title = (ed.title || '').trim() || firstLine(ed.body).slice(0, 40);
      const book = incoming ? cur0.book : ed.book;
      if (ed.id) this.setState((st) => ({ notes: st.notes.map((x) => (x.id === ed.id ? { ...x, title, body: ed.body, book, updated: Date.now() } : x)), noteEdit: { ...ed, title, orig: { title, body: ed.body, book }, err: '' } }));
      else { const id = 'n' + Date.now(); this.setState((st) => ({ notes: [{ id, book: ed.book, title, body: ed.body, updated: Date.now(), pinned: false }].concat(st.notes), noteEdit: { id, book: ed.book, title, body: ed.body, orig: { title, body: ed.body, book: ed.book }, err: '' } })); }
      flash(incoming ? 'Saved. ' + cur0.owner + ' will see your changes.' : 'Saved.');
    },
    remove: ed.id && !incoming ? () => { this.setState((st) => ({ notes: st.notes.filter((x) => x.id !== ed.id), added: Object.fromEntries(Object.entries(st.added).map(([k, v]) => [k, v.filter((m) => ![].concat(m.file || []).some((f) => f.noteId === ed.id))])), noteEdit: null })); flash('Note deleted.'); } : null,
    disconnect: incoming ? () => { this.setState((st) => ({ notes: st.notes.filter((x) => x.id !== ed.id), noteEdit: null })); flash('Removed from your files. ' + cur0.owner + ' has to share it again.'); } : null,
    canShare: !!cur0 && !incoming,
    share: { open: !!sp && !!cur0 && !incoming, toggle: () => set({ noteShare: sp ? null : { kind: 'person', to: '', perm: 'view', err: '' } }),
      isPerson: !!sp && sp.kind === 'person', isForum: !!sp && sp.kind === 'forum', to: sp ? sp.to : '', perm: sp ? sp.perm : 'view', err: sp ? sp.err || '' : '', hasErr: !!(sp && sp.err),
      people, forums: forumsData.map((f) => ({ v: f.id, l: f.name })), connected, count: shares.length,
      setKind: (k) => () => set({ noteShare: { ...sp, kind: k, to: '', err: '' } }), onTo: (e) => set({ noteShare: { ...sp, to: e.target.value, err: '' } }), onPerm: (e) => set({ noteShare: { ...sp, perm: e.target.value } }),
      send: () => cur0 && sendShare(cur0),
      list: shares.map((sh) => ({ id: sh.id, label: shareLabel(sh), kindLabel: sh.kind === 'forum' ? 'Forum' : 'Person', perm: sh.perm, live: connected, when: when(sh.at),
        onPerm: (e) => setPerm(cur0, sh, e.target.value), stop: () => stopShare(cur0, sh) })) } } : { open: false };

  // ---- Files page: notes shared with me, and notes I share ----
  const noteCard = (n, extra) => ({ id: n.id, title: n.title || 'Untitled', preview: plain(n.body).split('\n').filter((l) => l.trim()).slice(0, 5).join('\n'), when: when(n.updated), color: '#FFD66B',
    open: () => (connected ? set({ section: 'notes', noteBook: null, noteEdit: editOf(n) }) : this.notify('Notes is disconnected. Add the My notes widget to Home to open notes.', '#FFB86B')), ...extra });
  const sharedIn = sorted(s.notes.filter((n) => !isMine(n) && (n.shares || []).some((x) => x.to === 'Sam'))).map((n) => noteCard(n, { by: n.owner, access: n.access, accessLabel: permLabel(n.access), canEdit: n.access === 'edit',
    remove: () => { this.setState((st) => ({ notes: st.notes.filter((x) => x.id !== n.id), noteEdit: st.noteEdit && st.noteEdit.id === n.id ? null : st.noteEdit })); flash('Removed from your files. ' + n.owner + ' has to share it again.'); } }));
  const sharedOut = mine.flatMap((n) => (n.shares || []).map((sh) => noteCard(n, { shareId: sh.id, to: shareLabel(sh), accessLabel: permLabel(sh.perm), live: connected, stop: () => stopShare(n, sh) })));
  const sharedNotes = { in: sharedIn, out: sharedOut, hasIn: sharedIn.length > 0, hasOut: sharedOut.length > 0, connected, disconnected: !connected };

  const notesPage = { books, inBook: !!cur, book: cur ? { id: cur.id, name: cur.name, color: cur.color, count: inBook.length } : null, notes: inBook, hasNotes: inBook.length > 0, q: s.noteQ, onQ: (e) => set({ noteQ: e.target.value }), searching: !cur && !!q, matches: allMatches,
    back: () => set({ noteBook: null, noteQ: '' }), newNote: () => set({ noteShare: null, noteEdit: { id: '', book: cur ? cur.id : s.noteBooks[0].id, title: '', body: '', orig: { title: '', body: '', book: '' }, err: '' } }),
    newBook: { open: s.noteNewBook !== null, name: s.noteNewBook || '', toggle: () => set({ noteNewBook: s.noteNewBook === null ? '' : null }), onName: (e) => set({ noteNewBook: e.target.value }),
      create: () => { const n = (s.noteNewBook || '').trim(); if (!n) return; const colors = ['#FFD66B', '#7CC4FF', '#C9B6FF', '#6EE7C8', '#F49AC1', '#FFB86B']; this.setState((st) => ({ noteBooks: st.noteBooks.concat([{ id: 'nb' + Date.now(), name: n, color: colors[st.noteBooks.length % colors.length] }]), noteNewBook: null })); flash('Notebook created.'); } },
    editor, flash: s.noteFlash, hasFlash: !!s.noteFlash, total: mine.length };
  // Home widget: a quick note with a Save button that files it under Quick notes.
  const noteWidget = { text: s.note, canSave: !!(s.note || '').trim(), recent: sorted(mine).slice(0, 3).map((n) => ({ id: n.id, title: n.title || firstLine(n.body), when: when(n.updated), color: bookOf(n.book).color, open: () => set({ section: 'notes', noteBook: n.book, noteEdit: editOf(n) }) })),
    save: () => { const body = (s.note || '').trim(); if (!body) return; this.setState((st) => ({ notes: [{ id: 'n' + Date.now(), book: 'nb1', title: firstLine(body).slice(0, 40), body, updated: Date.now(), pinned: false }].concat(st.notes), note: '' })); flash('Saved to Quick notes.'); },
    openAll: () => set({ section: 'notes', noteBook: null, noteQ: '' }), flash: s.noteFlash, count: mine.length };
  return { notesPage, noteWidget, sharedNotes };
}
