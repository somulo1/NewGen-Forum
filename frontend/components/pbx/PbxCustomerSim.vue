<!-- Calls management: "Call in as a customer". Plays exactly what a caller hears (greeting, phone menu, queue position and the care
     tune) and then rings this agent's phone, so the whole journey can be checked before real customers hear it. A simulation. -->
<template>
<div class="hub-overlay pbx-sim-wrap" @click.self="hangUp">
  <div role="dialog" aria-modal="true" aria-label="Call in as a customer" class="pbx-sim">
    <div class="pbx-row between"><strong>Calling as a customer</strong><button type="button" class="pbx-x" aria-label="Hang up and close" @click="hangUp">×</button></div>
    <div class="pbx-sim-phone">
      <span class="pbx-sub">{{ who }}</span>
      <strong class="pbx-sim-num">0709 000 200</strong>
      <span class="pbx-sim-time">{{ timer }}</span>
      <span class="pbx-sim-stage" role="status" aria-live="polite">{{ stageText }}</span>
      <div v-if="stage === 'menu'" class="pbx-sim-opts"><span v-for="o in menu" :key="o.key"><strong>{{ o.key }}</strong> {{ o.to }}</span></div>
      <div class="pbx-keys small"><button v-for="k in keys" :key="k" type="button" :disabled="stage !== 'menu' && stage !== 'queue'" @click="press(k)"><strong>{{ k }}</strong></button></div>
      <button type="button" class="pbx-btn stop big" @click="hangUp">Hang up</button>
    </div>
    <label class="pbx-field">Call in as<select v-model="as" :disabled="stage !== 'idle'"><option value="">A new caller (unknown number)</option><option v-for="c in pbx.customers" :key="c.id" :value="c.id">{{ c.name }} · {{ c.product }}</option></select></label>
    <button v-if="stage === 'idle'" type="button" class="pbx-btn go" @click="start">Call our customer care</button>
    <span class="pbx-sub">You hear what the customer hears. When you are put through, this phone rings with the customer's call: answer it on the ringing card to finish the test.</span>
  </div>
</div>
</template>

<script>
import { pbx, simulateIncoming, customerOf } from '~/logic/pbx/store.js';
import { startTune, stopTune, tuneLevel, speak, stopSpeech, playClip, keyTone, startRingback, stopRingback, stopRing, GREETING } from '~/logic/pbx/sounds.js';

const wait = (ms) => new Promise((r) => setTimeout(r, ms));

export default {
  name: 'PbxCustomerSim',
  props: { greeting: { type: [Object, Boolean], default: false } },
  emits: ['close'],
  data() { return { pbx, as: '', stage: 'idle', started: 0, now: Date.now(), key: '', run: 0, keys: ['1', '2', '3', '4', '5', '6', '7', '8', '9', '*', '0', '#'], queue: '', position: 0, t: 0 }; },
  computed: {
    menu() { return this.pbx.routing.ivr.options; },
    who() { const c = this.as && customerOf(this.as); return c ? 'Calling from ' + c.name + '’s phone' : 'Calling from an unknown number'; },
    timer() { if (!this.started) return '0:00'; const s = Math.floor((this.now - this.started) / 1000); return Math.floor(s / 60) + ':' + String(s % 60).padStart(2, '0'); },
    stageText() {
      return { idle: 'Ready to call', ringing: 'Ringing…', greeting: 'Listening to the greeting', menu: 'Choose an option on the keypad', queue: 'Waiting in the ' + this.queue + ' queue · you are number ' + this.position, agent: 'Being put through to an agent…', answered: 'An agent answered. Test finished.', ended: 'Call ended' }[this.stage];
    }
  },
  watch: {
    // the agent answered the ringing card (or declined it): the customer's side ends here
    'pbx.call'(c) { if (c && this.stage === 'agent') { stopTune(); stopSpeech(); this.stage = 'answered'; } },
    'pbx.incoming'(v) { if (!v && this.stage === 'agent' && !this.pbx.call) { stopTune(); this.stage = 'ended'; speak('Sorry, all our agents are busy. Please leave a message after the tone.'); } }
  },
  mounted() { this.t = setInterval(() => { this.now = Date.now(); }, 500); },
  beforeUnmount() { clearInterval(this.t); this.stopAll(); },
  methods: {
    stopAll() { this.run++; stopTune(); stopSpeech(); stopRingback(); },
    alive(run) { return run === this.run; },
    async say(text, run) { if (this.alive(run)) await speak(text, { duck: true, voice: this.pbx.settings.voiceName }); },
    async start() {
      const run = ++this.run; this.started = Date.now(); this.stage = 'ringing';
      startRingback(); await wait(2600); stopRingback(); if (!this.alive(run)) return;
      // 1. the greeting over the care tune
      this.stage = 'greeting';
      startTune(this.pbx.settings.tune, false);
      const g = this.greeting;
      if (g && g.url) await playClip(g.url); else if (g) await this.say(g.text || GREETING, run);   // no greeting when it is switched off
      if (!this.alive(run)) return;
      // 2. the phone menu: read the options, wait for a key, repeat as set in Routing, then go to the first queue
      this.stage = 'menu'; this.key = '';
      for (let i = 0; i <= this.pbx.routing.ivr.retries && !this.key; i++) {
        await this.say(this.pbx.routing.ivr.greeting, run);
        for (let w = 0; w < this.pbx.routing.ivr.timeoutSec * 10 && !this.key && this.alive(run); w++) await wait(100);
      }
      if (!this.alive(run)) return;
      const opt = this.menu.find((o) => o.key === this.key) || this.menu.find((o) => o.key === '0') || this.menu[0];
      if (opt && opt.to === 'Repeat menu') { this.start(); return; }
      this.queue = opt ? opt.to : 'Customer care';
      // 3. the queue: place in line, the tune, and the call-back offer
      this.stage = 'queue'; this.position = Math.max(1, ((this.pbx.queues.find((q) => q.name === this.queue) || {}).waiting || 0) + 1);
      if (this.pbx.routing.announcePosition) await this.say('You are caller number ' + this.position + ' in the queue. Thank you for waiting.', run);
      if (this.pbx.routing.callbacks) await this.say('To keep your place and have us call you back, press 1 at any time.', run);
      this.key = '';
      for (let w = 0; w < 60 && this.alive(run); w++) {
        if (this.key === '1' && this.pbx.routing.callbacks) {
          await this.say('Thank you. We will call you back on this number shortly. Goodbye.', run);
          this.pbx.callbacks.unshift({ id: 'cb' + Date.now(), number: (customerOf(this.as) || {}).phone || '+254700000000', customerId: this.as || null, queue: this.queue, askedAt: Date.now(), status: 'Waiting' });
          this.stopAll(); this.stage = 'ended'; return;
        }
        await wait(100);
        if (w % 20 === 19 && this.position > 1) this.position--;
      }
      if (!this.alive(run)) return;
      // 4. put through: the tune keeps playing for the caller while this agent's phone rings
      this.stage = 'agent'; tuneLevel(0.35);
      simulateIncoming(this.queue, this.as || null);
      if (!this.pbx.incoming) { stopTune(); this.stage = 'ended'; await this.say('Sorry, the agent is not taking calls right now. Please leave a message after the tone.', ++this.run); }
    },
    press(k) { keyTone(k); this.key = k; },
    hangUp() { this.stopAll(); if (this.stage === 'agent' && this.pbx.incoming) { stopRing(); this.pbx.incoming = null; } this.$emit('close'); }
  }
};
</script>

<style>
.pbx-sim-wrap { position: fixed; inset: 0; z-index: 85; background: rgba(5,10,8,0.6); display: flex; align-items: center; justify-content: center; padding: 16px; }
.pbx-sim { width: 380px; max-width: 100%; max-height: 100%; overflow-y: auto; box-sizing: border-box; padding: 18px; border-radius: 20px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-2A3832,#2A3832); display: flex; flex-direction: column; gap: 12px; color: var(--k-E8EFEA,#E8EFEA); }
.pbx-sim-phone { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 16px; border-radius: 18px; background: var(--k-121B17,#121B17); text-align: center; }
.pbx-sim-num { font-family: 'Space Grotesk', sans-serif; font-size: 22px; }
.pbx-sim-time { font-variant-numeric: tabular-nums; color: var(--k-9DB0A6,#9DB0A6); }
.pbx-sim-stage { min-height: 20px; font-size: 13px; color: var(--k-B8F36B,#B8F36B); font-weight: 600; }
.pbx-sim-opts { display: flex; flex-wrap: wrap; justify-content: center; gap: 6px; font-size: 12px; }
.pbx-sim-opts span { padding: 3px 8px; border-radius: 10px; background: var(--k-1D2A25,#1D2A25); }
.pbx-sim .pbx-keys { width: 100%; max-width: 260px; }
.pbx-sim .pbx-keys button:disabled { opacity: 0.45; cursor: default; }
</style>
