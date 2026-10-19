// Communications: announcements, memos and newsletters (the Announcements, Memos and Newsletter widgets and the
// Communications page share these). Content is sealed for the audience; this reads, marks read, and publishes.
import { newConversationKey, wrap, newId, b64, unb64 } from '../e2ee.js';

export const name = 'comms';

const dayOf = (iso) => (iso || '').slice(0, 10);

export async function load(logic, srv) {
  const { comms } = await srv.request('GET', '/comms');
  const { total } = await srv.request('GET', '/people?limit=1');
  const rows = await Promise.all(comms.map(async (c) => {
    let pl = null;
    try {
      const sealed = await srv.request('GET', `/comms/${c.id}/content`);
      pl = await srv.vault.openItem({ conversation: sealed.conversation, id: sealed.message, sender: sealed.sender_account, senderPerson: sealed.sender_person, senderDevice: sealed.device, epoch: sealed.epoch, envelope: sealed.envelope });
    } catch (e) { /* not visible to this person */ }
    const dept = srv.dirByUid && srv.dirByUid.get(c.by) ? srv.dirByUid.get(c.by).dept : '';
    return {
      id: c.ref, uid: c.id, type: c.type, title: pl ? pl.title : '(sealed)', body: pl ? pl.body || '' : '', by: c.by_name, byDept: dept,
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

// c = { type, title, body, wall, forum (server conversation id or ''), sel: { 'departments:Finance': true } }
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
  let pending = 0;
  for (const t of targets) {
    const conv = await t.conv();
    const sealed = await srv.vault.seal(conv, { title: c.title, body: c.body, to, type: c.type, from: srv.me.name }, newId());
    const res = await srv.request('POST', '/comms', { type: c.type, conversation: conv, message: sealed.id, epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope });
    if (res.approval) pending++;
  }
  await load(logic, srv);
  return { sent: targets.length, pending };
}
