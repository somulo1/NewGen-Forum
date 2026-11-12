<!-- Calls management > Campaigns: lists of customers to call, dialled one after another with retry rules, and the do-not-call list -->
<template>
<section class="pbx-mon">
  <div v-for="k in pbx.campaigns" :key="k.id" class="pbx-card">
    <div class="pbx-row between">
      <span><strong>{{ k.name }}</strong><span class="pbx-sub block">{{ k.queue }} · {{ k.list.length }} customers · up to {{ k.retries + 1 }} tries, {{ k.retryMin }} min apart · {{ k.window[0] }}–{{ k.window[1] }}</span></span>
      <span :class="['pbx-pill', { Running: 'ok', Paused: 'mid', Finished: 'off', Draft: 'off' }[k.status] || 'off']">{{ k.status }}</span>
    </div>
    <div class="pbx-progress" role="img" :aria-label="done(k) + ' of ' + k.list.length + ' reached'"><span :style="`width: ${pct(k)}%`"></span></div>
    <div class="pbx-row">
      <button type="button" class="pbx-btn go small" :disabled="!!pbx.call" @click="next(k)">{{ k.status === 'Running' ? 'Call the next customer' : 'Start' }}</button>
      <button v-if="k.status === 'Running'" type="button" class="pbx-btn ghost small" @click="k.status = 'Paused'">Pause</button>
      <label class="pbx-field inline">Tries<input type="number" min="1" max="6" :value="k.retries + 1" @change="k.retries = Math.max(0, +$event.target.value - 1)"></label>
      <label class="pbx-field inline">Gap (min)<input type="number" min="5" max="1440" v-model.number="k.retryMin"></label>
      <button type="button" class="pbx-link" @click="open = open === k.id ? null : k.id">{{ open === k.id ? 'Hide list' : 'Show list' }}</button>
    </div>
    <span v-if="msg[k.id]" role="status" class="pbx-sub">{{ msg[k.id] }}</span>
    <div v-if="open === k.id" class="pbx-table-wrap"><table class="pbx-table">
      <thead><tr><th>Customer</th><th>Tries</th><th>Last result</th><th>Set result</th></tr></thead>
      <tbody><tr v-for="r in k.list" :key="r.customerId">
        <td><strong>{{ cust(r).name }}</strong><span class="pbx-sub block">{{ cust(r).account }} · {{ cust(r).status }}</span><span v-if="isDnc(r)" class="pbx-pill off">Do not call</span></td>
        <td>{{ r.tries }}</td><td>{{ r.result || '—' }}</td>
        <td><select class="pbx-search" :value="r.result" :aria-label="'Result for ' + cust(r).name" @change="campaignResult(k, r.customerId, $event.target.value)"><option value="">—</option><option v-for="o in outcomes" :key="o">{{ o }}</option></select></td>
      </tr></tbody>
    </table></div>
  </div>

  <div class="pbx-grid2">
    <form class="pbx-card" @submit.prevent="create">
      <h2>New campaign</h2>
      <label class="pbx-field">Name<input v-model="name" placeholder="e.g. December arrears"></label>
      <label class="pbx-field">Queue<select v-model="queue"><option v-for="q in pbx.queues" :key="q.name">{{ q.name }}</option></select></label>
      <span class="pbx-sub">Customers to call</span>
      <div class="pbx-tags"><button v-for="c in pbx.customers" :key="c.id" type="button" :class="['pbx-tag', { on: picked.includes(c.id) }]" :aria-pressed="picked.includes(c.id)" @click="pick(c.id)">{{ c.name }}</button></div>
      <div class="pbx-row"><button type="button" class="pbx-link" @click="picked = pbx.customers.filter((c) => /arrears/.test(c.status)).map((c) => c.id)">Everyone in arrears</button><button type="button" class="pbx-link" @click="picked = pbx.customers.filter((c) => c.kyc !== 'Verified').map((c) => c.id)">KYC not finished</button></div>
      <button type="submit" class="pbx-btn go" :disabled="!name.trim() || !picked.length">Create campaign</button>
    </form>
    <div class="pbx-card">
      <h2>Do not call</h2>
      <span class="pbx-sub">Customers who asked not to be called. Campaigns skip them and the dial pad warns before calling.</span>
      <ul class="pbx-list"><li v-for="n in pbx.dnc" :key="n"><span class="grow">{{ label(n) }}</span><button type="button" class="pbx-link" @click="removeDnc(n)">Remove</button></li><li v-if="!pbx.dnc.length" class="pbx-empty">Nobody yet.</li></ul>
      <form class="pbx-row" @submit.prevent="addDnc(dncNo); dncNo = ''"><input v-model="dncNo" class="grow pbx-search" placeholder="Number" aria-label="Number to add to do not call"><button type="submit" class="pbx-btn ghost small" :disabled="!dncNo.trim()">Add</button></form>
    </div>
  </div>
</section>
</template>

<script>
import { pbx, campaignDial, campaignResult, newCampaign, addDnc, removeDnc, customerOf, customerByNumber, showNumber, norm } from '~/logic/pbx/store.js';

export default {
  name: 'PbxCampaigns',
  data() { return { pbx, open: null, msg: {}, name: '', queue: 'Collections', picked: [], dncNo: '', outcomes: ['No answer', 'Busy', 'Promise to pay', 'Payment made', 'Call back requested', 'Wrong number', 'Do not call', 'Resolved'] }; },
  methods: {
    campaignResult, addDnc, removeDnc,
    cust(r) { return customerOf(r.customerId) || { name: 'Customer', account: '', status: '' }; },
    isDnc(r) { return this.pbx.dnc.includes(norm(this.cust(r).phone)); },
    done(k) { return k.list.filter((r) => /Promise|Payment|Resolved|Do not call|Wrong/.test(r.result)).length; },
    pct(k) { return k.list.length ? Math.round((this.done(k) / k.list.length) * 100) : 0; },
    next(k) { const e = campaignDial(k); this.msg = { ...this.msg, [k.id]: e || 'Calling the next customer…' }; },
    pick(id) { const i = this.picked.indexOf(id); if (i === -1) this.picked.push(id); else this.picked.splice(i, 1); },
    create() { newCampaign(this.name.trim(), this.queue, this.picked); this.name = ''; this.picked = []; },
    label(n) { const c = customerByNumber(n); return (c ? c.name + ' · ' : '') + showNumber(n); }
  }
};
</script>
