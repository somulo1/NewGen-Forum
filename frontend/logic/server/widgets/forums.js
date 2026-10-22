// Forums: the forums you are in (sidebar, panel chips, the My forums widget and your profile), their members, unread counts,
// who is online, every message, polls and posting rules. A forum is an encrypted conversation: its name and messages are
// sealed for its members. This reads them, sends (text, replies, files, shared posts), votes in polls, changes the posting
// rules, creates forums and invites people. New messages arrive over the realtime socket, so an open forum updates by itself.
import { newConversationKey, wrap, newId, b64, unb64 } from '../e2ee.js';
import * as files from '../files.js';
import { connectRealtime } from '../realtime.js';
import { messageTone, alertTone } from '~/logic/tones.js';

export const name = 'forums';

// The forum ids the chat-rule and poll screens used before the server; a forum with another name gets its server id.
const LOCAL_ID = { 'IT Operations': 'it', 'Product Builders': 'dev', 'Weekend Gamers': 'game', 'Choir & Worship': 'choir', 'Risk & Compliance': 'risk' };
const PRIVACY = { invite: 'Invite only', open: 'Open to everyone' };
const PALETTE = ['#B8F36B', '#7CC4FF', '#FFB86B', '#F49AC1', '#6EE7C8', '#C9B6FF', '#FFD66B'];
const PAGE = 100, MAX_MESSAGES = 300;

// The API allows a burst of requests per person; the app's first load can use it all. A "slow down" answer is retried.
async function ask(srv, method, path, body) {
  for (let i = 0; ; i++) {
    try { return await srv.request(method, path, body); } catch (e) {
      if (!(e && e.status === 429) || i >= 5) throw e;
      await new Promise((r) => setTimeout(r, 1500 * (i + 1)));
    }
  }
}

// Runs jobs at most n at a time, so loading many forums does not use up the request burst at once.
async function pool(items, n, job) {
  const out = new Array(items.length);
  let next = 0;
  await Promise.all(Array.from({ length: Math.min(n, items.length) }, async () => {
    while (next < items.length) { const i = next++; out[i] = await job(items[i], i); }
  }));
  return out;
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
const cap = (r) => ((r || 'member')[0].toUpperCase() + (r || 'member').slice(1));
const localId = (c) => LOCAL_ID[c.name] || c.id;
const serverId = (srv, local) => (srv.forumServerId && srv.forumServerId.get(local)) || '';
export const isLive = (srv, local) => !!(srv && serverId(srv, local));
export const liveForum = (logic, local) => (logic.state.forumsLive || []).find((f) => f.id === local) || null;

async function allMessages(srv, conv) {
  const out = [];
  let before = '';
  while (out.length < MAX_MESSAGES) {
    // a failed page fails the whole read, so a reload keeps what is on screen instead of showing an empty forum
    const { messages } = await ask(srv, 'GET', `/conversations/${conv}/messages?limit=${PAGE}${before ? '&before=' + before : ''}`);
    out.push(...messages);
    if (messages.length < PAGE) break;
    before = messages[messages.length - 1].id;
  }
  return out;
}

// ---- posting rules ----
// The server limits who may post by the organization's roles (keys); the screens show role names. Which roles exist is data.
const roleNames = (logic) => Object.fromEntries((logic.state.serverRoles || []).map((r) => [r.id, r.name]));
const forumRoleOf = (role) => (role === 'owner' || role === 'admin' ? 'Admin' : role === 'moderator' ? 'Moderator' : role === 'guest' ? 'Guest' : 'Member');

// Same decision the server makes (it enforces it on every send). names: role key -> role name, for the message.
export function mayPost(rules, person, roleKey, names = {}) {
  const r = rules || { post_mode: 'everyone', roles: [], blocked: [], allowed: [] };
  if ((r.blocked || []).includes(person)) return { allowed: false, reason: 'You have been blocked from chatting in this forum.' };
  if ((r.allowed || []).includes(person)) return { allowed: true, reason: '' };
  if (r.post_mode === 'list') return { allowed: false, reason: 'Chat is locked in this forum. Only admins can post.' };
  if (r.post_mode === 'roles' && !(r.roles || []).includes(roleKey)) {
    const list = (r.roles || []).map((k) => names[k]).filter(Boolean);
    return { allowed: false, reason: 'Chat in this forum is limited to ' + (list.length ? list.join(', ') : 'nobody') + '.' };
  }
  return { allowed: true, reason: '' };
}

// The rules as the chat-rules screen shows them: { mode: everyone | roles | locked, roles: [role names], blocked: [names], allowed: [names] }.
export function policyOf(f, names = {}) {
  const r = f.rules || { post_mode: 'everyone', roles: [], blocked: [], allowed: [] };
  const name = (id) => (f.memberList.find((m) => m.person === id) || {}).name;
  const admins = new Set(f.memberList.filter((m) => m.role === 'owner' || m.role === 'admin').map((m) => m.person));
  return {
    mode: r.post_mode === 'list' ? 'locked' : r.post_mode,
    roles: (r.roles || []).map((k) => names[k]).filter(Boolean),
    blocked: (r.blocked || []).map(name).filter(Boolean),
    // in a locked forum the admins are written into the list; the screen shows only the people allowed beyond them
    allowed: (r.allowed || []).filter((id) => r.post_mode !== 'list' || !admins.has(id)).map(name).filter(Boolean)
  };
}

// patch = any of { mode, roles (role names), blocked, allowed } in the screen's terms. Saves the whole rule set.
export async function saveRules(logic, srv, local, patch) {
  const f = liveForum(logic, local);
  if (!f) return;
  const names = roleNames(logic);
  const keyOf = Object.fromEntries(Object.entries(names).map(([k, n]) => [n, k]));
  const cur = policyOf(f, names);
  const next = { ...cur, ...patch };
  const id = (n) => (f.memberList.find((m) => m.name === n) || {}).person;
  const admins = f.memberList.filter((m) => m.role === 'owner' || m.role === 'admin').map((m) => m.person);
  const allowed = [...new Set([...next.allowed.map(id).filter(Boolean), ...(next.mode === 'locked' ? admins : [])])];
  const body = {
    post_mode: next.mode === 'locked' ? 'list' : next.mode,
    roles: next.mode === 'roles' ? next.roles.map((n) => keyOf[n]).filter(Boolean) : [],
    blocked: next.blocked.map(id).filter(Boolean), allowed
  };
  logic.setState((st) => ({ forumsLive: st.forumsLive.map((x) => (x.id === local ? { ...x, rules: body } : x)) }));
  try { await ask(srv, 'PUT', `/conversations/${serverId(srv, local)}/rules`, body); } catch (e) { await reloadForum(logic, srv, local); throw e; }
}

// ---- reading ----

// A shared wall post carries copies of its pictures (made when it was shared); they open like any attachment here.
function sharedOf(sh, todo) {
  if (!sh || !sh.snap || !sh.snap.media) return sh || null;
  return { ...sh, snap: { ...sh.snap, media: sh.snap.media.filter((f) => f.fileId || !/^blob:/.test(f.url || '')).map((f) => { if (!f.fileId) return f; todo.push(f); return { ...f, url: '', ref: f }; }) } };
}

// Turns the sealed messages of one forum (newest first) into the chat items the panels show (newest first).
function toItems(srv, opened) {
  const byId = new Map(opened.map((x) => [x.m.id, x]));
  const children = new Map();
  opened.forEach((x) => { if (x.m.reply_to) children.set(x.m.reply_to, (children.get(x.m.reply_to) || 0) + 1); });
  const nameOf = (m) => (m.sender_account === srv.me.account ? 'Sam' : m.sender_name);
  const snippet = (pl) => { const t = (pl && (pl.text || (pl.file && pl.file[0] && pl.file[0].name))) || ''; return t.length > 90 ? t.slice(0, 90) + '…' : t; };
  const todo = [];
  const items = opened.filter((x) => x.m.type === 'text').map(({ m, pl }) => {
    const mine = m.sender_account === srv.me.account;
    const parent = m.reply_to ? byId.get(m.reply_to) : null;
    const file = (pl.file || []).map((f) => { const it = { name: f.name, meta: f.meta || '', isImg: !!f.isImg, isVideo: !!f.isVideo, url: '', fileId: f.fileId, ref: f }; todo.push(f); return it; });
    if (pl.note) file.push({ name: (pl.note.title || 'Untitled') + '.note', meta: 'Note · ' + (pl.note.perm === 'edit' ? 'Can edit' : 'Read only'), noteId: pl.note.conv, shareId: pl.note.id, url: '' });
    return {
      id: m.id, mid: m.id, noteShare: pl.note || null, who: nameOf(m), ...(mine ? { initials: ME.initials, color: ME.color } : lookOf(srv, m.sender_person, m.sender_name)),
      time: ago(m.created_at), at: m.created_at, text: pl.text || '', file: file.length ? file : undefined, shared: sharedOf(pl.shared, todo),
      replyTo: parent && parent.m.type === 'text' ? { who: parent.m.sender_account === srv.me.account ? 'You' : parent.m.sender_name, snippet: snippet(parent.pl) } : null, replies: children.get(m.id) || 0
    };
  });
  return { items, todo };
}

// Polls are messages too: a poll, one vote per person (the latest counts) and lock / unlock notes that point back at it.
function toPolls(srv, local, opened, members) {
  const mine = (m) => m.sender_account === srv.me.account;
  const asc = opened.slice().sort((a, b) => (a.m.id < b.m.id ? -1 : 1));
  return asc.filter((x) => x.m.type === 'poll').map(({ m, pl }) => {
    const votes = new Map();    // account -> { option, mid }
    let locked = false;
    asc.forEach((x) => {
      if (x.m.reply_to !== m.id) return;
      // one vote per person (their latest), only for a real option, and none cast while the poll was locked
      if (x.m.type === 'vote' && !locked && (pl.options || []).some((o) => o.id === x.pl.option)) votes.set(x.m.sender_account, { option: x.pl.option, mid: x.m.id });
      if (x.m.type === 'poll_lock') locked = x.pl.locked !== false;
    });
    const my = votes.get(srv.me.account);
    return {
      id: m.id, mid: m.id, forum: local, question: pl.question || '', live: true,
      options: (pl.options || []).map((o) => ({ id: o.id, label: o.label, votes: [...votes.entries()].filter(([acct, v]) => v.option === o.id && acct !== srv.me.account).length })),
      by: mine(m) ? 'You' : m.sender_name, at: new Date(m.created_at).getTime(), locked, blocked: [], myVote: my ? my.option : null, myVoteMid: my ? my.mid : '', members
    };
  }).reverse();
}

// What this device holds of each forum, by server id: the conversation, its rules, every message read so far (by id) and
// the catch-up marker. A reload asks the server only for what changed after the marker.
const held = (srv) => srv.forumHeld || (srv.forumHeld = new Map());
const WANTED = ['text', 'poll', 'vote', 'poll_lock'];
const SYNC_PAGES = 5;   // more changed than this since the last look: re-read the history instead

// A message never changes once sent, so each is decrypted once and remembered (only ones that opened).
async function openAll(srv, msgs) {
  const cache = srv.forumPayloads || (srv.forumPayloads = new Map());
  return Promise.all(msgs.map(async (m) => {
    if (cache.has(m.id)) return { m, pl: cache.get(m.id) };
    const pl = await srv.vault.openMessage(m);
    if (pl) cache.set(m.id, pl);
    return { m, pl };
  }));
}

// The catch-up marker after a full read: a little before the newest message, so nothing stored around it is skipped.
const markerAfter = (msgs) => (msgs.length ? new Date(new Date(msgs[0].created_at).getTime() - 120000).toISOString() : '2000-01-01T00:00:00Z');

const fetchMeta = (srv, sid) => Promise.all([
  ask(srv, 'GET', '/conversations/' + sid),
  ask(srv, 'GET', `/conversations/${sid}/rules`).catch(() => null)
]);

// Reads a forum from scratch: its details, rules and recent history.
async function fetchFull(srv, sid) {
  const [[detail, rules], msgs] = await Promise.all([fetchMeta(srv, sid), allMessages(srv, sid)]);
  const h = { detail, rules, msgs: new Map(msgs.map((m) => [m.id, m])), since: markerAfter(msgs) };
  held(srv).set(sid, h);
  return h;
}

// Brings a held forum up to date from its marker. Returns the messages that are new to this device, or null when a full
// read is the better choice (too much changed, or the server cannot answer catch-up reads).
async function catchUp(srv, sid, h) {
  const fresh = [];
  let since = h.since, after = '';
  for (let i = 0; i < SYNC_PAGES; i++) {
    const page = await ask(srv, 'GET', `/conversations/${sid}/changes?since=${encodeURIComponent(since)}${after ? '&after=' + after : ''}&limit=200`).catch((e) => {
      if (e && (e.status === 404 || e.status === 405)) return null;
      throw e;
    });
    if (!page) return null;
    for (const m of page.messages) {
      if (!h.msgs.has(m.id) && !m.deleted) fresh.push(m);
      h.msgs.set(m.id, m);
    }
    since = page.next_since; after = page.next_after || '';
    if (!page.more) {
      // keep only the newest ones, as a full read would
      if (h.msgs.size > MAX_MESSAGES) [...h.msgs.keys()].sort().slice(0, h.msgs.size - MAX_MESSAGES).forEach((id) => h.msgs.delete(id));
      h.since = since;
      return fresh;
    }
  }
  return null;
}

// Turns what is held of a forum into the forum the screens show.
async function build(srv, h) {
  const { detail, rules } = h;
  const msgs = [...h.msgs.values()].sort((a, b) => (a.id < b.id ? 1 : -1));
  const opened = (await openAll(srv, msgs.filter((m) => !m.deleted && WANTED.includes(m.type)))).filter((x) => x.pl);
  const { items, todo } = toItems(srv, opened);
  const members = (detail.members || []).map((x) => {
    const d = srv.dirByUid && srv.dirByUid.get(x.person);
    return { person: x.person, account: x.account, name: x.account === srv.me.account ? srv.me.name : x.name, initials: x.initials, color: x.color, role: x.role, forumRole: forumRoleOf(x.role), roleKey: x.account === srv.me.account ? srv.me.role : (d ? d.role : ''), title: d ? d.title : '' };
  });
  const local = localId(detail);
  const live = msgs.filter((m) => !m.deleted);
  return {
    forum: {
      id: local, sid: detail.id, live: true, name: detail.name, initials: initialsOf(detail.name) || detail.initials,
      color: detail.color || '#7CC4FF', unread: detail.unread || 0, members: detail.member_count || members.length, online: 0, privacy: PRIVACY[detail.privacy] || 'Invite only', about: detail.about || '',
      myRole: cap(detail.my_role), myForumRole: forumRoleOf(detail.my_role), epoch: detail.epoch, created: detail.created_at, memberList: members, updates: items, polls: toPolls(srv, local, opened, detail.member_count || members.length),
      rules: rules || { post_mode: 'everyone', roles: [], blocked: [], allowed: [] }, newest: live.length ? live[0].id : ''
    },
    todo
  };
}

// A forum this device already holds is caught up rather than read again (a reload of the list after joining a forum).
async function fetchForum(srv, c) {
  const h = held(srv).get(c.id);
  if (h) {
    const [[detail, rules], fresh] = await Promise.all([fetchMeta(srv, c.id), catchUp(srv, c.id, h)]);
    h.detail = detail; h.rules = rules || h.rules;
    if (fresh !== null) return build(srv, h);
  }
  return build(srv, await fetchFull(srv, c.id));
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
  const on = srv.online || new Set([srv.me.person]);
  logic.setState((st) => ({ forumsLive: (st.forumsLive || []).map((f) => ({ ...f, online: f.memberList.filter((m) => on.has(m.person)).length, memberList: f.memberList.map((m) => ({ ...m, online: on.has(m.person) })) })) }));
}

function loadFiles(logic, srv, todo) {
  todo.forEach((f) => files.openUrl(srv, f).then((url) => {
    const put = (list) => list.map((x) => (x.fileId === f.fileId ? { ...x, url } : x));
    logic.setState((st) => ({ forumsLive: (st.forumsLive || []).map((fo) => ({ ...fo, updates: fo.updates.map((u) => {
      let n = u;
      if (n.file && n.file.some((x) => x.fileId === f.fileId)) n = { ...n, file: put(n.file) };
      if (n.shared && n.shared.snap && (n.shared.snap.media || []).some((x) => x.fileId === f.fileId)) n = { ...n, shared: { ...n.shared, snap: { ...n.shared.snap, media: put(n.shared.snap.media) } } };
      return n;
    }) })) }));
  }).catch(() => {}));
}

export async function load(logic, srv) {
  const { conversations } = await ask(srv, 'GET', '/conversations?kind=forum&limit=200');
  const prev = new Map((logic.state.forumsLive || []).map((f) => [f.sid, f]));
  // a forum that fails to load keeps what was shown before (or is left out until the next reload) instead of failing them all
  const built = (await pool(conversations, 8, (c) => fetchForum(srv, c).catch((e) => {
    console.warn('Could not load forum ' + c.id + ':', e);
    return prev.has(c.id) ? { forum: prev.get(c.id), todo: [] } : null;
  }))).filter(Boolean);
  built.sort((a, b) => (a.forum.created < b.forum.created ? -1 : 1));
  srv.forumServerId = new Map(built.map((b) => [b.forum.id, b.forum.sid]));
  srv.forumLocalId = new Map(built.map((b) => [b.forum.sid, b.forum.id]));
  srv.forumConv = Object.fromEntries(built.map((b) => [b.forum.id, b.forum.sid]));   // used when a memo or announcement is sent to a forum
  logic.setState({ forumsLive: built.map((b) => b.forum) });
  built.forEach((b) => loadFiles(logic, srv, b.todo));
  watch(logic, srv);   // subscribe before presence so no message is missed while it loads
  loadPresence(logic, srv).catch(() => {});
}

// Reloads one forum (its messages, members, polls, rules and unread count) and, if you are looking at it, marks it read.
// One reload per forum at a time: a reload asked for while one runs waits and runs once after it, so a burst of messages
// costs two reads, not one each, and an older read can never overwrite a newer one.
// opts.meta: also re-read the forum's details and rules (members, name, unread count). Without it only messages are
// brought up to date, which is what a new or removed message needs.
// opts.incoming: messages the socket delivered whole (sealed, with their ids); they are applied without asking the server.
// opts.gone: ids of messages removed. Both make the forum current at once, so a message appears as soon as it arrives.
export function reloadForum(logic, srv, local, opts = { meta: true }) {
  const q = srv.forumReloads || (srv.forumReloads = new Map());
  const cur = q.get(local);
  if (cur) {
    cur.again = true; cur.meta = cur.meta || !!opts.meta;
    cur.incoming.push(...(opts.incoming || [])); (opts.gone || []).forEach((id) => cur.gone.add(id));
    return cur.done;
  }
  const entry = { again: false, meta: !!opts.meta, incoming: [...(opts.incoming || [])], gone: new Set(opts.gone || []) };
  entry.done = (async () => {
    try {
      do {
        const run = { meta: entry.meta, incoming: entry.incoming, gone: entry.gone };
        entry.again = false; entry.meta = false; entry.incoming = []; entry.gone = new Set();
        await reloadOnce(logic, srv, local, run);
      } while (entry.again);
    } finally { q.delete(local); }
  })();
  q.set(local, entry);
  return entry.done;
}

async function reloadOnce(logic, srv, local, { meta, incoming = [], gone = new Set() }) {
  const sid = serverId(srv, local);
  if (!sid) return;
  let h = held(srv).get(sid);
  let fresh = null;
  if (h && !meta && (incoming.length || gone.size)) {
    // the socket handed us exactly what changed: no request, no waiting
    fresh = [];
    for (const m of incoming) {
      if (m.conversation && m.conversation !== sid) continue;
      if (!h.msgs.has(m.id) && !m.deleted) fresh.push(m);
      h.msgs.set(m.id, m);
    }
    gone.forEach((id) => { const old = h.msgs.get(id); if (old) h.msgs.set(id, { ...old, deleted: true, envelope: '' }); });
    if (h.msgs.size > MAX_MESSAGES) [...h.msgs.keys()].sort().slice(0, h.msgs.size - MAX_MESSAGES).forEach((id) => h.msgs.delete(id));
  } else if (h) {
    // details and the message catch-up run side by side; if catching up is not possible, read the history again
    const [m, f] = await Promise.all([meta ? fetchMeta(srv, sid) : null, catchUp(srv, sid, h)]);
    if (m) { h.detail = m[0]; h.rules = m[1] || h.rules; }
    fresh = f;
    incoming.forEach((m) => { if (!h.msgs.has(m.id) && !m.deleted) fresh && fresh.push(m); h.msgs.set(m.id, m); });
  }
  if (!h || fresh === null) h = await fetchFull(srv, sid);
  const { forum, todo } = await build(srv, h);
  const open = logic.state.forum === local && logic.state.panelKind === 'forum' && !['closed', 'min'].includes(logic.state.mode);
  const prev = liveForum(logic, local);
  // without fresh details, the unread count moves by the new messages from others
  const others = (fresh || []).filter((m) => m.type === 'text' && m.sender_account !== srv.me.account).length;
  const unread = open ? 0 : meta || !prev ? forum.unread : prev.unread + others;
  if (!meta) h.detail = { ...h.detail, unread };
  const on = srv.online || new Set();
  const merged = { ...forum, online: prev ? prev.online : 0, unread, memberList: forum.memberList.map((m) => ({ ...m, online: on.has(m.person) })) };
  logic.setState((st) => ({ forumsLive: (st.forumsLive || []).map((f) => (f.id === local ? merged : f)) }));
  applyOnline(logic, srv);
  loadFiles(logic, srv, todo);
  if (open && forum.newest) ask(srv, 'POST', `/conversations/${sid}/read`, { message: forum.newest }).catch(() => {});
}

// You opened a forum: its unread count goes to zero here and on the server.
export function markRead(logic, srv, local) {
  const sid = serverId(srv, local);
  if (!sid) return;
  const f = liveForum(logic, local);
  if (!f || !f.newest) return;
  if (f.unread) logic.setState((st) => ({ forumsLive: st.forumsLive.map((x) => (x.id === local ? { ...x, unread: 0 } : x)) }));
  ask(srv, 'POST', `/conversations/${sid}/read`, { message: f.newest }).catch(() => {});
}

// ---- messages ----

const post = async (srv, sid, type, payload, replyTo) => {
  const sealed = await srv.vault.seal(sid, payload);
  return ask(srv, 'POST', `/conversations/${sid}/messages`, { ...sealed, type, ...(replyTo ? { reply_to: replyTo } : {}) });
};
const unpost = (srv, sid, mid) => ask(srv, 'DELETE', `/conversations/${sid}/messages/${mid}`);

// m = { text, files: [File], replyTo: message id | '', shared: { sid, postId, snap } | null }
export async function send(logic, srv, local, m) {
  const sid = serverId(srv, local);
  if (!sid) throw new Error('This forum is not connected to the server.');
  const refs = [];
  for (const f of m.files || []) refs.push(await files.upload(srv, sid, f));
  // The pictures of a shared wall post belong to the wall, whose key forum members may not hold. Copy each into this forum
  // (decrypt it here, encrypt it again for the forum) so everyone in the forum can open it.
  let shared;
  if (m.shared) {
    const media = [];
    for (const it of (m.shared.snap && m.shared.snap.media) || []) {
      try {
        const src = it.ref ? await files.openUrl(srv, it.ref) : (/^blob:|^\//.test(it.url || '') ? it.url : '');
        if (!src) continue;
        const blob = await (await fetch(src)).blob();
        const up = await files.upload(srv, sid, new File([blob], it.name || 'picture', { type: it.type || blob.type || '' }));
        media.push({ fileId: up.fileId, name: up.name, meta: up.meta, type: up.type, isImg: up.isImg, isVideo: up.isVideo, key: up.key });
      } catch (e) { /* a picture that cannot be copied is left out; the text still goes */ }
    }
    shared = { ...m.shared, snap: m.shared.snap ? { ...m.shared.snap, media } : undefined };
  }
  const payload = { text: m.text || '' };
  if (refs.length) payload.file = refs.map(({ fileId, name, meta, type, isImg, isVideo, key }) => ({ fileId, name, meta, type, isImg, isVideo, key }));
  if (shared) payload.shared = shared;
  if (m.note) payload.note = m.note;   // a note shared into the forum: members open it from the Files tab
  await post(srv, sid, 'text', payload, m.replyTo);
  await reloadForum(logic, srv, local, { meta: false });
}

export async function remove(logic, srv, local, mid) {
  await unpost(srv, serverId(srv, local), mid);
  await reloadForum(logic, srv, local, { meta: false });
}

// ---- polls ----

export function findPoll(logic, pollId) {
  for (const f of logic.state.forumsLive || []) { const p = f.polls.find((x) => x.id === pollId); if (p) return { forum: f, poll: p }; }
  return null;
}

export async function createPoll(logic, srv, local, question, options) {
  await post(srv, serverId(srv, local), 'poll', { question, options: options.map((label, i) => ({ id: 'o' + (i + 1), label })) });
  await reloadForum(logic, srv, local, { meta: false });
}

// Tapping your vote again takes it back; tapping another option changes it.
export async function votePoll(logic, srv, pollId, optionId) {
  const hit = findPoll(logic, pollId);
  if (!hit) return;
  const { forum, poll } = hit;
  const sid = serverId(srv, forum.id);
  if (poll.myVoteMid) await unpost(srv, sid, poll.myVoteMid);
  if (poll.myVote !== optionId) await post(srv, sid, 'vote', { option: optionId }, poll.mid);
  await reloadForum(logic, srv, forum.id, { meta: false });
}

export async function lockPoll(logic, srv, pollId) {
  const hit = findPoll(logic, pollId);
  if (!hit) return;
  await post(srv, serverId(srv, hit.forum.id), 'poll_lock', { locked: !hit.poll.locked }, hit.poll.mid);
  await reloadForum(logic, srv, hit.forum.id, { meta: false });
}

export async function deletePoll(logic, srv, pollId) {
  const hit = findPoll(logic, pollId);
  if (!hit) return;
  await unpost(srv, serverId(srv, hit.forum.id), hit.poll.mid);
  await reloadForum(logic, srv, hit.forum.id, { meta: false });
}

// ---- making forums and inviting people ----

// "Brian Kiprop, amina@company.com" -> person ids. Anything that matches nobody is reported, not skipped.
function resolvePeople(srv, text, logic) {
  const ids = [], missing = [];
  String(text || '').split(/[,;\n]/).map((t) => t.trim()).filter(Boolean).forEach((t) => {
    const low = t.toLowerCase();
    const d = (logic.dir || []).find((x) => x.name.toLowerCase() === low || (x.email && x.email.toLowerCase() === low));
    if (d) ids.push(d.uid); else missing.push(t);
  });
  if (missing.length) throw new Error('Nobody in the directory matches ' + missing.join(', ') + '.');
  return [...new Set(ids)];
}

// f = { name, about, privacy: 'invite' | 'open', members: 'names or emails, comma separated' }
export async function createForum(logic, srv, f) {
  const name = (f.name || '').trim();
  if (!name) throw new Error('Give the forum a name.');
  const typed = String(f.members || '').trim() ? resolvePeople(srv, f.members, logic) : [];
  const members = [...new Set([...(f.memberIds || []), ...typed])].filter((id) => id !== srv.me.person);
  const { devices } = await ask(srv, 'POST', '/e2ee/people/devices', { people: [...new Set([srv.me.person, ...members])] });
  const key = newConversationKey();
  const keys = await Promise.all(devices.map(async (d) => ({ epoch: 1, device: d.id, wrapped_key: b64(await wrap(key, unb64(d.identity_key))) })));
  const color = PALETTE[(logic.state.forumsLive || []).length % PALETTE.length];
  const made = await ask(srv, 'POST', '/conversations', { kind: 'forum', name, about: (f.about || '').trim(), color, privacy: f.privacy === 'open' ? 'open' : 'invite', members, keys });
  await load(logic, srv);
  return srv.forumLocalId.get(made.id) || made.id;
}

// names = 'names or emails, comma separated'. People already in the forum are left as they are.
export async function invite(logic, srv, local, names, ids = []) {
  const f = liveForum(logic, local);
  if (!f) throw new Error('This forum is not connected to the server.');
  const sid = f.sid;
  const have = new Set(f.memberList.map((m) => m.person));
  const people = [...new Set([...ids, ...(String(names || '').trim() ? resolvePeople(srv, names, logic) : [])])].filter((id) => !have.has(id));
  if (!people.length) throw new Error('They are all in this forum already.');
  const conv = await srv.vault.conversation(sid, true);
  const key = await srv.vault.keyFor(sid, conv.epoch);
  if (!key) throw new Error('This device does not have the key for this forum yet.');
  const { devices } = await ask(srv, 'POST', '/e2ee/people/devices', { people });
  const keys = await srv.vault.wrapFor(sid, conv.epoch, key, devices);
  await ask(srv, 'POST', `/conversations/${sid}/members`, { people, device: srv.vault.deviceId, keys });
  srv.vault.convs.delete(sid);
  await reloadForum(logic, srv, local);
  return people.length;
}

// ---- realtime ----

function watch(logic, srv) {
  const rt = srv.rt || (srv.rt = connectRealtime(srv));   // whichever widget loads first makes the socket
  if (!srv.forumsWatching) {
    srv.forumsWatching = true;
    const timers = new Map();
    const metaDue = new Set();
    const later = (local, meta = true) => {
      if (meta) metaDue.add(local);
      clearTimeout(timers.get(local));
      timers.set(local, setTimeout(() => { const m = metaDue.delete(local); reloadForum(logic, srv, local, { meta: m }).catch(() => {}); }, 250));
    };
    const forumOf = (ev) => (ev.channel && srv.forumLocalId ? srv.forumLocalId.get(String(ev.channel).split(':').pop()) : null);
    rt.on('message.new', (ev) => {
      const f = forumOf(ev);
      if (!f) return;
      const m = ev.data || {};
      // a complete message (it carries its sealed envelope) is applied right away; anything else falls back to a catch-up read
      if (m.id && m.envelope) reloadForum(logic, srv, f, { meta: false, incoming: [m] }).catch(() => {}); else later(f, false);
      if (m.sender_account === srv.me.account || ['reaction', 'vote', 'poll_lock', 'typing', 'system'].includes(m.type)) return;
      const st = logic.state;
      const onScreen = st.forum === f && st.panelKind === 'forum' && !['closed', 'min'].includes(st.mode);
      try { (onScreen ? messageTone : alertTone)(); } catch (e) { /* sound is optional */ }
    });
    rt.on('message.deleted', (ev) => { const f = forumOf(ev); const d = ev.data || {}; if (f && d.id) reloadForum(logic, srv, f, { meta: false, gone: [d.id] }).catch(() => {}); else if (f) later(f, false); });
    ['member.added', 'member.removed', 'member.joined'].forEach((t) => rt.on(t, (ev) => { const f = forumOf(ev); if (f) later(f); }));
    rt.on('conversation.new', () => { load(logic, srv).catch(() => {}); });
    rt.on('conversation.removed', () => { load(logic, srv).catch(() => {}); });
    rt.on('presence', (ev) => {
      const d = ev.data || {};
      if (!d.person) return;
      if (d.online || d.person === srv.me.person) srv.online.add(d.person); else srv.online.delete(d.person);   // this tab holds a socket, so you are online
      applyOnline(logic, srv);
    });
    // after a reconnect, catch up on what was missed: a few forums at a time, not all at once
    rt.on('open', () => { pool((logic.state.forumsLive || []).map((f) => f.id), 3, (id) => reloadForum(logic, srv, id).catch(() => {})); });
  }
  (logic.state.forumsLive || []).forEach((f) => rt.sub(f.sid));
}
