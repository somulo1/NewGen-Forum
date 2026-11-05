<!-- A whiteboard and diagram canvas: shapes, sticky notes, text, connectors and pen. Works with mouse and touch. -->
<template>
<div style="display: flex; flex-direction: column; gap: 8px; min-width: 0">
<div role="toolbar" aria-label="Whiteboard tools" style="display: flex; flex-wrap: wrap; gap: 4px; align-items: center">
<button v-for="t in tools" :key="t.k" type="button" class="pp-btn" :aria-pressed="tool === t.k" :style="tool === t.k ? 'background: var(--k-22302A,#22302A)' : ''" @click="pick(t.k)">{{ t.l }}</button>
<label style="display: inline-flex; flex-direction: row; align-items: center; gap: 4px; width: auto; font-size: 12px">Colour<input type="color" v-model="color" aria-label="Colour" style="width: 32px; height: 28px; padding: 0; border: 0; background: transparent" @change="recolour"></label>
<button type="button" class="pp-btn" :disabled="!history.length" @click="undo">Undo</button>
<button type="button" class="pp-btn" :disabled="!sel" @click="remove">Delete</button>
</div>
<input v-if="selShape" type="text" :value="selShape.text" aria-label="Text in the selected shape" placeholder="Text in this shape" class="pp-input" @input="setText">
<span class="pp-mut">{{ hint }}</span>
<div style="overflow: auto; max-height: 70vh; border: 1px solid var(--k-2A3832,#2A3832); border-radius: 10px; background: #fff">
<svg ref="svg" width="1400" height="900" role="img" aria-label="Whiteboard canvas" style="display: block; touch-action: none; user-select: none" @pointerdown="down" @pointermove="move" @pointerup="up" @pointercancel="up">
<defs><marker id="be-a" markerWidth="10" markerHeight="8" refX="9" refY="4" orient="auto"><path d="M0 0L10 4L0 8z" fill="#555"/></marker></defs>
<polyline v-for="p in b.pen" :key="p.id" :data-id="p.id" fill="none" :stroke="p.color" :stroke-width="sel === p.id ? 4 : 2" :points="p.pts.map((q) => q.join(',')).join(' ')"/>
<line v-for="l in shownLines" :key="l.id" :data-id="l.id" :x1="l.a.x" :y1="l.a.y" :x2="l.z.x" :y2="l.z.y" :stroke="sel === l.id ? '#1F6FB5' : '#555'" stroke-width="2" marker-end="url(#be-a)"/>
<g v-for="s in b.shapes" :key="s.id" :data-id="s.id" style="cursor: move">
<ellipse v-if="s.t === 'ellipse'" :cx="s.x + s.w / 2" :cy="s.y + s.h / 2" :rx="s.w / 2" :ry="s.h / 2" :fill="s.color" :stroke="ring(s)" :stroke-width="sel === s.id ? 3 : 1"/>
<polygon v-else-if="s.t === 'diamond'" :points="`${s.x + s.w / 2},${s.y} ${s.x + s.w},${s.y + s.h / 2} ${s.x + s.w / 2},${s.y + s.h} ${s.x},${s.y + s.h / 2}`" :fill="s.color" :stroke="ring(s)" :stroke-width="sel === s.id ? 3 : 1"/>
<rect v-else :x="s.x" :y="s.y" :width="s.w" :height="s.h" rx="6" :fill="s.t === 'text' ? 'transparent' : s.color" :stroke="s.t === 'text' && sel !== s.id ? 'transparent' : ring(s)" :stroke-width="sel === s.id ? 3 : 1"/>
<text :x="s.x + s.w / 2" :y="s.y + s.h / 2" text-anchor="middle" dominant-baseline="middle" font-size="14" fill="#111" style="pointer-events: none">{{ s.text }}</text>
<rect v-if="sel === s.id" data-resize="1" :x="s.x + s.w - 6" :y="s.y + s.h - 6" width="12" height="12" fill="#1F6FB5" style="cursor: nwse-resize"/>
</g>
</svg>
</div>
</div>
</template>

<script>
import { parseBoard, centre } from '~/logic/boardSvg.js';

const uid = () => Math.random().toString(36).slice(2, 10);
const TOOLS = [{ k: 'select', l: 'Select' }, { k: 'sticky', l: 'Sticky note' }, { k: 'rect', l: 'Box' }, { k: 'ellipse', l: 'Circle' }, { k: 'diamond', l: 'Decision' }, { k: 'text', l: 'Text' }, { k: 'line', l: 'Connector' }, { k: 'pen', l: 'Pen' }];
const SIZE = { sticky: [140, 110], rect: [150, 70], ellipse: [130, 80], diamond: [130, 90], text: [150, 36] };

export default {
  name: 'BoardEditor',
  props: { modelValue: { type: String, default: '' } },
  emits: ['update:modelValue'],
  data() { return { b: parseBoard(this.modelValue), tool: 'select', color: '#FFE08A', sel: null, drag: null, from: null, history: [], tools: TOOLS }; },
  computed: {
    selShape() { return this.b.shapes.find((s) => s.id === this.sel) || null; },
    shownLines() { const m = new Map(this.b.shapes.map((s) => [s.id, s])); return this.b.lines.filter((l) => m.has(l.from) && m.has(l.to)).map((l) => ({ ...l, a: centre(m.get(l.from)), z: centre(m.get(l.to)) })); },
    hint() { if (this.tool === 'line') return this.from ? 'Now click the shape it goes to.' : 'Click the shape it starts from.'; if (this.tool === 'pen') return 'Draw on the board.'; if (this.tool === 'select') return 'Drag to move. Drag the blue corner to resize.'; return 'Click the board to place it.'; }
  },
  methods: {
    pick(k) { this.tool = k; this.from = null; },
    ring(s) { return this.sel === s.id ? '#1F6FB5' : '#555'; },
    pos(e) { const r = this.$refs.svg.getBoundingClientRect(); return [Math.round(e.clientX - r.left), Math.round(e.clientY - r.top)]; },
    snap() { this.history.push(JSON.stringify(this.b)); if (this.history.length > 50) this.history.shift(); },
    save() { this.$emit('update:modelValue', JSON.stringify(this.b)); },
    undo() { this.b = JSON.parse(this.history.pop()); this.sel = null; this.save(); },
    remove() {
      this.snap();
      const id = this.sel;
      this.b.shapes = this.b.shapes.filter((s) => s.id !== id);
      this.b.pen = this.b.pen.filter((s) => s.id !== id);
      this.b.lines = this.b.lines.filter((l) => l.id !== id && l.from !== id && l.to !== id);
      this.sel = null; this.save();
    },
    recolour() { if (this.selShape) { this.snap(); this.selShape.color = this.color; this.save(); } },
    setText(e) { this.selShape.text = e.target.value; this.save(); },
    down(e) {
      const [x, y] = this.pos(e);
      const g = e.target.closest('[data-id]');
      const id = g ? g.getAttribute('data-id') : null;
      this.$refs.svg.setPointerCapture(e.pointerId);
      if (this.tool === 'pen') { this.snap(); const p = { id: uid(), color: this.color === '#FFE08A' ? '#333333' : this.color, pts: [[x, y]] }; this.b.pen.push(p); this.drag = { pen: p }; return; }
      if (this.tool === 'line') {
        if (!id || !this.b.shapes.some((s) => s.id === id)) return;
        if (!this.from) { this.from = id; this.sel = id; return; }
        if (id !== this.from) { this.snap(); this.b.lines.push({ id: uid(), from: this.from, to: id }); this.save(); }
        this.from = null; return;
      }
      if (SIZE[this.tool]) {
        this.snap();
        const [w, h] = SIZE[this.tool];
        const s = { id: uid(), t: this.tool, x: x - w / 2, y: y - h / 2, w, h, text: '', color: this.tool === 'sticky' ? this.color : this.color === '#FFE08A' ? '#DCEBFA' : this.color };
        this.b.shapes.push(s); this.sel = s.id; this.tool = 'select'; this.save(); return;
      }
      this.sel = id;
      const s = this.selShape;
      if (s) { this.snap(); this.drag = { s, x, y, sx: s.x, sy: s.y, sw: s.w, sh: s.h, resize: !!e.target.getAttribute('data-resize') }; }
    },
    move(e) {
      if (!this.drag) return;
      const [x, y] = this.pos(e);
      const d = this.drag;
      if (d.pen) { d.pen.pts.push([x, y]); return; }
      if (d.resize) { d.s.w = Math.max(30, d.sw + x - d.x); d.s.h = Math.max(24, d.sh + y - d.y); } else { d.s.x = d.sx + x - d.x; d.s.y = d.sy + y - d.y; }
    },
    up() { if (this.drag) { this.drag = null; this.save(); } }
  }
};
</script>
