// Offline use for the Hub phone app. In the app every read keeps its answer (encrypted on the phone) and falls back to it when
// there is no connection; messages written offline wait in the outbox and go out in order when the connection is back; the
// pages themselves are kept by a service worker so the app opens and moves between pages offline. In a browser none of this
// runs: requests go straight to the server as they always have.
import * as store from './store.js';
import * as cache from './cache.js';
import * as outbox from './outbox.js';
import { setStatus, flash } from './banner.js';

export const enabled = () => typeof window !== 'undefined' && !!(window.HubNative && window.HubNative.app) &&
  typeof indexedDB !== 'undefined' && !!(window.crypto && window.crypto.subtle);

const scopeOf = (s) => (s && s.account ? s.account + ':' + ((s.org && s.org.id) || '') : '');
// no connection, or the server is still waking up behind its proxy: nothing was done, so reads use the kept copy and
// messages wait (a message sent twice is refused as a duplicate, so retrying is safe)
const unreachable = (e) => !!e && (e instanceof TypeError || [0, 502, 503, 504].includes(e.status));
const OFFLINE_WRITE = 'Hub cannot be reached right now. This needs a connection; try again when you are back online.';

let scope = '', post = null, offline = false, wired = false, timer = 0;

async function paint() { setStatus(offline, await outbox.count(scope).catch(() => 0)); }

function reachable(ok) {
  if (offline === !ok) return;
  offline = !ok;
  paint();
  if (ok) soon(300);
}

function soon(ms) {
  clearTimeout(timer);
  timer = setTimeout(() => {
    if (!post || !scope || offline) return;
    outbox.flush(scope, post, (msg, e) => flash('A message you wrote offline could not be sent: ' + (e.message || 'the server refused it') + '.')).then(paint);
  }, ms);
}

function wire() {
  if (wired) return;
  wired = true;
  window.addEventListener('online', () => { offline = false; paint(); soon(500); });
  window.addEventListener('offline', () => reachable(false));
  document.addEventListener('visibilitychange', () => { if (!document.hidden) soon(800); });
  window.addEventListener('hub:signed-out', () => { forget(); });
  outbox.onChange(paint);
  setInterval(() => { if (!offline) soon(0); }, 20000);
}

function use(session, call) {
  const s = scopeOf(session);
  if (!s) return '';
  post = (path, body) => call('POST', path, body);
  if (s !== scope) { scope = s; wire(); offline = typeof navigator !== 'undefined' && navigator.onLine === false; paint(); soon(1500); }
  return s;
}

// call(method, path, body): the normal authorized request to the server
export async function through(method, path, body, session, call) {
  const s = use(session, call);
  if (!s) return call(method, path, body);
  if (method === 'GET') {
    try {
      const data = await call(method, path, body);
      reachable(true);
      cache.remember(s, path, data).catch(() => {});
      return cache.withPending(s, path, data);
    } catch (e) {
      if (!unreachable(e)) throw e;
      reachable(false);
      const kept = (await cache.offlineChanges(s, path)) ?? (await cache.recall(s, path).catch(() => undefined));
      if (kept === undefined) throw e;
      return cache.withPending(s, path, kept);
    }
  }
  try {
    const out = await call(method, path, body);
    reachable(true);
    return out;
  } catch (e) {
    if (!unreachable(e)) throw e;
    reachable(false);
    if (outbox.queueable(method, path, body)) return outbox.queue(s, session, path, body);
    e.message = OFFLINE_WRITE;
    throw e;
  }
}

// file contents (already encrypted end to end, sealed again at rest): what was opened once opens again offline
export async function file(path, session, call) {
  const s = scopeOf(session);
  try {
    const bytes = await call();
    if (s) { reachable(true); cache.rememberFile(s, path, bytes).catch(() => {}); }
    return bytes;
  } catch (e) {
    const kept = s && unreachable(e) ? await cache.recallFile(s, path).catch(() => undefined) : undefined;
    if (!kept) throw e;
    reachable(false);
    return kept;
  }
}

// Signing out erases every offline copy and anything still waiting to be sent.
export function forget() {
  scope = ''; post = null; offline = false;
  setStatus(false, 0);
  return store.wipe();
}

// The pages (scripts, styles, icons) are kept by a service worker, so the app opens and moves between pages offline.
// Registered only inside the app; a new deploy installs a new worker that replaces the old copies.
export function keepShell() {
  if (!enabled() || import.meta.dev || !('serviceWorker' in navigator)) return;
  const cfg = (window.__NUXT__ && window.__NUXT__.config && window.__NUXT__.config.app) || {};
  navigator.serviceWorker.register('/hub-sw.js?v=' + encodeURIComponent(cfg.buildId || 'app'), { scope: '/' }).catch(() => {});
}
