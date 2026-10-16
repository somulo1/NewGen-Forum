// Sharing a post outside Hub: opens the other site with the whole post (author, text, poll options and image links) filled in.
// Hub cannot attach local files to another site, so images go as links; uploaded images that only exist in this browser are listed by name.
const abs = (u) => (u && /^(blob:|data:)/.test(u) ? '' : u ? (/^https?:/.test(u) ? u : (typeof location !== 'undefined' ? location.origin : '') + u) : '');

export function postText(p, pollLabels) {
  const lines = [p.author + (p.meta ? ' · ' + p.meta : ''), '', p.text || ''];
  if (pollLabels && pollLabels.length) lines.push('', 'Poll:', ...pollLabels.map((l) => '• ' + l));
  (p.media || []).forEach((m) => { const u = abs(m.url); lines.push(u ? (m.isImg ? 'Image: ' : m.isVideo ? 'Video: ' : 'File: ') + u : 'Attached: ' + m.name); });
  return lines.join('\n').trim();
}

export function shareUrl(app, p, pollLabels) {
  const text = postText(p, pollLabels), enc = encodeURIComponent, site = typeof location !== 'undefined' ? location.origin : '';
  switch (app) {
    case 'WhatsApp': return 'https://wa.me/?text=' + enc(text);
    case 'X': return 'https://twitter.com/intent/tweet?text=' + enc(text.slice(0, 270));
    case 'LinkedIn': return 'https://www.linkedin.com/feed/?shareActive=true&text=' + enc(text);
    case 'Facebook': return 'https://www.facebook.com/sharer/sharer.php?u=' + enc(site) + '&quote=' + enc(text);
    default: return 'mailto:?subject=' + enc('Post from ' + p.author) + '&body=' + enc(text);
  }
}
