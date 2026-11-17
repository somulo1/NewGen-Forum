// Files the web interface hands over (a download, an attachment to open, an export): written to the app's cache and offered
// through the system sheet, where the person can open, save or send them. The cache is the app's own; nothing is kept elsewhere.
import { File, Paths } from 'expo-file-system';
import * as Sharing from 'expo-sharing';

const EXT: Record<string, string> = {
  'application/pdf': 'pdf', 'image/png': 'png', 'image/jpeg': 'jpg', 'image/gif': 'gif', 'image/webp': 'webp', 'image/svg+xml': 'svg',
  'text/csv': 'csv', 'text/plain': 'txt', 'text/html': 'html', 'application/json': 'json', 'application/zip': 'zip',
  'audio/wav': 'wav', 'audio/webm': 'webm', 'audio/mpeg': 'mp3', 'audio/mp4': 'm4a', 'video/mp4': 'mp4', 'video/webm': 'webm',
  'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet': 'xlsx',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document': 'docx',
  'application/vnd.openxmlformats-officedocument.presentationml.presentation': 'pptx'
};

function safeName(name: string, mime: string): string {
  let n = String(name || '').split(/[\\/]/).pop() || '';
  n = n.replace(/[^\w.\- ()]+/g, '_').replace(/^\.+/, '').slice(0, 120);
  if (!n) n = 'hub-file-' + new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-');
  if (!/\.[a-z0-9]{1,5}$/i.test(n) && EXT[mime]) n += '.' + EXT[mime];
  return n;
}

export async function saveAndShare(name: string, mime: string, base64: string): Promise<void> {
  const file = new File(Paths.cache, safeName(name, mime));
  if (file.exists) file.delete();
  file.create();
  await Promise.resolve(file.write(base64, { encoding: 'base64' }));
  if (await Sharing.isAvailableAsync()) {
    await Sharing.shareAsync(file.uri, { mimeType: mime || 'application/octet-stream', dialogTitle: file.name });
  }
}
