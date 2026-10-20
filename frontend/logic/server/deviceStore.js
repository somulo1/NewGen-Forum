// Keeps this browser's device private keys encrypted at rest. The keys are sealed with an AES-256-GCM key that the browser holds
// as a non-extractable CryptoKey in IndexedDB, so a copy of the page storage (or a script that only reads localStorage) gets
// ciphertext, not keys. Old plain entries are sealed the first time they are read. Where IndexedDB or WebCrypto are not
// available the old plain storage is used, as before.
const DB = 'hub-keystore', STORE = 'keys';
const enc = new TextEncoder(), dec = new TextDecoder();
const b64 = (u) => btoa(String.fromCharCode(...u));
const unb64 = (s) => Uint8Array.from(atob(s), (c) => c.charCodeAt(0));

function idb() {
  return new Promise((resolve, reject) => {
    const r = indexedDB.open(DB, 1);
    r.onupgradeneeded = () => r.result.createObjectStore(STORE);
    r.onsuccess = () => resolve(r.result);
    r.onerror = () => reject(r.error);
  });
}
const idbGet = (db, k) => new Promise((res, rej) => { const q = db.transaction(STORE).objectStore(STORE).get(k); q.onsuccess = () => res(q.result); q.onerror = () => rej(q.error); });
const idbPut = (db, k, v) => new Promise((res, rej) => { const q = db.transaction(STORE, 'readwrite').objectStore(STORE).put(v, k); q.onsuccess = () => res(); q.onerror = () => rej(q.error); });

async function wrappingKey(slot, create) {
  const db = await idb();
  let key = await idbGet(db, slot);
  if (!key && create) {
    key = await crypto.subtle.generateKey({ name: 'AES-GCM', length: 256 }, false, ['encrypt', 'decrypt']);
    await idbPut(db, slot, key);
  }
  return key || null;
}

export async function saveDevice(slot, obj) {
  const plain = JSON.stringify(obj);
  try {
    const key = await wrappingKey(slot, true);
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const ct = new Uint8Array(await crypto.subtle.encrypt({ name: 'AES-GCM', iv, additionalData: enc.encode('hub/device-store/v1|' + slot) }, key, enc.encode(plain)));
    localStorage.setItem(slot, JSON.stringify({ v: 2, iv: b64(iv), ct: b64(ct) }));
  } catch (e) {
    try { localStorage.setItem(slot, plain); } catch (x) { /* keys last for this tab only */ }
  }
}

export async function loadDevice(slot) {
  let raw;
  try { raw = JSON.parse(localStorage.getItem(slot) || 'null'); } catch (e) { return null; }
  if (!raw) return null;
  if (raw.v !== 2) { await saveDevice(slot, raw); return raw; }   // an old plain entry: seal it now
  try {
    const key = await wrappingKey(slot, false);
    if (!key) return null;
    const pt = await crypto.subtle.decrypt({ name: 'AES-GCM', iv: unb64(raw.iv), additionalData: enc.encode('hub/device-store/v1|' + slot) }, key, unb64(raw.ct));
    return JSON.parse(dec.decode(pt));
  } catch (e) { return null; }
}
