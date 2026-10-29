// Keeps this account's device keys safe on the server, locked with a key made from the password, so signing in from another browser
// or phone gives the same keys back: posts, forums and private chats open there too, and nothing is lost when the browser changes.
// The server stores only ciphertext (AES-256-GCM under a PBKDF2-SHA256 key made from the password and a random salt).
//
//   sign in  -> onLogin: fetch the backup; if the password opens it, the keys replace this browser's; if there is none (or it was made
//               under an old password), a fresh one is prepared and boot uploads it once the device is known
//   boot     -> finish: upload the prepared backup
//   password -> rewrap: lock the same keys with the new password
import { loadDevice, saveDevice } from './deviceStore.js';

const enc = new TextEncoder(), dec = new TextDecoder();
const b64 = (u) => btoa(String.fromCharCode(...u));
const unb64 = (s) => Uint8Array.from(atob(s), (c) => c.charCodeAt(0));
const ITER = 600000;
const PENDING = 'hub-keybackup';
const slotOf = (account) => 'hub-device:' + account;
const aad = (account) => enc.encode('hub/key-backup/v1|' + account);

async function keyFrom(password, salt, iter) {
  const base = await crypto.subtle.importKey('raw', enc.encode(password), 'PBKDF2', false, ['deriveBits']);
  return new Uint8Array(await crypto.subtle.deriveBits({ name: 'PBKDF2', hash: 'SHA-256', salt, iterations: iter }, base, 256));
}
const aes = (raw, usage) => crypto.subtle.importKey('raw', raw, 'AES-GCM', false, [usage]);

async function seal(keyBytes, kdf, deviceExport, account) {
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const ct = new Uint8Array(await crypto.subtle.encrypt({ name: 'AES-GCM', iv, additionalData: aad(account) }, await aes(keyBytes, 'encrypt'), enc.encode(JSON.stringify(deviceExport))));
  const blob = new Uint8Array(12 + ct.length); blob.set(iv); blob.set(ct, 12);
  return { blob: b64(blob), kdf };
}

async function open(password, backup, account) {
  try {
    const k = backup.kdf || {};
    const raw = await keyFrom(password, unb64(k.salt), k.iter || ITER);
    const blob = unb64(backup.blob);
    const pt = await crypto.subtle.decrypt({ name: 'AES-GCM', iv: blob.subarray(0, 12), additionalData: aad(account) }, await aes(raw, 'decrypt'), blob.subarray(12));
    const o = JSON.parse(dec.decode(pt));
    return o && o.agree && o.sign ? o : null;
  } catch (e) { return null; }   // wrong password for this backup
}

// A prepared backup waits in this tab's session storage until the device is known (a few seconds), then it is removed.
async function prepare(password, account) {
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const key = await keyFrom(password, salt, ITER);
  try { sessionStorage.setItem(PENDING, JSON.stringify({ account, key: b64(key), kdf: { alg: 'PBKDF2-SHA256', iter: ITER, salt: b64(salt), v: 1 } })); } catch (e) { /* no storage: the backup waits for the next sign-in */ }
}
export const clearPending = () => { try { sessionStorage.removeItem(PENDING); } catch (e) { /* nothing */ } };

// request(method, path, body) is the signed-in call for the session that was just made.
export async function onLogin(request, account, password) {
  clearPending();
  let backup = null;
  try { backup = await request('GET', '/e2ee/backup'); } catch (e) { if (!e || e.status !== 404) return; /* cannot tell: leave things as they are */ }
  if (backup) {
    const dev = await open(password, backup, account);
    if (dev) { await saveDevice(slotOf(account), dev); return; }   // this browser now holds the account's keys
  }
  await prepare(password, account);   // none yet, or made under another password: make one from the keys this browser ends up with
}

// Called by boot once the device is known. Uploads the prepared backup (if any) of exactly these keys.
export async function finish(request, account, device) {
  let p = null;
  try { p = JSON.parse(sessionStorage.getItem(PENDING) || 'null'); } catch (e) { /* none */ }
  if (!p || p.account !== account) return;
  try {
    const body = await seal(unb64(p.key), p.kdf, device.export(), account);
    await request('PUT', '/e2ee/backup', body);
  } finally { clearPending(); }
}

// The password changed: lock this browser's keys with the new one, so the next browser still gets them.
export async function rewrap(request, account, newPassword) {
  const saved = await loadDevice(slotOf(account));
  if (!saved) { await prepare(newPassword, account); return; }
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const key = await keyFrom(newPassword, salt, ITER);
  const body = await seal(key, { alg: 'PBKDF2-SHA256', iter: ITER, salt: b64(salt), v: 1 }, saved, account);
  await request('PUT', '/e2ee/backup', body);
}
