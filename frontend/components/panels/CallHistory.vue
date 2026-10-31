<!-- Call history of the open chat (forum or private). Works in the docked panel and the expanded screens. -->
<template>
<div style="display: flex; flex-direction: column; gap: 8px; min-width: 0">
<div style="display: flex; align-items: center; gap: 8px"><strong style="flex: 1; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Calls with {{ callHistory.who }}</strong>
<template v-if="callHistory.rows.length > 1"><button v-if="!confirmAll" type="button" @click="confirmAll = true" style="height: 28px; padding: 0 12px; border: 1px solid var(--k-2A3832,#2A3832); border-radius: 14px; background: transparent; color: var(--k-9DB0A6,#9DB0A6); cursor: pointer; font-size: 12px">Clear all</button>
<template v-else><button type="button" :disabled="busy" @click="clearAllNow" style="height: 28px; padding: 0 12px; border: 0; border-radius: 14px; background: #FF9B8A; color: #0F1714; cursor: pointer; font-size: 12px; font-weight: 600">Delete all {{ callHistory.rows.length }} for good</button><button type="button" @click="confirmAll = false" style="height: 28px; padding: 0 10px; border: 0; background: transparent; color: var(--k-9DB0A6,#9DB0A6); cursor: pointer; font-size: 12px">Keep</button></template></template></div>
<span v-if="callHistory.rows.length" style="font-size: 11px; color: var(--k-6F8178,#6F8178); margin-top: -4px">Swipe a call to the left, then tap Delete, to remove it from your history for good.</span>
<span v-if="err" role="alert" style="font-size: 12px; color: #FF9B8A">{{ err }}</span>
<template v-if="callHistory.empty"><p style="margin: 8px 0; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">No calls yet.</p></template>
<template v-for="r in callHistory.rows" :key="r.id">
<div style="position: relative; overflow: hidden; border-radius: 10px">
<button type="button" :disabled="busy" @click="remove(r)" aria-label="Delete this call for good" style="position: absolute; top: 0; right: 0; bottom: 0; width: 88px; border: 0; border-radius: 0 10px 10px 0; background: #FF9B8A; color: #0F1714; font-weight: 700; font-size: 12px; cursor: pointer">Delete</button>
<div :style="`display: flex; align-items: center; gap: 10px; padding: 8px 10px; border-radius: 10px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); min-width: 0; touch-action: pan-y; user-select: none; position: relative; transform: translateX(${shift(r)}px); transition: ${drag && drag.id === r.id ? 'none' : 'transform .18s ease'}`" @pointerdown="down($event, r)" @pointermove="move($event, r)" @pointerup="up(r)" @pointercancel="up(r)" @pointerleave="up(r)">
<span aria-hidden="true" :style="`width: 30px; height: 30px; flex: none; border-radius: 50%; background: var(--k-1D2A25,#1D2A25); color: ${r.color}; display: flex; align-items: center; justify-content: center`"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 4h3.5l1.6 4-2 1.4a11 11 0 0 0 5.5 5.5l1.4-2 4 1.6V18a2 2 0 0 1-2 2A15 15 0 0 1 3 6a2 2 0 0 1 2-2z" stroke-width="1.6"/><template v-if="r.icon === 'in'"><path d="M21 3l-6 6M15 5v4h4"/></template><template v-else-if="r.icon === 'out'"><path d="M15 9l6-6M17 3h4v4"/></template><template v-else-if="r.icon === 'missedIn'"><path d="M21 3l-6 6M15 5v4h4"/><path d="M16.5 14.5l4.5 4.5M21 14.5L16.5 19" stroke-width="1.6"/></template><template v-else-if="r.icon === 'missedOut'"><path d="M15 9l6-6M17 3h4v4"/><path d="M16.5 14.5l4.5 4.5M21 14.5L16.5 19" stroke-width="1.6"/></template><template v-else-if="r.icon === 'declined'"><path d="M15 3l6 6M21 3l-6 6"/></template><template v-else><circle cx="18" cy="6" r="3.5"/><path d="M15.5 8.5l5-5"/></template></svg></span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><span :style="`font-size: 13px; font-weight: 600; color: ${r.color}; overflow: hidden; text-overflow: ellipsis; white-space: nowrap`">{{ r.title }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ r.sub }}</span></div>
</div>
</div>
</template>
</div>
</template>

<script>
import { hubBindings } from '~/composables/useHub';
const REVEAL = 88;
export default {
  name: 'CallHistory',
  ...hubBindings(['callHistory']),
  data() { return { drag: null, open: '', busy: false, confirmAll: false, err: '' }; },
  methods: {
    shift(r) { return this.drag && this.drag.id === r.id ? this.drag.dx : this.open === r.id ? -REVEAL : 0; },
    down(e, r) { this.drag = { id: r.id, x0: e.clientX, y0: e.clientY, base: this.open === r.id ? -REVEAL : 0, dx: this.open === r.id ? -REVEAL : 0, on: false }; },
    move(e, r) {
      const d = this.drag; if (!d || d.id !== r.id) return;
      const dx = e.clientX - d.x0, dy = e.clientY - d.y0;
      if (!d.on) { if (Math.abs(dy) > 10 && Math.abs(dy) > Math.abs(dx)) { this.drag = null; return; } if (Math.abs(dx) < 8) return; d.on = true; try { e.currentTarget.setPointerCapture(e.pointerId); } catch (x) { /* not capturable */ } }
      d.dx = Math.max(-REVEAL - 12, Math.min(0, d.base + dx));
    },
    up(r) {
      const d = this.drag; if (!d || d.id !== r.id) return;
      if (d.on) this.open = d.dx < -REVEAL / 2 ? r.id : '';
      this.drag = null;
    },
    async remove(r) { await this.run(() => r.remove()); this.open = ''; },
    async clearAllNow() { await this.run(() => this.callHistory.clearAll()); this.confirmAll = false; },
    async run(job) {
      if (this.busy) return;
      this.busy = true; this.err = '';
      try { await job(); } catch (e) { this.err = (e && e.message) || 'Could not delete it.'; }
      this.busy = false;
    }
  }
};
</script>
