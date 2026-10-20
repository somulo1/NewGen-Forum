// Initial state: notes (notebooks hold notes, like folders hold files).
export function notesState() {
  return {
    noteBooks: [],
    notes: [],

    noteShare: null, noteBook: null, noteEdit: null, noteQ: '', noteNewBook: null, noteFlash: ''
  };
}
