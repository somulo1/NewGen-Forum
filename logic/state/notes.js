// Initial state: notes (notebooks hold notes, like folders hold files).
export function notesState() {
  const h = 3600000, now = Date.now();
  return {
    noteBooks: [{ id: 'nb1', name: 'Quick notes', color: '#FFD66B' }, { id: 'nb2', name: 'Meetings', color: '#7CC4FF' }, { id: 'nb3', name: 'Ideas', color: '#C9B6FF' }],
    notes: [
      { id: 'n1', book: 'nb2', title: 'Risk review prep', body: 'Collect access review evidence.\nAsk Amina for the audit list.\nBook the room for Thursday.', updated: now - 3 * h, pinned: true },
      { id: 'n2', book: 'nb2', title: 'Town hall questions', body: 'Ask about the new leave policy and the office move.', updated: now - 30 * h, pinned: false },
      { id: 'n3', book: 'nb3', title: 'Self-service password reset', body: 'Let people reset their own password with a one-time code and security questions.', updated: now - 52 * h, pinned: false },
      { id: 'n4', book: 'nb1', title: 'Call Brian about the VPN certificate', body: 'Renew before Friday.', updated: now - 5 * h, pinned: false }
    ],
    noteBook: null, noteEdit: null, noteQ: '', noteNewBook: null, noteFlash: ''
  };
}
