// Ready-made diagrams people can start from. Every one is ordinary shapes and connectors, so all of it can be moved and edited.
const build = (name, group, nodes, edges, bg = 'white') => {
  const shapes = nodes.map(([id, t, x, y, w, h, text, color]) => ({ id, t, x, y, w, h, text: text || '', ...(color ? { color } : {}) }));
  const lines = edges.map(([from, to, label, opts], i) => ({ id: 'e' + i, from, to, fromPort: null, toPort: null, kind: 'elbow', end: 'arrow', pts: [], label: label || '', ...(opts || {}) }));
  return { name, group, board: { bg, shapes, lines, pen: [] } };
};
const G = '#CDEBC0', B = '#DCEBFA', Y = '#FFE9A8', R = '#F7CFC8', P = '#E3D8F5';

export const BOARD_TEMPLATES = [
  build('Chatbot conversation flow', 'Flowcharts', [
    ['s', 'terminator', 300, 20, 160, 50, 'User sends message', G], ['i', 'process', 300, 110, 160, 60, 'Detect intent'], ['d', 'decision', 290, 210, 180, 90, 'Intent understood?', Y],
    ['a', 'process', 80, 340, 170, 60, 'Look up answer'], ['q', 'process', 510, 340, 170, 60, 'Ask to rephrase'], ['h', 'decision', 75, 440, 180, 90, 'Answer found?', Y],
    ['r', 'process', 80, 580, 170, 60, 'Send the reply'], ['g', 'process', 320, 580, 170, 60, 'Hand over to an agent'], ['e', 'terminator', 80, 690, 170, 50, 'End of chat', R], ['kb', 'database', 310, 345, 90, 100, 'Knowledge base']
  ], [['s', 'i'], ['i', 'd'], ['d', 'a', 'Yes'], ['d', 'q', 'No'], ['q', 's', 'Try again', { dash: 'dashed' }], ['a', 'kb', '', { start: 'arrow', dash: 'dashed' }], ['a', 'h'], ['h', 'r', 'Yes'], ['h', 'g', 'No'], ['r', 'e'], ['g', 'e']]),
  build('Login flow', 'Flowcharts', [
    ['s', 'terminator', 260, 20, 150, 50, 'Start', G], ['f', 'data', 245, 100, 180, 60, 'Enter email and password'], ['c', 'decision', 245, 200, 180, 90, 'Correct?', Y], ['m', 'decision', 245, 330, 180, 90, 'MFA passed?', Y],
    ['h', 'process', 260, 460, 150, 60, 'Open home page'], ['e', 'terminator', 260, 560, 150, 50, 'End', R], ['l', 'process', 490, 215, 160, 60, 'Count failed attempt'], ['b', 'decision', 480, 330, 180, 90, 'Five failures?', Y], ['k', 'process', 490, 460, 160, 60, 'Block the account']
  ], [['s', 'f'], ['f', 'c'], ['c', 'm', 'Yes'], ['m', 'h', 'Yes'], ['h', 'e'], ['c', 'l', 'No'], ['m', 'l', 'No'], ['l', 'b'], ['b', 'f', 'No', { dash: 'dashed' }], ['b', 'k', 'Yes'], ['k', 'e']]),
  build('Decision tree', 'Flowcharts', [
    ['a', 'decision', 310, 20, 180, 90, 'Budget approved?', Y], ['b', 'decision', 90, 170, 180, 90, 'Urgent?', Y], ['c', 'decision', 530, 170, 180, 90, 'Can it wait?', Y],
    ['d', 'process', 20, 320, 140, 60, 'Buy now'], ['e', 'process', 200, 320, 140, 60, 'Plan next month'], ['f', 'process', 460, 320, 140, 60, 'Next budget'], ['g', 'process', 640, 320, 140, 60, 'Ask for funds']
  ], [['a', 'b', 'Yes'], ['a', 'c', 'No'], ['b', 'd', 'Yes'], ['b', 'e', 'No'], ['c', 'f', 'Yes'], ['c', 'g', 'No']]),
  build('Approval swimlanes', 'Flowcharts', [
    ['l1', 'lane', 20, 20, 820, 130, 'Staff', '#F4F7F5'], ['l2', 'lane', 20, 150, 820, 150, 'Manager', '#F4F7F5'], ['l3', 'lane', 20, 300, 820, 130, 'Finance', '#F4F7F5'],
    ['a', 'terminator', 90, 60, 130, 50, 'Need arises', G], ['b', 'process', 270, 55, 150, 60, 'Write request'], ['c', 'decision', 260, 180, 170, 90, 'Approve?', Y], ['d', 'process', 490, 195, 140, 60, 'Send back'], ['e', 'process', 270, 335, 150, 60, 'Pay'], ['f', 'terminator', 500, 340, 130, 50, 'Done', R]
  ], [['a', 'b'], ['b', 'c'], ['c', 'e', 'Yes'], ['c', 'd', 'No'], ['d', 'b', '', { dash: 'dashed' }], ['e', 'f']]),

  build('DFD level 0 (context)', 'Data flow', [
    ['p', 'dfdprocess', 330, 170, 170, 170, '0 Order system'], ['c', 'dfdentity', 40, 215, 150, 80, 'Customer'], ['k', 'dfdentity', 640, 60, 150, 80, 'Kitchen'], ['m', 'dfdentity', 640, 370, 150, 80, 'Manager']
  ], [['c', 'p', 'Order'], ['p', 'c', 'Receipt', { kind: 'curve' }], ['p', 'k', 'Food order'], ['k', 'p', 'Order ready', { kind: 'curve' }], ['p', 'm', 'Daily report']]),
  build('DFD level 1', 'Data flow', [
    ['c', 'dfdentity', 30, 180, 140, 70, 'Customer'], ['p1', 'dfdprocess', 260, 40, 130, 130, '1.0 Take order'], ['p2', 'dfdprocess', 520, 40, 130, 130, '2.0 Take payment'], ['p3', 'dfdprocess', 390, 300, 130, 130, '3.0 Prepare order'],
    ['d1', 'dfdstore', 240, 220, 170, 50, 'D1 Orders'], ['d2', 'dfdstore', 720, 80, 170, 50, 'D2 Payments'], ['k', 'dfdentity', 660, 330, 140, 70, 'Kitchen']
  ], [['c', 'p1', 'Order details'], ['p1', 'd1', 'New order'], ['p1', 'p2', 'Amount due'], ['p2', 'd2', 'Payment record'], ['d1', 'p3', 'Open orders'], ['p3', 'k', 'Ticket'], ['k', 'p3', 'Ready', { kind: 'curve' }], ['p3', 'c', 'Food', { kind: 'curve' }]]),
  build('Data pipeline', 'Data flow', [
    ['a', 'server', 30, 60, 90, 90, 'Source apps'], ['b', 'queue', 190, 75, 140, 60, 'Event queue'], ['c', 'process', 400, 70, 140, 70, 'Clean and join'], ['d', 'datalake', 610, 50, 120, 100, 'Data lake'], ['e', 'warehouse', 610, 220, 120, 100, 'Warehouse'], ['f', 'chart', 400, 240, 90, 80, 'Dashboards']
  ], [['a', 'b'], ['b', 'c'], ['c', 'd'], ['d', 'e'], ['e', 'f']]),

  build('Shop database (ERD)', 'Database', [
    ['c', 'dbtable', 40, 40, 190, 130, 'customers\nPK id\nname\nemail'], ['o', 'dbtable', 330, 40, 190, 130, 'orders\nPK id\nFK customer_id\ncreated_at'], ['i', 'dbtable', 330, 240, 190, 130, 'order_items\nPK id\nFK order_id\nFK product_id\nquantity'], ['p', 'dbtable', 620, 240, 190, 130, 'products\nPK id\nname\nprice']
  ], [['c', 'o', 'places', { end: 'many', start: 'one', kind: 'elbow' }], ['o', 'i', 'has', { end: 'many', start: 'one' }], ['p', 'i', 'in', { end: 'many', start: 'one' }]]),
  build('ERD (Chen notation)', 'Database', [
    ['s', 'entity', 60, 120, 140, 70, 'Student'], ['r', 'relationship', 300, 110, 140, 90, 'Enrols'], ['c', 'entity', 540, 120, 140, 70, 'Course'], ['a1', 'keyattribute', 20, 10, 110, 50, 'Student ID'], ['a2', 'attribute', 150, 10, 100, 50, 'Name'], ['a3', 'keyattribute', 510, 10, 110, 50, 'Code'], ['a4', 'attribute', 640, 10, 100, 50, 'Title'], ['a5', 'attribute', 320, 10, 100, 50, 'Grade']
  ], [['s', 'r', 'M', { end: 'none' }], ['r', 'c', 'N', { end: 'none' }], ['a1', 's', '', { end: 'none', kind: 'straight' }], ['a2', 's', '', { end: 'none', kind: 'straight' }], ['a3', 'c', '', { end: 'none', kind: 'straight' }], ['a4', 'c', '', { end: 'none', kind: 'straight' }], ['a5', 'r', '', { end: 'none', kind: 'straight' }]]),

  build('Class diagram', 'UML', [
    ['a', 'umlclass', 280, 20, 190, 120, 'Account\n\n+ id\n+ email\n\n+ login()'], ['b', 'umlclass', 70, 230, 190, 120, 'Customer\n\n+ address\n\n+ placeOrder()'], ['c', 'umlclass', 490, 230, 190, 120, 'Admin\n\n+ level\n\n+ report()'], ['d', 'umlclass', 70, 430, 190, 120, 'Order\n\n+ total\n\n+ pay()']
  ], [['b', 'a', '', { end: 'triangle' }], ['c', 'a', '', { end: 'triangle' }], ['b', 'd', 'places', { start: 'odiamond', end: 'open' }]]),
  build('Use cases', 'UML', [
    ['f', 'frame', 220, 20, 360, 400, 'Hub'], ['u', 'actor', 70, 120, 60, 110, 'Staff'], ['m', 'actor', 660, 180, 60, 110, 'Manager'], ['a', 'usecase', 300, 70, 200, 70, 'Raise ticket'], ['b', 'usecase', 300, 180, 200, 70, 'Track ticket'], ['c', 'usecase', 300, 290, 200, 70, 'Approve request']
  ], [['u', 'a', '', { end: 'none', kind: 'straight' }], ['u', 'b', '', { end: 'none', kind: 'straight' }], ['m', 'c', '', { end: 'none', kind: 'straight' }], ['m', 'b', '', { end: 'none', kind: 'straight' }]]),
  build('Sequence diagram', 'UML', [
    ['a', 'lifeline', 60, 20, 130, 400, 'Browser'], ['b', 'lifeline', 300, 20, 130, 400, 'API'], ['c', 'lifeline', 540, 20, 130, 400, 'Database']
  ], [['a', 'b', 'POST /login', { fromPort: 1, toPort: 3, kind: 'straight' }]].concat([])),
  build('State machine', 'UML', [
    ['i', 'initial', 40, 130, 30, 30], ['a', 'state', 130, 110, 140, 70, 'Raised'], ['b', 'state', 350, 110, 140, 70, 'In progress'], ['c', 'state', 570, 110, 140, 70, 'Resolved'], ['d', 'state', 570, 280, 140, 70, 'Closed'], ['z', 'final', 760, 300, 34, 34]
  ], [['i', 'a'], ['a', 'b', 'assign'], ['b', 'c', 'solve'], ['c', 'b', 'reopen', { kind: 'curve', dash: 'dashed' }], ['c', 'd', 'close'], ['d', 'z']]),

  build('Three-tier web app', 'Network and cloud', [
    ['u', 'users', 40, 170, 80, 70, 'Users'], ['w', 'firewall', 180, 165, 70, 80, 'Firewall'], ['l', 'loadbalancer', 310, 170, 90, 70, 'Load balancer'], ['a', 'server', 470, 50, 90, 90, 'App 1'], ['b', 'server', 470, 270, 90, 90, 'App 2'], ['c', 'cache', 640, 40, 110, 70, 'Cache'], ['d', 'dbstack', 640, 170, 110, 110, 'Primary DB'], ['e', 'storage', 640, 340, 110, 70, 'Files']
  ], [['u', 'w'], ['w', 'l'], ['l', 'a'], ['l', 'b'], ['a', 'c'], ['a', 'd'], ['b', 'd'], ['b', 'e']]),
  build('Office network', 'Network and cloud', [
    ['i', 'cloud', 330, 10, 160, 90, 'Internet'], ['r', 'router', 370, 150, 80, 60, 'Router'], ['f', 'firewall', 375, 250, 70, 80, 'Firewall'], ['s', 'switch', 340, 380, 140, 40, 'Switch'], ['a', 'desktop', 100, 490, 90, 80, 'Desktop'], ['b', 'laptop', 280, 490, 100, 70, 'Laptop'], ['c', 'printer', 450, 490, 90, 70, 'Printer'], ['d', 'server', 630, 480, 90, 90, 'File server']
  ], [['i', 'r', '', { end: 'none' }], ['r', 'f', '', { end: 'none' }], ['f', 's', '', { end: 'none' }], ['s', 'a', '', { end: 'none' }], ['s', 'b', '', { end: 'none' }], ['s', 'c', '', { end: 'none' }], ['s', 'd', '', { end: 'none' }]]),
  build('Build and release pipeline', 'Network and cloud', [
    ['a', 'laptop', 20, 60, 100, 70, 'Developer'], ['b', 'cylinderrepo', 170, 50, 100, 90, 'Repository'], ['c', 'gear', 330, 55, 80, 80, 'Build'], ['d', 'check', 470, 55, 80, 80, 'Tests'], ['e', 'container', 610, 50, 110, 90, 'Image'], ['f', 'server', 620, 240, 90, 90, 'Production']
  ], [['a', 'b', 'push'], ['b', 'c'], ['c', 'd'], ['d', 'e'], ['e', 'f', 'deploy']]),

  build('Organization chart', 'People and ideas', [
    ['a', 'rect', 300, 20, 170, 60, 'Managing director', P], ['b', 'rect', 70, 140, 160, 60, 'Finance', B], ['c', 'rect', 305, 140, 160, 60, 'Operations', B], ['d', 'rect', 540, 140, 160, 60, 'Technology', B], ['e', 'rect', 430, 260, 140, 60, 'Software', Y], ['f', 'rect', 620, 260, 140, 60, 'Support', Y]
  ], [['a', 'b', '', { end: 'none' }], ['a', 'c', '', { end: 'none' }], ['a', 'd', '', { end: 'none' }], ['d', 'e', '', { end: 'none' }], ['d', 'f', '', { end: 'none' }]]),
  build('Mind map', 'People and ideas', [
    ['c', 'ellipse', 310, 190, 170, 100, 'Project idea', Y], ['a', 'rect', 80, 60, 130, 50, 'Goals', G], ['b', 'rect', 80, 370, 130, 50, 'Risks', R], ['d', 'rect', 590, 60, 130, 50, 'People', B], ['e', 'rect', 590, 370, 130, 50, 'Budget', P]
  ], ['a', 'b', 'd', 'e'].map((x) => ['c', x, '', { kind: 'curve', end: 'none' }]))
];
// a sequence diagram's messages run between points on the lifelines, not the shapes' edges
const seq = BOARD_TEMPLATES.find((t) => t.name === 'Sequence diagram').board;
seq.lines = [['POST /login', 90, 'a', 'b', ''], ['SELECT user', 160, 'b', 'c', ''], ['user row', 230, 'c', 'b', 'dashed'], ['token', 300, 'b', 'a', 'dashed']].map(([label, y, f, t, dash], i) => {
  const sx = { a: 125, b: 365, c: 605 };
  return { id: 'm' + i, from: null, to: null, x1: sx[f], y1: y, x2: sx[t], y2: y, kind: 'straight', end: 'arrow', dash: dash || 'solid', label, pts: [] };
});
