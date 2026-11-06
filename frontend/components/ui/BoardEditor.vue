<!-- Diagrams and whiteboards: pick shapes from the library, drag between their connection points, bend connectors by their handles. -->
<template>
<div ref="anchor" style="min-width: 0">
<Teleport to="body" :disabled="!full">
<div ref="root" :class="['bd', { 'bd-full': full }, full && lightMode ? 'hub-light' : '']" tabindex="0" @keydown="onKey" @keyup="onKeyUp">
<div class="bd-top" role="toolbar" aria-label="Diagram tools">
<button v-for="t in tools" :key="t.k" type="button" class="bd-btn" :aria-pressed="tool === t.k" :class="{ on: tool === t.k }" @click="tool = t.k">{{ t.l }}</button>
<span class="bd-sep"></span>
<button type="button" class="bd-btn" :disabled="!past.length" @click="undo">Undo</button><button type="button" class="bd-btn" :disabled="!future.length" @click="redo">Redo</button>
<button type="button" class="bd-btn" :disabled="!selShapes.length" @click="duplicate">Duplicate</button><button type="button" class="bd-btn" :disabled="!hasSel" @click="removeSel">Delete</button>
<button type="button" class="bd-btn" :disabled="!hasSel" title="Ctrl+L" @click="toggleLock">{{ allLocked ? 'Unlock' : 'Lock' }}</button>
<button type="button" class="bd-btn" :disabled="!selShapes.length" @click="order(1)">To front</button><button type="button" class="bd-btn" :disabled="!selShapes.length" @click="order(-1)">To back</button>
<span class="bd-sep"></span>
<label class="bd-check">Background<select class="bd-sel" :value="b.bg" @change="setBg($event.target.value)"><option v-for="o in backgrounds" :key="o.v" :value="o.v">{{ o.l }}</option></select></label>
<label class="bd-check"><input type="checkbox" v-model="grid">Snap</label>
<span class="bd-zoom" role="group" aria-label="Zoom"><button type="button" class="bd-btn" title="Zoom out (Ctrl and -)" aria-label="Zoom out" @click="setZoom(zoom - 0.1)">−</button><input type="range" min="20" max="300" step="10" aria-label="Zoom level" :value="Math.round(zoom * 100)" @input="setZoom($event.target.value / 100)"><button type="button" class="bd-btn" title="Zoom in (Ctrl and +)" aria-label="Zoom in" @click="setZoom(zoom + 0.1)">+</button><button type="button" class="bd-btn" title="Back to 100% (Ctrl and 0)" @click="setZoom(1)">{{ Math.round(zoom * 100) }}%</button><button type="button" class="bd-btn" title="Show the whole drawing" @click="fit">Fit</button></span>
<button type="button" class="bd-btn" @click="exportSvg">SVG</button><button type="button" class="bd-btn" @click="exportPng">PNG</button>
<button type="button" class="bd-btn" @click="libOpen = !libOpen">{{ libOpen ? 'Hide shapes' : 'Show shapes' }}</button>
<button type="button" :class="['bd-btn', { 'bd-focus': !full }]" :title="full ? 'Back to the page (Esc)' : 'Draw on the whole screen'" @click="toggleFull">{{ full ? 'Exit focus' : 'Focus mode' }}</button>
</div>

<div v-if="hasSel && allLocked" class="bd-props bd-mut">Locked. Press Unlock to change it.</div>
<div v-else-if="selShapes.length" class="bd-props" aria-label="Shape settings">
<label v-if="selShapes.length === 1">Text<input type="text" :value="selShapes[0].text" @input="setShape('text', $event.target.value)"></label>
<label>Fill<input type="color" :value="selShapes[0].color || '#DCEBFA'" @input="setShape('color', $event.target.value)" @change="commit"></label>
<label>Line<input type="color" :value="selShapes[0].stroke || '#555555'" @input="setShape('stroke', $event.target.value)" @change="commit"></label>
<label>Text colour<input type="color" :value="selShapes[0].textColor || '#111111'" @input="setShape('textColor', $event.target.value)" @change="commit"></label>
<label>Size<input type="number" min="9" max="40" :value="selShapes[0].size || 14" @input="setShape('size', Number($event.target.value))"></label>
<label>Shape<select :value="selShapes[0].t" @change="setShape('t', $event.target.value); commit()"><optgroup v-for="g in groups" :key="g.group" :label="g.group"><option v-for="x in g.shapes" :key="x.key" :value="x.key">{{ x.label }}</option></optgroup></select></label>
</div>
<div v-else-if="selLine" class="bd-props" aria-label="Connector settings">
<label>Label<input type="text" :value="selLine.label || ''" @input="setLine('label', $event.target.value)"></label>
<label>Route<select :value="selLine.kind || 'straight'" @change="setLine('kind', $event.target.value); commit()"><option value="straight">Straight</option><option value="elbow">Right angles</option><option value="curve">Curved</option></select></label>
<label>Pattern<select :value="selLine.dash || 'solid'" @change="setLine('dash', $event.target.value); commit()"><option value="solid">Solid</option><option value="dashed">Dashed</option><option value="dotted">Dotted</option></select></label>
<label>Start<select :value="selLine.start || 'none'" @change="setLine('start', $event.target.value); commit()"><option v-for="e in endTypes" :key="e.v" :value="e.v">{{ e.l }}</option></select></label>
<label>End<select :value="selLine.end === undefined ? 'arrow' : selLine.end" @change="setLine('end', $event.target.value); commit()"><option v-for="e in endTypes" :key="e.v" :value="e.v">{{ e.l }}</option></select></label>
<label>Thickness<input type="number" min="1" max="8" :value="selLine.width || 2" @input="setLine('width', Number($event.target.value))"></label>
<label>Colour<input type="color" :value="selLine.color || '#555555'" @input="setLine('color', $event.target.value)" @change="commit"></label>
<button v-if="(selLine.pts || []).length" type="button" class="bd-btn" @click="setLine('pts', []); commit()">Remove bends</button>
<button type="button" class="bd-btn" @click="reverse">Swap ends</button>
</div>
<div v-else class="bd-props bd-mut">{{ hint }}</div>

<div class="bd-body">
<div v-if="libOpen" class="bd-lib" aria-label="Shape library">
<label class="bd-find"><span class="bd-sr">Find a shape</span><input type="search" v-model="find" placeholder="Find a shape"></label>
<span v-if="!shownGroups.length" class="bd-mut" style="padding: 6px">No shape has that name.</span>
<details v-for="g in shownGroups" :key="g.group" open>
<summary>{{ g.group }}</summary>
<div class="bd-lib-grid">
<button v-for="x in g.shapes" :key="x.key" type="button" :class="{ on: placing === x.key }" :title="x.label" :aria-label="x.label" draggable="true" @dragstart="dragNew = x.key" @click="placing = placing === x.key ? null : x.key">
<svg :viewBox="`-10 -10 ${x.w + 20} ${x.h + (x.textBelow ? 30 : 20)}`" width="38" height="28" aria-hidden="true"><path :d="x.d(x.w, x.h)" class="bd-ico" :fill="x.noBox || x.invisibleBox ? 'none' : undefined" :stroke="x.invisibleBox || x.noBorder ? 'none' : undefined" :stroke-dasharray="x.noBox ? '10 8' : undefined" stroke-width="5"/><path v-if="x.more" :d="x.more(x.w, x.h)" fill="none" stroke="currentColor" stroke-width="5"/></svg>
<span>{{ x.label }}</span>
</button>
</div>
</details>
</div>

<div ref="scroller" class="bd-canvas" :style="`background: ${paperColor}`" @dragover.prevent @drop.prevent="onDrop" @wheel="onWheel">
<div :style="`position: relative; width: ${W * zoom}px; height: ${H * zoom}px`">
<svg ref="svg" :width="W * zoom" :height="H * zoom" :viewBox="`0 0 ${W} ${H}`" role="img" aria-label="Diagram canvas" :style="`display: block; touch-action: none; user-select: none; cursor: ${cursor}`" @pointerdown="down" @pointermove="move" @pointerup="up" @pointercancel="up" @dblclick="dbl">
<defs v-html="pattern"></defs>
<rect v-if="pattern" :width="W" :height="H" fill="url(#bd-bg)" pointer-events="none"/>
<polyline v-for="p in b.pen" :key="p.id" :data-pen="p.id" fill="none" :stroke="p.color" :stroke-width="sel.has(p.id) ? 4 : 2" :points="p.pts.map((q) => q.join(',')).join(' ')"/>
<g v-for="s in drawn" :key="s.id" :data-shape="s.id" :transform="`translate(${s.x} ${s.y})`" style="cursor: move">
<path :d="s.g.d" :fill="s.g.fill" :stroke="s.g.stroke" stroke-width="1.5"/>
<path v-if="s.g.more" :d="s.g.more" fill="none" :stroke="s.g.moreStroke" stroke-width="1.5" pointer-events="none"/>
<text v-if="s.locked" :x="s.w - 8" y="12" font-size="11" text-anchor="middle" pointer-events="none">🔒</text>
<text v-for="(t, ti) in s.g.lines" :key="ti" :x="t.x" :y="t.y" text-anchor="middle" dominant-baseline="middle" :font-size="s.g.size" :fill="s.g.textColor" pointer-events="none">{{ t.t }}</text>
</g>
<g v-for="l in routed" :key="l.id">
<path :d="l.r.d" fill="none" stroke="transparent" stroke-width="14" :data-line="l.id" style="cursor: pointer"/>
<path :d="l.r.d" fill="none" :stroke="sel.has(l.id) ? '#1F6FB5' : l.r.color" :stroke-width="l.r.width" :stroke-dasharray="l.r.dash" pointer-events="none"/>
<path v-for="(g, gi) in l.r.glyphs" :key="gi" :d="g.d" :stroke="sel.has(l.id) ? '#1F6FB5' : l.r.color" stroke-width="1.5" :fill="g.fill === 'line' ? (sel.has(l.id) ? '#1F6FB5' : l.r.color) : g.fill === 'white' ? paperColor : 'none'" pointer-events="none"/>
<text v-if="l.r.label" :x="l.r.mid.x" :y="l.r.mid.y" text-anchor="middle" dominant-baseline="middle" font-size="12" :fill="b.bg === 'black' ? '#E8EFEA' : '#111'" :stroke="paperColor" stroke-width="4" paint-order="stroke" pointer-events="none">{{ l.r.label }}</text>
</g>
<!-- handles of the selected connector -->
<template v-if="selLine && selRoute && !selLine.locked">
<circle v-for="h in selRoute.adds" :key="'add' + h.at" :cx="h.x" :cy="h.y" r="5" fill="#fff" stroke="#1F6FB5" stroke-width="2" :data-add="h.at" style="cursor: grab"><title>Drag to bend</title></circle>
<circle v-for="(p, pi) in selRoute.pts" :key="'way' + pi" :cx="p[0]" :cy="p[1]" r="6" fill="#1F6FB5" :data-way="pi" style="cursor: grab"><title>Drag to move. Double-click to remove.</title></circle>
<rect :x="selRoute.a.x - 6" :y="selRoute.a.y - 6" width="12" height="12" fill="#3C8D2F" data-end="from" style="cursor: crosshair"/>
<rect :x="selRoute.z.x - 6" :y="selRoute.z.y - 6" width="12" height="12" fill="#C0392B" data-end="to" style="cursor: crosshair"/>
</template>
<!-- resize handles -->
<g v-if="selShapes.length === 1 && !drag && !selShapes[0].locked">
<rect :x="selShapes[0].x" :y="selShapes[0].y" :width="selShapes[0].w" :height="selShapes[0].h" fill="none" stroke="#1F6FB5" stroke-dasharray="4 3" pointer-events="none"/>
<rect v-for="h in resizeHandles" :key="h.k" :x="h.x - 5" :y="h.y - 5" width="10" height="10" fill="#fff" stroke="#1F6FB5" stroke-width="2" :data-resize="h.k" :style="`cursor: ${h.c}`"/>
</g>
<g v-else>
<rect v-for="s in selShapes" :key="s.id" :x="s.x - 3" :y="s.y - 3" :width="s.w + 6" :height="s.h + 6" fill="none" stroke="#1F6FB5" stroke-dasharray="4 3" pointer-events="none"/>
</g>
<!-- connection points -->
<template v-for="s in portShapes" :key="'p' + s.id"><circle v-for="p in s.ports" :key="p.i" :cx="p.x" :cy="p.y" :r="snap && snap.shape === s.id && snap.port === p.i ? 8 : 5" :fill="snap && snap.shape === s.id && snap.port === p.i ? '#3C8D2F' : '#fff'" stroke="#3C8D2F" stroke-width="2" :data-port="s.id + ':' + p.i" style="cursor: crosshair"><title>Drag to connect</title></circle></template>
<path v-if="drag && drag.ghost" :d="drag.ghost" fill="none" stroke="#1F6FB5" stroke-width="2" stroke-dasharray="5 4" pointer-events="none"/>
<rect v-if="drag && drag.type === 'marquee'" :x="Math.min(drag.x0, drag.x)" :y="Math.min(drag.y0, drag.y)" :width="Math.abs(drag.x - drag.x0)" :height="Math.abs(drag.y - drag.y0)" fill="#1F6FB522" stroke="#1F6FB5" pointer-events="none"/>
</svg>
<textarea v-if="editing" ref="inline" :value="editing.text" aria-label="Text in this shape" :style="`position: absolute; left: ${editing.x * zoom}px; top: ${editing.y * zoom}px; width: ${editing.w * zoom}px; height: ${editing.h * zoom}px; box-sizing: border-box; padding: 4px; border: 2px solid #1F6FB5; background: #fffffff0; color: #111; text-align: center; font: inherit; font-size: 14px; resize: none`" @input="setShape('text', $event.target.value)" @blur="editing = null; commit()" @keydown.stop="$event.key === 'Escape' && $event.target.blur()" @pointerdown.stop></textarea>
</div>
</div>
</div>
</div>
</Teleport>
</div>
</template>

<script>
import { BOARD_BACKGROUNDS, SHAPES, SHAPE_GROUPS, boardBox, boardInk, boardPaper, boardPattern, boardSvg, drawShape, parseBoard, portsOf, routeLine } from '~/logic/boardSvg.js';

const uid = () => Math.random().toString(36).slice(2, 10);
const END_TYPES = [{ v: 'none', l: 'None' }, { v: 'arrow', l: 'Arrow' }, { v: 'open', l: 'Open arrow' }, { v: 'triangle', l: 'Hollow triangle' }, { v: 'diamond', l: 'Filled diamond' }, { v: 'odiamond', l: 'Hollow diamond' }, { v: 'circle', l: 'Circle' }, { v: 'one', l: 'One (crow’s foot)' }, { v: 'many', l: 'Many (crow’s foot)' }, { v: 'zeromany', l: 'Zero or many' }];
const HANDLES = [['nw', 0, 0, 'nwse-resize'], ['n', 0.5, 0, 'ns-resize'], ['ne', 1, 0, 'nesw-resize'], ['e', 1, 0.5, 'ew-resize'], ['se', 1, 1, 'nwse-resize'], ['s', 0.5, 1, 'ns-resize'], ['sw', 0, 1, 'nesw-resize'], ['w', 0, 0.5, 'ew-resize']];

export default {
  name: 'BoardEditor',
  props: { modelValue: { type: String, default: '' } },
  emits: ['update:modelValue'],
  data() {
    return { b: parseBoard(this.modelValue), libOpen: true, find: '', lightMode: false, panning: false, zoom: 1, grid: true, full: false, tool: 'select', placing: null, dragNew: null, sel: new Set(), hover: null, snap: null, drag: null, editing: null, past: [], future: [], clip: null, backgrounds: BOARD_BACKGROUNDS, groups: SHAPE_GROUPS, endTypes: END_TYPES, tools: [{ k: 'select', l: 'Select' }, { k: 'line', l: 'Line' }, { k: 'pen', l: 'Pen' }, { k: 'hand', l: 'Move view' }], penColor: '#333333' };
  },
  computed: {
    box() { return boardBox(this.b); },
    // the sheet grows with the drawing, so there is always room to keep going
    W() { return Math.max(2400, this.box.x + this.box.w + 800); },
    H() { return Math.max(1600, this.box.y + this.box.h + 600); },
    shownGroups() { const q = this.find.trim().toLowerCase(); return q ? this.groups.map((g) => ({ ...g, shapes: g.shapes.filter((x) => x.label.toLowerCase().includes(q)) })).filter((g) => g.shapes.length) : this.groups; },
    pattern() { return boardPattern(this.b.bg, 'bd-bg'); },
    paperColor() { return boardPaper(this.b.bg); },
    byId() { return new Map(this.b.shapes.map((s) => [s.id, s])); },
    hasSel() { return this.sel.size > 0; },
    selected() { return [...this.b.shapes, ...this.b.lines, ...this.b.pen].filter((x) => this.sel.has(x.id)); },
    allLocked() { const l = this.selected; return l.length > 0 && l.every((x) => x.locked); },
    selShapes() { return this.b.shapes.filter((s) => this.sel.has(s.id)); },
    selLine() { return this.sel.size === 1 ? this.b.lines.find((l) => this.sel.has(l.id)) || null : null; },
    selRoute() { return this.selLine ? routeLine(this.selLine, this.byId, boardInk(this.b.bg)) : null; },
    drawn() { return this.b.shapes.map((s) => ({ ...s, g: drawShape(s, this.b.bg) })); },
    routed() { return this.b.lines.filter((l) => (!l.from || this.byId.has(l.from)) && (!l.to || this.byId.has(l.to))).map((l) => ({ id: l.id, r: routeLine(l, this.byId, boardInk(this.b.bg)) })); },
    resizeHandles() { const s = this.selShapes[0]; return HANDLES.map(([k, fx, fy, c]) => ({ k, c, x: s.x + fx * s.w, y: s.y + fy * s.h })); },
    linking() { return this.tool === 'line' || (this.drag && (this.drag.type === 'newline' || this.drag.type === 'end')); },
    portShapes() {
      const l = this.linking ? this.b.shapes : this.b.shapes.filter((s) => s.id === this.hover || (this.sel.size === 1 && this.sel.has(s.id)));
      return this.drag && this.drag.type !== 'newline' && this.drag.type !== 'end' ? [] : l.map((s) => ({ id: s.id, ports: portsOf(s) }));
    },
    cursor() { if (this.panning || this.tool === 'hand') return this.drag && this.drag.type === 'pan' ? 'grabbing' : 'grab'; return this.placing ? 'copy' : this.tool === 'pen' ? 'crosshair' : this.tool === 'line' ? 'crosshair' : 'default'; },
    hint() { if (this.placing) return 'Click the canvas to place ' + SHAPES[this.placing].label.toLowerCase() + '.'; if (this.tool === 'line') return 'Drag from one shape’s point to another’s, or draw a free line.'; if (this.tool === 'pen') return 'Draw freely.'; return 'Pick a shape, or drag one in. Hover a shape and drag from a green point to connect. Double-click to write.'; }
  },
  created() { this.touches = new Map(); this.pinch = null; },
  beforeUnmount() { if (this.full) document.body.style.overflow = ''; },
  methods: {
    pos(e) { const r = this.$refs.svg.getBoundingClientRect(); return [(e.clientX - r.left) / this.zoom, (e.clientY - r.top) / this.zoom]; },
    g(v) { return this.grid ? Math.round(v / 10) * 10 : Math.round(v); },
    snapshot() { this.past.push(JSON.stringify(this.b)); if (this.past.length > 100) this.past.shift(); this.future = []; },
    commit() { this.$emit('update:modelValue', JSON.stringify(this.b)); },
    undo() { this.future.push(JSON.stringify(this.b)); this.b = JSON.parse(this.past.pop()); this.sel = new Set(); this.commit(); },
    redo() { this.past.push(JSON.stringify(this.b)); this.b = JSON.parse(this.future.pop()); this.sel = new Set(); this.commit(); },
    // keeps the point under (cx, cy) in the same place on screen
    setZoom(z, cx, cy) {
      const el = this.$refs.scroller, old = this.zoom;
      const next = Math.min(3, Math.max(0.2, Math.round(z * 100) / 100));
      const px = cx === undefined ? el.clientWidth / 2 : cx, py = cy === undefined ? el.clientHeight / 2 : cy;
      const wx = (el.scrollLeft + px) / old, wy = (el.scrollTop + py) / old;
      this.zoom = next;
      this.$nextTick(() => { el.scrollLeft = wx * next - px; el.scrollTop = wy * next - py; });
    },
    fit() {
      const el = this.$refs.scroller, b = this.box;
      const z = Math.min(1.5, Math.max(0.2, Math.min(el.clientWidth / b.w, el.clientHeight / b.h)));
      this.zoom = Math.floor(z * 100) / 100;
      this.$nextTick(() => { el.scrollLeft = b.x * this.zoom; el.scrollTop = b.y * this.zoom; });
    },
    onWheel(e) {
      if (!e.ctrlKey && !e.metaKey) return;
      e.preventDefault();
      const r = this.$refs.scroller.getBoundingClientRect();
      this.setZoom(this.zoom * (e.deltaY < 0 ? 1.1 : 0.9), e.clientX - r.left, e.clientY - r.top);
    },
    toggleFull() {
      if (!this.full) this.lightMode = !!(this.$refs.anchor.closest('.hub-light') || document.querySelector('.hub-light'));
      this.full = !this.full;
      document.body.style.overflow = this.full ? 'hidden' : '';
      this.$nextTick(() => this.$refs.root.focus({ preventScroll: true }));
    },
    setShape(k, v) { if (!this.pendingEdit) { this.snapshot(); this.pendingEdit = true; setTimeout(() => { this.pendingEdit = false; }, 800); } this.selShapes.filter((s) => !s.locked).forEach((s) => { s[k] = v; }); if (this.editing && k === 'text') this.editing.text = v; this.commit(); },
    setLine(k, v) { if (!this.pendingEdit) { this.snapshot(); this.pendingEdit = true; setTimeout(() => { this.pendingEdit = false; }, 800); } this.selLine[k] = v; this.commit(); },
    reverse() { this.snapshot(); const l = this.selLine; [l.from, l.to] = [l.to, l.from]; [l.fromPort, l.toPort] = [l.toPort, l.fromPort]; [l.x1, l.x2] = [l.x2, l.x1]; [l.y1, l.y2] = [l.y2, l.y1]; l.pts = (l.pts || []).slice().reverse(); this.commit(); },
    addShape(key, x, y) {
      this.snapshot();
      const sp = SHAPES[key];
      const s = { id: uid(), t: key, x: this.g(x - sp.w / 2), y: this.g(y - sp.h / 2), w: sp.w, h: sp.h, text: '', ...(sp.fill ? { color: sp.fill } : {}) };
      this.b.shapes.push(s);
      this.sel = new Set([s.id]);
      this.commit();
    },
    onDrop(e) { if (!this.dragNew) return; const [x, y] = this.pos(e); this.addShape(this.dragNew, x, y); this.dragNew = null; },
    removeSel() {
      this.snapshot();
      const on = new Set(this.selected.filter((x) => !x.locked).map((x) => x.id));
      this.b.shapes = this.b.shapes.filter((s) => !on.has(s.id));
      this.b.pen = this.b.pen.filter((p) => !on.has(p.id));
      this.b.lines = this.b.lines.filter((l) => l.locked || (!on.has(l.id) && !on.has(l.from) && !on.has(l.to)));
      this.sel = new Set(); this.commit();
    },
    setBg(v) { this.snapshot(); this.b.bg = v; this.commit(); },
    toggleLock() { this.snapshot(); const v = !this.allLocked; this.selected.forEach((x) => { x.locked = v; }); this.commit(); },
    copyOut() { const on = this.sel; this.clip = JSON.stringify({ shapes: this.b.shapes.filter((s) => on.has(s.id)), lines: this.b.lines.filter((l) => on.has(l.from) && on.has(l.to)) }); },
    paste() {
      if (!this.clip) return;
      this.snapshot();
      const c = JSON.parse(this.clip), map = new Map();
      c.shapes.forEach((s) => { const id = uid(); map.set(s.id, id); this.b.shapes.push({ ...s, locked: false, id, x: s.x + 30, y: s.y + 30 }); });
      c.lines.forEach((l) => this.b.lines.push({ ...l, id: uid(), from: map.get(l.from), to: map.get(l.to), pts: (l.pts || []).map(([x, y]) => [x + 30, y + 30]) }));
      this.sel = new Set(map.values()); this.clip = JSON.stringify({ shapes: this.selShapes, lines: [] }); this.commit();
    },
    duplicate() { this.copyOut(); this.paste(); },
    order(dir) { this.snapshot(); const on = this.b.shapes.filter((s) => this.sel.has(s.id)), rest = this.b.shapes.filter((s) => !this.sel.has(s.id)); this.b.shapes = dir > 0 ? rest.concat(on) : on.concat(rest); this.commit(); },
    shapeAt(x, y, pad = 8) { for (let i = this.b.shapes.length - 1; i >= 0; i--) { const s = this.b.shapes[i]; if (x >= s.x - pad && x <= s.x + s.w + pad && y >= s.y - pad && y <= s.y + s.h + pad) return s; } return null; },
    nearPort(x, y, skip) {
      let best = null;
      this.b.shapes.forEach((s) => { if (s.id === skip) return; portsOf(s).forEach((p) => { const d = Math.hypot(p.x - x, p.y - y); if (d < 22 && (!best || d < best.d)) best = { shape: s.id, port: p.i, d, x: p.x, y: p.y }; }); });
      if (best) return best;
      const s = this.shapeAt(x, y, 0);
      return s && s.id !== skip ? { shape: s.id, port: null, x, y } : null;
    },
    down(e) {
      this.$refs.root.focus({ preventScroll: true });
      const [x, y] = this.pos(e), t = e.target, ds = t.dataset || {};
      this.$refs.svg.setPointerCapture(e.pointerId);
      // two fingers: pinch to zoom
      this.touches.set(e.pointerId, [e.clientX, e.clientY]);
      if (this.touches.size === 2) { const [a, b] = [...this.touches.values()]; this.pinch = { d: Math.hypot(a[0] - b[0], a[1] - b[1]), z: this.zoom }; this.drag = null; return; }
      if (this.tool === 'hand' || this.panning || e.button === 1) { const el = this.$refs.scroller; this.drag = { type: 'pan', cx: e.clientX, cy: e.clientY, sl: el.scrollLeft, st: el.scrollTop }; return; }
      if (this.placing) { this.addShape(this.placing, x, y); this.placing = null; return; }
      if (this.tool === 'pen') { this.snapshot(); const p = { id: uid(), color: this.penColor, pts: [[Math.round(x), Math.round(y)]] }; this.b.pen.push(p); this.drag = { type: 'pen', p }; return; }
      if (ds.port) { const [sid, i] = ds.port.split(':'); this.drag = { type: 'newline', from: sid, fromPort: Number(i), x, y }; return; }
      if (this.tool === 'line') { this.drag = { type: 'newline', from: null, x1: x, y1: y, x, y }; return; }
      if (ds.resize) { const s = this.selShapes[0]; this.snapshot(); this.drag = { type: 'resize', k: ds.resize, s, x0: x, y0: y, o: { x: s.x, y: s.y, w: s.w, h: s.h } }; return; }
      if (ds.way !== undefined) { this.snapshot(); this.drag = { type: 'way', i: Number(ds.way) }; return; }
      if (ds.add !== undefined) { this.snapshot(); const pts = (this.selLine.pts || []).slice(); pts.splice(Number(ds.add), 0, [Math.round(x), Math.round(y)]); this.selLine.pts = pts; this.drag = { type: 'way', i: Number(ds.add) }; return; }
      if (ds.end) { this.snapshot(); this.drag = { type: 'end', which: ds.end, x, y }; return; }
      if (ds.line) { this.sel = new Set([ds.line]); return; }
      const g = t.closest && t.closest('[data-shape]');
      const id = g ? g.getAttribute('data-shape') : t.dataset && t.dataset.pen;
      if (id) {
        if (e.shiftKey) { const n = new Set(this.sel); if (n.has(id)) n.delete(id); else n.add(id); this.sel = n; return; }
        if (!this.sel.has(id)) this.sel = new Set([id]);
        const list = this.selShapes.filter((s) => !s.locked).map((s) => ({ s, x: s.x, y: s.y }));
        if (list.length) { this.drag = { type: 'move', list, x0: x, y0: y, moved: false }; }
        return;
      }
      if (!e.shiftKey) this.sel = new Set();
      this.drag = { type: 'marquee', x0: x, y0: y, x, y };
    },
    move(e) {
      if (this.touches.has(e.pointerId)) this.touches.set(e.pointerId, [e.clientX, e.clientY]);
      if (this.pinch && this.touches.size === 2) { const [a, b] = [...this.touches.values()]; const r = this.$refs.scroller.getBoundingClientRect(); this.setZoom(this.pinch.z * (Math.hypot(a[0] - b[0], a[1] - b[1]) / this.pinch.d), (a[0] + b[0]) / 2 - r.left, (a[1] + b[1]) / 2 - r.top); return; }
      if (this.drag && this.drag.type === 'pan') { const el = this.$refs.scroller; el.scrollLeft = this.drag.sl - (e.clientX - this.drag.cx); el.scrollTop = this.drag.st - (e.clientY - this.drag.cy); return; }
      const [x, y] = this.pos(e), d = this.drag;
      if (!d) { const s = this.shapeAt(x, y); this.hover = s ? s.id : null; return; }
      if (d.type === 'pen') { d.p.pts.push([Math.round(x), Math.round(y)]); return; }
      if (d.type === 'move') {
        if (!d.moved && Math.hypot(x - d.x0, y - d.y0) > 3) { this.snapshot(); d.moved = true; }
        if (d.moved) d.list.forEach((m) => { m.s.x = this.g(m.x + x - d.x0); m.s.y = this.g(m.y + y - d.y0); });
        return;
      }
      if (d.type === 'resize') {
        const { o, k, s } = d, dx = x - d.x0, dy = y - d.y0, min = 24;
        let { x: nx, y: ny, w, h } = o;
        if (k.includes('e')) w = Math.max(min, this.g(o.w + dx));
        if (k.includes('s')) h = Math.max(min, this.g(o.h + dy));
        if (k.includes('w')) { w = Math.max(min, this.g(o.w - dx)); nx = o.x + o.w - w; }
        if (k.includes('n')) { h = Math.max(min, this.g(o.h - dy)); ny = o.y + o.h - h; }
        Object.assign(s, { x: nx, y: ny, w, h });
        return;
      }
      if (d.type === 'way') { const pts = this.selLine.pts.slice(); pts[d.i] = [this.g(x), this.g(y)]; this.selLine.pts = pts; return; }
      if (d.type === 'marquee') { d.x = x; d.y = y; return; }
      if (d.type === 'newline') {
        this.snap = this.nearPort(x, y, d.from);
        const a = d.from ? portsOf(this.byId.get(d.from))[d.fromPort] : { x: d.x1, y: d.y1 };
        const z = this.snap || { x, y };
        d.ghost = `M${a.x} ${a.y}L${z.x} ${z.y}`;
        d.x = x; d.y = y;
        return;
      }
      if (d.type === 'end') {
        const l = this.selLine, other = d.which === 'from' ? l.to : l.from;
        this.snap = this.nearPort(x, y, other);
        const set = d.which === 'from' ? ['from', 'fromPort', 'x1', 'y1'] : ['to', 'toPort', 'x2', 'y2'];
        if (this.snap) { l[set[0]] = this.snap.shape; l[set[1]] = this.snap.port; } else { l[set[0]] = null; l[set[1]] = null; l[set[2]] = Math.round(x); l[set[3]] = Math.round(y); }
      }
    },
    up(e) {
      if (e) this.touches.delete(e.pointerId);
      if (this.touches.size < 2) this.pinch = null;
      const d = this.drag;
      this.drag = null;
      if (!d || d.type === 'pan') return;
      if (d.type === 'newline') {
        const t = this.snap; this.snap = null;
        const far = d.from || Math.hypot(d.x - d.x1, d.y - d.y1) > 8;
        if ((t || far) && !(d.from && !t && !far)) {
          if (!d.from && !t && !far) return;
          this.snapshot();
          const l = { id: uid(), from: d.from || null, fromPort: d.from ? d.fromPort : null, x1: Math.round(d.x1 || 0), y1: Math.round(d.y1 || 0), to: t ? t.shape : null, toPort: t ? t.port : null, x2: Math.round(d.x), y2: Math.round(d.y), kind: this.lastKind || 'elbow', end: 'arrow', pts: [] };
          if (d.from && !t && Math.hypot(d.x - portsOf(this.byId.get(d.from))[d.fromPort].x, d.y - portsOf(this.byId.get(d.from))[d.fromPort].y) < 12) { this.past.pop(); return; }
          this.b.lines.push(l);
          this.sel = new Set([l.id]);
          this.tool = 'select';
          this.commit();
        }
        return;
      }
      this.snap = null;
      if (d.type === 'marquee') {
        const x0 = Math.min(d.x0, d.x), x1 = Math.max(d.x0, d.x), y0 = Math.min(d.y0, d.y), y1 = Math.max(d.y0, d.y);
        if (x1 - x0 > 4 || y1 - y0 > 4) this.sel = new Set(this.b.shapes.filter((s) => s.x >= x0 && s.y >= y0 && s.x + s.w <= x1 && s.y + s.h <= y1).map((s) => s.id));
        return;
      }
      if (d.type !== 'move' || d.moved) this.commit();
    },
    dbl(e) {
      const ds = e.target.dataset || {};
      if (ds.way !== undefined) { this.snapshot(); const pts = this.selLine.pts.slice(); pts.splice(Number(ds.way), 1); this.selLine.pts = pts; this.commit(); return; }
      if (ds.line && this.selLine && !this.selLine.locked) { const v = window.prompt('Text on this connector', this.selLine.label || ''); if (v !== null) { this.setLine('label', v); this.commit(); } return; }
      const g = e.target.closest && e.target.closest('[data-shape]');
      if (!g) return;
      const s = this.byId.get(g.getAttribute('data-shape'));
      this.sel = new Set([s.id]);
      if (s.locked) return;
      this.editing = { x: s.x, y: s.y, w: Math.max(s.w, 80), h: Math.max(s.h, 40), text: s.text };
      this.$nextTick(() => this.$refs.inline && this.$refs.inline.focus());
    },
    onKey(e) {
      if (/^(INPUT|TEXTAREA|SELECT)$/.test(e.target.tagName)) return;
      const c = e.ctrlKey || e.metaKey, k = e.key.toLowerCase();
      if (k === ' ') { e.preventDefault(); this.panning = true; return; }
      if (c && (k === '=' || k === '+')) { e.preventDefault(); this.setZoom(this.zoom + 0.1); return; }
      if (c && k === '-') { e.preventDefault(); this.setZoom(this.zoom - 0.1); return; }
      if (c && k === '0') { e.preventDefault(); this.setZoom(1); return; }
      if (k === 'escape' && this.full && !this.hasSel && !this.placing && this.tool === 'select') { this.toggleFull(); return; }
      if ((k === 'delete' || k === 'backspace') && this.hasSel) { e.preventDefault(); this.removeSel(); return; }
      if (c && k === 'z') { e.preventDefault(); if (e.shiftKey ? this.future.length : this.past.length) (e.shiftKey ? this.redo : this.undo)(); return; }
      if (c && k === 'y') { e.preventDefault(); if (this.future.length) this.redo(); return; }
      if (c && k === 'l' && this.hasSel) { e.preventDefault(); this.toggleLock(); return; }
      if (c && k === 'c') { this.copyOut(); return; }
      if (c && k === 'v') { e.preventDefault(); this.paste(); return; }
      if (c && k === 'd') { e.preventDefault(); this.duplicate(); return; }
      if (c && k === 'a') { e.preventDefault(); this.sel = new Set(this.b.shapes.map((s) => s.id)); return; }
      if (k === 'escape') { this.sel = new Set(); this.placing = null; this.tool = 'select'; return; }
      const step = { arrowleft: [-1, 0], arrowright: [1, 0], arrowup: [0, -1], arrowdown: [0, 1] }[k];
      if (step && this.selShapes.length) { e.preventDefault(); this.snapshot(); const n = e.shiftKey ? 20 : 2; this.selShapes.filter((s) => !s.locked).forEach((s) => { s.x += step[0] * n; s.y += step[1] * n; }); this.commit(); }
    },
    onKeyUp(e) { if (e.key === ' ') this.panning = false; },
    save(name, url) { const a = document.createElement('a'); a.href = url; a.download = name; a.click(); },
    exportSvg() { const u = URL.createObjectURL(new Blob([boardSvg(JSON.stringify(this.b))], { type: 'image/svg+xml' })); this.save('diagram.svg', u); setTimeout(() => URL.revokeObjectURL(u), 2000); },
    exportPng() {
      const img = new Image();
      img.onload = () => { const c = document.createElement('canvas'); c.width = img.naturalWidth * 2; c.height = img.naturalHeight * 2; const x = c.getContext('2d'); x.fillStyle = '#fff'; x.fillRect(0, 0, c.width, c.height); x.drawImage(img, 0, 0, c.width, c.height); this.save('diagram.png', c.toDataURL('image/png')); };
      img.src = 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(boardSvg(JSON.stringify(this.b)));
    }
  }
};
</script>

<style>
.bd { display: flex; flex-direction: column; gap: 8px; min-width: 0; outline: none; color: var(--k-E8EFEA,#E8EFEA); }
.bd-full { position: fixed; inset: 0; z-index: 3000; box-sizing: border-box; padding: 10px; background: var(--k-0F1714,#0F1714); }
.bd-zoom { display: inline-flex; align-items: center; gap: 4px; }
.bd-zoom input[type=range] { width: 90px; accent-color: var(--k-B8F36B,#B8F36B); }
.bd-btn.bd-focus { background: #B8F36B; border-color: #B8F36B; color: #0F1714; font-weight: 600; }
.bd-sr { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); }
.bd-top, .bd-props { display: flex; flex-wrap: wrap; gap: 4px 6px; align-items: center; }
.bd-props { min-height: 40px; padding: 6px 10px; border-radius: 10px; background: var(--k-121B17,#121B17); font-size: 12px; }
.bd-props label { display: inline-flex !important; flex-direction: row !important; align-items: center !important; width: auto !important; flex: none !important; gap: 4px !important; color: var(--k-9DB0A6,#9DB0A6); }
.bd-props input[type=text] { width: 170px !important; }
.bd-props input[type=number] { width: 54px !important; }
.bd-props input[type=text], .bd-props input[type=number], .bd-props select { height: 28px; padding: 0 6px; border-radius: 6px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font: inherit; font-size: 12px; }
.bd-props input[type=color] { width: 28px !important; height: 26px; padding: 0; border: 1px solid var(--k-2A3832,#2A3832); border-radius: 5px; background: transparent; }
.bd-btn { min-height: 30px; padding: 0 10px; border-radius: 8px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font: inherit; font-size: 12px; }
.bd-btn.on { background: var(--k-22302A,#22302A); border-color: var(--k-B8F36B,#B8F36B); }
.bd-btn:disabled { opacity: .4; cursor: default; }
.bd-sel { height: 30px; padding: 0 6px; border-radius: 8px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font: inherit; font-size: 12px; }
.bd-sep { width: 1px; height: 20px; margin: 0 2px; background: var(--k-2A3832,#2A3832); }
.bd-mut { font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); }
.bd-check { display: inline-flex !important; flex-direction: row !important; align-items: center !important; width: auto !important; flex: none !important; gap: 4px !important; font-size: 12px; }
.bd-body { display: flex; flex-wrap: wrap; gap: 8px; min-width: 0; }
.bd-lib { flex: 0 1 236px; max-width: 100%; max-height: 72vh; box-sizing: border-box; overflow-y: auto; padding: 8px; border-radius: 10px; border: 1px solid var(--k-24322C,#24322C); background: var(--k-121B17,#121B17); scrollbar-width: thin; }
.bd-find { display: block !important; margin-bottom: 6px; }
.bd-find input { width: 100%; height: 32px; box-sizing: border-box; padding: 0 10px; border-radius: 8px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font: inherit; font-size: 12px; }
.bd-lib summary { position: sticky; top: -8px; z-index: 1; cursor: pointer; padding: 6px 4px; font-size: 12px; font-weight: 700; color: var(--k-C9D6CF,#C9D6CF); background: var(--k-121B17,#121B17); }
.bd-lib-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 4px; padding-bottom: 6px; }
.bd-lib-grid button { display: flex; flex-direction: column; align-items: center; gap: 3px; min-height: 58px; padding: 6px 2px; border: 1px solid var(--k-24322C,#24322C); border-radius: 8px; background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); cursor: grab; font: inherit; font-size: 10px; line-height: 1.15; text-align: center; }
.bd-lib-grid button:hover { border-color: var(--k-9DB0A6,#9DB0A6); background: var(--k-22302A,#22302A); }
.bd-lib-grid button.on { border: 2px solid var(--k-B8F36B,#B8F36B); background: var(--k-22302A,#22302A); }
.bd-ico { fill: #7CC4FF40; stroke: currentColor; }
.bd-canvas { flex: 999 1 400px; min-width: 0; height: 72vh; overflow: auto; border: 1px solid var(--k-2A3832,#2A3832); border-radius: 10px; scrollbar-width: thin; }
.bd-full .bd-body { flex: 1; min-height: 0; flex-wrap: nowrap; }
.bd-full .bd-lib, .bd-full .bd-canvas { height: 100%; max-height: none; }
.bd-canvas::-webkit-scrollbar, .bd-lib::-webkit-scrollbar { display: block; width: 8px; height: 8px; }
.bd-canvas::-webkit-scrollbar-thumb, .bd-lib::-webkit-scrollbar-thumb { background: #9aa; border-radius: 4px; }
</style>
