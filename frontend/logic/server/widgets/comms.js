// Communications: announcements, memos and newsletters (the Announcements, Memos and Newsletter widgets and the
// Communications page share these). Content is sealed for the audience; this reads, marks read, and publishes.
import { newConversationKey, wrap, newId, b64, unb64 } from '../e2ee.js';
import * as files from '../files.js';

export const name = 'comms';

const dayOf = (iso) => (iso || '').slice(0, 10);

export async function load(logic, srv) {
  const { comms } = await srv.request('GET', '/comms');
  const { total } = await srv.request('GET', '/people?limit=1');
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
      status: c.status === 'Pending approval' ? 'Awaiting approval' : c.status, read: Math.min(100, Math.round((c.read / Math.max(1, total)) * 100)), iRead: !!c.i_read
    };
  }));
  rows.sort((a, b) => (a.at < b.at ? 1 : -1));
  logic.setState({ comms: rows });
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
export async function insights(srv, c) {
  return (await srv.request('GET', `/comms/${c.uid}/insights`)).places;
}

// Opens an attached file in a new tab.
export async function openFile(srv, f) {
  window.open(await files.openUrl(srv, f), '_blank');
}

// c = { type, title, summary, body, fields, files: [File], wall, forum (server conversation id or ''), sel: { 'departments:Finance': true } }
export async function publish(logic, srv, c) {
  const targets = [];
  const to = [];
  if (c.wall) { targets.push({ conv: () => audienceWall(srv, 'org', null, 'Community wall') }); to.push('Community wall'); }
  if (c.forum && c.forum.id) { targets.push({ conv: async () => c.forum.id }); to.push(c.forum.name + ' forum'); }
  Object.keys(c.sel || {}).filter((k) => c.sel[k]).forEach((k) => {
    const [level, nm] = [k.slice(0, k.indexOf(':')), k.slice(k.indexOf(':') + 1)];
    const ref = srv.structure[level] && srv.structure[level].get(nm);
    if (ref) { targets.push({ conv: () => audienceWall(srv, AUDIENCE[level], ref, nm + ' wall') }); to.push(nm); }
  });
  for (const f of c.files || []) { const bad = fileProblem(srv, f); if (bad) throw new Error(bad); }
  let pending = 0;
  const batch = newId();
  for (const t of targets) {
    const conv = await t.conv();
    // each place holds its own copy of the files, encrypted for that audience
    const refs = [];
    for (const f of c.files || []) { const u = await files.upload(srv, conv, f); refs.push({ fileId: u.fileId, name: u.name, meta: u.meta, type: u.type, isImg: u.isImg, key: u.key }); }
    const sealed = await srv.vault.seal(conv, { title: c.title, summary: c.summary || '', body: c.body, fields: c.fields || {}, files: refs, to, type: c.type, from: srv.me.name }, newId());
    const details = srv.emailDetails === false ? undefined : { title: c.title, summary: c.summary || '', body: c.body || '', fields: Object.entries(c.fields || {}).map(([k, v]) => [labelOf(c.type, k), String(v)]), files: (c.files || []).map((f) => f.name) };
    const res = await srv.request('POST', '/comms', { type: c.type, batch, details, conversation: conv, message: sealed.id, epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope });
    if (res.approval) pending++;
  }
  await load(logic, srv);
  return { sent: targets.length, pending };
}
