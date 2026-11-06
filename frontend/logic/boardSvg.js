import { MORE_SHAPES } from './boardShapesMore.js';
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

const fx = (w, h, l) => poly(w, h, l);
const can = (w, h, r) => `M0 ${r}A${w / 2} ${r} 0 0 1 ${w} ${r}V${h - r}A${w / 2} ${r} 0 0 1 0 ${h - r}Z`;
const canTop = (w, r) => `M0 ${r}A${w / 2} ${r} 0 0 0 ${w} ${r}`;
const box = (w, h) => `M0 0H${w}V${h}H0Z`;
const below = (o) => ({ textBelow: true, ...o });
Object.assign(SHAPES, {
  // more basics
  rounded: { label: 'Rounded box', group: 'Basic', w: 150, h: 70, d: (w, h) => round(w, h, 16), ports: WITH_CORNERS },
  parallelogram: { label: 'Parallelogram', group: 'Basic', w: 150, h: 70, d: (w, h) => fx(w, h, [[0.2, 0], [1, 0], [0.8, 1], [0, 1]]) },
  trapezoid: { label: 'Trapezoid', group: 'Basic', w: 150, h: 70, d: (w, h) => fx(w, h, [[0.2, 0], [0.8, 0], [1, 1], [0, 1]]) },
  cube: { label: 'Cube', group: 'Basic', w: 110, h: 100, d: (w, h) => fx(w, h, [[0, 0.2], [0.2, 0], [1, 0], [1, 0.8], [0.8, 1], [0, 1]]), more: (w, h) => `M0 ${h * 0.2}H${w * 0.8}V${h}M${w * 0.8} ${h * 0.2}L${w} 0` },
  plus: { label: 'Plus', group: 'Basic', w: 90, h: 90, d: (w, h) => fx(w, h, [[0.35, 0], [0.65, 0], [0.65, 0.35], [1, 0.35], [1, 0.65], [0.65, 0.65], [0.65, 1], [0.35, 1], [0.35, 0.65], [0, 0.65], [0, 0.35], [0.35, 0.35]]) },
  cross: { label: 'Cross mark', group: 'Basic', w: 90, h: 90, d: (w, h) => fx(w, h, [[0.2, 0], [0.5, 0.3], [0.8, 0], [1, 0.2], [0.7, 0.5], [1, 0.8], [0.8, 1], [0.5, 0.7], [0.2, 1], [0, 0.8], [0.3, 0.5], [0, 0.2]]) },
  heart: { label: 'Heart', group: 'Basic', w: 100, h: 90, d: (w, h) => `M${w / 2} ${h}C${-w * 0.2} ${h * 0.5} ${w * 0.1} ${-h * 0.15} ${w / 2} ${h * 0.22}C${w * 0.9} ${-h * 0.15} ${w * 1.2} ${h * 0.5} ${w / 2} ${h}Z` },
  callout: { label: 'Speech bubble', group: 'Basic', w: 150, h: 90, d: (w, h) => `M6 0H${w - 6}Q${w} 0 ${w} 6V${h * 0.7}Q${w} ${h * 0.76} ${w - 6} ${h * 0.76}H${w * 0.4}L${w * 0.2} ${h}V${h * 0.76}H6Q0 ${h * 0.76} 0 ${h * 0.7}V6Q0 0 6 0Z` },
  ring: { label: 'Ring', group: 'Basic', w: 90, h: 90, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => oval(w / 2, h / 2, w * 0.3, h * 0.3) },

  // flowchart extras
  multidoc: { label: 'Many documents', group: 'Flowchart', w: 150, h: 90, d: (w, h) => `M${w * 0.1} 0H${w}V${h * 0.7}H${w * 0.9}V${h * 0.85}C${w * 0.65} ${h * 0.7} ${w * 0.3} ${h * 1.1} 0 ${h * 0.88}V${h * 0.25}H${w * 0.1}Z`, more: (w, h) => `M0 ${h * 0.25}H${w * 0.9}V${h * 0.7}` },
  stored: { label: 'Stored data', group: 'Flowchart', w: 140, h: 70, d: (w, h) => `M${w * 0.15} 0H${w}C${w * 0.82} ${h * 0.3} ${w * 0.82} ${h * 0.7} ${w} ${h}H${w * 0.15}C${-w * 0.05} ${h * 0.7} ${-w * 0.05} ${h * 0.3} ${w * 0.15} 0Z` },
  internal: { label: 'Internal storage', group: 'Flowchart', w: 130, h: 80, d: box, more: (w, h) => `M${w * 0.15} 0V${h}M0 ${h * 0.2}H${w}`, ports: WITH_CORNERS },
  display: { label: 'Display', group: 'Flowchart', w: 140, h: 70, d: (w, h) => `M${w * 0.15} 0H${w * 0.8}C${w * 1.07} ${h * 0.25} ${w * 1.07} ${h * 0.75} ${w * 0.8} ${h}H${w * 0.15}L0 ${h / 2}Z` },
  merge: { label: 'Merge', group: 'Flowchart', w: 100, h: 80, d: (w, h) => fx(w, h, [[0, 0], [1, 0], [0.5, 1]]), ports: [[0.5, 0], [0.75, 0.5], [0.5, 1], [0.25, 0.5]] },
  extract: { label: 'Extract', group: 'Flowchart', w: 100, h: 80, d: (w, h) => fx(w, h, [[0.5, 0], [1, 1], [0, 1]]), ports: [[0.5, 0], [0.75, 0.5], [0.5, 1], [0.25, 0.5]] },
  sort: { label: 'Sort', group: 'Flowchart', w: 90, h: 100, d: (w, h) => fx(w, h, [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5]]), more: (w, h) => `M0 ${h / 2}H${w}` },
  collate: { label: 'Collate', group: 'Flowchart', w: 90, h: 100, d: (w, h) => fx(w, h, [[0, 0], [1, 0], [0, 1], [1, 1]]), ports: [[0.5, 0], [0.5, 1], [0.5, 0.5]] },
  orjunction: { label: 'Or', group: 'Flowchart', w: 60, h: 60, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => `M${w / 2} 0V${h}M0 ${h / 2}H${w}` },
  sumjunction: { label: 'Summing', group: 'Flowchart', w: 60, h: 60, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => `M${w * 0.15} ${h * 0.15}L${w * 0.85} ${h * 0.85}M${w * 0.85} ${h * 0.15}L${w * 0.15} ${h * 0.85}` },
  loop: { label: 'Loop limit', group: 'Flowchart', w: 130, h: 70, d: (w, h) => fx(w, h, [[0.15, 0], [0.85, 0], [1, 0.3], [1, 1], [0, 1], [0, 0.3]]) },
  card: { label: 'Card', group: 'Flowchart', w: 130, h: 80, d: (w, h) => fx(w, h, [[0.15, 0], [1, 0], [1, 1], [0, 1], [0, 0.2]]) },
  tape: { label: 'Paper tape', group: 'Flowchart', w: 130, h: 80, d: (w, h) => `M0 ${h * 0.12}C${w * 0.3} ${-h * 0.1} ${w * 0.7} ${h * 0.34} ${w} ${h * 0.12}V${h * 0.88}C${w * 0.7} ${h * 1.1} ${w * 0.3} ${h * 0.66} 0 ${h * 0.88}Z` },

  // DFD extras
  gane: { label: 'Process (rounded)', group: 'Data flow (DFD)', w: 130, h: 110, d: (w, h) => round(w, h, 14), more: (w) => `M0 28H${w}`, textTop: false, ports: WITH_CORNERS },
  dfdstore2: { label: 'Data store (open)', group: 'Data flow (DFD)', w: 170, h: 50, d: box, noBorder: true, more: (w, h) => `M0 0H${w}M0 ${h}H${w}`, ports: [[0.5, 0], [0.5, 1], [0, 0.5], [1, 0.5]] },
  dfdboundary: { label: 'System boundary', group: 'Data flow (DFD)', w: 300, h: 220, noBox: true, textTop: true, d: box, more: (w, h) => `M0 0H${w}V${h}H0Z`, ports: WITH_CORNERS },

  // ERD extras
  keyattribute: { label: 'Key attribute', group: 'Database (ERD)', w: 120, h: 60, underline: true, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2) },
  multivalued: { label: 'Many-valued', group: 'Database (ERD)', w: 120, h: 60, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => oval(w / 2, h / 2, w / 2 - 5, h / 2 - 5) },
  weakrel: { label: 'Weak relationship', group: 'Database (ERD)', w: 140, h: 90, d: (w, h) => fx(w, h, [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5]]), more: (w, h) => fx(w, h, [[0.5, 0.12], [0.88, 0.5], [0.5, 0.88], [0.12, 0.5]]) },
  associative: { label: 'Associative entity', group: 'Database (ERD)', w: 150, h: 80, d: box, more: (w, h) => fx(w, h, [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5]]), ports: WITH_CORNERS },
  dbtable: { label: 'SQL table', group: 'Database (ERD)', w: 190, h: 130, textTop: true, d: box, more: (w) => `M0 28H${w}`, ports: WITH_CORNERS },

  // real database and storage pictures
  dbsingle: { label: 'Database', group: 'Databases & storage', w: 100, h: 110, ...below({}), d: (w, h) => can(w, h, h * 0.12), more: (w, h) => canTop(w, h * 0.12) },
  dbstack: { label: 'Database (disks)', group: 'Databases & storage', w: 110, h: 110, ...below({}), d: (w, h) => can(w, h, h * 0.11), more: (w, h) => { const r = h * 0.11; return `${canTop(w, r)}M0 ${h * 0.38}A${w / 2} ${r} 0 0 0 ${w} ${h * 0.38}M0 ${h * 0.64}A${w / 2} ${r} 0 0 0 ${w} ${h * 0.64}`; } },
  dbcluster: { label: 'Database cluster', group: 'Databases & storage', w: 130, h: 100, ...below({}), d: (w, h) => `M${w * 0.3} ${h * 0.1}H${w}V${h * 0.8}H${w * 0.7}V${h}H0V${h * 0.2}H${w * 0.3}Z`, more: (w, h) => `${canTop(w * 0.65, h * 0.1).replace('M0', 'M0')}M0 ${h * 0.45}H${w * 0.7}M${w * 0.7} ${h * 0.2}V${h * 0.8}M0 ${h * 0.2}H${w * 0.7}` },
  cylinderrepo: { label: 'Repository', group: 'Databases & storage', w: 100, h: 90, ...below({}), d: (w, h) => can(w, h, h * 0.13), more: (w, h) => `${canTop(w, h * 0.13)}M${w * 0.3} ${h * 0.45}L${w * 0.15} ${h * 0.6}L${w * 0.3} ${h * 0.75}M${w * 0.7} ${h * 0.45}L${w * 0.85} ${h * 0.6}L${w * 0.7} ${h * 0.75}` },
  cache: { label: 'Cache', group: 'Databases & storage', w: 110, h: 70, ...below({}), d: (w, h) => round(w, h, 6), more: (w, h) => `M${w * 0.55} ${h * 0.12}L${w * 0.38} ${h * 0.55}H${w * 0.52}L${w * 0.45} ${h * 0.88}L${w * 0.65} ${h * 0.42}H${w * 0.5}Z` },
  storage: { label: 'File storage', group: 'Databases & storage', w: 110, h: 70, ...below({}), d: box, more: (w, h) => `M0 ${h / 2}H${w}M${w * 0.8} ${h * 0.25}h8M${w * 0.8} ${h * 0.75}h8` , ports: WITH_CORNERS },
  bucket: { label: 'Object bucket', group: 'Databases & storage', w: 90, h: 100, ...below({}), d: (w, h) => `M0 ${h * 0.12}L${w * 0.14} ${h}H${w * 0.86}L${w} ${h * 0.12}Z`, more: (w, h) => `M0 ${h * 0.12}A${w / 2} ${h * 0.12} 0 0 1 ${w} ${h * 0.12}A${w / 2} ${h * 0.12} 0 0 1 0 ${h * 0.12}` },
  datalake: { label: 'Data lake', group: 'Databases & storage', w: 120, h: 100, ...below({}), d: (w, h) => can(w, h, h * 0.13), more: (w, h) => `${canTop(w, h * 0.13)}M${w * 0.12} ${h * 0.55}q${w * 0.1} -${h * 0.1} ${w * 0.2} 0t${w * 0.2} 0t${w * 0.2} 0M${w * 0.12} ${h * 0.75}q${w * 0.1} -${h * 0.1} ${w * 0.2} 0t${w * 0.2} 0t${w * 0.2} 0` },
  warehouse: { label: 'Data warehouse', group: 'Databases & storage', w: 120, h: 100, ...below({}), d: (w, h) => fx(w, h, [[0, 0.35], [0.5, 0], [1, 0.35], [1, 1], [0, 1]]), more: (w, h) => `M${w * 0.2} ${h}V${h * 0.55}H${w * 0.8}V${h}M${w * 0.2} ${h * 0.7}H${w * 0.8}M${w * 0.2} ${h * 0.85}H${w * 0.8}` },
  queue: { label: 'Message queue', group: 'Databases & storage', w: 140, h: 60, ...below({}), d: (w, h) => `M${w * 0.08} 0H${w * 0.92}A${w * 0.08} ${h / 2} 0 0 1 ${w * 0.92} ${h}H${w * 0.08}A${w * 0.08} ${h / 2} 0 0 1 ${w * 0.08} 0Z`, more: (w, h) => `M${w * 0.92} 0A${w * 0.08} ${h / 2} 0 0 0 ${w * 0.92} ${h}M${w * 0.6} 0V${h}M${w * 0.4} 0V${h}M${w * 0.2} 0V${h}` },
  backup: { label: 'Backup tape', group: 'Databases & storage', w: 100, h: 80, ...below({}), d: (w, h) => round(w, h, 4), more: (w, h) => `${oval(w * 0.3, h * 0.45, h * 0.18, h * 0.18)}${oval(w * 0.7, h * 0.45, h * 0.18, h * 0.18)}M${w * 0.3} ${h * 0.63}H${w * 0.7}` },

  // UML extras
  lifeline: { label: 'Lifeline', group: 'UML', w: 130, h: 400, textHead: true, invisibleBox: true, d: box, more: (w, h) => `M0 0H${w}V50H0ZM${w / 2} 50V${h}`, ports: [[0.5, 0], [0.5, 1]] },
  state: { label: 'State', group: 'UML', w: 140, h: 70, d: (w, h) => round(w, h, 18), ports: WITH_CORNERS },
  initial: { label: 'Start point', group: 'UML', w: 30, h: 30, fill: '#333333', d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2) },
  final: { label: 'End point', group: 'UML', w: 34, h: 34, d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => `${oval(w / 2, h / 2, w * 0.28, h * 0.28)}` },
  forkbar: { label: 'Fork / join', group: 'UML', w: 160, h: 10, fill: '#333333', d: box, ports: [[0.5, 0], [0.5, 1], [0.2, 0], [0.8, 0], [0.2, 1], [0.8, 1]] },
  package: { label: 'Package', group: 'UML', w: 160, h: 110, d: (w, h) => `M0 0H${w * 0.45}V20H${w}V${h}H0Z`, more: () => 'M0 20H72', ports: WITH_CORNERS },
  iface: { label: 'Interface', group: 'UML', w: 44, h: 44, ...below({}), d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2) },
  activation: { label: 'Activation', group: 'UML', w: 16, h: 90, d: box, ports: WITH_CORNERS },

  // BPMN
  bstart: { label: 'Start event', group: 'Business process (BPMN)', w: 50, h: 50, ...below({}), fill: '#CDEBC0', d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2) },
  bend: { label: 'End event', group: 'Business process (BPMN)', w: 50, h: 50, ...below({}), fill: '#F7CFC8', d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => oval(w / 2, h / 2, w / 2 - 3, h / 2 - 3) },
  btimer: { label: 'Timer event', group: 'Business process (BPMN)', w: 50, h: 50, ...below({}), d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => `${oval(w / 2, h / 2, w / 2 - 5, h / 2 - 5)}M${w / 2} ${h * 0.25}V${h / 2}L${w * 0.68} ${h * 0.62}` },
  bmessage: { label: 'Message event', group: 'Business process (BPMN)', w: 50, h: 50, ...below({}), d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => `M${w * 0.25} ${h * 0.33}H${w * 0.75}V${h * 0.67}H${w * 0.25}ZL${w / 2} ${h * 0.52}L${w * 0.75} ${h * 0.33}` },
  btask: { label: 'Task', group: 'Business process (BPMN)', w: 140, h: 70, d: (w, h) => round(w, h, 10), ports: WITH_CORNERS },
  bxor: { label: 'Either-or gateway', group: 'Business process (BPMN)', w: 60, h: 60, ...below({}), d: (w, h) => fx(w, h, [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5]]), more: (w, h) => `M${w * 0.35} ${h * 0.35}L${w * 0.65} ${h * 0.65}M${w * 0.65} ${h * 0.35}L${w * 0.35} ${h * 0.65}` },
  band: { label: 'Parallel gateway', group: 'Business process (BPMN)', w: 60, h: 60, ...below({}), d: (w, h) => fx(w, h, [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5]]), more: (w, h) => `M${w / 2} ${h * 0.28}V${h * 0.72}M${w * 0.28} ${h / 2}H${w * 0.72}` },
  bor: { label: 'Inclusive gateway', group: 'Business process (BPMN)', w: 60, h: 60, ...below({}), d: (w, h) => fx(w, h, [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5]]), more: (w, h) => oval(w / 2, h / 2, w * 0.2, h * 0.2) },
  bdata: { label: 'Data object', group: 'Business process (BPMN)', w: 60, h: 80, ...below({}), d: (w, h) => `M0 0H${w - 16}L${w} 16V${h}H0Z`, more: (w) => `M${w - 16} 0V16H${w}` },

  // network and cloud
  server: { label: 'Server', group: 'Network & cloud', w: 90, h: 90, ...below({}), d: (w, h) => round(w, h, 4), more: (w, h) => `M0 ${h / 3}H${w}M0 ${(h * 2) / 3}H${w}${oval(w * 0.82, h / 6, 3, 3)}${oval(w * 0.82, h / 2, 3, 3)}${oval(w * 0.82, (h * 5) / 6, 3, 3)}M${w * 0.12} ${h / 6}H${w * 0.55}M${w * 0.12} ${h / 2}H${w * 0.55}M${w * 0.12} ${(h * 5) / 6}H${w * 0.55}`, ports: WITH_CORNERS },
  firewall: { label: 'Firewall', group: 'Network & cloud', w: 70, h: 80, ...below({}), fill: '#F7CFC8', d: box, more: (w, h) => `M0 ${h * 0.25}H${w}M0 ${h * 0.5}H${w}M0 ${h * 0.75}H${w}M${w / 2} 0V${h * 0.25}M${w * 0.25} ${h * 0.25}V${h * 0.5}M${w * 0.75} ${h * 0.25}V${h * 0.5}M${w / 2} ${h * 0.5}V${h * 0.75}M${w * 0.25} ${h * 0.75}V${h}M${w * 0.75} ${h * 0.75}V${h}`, ports: WITH_CORNERS },
  router: { label: 'Router', group: 'Network & cloud', w: 80, h: 60, ...below({}), d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => `M${w * 0.2} ${h * 0.5}H${w * 0.8}M${w / 2} ${h * 0.15}V${h * 0.85}M${w * 0.7} ${h * 0.38}L${w * 0.8} ${h * 0.5}L${w * 0.7} ${h * 0.62}M${w * 0.3} ${h * 0.38}L${w * 0.2} ${h * 0.5}L${w * 0.3} ${h * 0.62}` },
  switch: { label: 'Network switch', group: 'Network & cloud', w: 140, h: 40, ...below({}), d: (w, h) => round(w, h, 4), more: (w, h) => [0.15, 0.3, 0.45, 0.6, 0.75].map((f) => `M${w * f} ${h * 0.35}h10v${h * 0.3}h-10Z`).join(''), ports: [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5], [0.15, 1], [0.85, 1]] },
  loadbalancer: { label: 'Load balancer', group: 'Network & cloud', w: 90, h: 70, ...below({}), d: (w, h) => round(w, h, 6), more: (w, h) => `M${w * 0.15} ${h / 2}H${w * 0.45}M${w * 0.45} ${h / 2}L${w * 0.8} ${h * 0.2}M${w * 0.45} ${h / 2}H${w * 0.8}M${w * 0.45} ${h / 2}L${w * 0.8} ${h * 0.8}` },
  gateway: { label: 'API gateway', group: 'Network & cloud', w: 90, h: 80, ...below({}), d: (w, h) => fx(w, h, [[0.2, 0], [0.8, 0], [1, 0.5], [0.8, 1], [0.2, 1], [0, 0.5]]), more: (w, h) => `M${w * 0.32} ${h * 0.3}L${w * 0.2} ${h * 0.5}L${w * 0.32} ${h * 0.7}M${w * 0.68} ${h * 0.3}L${w * 0.8} ${h * 0.5}L${w * 0.68} ${h * 0.7}` },
  container: { label: 'Container', group: 'Network & cloud', w: 110, h: 90, ...below({}), d: (w, h) => fx(w, h, [[0.5, 0], [1, 0.25], [1, 0.75], [0.5, 1], [0, 0.75], [0, 0.25]]), more: (w, h) => `M0 ${h * 0.25}L${w / 2} ${h / 2}L${w} ${h * 0.25}M${w / 2} ${h / 2}V${h}` },
  func: { label: 'Function', group: 'Network & cloud', w: 80, h: 80, ...below({}), d: box, more: (w, h) => `M${w * 0.62} ${h * 0.2}Q${w * 0.45} ${h * 0.15} ${w * 0.45} ${h * 0.4}V${h * 0.65}Q${w * 0.45} ${h * 0.85} ${w * 0.3} ${h * 0.8}M${w * 0.3} ${h * 0.45}H${w * 0.6}`, ports: WITH_CORNERS },
  globe: { label: 'Internet', group: 'Network & cloud', w: 80, h: 80, ...below({}), d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => `${oval(w / 2, h / 2, w * 0.22, h / 2)}M0 ${h / 2}H${w}M${w * 0.07} ${h * 0.25}H${w * 0.93}M${w * 0.07} ${h * 0.75}H${w * 0.93}` },
  antenna: { label: 'Wi-Fi', group: 'Network & cloud', w: 70, h: 70, ...below({}), invisibleBox: true, d: box, more: (w, h) => `M${w * 0.1} ${h * 0.4}Q${w / 2} 0 ${w * 0.9} ${h * 0.4}M${w * 0.25} ${h * 0.58}Q${w / 2} ${h * 0.3} ${w * 0.75} ${h * 0.58}M${w * 0.38} ${h * 0.75}Q${w / 2} ${h * 0.6} ${w * 0.62} ${h * 0.75}${oval(w / 2, h * 0.88, 4, 4)}` },

  // computers and people
  desktop: { label: 'Desktop', group: 'Devices & people', w: 90, h: 80, ...below({}), d: (w, h) => round(w, h * 0.72, 4), more: (w, h) => `M${w / 2} ${h * 0.72}V${h * 0.92}M${w * 0.28} ${h * 0.95}H${w * 0.72}`, ports: [[0.5, 0], [1, 0.36], [0.5, 1], [0, 0.36]] },
  laptop: { label: 'Laptop', group: 'Devices & people', w: 100, h: 70, ...below({}), d: (w, h) => fx(w, h, [[0.12, 0], [0.88, 0], [0.88, 0.75], [1, 1], [0, 1], [0.12, 0.75]]), more: (w, h) => `M${w * 0.12} ${h * 0.75}H${w * 0.88}` },
  phone: { label: 'Phone', group: 'Devices & people', w: 50, h: 90, ...below({}), d: (w, h) => round(w, h, 7), more: (w, h) => `M0 ${h * 0.12}H${w}M0 ${h * 0.85}H${w}` },
  tablet: { label: 'Tablet', group: 'Devices & people', w: 80, h: 100, ...below({}), d: (w, h) => round(w, h, 7), more: (w, h) => `M${w * 0.1} ${h * 0.08}H${w * 0.9}V${h * 0.88}H${w * 0.1}Z` },
  printer: { label: 'Printer', group: 'Devices & people', w: 90, h: 70, ...below({}), d: (w, h) => fx(w, h, [[0, 0.3], [0.2, 0.3], [0.2, 0], [0.8, 0], [0.8, 0.3], [1, 0.3], [1, 0.8], [0.8, 0.8], [0.8, 1], [0.2, 1], [0.2, 0.8], [0, 0.8]]), more: (w, h) => `M${w * 0.2} ${h * 0.3}H${w * 0.8}M${w * 0.2} ${h * 0.65}H${w * 0.8}` },
  browser: { label: 'Web browser', group: 'Devices & people', w: 120, h: 90, ...below({}), d: (w, h) => round(w, h, 5), more: (w) => `M0 18H${w}${oval(10, 9, 3, 3)}${oval(20, 9, 3, 3)}${oval(30, 9, 3, 3)}`, ports: WITH_CORNERS },
  user: { label: 'Person', group: 'Devices & people', w: 60, h: 80, ...below({}), d: (w, h) => `${oval(w / 2, h * 0.25, w * 0.25, h * 0.22)}M0 ${h}C0 ${h * 0.5} ${w} ${h * 0.5} ${w} ${h}Z` },
  users: { label: 'People', group: 'Devices & people', w: 80, h: 70, ...below({}), d: (w, h) => `${oval(w * 0.32, h * 0.25, w * 0.17, h * 0.2)}${oval(w * 0.7, h * 0.3, w * 0.14, h * 0.17)}M0 ${h}C0 ${h * 0.45} ${w * 0.64} ${h * 0.45} ${w * 0.64} ${h}ZM${w * 0.68} ${h}H${w}C${w} ${h * 0.55} ${w * 0.65} ${h * 0.5} ${w * 0.55} ${h * 0.58}Q${w * 0.68} ${h * 0.75} ${w * 0.68} ${h}Z` },

  // containers and symbols
  lane: { label: 'Swimlane', group: 'Containers', w: 820, h: 130, textSide: true, d: box, more: (w, h) => `M40 0V${h}`, ports: WITH_CORNERS },
  pool: { label: 'Group with title', group: 'Containers', w: 300, h: 200, textTop: true, d: (w, h) => round(w, h, 6), more: (w) => `M0 28H${w}`, ports: WITH_CORNERS },
  frame: { label: 'Frame', group: 'Containers', w: 300, h: 200, noBox: true, textTop: true, d: box, more: (w, h) => `M0 0H${w}V${h}H0Z`, ports: WITH_CORNERS },
  brace: { label: 'Brace', group: 'Containers', w: 30, h: 120, noBox: true, d: box, more: (w, h) => `M${w} 0Q${w / 2} 0 ${w / 2} ${h * 0.2}V${h * 0.4}Q${w / 2} ${h / 2} 0 ${h / 2}Q${w / 2} ${h / 2} ${w / 2} ${h * 0.6}V${h * 0.8}Q${w / 2} ${h} ${w} ${h}` },

  arrowl: { label: 'Arrow left', group: 'Arrows', w: 140, h: 70, d: (w, h) => fx(w, h, [[1, 0.25], [0.35, 0.25], [0.35, 0], [0, 0.5], [0.35, 1], [0.35, 0.75], [1, 0.75]]) },
  arrowu: { label: 'Arrow up', group: 'Arrows', w: 70, h: 120, d: (w, h) => fx(w, h, [[0.25, 1], [0.25, 0.35], [0, 0.35], [0.5, 0], [1, 0.35], [0.75, 0.35], [0.75, 1]]) },
  arrowd: { label: 'Arrow down', group: 'Arrows', w: 70, h: 120, d: (w, h) => fx(w, h, [[0.25, 0], [0.25, 0.65], [0, 0.65], [0.5, 1], [1, 0.65], [0.75, 0.65], [0.75, 0]]) },
  arrowboth: { label: 'Two-way arrow', group: 'Arrows', w: 150, h: 60, d: (w, h) => fx(w, h, [[0, 0.5], [0.2, 0], [0.2, 0.25], [0.8, 0.25], [0.8, 0], [1, 0.5], [0.8, 1], [0.8, 0.75], [0.2, 0.75], [0.2, 1]]) },
  chevron: { label: 'Chevron', group: 'Arrows', w: 130, h: 60, d: (w, h) => fx(w, h, [[0, 0], [0.75, 0], [1, 0.5], [0.75, 1], [0, 1], [0.25, 0.5]]) },
  pentarrow: { label: 'Step arrow', group: 'Arrows', w: 140, h: 60, d: (w, h) => fx(w, h, [[0, 0], [0.8, 0], [1, 0.5], [0.8, 1], [0, 1]]) },

  gear: { label: 'Settings', group: 'Symbols', w: 80, h: 80, ...below({}), d: (w, h) => { const p = []; for (let i = 0; i < 16; i++) { const r = i % 2 ? 0.5 : 0.38, a = (i * Math.PI) / 8; p.push([0.5 + r * Math.cos(a), 0.5 + r * Math.sin(a)]); } return poly(w, h, p); }, more: (w, h) => oval(w / 2, h / 2, w * 0.15, h * 0.15) },
  lock: { label: 'Lock', group: 'Symbols', w: 60, h: 80, ...below({}), d: (w, h) => `M0 ${h * 0.42}H${w}V${h}H0Z`, more: (w, h) => `M${w * 0.2} ${h * 0.42}V${h * 0.25}A${w * 0.3} ${h * 0.22} 0 0 1 ${w * 0.8} ${h * 0.25}V${h * 0.42}` },
  key: { label: 'Key', group: 'Symbols', w: 90, h: 50, ...below({}), d: (w, h) => `${oval(h / 2, h / 2, h / 2, h / 2)}M${h} ${h * 0.4}H${w}V${h * 0.75}H${w * 0.85}V${h * 0.6}H${h}Z` },
  mail: { label: 'Email', group: 'Symbols', w: 90, h: 60, ...below({}), d: box, more: (w, h) => `M0 0L${w / 2} ${h * 0.6}L${w} 0`, ports: WITH_CORNERS },
  folder: { label: 'Folder', group: 'Symbols', w: 90, h: 70, ...below({}), d: (w, h) => `M0 0H${w * 0.4}L${w * 0.5} ${h * 0.18}H${w}V${h}H0Z` },
  file: { label: 'File', group: 'Symbols', w: 65, h: 85, ...below({}), d: (w, h) => `M0 0H${w - 18}L${w} 18V${h}H0Z`, more: (w, h) => `M${w - 18} 0V18H${w}M10 ${h * 0.45}H${w - 10}M10 ${h * 0.62}H${w - 10}M10 ${h * 0.79}H${w - 10}` },
  chart: { label: 'Chart', group: 'Symbols', w: 90, h: 80, ...below({}), d: box, more: (w, h) => `M${w * 0.2} ${h * 0.85}V${h * 0.55}M${w * 0.4} ${h * 0.85}V${h * 0.3}M${w * 0.6} ${h * 0.85}V${h * 0.45}M${w * 0.8} ${h * 0.85}V${h * 0.15}`, ports: WITH_CORNERS },
  check: { label: 'Tick', group: 'Symbols', w: 80, h: 80, ...below({}), fill: '#CDEBC0', d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => `M${w * 0.25} ${h * 0.52}L${w * 0.43} ${h * 0.7}L${w * 0.75} ${h * 0.32}` },
  warning: { label: 'Warning', group: 'Symbols', w: 80, h: 70, ...below({}), fill: '#FFE9A8', d: (w, h) => fx(w, h, [[0.5, 0], [1, 1], [0, 1]]), more: (w, h) => `M${w / 2} ${h * 0.35}V${h * 0.68}M${w / 2} ${h * 0.78}V${h * 0.85}`, ports: [[0.5, 0], [0.75, 0.5], [0.5, 1], [0.25, 0.5]] },
  clock: { label: 'Clock', group: 'Symbols', w: 70, h: 70, ...below({}), d: (w, h) => oval(w / 2, h / 2, w / 2, h / 2), more: (w, h) => `M${w / 2} ${h * 0.2}V${h / 2}L${w * 0.7} ${h * 0.62}` },
  bulb: { label: 'Idea', group: 'Symbols', w: 60, h: 85, ...below({}), d: (w, h) => `M${w * 0.32} ${h * 0.72}C${w * 0.28} ${h * 0.55} 0 ${h * 0.5} 0 ${h * 0.3}A${w / 2} ${h * 0.3} 0 0 1 ${w} ${h * 0.3}C${w} ${h * 0.5} ${w * 0.72} ${h * 0.55} ${w * 0.68} ${h * 0.72}Z`, more: (w, h) => `M${w * 0.34} ${h * 0.82}H${w * 0.66}M${w * 0.4} ${h * 0.92}H${w * 0.6}` }
});
Object.assign(SHAPES, MORE_SHAPES);
export const SHAPE_GROUPS = [...new Set(Object.values(SHAPES).map((s) => s.group))].map((group) => ({ group, shapes: Object.entries(SHAPES).filter(([, s]) => s.group === group).map(([key, s]) => ({ key, ...s })) }));
const specOf = (s) => SHAPES[s.t] || SHAPES.rect;
export const nameOfShape = (t) => (SHAPES[t] || SHAPES.rect).label;

export const parseBoard = (body) => { try { const b = JSON.parse(body || '{}'); return { bg: b.bg || 'white', shapes: b.shapes || [], lines: b.lines || [], pen: b.pen || [] }; } catch (e) { return { bg: 'white', shapes: [], lines: [], pen: [] }; } };

// the sheet a diagram is drawn on
export const BOARD_BACKGROUNDS = [{ v: 'white', l: 'Whiteboard' }, { v: 'graph', l: 'Graph paper' }, { v: 'dots', l: 'Dotted' }, { v: 'lined', l: 'Lined' }, { v: 'black', l: 'Blackboard' }];
export const boardPattern = (bg, id) => ({
  graph: `<pattern id="${id}" width="50" height="50" patternUnits="userSpaceOnUse"><path d="M10 0V50M20 0V50M30 0V50M40 0V50M0 10H50M0 20H50M0 30H50M0 40H50" fill="none" stroke="#DCE8F1" stroke-width="0.6"/><path d="M50 0H0V50" fill="none" stroke="#A9C6DC" stroke-width="1"/></pattern>`,
  dots: `<pattern id="${id}" width="20" height="20" patternUnits="userSpaceOnUse"><circle cx="1" cy="1" r="1.1" fill="#B4BEB9"/></pattern>`,
  lined: `<pattern id="${id}" width="20" height="28" patternUnits="userSpaceOnUse"><path d="M0 27.5H20" stroke="#CFE0EE"/></pattern>`
}[bg] || '');
export const boardInk = (bg) => (bg === 'black' ? '#E8EFEA' : '#555555');
export const boardPaper = (bg) => (bg === 'black' ? '#1B2420' : '#FFFFFF');
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
export function routeLine(l, byId, ink = '#555555') {
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
  return { d, a, z, mid, glyphs, adds, pts, dash: l.dash === 'dashed' ? '8 6' : l.dash === 'dotted' ? '2 5' : '', color: col(l.color, ink), width: Math.min(8, Math.max(1, num(l.width) || 2)), label: l.label || '' };
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
export function drawShape(s, bg) {
  const sp = specOf(s);
  const w = num(s.w), h = num(s.h);
  const size = Math.min(40, Math.max(9, num(s.size) || 14));
  const lines = wrapText(s.text, sp.textBelow ? Math.max(w, 140) : w, size);   // a name below a small picture may be wider than it
  const y0 = sp.textTop ? 16 : sp.textHead ? 25 : sp.textBelow ? h + 14 : h / 2 - ((lines.length - 1) * size * 1.2) / 2;
  return {
    d: sp.d(w, h), more: sp.more ? sp.more(w, h) : '',
    fill: sp.noBox || sp.invisibleBox ? 'transparent' : col(s.color, sp.fill || '#DCEBFA'),
    stroke: sp.noBox || sp.invisibleBox || sp.noBorder ? 'transparent' : col(s.stroke, '#555555'),
    moreStroke: col(s.stroke, boardInk(bg)), textColor: col(s.textColor, bg === 'black' && (sp.noBox || sp.invisibleBox || sp.noBorder) ? '#E8EFEA' : '#111111'), size,
    underline: !!sp.underline, side: !!sp.textSide,
    lines: lines.map((t, i) => ({ t, x: sp.textSide ? 20 : w / 2, y: y0 + i * size * 1.2 }))
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
    const r = routeLine(l, byId, boardInk(b.bg));
    lineParts.push('<path d="' + esc(r.d) + '" fill="none" stroke="' + r.color + '" stroke-width="' + r.width + '"' + (r.dash ? ' stroke-dasharray="' + r.dash + '"' : '') + '/>');
    r.glyphs.forEach((g) => lineParts.push('<path d="' + esc(g.d) + '" stroke="' + r.color + '" stroke-width="1.5" fill="' + (g.fill === 'line' ? r.color : g.fill === 'white' ? boardPaper(b.bg) : 'none') + '"/>'));
    if (r.label) lineParts.push('<text x="' + r.mid.x + '" y="' + r.mid.y + '" text-anchor="middle" dominant-baseline="middle" font-family="sans-serif" font-size="12" fill="' + (b.bg === 'black' ? '#E8EFEA' : '#111') + '" stroke="' + boardPaper(b.bg) + '" stroke-width="4" paint-order="stroke">' + esc(r.label) + '</text>');
  });
  b.shapes.forEach((s) => {
    const g = drawShape(s, b.bg);
    parts.push('<g transform="translate(' + num(s.x) + ' ' + num(s.y) + ')"><path d="' + esc(g.d) + '" fill="' + g.fill + '" stroke="' + g.stroke + '"/>' + (g.more ? '<path d="' + esc(g.more) + '" fill="none" stroke="' + g.moreStroke + '"/>' : '') + g.lines.map((t) => '<text x="' + t.x + '" y="' + t.y + '" text-anchor="middle" dominant-baseline="middle" font-family="sans-serif" font-size="' + g.size + '" fill="' + g.textColor + '">' + esc(t.t) + '</text>').join('') + '</g>');
  });
  return '<svg xmlns="http://www.w3.org/2000/svg" viewBox="' + [box.x, box.y, box.w, box.h].join(' ') + '" width="' + Math.round(box.w) + '" height="' + Math.round(box.h) + '" style="max-width:100%;height:auto;background:' + boardPaper(b.bg) + ';border-radius:8px"><defs>' + boardPattern(b.bg, 'bgp') + '</defs>' + (boardPattern(b.bg, 'bgp') ? '<rect x="' + box.x + '" y="' + box.y + '" width="' + box.w + '" height="' + box.h + '" fill="url(#bgp)"/>' : '<rect x="' + box.x + '" y="' + box.y + '" width="' + box.w + '" height="' + box.h + '" fill="' + boardPaper(b.bg) + '"/>') + parts.join('') + lineParts.join('') + '</svg>';
}
export const boardDataUrl = (body) => 'data:image/svg+xml;charset=utf-8,' + encodeURIComponent(boardSvg(body));
