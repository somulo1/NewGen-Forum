<!-- In-page file viewer: images, video, PDF, text and CSV open here; anything else offers a download. -->
<template>
<div v-if="viewer" class="hub-overlay" role="dialog" aria-modal="true" :aria-label="`Viewing ${file.name}`" @click.self="close" tabindex="-1" ref="root" @keydown="onKey" style="position: fixed; inset: 0; z-index: 80; background: rgba(5,10,8,0.88); display: flex; flex-direction: column; outline: none">
<div style="flex: none; display: flex; align-items: center; gap: 10px; padding: 12px 16px; color: var(--k-E8EFEA,#E8EFEA)">
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 15px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ file.name }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ file.meta }}<template v-if="viewer.files.length > 1"> · {{ viewer.index + 1 }} of {{ viewer.files.length }}</template></span></div>
<a :href="file.url" :download="file.name" style="height: 36px; padding: 0 16px; border-radius: 18px; background: #B8F36B; color: #0F1714; font-weight: 600; font-size: 14px; display: flex; align-items: center; text-decoration: none">Download</a>
<button type="button" aria-label="Close viewer" @click="close" style="width: 36px; height: 36px; border-radius: 50%; border: 1px solid var(--k-3A4A42,#3A4A42); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">✕</button>
</div>
<div @click.self="close" style="flex: 1; min-height: 0; position: relative; display: flex; align-items: center; justify-content: center; padding: 0 56px 8px; box-sizing: border-box">
<button v-if="many" type="button" aria-label="Previous file" @click="step(-1)" style="position: absolute; left: 8px; top: 50%; transform: translateY(-50%); width: 40px; height: 40px; border-radius: 50%; border: 1px solid var(--k-3A4A42,#3A4A42); background: rgba(15,23,20,.7); color: #fff; cursor: pointer; font-size: 18px">‹</button>
<img v-if="kind === 'image'" :src="file.url" :alt="file.name" style="max-width: 100%; max-height: 100%; object-fit: contain; border-radius: 8px">
<video v-else-if="kind === 'video'" :key="file.url" :src="file.url" controls autoplay style="max-width: 100%; max-height: 100%; border-radius: 8px; background: #000"></video>
<iframe v-else-if="kind === 'pdf'" :src="file.url" :title="file.name" style="width: 100%; height: 100%; border: 0; border-radius: 8px; background: #fff"></iframe>
<div v-else-if="kind === 'csv'" style="max-width: 100%; max-height: 100%; overflow: auto; border-radius: 8px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-2A3832,#2A3832)">
<table style="border-collapse: collapse; font-size: 14px; color: var(--k-E8EFEA,#E8EFEA)"><template v-for="(r, ri) in rows" :key="ri"><tr><template v-for="(c, ci) in r" :key="ci"><component :is="ri === 0 ? 'th' : 'td'" style="padding: 8px 14px; border-bottom: 1px solid var(--k-24322C,#24322C); text-align: left">{{ c }}</component></template></tr></template></table>
</div>
<pre v-else-if="kind === 'text'" style="margin: 0; max-width: 100%; max-height: 100%; overflow: auto; padding: 18px; border-radius: 8px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-2A3832,#2A3832); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; white-space: pre-wrap">{{ text }}</pre>
<div v-else style="max-width: 360px; padding: 28px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-2A3832,#2A3832); color: var(--k-E8EFEA,#E8EFEA); text-align: center; display: flex; flex-direction: column; gap: 10px"><strong>No preview for this file type</strong><span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ file.name }} can still be downloaded.</span></div>
<button v-if="many" type="button" aria-label="Next file" @click="step(1)" style="position: absolute; right: 8px; top: 50%; transform: translateY(-50%); width: 40px; height: 40px; border-radius: 50%; border: 1px solid var(--k-3A4A42,#3A4A42); background: rgba(15,23,20,.7); color: #fff; cursor: pointer; font-size: 18px">›</button>
</div>
<div v-if="many" style="flex: none; display: flex; gap: 6px; justify-content: center; padding: 8px 16px 14px; overflow-x: auto">
<template v-for="(f, i) in viewer.files" :key="i"><button type="button" :aria-label="`Show ${f.name}`" @click="viewer = { files: viewer.files, index: i }" :style="`flex: none; width: 56px; height: 42px; padding: 0; border-radius: 8px; overflow: hidden; border: 2px solid ${i === viewer.index ? 'var(--k-B8F36B,#B8F36B)' : 'var(--k-2A3832,#2A3832)'}; background: var(--k-16201C,#16201C); color: var(--k-C9D6CF,#C9D6CF); font-size: 10px; cursor: pointer`"><img v-if="f.isImg" :src="f.url" alt="" style="width: 100%; height: 100%; object-fit: cover"><span v-else>{{ ext(f.name) }}</span></button></template>
</div>
</div>
</template>

<script>
const EXT = (n) => ((n || '').split('.').pop() || '').toLowerCase();
export default {
  name: 'FileViewer',
  setup() { return { viewer: useState('fileViewer', () => null) }; },
  data() { return { text: '', rows: [] }; },
  computed: {
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
