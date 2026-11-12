<!-- Calls management: the customer phone system (softphone, call history, customer messaging, voicemail, queue monitoring).
     Front end only for now: calls are simulated in logic/pbx/store.js until the PBX is connected. -->
<template>
<div class="pbx">
  <header class="pbx-head">
    <div class="pbx-title">
      <h1>Calls management</h1>
      <span class="pbx-sub">Extension {{ pbx.settings.extension }} · Caller ID {{ pbx.settings.callerId }}</span>
    </div>
    <span class="pbx-demo" title="No calls leave this browser until the phone system is connected">Simulation · not connected to the PBX</span>
    <label class="pbx-presence">
      <span :class="['pbx-dot', presenceTone]" aria-hidden="true"></span>
      <span class="pbx-vh">My status</span>
      <select :value="pbx.presence" @change="setPresence($event.target.value)">
        <option v-for="p in presences" :key="p" :value="p">{{ p }}</option>
      </select>
    </label>
    <button type="button" class="pbx-btn ghost" @click="simulateIncoming()" title="Ring this phone with a sample customer call">Simulate inbound call</button>
  </header>

  <!-- ringing -->
  <div v-if="pbx.incoming" class="pbx-ring" role="alertdialog" aria-label="Incoming call">
    <div class="pbx-ring-who">
      <span class="pbx-ring-pulse" aria-hidden="true"></span>
      <div>
        <strong>{{ incomingCustomer ? incomingCustomer.name : 'Unknown caller' }}</strong>
        <span>{{ showNumber(pbx.incoming.number) }} · {{ pbx.incoming.queue }} queue · waited {{ pbx.incoming.waited }}s</span>
        <span v-if="incomingCustomer" class="pbx-ring-crm">{{ incomingCustomer.account }} · {{ incomingCustomer.product }} · {{ incomingCustomer.status }} · KYC {{ incomingCustomer.kyc }}</span>
      </div>
    </div>
    <div class="pbx-row">
      <button type="button" class="pbx-btn go" @click="answer">Answer</button>
      <button type="button" class="pbx-btn ghost" @click="decline(true)">To voicemail</button>
      <button type="button" class="pbx-btn stop" @click="decline(false)">Decline</button>
    </div>
  </div>

  <!-- live call -->
  <section v-if="pbx.call" class="pbx-live" aria-label="Current call">
    <div class="pbx-live-top">
      <div class="pbx-live-who">
        <UserAvatar :name="liveName" :initials="initials(liveName)" color="#7CC4FF" :size="44" :font-size="15"/>
        <div>
          <strong>{{ liveName }}</strong>
          <span>{{ showNumber(pbx.call.number, revealed[pbx.call.number]) }} <button v-if="pbx.settings.maskNumbers && !revealed[pbx.call.number]" type="button" class="pbx-link" @click="reveal(pbx.call.number)">Show number</button></span>
          <span class="pbx-state">{{ pbx.call.dir === 'in' ? 'Inbound' : 'Outbound' }} · {{ callState }}<template v-if="pbx.call.queue"> · {{ pbx.call.queue }}</template><template v-if="pbx.call.recording"> · <span class="pbx-rec">● {{ pbx.call.recPaused ? 'Recording paused' : 'Recording' }}</span></template></span>
        </div>
      </div>
      <span class="pbx-timer" aria-live="off">{{ liveTimer }}</span>
    </div>

    <div v-if="liveCustomer" class="pbx-crm">
      <div><span>Account</span><strong>{{ liveCustomer.account }}</strong></div>
      <div><span>Product</span><strong>{{ liveCustomer.product }}</strong></div>
      <div><span>Balance</span><strong>{{ money(liveCustomer.balance) }}</strong></div>
      <div><span>Status</span><strong :class="{ warn: /arrears|Frozen/.test(liveCustomer.status) }">{{ liveCustomer.status }}</strong></div>
      <div><span>KYC</span><strong :class="{ warn: liveCustomer.kyc !== 'Verified' }">{{ liveCustomer.kyc }}</strong></div>
    </div>

    <div v-if="pbx.call.transferMode === 'consulting'" class="pbx-note">
      Talking to <strong>{{ pbx.call.transferTo }}</strong> while the customer waits on hold.
      <div class="pbx-row"><button type="button" class="pbx-btn go" @click="completeTransfer">Complete transfer</button><button type="button" class="pbx-btn ghost" @click="cancelTransfer">Back to the customer</button></div>
    </div>

    <div class="pbx-controls" role="toolbar" aria-label="Call controls">
      <button type="button" :class="['pbx-ctl', { on: pbx.call.muted }]" :aria-pressed="pbx.call.muted" @click="pbx.call.muted = !pbx.call.muted">{{ pbx.call.muted ? 'Unmute' : 'Mute' }}</button>
      <button type="button" :class="['pbx-ctl', { on: pbx.call.onHold }]" :aria-pressed="pbx.call.onHold" @click="toggleHold">{{ pbx.call.onHold ? 'Resume' : 'Hold' }}</button>
      <button type="button" :class="['pbx-ctl', { on: panel === 'keys' }]" @click="togglePanel('keys')">Keypad</button>
      <button type="button" :class="['pbx-ctl', { on: panel === 'transfer' }]" @click="togglePanel('transfer')">Transfer</button>
      <button type="button" :class="['pbx-ctl', { on: panel === 'conf' }]" @click="togglePanel('conf')">Add caller</button>
      <button type="button" :class="['pbx-ctl', { on: pbx.call.recording && !pbx.call.recPaused }]" @click="toggleRecord">{{ !pbx.call.recording ? 'Record' : pbx.call.recPaused ? 'Resume recording' : 'Pause recording' }}</button>
      <button type="button" class="pbx-ctl" @click="park">Park</button>
      <button type="button" class="pbx-ctl end" @click="endCall()">End call</button>
    </div>

    <div v-if="panel === 'keys'" class="pbx-subpanel">
      <div class="pbx-keys small"><button v-for="k in keys" :key="k[0]" type="button" @click="sendDtmf(k[0])"><strong>{{ k[0] }}</strong></button></div>
      <span class="pbx-sub">Tones sent: {{ pbx.call.dtmf || 'none yet' }}</span>
    </div>
    <div v-if="panel === 'transfer'" class="pbx-subpanel">
      <label class="pbx-field">Transfer to (extension, queue or number)
        <input v-model="transferTarget" list="pbx-targets" placeholder="e.g. 1003 or Collections">
      </label>
      <datalist id="pbx-targets"><option v-for="t in targets" :key="t" :value="t"></option></datalist>
      <div class="pbx-row"><button type="button" class="pbx-btn go" :disabled="!transferTarget" @click="doTransfer('attended')">Talk first, then transfer</button><button type="button" class="pbx-btn ghost" :disabled="!transferTarget" @click="doTransfer('blind')">Transfer now</button></div>
    </div>
    <div v-if="panel === 'conf'" class="pbx-subpanel">
      <label class="pbx-field">Add someone to this call (up to 5)
        <input v-model="confTarget" list="pbx-targets" placeholder="Extension, colleague or number">
      </label>
      <div class="pbx-row"><button type="button" class="pbx-btn go" :disabled="!confTarget" @click="addToConference(confTarget); confTarget = ''">Add to call</button></div>
      <ul v-if="pbx.call.conference.length" class="pbx-list">
        <li v-for="(p, i) in pbx.call.conference" :key="i"><span>{{ p.who }}</span><button type="button" class="pbx-link" @click="dropFromConference(i)">Remove</button></li>
      </ul>
    </div>
    <label class="pbx-field">Call notes
      <textarea v-model="pbx.call.note" rows="2" placeholder="What was agreed. Saved with the call."></textarea>
    </label>
  </section>

  <!-- other calls held or parked -->
  <div v-if="pbx.held.length || pbx.parked.length" class="pbx-strip">
    <div v-for="h in pbx.held" :key="h.id" class="pbx-chip"><span>On hold · {{ nameFor(h) }}</span><button type="button" class="pbx-link" @click="swapTo(h.id)">Switch to</button></div>
    <div v-for="p in pbx.parked" :key="p.slot" class="pbx-chip"><span>Parked on {{ p.slot }} · {{ nameFor(p.call) }}</span><button type="button" class="pbx-link" @click="unpark(p.slot)">Pick up</button></div>
  </div>

  <!-- after-call work -->
  <section v-if="pbx.wrap" class="pbx-wrap" aria-label="After-call work">
    <div class="pbx-row between"><strong>Wrap up the call</strong><span class="pbx-sub">Closes by itself in {{ pbx.wrap.left }}s</span></div>
    <div class="pbx-tags"><button v-for="d in pbx.dispositions" :key="d" type="button" :class="['pbx-tag', { on: pbx.wrap.disposition === d }]" @click="pbx.wrap.disposition = d">{{ d }}</button></div>
    <textarea v-model="pbx.wrap.note" rows="2" placeholder="Notes for the customer record"></textarea>
    <div class="pbx-row"><button type="button" class="pbx-btn go" @click="finishWrap">Save and get ready</button></div>
  </section>

  <nav class="pbx-tabs" role="tablist" aria-label="Calls management">
    <button v-for="t in tabs" :key="t.k" type="button" role="tab" :aria-selected="tab === t.k" :class="{ on: tab === t.k }" @click="tab = t.k">{{ t.l }}<span v-if="t.n" class="pbx-count">{{ t.n }}</span></button>
  </nav>

  <!-- dial pad -->
  <section v-if="tab === 'dial'" class="pbx-grid2">
    <div class="pbx-card">
      <input class="pbx-number" v-model="number" inputmode="tel" placeholder="Number or customer name" aria-label="Number to call" @keydown.enter="callNumber">
      <ul v-if="numberMatches.length" class="pbx-list">
        <li v-for="c in numberMatches" :key="c.id"><button type="button" class="pbx-link" @click="number = c.phone">{{ c.name }}</button><span class="pbx-sub">{{ showNumber(c.phone) }}</span></li>
      </ul>
      <div class="pbx-keys"><button v-for="k in keys" :key="k[0]" type="button" @click="number += k[0]"><strong>{{ k[0] }}</strong><small>{{ k[1] }}</small></button></div>
      <div class="pbx-row center">
        <button type="button" class="pbx-btn ghost" @click="number = number.slice(0, -1)" aria-label="Delete last digit">⌫</button>
        <button type="button" class="pbx-btn go big" @click="callNumber">Call</button>
      </div>
      <label class="pbx-field">Call out as
        <select v-model="pbx.settings.callerId" @change="saveSettings"><option v-for="c in pbx.settings.callerIds" :key="c">{{ c }}</option></select>
      </label>
      <span v-if="dialErr" role="alert" class="pbx-err">{{ dialErr }}</span>
    </div>
    <div class="pbx-card">
      <h2>Recent</h2>
      <ul class="pbx-list">
        <li v-for="h in pbx.history.slice(0, 8)" :key="h.id">
          <span :class="['pbx-dir', dirTone(h)]">{{ dirIcon(h) }}</span>
          <span class="grow"><strong>{{ nameFor(h) }}</strong><span class="pbx-sub">{{ h.result }} · {{ fmtWhen(h.at) }}</span></span>
          <button type="button" class="pbx-btn small go" @click="callTo(h.number)">Call</button>
        </li>
      </ul>
    </div>
  </section>

  <!-- history -->
  <section v-if="tab === 'history'" class="pbx-card">
    <div class="pbx-row wrap">
      <input v-model="hq" type="search" placeholder="Search name, number or note" aria-label="Search call history" class="pbx-search">
      <div class="pbx-seg" role="radiogroup" aria-label="Show"><button v-for="f in ['All', 'Inbound', 'Outbound', 'Missed']" :key="f" type="button" role="radio" :aria-checked="hf === f" :class="{ on: hf === f }" @click="hf = f">{{ f }}</button></div>
      <button type="button" class="pbx-btn ghost small" @click="exportCsv">Export CSV</button>
    </div>
    <div class="pbx-table-wrap">
      <table class="pbx-table">
        <thead><tr><th></th><th>Customer</th><th>Result</th><th>Outcome</th><th>Duration</th><th>When</th><th>Recording</th><th></th></tr></thead>
        <tbody>
          <tr v-for="h in historyRows" :key="h.id">
            <td><span :class="['pbx-dir', dirTone(h)]">{{ dirIcon(h) }}</span></td>
            <td><strong>{{ nameFor(h) }}</strong><span class="pbx-sub block">{{ showNumber(h.number) }}</span><span v-if="h.note" class="pbx-sub block clip" :title="h.note">{{ h.note }}</span></td>
            <td>{{ h.result }}</td>
            <td>{{ h.disposition || '—' }}</td>
            <td>{{ h.duration ? fmtDur(h.duration) : '—' }}</td>
            <td class="nowrap">{{ fmtWhen(h.at) }}</td>
            <td><button v-if="h.recorded" type="button" class="pbx-link" @click="playing = playing === h.id ? null : h.id">{{ playing === h.id ? 'Stop' : 'Play' }}</button><span v-else>—</span></td>
            <td><button type="button" class="pbx-btn small go" @click="callTo(h.number)">Call back</button></td>
          </tr>
          <tr v-if="!historyRows.length"><td colspan="8" class="pbx-empty">No calls match.</td></tr>
        </tbody>
      </table>
    </div>
    <p v-if="playing" class="pbx-sub">Playing the recording (sample). PBX: recordings stream from the phone system's storage.</p>
  </section>

  <!-- customers -->
  <section v-if="tab === 'contacts'" class="pbx-card">
    <input v-model="cq" type="search" placeholder="Search customers by name, account or number" aria-label="Search customers" class="pbx-search">
    <div class="pbx-cards">
      <article v-for="c in customerRows" :key="c.id" class="pbx-cust">
        <div class="pbx-row"><UserAvatar :name="c.name" :initials="initials(c.name)" color="#C9B6FF" :size="36" :font-size="12"/><span class="grow"><strong>{{ c.name }}</strong><span class="pbx-sub block">{{ c.account }} · {{ c.product }}</span></span></div>
        <span class="pbx-sub">{{ showNumber(c.phone, revealed[c.phone]) }} <button v-if="pbx.settings.maskNumbers && !revealed[c.phone]" type="button" class="pbx-link" @click="reveal(c.phone)">Show</button></span>
        <span class="pbx-sub">{{ c.status }} · {{ money(c.balance) }} · KYC {{ c.kyc }}</span>
        <div v-if="c.tags.length" class="pbx-tags"><span v-for="t in c.tags" :key="t" class="pbx-tag static">{{ t }}</span></div>
        <div class="pbx-row"><button type="button" class="pbx-btn small go" @click="callTo(c.phone)">Call</button><button type="button" class="pbx-btn small ghost" @click="messageTo(c)">Message</button></div>
      </article>
    </div>
  </section>

  <!-- messages -->
  <section v-if="tab === 'messages'" class="pbx-msgs">
    <aside class="pbx-card pbx-threads">
      <button v-for="t in pbx.threads" :key="t.id" type="button" :class="['pbx-thread', { on: thread === t.id }]" @click="openThread(t)">
        <span class="grow"><strong>{{ (customerOf(t.customerId) || {}).name || 'Customer' }}</strong><span class="pbx-sub block clip">{{ t.msgs.length ? t.msgs[t.msgs.length - 1].text : 'No messages yet' }}</span></span>
        <span class="pbx-tag static">{{ t.channel }}</span><span v-if="t.unread" class="pbx-count">{{ t.unread }}</span>
      </button>
    </aside>
    <div class="pbx-card pbx-convo">
      <template v-if="curThread">
        <div class="pbx-row between"><strong>{{ (customerOf(curThread.customerId) || {}).name }}</strong>
          <div class="pbx-seg" role="radiogroup" aria-label="Channel"><button v-for="ch in ['SMS', 'WhatsApp']" :key="ch" type="button" role="radio" :aria-checked="curThread.channel === ch" :class="{ on: curThread.channel === ch }" @click="curThread.channel = ch">{{ ch }}</button></div>
        </div>
        <div class="pbx-bubbles">
          <div v-for="(m, i) in curThread.msgs" :key="i" :class="['pbx-bubble', m.from === 'me' ? 'me' : 'them']"><span>{{ m.text }}</span><small>{{ fmtWhen(m.at) }}<template v-if="m.from === 'me'"> · {{ m.status || 'Delivered' }}</template></small></div>
        </div>
        <div class="pbx-tags"><button v-for="q in templates" :key="q.l" type="button" class="pbx-tag" @click="draft = q.t(customerOf(curThread.customerId))">{{ q.l }}</button></div>
        <form class="pbx-row" @submit.prevent="send"><input v-model="draft" class="grow pbx-search" placeholder="Write a message" aria-label="Message"><button type="submit" class="pbx-btn go" :disabled="!draft.trim()">Send</button></form>
        <span class="pbx-sub">{{ draft.length }} characters · {{ Math.max(1, Math.ceil(draft.length / 160)) }} SMS</span>
      </template>
      <p v-else class="pbx-empty">Choose a conversation, or message a customer from Customers.</p>
    </div>
  </section>

  <!-- voicemail -->
  <section v-if="tab === 'voicemail'" class="pbx-card">
    <ul class="pbx-list">
      <li v-for="v in pbx.voicemail" :key="v.id" class="col">
        <div class="pbx-row between"><span><strong>{{ v.customerId ? customerOf(v.customerId).name : showNumber(v.number) }}</strong><span v-if="!v.heard" class="pbx-tag static">New</span></span><span class="pbx-sub">{{ fmtWhen(v.at) }} · {{ fmtDur(v.duration) }}</span></div>
        <p class="pbx-transcript">"{{ v.transcript }}"</p>
        <div class="pbx-row"><button type="button" class="pbx-btn small ghost" @click="markHeard(v)">{{ v.heard ? 'Play again' : 'Play' }}</button><button type="button" class="pbx-btn small go" @click="callTo(v.number)">Call back</button><button type="button" class="pbx-btn small stop" @click="deleteVoicemail(v.id)">Delete</button></div>
      </li>
      <li v-if="!pbx.voicemail.length" class="pbx-empty">No voicemail.</li>
    </ul>
    <p class="pbx-sub">Greeting: {{ pbx.settings.greeting }}. Change it in Settings.</p>
  </section>

  <!-- monitoring / wallboard -->
  <section v-if="tab === 'monitor'" class="pbx-mon">
    <div class="pbx-kpis">
      <div class="pbx-kpi"><span>Waiting now</span><strong>{{ totals.waiting }}</strong></div>
      <div class="pbx-kpi"><span>Longest wait</span><strong>{{ totals.longest }}s</strong></div>
      <div class="pbx-kpi"><span>Answered today</span><strong>{{ totals.answered }}</strong></div>
      <div class="pbx-kpi"><span>Abandoned</span><strong>{{ totals.abandoned }}</strong></div>
      <div class="pbx-kpi"><span>Agents available</span><strong>{{ totals.free }} / {{ pbx.agents.length }}</strong></div>
    </div>
    <div class="pbx-card">
      <h2>Queues</h2>
      <div class="pbx-table-wrap"><table class="pbx-table">
        <thead><tr><th>Queue</th><th>Waiting</th><th>Longest</th><th>Answered</th><th>Abandoned</th><th>Service level</th><th>Avg wait</th><th>Avg talk</th><th>I take calls</th></tr></thead>
        <tbody><tr v-for="q in pbx.queues" :key="q.name">
          <td><strong>{{ q.name }}</strong><span class="pbx-sub block">{{ q.number }}</span></td>
          <td :class="{ warn: q.waiting > 2 }">{{ q.waiting }}</td><td :class="{ warn: q.longest > q.target * 3 }">{{ q.longest }}s</td><td>{{ q.answered }}</td><td>{{ q.abandoned }}</td>
          <td><span :class="['pbx-sla', q.sla >= 90 ? 'ok' : q.sla >= 80 ? 'mid' : 'bad']">{{ q.sla }}%</span><span class="pbx-sub block">answered within {{ q.target }}s</span></td>
          <td>{{ q.avgWait }}s</td><td>{{ fmtDur(q.avgTalk * 1000) }}</td>
          <td><label class="pbx-switch"><input type="checkbox" :checked="!!pbx.settings.queues[q.name]" @change="setAgentQueue(q.name, $event.target.checked)"><span class="pbx-vh">Take {{ q.name }} calls</span></label></td>
        </tr></tbody>
      </table></div>
    </div>
    <div class="pbx-card">
      <h2>Agents</h2>
      <div class="pbx-table-wrap"><table class="pbx-table">
        <thead><tr><th>Agent</th><th>Status</th><th>For</th><th>Queues</th><th>Calls today</th><th>Talk time</th><th>Supervise</th></tr></thead>
        <tbody><tr v-for="a in pbx.agents" :key="a.ext">
          <td><strong>{{ a.name }}</strong><span class="pbx-sub block">Ext {{ a.ext }}</span></td>
          <td><span :class="['pbx-pill', agentTone(a.status)]">{{ a.status }}</span></td>
          <td>{{ a.since ? fmtDur(pbx.tick - a.since) : '—' }}</td>
          <td class="pbx-sub">{{ a.queues.join(', ') }}</td>
          <td>{{ a.calls }}</td><td>{{ fmtDur(a.talk) }}</td>
          <td><div v-if="a.status === 'On a call' && !a.me" class="pbx-row"><button v-for="h in ['Listen', 'Whisper', 'Barge in']" :key="h" type="button" class="pbx-btn small ghost" @click="superviseNow(a, h)">{{ h }}</button></div><span v-else class="pbx-sub">—</span></td>
        </tr></tbody>
      </table></div>
      <p v-if="supNote" class="pbx-note" role="status">{{ supNote }}</p>
    </div>
    <div class="pbx-card">
      <h2>Compliance log</h2>
      <ul class="pbx-list"><li v-for="(a, i) in pbx.audit" :key="i"><span class="grow">{{ a.what }}</span><span class="pbx-sub">{{ fmtWhen(a.at) }}</span></li><li v-if="!pbx.audit.length" class="pbx-empty">Nothing yet. Showing a customer's number and listening in are recorded here.</li></ul>
    </div>
  </section>

  <!-- settings -->
  <section v-if="tab === 'settings'" class="pbx-grid2">
    <div class="pbx-card">
      <h2>Calls to me</h2>
      <label class="pbx-check"><input type="checkbox" v-model="pbx.settings.dnd" @change="saveSettings"> Do not disturb (calls go to voicemail)</label>
      <label class="pbx-check"><input type="checkbox" v-model="pbx.settings.autoAnswer" @change="saveSettings"> Answer queue calls automatically</label>
      <label class="pbx-field">Forward every call to<input v-model="pbx.settings.forwardAlways" @change="saveSettings" placeholder="Leave empty for no forwarding"></label>
      <label class="pbx-field">When I am busy, forward to<input v-model="pbx.settings.forwardBusy" @change="saveSettings" placeholder="Voicemail"></label>
      <label class="pbx-field">When I don't answer, forward to<input v-model="pbx.settings.forwardNoAnswer" @change="saveSettings" placeholder="Voicemail"></label>
      <label class="pbx-field">Ring me for (seconds)<input type="number" min="5" max="120" v-model.number="pbx.settings.noAnswerSec" @change="saveSettings"></label>
      <label class="pbx-field">Ring on<select v-model="pbx.settings.ringDevice" @change="saveSettings"><option>Browser</option><option>Desk phone</option><option>Mobile app</option><option>All at once</option></select></label>
    </div>
    <div class="pbx-card">
      <h2>Calls and privacy</h2>
      <label class="pbx-check"><input type="checkbox" v-model="pbx.settings.recordAll" @change="saveSettings"> Record every call (customers hear the recording notice)</label>
      <label class="pbx-check"><input type="checkbox" v-model="pbx.settings.maskNumbers" @change="saveSettings"> Hide customer numbers until I choose to show one</label>
      <label class="pbx-field">After-call work time (seconds)<input type="number" min="0" max="300" v-model.number="pbx.settings.wrapUpSec" @change="saveSettings"></label>
      <label class="pbx-field">Voicemail greeting<select v-model="pbx.settings.greeting" @change="saveSettings"><option>Default greeting</option><option>Out of office</option><option>After hours</option></select></label>
      <h2>Blocked numbers</h2>
      <ul class="pbx-list"><li v-for="(b, i) in pbx.settings.blocked" :key="b"><span class="grow">{{ showNumber(b) }}</span><button type="button" class="pbx-link" @click="pbx.settings.blocked.splice(i, 1); saveSettings()">Unblock</button></li></ul>
      <form class="pbx-row" @submit.prevent="block"><input v-model="blockNo" class="grow pbx-search" placeholder="Number to block" aria-label="Number to block"><button type="submit" class="pbx-btn ghost">Block</button></form>
    </div>
  </section>
</div>
</template>

<script>
import { pbx, dial, simulateIncoming, answer, decline, endCall, toggleHold, swapTo, sendDtmf, transfer, completeTransfer, cancelTransfer, addToConference, dropFromConference, park, unpark, finishWrap,
  sendMessage, startThread, supervise, setAgentQueue, setPresence, saveSettings, markHeard, deleteVoicemail, startClock, stopClock, customerOf, customerByNumber, showNumber, fmtDur, fmtWhen, money, norm, logAudit } from '~/logic/pbx/store.js';

const KEYS = [['1', ''], ['2', 'ABC'], ['3', 'DEF'], ['4', 'GHI'], ['5', 'JKL'], ['6', 'MNO'], ['7', 'PQRS'], ['8', 'TUV'], ['9', 'WXYZ'], ['*', ''], ['0', '+'], ['#', '']];

export default {
  name: 'CallsManagementSection',
  data() {
    return { pbx, tab: 'dial', number: '', dialErr: '', panel: '', transferTarget: '', confTarget: '', hq: '', hf: 'All', cq: '', thread: null, draft: '', playing: null, revealed: {}, supNote: '', blockNo: '',
      keys: KEYS, presences: ['Available', 'Away', 'Busy', 'Lunch', 'Do not disturb', 'Off work'],
      templates: [
        { l: 'Payment reminder', t: (c) => 'Hello ' + (c ? c.name.split(' ')[0] : '') + ', a reminder that your payment on ' + (c ? c.account : 'your account') + ' is due. Reply PAY for the paybill details.' },
        { l: 'Missed your call', t: (c) => 'Hello ' + (c ? c.name.split(' ')[0] : '') + ', we tried to reach you. Please call us back on 0709 000 000.' },
        { l: 'KYC documents', t: () => 'Please upload a clear photo of your ID and a selfie in the app to finish verification.' }
      ] };
  },
  mounted() { startClock(); },
  beforeUnmount() { stopClock(); },
  computed: {
    tabs() {
      const missed = this.pbx.history.filter((h) => /Missed/.test(h.result) && Date.now() - h.at < 86400000).length;
      return [{ k: 'dial', l: 'Dial pad' }, { k: 'history', l: 'Call history', n: missed }, { k: 'contacts', l: 'Customers' }, { k: 'messages', l: 'Messages', n: this.pbx.threads.reduce((a, t) => a + t.unread, 0) },
        { k: 'voicemail', l: 'Voicemail', n: this.pbx.voicemail.filter((v) => !v.heard).length }, { k: 'monitor', l: 'Monitoring' }, { k: 'settings', l: 'Settings' }];
    },
    incomingCustomer() { return this.pbx.incoming && this.pbx.incoming.customerId ? customerOf(this.pbx.incoming.customerId) : null; },
    liveCustomer() { return this.pbx.call && this.pbx.call.customerId ? customerOf(this.pbx.call.customerId) : null; },
    liveName() { return this.liveCustomer ? this.liveCustomer.name : 'Unknown number'; },
    callState() { const c = this.pbx.call; return c.onHold ? 'On hold' : c.state; },
    liveTimer() { const c = this.pbx.call; void this.pbx.tick; return c.answeredAt ? fmtDur(Date.now() - c.answeredAt) : '…'; },
    targets() { return this.pbx.agents.filter((a) => !a.me).map((a) => a.ext + ' ' + a.name).concat(this.pbx.queues.map((q) => q.name)); },
    numberMatches() {
      const q = this.number.trim().toLowerCase(); if (q.length < 2 || /^[\d+*#\s]+$/.test(q)) return [];
      return this.pbx.customers.filter((c) => c.name.toLowerCase().includes(q) || c.account.toLowerCase().includes(q)).slice(0, 5);
    },
    historyRows() {
      const q = this.hq.trim().toLowerCase();
      return this.pbx.history.filter((h) => (this.hf === 'All' || (this.hf === 'Inbound' && h.dir === 'in') || (this.hf === 'Outbound' && h.dir === 'out') || (this.hf === 'Missed' && /Missed|voicemail|Declined/.test(h.result))) &&
        (!q || [this.nameFor(h), h.number, h.note, h.disposition].join(' ').toLowerCase().includes(q)));
    },
    customerRows() { const q = this.cq.trim().toLowerCase(); return this.pbx.customers.filter((c) => !q || [c.name, c.account, c.phone, c.product, c.tags.join(' ')].join(' ').toLowerCase().includes(q)); },
    curThread() { return this.pbx.threads.find((t) => t.id === this.thread) || null; },
    totals() {
      const Q = this.pbx.queues;
      return { waiting: Q.reduce((a, q) => a + q.waiting, 0), longest: Math.max(0, ...Q.map((q) => q.longest)), answered: Q.reduce((a, q) => a + q.answered, 0), abandoned: Q.reduce((a, q) => a + q.abandoned, 0), free: this.pbx.agents.filter((a) => a.status === 'Available').length };
    },
    presenceTone() { return { Available: 'ok', Away: 'mid', Lunch: 'mid', Busy: 'bad', 'Do not disturb': 'bad', 'Off work': 'off' }[this.pbx.presence] || 'off'; }
  },
  methods: {
    simulateIncoming, answer, decline, endCall, toggleHold, swapTo, sendDtmf, completeTransfer, cancelTransfer, addToConference, dropFromConference, park, unpark, finishWrap, setAgentQueue, setPresence, saveSettings, markHeard, deleteVoicemail,
    customerOf, showNumber, fmtDur, fmtWhen, money,
    initials(n) { return String(n || '?').split(/\s+/).filter(Boolean).slice(0, 2).map((w) => w[0]).join('').toUpperCase(); },
    nameFor(h) { const c = h.customerId ? customerOf(h.customerId) : customerByNumber(h.number); return c ? c.name : showNumber(h.number); },
    dirIcon(h) { return /Missed|Declined|voicemail/.test(h.result) ? '↙✕' : h.dir === 'in' ? '↙' : '↗'; },
    dirTone(h) { return /Missed|Declined|voicemail/.test(h.result) ? 'bad' : h.dir === 'in' ? 'in' : 'out'; },
    agentTone(s) { return { Available: 'ok', 'On a call': 'busy', Ringing: 'busy', 'Wrap-up': 'mid', Away: 'mid', Lunch: 'mid' }[s] || 'off'; },
    callNumber() { const n = this.numberMatches.length === 1 ? this.numberMatches[0].phone : this.number; this.dialErr = dial(n); if (!this.dialErr) this.number = ''; },
    callTo(n) { this.dialErr = dial(n); if (this.dialErr) this.tab = 'dial'; },
    togglePanel(p) { this.panel = this.panel === p ? '' : p; },
    toggleRecord() { const c = this.pbx.call; if (!c.recording) { c.recording = true; c.recPaused = false; } else c.recPaused = !c.recPaused; },
    doTransfer(mode) { transfer(this.transferTarget, mode); this.transferTarget = ''; this.panel = ''; },
    reveal(n) { this.revealed = { ...this.revealed, [n]: true }; logAudit('Showed the number ' + showNumber(n) + ' (' + (customerByNumber(n) || { name: 'unknown caller' }).name + ')'); },
    messageTo(c) { this.thread = startThread(c.id); this.tab = 'messages'; },
    openThread(t) { this.thread = t.id; t.unread = 0; },
    send() { sendMessage(this.thread, this.draft); this.draft = ''; },
    superviseNow(a, how) { supervise(a, how); this.supNote = how + ' started on ' + a.name + '\'s call (simulated). ' + (how === 'Listen' ? 'Nobody hears you.' : how === 'Whisper' ? 'Only ' + a.name + ' hears you.' : 'Both sides hear you.'); },
    block() { const n = norm(this.blockNo); if (n && !this.pbx.settings.blocked.includes(n)) { this.pbx.settings.blocked.push(n); saveSettings(); } this.blockNo = ''; },
    exportCsv() {
      const esc = (v) => '"' + String(v == null ? '' : v).replace(/"/g, '""') + '"';
      const rows = [['Direction', 'Customer', 'Number', 'Result', 'Outcome', 'Duration (s)', 'When', 'Notes']].concat(this.historyRows.map((h) => [h.dir === 'in' ? 'Inbound' : 'Outbound', this.nameFor(h), showNumber(h.number), h.result, h.disposition, Math.round(h.duration / 1000), new Date(h.at).toISOString(), h.note]));
      const url = URL.createObjectURL(new Blob([rows.map((r) => r.map(esc).join(',')).join('\n')], { type: 'text/csv' }));
      const a = document.createElement('a'); a.href = url; a.download = 'call-history.csv'; a.click(); setTimeout(() => URL.revokeObjectURL(url), 1000);
    }
  }
};
</script>

<style scoped>
.pbx { display: flex; flex-direction: column; gap: 16px; max-width: 1280px; }
.pbx h1 { margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 24px; }
.pbx h2 { margin: 0 0 10px; font-size: 15px; }
.pbx-head { display: flex; flex-wrap: wrap; align-items: center; gap: 10px 14px; }
.pbx-title { flex: 1 1 260px; display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.pbx-sub { font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); }
.block { display: block; }
.clip { max-width: 260px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.nowrap { white-space: nowrap; }
.grow { flex: 1; min-width: 0; }
.warn { color: var(--k-FFB86B,#FFB86B); }
.pbx-vh { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); }
.pbx-demo { font-size: 12px; padding: 4px 10px; border-radius: 12px; background: var(--k-3A2E1C,#3A2E1C); color: var(--k-FFB86B,#FFB86B); }
.pbx-presence { position: relative; display: inline-flex; align-items: center; gap: 8px; padding: 0 4px 0 12px; height: 38px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); }
.pbx-presence select { height: 34px; border: 0; background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-size: 13px; }
.pbx-dot { width: 10px; height: 10px; border-radius: 50%; }
.pbx-dot.ok { background: #B8F36B; } .pbx-dot.mid { background: #FFB86B; } .pbx-dot.bad { background: #FF9B8A; } .pbx-dot.off { background: #6B7A72; }
.pbx-btn { height: 38px; padding: 0 16px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; font-size: 13px; cursor: pointer; white-space: nowrap; }
.pbx-btn.go { background: #B8F36B; color: #0F1714; border-color: #B8F36B; }
.pbx-btn.stop { background: #FF9B8A; color: #0F1714; border-color: #FF9B8A; }
.pbx-btn.small { height: 32px; padding: 0 12px; font-size: 12px; }
.pbx-btn.big { height: 52px; min-width: 160px; border-radius: 26px; font-size: 16px; }
.pbx-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.pbx-link { border: 0; background: none; padding: 0; color: var(--k-7CC4FF,#7CC4FF); cursor: pointer; font-size: 12px; font-weight: 600; }
.pbx-row { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.pbx-row.between { justify-content: space-between; }
.pbx-row.center { justify-content: center; }
.pbx-card { padding: 16px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 12px; min-width: 0; }
.pbx-grid2 { display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 340px), 1fr)); gap: 16px; }
.pbx-ring { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 12px; padding: 16px; border-radius: 16px; background: var(--k-1C2B22,#1C2B22); border: 1px solid #B8F36B; }
.pbx-ring-who { display: flex; gap: 14px; align-items: center; min-width: 0; }
.pbx-ring-who > div { display: flex; flex-direction: column; gap: 2px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); }
.pbx-ring-who strong { font-size: 17px; color: var(--k-E8EFEA,#E8EFEA); }
.pbx-ring-crm { font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); }
.pbx-ring-pulse { width: 16px; height: 16px; flex: none; border-radius: 50%; background: #B8F36B; animation: pbxpulse 1s infinite; }
@keyframes pbxpulse { 0% { box-shadow: 0 0 0 0 rgba(184,243,107,.6); } 100% { box-shadow: 0 0 0 14px rgba(184,243,107,0); } }
@media (prefers-reduced-motion: reduce) { .pbx-ring-pulse { animation: none; } }
.pbx-live { padding: 16px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-3A5446,#3A5446); display: flex; flex-direction: column; gap: 12px; }
.pbx-live-top { display: flex; justify-content: space-between; align-items: center; gap: 12px; flex-wrap: wrap; }
.pbx-live-who { display: flex; gap: 12px; align-items: center; min-width: 0; }
.pbx-live-who > div { display: flex; flex-direction: column; gap: 2px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); }
.pbx-live-who strong { font-size: 18px; color: var(--k-E8EFEA,#E8EFEA); }
.pbx-state { font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); }
.pbx-rec { color: #FF9B8A; }
.pbx-timer { font-family: 'Space Grotesk', sans-serif; font-size: 28px; font-weight: 700; font-variant-numeric: tabular-nums; }
.pbx-crm { display: grid; grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap: 8px; padding: 10px 12px; border-radius: 12px; background: var(--k-121B17,#121B17); }
.pbx-crm div { display: flex; flex-direction: column; gap: 2px; font-size: 13px; }
.pbx-crm span { font-size: 11px; color: var(--k-9DB0A6,#9DB0A6); }
.pbx-controls { display: flex; flex-wrap: wrap; gap: 8px; }
.pbx-ctl { height: 40px; padding: 0 14px; border-radius: 12px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px; font-weight: 600; }
.pbx-ctl.on { background: var(--k-2C3D34,#2C3D34); border-color: #B8F36B; }
.pbx-ctl.end { background: #FF6B5A; border-color: #FF6B5A; color: #fff; margin-left: auto; }
.pbx-subpanel { display: flex; flex-direction: column; gap: 10px; padding: 12px; border-radius: 12px; background: var(--k-121B17,#121B17); }
.pbx-field { display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); }
.pbx-field input, .pbx-field select, .pbx-field textarea, .pbx textarea, .pbx-search { box-sizing: border-box; width: 100%; min-height: 40px; padding: 8px 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font: inherit; font-size: 14px; }
.pbx-search { width: auto; flex: 1 1 220px; }
.pbx-check { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.pbx-note { padding: 10px 12px; border-radius: 12px; background: var(--k-1A2520,#1A2520); font-size: 13px; display: flex; flex-direction: column; gap: 8px; }
.pbx-strip { display: flex; flex-wrap: wrap; gap: 8px; }
.pbx-chip { display: inline-flex; align-items: center; gap: 10px; padding: 6px 12px; border-radius: 16px; background: var(--k-22302A,#22302A); font-size: 13px; }
.pbx-wrap { padding: 14px 16px; border-radius: 16px; border: 1px solid #FFB86B; background: var(--k-16201C,#16201C); display: flex; flex-direction: column; gap: 10px; }
.pbx-tags { display: flex; flex-wrap: wrap; gap: 6px; }
.pbx-tag { height: 28px; padding: 0 10px; border-radius: 14px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-C9D6CF,#C9D6CF); font-size: 12px; cursor: pointer; display: inline-flex; align-items: center; }
.pbx-tag.on { background: #B8F36B; color: #0F1714; border-color: #B8F36B; font-weight: 600; }
.pbx-tag.static { cursor: default; height: 22px; margin-left: 6px; }
.pbx-tabs { display: flex; gap: 4px; overflow-x: auto; border-bottom: 1px solid var(--k-24322C,#24322C); }
.pbx-tabs button { flex: none; min-height: 40px; padding: 0 14px; border: 0; border-bottom: 2px solid transparent; background: transparent; color: var(--k-9DB0A6,#9DB0A6); cursor: pointer; font-size: 14px; font-weight: 500; display: inline-flex; align-items: center; gap: 6px; }
.pbx-tabs button.on { color: var(--k-E8EFEA,#E8EFEA); border-bottom-color: #B8F36B; font-weight: 600; }
.pbx-count { min-width: 18px; height: 18px; padding: 0 5px; box-sizing: border-box; border-radius: 9px; background: #FF9B8A; color: #0F1714; font-size: 11px; font-weight: 700; display: inline-flex; align-items: center; justify-content: center; }
.pbx-number { box-sizing: border-box; width: 100%; height: 56px; padding: 0 14px; border-radius: 14px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-family: 'Space Grotesk', sans-serif; font-size: 24px; text-align: center; }
.pbx-keys { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; max-width: 320px; width: 100%; margin: 0 auto; }
.pbx-keys button { height: 60px; border-radius: 16px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 2px; }
.pbx-keys button strong { font-size: 22px; } .pbx-keys button small { font-size: 10px; letter-spacing: 0.1em; color: var(--k-9DB0A6,#9DB0A6); min-height: 12px; }
.pbx-keys.small button { height: 44px; }
.pbx-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 4px; }
.pbx-list li { display: flex; align-items: center; gap: 10px; padding: 8px 6px; border-radius: 10px; font-size: 13px; }
.pbx-list li.col { flex-direction: column; align-items: stretch; border-bottom: 1px solid var(--k-24322C,#24322C); }
.pbx-list li > .grow { display: flex; flex-direction: column; }
.pbx-dir { display: inline-flex; min-width: 28px; justify-content: center; font-weight: 700; }
.pbx-dir.in { color: #7CC4FF; } .pbx-dir.out { color: #B8F36B; } .pbx-dir.bad { color: #FF9B8A; }
.pbx-seg { display: inline-flex; padding: 3px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-2A3832,#2A3832); }
.pbx-seg button { height: 30px; padding: 0 12px; border: 0; border-radius: 9px; background: transparent; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; font-size: 13px; font-weight: 600; }
.pbx-seg button.on { background: #B8F36B; color: #0F1714; }
.pbx-table-wrap { overflow-x: auto; }
.pbx-table { width: 100%; min-width: 760px; border-collapse: collapse; font-size: 13px; }
.pbx-table th { text-align: left; font-size: 12px; font-weight: 600; color: var(--k-9DB0A6,#9DB0A6); padding: 8px; border-bottom: 1px solid var(--k-24322C,#24322C); }
.pbx-table td { padding: 10px 8px; border-bottom: 1px solid var(--k-1E2A25,#1E2A25); vertical-align: top; }
.pbx-empty { padding: 16px; text-align: center; color: var(--k-9DB0A6,#9DB0A6); font-size: 13px; }
.pbx-err { color: #FF9B8A; font-size: 13px; }
.pbx-cards { display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 260px), 1fr)); gap: 12px; }
.pbx-cust { padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 8px; }
.pbx-cust .grow { display: flex; flex-direction: column; }
.pbx-msgs { display: grid; grid-template-columns: minmax(220px, 300px) 1fr; gap: 16px; }
@media (max-width: 760px) { .pbx-msgs { grid-template-columns: 1fr; } }
.pbx-threads { gap: 4px; padding: 8px; }
.pbx-thread { display: flex; align-items: center; gap: 8px; padding: 10px; border: 0; border-radius: 10px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left; font: inherit; font-size: 13px; }
.pbx-thread.on { background: var(--k-22302A,#22302A); }
.pbx-thread .grow { display: flex; flex-direction: column; }
.pbx-convo { min-height: 360px; }
.pbx-bubbles { flex: 1; display: flex; flex-direction: column; gap: 6px; max-height: 420px; overflow-y: auto; }
.pbx-bubble { max-width: 78%; padding: 8px 12px; border-radius: 14px; display: flex; flex-direction: column; gap: 4px; font-size: 14px; }
.pbx-bubble small { font-size: 11px; color: var(--k-9DB0A6,#9DB0A6); }
.pbx-bubble.me { margin-left: auto; background: var(--k-2C3D34,#2C3D34); border-radius: 14px 4px 14px 14px; }
.pbx-bubble.them { background: var(--k-1A2520,#1A2520); border-radius: 4px 14px 14px 14px; }
.pbx-transcript { margin: 4px 0; font-size: 14px; color: var(--k-C9D6CF,#C9D6CF); }
.pbx-mon { display: flex; flex-direction: column; gap: 16px; }
.pbx-kpis { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 12px; }
.pbx-kpi { padding: 14px 16px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 6px; }
.pbx-kpi span { font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); }
.pbx-kpi strong { font-family: 'Space Grotesk', sans-serif; font-size: 26px; }
.pbx-sla { font-weight: 700; } .pbx-sla.ok { color: #B8F36B; } .pbx-sla.mid { color: #FFB86B; } .pbx-sla.bad { color: #FF9B8A; }
.pbx-pill { display: inline-flex; padding: 2px 10px; border-radius: 10px; font-size: 12px; font-weight: 600; }
.pbx-pill.ok { background: rgba(184,243,107,.15); color: #B8F36B; } .pbx-pill.busy { background: rgba(124,196,255,.15); color: #7CC4FF; }
.pbx-pill.mid { background: rgba(255,184,107,.15); color: #FFB86B; } .pbx-pill.off { background: rgba(107,122,114,.2); color: var(--k-9DB0A6,#9DB0A6); }
.pbx-switch input { width: 18px; height: 18px; }
</style>
