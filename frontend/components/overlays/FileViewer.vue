<!-- In-page file viewer: images, video, PDF, text and CSV open here; anything else offers a download. -->
<template>
<div v-if="viewer" :class="['hub-overlay', 'fv', { 'fv-light': isLight }]" role="dialog" aria-modal="true" :aria-label="`Viewing ${file.name}`" @click.self="close" tabindex="-1" ref="root" @keydown="onKey" style="position: fixed; inset: 0; z-index: 80; background: var(--fv-bg); display: flex; flex-direction: column; outline: none">
<div style="flex: none; display: flex; align-items: center; gap: 10px; padding: 12px 16px; color: var(--fv-fg)">
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 15px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ file.name }}</strong><span style="font-size: 12px; color: var(--fv-muted)">{{ file.meta }}<template v-if="viewer.files.length > 1"> · {{ viewer.index + 1 }} of {{ viewer.files.length }}</template></span></div>
<a :href="file.url" :download="file.name" style="height: 36px; padding: 0 16px; border-radius: 18px; background: #B8F36B; color: #0F1714; font-weight: 600; font-size: 14px; display: flex; align-items: center; text-decoration: none">Download</a>
<button type="button" aria-label="Close viewer" @click="close" class="fv-btn" style="width: 38px; height: 38px; font-size: 16px">✕</button>
</div>
<div @click.self="close" style="flex: 1; min-height: 0; position: relative; display: flex; align-items: center; justify-content: center; padding: 0 76px 8px; box-sizing: border-box">
<button v-if="many" type="button" class="fv-btn fv-arrow fv-prev" aria-label="Previous file" title="Previous (←)" @click="step(-1)"><svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round"><path d="M15 5l-7 7 7 7"></path></svg></button>
<img v-if="kind === 'image'" :src="file.url" :alt="file.name" style="max-width: 100%; max-height: 100%; object-fit: contain; border-radius: 8px">
<video v-else-if="kind === 'video'" :key="file.url" :src="file.url" controls autoplay style="max-width: 100%; max-height: 100%; border-radius: 8px; background: #000"></video>
<iframe v-else-if="kind === 'pdf'" :src="file.url" :title="file.name" style="width: 100%; height: 100%; border: 0; border-radius: 8px; background: #fff"></iframe>
<div v-else-if="kind === 'csv'" style="max-width: 100%; max-height: 100%; overflow: auto; border-radius: 8px; background: var(--fv-panel); border: 1px solid var(--fv-line)">
<table style="border-collapse: collapse; font-size: 14px; color: var(--fv-fg)"><template v-for="(r, ri) in rows" :key="ri"><tr><template v-for="(c, ci) in r" :key="ci"><component :is="ri === 0 ? 'th' : 'td'" style="padding: 8px 14px; border-bottom: 1px solid var(--k-24322C,#24322C); text-align: left">{{ c }}</component></template></tr></template></table>
</div>
<pre v-else-if="kind === 'text'" style="margin: 0; max-width: 100%; max-height: 100%; overflow: auto; padding: 18px; border-radius: 8px; background: var(--fv-panel); border: 1px solid var(--fv-line); color: var(--fv-fg); font-size: 14px; white-space: pre-wrap">{{ text }}</pre>
<div v-else style="max-width: 360px; padding: 28px; border-radius: 16px; background: var(--fv-panel); border: 1px solid var(--fv-line); color: var(--fv-fg); text-align: center; display: flex; flex-direction: column; gap: 10px"><strong>No preview for this file type</strong><span style="font-size: 13px; color: var(--fv-muted)">{{ file.name }} can still be downloaded.</span></div>
<button v-if="many" type="button" class="fv-btn fv-arrow fv-next" aria-label="Next file" title="Next (→)" @click="step(1)"><svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round"><path d="M9 5l7 7-7 7"></path></svg></button>
</div>
<div v-if="many" style="flex: none; display: flex; gap: 6px; justify-content: center; padding: 8px 16px 14px; overflow-x: auto">
<template v-for="(f, i) in viewer.files" :key="i"><button type="button" :aria-label="`Show ${f.name}`" @click="viewer = { files: viewer.files, index: i }" :style="`flex: none; width: 56px; height: 42px; padding: 0; border-radius: 8px; overflow: hidden; border: 2px solid ${i === viewer.index ? '#B8F36B' : 'var(--fv-line)'}; background: var(--fv-panel); color: var(--fv-fg); opacity: ${i === viewer.index ? 1 : 0.8}; font-size: 10px; cursor: pointer`"><img v-if="f.isImg" :src="f.url" alt="" style="width: 100%; height: 100%; object-fit: cover"><span v-else>{{ ext(f.name) }}</span></button></template>
</div>
</div>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

const EXT = (n) => ((n || '').split('.').pop() || '').toLowerCase();
export default {
  name: 'FileViewer',
  setup() { return { viewer: useState('fileViewer', () => null) }; },
  data() { return { text: '', rows: [] }; },
  computed: {
    ...hubBindings(['isLight']).computed,
    file() { return this.viewer ? this.viewer.files[this.viewer.index] : {}; },
    many() { return this.viewer && this.viewer.files.length > 1; },
    kind() {
      const f = this.file, e = EXT(f.name);
      if (f.isImg) return 'image';
      if (f.isVideo) return 'video';
      if (e === 'pdf') return 'pdf';
      if (e === 'csv') return 'csv';
      if (['txt', 'md', 'log', 'json'].indexOf(e) !== -1) return 'text';
      return 'other';
    }
  },
  watch: {
    file: { immediate: true, handler(f) { this.load(f); } },
    viewer(v) { if (v) this.$nextTick(() => this.$refs.root && this.$refs.root.focus()); }
  },
  methods: {
    ext: EXT,
    close() { this.viewer = null; },
    step(d) { const n = this.viewer.files.length; this.viewer = { files: this.viewer.files, index: (this.viewer.index + d + n) % n }; },
    onKey(e) { if (e.key === 'Escape') this.close(); else if (this.many && e.key === 'ArrowRight') this.step(1); else if (this.many && e.key === 'ArrowLeft') this.step(-1); },
    async load(f) {
      this.text = ''; this.rows = [];
      if (!f || !f.url || (this.kind !== 'text' && this.kind !== 'csv')) return;
      try {
        const t = await (await fetch(f.url)).text();
        if (this.kind === 'csv') this.rows = t.trim().split(/\r?\n/).map((l) => l.split(',')); else this.text = t;
      } catch (e) { this.text = 'Could not load this file.'; }
    }
  }
};
</script>

<style scoped>
/* The viewer follows the theme: a dark backdrop with light controls, or a light backdrop with dark controls. The arrows are solid
   circles with a strong contrast against both the backdrop and any picture behind them. */
.fv { --fv-bg: rgba(5, 10, 8, 0.92); --fv-fg: #F3F7F4; --fv-muted: #B5C4BB; --fv-panel: #16201C; --fv-line: #34453D; --fv-btn-bg: rgba(255, 255, 255, 0.94); --fv-btn-fg: #0F1714; --fv-btn-line: rgba(0, 0, 0, 0.35); }
.fv.fv-light { --fv-bg: rgba(236, 242, 239, 0.95); --fv-fg: #0F1714; --fv-muted: #4A5C54; --fv-panel: #FFFFFF; --fv-line: #C4D2CA; --fv-btn-bg: rgba(15, 23, 20, 0.9); --fv-btn-fg: #FFFFFF; --fv-btn-line: rgba(255, 255, 255, 0.55); }
.fv-btn { display: flex; align-items: center; justify-content: center; padding: 0; border-radius: 50%; border: 1px solid var(--fv-btn-line); background: var(--fv-btn-bg); color: var(--fv-btn-fg); cursor: pointer; box-shadow: 0 4px 16px rgba(0, 0, 0, 0.35); transition: transform 0.12s ease; }
.fv-btn:hover { transform: scale(1.07); }
.fv-btn:focus-visible { outline: 3px solid #B8F36B; outline-offset: 2px; }
.fv-arrow { position: absolute; top: 50%; width: 52px; height: 52px; margin-top: -26px; z-index: 2; }
.fv-arrow:hover { transform: scale(1.07); }
.fv-prev { left: 12px; }
.fv-next { right: 12px; }
@media (max-width: 600px) { .fv-arrow { width: 44px; height: 44px; margin-top: -22px; } .fv-prev { left: 6px; } .fv-next { right: 6px; } }
</style>
