// "Someone is typing", event driven. While a person types, their app tells the conversation (at most every 2.5 seconds, only
// while the text really changes) and says so the moment they stop: they empty the box, send, switch chat or leave the tab.
// The other side shows the indicator from the first event until the "stopped" event, the person's message arriving, or 5 seconds
// without a new event (a closed laptop or lost connection can never leave it stuck). Keys are the screens' own chat ids.
const BEAT = 2500;
const EXPIRE = 5000;

export function watch(logic, srv) {
  const rt = srv.rt;
  if (!rt || srv.typingWatching) return;
  srv.typingWatching = true;
  const live = new Map();   // `${conversation}|${person}` -> { local, name, timer }
  const localOf = (sid) => (srv.forumLocalId && srv.forumLocalId.get(sid)) || sid;
  const publish = () => {
    const by = {};
    live.forEach((v) => { (by[v.local] || (by[v.local] = [])).push(v.name); });
    logic.setState({ typing: by });
  };
  const drop = (k) => { const v = live.get(k); if (!v) return; clearTimeout(v.timer); live.delete(k); publish(); };
  rt.on('typing', (ev) => {
    const d = ev.data || {};
    const sid = String(ev.channel || '').split(':').pop();
    if (!d.person || d.person === srv.me.person) return;
    const k = sid + '|' + d.person;
    if (d.on === false) { drop(k); return; }
    const had = live.get(k);
    if (had) clearTimeout(had.timer);
    live.set(k, { local: localOf(sid), name: (d.name || 'Someone').split(' ')[0], timer: setTimeout(() => drop(k), EXPIRE) });
    if (!had) publish();
  });
  // a message from them means they are no longer typing it
  rt.on('message.new', (ev) => {
    const m = ev.data || {};
    const sid = m.conversation || String(ev.channel || '').split(':').pop();
    if (m.sender_person) drop(sid + '|' + m.sender_person);
  });
  rt.on('presence', (ev) => { const d = ev.data || {}; if (d.online === false) [...live.keys()].filter((k) => k.endsWith('|' + d.person)).forEach(drop); });
  // my own connection dropped: whatever I was told is stale
  rt.on('open', () => { [...live.keys()].forEach((k) => { clearTimeout(live.get(k).timer); }); live.clear(); publish(); });
  document.addEventListener('visibilitychange', () => { if (document.hidden) signal(logic, srv, ''); });
}

// The conversation the open chat screen is on, as the server knows it.
function openConversation(logic, srv) {
  const st = logic.state;
  if (['closed', 'min'].includes(st.mode)) return '';
  if (st.panelKind === 'dm') return st.dmId && !String(st.dmId).startsWith('dm-') ? st.dmId : '';
  return (srv.forumServerId && srv.forumServerId.get(st.forum)) || '';
}

// Called whenever the draft changes (and with '' when it is sent or cleared).
export function signal(logic, srv, text) {
  const rt = srv && srv.rt;
  if (!rt) return;
  const sent = srv.typingSent || (srv.typingSent = { sid: '', at: 0 });
  const sid = text && String(text).replace(/<[^>]*>/g, '').trim() ? openConversation(logic, srv) : '';
  if (sent.sid && sent.sid !== sid) { rt.typing(sent.sid, false); sent.sid = ''; sent.at = 0; }
  if (!sid) return;
  const now = Date.now();
  if (sent.sid === sid && now - sent.at < BEAT) return;
  sent.sid = sid; sent.at = now;
  rt.typing(sid, true);
}
