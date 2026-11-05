// Whiteboards and diagrams: shapes, connectors and pen strokes kept as data. This draws one as a picture for pages and PDFs.
const esc = (t) => String(t ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
const num = (v) => (Number.isFinite(Number(v)) ? Number(v) : 0);
const col = (c) => (/^#[0-9a-f]{3,8}$/i.test(c || '') ? c : '#FFE08A');

export const parseBoard = (body) => { try { const b = JSON.parse(body || '{}'); return { shapes: b.shapes || [], lines: b.lines || [], pen: b.pen || [] }; } catch (e) { return { shapes: [], lines: [], pen: [] }; } };
export const centre = (s) => ({ x: num(s.x) + num(s.w) / 2, y: num(s.y) + num(s.h) / 2 });

export function boardBox(b) {
  const xs = [], ys = [];
  b.shapes.forEach((s) => { xs.push(num(s.x), num(s.x) + num(s.w)); ys.push(num(s.y), num(s.y) + num(s.h)); });
  b.pen.forEach((p) => (p.pts || []).forEach(([x, y]) => { xs.push(num(x)); ys.push(num(y)); }));
  if (!xs.length) return { x: 0, y: 0, w: 400, h: 200 };
  const x = Math.min(...xs) - 20, y = Math.min(...ys) - 20;
  return { x, y, w: Math.max(...xs) + 20 - x, h: Math.max(...ys) + 20 - y };
}

export function boardSvg(body) {
  const b = parseBoard(body);
  const box = boardBox(b);
  const byId = new Map(b.shapes.map((s) => [s.id, s]));
  const parts = [];
  b.pen.forEach((p) => parts.push('<polyline fill="none" stroke="' + col(p.color) + '" stroke-width="2" points="' + (p.pts || []).map(([x, y]) => num(x) + ',' + num(y)).join(' ') + '"/>'));
  b.lines.forEach((l) => { const f = byId.get(l.from), t = byId.get(l.to); if (!f || !t) return; const a = centre(f), z = centre(t); parts.push('<line x1="' + a.x + '" y1="' + a.y + '" x2="' + z.x + '" y2="' + z.y + '" stroke="#555" stroke-width="2" marker-end="url(#a)"/>'); });
  b.shapes.forEach((s) => {
    const x = num(s.x), y = num(s.y), w = num(s.w), h = num(s.h), f = col(s.color);
    if (s.t === 'ellipse') parts.push('<ellipse cx="' + (x + w / 2) + '" cy="' + (y + h / 2) + '" rx="' + w / 2 + '" ry="' + h / 2 + '" fill="' + f + '" stroke="#555"/>');
    else if (s.t === 'diamond') parts.push('<polygon points="' + (x + w / 2) + ',' + y + ' ' + (x + w) + ',' + (y + h / 2) + ' ' + (x + w / 2) + ',' + (y + h) + ' ' + x + ',' + (y + h / 2) + '" fill="' + f + '" stroke="#555"/>');
    else if (s.t !== 'text') parts.push('<rect x="' + x + '" y="' + y + '" width="' + w + '" height="' + h + '" rx="6" fill="' + f + '" stroke="#555"/>');
    parts.push('<text x="' + (x + w / 2) + '" y="' + (y + h / 2) + '" text-anchor="middle" dominant-baseline="middle" font-family="sans-serif" font-size="14" fill="#111">' + esc(s.text) + '</text>');
  });
  return '<svg xmlns="http://www.w3.org/2000/svg" viewBox="' + [box.x, box.y, box.w, box.h].join(' ') + '" width="' + Math.round(box.w) + '" height="' + Math.round(box.h) + '" style="max-width:100%;height:auto;background:#fff;border-radius:8px"><defs><marker id="a" markerWidth="10" markerHeight="8" refX="9" refY="4" orient="auto"><path d="M0 0L10 4L0 8z" fill="#555"/></marker></defs>' + parts.join('') + '</svg>';
}
export const boardDataUrl = (body) => 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(boardSvg(body));
