// Catching up after a dropped connection without asking about every chat. One request lists the conversations with the time each
// last changed (a message, an edit, a like, a removal); only the ones that changed since this device last looked are read again.
// Someone with 60 chats who was offline for a minute used to send 60 requests on reconnecting; now it is one, plus one per chat that
// really changed. A random pause comes first, so thousands of devices that lost the connection together do not return together.
import { spread } from './backoff.js';

const SAFETY_MS = 1000;   // chats that changed in the last second before the mark are read again, so a change on the boundary is never missed

export async function changedSince(srv, kind, mark) {
  const { conversations } = await srv.request('GET', `/conversations?kind=${kind}&limit=200`);
  // an older server (or a first look with no mark) cannot say what changed: then everything is read again, as before
  if (!mark || conversations.some((c) => !c.changed_at)) return { changed: conversations.map((c) => c.id), all: conversations.map((c) => c.id), mark: markOf(conversations) };
  const at = (c) => Date.parse(c.changed_at);
  const cut = Date.parse(mark) - SAFETY_MS;
  let top = Date.parse(mark);
  const changed = [];
  for (const c of conversations) {
    if (at(c) > cut) changed.push(c.id);
    if (at(c) > top) top = at(c);
  }
  return { changed, all: conversations.map((c) => c.id), mark: new Date(top).toISOString() };
}

// the newest change among a list of conversations (the starting mark after a full load)
export function markOf(conversations) {
  let top = 0;
  for (const c of conversations) { const t = c.changed_at ? Date.parse(c.changed_at) : 0; if (t > top) top = t; }
  return top ? new Date(top).toISOString() : '';
}

export const pause = (maxMs = 2500) => new Promise((r) => setTimeout(r, spread(maxMs)));
