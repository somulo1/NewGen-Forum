// Files: the folders on the Files page and the files in them. A folder is an encrypted conversation; each file's name and key are
// sealed for its members and the bytes are stored encrypted (see ../files.js).
import * as files from '../files.js';

export const name = 'folders';

const ago = (iso) => { const m = Math.max(0, Math.round((Date.now() - new Date(iso)) / 60000)); return m < 60 ? (m < 1 ? 'Just now' : m + ' min ago') : m < 1440 ? Math.round(m / 60) + ' h ago' : Math.round(m / 1440) === 1 ? 'Yesterday' : Math.round(m / 1440) + ' days ago'; };

async function readFolder(srv, c) {
  const [{ files: list }, detail] = await Promise.all([srv.request('GET', `/folders/${c.id}/files`), srv.request('GET', '/conversations/' + c.id)]);
  const opened = await Promise.all((list || []).filter((f) => f.status !== 'deleted').map(async (f) => {
    const up = srv.dirByUid && srv.dirByUid.get(f.uploader);
    const pl = await srv.vault.openItem({ conversation: c.id, id: f.id, sender: up ? up.account : srv.me.account, senderPerson: f.uploader, senderDevice: f.author_device, epoch: f.epoch, envelope: f.envelope });
    return pl ? { name: pl.name, kind: pl.kind || 'ZIP', size: pl.size_label || '', by: pl.by || f.uploader_name, when: ago(f.created_at), at: f.created_at, fileId: f.id, key: pl.file_key, type: pl.type || '', isImg: pl.kind === 'IMG', url: '' } : null;
  }));
  const items = opened.filter(Boolean).sort((a, b) => (a.at < b.at ? 1 : -1));
  return { id: c.id, name: detail.name || c.name, color: detail.color || '#7CC4FF', audience: (detail.member_count || 1) + ' people', updated: items.length ? items[0].when : 'empty', files: items };
}

export async function load(logic, srv) {
  const { conversations } = await srv.request('GET', '/conversations?kind=folder&limit=200');
  const folders = (await Promise.all(conversations.map((c) => readFolder(srv, c).catch(() => null)))).filter(Boolean);
  logic.setState({ folders });
}

export async function upload(logic, srv, folder, list) {
  for (const f of list) await files.upload(srv, folder, f);
  await load(logic, srv);
}

export async function remove(logic, srv, f) {
  await srv.request('DELETE', '/files/' + f.fileId);
  await load(logic, srv);
}

// Makes a link for someone outside Hub. The link key stays in the part after # and never reaches the server.
export async function share(srv, f) {
  const lk = globalThis.crypto.getRandomValues(new Uint8Array(32));
  const iv = globalThis.crypto.getRandomValues(new Uint8Array(12));
  const k = await globalThis.crypto.subtle.importKey('raw', lk, 'AES-GCM', false, ['encrypt']);
  const ct = new Uint8Array(await globalThis.crypto.subtle.encrypt({ name: 'AES-GCM', iv }, k, new TextEncoder().encode(f.key)));
  const env = new Uint8Array(12 + ct.length); env.set(iv); env.set(ct, 12);
  const b64 = (u) => btoa(String.fromCharCode(...u)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  const r = await srv.request('POST', `/files/${f.fileId}/shares`, { expires_in_hours: 24, max_downloads: 0, key_envelope: btoa(String.fromCharCode(...env)) });
  const id = (r.share && r.share.token) || '';
  const link = location.origin + '/share/' + id + '#' + b64(lk);
  try { await navigator.clipboard.writeText(link); } catch { /* clipboard can be blocked */ }
  return { link, pending: !!r.approval };
}

// Decrypts a file and opens it in a new tab.
export async function open(srv, f) {
  const url = await files.openUrl(srv, { fileId: f.fileId, key: f.key, type: f.type, isImg: f.isImg });
  window.open(url, '_blank', 'noopener');
}
