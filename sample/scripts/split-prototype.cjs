// Splits prototype-source/generated/HubApp.vue (the exact port) into a structured Nuxt app:
// layouts/default.vue, pages/*, components/{layout,sections,panels,overlays}/*, logic/HubLogic.js.
// Markup and inline styles are moved verbatim, so the design stays pixel-identical.
const fs = require('fs');
const path = require('path');
const ROOT = path.resolve(__dirname, '..');
const src = fs.readFileSync(path.join(ROOT, 'prototype-source/generated/HubApp.vue'), 'utf8');
const tStart = src.indexOf('<template>') + '<template>'.length;
const tEnd = src.lastIndexOf('</template>', src.indexOf('<script>'));
let tpl = src.slice(tStart, tEnd);
const script = src.slice(src.indexOf('<script>') + 8, src.lastIndexOf('</script>'));

const VOID = new Set(['input', 'img', 'br', 'hr', 'meta', 'link', 'source', 'area', 'col', 'embed', 'wbr']);
// find the full element that starts with the first tag matching `opener` (a string search)
function blockAt(text, opener, from = 0) {
  const s = text.indexOf(opener, from);
  if (s < 0) throw new Error('not found: ' + opener);
  const re = /<(\/?)([a-zA-Z][\w-]*)((?:"[^"]*"|'[^']*'|[^'">])*?)(\/?)>/g;
  re.lastIndex = s;
  let depth = 0, m;
  while ((m = re.exec(text))) {
    const [all, close, tag, , self] = m;
    if (close) depth--; else if (!self && !VOID.has(tag.toLowerCase())) depth++;
    if (depth === 0) return [s, m.index + all.length];
  }
  throw new Error('unclosed: ' + opener);
}

// expression roots used in a chunk of markup
function rootsOf(chunk) {
  const exprs = [], loops = new Set();
  chunk.replace(/v-for="\((\w+), (\w+)\) in ([^"]+)"/g, (m, a, i, x) => { loops.add(a); loops.add(i); exprs.push(x); });
  chunk.replace(/v-if="([^"]+)"/g, (m, x) => exprs.push(x));
  chunk.replace(/\s[:@][\w-]+="([^"]*)"/g, (m, x) => {
    if (x.startsWith('`')) x.replace(/\$\{([^}]+)\}/g, (mm, y) => exprs.push(y)); else exprs.push(x);
  });
  chunk.replace(/\{\{\s*([^}]+?)\s*\}\}/g, (m, x) => exprs.push(x));
  const roots = new Set();
  exprs.forEach((x) => { const r = x.trim().match(/^[A-Za-z_$][\w$]*/); if (r && !loops.has(r[0]) && !['true', 'false', 'null', 'undefined'].includes(r[0])) roots.add(r[0]); });
  return [...roots].sort();
}

const files = {};
function component(dir, name, markup, note) {
  const names = rootsOf(markup);
  files[`components/${dir}/${name}.vue`] = `<!-- ${note} -->
<template>${markup}
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: '${name}',
  ...hubBindings(${JSON.stringify(names)})
};
</script>
`;
}

// ---- pull out blocks, replacing them with component tags ----
const take = (opener, replacement, from) => {
  const [a, b] = blockAt(tpl, opener, from ? tpl.indexOf(from) : 0);
  const chunk = tpl.slice(a, b);
  tpl = tpl.slice(0, a) + replacement + tpl.slice(b);
  return chunk;
};

// overlays & panels (outside <main>)
const overlays = [
  ['<template v-if="previewing">', 'layout', 'PreviewBanner', 'Banner shown while previewing the app as another role'],
  ['<header class="hub-header"', 'layout', 'AppHeader', 'Top bar: menu (mobile), logo, search, theme, notifications, profile'],
  ['<template v-if="navOpen">', 'layout', 'NavBackdrop', 'Mobile: dark backdrop behind the slide-out menu'],
  ['<nav aria-label="Main"', 'layout', 'SideNav', 'Sidebar: main pages, admin, messages, organization, forums, profile'],
  ['<template v-if="isDocked">', 'panels', 'ForumPanel', 'Docked side panel for a forum or private chat (full screen on mobile)'],
  ['<template v-if="isFullDm">', 'panels', 'PrivateChatFull', 'Private chat in full screen'],
  ['<template v-if="isFullForum">', 'panels', 'ForumFull', 'Forum in full screen'],
  ['<template v-if="awOpen">', 'overlays', 'AwardDialog', 'Give an award dialog'],
  ['<template v-if="evOpen">', 'overlays', 'EventDialog', 'Calendar: new event / meeting / personal pin dialog'],
  ['<template v-if="cmOpen">', 'overlays', 'CommComposer', 'Memo / announcement / newsletter composer'],
  ['<template v-if="personOpen">', 'overlays', 'PersonDrawer', 'Person profile drawer'],
  ['<template v-if="taskOpen">', 'overlays', 'TaskDrawer', 'Project task drawer'],
  ['<template v-if="settingsOpen">', 'overlays', 'WidgetSettingsDrawer', 'Widget settings drawer'],
  ['<template v-if="inCall">', 'overlays', 'CallScreen', 'Voice / video call screen'],
  ['<template v-if="hasToast">', 'overlays', 'AppToast', 'Toast notification'],
  ['<template v-if="galleryOpen">', 'overlays', 'WidgetGallery', 'Add-widget gallery'],
  ['<template v-if="showTabbar">', 'layout', 'MobileTabBar', 'Mobile bottom tab bar'],
  ['<template v-if="isMin">', 'panels', 'PanelBubble', 'Minimized forum / chat bubble with its menu']
];
overlays.forEach(([op, dir, name, note]) => component(dir, name, take(op, `<${name} />`), note));

// sections inside <main> → one component each (grouped), pages render them
const sectionDefs = [
  ['HomeSection', ['<template v-if="show.home">'], '/', 'home', 'Home: awards ribbon + widget masonry'],
  ['ProfileSection', ['<template v-if="show.profile">'], '/profile', 'profile', 'My profile'],
  ['WallSection', ['<template v-if="show.feed">'], '/wall', 'feed', 'Community wall'],
  ['MeetingsSection', ['<template v-if="show.meetings">'], '/meetings', 'meetings', 'Meetings & standups'],
  ['TicketsSection', ['<template v-if="showTicketList">', '<template v-if="showTicketView">'], '/tickets', 'tickets', 'Tickets list and ticket details'],
  ['CalendarSection', ['<template v-if="show.calendar">'], '/calendar', 'calendar', 'Calendar'],
  ['CommsSection', ['<template v-if="show.comms">'], '/communications', 'comms', 'Communications (memos, newsletter, announcements)'],
  ['StructureSection', ['<template v-if="show.structure">'], '/structure', 'structure', 'Organization structure'],
  ['OrgUnitSection', ['<template v-if="show.orgunit">'], '/org', 'orgunit', 'Department / sector / branch / unit dashboard'],
  ['PeopleSection', ['<template v-if="show.people">'], '/people', 'people', 'People directory'],
  ['SecuritySection', ['<template v-if="show.security">'], '/security', 'security', 'Security & logs'],
  ['RolesSection', ['<template v-if="show.roles">'], '/roles', 'roles', 'Roles & rights'],
  ['ProjectsSection', ['<template v-if="showProjectList">', '<template v-if="showProjectView">'], '/projects', 'projects', 'Projects list and project view'],
  ['FilesSection', ['<template v-if="showFolderList">', '<template v-if="showFolderView">'], '/files', 'files', 'Files and folder view'],
  ['SecureShareSection', ['<template v-if="show.share">'], '/secure-share', 'share', 'Secure share (one-time links)'],
  ['ApprovalsSection', ['<template v-if="show.approvals">'], '/approvals', 'approvals', 'Approvals inbox']
];
const routes = {};
sectionDefs.forEach(([name, openers, route, section, note]) => {
  const parts = openers.map((op) => take(op, ''));
  component('sections', name, parts.join('\n\n'), note);
  routes[section] = route;
  const file = route === '/' ? 'pages/index.vue' : `pages${route}.vue`;
  files[file] = `<template>\n  <${name} />\n</template>\n`;
});
// legacy blocks that never render
while (tpl.indexOf('<template v-if="retired">') !== -1) take('<template v-if="retired">', '');

// <main> keeps its opening tag; contents become the page slot
const [ma, mb] = blockAt(tpl, '<main class="hub-main"');
const mainOpen = tpl.slice(ma, tpl.indexOf('>', ma) + 1);
tpl = tpl.slice(0, ma) + mainOpen + '\n<slot />\n</main>' + tpl.slice(mb);
tpl = tpl.replace(/\n{3,}/g, '\n\n');

const layoutNames = rootsOf(tpl);
files['layouts/default.vue'] = `<!-- App shell: header, sidebar, page slot, docked panel and every overlay. -->
<template>${tpl}
</template>

<script>
import { hubBindings, useHubShell } from '~/composables/useHub';

export default {
  name: 'DefaultLayout',
  ...hubBindings(${JSON.stringify(layoutNames)}),
  setup() { useHubShell(); }
};
</script>
`;

// logic module
const logic = script
  .replace(/import DataTable[^\n]*\n/, '')
  .replace(/import \{ DCLogic, useDC \}[^\n]*\n/, "import { DCLogic } from '~/utils/dcLogic.js';\n")
  .replace('class Component extends DCLogic', '// All Hub state, sample data, business rules and actions.\n// renderVals() computes everything the templates show; actions call setState().\nexport class HubLogic extends DCLogic')
  .replace(/\nexport default \{[\s\S]*$/, '\n');
files['logic/HubLogic.js'] = logic.trim() + '\n';
files['logic/routes.js'] = `// Section ↔ URL map used by the layout to keep the address bar in sync.\nexport const sectionRoutes = ${JSON.stringify(routes, null, 2)};\n`;

for (const [f, c] of Object.entries(files)) {
  const p = path.join(ROOT, f);
  fs.mkdirSync(path.dirname(p), { recursive: true });
  fs.writeFileSync(p, c);
}
console.log('wrote', Object.keys(files).length, 'files');
