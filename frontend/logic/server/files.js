// Encrypted attachments in any container (wall, forum, chat, ticket, project, event, folder). The bytes are AES-256-GCM
// under a fresh file key (layout nonce | ciphertext, as the seeder writes); the file key and name are sealed for the
// container's members in the file's description. The server stores only ciphertext.
import { b64, unb64 } from './e2ee.js';
import { newId } from './e2ee.js';

const subtle = () => globalThis.crypto.subtle;

const fmtSize = (n) => (n > 1048576 ? (n / 1048576).toFixed(1) + ' MB' : Math.max(1, Math.round(n / 1024)) + ' KB');
const kindOf = (name, type) => {
  if ((type || '').startsWith('image/')) return 'IMG';
  if ((type || '').startsWith('video/')) return 'VID';
  const ext = (name.split('.').pop() || '').toUpperCase();
  return { PDF: 'PDF', DOC: 'DOC', DOCX: 'DOC', XLS: 'XLS', XLSX: 'XLS', CSV: 'XLS', PPT: 'PPT', PPTX: 'PPT', ZIP: 'ZIP', PNG: 'IMG', JPG: 'IMG', JPEG: 'IMG' }[ext] || 'FILE';
};

// Uploads a browser File into a container. Returns what a message or record should carry about it.
export async function upload(srv, container, file) {
  const limit = (srv.maxVideoMB || 50) * 1048576;
  if (kindOf(file.name, file.type) === 'VID' && file.size > limit) throw new Error('Videos can be up to ' + (srv.maxVideoMB || 50) + ' MB. "' + file.name + '" is ' + fmtSize(file.size) + '. Your administrator sets this limit in Settings.');
  const raw = new Uint8Array(await file.arrayBuffer());
  const fileKey = globalThis.crypto.getRandomValues(new Uint8Array(32));
  const nonce = globalThis.crypto.getRandomValues(new Uint8Array(12));
  const k = await subtle().importKey('raw', fileKey, 'AES-GCM', false, ['encrypt']);
  const id = newId();   // the file's id is part of what is authenticated, so its bytes cannot be swapped for another file's
  const ct = new Uint8Array(await subtle().encrypt({ name: 'AES-GCM', iv: nonce, additionalData: new TextEncoder().encode('hub/file/v1|' + id) }, k, raw));
  const body = new Uint8Array(nonce.length + ct.length);
  body.set(nonce); body.set(ct, nonce.length);
  const info = { name: file.name, kind: kindOf(file.name, file.type), type: file.type || '', size_label: fmtSize(file.size), by: srv.me.name, when: new Date().toISOString(), file_key: b64(fileKey) };
  const sealed = await srv.vault.seal(container, info, id);
  await srv.request('POST', `/folders/${container}/files`, { id: sealed.id, epoch: sealed.epoch, device: sealed.device, envelope: sealed.envelope });
  await srv.bytes('PUT', `/files/${sealed.id}/content`, body);
  urls.set(sealed.id, URL.createObjectURL(file));   // what was just sent is shown from here, not downloaded back
  return { fileId: sealed.id, name: file.name, meta: info.size_label, type: info.type, isImg: info.kind === 'IMG', isVideo: info.kind === 'VID', key: info.file_key };
}

// The type of a file comes from whoever uploaded it, and a blob: address belongs to this site. A page type (HTML, SVG, XML) opened
// from one would run the sender's script here, so only types that cannot run script keep their type; everything else is a download.
const SAFE_TYPES = /^(image\/(png|jpeg|gif|webp|avif|bmp)|video\/(mp4|webm|ogg|quicktime)|audio\/(mpeg|ogg|wav|webm|mp4)|application\/pdf|text\/(plain|csv))$/;
const safeType = (t) => (SAFE_TYPES.test(String(t || '').toLowerCase().split(';')[0].trim()) ? String(t).toLowerCase().split(';')[0].trim() : 'application/octet-stream');

const urls = new Map();
const opening = new Map();
// Downloads and decrypts an attachment; returns an object URL (cached for the session; one download per file at a time).
// Waits for the first few of a page's pictures and files to be ready, but never more than a few seconds: a page counts as loaded when
// what is on its first screen is in, and the rest fills in behind.
export const settle = (promises, n = 12, ms = 6000) => Promise.race([Promise.all(promises.slice(0, n)), new Promise((r) => setTimeout(r, ms))]);

export async function openUrl(srv, ref) {
  if (urls.has(ref.fileId)) return urls.get(ref.fileId);
  if (!opening.has(ref.fileId)) opening.set(ref.fileId, fetchUrl(srv, ref).catch(async (e) => { await new Promise((r) => setTimeout(r, 1500)); try { return await fetchUrl(srv, ref); } catch (e2) { console.warn('Could not open attachment ' + ref.fileId + ':', e2); throw e2; } }).finally(() => opening.delete(ref.fileId)));
  return opening.get(ref.fileId);
}
// A file's encrypted bytes never change (its id is part of what is authenticated), so this browser keeps the copy it downloaded:
// the next time the wall opens, the picture comes from here instead of crossing the network again. What is kept is the same
// ciphertext the server holds; it cannot be read without the key inside the post.
const CACHE = 'hub-files-v1';
async function cachedBytes(srv, id) {
  const url = '/hub-file/' + id;
  try {
    if (globalThis.caches) {
      const c = await caches.open(CACHE);
      const hit = await c.match(url);
      if (hit) return new Uint8Array(await hit.arrayBuffer());
      const data = await srv.bytes('GET', `/files/${id}/content`);
      c.put(url, new Response(data)).catch(() => {});
      return data;
    }
  } catch (e) { /* no cache here: download as usual */ }
  return srv.bytes('GET', `/files/${id}/content`);
}

// A few downloads at a time: each one costs the server a database check, so a wall of twenty pictures must not ask for all at once
// and push the posts' own requests to the back of the line. Newest posts are asked for first.
let running = 0; const waiting = [];
const slot = () => new Promise((go) => { if (running < 4) { running++; go(); } else waiting.push(go); });
const free = () => { const n = waiting.shift(); if (n) n(); else running--; };

async function fetchUrl(srv, ref) {
  await slot();
  let data;
  try { data = await cachedBytes(srv, ref.fileId); } finally { free(); }
  const k = await subtle().importKey('raw', unb64(ref.key), 'AES-GCM', false, ['decrypt']);
  // files made before the id was authenticated have no extra data: try the new form first, then the old
  let pt;
  try { pt = await subtle().decrypt({ name: 'AES-GCM', iv: data.subarray(0, 12), additionalData: new TextEncoder().encode('hub/file/v1|' + ref.fileId) }, k, data.subarray(12)); }
  catch (e) { pt = await subtle().decrypt({ name: 'AES-GCM', iv: data.subarray(0, 12) }, k, data.subarray(12)); }
  const url = URL.createObjectURL(new Blob([pt], { type: safeType(ref.type) }));
  urls.set(ref.fileId, url);
  return url;
}
