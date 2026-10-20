<!-- Who uses the system: people who used it today (each counted once however often they signed in), and who has been away. -->
<template>
<section v-if="data" aria-label="System use" style="padding: 18px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 14px">
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); gap: 12px">
<div style="padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; gap: 2px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">People who used Hub today</span><strong style="font-size: 28px; font-family: 'Space Grotesk', sans-serif">{{ data.today }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">of {{ data.active }} · </span></div>
<button type="button" @click="open = open === 'w' ? '' : 'w'" :aria-expanded="open === 'w'" style="text-align: left; cursor: pointer; border: 0; color: inherit; padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; gap: 2px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Not used for 1 week</span><strong style="font-size: 28px; font-family: 'Space Grotesk', sans-serif; color: #FFB86B">{{ data.inactive_1w.length }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">7 to 13 days · tap to see who</span></button>
<button type="button" @click="open = open === 'f' ? '' : 'f'" :aria-expanded="open === 'f'" style="text-align: left; cursor: pointer; border: 0; color: inherit; padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; gap: 2px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Not used for 2 weeks or more</span><strong style="font-size: 28px; font-family: 'Space Grotesk', sans-serif; color: #FF9B8A">{{ data.inactive_2w.length }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">14 days or more, or never · tap to see who</span></button>
</div>
<template v-if="open">
<div style="display: flex; flex-direction: column; gap: 6px; max-height: 280px; overflow-y: auto">
<span v-if="!list.length" style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Nobody.</span>
<div v-for="p in list" :key="p.person" style="display: flex; flex-wrap: wrap; gap: 4px 12px; padding: 8px 12px; border-radius: 10px; background: var(--k-121B17,#121B17)"><strong style="font-size: 14px; flex: 1 1 180px">{{ p.name }}<span v-if="p.title" style="font-weight: 400; color: var(--k-9DB0A6,#9DB0A6)"> · {{ p.title }}</span></strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ p.last_seen ? 'Last seen ' + day(p.last_seen) : 'Never signed in' }}</span></div>
</div>
</template>
</section>
<div v-else-if="err" role="alert" style="padding: 10px 14px; border-radius: 12px; background: var(--k-3A2420,#3A2420); color: #FFD3CA; font-size: 13px">{{ err }}</div>
</template>

<script>
import { useHubStore } from '~/composables/useHub';

export default {
  name: 'UsageCard',
  data() { return { data: null, open: '', err: '', timer: 0 }; },
  computed: { list() { return this.open === 'w' ? this.data.inactive_1w : this.data.inactive_2w; } },
  mounted() { this.load(); this.timer = setInterval(() => this.load(), 60000); },
  beforeUnmount() { clearInterval(this.timer); },
  methods: {
    day(d) { return new Date(d).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }); },
    async load() {
      const srv = useHubStore().logic.srv;
      if (!srv) { setTimeout(() => this.load(), 1000); return; }
      try { this.data = await srv.request('GET', '/security/activity'); this.err = ''; } catch (e) { if (!this.data) this.err = (e && e.message) || 'Could not read system use.'; }
    }
  }
};
</script>
