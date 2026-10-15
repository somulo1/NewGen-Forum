// Converts a Claude Design .dc.html file into a Vue 3 SFC that renders identically.
const fs = require('fs');
const [,, src, out, compName] = process.argv;
const html = fs.readFileSync(src, 'utf8');
const helmet = (html.match(/<helmet>([\s\S]*?)<\/helmet>/) || [, ''])[1];
const css = (helmet.match(/<style>([\s\S]*?)<\/style>/) || [, ''])[1];
let markup = html.split('</helmet>')[1].split('</x-dc>')[0];
let script = html.match(/<script type="text\/x-dc"[^>]*>([\s\S]*?)<\/script>/)[1];

const loopVars = new Set();
const exprs = [];
const hole = (s) => s.replace(/^\{\{\s*|\s*\}\}$/g, '');
markup = markup.replace(/\s+hint-[a-z-]+="[^"]*"/g, '');
markup = markup.replace(/<sc-if value="\{\{\s*([^}]+?)\s*\}\}">/g, (m, x) => { exprs.push(x); return `<template v-if="${x}">`; }).replace(/<\/sc-if>/g, '</template>');
markup = markup.replace(/<sc-for list="\{\{\s*([^}]+?)\s*\}\}" as="(\w+)">/g, (m, x, a) => { exprs.push(x); loopVars.add(a); return `<template v-for="(${a}, ${a}__i) in ${x}" :key="${a}__i">`; }).replace(/<\/sc-for>/g, '</template>');
markup = markup.replace(/<dc-import name="(\w+)" t="\{\{\s*([^}]+?)\s*\}\}"[^>]*><\/dc-import>/g, (m, n, x) => { exprs.push(x); return `<Data${n} :t="${x}" />`; });
const changeTag = (tag) => /^<select/.test(tag) || /type="(checkbox|radio|file)"/.test(tag);
markup = markup.replace(/<[a-zA-Z][^<>]*>/g, (tag) => {
  if (tag.startsWith('<template') || tag.startsWith('<Data')) return tag;
  return tag.replace(/\s([a-zA-Z][\w-]*)="([^"]*\{\{[^"]*)"/g, (m, name, val) => {
    const pure = /^\{\{\s*[^}]+?\s*\}\}$/.test(val);
    if (/^on[A-Z]/.test(name)) {
      let ev = name.slice(2).toLowerCase();
      if (ev === 'change' && !changeTag(tag)) ev = 'input';
      const x = hole(val); exprs.push(x); return ` @${ev}="${x}"`;
    }
    if (pure) { const x = hole(val); exprs.push(x); return ` :${name}="${x}"`; }
    const lit = val.replace(/`/g, '\\`').replace(/\{\{\s*([^}]+?)\s*\}\}/g, (mm, x) => { exprs.push(x); return '${' + x + '}'; });
    return ` :${name}="\`${lit}\`"`;
  });
});
markup.replace(/\{\{\s*([^}]+?)\s*\}\}/g, (m, x) => { exprs.push(x); return m; });
const roots = new Set();
exprs.forEach((x) => { const r = x.trim().match(/^[A-Za-z_$][\w$]*/); if (r && !loopVars.has(r[0]) && !['true', 'false', 'null'].includes(r[0])) roots.add(r[0]); });
const names = [...roots].sort();
const bad = names.filter((n) => /^[$_]/.test(n));
if (bad.length) console.error('reserved names:', bad);
const comps = markup.includes('<DataTable') ? "import DataTable from './DataTable.vue';\n" : '';
const sfc = `<!-- Generated from ${src.split('/').pop()} — exact port of the Claude Design prototype. -->
<template>
<div class="dc-root" style="display: contents">${markup}</div>
</template>

<script>
${comps}import { DCLogic, useDC } from '../../utils/dcLogic.js';

${script.trim()}

export default {
  name: '${compName}',${comps ? "\n  components: { DataTable }," : ''}
  props: { t: { type: Object, default: null } },
  ...useDC(Component, ${JSON.stringify(names)})
};
</script>
`;
fs.writeFileSync(out, sfc);
fs.writeFileSync(out.replace(/\.vue$/, '.css'), css);
console.log(compName, 'roots:', names.length, 'loopVars:', loopVars.size);
