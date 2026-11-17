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

// Sends what is waiting, oldest first. Stops at the first message the server cannot take yet, so the order is kept.
// onRefused(message, error): the server will never take it (the chat is gone, posting is closed, ...); it is removed.
export function flush(scope, post, onRefused) {
  if (flushing) return flushing;
  flushing = (async () => {
    for (const [name] of await store.list('outbox', prefix(scope))) {
      const item = await store.get('outbox', name);
      if (!item) continue;
      try {
        await post(item.path, item.body);
      } catch (e) {
        if (later(e)) return;
        if (!(e.status === 409 && e.code === 'duplicate')) onRefused(item.msg, e);
      }
      await store.del('outbox', name);
      changed();
    }
  })().catch(() => {}).finally(() => { flushing = null; });
  return flushing;
}
