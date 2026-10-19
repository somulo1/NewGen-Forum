<!-- Someone is calling. Only shown when your privacy settings let the call through. -->
<template><template v-if="incomingView">
<div ref="card" role="alertdialog" aria-label="Incoming call" title="Drag to move" @pointerdown="dragStart" :style="(pos ? `left: ${pos.x}px; top: ${pos.y}px; right: auto; ` : 'right: 20px; top: 84px; ') + 'position: fixed; z-index: 90; touch-action: none; cursor: grab; user-select: none;'" style=" width: 320px; max-width: calc(100vw - 28px); box-sizing: border-box; padding: 16px; border-radius: 18px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-B8F36B,#B8F36B); box-shadow: 0 16px 40px rgba(0,0,0,0.5); display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 12px; color: var(--k-B8F36B,#B8F36B); font-weight: 600">● Incoming {{ incomingView.kindLabel.toLowerCase() }}</span><strong style="font-size: 17px">{{ incomingView.name }}</strong></div>
<div style="display: flex; gap: 8px"><button @click="incomingView.accept" style="flex: 1; height: 42px; border-radius: 21px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; cursor: pointer">Accept</button><button @click="incomingView.decline" style="flex: 1; height: 42px; border-radius: 21px; border: 0; background: #E5484D; color: #fff; font-weight: 700; cursor: pointer">Decline</button></div>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'IncomingCall',
  data() { return { pos: null }; },
  watch: { incomingView(v) { if (!v) this.pos = null; } },
  methods: {
    // Drag the card by anywhere except its buttons; it stays inside the window.
    dragStart(e) {
      if (e.target.closest('button')) return;
      const el = this.$refs.card, r = el.getBoundingClientRect(), dx = e.clientX - r.left, dy = e.clientY - r.top;
      const move = (m) => { this.pos = { x: Math.max(0, Math.min(window.innerWidth - r.width, m.clientX - dx)), y: Math.max(0, Math.min(window.innerHeight - r.height, m.clientY - dy)) }; };
      const up = () => { window.removeEventListener('pointermove', move); window.removeEventListener('pointerup', up); };
      window.addEventListener('pointermove', move); window.addEventListener('pointerup', up);
      e.preventDefault();
    }
  },
  ...hubBindings(["incomingView"])
};
</script>
