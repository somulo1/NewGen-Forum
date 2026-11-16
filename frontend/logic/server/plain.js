// The community wall (every wall: the organization's and each department's, branch's ...) is not end-to-end encrypted: the server
// reads, checks and keeps its posts, comments, moments, likes and votes. Forums and private chats stay end-to-end encrypted.
// These helpers send to the right kind: a plain body and plain attachments for a wall, a sealed envelope everywhere else.
import * as files from './files.js';
import { newId } from './e2ee.js';

const kinds = new Map();   // conversation id -> kind
export async function isPlainConv(srv, conv) {
  if (!kinds.has(conv)) {
    const c = await srv.vault.conversation(conv).catch(() => null);
    kinds.set(conv, c ? c.kind : '');
  }
  return kinds.get(conv) === 'wall';
}
export const markPlain = (conv) => kinds.set(conv, 'wall');

// { id, body } for a wall, or the vault's sealed { id, epoch, device, envelope } elsewhere
export async function sealFor(srv, conv, payload, id) {
  if (await isPlainConv(srv, conv)) return { id: id || newId(), body: payload, plain: true };
  return srv.vault.seal(conv, payload, id);
}

// an attachment: a plain file for a wall (no key), an encrypted one elsewhere
export async function uploadFor(srv, conv, f) {
  if (!(await isPlainConv(srv, conv))) return files.upload(srv, conv, f);
  const q = new URLSearchParams({ name: f.name, type: f.type || '' });
  const up = await srv.bytes('PUT', `/conversations/${conv}/plain-files?${q}`, new Uint8Array(await f.arrayBuffer()));
  const kind = files.kindOf(f.name, f.type);
  const size = f.size >= 1048576 ? (f.size / 1048576).toFixed(1) + ' MB' : Math.max(1, Math.round(f.size / 1024)) + ' KB';
  files.remember(up.id, URL.createObjectURL(f));
  return { fileId: up.id, name: f.name, meta: size, type: f.type || '', isImg: kind === 'IMG', isVideo: kind === 'VID', plain: true };
}
