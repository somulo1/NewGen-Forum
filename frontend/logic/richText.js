const escapeHtml = (text) => String(text)
  .replace(/&/g, '&amp;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;')
  .replace(/"/g, '&quot;')
  .replace(/'/g, '&#39;');

const LINK_STYLE = 'color:var(--k-B8F36B,#B8F36B);text-decoration:underline;overflow-wrap:anywhere';
const linkTag = (href, labelHtml) => '<a href="' + escapeHtml(href) + '" target="_blank" rel="noopener noreferrer" style="' + LINK_STYLE + '">' + labelHtml + '</a>';
const hold = (links, tag) => { links.push(tag); return '\u0000' + (links.length - 1) + '\u0000'; };
const putBack = (html, links) => html.replace(/\u0000(\d+)\u0000/g, (marker, index) => links[Number(index)] || marker);
const marks = (html) => html
  .replace(/\*\*([^*\n]+)\*\*/g, '<strong>$1</strong>')
  .replace(/(?<!\*)\*([^*\n]+)\*(?!\*)/g, '<em>$1</em>')
  .replace(/\+\+([^+\n]+)\+\+/g, '<u>$1</u>')
  .replace(/\n/g, '<br>');

// [label](address), made with the link button
function setLinks(escaped, links) {
  return escaped.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (match, label, target) => {
    const rawTarget = target.replace(/&amp;/g, '&');
    const candidate = /^[a-z][a-z\d+.-]*:/i.test(rawTarget) ? rawTarget : 'https://' + rawTarget;
    try {
      const url = new URL(candidate);
      if (!['http:', 'https:', 'mailto:'].includes(url.protocol)) return match;
      return hold(links, linkTag(url.toString(), marks(label)));
    } catch {
      return match;
    }
  });
}

// Addresses simply typed or pasted: https://..., www..., example.com/path and name@example.com. The text is already escaped.
const BARE = /(?<![\w@/.\-\u0000])(?:([\w.+-]+@[a-z\d-]+(?:\.[a-z\d-]+)+)|((?:https?:\/\/|www\.)[^\s\u0000]+|(?:[a-z\d-]+\.)+(?:com|org|net|edu|gov|io|co|ai|app|dev|info|biz|ke)(?![\w-])(?:[/?#][^\s\u0000]*)?))/gi;
const closes = { ')': '(', ']': '[', '}': '{' };
// Punctuation around an address belongs to the sentence: "(see https://a.com/x)." links only https://a.com/x
function trimAddress(text) {
  const quote = text.search(/&(?:quot|#39|lt|gt);/);
  let end = quote === -1 ? text.length : quote;
  for (;;) {
    const ch = text[end - 1];
    if (ch && '.,;:!?*+_~'.includes(ch)) { end--; continue; }
    if (text.slice(0, end).endsWith('&amp;')) { end -= 5; continue; }
    if (closes[ch]) {
      const part = text.slice(0, end);
      if (part.split(ch).length > part.split(closes[ch]).length) { end--; continue; }
    }
    return end;
  }
}
function bareLinks(escaped, links) {
  return escaped.replace(BARE, (match, email, address) => {
    if (email) {
      const end = trimAddress(email);
      return hold(links, linkTag('mailto:' + email.slice(0, end).replace(/&amp;/g, '&'), email.slice(0, end))) + email.slice(end);
    }
    const end = trimAddress(address);
    const shown = address.slice(0, end), rest = address.slice(end);
    const raw = shown.replace(/&amp;/g, '&');
    try {
      const url = new URL(/^https?:\/\//i.test(raw) ? raw : 'https://' + raw);
      if (!url.hostname.includes('.') && url.hostname !== 'localhost') return match;
      return hold(links, linkTag(url.toString(), shown)) + rest;
    } catch {
      return match;
    }
  });
}

export function richTextHtml(value) {
  const links = [];
  return putBack(marks(bareLinks(setLinks(escapeHtml(value || ''), links), links)), links);
}

// For text with no formatting marks (memos, job posts, project goals): only the links are made clickable. Line breaks are left to CSS.
export function linkTextHtml(value) {
  const links = [];
  return putBack(bareLinks(setLinks(escapeHtml(value ?? ''), links), links), links);
}

// A tagged person (@Name) in a comment becomes a link that opens their profile (the click is handled where the list is shown).
export function mentionLinker(names) {
  const list = [...new Set(names)].filter(Boolean).sort((a, b) => b.length - a.length);
  if (!list.length) return (h) => h;
  const re = new RegExp('(?<![\\w/:.-])@(' + list.map((n) => escapeHtml(n).replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|') + ')(?![\\w])', 'g');
  return (html) => html.replace(re, (m, name) => '<a href="#" data-mention="' + name + '" style="color:var(--k-C9B6FF,#C9B6FF);font-weight:600;text-decoration:none">@' + name + '</a>');
}


// #hashtags: found in a text, and turned into links the wall filters by.
const TAG = /(^|[\s>(])#([\p{L}\p{N}_]{2,40})/gu;
export function hashtagsOf(text) {
  const out = [];
  String(text || '').replace(TAG, (m, pre, tag) => { out.push(tag); return m; });
  return out;
}
export function linkHashtags(html) {
  return html.replace(TAG, (m, pre, tag) => pre + '<a href="#" data-hashtag="' + tag + '" style="color:var(--k-7CC4FF,#7CC4FF);font-weight:600;text-decoration:none">#' + tag + '</a>');
}
