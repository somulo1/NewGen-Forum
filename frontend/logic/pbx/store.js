// Calls management: the customer phone system (a softphone and call centre, in the manner of a business PBX app).
// FRONT END ONLY. Nothing here reaches a phone network or the Hub server: calls, messages, voicemail and the wallboard are
// simulated in the browser so the screens can be built and reviewed. Every place that must talk to the real PBX is marked "PBX:".
// What the viewer changes (history, notes, settings) is kept in this browser only, so a reload does not lose the demo.
import { reactive } from 'vue';
import * as snd from './sounds.js';

const KEY = 'hub.pbx.v1';
const now = () => Date.now();
const MIN = 60000, HOUR = 3600000, DAY = 86400000;

// ---- sample data (replace with the PBX and CRM feeds) ----
const CUSTOMERS = [
  { id: 'c1', name: 'Grace Wanjiku', phone: '+254712345678', account: 'LN-20431', product: 'Personal loan', balance: 48250, status: 'Current', kyc: 'Verified', tags: ['VIP'], lastContact: now() - 2 * DAY },
  { id: 'c2', name: 'Peter Otieno', phone: '+254722987654', account: 'LN-18822', product: 'Business loan', balance: 312000, status: 'In arrears (12 days)', kyc: 'Verified', tags: ['Collections'], lastContact: now() - 5 * HOUR },
  { id: 'c3', name: 'Mercy Achieng', phone: '+254733112233', account: 'WL-55210', product: 'Wallet', balance: 1520, status: 'Active', kyc: 'Pending documents', tags: ['KYC'], lastContact: now() - 9 * DAY },
  { id: 'c4', name: 'John Kamau', phone: '+254701445566', account: 'LN-30017', product: 'Asset finance', balance: 905400, status: 'Current', kyc: 'Verified', tags: [], lastContact: now() - 30 * DAY },
  { id: 'c5', name: 'Faith Njeri', phone: '+254745778899', account: 'SV-77120', product: 'Savings', balance: 64000, status: 'Active', kyc: 'Verified', tags: ['New'], lastContact: now() - 1 * DAY },
  { id: 'c6', name: 'Brian Mutua', phone: '+254790223344', account: 'LN-41200', product: 'Personal loan', balance: 22100, status: 'In arrears (40 days)', kyc: 'Verified', tags: ['Collections', 'Promise to pay'], lastContact: now() - 3 * HOUR },
  { id: 'c7', name: 'Aisha Hassan', phone: '+254711556677', account: 'WL-60321', product: 'Wallet', balance: 300, status: 'Frozen', kyc: 'Flagged', tags: ['Fraud review'], lastContact: now() - 12 * DAY }
];

const AGENTS = [
  { ext: '1001', name: 'You', me: true, status: 'Available', queues: ['Collections', 'Customer care'], calls: 0, talk: 0 },
  { ext: '1002', name: 'Kevin Ondede', status: 'On a call', queues: ['Collections'], calls: 14, talk: 52 * MIN, since: now() - 3 * MIN, with: '+254722987654' },
  { ext: '1003', name: 'Salmon Eze', status: 'Available', queues: ['Customer care', 'Onboarding'], calls: 9, talk: 31 * MIN },
  { ext: '1004', name: 'Daisy Moraa', status: 'Wrap-up', queues: ['Customer care'], calls: 11, talk: 40 * MIN, since: now() - 40000 },
  { ext: '1005', name: 'Tom Kiprono', status: 'Away', queues: ['Onboarding'], calls: 6, talk: 18 * MIN },
  { ext: '1006', name: 'Lucy Wairimu', status: 'Ringing', queues: ['Collections', 'Customer care'], calls: 12, talk: 45 * MIN, since: now() - 8000 },
  { ext: '1007', name: 'Ian Barasa', status: 'Logged out', queues: ['Fraud'], calls: 0, talk: 0 }
];

const QUEUES = [
  { name: 'Customer care', number: '6001', waiting: 3, longest: 95, answered: 128, abandoned: 6, sla: 91, target: 20, avgWait: 18, avgTalk: 214 },
  { name: 'Collections', number: '6002', waiting: 1, longest: 40, answered: 74, abandoned: 9, sla: 84, target: 30, avgWait: 26, avgTalk: 296 },
  { name: 'Onboarding', number: '6003', waiting: 0, longest: 0, answered: 41, abandoned: 1, sla: 97, target: 20, avgWait: 9, avgTalk: 180 },
  { name: 'Fraud', number: '6004', waiting: 0, longest: 0, answered: 12, abandoned: 0, sla: 100, target: 10, avgWait: 4, avgTalk: 402 }
];

const DISPOSITIONS = ['Resolved', 'Promise to pay', 'Payment made', 'Call back requested', 'Wrong number', 'No answer', 'Escalated', 'KYC follow-up', 'Fraud report', 'Complaint'];

// ---- call routing (PBX: these become the PBX's IVR, ring group, time condition and queue settings) ----
function seedRouting() {
  return {
    ivr: { greeting: 'Welcome to our customer care. For loans press 1. For your wallet press 2. To report fraud press 3. To speak to someone press 0.', timeoutSec: 8, retries: 2,
      options: [{ key: '1', to: 'Collections' }, { key: '2', to: 'Customer care' }, { key: '3', to: 'Fraud' }, { key: '0', to: 'Customer care' }, { key: '9', to: 'Repeat menu' }] },
    groups: [{ name: 'Collections team', strategy: 'Longest idle first', members: ['1001', '1002', '1006'], ringSec: 20 }, { name: 'Care team', strategy: 'Ring all', members: ['1001', '1003', '1004'], ringSec: 25 }],
    hours: { mon: ['08:00', '18:00'], tue: ['08:00', '18:00'], wed: ['08:00', '18:00'], thu: ['08:00', '18:00'], fri: ['08:00', '18:00'], sat: ['09:00', '13:00'], sun: null },
    afterHours: 'Voicemail', holidays: [{ date: '2026-12-12', name: 'Jamhuri Day' }, { date: '2026-12-25', name: 'Christmas Day' }, { date: '2026-12-26', name: 'Boxing Day' }],
    callbacks: true, announcePosition: true, announceEverySec: 45, maxWaitSec: 300
  };
}
function seedCampaigns() {
  return [
    { id: 'k1', name: 'Arrears follow-up (Nov)', queue: 'Collections', status: 'Paused', retries: 2, retryMin: 60, window: ['09:00', '17:00'], list: [{ customerId: 'c2', tries: 1, result: 'No answer' }, { customerId: 'c6', tries: 1, result: 'Promise to pay' }, { customerId: 'c4', tries: 0, result: '' }] },
    { id: 'k2', name: 'KYC reminders', queue: 'Onboarding', status: 'Draft', retries: 1, retryMin: 120, window: ['10:00', '16:00'], list: [{ customerId: 'c3', tries: 0, result: '' }, { customerId: 'c7', tries: 0, result: '' }] }
  ];
}
function seedCallbacks() {
  return [{ id: 'cb1', number: '+254745778899', customerId: 'c5', queue: 'Customer care', askedAt: now() - 25 * MIN, status: 'Waiting' }];
}

function seedHistory() {
  const rows = [
    ['in', 'c2', 'Answered', 4 * MIN + 12000, 'Promise to pay', 2 * HOUR, true],
    ['out', 'c6', 'Answered', 6 * MIN, 'Promise to pay', 3 * HOUR, true],
    ['in', null, 'Missed', 0, '', 4 * HOUR, false, '+254799000111'],
    ['out', 'c3', 'No answer', 0, 'No answer', 6 * HOUR, false],
    ['in', 'c5', 'Answered', 2 * MIN + 40000, 'Resolved', DAY, true],
    ['out', 'c1', 'Answered', 9 * MIN, 'Resolved', 2 * DAY, true],
    ['in', 'c7', 'Answered', 12 * MIN, 'Fraud report', 3 * DAY, true],
    ['in', null, 'Voicemail', 0, '', 3 * DAY + HOUR, false, '+254708123456']
  ];
  return rows.map(([dir, cid, result, dur, disp, ago, rec, raw], i) => ({ id: 'h' + i, dir, customerId: cid, number: raw || CUSTOMERS.find((c) => c.id === cid).phone, result, duration: dur, disposition: disp, at: now() - ago, recorded: rec, note: '', agent: 'You' }));
}

function seedThreads() {
  return [
    { id: 't1', customerId: 'c6', channel: 'SMS', unread: 1, msgs: [{ from: 'me', text: 'Hello Brian, a reminder that KES 4,500 is due on Friday. Reply PAY to get the paybill details.', at: now() - DAY }, { from: 'them', text: 'PAY', at: now() - 20 * HOUR }, { from: 'me', text: 'Paybill 400200, account LN-41200. Thank you.', at: now() - 20 * HOUR + MIN }, { from: 'them', text: 'Will pay on Friday evening', at: now() - 2 * HOUR }] },
    { id: 't2', customerId: 'c3', channel: 'WhatsApp', unread: 0, msgs: [{ from: 'me', text: 'Hi Mercy, please upload a clear photo of your ID to finish verification.', at: now() - 3 * DAY }, { from: 'them', text: 'Sent it now', at: now() - 2 * DAY }] },
    { id: 't3', customerId: 'c1', channel: 'SMS', unread: 0, msgs: [{ from: 'me', text: 'Your loan top-up of KES 20,000 has been approved.', at: now() - 2 * DAY }] }
  ];
}

function seedVoicemail() {
  return [
    { id: 'v1', box: 'Customer care', number: '+254708123456', customerId: null, at: now() - 3 * DAY - HOUR, duration: 38000, heard: false, transcript: 'Hi, I was charged twice for my wallet top-up yesterday. Please call me back on this number.' },
    { id: 'v2', box: 'My voicemail', number: '+254722987654', customerId: 'c2', at: now() - 26 * HOUR, duration: 21000, heard: true, transcript: 'This is Peter, about the business loan. I will be in the office after 2 pm.' }
  ];
}

const DEFAULT_SETTINGS = {
  extension: '1001', callerId: 'Main line (0709 000 000)', callerIds: ['Main line (0709 000 000)', 'Collections (0709 000 100)', 'Customer care (0709 000 200)'],
  forwardAlways: '', forwardBusy: '', forwardNoAnswer: '', noAnswerSec: 25, dnd: false, autoAnswer: false,
  ringDevice: 'Browser', mic: '', speaker: '', alertWaitSec: 60, alertSla: 85, wrapUpSec: 30, tune: 'Calm care', tuneGreeting: true, maskNumbers: true, greeting: 'Default greeting', blocked: ['+254799000111'],
  queues: { 'Collections': true, 'Customer care': true, 'Onboarding': false, 'Fraud': false }
};

function load() {
  try { const raw = typeof localStorage !== 'undefined' && localStorage.getItem(KEY); return raw ? JSON.parse(raw) : null; } catch (e) { return null; }
}

const saved = load() || {};
export const pbx = reactive({
  customers: CUSTOMERS,
  agents: AGENTS,
  queues: QUEUES,
  dispositions: DISPOSITIONS,
  history: saved.history || seedHistory(),
  threads: saved.threads || seedThreads(),
  voicemail: saved.voicemail || seedVoicemail(),
  routing: saved.routing || seedRouting(),
  campaigns: saved.campaigns || seedCampaigns(),
  callbacks: saved.callbacks || seedCallbacks(),
  dnc: saved.dnc || ['+254711556677'],              // customers who asked not to be called (do-not-call list)
  greetings: saved.greetings || [],                  // my recorded voicemail greetings (data URLs, this browser only)
  alerts: [],
  alertsOff: {},                                     // alerts closed with ×: they stay away until that problem clears                                        // wallboard alerts raised by the clock
  quality: { mos: 4.3, jitter: 8, loss: 0.2, rtt: 64 },   // the live call's line quality
  settings: { ...DEFAULT_SETTINGS, ...(saved.settings || {}) },
  presence: saved.presence || 'Available',          // Available, Away, Busy, Do not disturb, Lunch, Off work
  call: null,                                        // the live call, see newCall()
  held: [],                                          // other calls on hold while I talk on one
  incoming: null,                                    // a call ringing for me
  wrap: null,                                        // after-call work: { callId, left, disposition, note }
  parked: [],                                        // calls parked on a slot anyone can pick up
  tick: now(),
  audit: []                                          // who revealed a number or listened in, for compliance
});

function persist() {
  try { localStorage.setItem(KEY, JSON.stringify({ history: pbx.history.slice(0, 300), threads: pbx.threads, voicemail: pbx.voicemail, settings: pbx.settings, presence: pbx.presence, routing: pbx.routing, campaigns: pbx.campaigns, callbacks: pbx.callbacks, dnc: pbx.dnc, greetings: pbx.greetings })); } catch (e) { /* private window */ }
}

// ---- helpers ----
export const customerOf = (id) => pbx.customers.find((c) => c.id === id) || null;
export const customerByNumber = (n) => pbx.customers.find((c) => norm(c.phone) === norm(n)) || null;
export function norm(n) {
  let d = String(n || '').replace(/[^\d+]/g, '');
  if (/^0[17]\d{8}$/.test(d)) d = '+254' + d.slice(1);
  if (/^254\d{9}$/.test(d)) d = '+' + d;
  return d;
}
// numbers are shown masked unless the agent reveals one (and the reveal is logged): customer numbers are personal data
export function showNumber(n, reveal) {
  const d = norm(n);
  if (!pbx.settings.maskNumbers || reveal || d.length < 8) return pretty(d);
  return pretty(d).replace(/\d(?=(?:\D*\d){3})/g, (m, i) => (i < 5 ? m : '•'));
}
export function pretty(d) {
  const m = /^\+254(\d{3})(\d{3})(\d{3})$/.exec(d);
  return m ? '+254 ' + m[1] + ' ' + m[2] + ' ' + m[3] : d;
}
export function fmtDur(ms) {
  const s = Math.max(0, Math.round(ms / 1000));
  const h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60), x = s % 60;
  return (h ? h + ':' + String(m).padStart(2, '0') : m) + ':' + String(x).padStart(2, '0');
}
export function fmtWhen(t) {
  const d = new Date(t), today = new Date();
  const same = d.toDateString() === today.toDateString();
  const yest = new Date(today.getTime() - DAY).toDateString() === d.toDateString();
  const time = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  return same ? time : yest ? 'Yesterday ' + time : d.toLocaleDateString([], { day: 'numeric', month: 'short' }) + ' ' + time;
}
export const money = (n) => 'KES ' + Number(n || 0).toLocaleString();
export function logAudit(what) { pbx.audit.unshift({ at: now(), what }); pbx.audit.splice(50); }

// ---- calls (PBX: each of these becomes a request to the PBX's call-control API, and the state comes back by events) ----
let seq = 0;
function newCall(dir, number, extra) {
  const c = customerByNumber(number);
  return { id: 'call' + (++seq) + '-' + now(), dir, number: norm(number), customerId: c ? c.id : null, state: dir === 'out' ? 'Dialling' : 'Talking', startedAt: now(), answeredAt: dir === 'in' ? now() : null,
    muted: false, onHold: false, recording: true,   // every call is recorded, start to end, and cannot be paused: calls must be auditable
     dtmf: '', note: '', conference: [], transferTo: '', transferMode: null, queue: (extra && extra.queue) || '', ...extra };
}

export function dial(number, opts) {
  opts = opts || {};
  const n = norm(number);
  if (!n || n.replace(/\D/g, '').length < 3) return 'Enter a number to call.';
  if (pbx.settings.blocked.includes(n)) return 'This number is blocked. Remove it from Settings > Blocked numbers to call it.';
  if (pbx.dnc.includes(n) && !opts.override) return 'This customer is on the do-not-call list. Remove them in Campaigns > Do not call to call them.';
  if (pbx.call) holdCurrent();
  pbx.call = newCall('out', n, { callerId: pbx.settings.callerId, campaign: opts.campaign || '' });
  const id = pbx.call.id;
  snd.startRingback();
  // PBX: the far end answers (or not); simulated after a few seconds
  setTimeout(() => { if (pbx.call && pbx.call.id === id && pbx.call.state === 'Dialling') pbx.call.state = 'Ringing'; }, 900);
  setTimeout(() => {
    if (!pbx.call || pbx.call.id !== id || pbx.call.state !== 'Ringing') return;
    snd.stopRingback();
    if (Math.random() < 0.15) { endCall('No answer'); return; }
    pbx.call.state = 'Talking'; pbx.call.answeredAt = now(); snd.connected();
  }, 3200);
  return '';
}

export function simulateIncoming(queue) {
  if (pbx.incoming) return;
  if (pbx.settings.dnd || pbx.presence === 'Do not disturb' || pbx.presence === 'Off work') {
    pbx.history.unshift({ id: 'h' + now(), dir: 'in', customerId: null, number: '+2547' + Math.floor(10000000 + Math.random() * 89999999), result: 'Missed (DND)', duration: 0, disposition: '', at: now(), recorded: false, note: '', agent: 'You' });
    persist(); return;
  }
  const c = pbx.customers[Math.floor(Math.random() * pbx.customers.length)];
  snd.startRing();
  pbx.incoming = { number: c.phone, customerId: c.id, queue: queue || (Math.random() < 0.5 ? 'Customer care' : 'Collections'), at: now(), waited: Math.floor(Math.random() * 60) };
  if (pbx.settings.autoAnswer) setTimeout(answer, 1500);
}

export function answer() {
  const inc = pbx.incoming; if (!inc) return;
  if (pbx.call) holdCurrent();
  pbx.call = newCall('in', inc.number, { queue: inc.queue });
  pbx.incoming = null; snd.stopRing(); snd.connected();
}
export function decline(toVoicemail) {
  const inc = pbx.incoming; if (!inc) return;
  snd.stopRing();
  pbx.history.unshift({ id: 'h' + now(), dir: 'in', customerId: inc.customerId, number: inc.number, result: toVoicemail ? 'Sent to voicemail' : 'Declined', duration: 0, disposition: '', at: now(), recorded: false, note: '', agent: 'You' });
  pbx.incoming = null; persist();
}

function holdCurrent() { if (!pbx.call) return; pbx.call.onHold = true; pbx.held.push(pbx.call); pbx.call = null; }
export function toggleHold() { if (pbx.call) pbx.call.onHold = !pbx.call.onHold; }
export function swapTo(id) {
  const i = pbx.held.findIndex((h) => h.id === id); if (i === -1) return;
  const next = pbx.held.splice(i, 1)[0];
  if (pbx.call) { pbx.call.onHold = true; pbx.held.push(pbx.call); }
  next.onHold = false; pbx.call = next;
}
export function sendDtmf(k) { snd.keyTone(k); if (pbx.call) pbx.call.dtmf = (pbx.call.dtmf + k).slice(-24); }

export function endCall(result) {
  const c = pbx.call; if (!c) return;
  const answered = !!c.answeredAt;
  snd.stopRingback(); snd.hangup();
  const row = { id: 'h' + now(), dir: c.dir, customerId: c.customerId, number: c.number, result: result || (answered ? 'Answered' : 'Cancelled'), duration: answered ? now() - c.answeredAt : 0, disposition: '', at: c.startedAt, recorded: answered, note: c.note, agent: 'You', queue: c.queue, conference: c.conference.slice() };
  pbx.history.unshift(row);
  pbx.call = pbx.held.length ? Object.assign(pbx.held.pop(), { onHold: false }) : null;
  if (answered) pbx.wrap = { historyId: row.id, left: pbx.settings.wrapUpSec, disposition: '', note: c.note };
  persist();
}

// transfer: blind sends it straight on; attended talks to the other side first, then completes or cancels
export function transfer(target, mode) {
  const c = pbx.call; if (!c || !target) return;
  if (mode === 'blind') { c.note = (c.note ? c.note + '\n' : '') + 'Transferred to ' + target + '.'; endCall('Transferred to ' + target); return; }
  c.transferTo = target; c.transferMode = 'consulting'; c.onHold = true;
}
export function completeTransfer() { const c = pbx.call; if (!c || !c.transferTo) return; endCall('Transferred to ' + c.transferTo + ' (attended)'); }
export function cancelTransfer() { const c = pbx.call; if (!c) return; c.transferTo = ''; c.transferMode = null; c.onHold = false; }
export function addToConference(who) { const c = pbx.call; if (!c || !who || c.conference.length >= 5) return; c.conference.push({ who, joinedAt: now() }); }
export function dropFromConference(i) { const c = pbx.call; if (c) c.conference.splice(i, 1); }
export function park() {
  const c = pbx.call; if (!c) return;
  const slot = 700 + pbx.parked.length + 1;
  pbx.parked.push({ slot, call: { ...c, onHold: true }, at: now() });
  pbx.call = pbx.held.length ? Object.assign(pbx.held.pop(), { onHold: false }) : null;
}
export function unpark(slot) {
  const i = pbx.parked.findIndex((p) => p.slot === slot); if (i === -1) return;
  if (pbx.call) holdCurrent();
  pbx.call = Object.assign(pbx.parked.splice(i, 1)[0].call, { onHold: false });
}

export function finishWrap() {
  const w = pbx.wrap; if (!w) return;
  const h = pbx.history.find((x) => x.id === w.historyId);
  if (h) { h.disposition = w.disposition; h.note = w.note; }
  pbx.wrap = null; persist();
}

// ---- messages ----
export function sendMessage(threadId, text, channel) {
  const t = pbx.threads.find((x) => x.id === threadId); if (!t || !text.trim()) return;
  t.msgs.push({ from: 'me', text: text.trim(), at: now(), status: 'Sent' }); t.unread = 0;
  if (channel) t.channel = channel;
  persist();
}
export function startThread(customerId, channel) {
  let t = pbx.threads.find((x) => x.customerId === customerId);
  if (!t) { t = { id: 't' + now(), customerId, channel: channel || 'SMS', unread: 0, msgs: [] }; pbx.threads.unshift(t); persist(); }
  return t.id;
}

// ---- monitoring (PBX: supervisor features need the PBX's spy / whisper / barge commands) ----
export function supervise(agent, how) { logAudit(how + ' on ' + agent.name + ' (ext ' + agent.ext + ')'); return how; }
export function setAgentQueue(name, on) { pbx.settings.queues[name] = on; persist(); }
export function setPresence(p) { pbx.presence = p; const me = pbx.agents.find((a) => a.me); if (me) me.status = p === 'Available' ? 'Available' : p; persist(); }
export function saveSettings() { persist(); }
export function markHeard(v) { v.heard = true; persist(); }
export function deleteVoicemail(id) { pbx.voicemail = pbx.voicemail.filter((v) => v.id !== id); persist(); }

// ---- callbacks, campaigns, do-not-call, greetings ----
export function callBack(cb) { const err = dial(cb.number); if (!err) { cb.status = 'Called back'; persist(); } return err; }
export function dropCallback(id) { pbx.callbacks = pbx.callbacks.filter((c) => c.id !== id); persist(); }
export function saveRouting() { persist(); }
export function dismissAlert(a) { pbx.alertsOff[a.queue] = a.text; pbx.alerts = pbx.alerts.filter((x) => x !== a); }
export function campaignNext(k) {
  // the next customer due: not reached yet, tries left, and not on the do-not-call list
  return k.list.find((r) => !/Promise to pay|Payment made|Resolved|Do not call/.test(r.result) && r.tries <= k.retries && !pbx.dnc.includes(norm((customerOf(r.customerId) || {}).phone))) || null;
}
export function campaignDial(k) {
  const r = campaignNext(k); if (!r) { k.status = 'Finished'; persist(); return 'Everyone on this list has been reached or has no tries left.'; }
  const c = customerOf(r.customerId);
  const err = dial(c.phone, { campaign: k.name });
  if (!err) { r.tries++; r.result = 'Dialled'; k.status = 'Running'; persist(); }
  return err;
}
export function campaignResult(k, customerId, result) { const r = k.list.find((x) => x.customerId === customerId); if (r) { r.result = result; if (result === 'Do not call') addDnc(customerOf(customerId).phone); persist(); } }
export function newCampaign(name, queue, ids) { pbx.campaigns.unshift({ id: 'k' + now(), name, queue, status: 'Draft', retries: 2, retryMin: 60, window: ['09:00', '17:00'], list: ids.map((id) => ({ customerId: id, tries: 0, result: '' })) }); persist(); }
export function addDnc(n) { n = norm(n); if (n && !pbx.dnc.includes(n)) { pbx.dnc.push(n); persist(); } }
export function removeDnc(n) { pbx.dnc = pbx.dnc.filter((x) => x !== n); persist(); }
export function addGreeting(g) { pbx.greetings.unshift(g); pbx.greetings.splice(5); persist(); }
export function removeGreeting(i) { pbx.greetings.splice(i, 1); persist(); }

// ---- the clock: call timers, wrap-up countdown, and a live-looking wallboard ----
let timer = 0;
export function startClock() {
  if (timer || typeof window === 'undefined') return;
  timer = window.setInterval(() => {
    pbx.tick = now();
    if (pbx.wrap && pbx.wrap.left > 0) { pbx.wrap.left--; if (pbx.wrap.left === 0) finishWrap(); }
    if (pbx.call && pbx.call.answeredAt) {
      const q = pbx.quality, j = (x, d, lo, hi) => Math.min(hi, Math.max(lo, x + (Math.random() - 0.5) * d));
      q.jitter = Math.round(j(q.jitter, 6, 2, 60)); q.loss = +j(q.loss, 0.6, 0, 6).toFixed(1); q.rtt = Math.round(j(q.rtt, 20, 30, 400));
      q.mos = +Math.max(1, Math.min(4.5, 4.5 - q.jitter / 40 - q.loss / 2 - q.rtt / 600)).toFixed(1);
    }
    pbx.queues.forEach((q) => {
      const key = q.name;
      const bad = q.longest > pbx.settings.alertWaitSec ? 'Longest wait in ' + key + ' is over ' + pbx.settings.alertWaitSec + 's' : q.sla < pbx.settings.alertSla ? key + ' service level is below ' + pbx.settings.alertSla + '%' : '';
      const i = pbx.alerts.findIndex((a) => a.queue === key);
      if (!bad) delete pbx.alertsOff[key];
      if (bad && i === -1 && pbx.alertsOff[key] !== bad) pbx.alerts.unshift({ queue: key, text: bad, at: now() });
      if (!bad && i !== -1) pbx.alerts.splice(i, 1);
    });
    if (Math.random() < 0.08) {
      const q = pbx.queues[Math.floor(Math.random() * pbx.queues.length)];
      q.waiting = Math.max(0, q.waiting + (Math.random() < 0.5 ? -1 : 1));
      q.longest = q.waiting ? q.longest + 3 : 0;
      if (Math.random() < 0.4) q.answered++;
    }
  }, 1000);
}
export function stopClock() { if (timer) { clearInterval(timer); timer = 0; } }
