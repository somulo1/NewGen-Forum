// Connects the Hub logic to the backend after sign-in: sets up this device's keys, then loads each widget's data from the
// API into the logic state. Each widget has its own file in widgets/ exporting load(logic, srv) and, where it writes, actions.
import { Device, unb64, b64 } from './e2ee.js';
import { Vault } from './vault.js';
import * as people from './widgets/people.js';
import * as comms from './widgets/comms.js';
import * as prefs from './widgets/prefs.js';
import * as calendar from './widgets/calendar.js';
import * as tickets from './widgets/tickets.js';
import * as projects from './widgets/projects.js';
import * as awards from './widgets/awards.js';
import * as profile from './widgets/profile.js';
import * as approvals from './widgets/approvals.js';
import * as notifications from './widgets/notifications.js';

const WIDGETS = [people, profile, prefs, comms, calendar, tickets, projects, awards, approvals, notifications];

async function openDevice(request, session) {
  const account = session.account;
  const mine = (await request('GET', '/e2ee/devices')).devices.filter((d) => !d.revoked);
  const has = (dev) => mine.find((d) => d.identity_key === b64(dev.identity) && d.signing_key === b64(dev.signing));
  // 1. a seeded demo person: the device is derived from their email (development only)
  const demo = Device.fromDemo(session.email);
  if (has(demo)) return { device: demo, deviceId: has(demo).id };
  // 2. a device this browser registered earlier
  const slot = 'hub-device:' + account;
  try {
    const saved = JSON.parse(localStorage.getItem(slot) || 'null');
    if (saved) { const d = Device.import(saved); if (has(d)) return { device: d, deviceId: has(d).id }; }
  } catch (e) { /* unreadable: register a new one */ }
  // 3. register a new device for this browser
  const device = Device.generate();
  const reg = await request('POST', '/e2ee/devices', {
    label: 'Web browser', identity_key: b64(device.identity), signing_key: b64(device.signing), binding_sig: b64(device.binding(account))
  });
  try { localStorage.setItem(slot, JSON.stringify(device.export())); } catch (e) { /* keys last for this tab only */ }
  return { device, deviceId: reg.id };
}

export async function startServerSync(logic, request, session) {
  if (logic.srv) return logic.srv;
  const { device, deviceId } = await openDevice(request, session);
  const me = await request('GET', '/me');
  const vault = new Vault(request, { org: session.org.id, account: session.account, person: session.org.person, device, deviceId });
  const srv = { request, vault, session, me, personOf: new Map(), nameOf: new Map(), reload: {} };
  logic.srv = srv;
  vault.custody().catch(() => {});   // hand keys to colleagues' new devices if they are waiting
  for (const w of WIDGETS) {
    try { await w.load(logic, srv); } catch (e) { console.warn('Could not load ' + (w.name || 'a widget') + ':', e); }
  }
  logic.forceUpdate();
  return srv;
}

export { unb64 };
