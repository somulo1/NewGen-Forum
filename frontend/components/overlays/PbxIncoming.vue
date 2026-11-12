<!-- A customer is calling the phone system: the same ringing card Hub uses for every call, shown on any page, and it rings. -->
<template><template v-if="inc">
<div ref="card" role="alertdialog" aria-label="Incoming customer call" title="Drag to move" @pointerdown="dragStart" :style="(pos ? `left: ${pos.x}px; top: ${pos.y}px; right: auto; ` : 'right: 20px; top: 84px; ') + 'position: fixed; z-index: 90; touch-action: none; cursor: grab; user-select: none;'" style="width: 320px; max-width: calc(100vw - 28px); box-sizing: border-box; padding: 16px; border-radius: 18px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-B8F36B,#B8F36B); box-shadow: 0 16px 40px rgba(0,0,0,0.5); display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 12px; color: var(--k-B8F36B,#B8F36B); font-weight: 600">● Incoming customer call · {{ inc.queue }}</span><strong style="font-size: 17px">{{ who ? who.name : 'Unknown caller' }}</strong><span v-if="who" style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ who.account }} · {{ who.status }}</span></div>
<div style="display: flex; gap: 8px"><button @click="accept" style="flex: 1; height: 42px; border-radius: 21px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; cursor: pointer">Accept</button><button @click="decline(false)" style="flex: 1; height: 42px; border-radius: 21px; border: 0; background: #E5484D; color: #fff; font-weight: 700; cursor: pointer">Decline</button></div>
<button @click="decline(true)" style="height: 32px; border-radius: 16px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-C9D6CF,#C9D6CF); font-size: 12px; cursor: pointer">Send to voicemail</button>
</div>
</template></template>

<script>
import { pbx, answer, decline, customerOf } from '~/logic/pbx/store.js';
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'PbxIncoming',
  data() { return { pbx, pos: null }; },
  computed: {
    ...hubBindings(['go']).computed,
    inc() { return this.pbx.incoming; },
    who() { return this.inc && this.inc.customerId ? customerOf(this.inc.customerId) : null; }
  },
  watch: { inc(v) { if (!v) this.pos = null; } },
  methods: {
    decline,
    // answering takes you to the call, wherever you were
    accept() { answer(); if (this.go && this.go.pbx) this.go.pbx(); },
    dragStart(e) {
      if (e.target.closest('button')) return;
      const el = this.$refs.card, r = el.getBoundingClientRect(), dx = e.clientX - r.left, dy = e.clientY - r.top;
      const ui = window.__uiZ || 1;
      const move = (m) => { this.pos = { x: Math.max(0, Math.min(window.innerWidth - r.width, m.clientX - dx)) / ui, y: Math.max(0, Math.min(window.innerHeight - r.height, m.clientY - dy)) / ui }; };
      const up = () => { window.removeEventListener('pointermove', move); window.removeEventListener('pointerup', up); };
      window.addEventListener('pointermove', move); window.addEventListener('pointerup', up);
      e.preventDefault();
    }
  }
};
</script>
