<!-- Calls management > Routing: the phone menu, ring groups, opening hours and holidays, and queue callbacks (sample until the PBX is connected) -->
<template>
<section class="pbx-grid2">
  <div class="pbx-card">
    <h2>Phone menu</h2>
    <span class="pbx-sub">What callers hear and the key they press. Played with the sample voice until the phone system plays the real one.</span>
    <label class="pbx-field">Greeting<textarea v-model="r.ivr.greeting" rows="3" @change="save"></textarea></label>
    <div class="pbx-row"><button type="button" class="pbx-btn ghost small" @click="speak(r.ivr.greeting)">Listen</button></div>
    <ul class="pbx-list">
      <li v-for="(o, i) in r.ivr.options" :key="i">
        <span class="pbx-key">{{ o.key }}</span>
        <select v-model="o.to" class="grow pbx-search" :aria-label="'Key ' + o.key + ' goes to'" @change="save"><option v-for="d in destinations" :key="d">{{ d }}</option></select>
        <button type="button" class="pbx-link" @click="r.ivr.options.splice(i, 1); save()">Remove</button>
      </li>
    </ul>
    <form class="pbx-row" @submit.prevent="addKey">
      <select v-model="newKey" class="pbx-search" aria-label="Key"><option v-for="k in freeKeys" :key="k">{{ k }}</option></select>
      <select v-model="newTo" class="grow pbx-search" aria-label="Goes to"><option v-for="d in destinations" :key="d">{{ d }}</option></select>
      <button type="submit" class="pbx-btn ghost small" :disabled="!newKey">Add key</button>
    </form>
    <div class="pbx-row">
      <label class="pbx-field">Wait for a key (s)<input type="number" min="3" max="30" v-model.number="r.ivr.timeoutSec" @change="save"></label>
      <label class="pbx-field">Repeat the menu<input type="number" min="0" max="5" v-model.number="r.ivr.retries" @change="save"></label>
    </div>
  </div>

  <div class="pbx-card">
    <h2>Ring groups</h2>
    <div v-for="(g, gi) in r.groups" :key="gi" class="pbx-sub-card">
      <div class="pbx-row between"><strong>{{ g.name }}</strong><button type="button" class="pbx-link" @click="r.groups.splice(gi, 1); save()">Delete</button></div>
      <label class="pbx-field">Who rings first<select v-model="g.strategy" @change="save"><option v-for="x in strategies" :key="x">{{ x }}</option></select></label>
      <label class="pbx-field">Ring each for (s)<input type="number" min="5" max="120" v-model.number="g.ringSec" @change="save"></label>
      <div class="pbx-tags"><button v-for="a in agents" :key="a.ext" type="button" :class="['pbx-tag', { on: g.members.includes(a.ext) }]" :aria-pressed="g.members.includes(a.ext)" @click="toggleMember(g, a.ext)">{{ a.ext }} {{ a.name }}</button></div>
    </div>
    <form class="pbx-row" @submit.prevent="addGroup"><input v-model="groupName" class="grow pbx-search" placeholder="New ring group name" aria-label="New ring group name"><button type="submit" class="pbx-btn ghost small" :disabled="!groupName.trim()">Add group</button></form>
  </div>

  <div class="pbx-card">
    <h2>Opening hours</h2>
    <div v-for="d in days" :key="d.k" class="pbx-row">
      <label class="pbx-check" style="width: 110px"><input type="checkbox" :checked="!!r.hours[d.k]" @change="r.hours[d.k] = $event.target.checked ? ['08:00', '17:00'] : null; save()"> {{ d.l }}</label>
      <template v-if="r.hours[d.k]"><input type="time" v-model="r.hours[d.k][0]" class="pbx-search" :aria-label="d.l + ' opens'" @change="save"><span>to</span><input type="time" v-model="r.hours[d.k][1]" class="pbx-search" :aria-label="d.l + ' closes'" @change="save"></template>
      <span v-else class="pbx-sub">Closed</span>
    </div>
    <label class="pbx-field">When closed, calls go to<select v-model="r.afterHours" @change="save"><option>Voicemail</option><option>After-hours message, then hang up</option><option>Fraud line (always open)</option></select></label>
    <span :class="['pbx-pill', openNow ? 'ok' : 'off']">{{ openNow ? 'Open now' : 'Closed now' }}</span>
    <h2>Holidays</h2>
    <ul class="pbx-list"><li v-for="(h, i) in r.holidays" :key="h.date"><span class="grow">{{ h.name }}</span><span class="pbx-sub">{{ h.date }}</span><button type="button" class="pbx-link" @click="r.holidays.splice(i, 1); save()">Remove</button></li></ul>
    <form class="pbx-row" @submit.prevent="addHoliday"><input type="date" v-model="hDate" class="pbx-search" aria-label="Holiday date"><input v-model="hName" class="grow pbx-search" placeholder="Holiday name" aria-label="Holiday name"><button type="submit" class="pbx-btn ghost small" :disabled="!hDate || !hName.trim()">Add</button></form>
  </div>

  <div class="pbx-card">
    <h2>Waiting callers</h2>
    <label class="pbx-check"><input type="checkbox" v-model="r.announcePosition" @change="save"> Tell callers their place in the queue</label>
    <label class="pbx-field">Every (s)<input type="number" min="15" max="300" v-model.number="r.announceEverySec" @change="save" :disabled="!r.announcePosition"></label>
    <div class="pbx-row"><button type="button" class="pbx-btn ghost small" @click="speak('You are caller number 3 in the queue. Your expected wait is about 2 minutes. Press 1 at any time and we will call you back without losing your place.')">Listen to a sample</button></div>
    <label class="pbx-check"><input type="checkbox" v-model="r.callbacks" @change="save"> Offer a call back instead of waiting (press 1)</label>
    <label class="pbx-field">Longest wait before voicemail (s)<input type="number" min="30" max="1800" v-model.number="r.maxWaitSec" @change="save"></label>
    <h2>Call-back requests</h2>
    <ul class="pbx-list">
      <li v-for="cb in pbx.callbacks" :key="cb.id"><span class="grow"><strong>{{ name(cb) }}</strong><span class="pbx-sub block">{{ cb.queue }} · asked {{ fmtWhen(cb.askedAt) }} · {{ cb.status }}</span></span>
        <button type="button" class="pbx-btn small go" @click="err = callBack(cb)">Call back</button><button type="button" class="pbx-link" @click="dropCallback(cb.id)">Done</button></li>
      <li v-if="!pbx.callbacks.length" class="pbx-empty">Nobody is waiting for a call back.</li>
    </ul>
    <span v-if="err" role="alert" class="pbx-err">{{ err }}</span>
  </div>
</section>
</template>

<script>
import { pbx, saveRouting, callBack, dropCallback, customerOf, showNumber, fmtWhen } from '~/logic/pbx/store.js';
import { speak } from '~/logic/pbx/sounds.js';

const DAYS = [['mon', 'Monday'], ['tue', 'Tuesday'], ['wed', 'Wednesday'], ['thu', 'Thursday'], ['fri', 'Friday'], ['sat', 'Saturday'], ['sun', 'Sunday']].map(([k, l]) => ({ k, l }));

export default {
  name: 'PbxRouting',
  data() { return { pbx, newKey: '', newTo: 'Customer care', groupName: '', hDate: '', hName: '', err: '', days: DAYS, strategies: ['Ring all', 'Longest idle first', 'In order', 'Fewest calls first', 'Random'] }; },
  computed: {
    r() { return this.pbx.routing; },
    agents() { return this.pbx.agents; },
    destinations() { return this.pbx.queues.map((q) => q.name).concat(this.pbx.routing.groups.map((g) => g.name), ['Voicemail', 'Repeat menu', 'Operator (1001)']); },
    freeKeys() { const used = this.r.ivr.options.map((o) => o.key); return ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '*', '#'].filter((k) => !used.includes(k)); },
    openNow() {
      const d = new Date(), key = ['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat'][d.getDay()], h = this.r.hours[key];
      if (!h || this.r.holidays.some((x) => x.date === d.toISOString().slice(0, 10))) return false;
      const t = String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0');
      return t >= h[0] && t < h[1];
    }
  },
  methods: {
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
