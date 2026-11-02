// Which notification sounds a person wants. Only the short chime that goes with a notice (a toast in the corner, or a new item in the
// bell) is covered here: the sounds of calls, meetings and open chats are separate and not affected.
// A person's choices are kept in their preferences (notifTones); anything they never touched uses these defaults.

// the groups of notices a person can switch, in the order they are shown
export const SOUND_KINDS = [
  { key: 'done', group: 'What I just did', label: 'Confirmations', hint: '"Saved", "Sent", "Done": a successful action you made', on: false },
  { key: 'warn', group: 'What I just did', label: 'Warnings', hint: '"Not allowed", "You already voted"', on: true },
  { key: 'error', group: 'What I just did', label: 'Problems', hint: 'Something did not go through', on: true },
  { key: 'info', group: 'What I just did', label: 'Information', hint: 'Other short notices', on: false },
  { key: 'ticket', group: 'In my notifications', label: 'Tickets', hint: 'Assigned, overdue, solved', on: true },
  { key: 'comm', group: 'In my notifications', label: 'Communications', hint: 'Memos, announcements, posts for you', on: true },
  { key: 'approval', group: 'In my notifications', label: 'Approvals and awards', hint: 'Requests waiting for you, decisions, awards', on: true },
  { key: 'alert', group: 'In my notifications', label: 'Security alerts', hint: 'Alerts raised for your role', on: true },
  { key: 'other', group: 'In my notifications', label: 'Everything else', hint: 'Any other new notification', on: true }
];

export const DEFAULTS = Object.fromEntries(SOUND_KINDS.map((k) => [k.key, k.on]));

// prefs: { all?: boolean, [key]: boolean } (what the person chose); true when this kind of notice should chime
export function rings(prefs, key) {
  const p = prefs || {};
  if (p.all === false) return false;
  return key in p ? !!p[key] : !!DEFAULTS[key];
}

// what a person has chosen for one kind (ignoring the master switch), for showing the switch
export const chosen = (prefs, key) => (prefs && key in prefs ? !!prefs[key] : !!DEFAULTS[key]);

// a toast's colour says what kind of notice it is
export function kindOfToast(color) {
  switch (String(color || '').toUpperCase()) {
    case '#B8F36B': return 'done';
    case '#FFB86B': case '#FFD66B': return 'warn';
    case '#FF9B8A': return 'error';
    default: return 'info';
  }
}

// the kind of a notification in the bell, as the server names it
export function kindOfNote(kind) {
  if (kind === 'ticket' || kind === 'comm' || kind === 'alert') return kind;
  if (kind === 'approval' || kind === 'award') return 'approval';
  return 'other';
}
