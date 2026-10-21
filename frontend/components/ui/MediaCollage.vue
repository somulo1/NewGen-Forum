<!-- Files of one post, fitted to the card. Every picture or video is shown whole (never cropped) inside a thin frame; what the
     picture does not fill is the same picture, blurred and dimmed, so tall screenshots, wide banners and small icons all look
     tidy. One file uses its natural proportions (up to a height limit); several tile in a fixed grid of at most four, the last
     showing how many more there are. Click opens the viewer. -->
<template>
<div :style="framed ? 'display: flex; flex-direction: column; gap: 4px; min-width: 0; max-width: 100%; padding: 3px; border-radius: 14px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-0F1714,#0F1714)' : 'display: flex; flex-direction: column; gap: 6px; min-width: 0'">
<button v-if="visual.length === 1" type="button" class="mc-cell mc-single" @click="open(visual[0])" :aria-label="`Open ${visual[0].name}`" :style="`max-height: ${maxH}px`">
<span v-if="visual[0].isImg && visual[0].url" class="mc-bg" :style="bg(visual[0])"></span>
<img v-if="visual[0].isImg" class="mc-img" :src="visual[0].url" :alt="visual[0].name" :style="`max-height: ${maxH}px`" loading="lazy">
<template v-else><video class="mc-img" :src="visual[0].url" preload="metadata" muted playsinline :style="`max-height: ${maxH}px`"></video><span class="mc-play" aria-hidden="true">▶</span></template>
</button>
<div v-else-if="visual.length > 1" :style="gridStyle">
<template v-for="(m, i) in shown" :key="i">
<button type="button" class="mc-cell mc-fill" @click="open(m)" :aria-label="`Open ${m.name}`" :style="cell(i)">
<span v-if="m.isImg && m.url" class="mc-bg" :style="bg(m)"></span>
<img v-if="m.isImg" class="mc-img" :src="m.url" :alt="m.name" loading="lazy">
<template v-else><video class="mc-img" :src="m.url" preload="metadata" muted playsinline></video><span class="mc-play" aria-hidden="true">▶</span></template>
<span v-if="i === shown.length - 1 && extra > 0" class="mc-more" aria-hidden="true">+{{ extra }}</span>
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
  props: { files: { type: Array, default: () => [] }, framed: Boolean, maxH: { type: Number, default: 520 }, row: { type: Number, default: 210 } },
  setup() { return { viewer: useState('fileViewer', () => null) }; },
  computed: {
    visual() { return this.files.filter((m) => m.isImg || m.isVideo); },
    shown() { return this.visual.slice(0, 4); },
    extra() { return Math.max(0, this.visual.length - 4); },
    others() { return this.files.filter((m) => !m.isImg && !m.isVideo); },
    // 2 = side by side, 3 = one tall + two, 4+ = 2x2 (the fourth says how many more)
    gridStyle() {
      const n = this.shown.length;
      const r = this.row;
      const rows = n === 2 ? Math.round(r * 1.15) + 'px' : r + 'px ' + r + 'px';
      return `display: grid; gap: 3px; grid-template-columns: 1fr 1fr; grid-template-rows: ${rows}`;
    }
  },
  methods: {
    open(m) { this.viewer = { files: this.files, index: this.files.indexOf(m) }; },
    bg(m) { return { backgroundImage: 'url("' + String(m.url).replace(/"/g, '%22') + '")' }; },
    cell(i) { return this.shown.length === 3 && i === 0 ? 'grid-row: span 2' : ''; }
  }
};
</script>

<style scoped>
.mc-cell { position: relative; display: flex; align-items: center; justify-content: center; min-width: 0; min-height: 0; padding: 0; overflow: hidden; border: 1px solid rgba(255, 255, 255, 0.08); border-radius: 12px; background: #0F1714; cursor: zoom-in; width: 100%; }
.mc-single { min-height: 96px; }
.mc-bg { position: absolute; inset: -24px; background-size: cover; background-position: center; filter: blur(20px) brightness(0.5) saturate(1.15); pointer-events: none; }
.mc-img { position: relative; display: block; max-width: 100%; object-fit: contain; background: transparent; }
.mc-single .mc-img { width: auto; height: auto; max-width: 100%; }
.mc-fill .mc-img { width: 100%; height: 100%; }
.mc-play { position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%); width: 46px; height: 46px; border-radius: 50%; background: rgba(15, 23, 20, 0.72); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 18px; pointer-events: none; }
.mc-more { position: absolute; inset: 0; display: flex; align-items: center; justify-content: center; background: rgba(8, 14, 11, 0.62); color: #fff; font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700; pointer-events: none; }
</style>
