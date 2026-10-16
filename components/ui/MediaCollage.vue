<!-- Files of one post, fitted to the card: one file fills it, several tile together like a collage. Click opens the viewer. -->
<template>
<div :style="framed ? 'display: flex; flex-direction: column; gap: 4px; min-width: 0; max-width: 100%; padding: 3px; border-radius: 14px; border: 1px solid #2A3832; background: #0F1714' : 'display: flex; flex-direction: column; gap: 6px; min-width: 0'">
<template v-if="visual.length === 1">
<button type="button" @click="open(visual[0])" :aria-label="`Open ${visual[0].name}`" style="position: relative; display: block; width: 100%; padding: 0; border: 0; background: #0F1714; cursor: zoom-in; overflow: hidden; border-radius: 12px">
<img v-if="visual[0].isImg" :src="visual[0].url" :alt="visual[0].name" :style="`display: block; width: 100%; max-height: ${maxH}px; object-fit: contain`">
<template v-else><video :src="visual[0].url" preload="metadata" muted playsinline :style="`display: block; width: 100%; max-height: ${maxH}px; background: #000`"></video><span class="mc-play" aria-hidden="true">▶</span></template>
</button>
</template>
<div v-else-if="visual.length > 1" :style="gridStyle">
<template v-for="(m, i) in visual" :key="i">
<button type="button" @click="open(m)" :aria-label="`Open ${m.name}`" :style="`position: relative; padding: 0; border: 0; overflow: hidden; border-radius: 10px; background: #0F1714; cursor: zoom-in; min-width: 0; min-height: 0; ${cell(i)}`">
<img v-if="m.isImg" :src="m.url" :alt="m.name" style="display: block; width: 100%; height: 100%; object-fit: cover">
<template v-else><video :src="m.url" preload="metadata" muted playsinline style="display: block; width: 100%; height: 100%; object-fit: cover; background: #000"></video><span class="mc-play" aria-hidden="true">▶</span></template>
</button>
</template>
</div>
<div v-if="others.length" style="display: flex; flex-wrap: wrap; gap: 6px; min-width: 0">
<template v-for="(m, i) in others" :key="i"><FileChip :name="m.name" :meta="m.meta" :url="m.url" :group="files" :index="files.indexOf(m)"/></template>
</div>
</div>
</template>

<script>
export default {
  name: 'MediaCollage',
  props: { files: { type: Array, default: () => [] }, framed: Boolean, maxH: { type: Number, default: 460 }, row: { type: Number, default: 170 } },
  setup() { return { viewer: useState('fileViewer', () => null) }; },
  computed: {
    visual() { return this.files.filter((m) => m.isImg || m.isVideo); },
    others() { return this.files.filter((m) => !m.isImg && !m.isVideo); },
    // 2 = side by side, 3 = one tall + two, 4 = 2x2, 5 = two over three.
    gridStyle() {
      const n = this.visual.length;
      const cols = n === 5 ? 'repeat(6, 1fr)' : '1fr 1fr';
      const r = this.row;
      const rows = n === 2 ? (r + 40) + 'px' : n === 5 ? (r + 40) + 'px ' + (r - 20) + 'px' : r + 'px ' + r + 'px';
      return `display: grid; gap: 4px; grid-template-columns: ${cols}; grid-template-rows: ${rows}`;
    }
  },
  methods: {
    open(m) { this.viewer = { files: this.files, index: this.files.indexOf(m) }; },
    cell(i) {
      const n = this.visual.length;
      if (n === 3 && i === 0) return 'grid-row: span 2;';
      if (n === 5) return i < 2 ? 'grid-column: span 3;' : 'grid-column: span 2;';
      return '';
    }
  }
};
</script>

<style scoped>
.mc-play { position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%); width: 46px; height: 46px; border-radius: 50%; background: rgba(15,23,20,0.72); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 18px; pointer-events: none; }
</style>
