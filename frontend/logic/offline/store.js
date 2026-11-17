// Encrypted storage for the phone app's offline copies. Every record is sealed with AES-256-GCM under a key the WebView keeps
// as a non-extractable CryptoKey, so the stored data is unreadable outside this app; signing out deletes the whole database.
const DB = 'hub-offline', VERSION = 1;
const STORES = ['keys', 'cache', 'files', 'outbox'];
const enc = new TextEncoder(), dec = new TextDecoder();

let dbp = null, keyp = null;

const req = (r) => new Promise((resolve, reject) => { r.onsuccess = () => resolve(r.result); r.onerror = () => reject(r.error); });

function db() {
  if (!dbp) {
    dbp = new Promise((resolve, reject) => {
      const r = indexedDB.open(DB, VERSION);
      r.onupgradeneeded = () => STORES.forEach((s) => { if (!r.result.objectStoreNames.contains(s)) r.result.createObjectStore(s).createIndex('at', 'at'); });
      r.onsuccess = () => resolve(r.result);
      r.onerror = () => reject(r.error);
    });
    dbp.catch(() => { dbp = null; });
  }
  return dbp;
}

const tx = async (store, mode, fn) => { const d = await db(); return req(fn(d.transaction(store, mode).objectStore(store))); };

function key() {
  if (!keyp) {
    keyp = (async () => {
      let k = await tx('keys', 'readonly', (s) => s.get('main'));
      if (!k) {
        k = await crypto.subtle.generateKey({ name: 'AES-GCM', length: 256 }, false, ['encrypt', 'decrypt']);
        await tx('keys', 'readwrite', (s) => s.put(k, 'main'));
      }
      return k;
    })();
    keyp.catch(() => { keyp = null; });
  }
  return keyp;
}

// the record's name is bound into the seal, so a sealed value moved to another name does not open
async function seal(name, value) {
  const bin = value instanceof Uint8Array;
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const data = bin ? value : enc.encode(JSON.stringify(value));
  const ct = await crypto.subtle.encrypt({ name: 'AES-GCM', iv, additionalData: enc.encode(name) }, await key(), data);
  return { iv, ct: new Uint8Array(ct), bin };
}

async function unseal(name, rec) {
  const pt = new Uint8Array(await crypto.subtle.decrypt({ name: 'AES-GCM', iv: rec.iv, additionalData: enc.encode(name) }, await key(), rec.ct));
  return rec.bin ? pt : JSON.parse(dec.decode(pt));
}

export async function put(store, name, value, extra = {}) {
  const rec = { ...(await seal(store + '|' + name, value)), at: Date.now(), ...extra };
  await tx(store, 'readwrite', (s) => s.put(rec, name));
}

export async function get(store, name) {
  const rec = await tx(store, 'readonly', (s) => s.get(name));
  if (!rec) return undefined;
  try { return await unseal(store + '|' + name, rec); } catch (e) { await del(store, name); return undefined; }
}

export const del = (store, name) => tx(store, 'readwrite', (s) => s.delete(name));

// Keeps the newest max records of a store: the oldest are deleted, reading only their names.
export async function trim(store, max) {
  const d = await db();
  return new Promise((resolve, reject) => {
    const t = d.transaction(store, 'readwrite');
    const s = t.objectStore(store);
    const count = s.count();
    count.onsuccess = () => {
      let extra = count.result - max;
      if (extra <= 0) return;
      const r = s.index('at').openKeyCursor();
      r.onsuccess = () => { const c = r.result; if (!c || extra-- <= 0) return; s.delete(c.primaryKey); c.continue(); };
    };
    t.oncomplete = () => resolve();
    t.onerror = () => reject(t.error);
  });
}

// [name, record metadata] of every record whose name starts with prefix, oldest first (for the small outbox)
export async function list(store, prefix = '') {
  const d = await db();
  return new Promise((resolve, reject) => {
    const out = [];
    const range = prefix ? IDBKeyRange.bound(prefix, prefix + '￿') : undefined;
    const r = d.transaction(store, 'readonly').objectStore(store).openCursor(range);
    r.onsuccess = () => {
      const c = r.result;
      if (!c) { resolve(out.sort((a, b) => a[1].at - b[1].at)); return; }
      const { iv, ct, ...meta } = c.value;
      out.push([c.key, { ...meta, size: ct.byteLength }]);
      c.continue();
    };
    r.onerror = () => reject(r.error);
  });
}

// Signing out: everything the app kept for offline use is erased, including the key that sealed it.
export function wipe() {
  const closing = dbp ? dbp.then((d) => d.close()).catch(() => {}) : Promise.resolve();
  dbp = null; keyp = null;
  return closing.then(() => req(indexedDB.deleteDatabase(DB))).catch(() => {});
}
