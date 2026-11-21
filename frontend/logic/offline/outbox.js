// Everything done without a connection waits here, encrypted, and goes to the server in the order it was done once the
// connection is back: messages (sealed by the page exactly as they would be sent, with an id that carries the time they were
// written), every other change, and the encrypted contents of attachments. Anything that waited more than 48 hours without a
// connection is not sent: it is dropped and the person is told (the server refuses it as well). Sending one twice is harmless.
import * as store from './store.js';

const SEND = /^\/conversations\/([0-9a-f-]{36})\/messages$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
const MAX_WAITING = 500;
export const MAX_WAIT_MS = 48 * 3600 * 1000;

const listeners = new Set();
const changed = () => listeners.forEach((fn) => { try { fn(); } catch (e) { /* a listener failing does not stop the others */ } });
export const onChange = (fn) => { listeners.add(fn); return () => listeners.delete(fn); };

export const queueable = (method, path, body) => method === 'POST' && SEND.test(path) && !!body && UUID.test(String(body.id || ''));

// Every change may wait for the connection, except the few whose answer the page needs before it can go on (the server gives the
// new thing its number or keys) and the sign-in and live-call machinery, which only make sense live.
const ID = '[0-9a-f-]{36}';
const LIVE_ONLY = [
  /^\/auth\//, /^\/realtime\//, /^\/push\//, /^\/calls(\/|$)/, /^\/e2ee\//, /^\/conversations$/, /^\/comms$/, /^\/awards$/,
  /^\/projects$/, new RegExp(`^/projects/${ID}/gates/`), /^\/roles$/, /^\/groups$/, new RegExp(`^/(roles|groups)/${ID}/rights`),
  /^\/people$/, new RegExp(`^/people/${ID}(/rights)`), /^\/settings\/subscription$/, new RegExp(`^/files/${ID}/shares$`),
  new RegExp(`^/tickets/${ID}/route$`)
];
// PATCH /people/<id> answers with what was applied (a role change may wait for approval), so it needs the server too
export const editable = (method, path) => method !== 'GET' && !LIVE_ONLY.some((re) => re.test(path)) && !(method === 'PATCH' && /^\/people\//.test(path));

const EDIT = '~edit|';

// Keeps an edit made without a connection. It goes to the server in the order it was made, carrying the time it was made, so
// the server keeps the latest change when someone else changed the same thing meanwhile.
export async function queueEdit(scope, method, path, body) {
  if ((await store.list('outbox', prefix(scope))).length >= MAX_WAITING) throw new Error('Too many changes are waiting to be sent. Connect to the internet to send them.');
  const at = Date.now();
  await store.put('outbox', prefix(scope) + EDIT + at + '|' + Math.random().toString(36).slice(2, 10), { method, path, body: body ?? null, at });
  changed();
  return { queued: true, ...(body && typeof body === 'object' ? body : {}) };
}

// The encrypted contents of an attachment, kept until its file record (queued just before it) is on the server.
export async function queueBytes(scope, path, bytes) {
  const at = Date.now();
  await store.put('outbox', prefix(scope) + '~file|' + at + '|' + Math.random().toString(36).slice(2, 10), { method: 'PUT', path, bytes, at });
  changed();
}

// the edits still waiting, oldest first, so reads can show them in place
export async function pendingEdits(scope) {
  const out = [];
  for (const [name] of await store.list('outbox', prefix(scope) + EDIT)) {
    const item = await store.get('outbox', name);
    if (item) out.push(item);
  }
  return out;
}

const writtenAt = (id) => new Date(parseInt(id.replace(/-/g, '').slice(0, 12), 16)).toISOString();

// what the server would answer for this message, so the page shows it as sent
function asStored(session, conv, body) {
  return {
    id: body.id, conversation: conv, type: body.type || 'text', reply_to: body.reply_to || null, epoch: body.epoch || 0,
    sender_account: session.account, sender_person: (session.org && session.org.person) || '', sender_device: body.device || '',
    sender_name: session.name || '', envelope: body.envelope || '', ...(body.body ? { body: body.body } : {}),
    created_at: writtenAt(body.id), deleted: false, queued: true
  };
}

const prefix = (scope, conv = '') => scope + '|' + (conv ? conv + '|' : '');

export async function queue(scope, session, path, body) {
  const conv = SEND.exec(path)[1];
  if ((await store.list('outbox', prefix(scope))).length >= MAX_WAITING) throw new Error('Too many messages are waiting to be sent. Connect to the internet to send them.');
  const msg = asStored(session, conv, body);
  await store.put('outbox', prefix(scope, conv) + body.id, { path, body, msg });
  changed();
  return msg;
}

export async function pendingFor(scope, conv) {
  const out = [];
  for (const [name] of await store.list('outbox', prefix(scope, conv))) {
    const item = await store.get('outbox', name);
    if (item) out.push(item.msg);
  }
  return out;
}

export const count = async (scope) => (scope ? (await store.list('outbox', prefix(scope))).length : 0);

// the server did not answer, or asked to wait: try again later
// Without a connection nothing is sent, but the 48 hours still run: what is older is dropped now, and the person is told.
export async function expire(scope, onExpired) {
  let gone = 0;
  for (const [name, meta] of await store.list('outbox', prefix(scope))) {
    if (Date.now() - meta.at <= MAX_WAIT_MS) continue;
    const item = await store.get('outbox', name);
    await store.del('outbox', name);
    if (item) { gone++; onExpired(item); }
  }
  if (gone) changed();
  return gone;
}

const later = (e) => !e || e.status === 0 || e.status === 401 || e.status === 429 || e.status >= 500 || e instanceof TypeError;

let flushing = null;

// Sends what is waiting (messages and edits together), oldest first. Stops at the first one the server cannot take yet, so the
// order is kept. send(method, path, body, headers) is the normal request. onRefused(item, error): the server will never take
// it (the chat is gone, the right was removed, ...); it is removed. Resolves to how many went out.
export function flush(scope, send, sendBytes, onRefused) {
  if (flushing) return flushing;
  flushing = (async () => {
    let sent = 0;
    for (const [name, meta] of await store.list('outbox', prefix(scope))) {
      const item = await store.get('outbox', name);
      if (!item) continue;
      const at = item.at || meta.at;
      if (Date.now() - at > MAX_WAIT_MS) {
        onRefused(item, new Error('it waited more than 48 hours without a connection'));
        await store.del('outbox', name);
        changed();
        continue;
      }
      try {
        if (item.bytes) await sendBytes(item.path, item.bytes, at);
        else await send(item.method || 'POST', item.path, item.body ?? undefined, { 'X-Hub-Written-At': String(at) });
        sent++;
      } catch (e) {
        if (later(e)) return sent;
        // already stored, or someone changed it later (the newer change is kept): nothing to tell
        if (!(e.status === 409 && (e.code === 'duplicate' || e.code === 'stale_edit'))) onRefused(item, e);
      }
      await store.del('outbox', name);
      changed();
    }
    return sent;
  })().catch(() => 0).finally(() => { flushing = null; });
  return flushing;
}
