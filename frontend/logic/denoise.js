// AI noise cancellation for calls, meetings and standups, always on.
// The browser's own noise suppression handles steady background noise. For loud noise right next to the speaker (people
// talking nearby, a TV, traffic, typing, a crowd), every microphone track also goes through RNNoise, a small neural network
// trained to separate speech from noise, then a noise gate that silences what is left between words. It all runs on the
// person's own device (WebAssembly in an audio worklet): no audio leaves for processing, and it adds about 10 ms of delay.
import { RnnoiseWorkletNode, NoiseGateWorkletNode, loadRnnoise } from '@sapphi-red/web-noise-suppressor';
import rnnoiseWorklet from '@sapphi-red/web-noise-suppressor/rnnoiseWorklet.js?url';
import noiseGateWorklet from '@sapphi-red/web-noise-suppressor/noiseGateWorklet.js?url';
import rnnoiseWasm from '@sapphi-red/web-noise-suppressor/rnnoise.wasm?url';
import rnnoiseSimdWasm from '@sapphi-red/web-noise-suppressor/rnnoise_simd.wasm?url';

let wasm = null;   // the model, downloaded once (~100 KB) and kept
const loadModel = () => (wasm = wasm || loadRnnoise({ url: rnnoiseWasm, simdUrl: rnnoiseSimdWasm }).catch((e) => { wasm = null; throw e; }));

/**
 * Returns a cleaned copy of a microphone track, or null when the browser cannot run it (the call then uses the plain
 * microphone, which still has the browser's own noise suppression).
 * @param {MediaStreamTrack} track  the microphone track
 * @param {{keepSource?: boolean}} opts  keepSource: the caller owns the source track (LiveKit); otherwise stopping the
 *   cleaned track also stops the microphone.
 */
export async function denoiseTrack(track, opts = {}) {
  if (!track || track.kind !== 'audio' || typeof window === 'undefined' || !window.AudioContext || !window.AudioWorkletNode) return null;
  let ctx = null;
  try {
    ctx = new AudioContext({ sampleRate: 48000 });   // RNNoise works on 48 kHz audio
    const [bin] = await Promise.all([loadModel(), ctx.audioWorklet.addModule(rnnoiseWorklet), ctx.audioWorklet.addModule(noiseGateWorklet)]);
    const src = ctx.createMediaStreamSource(new MediaStream([track]));
    const rnnoise = new RnnoiseWorkletNode(ctx, { maxChannels: 1, wasmBinary: bin });
    // the gate opens for speech and closes for the faint noise RNNoise leaves between words
    const gate = new NoiseGateWorkletNode(ctx, { openThreshold: -50, closeThreshold: -58, holdMs: 120, maxChannels: 1 });
    const out = ctx.createMediaStreamDestination();
    src.connect(rnnoise); rnnoise.connect(gate); gate.connect(out);
    if (ctx.state !== 'running') ctx.resume().catch(() => {});
    const clean = out.stream.getAudioTracks()[0];
    const stopClean = clean.stop.bind(clean);
    const stop = () => {
      try { src.disconnect(); rnnoise.disconnect(); gate.disconnect(); rnnoise.destroy(); } catch (e) { /* already gone */ }
      ctx.close().catch(() => {});
      if (!opts.keepSource) track.stop();
    };
    // stopping the cleaned track (leaving the call) releases the microphone and the audio engine too
    clean.stop = () => { stopClean(); stop(); };
    track.addEventListener('ended', () => { try { stopClean(); } catch (e) { /* gone */ } stop(); });
    return { track: clean, stop };
  } catch (e) {
    if (ctx) ctx.close().catch(() => {});
    return null;
  }
}

// Puts the cleaned microphone into a stream in place of the raw one (video tracks are kept as they are).
export async function denoiseStream(stream) {
  if (!stream) return stream;
  const mic = stream.getAudioTracks()[0];
  const d = mic && await denoiseTrack(mic);
  if (!d) return stream;
  return new MediaStream([d.track, ...stream.getVideoTracks()]);
}

// A LiveKit track processor (large meetings and standups): the microphone LiveKit publishes is replaced by the cleaned one.
export function livekitDenoiser() {
  let d = null;
  const proc = {
    name: 'hub-rnnoise',
    processedTrack: undefined,
    async init({ track }) { d = await denoiseTrack(track, { keepSource: true }); proc.processedTrack = d ? d.track : undefined; },
    async restart(opts) { if (d) d.stop(); await proc.init(opts); },
    async destroy() { if (d) d.stop(); d = null; proc.processedTrack = undefined; }
  };
  return proc;
}
