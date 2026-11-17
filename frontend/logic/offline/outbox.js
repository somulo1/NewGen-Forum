// Messages written without a connection (private chats, forums, the wall). Each is already sealed by the page exactly as it would
// be sent, with an id that carries the time it was written; it waits here, encrypted, and goes to the server in the order it was
// written once the connection is back. Sending one twice is harmless: the server refuses a repeated id as a duplicate.
import * as store from './store.js';

const SEND = /^\/conversations\/([0-9a-f-]{36})\/messages$/;
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;
const MAX_WAITING = 500;

const listeners = new Set();
const changed = () => listeners.forEach((fn) => { try { fn(); } catch (e) { /* a listener failing does not stop the others */ } });
export const onChange = (fn) => { listeners.add(fn); return () => listeners.delete(fn); };

export const queueable = (method, path, body) => method === 'POST' && SEND.test(path) && !!body && UUID.test(String(body.id || ''));

// Edits that may wait for the connection: changes to one's own work (notes, calendar entries, project plans, tickets, reactions,
// read marks, the home layout). Anything about security, people, roles, approvals or the organization's settings, and anything
// that needs the server to give it a number first, needs a connection.
const ID = '[0-9a-f-]{36}';
const EDITS = [
  ['POST', new RegExp(`^/containers/${ID}/records$`), (b) => !!b && UUID.test(String(b.id || ''))],
  ['PUT', new RegExp(`^/containers/${ID}/records/${ID}$`)], ['DELETE', new RegExp(`^/containers/${ID}/records/${ID}$`)],
  ['PUT', new RegExp(`^/projects/${ID}/items/${ID}$`)], ['DELETE', new RegExp(`^/projects/${ID}/items/${ID}$`)],
  ['DELETE', new RegExp(`^/conversations/${ID}/messages/${ID}$`)], ['POST', new RegExp(`^/conversations/${ID}/read$`)],
  ['POST', new RegExp(`^/tickets/${ID}/(comments|act|acknowledge|chase)$`)],
  ['PUT', new RegExp(`^/wall/${ID}/(reaction|vote)$`)],
  ['POST', /^\/comms\/seen$/], ['POST', new RegExp(`^/comms/${ID}/(read|acknowledge)$`)],
  ['POST', /^\/notifications\/read$/], ['POST', /^\/mentions$/], ['PUT', /^\/me\/prefs$/]
];
export const editable = (method, path, body) => EDITS.some(([m, re, ok]) => m === method && re.test(path) && (!ok || ok(body)));

const EDIT = '~edit|';

// Keeps an edit made without a connection. It goes to the server in the order it was made, carrying the time it was made, so
// the server keeps the latest change when someone else changed the same thing meanwhile.
export async function queueEdit(scope, method, path, body) {
  if ((await store.list('outbox', prefix(scope))).length >= MAX_WAITING) throw new Error('Too many changes are waiting to be sent. Connect to the internet to send them.');
  const at = Date.now();
  await store.put('outbox', prefix(scope) + EDIT + at + '|' + Math.random().toString(36).slice(2, 10), { method, path, body: body ?? null, at });
  changed();
  return { queued: true, ...(body && body.id ? { id: body.id } : {}) };
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
    created_at: writtenAt(body.id), deleted: false
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
const later = (e) => !e || e.status === 0 || e.status === 401 || e.status === 429 || e.status >= 500 || e instanceof TypeError;

let flushing = null;

// Sends what is waiting (messages and edits together), oldest first. Stops at the first one the server cannot take yet, so the
// order is kept. send(method, path, body, headers) is the normal request. onRefused(item, error): the server will never take
// it (the chat is gone, the right was removed, ...); it is removed. Resolves to how many went out.
export function flush(scope, send, onRefused) {
  if (flushing) return flushing;
  flushing = (async () => {
    let sent = 0;
    for (const [name] of await store.list('outbox', prefix(scope))) {
      const item = await store.get('outbox', name);
      if (!item) continue;
      try {
        if (item.method) await send(item.method, item.path, item.body ?? undefined, { 'X-Hub-Written-At': String(item.at) });
        else await send('POST', item.path, item.body);
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
