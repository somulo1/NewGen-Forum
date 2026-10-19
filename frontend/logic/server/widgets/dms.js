// Private messages: one encrypted conversation per pair of people. Lists your chats with unread counts, opens and sends
// (text, replies, files), deletes your own messages, marks chats read, and receives new messages live over the realtime
// socket. A message for the chat you are looking at plays a soft pop; one for a closed chat plays the brighter alert.
import { newConversationKey, wrap, b64, unb64 } from '../e2ee.js';
import { connectRealtime } from '../realtime.js';
import * as files from '../files.js';
import { messageTone, alertTone } from '../../tones.js';

export const name = 'dms';

const hm = (iso) => {
  const d = new Date(iso);
  const day = (x) => x.toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' });
  const t = d.toLocaleTimeString('en-GB', { hour: 'numeric', minute: '2-digit', hour12: true, timeZone: 'Africa/Nairobi' }).toUpperCase();
  if (day(d) === day(new Date())) return t;
  if (day(d) === day(new Date(Date.now() - 86400000))) return 'Yesterday';
  return d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', timeZone: 'Africa/Nairobi' });
};

async function toMsg(srv, m) {
  const pl = await srv.vault.openMessage(m);
  if (!pl) return null;
  const mine = m.sender_account === srv.me.account;
  const media = (pl.files || []).map((f) => ({ name: f.name, meta: f.meta || '', isImg: !!f.isImg, isVideo: !!f.isVideo, url: '', ref: f }));
  return { id: m.id, mid: m.id, from: mine ? 'me' : m.sender_name, text: pl.text || '', file: media.length ? media : undefined, replyTo: pl.reply || null, time: hm(m.created_at), at: m.created_at };
}

// Fills in decrypted attachment URLs once they are ready.
function loadFiles(logic, srv, dmId, msgs) {
  msgs.forEach((m) => (m.file || []).forEach((f) => files.openUrl(srv, f.ref).then((url) => {
    logic.setState((st) => ({ dms: st.dms.map((d) => (d.id !== dmId ? d : { ...d, msgs: d.msgs.map((x) => (x.id !== m.id ? x : { ...x, file: x.file.map((y) => (y.ref && y.ref.fileId === f.ref.fileId ? { ...y, url } : y)) })) })) }));
  }).catch(() => {})));
}

async function history(srv, conv) {
  const { messages } = await srv.request('GET', `/conversations/${conv}/messages?limit=100`);
  const out = (await Promise.all(messages.filter((m) => !m.deleted).map((m) => toMsg(srv, m)))).filter(Boolean);
  return out.sort((a, b) => (a.id < b.id ? -1 : 1));
}

export async function load(logic, srv) {
  const { conversations } = await srv.request('GET', '/conversations?kind=dm&limit=200');
  const dms = await Promise.all(conversations.map(async (c) => {
    const other = (c.members || []).find((m) => m.account !== srv.me.account) || {};
    const msgs = await history(srv, c.id).catch(() => []);
    return { id: c.id, with: other.name || 'Unknown', person: other.person, unread: c.unread || 0, msgs };
  }));
  logic.setState({ dms });
  dms.forEach((d) => loadFiles(logic, srv, d.id, d.msgs));
  live(logic, srv);
}

const isOpen = (st, id) => st.panelKind === 'dm' && st.dmId === id && st.mode !== 'closed' && st.mode !== 'min';

function live(logic, srv) {
  const rt = srv.rt || (srv.rt = connectRealtime(srv));
  logic.state.dms.forEach((d) => rt.sub(d.id));
  if (srv.dmLive) return;
  srv.dmLive = true;
  // someone opened a new chat with me
  rt.on('conversation.new', async (ev) => {
    const c = ev.data || {};
    if (c.kind !== 'dm' || logic.state.dms.some((d) => d.id === c.conversation)) return;
    await load(logic, srv);
  });
  rt.on('message.new', async (ev) => {
    const m = ev.data || {};
    const dm = logic.state.dms.find((d) => d.id === m.conversation);
    if (!dm || dm.msgs.some((x) => x.id === m.id)) return;
    const msg = await toMsg(srv, m);
    if (!msg) return;
    const mine = msg.from === 'me';
    const pr = (logic.state.dmPrivacy || {})[dm.with] || {};
    if (!mine && (pr.blocked || pr.noChat)) return;
    const open = isOpen(logic.state, dm.id);
    logic.setState((st) => ({ dms: st.dms.map((d) => (d.id !== dm.id ? d : { ...d, unread: mine || open ? d.unread : d.unread + 1, msgs: d.msgs.filter((x) => !x.pending || x.text !== msg.text).concat([msg]) })) }));
    loadFiles(logic, srv, dm.id, [msg]);
    if (!mine) {
      if (open) { messageTone(); markRead(logic, srv, dm.id); } else alertTone();
    }
  });
  rt.on('message.deleted', (ev) => {
    const d = ev.data || {};
    logic.setState((st) => ({ dms: st.dms.map((x) => (x.id !== d.conversation ? x : { ...x, msgs: x.msgs.filter((y) => y.id !== d.message && y.id !== d.id) })) }));
  });
}

// Opening a chat with someone you have never written to creates the pair conversation (with keys for both people).
export async function ensure(logic, srv, personName) {
  const have = logic.state.dms.find((d) => d.with === personName && !String(d.id).startsWith('dm-'));
  if (have) return have.id;
  const pid = srv.personOf.get(personName);
  if (!pid) throw new Error(personName + ' does not have an account yet.');
  const { devices } = await srv.request('POST', '/e2ee/people/devices', { people: [srv.me.person, pid] });
  const key = newConversationKey();
  const keys = await Promise.all(devices.map(async (d) => ({ epoch: 1, device: d.id, wrapped_key: b64(await wrap(key, unb64(d.identity_key))) })));
  const c = await srv.request('POST', '/conversations', { kind: 'dm', members: [pid], keys });
  const msgs = await history(srv, c.id).catch(() => []);
  logic.setState((st) => {
    const local = st.dms.find((d) => d.with === personName);
    const row = { id: c.id, with: personName, person: pid, unread: 0, msgs };
    return { dms: local ? st.dms.map((d) => (d === local ? row : d)) : [row].concat(st.dms), dmId: st.dmId === (local && local.id) ? c.id : st.dmId };
  });
  (srv.rt || (srv.rt = connectRealtime(srv))).sub(c.id);
  return c.id;
}

// msg = { text, files: [File], replyTo }
export async function send(logic, srv, dmId, msg) {
  const dm = logic.state.dms.find((d) => d.id === dmId);
  const conv = String(dmId).startsWith('dm-') ? await ensure(logic, srv, dm.with) : dmId;
  const up = [];
  for (const f of msg.files || []) up.push(await files.upload(srv, conv, f));
  const sealed = await srv.vault.seal(conv, { text: msg.text || '', reply: msg.replyTo || null, files: up.map(({ fileId, name, meta, type, isImg, isVideo, key }) => ({ fileId, name, meta, type, isImg, isVideo, key })) });
  await srv.request('POST', `/conversations/${conv}/messages`, { ...sealed, type: 'text' });
  return conv;
}

export async function remove(logic, srv, dmId, mid) {
  await srv.request('DELETE', `/conversations/${dmId}/messages/${mid}`);
  logic.setState((st) => ({ dms: st.dms.map((d) => (d.id !== dmId ? d : { ...d, msgs: d.msgs.filter((m) => m.id !== mid) })) }));
}

export async function markRead(logic, srv, dmId) {
  const dm = logic.state.dms.find((d) => d.id === dmId);
  if (!dm || String(dmId).startsWith('dm-') || !dm.msgs.length) return;
  const last = dm.msgs[dm.msgs.length - 1];
  if (!last.mid) return;
  await srv.request('POST', `/conversations/${dmId}/read`, { message: last.mid }).catch(() => {});
}

// Reloads one chat's recent history (after sending, so the stored copy replaces the one shown while sending).
export async function refresh(logic, srv, dmId) {
  const msgs = await history(srv, dmId);
  logic.setState((st) => ({ dms: st.dms.map((d) => (d.id !== dmId ? d : { ...d, msgs: msgs.concat(d.msgs.filter((m) => m.pending)) })) }));
  loadFiles(logic, srv, dmId, msgs);
}
