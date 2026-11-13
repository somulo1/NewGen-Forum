<!-- Calls management: the customer phone system (softphone, call history, customer messaging, voicemail, queue monitoring).
     Front end only for now: calls are simulated in logic/pbx/store.js until the PBX is connected. -->
<template>
<div class="pbx">
  <div class="pbx-top">
  <header class="pbx-head">
    <h1>Calls management</h1>
    <span class="pbx-demo" title="Calls, sounds and data are samples until the phone system is connected">Sample mode</span>
    <label class="pbx-presence">
      <span :class="['pbx-dot', presenceTone]" aria-hidden="true"></span>
      <span class="pbx-vh">My status</span>
      <select :value="pbx.presence" @change="setPresence($event.target.value)">
        <option v-for="p in presences" :key="p" :value="p">{{ p }} · ext {{ pbx.settings.extension }}</option>
      </select>
    </label>
  </header>
  <nav class="pbx-tabs" role="tablist" aria-label="Calls management">
    <button v-for="t in tabs" :key="t.k" type="button" role="tab" :aria-selected="tab === t.k" :class="{ on: tab === t.k }" @click="tab = t.k">{{ t.l }}<span v-if="t.n" class="pbx-count">{{ t.n }}</span></button>
  </nav>
  </div>
  <div class="pbx-layout">
  <!-- dial pad -->
  <aside class="pbx-dialcol" aria-label="Dial pad">
    <div class="pbx-card">
      <input class="pbx-number" v-model="number" inputmode="tel" placeholder="Number or customer name" aria-label="Number to call" @keydown.enter="callNumber">
      <ul v-if="numberMatches.length" class="pbx-list">
        <li v-for="c in numberMatches" :key="c.id"><button type="button" class="pbx-link" @click="number = c.phone">{{ c.name }}</button><span class="pbx-sub">{{ showNumber(c.phone) }}</span></li>
      </ul>
      <div class="pbx-keys"><button v-for="k in keys" :key="k[0]" type="button" @click="number += k[0]; keyTone(k[0])"><strong>{{ k[0] }}</strong><small>{{ k[1] }}</small></button></div>
      <div class="pbx-row center">
        <button type="button" class="pbx-btn ghost" @click="number = number.slice(0, -1)" aria-label="Delete last digit">⌫</button>
        <button type="button" class="pbx-btn go big" @click="callNumber">Call</button>
      </div>
      <label class="pbx-field">Call out as
        <select v-model="pbx.settings.callerId" @change="saveSettings"><option v-for="c in pbx.settings.callerIds" :key="c">{{ c }}</option></select>
      </label>
      <span v-if="dialErr" role="alert" class="pbx-err">{{ dialErr }}</span>
    </div>
  </aside>
  <div class="pbx-center">
  <!-- after-call work -->
  <section v-if="pbx.wrap && !pbx.call" class="pbx-wrap" aria-label="After-call work">
    <div class="pbx-row between"><strong>How did the call end?</strong><span class="pbx-sub">Saves by itself in {{ pbx.wrap.left }}s</span></div>
    <div class="pbx-tags"><button v-for="d in pbx.dispositions" :key="d" type="button" :class="['pbx-tag', { on: pbx.wrap.disposition === d }]" @click="pbx.wrap.disposition = d">{{ d }}</button></div>
    <textarea v-model="pbx.wrap.note" rows="2" placeholder="Notes for the customer record"></textarea>
    <div class="pbx-row"><button type="button" class="pbx-btn go" @click="finishWrap">Save</button></div>
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
        <thead><tr><th></th><th>Customer</th><th>Result</th><th>Outcome</th><th>Duration</th><th>When</th><th></th></tr></thead>
        <tbody>
          <tr v-for="h in historyRows" :key="h.id">
            <td><span :class="['pbx-dir', dirTone(h)]">{{ dirIcon(h) }}</span></td>
            <td><strong>{{ nameFor(h) }}</strong><span class="pbx-sub block">{{ showNumber(h.number) }}</span><span v-if="h.note" class="pbx-sub block clip" :title="h.note">{{ h.note }}</span></td>
            <td>{{ h.result }}</td>
            <td>{{ h.disposition || '—' }}</td>
            <td>{{ h.duration ? fmtDur(h.duration) : '—' }}</td>
            <td class="nowrap">{{ fmtWhen(h.at) }}</td>
            <td><button type="button" class="pbx-btn small go" @click="callTo(h.number)">Call back</button></td>
          </tr>
          <tr v-if="!historyRows.length"><td colspan="7" class="pbx-empty">No calls match.</td></tr>
        </tbody>
      </table>
    </div>
  </section>

  <!-- customers -->
  <section v-if="tab === 'contacts'" class="pbx-card">
    <div class="pbx-seg" role="radiogroup" aria-label="Show" style="align-self: flex-start"><button v-for="w in ['Customers', 'Colleagues']" :key="w" type="button" role="radio" :aria-checked="who === w" :class="{ on: who === w }" @click="who = w">{{ w }}</button></div>
    <ul v-if="who === 'Colleagues'" class="pbx-list">
      <li v-for="a in pbx.agents.filter((x) => !x.me)" :key="a.ext"><UserAvatar :name="a.name" :initials="initials(a.name)" color="#B8F36B" :size="32" :font-size="11"/><span class="grow"><strong>{{ a.name }}</strong><span class="pbx-sub block">Ext {{ a.ext }} · {{ a.queues.join(', ') }}</span></span><span :class="['pbx-pill', agentTone(a.status)]">{{ a.status }}</span><button type="button" class="pbx-btn small go" :disabled="a.status === 'Logged out'" @click="callTo(a.ext)">Call</button></li>
    </ul>
    <template v-else>
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
    </template>
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
    <div class="pbx-seg" role="radiogroup" aria-label="Voicemail box" style="align-self: flex-start; flex-wrap: wrap"><button v-for="b in boxes" :key="b" type="button" role="radio" :aria-checked="box === b" :class="{ on: box === b }" @click="box = b">{{ b }}</button></div>
    <ul class="pbx-list">
      <li v-for="v in pbx.voicemail.filter((x) => box === 'All boxes' || (x.box || 'My voicemail') === box)" :key="v.id" class="col">
        <div class="pbx-row between"><span><strong>{{ v.customerId ? customerOf(v.customerId).name : showNumber(v.number) }}</strong><span v-if="!v.heard" class="pbx-tag static">New</span></span><span class="pbx-sub">{{ fmtWhen(v.at) }} · {{ fmtDur(v.duration) }}</span></div>
        <p class="pbx-transcript">"{{ v.transcript }}"</p>
        <div class="pbx-row"><button type="button" class="pbx-btn small ghost" @click="markHeard(v)">{{ v.heard ? 'Play again' : 'Play' }}</button><button type="button" class="pbx-btn small go" @click="callTo(v.number)">Call back</button><button type="button" class="pbx-btn small stop" @click="deleteVoicemail(v.id)">Delete</button></div>
      </li>
      <li v-if="!pbx.voicemail.length" class="pbx-empty">No voicemail.</li>
    </ul>
  </section>

  <!-- monitoring / wallboard -->
  <section v-if="tab === 'monitor'" class="pbx-mon">
    <div v-for="a in pbx.alerts" :key="a.queue" class="pbx-alert" role="alert">⚠ {{ a.text }} <span class="pbx-sub">since {{ fmtWhen(a.at) }}</span></div>
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

  <PbxRecordings v-if="tab === 'recordings'" />
  <PbxCampaigns v-if="tab === 'campaigns'" />
  <PbxReports v-if="tab === 'reports'" />
  <PbxRouting v-if="tab === 'routing'" />

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
      <span class="pbx-sub">Every call is recorded from start to end and cannot be paused, so all calls can be audited. Customers hear the recording notice.</span>
      <label class="pbx-check"><input type="checkbox" v-model="pbx.settings.maskNumbers" @change="saveSettings"> Hide customer numbers until I choose to show one</label>
      <label class="pbx-field">After-call work time (seconds)<input type="number" min="0" max="300" v-model.number="pbx.settings.wrapUpSec" @change="saveSettings"></label>
      <h2>Audio devices</h2>
      <label class="pbx-field">Microphone<select v-model="pbx.settings.mic" @change="saveSettings"><option value="">System default</option><option v-for="d in devices.mics" :key="d.id" :value="d.id">{{ d.label }}</option></select></label>
      <label class="pbx-field">Speaker<select v-model="pbx.settings.speaker" @change="saveSettings"><option value="">System default</option><option v-for="d in devices.speakers" :key="d.id" :value="d.id">{{ d.label }}</option></select></label>
      <button v-if="!devices.asked" type="button" class="pbx-btn ghost small" @click="loadDevices(true)">Show my devices</button>
      <h2>My voicemail greeting</h2>
      <div class="pbx-row"><button type="button" :class="['pbx-btn', recording ? 'stop' : 'ghost']" @click="recordGreeting">{{ recording ? 'Stop recording (' + recSecs + 's)' : 'Record a greeting' }}</button><span v-if="recErr" role="alert" class="pbx-err">{{ recErr }}</span></div>
      <ul class="pbx-list"><li v-for="(g, i) in pbx.greetings" :key="g.at"><audio :src="g.url" controls preload="none" style="height: 32px; max-width: 200px"></audio><span class="grow pbx-sub">{{ fmtWhen(g.at) }}</span><button type="button" class="pbx-link" @click="pbx.settings.greeting = 'My recording ' + fmtWhen(g.at); saveSettings()">Use</button><button type="button" class="pbx-link" @click="removeGreeting(i)">Delete</button></li></ul>
      <label class="pbx-field">Greeting in use<select v-model="pbx.settings.greeting" @change="saveSettings"><option>Default greeting</option><option>Out of office</option><option>After hours</option><option v-for="g in pbx.greetings" :key="g.at">My recording {{ fmtWhen(g.at) }}</option></select></label>
      <h2>Wallboard alerts</h2>
      <label class="pbx-field">Alert when someone waits longer than (s)<input type="number" min="10" max="900" v-model.number="pbx.settings.alertWaitSec" @change="saveSettings"></label>
      <label class="pbx-field">Alert when service level falls below (%)<input type="number" min="50" max="100" v-model.number="pbx.settings.alertSla" @change="saveSettings"></label>
      <h2>Customer care tune</h2>
      <span class="pbx-sub">What customers hear while they wait in a queue or are on hold. A sample tune with a sample AI greeting until the phone system plays the real one.</span>
      <label class="pbx-field">Tune<select v-model="pbx.settings.tune" @change="saveSettings"><option v-for="t in tuneNames" :key="t">{{ t }}</option></select></label>
      <label class="pbx-check"><input type="checkbox" v-model="pbx.settings.tuneGreeting" @change="saveSettings"> Play the spoken greeting first</label>
      <div class="pbx-row"><button type="button" class="pbx-btn ghost" @click="previewTune">{{ tuneOn ? 'Stop' : 'Listen' }}</button><button type="button" class="pbx-btn ghost" @click="simulateIncoming()">Test an inbound call</button></div>
      <h2>Blocked numbers</h2>
      <ul class="pbx-list"><li v-for="(b, i) in pbx.settings.blocked" :key="b"><span class="grow">{{ showNumber(b) }}</span><button type="button" class="pbx-link" @click="pbx.settings.blocked.splice(i, 1); saveSettings()">Unblock</button></li></ul>
      <form class="pbx-row" @submit.prevent="block"><input v-model="blockNo" class="grow pbx-search" placeholder="Number to block" aria-label="Number to block"><button type="submit" class="pbx-btn ghost">Block</button></form>
    </div>
  </section>
  </div>
  <!-- the live call: built like Hub's voice call screen -->
  <aside v-if="pbx.call" class="pbx-callpanel" aria-label="Current call" :style="{ width: callW + 'px' }">
    <button type="button" class="pbx-grip" aria-label="Drag to resize the call panel. Arrow keys also resize." @pointerdown="startResize" @keydown.left.prevent="callW = clampW(callW + 24)" @keydown.right.prevent="callW = clampW(callW - 24)"><span></span></button>
    <div class="pbx-callscreen">
    <div class="pbx-cs-top">
      <span class="pbx-cs-q">{{ pbx.call.dir === 'in' ? 'Inbound' : 'Outbound' }}<template v-if="pbx.call.queue"> · {{ pbx.call.queue }}</template><template v-else> · as {{ pbx.call.callerId }}</template></span>
      <span class="pbx-chip-rec" title="Every call is recorded for audit"><span aria-hidden="true"></span>REC</span>
      <span v-if="pbx.call.answeredAt" :class="['pbx-chip-q', qualityTone]" :title="`Line quality ${pbx.quality.mos} of 4.5 · jitter ${pbx.quality.jitter} ms · loss ${pbx.quality.loss}% · delay ${pbx.quality.rtt} ms`">{{ qualityLabel }}</span>
      <span class="pbx-chip-time">{{ liveTimer }}</span>
    </div>
    <div class="pbx-cs-body">
      <div class="pbx-cs-main">
        <UserAvatar :class="['pbx-cs-face', { talking: callState === 'Talking' && !pbx.call.muted, ringing: /Dialling|Ringing/.test(callState) }]" :name="liveName" :initials="initials(liveName)" color="#7CC4FF" :size="104" :font-size="34"/>
        <strong class="pbx-cs-name">{{ liveName }}</strong>
        <span class="pbx-sub">{{ showNumber(pbx.call.number, revealed[pbx.call.number]) }} <button v-if="pbx.settings.maskNumbers && !revealed[pbx.call.number]" type="button" class="pbx-link" @click="reveal(pbx.call.number)">Show</button></span>
        <span :class="['pbx-cs-state', { hold: pbx.call.onHold }]">{{ pbx.call.transferMode === 'consulting' ? 'Talking to ' + pbx.call.transferTo + ' · customer on hold' : pbx.call.onHold ? 'On hold · customer hears ' + pbx.settings.tune : callState }}</span>
        <div v-if="liveCustomer" class="pbx-cs-facts">
          <span>{{ liveCustomer.account }}</span><span>{{ liveCustomer.product }}</span><span>{{ money(liveCustomer.balance) }}</span>
          <span :class="{ warn: /arrears|Frozen/.test(liveCustomer.status) }">{{ liveCustomer.status }}</span><span :class="{ warn: liveCustomer.kyc !== 'Verified' }">KYC {{ liveCustomer.kyc }}</span>
        </div>
        <div v-if="pbx.call.transferMode === 'consulting'" class="pbx-row center"><button type="button" class="pbx-btn go" @click="completeTransfer">Complete transfer</button><button type="button" class="pbx-btn ghost" @click="cancelTransfer">Back to the customer</button></div>
        <div v-if="pbx.call.conference.length" class="pbx-cs-facts"><span v-for="(p, i) in pbx.call.conference" :key="i">{{ p.who }} <button type="button" class="pbx-link" :aria-label="'Remove ' + p.who" @click="dropFromConference(i)">×</button></span></div>
        <div v-if="pbx.held.length || pbx.parked.length" class="pbx-strip">
          <div v-for="h in pbx.held" :key="h.id" class="pbx-chip"><span>On hold · {{ nameFor(h) }}</span><button type="button" class="pbx-link" @click="swapTo(h.id)">Switch</button></div>
          <div v-for="p in pbx.parked" :key="p.slot" class="pbx-chip"><span>Parked {{ p.slot }} · {{ nameFor(p.call) }}</span><button type="button" class="pbx-link" @click="unpark(p.slot)">Pick up</button></div>
        </div>
      </div>
      <aside v-if="panel" class="pbx-cs-side" :aria-label="panelTitle">
        <div class="pbx-row between"><strong>{{ panelTitle }}</strong><button type="button" class="pbx-x" aria-label="Close" @click="panel = ''">×</button></div>
        <template v-if="panel === 'keys'">
          <div class="pbx-keys small"><button v-for="k in keys" :key="k[0]" type="button" @click="sendDtmf(k[0])"><strong>{{ k[0] }}</strong></button></div>
          <span class="pbx-sub">Sent: {{ pbx.call.dtmf || 'nothing yet' }}</span>
        </template>
        <template v-if="panel === 'transfer' || panel === 'conf'">
          <input v-model="target" list="pbx-targets" class="pbx-search" :placeholder="panel === 'transfer' ? 'Extension, queue or number' : 'Who to add (up to 5)'" :aria-label="panelTitle">
          <datalist id="pbx-targets"><option v-for="t in targets" :key="t" :value="t"></option></datalist>
          <div v-if="panel === 'transfer'" class="pbx-row"><button type="button" class="pbx-btn go" :disabled="!target" @click="doTransfer('attended')">Talk first</button><button type="button" class="pbx-btn ghost" :disabled="!target" @click="doTransfer('blind')">Transfer now</button></div>
          <button v-else type="button" class="pbx-btn go" :disabled="!target" @click="addToConference(target); target = ''">Add to call</button>
        </template>
        <textarea v-if="panel === 'notes'" v-model="pbx.call.note" rows="8" placeholder="What was agreed. Saved with the call."></textarea>
      </aside>
    </div>
    <div class="pbx-cs-bar" role="toolbar" aria-label="Call controls">
      <button v-for="c in controls" :key="c.k" type="button" :class="['pbx-ctl', c.cls]" :aria-pressed="c.pressed" :aria-label="c.label" @click="c.run"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" v-html="c.icon"></svg><span>{{ c.label }}</span></button>
    </div>
    </div>
  </aside>
  </div>
</div>
</template>

<script>
import { TUNES, startTune, stopTune, keyTone } from '~/logic/pbx/sounds.js';
import { removeGreeting, addGreeting } from '~/logic/pbx/store.js';
import { pbx, dial, simulateIncoming, answer, decline, endCall, toggleHold, swapTo, sendDtmf, transfer, completeTransfer, cancelTransfer, addToConference, dropFromConference, park, unpark, finishWrap,
  sendMessage, startThread, supervise, setAgentQueue, setPresence, saveSettings, markHeard, deleteVoicemail, startClock, stopClock, customerOf, customerByNumber, showNumber, fmtDur, fmtWhen, money, norm, logAudit } from '~/logic/pbx/store.js';

const KEYS = [['1', ''], ['2', 'ABC'], ['3', 'DEF'], ['4', 'GHI'], ['5', 'JKL'], ['6', 'MNO'], ['7', 'PQRS'], ['8', 'TUV'], ['9', 'WXYZ'], ['*', ''], ['0', '+'], ['#', '']];

export default {
  name: 'CallsManagementSection',
  data() {
    return { pbx, tab: 'history', callW: 380, number: '', dialErr: '', panel: '', target: '', who: 'Customers', box: 'All boxes', devices: { mics: [], speakers: [], asked: false }, recording: null, recSecs: 0, recErr: '', tuneOn: false, hearHold: false, hq: '', hf: 'All', cq: '', thread: null, draft: '', playing: null, revealed: {}, supNote: '', blockNo: '',
      keys: KEYS, presences: ['Available', 'Away', 'Busy', 'Lunch', 'Do not disturb', 'Off work'],
      templates: [
        { l: 'Payment reminder', t: (c) => 'Hello ' + (c ? c.name.split(' ')[0] : '') + ', a reminder that your payment on ' + (c ? c.account : 'your account') + ' is due. Reply PAY for the paybill details.' },
        { l: 'Missed your call', t: (c) => 'Hello ' + (c ? c.name.split(' ')[0] : '') + ', we tried to reach you. Please call us back on 0709 000 000.' },
        { l: 'KYC documents', t: () => 'Please upload a clear photo of your ID and a selfie in the app to finish verification.' }
      ] };
  },
  mounted() { startClock(); this.loadDevices(false); try { const w = +localStorage.getItem('hub.pbx.callW'); if (w) this.callW = this.clampW(w); } catch (e) { /* private window */ } },
  beforeUnmount() { stopClock(); stopTune(); },
  watch: { 'pbx.call.onHold'(v) { if (!v && this.hearHold) { stopTune(); this.hearHold = false; } }, 'pbx.call'(v) { if (!v) { this.panel = ''; if (this.hearHold) { stopTune(); this.hearHold = false; } } } },
  computed: {
    tabs() {
      const missed = this.pbx.history.filter((h) => /Missed/.test(h.result) && Date.now() - h.at < 86400000).length;
      return [{ k: 'history', l: 'Call history', n: missed }, { k: 'contacts', l: 'Contacts' }, { k: 'messages', l: 'Messages', n: this.pbx.threads.reduce((a, t) => a + t.unread, 0) },
        { k: 'voicemail', l: 'Voicemail', n: this.pbx.voicemail.filter((v) => !v.heard).length }, { k: 'recordings', l: 'Recordings' }, { k: 'campaigns', l: 'Campaigns' },
        { k: 'monitor', l: 'Monitoring', n: this.pbx.alerts.length }, { k: 'reports', l: 'Reports' }, { k: 'routing', l: 'Routing', n: this.pbx.callbacks.filter((c) => c.status === 'Waiting').length }, { k: 'settings', l: 'Settings' }];
    },
    controls() {
      const c = this.pbx.call; if (!c) return [];
      const ic = {'mic': '<rect x="9" y="3" width="6" height="11" rx="3"></rect><path d="M5 11a7 7 0 0 0 14 0M12 18v3"></path>', 'micoff': '<path d="M9 9v2a3 3 0 0 0 5 2.2M15 9.3V6a3 3 0 0 0-5.7-1.3M5 11a7 7 0 0 0 11.5 5.4M19 11a7 7 0 0 1-.6 2.8M12 18v3M3 3l18 18"></path>', 'hold': '<path d="M9 5v14M15 5v14"></path>', 'keys': '<circle cx="6" cy="5" r="1.2"></circle><circle cx="12" cy="5" r="1.2"></circle><circle cx="18" cy="5" r="1.2"></circle><circle cx="6" cy="11" r="1.2"></circle><circle cx="12" cy="11" r="1.2"></circle><circle cx="18" cy="11" r="1.2"></circle><circle cx="6" cy="17" r="1.2"></circle><circle cx="12" cy="17" r="1.2"></circle><circle cx="18" cy="17" r="1.2"></circle><circle cx="12" cy="22" r="1.2"></circle>', 'transfer': '<path d="M4 8h13l-3-3M20 16H7l3 3"></path>', 'add': '<circle cx="9" cy="8" r="3.5"></circle><path d="M2 20c0-3.5 3-6 7-6s7 2.5 7 6M19 8v6M16 11h6"></path>', 'rec': '<circle cx="12" cy="12" r="8"></circle><circle cx="12" cy="12" r="3.5" fill="currentColor"></circle>', 'park': '<rect x="4" y="3" width="16" height="18" rx="3"></rect><path d="M10 16V8h3a2.5 2.5 0 0 1 0 5h-3"></path>', 'notes': '<path d="M6 3h9l4 4v14H6z"></path><path d="M9 12h7M9 16h5"></path>', 'tune': '<path d="M9 18V5l11-2v13"></path><circle cx="6" cy="18" r="3"></circle><circle cx="17" cy="16" r="3"></circle>', 'end': '<path d="M3 15c5-5 13-5 18 0l-2.5 2.5-3.5-1.5v-2.5a10 10 0 0 0-6 0V16l-3.5 1.5z" fill="currentColor"></path>'};
      const tog = (k) => () => { this.panel = this.panel === k ? '' : k; };
      return [
        { k: 'mute', label: c.muted ? 'Unmute' : 'Mute', icon: c.muted ? ic.micoff : ic.mic, pressed: c.muted, cls: { warn: c.muted }, run: () => { c.muted = !c.muted; } },
        { k: 'hold', label: c.onHold ? 'Resume' : 'Hold', icon: ic.hold, pressed: c.onHold, cls: { on: c.onHold }, run: () => toggleHold() },
        ...(c.onHold ? [{ k: 'tune', label: this.hearHold ? 'Stop tune' : 'Hear tune', icon: ic.tune, pressed: this.hearHold, cls: { on: this.hearHold }, run: () => this.toggleHear() }] : []),
        { k: 'keys', label: 'Keypad', icon: ic.keys, pressed: this.panel === 'keys', cls: { on: this.panel === 'keys' }, run: tog('keys') },
        { k: 'transfer', label: 'Transfer', icon: ic.transfer, pressed: this.panel === 'transfer', cls: { on: this.panel === 'transfer' }, run: tog('transfer') },
        { k: 'conf', label: 'Add', icon: ic.add, pressed: this.panel === 'conf', cls: { on: this.panel === 'conf' }, run: tog('conf') },
        { k: 'park', label: 'Park', icon: ic.park, cls: {}, run: () => park() },
        { k: 'notes', label: 'Notes', icon: ic.notes, pressed: this.panel === 'notes', cls: { on: this.panel === 'notes' }, run: tog('notes') },
        { k: 'end', label: 'End', icon: ic.end, cls: { end: true }, run: () => endCall() }
      ];
    },
    panelTitle() { return { keys: 'Keypad', transfer: 'Transfer the call', conf: 'Add someone to the call', notes: 'Call notes' }[this.panel] || ''; },
    boxes() { return ['All boxes', 'My voicemail'].concat(this.pbx.queues.map((q) => q.name)); },
    qualityTone() { const m = this.pbx.quality.mos; return m >= 4 ? 'ok' : m >= 3.5 ? 'mid' : 'bad'; },
    qualityLabel() { const m = this.pbx.quality.mos; return m >= 4 ? 'Clear line' : m >= 3.5 ? 'Fair line' : 'Poor line'; },
    tuneNames() { return Object.keys(TUNES); },
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
    customerOf, showNumber, fmtDur, fmtWhen, money, keyTone,
    initials(n) { return String(n || '?').split(/\s+/).filter(Boolean).slice(0, 2).map((w) => w[0]).join('').toUpperCase(); },
    nameFor(h) { const c = h.customerId ? customerOf(h.customerId) : customerByNumber(h.number); return c ? c.name : showNumber(h.number); },
    dirIcon(h) { return /Missed|Declined|voicemail/.test(h.result) ? '↙✕' : h.dir === 'in' ? '↙' : '↗'; },
    dirTone(h) { return /Missed|Declined|voicemail/.test(h.result) ? 'bad' : h.dir === 'in' ? 'in' : 'out'; },
    agentTone(s) { return { Available: 'ok', 'On a call': 'busy', Ringing: 'busy', 'Wrap-up': 'mid', Away: 'mid', Lunch: 'mid' }[s] || 'off'; },
    callNumber() { const n = this.numberMatches.length === 1 ? this.numberMatches[0].phone : this.number; this.dialErr = dial(n); if (!this.dialErr) this.number = ''; },
    callTo(n) { this.dialErr = dial(n); },
    togglePanel(p) { this.panel = this.panel === p ? '' : p; },
    doTransfer(mode) { transfer(this.target, mode); this.target = ''; this.panel = ''; },
    removeGreeting,
    clampW(w) { return Math.max(320, Math.min(Math.round(window.innerWidth * 0.6), w)); },
    // drag the panel's left edge to widen or narrow the call, like the chat side panel
    startResize(e) {
      const x0 = e.clientX, w0 = this.callW;
      const move = (m) => { this.callW = this.clampW(w0 + (x0 - m.clientX)); };
      const up = () => { window.removeEventListener('pointermove', move); window.removeEventListener('pointerup', up); try { localStorage.setItem('hub.pbx.callW', String(this.callW)); } catch (err) { /* private window */ } };
      window.addEventListener('pointermove', move); window.addEventListener('pointerup', up); e.preventDefault();
    },
    async loadDevices(ask) {
      try {
        if (!navigator.mediaDevices) return;
        if (ask) { const st = await navigator.mediaDevices.getUserMedia({ audio: true }); st.getTracks().forEach((t) => t.stop()); }
        const list = await navigator.mediaDevices.enumerateDevices();
        const lab = (d, i, kind) => d.label || kind + ' ' + (i + 1);
        this.devices = { asked: list.some((d) => d.label), mics: list.filter((d) => d.kind === 'audioinput').map((d, i) => ({ id: d.deviceId, label: lab(d, i, 'Microphone') })), speakers: list.filter((d) => d.kind === 'audiooutput').map((d, i) => ({ id: d.deviceId, label: lab(d, i, 'Speaker') })) };
      } catch (e) { this.recErr = 'The browser did not allow the microphone.'; }
    },
    // record a voicemail greeting with the microphone (kept in this browser until the phone system stores greetings)
    async recordGreeting() {
      if (this.recording) { this.recording.stop(); return; }
      this.recErr = '';
      try {
        const st = await navigator.mediaDevices.getUserMedia({ audio: this.pbx.settings.mic ? { deviceId: this.pbx.settings.mic } : true });
        const rec = new MediaRecorder(st), parts = [];
        rec.ondataavailable = (e) => parts.push(e.data);
        rec.onstop = () => {
          st.getTracks().forEach((t) => t.stop()); clearInterval(this._recT); this.recording = null;
          const fr = new FileReader(); fr.onload = () => addGreeting({ at: Date.now(), url: fr.result }); fr.readAsDataURL(new Blob(parts, { type: rec.mimeType }));
        };
        rec.start(); this.recording = rec; this.recSecs = 0;
        this._recT = setInterval(() => { this.recSecs++; if (this.recSecs >= 30) rec.stop(); }, 1000);
      } catch (e) { this.recErr = 'The browser did not allow the microphone.'; }
    },
    previewTune() { if (this.tuneOn) { stopTune(); this.tuneOn = false; return; } startTune(this.pbx.settings.tune, this.pbx.settings.tuneGreeting); this.tuneOn = true; },
    // on hold the customer hears the care tune; the agent can listen in to what they hear
    toggleHear() { if (this.hearHold) { stopTune(); this.hearHold = false; } else { startTune(this.pbx.settings.tune, false); this.hearHold = true; } },
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

<style>
/* pbx- classes are shared by the Calls management tabs (components/pbx) */
.pbx { display: flex; flex-direction: column; gap: 16px; max-width: 1280px; }
.pbx h1 { margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 24px; }
.pbx h2 { margin: 0 0 10px; font-size: 15px; }
.pbx-head h1 { flex: 1 1 auto; }
.pbx-note { padding: 10px 12px; border-radius: 12px; background: var(--k-1A2520,#1A2520); font-size: 13px; }
.pbx-head { display: flex; flex-wrap: wrap; align-items: center; gap: 10px 14px; }
.pbx-title { flex: 1 1 260px; display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.pbx-sub { font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); }
.pbx .block { display: block; }
.pbx .clip { max-width: 260px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.pbx .nowrap { white-space: nowrap; }
.pbx .grow { flex: 1; min-width: 0; }
.pbx .warn { color: var(--k-FFB86B,#FFB86B); }
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
.pbx-field { display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); }
.pbx-field input, .pbx-field select, .pbx-field textarea, .pbx textarea, .pbx-search { box-sizing: border-box; width: 100%; min-height: 40px; padding: 8px 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font: inherit; font-size: 14px; }
.pbx-search { width: auto; flex: 1 1 220px; }
.pbx-check { display: flex; align-items: center; gap: 8px; font-size: 13px; }
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
.pbx-callscreen { display: flex; flex-direction: column; min-height: min(560px, calc(100vh - 220px)); border-radius: 20px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C); overflow: hidden; }
.pbx-cs-top { display: flex; align-items: center; gap: 10px; padding: 12px 18px; }
.pbx-cs-q { flex: 1; min-width: 0; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.pbx-chip-rec { height: 30px; padding: 0 12px; border-radius: 15px; background: var(--k-3A1E1A,#3A1E1A); color: var(--k-FF9B8A,#FF9B8A); font-size: 12px; font-weight: 700; display: flex; align-items: center; gap: 6px; }
.pbx-chip-rec span { width: 8px; height: 8px; border-radius: 50%; background: #FF9B8A; }
.pbx-chip-time { height: 30px; padding: 0 12px; border-radius: 15px; background: var(--k-16201C,#16201C); font-family: 'Space Grotesk', sans-serif; font-size: 14px; display: flex; align-items: center; font-variant-numeric: tabular-nums; }
.pbx-cs-body { flex: 1; min-height: 0; display: flex; gap: 12px; padding: 0 18px 12px; }
.pbx-cs-main { flex: 1; min-width: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 10px; padding: 20px 0; text-align: center; }
.pbx-cs-face { opacity: 0.9; }
.pbx-cs-face.talking { animation: ring 1.2s ease-out infinite; opacity: 1; }
.pbx-cs-face.ringing { animation: pbxbreathe 1.6s ease-in-out infinite; }
@keyframes pbxbreathe { 50% { transform: scale(1.05); } }
@media (prefers-reduced-motion: reduce) { .pbx-cs-face.talking, .pbx-cs-face.ringing { animation: none; } }
.pbx-cs-name { font-family: 'Space Grotesk', sans-serif; font-size: 22px; }
.pbx-cs-state { font-size: 14px; color: var(--k-B8F36B,#B8F36B); font-weight: 600; }
.pbx-cs-state.hold { color: var(--k-FFB86B,#FFB86B); }
.pbx-cs-facts { display: flex; flex-wrap: wrap; justify-content: center; gap: 6px; max-width: 640px; }
.pbx-cs-facts span { padding: 4px 10px; border-radius: 12px; background: var(--k-16201C,#16201C); font-size: 12px; color: var(--k-C9D6CF,#C9D6CF); }
.pbx-cs-side { flex: 0 0 300px; max-width: 100%; display: flex; flex-direction: column; gap: 12px; padding: 14px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); }
@media (max-width: 760px) { .pbx-cs-body { flex-direction: column; } .pbx-cs-side { flex-basis: auto; } }
.pbx-x { width: 32px; height: 32px; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 18px; }
.pbx-cs-bar { flex: none; display: flex; flex-wrap: wrap; justify-content: center; gap: 10px; padding: 12px 18px 18px; }
.pbx-ctl { width: 66px; height: 60px; border-radius: 16px; border: 0; cursor: pointer; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 3px; background: var(--k-1D2A25,#1D2A25); color: var(--k-E8EFEA,#E8EFEA); }
.pbx-ctl span { font-size: 11px; white-space: nowrap; }
.pbx-ctl.on { background: #B8F36B; color: #0F1714; }
.pbx-ctl.warn, .pbx-ctl.rec { background: #FF9B8A; color: #0F1714; }
.pbx-ctl.end { background: #E5484D; color: #fff; }
@media (max-width: 520px) { .pbx-ctl { width: 54px; height: 54px; } .pbx-cs-bar { gap: 6px; padding: 10px; } }
.pbx-dial { display: flex; justify-content: center; }
.pbx-dial > .pbx-card { width: 100%; max-width: 420px; }
.pbx-chip-q { height: 30px; padding: 0 12px; border-radius: 15px; font-size: 12px; font-weight: 600; display: flex; align-items: center; }
.pbx-chip-q.ok { background: rgba(184,243,107,.12); color: #B8F36B; } .pbx-chip-q.mid { background: rgba(255,184,107,.14); color: #FFB86B; } .pbx-chip-q.bad { background: rgba(255,155,138,.16); color: #FF9B8A; }
.pbx-alert { padding: 10px 14px; border-radius: 12px; background: rgba(255,155,138,.12); border: 1px solid #FF9B8A; color: var(--k-FFD3CA,#FFD3CA); font-size: 13px; }
.pbx-key { width: 32px; height: 32px; flex: none; border-radius: 8px; background: var(--k-22302A,#22302A); display: flex; align-items: center; justify-content: center; font-weight: 700; }
.pbx-sub-card { padding: 12px; border-radius: 12px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; gap: 10px; }
.pbx-field.inline { flex-direction: row; align-items: center; gap: 8px; }
.pbx-field.inline input, .pbx-field.inline select { width: auto; min-width: 70px; }
.pbx-progress { height: 8px; border-radius: 4px; background: var(--k-121B17,#121B17); overflow: hidden; }
.pbx-progress span { display: block; height: 100%; background: #B8F36B; }
.pbx-legend { display: flex; gap: 14px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); }
.pbx-legend i { display: inline-block; width: 10px; height: 10px; border-radius: 3px; margin-right: 6px; vertical-align: -1px; }
.pbx-legend .in, .pbx-chart .in { background: #7CC4FF; fill: #7CC4FF; } .pbx-legend .out, .pbx-chart .out { background: #B8F36B; fill: #B8F36B; } .pbx-legend .bad, .pbx-chart .bad { background: #FF9B8A; fill: #FF9B8A; }
.pbx-chart { width: 100%; height: auto; }
.pbx-gridline { stroke: var(--k-24322C,#24322C); stroke-width: 1; }
.pbx-axis { fill: var(--k-9DB0A6,#9DB0A6); font-size: 10px; }
.pbx-hours { display: flex; align-items: flex-end; gap: 6px; height: 140px; }
.pbx-hour { flex: 1; height: 100%; display: flex; flex-direction: column; justify-content: flex-end; align-items: center; gap: 4px; }
.pbx-hour span { width: 100%; border-radius: 4px 4px 0 0; background: #7CC4FF; }
.pbx-hour small { font-size: 10px; color: var(--k-9DB0A6,#9DB0A6); }
.pbx-bar { flex: 0 0 40%; height: 8px; border-radius: 4px; background: var(--k-121B17,#121B17); overflow: hidden; }
.pbx-bar span { display: block; height: 100%; background: #C9B6FF; }
.pbx-row.wrap { flex-wrap: wrap; }
.pbx-top { position: sticky; top: -24px; z-index: 20; margin-top: -24px; padding-top: 24px; background: var(--k-0F1714,#0F1714); display: flex; flex-direction: column; gap: 12px; }
.pbx-layout { display: flex; align-items: flex-start; gap: 16px; }
.pbx-dialcol { flex: 0 0 300px; position: sticky; top: 120px; }
.pbx-dialcol > .pbx-card { width: 100%; max-width: none; }
.pbx-center { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 16px; }
.pbx-callpanel { flex: none; position: sticky; top: 120px; max-height: calc(100vh - 200px); display: flex; border-radius: 20px; }
.pbx-callpanel .pbx-callscreen { flex: 1; min-width: 0; min-height: 0; max-height: 100%; overflow-y: auto; }
.pbx-grip { position: absolute; left: -7px; top: 0; bottom: 0; width: 14px; padding: 0; border: 0; background: transparent; cursor: ew-resize; z-index: 3; touch-action: none; display: flex; align-items: center; justify-content: center; }
.pbx-grip span { width: 4px; height: 44px; border-radius: 2px; background: var(--k-3A4A42,#3A4A42); }
.pbx-grip:hover span, .pbx-grip:focus-visible span { background: #B8F36B; }
/* the call in the side panel is compact, like Hub's side-panel call screen */
.pbx-callpanel .pbx-cs-body { flex-direction: column; padding: 0 14px 10px; }
.pbx-callpanel .pbx-cs-side { flex-basis: auto; }
.pbx-callpanel .pbx-cs-main { padding: 10px 0; }
.pbx-callpanel .pbx-cs-face { width: 72px !important; height: 72px !important; font-size: 24px !important; }
.pbx-callpanel .pbx-cs-name { font-size: 18px; }
.pbx-callpanel .pbx-cs-bar { gap: 6px; padding: 10px 12px 14px; }
.pbx-callpanel .pbx-ctl { width: 46px; height: 46px; border-radius: 14px; }
.pbx-callpanel .pbx-ctl span { display: none; }
.pbx-callpanel .pbx-cs-top { flex-wrap: wrap; padding: 10px 14px; }
.pbx-callscreen { min-height: 0 !important; }
@media (max-width: 1100px) {
  .pbx-layout { flex-direction: column; align-items: stretch; }
  .pbx-dialcol { position: static; flex-basis: auto; }
  .pbx-callpanel { position: fixed; inset: 64px 0 64px 0; width: auto !important; max-height: none; z-index: 60; border-radius: 0; background: var(--k-121B17,#121B17); }
  .pbx-grip { display: none; }
}
</style>
