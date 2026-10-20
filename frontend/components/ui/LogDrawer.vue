<!-- One audit log entry in full, on the right: who did what, when and from where, the alert it raised, what else that person
     did, and the proof that the entry is untouched. -->
<template>
<div v-if="id" class="hub-overlay hub-right-drawer" @click.self="close" style="position: fixed; inset: 0; z-index: 64; background: rgba(5,10,8,0.55); display: flex; justify-content: flex-end">
<aside role="dialog" aria-label="Log entry details" aria-modal="true" style="width: 480px; max-width: 100%; height: 100%; background: var(--k-16201C,#16201C); border-left: 1px solid var(--k-2A3832,#2A3832); box-shadow: -20px 0 50px rgba(0,0,0,0.4); display: flex; flex-direction: column">
<div style="flex: none; display: flex; align-items: flex-start; gap: 12px; padding: 18px 20px; border-bottom: 1px solid var(--k-24322C,#24322C)">
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6px">
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Log entry<template v-if="d && d.seq"> · #{{ d.seq }}</template></span>
<h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 20px; line-height: 1.3; overflow-wrap: anywhere">{{ d ? d.action : err ? 'Could not load this entry' : 'Loading…' }}</h2>
<div v-if="d" style="display: flex; flex-wrap: wrap; gap: 6px"><span :style="pill(resColor(d.result))">{{ d.result }}</span><span :style="pill('#7CC4FF')">{{ d.module }}</span><span v-if="d.rule" :style="pill('#FF9B8A')">Flagged</span></div>
</div>
<button type="button" aria-label="Close details" @click="close" style="flex: none; width: 40px; height: 40px; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">✕</button>
</div>
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 18px 20px; display: flex; flex-direction: column; gap: 18px">
<div v-if="err" role="alert" style="padding: 12px 14px; border-radius: 12px; background: var(--k-3A2420,#3A2420); color: #FFD3CA; font-size: 14px; line-height: 1.5; display: flex; flex-direction: column; gap: 10px">{{ err }}<button type="button" @click="load(id)" style="align-self: flex-start; height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600; cursor: pointer">Try again</button></div>
<div v-else-if="!d" role="status" style="font-size: 14px; color: var(--k-9DB0A6,#9DB0A6)">Loading the entry…</div>
<template v-if="d">
<dl style="margin: 0; display: grid; grid-template-columns: 110px minmax(0, 1fr); gap: 10px 12px; font-size: 14px">
<dt style="color: var(--k-9DB0A6,#9DB0A6)">Who</dt><dd style="margin: 0; overflow-wrap: anywhere"><strong>{{ d.user }}</strong></dd>
<dt style="color: var(--k-9DB0A6,#9DB0A6)">When</dt><dd style="margin: 0">{{ when(d.at) }}<span style="display: block; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ ago(d.at) }}</span></dd>
<dt style="color: var(--k-9DB0A6,#9DB0A6)">Address</dt><dd style="margin: 0; font-family: ui-monospace, monospace">{{ d.ip || '—' }}</dd>
<dt style="color: var(--k-9DB0A6,#9DB0A6)">Place</dt><dd style="margin: 0">{{ d.location || '—' }}</dd>
<dt style="color: var(--k-9DB0A6,#9DB0A6)">Module</dt><dd style="margin: 0">{{ d.module }}</dd>
</dl>

<section v-if="d.alert" style="padding: 14px; border-radius: 14px; background: var(--k-3A2420,#3A2420); border: 1px solid var(--k-5A3A33,#5A3A33); display: flex; flex-direction: column; gap: 6px">
<span style="font-size: 12px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; color: #FF9B8A">Raised an alert</span>
<strong style="font-size: 14px; overflow-wrap: anywhere">{{ d.alert.ref }} · {{ d.alert.title }}</strong>
<span style="font-size: 12px; color: var(--k-C9D6CF,#C9D6CF)">{{ d.alert.severity }} severity · {{ d.alert.status }}</span>
<button type="button" @click="openAlerts" style="align-self: flex-start; height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600; cursor: pointer">Open the alert</button>
</section>

<section style="display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 10px">
<div style="padding: 12px; border-radius: 12px; background: var(--k-121B17,#121B17); text-align: center"><strong style="font-size: 20px">{{ d.same_user }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">entries by them, 24 h</div></div>
<div :style="`padding: 12px; border-radius: 12px; background: var(--k-121B17,#121B17); text-align: center; ${d.failed_today ? 'box-shadow: inset 0 0 0 1px #FF9B8A66' : ''}`"><strong :style="`font-size: 20px; color: ${d.failed_today ? '#FF9B8A' : 'inherit'}`">{{ d.failed_today }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">failed or blocked</div></div>
<div style="padding: 12px; border-radius: 12px; background: var(--k-121B17,#121B17); text-align: center"><strong style="font-size: 20px">{{ d.same_ip }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">from this address</div></div>
</section>

<section v-if="d.related.length" style="display: flex; flex-direction: column; gap: 8px">
<strong style="font-size: 14px">What else {{ d.user }} did before this</strong>
<button v-for="r in d.related" :key="r.id" type="button" @click="show(r.id)" style="display: flex; flex-direction: column; gap: 2px; text-align: left; padding: 10px 12px; border-radius: 10px; border: 1px solid var(--k-24322C,#24322C); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer"><span style="font-size: 13px; overflow-wrap: anywhere">{{ r.action }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ ago(r.at) }} · <span :style="`color: ${resColor(r.result)}`">{{ r.result }}</span> · {{ r.module }}</span></button>
</section>

<section v-if="d.sealed !== undefined && d.sealed !== null" :style="`padding: 14px; border-radius: 14px; border: 1px solid ${d.sealed ? '#2C3D34' : '#FF9B8A'}; background: ${d.sealed ? 'var(--k-1E3326,#1E3326)' : 'var(--k-3A2420,#3A2420)'}; display: flex; flex-direction: column; gap: 6px`">
<strong :style="`font-size: 14px; color: ${d.sealed ? '#B8F36B' : '#FF9B8A'}`">{{ d.sealed ? '✓ Untouched' : '✕ This entry does not match its seal' }}</strong>
<span style="font-size: 12px; color: var(--k-C9D6CF,#C9D6CF); line-height: 1.5">{{ d.sealed ? 'It is entry #' + d.seq + ' in the tamper-proof log and still matches what was written.' : 'It was changed after it was written. Treat it as unreliable and check who had direct access to the database.' }}</span>
<span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6); font-family: ui-monospace, monospace; overflow-wrap: anywhere">seal {{ d.hash.slice(0, 24) }}…<br>after {{ d.prev_hash.slice(0, 24) }}…</span>
</section>
</template>
</div>
</aside>
</div>
</template>

<script>
import { useHubStore } from '~/composables/useHub';

export default {
  name: 'LogDrawer',
  data() { return { d: null, err: '', seen: '' }; },
  computed: {
    logic() { return useHubStore().logic; },
    id() { return this.logic.state.logOpen || ''; }
  },
  watch: { id: { immediate: true, handler(v) { if (v) this.load(v); else { this.d = null; this.err = ''; } } } },
  mounted() { this.key = (e) => { if (e.key === 'Escape' && this.id) this.close(); }; window.addEventListener('keydown', this.key); },
  beforeUnmount() { window.removeEventListener('keydown', this.key); },
  methods: {
    pill(c) { return `padding: 3px 10px; border-radius: 12px; font-size: 12px; font-weight: 700; background: ${c}22; color: ${c}`; },
    resColor(r) { return r === 'Success' ? '#9DB0A6' : r === 'Failed' ? '#FF9B8A' : '#FFB86B'; },
    when(iso) { return new Date(iso).toLocaleString('en-GB', { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit', timeZone: 'Africa/Nairobi' }); },
    ago(iso) { const m = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 60000)); return m < 1 ? 'Just now' : m < 60 ? m + ' min ago' : m < 1440 ? Math.round(m / 60) + ' h ago' : Math.round(m / 1440) + ' days ago'; },
    close() { this.logic.setState({ logOpen: '' }); },
    show(id) { this.logic.setState({ logOpen: id }); },
    openAlerts() { this.logic.setState({ logOpen: '', secTab: 'alerts' }); },
    async load(id) {
      this.d = null; this.err = '';
      const srv = this.logic.srv;
      if (!srv) { this.err = 'Details need a connection to the server.'; return; }
      try { const d = await srv.request('GET', '/security/logs/' + id); if (this.id === id) this.d = d; } catch (e) { if (this.id === id) this.err = e && e.status === 404 ? 'The server does not know this view yet. It is running an older version: restart the API, then try again.' : (e && e.message) || 'Could not load this entry.'; }
    }
  }
};
</script>
