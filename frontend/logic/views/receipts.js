// The ticks on my own messages: 'wait' (one tick) while the message is still waiting to be sent, 'sent' (two grey) once the server
// has it, 'read' (two blue) once it has been read: in a private chat by the other person, in a forum by everyone in it.
const stamp = (v) => Date.parse(v) || 0;

// sent: the server has already answered for this message, though its stored copy has not replaced the temporary one yet
const waiting = (m) => !!m.queued || (!m.sent && (!m.mid || !!m.pending));

// others: the people besides me who could read it ({ person, name }); places: person -> when they have read up to.
export function receipt(m, others, places, mine) {
  if (!mine) return {};
  if (waiting(m)) return { tick: 'wait' };
  const t = stamp(m.at);
  const done = (p) => t > 0 && stamp(places[p.person]) >= t;
  const read = others.filter(done), unread = others.filter((p) => !done(p));
  return { tick: others.length && !unread.length ? 'read' : 'sent', seen: { read, unread, total: others.length } };
}

export const othersOf = (members, me) => (members || []).filter((p) => p.person && p.person !== me).map((p) => ({ person: p.person, name: p.name || 'Someone' }));

// a private chat has one other person; they may not be known by id until they have a place, so whoever is not me counts
export const dmOthers = (places, me) => Object.keys(places).filter((p) => p !== me).map((person) => ({ person, name: '' }));
