// The last answer to each read, kept so the app shows what it last saw when there is no connection. Answers that only make
// sense live (a call ringing now, who is online now) are never kept. Messages waiting in the outbox are part of every read of
// their conversation, so they show in their place until the server has them.
import * as store from './store.js';
import { pendingFor, pendingEdits } from './outbox.js';

const LIVE = [/^\/calls\/live/, /^\/presence/, /^\/realtime\//, /^\/auth\//, /\/changes\?/];
const MAX_ENTRIES = 1500;
const MAX_FILE_BYTES = 8 * 1024 * 1024;
const MAX_FILES = 300;

const MESSAGES = /^\/conversations\/([0-9a-f-]{36})\/messages(\?.*)?$/;
const CHANGES = /^\/conversations\/([0-9a-f-]{36})\/changes\?(.*)$/;

export const keepable = (path) => !LIVE.some((re) => re.test(path));

let writes = 0;

export async function remember(scope, path, data) {
  if (!keepable(path)) return;
  await store.put('cache', scope + path, data);
  if (++writes % 50 === 0) store.trim('cache', MAX_ENTRIES).catch(() => {});
}

export const recall = (scope, path) => store.get('cache', scope + path);

// encrypted file contents (pictures and attachments in chats), so what was opened once opens again offline
export async function rememberFile(scope, path, bytes) {
  if (!(bytes instanceof Uint8Array) || bytes.byteLength > MAX_FILE_BYTES) return;
  await store.put('files', scope + path, bytes);
  if (++writes % 20 === 0) store.trim('files', MAX_FILES).catch(() => {});
}

export const recallFile = (scope, path) => store.get('files', scope + path);

// The edits still waiting show in every read they belong to: a record or plan item changed, added or deleted offline looks that
// way in its list until the server has it. Only fields the listed thing already has are changed.
const lists = (data) => (data && typeof data === 'object' ? Object.values(data).filter((v) => Array.isArray(v) && v.some((x) => x && typeof x === 'object' && 'id' in x)) : []);
const parentOf = (p) => p.slice(0, p.lastIndexOf('/'));
const sameList = (editParent, base) => editParent === base || editParent.replace(/\/items$/, '/plan') === base;

function overlay(data, edits, base, query, session) {
  let out = data;
  const copy = () => { if (out === data) out = JSON.parse(JSON.stringify(data)); return out; };
  for (const e of edits) {
    const id = e.path.slice(e.path.lastIndexOf('/') + 1);
    if (e.method === 'PUT' && e.path === base && base === '/me/prefs') { out = { ...copy(), ...(e.body || {}) }; continue; }
    if (e.method === 'POST' && e.path === base && e.body && e.body.id) {
      if (query.get('collection') && query.get('collection') !== e.body.collection) continue;
      const now = new Date(e.at).toISOString();
      for (const list of lists(copy())) if (!list.some((x) => x.id === e.body.id)) list.push({ ...e.body, author: session.account, author_name: session.name, created_at: now, updated_at: now });
      continue;
    }
    if (!sameList(parentOf(e.path), base)) continue;
    for (const list of lists(copy())) {
      const i = list.findIndex((x) => x && x.id === id);
      if (e.method === 'DELETE') { if (i >= 0) list.splice(i, 1); continue; }
      if (i >= 0) { for (const k of Object.keys(e.body || {})) if (k in list[i]) list[i][k] = e.body[k]; }
      else if (e.body && list.length && Object.keys(e.body).every((k) => k in list[0])) list.push({ id, ...e.body });
    }
  }
  return out;
}

// The newest page of a conversation (no "before" marker) gains the messages still waiting to be sent, newest first like the server.
export async function withPending(scope, path, data, session) {
  const [base, q] = path.split('?');
  const edits = await pendingEdits(scope).catch(() => []);
  if (edits.length) data = overlay(data, edits, base, new URLSearchParams(q || ''), session || {});
  const m = MESSAGES.exec(path);
  if (!m || !data || !Array.isArray(data.messages) || /[?&]before=/.test(m[2] || '')) return data;
  const waiting = await pendingFor(scope, m[1]);
  if (!waiting.length) return data;
  const have = new Set(data.messages.map((x) => x.id));
  const messages = data.messages.concat(waiting.filter((x) => !have.has(x.id))).sort((a, b) => (a.id < b.id ? 1 : -1));
  return { ...data, messages };
}

// Offline, "what changed since my marker" has one honest answer: nothing this phone has not already seen, apart from its own
// messages that are waiting to go. The marker stays where it was, so the real catch-up runs once the connection is back.
export async function offlineChanges(scope, path) {
  const m = CHANGES.exec(path);
  if (!m) return undefined;
  const since = new URLSearchParams(m[2]).get('since') || '';
  return { messages: await pendingFor(scope, m[1]), next_since: since, next_after: '', more: false };
}
