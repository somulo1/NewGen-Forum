// Project pages are written in the notebook's editor, which saves HTML. Pages written before it are plain text and open as paragraphs.
const esc = (t) => String(t).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
const isHtml = (b) => /^\s*<(p|h[1-6]|ul|ol|table|blockquote|pre|hr|div)\b/i.test(b || '');
const attr = (t) => esc(t).replace(/"/g, '&quot;');
// For reading: page and task references become links, and text with a comment on it is marked. It never changes what is stored.
export function decorate(html, keys = [], comments = []) {
  const known = new Set(keys);
  const left = comments.filter((c) => c.quote && !c.resolved).map((c) => ({ id: c.id, quote: esc(c.quote) }));
  return String(html).split(/(<[^>]+>)/).map((part, i) => {
    if (i % 2) return part;
    let t = part;
    left.forEach((c) => {
      if (c.done) return;
      const at = t.indexOf(c.quote);
      if (at === -1) return;
      c.done = true;
      t = t.slice(0, at) + '\u0001' + c.id + '\u0002' + c.quote + '\u0003' + t.slice(at + c.quote.length);
    });
    t = t.replace(/\[\[([^\]\u0001-\u0003]{1,200})\]\]/g, (m, title) => '<span data-ref="page:' + attr(title) + '">' + title + '</span>');
    t = t.replace(/\b[A-Z]{2,5}-\d+-\d+\b/g, (m) => (known.has(m) ? '<span data-ref="task:' + m + '">' + m + '</span>' : m));
    return t.replace(/\u0001([^\u0002]*)\u0002/g, (m, id) => '<span data-comment="' + attr(id) + '">').replace(/\u0003/g, '</span>');
  }).join('');
}

export const pageHtml = (body) => (isHtml(body) ? body : String(body || '').split('\n').map((l) => '<p>' + esc(l) + '</p>').join(''));
// The words of a page, one block a line: for search, links and comparing versions.
export function pageText(body) {
  if (!isHtml(body)) return String(body || '');
  return String(body).replace(/<\/(p|h[1-6]|li|tr|blockquote|pre|div)>|<hr[^>]*>|<br[^>]*>/gi, '\n').replace(/<\/t[dh]>/gi, ' | ').replace(/<[^>]+>/g, '')
    .replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/&nbsp;/g, ' ').replace(/&amp;/g, '&');
}

// What changed between two versions, line by line (longest common subsequence).
export function lineDiff(a, b) {
  const A = String(a || '').split('\n').slice(0, 800), B = String(b || '').split('\n').slice(0, 800);
  const T = Array.from({ length: A.length + 1 }, () => new Uint16Array(B.length + 1));
  for (let i = A.length - 1; i >= 0; i--) for (let j = B.length - 1; j >= 0; j--) T[i][j] = A[i] === B[j] ? T[i + 1][j + 1] + 1 : Math.max(T[i + 1][j], T[i][j + 1]);
  const out = [];
  let i = 0, j = 0;
  while (i < A.length || j < B.length) {
    if (i < A.length && j < B.length && A[i] === B[j]) { out.push({ k: 'same', t: A[i] }); i++; j++; }
    else if (j < B.length && (i >= A.length || T[i][j + 1] >= T[i + 1][j])) out.push({ k: 'add', t: B[j++] });
    else out.push({ k: 'del', t: A[i++] });
  }
  return out;
}
