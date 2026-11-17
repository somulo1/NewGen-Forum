// The phone's own key. Made once, on first use, from the system's secure random source; the private half is kept in the
// phone's secure storage (iOS Keychain / Android Keystore-encrypted storage) as "this device only": it is not copied to
// backups, not moved to a new phone, and never handed to the web pages. Only signatures leave this file, and only over the two
// messages the server expects (a sign-in for an email, and a refresh of a session), so a page cannot get anything else signed.
import { ed25519 } from '@noble/curves/ed25519.js';
import * as Application from 'expo-application';
import * as Crypto from 'expo-crypto';
import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

const SLOT = 'hub.device.ed25519.v1';
const OPTS: SecureStore.SecureStoreOptions = { keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY };

const ALPHA = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/';
export function toB64(u: Uint8Array): string {
  let out = '';
  for (let i = 0; i < u.length; i += 3) {
    const a = u[i], b = u[i + 1], c = u[i + 2];
    out += ALPHA[a >> 2] + ALPHA[((a & 3) << 4) | ((b ?? 0) >> 4)];
    out += b === undefined ? '=' : ALPHA[((b & 15) << 2) | ((c ?? 0) >> 6)];
    out += c === undefined ? '=' : ALPHA[c & 63];
  }
  return out;
}
function fromB64(s: string): Uint8Array {
  const clean = s.replace(/[^A-Za-z0-9+/]/g, '');
  const out = new Uint8Array(Math.floor((clean.length * 3) / 4));
  let bits = 0, acc = 0, j = 0;
  for (const ch of clean) {
    acc = (acc << 6) | ALPHA.indexOf(ch);
    bits += 6;
    if (bits >= 8) { bits -= 8; out[j++] = (acc >> bits) & 255; }
  }
  return out.slice(0, j);
}

let secret: Uint8Array | null = null;

async function key(): Promise<Uint8Array> {
  if (secret) return secret;
  const kept = await SecureStore.getItemAsync(SLOT, OPTS);
  if (kept) {
    const k = fromB64(kept);
    if (k.length === 32) { secret = k; return k; }
  }
  const fresh = Crypto.getRandomBytes(32);
  await SecureStore.setItemAsync(SLOT, toB64(fresh), OPTS);
  secret = fresh;
  return fresh;
}

const enc = (s: string) => new TextEncoder().encode(s);
const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const EMAIL = /^[^\s@|]{1,200}@[^\s@|]{1,120}$/;

export type Proof = { key: string; ts: number; sig: string; platform: string; device: string };

// The phone itself, as the system names it for this app (Android ID, iOS identifier for vendor). Unlike the key, which Android
// erases together with the app, it survives reinstalling or updating Hub, so doing that on the same phone is not taken for a new
// phone. Only a hash of it leaves the phone, and the key's signature covers it.
let phoneId: Promise<string> | null = null;
function thisPhone(): Promise<string> {
  if (!phoneId) {
    phoneId = (async () => {
      let raw = '';
      try { raw = Platform.OS === 'android' ? Application.getAndroidId() : (await Application.getIosIdForVendorAsync()) || ''; } catch (e) { raw = ''; }
      return raw ? Crypto.digestStringAsync(Crypto.CryptoDigestAlgorithm.SHA256, 'hub-phone|' + raw) : '';
    })();
  }
  return phoneId;
}

// proof signs "hub-device-<kind>|<subject>|<unix seconds>"; kind is "login" (subject: the email, then this phone's hash) or
// "refresh" (the session id).
export async function proof(kind: string, subject: string): Promise<Proof> {
  const s = String(subject || '').trim();
  if (kind === 'login' ? !EMAIL.test(s) : kind === 'refresh' ? !UUID.test(s) : true) throw new Error('refused');
  const k = await key();
  const device = kind === 'login' ? await thisPhone() : '';
  const ts = Math.floor(Date.now() / 1000);
  const msg = 'hub-device-' + kind + '|' + s.toLowerCase() + (device ? '|' + device : '') + '|' + ts;
  return { key: toB64(ed25519.getPublicKey(k)), ts, sig: toB64(ed25519.sign(enc(msg), k)), platform: Platform.OS + ' ' + String(Platform.Version), device };
}

// readable at start-up so a broken secure store is found before someone tries to sign in
export async function ready(): Promise<boolean> {
  try { await key(); return true; } catch (e) { return false; }
}
