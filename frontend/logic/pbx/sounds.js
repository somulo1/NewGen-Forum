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
// Written the way a care agent speaks: short phrases, commas where a person breathes, a warm close.
export const GREETING = 'Hello, and thank you for calling our customer care. Your call is important to us, and we are glad you reached out. For your security and ours, this call may be recorded. Please stay on the line. One of our team will be with you shortly.';

// ---- the tune: a soft pad and melody, scheduled ahead on the audio clock so it never stutters or clicks ----
let loop = null;   // { timer, bus, music, next, step }
export function tunePlaying() { return !!loop; }

// one soft note: slow attack and release through a warm low-pass filter, so it glides instead of beeping
function pad(a, dest, freqs, t0, len, vol, wave, cutoff) {
  const g = a.createGain(), f = a.createBiquadFilter();
  f.type = 'lowpass'; f.frequency.value = cutoff || 1400; f.Q.value = 0.4;
  const att = Math.min(0.35, len * 0.3), rel = Math.min(0.6, len * 0.45);
  g.gain.setValueAtTime(0.0001, t0);
  g.gain.linearRampToValueAtTime(vol, t0 + att);
  g.gain.setValueAtTime(vol, t0 + len - rel);
  g.gain.linearRampToValueAtTime(0.0001, t0 + len);
  f.connect(g); g.connect(dest);
  freqs.forEach((hz, i) => {
    // two slightly detuned voices per note make it sound full and smooth (a "chorus")
    [-4, 4].forEach((cents) => { const o = a.createOscillator(); o.type = wave; o.frequency.value = hz; o.detune.value = cents + i; o.connect(f); o.start(t0); o.stop(t0 + len + 0.05); });
  });
}

export function startTune(name, withGreeting) {
  stopTune();
  const a = audio(); if (!a) return;
  const t = TUNES[name] || TUNES['Calm care'];
  const bus = a.createGain(); bus.gain.value = 0.9; bus.connect(a.destination);
  const music = a.createGain(); music.gain.value = 1;
  // a short echo gives the room-like tail callers know from hold music
  const dly = a.createDelay(1); dly.delayTime.value = 0.32; const fb = a.createGain(); fb.gain.value = 0.28; const wet = a.createGain(); wet.gain.value = 0.35;
  music.connect(bus); music.connect(dly); dly.connect(fb); fb.connect(dly); dly.connect(wet); wet.connect(bus);
  const beat = 60 / t.bpm, bar = beat * 4, steps = t.melody.length;
  const st = { timer: 0, bus, music, next: a.currentTime + 0.1, step: 0 };
  // look 0.5 s ahead every 100 ms and book the notes due in that window, on the audio clock (not the page's timers)
  const schedule = () => {
    while (st.next < a.currentTime + 0.5) {
      const i = st.step % steps;
      if (i % 4 === 0) pad(a, music, t.chords[(i / 4) % t.chords.length], st.next, bar * 1.02, 0.03, 'triangle', 900);
      const f = t.melody[i]; if (f) pad(a, music, [f], st.next, beat * 1.4, 0.045, 'sine', 2200);
      st.next += beat; st.step++;
    }
  };
  schedule(); st.timer = setInterval(schedule, 100);
  loop = st;
  if (withGreeting) speak(GREETING, { duck: true });
}
export function stopTune() {
  if (loop) {
    clearInterval(loop.timer);
    const b = loop.bus;
    try { b.gain.setTargetAtTime(0.0001, audio().currentTime, 0.25); setTimeout(() => b.disconnect(), 1500); } catch (e) { /* closed */ }
    loop = null;
  }
  stopSpeech();
}

// ---- the sample "AI" voice ----
// The browser's own speech, made to sound like a care agent: the most natural voice it has, read sentence by sentence with
// pauses between them, a slightly slower pace, a gentle lift on questions and a calmer, lower close. Long text is never sent
// in one piece (some browsers cut a long utterance off part way, which is what made it break; short pieces never reach that limit).
let speechRun = 0;
function voices() { return (typeof window !== 'undefined' && window.speechSynthesis && window.speechSynthesis.getVoices()) || []; }
const voicesReady = () => new Promise((res) => {
  if (voices().length) return res();
  const ss = window.speechSynthesis; const done = () => { ss.removeEventListener && ss.removeEventListener('voiceschanged', done); res(); };
  ss.addEventListener ? ss.addEventListener('voiceschanged', done) : (ss.onvoiceschanged = done);
  setTimeout(res, 1500);
});
// the best English voice the device has: neural and "natural" voices first, then known smooth female voices, then any English one
export function bestVoice() {
  const en = voices().filter((v) => /^en([-_]|$)/i.test(v.lang));
  const score = (v) => {
    let s = 0; const n = v.name;
    if (/natural|neural|online/i.test(n)) s += 50;
    if (/Google UK English Female|Google US English|Samantha|Serena|Karen|Moira|Tessa|Fiona|Libby|Sonia|Aria|Jenny|Ava|Allison|Susan|Hazel|Zira|female/i.test(n)) s += 20;
    if (/en[-_](GB|KE|IE|ZA)/i.test(v.lang)) s += 6; else if (/en[-_](US|AU)/i.test(v.lang)) s += 4;
    if (/compact|espeak|robot|novelty|whisper|bells|bad news|zarvox|trinoids|albert|fred|junior/i.test(n)) s -= 60;
    if (!v.localService) s += 3;   // network voices are usually the higher-quality ones
    return s;
  };
  return en.sort((a, b) => score(b) - score(a))[0] || null;
}
export function stopSpeech() {
  speechRun++;
  if (typeof window !== 'undefined' && window.speechSynthesis) window.speechSynthesis.cancel();
}
// break the text into sentences, and long sentences at their commas, so each piece is read with its own natural intonation
function phrases(text) {
  return String(text).replace(/\s+/g, ' ').match(/[^.!?]+[.!?]*/g).map((x) => x.trim()).filter(Boolean)
    .flatMap((sen) => (sen.length > 140 ? sen.split(/(?<=,)\s+/) : [sen]));
}
export async function speak(text, opts) {
  if (typeof window === 'undefined' || !window.speechSynthesis) return;
  stopSpeech();
  const run = speechRun, ss = window.speechSynthesis;
  await voicesReady();
  if (run !== speechRun) return;
  const v = bestVoice(), parts = phrases(text);
  // the music steps back while the voice speaks, then comes up again
  const duck = (down) => { if (opts && opts.duck && loop) { try { loop.music.gain.setTargetAtTime(down ? 0.35 : 1, audio().currentTime, 0.4); } catch (e) { /* closed */ } } };
  duck(true);
  for (let i = 0; i < parts.length; i++) {
    if (run !== speechRun) return;
    const p = parts[i], last = i === parts.length - 1, q = /\?$/.test(p);
    await new Promise((res) => {
      const u = new SpeechSynthesisUtterance(p);
      if (v) { u.voice = v; u.lang = v.lang; }
      u.rate = last ? 0.88 : 0.93;            // unhurried, and slower still for the closing line
      u.pitch = q ? 1.08 : last ? 0.97 : 1.02; // a lift on questions, a settled close
      u.volume = 1;
      u.onend = u.onerror = () => res();
      ss.speak(u);
      setTimeout(res, 2000 + p.length * 120); // never wait for ever if the browser drops the end event
    });
    if (run !== speechRun) return;
    // a breath between sentences, a little longer after a full stop than after a comma
    await new Promise((r) => setTimeout(r, /,$/.test(p) ? 180 : 380));
  }
  if (run === speechRun) duck(false);
}
