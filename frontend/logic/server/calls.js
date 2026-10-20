// Voice and video calls: private calls, forum calls and meetings. The API decides who may start or join and hands out a
// one-time pass; the room runs on the calls service, which only relays signalling. Media flows directly between the
// browsers (WebRTC mesh, up to 8 people), encrypted end to end by DTLS-SRTP. Bigger forum calls and meetings run on a media
// server instead (see sfuCall.js): the API answers with mode "sfu" and the pass for it.
//   start/join   -> POST /calls or /calls/{id}/join -> { call, token, url }
//   room socket  -> welcome { you, peers } | peer-joined | peer-left | offer/answer/ice (relayed) | state | chat | mod | ended
import { connectRealtime } from './realtime.js';
import { connectSfu, leaveSfu, applySfu, sendChatSfu, sfuActive, control } from './sfuCall.js';
import { startRing, stopRing, startRingback, stopRingback, connected, hangup, joinedTone, handTone } from '../tones.js';

const ICE = [{ urls: 'stun:stun.l.google.com:19302' }];

// Live media, outside the reactive state (MediaStreams must not be proxied). Components read them with streamFor().
const media = { local: null, screen: null, remote: new Map(), audio: new Map(), tracks: new Map(), version: 0, listeners: new Set() };
export const streamFor = (key) => (key === 'local' ? media.screen || media.local : media.remote.get(key) || null);
// In a big room the video comes as media-server tracks (the server picks the quality from the size they are drawn at).
export const trackFor = (key) => media.tracks.get(key) || null;
export const onMedia = (fn) => { media.listeners.add(fn); return () => media.listeners.delete(fn); };
const changed = () => { media.version++; media.listeners.forEach((fn) => fn(media.version)); };

let room = null;   // { ws, me, pcs: Map(peerId -> RTCPeerConnection), callId, host }

function setCall(logic, patch) {
  logic.setState((st) => (st.call ? { call: { ...st.call, ...patch } } : {}));
}

async function getMedia(kind) {
  try {
    return await navigator.mediaDevices.getUserMedia({ audio: true, video: kind === 'video' });
  } catch (e) {
    // no camera or permission refused for video: fall back to voice only
    if (kind === 'video') { try { return await navigator.mediaDevices.getUserMedia({ audio: true }); } catch (x) { /* nothing */ } }
    return null;
  }
}

function peerConn(logic, peerId) {
  if (room.pcs.has(peerId)) return room.pcs.get(peerId);
  const pc = new RTCPeerConnection({ iceServers: ICE });
  room.pcs.set(peerId, pc);
  const local = media.local;
  if (local) local.getTracks().forEach((t) => pc.addTrack(t, local));
  pc.onicecandidate = (e) => { if (e.candidate) send({ type: 'ice', to: peerId, data: e.candidate }); };
  pc.ontrack = (e) => {
    const s = e.streams[0] || new MediaStream([e.track]);
    media.remote.set(peerId, s);
    if (!media.audio.has(peerId)) {
      const a = new Audio(); a.autoplay = true; a.srcObject = s; a.play().catch(() => {});
      media.audio.set(peerId, a);
    } else media.audio.get(peerId).srcObject = s;
    changed();
  };
  pc.onconnectionstatechange = () => {
    if (pc.connectionState === 'connected') { stopRingback(); setCall(logic, { connected: true }); }
  };
  return pc;
}

function send(o) { if (room && room.ws && room.ws.readyState === 1) room.ws.send(JSON.stringify(o)); }

function peerView(p) {
  return { id: p.id, person: p.person, name: p.name, role: p.role, mic: p.mic !== false, cam: !!p.cam, hand: !!p.hand, sharing: !!p.sharing };
}

// opts.ringback: this is a private call I started, so I hear it ringing until the other person answers.
async function connectRoom(logic, srv, entry, opts = {}) {
  leaveRoom(false);
  if (entry.mode === 'sfu') {
    const ctx = { logic, srv, media, changed, setCall, callContext: () => entry.call.context || 'meeting', setRoom: (r) => { room = { ws: null, pcs: new Map(), me: null, ...r }; } };
    try { await connectSfu(ctx, entry, opts); } catch (e) { leaveRoom(false); throw e; }
    return;
  }
  if (opts.ringback) startRingback();
  const ctx = entry.call.context || 'meeting';
  media.local = await getMedia(entry.call.kind);
  if (!media.local) logic.notify('Hub cannot use your microphone. Check the browser permission; you can still listen.', '#FFB86B');
  changed();
  const ws = new WebSocket(entry.url + '?token=' + encodeURIComponent(entry.token));
  room = { ws, me: null, pcs: new Map(), callId: entry.call.id, host: entry.call.host, ctx };
  setCall(logic, { callId: entry.call.id, peers: [], me: null });
  ws.onmessage = async (e) => {
    let m; try { m = JSON.parse(e.data); } catch (x) { return; }
    const peers = () => (logic.state.call && logic.state.call.peers) || [];
    switch (m.type) {
      case 'welcome':
        room.me = m.you.id;
        setCall(logic, { me: m.you.id, myRole: m.you.role, peers: (m.peers || []).map(peerView) });
        // the newcomer calls everyone already in the room
        for (const p of m.peers || []) {
          const pc = peerConn(logic, p.id);
          const offer = await pc.createOffer();
          await pc.setLocalDescription(offer);
          send({ type: 'offer', to: p.id, data: offer });
        }
        // joining a call that is already going: you hear that you are in
        if ((m.peers || []).length) { stopRingback(); connected(); }
        sendState(logic);
        break;
      case 'peer-joined':
        setCall(logic, { peers: peers().filter((p) => p.id !== m.peer.id).concat([peerView(m.peer)]) });
        stopRingback();
        // a private call: they answered. A forum call or meeting: someone joined, with a notice saying who.
        if (ctx === 'dm') connected();
        else { joinedTone(); logic.notify((m.peer.name || 'Someone') + ' joined the call.', '#B8F36B'); }
        break;
      case 'peer-left':
        dropPeer(m.id);
        setCall(logic, { peers: peers().filter((p) => p.id !== m.id) });
        if (logic.state.call && logic.state.call.context === 'dm' && !peers().filter((p) => p.id !== m.id).length) logic.endCall();
        break;
      case 'offer': {
        const pc = peerConn(logic, m.from);
        await pc.setRemoteDescription(m.data);
        const ans = await pc.createAnswer();
        await pc.setLocalDescription(ans);
        send({ type: 'answer', to: m.from, data: ans });
        break;
      }
      case 'answer': { const pc = room.pcs.get(m.from); if (pc) await pc.setRemoteDescription(m.data); break; }
      case 'ice': { const pc = room.pcs.get(m.from); if (pc && m.data) await pc.addIceCandidate(m.data).catch(() => {}); break; }
      case 'state': {
        const before = peers().find((p) => p.id === m.peer.id);
        setCall(logic, { peers: peers().map((p) => (p.id === m.peer.id ? peerView(m.peer) : p)) });
        if (m.peer.id !== room.me && m.peer.hand && !(before && before.hand)) {
          handTone();
          logic.notify((m.peer.name || (before && before.name) || 'Someone') + ' raised their hand.', '#FFD66B');
        }
        break;
      }
      case 'chat': {
        const who = (peers().find((p) => p.id === m.from) || {}).name || 'Someone';
        let text = ''; try { text = (typeof m.data === 'string' ? JSON.parse(m.data) : m.data).text; } catch (x) { /* ignore */ }
        if (text) logic.setState((st) => ({ callMsgs: st.callMsgs.concat([{ who, text }]) }));
        break;
      }
      case 'mod':
        if (m.action === 'mute') { logic.setState({ mic: false }); applyLocal(logic); logic.notify('The host muted you.', '#FFB86B'); }
        if (m.action === 'cam_off') { logic.setState({ cam: false }); applyLocal(logic); logic.notify('The host turned off your camera.', '#FFB86B'); }
        if (m.action === 'remove') { logic.notify('The host removed you from the call.', '#FF9B8A'); logic.endCall(); }
        break;
      case 'ended':
        logic.notify('The call was ended by the host.', '#7CC4FF');
        logic.endCall();
        break;
      default:
    }
  };
  ws.onclose = () => { if (room && room.ws === ws && logic.state.call) { logic.notify('The call connection closed.', '#FFB86B'); logic.endCall(); } };
}

function dropPeer(id) {
  const pc = room && room.pcs.get(id);
  if (pc) { pc.close(); room.pcs.delete(id); }
  const a = media.audio.get(id); if (a) { a.srcObject = null; media.audio.delete(id); }
  media.remote.delete(id);
  changed();
}

export function leaveRoom(sound = true) {
  leaveSfu();
  media.tracks.clear();
  if (room) {
    room.pcs.forEach((pc) => pc.close());
    try { if (room.ws) room.ws.close(); } catch (e) { /* closed */ }
  }
  room = null;
  [media.local, media.screen].forEach((s) => s && s.getTracks().forEach((t) => t.stop()));
  media.local = null; media.screen = null;
  media.audio.forEach((a) => { a.srcObject = null; });
  media.audio.clear(); media.remote.clear();
  stopRingback(); stopRing();
  if (sound) hangup();
  changed();
}

// Mic, camera, screen share and hand follow the app's state; tell the room.
export async function applyLocal(logic) {
  if (sfuActive()) return applySfu({ logic, media, changed, setCall });
  const s = logic.state;
  if (media.local) {
    media.local.getAudioTracks().forEach((t) => { t.enabled = !!s.mic; });
    media.local.getVideoTracks().forEach((t) => { t.enabled = !!s.cam; });
  }
  if (s.sharing && !media.screen && navigator.mediaDevices.getDisplayMedia) {
    try {
      media.screen = await navigator.mediaDevices.getDisplayMedia({ video: true });
      const track = media.screen.getVideoTracks()[0];
      track.onended = () => logic.setState({ sharing: false });
      room && room.pcs.forEach((pc) => { const snd = pc.getSenders().find((x) => x.track && x.track.kind === 'video'); if (snd) snd.replaceTrack(track); else pc.addTrack(track, media.screen); });
    } catch (e) { logic.setState({ sharing: false }); }
  } else if (!s.sharing && media.screen) {
    media.screen.getTracks().forEach((t) => t.stop());
    media.screen = null;
    const cam = media.local && media.local.getVideoTracks()[0];
    room && room.pcs.forEach((pc) => { const snd = pc.getSenders().find((x) => x.track && x.track.kind === 'video'); if (snd) snd.replaceTrack(cam || null); });
  }
  changed();
  sendState(logic);
}

function sendState(logic) {
  const s = logic.state;
  send({ type: 'state', mic: !!s.mic, cam: !!s.cam, sharing: !!s.sharing, hand: !!s.hand });
}

export function sendChat(text) { if (sfuActive()) { sendChatSfu(text); return; } send({ type: 'chat', data: JSON.stringify({ text }) }); }

// Room controls in a big room: mute | cam_off | remove | mute_all | lock | unlock | end | stage_on | stage_off
export function roomControl(logic, srv, action, identity = '') {
  const c = logic.state.call;
  if (!c || !c.callId) return Promise.resolve();
  return control(srv, c.callId, action, identity).catch((e) => logic.notify(e.message || 'Could not do that.', '#FF9B8A'));
}

export function moderate(logic, peerName, what) {
  if (sfuActive()) {
    const p = ((logic.state.call && logic.state.call.peers) || []).find((x) => x.name === peerName);
    if (p) roomControl(logic, logic.srv, what === 'mic' ? 'mute' : 'cam_off', p.id);
    return;
  }
  const p = ((logic.state.call && logic.state.call.peers) || []).find((x) => x.name === peerName);
  if (p) send({ type: 'mod', action: what === 'mic' ? 'mute' : 'cam_off', target: p.id });
}

// ---- starting, joining, ringing ----

// Which conversation a call belongs to: the private chat, the forum, or the calendar event's container.
async function conversationFor(logic, srv, context, ref) {
  if (context === 'dm') {
    const dmsW = await import('./widgets/dms.js');
    return dmsW.ensure(logic, srv, ref);
  }
  if (context === 'forum') return (srv.forumServerId && srv.forumServerId.get(ref)) || null;
  if (ref && /^[0-9a-f-]{36}$/.test(ref)) return ref;
  return null;
}

export async function start(logic, srv, { kind, context, ref }) {
  const conversation = await conversationFor(logic, srv, context, ref);
  const ctx = context === 'dm' ? 'dm' : context === 'forum' ? 'forum' : 'meeting';
  const entry = await srv.request('POST', '/calls', { kind, context: ctx, ...(conversation ? { conversation } : {}) });
  await connectRoom(logic, srv, entry, { ringback: ctx === 'dm' });
}

export async function join(logic, srv, callId) {
  const entry = await srv.request('POST', `/calls/${callId}/join`, {});
  await connectRoom(logic, srv, entry);
}

export async function decline(srv, callId) {
  await srv.request('POST', `/calls/${callId}/decline`, {}).catch(() => {});
}

// Incoming rings and calls that start in my forums or meetings arrive on my account channel.
export function listen(logic, srv) {
  const rt = srv.rt || (srv.rt = connectRealtime(srv));
  if (srv.callsLive) return;
  srv.callsLive = true;
  srv.doneCalls = srv.doneCalls || new Set();   // calls already answered or ended: a late ring for one is ignored
  const onCall = (ev) => {
    const d = ev.data || {};
    if (srv.doneCalls.has(d.call)) return;
    if (logic.state.call) { logic.notify('Missed a call from ' + (d.from_name || 'a colleague') + ' while you were in a call.', '#FFB86B'); return; }
    const pr = (logic.state.dmPrivacy || {})[d.from_name] || {};
    if (d.context === 'dm' && (pr.blocked || pr.noCall || (d.kind === 'video' && pr.noVideo))) { decline(srv, d.call); return; }
    logic.setState({ incoming: { name: d.from_name, kind: d.kind, callId: d.call, context: d.context, conversation: d.conversation } });
    // a private call rings until answered; a forum call or meeting shows the banner (and meetings a notice) without ringing
    if (d.context === 'dm') startRing();
    else if (d.context !== 'forum') logic.notify((d.from_name || 'A colleague') + ' started a ' + (d.context === 'standup' ? 'standup' : 'meeting') + '. Join from the banner.', '#7CC4FF');
    clearTimeout(srv.ringT);
    srv.ringT = setTimeout(() => { if (logic.state.incoming && logic.state.incoming.callId === d.call) { stopRing(); logic.setState({ incoming: null }); logic.notify('Missed ' + (d.kind === 'video' ? 'video call' : 'call') + ' from ' + d.from_name + '.', '#FFB86B'); loadHistory(logic, srv); } }, 45000);
  };
  rt.on('call.incoming', onCall);
  rt.on('call.started', onCall);
  // Answered or ended somewhere else (another tab or device of mine, or the caller hung up): stop ringing here.
  rt.on('call.answered', (ev) => {
    const d = ev.data || {};
    srv.doneCalls.add(d.call);
    const inc = logic.state.incoming;
    if (inc && inc.callId === d.call) { clearTimeout(srv.ringT); stopRing(); logic.setState({ incoming: null }); }
    if (logic.state.call && logic.state.call.callId === d.call) stopRingback();
  });
  rt.on('call.ended', (ev) => {
    const d = ev.data || {};
    srv.doneCalls.add(d.call);
    const inc = logic.state.incoming;
    if (inc && inc.callId === d.call) { clearTimeout(srv.ringT); stopRing(); logic.setState({ incoming: null }); logic.notify('Missed a call from ' + (inc.name || 'a colleague') + '.', '#FFB86B'); }
  });
  rt.on('call.declined', (ev) => {
    const d = ev.data || {};
    if (logic.state.call && logic.state.call.callId === d.call) { logic.notify((d.by_name || 'They') + ' declined the call.', '#FF9B8A'); logic.endCall(); }
  });
}

// The Calls tab in each chat: real history from the server.
export async function loadHistory(logic, srv) {
  const { calls } = await srv.request('GET', '/calls/history?limit=200');
  const localOf = (c) => {
    if (c.context === 'forum' && srv.forumLocalId) return srv.forumLocalId.get(c.conversation) || c.conversation;
    if (c.context === 'dm') return (c.people || []).find((n) => n !== srv.me.name) || c.host_name;
    return c.conversation || '';
  };
  const dir = (c) => (c.status === 'missed' ? 'missed' : c.status === 'declined' ? 'declined' : c.direction || 'out');
  logic.setState({ callLog: calls.map((c) => ({ id: c.id, ctx: c.context === 'meeting' || c.context === 'standup' ? 'meeting' : c.context, ref: localOf(c), kind: c.kind, dir: dir(c), at: new Date(c.started_at).getTime(), dur: c.duration_seconds || 0 })) });
}
