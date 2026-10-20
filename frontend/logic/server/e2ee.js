// Browser/Node implementation of Hub's end-to-end encryption format (mirrors backend/internal/e2ee/e2ee.go byte for byte).
//   device keys   X25519 (agreement) + Ed25519 (signing); private halves stay on the device
//   key wrap      ephemeral X25519 -> HKDF-SHA256 -> AES-256-GCM, layout: 1 | eph(32) | nonce(12) | ct+tag
//   message       AES-256-GCM under the conversation key, AAD binds org/conversation/sender/id/epoch, then Ed25519-signed
import { x25519, ed25519 } from '@noble/curves/ed25519';
import { hkdf } from '@noble/hashes/hkdf';
import { sha256 } from '@noble/hashes/sha256';
import { hmac } from '@noble/hashes/hmac';

const VERSION = 1;
const NONCE = 12;
const SIG = 64;
const TAG = 16;
const WRAP_MIN = 1 + 32 + NONCE + 32 + TAG;
const enc = new TextEncoder();
const dec = new TextDecoder();
const subtle = () => globalThis.crypto.subtle;
const rand = (n) => globalThis.crypto.getRandomValues(new Uint8Array(n));
const cat = (...a) => { const o = new Uint8Array(a.reduce((n, x) => n + x.length, 0)); let i = 0; for (const x of a) { o.set(x, i); i += x.length; } return o; };

export const b64 = (u8) => { let s = ''; for (let i = 0; i < u8.length; i += 0x8000) s += String.fromCharCode.apply(null, u8.subarray(i, i + 0x8000)); return btoa(s); };
export const unb64 = (s) => Uint8Array.from(atob(s), (c) => c.charCodeAt(0));

const aesKey = (raw, usage) => subtle().importKey('raw', raw, 'AES-GCM', false, [usage]);
const gcmSeal = async (key, nonce, pt, aad) => new Uint8Array(await subtle().encrypt({ name: 'AES-GCM', iv: nonce, additionalData: aad, tagLength: 128 }, await aesKey(key, 'encrypt'), pt));
const gcmOpen = async (key, nonce, ct, aad) => new Uint8Array(await subtle().decrypt({ name: 'AES-GCM', iv: nonce, additionalData: aad, tagLength: 128 }, await aesKey(key, 'decrypt'), ct));

export class Device {
  // agree / sign are 32-byte private seeds
  constructor(agree, sign) { this.agree = agree; this.sign = sign; this.identity = x25519.getPublicKey(agree); this.signing = ed25519.getPublicKey(sign); }
  static generate() { return new Device(rand(32), rand(32)); }
  static fromSeed(seed) {
    return new Device(hkdf(sha256, seed, undefined, 'hub/e2ee/x25519', 32), hkdf(sha256, seed, undefined, 'hub/e2ee/ed25519', 32));
  }
  // Demo people only (backend/internal/seeder DeviceSeed): lets a development client act as a seeded person.
  static fromDemo(email, secret = 'hub-demo-seed') {
    return Device.fromSeed(hmac(sha256, enc.encode(secret), enc.encode('hub-demo-device|' + email)));
  }
  export() { return { agree: b64(this.agree), sign: b64(this.sign) }; }
  static import(o) { return new Device(unb64(o.agree), unb64(o.sign)); }
  // Proves the two keys belong together and to this account (POST /e2ee/devices).
  binding(accountId) { return ed25519.sign(cat(enc.encode('hub/device/v1|'), this.identity, uuidBytes(accountId)), this.sign); }

  async unwrap(wrapped, context = enc.encode('ctx')) {
    if (wrapped.length < WRAP_MIN || wrapped[0] !== VERSION) throw new Error('e2ee: malformed wrapped key');
    const eph = wrapped.subarray(1, 33);
    const k = hkdf(sha256, x25519.getSharedSecret(this.agree, eph), cat(eph, this.identity), 'hub/e2ee/wrap/v1', 32);
    return gcmOpen(k, wrapped.subarray(33, 33 + NONCE), wrapped.subarray(33 + NONCE), context);
  }
  async seal(convKey, plaintext, c) {
    const nonce = rand(NONCE);
    const body = cat(Uint8Array.of(VERSION), nonce, await gcmSeal(convKey, nonce, plaintext, aad(c)));
    return cat(body, ed25519.sign(signedBytes(body, c), this.sign));
  }
}

export async function wrap(convKey, recipientIdentity, context = enc.encode('ctx')) {
  const eph = rand(32);
  const ephPub = x25519.getPublicKey(eph);
  const k = hkdf(sha256, x25519.getSharedSecret(eph, recipientIdentity), cat(ephPub, recipientIdentity), 'hub/e2ee/wrap/v1', 32);
  const nonce = rand(NONCE);
  return cat(Uint8Array.of(VERSION), ephPub, nonce, await gcmSeal(k, nonce, convKey, context));
}

export const newConversationKey = () => rand(32);

export async function open(convKey, envelope, senderSigning, c) {
  if (envelope.length < 1 + NONCE + TAG + SIG || envelope[0] !== VERSION) throw new Error('e2ee: malformed envelope');
  const body = envelope.subarray(0, envelope.length - SIG);
  if (!ed25519.verify(envelope.subarray(envelope.length - SIG), signedBytes(body, c), senderSigning)) throw new Error('e2ee: signature does not match the claimed sender');
  return gcmOpen(convKey, body.subarray(1, 1 + NONCE), body.subarray(1 + NONCE), aad(c));
}

// c = { org, conversation, sender (account id), id (message/record id), epoch }
const aad = (c) => enc.encode(`hub/msg/v1|${c.org}|${c.conversation}|${c.sender}|${c.id}|${c.epoch}`);
const signedBytes = (body, c) => cat(aad(c), sha256(body));
const uuidBytes = (id) => Uint8Array.from(id.replace(/-/g, '').match(/../g).map((h) => parseInt(h, 16)));

// A short code two people can read to each other to be sure they hold the same keys: 12 groups of 5 digits made from every active
// device key of the person (sorted, so both sides get the same code).
export function safetyCode(identityKeys) {
  const h = sha256(cat(enc.encode('hub/safety/v1'), ...[...identityKeys].map(b64).sort().map((s) => enc.encode(s + '|'))));
  const groups = [];
  for (let i = 0; i < 12; i++) groups.push(String(((h[i * 2] << 8) | h[i * 2 + 1]) % 100000).padStart(5, '0'));
  return groups.join(' ');
}

export const toText = (u8) => dec.decode(u8);
export const fromText = (s) => enc.encode(s);
export const newId = () => {
  // UUIDv7-like, time ordered (the server sorts and pages by id)
  const b = rand(16); const t = BigInt(Date.now());
  for (let i = 0; i < 6; i++) b[i] = Number((t >> BigInt(8 * (5 - i))) & 0xffn);
  b[6] = (b[6] & 0x0f) | 0x70; b[8] = (b[8] & 0x3f) | 0x80;
  const h = [...b].map((x) => x.toString(16).padStart(2, '0')).join('');
  return `${h.slice(0, 8)}-${h.slice(8, 12)}-${h.slice(12, 16)}-${h.slice(16, 20)}-${h.slice(20)}`;
};
