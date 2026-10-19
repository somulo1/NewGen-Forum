const escapeHtml = (text) => String(text)
  .replace(/&/g, '&amp;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;')
  .replace(/"/g, '&quot;')
  .replace(/'/g, '&#39;');

export function richTextHtml(value) {
  const escaped = escapeHtml(value || '');
  const links = [];
  const linked = escaped.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (match, label, target) => {
    const rawTarget = target.replace(/&amp;/g, '&');
    const candidate = /^[a-z][a-z\d+.-]*:/i.test(rawTarget) ? rawTarget : 'https://' + rawTarget;
    try {
      const url = new URL(candidate);
      if (!['http:', 'https:', 'mailto:'].includes(url.protocol)) return match;
      const href = escapeHtml(url.toString());
      const labelHtml = label
        .replace(/\*\*([^*\n]+)\*\*/g, '<strong>$1</strong>')
        .replace(/(?<!\*)\*([^*\n]+)\*(?!\*)/g, '<em>$1</em>')
        .replace(/\+\+([^+\n]+)\+\+/g, '<u>$1</u>')
        .replace(/\n/g, '<br>');
      const marker = '\u0000' + links.length + '\u0000';
      links.push('<a href="' + href + '" target="_blank" rel="noopener noreferrer" style="color:var(--k-B8F36B,#B8F36B);text-decoration:underline">' + labelHtml + '</a>');
      return marker;
    } catch {
      return match;
    }
  });
  const formatted = linked
    .replace(/\*\*([^*\n]+)\*\*/g, '<strong>$1</strong>')
    .replace(/(?<!\*)\*([^*\n]+)\*(?!\*)/g, '<em>$1</em>')
    .replace(/\+\+([^+\n]+)\+\+/g, '<u>$1</u>')
    .replace(/\n/g, '<br>');
  return formatted.replace(/\u0000(\d+)\u0000/g, (marker, index) => links[Number(index)] || marker);
}
