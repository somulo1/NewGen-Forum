// Who is online right now. A person is online while their browser holds the realtime connection open (the server keeps that in Redis).
// The directory of online people lives in srv.online (a Set of person ids), shared with the forums' member lists. People are looked up
// when a screen needs them (a private chat's other person), then kept current by the server's live "presence" notices; after the
// connection to the server comes back, they are looked up again, because notices sent meanwhile were missed.
import { connectRealtime } from './realtime.js';

const tracked = (srv) => srv.presenceIds || (srv.presenceIds = new Set());

export const isOnline = (srv, uid) => !!(srv && uid && srv.online && srv.online.has(uid));

async function fetchNow(logic, srv, ids) {
  if (!ids.length) return;
  const on = srv.online || (srv.online = new Set([srv.me.person]));
  for (let i = 0; i < ids.length; i += 200) {
    const part = ids.slice(i, i + 200);
    try {
      const { online } = await srv.request('GET', '/presence?' + part.map((id) => 'id=' + id).join('&'));
      part.forEach((id) => { if (online && online[id]) on.add(id); else on.delete(id); });
    } catch (e) { /* best effort: the live notices keep it right afterwards */ }
  }
  on.add(srv.me.person);
  logic.forceUpdate();
}

// Starts following these people (and asks where they stand now).
export function track(logic, srv, ids) {
  const set = tracked(srv);
  const fresh = (ids || []).filter((id) => id && !set.has(id));
  fresh.forEach((id) => set.add(id));
  watch(logic, srv);
  return fetchNow(logic, srv, fresh);
}

function watch(logic, srv) {
  if (srv.presenceWatching) return;
  srv.presenceWatching = true;
  const rt = srv.rt || (srv.rt = connectRealtime(srv));
  const on = () => srv.online || (srv.online = new Set([srv.me.person]));
  rt.on('presence', (ev) => {
    const d = ev.data || {};
    if (!d.person) return;
    if (d.online || d.person === srv.me.person) on().add(d.person); else on().delete(d.person);
    logic.forceUpdate();
  });
  // after a reconnect, anything announced while the connection was down was missed: ask again (after a short random pause)
  rt.on('open', (ev) => {
    if (!ev || !ev.reconnect) return;
    setTimeout(() => fetchNow(logic, srv, [...tracked(srv)]), Math.random() * 2000);
  });
}
