// How a call's connection is doing, measured on this device every few seconds from the browser's WebRTC stats. It drives the
// "Weak connection" sign, and when the call ends one summary of numbers (no names, no content) goes to the server's monitoring.
const EVERY_MS = 5000;
const WEAK = { lossPct: 8, rttMs: 500, jitterMs: 100 };   // past any of these, speech and video visibly suffer
const FLIP_AFTER = 2;   // samples in a row before the sign changes, so one bad moment does not make it flicker

// One sample from raw stats reports. prev holds the packet counters of the last sample, because those only ever go up.
export function readSample(reports, prev = new Map()) {
  const next = new Map();
  const seen = new Set();
  let lost = 0, received = 0, rtt = null, jitter = 0, jitters = 0, sentLoss = 0;
  for (const report of reports) {
    if (!report) continue;
    report.forEach((s) => {
      if (seen.has(s.id)) return;
      seen.add(s.id);
      if (s.type === 'candidate-pair' && s.nominated && typeof s.currentRoundTripTime === 'number') rtt = Math.max(rtt || 0, s.currentRoundTripTime * 1000);
      else if (s.type === 'remote-inbound-rtp') {
        if (typeof s.fractionLost === 'number') sentLoss = Math.max(sentLoss, s.fractionLost * 100);
        if (rtt === null && typeof s.roundTripTime === 'number') rtt = s.roundTripTime * 1000;
      } else if (s.type === 'inbound-rtp') {
        const was = prev.get(s.id) || { lost: s.packetsLost || 0, received: s.packetsReceived || 0 };
        lost += Math.max(0, (s.packetsLost || 0) - was.lost);
        received += Math.max(0, (s.packetsReceived || 0) - was.received);
        next.set(s.id, { lost: s.packetsLost || 0, received: s.packetsReceived || 0 });
        if (typeof s.jitter === 'number') { jitter += s.jitter * 1000; jitters++; }
      }
    });
  }
  const lostPct = lost + received > 0 ? (lost / (lost + received)) * 100 : 0;
  return {
    sample: { rttMs: rtt, lossPct: Math.max(lostPct, sentLoss), jitterMs: jitters ? jitter / jitters : 0, measured: seen.size > 0 },
    counters: next
  };
}

export const isWeak = (s) => s.lossPct >= WEAK.lossPct || (s.rttMs !== null && s.rttMs >= WEAK.rttMs) || s.jitterMs >= WEAK.jitterMs;

let active = null;

// getReports() returns the current stats reports; onWeak(true|false) is called only when the state changes.
export function startQualityWatch({ mode, getReports, onWeak, every = EVERY_MS }) {
  stopQualityWatch();
  const w = { mode, counters: new Map(), samples: 0, rtt: 0, rtts: 0, loss: 0, jitter: 0, weakSamples: 0, reconnects: 0, joinMs: 0, weak: false, streak: 0, timer: null };
  active = w;
  w.timer = setInterval(async () => {
    let reports;
    try { reports = await getReports(); } catch (e) { return; }
    if (active !== w) return;
    const { sample, counters } = readSample(reports, w.counters);
    const first = w.counters.size === 0;   // the first sample only sets the counters
    w.counters = counters;
    if (!sample.measured || first) return;
    w.samples++;
    w.loss += sample.lossPct; w.jitter += sample.jitterMs;
    if (sample.rttMs !== null) { w.rtt += sample.rttMs; w.rtts++; }
    const weak = isWeak(sample);
    if (weak) w.weakSamples++;
    w.streak = weak === w.weak ? 0 : w.streak + 1;
    if (w.streak >= FLIP_AFTER) { w.weak = weak; w.streak = 0; onWeak(weak); }
  }, every);
}

export function noteReconnect() { if (active) active.reconnects++; }
export function noteJoined(ms) { if (active) active.joinMs = Math.round(ms); }

// Ends the watch and returns what the server records, or null when too little was measured.
export function stopQualityWatch() {
  const w = active;
  active = null;
  if (!w) return null;
  clearInterval(w.timer);
  if (!w.samples && !w.reconnects) return null;
  const avg = (sum, n) => (n ? Math.round((sum / n) * 10) / 10 : 0);
  return { mode: w.mode, samples: w.samples, rtt_ms: avg(w.rtt, w.rtts), loss_pct: avg(w.loss, w.samples), jitter_ms: avg(w.jitter, w.samples), weak_samples: w.weakSamples, reconnects: w.reconnects, join_ms: w.joinMs };
}
