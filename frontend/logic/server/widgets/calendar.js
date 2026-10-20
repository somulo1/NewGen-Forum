// Calendar and Upcoming events. Each calendar item is an "entry" record in its own event container (members = invitees).
// Dates, times, type and repeat are readable by the server; the title, people, notes and the like stay sealed.
import { newConversationKey, wrap, newId, b64, unb64 } from '../e2ee.js';

export const name = 'calendar';

const iso = (d) => d.toISOString().slice(0, 10);
const addDays = (s, n) => iso(new Date(new Date(s + 'T12:00:00Z').getTime() + n * 86400000));
const todayISO = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' });

async function fetchRange(logic, srv, from, to) {
  const { entries } = await srv.request('GET', `/calendar?from=${from}&to=${to}`);
  const got = await Promise.all(entries.map(async (r) => {
    const pl = (await srv.vault.openRecord(r)) || {};
    return {
      id: r.id, container: r.container, title: pl.title || '(sealed)', type: r.meta.type || 'Event', date: r.meta.date, start: r.meta.start || '', end: r.meta.end || '',
      repeat: r.meta.repeat || 'None', people: pl.people || [], forum: pl.forum || '', by: pl.by || r.author_name, notes: pl.notes || '', visibility: pl.visibility || '',
      remind: pl.remind || '', closedDates: pl.closedDates || [], payload: pl
    };
  }));
  const ids = new Set(got.map((e) => e.id));
  logic.setState((st) => ({ events: st.events.filter((e) => !ids.has(e.id) && !e.fromServer).concat(got.map((e) => ({ ...e, fromServer: true }))).concat(st.events.filter((e) => e.fromServer && !ids.has(e.id))) }));
}

export async function load(logic, srv) {
  srv.calLoaded = new Set();
  logic.setState({ events: [] });
  const t = todayISO();
  await fetchRange(logic, srv, addDays(t, -30), addDays(t, 69));
  ['-1', '0', '1', '2'].forEach((d) => srv.calLoaded.add(monthShift(t.slice(0, 7), +d)));
  const asked = new Map([...upcomingEvents(logic.state.events, t), ...sessionEvents(logic.state.events, t)].map((e) => [e.id, e]));
  await loadRsvps(logic, srv, [...asked.values()].slice(0, 60)).catch(() => {});
}

// The Upcoming events widget: the next few one-off events and meetings on this person's calendar.
// Meetings and standups coming up (including repeating ones): the Home page asks for the person's RSVP on each.
export function sessionEvents(events, today) {
  return events.filter((e) => (e.type === 'Meeting' || e.type === 'Standup') && e.container && (e.repeat !== 'None' || e.date >= today));
}

export function upcomingEvents(events, today) {
  return events.filter((e) => (e.type === 'Event' || e.type === 'Meeting') && e.repeat === 'None' && e.date >= today && !(e.closedDates || []).length).sort((a, b) => (a.date + a.start).localeCompare(b.date + b.start)).slice(0, 3);
}

function monthShift(ym, d) {
  let y = +ym.slice(0, 4), m = +ym.slice(5, 7) + d;
  while (m < 1) { m += 12; y--; }
  while (m > 12) { m -= 12; y++; }
  return y + '-' + String(m).padStart(2, '0');
}

// Called when the calendar page moves to another month.
export async function ensureMonth(logic, srv, ym) {
  if (!srv || srv.calLoaded.has(ym)) return;
  srv.calLoaded.add(ym);
  await fetchRange(logic, srv, addDays(ym + '-01', -7), addDays(monthShift(ym, 1) + '-01', 7));
}

const metaOf = (e) => ({ date: e.date, start: e.start || '', end: e.end || '', type: e.type, repeat: e.repeat || 'None' });
const payloadOf = (e) => ({ title: e.title, people: e.people || [], forum: e.forum || '', by: e.by, notes: e.notes || '', visibility: e.visibility || '', remind: e.remind || '', closedDates: e.closedDates || [] });

// e = { title, type, date, start, end, repeat, people (names), forum, notes, visibility, remind }
export async function create(logic, srv, e) {
  const members = (e.people || []).map((n) => srv.personOf.get(n)).filter(Boolean);
  const { devices } = await srv.request('POST', '/e2ee/people/devices', { people: [...new Set([srv.me.person, ...members])] });
  const key = newConversationKey();
  const keys = await Promise.all(devices.map(async (d) => ({ epoch: 1, device: d.id, wrapped_key: b64(await wrap(key, unb64(d.identity_key))) })));
  const conv = await srv.request('POST', '/conversations', { kind: 'event', members, keys });
  const sealed = await srv.vault.seal(conv.id, payloadOf(e));
  await srv.request('POST', `/containers/${conv.id}/records`, { id: sealed.id, collection: 'entry', epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope, meta: metaOf(e) });
  await fetchRange(logic, srv, addDays(e.date, -7), addDays(e.date, 35));
  return sealed.id;
}

export async function update(logic, srv, ev, patch) {
  const next = { ...ev, ...patch };
  const sealed = await srv.vault.seal(ev.container, payloadOf(next), ev.id);
  await srv.request('PUT', `/containers/${ev.container}/records/${ev.id}`, { collection: 'entry', epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope, meta: metaOf(next) });
  logic.setState((st) => ({ events: st.events.map((x) => (x.id === ev.id ? { ...x, ...patch } : x)) }));
}

export async function remove(logic, srv, ev) {
  await srv.request('DELETE', `/containers/${ev.container}/records/${ev.id}`);
  logic.setState((st) => ({ events: st.events.filter((x) => x.id !== ev.id) }));
}

// ---- RSVP (Upcoming events widget) ----
// One small "rsvp" record per person per event; "going" is readable so the server can count, the rest is sealed.
export async function loadRsvps(logic, srv, events) {
  srv.rsvp = srv.rsvp || {};
  const mine = {};
  await Promise.all(events.map(async (e) => {
    const { records } = await srv.request('GET', `/containers/${e.container}/records?collection=rsvp`);
    const r = records.find((x) => x.author_account === srv.me.account && !x.deleted);
    if (r) { srv.rsvp[e.id] = r.id; mine[e.id] = !!r.meta.going; }
  }));
  logic.setState((st) => ({ rsvp: { ...st.rsvp, ...mine } }));
}

export async function toggleRsvp(logic, srv, ev) {
  const going = !logic.state.rsvp[ev.id];
  logic.setState((st) => ({ rsvp: { ...st.rsvp, [ev.id]: going } }));   // optimistic
  try {
    const rid = srv.rsvp[ev.id];
    const sealed = await srv.vault.seal(ev.container, { going }, rid);
    const body = { collection: 'rsvp', parent: ev.id, epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope, meta: { going } };
    if (rid) await srv.request('PUT', `/containers/${ev.container}/records/${rid}`, body);
    else { await srv.request('POST', `/containers/${ev.container}/records`, { id: sealed.id, ...body }); srv.rsvp[ev.id] = sealed.id; }
  } catch (e) {
    logic.setState((st) => ({ rsvp: { ...st.rsvp, [ev.id]: !going } }));
    throw e;
  }
}
