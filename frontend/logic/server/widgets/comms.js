// Communications: announcements, memos and newsletters (the Announcements, Memos and Newsletter widgets and the
// Communications page share these). Content is sealed for the audience; this reads, marks read, and publishes.
import { newConversationKey, wrap, newId, b64, unb64 } from '../e2ee.js';
import * as files from '../files.js';

export const name = 'comms';

const dayOf = (iso) => (iso || '').slice(0, 10);

export async function load(logic, srv) {
  const { comms } = await srv.request('GET', '/comms');
  // all the sealed contents in one request (it was one request per communication)
  const sealedBy = new Map();
  for (let i = 0; i < comms.length; i += 200) {
    try { (await srv.request('POST', '/comms/contents', { ids: comms.slice(i, i + 200).map((c) => c.id) })).contents.forEach((x) => sealedBy.set(x.id, x)); } catch (e) { /* none visible */ }
  }
  const rows = await Promise.all(comms.map(async (c) => {
    let pl = null;
    try {
      const sealed = sealedBy.get(c.id);
      if (sealed) pl = await srv.vault.openItem({ conversation: sealed.conversation, id: sealed.message, sender: sealed.sender_account, senderPerson: sealed.sender_person, senderDevice: sealed.device, epoch: sealed.epoch, envelope: sealed.envelope });
    } catch (e) { /* not visible to this person */ }
    const dept = srv.dirByUid && srv.dirByUid.get(c.by) ? srv.dirByUid.get(c.by).dept : '';
    return {
      id: c.ref, uid: c.id, type: c.type, title: pl ? pl.title : '(sealed)', summary: pl ? pl.summary || '' : '', body: pl ? pl.body || '' : '', fields: pl && pl.fields ? pl.fields : {}, files: pl && pl.files ? pl.files : [], by: c.by_name, byDept: dept,
      date: dayOf(c.published_at || c.created_at), at: c.published_at || c.created_at, to: pl && pl.to ? pl.to : [],
      batch: c.batch || c.id, sealed: !pl,
      status: c.status === 'Pending approval' ? 'Awaiting approval' : c.status, // read % of the people it actually went to (members of that place, author excluded), never of the whole company
      readers: c.read || 0, audience: c.audience || 0, read: c.audience ? Math.min(100, Math.round(((c.read || 0) / c.audience) * 100)) : 0, iRead: !!c.i_read, iAcked: !!c.i_acked, mine: !!(srv.me && c.by === srv.me.person)
    };
  }));
  // one communication sent to several places is stored once per place (same batch): it is listed once
  const byBatch = new Map();
  rows.forEach((r) => { if (!byBatch.has(r.batch)) byBatch.set(r.batch, []); byBatch.get(r.batch).push(r); });
  const list = [...byBatch.values()].map((copies) => {
    const main = copies.find((x) => !x.sealed) || copies[0];
    const statuses = new Set(copies.map((x) => x.status));
    return { ...main, copies: copies.length,
      status: statuses.size === 1 ? main.status : statuses.has('Awaiting approval') ? 'Awaiting approval' : main.status,
      // copies sent to several places together: everyone reached across them, and everyone who read
      // each copy already carries the whole group's numbers, each person counted once: take them, never add copies up
      readers: Math.max(...copies.map((x) => x.readers)), audience: Math.max(...copies.map((x) => x.audience)), read: Math.max(...copies.map((x) => x.read)), iRead: copies.some((x) => x.iRead), iAcked: copies.some((x) => x.iAcked),
      at: copies.map((x) => x.at).sort().reverse()[0], date: dayOf(copies.map((x) => x.at).sort().reverse()[0]) };
  });
  list.sort((a, b) => (a.at < b.at ? 1 : -1));
  logic.setState({ comms: list });
}

export async function markRead(logic, srv, c) {
  await srv.request('POST', `/comms/${c.uid}/read`);
  logic.setState((st) => ({ comms: st.comms.map((x) => (x.uid === c.uid ? { ...x, iRead: true } : x)) }));
}

// the label of a form field, for emails
const LABELS = { priority: 'Priority', effective: 'Takes effect on', ref: 'Reference number', actionBy: 'Action needed by', actionFrom: 'Who must act', issue: 'Issue', team: 'Team', location: 'Location', employment: 'Employment type', closing: 'Closing date', apply: 'How to apply' };
const labelOf = (type, k) => LABELS[k] || k;

const AUDIENCE = { sectors: 'sector', branches: 'branch', units: 'unit', departments: 'department' };

// Finds the wall for an audience (or makes it, wrapping its key for everyone it covers) and returns its conversation id.
async function audienceWall(srv, audience, ref, label) {
  const walls = (await srv.request('GET', '/conversations?kind=wall&limit=200')).conversations;
  const hit = walls.find((w) => w.audience === audience && (audience === 'org' || w.audience_ref === ref));
  if (hit) return hit.id;
  const { people } = await srv.request('GET', `/conversations/audience-members?audience=${audience}${ref ? '&ref=' + ref : ''}`);
  const all = [...new Set([srv.me.person, ...people])];
  const { devices } = await srv.request('POST', '/e2ee/people/devices', { people: all });
  const key = newConversationKey();
  const keys = await Promise.all(devices.map(async (d) => ({ epoch: 1, device: d.id, wrapped_key: b64(await wrap(key, unb64(d.identity_key))) })));
  const made = await srv.request('POST', '/conversations', { kind: 'wall', name: label, audience, audience_ref: ref || undefined, color: '#F49AC1', keys });
  return made.id;
}

// Checks a file against the organization's limits (set on the Settings page).
export function fileProblem(srv, f) {
  const lim = srv.commFiles || { maxMB: 10, types: [] };
  const ext = (f.name.split('.').pop() || '').toLowerCase();
  if (lim.types.length && !lim.types.includes(ext)) return '"' + f.name + '" is not an allowed type. Allowed: ' + lim.types.join(', ') + '.';
  if (f.size > lim.maxMB * 1048576) return '"' + f.name + '" is over the ' + lim.maxMB + ' MB limit.';
  return '';
}

// Who read it, who reacted and who replied, for every place it went. Only people allowed to see insights get an answer.
// "Received and noted": a click, apart from just opening it. Its insights show who did.
export async function acknowledge(logic, srv, c) {
  await srv.request('POST', `/comms/${c.uid}/acknowledge`);
  logic.setState((st) => ({ comms: st.comms.map((x) => (x.uid === c.uid ? { ...x, iAcked: true, iRead: true } : x)) }));
}

// Everything the engagement overview needs, in one request.
export async function engagement(srv) {
  return (await srv.request('GET', '/comms/engagement')).comms;
}

// Reminds only the people a communication has not reached yet (at most once every 6 hours).
export async function remind(srv, c) {
  return (await srv.request('POST', `/comms/${c.uid}/remind`)).reminded;
}

// per place, plus `totals`: the whole communication with each person counted once, whatever places they met it in
export async function insights(srv, c) {
  const r = await srv.request('GET', `/comms/${c.uid}/insights`);
  return { places: r.places || [], totals: r.totals || null };
}

// Opens an attached file in a new tab.
export async function openFile(srv, f) {
  window.open(await files.openUrl(srv, f), '_blank', 'noopener');
}

// c = { type, title, summary, body, fields, files: [File], wall, forum (server conversation id or ''), sel: { 'departments:Finance': true } }
// Posts a published item again to the places it went to before. Each place gets a fresh copy sealed for it (with its own copy of
// the files), and only places the person belongs to can be posted into. It follows the same approval rules as a new item.
export async function repost(logic, srv, c) {
  const { places } = await srv.request('GET', `/comms/${c.uid}/places`);
  const mine = places.filter((p) => p.member);
  if (!mine.length) throw new Error('You are not a member of any place it went to, so it cannot be reposted by you.');
  // the files are opened once here and uploaded again for each place
  const blobs = await Promise.all((c.files || []).map(async (f) => { const b = await (await fetch(await files.openUrl(srv, f))).blob(); return new File([b], f.name, { type: f.type || b.type }); }));
  const batch = newId();
  let sent = 0, pending = 0;
  const failed = [];
  for (const p of mine) {
    try {
      const refs = [];
      for (const f of blobs) { const u = await files.upload(srv, p.conversation, f); refs.push({ fileId: u.fileId, name: u.name, meta: u.meta, type: u.type, isImg: u.isImg, key: u.key }); }
      const sealed = await srv.vault.seal(p.conversation, { title: c.title, summary: c.summary || '', body: c.body, fields: c.fields || {}, files: refs, to: c.to || [], type: c.type, from: srv.me.name, reposted: true }, newId());
      const details = srv.emailDetails === false ? undefined : { title: c.title, summary: c.summary || '', body: c.body || '', fields: Object.entries(c.fields || {}).map(([k, v]) => [labelOf(c.type, k), String(v)]), files: (c.files || []).map((f) => f.name) };
      const res = await srv.request('POST', '/comms', { type: c.type, batch, details, conversation: p.conversation, message: sealed.id, epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope, repost_of: c.uid });
      sent++;
      if (res.approval) pending++;
    } catch (e) { failed.push((p.name || (p.kind === 'forum' ? 'a forum' : 'a wall')) + ' (' + ((e && e.message) || 'it did not go through') + ')'); }
  }
  await load(logic, srv);
  if (failed.length) throw new Error((sent ? 'Reposted to ' + sent + ', but not to ' : 'Not reposted to ') + failed.join('; ') + '.');
  return { sent, pending, skipped: places.length - mine.length };
}

export async function publish(logic, srv, c) {
  const targets = [];
  const to = [];
  if (c.wall) { targets.push({ name: 'the Community wall', conv: () => audienceWall(srv, 'org', null, 'Community wall') }); to.push('Community wall'); }
  for (const fo of c.forums || (c.forum ? [c.forum] : [])) {
    if (!fo.id) throw new Error('The forum "' + fo.name + '" is not connected to the server yet. Reload the page and try again.');
    targets.push({ name: 'the ' + fo.name + ' forum', conv: async () => fo.id }); to.push(fo.name + ' forum');
  }
  Object.keys(c.sel || {}).filter((k) => c.sel[k]).forEach((k) => {
    const [level, nm] = [k.slice(0, k.indexOf(':')), k.slice(k.indexOf(':') + 1)];
    const ref = srv.structure[level] && srv.structure[level].get(nm);
    if (ref) { targets.push({ name: nm, conv: () => audienceWall(srv, AUDIENCE[level], ref, nm + ' wall') }); to.push(nm); }
  });
  for (const f of c.files || []) { const bad = fileProblem(srv, f); if (bad) throw new Error(bad); }
  let pending = 0;
  const batch = newId();
  const failed = [];
  let sent = 0;
  // every place is tried on its own: one that refuses (a forum whose rules stop you posting, say) never stops the others
  for (const t of targets) {
    try {
      const conv = await t.conv();
      // each place holds its own copy of the files, encrypted for that audience
      const refs = [];
      for (const f of c.files || []) { const u = await files.upload(srv, conv, f); refs.push({ fileId: u.fileId, name: u.name, meta: u.meta, type: u.type, isImg: u.isImg, key: u.key }); }
      const sealed = await srv.vault.seal(conv, { title: c.title, summary: c.summary || '', body: c.body, fields: c.fields || {}, files: refs, to, type: c.type, from: srv.me.name }, newId());
      const details = srv.emailDetails === false ? undefined : { title: c.title, summary: c.summary || '', body: c.body || '', fields: Object.entries(c.fields || {}).map(([k, v]) => [labelOf(c.type, k), String(v)]), files: (c.files || []).map((f) => f.name) };
      const res = await srv.request('POST', '/comms', { type: c.type, batch, details, conversation: conv, message: sealed.id, epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope });
      sent++;
      if (res.approval) pending++;
    } catch (e) { failed.push({ place: t.name, why: (e && e.message) || 'it did not go through' }); }
  }
  await load(logic, srv);
  if (failed.length) throw new Error((sent ? 'Sent to ' + sent + ' place' + (sent === 1 ? '' : 's') + ', but not to ' : 'Not sent to ') + failed.map((f) => f.place + ' (' + f.why + ')').join('; ') + '.');
  return { sent, pending };
}
