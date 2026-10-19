// Forums: the forums you are in (sidebar, panel chips, the My forums widget and your profile), their members and unread
// counts, who is online, and every message. A forum is an encrypted conversation: its name and messages are sealed for its
// members, and this reads, sends (text, replies, files, shared posts), deletes and marks read. New messages arrive over the
// realtime socket, so an open forum updates without a refresh.
import * as files from '../files.js';
import { connectRealtime } from '../realtime.js';
import { messageTone, alertTone } from '~/logic/tones.js';

export const name = 'forums';

// The forum ids the polls and chat rules still use. A forum with another name gets its server id.
const LOCAL_ID = { 'IT Operations': 'it', 'Product Builders': 'dev', 'Weekend Gamers': 'game', 'Choir & Worship': 'choir', 'Risk & Compliance': 'risk' };
const PRIVACY = { invite: 'Invite only', open: 'Open to everyone' };
const PAGE = 100;

// The API allows a burst of requests per person; the app's first load can use it all. A "slow down" answer is retried.
async function ask(srv, method, path, body) {
  for (let i = 0; ; i++) {
    try { return await srv.request(method, path, body); } catch (e) {
      if (!(e && e.status === 429) || i >= 5) throw e;
      await new Promise((r) => setTimeout(r, 1500 * (i + 1)));
    }
  }
}

const ago = (iso) => {
  const m = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 60000));
  if (m < 1) return 'Just now';
  if (m < 60) return m + ' min ago';
  const h = Math.round(m / 60);
  if (h < 24) return h + ' h ago';
  const d = Math.round(h / 24);
  return d === 1 ? 'Yesterday' : d < 7 ? d + ' days ago' : d < 14 ? 'Last week' : new Date(iso).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
};

const ME = { who: 'Sam', initials: 'S', color: '#F49AC1' };
const lookOf = (srv, person, name) => {
  const p = (srv.dirByUid && person && srv.dirByUid.get(person)) || null;
  return { initials: p ? p.initials : (name || '?').split(' ').map((x) => x[0]).join('').slice(0, 2).toUpperCase(), color: p ? p.color : '#C9B6FF' };
};

// "IT Operations" -> IT, "Product Builders" -> PB
const initialsOf = (n) => { const w = (n || '').split(/[\s&]+/).filter(Boolean); return w.length && /^[A-Z]{2,3}$/.test(w[0]) ? w[0] : w.map((x) => x[0]).join('').slice(0, 2).toUpperCase(); };
const localId = (srv, c) => LOCAL_ID[c.name] || c.id;
const serverId = (srv, local) => (srv.forumServerId && srv.forumServerId.get(local)) || '';
export const isLive = (srv, local) => !!(srv && serverId(srv, local));

// Turns the sealed messages of one forum (newest first) into the chat items the panels show (newest first).
function toItems(srv, msgs, opened) {
  const byId = new Map(opened.map((x) => [x.m.id, x]));
  const children = new Map();
  opened.forEach((x) => { if (x.m.reply_to) children.set(x.m.reply_to, (children.get(x.m.reply_to) || 0) + 1); });
  const nameOf = (m) => (m.sender_account === srv.me.account ? 'Sam' : m.sender_name);
  const snippet = (pl) => { const t = (pl && (pl.text || (pl.file && pl.file[0] && pl.file[0].name))) || ''; return t.length > 90 ? t.slice(0, 90) + '…' : t; };
  const todo = [];
  const items = opened.map(({ m, pl }) => {
    const mine = m.sender_account === srv.me.account;
    const parent = m.reply_to ? byId.get(m.reply_to) : null;
    const file = (pl.file || []).map((f) => { const it = { name: f.name, meta: f.meta || '', isImg: !!f.isImg, isVideo: !!f.isVideo, url: '', fileId: f.fileId, ref: f }; todo.push(f); return it; });
    return {
      id: m.id, mid: m.id, who: nameOf(m), ...(mine ? { initials: ME.initials, color: ME.color } : lookOf(srv, m.sender_person, m.sender_name)),
      time: ago(m.created_at), at: m.created_at, text: pl.text || '', file: file.length ? file : undefined, shared: pl.shared || null,
      replyTo: parent ? { who: parent.m.sender_account === srv.me.account ? 'You' : parent.m.sender_name, snippet: snippet(parent.pl) } : null, replies: children.get(m.id) || 0
    };
  });
  return { items, todo };
}

async function fetchForum(srv, c) {
  const [detail, page] = await Promise.all([
    ask(srv, 'GET', '/conversations/' + c.id),
    ask(srv, 'GET', `/conversations/${c.id}/messages?limit=${PAGE}`).catch(() => ({ messages: [] }))
  ]);
  const all = await Promise.all(page.messages.filter((m) => !m.deleted && m.type === 'text').map(async (m) => ({ m, pl: await srv.vault.openMessage(m) })));
  const opened = all.filter((x) => x.pl);
  const { items, todo } = toItems(srv, page.messages, opened);
  const members = (detail.members || []).map((x) => ({ person: x.person, account: x.account, name: x.account === srv.me.account ? srv.me.name : x.name, initials: x.initials, color: x.color, role: x.role }));
  return {
    forum: {
      id: localId(srv, detail), sid: detail.id, live: true, name: detail.name, initials: initialsOf(detail.name) || detail.initials,
      color: detail.color || '#7CC4FF', unread: detail.unread || 0, members: detail.member_count || members.length, online: 0, privacy: PRIVACY[detail.privacy] || 'Invite only', about: detail.about || '',
      myRole: ((detail.my_role || 'member')[0].toUpperCase() + (detail.my_role || 'member').slice(1)), epoch: detail.epoch, created: detail.created_at, memberList: members, updates: items, newest: page.messages.length ? page.messages[0].id : ''
    },
    todo
  };
}

// Fills in who is online from the server's presence (a person is online while their browser holds the socket open).
async function loadPresence(logic, srv) {
  const ids = [...new Set((logic.state.forumsLive || []).flatMap((f) => f.memberList.map((m) => m.person)))];
  const on = new Set();
  for (let i = 0; i < ids.length; i += 200) {
    try {
      const q = ids.slice(i, i + 200).map((id) => 'id=' + id).join('&');
      const { online } = await ask(srv, 'GET', '/presence?' + q);
      Object.keys(online || {}).forEach((id) => online[id] && on.add(id));
    } catch (e) { /* presence is best effort */ }
  }
  on.add(srv.me.person);
  srv.online = on;
  applyOnline(logic, srv);
}
function applyOnline(logic, srv) {
  logic.setState((st) => ({ forumsLive: (st.forumsLive || []).map((f) => ({ ...f, online: f.memberList.filter((m) => srv.online.has(m.person)).length, memberList: f.memberList.map((m) => ({ ...m, online: srv.online.has(m.person) })) })) }));
}

function loadFiles(logic, srv, todo) {
  todo.forEach((f) => files.openUrl(srv, f).then((url) => {
    logic.setState((st) => ({ forumsLive: (st.forumsLive || []).map((fo) => ({ ...fo, updates: fo.updates.map((u) => (u.file && u.file.some((x) => x.fileId === f.fileId) ? { ...u, file: u.file.map((x) => (x.fileId === f.fileId ? { ...x, url } : x)) } : u)) })) }));
  }).catch(() => {}));
}

export async function load(logic, srv) {
  const { conversations } = await ask(srv, 'GET', '/conversations?kind=forum&limit=200');
  const built = await Promise.all(conversations.map((c) => fetchForum(srv, c)));
  built.sort((a, b) => (a.forum.created < b.forum.created ? -1 : 1));
  srv.forumServerId = new Map(built.map((b) => [b.forum.id, b.forum.sid]));
  srv.forumLocalId = new Map(built.map((b) => [b.forum.sid, b.forum.id]));
  srv.forumConv = Object.fromEntries(built.map((b) => [b.forum.id, b.forum.sid]));   // used when a memo or announcement is sent to a forum
  logic.setState({ forumsLive: built.map((b) => b.forum) });
  built.forEach((b) => loadFiles(logic, srv, b.todo));
  await loadPresence(logic, srv);
  watch(logic, srv);
}

// Reloads one forum (its messages, members and unread count) and, if you are looking at it, marks it read.
export async function reloadForum(logic, srv, local) {
  const sid = serverId(srv, local);
  if (!sid) return;
  const { forum, todo } = await fetchForum(srv, { id: sid, name: (logic.state.forumsLive || []).find((f) => f.id === local)?.name });
  const open = logic.state.forum === local && logic.state.panelKind === 'forum' && !['closed', 'min'].includes(logic.state.mode);
  const prev = (logic.state.forumsLive || []).find((f) => f.id === local);
  const merged = { ...forum, online: prev ? prev.online : 0, unread: open ? 0 : forum.unread };
  logic.setState((st) => ({ forumsLive: (st.forumsLive || []).map((f) => (f.id === local ? { ...merged, memberList: merged.memberList.map((m) => ({ ...m, online: !!(srv.online && srv.online.has(m.person)) })) } : f)) }));
  applyOnline(logic, srv);
  loadFiles(logic, srv, todo);
  if (open && forum.newest) ask(srv, 'POST', `/conversations/${sid}/read`, { message: forum.newest }).catch(() => {});
}

// You opened a forum: its unread count goes to zero here and on the server.
export function markRead(logic, srv, local) {
  const sid = serverId(srv, local);
  if (!sid) return;
  const f = (logic.state.forumsLive || []).find((x) => x.id === local);
  if (!f || !f.newest) return;
  if (f.unread) logic.setState((st) => ({ forumsLive: st.forumsLive.map((x) => (x.id === local ? { ...x, unread: 0 } : x)) }));
  ask(srv, 'POST', `/conversations/${sid}/read`, { message: f.newest }).catch(() => {});
}

// m = { text, files: [File], replyTo: message id | '', shared: { sid, postId, snap } | null }
export async function send(logic, srv, local, m) {
  const sid = serverId(srv, local);
  if (!sid) throw new Error('This forum is not connected to the server.');
  const refs = [];
  for (const f of m.files || []) refs.push(await files.upload(srv, sid, f));
  const shared = m.shared ? { ...m.shared, snap: m.shared.snap ? { ...m.shared.snap, media: (m.shared.snap.media || []).map(({ name, meta, isImg, isVideo }) => ({ name, meta, isImg, isVideo })) } : undefined } : undefined;
  const payload = { text: m.text || '' };
  if (refs.length) payload.file = refs.map(({ fileId, name, meta, type, isImg, isVideo, key }) => ({ fileId, name, meta, type, isImg, isVideo, key }));
  if (shared) payload.shared = shared;
  const sealed = await srv.vault.seal(sid, payload);
  await ask(srv, 'POST', `/conversations/${sid}/messages`, { ...sealed, type: 'text', ...(m.replyTo ? { reply_to: m.replyTo } : {}) });
  await reloadForum(logic, srv, local);
}

export async function remove(logic, srv, local, mid) {
  await ask(srv, 'DELETE', `/conversations/${serverId(srv, local)}/messages/${mid}`);
  await reloadForum(logic, srv, local);
}

// ---- realtime ----

function watch(logic, srv) {
  const rt = srv.rt || (srv.rt = connectRealtime(srv));   // whichever widget loads first makes the socket
  if (!srv.forumsWatching) {
    srv.forumsWatching = true;
    const timers = new Map();
    const later = (local) => { clearTimeout(timers.get(local)); timers.set(local, setTimeout(() => reloadForum(logic, srv, local).catch(() => {}), 250)); };
    const forumOf = (ev) => (ev.channel && srv.forumLocalId ? srv.forumLocalId.get(String(ev.channel).split(':').pop()) : null);
    rt.on('message.new', (ev) => {
      const f = forumOf(ev);
      if (!f) return;
      later(f);
      const m = ev.data || {};
      if (m.sender_account === srv.me.account || m.type !== 'text') return;
      const st = logic.state;
      const onScreen = st.forum === f && st.panelKind === 'forum' && !['closed', 'min'].includes(st.mode);
      try { (onScreen ? messageTone : alertTone)(); } catch (e) { /* sound is optional */ }
    });
    ['message.deleted', 'member.added', 'member.removed', 'member.joined'].forEach((t) => rt.on(t, (ev) => { const f = forumOf(ev); if (f) later(f); }));
    rt.on('conversation.new', () => { load(logic, srv).catch(() => {}); });
    rt.on('conversation.removed', () => { load(logic, srv).catch(() => {}); });
    rt.on('presence', (ev) => {
      const d = ev.data || {};
      if (!d.person) return;
      if (d.online) srv.online.add(d.person); else srv.online.delete(d.person);
      applyOnline(logic, srv);
    });
    rt.on('open', () => { (logic.state.forumsLive || []).forEach((f) => later(f.id)); });
  }
  (logic.state.forumsLive || []).forEach((f) => rt.sub(f.sid));
}
