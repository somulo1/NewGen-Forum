<!-- Storage by person (Settings > Security): who stores the most, and a different limit for one person. -->
<template>
<section aria-label="Storage by person" style="display: flex; flex-direction: column; gap: 10px; padding-top: 12px; border-top: 1px solid var(--k-24322C,#24322C)">
<div><strong style="font-size: 14px">Storage by person</strong><span style="display: block; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">The people who store the most, and anyone with their own limit. Give someone more or less space than the default; "Default" puts them back on it.</span></div>
<input type="search" v-model="q" placeholder="Find a person" aria-label="Find a person" style="height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; max-width: 320px"/>
<template v-if="err"><div role="alert" style="padding: 10px 14px; border-radius: 12px; background: var(--k-3A2420,#3A2420); color: #FFD3CA; font-size: 13px">{{ err }}</div></template>
<template v-if="ok"><div role="status" style="font-size: 13px; color: #B8F36B">{{ ok }}</div></template>
<span v-if="loaded && !shown.length" style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Nobody has uploaded anything yet.</span>
<div v-for="p in shown" :key="p.person" style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding: 8px 12px; border-radius: 12px; background: var(--k-121B17,#121B17)">
<div style="flex: 1 1 200px; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px">{{ p.name }}</strong><span :style="`font-size: 12px; color: ${over(p) ? '#FF9B8A' : 'var(--k-9DB0A6,#9DB0A6)'}`">{{ size(p.used_bytes) }} of {{ p.quota_mb ? size(p.quota_mb * 1048576) : 'no limit' }}{{ p.custom ? ' (own limit)' : ' (default)' }}</span></div>
<select :value="p.custom ? String(p.quota_mb) : ''" :disabled="busy" @change="setLimit(p, $event.target.value)" :aria-label="`Storage limit for ${p.name}`" style="height: 36px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px">
<option value="">Default</option><option v-for="o in choices" :key="o" :value="o">{{ label(o) }}</option>
</select>
</div>
</section>
</template>

<script>
import { useHubStore } from '~/composables/useHub';

export default {
  name: 'StorageLimits',
  data() { return { people: [], q: '', err: '', ok: '', busy: false, loaded: false, choices: ['100', '250', '500', '1024', '2048', '5120', '10240', '20480', '51200', '102400', '0'] }; },
  computed: { shown() { const q = this.q.trim().toLowerCase(); return this.people.filter((p) => !q || p.name.toLowerCase().includes(q)); } },
  mounted() { this.load(); },
  methods: {
    srv() { return useHubStore().logic.srv; },
    size(b) { return b >= 1073741824 ? (b / 1073741824).toFixed(1) + ' GB' : b >= 1048576 ? (b / 1048576).toFixed(1) + ' MB' : Math.max(0, Math.round(b / 1024)) + ' KB'; },
    label(o) { return o === '0' ? 'No limit' : Number(o) >= 1024 ? Number(o) / 1024 + ' GB' : o + ' MB'; },
    over(p) { return p.quota_mb && p.used_bytes >= p.quota_mb * 1048576 * 0.9; },
    async load() {
      const srv = this.srv();
      if (!srv) { setTimeout(() => this.load(), 1000); return; }
      try { this.people = (await srv.request('GET', '/storage/people?limit=200')).people; this.loaded = true; } catch (e) { this.err = (e && e.message) || 'Could not read storage use.'; }
    },
    async setLimit(p, v) {
      this.busy = true; this.err = ''; this.ok = '';
      try {
        await this.srv().request('PUT', '/storage/people/' + p.person, { quota_mb: v === '' ? null : Number(v) });
        this.ok = p.name + (v === '' ? ' is back on the default limit.' : ' now has ' + (v === '0' ? 'no limit.' : this.label(v) + '.'));
        await this.load();
      } catch (e) { this.err = (e && e.message) || 'Could not change the limit.'; }
      this.busy = false;
    }
  }
};
</script>
