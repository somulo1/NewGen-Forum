<!-- How I did this month in each place I belong to: only my own figures, whatever my rights. It counts what was assigned to me
     (tickets, business requests, and communications places send each other), resolved or escalated on within the turnaround. -->
<template>
<section v-if="places.length || loading" aria-label="My performance" style="padding: 16px 20px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px"><strong style="flex: 1 1 auto; font-size: 15px">My performance</strong>
<select :value="period" @change="(e) => { period = e.target.value; load(); }" aria-label="Period" style="height: 34px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA)"><option value="month">This month</option><option value="quarter">This quarter</option><option value="year">This year</option></select></div>
<span v-if="loading" style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Loading…</span>
<div v-for="pl in places" :key="pl.unit.id" style="display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ pl.unit.name }}</span>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap: 8px">
<div v-for="c in cards(pl)" :key="c.label" style="padding: 10px 12px; border-radius: 12px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; gap: 2px">
<span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 20px; font-weight: 700; color: ${c.color}`">{{ c.value }}</span>
<span style="font-size: 12px">{{ c.label }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ c.note }}</span>
</div>
</div>
</div>
</section>
</template>
<script>
import { useHubStore } from '~/composables/useHub';
export default {
  name: 'MyPerformance',
  data() { return { places: [], loading: false, period: 'month' }; },
  mounted() { this.load(); },
  methods: {
    async load() {
      const srv = useHubStore().logic.srv;
      if (!srv) return;
      this.loading = true;
      try { this.places = (await srv.request('GET', '/org/my-performance?period=' + this.period)).places || []; } catch (e) { this.places = []; } finally { this.loading = false; }
    },
    cards(pl) {
      const m = pl.me, col = (p, n) => (!n ? '#9DB0A6' : p >= 95 ? '#B8F36B' : p >= 85 ? '#FFB86B' : '#FF9B8A');
      const card = (label, part) => ({ label, value: part.received ? part.pct + '%' : '—', color: col(part.pct, part.received), note: part.received + part.waiting ? part.ok + ' resolved or escalated on time of ' + part.received + (part.waiting ? ' · ' + part.waiting + ' not due yet' : '') : 'none assigned' });
      return [{ label: 'Overall', value: m.measured ? m.perf + '%' : '—', color: col(m.perf, m.measured), note: m.measured ? 'from ' + m.measured + ' decided' : 'Nothing due yet' },
        card('Tickets', m.tickets), ...(pl.unit.kind === 'department' ? [card('Business requests', m.requests)] : []), card('Communications', m.comms)];
    }
  }
};
</script>
