// The app's sounds, played with the Web Audio API from parameters the server serves (GET /v1/tones), so no sound files are
// needed and every tone can be changed in the database. Until the server's set has loaded, the built-in copy below plays.
// Browsers only allow sound after the person has clicked or typed once; before that these do nothing.
let ctx = null, ringTimer = null, backTimer = null, ringStop = null;
const looping = { ring: new Set(), ringback: new Set() };   // oscillators of a ring or ringback that has been scheduled; stopping the ring silences them at once

const BUILT_IN = {
  chime: { notes: [{ freqs: [880], at: 0, len: 0.14, vol: 0.18 }, { freqs: [1320], at: 0.15, len: 0.22, vol: 0.16 }] },
  ring: { repeat_ms: 2400, notes: [{ freqs: [440, 480], at: 0, len: 0.4, vol: 0.2 }, { freqs: [440, 480], at: 0.5, len: 0.4, vol: 0.2 }] },
  ringback: { repeat_ms: 3000, then: 'connected', then_after_ms: 3200, notes: [{ freqs: [440, 480], at: 0, len: 1.4, vol: 0.16 }] },
  connected: { notes: [{ freqs: [660], at: 0, len: 0.1, vol: 0.16 }, { freqs: [880], at: 0.12, len: 0.14, vol: 0.16 }] },
  hangup: { notes: [{ freqs: [480], at: 0, len: 0.18, vol: 0.16 }, { freqs: [480], at: 0.26, len: 0.18, vol: 0.16 }] },
  message: { notes: [{ freqs: [740], at: 0, len: 0.1, vol: 0.2 }, { freqs: [988], at: 0.1, len: 0.16, vol: 0.2 }] },
  joined: { notes: [{ freqs: [587], at: 0, len: 0.1, vol: 0.16 }, { freqs: [784], at: 0.11, len: 0.1, vol: 0.16 }, { freqs: [1175], at: 0.22, len: 0.18, vol: 0.16 }] },
  hand: { wave: 'triangle', notes: [{ freqs: [1319], at: 0, len: 0.08, vol: 0.2 }, { freqs: [1319], at: 0.12, len: 0.08, vol: 0.2 }, { freqs: [1760], at: 0.24, len: 0.16, vol: 0.2 }] },
  alert: { wave: 'triangle', notes: [{ freqs: [784], at: 0, len: 0.12, vol: 0.2 }, { freqs: [1047], at: 0.14, len: 0.12, vol: 0.2 }, { freqs: [1319], at: 0.28, len: 0.22, vol: 0.18 }] }
};
let tones = { ...BUILT_IN };

// Replaces the built-in set with the server's (called once after sign-in).
export async function loadTones(baseUrl) {
  try {
    const r = await fetch(baseUrl.replace(/\/$/, '') + '/v1/tones');
    if (!r.ok) return;
    const { tones: list } = await r.json();
    const next = { ...BUILT_IN };
    list.forEach((t) => { next[t.key] = t.params; });
    tones = next;
  } catch (e) { /* offline: keep the built-in set */ }
}

function audio() {
  if (typeof window === 'undefined') return null;
  try {
    if (!ctx) { const AC = window.AudioContext || window.webkitAudioContext; if (!AC) return null; ctx = new AC(); }
    return ctx;
  } catch (e) { return null; }
}

// Browsers keep audio switched off until the person clicks or types. Calls ring from the network, not from a click, so
// the audio is switched on at the first click or key anywhere in the app; after that every tone can play.
if (typeof window !== 'undefined') {
  const unlock = () => {
    const a = audio();
    if (a && a.state !== 'running') a.resume().catch(() => {});
    if (a && a.state === 'running') ['pointerdown', 'keydown', 'touchstart'].forEach((t) => window.removeEventListener(t, unlock, true));
  };
  ['pointerdown', 'keydown', 'touchstart'].forEach((t) => window.addEventListener(t, unlock, true));
}

// One note: the frequencies play together, shaped by the tone's attack and release so it does not click.
function note(p, n, loop) {
  const a = audio(); if (!a) return;
  const attack = p.attack || 0.02, release = p.release || 0.04, floor = p.floor || 0.0001;
  const t0 = a.currentTime + n.at;
  const g = a.createGain();
  g.gain.setValueAtTime(floor, t0);
  g.gain.exponentialRampToValueAtTime(n.vol, t0 + attack);
  g.gain.setValueAtTime(n.vol, t0 + Math.max(attack + 0.01, n.len - release));
  g.gain.exponentialRampToValueAtTime(floor, t0 + n.len);
  g.connect(a.destination);
  n.freqs.forEach((f) => { const o = a.createOscillator(); o.type = p.wave || 'sine'; o.frequency.value = f; o.connect(g); o.start(t0); o.stop(t0 + n.len + 0.02); if (loop) { looping[loop].add(o); o.onended = () => looping[loop].delete(o); } });
}

export function play(key) {
  const p = tones[key] || BUILT_IN[key];
  const a = audio();
  if (!p || !a) return;
  const loop = key === 'ring' || key === 'ringback' ? key : '';
  // a ring that was queued while the browser still blocked sound must not play late, after the call was answered
  const wanted = () => !loop || (key === 'ring' ? !!ringTimer : !!backTimer);
  const go = () => { if (wanted()) p.notes.forEach((n) => note(p, n, loop)); };
  if (a.state === 'running') go(); else a.resume().then(() => { if (a.state === 'running') go(); }).catch(() => {});
}

// A new notification.
export function chime() { play('chime'); }
// A message in the chat that is open on screen: a soft pop.
export function messageTone() { play('message'); }
// A message for a chat that is closed: brighter, so you know to open it.
export function alertTone() { play('alert'); }

// Incoming call: repeats until answered or declined.
export function startRing() {
  stopRing();
  ringTimer = setInterval(() => play('ring'), (tones.ring && tones.ring.repeat_ms) || 2400);
  play('ring');
}
function silence(k) { looping[k].forEach((o) => { try { o.stop(); } catch (e) { /* already stopped */ } }); looping[k].clear(); }
export function stopRing() { if (ringTimer) { clearInterval(ringTimer); ringTimer = null; } silence('ring'); }

// Outgoing call: ringing until the other side answers. Without `ms` it rings until stopRingback() is called.
export function startRingback(ms) {
  stopRingback();
  backTimer = setInterval(() => play('ringback'), (tones.ringback && tones.ringback.repeat_ms) || 3000);
  play('ringback');
  if (ms) ringStop = setTimeout(() => { stopRingback(); connected(); }, ms);
}
export function stopRingback() { if (backTimer) { clearInterval(backTimer); backTimer = null; } if (ringStop) { clearTimeout(ringStop); ringStop = null; } silence('ringback'); }

export function connected() { play('connected'); }
export function hangup() { play('hangup'); }
// Someone joined the call you are in.
export function joinedTone() { play('joined'); }
// Someone in the call raised their hand.
export function handTone() { play('hand'); }
export function stopAll() { stopRingback(); stopRing(); }
