// Community wall: posts, comments (with replies), likes/loves, shares, polls and moments. Everything is a message in the
// organization's wall conversation, sealed for its audience; reactions, comments and votes point at what they answer
// (reply_to), so counts come from real people. Photos and videos are encrypted attachments of the wall.
import * as files from '../files.js';

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

async function allMessages(srv, conv) {
  const out = [];
  let before = '';
  for (let i = 0; i < 40; i++) {
    const { messages } = await srv.request('GET', `/conversations/${conv}/messages?limit=100${before ? '&before=' + before : ''}`);
    out.push(...messages);
    if (messages.length < 100) break;
    before = messages[messages.length - 1].id;
  }
  return out;
}

export async function load(logic, srv) {
  const { conversations } = await srv.request('GET', '/conversations?kind=wall&limit=200');
  const wall = conversations.find((c) => c.audience === 'org') || conversations[0];
  if (!wall) { logic.setState({ posts: [], moments: [] }); return; }
  srv.wallId = wall.id;
  const msgs = (await allMessages(srv, wall.id)).filter((m) => !m.deleted);
  const opened = await Promise.all(msgs.map(async (m) => ({ m, pl: await srv.vault.openMessage(m) })));
  build(logic, srv, opened.filter((x) => x.pl).sort((a, b) => (a.m.id < b.m.id ? -1 : 1)));
}

function build(logic, srv, items) {
  const me = srv.me.account;
  const who = (m) => (m.sender_account === me ? 'Sam' : m.sender_name);
  const person = (m) => (srv.dirByUid && m.sender_person && srv.dirByUid.get(m.sender_person)) || null;
  const look = (m) => { const p = person(m); const n = m.sender_name || ''; return { initials: m.sender_account === me ? 'S' : (p ? p.initials : n.split(' ').map((x) => x[0]).join('').slice(0, 2)), color: m.sender_account === me ? '#F49AC1' : (p ? p.color : '#C9B6FF') }; };
  const byParent = new Map();
  items.forEach((x) => { if (x.m.reply_to) { const l = byParent.get(x.m.reply_to) || []; l.push(x); byParent.set(x.m.reply_to, l); } });
  const kids = (id, type) => (byParent.get(id) || []).filter((x) => x.m.type === type);
  srv.wallMine = {};   // my reaction / vote message per target, so they can be undone
  const comments = (id) => kids(id, 'comment').map(({ m, pl }) => ({ id: m.id, who: who(m), ...look(m), text: pl.text || '', time: ago(m.created_at), replies: comments(m.id), mine: m.sender_account === me }));
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
    const media = (pl.media || []).map((f) => { const it = { id: f.fileId, name: f.name, meta: f.meta || '', isImg: !!f.isImg, isVideo: !!f.isVideo, url: '', ref: f }; mediaToLoad.push(f); return it; });
    const p = person(m);
    const isComm = m.type !== 'post';
    return {
      id: m.id, author: who(m), ...look(m), mine: m.sender_account === me,
      meta: (isComm ? (m.type === 'memo' ? 'Memo' : 'Announcement') : (pl.department || (p ? p.dept : '') || 'Everyone')) + ' · ' + ago(m.created_at),
      text: isComm ? (pl.title || '') + (pl.body ? '\n' + pl.body : '') : (pl.text || ''), postBg: pl.postBg || undefined,
      likes: likes.length - (mineLike ? 1 : 0), extra: 0, shares: kids(m.id, 'share').length,
      photoLabel: pl.photo ? pl.photo.label : undefined, photoBg: pl.photo ? pl.photo.bg : undefined,
      comments: comments(m.id), media,
      pollData: pl.poll ? { opts: pl.poll.map((o) => ({ id: o.id, label: o.label, votes: votes.filter((v) => v.pl.option === o.id && v.m.sender_account !== me).length })) } : undefined,
      tags: pl.tags || [], poster: pl.poster || undefined
    };
  });
  const moments = items.filter((x) => x.m.type === 'moment').reverse().map(({ m, pl }) => {
    const names = (kind) => reacts(m.id, kind).map((x) => { if (x.m.sender_account === me) srv.wallMine[kind + ':' + m.id] = x.m.id; return who(x.m); });
    const media = pl.media ? (pl.media.fileId ? { ...pl.media, url: '' } : pl.media) : undefined;
    if (media && media.fileId) mediaToLoad.push(media);
    return { id: m.id, owner: who(m), ...look(m), bg: pl.bg || '#2B4A63', caption: pl.caption || '', media, likes: names('like'), loves: names('love') };
  });
  srv.applyingWall = true;
  logic.setState({ posts, moments, liked, pollVotes });
  srv.applyingWall = false;
  // decrypt attachments in the background and drop their URLs in when ready
  mediaToLoad.forEach((f) => files.openUrl(srv, f).then((url) => {
    logic.setState((st) => ({
      posts: st.posts.map((p) => (p.media && p.media.some((x) => x.id === f.fileId) ? { ...p, media: p.media.map((x) => (x.id === f.fileId ? { ...x, url } : x)) } : p)),
      moments: st.moments.map((x) => (x.media && x.media.fileId === f.fileId ? { ...x, media: { ...x.media, url } } : x))
    }));
  }).catch(() => {}));
}

async function send(srv, type, payload, replyTo) {
  const sealed = await srv.vault.seal(srv.wallId, payload);
  return srv.request('POST', `/conversations/${srv.wallId}/messages`, { ...sealed, type, ...(replyTo ? { reply_to: replyTo } : {}) });
}
const del = (srv, id) => srv.request('DELETE', `/conversations/${srv.wallId}/messages/${id}`);

// ---- actions ----

// post = { text, postBg, files: [File], poll: [labels] | null, audience }
export async function post(logic, srv, p) {
  const media = [];
  for (const f of p.files || []) media.push(await files.upload(srv, srv.wallId, f));
  await send(srv, 'post', { text: p.text, department: p.audience && p.audience !== 'Everyone' ? p.audience : (logic.state.meInfo || {}).dept || '', postBg: p.postBg || undefined, media: media.map(({ fileId, name, meta, type, isImg, isVideo, key }) => ({ fileId, name, meta, type, isImg, isVideo, key })), poll: p.poll ? p.poll.map((label, i) => ({ id: 'o' + i, label })) : undefined, tags: p.tags || undefined, poster: p.poster || undefined });
  await load(logic, srv);
}

export async function comment(logic, srv, targetId, text) {
  await send(srv, 'comment', { text }, targetId);
  await load(logic, srv);
}

export async function toggleReaction(logic, srv, targetId, kind) {
  const k = kind + ':' + targetId;
  if (srv.wallMine[k]) await del(srv, srv.wallMine[k]);
  else await send(srv, 'reaction', { kind }, targetId);
  await load(logic, srv);
}

export async function vote(logic, srv, postId, option) {
  if (srv.wallMine['vote:' + postId]) await del(srv, srv.wallMine['vote:' + postId]);
  await send(srv, 'vote', { option }, postId);
  await load(logic, srv);
}

export async function remove(logic, srv, id) {
  await del(srv, id);
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
  await send(srv, 'moment', { caption: m.caption, bg: m.bg, media });
  await load(logic, srv);
}
