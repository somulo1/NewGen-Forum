<!-- Searchable emoji picker with categories and recently used. Data comes from logic/emoji.js. -->
<template>
<div style="display: flex; flex-direction: column; gap: 6px; padding: 8px; border-radius: 14px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 16px 36px rgba(0,0,0,.5); max-height: 330px; box-sizing: border-box">
<input ref="q" v-model="q" type="search" placeholder="Search emoji" aria-label="Search emoji" style="height: 34px; padding: 0 12px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; min-width: 0">
<div v-if="!q" role="tablist" aria-label="Emoji categories" style="display: flex; gap: 2px; overflow-x: auto; flex: none">
<button v-if="recents.length" role="tab" type="button" :aria-selected="cat === 'recent'" title="Recently used" @click="cat = 'recent'" :style="tabStyle('recent')">🕘</button>
<template v-for="c in cats" :key="c.id"><button role="tab" type="button" :aria-selected="cat === c.id" :title="c.label" @click="cat = c.id" :style="tabStyle(c.id)">{{ c.icon }}</button></template>
</div>
<div style="flex: 1; min-height: 0; overflow-y: auto; min-width: 0">
<template v-if="!ready"><span style="display: block; padding: 18px; text-align: center; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Loading emoji…</span></template>
<template v-else>
<div v-if="shown.length" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(34px, 1fr)); gap: 2px">
<template v-for="e in shown" :key="e.emoji || e"><button type="button" :title="e.label || ''" :aria-label="e.label || e" @click="pick(e.emoji || e)" style="height: 34px; border: 0; border-radius: 8px; background: transparent; font-size: 21px; line-height: 1; cursor: pointer">{{ e.emoji || e }}</button></template>
</div>
<span v-else style="display: block; padding: 18px; text-align: center; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">No emoji match "{{ q }}".</span>
</template>
</div>
</div>
</template>

<script>
import { CATEGORIES, loadEmoji, recentEmoji, rememberEmoji, searchEmoji } from '~/logic/emoji.js';

export default {
  name: 'EmojiPicker',
  emits: ['pick'],
  data() { const r = recentEmoji(); return { q: '', cat: r.length ? 'recent' : 0, ready: false, data: null, recents: r, cats: CATEGORIES }; },
  computed: {
    shown() {
      if (!this.data) return [];
      if (this.q.trim()) return searchEmoji(this.data, this.q);
      if (this.cat === 'recent') return this.recents;
      return this.data.byGroup[this.cat] || [];
    }
  },
  async mounted() { this.data = await loadEmoji(); this.ready = true; this.$nextTick(() => this.$refs.q && this.$refs.q.focus()); },
  methods: {
    tabStyle(id) { const on = this.cat === id; return `flex: none; width: 34px; height: 30px; border: 0; border-radius: 8px; background: ${on ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; font-size: 17px; cursor: pointer; border-bottom: 2px solid ${on ? 'var(--k-B8F36B,#B8F36B)' : 'transparent'}`; },
    pick(ch) { rememberEmoji(ch); this.recents = recentEmoji(); this.$emit('pick', ch); }
  }
};
</script>
