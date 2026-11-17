// The last answer to each read, kept so the app shows what it last saw when there is no connection. Answers that only make
// sense live (a call ringing now, who is online now) are never kept. Messages waiting in the outbox are part of every read of
// their conversation, so they show in their place until the server has them.
import * as store from './store.js';
import { pendingFor } from './outbox.js';

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

// The newest page of a conversation (no "before" marker) gains the messages still waiting to be sent, newest first like the server.
export async function withPending(scope, path, data) {
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
