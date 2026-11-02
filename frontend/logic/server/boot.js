// Connects the Hub logic to the backend after sign-in: sets up this device's keys, then loads each widget's data from the
// API into the logic state. Each widget has its own file in widgets/ exporting load(logic, srv) and, where it writes, actions.
import { loadDevice, saveDevice } from './deviceStore.js';
import { Device, unb64, b64 } from './e2ee.js';
import { Vault } from './vault.js';
import { connectRealtime } from './realtime.js';
import * as keyBackup from './keyBackup.js';
import * as callsW from './calls.js';
import * as people from './widgets/people.js';
import * as comms from './widgets/comms.js';
import * as prefs from './widgets/prefs.js';
import * as calendar from './widgets/calendar.js';
import * as tickets from './widgets/tickets.js';
import * as projects from './widgets/projects.js';
import * as awards from './widgets/awards.js';
import * as profile from './widgets/profile.js';
import * as wall from './widgets/wall.js';
import * as forums from './widgets/forums.js';
import * as notes from './widgets/notes.js';
import * as folders from './widgets/folders.js';
import * as dms from './widgets/dms.js';
import * as roles from './widgets/roles.js';
import * as security from './widgets/security.js';
import * as approvals from './widgets/approvals.js';
import * as notifications from './widgets/notifications.js';
import * as orgW from './widgets/org.js';
import * as settingsW from './widgets/settings.js';
import * as typing from './typing.js';
import { createLazy } from './lazy.js';

// chats (forums, dms) come right after the directory they need, so they are not queued behind every other widget
// loaded when a page first needs them (see lazy.js)
const LAZY = [comms, calendar, tickets, projects, awards, approvals, wall, security, notes, folders];

async function openDevice(request, session) {
  const account = session.account;
  const mine = (await request('GET', '/e2ee/devices')).devices.filter((d) => !d.revoked);
  const has = (dev) => mine.find((d) => d.identity_key === b64(dev.identity) && d.signing_key === b64(dev.signing));
  // 1. a device this browser registered earlier
  const slot = 'hub-device:' + account;
  try {
    const saved = await loadDevice(slot);
    if (saved) { const d = Device.import(saved); if (has(d)) return { device: d, deviceId: has(d).id }; }
  } catch (e) { /* unreadable: register a new one */ }
  // 2. register a new device for this browser
  const device = Device.generate();
  const reg = await request('POST', '/e2ee/devices', {
    label: 'Web browser', identity_key: b64(device.identity), signing_key: b64(device.signing), binding_sig: b64(device.binding(account))
  });
  await saveDevice(slot, device.export());
  return { device, deviceId: reg.id };
}

export async function startServerSync(logic, request, session, bytes) {
  // The app in memory belongs to one account. If the session now belongs to another (someone signed in here, or in another tab of
  // this browser), nothing of the first may be shown or used: start again with a clean page.
  if (logic.srv) {
    if (logic.srv.session && session && logic.srv.session.account !== session.account && typeof window !== 'undefined') window.location.reload();
    return logic.srv;
  }
  const { device, deviceId } = await openDevice(request, session);
  keyBackup.finish(request, session.account, device).catch(() => {});   // the account's keys are kept safe on the server for the next browser
  const me = await request('GET', '/me');
  const vault = new Vault(request, { org: session.org.id, account: session.account, person: session.org.person, device, deviceId });
  const srv = { request, bytes, vault, session, me, personOf: new Map(), nameOf: new Map(), reload: {} };
  logic.srv = srv;
  logic.loaded = new Set();
  vault.onNewDevice = (acct, n) => {
    const who = (logic.dir.find((p) => p.uid === vault.personOfAcct.get(acct)) || {}).name || 'A colleague';
    logic.notify(who + ' now has ' + (n === 1 ? 'a new device' : n + ' new devices') + ' that can read your shared chats. If you did not expect that, compare security codes with them (open their profile card).', '#FFB86B');
  };
  vault.custody().catch(() => {});   // hand keys to colleagues' new devices if they are waiting
  // Keys move in real time, like a chat app: when this device cannot open something yet, the colleagues who hold the key are told
  // straight away and one of them hands it over within a second or two; the moment it arrives the screens open what was waiting.
  const rt = srv.rt || (srv.rt = connectRealtime(srv));
  typing.watch(logic, srv);
  let asked = false;
  rt.on('open', () => { if (asked) return; asked = true; srv.request('POST', '/e2ee/keys/request', {}).catch(() => {}); });
  rt.on('keys.requested', () => { setTimeout(() => vault.custody().catch(() => {}), Math.random() * 300); });
  let reloadT = 0;
  rt.on('keys.shared', () => {
    vault.keys.clear();
    clearTimeout(reloadT);
    reloadT = setTimeout(() => { [wall, forums, dms].filter((w) => logic.loaded.has(w.name)).forEach((w) => w.load(logic, srv).catch(() => {})); }, 250);
  });
  // ...and keep doing it while this tab is open: a colleague added to the wall or a forum after you signed in gets their key from
  // whoever of its members is online, so this must not wait for the next sign-in
  clearInterval(srv.custodyT);
  srv.custodyT = setInterval(() => { if (!document.hidden && logic.srv === srv) vault.custody().catch(() => {}); }, 30000);
  // Signed in: the sample data built into the app must not show while (or if) the real data loads. Every list the server
  // fills starts empty, and whatever was open from the samples is closed.
  logic.setState({
    projects: [], tickets: [], events: [], comms: [], posts: [], dms: [], notifs: [], notes: [], noteBooks: [],
    awardList: [], requests: [], callLog: [], callMsgs: [],
    projectOpen: null, ticketOpen: null
  });
  // Who may open what (the directory, the roles and rights, the organization's levels) loads first, so pages are decided at
  // once; everything else fills in behind. Until then no page is called "no access".
  // logic.loaded: which widgets have answered; a page shows grey placeholder blocks until the ones it needs have
  const loadOne = async (w) => {
    const t0 = performance.now();
    try { await w.load(logic, srv); } catch (e) { console.warn('Could not load ' + (w.name || 'a widget') + ':', e); }
    logic.loaded.add(w.name); logic.forceUpdate();
    console.debug('[hub] ' + (w.name || 'widget') + ' loaded in ' + Math.round(performance.now() - t0) + ' ms');
  };
  await loadOne(people);
  // The chats, notifications and settings (what the sidebar and header show on every page) start as soon as the directory is in, beside
  // the rights, the profile, the organization and the preferences that decide which pages open. Nothing else is fetched at sign-in:
  // each page loads its own widgets the first time it opens (see lazy.js), so a person who only reads the wall never downloads the
  // security logs, and sign-in puts a fraction of the load on the server that fetching all of it did.
  const shell = Promise.all([forums, dms, notifications, settingsW].map(loadOne));
  await Promise.all([roles, orgW, prefs, profile].map(loadOne));   // the home layout and the organization's widget settings come before anything is drawn
  logic._accessPending = false;
  logic.forceUpdate();
  await shell;
  const lazy = srv.lazy = createLazy(logic, LAZY);
  // the bell shows overdue tickets and shared notes show in chats, so those two arrive quietly a moment after the first screen
  lazy.later(['tickets', 'notes']);
  // incoming calls ring live; each chat's Calls tab shows the real history
  // a dashboard shows new tickets, communications and requests as they happen (its tabs then blink): drop the cached dashboards so the
  // open one is fetched again
  const staleDash = () => { if (srv.ouAsked) { srv.ouAsked.clear(); logic.forceUpdate(); } };
  // ticket ratings are fetched again the next time the list shows
  try { srv.rt && srv.rt.on('ticket.updated', () => { srv.fbTkPage = undefined; if (logic.state.tkFbList) logic.setState({ tkFbList: null }); }); } catch (e) { /* no live connection */ }
  ['ticket.updated', 'approval.changed', 'notification', 'comm.published'].forEach((t) => { try { srv.rt && srv.rt.on(t, staleDash); } catch (e) { /* no live connection */ } });
  callsW.listen(logic, srv);
  callsW.loadHistory(logic, srv).catch(() => {});
  setTimeout(() => callsW.checkFeedback(logic, srv).catch(() => {}), 4000);
  // what a live notification may need to refresh
  srv.reload.tickets = () => { srv.lazy.refresh(tickets).catch(() => {}); orgW.loadPerformance(logic, srv).catch(() => {}); };
  srv.reload.approvals = () => Promise.all([approvals, comms, awards].map((w) => srv.lazy.refresh(w))).catch(() => {});   // only what was ever opened
  srv.bootDone = true;
  logic.forceUpdate();
  return srv;
}

export { unb64 };
