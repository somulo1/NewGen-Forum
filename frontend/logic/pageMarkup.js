import { richTextHtml } from '~/logic/richText.js';

// Project pages: headings, lists, checklists, tables, quotes, code and rules from plain text. Everything goes through richTextHtml, which
// escapes it first. [[Page title]] links another page, and a task key such as PRJ-01-12 links its task.
const esc = (t) => String(t).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

export function pageHtml(body, taskKeys = []) {
  const keys = new Set(taskKeys);
  const inline = (t) => richTextHtml(t)
    .replace(/\[\[([^\]]{1,200})\]\]/g, (m, title) => '<a href="#" data-page="' + title + '">' + title + '</a>')
    .replace(/\b[A-Z]{2,5}-\d+-\d+\b/g, (m) => (keys.has(m) ? '<a href="#" data-task="' + m + '">' + m + '</a>' : m));
  const out = [];
  const lines = String(body || '').split('\n');
  for (let i = 0; i < lines.length; i++) {
    const l = lines[i];
    if (/^```/.test(l)) {
      const code = [];
      for (i++; i < lines.length && !/^```/.test(lines[i]); i++) code.push(lines[i]);
      out.push('<pre style="padding:10px 12px;border-radius:8px;background:var(--k-121B17,#121B17);overflow-x:auto;font-size:13px"><code>' + esc(code.join('\n')) + '</code></pre>');
    } else if (/^#{1,3} /.test(l)) {
      const n = l.indexOf(' ');
      out.push('<h' + (n + 2) + ' style="margin:14px 0 6px">' + inline(l.slice(n + 1)) + '</h' + (n + 2) + '>');
    } else if (/^---+$/.test(l.trim())) {
      out.push('<hr style="border:0;border-top:1px solid var(--k-2A3832,#2A3832);margin:12px 0">');
    } else if (/^> /.test(l)) {
      out.push('<blockquote style="margin:6px 0;padding:6px 12px;border-left:3px solid var(--k-7CC4FF,#7CC4FF)">' + inline(l.slice(2)) + '</blockquote>');
    } else if (/^\s*([-*]|\d+\.) /.test(l)) {
      const ordered = /^\s*\d+\./.test(l);
      const items = [];
      for (; i < lines.length && /^\s*([-*]|\d+\.) /.test(lines[i]); i++) {
        const t = lines[i].replace(/^\s*([-*]|\d+\.) /, '');
        const box = /^\[( |x)\] /i.exec(t);
        items.push('<li>' + (box ? (box[1] === ' ' ? '☐ ' : '☑ ') + inline(t.slice(4)) : inline(t)) + '</li>');
      }
      i--;
      out.push('<' + (ordered ? 'ol' : 'ul') + ' style="margin:6px 0;padding-left:22px">' + items.join('') + '</' + (ordered ? 'ol' : 'ul') + '>');
    } else if (/^\|.*\|\s*$/.test(l)) {
      const rows = [];
      for (; i < lines.length && /^\|.*\|\s*$/.test(lines[i]); i++) rows.push(lines[i].trim().slice(1, -1).split('|').map((c) => c.trim()));
      i--;
      const cell = (tag, c) => '<' + tag + ' style="padding:6px 10px;border:1px solid var(--k-2A3832,#2A3832);text-align:left">' + inline(c) + '</' + tag + '>';
      const body2 = rows.filter((r) => !r.every((c) => /^:?-+:?$/.test(c)));
      out.push('<div style="overflow-x:auto"><table style="border-collapse:collapse;margin:6px 0;font-size:13px">' + body2.map((r, n) => '<tr>' + r.map((c) => cell(n ? 'td' : 'th', c)).join('') + '</tr>').join('') + '</table></div>');
    } else if (!l.trim()) {
      out.push('<div style="height:8px"></div>');
    } else {
      out.push('<p style="margin:2px 0">' + inline(l) + '</p>');
    }
  }
  return out.join('');
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
