import { noteText, isMine, permLabel } from '~/logic/noteText.js';

// A note shared into a chat is a live link: it stops opening when the share is stopped, the note is deleted or Notes is switched off.
export function noteFile(file, s, openNote) {
  const note = s.notes.find((x) => x.container === file.noteId) || s.notes.find((x) => x.id === file.noteId);
  const share = note && isMine(note) ? (note.shares || []).find((x) => x.id === file.shareId) : null;
  // the link works while the share exists: for its owner that is the share entry, for everyone it was shared with it is having the note
  const live = !!note && (isMine(note) ? !!share : true);
  const perm = note ? (isMine(note) ? (share && share.perm) : note.access) : 'view';
  return { ...file, kind: 'NOTE', previewKind: 'note', color: '#FFD66B', isNote: true, live, noteLive: live, notePreview: live ? noteText(note.body).split('\n').filter((l) => l.trim()).slice(0, 6).join('\n') : '',
    meta: live ? 'Note · ' + permLabel(perm) + ' · live link' : 'Note · access removed', noteOpen: () => openNote(note.id) };
}

// A message's files, with the shared notes set apart so the bubble can show them as note cards and the Files tab can list them all.
export function splitFiles(m, s, openNote) {
  const all = [].concat(m.file || []).map((f) => (f.noteId ? noteFile(f, s, openNote) : f));
  const files = all.filter((f) => !f.isNote), notes = all.filter((f) => f.isNote);
  return { fileList: files, hasFile: files.length > 0, noteList: notes, hasNote: notes.length > 0, allFiles: all };
}
