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

const urls = new Map();
const opening = new Map();
// Downloads and decrypts an attachment; returns an object URL (cached for the session; one download per file at a time).
export async function openUrl(srv, ref) {
  if (urls.has(ref.fileId)) return urls.get(ref.fileId);
  if (!opening.has(ref.fileId)) opening.set(ref.fileId, fetchUrl(srv, ref).catch(async (e) => { await new Promise((r) => setTimeout(r, 1500)); try { return await fetchUrl(srv, ref); } catch (e2) { console.warn('Could not open attachment ' + ref.fileId + ':', e2); throw e2; } }).finally(() => opening.delete(ref.fileId)));
  return opening.get(ref.fileId);
}
async function fetchUrl(srv, ref) {
  const data = await srv.bytes('GET', `/files/${ref.fileId}/content`);
  const k = await subtle().importKey('raw', unb64(ref.key), 'AES-GCM', false, ['decrypt']);
  // files made before the id was authenticated have no extra data: try the new form first, then the old
  let pt;
  try { pt = await subtle().decrypt({ name: 'AES-GCM', iv: data.subarray(0, 12), additionalData: new TextEncoder().encode('hub/file/v1|' + ref.fileId) }, k, data.subarray(12)); }
  catch (e) { pt = await subtle().decrypt({ name: 'AES-GCM', iv: data.subarray(0, 12) }, k, data.subarray(12)); }
  const url = URL.createObjectURL(new Blob([pt], { type: ref.type || (ref.isImg ? 'image/*' : 'application/octet-stream') }));
  urls.set(ref.fileId, url);
  return url;
}
