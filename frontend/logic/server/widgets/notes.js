// Notes: notebooks and notes (the Notes page and the My notes widget), and notes shared with people or forums.
// Your own notes live in your private vault (a conversation only you are in); each note and notebook is a sealed record there.
// Sharing a note moves it into its own small conversation (a "shared note") that holds the note and whoever it is shared with:
// the owner decides who is in and whether they can only read or also edit. Everything is end-to-end encrypted.
// The Notes page edits state.notes / state.noteBooks directly; this module watches those edits and writes the differences.
import { newConversationKey, wrap, newId, b64, unb64 } from '../e2ee.js';
import { connectRealtime } from '../realtime.js';
import * as forumsW from './forums.js';

export const name = 'notes';

const QUICK = 'Quick notes';
const COLORS = ['#FFD66B', '#7CC4FF', '#C9B6FF', '#6EE7C8', '#F49AC1', '#FFB86B'];

async function ask(srv, method, path, body) {
  for (let i = 0; ; i++) {
    try { return await srv.request(method, path, body); } catch (e) {
      if (!(e && e.status === 429) || i >= 5) throw e;
      await new Promise((r) => setTimeout(r, 1500 * (i + 1)));
    }
  }
}

async function myKeys(srv, people) {
  const { devices } = await ask(srv, 'POST', '/e2ee/people/devices', { people: [...new Set([srv.me.person, ...people])] });
  const key = newConversationKey();
  return Promise.all(devices.map(async (d) => ({ epoch: 1, device: d.id, wrapped_key: b64(await wrap(key, unb64(d.identity_key))) })));
}

async function records(srv, conv) {
  const out = [];
  let before = '';
  for (let i = 0; i < 20; i++) {
    const { records: page } = await ask(srv, 'GET', `/containers/${conv}/records?limit=500${before ? '&before=' + before : ''}`);
    out.push(...page);
    if (page.length < 500) break;
    before = page[page.length - 1].id;
  }
  return out.filter((r) => !r.deleted);
}

const sealRecord = async (srv, conv, payload, id) => srv.vault.seal(conv, payload, id);
const put = async (srv, conv, collection, id, payload, meta, create) => {
  const sealed = await sealRecord(srv, conv, payload, id);
  const body = { collection, epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope, meta: meta || {} };
  if (create) await ask(srv, 'POST', `/containers/${conv}/records`, { id, ...body });
  else await ask(srv, 'PUT', `/containers/${conv}/records/${id}`, body);
};

const nameOf = (srv, person) => (person === srv.me.person ? srv.me.name : (srv.nameOf && srv.nameOf.get(person)) || '');

// ---- reading ----

async function readVault(srv) {
  let vault = (await ask(srv, 'GET', '/conversations?kind=note&limit=5')).conversations[0];
  if (!vault) vault = await ask(srv, 'POST', '/conversations', { kind: 'note', keys: await myKeys(srv, []) });
  srv.noteVault = vault.id;
  const recs = await records(srv, vault.id);
  const opened = await Promise.all(recs.map(async (r) => ({ r, pl: await srv.vault.openRecord(r) })));
  const books = opened.filter((x) => x.r.collection === 'notebook' && x.pl).map(({ r, pl }) => ({ id: r.id, name: pl.name || 'Notebook', color: pl.color || COLORS[0], _t: Date.parse(r.created_at) }))
    .sort((a, b) => a._t - b._t).map(({ _t, ...b }) => b);
  const notes = opened.filter((x) => x.r.collection === 'note' && x.pl).map(({ r, pl }) => ({ id: r.id, book: r.meta.book || '', title: pl.title || '', body: pl.body || '', updated: Date.parse(r.updated_at || r.created_at), pinned: !!r.meta.pinned }));
  return { books, notes };
}

// One shared note: its record, its members and what I may do.
async function readShared(srv, c) {
  const detail = await ask(srv, 'GET', '/conversations/' + c.id);
  const recs = await records(srv, c.id);
  const r = recs.find((x) => x.collection === 'note');
  if (!r) return null;
  const pl = await srv.vault.openRecord(r);
  if (!pl) return null;
  const members = detail.members || [];
  const owner = members.find((m) => m.role === 'owner');
  const mineRole = (members.find((m) => m.person === srv.me.person) || {}).role || '';
  const iOwn = !!owner && owner.person === srv.me.person;
  const n = { id: r.id, container: c.id, book: r.meta.book || '', title: pl.title || '', body: pl.body || '', updated: Date.parse(r.updated_at || r.created_at), pinned: !!r.meta.pinned, meta: r.meta, sharesRaw: pl.shares || [] };
  if (iOwn) {
    n.shares = (pl.shares || []).map((s) => ({ ...s, perm: s.kind === 'person' ? roleToPerm((members.find((m) => m.person === s.to) || {}).role) || s.perm : s.perm }));
  } else {
    n.owner = owner ? (owner.person === srv.me.person ? srv.me.name : owner.name) : 'Someone';
    n.access = mineRole === 'admin' ? 'edit' : 'view';
    n.shares = [];
    n.book = '';
  }
  return n;
}
const roleToPerm = (role) => (role === 'admin' ? 'edit' : role === 'member' ? 'view' : '');

function snapshot(notes, books) {
  const snap = new Map();
  books.forEach((b) => snap.set('b:' + b.id, JSON.stringify({ name: b.name, color: b.color })));
  notes.forEach((n) => snap.set('n:' + n.id, JSON.stringify({ book: n.book, title: n.title, body: n.body, pinned: !!n.pinned, container: n.container || '', access: n.access || '' })));
  return snap;
}

export async function load(logic, srv) {
  const vault = await readVault(srv);
  let books = vault.books;
  if (!books.length) {
    const quick = { id: newId(), name: QUICK, color: COLORS[0] };
    await put(srv, srv.noteVault, 'notebook', quick.id, { name: quick.name, color: quick.color }, {}, true);
    books = [quick];
  }
  const { conversations } = await ask(srv, 'GET', '/conversations?kind=sharednote&limit=200');
  const shared = (await Promise.all(conversations.map((c) => readShared(srv, c).catch(() => null)))).filter(Boolean);
  srv.sharedConvs = new Set(shared.map((n) => n.container));
  const notes = vault.notes.concat(shared);
  srv.noteSnap = snapshot(notes, books);
  srv.applyingNotes = true;
  logic.setState({ noteBooks: books, notes });
  srv.applyingNotes = false;
  watch(logic, srv);
  const rt = srv.rt || (srv.rt = connectRealtime(srv));
  shared.forEach((n) => rt.sub(n.container));
}

// A shared note changed (or appeared or went away): refresh it, or the whole list when it is new.
async function reloadShared(logic, srv, conv) {
  let n = null;
  try { n = await readShared(srv, { id: conv }); } catch (e) { n = null; }
  srv.applyingNotes = true;
  logic.setState((st) => {
    const rest = st.notes.filter((x) => x.container !== conv);
    return { notes: n ? [n].concat(rest) : rest, noteEdit: st.noteEdit && !n && st.noteEdit.container === conv ? null : st.noteEdit };
  });
  srv.applyingNotes = false;
  if (n) { srv.noteSnap.set('n:' + n.id, snapshot([n], []).get('n:' + n.id)); srv.sharedConvs.add(conv); (srv.rt || {}).sub && srv.rt.sub(conv); } else { srv.sharedConvs.delete(conv); }
}

// ---- writing: what the page changed in state.notes / state.noteBooks ----

async function flush(logic, srv) {
  const { notes, noteBooks } = logic.state;
  const now = snapshot(notes, noteBooks);
  const before = srv.noteSnap;
  srv.noteSnap = now;
  // notebooks
  for (const b of noteBooks) {
    const k = 'b:' + b.id;
    if (before.get(k) === now.get(k)) continue;
    await put(srv, srv.noteVault, 'notebook', b.id, { name: b.name, color: b.color }, {}, !before.has(k));
  }
  for (const [k] of before) {
    if (!k.startsWith('b:') || now.has(k)) continue;
    await ask(srv, 'DELETE', `/containers/${srv.noteVault}/records/${k.slice(2)}`).catch(() => {});
  }
  // notes
  for (const n of notes) {
    const k = 'n:' + n.id;
    if (before.get(k) === now.get(k)) continue;
    if (n.container) {
      if (n.owner && n.access !== 'edit') continue;                       // read only: nothing to save
      const meta = n.owner ? (n.meta || {}) : { ...(n.book ? { book: n.book } : {}), ...(n.pinned ? { pinned: true } : {}) };
      await put(srv, n.container, 'note', n.id, { title: n.title, body: n.body, shares: n.owner ? n.sharesRaw || [] : n.shares || [] }, meta, false);
    } else {
      await put(srv, srv.noteVault, 'note', n.id, { title: n.title, body: n.body }, { ...(n.book ? { book: n.book } : {}), ...(n.pinned ? { pinned: true } : {}) }, !before.has(k));
    }
  }
  for (const [k] of before) {
    if (!k.startsWith('n:') || now.has(k)) continue;
    const old = JSON.parse(before.get(k));
    const id = k.slice(2);
    if (old.container && !old.access) await ask(srv, 'DELETE', `/containers/${old.container}/records/${id}`).catch(() => {});   // I own it: it goes for everyone
    else if (!old.container) await ask(srv, 'DELETE', `/containers/${srv.noteVault}/records/${id}`).catch(() => {});
  }
}

function watch(logic, srv) {
  if (srv.notesWatching) return;
  srv.notesWatching = true;
  let busy = false, again = false, timer = null;
  const run = async () => {
    if (busy) { again = true; return; }
    busy = true;
    try { await flush(logic, srv); } catch (e) { logic.notify(e.message || 'Could not save your notes.', '#FF9B8A'); await load(logic, srv).catch(() => {}); } finally { busy = false; if (again) { again = false; run(); } }
  };
  const orig = logic.setState.bind(logic);
  logic.setState = (update, done) => { orig(update, done); if (!srv.applyingNotes) { clearTimeout(timer); timer = setTimeout(run, 500); } };
  const rt = srv.rt || (srv.rt = connectRealtime(srv));
  const later = (() => { const t = new Map(); return (conv) => { clearTimeout(t.get(conv)); t.set(conv, setTimeout(() => reloadShared(logic, srv, conv).catch(() => {}), 300)); }; })();
  const convOf = (ev) => String(ev.channel || '').split(':').pop();
  ['record.saved', 'record.deleted'].forEach((t) => rt.on(t, (ev) => { const c = convOf(ev); if (srv.sharedConvs && srv.sharedConvs.has(c)) later(c); }));
  ['member.added', 'member.removed'].forEach((t) => rt.on(t, (ev) => { const c = convOf(ev); if (srv.sharedConvs && srv.sharedConvs.has(c)) later(c); }));
  rt.on('conversation.new', (ev) => { const d = ev.data || {}; if (d.conversation && !(srv.forumLocalId && srv.forumLocalId.has(d.conversation))) { srv.noteLoad = srv.noteLoad || 0; clearTimeout(srv.noteLoad); srv.noteLoad = setTimeout(() => load(logic, srv).catch(() => {}), 800); } });
  rt.on('conversation.removed', (ev) => { const d = ev.data || {}; if (srv.sharedConvs && srv.sharedConvs.has(d.conversation)) later(d.conversation); });
}

// ---- sharing ----

async function addMembers(srv, conv, people) {
  const c = await srv.vault.conversation(conv, true);
  const have = new Set((c.members || []).map((m) => m.person));
  const add = [...new Set(people)].filter((p) => p && !have.has(p));
  if (!add.length) return [];
  const key = await srv.vault.keyFor(conv, c.epoch);
  if (!key) throw new Error('This device does not have the key for this note yet.');
  const { devices } = await ask(srv, 'POST', '/e2ee/people/devices', { people: add });
  await ask(srv, 'POST', `/conversations/${conv}/members`, { people: add, device: srv.vault.deviceId, keys: await srv.vault.wrapFor(conv, c.epoch, key, devices) });
  srv.vault.convs.delete(conv);
  return add;
}

async function removeMember(srv, conv, person) {
  const c = await srv.vault.conversation(conv, true);
  if (!(c.members || []).some((m) => m.person === person)) return;
  const stay = (c.members || []).filter((m) => m.person !== person).map((m) => m.person);
  const { devices } = await ask(srv, 'POST', '/e2ee/people/devices', { people: stay });
  const keys = await srv.vault.wrapFor(conv, c.epoch + 1, newConversationKey(), devices);
  await ask(srv, 'DELETE', `/conversations/${conv}/members/${person}`, { device: srv.vault.deviceId, keys });
  srv.vault.keys.delete(conv); srv.vault.convs.delete(conv);
}

// Moves a note into a shared container the first time it is shared (a record id is unique, so it gets a new one); returns the
// note as it is after.
async function ensureShared(logic, srv, n) {
  if (n.container) return n;
  const made = await ask(srv, 'POST', '/conversations', { kind: 'sharednote', members: [], keys: await myKeys(srv, []) });
  const conv = made.id;
  const id = newId();
  const meta = { ...(n.book ? { book: n.book } : {}), ...(n.pinned ? { pinned: true } : {}) };
  await put(srv, conv, 'note', id, { title: n.title, body: n.body, shares: [] }, meta, true);
  await ask(srv, 'DELETE', `/containers/${srv.noteVault}/records/${n.id}`).catch(() => {});
  srv.sharedConvs.add(conv);
  (srv.rt || {}).sub && srv.rt.sub(conv);
  const next = { ...n, id, container: conv, sharesRaw: [], shares: [], meta };
  srv.applyingNotes = true;
  logic.setState((st) => ({ notes: st.notes.map((x) => (x.id === n.id ? next : x)), noteEdit: st.noteEdit && st.noteEdit.id === n.id ? { ...st.noteEdit, id } : st.noteEdit }));
  srv.applyingNotes = false;
  srv.noteSnap = snapshot(logic.state.notes, logic.state.noteBooks);
  return next;
}

async function saveShares(logic, srv, n, sharesRaw) {
  await put(srv, n.container, 'note', n.id, { title: n.title, body: n.body, shares: sharesRaw }, n.meta || {}, false);
  const next = { ...n, sharesRaw, shares: sharesRaw };
  srv.applyingNotes = true;
  logic.setState((st) => ({ notes: st.notes.map((x) => (x.id === n.id ? next : x)) }));
  srv.applyingNotes = false;
  srv.noteSnap = snapshot(logic.state.notes, logic.state.noteBooks);
  return next;
}

// s = { kind: 'person' | 'forum', to: person name | forum local id, perm: 'view' | 'edit' }
export async function share(logic, srv, note, s) {
  await flush(logic, srv).catch(() => {});   // an unsaved edit goes first
  let n = await ensureShared(logic, srv, logic.state.notes.find((x) => x.id === note.id) || note);
  let people = [], toName = '', label = '', mid = '';
  if (s.kind === 'person') {
    const pid = srv.personOf.get(s.to);
    if (!pid) throw new Error('Pick a person from the directory.');
    people = [pid]; toName = s.to;
  } else {
    const f = forumsW.liveForum(logic, s.to);
    if (!f) throw new Error('Choose a forum you are in.');
    people = f.memberList.map((m) => m.person).filter((p) => p !== srv.me.person);
    toName = f.name; label = s.to;
  }
  const added = await addMembers(srv, n.container, people);
  const role = s.perm === 'edit' ? 'admin' : 'member';
  for (const pid of people) await ask(srv, 'PATCH', `/conversations/${n.container}/members/${pid}`, { role }).catch(() => {});
  const entry = { id: 'sh' + newId().slice(-12), kind: s.kind, to: s.kind === 'person' ? people[0] : s.to, toName, perm: s.perm === 'edit' ? 'edit' : 'view', at: Date.now(), members: added, mid: '' };
  if (s.kind === 'forum') {
    await forumsW.send(logic, srv, s.to, { text: '', files: [], note: { conv: n.container, title: n.title || 'Untitled', perm: entry.perm, id: entry.id } });
    const f = forumsW.liveForum(logic, s.to);
    const msg = f && f.updates.find((u) => u.noteShare && u.noteShare.id === entry.id);
    entry.mid = msg ? msg.mid : '';
  }
  n = await saveShares(logic, srv, n, (n.sharesRaw || []).concat([entry]));
  return n;
}

export async function setPerm(logic, srv, note, shareId, perm) {
  const n = logic.state.notes.find((x) => x.id === note.id);
  const sh = (n.sharesRaw || []).find((x) => x.id === shareId);
  if (!sh) return;
  const people = sh.kind === 'person' ? [sh.to] : (sh.members || []);
  for (const pid of people) await ask(srv, 'PATCH', `/conversations/${n.container}/members/${pid}`, { role: perm === 'edit' ? 'admin' : 'member' }).catch(() => {});
  await saveShares(logic, srv, n, n.sharesRaw.map((x) => (x.id === shareId ? { ...x, perm } : x)));
}

export async function stopShare(logic, srv, note, shareId) {
  const n = logic.state.notes.find((x) => x.id === note.id);
  const sh = (n.sharesRaw || []).find((x) => x.id === shareId);
  if (!sh) return;
  const others = new Set((n.sharesRaw || []).filter((x) => x.id !== shareId).flatMap((x) => (x.kind === 'person' ? [x.to] : x.members || [])));
  const leaving = (sh.kind === 'person' ? [sh.to] : sh.members || []).filter((p) => !others.has(p));
  for (const pid of leaving) await removeMember(srv, n.container, pid);
  if (sh.kind === 'forum' && sh.mid) await forumsW.remove(logic, srv, sh.to, sh.mid).catch(() => {});
  await saveShares(logic, srv, n, n.sharesRaw.filter((x) => x.id !== shareId));
}

// A note shared with me: take it out of my files (I leave it; the owner has to share it again).
export async function leave(logic, srv, note) {
  const n = logic.state.notes.find((x) => x.id === note.id);
  if (!n || !n.container) return;
  await removeMember(srv, n.container, srv.me.person);
  srv.sharedConvs.delete(n.container);
  srv.applyingNotes = true;
  logic.setState((st) => ({ notes: st.notes.filter((x) => x.id !== n.id), noteEdit: st.noteEdit && st.noteEdit.id === n.id ? null : st.noteEdit }));
  srv.applyingNotes = false;
  srv.noteSnap = snapshot(logic.state.notes, logic.state.noteBooks);
}

// The id of my Quick notes notebook (what the Home sticky note saves into).
export const quickBook = (logic) => { const b = logic.state.noteBooks || []; return (b.find((x) => x.name === QUICK) || b[0] || {}).id || ''; };
export { newId };
