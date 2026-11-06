// More of the diagram library: networking, servers, cloud, security, software, UI wireframes, charts, electrical, floor plans and more.
const poly = (w, h, l) => 'M' + l.map(([x, y]) => (x * w).toFixed(1) + ' ' + (y * h).toFixed(1)).join('L') + 'Z';
const oval = (cx, cy, rx, ry) => `M${cx - rx} ${cy}a${rx} ${ry} 0 1 0 ${rx * 2} 0a${rx} ${ry} 0 1 0 ${-rx * 2} 0Z`;
const round = (w, h, r) => `M${r} 0H${w - r}Q${w} 0 ${w} ${r}V${h - r}Q${w} ${h} ${w - r} ${h}H${r}Q0 ${h} 0 ${h - r}V${r}Q0 0 ${r} 0Z`;
const box = (w, h) => `M0 0H${w}V${h}H0Z`;
const circ = (w, h) => oval(w / 2, h / 2, w / 2, h / 2);
const corners = [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5], [0, 0], [1, 0], [1, 1], [0, 1]];
// a shape drawn as an outline with a picture on it, its name written below
const pic = (label, group, w, h, d, more, o = {}) => ({ label, group, w, h, d, more, textBelow: true, ...o });
// a shape with no outline of its own: the drawing is the shape
const line = (label, group, w, h, more, o = {}) => ({ label, group, w, h, d: box, more, invisibleBox: true, textBelow: true, ...o });
const lights = (w, h, ys) => ys.map((f) => `M${w * 0.1} ${h * f}H${w * 0.55}${oval(w * 0.8, h * f, 3, 3)}`).join('');
const dividers = (w, h, n) => Array.from({ length: n - 1 }, (_, i) => `M0 ${(h * (i + 1)) / n}H${w}`).join('');
const server = (extra) => (w, h) => dividers(w, h, 2) + lights(w, h, [0.25, 0.75]) + extra(w, h);
const shield = (w, h) => `M${w / 2} 0L${w} ${h * 0.15}V${h * 0.5}C${w} ${h * 0.78} ${w * 0.7} ${h * 0.92} ${w / 2} ${h}C${w * 0.3} ${h * 0.92} 0 ${h * 0.78} 0 ${h * 0.5}V${h * 0.15}Z`;
const cloudD = (w, h) => `M${w * 0.25} ${h * 0.85}C${w * 0.02} ${h * 0.85} ${-w * 0.02} ${h * 0.45} ${w * 0.2} ${h * 0.4}C${w * 0.2} ${h * 0.08} ${w * 0.5} ${-h * 0.05} ${w * 0.6} ${h * 0.22}C${w * 0.78} ${h * 0.1} ${w * 0.92} ${h * 0.3} ${w * 0.85} ${h * 0.45}C${w * 1.04} ${h * 0.5} ${w} ${h * 0.87} ${w * 0.78} ${h * 0.85}Z`;
const hex = (w, h) => poly(w, h, [[0.5, 0], [1, 0.25], [1, 0.75], [0.5, 1], [0, 0.75], [0, 0.25]]);

const NET = 'Networking', SRV = 'Servers & data centre', CLD = 'Cloud & software', SEC = 'Security', UI = 'Screens & wireframes', CHT = 'Charts', ELE = 'Electrical', FLR = 'Floor plan', BAS = 'Basic', MTH = 'Maths & logic', ORG = 'Business';

export const MORE_SHAPES = {
  // networking
  modem: pic('Modem', NET, 100, 50, (w, h) => round(w, h, 6), (w, h) => `${[0.2, 0.35, 0.5].map((f) => oval(w * f, h * 0.6, 3, 3)).join('')}M${w * 0.8} 0V${-h * 0.4}`),
  accesspoint: pic('Access point', NET, 90, 60, (w, h) => `M0 ${h * 0.5}Q${w / 2} ${h * 0.2} ${w} ${h * 0.5}V${h * 0.8}Q${w / 2} ${h * 1.1} 0 ${h * 0.8}Z`, (w, h) => `M${w * 0.3} ${h * 0.12}Q${w / 2} ${-h * 0.15} ${w * 0.7} ${h * 0.12}M${w * 0.4} ${h * 0.25}Q${w / 2} ${h * 0.12} ${w * 0.6} ${h * 0.25}`),
  nethub: pic('Network hub', NET, 120, 40, (w, h) => round(w, h, 4), (w, h) => `M${w * 0.15} ${h}V${h * 1.3}M${w * 0.5} ${h}V${h * 1.3}M${w * 0.85} ${h}V${h * 1.3}${oval(w / 2, h / 2, 4, 4)}`, { ports: [[0.5, 0], [1, 0.5], [0.5, 1], [0, 0.5], [0.15, 1], [0.85, 1]] }),
  bridge: pic('Network bridge', NET, 100, 50, box, (w, h) => `M${w * 0.2} ${h * 0.7}Q${w / 2} ${h * 0.05} ${w * 0.8} ${h * 0.7}M${w * 0.1} ${h * 0.7}H${w * 0.9}`, { ports: corners }),
  l3switch: pic('Layer 3 switch', NET, 100, 60, box, (w, h) => `M${w * 0.15} ${h * 0.35}H${w * 0.85}M${w * 0.7} ${h * 0.2}L${w * 0.85} ${h * 0.35}L${w * 0.7} ${h * 0.5}M${w * 0.85} ${h * 0.65}H${w * 0.15}M${w * 0.3} ${h * 0.5}L${w * 0.15} ${h * 0.65}L${w * 0.3} ${h * 0.8}`, { ports: corners }),
  wirelessrouter: pic('Wi-Fi router', NET, 90, 70, (w, h) => `M0 ${h * 0.55}H${w}V${h}H0Z`, (w, h) => `M${w * 0.2} ${h * 0.55}V${h * 0.1}M${w * 0.8} ${h * 0.55}V${h * 0.1}${oval(w * 0.2, h * 0.78, 3, 3)}${oval(w * 0.4, h * 0.78, 3, 3)}`),
  vpn: pic('VPN gateway', NET, 90, 80, shield, (w, h) => `M${w * 0.28} ${h * 0.35}L${w / 2} ${h * 0.65}L${w * 0.72} ${h * 0.35}`),
  proxy: pic('Proxy server', NET, 100, 70, (w, h) => round(w, h, 4), (w, h) => `M0 ${h / 2}H${w * 0.35}M${w * 0.65} ${h / 2}H${w}${oval(w / 2, h / 2, w * 0.15, h * 0.22)}`),
  dns: pic('DNS', NET, 90, 80, (w, h) => round(w, h, 4), (w, h) => `${oval(w * 0.3, h * 0.3, 6, 6)}${oval(w * 0.7, h * 0.3, 6, 6)}${oval(w / 2, h * 0.72, 6, 6)}M${w * 0.34} ${h * 0.37}L${w * 0.47} ${h * 0.65}M${w * 0.66} ${h * 0.37}L${w * 0.53} ${h * 0.65}`),
  cdn: pic('CDN', NET, 110, 80, circ, (w, h) => `${oval(w / 2, h / 2, 7, 7)}${[[0.2, 0.25], [0.8, 0.25], [0.2, 0.75], [0.8, 0.75]].map(([x, y]) => `M${w / 2} ${h / 2}L${w * x} ${h * y}${oval(w * x, h * y, 4, 4)}`).join('')}`),
  wan: pic('WAN', NET, 140, 80, cloudD, (w, h) => `M${w * 0.32} ${h * 0.55}H${w * 0.68}`),
  lan: line('LAN bus', NET, 240, 50, (w, h) => `M0 ${h * 0.25}H${w}M${w * 0.2} ${h * 0.25}V${h}M${w * 0.5} ${h * 0.25}V${h}M${w * 0.8} ${h * 0.25}V${h}`, { ports: [[0, 0.25], [1, 0.25], [0.2, 1], [0.5, 1], [0.8, 1]] }),
  subnet: { label: 'Subnet', group: NET, w: 260, h: 170, noBox: true, textTop: true, d: box, more: (w, h) => `M0 0H${w}V${h}H0Z`, ports: corners },
  vlan: pic('VLAN', NET, 120, 50, (w, h) => round(w, h, 25), (w, h) => `M${w * 0.2} ${h / 2}H${w * 0.8}M${w * 0.3} ${h * 0.3}V${h * 0.7}M${w * 0.5} ${h * 0.3}V${h * 0.7}M${w * 0.7} ${h * 0.3}V${h * 0.7}`),
  tower: line('Cell tower', NET, 70, 100, (w, h) => `M${w / 2} ${h * 0.15}L${w * 0.2} ${h}M${w / 2} ${h * 0.15}L${w * 0.8} ${h}M${w * 0.32} ${h * 0.65}H${w * 0.68}M${w * 0.26} ${h * 0.82}H${w * 0.74}${oval(w / 2, h * 0.15, 4, 4)}M${w * 0.28} 0Q${w * 0.15} ${h * 0.15} ${w * 0.28} ${h * 0.3}M${w * 0.72} 0Q${w * 0.85} ${h * 0.15} ${w * 0.72} ${h * 0.3}`),
  satellite: line('Satellite', NET, 90, 80, (w, h) => `M${w * 0.35} ${h * 0.35}H${w * 0.65}V${h * 0.65}H${w * 0.35}ZM${w * 0.35} ${h / 2}H0V${h * 0.25}M0 ${h * 0.75}V${h / 2}M${w * 0.65} ${h / 2}H${w}V${h * 0.25}M${w} ${h * 0.75}V${h / 2}M${w * 0.15} ${h * 0.3}V${h * 0.7}M${w * 0.85} ${h * 0.3}V${h * 0.7}`),
  dish: line('Satellite dish', NET, 80, 80, (w, h) => `M${w * 0.1} ${h * 0.15}Q${w * 0.15} ${h * 0.7} ${w * 0.75} ${h * 0.7}ZM${w * 0.42} ${h * 0.42}L${w * 0.8} ${h * 0.1}M${w * 0.45} ${h * 0.7}L${w * 0.3} ${h}M${w * 0.45} ${h * 0.7}L${w * 0.65} ${h}`),
  fiber: line('Fibre link', NET, 140, 30, (w, h) => `M0 ${h / 2}H${w}M0 ${h * 0.2}H${w}M0 ${h * 0.8}H${w}`, { ports: [[0, 0.5], [1, 0.5]] }),
  ipphone: pic('IP phone', NET, 80, 70, (w, h) => round(w, h, 5), (w, h) => `M${w * 0.14} ${h * 0.15}V${h * 0.85}M${w * 0.3} ${h * 0.2}H${w * 0.85}V${h * 0.4}H${w * 0.3}Z${[0.55, 0.72].map((y) => [0.4, 0.58, 0.76].map((x) => oval(w * x, h * y, 2, 2)).join('')).join('')}`),
  cctv: line('CCTV camera', NET, 90, 60, (w, h) => `M0 ${h * 0.1}H${w * 0.75}V${h * 0.55}H0ZM${w * 0.75} ${h * 0.22}L${w} ${h * 0.1}V${h * 0.55}L${w * 0.75} ${h * 0.43}M${w * 0.3} ${h * 0.55}V${h}M${w * 0.15} ${h}H${w * 0.45}`),
  iot: pic('IoT sensor', NET, 60, 60, circ, (w, h) => `${oval(w / 2, h / 2, 5, 5)}M${w * 0.28} ${h * 0.28}Q${w * 0.1} ${h / 2} ${w * 0.28} ${h * 0.72}M${w * 0.72} ${h * 0.28}Q${w * 0.9} ${h / 2} ${w * 0.72} ${h * 0.72}`),
  scanner: pic('Scanner', NET, 100, 50, (w, h) => round(w, h, 4), (w, h) => `M0 ${h * 0.35}H${w}M${w * 0.1} ${h * 0.68}H${w * 0.6}`),

  // servers and data centre
  webserver: pic('Web server', SRV, 90, 90, (w, h) => round(w, h, 4), server((w, h) => `${oval(w * 0.8, h * 0.5, 0, 0)}`)),
  appserver: pic('App server', SRV, 90, 90, (w, h) => round(w, h, 4), server((w, h) => `M${w * 0.1} ${h * 0.92}H${w * 0.4}`)),
  dbserver: pic('Database server', SRV, 90, 100, (w, h) => round(w, h, 4), (w, h) => `${dividers(w, h * 0.5, 2)}${lights(w, h * 0.5, [0.25, 0.75])}M${w * 0.25} ${h * 0.62}A${w * 0.25} ${h * 0.06} 0 0 1 ${w * 0.75} ${h * 0.62}V${h * 0.88}A${w * 0.25} ${h * 0.06} 0 0 1 ${w * 0.25} ${h * 0.88}ZM${w * 0.25} ${h * 0.62}A${w * 0.25} ${h * 0.06} 0 0 0 ${w * 0.75} ${h * 0.62}M0 ${h * 0.5}H${w}`),
  mailserver: pic('Mail server', SRV, 90, 90, (w, h) => round(w, h, 4), (w, h) => `${dividers(w, h, 2)}${lights(w, h, [0.25])}M${w * 0.2} ${h * 0.62}H${w * 0.8}V${h * 0.9}H${w * 0.2}ZL${w / 2} ${h * 0.78}L${w * 0.8} ${h * 0.62}`),
  fileserver: pic('File server', SRV, 90, 90, (w, h) => round(w, h, 4), (w, h) => `${dividers(w, h, 2)}${lights(w, h, [0.25])}M${w * 0.25} ${h * 0.9}V${h * 0.62}H${w * 0.45}L${w * 0.52} ${h * 0.68}H${w * 0.75}V${h * 0.9}Z`),
  rack: pic('Server rack', SRV, 80, 130, box, (w, h) => `${dividers(w, h, 6)}${lights(w, h, [1 / 12, 3 / 12, 5 / 12, 7 / 12, 9 / 12, 11 / 12])}`, { ports: corners }),
  bladeserver: pic('Blade server', SRV, 110, 80, box, (w, h) => `${[0.2, 0.4, 0.6, 0.8].map((f) => `M${w * f} 0V${h}`).join('')}${[0.1, 0.3, 0.5, 0.7, 0.9].map((f) => oval(w * f, h * 0.15, 2, 2)).join('')}`, { ports: corners }),
  mainframe: pic('Mainframe', SRV, 100, 120, box, (w, h) => `M${w / 2} 0V${h}${[0.25, 0.75].map((x) => [0.12, 0.25, 0.38].map((y) => oval(w * x, h * y, 3, 3)).join('')).join('')}M${w * 0.12} ${h * 0.6}H${w * 0.38}M${w * 0.62} ${h * 0.6}H${w * 0.88}M${w * 0.12} ${h * 0.75}H${w * 0.38}M${w * 0.62} ${h * 0.75}H${w * 0.88}`, { ports: corners }),
  nas: pic('NAS', SRV, 90, 80, (w, h) => round(w, h, 4), (w, h) => `${dividers(w, h, 3)}${[1 / 6, 0.5, 5 / 6].map((f) => oval(w * 0.8, h * f, 3, 3)).join('')}`),
  san: pic('SAN', SRV, 110, 80, box, (w, h) => `${dividers(w, h, 4)}${[0.125, 0.375, 0.625, 0.875].map((f) => [0.15, 0.3, 0.45, 0.6].map((x) => `M${w * x} ${h * f - 3}h6v6h-6Z`).join('')).join('')}`, { ports: corners }),
  harddisk: pic('Hard disk', SRV, 80, 90, (w, h) => round(w, h, 4), (w, h) => `${oval(w / 2, h * 0.42, w * 0.3, w * 0.3)}${oval(w / 2, h * 0.42, 3, 3)}M${w * 0.15} ${h * 0.88}H${w * 0.5}`),
  ups: pic('UPS / battery', SRV, 70, 100, box, (w, h) => `M${w * 0.58} ${h * 0.15}L${w * 0.35} ${h * 0.52}H${w * 0.52}L${w * 0.42} ${h * 0.85}L${w * 0.68} ${h * 0.44}H${w * 0.52}Z`, { ports: corners }),
  kvm: pic('Console', SRV, 90, 70, (w, h) => round(w, h, 4), (w, h) => `M${w * 0.15} ${h * 0.3}L${w * 0.3} ${h * 0.45}L${w * 0.15} ${h * 0.6}M${w * 0.38} ${h * 0.65}H${w * 0.65}`),
  thinclient: pic('Thin client', SRV, 60, 90, (w, h) => round(w, h, 4), (w, h) => `${oval(w / 2, h * 0.18, 4, 4)}M${w * 0.2} ${h * 0.4}H${w * 0.8}M${w * 0.2} ${h * 0.55}H${w * 0.8}`),
  datacentre: pic('Data centre', SRV, 120, 90, (w, h) => poly(w, h, [[0, 0.2], [1, 0.2], [1, 1], [0, 1]]), (w, h) => `M0 ${h * 0.2}L${w * 0.15} 0H${w * 0.85}L${w} ${h * 0.2}${[0.2, 0.45, 0.7].map((x) => `M${w * x} ${h * 0.35}h${w * 0.12}v${h * 0.5}h-${w * 0.12}Z`).join('')}`),

  // cloud and software
  vm: pic('Virtual machine', CLD, 100, 80, box, (w, h) => `M${w * 0.15} ${h * 0.18}H${w * 0.85}V${h * 0.82}H${w * 0.15}Z`, { dashMore: true, ports: corners }),
  hypervisor: pic('Hypervisor', CLD, 120, 80, box, (w, h) => `M0 ${h * 0.65}H${w}M${w * 0.1} ${h * 0.12}h${w * 0.35}v${h * 0.42}h-${w * 0.35}ZM${w * 0.55} ${h * 0.12}h${w * 0.35}v${h * 0.42}h-${w * 0.35}Z`, { ports: corners }),
  pod: pic('Pod', CLD, 80, 90, hex, (w, h) => `${oval(w / 2, h / 2, w * 0.15, h * 0.15)}`),
  k8s: pic('Cluster', CLD, 100, 100, (w, h) => poly(w, h, [[0.5, 0], [0.9, 0.2], [1, 0.62], [0.72, 1], [0.28, 1], [0, 0.62], [0.1, 0.2]]), (w, h) => `${oval(w / 2, h / 2, w * 0.1, h * 0.1)}${[[0.5, 0.15], [0.8, 0.33], [0.85, 0.65], [0.65, 0.85], [0.35, 0.85], [0.15, 0.65], [0.2, 0.33]].map(([x, y]) => `M${w / 2} ${h / 2}L${w * x} ${h * y}`).join('')}`),
  microservice: pic('Microservice', CLD, 90, 80, hex, (w, h) => `M${w * 0.32} ${h * 0.35}L${w * 0.2} ${h / 2}L${w * 0.32} ${h * 0.65}M${w * 0.68} ${h * 0.35}L${w * 0.8} ${h / 2}L${w * 0.68} ${h * 0.65}`),
  api: pic('API', CLD, 90, 60, (w, h) => round(w, h, 8), (w, h) => `M${w * 0.3} ${h * 0.25}L${w * 0.15} ${h / 2}L${w * 0.3} ${h * 0.75}M${w * 0.7} ${h * 0.25}L${w * 0.85} ${h / 2}L${w * 0.7} ${h * 0.75}M${w * 0.57} ${h * 0.2}L${w * 0.43} ${h * 0.8}`),
  webhook: line('Webhook', CLD, 70, 70, (w, h) => `${oval(w / 2, h * 0.2, 9, 9)}${oval(w * 0.15, h * 0.82, 9, 9)}${oval(w * 0.85, h * 0.82, 9, 9)}M${w / 2} ${h * 0.33}L${w * 0.25} ${h * 0.72}M${w * 0.28} ${h * 0.82}H${w * 0.72}M${w * 0.75} ${h * 0.72}L${w / 2} ${h * 0.33}`),
  broker: pic('Message broker', CLD, 110, 70, (w, h) => round(w, h, 6), (w, h) => `M0 ${h * 0.5}H${w * 0.3}M${w * 0.7} ${h * 0.25}H${w}M${w * 0.7} ${h * 0.75}H${w}M${w * 0.3} ${h * 0.3}H${w * 0.7}V${h * 0.7}H${w * 0.3}Z`),
  topic: pic('Event topic', CLD, 110, 50, (w, h) => round(w, h, 25), (w, h) => `M${w * 0.25} ${h * 0.3}V${h * 0.7}M${w * 0.45} ${h * 0.3}V${h * 0.7}M${w * 0.65} ${h * 0.3}V${h * 0.7}M${w * 0.72} ${h * 0.5}H${w * 0.88}`),
  cron: pic('Scheduled job', CLD, 80, 80, circ, (w, h) => `M${w / 2} ${h * 0.18}V${h / 2}H${w * 0.72}M${w * 0.14} ${h * 0.14}L${w * 0.05} ${h * 0.05}`),
  monitor: pic('Monitoring', CLD, 100, 80, (w, h) => round(w, h, 4), (w, h) => `M${w * 0.1} ${h * 0.55}H${w * 0.3}L${w * 0.4} ${h * 0.25}L${w * 0.55} ${h * 0.78}L${w * 0.65} ${h * 0.55}H${w * 0.9}`),
  logs: pic('Logs', CLD, 80, 90, box, (w, h) => `${[0.2, 0.4, 0.6, 0.8].map((f, i) => `M${w * 0.15} ${h * f}H${w * (0.85 - i * 0.12)}`).join('')}`, { ports: corners }),
  region: { label: 'Cloud region', group: CLD, w: 320, h: 200, noBox: true, textTop: true, d: box, more: (w, h) => `M0 0H${w}V${h}H0Z`, ports: corners },
  code: pic('Source code', CLD, 90, 80, box, (w, h) => `M0 ${h * 0.18}H${w}M${w * 0.32} ${h * 0.38}L${w * 0.18} ${h * 0.6}L${w * 0.32} ${h * 0.82}M${w * 0.68} ${h * 0.38}L${w * 0.82} ${h * 0.6}L${w * 0.68} ${h * 0.82}`, { ports: corners }),
  terminal: pic('Terminal', CLD, 100, 70, (w, h) => round(w, h, 5), (w, h) => `M${w * 0.15} ${h * 0.3}L${w * 0.3} ${h * 0.48}L${w * 0.15} ${h * 0.66}M${w * 0.38} ${h * 0.7}H${w * 0.62}`),
  bug: pic('Bug', CLD, 70, 80, (w, h) => oval(w / 2, h * 0.55, w * 0.3, h * 0.38), (w, h) => `M${w / 2} ${h * 0.2}V${h * 0.93}M${w * 0.2} ${h * 0.4}L0 ${h * 0.28}M${w * 0.8} ${h * 0.4}L${w} ${h * 0.28}M${w * 0.2} ${h * 0.58}H0M${w * 0.8} ${h * 0.58}H${w}M${w * 0.23} ${h * 0.75}L0 ${h * 0.88}M${w * 0.77} ${h * 0.75}L${w} ${h * 0.88}`),
  c4person: pic('C4 person', CLD, 110, 110, (w, h) => `M${w * 0.1} ${h * 0.45}Q${w * 0.1} ${h * 0.4} ${w * 0.3} ${h * 0.4}H${w * 0.7}Q${w * 0.9} ${h * 0.4} ${w * 0.9} ${h * 0.45}V${h * 0.9}Q${w * 0.9} ${h} ${w * 0.7} ${h}H${w * 0.3}Q${w * 0.1} ${h} ${w * 0.1} ${h * 0.9}Z${oval(w / 2, h * 0.2, h * 0.19, h * 0.19)}`, null),
  c4system: { label: 'C4 system', group: CLD, w: 170, h: 100, fill: '#BBD6F2', d: (w, h) => round(w, h, 8), ports: corners },

  // security
  shield: pic('Shield', SEC, 80, 95, shield, (w, h) => `M${w * 0.28} ${h * 0.48}L${w * 0.45} ${h * 0.65}L${w * 0.74} ${h * 0.32}`),
  ids: pic('Intrusion detection', SEC, 90, 90, shield, (w, h) => `${oval(w / 2, h * 0.45, w * 0.2, h * 0.14)}${oval(w / 2, h * 0.45, 3, 3)}`),
  vault: pic('Secrets vault', SEC, 90, 90, (w, h) => round(w, h, 6), (w, h) => `${oval(w / 2, h / 2, w * 0.25, h * 0.25)}M${w / 2} ${h * 0.25}V${h * 0.75}M${w * 0.25} ${h / 2}H${w * 0.75}`),
  certificate: pic('Certificate', SEC, 100, 70, box, (w, h) => `M${w * 0.12} ${h * 0.25}H${w * 0.6}M${w * 0.12} ${h * 0.45}H${w * 0.5}${oval(w * 0.76, h * 0.62, 11, 11)}M${w * 0.7} ${h * 0.8}V${h}L${w * 0.76} ${h * 0.92}L${w * 0.82} ${h}V${h * 0.8}`, { ports: corners }),
  identity: pic('Identity', SEC, 90, 70, (w, h) => round(w, h, 5), (w, h) => `${oval(w * 0.28, h * 0.4, 9, 9)}M${w * 0.12} ${h * 0.8}Q${w * 0.28} ${h * 0.5} ${w * 0.44} ${h * 0.8}M${w * 0.55} ${h * 0.35}H${w * 0.88}M${w * 0.55} ${h * 0.55}H${w * 0.88}`),
  fingerprint: line('Fingerprint', SEC, 70, 85, (w, h) => `M${w * 0.1} ${h * 0.75}V${h * 0.42}A${w * 0.4} ${h * 0.36} 0 0 1 ${w * 0.9} ${h * 0.42}V${h * 0.75}M${w * 0.28} ${h * 0.9}V${h * 0.45}A${w * 0.22} ${h * 0.2} 0 0 1 ${w * 0.72} ${h * 0.45}V${h * 0.9}M${w / 2} ${h}V${h * 0.45}`),
  threat: pic('Threat', SEC, 80, 80, (w, h) => poly(w, h, [[0.3, 0], [0.7, 0], [1, 0.3], [1, 0.7], [0.7, 1], [0.3, 1], [0, 0.7], [0, 0.3]]), (w, h) => `M${w / 2} ${h * 0.2}V${h * 0.6}M${w / 2} ${h * 0.72}V${h * 0.8}`, { fill: '#F7CFC8' }),

  // screens and wireframes
  uiwindow: { label: 'Window', group: UI, w: 260, h: 180, textTop: false, d: (w, h) => round(w, h, 5), more: (w) => `M0 24H${w}${oval(w - 14, 12, 4, 4)}${oval(w - 28, 12, 4, 4)}`, ports: corners },
  uimobile: { label: 'Mobile screen', group: UI, w: 130, h: 240, d: (w, h) => round(w, h, 14), more: (w, h) => `M0 22H${w}M0 ${h - 26}H${w}${oval(w / 2, h - 13, 6, 6)}`, ports: corners },
  uibutton: { label: 'Button', group: UI, w: 110, h: 38, fill: '#BBD6F2', d: (w, h) => round(w, h, 8), ports: corners },
  uiinput: { label: 'Text field', group: UI, w: 170, h: 36, fill: '#FFFFFF', d: (w, h) => round(w, h, 4), ports: corners },
  uidropdown: { label: 'Dropdown', group: UI, w: 170, h: 36, fill: '#FFFFFF', d: (w, h) => round(w, h, 4), more: (w, h) => `M${w - 24} ${h * 0.4}L${w - 16} ${h * 0.62}L${w - 8} ${h * 0.4}`, ports: corners },
  uicheckbox: pic('Tick box', UI, 26, 26, (w, h) => round(w, h, 3), (w, h) => `M${w * 0.22} ${h / 2}L${w * 0.42} ${h * 0.72}L${w * 0.78} ${h * 0.28}`, { fill: '#FFFFFF' }),
  uiradio: pic('Radio button', UI, 26, 26, circ, (w, h) => oval(w / 2, h / 2, w * 0.2, h * 0.2), { fill: '#FFFFFF' }),
  uitoggle: pic('Switch', UI, 56, 28, (w, h) => round(w, h, h / 2), (w, h) => oval(w - h / 2, h / 2, h * 0.36, h * 0.36)),
  uislider: line('Slider', UI, 170, 26, (w, h) => `M0 ${h / 2}H${w}${oval(w * 0.6, h / 2, 8, 8)}`),
  uinav: { label: 'Top bar', group: UI, w: 300, h: 44, d: box, more: (w, h) => `M${w - 90} ${h / 2}h20M${w - 60} ${h / 2}h20M${w - 30} ${h / 2}h20`, ports: corners },
  uisidebar: { label: 'Side menu', group: UI, w: 90, h: 220, textTop: true, d: box, more: (w, h) => [0.2, 0.32, 0.44, 0.56].map((f) => `M12 ${h * f}H${w - 12}`).join(''), ports: corners },
  uiimage: pic('Picture', UI, 120, 90, box, (w, h) => `M0 0L${w} ${h}M${w} 0L0 ${h}`, { ports: corners }),
  uicard: { label: 'Card', group: UI, w: 180, h: 130, d: (w, h) => round(w, h, 8), more: (w) => `M0 44H${w}`, ports: corners },
  uitable: { label: 'Data table', group: UI, w: 240, h: 120, textTop: true, d: box, more: (w, h) => `${dividers(w, h, 4)}M${w / 3} 0V${h}M${(w * 2) / 3} 0V${h}`, ports: corners },
  uitabs: { label: 'Tabs', group: UI, w: 240, h: 120, d: (w, h) => `M0 26H${w}V${h}H0Z`, more: () => 'M0 26V4H70V26M70 4H140V26', ports: corners },
  uisearch: pic('Search box', UI, 170, 36, (w, h) => round(w, h, h / 2), (w, h) => `${oval(w - 24, h * 0.45, 7, 7)}M${w - 19} ${h * 0.62}L${w - 12} ${h * 0.82}`, { fill: '#FFFFFF' }),
  uimodal: { label: 'Dialog', group: UI, w: 220, h: 140, d: (w, h) => round(w, h, 8), more: (w, h) => `M${w - 20} 10l10 10m0 -10l-10 10M${w - 80} ${h - 14}h60`, ports: corners },
  uiavatar: pic('Avatar', UI, 50, 50, circ, (w, h) => `${oval(w / 2, h * 0.38, w * 0.16, h * 0.16)}M${w * 0.22} ${h * 0.82}Q${w / 2} ${h * 0.45} ${w * 0.78} ${h * 0.82}`),

  // charts
  chartbar: pic('Bar chart', CHT, 100, 80, box, (w, h) => [[0.15, 0.5], [0.35, 0.3], [0.55, 0.6], [0.75, 0.15]].map(([x, y]) => `M${w * x} ${h * 0.9}V${h * y}h${w * 0.12}V${h * 0.9}`).join(''), { ports: corners }),
  chartline: pic('Line chart', CHT, 100, 80, box, (w, h) => `M${w * 0.1} ${h * 0.1}V${h * 0.9}H${w * 0.92}M${w * 0.12} ${h * 0.7}L${w * 0.35} ${h * 0.45}L${w * 0.55} ${h * 0.6}L${w * 0.85} ${h * 0.2}`, { ports: corners }),
  chartpie: pic('Pie chart', CHT, 90, 90, circ, (w, h) => `M${w / 2} ${h / 2}V0M${w / 2} ${h / 2}L${w} ${h * 0.65}M${w / 2} ${h / 2}L${w * 0.12} ${h * 0.82}`),
  chartdonut: pic('Doughnut chart', CHT, 90, 90, circ, (w, h) => `${oval(w / 2, h / 2, w * 0.25, h * 0.25)}M${w / 2} 0V${h * 0.25}M${w} ${h / 2}H${w * 0.75}`),
  chartarea: pic('Area chart', CHT, 100, 80, box, (w, h) => `M${w * 0.08} ${h * 0.92}V${h * 0.6}L${w * 0.3} ${h * 0.35}L${w * 0.55} ${h * 0.5}L${w * 0.92} ${h * 0.15}V${h * 0.92}Z`, { ports: corners }),
  chartscatter: pic('Scatter plot', CHT, 100, 80, box, (w, h) => `M${w * 0.1} ${h * 0.1}V${h * 0.9}H${w * 0.92}${[[0.25, 0.7], [0.35, 0.5], [0.5, 0.6], [0.6, 0.35], [0.8, 0.25]].map(([x, y]) => oval(w * x, h * y, 3, 3)).join('')}`, { ports: corners }),
  funnel: { label: 'Funnel', group: CHT, w: 130, h: 120, d: (w, h) => poly(w, h, [[0, 0], [1, 0], [0.62, 0.65], [0.62, 1], [0.38, 1], [0.38, 0.65]]), more: (w, h) => `M${w * 0.2} ${h * 0.33}H${w * 0.8}` },
  pyramid: { label: 'Pyramid', group: CHT, w: 140, h: 120, d: (w, h) => poly(w, h, [[0.5, 0], [1, 1], [0, 1]]), more: (w, h) => `M${w * 0.33} ${h / 3}H${w * 0.67}M${w * 0.17} ${(h * 2) / 3}H${w * 0.83}`, ports: [[0.5, 0], [0.75, 0.5], [0.5, 1], [0.25, 0.5]] },
  gauge: pic('Gauge', CHT, 100, 60, (w, h) => `M0 ${h}A${w / 2} ${h} 0 0 1 ${w} ${h}Z`, (w, h) => `M${w / 2} ${h}L${w * 0.72} ${h * 0.3}`),
  venn: { label: 'Venn diagram', group: CHT, w: 150, h: 100, invisibleBox: true, textBelow: true, d: box, more: (w, h) => `${oval(w * 0.36, h / 2, w * 0.36, h / 2)}${oval(w * 0.64, h / 2, w * 0.36, h / 2)}` },
  timelinebar: line('Timeline', CHT, 220, 40, (w, h) => `M0 ${h / 2}H${w}${[0.1, 0.35, 0.6, 0.85].map((f) => oval(w * f, h / 2, 6, 6)).join('')}`, { ports: [[0, 0.5], [1, 0.5], [0.1, 0.5], [0.35, 0.5], [0.6, 0.5], [0.85, 0.5]] }),

  // electrical
  resistor: line('Resistor', ELE, 110, 30, (w, h) => `M0 ${h / 2}H${w * 0.2}L${w * 0.27} 0L${w * 0.4} ${h}L${w * 0.53} 0L${w * 0.66} ${h}L${w * 0.75} ${h / 2}H${w}`, { ports: [[0, 0.5], [1, 0.5]] }),
  capacitor: line('Capacitor', ELE, 90, 40, (w, h) => `M0 ${h / 2}H${w * 0.43}M${w * 0.43} 0V${h}M${w * 0.57} 0V${h}M${w * 0.57} ${h / 2}H${w}`, { ports: [[0, 0.5], [1, 0.5]] }),
  inductor: line('Coil', ELE, 110, 30, (w, h) => `M0 ${h * 0.8}H${w * 0.15}${[0, 1, 2, 3].map(() => `a${w * 0.0875} ${h * 0.6} 0 0 1 ${w * 0.175} 0`).join('')}H${w}`, { ports: [[0, 0.8], [1, 0.8]] }),
  diode: line('Diode', ELE, 90, 40, (w, h) => `M0 ${h / 2}H${w * 0.35}L${w * 0.35} 0L${w * 0.65} ${h / 2}L${w * 0.35} ${h}V${h / 2}M${w * 0.65} 0V${h}M${w * 0.65} ${h / 2}H${w}`, { ports: [[0, 0.5], [1, 0.5]] }),
  led: line('LED', ELE, 90, 50, (w, h) => `M0 ${h * 0.6}H${w * 0.35}M${w * 0.35} ${h * 0.3}V${h * 0.9}L${w * 0.65} ${h * 0.6}ZM${w * 0.65} ${h * 0.3}V${h * 0.9}M${w * 0.65} ${h * 0.6}H${w}M${w * 0.55} ${h * 0.2}L${w * 0.7} 0M${w * 0.68} ${h * 0.3}L${w * 0.83} ${h * 0.1}`, { ports: [[0, 0.6], [1, 0.6]] }),
  battery: line('Battery', ELE, 90, 40, (w, h) => `M0 ${h / 2}H${w * 0.42}M${w * 0.42} 0V${h}M${w * 0.52} ${h * 0.25}V${h * 0.75}M${w * 0.52} ${h / 2}H${w}`, { ports: [[0, 0.5], [1, 0.5]] }),
  ground: line('Ground', ELE, 50, 40, (w, h) => `M${w / 2} 0V${h * 0.45}M0 ${h * 0.45}H${w}M${w * 0.2} ${h * 0.72}H${w * 0.8}M${w * 0.38} ${h}H${w * 0.62}`, { ports: [[0.5, 0]] }),
  elswitch: line('Switch', ELE, 90, 30, (w, h) => `M0 ${h * 0.8}H${w * 0.3}L${w * 0.68} 0M${w * 0.7} ${h * 0.8}H${w}`, { ports: [[0, 0.8], [1, 0.8]] }),
  acsource: pic('AC source', ELE, 60, 60, circ, (w, h) => `M${w * 0.22} ${h / 2}Q${w * 0.36} ${h * 0.15} ${w / 2} ${h / 2}T${w * 0.78} ${h / 2}`),
  lamp: pic('Lamp', ELE, 60, 60, circ, (w, h) => `M${w * 0.15} ${h * 0.15}L${w * 0.85} ${h * 0.85}M${w * 0.85} ${h * 0.15}L${w * 0.15} ${h * 0.85}`),
  transistor: pic('Transistor', ELE, 70, 70, circ, (w, h) => `M${w * 0.4} ${h * 0.2}V${h * 0.8}M0 ${h / 2}H${w * 0.4}M${w * 0.4} ${h * 0.4}L${w * 0.75} ${h * 0.1}M${w * 0.4} ${h * 0.6}L${w * 0.75} ${h * 0.9}`),
  andgate: pic('AND gate', ELE, 90, 60, (w, h) => `M0 0H${w * 0.5}A${w * 0.5} ${h / 2} 0 0 1 ${w * 0.5} ${h}H0Z`, null, { ports: [[0, 0.25], [0, 0.75], [1, 0.5]] }),
  orgate: pic('OR gate', ELE, 90, 60, (w, h) => `M0 0Q${w * 0.6} 0 ${w} ${h / 2}Q${w * 0.6} ${h} 0 ${h}Q${w * 0.25} ${h / 2} 0 0Z`, null, { ports: [[0.1, 0.25], [0.1, 0.75], [1, 0.5]] }),
  notgate: pic('NOT gate', ELE, 80, 60, (w, h) => poly(w, h, [[0, 0], [0.8, 0.5], [0, 1]]), (w, h) => oval(w * 0.88, h / 2, w * 0.08, w * 0.08), { ports: [[0, 0.5], [1, 0.5]] }),
  chip: pic('Chip', ELE, 90, 90, (w, h) => `M${w * 0.18} ${h * 0.18}H${w * 0.82}V${h * 0.82}H${w * 0.18}Z`, (w, h) => [0.3, 0.5, 0.7].map((f) => `M${w * f} 0V${h * 0.18}M${w * f} ${h * 0.82}V${h}M0 ${h * f}H${w * 0.18}M${w * 0.82} ${h * f}H${w}`).join('')),

  // floor plan
  room: { label: 'Room', group: FLR, w: 200, h: 150, noBox: true, d: box, more: (w, h) => `M0 0H${w}V${h}H0Z`, ports: corners },
  wall: { label: 'Wall', group: FLR, w: 200, h: 10, fill: '#555555', d: box, ports: [[0, 0.5], [1, 0.5]] },
  door: line('Door', FLR, 60, 60, (w, h) => `M0 ${h}V0A${w} ${h} 0 0 1 ${w} ${h}`),
  fwindow: line('Window', FLR, 90, 14, (w, h) => `M0 0H${w}V${h}H0ZM0 ${h / 2}H${w}`, { ports: [[0, 0.5], [1, 0.5]] }),
  desk: pic('Desk', FLR, 110, 60, box, null, { ports: corners }),
  chair: pic('Chair', FLR, 40, 40, (w, h) => round(w, h, 8), (w) => `M4 -4H${w - 4}`),
  meetingtable: pic('Meeting table', FLR, 130, 70, (w, h) => round(w, h, 34), null),
  stairs: pic('Stairs', FLR, 70, 110, box, (w, h) => dividers(w, h, 7), { ports: corners }),

  // more basic forms
  semicircle: { label: 'Half circle', group: BAS, w: 120, h: 60, d: (w, h) => `M0 ${h}A${w / 2} ${h} 0 0 1 ${w} ${h}Z`, ports: [[0.5, 0], [1, 1], [0.5, 1], [0, 1]] },
  pie: { label: 'Pie slice', group: BAS, w: 100, h: 100, d: (w, h) => `M${w / 2} ${h / 2}L${w / 2} 0A${w / 2} ${h / 2} 0 1 1 0 ${h / 2}Z` },
  lshape: { label: 'L shape', group: BAS, w: 100, h: 100, d: (w, h) => poly(w, h, [[0, 0], [0.4, 0], [0.4, 0.6], [1, 0.6], [1, 1], [0, 1]]), ports: [[0.2, 0], [1, 0.8], [0.5, 1], [0, 0.5]] },
  tag: { label: 'Tag', group: BAS, w: 130, h: 60, d: (w, h) => poly(w, h, [[0.2, 0], [1, 0], [1, 1], [0.2, 1], [0, 0.5]]), more: (w, h) => oval(w * 0.2, h / 2, 4, 4) },
  banner: { label: 'Banner', group: BAS, w: 170, h: 60, d: (w, h) => poly(w, h, [[0, 0], [1, 0], [0.92, 0.5], [1, 1], [0, 1], [0.08, 0.5]]) },
  sun: pic('Sun', BAS, 80, 80, (w, h) => oval(w / 2, h / 2, w * 0.25, h * 0.25), (w, h) => [[0.5, 0], [0.5, 1], [0, 0.5], [1, 0.5], [0.15, 0.15], [0.85, 0.15], [0.15, 0.85], [0.85, 0.85]].map(([x, y]) => `M${w * (0.5 + (x - 0.5) * 0.65)} ${h * (0.5 + (y - 0.5) * 0.65)}L${w * x} ${h * y}`).join('')),
  moon: { label: 'Moon', group: BAS, w: 70, h: 90, d: (w, h) => `M${w * 0.7} 0A${w * 0.7} ${h / 2} 0 1 0 ${w * 0.7} ${h}A${w * 0.55} ${h * 0.5} 0 0 1 ${w * 0.7} 0Z` },
  bolt: { label: 'Lightning', group: BAS, w: 70, h: 100, d: (w, h) => poly(w, h, [[0.6, 0], [0, 0.55], [0.42, 0.55], [0.3, 1], [1, 0.38], [0.55, 0.38]]) },
  hcylinder: { label: 'Side cylinder', group: BAS, w: 150, h: 70, d: (w, h) => `M${w * 0.1} 0H${w * 0.9}A${w * 0.1} ${h / 2} 0 0 1 ${w * 0.9} ${h}H${w * 0.1}A${w * 0.1} ${h / 2} 0 0 1 ${w * 0.1} 0Z`, more: (w, h) => `M${w * 0.9} 0A${w * 0.1} ${h / 2} 0 0 0 ${w * 0.9} ${h}` },

  // maths and logic
  meq: pic('Equals', MTH, 60, 60, circ, (w, h) => `M${w * 0.28} ${h * 0.4}H${w * 0.72}M${w * 0.28} ${h * 0.6}H${w * 0.72}`),
  mne: pic('Not equal', MTH, 60, 60, circ, (w, h) => `M${w * 0.28} ${h * 0.4}H${w * 0.72}M${w * 0.28} ${h * 0.6}H${w * 0.72}M${w * 0.65} ${h * 0.22}L${w * 0.35} ${h * 0.78}`),
  mplus: pic('Add', MTH, 60, 60, circ, (w, h) => `M${w / 2} ${h * 0.25}V${h * 0.75}M${w * 0.25} ${h / 2}H${w * 0.75}`),
  mminus: pic('Subtract', MTH, 60, 60, circ, (w, h) => `M${w * 0.25} ${h / 2}H${w * 0.75}`),
  mdiv: pic('Divide', MTH, 60, 60, circ, (w, h) => `M${w * 0.25} ${h / 2}H${w * 0.75}${oval(w / 2, h * 0.3, 2, 2)}${oval(w / 2, h * 0.7, 2, 2)}`),

  // business
  person2: line('Stick figure', ORG, 50, 100, (w, h) => `${oval(w / 2, h * 0.13, h * 0.12, h * 0.12)}M${w / 2} ${h * 0.25}V${h * 0.62}M0 ${h * 0.4}H${w}M${w / 2} ${h * 0.62}L0 ${h}M${w / 2} ${h * 0.62}L${w} ${h}`),
  building: pic('Company', ORG, 80, 100, box, (w, h) => `${[0.15, 0.38, 0.61].map((y) => [0.2, 0.6].map((x) => `M${w * x} ${h * y}h${w * 0.2}v${h * 0.13}h-${w * 0.2}Z`).join('')).join('')}M${w * 0.38} ${h}V${h * 0.82}H${w * 0.62}V${h}`, { ports: corners }),
  money: pic('Money', ORG, 100, 60, (w, h) => round(w, h, 4), (w, h) => `${oval(w / 2, h / 2, h * 0.28, h * 0.28)}${oval(w * 0.13, h / 2, 3, 3)}${oval(w * 0.87, h / 2, 3, 3)}`),
  cart: line('Shopping cart', ORG, 90, 80, (w, h) => `M0 ${h * 0.05}H${w * 0.18}L${w * 0.32} ${h * 0.65}H${w * 0.85}L${w} ${h * 0.2}H${w * 0.22}${oval(w * 0.38, h * 0.88, 6, 6)}${oval(w * 0.78, h * 0.88, 6, 6)}`),
  truck: line('Delivery', ORG, 110, 70, (w, h) => `M0 ${h * 0.1}H${w * 0.6}V${h * 0.75}H0ZM${w * 0.6} ${h * 0.3}H${w * 0.82}L${w} ${h * 0.52}V${h * 0.75}H${w * 0.6}${oval(w * 0.2, h * 0.85, 8, 8)}${oval(w * 0.78, h * 0.85, 8, 8)}`),
  calendar: pic('Calendar', ORG, 80, 80, (w, h) => round(w, h, 4), (w, h) => `M0 ${h * 0.25}H${w}M${w * 0.25} 0V${h * 0.14}M${w * 0.75} 0V${h * 0.14}${[0.45, 0.72].map((y) => [0.25, 0.5, 0.75].map((x) => oval(w * x, h * y, 2, 2)).join('')).join('')}`),
  flag: line('Milestone flag', ORG, 60, 90, (w, h) => `M${w * 0.1} 0V${h}M${w * 0.1} ${h * 0.05}H${w}L${w * 0.75} ${h * 0.28}L${w} ${h * 0.5}H${w * 0.1}`),
  target: pic('Target', ORG, 80, 80, circ, (w, h) => `${oval(w / 2, h / 2, w * 0.32, h * 0.32)}${oval(w / 2, h / 2, w * 0.14, h * 0.14)}`)
};
