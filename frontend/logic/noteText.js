// Note bodies are HTML (rich text); older ones are plain text. This gives the plain text for previews, search and word counts.
export function noteText(b) {
  b = b || '';
  if (!/^\s*</.test(b)) return b;
  return b.replace(/<\/(p|div|h[1-6]|li|tr|blockquote|pre)>|<br\s*\/?>/gi, '\n').replace(/<\/(td|th)>/gi, '\t').replace(/<[^>]*>/g, '')
    .replace(/&nbsp;/g, ' ').replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').replace(/&#39;/g, "'").replace(/&amp;/g, '&')
    .replace(/\t\n/g, '\n').replace(/\n{2,}/g, '\n').trim();
}

// Sharing is a live link, not a copy: it works for as long as the share exists. What is on anyone's Home page does not change it.
export const isMine = (n) => !n.owner || n.owner === 'Sam';
export const permLabel = (p) => (p === 'edit' ? 'Can edit' : 'Read only');
