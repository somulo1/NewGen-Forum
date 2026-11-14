// Voice and video calls: private calls, forum calls and meetings. The API decides who may start or join and hands out a
// one-time pass; the room runs on the calls service, which only relays signalling. Media flows directly between the
// browsers (WebRTC mesh, up to 8 people), encrypted end to end by DTLS-SRTP. Bigger forum calls and meetings run on a media
// server instead (see sfuCall.js): the API answers with mode "sfu" and the pass for it.
//   start/join   -> POST /calls or /calls/{id}/join -> { call, token, url }
//   room socket  -> welcome { you, peers } | peer-joined | peer-left | offer/answer/ice (relayed) | state | chat | mod | ended
import { connectRealtime } from './realtime.js';
import { denoiseStream, denoiseTrack } from '~/logic/denoise.js';
import * as notifyW from '~/logic/browserNotify.js';
import * as dmsW from './widgets/dms.js';
import { connectSfu, leaveSfu, applySfu, sendChatSfu, sfuActive, control, sfuStats, setSfuUpload, sfuVideoPaused, localScreenTrack } from './sfuCall.js';
import { startQualityWatch, stopQualityWatch, noteReconnect, noteJoined, newLadder, stepLadder, resetLadder, upBad, downBad } from './callQuality.js';
import { startRing, stopRing, startRingback, stopRingback, connected, hangup, joinedTone, handTone } from '../tones.js';
import { retryDelay } from './backoff.js';

const FALLBACK_ICE = [{ urls: 'stun:stun.l.google.com:19302' }, { urls: 'stun:stun.cloudflare.com:3478' }];

const media = { local: null, screen: null, remote: new Map(), audio: new Map(), tracks: new Map(), version: 0, listeners: new Set() };
// 'screen' is my own shared screen (the presenter's preview); 'local' is my camera (or my screen while I share)
export const streamFor = (key) => (key === 'screen' ? media.screen : key === 'local' ? media.screen || media.local : media.remote.get(key) || null);
export const trackFor = (key) => (key === 'screen' ? (sfuActive() ? localScreenTrack() : null) : media.tracks.get(key) || null);
export const onMedia = (fn) => { media.listeners.add(fn); return () => media.listeners.delete(fn); };
const changed = () => { media.version++; media.listeners.forEach((fn) => fn(media.version)); };

let room = null;   // { ws, me, pcs: Map(peerId -> RTCPeerConnection), callId, host }

// Device-to-device video: every device uploads a copy per person, so it is kept small, and smaller with more people.
const MESH_VIDEO = { width: { ideal: 640 }, height: { ideal: 360 }, frameRate: { ideal: 24, max: 30 } };
const meshBitrate = (peers) => (peers <= 1 ? 700000 : 350000);
const SCREEN_BITRATE = 1500000;
const REDUCED_BITRATE = 150000;

// Voice goes first when the connection is tight; video takes what is left, up to its limit.
async function tuneSenders(pc) {
  const peers = room ? room.pcs.size : 1;
  const screen = media.screen && media.screen.getVideoTracks()[0];
  const upLevel = net ? net.up.level : 0;
  for (const sender of pc.getSenders()) {
    if (!sender.track) continue;
    // someone who joined while my video is paused must not start receiving it
    if (upLevel === 2 && sender.track.kind === 'video' && sender.track !== screen) { await sender.replaceTrack(null).catch(() => {}); continue; }
    const params = sender.getParameters();
    if (!params.encodings || !params.encodings.length) continue;
    const video = sender.track.kind === 'video';
    params.encodings.forEach((e) => {
      e.priority = video ? 'low' : 'high';
      e.networkPriority = e.priority;
      if (video && sender.track === screen) { e.maxBitrate = SCREEN_BITRATE; e.scaleResolutionDownBy = 1; }
      else if (video) { e.maxBitrate = upLevel === 1 ? REDUCED_BITRATE : meshBitrate(peers); e.scaleResolutionDownBy = upLevel === 1 ? 2 : 1; }
    });
    await sender.setParameters(params).catch(() => {});
  }
}
const tuneAll = () => { if (room) room.pcs.forEach((pc) => tuneSenders(pc)); };

// Starts measuring this call's connection, and sends the summary to monitoring when it ends.
function watchQuality(logic, srv, mode, getReports) {
  if (!room) return;   // hung up while connecting
  room.srv = srv;
  net = { mode, up: newLadder(), down: newLadder(), downOff: false, lastCam: !!logic.state.cam, notice: '' };
  startQualityWatch({ mode, getReports, onWeak: (netWeak) => setCall(logic, { netWeak }), onSample: (s) => netStep(logic, s) });
}

// ---- falling back to voice when a connection cannot carry video ----
// A weak upload stops only my camera, so everyone sees my avatar and I still see them. A weak download stops only the video
// coming to me, and everyone still sees me. A shared screen is left alone. A camera the person turned off stays off.
let net = null;

function netStep(logic, sample) {
  if (!net || !room || !logic.state.call || logic.state.call.mode !== 'video') return;
  if (logic.state.cam || net.up.level > 0) {
    const before = net.up.level;
    if (stepLadder(net.up, upBad(sample)) !== before) applyUpLevel(logic);
  }
  if (net.mode === 'mesh') {
    stepLadder(net.down, downBad(sample));
    applyDownOff(logic);
  } else net.sfuDownPaused = downBad(sample) && sfuVideoPaused();
  showNetNotice(logic);
}

function showNetNotice(logic) {
  if (!net) return;
  const s = logic.state;
  const upOff = net.up.level === 2 && s.cam;
  const notice = upOff ? 'Your video is paused to keep your call stable.'
    : net.downOff ? 'Video from others is paused to keep your call stable.'
    : net.sfuDownPaused ? 'Some video is paused because your connection is weak.' : '';
  const canResume = !!(upOff || net.downOff);
  if (notice === net.notice && canResume === net.canResume) return;
  net.notice = notice; net.canResume = canResume;
  setCall(logic, { netNotice: notice, netCanResume: canResume, netDownPaused: !!net.downOff });
}

function applyUpLevel(logic) {
  const level = net.up.level;
  if (net.mode === 'sfu') { setSfuUpload(level); return; }
  if (room && level < 2) {
    const cam = media.local && media.local.getVideoTracks()[0];
    if (cam && !media.screen) room.pcs.forEach((pc) => {
      const tr = pc.getTransceivers().find((t) => t.receiver && t.receiver.track && t.receiver.track.kind === 'video');
      if (tr && !tr.sender.track) tr.sender.replaceTrack(cam).then(() => tuneSenders(pc)).catch(() => {});
    });
  }
  tuneAll();
  sendState(logic);
}

// Stops or restarts the video coming to me: the other side is told through a fresh offer.
function applyDownOff(logic) {
  const sharing = ((logic.state.call && logic.state.call.peers) || []).some((p) => p.sharing);
  const off = net.down.level === 2 && !sharing;
  if (off === net.downOff || !room) return;
  net.downOff = off;
  const flip = off ? { sendrecv: 'sendonly', recvonly: 'inactive' } : { sendonly: 'sendrecv', inactive: 'recvonly' };
  for (const [peerId, pc] of room.pcs) {
    const tr = pc.getTransceivers().find((t) => t.receiver && t.receiver.track && t.receiver.track.kind === 'video');
    if (!tr || !flip[tr.direction]) continue;
    tr.direction = flip[tr.direction];
    pc.createOffer().then((o) => pc.setLocalDescription(o).then(() => send({ type: 'offer', to: peerId, data: o }))).catch(() => {});
  }
}

// The "Turn back on" button, or the person switching their camera back on.
export function resumeVideo(logic) {
  if (!net) return;
  resetLadder(net.up); resetLadder(net.down);
  applyUpLevel(logic);
  if (net.mode === 'mesh') applyDownOff(logic);
  showNetNotice(logic);
}
function endQualityWatch() {
  const summary = stopQualityWatch();
  if (summary && room && room.srv) room.srv.request('POST', '/calls/quality', summary).catch(() => {});
}

function setCall(logic, patch) {
  logic.setState((st) => (st.call ? { call: { ...st.call, ...patch } } : {}));
}

// Noise cancellation is on for every call, meeting and standup: the browser removes background noise (keyboards, traffic, fans),
// cancels the echo of the other side coming back through the speakers, and levels the voice. Where the browser offers it
// (recent Chrome/Edge, Safari) voice isolation goes further and keeps only the speaker's voice. Unknown options are ignored.
export const VOICE = { echoCancellation: true, noiseSuppression: true, autoGainControl: true, voiceIsolation: true, channelCount: 1 };

async function getMedia(kind) {
  const md = navigator.mediaDevices;
  if (!md || !md.getUserMedia) return null;
  const tries = kind === 'video' ? [{ audio: VOICE, video: MESH_VIDEO }, { audio: VOICE }, { video: MESH_VIDEO }] : [{ audio: VOICE }, { audio: true }];
  for (const c of tries) {
    try { const raw = await md.getUserMedia(c); return await denoiseStream(raw); } catch (e) { /* busy, missing or refused: try the next */ }
  }
  return null;
}

function peerConn(logic, peerId) {
  if (room.pcs.has(peerId)) return room.pcs.get(peerId);
  const pc = new RTCPeerConnection({ iceServers: (room && room.ice) || FALLBACK_ICE });
  pc.queued = [];          // candidates that arrived before the other side's description
  pc.initiator = false;    // true on the side that made the first offer
  room.pcs.set(peerId, pc);
  const local = media.local;
  if (local) local.getTracks().forEach((t) => pc.addTrack(t, local));
  const has = (k) => !!(local && local.getTracks().some((t) => t.kind === k));
  if (!has('audio')) pc.addTransceiver('audio', { direction: 'recvonly' });
  if (room && room.kind === 'video' && !has('video')) pc.addTransceiver('video', { direction: 'recvonly' });
  pc.onicecandidate = (e) => { if (e.candidate) send({ type: 'ice', to: peerId, data: e.candidate }); };
  pc.ontrack = (e) => {
    const s = e.streams[0] || new MediaStream([e.track]);
    media.remote.set(peerId, s);
    if (!media.audio.has(peerId)) {
      const a = new Audio(); a.autoplay = true; a.srcObject = s; a.play().catch(() => {});
      media.audio.set(peerId, a);
    } else media.audio.get(peerId).srcObject = s;
    e.track.onunmute = () => changed();   // some browsers show black until the first frame arrives
    changed();
  };
  pc.onconnectionstatechange = () => {
    if (pc.connectionState === 'connected') { clearTimeout(pc.lostT); pc.restarts = 0; stopRingback(); setCall(logic, { connected: true }); tuneAll(); return; }
    if (pc.connectionState === 'disconnected') {
      // often a blip (wifi to mobile data): give it a few seconds, then try to reconnect
      clearTimeout(pc.lostT);
      pc.lostT = setTimeout(() => { if (room && room.pcs.get(peerId) === pc && pc.connectionState !== 'connected') lost(logic, peerId, pc); }, 6000);
    }
    if (pc.connectionState === 'failed') lost(logic, peerId, pc);
  };
  return pc;
}


// The path to a peer broke. Reconnect (the side that made the first offer asks for fresh candidates); when that does not work,
// say so: in a private call it ends, in a group call that person's tile goes away.
function lost(logic, peerId, pc) {
  if (!room || room.pcs.get(peerId) !== pc) return;
  pc.restarts = (pc.restarts || 0) + 1;
  noteReconnect();
  if (pc.restarts <= 2) {
    if (pc.initiator) {
      pc.restartIce();
      pc.createOffer({ iceRestart: true }).then((o) => pc.setLocalDescription(o).then(() => send({ type: 'offer', to: peerId, data: o }))).catch(() => {});
    }
    clearTimeout(pc.lostT);
    pc.lostT = setTimeout(() => { if (room && room.pcs.get(peerId) === pc && pc.connectionState !== 'connected') lost(logic, peerId, pc); }, 8000);
    return;
  }
  const c = logic.state.call;
  if (c && c.context === 'dm') { logic.notify('The call lost its connection and ended.', '#FFB86B'); logic.endCall(); return; }
  const name = ((c && c.peers) || []).find((p) => p.id === peerId);
  logic.notify((name ? name.name : 'Someone') + ' lost their connection.', '#FFB86B');
  dropPeer(peerId);
}

// Candidates are added once the other side's description is in; any that came earlier wait here.
async function flushCandidates(pc) {
  const q = pc.queued || [];
  pc.queued = [];
  for (const c of q) await pc.addIceCandidate(c).catch(() => {});
}

function send(o) { if (room && room.ws && room.ws.readyState === 1) room.ws.send(JSON.stringify(o)); }

function peerView(p) {
  return { id: p.id, person: p.person, name: p.name, role: p.role, mic: p.mic !== false, cam: !!p.cam, hand: !!p.hand, sharing: !!p.sharing, netPaused: !!p.net_paused };
}

// opts.ringback: this is a private call I started, so I hear it ringing until the other person answers.
async function connectRoom(logic, srv, entry, opts = {}) {
  leaveRoom(false);
  if (entry.mode === 'sfu') {
    const ctx = { logic, srv, media, changed, setCall, callContext: () => entry.call.context || 'meeting', setRoom: (r) => { room = { ws: null, pcs: new Map(), me: null, ...r }; } };
    const began = performance.now();
    try { await connectSfu(ctx, entry, opts); } catch (e) { leaveRoom(false); throw e; }
    watchQuality(logic, srv, 'sfu', sfuStats);
    noteJoined(performance.now() - began);
    return;
  }
  if (!logic.state.call) return;   // hung up while the call was still being set up: nothing to ring or open
  // The other person may have picked up before this side was even ready: then there is nothing to ring.
  const pickedUp = !!(srv.answeredCalls && srv.answeredCalls.has(entry.call.id));
  if (opts.ringback && !pickedUp) startRingback();
  const ctx = entry.call.context || 'meeting';
  media.local = await getMedia(entry.call.kind);
  if (!logic.state.call) { (media.local ? media.local.getTracks() : []).forEach((t) => t.stop()); media.local = null; stopRingback(); return; }   // hung up while the microphone was being asked for
  if (!media.local || !media.local.getAudioTracks().length) {
    logic.setState({ mic: false });
    logic.notify('No microphone is available to Hub (another app or browser tab may be using it, or it is blocked in the browser settings). You can still hear and see the others; they will see you as muted.', '#FFB86B');
  }
  if (entry.call.kind === 'video' && !(media.local && media.local.getVideoTracks().length)) {
    logic.setState({ cam: false });
    logic.notify('Your camera is not available, so the other person cannot see you. It may be in use by another app or browser tab, or blocked in the browser settings.', '#FFB86B');
  }
  changed();
  const ws = new WebSocket(entry.url + '?token=' + encodeURIComponent(entry.token));
  room = { ws, me: null, pcs: new Map(), callId: entry.call.id, host: entry.call.host, ctx, kind: entry.call.kind, ice: entry.ice && entry.ice.length ? entry.ice : FALLBACK_ICE };
  setCall(logic, { callId: entry.call.id, peers: [], me: null });
  watchQuality(logic, srv, 'mesh', () => Promise.all([...room.pcs.values()].map((pc) => pc.getStats().catch(() => null))));
  if (srv.answeredCalls && srv.answeredCalls.has(entry.call.id)) stopRingback();   // picked up while the microphone was being set up
  // messages are handled one after another, in the order they arrive: an offer must be in before the candidates that follow it
  let chain = Promise.resolve();
  const feed = (e) => { chain = chain.then(() => handle(e)).catch((x) => console.warn('call message failed', x)); };
  ws.onmessage = feed;
  const handle = async (e) => {
    let m; try { m = JSON.parse(e.data); } catch (x) { return; }
    const peers = () => (logic.state.call && logic.state.call.peers) || [];
    switch (m.type) {
      case 'welcome':
        room.me = m.you.id;
        room.resume = m.resume || '';   // the secret that lets this device take its place back if its connection drops
        setCall(logic, { me: m.you.id, myRole: m.you.role, peers: (m.peers || []).map(peerView) });
        // the newcomer calls everyone already in the room
        for (const p of m.peers || []) {
          const pc = peerConn(logic, p.id);
          pc.initiator = true;
          const offer = await pc.createOffer();
          await pc.setLocalDescription(offer);
          send({ type: 'offer', to: p.id, data: offer });
        }
        // joining a call that is already going: you hear that you are in
        if ((m.peers || []).length) { stopRingback(); connected(); }
        sendState(logic);
        break;
      case 'resumed': {
        // the connection to the call server came back and this device has its place again: bring the list of people up to date with what
        // happened while it was away (the media between devices carried on by itself)
        const now = (m.peers || []).map(peerView);
        for (const p of peers()) if (!now.some((x) => x.id === p.id)) dropPeer(p.id);
        setCall(logic, { peers: now });
        sendState(logic);
        break;
      }
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
        // nobody else is left (a private or forum call): the call ends for me too, at once
        if (logic.state.call && (logic.state.call.context === 'dm' || logic.state.call.context === 'forum') && !peers().filter((p) => p.id !== m.id).length) { logic.notify(logic.state.call.context === 'dm' ? ((logic.state.call.title || 'The other person').replace('Private call with ', '') + ' hung up.') : 'Everyone else left, so the call ended.', '#7CC4FF'); logic.endCall(); }
        break;
      case 'offer': {
        const pc = peerConn(logic, m.from);
        await pc.setRemoteDescription(m.data);
        await flushCandidates(pc);
        const ans = await pc.createAnswer();
        await pc.setLocalDescription(ans);
        send({ type: 'answer', to: m.from, data: ans });
        break;
      }
      case 'answer': { const pc = room.pcs.get(m.from); if (pc) { await pc.setRemoteDescription(m.data); await flushCandidates(pc); } break; }
      case 'ice': {
        const pc = room.pcs.get(m.from);
        if (!pc || !m.data) break;
        if (pc.remoteDescription) await pc.addIceCandidate(m.data).catch(() => {});
        else pc.queued.push(m.data);
        break;
      }
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

  let resuming = false;
  const reconnect = async () => {
    if (resuming) return;
    resuming = true;
    logic.notify('Connection lost. Reconnecting to the call…', '#FFB86B');
    const t0 = Date.now();
    for (let attempt = 1; room && logic.state.call && Date.now() - t0 < 30000; attempt++) {
      await new Promise((r) => setTimeout(r, attempt === 1 ? 400 : retryDelay(attempt)));
      if (!room || !logic.state.call) break;
      const q = '?resume=' + encodeURIComponent(room.me) + '&room=' + encodeURIComponent(entry.call.id) + '&key=' + encodeURIComponent(room.resume);
      const back = await new Promise((resolve) => {
        let w; let settled = false;
        const fin = (v) => { if (!settled) { settled = true; clearTimeout(giveUp); resolve(v); } };
        const giveUp = setTimeout(() => { try { w && w.close(); } catch (x) { /* already closed */ } fin(null); }, 6000);
        try { w = new WebSocket(entry.url + q); } catch (x) { fin(null); return; }
        w.onmessage = (e) => {
          let m; try { m = JSON.parse(e.data); } catch (x) { return; }
          if (!m || m.type !== 'resumed') return;
          if (!room || !logic.state.call) { try { w.close(); } catch (x) { /* closed */ } fin(null); return; }
          room.ws = w;
          watch(w);
          feed(e);   // the "resumed" message itself, then everything after it, goes through the normal handler
          fin(w);
        };
        w.onclose = () => fin(null);
        w.onerror = () => fin(null);
      });
      if (back) { resuming = false; logic.notify('Back in the call.', '#B8F36B'); return; }
    }
    resuming = false;
    if (room && logic.state.call) { logic.notify('The call connection closed.', '#FFB86B'); logic.endCall(); }
  };
  let opened = false;
  // handlers of the connection the room is using now (the first one, or the one that came back)
  const watch = (w) => {
    w.onmessage = feed;
    w.onclose = () => {
      if (!room || room.ws !== w || !logic.state.call) return;
      if (room.resume && room.me) { reconnect(); return; }
      logic.notify(opened ? 'The call connection closed.' : 'Could not reach the call server, or the call pass expired. Start the call again.', '#FFB86B');
      logic.endCall();
    };
  };
  ws.onopen = () => { opened = true; };
  watch(ws);
}

function dropPeer(id) {
  const pc = room && room.pcs.get(id);
  if (pc) { pc.close(); room.pcs.delete(id); }
  const a = media.audio.get(id); if (a) { a.srcObject = null; media.audio.delete(id); }
  media.remote.delete(id);
  changed();
}

export function leaveRoom(sound = true) {
  endQualityWatch();
  net = null;
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
// Gets one kind of track later, when the person turns the microphone or camera on after joining without it (the browser asks
// again), and adds it to every connection already open, so the others start hearing or seeing them. Those who already had it are
// not touched.
async function ensureTrack(kind) {
  if (media.local && media.local.getTracks().some((t) => t.kind === kind)) return true;
  const md = navigator.mediaDevices;
  if (!md || !md.getUserMedia || !room) return false;
  let got;
  try { got = await md.getUserMedia(kind === 'audio' ? { audio: VOICE } : { video: MESH_VIDEO }); } catch (e) { return false; }
  let track = got.getTracks()[0];
  if (!track) return false;
  if (kind === 'audio') { const d = await denoiseTrack(track); if (d) track = d.track; }   // AI noise cancellation, always on
  if (!media.local) media.local = new MediaStream();
  media.local.addTrack(track);
  for (const [peerId, pc] of room.pcs) {
    const tr = pc.getTransceivers().find((t) => t.receiver && t.receiver.track && t.receiver.track.kind === kind && !t.sender.track);
    try {
      if (tr) { tr.direction = 'sendrecv'; await tr.sender.replaceTrack(track); } else pc.addTrack(track, media.local);
      // a new track needs a fresh offer so the other side knows about it
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      send({ type: 'offer', to: peerId, data: offer });
    } catch (e) { /* that connection will pick it up when it reconnects */ }
  }
  tuneAll();
  changed();
  return true;
}

export async function applyLocal(logic) {
  const s = logic.state;
  if (net) {
    if (s.cam && !net.lastCam && net.up.level > 0) resumeVideo(logic);
    net.lastCam = !!s.cam;
    showNetNotice(logic);
  }
  if (sfuActive()) return applySfu({ logic, media, changed, setCall });
  const HOW = ' Allow it in the browser (the lock icon in the address bar), then tap it again.';
  if (room && s.mic && !(media.local && media.local.getAudioTracks().length)) {
    if (!(await ensureTrack('audio'))) { logic.setState({ mic: false }); logic.notify('Hub cannot use your microphone, so you are muted. You can still listen.' + HOW, '#FFB86B'); }
  }
  if (room && s.cam && room.kind === 'video' && !(media.local && media.local.getVideoTracks().length)) {
    if (!(await ensureTrack('video'))) { logic.setState({ cam: false }); logic.notify('Hub cannot use your camera, so it is off. The others can still see and hear each other.' + HOW, '#FFB86B'); }
  }
  if (media.local) {
    media.local.getAudioTracks().forEach((t) => { t.enabled = !!s.mic; });
    media.local.getVideoTracks().forEach((t) => { t.enabled = !!s.cam; });
  }
  if (s.sharing && !media.screen && navigator.mediaDevices.getDisplayMedia) {
    try {
      media.screen = await navigator.mediaDevices.getDisplayMedia({ video: { frameRate: { ideal: 15, max: 15 } } });
      const track = media.screen.getVideoTracks()[0];
      track.contentHint = 'text';   // keep slides and documents readable: fewer frames rather than blurred text
      changed();   // the presenter's own preview shows the shared screen at once
      track.onended = () => logic.setState({ sharing: false });
      await Promise.all([...(room ? room.pcs.values() : [])].map(async (pc) => { const snd = pc.getSenders().find((x) => x.track && x.track.kind === 'video'); if (snd) await snd.replaceTrack(track); else pc.addTrack(track, media.screen); }));
      tuneAll();
    } catch (e) { logic.setState({ sharing: false }); }
  } else if (!s.sharing && media.screen) {
    media.screen.getTracks().forEach((t) => t.stop());
    media.screen = null;
    changed();
    const cam = media.local && media.local.getVideoTracks()[0];
    // back to the camera, unless my video is paused for the connection
    await Promise.all([...(room ? room.pcs.values() : [])].map(async (pc) => { const snd = pc.getSenders().find((x) => x.track && x.track.kind === 'video'); if (snd) await snd.replaceTrack(net && net.up.level === 2 ? null : cam || null); }));
    tuneAll();
  }
  changed();
  sendState(logic);
}

function sendState(logic) {
  const s = logic.state;
  send({ type: 'state', mic: !!s.mic, cam: !!s.cam, sharing: !!s.sharing, hand: !!s.hand, net_paused: !!(net && net.up.level === 2 && s.cam) });
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
    return dmsW.ensure(logic, srv, ref);
  }
  if (context === 'forum') return (srv.forumServerId && srv.forumServerId.get(ref)) || null;
  if (ref && /^[0-9a-f-]{36}$/.test(ref)) return ref;
  return null;
}

export async function start(logic, srv, { kind, context, ref }) {
  srv.cancelStart = false; srv.startingCall = '';
  const conversation = await conversationFor(logic, srv, context, ref);
  const ctx = context === 'dm' ? 'dm' : context === 'forum' ? 'forum' : 'meeting';
  const entry = await srv.request('POST', '/calls', { kind, context: ctx, ...(conversation ? { conversation } : {}) });
  srv.startingCall = entry.call.id;
  // hung up before the call even existed on the server: tell it now, so nobody is left ringing
  if (srv.cancelStart || !logic.state.call) { srv.cancelStart = false; srv.request('POST', `/calls/${entry.call.id}/leave`, {}).catch(() => {}); return; }
  // A private call nobody picks up must not ring for ever, even if the server's "missed" notice never reaches this browser: after 50 s
  // (the server gives up at 45) it says so and hangs up.
  if (ctx === 'dm') {
    const id = entry.call.id;
    setTimeout(() => {
      const cur = logic.state.call;
      if (cur && cur.callId === id && !cur.connected && !(srv.answeredCalls && srv.answeredCalls.has(id))) { logic.notify('No answer.', '#7CC4FF'); logic.endCall(); }
    }, 50000);
  }
  await connectRoom(logic, srv, entry, { ringback: ctx === 'dm' });
}

// The person hung up. The server is told straight away (not after the room notices), so whoever is still being rung stops
// ringing at once when nobody had answered; if someone has already joined, the call just goes on without them.
export function hangUp(srv, call) {
  const id = (call && call.callId) || srv.startingCall;
  if (id) srv.request('POST', `/calls/${id}/leave`, {}).catch(() => {});
  else srv.cancelStart = true;   // the request that creates the call is still running: start() tells the server when it returns
}

export async function join(logic, srv, callId) {
  srv.cancelStart = false; srv.startingCall = '';
  const entry = await srv.request('POST', `/calls/${callId}/join`, {});
  await connectRoom(logic, srv, entry);
}

export async function decline(srv, callId) {
  await srv.request('POST', `/calls/${callId}/decline`, {}).catch(() => {});
}

// Calls in progress that I can see or join (my private chats, my forums, meetings I am invited to). Fetched first thing after
// sign-in, before the rest of the app loads, so a call ringing me during a reload shows its card at once, and a call already
// running in one of my chats glows where I can join it. Kept fresh by the call events and a light poll.
export async function loadLive(logic, srv, prefetched) {
  let res = prefetched ? await prefetched : null;
  if (!res) { try { res = await srv.request('GET', '/calls/live'); } catch (e) { return; } }
  const list = (res && res.calls) || [];
  logic.setState({ liveCalls: list });
  const mine = logic.state.call && logic.state.call.callId;
  const ring = list.find((c) => c.ringing_me && c.id !== mine && !(srv.doneCalls && srv.doneCalls.has(c.id)) && (c.context === 'dm' || c.context === 'forum'));
  if (ring && !logic.state.incoming && !logic.state.call) {
    logic.setState({ incoming: { name: ring.host_name, kind: ring.kind, callId: ring.id, context: ring.context, conversation: ring.conversation } });
    startRing();
    notifyW.show({ title: (ring.kind === 'video' ? 'Video call' : 'Call') + ' from ' + (ring.host_name || 'a colleague'), body: 'Click to answer in Hub', tag: 'call:' + ring.id, urgent: true });
    clearTimeout(srv.ringT);
    srv.ringT = setTimeout(() => { const inc = logic.state.incoming; if (inc && inc.callId === ring.id) { stopRing(); logic.setState({ incoming: null }); } }, 45000);
  }
}
let liveT = 0;
const refreshLive = (logic, srv) => { clearTimeout(liveT); liveT = setTimeout(() => loadLive(logic, srv), 400); };

// Incoming rings and calls that start in my forums or meetings arrive on my account channel.
export function listen(logic, srv) {
  const rt = srv.rt || (srv.rt = connectRealtime(srv));
  if (srv.callsLive) return;
  srv.callsLive = true;
  srv.doneCalls = srv.doneCalls || new Set();   // calls already answered or ended: a late ring for one is ignored
  srv.answeredCalls = srv.answeredCalls || new Set();   // calls the other side picked up (so a ring that has not started yet is not started)
  const onCall = (ev) => {
    const d = ev.data || {};
    if (srv.doneCalls.has(d.call)) return;
    if (logic.state.call) { logic.notify('Missed a call from ' + (d.from_name || 'a colleague') + ' while you were in a call.', '#FFB86B'); return; }
    const pr = (logic.state.dmPrivacy || {})[d.from_name] || {};
    if (d.context === 'dm' && (pr.blocked || pr.noCall || (d.kind === 'video' && pr.noVideo))) { decline(srv, d.call); return; }
    logic.setState({ incoming: { name: d.from_name, kind: d.kind, callId: d.call, context: d.context, conversation: d.conversation } });
    // a private call and a forum call ring until answered (every member of the forum gets it); a meeting shows the banner and a notice
    if (d.context === 'dm' || d.context === 'forum') startRing();
    else logic.notify((d.from_name || 'A colleague') + ' started a ' + (d.context === 'standup' ? 'standup' : 'meeting') + '. Join from the banner.', '#7CC4FF');
    notifyW.show({ title: (d.kind === 'video' ? 'Video call' : 'Call') + ' from ' + (d.from_name || 'a colleague'), body: d.context === 'dm' || d.context === 'forum' ? 'Click to answer in Hub' : 'A meeting started. Click to join.', tag: 'call:' + d.call, urgent: true });
    clearTimeout(srv.ringT);
    srv.ringT = setTimeout(() => { if (logic.state.incoming && logic.state.incoming.callId === d.call) { stopRing(); logic.setState({ incoming: null }); logic.notify('Missed ' + (d.kind === 'video' ? 'video call' : 'call') + ' from ' + d.from_name + '.', '#FFB86B'); loadHistory(logic, srv); } }, 45000);
  };
  rt.on('call.incoming', onCall);
  rt.on('call.started', onCall);
  // the list of calls in progress follows every call event, and a slow poll catches anything missed while offline
  ['call.incoming', 'call.started', 'call.answered', 'call.ended', 'call.declined', 'call.joined', 'call.left'].forEach((t) => { try { rt.on(t, () => refreshLive(logic, srv)); } catch (e) { /* no live connection */ } });
  clearInterval(srv.liveI); srv.liveI = setInterval(() => { if (!document.hidden) loadLive(logic, srv); }, 20000);
  // Answered or ended somewhere else (another tab or device of mine, or the caller hung up): stop ringing here.
  rt.on('call.answered', (ev) => {
    const d = ev.data || {};
    srv.doneCalls.add(d.call);
    srv.answeredCalls.add(d.call);
    const inc = logic.state.incoming;
    if (inc && inc.callId === d.call) { clearTimeout(srv.ringT); stopRing(); logic.setState({ incoming: null }); }
    // my ringback stops the moment they pick up. The call's id is known to the app before it is set on the call screen (which waits for
    // the microphone), so the call being started counts too.
    if ((logic.state.call && logic.state.call.callId === d.call) || srv.startingCall === d.call) stopRingback();
  });
  rt.on('call.ended', (ev) => {
    const d = ev.data || {};
    srv.doneCalls.add(d.call);
    const inc = logic.state.incoming;
    if (inc && inc.callId === d.call) { clearTimeout(srv.ringT); stopRing(); logic.setState({ incoming: null }); logic.notify('Missed a call from ' + (inc.name || 'a colleague') + '.', '#FFB86B'); }
    // the server ended the call I am in (everyone else left): hang up here too
    const cur = logic.state.call;
    if (cur && cur.callId === d.call) {
      const unanswered = !srv.answeredCalls.has(d.call) && !cur.connected;
      logic.notify(unanswered ? (cur.context === 'dm' ? 'No answer.' : 'Nobody joined.') : cur.context === 'dm' ? 'The call ended.' : 'Everyone else left, so the call ended.', '#7CC4FF');
      logic.endCall();
    }
  });
  rt.on('call.declined', (ev) => {
    const d = ev.data || {};
    if (logic.state.call && logic.state.call.callId === d.call) { logic.notify((d.by_name || 'They') + ' declined the call.', '#FF9B8A'); logic.endCall(); }
  });
}

// Deletes calls from my own history for good (the other people on them keep theirs).
export async function deleteFromLog(logic, srv, ids) {
  const real = ids.filter((id) => /^[0-9a-f]{8}-/.test(String(id)));
  if (real.length) await srv.request('POST', '/calls/history/delete', { ids: real });
  const gone = new Set(ids);
  logic.setState((st) => ({ callLog: st.callLog.filter((e) => !gone.has(e.id)) }));
}

// The Calls tab in each chat: real history from the server.
export async function loadHistory(logic, srv) {
  const { calls } = await srv.request('GET', '/calls/history?limit=200');
  const localOf = (c) => {
    if (c.context === 'forum' && srv.forumLocalId) return srv.forumLocalId.get(c.conversation) || c.conversation;
    if (c.context === 'dm') return (c.people || []).find((n) => n !== srv.me.name) || c.host_name;
    return c.conversation || '';
  };
  const dir = (c) => (c.status === 'missed' ? (c.direction === 'out' ? 'missedOut' : 'missedIn') : c.status === 'declined' ? 'declined' : c.direction || 'out');
  logic.setState({ callLog: calls.map((c) => ({ id: c.id, ctx: c.context === 'meeting' || c.context === 'standup' ? 'meeting' : c.context, ref: localOf(c), kind: c.kind, dir: dir(c), at: new Date(c.started_at).getTime(), dur: c.duration_seconds || 0 })) });
}

// Feedback: after a call, meeting or standup where someone else was in it at the same time (for a private call, one that was answered),
// everyone who was in it is asked once to rate it (1 to 5 stars, optional remark). Rating or cancelling both close it for good; only a
// tab closed with the card open asks again.
export async function checkFeedback(logic, srv) {
  if (logic.state.call || logic.state.callFb) return;
  const { due } = await srv.request('GET', '/calls/feedback');
  const next = (due || [])[0];
  if (!next || logic.state.call || logic.state.callFb) return;
  const what = next.context === 'meeting' ? 'meeting' : next.context === 'standup' ? 'standup' : next.kind === 'video' ? 'video call' : 'call';
  const with_ = (next.people || []).slice(0, 3).join(', ') + ((next.people || []).length > 3 ? ' and ' + (next.people.length - 3) + ' more' : '');
  logic.setState({ callFb: { id: next.call, what, with: with_, when: new Date(next.started_at).toLocaleString([], { weekday: 'short', hour: '2-digit', minute: '2-digit' }), stars: 0, remark: '', err: '', busy: false } });
}

export async function sendFeedback(logic, srv) {
  const fb = logic.state.callFb;
  if (!fb || fb.busy) return;
  if (!fb.stars) { logic.setState({ callFb: { ...fb, err: 'Pick 1 to 5 stars.' } }); return; }
  logic.setState({ callFb: { ...fb, busy: true, err: '' } });
  try {
    await srv.request('POST', `/calls/${fb.id}/feedback`, { stars: fb.stars, remark: fb.remark.trim() });
  } catch (e) {
    if (e.status !== 409) { logic.setState((st) => ({ callFb: st.callFb && { ...st.callFb, busy: false, err: e.message || 'That did not work.' } })); return; }
  }
  logic.setState({ callFb: null });
  logic.notify('Thanks for your feedback.', '#B8F36B');
}

export async function feedbackLater(logic, srv) {
  const fb = logic.state.callFb;
  if (!fb || fb.busy) return;
  logic.setState({ callFb: null });
  // not asked again: the server remembers, so other tabs and the next sign-in do not ask either
  await srv.request('POST', `/calls/${fb.id}/feedback/skip`).catch(() => {});
}
