<!-- Calls management > Recordings: every recorded call, searchable, with playback, download and how long recordings are kept -->
<template>
<section class="pbx-card">
  <div class="pbx-row wrap">
    <input v-model="q" type="search" class="pbx-search" placeholder="Search customer, number, outcome or note" aria-label="Search recordings">
    <input v-model="from" type="date" class="pbx-search" style="flex: 0 1 170px" aria-label="From date">
    <select v-model="outcome" class="pbx-search" style="flex: 0 1 200px" aria-label="Outcome"><option value="">Any outcome</option><option v-for="d in pbx.dispositions" :key="d">{{ d }}</option></select>
  </div>
  <div class="pbx-table-wrap"><table class="pbx-table">
    <thead><tr><th>Customer</th><th>Direction</th><th>Outcome</th><th>Length</th><th>When</th><th>Agent</th><th></th></tr></thead>
    <tbody>
      <tr v-for="h in pageOf('rec', rows)" :key="h.id">
        <td><strong>{{ name(h) }}</strong><span class="pbx-sub block">{{ showNumber(h.number) }}</span></td>
        <td>{{ h.dir === 'in' ? 'Inbound' : 'Outbound' }}</td><td>{{ h.disposition || '—' }}</td><td>{{ fmtDur(h.duration) }}</td><td class="nowrap">{{ fmtWhen(h.at) }}</td><td>{{ h.agent }}</td>
        <td><div class="pbx-row"><button type="button" class="pbx-btn small ghost" @click="play(h)">{{ playing === h.id ? 'Stop' : 'Play' }}</button><button type="button" class="pbx-btn small ghost" @click="download(h)">Download</button></div></td>
      </tr>
      <tr v-if="!rows.length"><td colspan="7" class="pbx-empty">No recordings match.</td></tr>
    </tbody>
  </table></div>
<PbxPager :size="25" :total="rows.length" :page="pg['rec'] || 1" @update:page="setPage('rec', $event)" />
  <div class="pbx-row">
    <label class="pbx-field inline">Keep recordings for<select v-model="pbx.settings.keepDays" @change="saveSettings"><option :value="90">90 days</option><option :value="180">6 months</option><option :value="365">1 year</option><option :value="2555">7 years</option></select></label>
    <span class="pbx-sub">Played recordings and downloads are written to the compliance log. Playback here is a sample; the phone system streams the real audio.</span>
  </div>
</section>
</template>

<script>
import { pbx, customerOf, customerByNumber, showNumber, fmtDur, fmtWhen, logAudit, saveSettings } from '~/logic/pbx/store.js';
import { speak, stopSpeech } from '~/logic/pbx/sounds.js';

export default {
  name: 'PbxRecordings',
  data() { return { pg: {}, pbx, q: '', from: '', outcome: '', playing: null }; },
  created() { if (!this.pbx.settings.keepDays) this.pbx.settings.keepDays = 365; },
  beforeUnmount() { if (this.playing) stopSpeech(); },
  computed: {
    rows() {
      const q = this.q.trim().toLowerCase(), from = this.from ? new Date(this.from).getTime() : 0;
      return this.pbx.history.filter((h) => h.recorded && h.at >= from && (!this.outcome || h.disposition === this.outcome) && (!q || [this.name(h), h.number, h.note, h.disposition].join(' ').toLowerCase().includes(q)));
    }
  },
  methods: {
    pageOf(key, rows, size) { const n = size || 25, p = this.pg[key] || 1; return rows.slice((p - 1) * n, p * n); },
    setPage(key, p) { this.pg = { ...this.pg, [key]: p }; },
    showNumber, fmtDur, fmtWhen, saveSettings,
    name(h) { const c = h.customerId ? customerOf(h.customerId) : customerByNumber(h.number); return c ? c.name : showNumber(h.number); },
    play(h) {
      if (this.playing === h.id) { stopSpeech(); this.playing = null; return; }
      this.playing = h.id; logAudit('Played the recording of the call with ' + this.name(h) + ' (' + fmtWhen(h.at) + ')');
      speak('Sample recording. Call with ' + this.name(h) + '. ' + (h.note || 'No notes were saved for this call.'));
    },
    // a short placeholder audio file, so the download flow can be tried before real recordings exist
    download(h) {
      logAudit('Downloaded the recording of the call with ' + this.name(h));
      const rate = 8000, secs = 1, n = rate * secs, buf = new ArrayBuffer(44 + n * 2), v = new DataView(buf);
      const w = (o, s) => { for (let i = 0; i < s.length; i++) v.setUint8(o + i, s.charCodeAt(i)); };
      w(0, 'RIFF'); v.setUint32(4, 36 + n * 2, true); w(8, 'WAVE'); w(12, 'fmt '); v.setUint32(16, 16, true); v.setUint16(20, 1, true); v.setUint16(22, 1, true);
      v.setUint32(24, rate, true); v.setUint32(28, rate * 2, true); v.setUint16(32, 2, true); v.setUint16(34, 16, true); w(36, 'data'); v.setUint32(40, n * 2, true);
      for (let i = 0; i < n; i++) v.setInt16(44 + i * 2, Math.sin((i / rate) * 2 * Math.PI * 440) * 3000 * (1 - i / n), true);
      const url = URL.createObjectURL(new Blob([buf], { type: 'audio/wav' })), a = document.createElement('a');
      a.href = url; a.download = 'call-' + new Date(h.at).toISOString().slice(0, 16).replace(/[:T]/g, '-') + '.wav'; a.click(); setTimeout(() => URL.revokeObjectURL(url), 1000);
    }
  }
};
</script>
