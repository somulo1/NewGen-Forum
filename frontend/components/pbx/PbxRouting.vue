<!-- Calls management > Routing: the phone menu, ring groups, opening hours and holidays, and queue callbacks (sample until the PBX is connected) -->
<template>
<section class="pbx-flow">
  <p class="pbx-sub pbx-flow-intro">How a customer's call travels, in order. Open a step to change it; <strong>Call in as a customer</strong> (top of the page) lets you hear the result.</p>

  <!-- 1. hours -->
  <div :class="['pbx-step', { open: open === 'hours' }]">
    <button type="button" class="pbx-step-head" :aria-expanded="open === 'hours'" @click="toggle('hours')">
      <span class="pbx-step-n">1</span>
      <span class="grow"><strong>Are we open?</strong><span class="pbx-sub block">{{ hoursSummary }}</span></span>
      <span :class="['pbx-pill', openNow ? 'ok' : 'off']">{{ openNow ? 'Open now' : 'Closed now' }}</span>
    </button>
    <div v-if="open === 'hours'" class="pbx-step-body">
      <div class="pbx-hours-grid">
        <template v-for="d in days" :key="d.k">
          <label class="pbx-check"><input type="checkbox" :checked="!!r.hours[d.k]" @change="r.hours[d.k] = $event.target.checked ? ['08:00', '17:00'] : null; save()"> {{ d.l.slice(0, 3) }}</label>
          <span v-if="r.hours[d.k]" class="pbx-row nowrap"><input type="time" v-model="r.hours[d.k][0]" class="pbx-time" :aria-label="d.l + ' opens'" @change="save"> – <input type="time" v-model="r.hours[d.k][1]" class="pbx-time" :aria-label="d.l + ' closes'" @change="save"></span>
          <span v-else class="pbx-sub">Closed</span>
        </template>
      </div>
      <label class="pbx-field inline">When closed<select v-model="r.afterHours" @change="save"><option>Voicemail</option><option>After-hours message, then hang up</option><option>Fraud line (always open)</option></select></label>
      <div class="pbx-row"><span class="pbx-sub">Holidays (closed all day):</span>
        <span v-for="(h, i) in r.holidays" :key="h.date" class="pbx-chip">{{ h.name }} · {{ h.date.slice(5) }}<button type="button" class="pbx-link" :aria-label="'Remove ' + h.name" @click="r.holidays.splice(i, 1); save()">×</button></span>
      </div>
      <form class="pbx-row" @submit.prevent="addHoliday"><input type="date" v-model="hDate" class="pbx-time" aria-label="Holiday date"><input v-model="hName" class="pbx-search" placeholder="Holiday name" aria-label="Holiday name"><button type="submit" class="pbx-btn ghost small" :disabled="!hDate || !hName.trim()">Add holiday</button></form>
    </div>
  </div>

  <!-- 2. menu -->
  <div :class="['pbx-step', { open: open === 'menu' }]">
    <button type="button" class="pbx-step-head" :aria-expanded="open === 'menu'" @click="toggle('menu')">
      <span class="pbx-step-n">2</span>
      <span class="grow"><strong>Phone menu: the caller picks a key</strong><span class="pbx-sub block clip-line">{{ menuSummary }}</span></span>
    </button>
    <div v-if="open === 'menu'" class="pbx-step-body">
      <label class="pbx-field">What callers hear<textarea v-model="r.ivr.greeting" rows="2" @change="save"></textarea></label>
      <div class="pbx-keyrows">
        <div v-for="(o, i) in r.ivr.options" :key="o.key" class="pbx-keyrow">
          <span class="pbx-key">{{ o.key }}</span><span aria-hidden="true">→</span>
          <select v-model="o.to" class="pbx-search" :aria-label="'Key ' + o.key + ' goes to'" @change="save"><option v-for="d in destinations" :key="d">{{ d }}</option></select>
          <button type="button" class="pbx-link" :aria-label="'Remove key ' + o.key" @click="r.ivr.options.splice(i, 1); save()">×</button>
        </div>
      </div>
      <form class="pbx-row" @submit.prevent="addKey">
        <select v-model="newKey" class="pbx-time" aria-label="Key"><option value="" disabled>Key</option><option v-for="k in freeKeys" :key="k">{{ k }}</option></select>
        <select v-model="newTo" class="pbx-search" aria-label="Goes to"><option v-for="d in destinations" :key="d">{{ d }}</option></select>
        <button type="submit" class="pbx-btn ghost small" :disabled="!newKey">Add key</button>
      </form>
      <div class="pbx-row">
        <label class="pbx-field inline">Wait<input type="number" min="3" max="30" v-model.number="r.ivr.timeoutSec" @change="save"> s for a key</label>
        <label class="pbx-field inline">Repeat<input type="number" min="0" max="5" v-model.number="r.ivr.retries" @change="save"> times</label>
        <button type="button" class="pbx-btn ghost small" @click="speak(r.ivr.greeting)">Listen</button>
      </div>
    </div>
  </div>

  <!-- 3. who answers -->
  <div :class="['pbx-step', { open: open === 'groups' }]">
    <button type="button" class="pbx-step-head" :aria-expanded="open === 'groups'" @click="toggle('groups')">
      <span class="pbx-step-n">3</span>
      <span class="grow"><strong>Who answers: ring groups</strong><span class="pbx-sub block clip-line">{{ groupsSummary }}</span></span>
    </button>
    <div v-if="open === 'groups'" class="pbx-step-body">
      <div v-for="(g, gi) in r.groups" :key="gi" class="pbx-grouprow">
        <div class="pbx-row">
          <strong class="grow">{{ g.name }}</strong>
          <select v-model="g.strategy" class="pbx-time" :aria-label="g.name + ': who rings first'" @change="save"><option v-for="x in strategies" :key="x">{{ x }}</option></select>
          <label class="pbx-field inline"><input type="number" min="5" max="120" v-model.number="g.ringSec" @change="save" :aria-label="g.name + ': ring seconds'"> s each</label>
          <button type="button" class="pbx-link" :aria-label="'Delete ' + g.name" @click="r.groups.splice(gi, 1); save()">×</button>
        </div>
        <div class="pbx-tags"><button v-for="a in agents" :key="a.ext" type="button" :class="['pbx-tag', { on: g.members.includes(a.ext) }]" :aria-pressed="g.members.includes(a.ext)" @click="toggleMember(g, a.ext)">{{ a.name }}</button></div>
      </div>
      <form class="pbx-row" @submit.prevent="addGroup"><input v-model="groupName" class="pbx-search" placeholder="New group name" aria-label="New ring group name"><button type="submit" class="pbx-btn ghost small" :disabled="!groupName.trim()">Add group</button></form>
    </div>
  </div>

  <!-- 4. waiting -->
  <div :class="['pbx-step', { open: open === 'wait' }]">
    <button type="button" class="pbx-step-head" :aria-expanded="open === 'wait'" @click="toggle('wait')">
      <span class="pbx-step-n">4</span>
      <span class="grow"><strong>While they wait</strong><span class="pbx-sub block">{{ waitSummary }}</span></span>
      <span v-if="waiting" class="pbx-count">{{ waiting }}</span>
    </button>
    <div v-if="open === 'wait'" class="pbx-step-body">
      <div class="pbx-row">
        <label class="pbx-check"><input type="checkbox" v-model="r.announcePosition" @change="save"> Tell them their place every</label>
        <input type="number" min="15" max="300" class="pbx-time" v-model.number="r.announceEverySec" @change="save" :disabled="!r.announcePosition" aria-label="Seconds between announcements"> s
      </div>
      <label class="pbx-check"><input type="checkbox" v-model="r.callbacks" @change="save"> Offer a call back (press 1) so they don't have to wait</label>
      <label class="pbx-field inline">After<input type="number" min="30" max="1800" v-model.number="r.maxWaitSec" @change="save"> s of waiting, send to voicemail</label>
      <strong>Call-back requests</strong>
      <ul class="pbx-list">
        <li v-for="cb in pbx.callbacks" :key="cb.id"><span class="grow"><strong>{{ name(cb) }}</strong> <span class="pbx-sub">{{ cb.queue }} · {{ fmtWhen(cb.askedAt) }} · {{ cb.status }}</span></span>
          <button type="button" class="pbx-btn small go" @click="err = callBack(cb)">Call back</button><button type="button" class="pbx-link" @click="dropCallback(cb.id)">Done</button></li>
        <li v-if="!pbx.callbacks.length" class="pbx-sub">Nobody is waiting for a call back.</li>
      </ul>
      <span v-if="err" role="alert" class="pbx-err">{{ err }}</span>
    </div>
  </div>
</section>
</template>

<style>
.pbx-flow { display: flex; flex-direction: column; gap: 8px; max-width: 920px; }
.pbx-flow-intro { margin: 0 0 4px; }
.pbx-step { border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); }
.pbx-step.open { border-color: var(--k-3A5446,#3A5446); }
.pbx-step-head { width: 100%; display: flex; align-items: center; gap: 12px; padding: 10px 14px; border: 0; background: transparent; color: inherit; cursor: pointer; text-align: left; font: inherit; }
.pbx-step-head .grow { display: flex; flex-direction: column; min-width: 0; }
.pbx-step-n { width: 26px; height: 26px; flex: none; border-radius: 50%; background: var(--k-22302A,#22302A); color: #B8F36B; font-weight: 700; font-size: 13px; display: flex; align-items: center; justify-content: center; }
.pbx-step-body { display: flex; flex-direction: column; gap: 10px; padding: 0 14px 14px 52px; }
@media (max-width: 600px) { .pbx-step-body { padding-left: 14px; } }
.clip-line { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.pbx-hours-grid { display: grid; grid-template-columns: 80px 1fr; gap: 6px 10px; align-items: center; max-width: 360px; }
.pbx-time { box-sizing: border-box; height: 34px; padding: 0 8px; border-radius: 8px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font: inherit; font-size: 13px; }
.pbx-keyrows { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 260px), 1fr)); gap: 6px; }
.pbx-keyrow { display: flex; align-items: center; gap: 8px; }
.pbx-keyrow .pbx-search, .pbx-step form .pbx-search { flex: 1 1 140px; min-height: 34px; padding: 4px 10px; }
.pbx-grouprow { display: flex; flex-direction: column; gap: 6px; padding: 10px 0; border-top: 1px solid var(--k-24322C,#24322C); }
.pbx-step .pbx-field.inline input { width: 70px; height: 34px; min-height: 0; padding: 0 8px; }
</style>

<script>
import { pbx, saveRouting, callBack, dropCallback, customerOf, showNumber, fmtWhen } from '~/logic/pbx/store.js';
import { speak } from '~/logic/pbx/sounds.js';

const DAYS = [['mon', 'Monday'], ['tue', 'Tuesday'], ['wed', 'Wednesday'], ['thu', 'Thursday'], ['fri', 'Friday'], ['sat', 'Saturday'], ['sun', 'Sunday']].map(([k, l]) => ({ k, l }));

export default {
  name: 'PbxRouting',
  data() { return { pbx, open: 'menu', newKey: '', newTo: 'Customer care', groupName: '', hDate: '', hName: '', err: '', days: DAYS, strategies: ['Ring all', 'Longest idle first', 'In order', 'Fewest calls first', 'Random'] }; },
  computed: {
    r() { return this.pbx.routing; },
    agents() { return this.pbx.agents; },
    destinations() { return this.pbx.queues.map((q) => q.name).concat(this.pbx.routing.groups.map((g) => g.name), ['Voicemail', 'Repeat menu', 'Operator (1001)']); },
    freeKeys() { const used = this.r.ivr.options.map((o) => o.key); return ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '*', '#'].filter((k) => !used.includes(k)); },
    // one-line summaries, so each step says what it does without opening it
    hoursSummary() {
      const on = this.days.filter((d) => this.r.hours[d.k]);
      const wk = on.filter((d) => !['sat', 'sun'].includes(d.k)), h = wk[0] && this.r.hours[wk[0].k];
      return (h ? 'Weekdays ' + h[0] + '–' + h[1] : 'Closed on weekdays') + (this.r.hours.sat ? ' · Sat ' + this.r.hours.sat.join('–') : '') + (this.r.hours.sun ? ' · Sun ' + this.r.hours.sun.join('–') : '') + ' · closed: ' + this.r.afterHours.toLowerCase() + ' · ' + this.r.holidays.length + ' holidays';
    },
    menuSummary() { return this.r.ivr.options.map((o) => o.key + ' ' + o.to).join(' · ') || 'No keys yet: callers go straight to Customer care'; },
    groupsSummary() { return this.r.groups.map((g) => g.name + ' (' + g.members.length + ', ' + g.strategy.toLowerCase() + ')').join(' · ') || 'No groups yet'; },
    waitSummary() { return [this.r.announcePosition ? 'place in line every ' + this.r.announceEverySec + 's' : 'no place announcements', this.r.callbacks ? 'call back offered' : 'no call back', 'voicemail after ' + this.r.maxWaitSec + 's'].join(' · '); },
    waiting() { return this.pbx.callbacks.filter((c) => c.status === 'Waiting').length; },
    openNow() {
      const d = new Date(), key = ['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat'][d.getDay()], h = this.r.hours[key];
      if (!h || this.r.holidays.some((x) => x.date === d.toISOString().slice(0, 10))) return false;
      const t = String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0');
      return t >= h[0] && t < h[1];
    }
  },
  methods: {
    toggle(k) { this.open = this.open === k ? '' : k; },
    speak, callBack, dropCallback, fmtWhen,
    save: saveRouting,
    name(cb) { const c = cb.customerId && customerOf(cb.customerId); return c ? c.name : showNumber(cb.number); },
    addKey() { this.r.ivr.options.push({ key: this.newKey, to: this.newTo }); this.r.ivr.options.sort((a, b) => a.key.localeCompare(b.key)); this.newKey = ''; this.save(); },
    toggleMember(g, ext) { const i = g.members.indexOf(ext); if (i === -1) g.members.push(ext); else g.members.splice(i, 1); this.save(); },
    addGroup() { this.r.groups.push({ name: this.groupName.trim(), strategy: 'Ring all', members: [], ringSec: 20 }); this.groupName = ''; this.save(); },
    addHoliday() { this.r.holidays.push({ date: this.hDate, name: this.hName.trim() }); this.r.holidays.sort((a, b) => a.date.localeCompare(b.date)); this.hDate = ''; this.hName = ''; this.save(); }
  }
};
</script>
