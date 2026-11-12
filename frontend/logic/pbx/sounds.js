// Calls management sounds. Ringing, ringback, connect and hang-up are the app's own call tones (logic/tones.js), so customer calls
// sound like every other call in Hub. Key tones and the customer care tune are sampled here with the Web Audio API, and the care
// tune's greeting is read by the browser's built-in voice: a sample until the PBX plays the real recordings to callers.
import * as tones from '../tones.js';

export const { startRing, stopRing, startRingback, stopRingback, connected, hangup } = tones;

let ctx = null;
function audio() {
  if (typeof window === 'undefined') return null;
  try {
    if (!ctx) { const AC = window.AudioContext || window.webkitAudioContext; if (!AC) return null; ctx = new AC(); }
    if (ctx.state !== 'running') ctx.resume().catch(() => {});
    return ctx;
  } catch (e) { return null; }
}

function tone(freqs, at, len, vol, wave, dest) {
  const a = audio(); if (!a) return [];
  const t0 = a.currentTime + at, g = a.createGain();
  g.gain.setValueAtTime(0.0001, t0);
  g.gain.exponentialRampToValueAtTime(vol, t0 + Math.min(0.04, len / 3));
  g.gain.setValueAtTime(vol, t0 + Math.max(0.05, len - 0.08));
  g.gain.exponentialRampToValueAtTime(0.0001, t0 + len);
  g.connect(dest || a.destination);
  return freqs.map((f) => { const o = a.createOscillator(); o.type = wave || 'sine'; o.frequency.value = f; o.connect(g); o.start(t0); o.stop(t0 + len + 0.03); return o; });
}

// the standard telephone key tones (DTMF): a low and a high frequency for each key
const DTMF = { 1: [697, 1209], 2: [697, 1336], 3: [697, 1477], 4: [770, 1209], 5: [770, 1336], 6: [770, 1477], 7: [852, 1209], 8: [852, 1336], 9: [852, 1477], '*': [941, 1209], 0: [941, 1336], '#': [941, 1477] };
export function keyTone(k) { const f = DTMF[k]; if (f) tone(f, 0, 0.12, 0.12); }

// ---- the customer care tune (in the manner of a caller ring-back tune): a soft looping melody under a spoken greeting ----
export const TUNES = {
  'Calm care': { bpm: 84, chords: [[220, 277, 330], [196, 247, 294], [175, 220, 262], [196, 247, 294]], melody: [659, 587, 523, 587, 659, 659, 659, 0, 587, 587, 587, 0, 659, 784, 784, 0] },
  'Bright fintech': { bpm: 112, chords: [[262, 330, 392], [220, 277, 330], [175, 220, 262], [196, 247, 294]], melody: [784, 659, 784, 880, 784, 659, 523, 0, 587, 659, 698, 659, 587, 523, 587, 0] },
  'Warm afro': { bpm: 100, chords: [[196, 247, 294], [262, 330, 392], [220, 277, 330], [175, 220, 262]], melody: [587, 0, 659, 587, 523, 0, 440, 523, 587, 0, 659, 784, 659, 0, 587, 0] }
};
export const GREETING = 'Thank you for calling our customer care. Your call is important to us. For your security, this call may be recorded. Please hold while we connect you to the next available agent.';

let loop = null;   // { timer, gain, voices }
export function tunePlaying() { return !!loop; }
export function startTune(name, withGreeting) {
  stopTune();
  const a = audio(); if (!a) return;
  const t = TUNES[name] || TUNES['Calm care'];
  const bus = a.createGain(); bus.gain.value = 0.55; bus.connect(a.destination);
  const beat = 60 / t.bpm, bar = beat * 4;
  const play = () => {
    t.chords.forEach((c, i) => tone(c, i * bar, bar * 0.95, 0.035, 'triangle', bus));
    t.melody.forEach((f, i) => { if (f) tone([f], i * beat, beat * 0.9, 0.05, 'sine', bus); });
  };
  play();
  loop = { timer: setInterval(play, t.chords.length * bar * 1000), bus };
  if (withGreeting) speak(GREETING);
}
export function stopTune() {
  if (!loop) return;
  clearInterval(loop.timer);
  try { loop.bus.gain.setTargetAtTime(0.0001, audio().currentTime, 0.15); setTimeout(((b) => () => b.disconnect())(loop.bus), 800); } catch (e) { /* closed */ }
  loop = null;
  if (typeof window !== 'undefined' && window.speechSynthesis) window.speechSynthesis.cancel();
}
// the sample "AI" voice: the browser's own speech, a calm English voice when there is one
export function speak(text) {
  if (typeof window === 'undefined' || !window.speechSynthesis) return;
  const u = new SpeechSynthesisUtterance(text);
  const v = window.speechSynthesis.getVoices().find((x) => /en[-_](GB|KE|US)/i.test(x.lang) && /female|Samantha|Google UK English Female|Zira/i.test(x.name))
    || window.speechSynthesis.getVoices().find((x) => /^en/i.test(x.lang));
  if (v) u.voice = v;
  u.rate = 0.95; u.pitch = 1; u.volume = 0.9;
  window.speechSynthesis.cancel(); window.speechSynthesis.speak(u);
}
