// Big calls: voice and video rooms that run on a media server (SFU) instead of device-to-device. Every device sends its
// audio and video to the media server once; the server forwards to everyone else, so a room can hold hundreds of people.
//
// A big room is a stage and an audience. People on the stage (hosts, moderators, anyone the host brings up, and everybody in
// a small room) publish audio and video. The audience only listens, so each listener receives the few people on stage, never
// hundreds of streams. The screen shows the stage as tiles and the whole room as a searchable people list.
//
// Media is encrypted on the device before it leaves (insertable streams) with a key made from the conversation's key, which
// only members have, so the media server relays video it cannot watch. In-call chat is sealed the same way.
import { startRingback, stopRingback, connected, joinedTone, handTone } from '../tones.js';

const enc = new TextEncoder();
const dec = new TextDecoder();
const STAGE_TILES = 24;      // most tiles drawn at once
const ROSTER_EVERY = 900;    // the people list is refreshed at most this often (ms)
const TILES_EVERY = 250;

const S = { room: null, key: null, chatKey: null, vids: new Map(), audio: new Map(), timers: {}, canSpeak: false, info: null };
export const sfuActive = () => !!S.room;

async function hkdf(raw, salt, info) {
  const k = await crypto.subtle.importKey('raw', raw, 'HKDF', false, ['deriveBits']);
  return new Uint8Array(await crypto.subtle.deriveBits({ name: 'HKDF', hash: 'SHA-256', salt: enc.encode(salt), info: enc.encode(info) }, k, 256));
}

// The call key: made from the conversation's current key, so only its members can open the media. Calls that belong to no
// conversation have no such key; their media is still encrypted to the media server, but not end to end.
async function callKeys(srv, entry) {
  const conv = entry.call.conversation;
  if (!conv || !srv.vault) return null;
  try {
    const c = await srv.vault.conversation(conv, true);
    const raw = await srv.vault.keyFor(conv, c.epoch);
    if (!raw) return null;
    const bytes = raw instanceof Uint8Array ? raw : new Uint8Array(raw);
    return { media: await hkdf(bytes, entry.call.id, 'hub-call-media'), chat: await hkdf(bytes, entry.call.id, 'hub-call-chat') };
  } catch (e) { return null; }
}

async function sealChat(text) {
  if (!S.chatKey) return enc.encode(JSON.stringify({ text }));
  const key = await crypto.subtle.importKey('raw', S.chatKey, 'AES-GCM', false, ['encrypt']);
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const ct = new Uint8Array(await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, enc.encode(JSON.stringify({ text }))));
  const out = new Uint8Array(1 + 12 + ct.length);
  out[0] = 1; out.set(iv, 1); out.set(ct, 13);
  return out;
}
async function openChat(bytes) {
  if (bytes[0] === 1 && S.chatKey) {
    const key = await crypto.subtle.importKey('raw', S.chatKey, 'AES-GCM', false, ['decrypt']);
    return JSON.parse(dec.decode(await crypto.subtle.decrypt({ name: 'AES-GCM', iv: bytes.slice(1, 13) }, key, bytes.slice(13)))).text;
  }
  return JSON.parse(dec.decode(bytes)).text;
}

const metaOf = (p) => { try { return JSON.parse(p.metadata || '{}'); } catch (e) { return {}; } };
const canPublish = (p) => !!(p.permissions && p.permissions.canPublish);
const personView = (p) => ({
  id: p.identity, person: metaOf(p).person || '', name: p.name || 'Someone', role: metaOf(p).role || 'participant',
  mic: p.isMicrophoneEnabled, cam: p.isCameraEnabled, sharing: p.isScreenShareEnabled, hand: (p.attributes || {}).hand === '1', stage: canPublish(p), speaking: !!p.isSpeaking
});

// ---- what the screen shows ----

// The tiles: people who can speak, the loudest first, then those with a camera, capped so a big room stays light.
function stageOf(room) {
  const rem = [...room.remoteParticipants.values()].filter(canPublish).map(personView);
  const loud = new Map(room.activeSpeakers.map((p, i) => [p.identity, i]));
  rem.sort((a, b) => (loud.has(a.id) ? loud.get(a.id) : 99) - (loud.has(b.id) ? loud.get(b.id) : 99) || (b.cam - a.cam) || (b.sharing - a.sharing));
  return rem.slice(0, STAGE_TILES);
}

function refresh(ctx, now = false) {
  const { logic } = ctx;
  const run = () => {
    S.timers.tiles = 0;
    if (!S.room) return;
    const room = S.room;
    const peers = stageOf(room);
    const sp = new Set(room.activeSpeakers.map((p) => p.identity));
    ctx.setCall(logic, { peers: peers.map((p) => ({ ...p, speaking: sp.has(p.id) })), total: room.remoteParticipants.size + 1, canSpeak: S.canSpeak, speakingNow: [...sp] });
    syncVideo(ctx, peers);
  };
  if (now) { clearTimeout(S.timers.tiles); run(); return; }
  if (!S.timers.tiles) S.timers.tiles = setTimeout(run, TILES_EVERY);
}

// The whole room, for the people list. Kept out of the tiles path and refreshed slowly: with hundreds of people a join or
// leave every moment must not redraw the screen every moment.
function refreshRoster(ctx) {
  if (S.timers.roster) return;
  S.timers.roster = setTimeout(() => {
    S.timers.roster = 0;
    if (!S.room) return;
    const list = [...S.room.remoteParticipants.values()].map(personView);
    list.sort((a, b) => (b.hand - a.hand) || (b.stage - a.stage) || a.name.localeCompare(b.name));
    ctx.setCall(ctx.logic, { roster: list });
  }, ROSTER_EVERY);
}

// The video of each tile: LiveKit decides which quality to send from the size the tile is drawn at.
function syncVideo(ctx, peers) {
  const want = new Set(peers.map((p) => p.id));
  for (const [id] of ctx.media.tracks) if (id !== 'local' && !want.has(id)) ctx.media.tracks.delete(id);
  for (const p of peers) {
    const v = S.vids.get(p.id);
    const t = v && (v.screen || v.camera);
    if (t && ctx.media.tracks.get(p.id) !== t) ctx.media.tracks.set(p.id, t);
  }
  ctx.changed();
}

// ---- joining ----

export async function connectSfu(ctx, entry, opts = {}) {
  const { logic, srv, media } = ctx;
  const info = entry.sfu;
  S.info = info; S.canSpeak = !!info.can_speak;
  if (opts.ringback) startRingback();
  const lk = await import('livekit-client');
  const keys = await callKeys(srv, entry);
  S.key = keys && keys.media; S.chatKey = keys && keys.chat;
  const opt = {
    adaptiveStream: true, dynacast: true,
    publishDefaults: { simulcast: true, dtx: true, red: true, videoCodec: 'vp8' },
    videoCaptureDefaults: { resolution: lk.VideoPresets.h540.resolution },
    audioCaptureDefaults: { echoCancellation: true, noiseSuppression: true, autoGainControl: true }
  };
  if (S.key) {
    const keyProvider = new lk.ExternalE2EEKeyProvider();
    await keyProvider.setKey(S.key);
    opt.e2ee = { keyProvider, worker: new Worker(new URL('livekit-client/e2ee-worker', import.meta.url)) };
  }
  const room = new lk.Room(opt);
  S.room = room;
  ctx.setRoom({ sfu: true, callId: entry.call.id, host: entry.call.host, ctx: entry.call.context || 'meeting' });
  ctx.setCall(logic, { callId: entry.call.id, peers: [], me: info.identity, myRole: info.role, sfu: true, canSpeak: S.canSpeak, limit: info.limit, total: 1, roster: [], e2ee: !!S.key });
  const E = lk.RoomEvent;
  const re = () => refresh(ctx);

  room.on(E.ParticipantConnected, (p) => {
    re(); refreshRoster(ctx);
    if (room.remoteParticipants.size <= 30) { joinedTone(); logic.notify((p.name || 'Someone') + ' joined the call.', '#B8F36B'); }
    stopRingback();
    if (ctx.callContext() === 'dm') connected();
  });
  room.on(E.ParticipantDisconnected, (p) => {
    for (const key of [p.identity]) { S.vids.delete(key); media.tracks.delete(key); }
    re(); refreshRoster(ctx);
  });
  room.on(E.ActiveSpeakersChanged, re);
  room.on(E.TrackMuted, (pub, p) => { if (p.isLocal) syncLocalState(ctx); re(); });
  room.on(E.TrackUnmuted, (pub, p) => { if (p.isLocal) syncLocalState(ctx); re(); });
  room.on(E.TrackPublished, re);
  room.on(E.TrackUnpublished, re);
  room.on(E.ParticipantAttributesChanged, (changed, p) => {
    if (!p.isLocal && 'hand' in changed && changed.hand === '1') { handTone(); logic.notify((p.name || 'Someone') + ' raised their hand.', '#FFD66B'); }
    re(); refreshRoster(ctx);
  });
  room.on(E.ParticipantPermissionsChanged, async (prev, p) => {
    if (p.isLocal) await onMyPermission(ctx, p);
    re(); refreshRoster(ctx);
  });
  room.on(E.TrackSubscribed, (track, pub, p) => {
    if (track.kind === 'audio') {
      const el = track.attach(); el.style.display = 'none'; document.body.appendChild(el);
      S.audio.set(pub.trackSid, el);
    } else if (track.kind === 'video') {
      const v = S.vids.get(p.identity) || {};
      v[pub.source === lk.Track.Source.ScreenShare ? 'screen' : 'camera'] = track;
      S.vids.set(p.identity, v);
      re();
    }
  });
  room.on(E.TrackUnsubscribed, (track, pub, p) => {
    if (track.kind === 'audio') { const el = S.audio.get(pub.trackSid); if (el) { track.detach(el); el.remove(); S.audio.delete(pub.trackSid); } }
    else if (track.kind === 'video') {
      const v = S.vids.get(p.identity);
      if (v) { delete v[pub.source === lk.Track.Source.ScreenShare ? 'screen' : 'camera']; if (!v.camera && !v.screen) S.vids.delete(p.identity); }
      media.tracks.delete(p.identity); re();
    }
  });
  room.on(E.DataReceived, async (payload, p, kind, topic) => {
    if (topic !== 'chat' || !p) return;
    try { const text = await openChat(payload); if (text) logic.setState((st) => ({ callMsgs: st.callMsgs.concat([{ who: p.name || 'Someone', text }]) })); } catch (e) { /* not for this key */ }
  });
  room.on(E.EncryptionError, () => logic.notify('A participant\'s media could not be decrypted.', '#FFB86B'));
  room.on(E.Disconnected, (reason) => {
    if (S.room !== room) return;
    const R = lk.DisconnectReason;
    if (reason === R.PARTICIPANT_REMOVED) logic.notify('The host removed you from the call.', '#FF9B8A');
    else if (reason === R.ROOM_DELETED) logic.notify('The call was ended by the host.', '#7CC4FF');
    else if (reason === R.DUPLICATE_IDENTITY) logic.notify('You joined this call from another place.', '#FFB86B');
    else logic.notify('The call connection closed.', '#FFB86B');
    if (logic.state.call) logic.endCall();
  });
  room.on(E.Reconnecting, () => logic.notify('Reconnecting to the call…', '#FFB86B'));
  room.on(E.Reconnected, () => logic.notify('Back in the call.', '#B8F36B'));

  await room.connect(info.url, info.token, { autoSubscribe: true });
  if (S.key) await room.setE2EEEnabled(true);
  ctx.setCall(logic, { me: room.localParticipant.identity, connected: true });
  if (room.remoteParticipants.size || ctx.callContext() !== 'dm') { stopRingback(); connected(); }
  refresh(ctx, true); refreshRoster(ctx);
  await applySfu(ctx);
}

// What the host's controls changed on my side.
function syncLocalState(ctx) {
  const lp = S.room && S.room.localParticipant;
  if (!lp) return;
  const st = ctx.logic.state;
  const patch = {};
  if (st.mic !== lp.isMicrophoneEnabled && S.canSpeak) patch.mic = lp.isMicrophoneEnabled;
  if (st.cam !== lp.isCameraEnabled && S.canSpeak && st.call && st.call.mode === 'video') patch.cam = lp.isCameraEnabled;
  if (Object.keys(patch).length) {
    ctx.logic.setState(patch);
    if ('mic' in patch && !patch.mic) ctx.logic.notify('The host muted you.', '#FFB86B');
    if ('cam' in patch && !patch.cam) ctx.logic.notify('The host turned off your camera.', '#FFB86B');
  }
}

async function onMyPermission(ctx, me) {
  const now = canPublish(me);
  if (now === S.canSpeak) return;
  S.canSpeak = now;
  ctx.setCall(ctx.logic, { canSpeak: now });
  if (now) {
    ctx.logic.setState({ hand: false, mic: true, cam: !!(ctx.logic.state.call && ctx.logic.state.call.mode === 'video') });
    ctx.logic.notify('The host brought you on stage. You can speak now.', '#B8F36B');
    await applySfu(ctx);
  } else {
    ctx.logic.setState({ mic: false, cam: false, sharing: false });
    ctx.logic.notify('You are back in the audience.', '#FFB86B');
    sendHand(ctx);
  }
}

function sendHand(ctx) {
  const lp = S.room && S.room.localParticipant;
  if (lp) lp.setAttributes({ hand: ctx.logic.state.hand ? '1' : '0' }).catch(() => {});
}

// Mic, camera, screen share and hand follow the app's state.
export async function applySfu(ctx) {
  const room = S.room;
  if (!room) return;
  const lp = room.localParticipant;
  const s = ctx.logic.state;
  const video = s.call && s.call.mode === 'video';
  if (!S.canSpeak) {
    if (s.mic || s.cam || s.sharing) {
      // a listener who presses the microphone is asking to speak
      ctx.logic.setState({ mic: false, cam: false, sharing: false, hand: true });
      ctx.logic.notify('You are in the audience. Your hand is raised; the host can bring you on stage.', '#FFD66B');
    }
    sendHand(ctx);
    return;
  }
  try {
    if (!!s.mic !== lp.isMicrophoneEnabled) await lp.setMicrophoneEnabled(!!s.mic);
    if (video && !!s.cam !== lp.isCameraEnabled) await lp.setCameraEnabled(!!s.cam);
    if (video && !!s.sharing !== lp.isScreenShareEnabled) await lp.setScreenShareEnabled(!!s.sharing, { audio: true });
  } catch (e) {
    if (video && s.cam) ctx.logic.setState({ cam: false });
    if (s.sharing) ctx.logic.setState({ sharing: false });
    if (!lp.isMicrophoneEnabled && s.mic) ctx.logic.notify('Hub cannot use your microphone. Check the browser permission; you can still listen.', '#FFB86B');
  }
  const cam = [...lp.videoTrackPublications.values()].find((p) => p.source === 'camera');
  const scr = [...lp.videoTrackPublications.values()].find((p) => p.source === 'screen_share');
  const t = (scr && scr.videoTrack) || (cam && cam.videoTrack);
  if (t) ctx.media.tracks.set('local', t); else ctx.media.tracks.delete('local');
  ctx.changed();
  sendHand(ctx);
}

export async function sendChatSfu(text) {
  if (!S.room) return;
  await S.room.localParticipant.publishData(await sealChat(text), { reliable: true, topic: 'chat' });
}

export function leaveSfu() {
  const room = S.room;
  S.room = null;
  Object.values(S.timers).forEach((t) => clearTimeout(t));
  S.timers = {};
  S.audio.forEach((el) => el.remove());
  S.audio.clear(); S.vids.clear(); S.key = S.chatKey = null; S.canSpeak = false;
  if (room) { try { room.disconnect(); } catch (e) { /* closed */ } }
}

// A room control the host or a moderator asked for: mute | cam_off | remove | mute_all | lock | unlock | end | stage_on | stage_off
export async function control(srv, callId, action, identity = '') {
  await srv.request('POST', `/calls/${callId}/control`, { action, identity });
}
