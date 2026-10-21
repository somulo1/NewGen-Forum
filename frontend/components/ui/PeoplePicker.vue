<!-- Find people already in Hub by typing part of a name, title or department, and pick them. In "instant" mode a pick is handed to
     the parent at once (to add them straight away); otherwise picked people show as chips and can be taken off again. -->
<template>
<div ref="box" style="position: relative; display: flex; flex-direction: column; gap: 8px; min-width: 0">
<template v-if="!instant && chips.length"><div style="display: flex; flex-wrap: wrap; gap: 6px">
<template v-for="c in chips" :key="c.uid"><span style="display: inline-flex; align-items: center; gap: 6px; height: 28px; padding: 0 6px 0 10px; border-radius: 14px; background: var(--k-22302A,#22302A); border: 1px solid var(--k-2A3832,#2A3832); font-size: 12px">{{ c.name }}<button type="button" :aria-label="`Remove ${c.name}`" @click="$emit('remove', c.uid)" style="width: 18px; height: 18px; padding: 0; border: 0; border-radius: 50%; background: transparent; color: var(--k-9DB0A6,#9DB0A6); cursor: pointer; font-size: 14px; line-height: 1">×</button></span></template>
</div></template>
<input type="search" v-model="q" @focus="open = true" @input="open = true" @keydown.down.prevent="move(1)" @keydown.up.prevent="move(-1)" @keydown.enter.prevent="chooseActive" @keydown.esc="open = false" :placeholder="placeholder" :aria-label="placeholder" autocomplete="off" role="combobox" :aria-expanded="open && matches.length > 0" style="height: 36px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px; min-width: 0; width: 100%; box-sizing: border-box"/>
<div v-if="open && q.trim()" role="listbox" style="position: absolute; left: 0; right: 0; top: 100%; z-index: 80; margin-top: 4px; max-height: 260px; overflow-y: auto; padding: 4px; border-radius: 12px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 14px 30px rgba(0,0,0,0.45)">
<template v-for="(m, i) in matches" :key="m.uid"><button type="button" role="option" :aria-selected="i === active" @mousedown.prevent="choose(m)" @mouseenter="active = i" :style="`display: flex; align-items: center; gap: 10px; width: 100%; padding: 6px 8px; border: 0; border-radius: 8px; background: ${i === active ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left`">
<span :style="`flex: none; width: 30px; height: 30px; border-radius: 50%; background: ${m.color || '#7CC4FF'}; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center`">{{ m.initials }}</span>
<span style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ m.name }}</strong><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6); overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ [m.title, m.dept].filter(Boolean).join(' · ') }}</span></span>
<span v-if="instant" style="flex: none; font-size: 11px; font-weight: 700; color: #B8F36B">Add</span></button></template>
<div v-if="!matches.length" style="padding: 10px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Nobody found. Only people who already have a Hub account can be added.</div>
</div>
</div>
</template>

<script>
export default {
  name: 'PeoplePicker',
  props: {
    people: { type: Array, default: () => [] },      // { uid, name, title, dept, initials, color }
    exclude: { type: Array, default: () => [] },     // uids that cannot be picked (already in)
    selected: { type: Array, default: () => [] },    // uids shown as chips (not instant)
    instant: Boolean,
    placeholder: { type: String, default: 'Search people by name…' }
  },
  emits: ['pick', 'remove'],
  data() { return { q: '', open: false, active: 0 }; },
  computed: {
    chips() { return this.selected.map((u) => this.people.find((p) => p.uid === u)).filter(Boolean); },
    matches() {
      const t = this.q.trim().toLowerCase();
      if (!t) return [];
      const skip = new Set(this.exclude.concat(this.selected));
      return this.people.filter((p) => !skip.has(p.uid) && [p.name, p.title, p.dept, p.email].some((x) => (x || '').toLowerCase().includes(t)))
        .sort((a, b) => (a.name.toLowerCase().startsWith(t) ? 0 : 1) - (b.name.toLowerCase().startsWith(t) ? 0 : 1) || a.name.localeCompare(b.name)).slice(0, 8);
    }
  },
  watch: { q() { this.active = 0; } },
  mounted() { this._off = (e) => { if (this.$refs.box && !this.$refs.box.contains(e.target)) this.open = false; }; document.addEventListener('pointerdown', this._off, true); },
  beforeUnmount() { document.removeEventListener('pointerdown', this._off, true); },
  methods: {
    move(d) { if (this.matches.length) this.active = (this.active + d + this.matches.length) % this.matches.length; },
    chooseActive() { if (this.matches[this.active]) this.choose(this.matches[this.active]); },
    choose(m) { this.$emit('pick', m.uid); this.q = ''; this.active = 0; if (this.instant) this.open = false; }
  }
};
</script>
