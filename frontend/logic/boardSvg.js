// Whiteboards and diagrams, kept as data: shapes, connectors and pen strokes. The shape library and connector routing live here, so the
// editor, pages and PDFs all draw the same picture.
const esc = (t) => String(t ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
const num = (v) => (Number.isFinite(Number(v)) ? Number(v) : 0);
const col = (c, d) => (/^#[0-9a-f]{3,8}$/i.test(c || '') ? c : d);
const poly = (w, h, l) => 'M' + l.map(([x, y]) => (x * w).toFixed(1) + ' ' + (y * h).toFixed(1)).join('L') + 'Z';
const oval = (cx, cy, rx, ry) => `M${cx - rx} ${cy}a${rx} ${ry} 0 1 0 ${rx * 2} 0a${rx} ${ry} 0 1 0 ${-rx * 2} 0Z`;
const round = (w, h, r) => `M${r} 0H${w - r}Q${w} 0 ${w} ${r}V${h - r}Q${w} ${h} ${w - r} ${h}H${r}Q0 ${h} 0 ${h - r}V${r}Q0 0 ${r} 0Z`;
const star = (w, h) => { const p = []; for (let i = 0; i < 10; i++) { const r = i % 2 ? 0.22 : 0.5, a = -Math.PI / 2 + (i * Math.PI) / 5; p.push([0.5 + r * Math.cos(a), 0.5 + r * Math.sin(a)]); } return poly(w, h, p); };
const SIDES = [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5]];
const WITH_CORNERS = [...SIDES, [0, 0], [1, 0], [1, 1], [0, 1]];

// Each shape: how it draws (d: filled outline, more: lines on top), and where connectors may attach (fractions of its box).
export const SHAPES = {
  rect: { label: 'Box', group: 'Basic', w: 150, h: 70, d: (w, h) => round(w, h, 6), ports: WITH_CORNERS },
  sticky: { label: 'Sticky note', group: 'Basic', w: 140, h: 110, fill: '#FFE08A', d: (w, h) => round(w, h, 4), ports: WITH_CORNERS },
  ellipse: { label: 'Circle', group: 'Basic', w: 110, h: 110, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2) },
  triangle: { label: 'Triangle', group: 'Basic', w: 120, h: 100, d: (w, h) => poly(w, h, [[0.5, 0], [1, 1], [0, 1]]), ports: [[0.5, 0], [0.75, 0.5], [0.5, 1], [0.25, 0.5], [0, 1], [1, 1]] },
  rtriangle: { label: 'Right triangle', group: 'Basic', w: 120, h: 100, d: (w, h) => poly(w, h, [[0, 0], [1, 1], [0, 1]]), ports: [[0, 0], [0.5, 0.5], [0.5, 1], [0, 0.5], [1, 1], [0, 1]] },
  diamond: { label: 'Diamond', group: 'Basic', w: 130, h: 90, d: (w, h) => poly(w, h, [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5]]) },
  pentagon: { label: 'Pentagon', group: 'Basic', w: 110, h: 105, d: (w, h) => poly(w, h, [[0.5, 0], [1, 0.38], [0.81, 1], [0.19, 1], [0, 0.38]]), ports: [[0.5, 0], [1, 0.38], [0.5, 1], [0, 0.38]] },
  hexagon: { label: 'Hexagon', group: 'Basic', w: 130, h: 80, d: (w, h) => poly(w, h, [[0.2, 0], [0.8, 0], [1, 0.5], [0.8, 1], [0.2, 1], [0, 0.5]]) },
  octagon: { label: 'Octagon', group: 'Basic', w: 100, h: 100, d: (w, h) => poly(w, h, [[0.3, 0], [0.7, 0], [1, 0.3], [1, 0.7], [0.7, 1], [0.3, 1], [0, 0.7], [0, 0.3]]) },
  star: { label: 'Star', group: 'Basic', w: 110, h: 110, d: star, ports: [[0.5, 0], [0.98, 0.35], [0.8, 0.9], [0.2, 0.9], [0.02, 0.35]] },
  arrowr: { label: 'Block arrow', group: 'Basic', w: 140, h: 70, d: (w, h) => poly(w, h, [[0, 0.25], [0.65, 0.25], [0.65, 0], [1, 0.5], [0.65, 1], [0.65, 0.75], [0, 0.75]]), ports: [[1, 0.5], [0, 0.5], [0.65, 0], [0.65, 1]] },
  cloud: { label: 'Cloud', group: 'Basic', w: 150, h: 90, d: (w, h) => `M${w * 0.25} ${h * 0.85}C${w * 0.02} ${h * 0.85} ${-w * 0.02} ${h * 0.45} ${w * 0.2} ${h * 0.4}C${w * 0.2} ${h * 0.08} ${w * 0.5} ${-h * 0.05} ${w * 0.6} ${h * 0.22}C${w * 0.78} ${h * 0.1} ${w * 0.92} ${h * 0.3} ${w * 0.85} ${h * 0.45}C${w * 1.04} ${h * 0.5} ${w} ${h * 0.87} ${w * 0.78} ${h * 0.85}Z` },
  text: { label: 'Text', group: 'Basic', w: 150, h: 40, noBox: true, d: (w, h) => `M0 0H${w}V${h}H0Z`, ports: SIDES },

  process: { label: 'Process', group: 'Flowchart', w: 150, h: 70, d: (w, h) => `M0 0H${w}V${h}H0Z`, ports: WITH_CORNERS },
  decision: { label: 'Decision', group: 'Flowchart', w: 140, h: 90, d: (w, h) => poly(w, h, [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5]]) },
  terminator: { label: 'Start / end', group: 'Flowchart', w: 140, h: 56, d: (w, h) => round(w, h, Math.min(h, w) / 2) },
  data: { label: 'Input / output', group: 'Flowchart', w: 150, h: 70, d: (w, h) => poly(w, h, [[0.18, 0], [1, 0], [0.82, 1], [0, 1]]), ports: [[0.59, 0], [0.91, 0.5], [0.41, 1], [0.09, 0.5]] },
  subprocess: { label: 'Subprocess', group: 'Flowchart', w: 150, h: 70, d: (w, h) => `M0 0H${w}V${h}H0Z`, more: (w, h) => `M${w * 0.1} 0V${h}M${w * 0.9} 0V${h}`, ports: WITH_CORNERS },
  document: { label: 'Document', group: 'Flowchart', w: 140, h: 85, d: (w, h) => `M0 0H${w}V${h * 0.85}C${w * 0.75} ${h * 0.65} ${w * 0.3} ${h * 1.1} 0 ${h * 0.88}Z` },
  database: { label: 'Database', group: 'Flowchart', w: 100, h: 110, d: (w, h) => { const r = h * 0.12; return `M0 ${r}A${w / 2} ${r} 0 0 1 ${w} ${r}V${h - r}A${w / 2} ${r} 0 0 1 0 ${h - r}Z`; }, more: (w, h) => { const r = h * 0.12; return `M0 ${r}A${w / 2} ${r} 0 0 0 ${w} ${r}`; } },
  manual: { label: 'Manual input', group: 'Flowchart', w: 140, h: 75, d: (w, h) => poly(w, h, [[0, 0.25], [1, 0], [1, 1], [0, 1]]), ports: [[0.5, 0.12], [1, 0.5], [0.5, 1], [0, 0.6]] },
  preparation: { label: 'Preparation', group: 'Flowchart', w: 150, h: 70, d: (w, h) => poly(w, h, [[0.2, 0], [0.8, 0], [1, 0.5], [0.8, 1], [0.2, 1], [0, 0.5]]) },
  delay: { label: 'Delay', group: 'Flowchart', w: 130, h: 70, d: (w, h) => `M0 0H${w * 0.6}A${w * 0.4} ${h / 2} 0 0 1 ${w * 0.6} ${h}H0Z` },
  connector: { label: 'On-page link', group: 'Flowchart', w: 50, h: 50, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2) },
  offpage: { label: 'Off-page link', group: 'Flowchart', w: 80, h: 80, d: (w, h) => poly(w, h, [[0, 0], [1, 0], [1, 0.7], [0.5, 1], [0, 0.7]]) },

  dfdprocess: { label: 'DFD process', group: 'Data flow (DFD)', w: 120, h: 120, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), ports: [...SIDES, [0.15, 0.15], [0.85, 0.15], [0.85, 0.85], [0.15, 0.85]] },
  dfdentity: { label: 'External entity', group: 'Data flow (DFD)', w: 130, h: 70, d: (w, h) => `M0 0H${w}V${h}H0Z`, ports: WITH_CORNERS },
  dfdstore: { label: 'Data store', group: 'Data flow (DFD)', w: 170, h: 50, d: (w, h) => `M0 0H${w}V${h}H0Z`, noBorder: true, more: (w, h) => `M${w} 0H0V${h}H${w}M${w * 0.2} 0V${h}`, ports: [[0.5, 0], [0.5, 1], [0, 0.5], [1, 0.5]] },

  entity: { label: 'Entity', group: 'Database (ERD)', w: 140, h: 70, d: (w, h) => `M0 0H${w}V${h}H0Z`, ports: WITH_CORNERS },
  weakentity: { label: 'Weak entity', group: 'Database (ERD)', w: 140, h: 70, d: (w, h) => `M0 0H${w}V${h}H0Z`, more: (w, h) => `M5 5H${w - 5}V${h - 5}H5Z`, ports: WITH_CORNERS },
  relationship: { label: 'Relationship', group: 'Database (ERD)', w: 130, h: 80, d: (w, h) => poly(w, h, [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5]]) },
  attribute: { label: 'Attribute', group: 'Database (ERD)', w: 120, h: 60, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2) },
  table: { label: 'Table', group: 'Database (ERD)', w: 160, h: 130, textTop: true, d: (w, h) => `M0 0H${w}V${h}H0Z`, more: (w) => `M0 28H${w}`, ports: WITH_CORNERS },

  umlclass: { label: 'Class', group: 'UML', w: 160, h: 120, textTop: true, d: (w, h) => `M0 0H${w}V${h}H0Z`, more: (w, h) => `M0 28H${w}M0 ${h * 0.62}H${w}`, ports: WITH_CORNERS },
  actor: { label: 'Actor', group: 'UML', w: 60, h: 110, textBelow: true, invisibleBox: true, d: (w, h) => `M0 0H${w}V${h}H0Z`, more: (w, h) => `${oval(w / 2, h * 0.14, h * 0.13, h * 0.13)}M${w / 2} ${h * 0.27}V${h * 0.62}M0 ${h * 0.4}H${w}M${w / 2} ${h * 0.62}L0 ${h}M${w / 2} ${h * 0.62}L${w} ${h}` },
  usecase: { label: 'Use case', group: 'UML', w: 150, h: 70, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2) },
  component: { label: 'Component', group: 'UML', w: 150, h: 80, d: (w, h) => `M0 0H${w}V${h}H0Z`, more: () => 'M-8 16H8V28H-8ZM-8 44H8V56H-8Z', ports: WITH_CORNERS },
  note: { label: 'Note', group: 'UML', w: 130, h: 80, d: (w, h) => `M0 0H${w - 18}L${w} 18V${h}H0Z`, more: (w) => `M${w - 18} 0V18H${w}` }
};
export const SHAPE_GROUPS = [...new Set(Object.values(SHAPES).map((s) => s.group))].map((group) => ({ group, shapes: Object.entries(SHAPES).filter(([, s]) => s.group === group).map(([key, s]) => ({ key, ...s })) }));
const specOf = (s) => SHAPES[s.t] || SHAPES.rect;

export const parseBoard = (body) => { try { const b = JSON.parse(body || '{}'); return { shapes: b.shapes || [], lines: b.lines || [], pen: b.pen || [] }; } catch (e) { return { shapes: [], lines: [], pen: [] }; } };
export const centre = (s) => ({ x: num(s.x) + num(s.w) / 2, y: num(s.y) + num(s.h) / 2 });

export function portsOf(s) {
  return (specOf(s).ports || SIDES).map(([fx, fy], i) => {
    const dx = fx - 0.5, dy = fy - 0.5;
    const dir = Math.abs(dx) >= Math.abs(dy) ? { x: Math.sign(dx) || 1, y: 0 } : { x: 0, y: Math.sign(dy) || 1 };
    return { i, x: num(s.x) + fx * num(s.w), y: num(s.y) + fy * num(s.h), dir };
  });
}

// where a connector's two ends are. An end with no chosen point takes the nearest one, and follows as shapes move.
function ends(l, byId) {
  const f = byId.get(l.from), t = byId.get(l.to);
  const free = (x, y) => [{ x: num(x), y: num(y), dir: { x: 0, y: 0 } }];
  const A = f ? portsOf(f) : free(l.x1, l.y1), Z = t ? portsOf(t) : free(l.x2, l.y2);
  const As = f && l.fromPort !== undefined && l.fromPort !== null && A[l.fromPort] ? [A[l.fromPort]] : A;
  const Zs = t && l.toPort !== undefined && l.toPort !== null && Z[l.toPort] ? [Z[l.toPort]] : Z;
  let best = null;
  As.forEach((a) => Zs.forEach((z) => { const d = (a.x - z.x) ** 2 + (a.y - z.y) ** 2; if (!best || d < best.d) best = { a, z, d }; }));
  return best;
}

function elbow(a, z, pts) {
  const S = { x: a.x + a.dir.x * 20, y: a.y + a.dir.y * 20 }, E = { x: z.x + z.dir.x * 20, y: z.y + z.dir.y * 20 };
  const out = [[a.x, a.y], [S.x, S.y]];
  let horiz = a.dir.x !== 0 || (a.dir.y === 0 && Math.abs(E.x - S.x) >= Math.abs(E.y - S.y));
  if (!pts.length && !!a.dir.x === !!z.dir.x && (a.dir.x || a.dir.y) && (z.dir.x || z.dir.y)) {
    if (a.dir.x) { const m = (S.x + E.x) / 2; out.push([m, S.y], [m, E.y]); } else { const m = (S.y + E.y) / 2; out.push([S.x, m], [E.x, m]); }
  } else {
    let u = S;
    [...pts.map(([x, y]) => ({ x, y })), E].forEach((v) => {
      if (u.x !== v.x && u.y !== v.y) { out.push(horiz ? [v.x, u.y] : [u.x, v.y]); } else horiz = !horiz;
      out.push([v.x, v.y]);
      u = v;
    });
  }
  out.push([E.x, E.y], [z.x, z.y]);
  return out.filter((p, i) => !i || p[0] !== out[i - 1][0] || p[1] !== out[i - 1][1]);
}

function glyph(p, ang, type) {
  const c = Math.cos(ang), s = Math.sin(ang);
  const P = (x, y) => (p.x + x * c - y * s).toFixed(1) + ' ' + (p.y + x * s + y * c).toFixed(1);
  const circle = (cx, r) => { const m = P(cx - r, 0), n = P(cx + r, 0); return `M${m}A${r} ${r} 0 1 0 ${n}A${r} ${r} 0 1 0 ${m}Z`; };
  switch (type) {
    case 'arrow': return { d: `M${P(0, 0)}L${P(-12, -6)}L${P(-12, 6)}Z`, fill: 'line' };
    case 'open': return { d: `M${P(-12, -6)}L${P(0, 0)}L${P(-12, 6)}`, fill: 'none' };
    case 'triangle': return { d: `M${P(0, 0)}L${P(-13, -7)}L${P(-13, 7)}Z`, fill: 'white' };
    case 'diamond': return { d: `M${P(0, 0)}L${P(-9, -5)}L${P(-18, 0)}L${P(-9, 5)}Z`, fill: 'line' };
    case 'odiamond': return { d: `M${P(0, 0)}L${P(-9, -5)}L${P(-18, 0)}L${P(-9, 5)}Z`, fill: 'white' };
    case 'circle': return { d: circle(-5, 5), fill: 'white' };
    case 'one': return { d: `M${P(-8, -7)}L${P(-8, 7)}M${P(-13, -7)}L${P(-13, 7)}`, fill: 'none' };
    case 'many': return { d: `M${P(-12, 0)}L${P(0, -7)}M${P(-12, 0)}L${P(0, 7)}M${P(-15, -7)}L${P(-15, 7)}`, fill: 'none' };
    case 'zeromany': return { d: `M${P(-12, 0)}L${P(0, -7)}M${P(-12, 0)}L${P(0, 7)}${circle(-17, 5)}`, fill: 'white' };
    default: return null;
  }
}

// everything needed to draw one connector
export function routeLine(l, byId) {
  const e = ends(l, byId);
  const { a, z } = e;
  const pts = (l.pts || []).map(([x, y]) => [num(x), num(y)]);
  const kind = l.kind || 'straight';
  const control = [[a.x, a.y], ...pts, [z.x, z.y]];
  let d, path;
  if (kind === 'elbow') { path = elbow(a, z, pts); d = 'M' + path.map((p) => p.join(' ')).join('L'); } else if (kind === 'curve') {
    path = control;
    if (!pts.length) {
      const k = Math.max(30, Math.hypot(a.x - z.x, a.y - z.y) / 2);
      const c1 = [a.x + (a.dir.x || 0) * k, a.y + (a.dir.y || 0) * k], c2 = [z.x + (z.dir.x || 0) * k, z.y + (z.dir.y || 0) * k];
      d = `M${a.x} ${a.y}C${c1.join(' ')} ${c2.join(' ')} ${z.x} ${z.y}`;
      path = [[a.x, a.y], c1, c2, [z.x, z.y]];
    } else {
      d = 'M' + control[0].join(' ');
      for (let i = 0; i < control.length - 1; i++) {
        const p0 = control[i - 1] || control[i], p1 = control[i], p2 = control[i + 1], p3 = control[i + 2] || p2;
        d += `C${p1[0] + (p2[0] - p0[0]) / 6} ${p1[1] + (p2[1] - p0[1]) / 6} ${p2[0] - (p3[0] - p1[0]) / 6} ${p2[1] - (p3[1] - p1[1]) / 6} ${p2[0]} ${p2[1]}`;
      }
    }
  } else { path = control; d = 'M' + control.map((p) => p.join(' ')).join('L'); }
  const angle = (from, to) => Math.atan2(to[1] - from[1], to[0] - from[0]);
  const n = path.length;
  const glyphs = [glyph(z, angle(path[n - 2], path[n - 1]), l.end === undefined ? 'arrow' : l.end), glyph(a, angle(path[1], path[0]), l.start || 'none')].filter(Boolean);
  const mi = Math.floor((n - 1) / 2);
  const mid = n % 2 ? { x: path[mi][0], y: path[mi][1] } : { x: (path[mi][0] + path[mi + 1][0]) / 2, y: (path[mi][1] + path[mi + 1][1]) / 2 };
  // the hollow points that make a new bend when dragged
  const adds = control.slice(0, -1).map((p, i) => ({ at: i, x: !pts.length && kind !== 'straight' ? mid.x : (p[0] + control[i + 1][0]) / 2, y: !pts.length && kind !== 'straight' ? mid.y : (p[1] + control[i + 1][1]) / 2 }));
  return { d, a, z, mid, glyphs, adds, pts, dash: l.dash === 'dashed' ? '8 6' : l.dash === 'dotted' ? '2 5' : '', color: col(l.color, '#555555'), width: Math.min(8, Math.max(1, num(l.width) || 2)), label: l.label || '' };
}

export function wrapText(text, w, size = 14) {
  const per = Math.max(3, Math.floor((w - 8) / (size * 0.55)));
  const out = [];
  String(text ?? '').split('\n').forEach((para) => {
    let line = '';
    para.split(' ').forEach((word) => {
      let rest = word;
      while (rest.length > per) { if (line) { out.push(line); line = ''; } out.push(rest.slice(0, per)); rest = rest.slice(per); }
      if ((line + ' ' + rest).trim().length > per) { out.push(line); line = rest; } else line = (line + ' ' + rest).trim();
    });
    out.push(line);
  });
  return out;
}

// how one shape draws: fill and line colours resolved, its text placed
export function drawShape(s) {
  const sp = specOf(s);
  const w = num(s.w), h = num(s.h);
  const size = Math.min(40, Math.max(9, num(s.size) || 14));
  const lines = wrapText(s.text, w, size);
  const y0 = sp.textTop ? 16 : sp.textBelow ? h + 14 : h / 2 - ((lines.length - 1) * size * 1.2) / 2;
  return {
    d: sp.d(w, h), more: sp.more ? sp.more(w, h) : '',
    fill: sp.noBox || sp.invisibleBox ? 'transparent' : col(s.color, sp.fill || '#DCEBFA'),
    stroke: sp.noBox || sp.invisibleBox || sp.noBorder ? 'transparent' : col(s.stroke, '#555555'),
    moreStroke: col(s.stroke, '#555555'), textColor: col(s.textColor, '#111111'), size,
    lines: lines.map((t, i) => ({ t, x: w / 2, y: y0 + i * size * 1.2 }))
  };
}

export function boardBox(b) {
  const xs = [], ys = [];
  b.shapes.forEach((s) => { xs.push(num(s.x), num(s.x) + num(s.w)); ys.push(num(s.y), num(s.y) + num(s.h) + (specOf(s).textBelow ? 24 : 0)); });
  b.pen.forEach((p) => (p.pts || []).forEach(([x, y]) => { xs.push(num(x)); ys.push(num(y)); }));
  b.lines.forEach((l) => { (l.pts || []).forEach(([x, y]) => { xs.push(num(x)); ys.push(num(y)); }); if (!l.from) { xs.push(num(l.x1)); ys.push(num(l.y1)); } if (!l.to) { xs.push(num(l.x2)); ys.push(num(l.y2)); } });
  if (!xs.length) return { x: 0, y: 0, w: 400, h: 200 };
  const x = Math.min(...xs) - 30, y = Math.min(...ys) - 30;
  return { x, y, w: Math.max(...xs) + 30 - x, h: Math.max(...ys) + 30 - y };
}

export function boardSvg(body) {
  const b = parseBoard(body);
  const box = boardBox(b);
  const byId = new Map(b.shapes.map((s) => [s.id, s]));
  const parts = [];
  b.pen.forEach((p) => parts.push('<polyline fill="none" stroke="' + col(p.color, '#333333') + '" stroke-width="2" points="' + (p.pts || []).map(([x, y]) => num(x) + ',' + num(y)).join(' ') + '"/>'));
  const lineParts = [];
  b.lines.forEach((l) => {
    if ((l.from && !byId.get(l.from)) || (l.to && !byId.get(l.to))) return;
    const r = routeLine(l, byId);
    lineParts.push('<path d="' + esc(r.d) + '" fill="none" stroke="' + r.color + '" stroke-width="' + r.width + '"' + (r.dash ? ' stroke-dasharray="' + r.dash + '"' : '') + '/>');
    r.glyphs.forEach((g) => lineParts.push('<path d="' + esc(g.d) + '" stroke="' + r.color + '" stroke-width="1.5" fill="' + (g.fill === 'line' ? r.color : g.fill === 'white' ? '#fff' : 'none') + '"/>'));
    if (r.label) lineParts.push('<text x="' + r.mid.x + '" y="' + r.mid.y + '" text-anchor="middle" dominant-baseline="middle" font-family="sans-serif" font-size="12" fill="#111" stroke="#fff" stroke-width="4" paint-order="stroke">' + esc(r.label) + '</text>');
  });
  b.shapes.forEach((s) => {
    const g = drawShape(s);
    parts.push('<g transform="translate(' + num(s.x) + ' ' + num(s.y) + ')"><path d="' + esc(g.d) + '" fill="' + g.fill + '" stroke="' + g.stroke + '"/>' + (g.more ? '<path d="' + esc(g.more) + '" fill="none" stroke="' + g.moreStroke + '"/>' : '') + g.lines.map((t) => '<text x="' + t.x + '" y="' + t.y + '" text-anchor="middle" dominant-baseline="middle" font-family="sans-serif" font-size="' + g.size + '" fill="' + g.textColor + '">' + esc(t.t) + '</text>').join('') + '</g>');
  });
  return '<svg xmlns="http://www.w3.org/2000/svg" viewBox="' + [box.x, box.y, box.w, box.h].join(' ') + '" width="' + Math.round(box.w) + '" height="' + Math.round(box.h) + '" style="max-width:100%;height:auto;background:#fff;border-radius:8px">' + parts.join('') + lineParts.join('') + '</svg>';
}
export const boardDataUrl = (body) => 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(boardSvg(body));
