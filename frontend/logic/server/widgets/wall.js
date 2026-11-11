// Community wall: posts, comments (with replies), likes/loves, shares, polls and moments. Everything is a message in the
// organization's wall conversation, sealed for its audience; reactions, comments and votes point at what they answer
// (reply_to), so counts come from real people. Photos and videos are encrypted attachments of the wall.
import * as mentions from '../mentions.js';
import { connectRealtime } from '../realtime.js';
import * as files from '../files.js';
import { newConversationKey, wrap, b64, unb64 } from '../e2ee.js';

export const name = 'wall';

const ago = (iso) => {
  const m = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 60000));
  if (m < 1) return 'Just now';
  if (m < 60) return m + ' min ago';
  const h = Math.round(m / 60);
  if (h < 24) return h + ' h ago';
  const d = Math.round(h / 24);
  return d === 1 ? 'Yesterday' : d + ' days ago';
};

// The wall shows as soon as its newest page is in; older pages follow in the background (each page costs a round trip to the server).
// What an earlier load already holds is kept, so a like or a comment only fetches the newest page again.
async function page(srv, conv, before) {
  const { messages } = await srv.request('GET', `/conversations/${conv}/messages?limit=100${before ? '&before=' + before : ''}`);
  return messages;
}

async function render(logic, srv, all) {
  const open = srv.wallOpened || (srv.wallOpened = new Map());
  const msgs = [...all.values()].filter((m) => !m.deleted);
  const opened = await Promise.all(msgs.map(async (m) => {
    if (!open.has(m.id)) open.set(m.id, await srv.vault.openMessage(m));
    return { m, pl: open.get(m.id) };
  }));
  return build(logic, srv, opened.filter((x) => x.pl).sort((a, b) => (a.m.id < b.m.id ? -1 : 1)));
}

// The wall follows the live connection: a vote, like, comment or new post made by anyone shows up within a moment, so poll
// percentages and counts move while the page is open.
function watchLive(logic, srv) {
  const rt = srv.rt || (srv.rt = connectRealtime(srv));
  rt.sub(srv.wallId);
  if (srv.wallWatching) return;
  srv.wallWatching = true;
  let timer = 0;
  const refresh = () => { clearTimeout(timer); timer = setTimeout(() => load(logic, srv).catch(() => {}), 400); };
  const mine = (ev) => String(ev.channel || '').split(':').pop() === srv.wallId;
  rt.on('message.new', (ev) => { if (mine(ev) && (ev.data || {}).sender_account !== srv.me.account) refresh(); });
  rt.on('message.deleted', (ev) => { if (mine(ev)) refresh(); });
  rt.on('comm.published', (ev) => { if (mine(ev)) refresh(); });
  rt.on('open', () => refresh());
}

// What changed on the wall since the last look: { fresh: [new messages], gone: Set of removed ids }, or null when reading the newest
// page again is the better choice (nothing held yet, too much changed, or the server cannot answer catch-up reads).
// A like, a comment or a new post by a colleague now costs one small request, not a hundred messages.
async function changes(srv, conv) {
  let since = srv.wallSince && srv.wallSince.conv === conv ? srv.wallSince.at : '', after = '';
  if (!since) return null;
  const fresh = [], gone = new Set();
  for (let i = 0; i < 5; i++) {
    const res = await srv.request('GET', `/conversations/${conv}/changes?since=${encodeURIComponent(since)}${after ? '&after=' + after : ''}&limit=200`).catch((e) => {
      if (e && (e.status === 404 || e.status === 405)) return null;
      throw e;
    });
    if (!res) return null;
    for (const m of res.messages) if (m.deleted) gone.add(m.id); else fresh.push(m);
    since = res.next_since; after = res.next_after || '';
    if (!res.more) { srv.wallSince = { conv, at: since }; return { fresh, gone }; }
  }
  return null;
}
// the marker after a full read: a little before the newest message, so nothing stored around it is skipped
const markerAfter = (msgs) => (msgs.length ? new Date(new Date(msgs[0].created_at).getTime() - 120000).toISOString() : '2000-01-01T00:00:00Z');

export async function load(logic, srv) {
  // the wall's id never changes, so after the first look it is not asked for again
  let wall = srv.wallId && srv.wallAllId === srv.wallId ? { id: srv.wallId } : null;
  if (!wall) {
    const { conversations } = await srv.request('GET', '/conversations?kind=wall&limit=200');
    wall = conversations.find((c) => c.audience === 'org') || conversations[0];
  }
  if (!wall) { logic.setState({ posts: [], moments: [] }); return; }
  srv.wallId = wall.id;
  watchLive(logic, srv);
  const all = srv.wallAll && srv.wallAllId === wall.id ? srv.wallAll : new Map();   // messages held from earlier loads
  srv.wallAll = all; srv.wallAllId = wall.id;
  const delta = all.size ? await changes(srv, wall.id).catch(() => null) : null;
  if (delta) {
    delta.gone.forEach((id) => all.delete(id));
    delta.fresh.forEach((m) => all.set(m.id, m));
    await render(logic, srv, all);   // from what is held (no request): also drops moments that passed their 24 hours
    return;
  }
  const first = await page(srv, wall.id, '');
  srv.wallSince = { conv: wall.id, at: markerAfter(first) };
  first.forEach((m) => all.set(m.id, m));
  await render(logic, srv, all);
  if (first.length < 100 || srv.wallOlderDone === wall.id) return;
  srv.wallOlderDone = wall.id;
  // older posts (and the likes and comments on them), a page at a time, without holding up the screen
  (async () => {
    let before = first[first.length - 1].id;
    for (let i = 0; i < 2; i++) {
      const older = await page(srv, wall.id, before);
      older.forEach((m) => { if (!all.has(m.id)) all.set(m.id, m); });
      await render(logic, srv, all);
      if (older.length < 100) break;
      before = older[older.length - 1].id;
    }
  })().catch(() => { srv.wallOlderDone = null; });
}

function build(logic, srv, items) {
  const me = srv.me.account;
  const who = (m) => (m.sender_account === me ? srv.me.name : m.sender_name);
  const person = (m) => (srv.dirByUid && m.sender_person && srv.dirByUid.get(m.sender_person)) || null;
  const look = (m) => { const p = person(m); const n = m.sender_name || ''; return { uid: m.sender_person || '', self: m.sender_account === me, initials: m.sender_account === me ? (srv.me.name || 'You').split(' ').map((x) => x[0]).join('').slice(0, 2).toUpperCase() : (p ? p.initials : n.split(' ').map((x) => x[0]).join('').slice(0, 2)), color: m.sender_account === me ? '#F49AC1' : (p ? p.color : '#C9B6FF') }; };
  const byParent = new Map();
  items.forEach((x) => { if (x.m.reply_to) { const l = byParent.get(x.m.reply_to) || []; l.push(x); byParent.set(x.m.reply_to, l); } });
  const kids = (id, type) => (byParent.get(id) || []).filter((x) => x.m.type === type);
  srv.wallMine = {};   // my reaction / vote message per target, so they can be undone
  // a comment or reply: who wrote it, its likes and dislikes (one of each per person, counted from signed reactions) and its replies
  const comments = (id) => kids(id, 'comment').map(({ m, pl }) => {
    const likes = reacts(m.id, 'like'), dislikes = reacts(m.id, 'dislike');
    const mineL = likes.find((x) => x.m.sender_account === me), mineD = dislikes.find((x) => x.m.sender_account === me);
    if (mineL) srv.wallMine['like:' + m.id] = mineL.m.id;
    if (mineD) srv.wallMine['dislike:' + m.id] = mineD.m.id;
    return { id: m.id, who: who(m), ...look(m), text: pl.text || '', time: ago(m.created_at), replies: comments(m.id), mine: m.sender_account === me,
      likes: likes.length, dislikes: dislikes.length, iLiked: !!mineL, iDisliked: !!mineD };
  });
  // Counts are never stored: each like or vote is a signed message (ones whose signature fails never get here), and each
  // person counts once, with their latest one, however many the server holds.
  const onePer = (list) => [...new Map(list.map((x) => [x.m.sender_account, x])).values()];
  const reacts = (id, kind) => onePer(kids(id, 'reaction').filter((x) => (x.pl.kind || 'like') === kind));
  const liked = {}, pollVotes = {};
  const mediaToLoad = [];
  const posts = items.filter((x) => ['post', 'announcement', 'memo'].includes(x.m.type)).reverse().map(({ m, pl }) => {
    const likes = reacts(m.id, 'like');
    const mineLike = likes.find((x) => x.m.sender_account === me);
    if (mineLike) { liked[m.id] = true; srv.wallMine['like:' + m.id] = mineLike.m.id; }
    const votes = onePer(kids(m.id, 'vote')).filter((v) => (pl.poll || []).some((o) => o.id === v.pl.option));
    const myVote = votes.find((x) => x.m.sender_account === me);
    if (myVote) { pollVotes[m.id] = myVote.pl.option; srv.wallMine['vote:' + m.id] = myVote.m.id; }
    // a newsletter or job post sent from Communications is stored as a "post" message, but carries a title and body, not text
    const fromComms = m.type !== 'post' || (pl.title !== undefined && pl.text === undefined);
    const media = (pl.media || pl.files || []).map((f) => { const it = { id: f.fileId, name: f.name, meta: f.meta || '', isImg: !!f.isImg, isVideo: !!f.isVideo, url: '', ref: f }; mediaToLoad.push(f); return it; });
    const p = person(m);
    const isComm = fromComms;
    const commLabel = pl.type || (m.type === 'memo' ? 'Memo' : 'Announcement');
    return {
      id: m.id, author: who(m), ...look(m), mine: m.sender_account === me,
      headline: m.sender_account === me ? ((logic.state.profile || {}).headline || '') : (p ? (p.headline || '') : ''),   // what they wrote about themselves, shown under their name
      meta: (isComm ? commLabel : (pl.department || (p ? p.dept : '') || 'Everyone')) + ' · ' + ago(m.created_at),
      text: isComm ? (pl.title || '') + (pl.summary ? '\n' + pl.summary : '') + (pl.body ? '\n' + pl.body : '') : (pl.text || ''), postBg: pl.postBg || undefined,
      likes: likes.length - (mineLike ? 1 : 0), likers: likes.filter((x) => x.m.sender_account !== me).map((x) => ({ uid: x.m.sender_person || '', name: who(x.m) })), extra: 0, shares: kids(m.id, 'share').length,
      photoLabel: pl.photo ? pl.photo.label : undefined, photoBg: pl.photo ? pl.photo.bg : undefined,
      comments: comments(m.id), media,
      pollData: pl.poll ? { opts: pl.poll.map((o) => ({ id: o.id, label: o.label, votes: votes.filter((v) => v.pl.option === o.id && v.m.sender_account !== me).length })) } : undefined,
      tags: pl.tags || [], poster: pl.poster || undefined
    };
  });
  const moments = items.filter((x) => x.m.type === 'moment' && Date.now() - new Date(x.m.created_at).getTime() < 24 * 3600 * 1000).reverse().map(({ m, pl }) => {
    const names = (kind) => reacts(m.id, kind).map((x) => { if (x.m.sender_account === me) srv.wallMine[kind + ':' + m.id] = x.m.id; return who(x.m); });
    const media = pl.media ? (pl.media.fileId ? { ...pl.media, url: '' } : pl.media) : undefined;
    if (media && media.fileId) mediaToLoad.push(media);
    return { id: m.id, owner: who(m), ...look(m), bg: pl.bg || '#2B4A63', caption: pl.caption || '', media, likes: names('like'), loves: names('love') };
  });
  srv.applyingWall = true;
  logic.setState({ posts, moments, liked, pollVotes });
  srv.applyingWall = false;
  // decrypt attachments and drop their URLs in as they are ready
  const opening = mediaToLoad.map((f) => files.openUrl(srv, f).then((url) => {
    logic.setState((st) => ({
      posts: st.posts.map((p) => (p.media && p.media.some((x) => x.id === f.fileId) ? { ...p, media: p.media.map((x) => (x.id === f.fileId ? { ...x, url } : x)) } : p)),
      moments: st.moments.map((x) => (x.media && x.media.fileId === f.fileId ? { ...x, media: { ...x.media, url } } : x))
    }));
  }).catch(() => {}));
  // The wall counts as loaded when the pictures and videos of the newest posts are in too (the rest, or a slow one after a few
  // seconds, carry on behind), so the page opens complete instead of filling in while you look at it.
  return Promise.race([Promise.all(opening.slice(0, 12)), new Promise((r) => setTimeout(r, 6000))]);
}

// A new organization has no wall until someone first posts: the first post makes it, sealing its key for every person in the
// organization (later members are added to it by the server). Someone without the right to make walls is told to ask.
export async function ensureWall(srv) {
  if (srv.wallId) return srv.wallId;
  const { conversations } = await srv.request('GET', '/conversations?kind=wall&limit=200');
  let wall = conversations.find((c) => c.audience === 'org') || conversations[0];
  if (!wall) {
    try {
      const { people } = await srv.request('GET', '/conversations/audience-members?audience=org');
      const { devices } = await srv.request('POST', '/e2ee/people/devices', { people: [...new Set([srv.me.person, ...people])] });
      const key = newConversationKey();
      const keys = await Promise.all(devices.map(async (d) => ({ epoch: 1, device: d.id, wrapped_key: b64(await wrap(key, unb64(d.identity_key))) })));
      wall = await srv.request('POST', '/conversations', { kind: 'wall', name: 'Community wall', audience: 'org', color: '#F49AC1', keys });
    } catch (e) {
      if (e && (e.status === 403 || e.code === 'forbidden')) throw new Error('The community wall has not been set up yet. Ask an administrator to post first.');
      throw e;
    }
  }
  srv.wallId = wall.id;
  return wall.id;
}

async function send(srv, type, payload, replyTo, fileIds) {
  await ensureWall(srv);
  const sealed = await srv.vault.seal(srv.wallId, payload);
  return srv.request('POST', `/conversations/${srv.wallId}/messages`, { ...sealed, type, ...(replyTo ? { reply_to: replyTo } : {}), ...(fileIds && fileIds.length ? { files: fileIds } : {}) });
}
const del = (srv, id) => srv.request('DELETE', `/conversations/${srv.wallId}/messages/${id}`);

// ---- actions ----

// post = { text, postBg, files: [File], poll: [labels] | null, audience }
export async function post(logic, srv, p) {
  await ensureWall(srv);
  const media = [];
  media.push(...await Promise.all((p.files || []).map((f) => files.upload(srv, srv.wallId, f))));   // all at once
  await send(srv, 'post', { text: p.text, department: p.audience && p.audience !== 'Everyone' ? p.audience : (logic.state.meInfo || {}).dept || '', postBg: p.postBg || undefined, media: media.map(({ fileId, name, meta, type, isImg, isVideo, key }) => ({ fileId, name, meta, type, isImg, isVideo, key })), poll: p.poll ? p.poll.map((label, i) => ({ id: 'o' + i, label })) : undefined, tags: p.tags || undefined, poster: p.poster || undefined }, undefined, media.map((f) => f.fileId));
  if (p.text && p.text.indexOf('@') !== -1) mentions.tell(srv, 'wall', '', mentions.names(srv, p.text).filter((d) => !(p.tags || []).includes(d.name)));
  if ((p.tags || []).length) mentions.tell(srv, 'wall', '', (p.tags || []).map((n) => [...srv.dirByUid.values()].find((d) => d.name === n)).filter(Boolean));
  await load(logic, srv);
}

export async function comment(logic, srv, targetId, text) {
  if (text && text.indexOf('@') !== -1) mentions.tell(srv, 'wall', '', mentions.names(srv, text));
  await send(srv, 'comment', { text }, targetId);
  await load(logic, srv);
}

// Like or dislike a comment or reply: one or the other, and tapping it again takes it back.
export async function reactToComment(logic, srv, commentId, kind) {
  const other = kind === 'like' ? 'dislike' : 'like';
  const old = srv.wallMine[other + ':' + commentId];
  if (old) { await del(srv, old); delete srv.wallMine[other + ':' + commentId]; }
  const mine = srv.wallMine[kind + ':' + commentId];
  if (mine) { await del(srv, mine); delete srv.wallMine[kind + ':' + commentId]; } else await send(srv, 'reaction', { kind }, commentId);
  await load(logic, srv);
}

export async function toggleReaction(logic, srv, targetId, kind) {
  const k = kind + ':' + targetId;
  if (srv.wallMine[k]) await del(srv, srv.wallMine[k]);
  else await send(srv, 'reaction', { kind }, targetId);
  await load(logic, srv);
}

// A vote is final: the first one counts and cannot be taken back or changed.
export async function vote(logic, srv, postId, option) {
  if (srv.wallMine['vote:' + postId]) throw new Error('You have already voted on this poll. Your first vote is final.');
  try { await send(srv, 'vote', { option }, postId); } catch (e) { await load(logic, srv).catch(() => {}); throw e; }
  await load(logic, srv);
}

export async function remove(logic, srv, id) {
  const gone = (st) => ({ posts: st.posts.filter((x) => x.id !== id), moments: st.moments.filter((x) => x.id !== id) });
  logic.setState(gone);
  if (String(id).startsWith('pending-')) { await load(logic, srv); return; }   // still being saved: the next load shows whatever the server has
  try { await del(srv, id); } catch (e) { await load(logic, srv).catch(() => {}); throw e; }
  await load(logic, srv);
}

export async function shared(logic, srv, postId, where) {
  await send(srv, 'share', { note: where || '' }, postId);
  await load(logic, srv);
}

// moment = { caption, bg, file }
export async function addMoment(logic, srv, m) {
  let media;
  if (m.file) { const f = await files.upload(srv, srv.wallId, m.file); media = { fileId: f.fileId, name: f.name, type: f.type, isImg: f.isImg, isVideo: f.isVideo, key: f.key }; }
  await send(srv, 'moment', { caption: m.caption, bg: m.bg, media }, undefined, media ? [media.fileId] : []);
  await load(logic, srv);
}
