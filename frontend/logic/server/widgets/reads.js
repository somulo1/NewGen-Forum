// Who has read how far in each chat, for the delivery ticks. The server keeps each member's reading place (when the last message
// they read was sent); a message is read by someone when it was sent at or before their place. Places only move forward.
const stamp = (v) => Date.parse(v) || 0;
const MIN_GAP_MS = 1500;

const merge = (logic, conv, next) => logic.setState((st) => {
  const have = (st.reads || {})[conv] || {};
  const out = { ...have };
  Object.entries(next).forEach(([person, at]) => { if (stamp(at) >= stamp(out[person])) out[person] = at; });
  return { reads: { ...(st.reads || {}), [conv]: out } };
});

export async function load(logic, srv, conv) {
  if (!conv || String(conv).startsWith('dm-')) return;
  const seen = srv.readsAt || (srv.readsAt = new Map());
  if (Date.now() - (seen.get(conv) || 0) < MIN_GAP_MS) return;
  seen.set(conv, Date.now());
  try {
    const r = await srv.request('GET', `/conversations/${conv}/reads`);
    merge(logic, conv, Object.fromEntries((r.readers || []).map((x) => [x.person, x.at])));
  } catch (e) { /* the ticks stay as they were */ }
}

// Someone read a chat I am in: their place moves, and my ticks follow at once.
export function attach(logic, srv, rt) {
  if (srv.readsLive) return;
  srv.readsLive = true;
  rt.on('conversation.read', (ev) => {
    const d = ev.data || {};
    if (d.conversation && d.person) merge(logic, d.conversation, { [d.person]: d.at });
  });
  rt.on('open', () => {
    const st = logic.state;
    const open = st.panelKind === 'dm' ? st.dmId : st.forum;
    srv.readsAt && srv.readsAt.clear();
    load(logic, srv, open);
  });
}
