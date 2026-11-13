<!-- Calls management > Reports: calls by day and by hour, results per agent and per queue, with export (sample figures plus your own calls) -->
<template>
<section class="pbx-mon">
  <div class="pbx-row wrap">
    <div class="pbx-seg" role="radiogroup" aria-label="Period"><button v-for="p in periods" :key="p.d" type="button" role="radio" :aria-checked="days === p.d" :class="{ on: days === p.d }" @click="days = p.d">{{ p.l }}</button></div>
    <select v-model="queue" class="pbx-search" aria-label="Queue" style="flex: 0 1 220px"><option value="">All queues</option><option v-for="q in pbx.queues" :key="q.name">{{ q.name }}</option></select>
    <span class="grow"></span>
    <button type="button" class="pbx-btn ghost small" @click="exportCsv">Export CSV</button>
  </div>

  <div class="pbx-kpis">
    <div class="pbx-kpi"><span>Calls</span><strong>{{ sum.total }}</strong></div>
    <div class="pbx-kpi"><span>Answered</span><strong>{{ pctOf(sum.answered, sum.total) }}%</strong></div>
    <div class="pbx-kpi"><span>Abandoned</span><strong>{{ pctOf(sum.abandoned, sum.total) }}%</strong></div>
    <div class="pbx-kpi"><span>Average wait</span><strong>{{ sum.wait }}s</strong></div>
    <div class="pbx-kpi"><span>Average talk</span><strong>{{ fmtDur(sum.talk * 1000) }}</strong></div>
  </div>

  <div class="pbx-card">
    <h2>Calls by day</h2>
    <div class="pbx-legend"><span><i class="in"></i>Inbound</span><span><i class="out"></i>Outbound</span><span><i class="bad"></i>Abandoned</span></div>
    <svg ref="chart" class="pbx-chart" :style="`height: ${H + 22}px`" :viewBox="`0 0 ${W} ${H + 22}`" role="img" :aria-label="'Calls per day over the last ' + days + ' days'">
      <line v-for="g in grid" :key="g.v" :x1="0" :x2="W" :y1="g.y" :y2="g.y" class="pbx-gridline"/>
      <text v-for="g in grid" :key="'t' + g.v" x="0" :y="g.y - 3" class="pbx-axis">{{ g.v }}</text>
      <g v-for="(d, i) in series" :key="d.label">
        <title>{{ d.label }}: {{ d.inb }} inbound, {{ d.out }} outbound, {{ d.ab }} abandoned</title>
        <rect :x="x(i)" :y="y(d.inb + d.out + d.ab)" :width="bw" :height="h(d.ab)" class="bad"/>
        <rect :x="x(i)" :y="y(d.inb + d.out)" :width="bw" :height="h(d.out)" class="out"/>
        <rect :x="x(i)" :y="y(d.inb)" :width="bw" :height="h(d.inb)" class="in"/>
        <text v-if="series.length <= 14 || i % 5 === 0" :x="x(i) + bw / 2" :y="H + 16" text-anchor="middle" class="pbx-axis">{{ d.short }}</text>
      </g>
    </svg>
  </div>

  <div class="pbx-grid2">
    <div class="pbx-card">
      <h2>Busiest hours</h2>
      <div class="pbx-hours"><div v-for="hr in hours" :key="hr.h" class="pbx-hour" :title="hr.h + ':00 · ' + hr.n + ' calls'"><span :style="`height: ${hr.pct}%`"></span><small>{{ hr.h }}</small></div></div>
    </div>
    <div class="pbx-card">
      <h2>Outcomes</h2>
      <ul class="pbx-list"><li v-for="o in outcomes" :key="o.name"><span class="grow">{{ o.name }}</span><span class="pbx-bar"><span :style="`width: ${o.pct}%`"></span></span><span class="pbx-sub" style="width: 40px; text-align: right">{{ o.n }}</span></li></ul>
    </div>
  </div>

  <div class="pbx-card">
    <h2>By agent</h2>
    <div class="pbx-table-wrap"><table class="pbx-table">
      <thead><tr><th>Agent</th><th>Calls</th><th>Answered</th><th>Missed</th><th>Talk time</th><th>Average call</th><th>Promise to pay</th></tr></thead>
      <tbody><tr v-for="a in pageOf('ag', byAgent)" :key="a.ext"><td><strong>{{ a.name }}</strong><span class="pbx-sub block">Ext {{ a.ext }}</span></td><td>{{ a.calls }}</td><td>{{ a.answered }}</td><td>{{ a.missed }}</td><td>{{ fmtDur(a.talk * 1000) }}</td><td>{{ fmtDur((a.answered ? a.talk / a.answered : 0) * 1000) }}</td><td>{{ a.ptp }}</td></tr></tbody>
    </table></div>
<PbxPager :size="25" :total="byAgent.length" :page="pg['ag'] || 1" @update:page="setPage('ag', $event)" />
  </div>
</section>
</template>

<script>
import { pbx, fmtDur } from '~/logic/pbx/store.js';

// steady sample figures: the same every time for a given day and queue, so the charts do not jump around
function rnd(seed) { const x = Math.sin(seed) * 10000; return x - Math.floor(x); }

export default {
  name: 'PbxReports',
  data() { return { pg: {}, pbx, days: 14, queue: '', W: 720, H: 130, periods: [{ d: 7, l: '7 days' }, { d: 14, l: '14 days' }, { d: 30, l: '30 days' }] }; },
  computed: {
    weight() { const i = this.pbx.queues.findIndex((q) => q.name === this.queue); return i === -1 ? 1 : [0.45, 0.3, 0.17, 0.08][i] || 0.1; },
    series() {
      const out = [], day = 86400000, today = new Date(); today.setHours(0, 0, 0, 0);
      for (let i = this.days - 1; i >= 0; i--) {
        const t = today.getTime() - i * day, d = new Date(t), wk = d.getDay() === 0 ? 0.25 : d.getDay() === 6 ? 0.55 : 1, s = t / day + this.queue.length;
        const mine = this.pbx.history.filter((h) => h.at >= t && h.at < t + day && (!this.queue || h.queue === this.queue));
        out.push({ label: d.toDateString(), short: d.getDate() + '/' + (d.getMonth() + 1),
          inb: Math.round((120 + rnd(s) * 60) * wk * this.weight) + mine.filter((h) => h.dir === 'in' && !/Missed/.test(h.result)).length,
          out: Math.round((70 + rnd(s + 1) * 50) * wk * this.weight) + mine.filter((h) => h.dir === 'out').length,
          ab: Math.round((4 + rnd(s + 2) * 12) * wk * this.weight) + mine.filter((h) => /Missed/.test(h.result)).length });
      }
      return out;
    },
    max() { return Math.max(10, ...this.series.map((d) => d.inb + d.out + d.ab)); },
    bw() { return Math.max(4, this.W / this.series.length - 6); },
    grid() { const step = Math.ceil(this.max / 4 / 10) * 10; return [0, 1, 2, 3, 4].map((k) => ({ v: k * step, y: this.y(k * step) })).filter((g) => g.v <= this.max * 1.05); },
    sum() {
      const S = this.series, total = S.reduce((a, d) => a + d.inb + d.out + d.ab, 0), ab = S.reduce((a, d) => a + d.ab, 0);
      return { total, answered: total - ab, abandoned: ab, wait: Math.round(14 + rnd(this.days) * 10), talk: Math.round(190 + rnd(this.days + 3) * 80) };
    },
    hours() {
      const H = [8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18].map((h) => ({ h, n: Math.round((h === 13 ? 30 : h >= 10 && h <= 11 ? 90 : h >= 15 && h <= 16 ? 75 : 50) * (0.8 + rnd(h + this.days) * 0.4) * this.weight) }));
      const m = Math.max(...H.map((x) => x.n), 1); return H.map((x) => ({ ...x, pct: Math.round((x.n / m) * 100) }));
    },
    outcomes() {
      const counts = {}; this.pbx.dispositions.forEach((d, i) => { counts[d] = Math.round((40 - i * 3.5) * (0.7 + rnd(i + this.days) * 0.6) * this.weight * (this.days / 14)); });
      this.pbx.history.forEach((h) => { if (h.disposition) counts[h.disposition] = (counts[h.disposition] || 0) + 1; });
      const m = Math.max(...Object.values(counts), 1);
      return Object.entries(counts).sort((a, b) => b[1] - a[1]).map(([name, n]) => ({ name, n, pct: Math.round((n / m) * 100) }));
    },
    byAgent() {
      return this.pbx.agents.map((a, i) => {
        const f = this.days / 14, calls = a.me ? this.pbx.history.length : Math.round((a.calls + 20) * 7 * f * (0.8 + rnd(i) * 0.4));
        const answered = a.me ? this.pbx.history.filter((h) => h.duration).length : Math.round(calls * (0.86 + rnd(i + 1) * 0.1));
        const talk = a.me ? Math.round(this.pbx.history.reduce((s, h) => s + h.duration, 0) / 1000) : answered * Math.round(180 + rnd(i + 2) * 120);
        return { ...a, calls, answered, missed: calls - answered, talk, ptp: a.me ? this.pbx.history.filter((h) => h.disposition === 'Promise to pay').length : Math.round(answered * 0.12 * (a.queues.includes('Collections') ? 1 : 0.1)) };
      }).filter((a) => !this.queue || a.queues.includes(this.queue));
    }
  },
  // the chart is drawn at the width it really has, so it fills the card without growing tall
  mounted() { const el = this.$refs.chart; if (el && window.ResizeObserver) { this._ro = new ResizeObserver(([e]) => { this.W = Math.max(280, Math.round(e.contentRect.width)); }); this._ro.observe(el); } },
  beforeUnmount() { if (this._ro) this._ro.disconnect(); },
  methods: {
    pageOf(key, rows, size) { const n = size || 25, p = this.pg[key] || 1; return rows.slice((p - 1) * n, p * n); },
    setPage(key, p) { this.pg = { ...this.pg, [key]: p }; },
    fmtDur,
    pctOf(a, b) { return b ? Math.round((a / b) * 100) : 0; },
    x(i) { return i * (this.W / this.series.length) + 3; },
    y(v) { return this.H - (v / this.max) * (this.H - 10); },
    h(v) { return (v / this.max) * (this.H - 10); },
    exportCsv() {
      const rows = [['Day', 'Inbound', 'Outbound', 'Abandoned']].concat(this.series.map((d) => [d.label, d.inb, d.out, d.ab]), [[], ['Agent', 'Calls', 'Answered', 'Missed', 'Talk (s)', 'Promise to pay']], this.byAgent.map((a) => [a.name, a.calls, a.answered, a.missed, a.talk, a.ptp]));
      const url = URL.createObjectURL(new Blob([rows.map((r) => r.map((v) => '"' + String(v).replace(/"/g, '""') + '"').join(',')).join('\n')], { type: 'text/csv' }));
      const a = document.createElement('a'); a.href = url; a.download = 'call-report-' + this.days + 'd.csv'; a.click(); setTimeout(() => URL.revokeObjectURL(url), 1000);
    }
  }
};
</script>
