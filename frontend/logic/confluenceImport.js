// Reads a Confluence space export (Space tools > Content tools > Export > HTML): one HTML file a page, attachments in attachments/<page id>/.
// Everything happens in the browser. It returns pages parents-first; the caller saves them.

// a .zip, read with the browser's own decompression
export async function readZip(file) {
  const buf = await file.arrayBuffer();
  const v = new DataView(buf);
  let end = -1;
  for (let i = buf.byteLength - 22; i >= 0 && i >= buf.byteLength - 66000; i--) if (v.getUint32(i, true) === 0x06054b50) { end = i; break; }
  if (end < 0) throw new Error('That file is not a .zip export.');
  const count = v.getUint16(end + 10, true);
  let p = v.getUint32(end + 16, true);
  const out = new Map();
  const dec = new TextDecoder();
  for (let n = 0; n < count && v.getUint32(p, true) === 0x02014b50; n++) {
    const method = v.getUint16(p + 10, true), size = v.getUint32(p + 20, true), nl = v.getUint16(p + 28, true), xl = v.getUint16(p + 30, true), cl = v.getUint16(p + 32, true), at = v.getUint32(p + 42, true);
    const name = dec.decode(new Uint8Array(buf, p + 46, nl));
    p += 46 + nl + xl + cl;
    if (name.endsWith('/')) continue;
    out.set(name, async () => {
      const start = at + 30 + v.getUint16(at + 26, true) + v.getUint16(at + 28, true);
      const raw = new Blob([new Uint8Array(buf, start, size)]);
      if (method === 0) return raw;
      if (method !== 8) throw new Error('Unsupported compression in the export.');
      return new Response(raw.stream().pipeThrough(new DecompressionStream('deflate-raw'))).blob();
    });
  }
  return out;
}

const MIME = { png: 'image/png', jpg: 'image/jpeg', jpeg: 'image/jpeg', gif: 'image/gif', webp: 'image/webp', svg: '', pdf: 'application/pdf' };
const extOf = (n) => (n.split('.').pop() || '').toLowerCase();

// Pictures live inside the page, so they are made small enough to keep the page light.
export async function imageDataUrl(blob, max = 1400) {
  const bmp = await createImageBitmap(blob);
  const k = Math.min(1, max / Math.max(bmp.width, bmp.height));
  const c = document.createElement('canvas');
  c.width = Math.round(bmp.width * k); c.height = Math.round(bmp.height * k);
  c.getContext('2d').drawImage(bmp, 0, 0, c.width, c.height);
  bmp.close();
  const png = c.toDataURL('image/png');
  return png.length < 300000 ? png : c.toDataURL('image/jpeg', 0.82);
}

const MAX_BODY = 6 << 20;

export async function parseExport(file, onProgress = () => {}) {
  const zip = await readZip(file);
  const names = [...zip.keys()];
  if (names.some((n) => /(^|\/)entities\.xml$/.test(n))) throw new Error('This is an XML export. In Confluence choose the HTML export instead, then try again.');
  const html = names.filter((n) => /\.html$/i.test(n) && !/(^|\/)(styles|images|attachments)\//.test(n));
  if (!html.length) throw new Error('No pages found. Use a Confluence HTML space export.');
  const base = (n) => n.split('/').pop();
  const root = html[0].includes('/') ? html[0].slice(0, html[0].lastIndexOf('/') + 1) : '';
  const parser = new DOMParser();
  const docs = new Map();
  for (const n of html) {
    if (base(n) === 'index.html') continue;
    const d = parser.parseFromString(await (await zip.get(n)()).text(), 'text/html');
    const main = d.querySelector('#main-content') || d.querySelector('.wiki-content');
    if (!main) continue;
    const raw = (d.querySelector('#title-text') || d.querySelector('title') || {}).textContent || base(n);
    const title = raw.replace(/^[^:]*\s:\s/, '').trim().slice(0, 190) || base(n);
    const crumbs = [...d.querySelectorAll('#breadcrumbs a')].map((a) => base(a.getAttribute('href') || '')).filter((h) => h && h !== 'index.html');
    docs.set(base(n), { file: base(n), id: (/(\d+)\.html$/.exec(base(n)) || [])[1] || '', title, main, parentFile: crumbs[crumbs.length - 1] || '' });
  }
  // titles must be unique for page links
  const seen = new Map();
  docs.forEach((x) => { const k = x.title.toLowerCase(); seen.set(k, (seen.get(k) || 0) + 1); if (seen.get(k) > 1) x.title += ' (' + seen.get(k) + ')'; });
  const pages = [];
  const problems = [];
  let done = 0;
  for (const x of docs.values()) {
    x.main.querySelectorAll('script,style,iframe,object,embed,form').forEach((e) => e.remove());
    for (const a of x.main.querySelectorAll('a[href]')) {
      const h = a.getAttribute('href');
      const target = docs.get(base(h.split('#')[0]));
      if (target && !/^[a-z]+:/i.test(h)) a.replaceWith(x.main.ownerDocument.createTextNode('[[' + target.title + ']]'));
      else if (h.startsWith('attachments/') || !/^(https?:|mailto:)/i.test(h)) a.removeAttribute('href');
    }
    const inline = new Set();
    for (const img of x.main.querySelectorAll('img')) {
      const src = (img.getAttribute('src') || '').split('?')[0];
      const get = zip.get(root + src);
      if (/^https?:/i.test(src)) continue;
      if (!get || !MIME[extOf(src)] || !MIME[extOf(src)].startsWith('image/')) { img.remove(); if (get) problems.push(x.title + ': a picture stayed as an attachment'); continue; }
      try { img.setAttribute('src', await imageDataUrl(await get())); inline.add(root + src); [...img.attributes].forEach((t) => { if (!['src', 'alt'].includes(t.name)) img.removeAttribute(t.name); }); } catch (e) { img.remove(); problems.push(x.title + ': a picture could not be read'); }
    }
    let body = x.main.innerHTML.trim();
    if (body.length > MAX_BODY) {
      x.main.querySelectorAll('img').forEach((i) => i.remove());
      inline.clear();
      body = x.main.innerHTML.trim();
      problems.push(x.title + ': too many large pictures, so they became attachments');
    }
    const files = names.filter((n) => x.id && n.startsWith(root + 'attachments/' + x.id + '/') && !inline.has(n));
    pages.push({ file: x.file, title: x.title, parentFile: docs.has(x.parentFile) ? x.parentFile : '', body: /^\s*</.test(body) ? body : '<p>' + body + '</p>', attachments: files.map((n) => ({ name: base(n), read: zip.get(n) })) });
    onProgress(++done, docs.size);
  }
  // parents first
  const ordered = [];
  const put = (parent) => pages.filter((g) => g.parentFile === parent).forEach((g) => { ordered.push(g); put(g.file); });
  put('');
  pages.forEach((g) => { if (!ordered.includes(g)) ordered.push(g); });
  return { pages: ordered, problems };
}
