// Phone-style tones made with the Web Audio API, so no sound files are needed.
// Browsers only allow sound after the person has clicked or typed once; before that these do nothing.
let ctx = null, ringTimer = null, ringStop = null;

function audio() {
  if (typeof window === 'undefined') return null;
  try {
    if (!ctx) { const AC = window.AudioContext || window.webkitAudioContext; if (!AC) return null; ctx = new AC(); }
    if (ctx.state === 'suspended') ctx.resume();
    return ctx;
  } catch (e) { return null; }
}

// One soft beep: frequency list plays together (a dual tone), shaped so it does not click.
function beep(freqs, at, len, vol) {
  const a = audio(); if (!a) return;
  const t0 = a.currentTime + at;
  const g = a.createGain();
  g.gain.setValueAtTime(0.0001, t0);
  g.gain.exponentialRampToValueAtTime(vol, t0 + 0.02);
  g.gain.setValueAtTime(vol, t0 + Math.max(0.03, len - 0.04));
  g.gain.exponentialRampToValueAtTime(0.0001, t0 + len);
  g.connect(a.destination);
  freqs.forEach((f) => { const o = a.createOscillator(); o.type = 'sine'; o.frequency.value = f; o.connect(g); o.start(t0); o.stop(t0 + len + 0.02); });
}

// A new notification or message: two short rising notes.
export function chime() { beep([880], 0, 0.14, 0.18); beep([1320], 0.15, 0.22, 0.16); }

// Incoming call: a repeating two-burst ring, like a desk phone.
export function startRing() {
  stopRing();
  const once = () => { beep([440, 480], 0, 0.4, 0.2); beep([440, 480], 0.5, 0.4, 0.2); };
  once(); ringTimer = setInterval(once, 2400);
}
export function stopRing() { if (ringTimer) { clearInterval(ringTimer); ringTimer = null; } }

// Outgoing call: the long "ringing" tone the caller hears, until the other side answers.
export function startRingback(ms) {
  stopRingback();
  const once = () => beep([440, 480], 0, 1.4, 0.16);
  once(); ringTimer = setInterval(once, 3000);
  ringStop = setTimeout(() => { stopRingback(); connected(); }, ms || 3200);
}
export function stopRingback() { stopRing(); if (ringStop) { clearTimeout(ringStop); ringStop = null; } }

// Answered / call connected: a short up-blip. Hang-up or declined: a low double beep.
export function connected() { beep([660], 0, 0.1, 0.16); beep([880], 0.12, 0.14, 0.16); }
export function hangup() { beep([480], 0, 0.18, 0.16); beep([480], 0.26, 0.18, 0.16); }
export function stopAll() { stopRingback(); }
