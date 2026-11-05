<!-- Word-style rich text editor for notes. Pasted Word / web / Excel content keeps its bold, headings, lists, colours and tables. -->
<template>
<div class="nr-wrap" @click="onDocClick">
<div v-if="editor && editable" class="nr-bar" role="toolbar" aria-label="Formatting" @mousedown="keepFocus">
<div class="nr-grp">
<button type="button" class="nr-b" title="Undo (Ctrl+Z)" aria-label="Undo" :disabled="!can('undo')" @click="run(c => c.undo())">↶</button>
<button type="button" class="nr-b" title="Redo (Ctrl+Y)" aria-label="Redo" :disabled="!can('redo')" @click="run(c => c.redo())">↷</button>
</div>
<div class="nr-grp">
<select class="nr-s" aria-label="Text style" :value="block" @change="setBlock($event.target.value)">
<option value="p">Normal</option><option value="1">Heading 1</option><option value="2">Heading 2</option><option value="3">Heading 3</option><option value="4">Heading 4</option>
</select>
<select class="nr-s" aria-label="Font" :value="attr('fontFamily')" @change="setFont($event.target.value)">
<option value="">Font</option>
<option v-for="f in fonts" :key="f.v" :value="f.v">{{ f.l }}</option>
</select>
<select class="nr-s nr-s-sm" aria-label="Font size" :value="attr('fontSize')" @change="setSize($event.target.value)">
<option value="">Size</option>
<option v-for="z in sizes" :key="z" :value="z + 'px'">{{ z }}</option>
</select>
</div>
<div class="nr-grp">
<button type="button" class="nr-b" :class="{ on: on('bold') }" title="Bold (Ctrl+B)" aria-label="Bold" @click="run(c => c.toggleBold())"><b>B</b></button>
<button type="button" class="nr-b" :class="{ on: on('italic') }" title="Italic (Ctrl+I)" aria-label="Italic" @click="run(c => c.toggleItalic())"><i>I</i></button>
<button type="button" class="nr-b" :class="{ on: on('underline') }" title="Underline (Ctrl+U)" aria-label="Underline" @click="run(c => c.toggleUnderline())"><u>U</u></button>
<button type="button" class="nr-b" :class="{ on: on('strike') }" title="Strikethrough" aria-label="Strikethrough" @click="run(c => c.toggleStrike())"><s>S</s></button>
<button type="button" class="nr-b" :class="{ on: on('subscript') }" title="Subscript" aria-label="Subscript" @click="run(c => c.toggleSubscript())">x<sub>2</sub></button>
<button type="button" class="nr-b" :class="{ on: on('superscript') }" title="Superscript" aria-label="Superscript" @click="run(c => c.toggleSuperscript())">x<sup>2</sup></button>
<label class="nr-b nr-col" title="Text colour"><span :style="`border-bottom: 3px solid ${attr('color') || '#3A3520'}`">A</span><input type="color" aria-label="Text colour" :value="hex(attr('color'), '#3A3520')" @input="run(c => c.setColor($event.target.value))"></label>
<label class="nr-b nr-col" title="Highlight colour"><span :style="`border-bottom: 3px solid ${hl || 'var(--k-FFE066,#FFE066)'}`">ab</span><input type="color" aria-label="Highlight colour" :value="hex(hl, '#FFE066')" @input="run(c => c.setHighlight({ color: $event.target.value }))"></label>
<button type="button" class="nr-b" title="Clear formatting" aria-label="Clear formatting" @click="run(c => c.unsetAllMarks().clearNodes())">T<small>x</small></button>
</div>
<div class="nr-grp">
<button type="button" class="nr-b" :class="{ on: editor.isActive({ textAlign: 'left' }) }" title="Align left" aria-label="Align left" @click="run(c => c.setTextAlign('left'))">⯇</button>
<button type="button" class="nr-b" :class="{ on: editor.isActive({ textAlign: 'center' }) }" title="Centre" aria-label="Centre" @click="run(c => c.setTextAlign('center'))">≡</button>
<button type="button" class="nr-b" :class="{ on: editor.isActive({ textAlign: 'right' }) }" title="Align right" aria-label="Align right" @click="run(c => c.setTextAlign('right'))">⯈</button>
<button type="button" class="nr-b" :class="{ on: editor.isActive({ textAlign: 'justify' }) }" title="Justify" aria-label="Justify" @click="run(c => c.setTextAlign('justify'))">☰</button>
</div>
<div class="nr-grp">
<button type="button" class="nr-b" :class="{ on: on('bulletList') }" title="Bullet list" aria-label="Bullet list" @click="run(c => c.toggleBulletList())">• ≡</button>
<button type="button" class="nr-b" :class="{ on: on('orderedList') }" title="Numbered list" aria-label="Numbered list" @click="run(c => c.toggleOrderedList())">1. ≡</button>
<button type="button" class="nr-b" :class="{ on: on('taskList') }" title="Checklist" aria-label="Checklist" @click="run(c => c.toggleTaskList())">☑</button>
<button type="button" class="nr-b" title="Decrease indent" aria-label="Decrease indent" @click="outdent">⇤</button>
<button type="button" class="nr-b" title="Increase indent" aria-label="Increase indent" @click="indent">⇥</button>
</div>
<div class="nr-grp">
<button type="button" class="nr-b" :class="{ on: on('blockquote') }" title="Quote" aria-label="Quote" @click="run(c => c.toggleBlockquote())">❝</button>
<button type="button" class="nr-b" :class="{ on: on('codeBlock') }" title="Code block" aria-label="Code block" @click="run(c => c.toggleCodeBlock())">&lt;/&gt;</button>
<button type="button" class="nr-b" title="Horizontal line" aria-label="Horizontal line" @click="run(c => c.setHorizontalRule())">―</button>
<button type="button" class="nr-b" :class="{ on: on('link') }" title="Link" aria-label="Link" @click="setLink">🔗</button>
<select v-if="macros" class="nr-s" aria-label="Insert a macro" value="" @change="insertMacro($event.target.value); $event.target.value = ''"><option value="">Insert…</option><option value="info">Info panel</option><option value="note">Note panel</option><option value="warning">Warning panel</option><option value="success">Success panel</option><option value="expand">Expand section</option><option value="status">Status label</option><option value="toc">Table of contents</option><option value="children">Child pages</option><option value="tasks">Open tasks</option><option value="board">Whiteboard or diagram</option></select>
<label class="nr-b" title="Insert a picture" aria-label="Insert a picture" style="position: relative; display: inline-flex; align-items: center; justify-content: center; cursor: pointer">🖼<input type="file" accept="image/*" aria-label="Insert a picture" @change="pickImage" style="position: absolute; inset: 0; opacity: 0; cursor: pointer"></label>
</div>
<div class="nr-grp">
<span class="nr-pop-wrap">
<button type="button" class="nr-b" :class="{ on: on('table') }" title="Insert table" aria-label="Insert table" aria-haspopup="true" :aria-expanded="grid.open" @click="grid.open = !grid.open">▦ Table</button>
<div v-if="grid.open" class="nr-grid-pop" @mouseleave="grid.r = 0; grid.c = 0">
<div class="nr-grid"><template v-for="r in 8" :key="r"><span v-for="c in 8" :key="c" class="nr-cell" :class="{ hot: r <= grid.r && c <= grid.c }" @mouseenter="grid.r = r; grid.c = c" @click="insertTable(r, c)"></span></template></div>
<div class="nr-grid-lbl">{{ grid.r ? grid.r + ' × ' + grid.c : 'Choose size' }}</div>
</div>
</span>
<EmojiButton class="nr-emoji" :size="30" shape="6px" @pick="addEmoji"/>
</div>
<div v-if="on('table')" class="nr-grp nr-tbl" aria-label="Table tools">
<button type="button" class="nr-b nr-t" @click="run(c => c.addRowBefore())">Row ↑</button>
<button type="button" class="nr-b nr-t" @click="run(c => c.addRowAfter())">Row ↓</button>
<button type="button" class="nr-b nr-t" @click="run(c => c.deleteRow())">Del row</button>
<button type="button" class="nr-b nr-t" @click="run(c => c.addColumnBefore())">Col ←</button>
<button type="button" class="nr-b nr-t" @click="run(c => c.addColumnAfter())">Col →</button>
<button type="button" class="nr-b nr-t" @click="run(c => c.deleteColumn())">Del col</button>
<button type="button" class="nr-b nr-t" @click="run(c => c.mergeCells())" :disabled="!can('mergeCells')">Merge</button>
<button type="button" class="nr-b nr-t" @click="run(c => c.splitCell())" :disabled="!can('splitCell')">Split</button>
<button type="button" class="nr-b nr-t" @click="run(c => c.toggleHeaderRow())">Header row</button>
<button type="button" class="nr-b nr-t nr-del" @click="run(c => c.deleteTable())">Delete table</button>
</div>
</div>
<EditorContent :editor="editor" class="nr-doc" />
</div>
</template>

<script setup>
import { ref, shallowRef, reactive, watch, onMounted, onBeforeUnmount } from 'vue';
import { Editor, EditorContent } from '@tiptap/vue-3';
import { Node, Mark, mergeAttributes, generateJSON, getSchema } from '@tiptap/core';
import Collaboration from '@tiptap/extension-collaboration';
import { prosemirrorToYXmlFragment } from '@tiptap/y-tiptap';
import { Node as PMNode } from '@tiptap/pm/model';
import { imageDataUrl } from '~/logic/confluenceImport.js';
import StarterKit from '@tiptap/starter-kit';
import { Table, TableRow, TableCell, TableHeader } from '@tiptap/extension-table';
import { TextStyleKit } from '@tiptap/extension-text-style';
import Highlight from '@tiptap/extension-highlight';
import TextAlign from '@tiptap/extension-text-align';
import Placeholder from '@tiptap/extension-placeholder';
import Subscript from '@tiptap/extension-subscript';
import Superscript from '@tiptap/extension-superscript';
import { TaskList, TaskItem } from '@tiptap/extension-list';

const props = defineProps({ modelValue: { type: String, default: '' }, editable: { type: Boolean, default: true }, live: { type: Object, default: null }, liveSeed: { type: String, default: '' }, macros: { type: Boolean, default: false } });
const emit = defineEmits(['update:modelValue', 'ref', 'comment']);

// A picture lives inside the note or page itself, as data: nothing is fetched from elsewhere unless it was pasted as a web address.
const safeSrc = (v) => /^(data:image\/(png|jpeg|gif|webp);base64,|data:image\/svg\+xml;charset=utf-8,|https:\/\/)/i.test(v || '');
const HubImage = Node.create({
  name: 'image', group: 'block', atom: true, draggable: true,
  addAttributes() { return { src: { default: null }, alt: { default: '' } }; },
  parseHTML() { return [{ tag: 'img[src]', getAttrs: (el) => (safeSrc(el.getAttribute('src')) ? {} : false) }]; },
  renderHTML({ HTMLAttributes }) { return ['img', mergeAttributes(HTMLAttributes, safeSrc(HTMLAttributes.src) ? {} : { src: '' })]; }
});
// Only shown while reading: a link to a page or task, and text someone commented on.
const HubMark = Mark.create({
  name: 'hubMark',
  addAttributes() { return { ref: { default: null, parseHTML: (el) => el.getAttribute('data-ref'), renderHTML: (a) => (a.ref ? { 'data-ref': a.ref } : {}) }, comment: { default: null, parseHTML: (el) => el.getAttribute('data-comment'), renderHTML: (a) => (a.comment ? { 'data-comment': a.comment } : {}) } }; },
  parseHTML() { return [{ tag: 'span[data-ref]' }, { tag: 'span[data-comment]' }]; },
  renderHTML({ HTMLAttributes }) { return ['span', HTMLAttributes, 0]; }
});
// Confluence-style macros, for pages
const Panel = Node.create({
  name: 'panel', group: 'block', content: 'block+', defining: true,
  addAttributes() { return { kind: { default: 'info', parseHTML: (el) => el.getAttribute('data-panel'), renderHTML: (a) => ({ 'data-panel': a.kind }) } }; },
  parseHTML() { return [{ tag: 'div[data-panel]' }]; },
  renderHTML({ HTMLAttributes }) { return ['div', HTMLAttributes, 0]; }
});
const Expand = Node.create({
  name: 'expand', group: 'block', content: 'block+', defining: true,
  addAttributes() { return { title: { default: 'More', parseHTML: (el) => el.getAttribute('data-expand'), renderHTML: (a) => ({ 'data-expand': a.title }) } }; },
  parseHTML() { return [{ tag: 'div[data-expand]' }]; },
  renderHTML({ HTMLAttributes }) { return ['div', HTMLAttributes, 0]; }
});
const Macro = Node.create({
  name: 'macro', group: 'block', atom: true,
  addAttributes() { return { kind: { default: 'toc', parseHTML: (el) => el.getAttribute('data-macro'), renderHTML: (a) => ({ 'data-macro': a.kind }) }, arg: { default: '', parseHTML: (el) => el.getAttribute('data-arg') || '', renderHTML: (a) => (a.arg ? { 'data-arg': a.arg } : {}) } }; },
  parseHTML() { return [{ tag: 'div[data-macro]' }]; },
  renderHTML({ HTMLAttributes }) { return ['div', HTMLAttributes]; }
});
const Status = Mark.create({
  name: 'status',
  addAttributes() { return { color: { default: 'blue', parseHTML: (el) => el.getAttribute('data-status'), renderHTML: (a) => ({ 'data-status': a.color }) } }; },
  parseHTML() { return [{ tag: 'span[data-status]' }]; },
  renderHTML({ HTMLAttributes }) { return ['span', HTMLAttributes, 0]; }
});
const MAX_IMAGE = 15 << 20;
const addImage = async (ed, file) => { if (file.size > MAX_IMAGE) { window.alert('That picture is over 15 MB.'); return; } try { ed.chain().focus().insertContent({ type: 'image', attrs: { src: await imageDataUrl(file), alt: file.name } }).run(); } catch (e) { window.alert('That picture could not be read.'); } };
const imagesOf = (list) => Array.from(list || []).filter((f) => /^image\/(png|jpeg|gif|webp)/.test(f.type));

const fonts = [
  { l: 'Default', v: 'DM Sans, sans-serif' }, { l: 'Space Grotesk', v: 'Space Grotesk, sans-serif' }, { l: 'Arial', v: 'Arial, sans-serif' },
  { l: 'Calibri', v: 'Calibri, Carlito, sans-serif' }, { l: 'Times New Roman', v: '"Times New Roman", serif' }, { l: 'Georgia', v: 'Georgia, serif' },
  { l: 'Courier New', v: '"Courier New", monospace' }
];
const sizes = [10, 11, 12, 14, 16, 18, 20, 24, 28, 36];

let unmounted = false;
const editor = shallowRef(null);
const tick = ref(0);
const grid = reactive({ open: false, r: 0, c: 0 });
const hl = ref('');

const esc = (t) => t.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
// Notes saved before rich text are plain text: turn each line into a paragraph.
const toHtml = (v) => (!v ? '' : /^\s*</.test(v) ? v : v.split('\n').map((l) => '<p>' + esc(l) + '</p>').join(''));

onMounted(async () => {
  const baseExt = () => [
      StarterKit.configure({ ...(props.live ? { undoRedo: false } : {}), link: { openOnClick: false, autolink: true, HTMLAttributes: { rel: 'noopener noreferrer', target: '_blank' } } }),
      HubImage, HubMark, Panel, Expand, Macro, Status, TextStyleKit, Subscript, Superscript, Highlight.configure({ multicolor: true }),
      TextAlign.configure({ types: ['heading', 'paragraph'] }),
      Table.configure({ resizable: true }), TableRow, TableHeader, TableCell,
      TaskList, TaskItem.configure({ nested: true })
    ];
  const seedLive = async () => {
    const found = await props.live.ready();
    if (unmounted) return false;
    if (found) return true;
    const exts = baseExt();
    const seed = toHtml(props.liveSeed);
    if (seed) { const schema = getSchema(exts); const node = PMNode.fromJSON(schema, generateJSON(seed, exts)); props.live.seed(() => prosemirrorToYXmlFragment(node, props.live.doc.getXmlFragment('default'))); }
    return true;
  };
  if (props.live && !(await seedLive())) return;
  const ed = new Editor({
    ...(props.live ? {} : { content: toHtml(props.modelValue) }),
    editable: props.editable,
    extensions: [...baseExt(), ...(props.live ? [Collaboration.configure({ document: props.live.doc, field: 'default' })] : []), Placeholder.configure({ placeholder: 'Start writing, or paste a document here…' })],
    editorProps: { attributes: { 'aria-label': 'Note text', role: 'textbox', 'aria-multiline': 'true', spellcheck: 'true' },
      handlePaste: (v, e) => { const l = imagesOf(e.clipboardData && e.clipboardData.files); if (!l.length) return false; l.forEach((f) => addImage(editor.value, f)); return true; },
      handleDrop: (v, e) => { const l = imagesOf(e.dataTransfer && e.dataTransfer.files); if (!l.length) return false; e.preventDefault(); l.forEach((f) => addImage(editor.value, f)); return true; } },
    onUpdate: ({ editor: e }) => emit('update:modelValue', e.isEmpty ? '' : e.getHTML()),
    onTransaction: () => { tick.value++; hl.value = ed.getAttributes('highlight').color || ''; }
  });
  editor.value = ed;
  // a restored version replaces the text for everyone in the session
  if (props.live && props.modelValue && props.modelValue !== props.liveSeed) ed.commands.setContent(toHtml(props.modelValue));
});
onBeforeUnmount(() => { unmounted = true; if (editor.value) editor.value.destroy(); });

const on = (n) => { tick.value; return !!editor.value && editor.value.isActive(n); };
const attr = (k) => { tick.value; return (editor.value && editor.value.getAttributes('textStyle')[k]) || ''; };
const can = (n) => { tick.value; return !!editor.value && editor.value.can()[n](); };
const block = ref('p');
const keepFocus = (e) => { if (!/^(SELECT|INPUT|OPTION)$/.test(e.target.tagName)) e.preventDefault(); };
const run = (fn) => { fn(editor.value.chain().focus()).run(); };
const setBlock = (v) => { v === 'p' ? run((c) => c.setParagraph()) : run((c) => c.setHeading({ level: +v })); };
const setFont = (v) => run((c) => (v ? c.setFontFamily(v) : c.unsetFontFamily()));
const setSize = (v) => run((c) => (v ? c.setFontSize(v) : c.unsetFontSize()));
const hex = (v, d) => (/^#[0-9a-f]{6}$/i.test(v || '') ? v : d);
const indent = () => run((c) => (editor.value.isActive('listItem') ? c.sinkListItem('listItem') : editor.value.isActive('taskItem') ? c.sinkListItem('taskItem') : c.setBlockquote()));
const outdent = () => run((c) => (editor.value.isActive('listItem') ? c.liftListItem('listItem') : editor.value.isActive('taskItem') ? c.liftListItem('taskItem') : c.unsetBlockquote()));
const pickImage = (e) => { const f = e.target.files && e.target.files[0]; e.target.value = ''; if (f) addImage(editor.value, f); };
// While reading, links work with one click. While writing, Ctrl or Cmd and click opens them.
const onDocClick = (e) => {
  const t = e.target;
  const ex = !props.editable && t.closest && t.closest('[data-expand]');
  if (ex && e.offsetY < 30 && t === ex) { ex.classList.toggle('nr-closed'); return; }
  const ref = t.closest && t.closest('[data-ref]');
  if (ref) { emit('ref', ref.getAttribute('data-ref')); return; }
  const c = t.closest && t.closest('[data-comment]');
  if (c) { emit('comment', c.getAttribute('data-comment')); return; }
  const a = t.closest && t.closest('.ProseMirror a[href]');
  if (a && (!props.editable || e.ctrlKey || e.metaKey) && /^(https?:|mailto:)/i.test(a.getAttribute('href'))) { e.preventDefault(); window.open(a.getAttribute('href'), '_blank', 'noopener,noreferrer'); }
};
const insertMacro = (v) => {
  const ed = editor.value;
  if (['info', 'note', 'warning', 'success'].includes(v)) run((c) => c.insertContent({ type: 'panel', attrs: { kind: v }, content: [{ type: 'paragraph' }] }));
  else if (v === 'expand') { const t = window.prompt('Title of the expandable section', 'More'); if (t) run((c) => c.insertContent({ type: 'expand', attrs: { title: t }, content: [{ type: 'paragraph' }] })); }
  else if (v === 'status') { const t = window.prompt('Status text', 'IN PROGRESS'); const col = t && window.prompt('Colour: grey, blue, green, yellow or red', 'blue'); if (t) run((c) => c.insertContent({ type: 'text', text: t, marks: [{ type: 'status', attrs: { color: ['grey', 'blue', 'green', 'yellow', 'red'].includes(col) ? col : 'blue' } }] }).insertContent(' ')); }
  else if (v === 'board') { const t = window.prompt('Title of the whiteboard or diagram page to show here'); if (t) run((c) => c.insertContent({ type: 'macro', attrs: { kind: 'board', arg: t } })); }
  else if (v) run((c) => c.insertContent({ type: 'macro', attrs: { kind: v, arg: '' } }));
  void ed;
};
const setLink = () => {
  const cur = editor.value.getAttributes('link').href || '';
  const url = window.prompt('Link address (leave empty to remove)', cur);
  if (url === null) return;
  url.trim() ? run((c) => c.extendMarkRange('link').setLink({ href: /^[a-z]+:|^\//i.test(url.trim()) ? url.trim() : 'https://' + url.trim() })) : run((c) => c.extendMarkRange('link').unsetLink());
};
const insertTable = (rows, cols) => { grid.open = false; grid.r = grid.c = 0; run((c) => c.insertTable({ rows, cols, withHeaderRow: true })); };
const addEmoji = (ch) => run((c) => c.insertContent(ch));

// The block dropdown follows the cursor.
const syncBlock = () => { const e = editor.value; if (!e) return; const h = [1, 2, 3, 4].find((l) => e.isActive('heading', { level: l })); block.value = h ? String(h) : 'p'; };
watch(tick, syncBlock);
watch(() => props.editable, (v) => { if (editor.value) editor.value.setEditable(v); });
// A note opened later with different content (e.g. switching notes) is loaded in.
watch(() => props.modelValue, (v) => { const e = editor.value; if (!props.live && e && v !== (e.isEmpty ? '' : e.getHTML())) e.commands.setContent(toHtml(v), { emitUpdate: false }); });
</script>

<style>
.nr-wrap { display: flex; flex-direction: column; gap: 10px; flex: 1; min-width: 0; }
.nr-bar { display: flex; flex-wrap: wrap; gap: 6px 10px; align-items: center; padding: 8px 10px; border-radius: 10px; background: #EFE7C6; border: 1px solid #DDD2A6; position: sticky; top: 8px; z-index: 5; box-shadow: 0 4px 12px rgba(0,0,0,0.25); }
.nr-grp { display: flex; align-items: center; gap: 2px; padding-right: 10px; border-right: 1px solid #D3C78F; flex-wrap: wrap; }
.nr-grp:last-child { border-right: 0; padding-right: 0; }
.nr-tbl { flex-basis: 100%; border-right: 0; border-top: 1px dashed #D3C78F; padding-top: 6px; }
.nr-b { min-width: 30px; height: 30px; padding: 0 7px; border: 0; border-radius: 6px; background: transparent; color: #2A2716; font: 600 13px 'DM Sans', sans-serif; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; position: relative; }
.nr-b:hover:not(:disabled) { background: #E0D49C; }
.nr-b.on { background: #E0B93A; color: #1F1B0B; }
.nr-b:disabled { opacity: 0.35; cursor: default; }
.nr-t { font-weight: 500; font-size: 12px; }
.nr-del { color: #A8351F; }
.nr-col input { position: absolute; inset: 0; opacity: 0; width: 100%; height: 100%; cursor: pointer; }
.nr-s { height: 30px; border-radius: 6px; border: 1px solid #D3C78F; background: #FBF6E3; color: #2A2716; font-size: 12px; max-width: 128px; }
.nr-s-sm { max-width: 62px; }
.nr-pop-wrap { position: relative; }
.nr-emoji button { color: #2A2716 !important; }
.nr-emoji button:hover { background: #E0D49C !important; }
.nr-grid-pop { position: absolute; top: 34px; left: 0; z-index: 10; padding: 8px; background: #FBF6E3; border: 1px solid #D3C78F; border-radius: 8px; box-shadow: 0 8px 20px rgba(0,0,0,0.35); }
.nr-grid { display: grid; grid-template-columns: repeat(8, 18px); gap: 3px; }
.nr-cell { width: 18px; height: 18px; border: 1px solid #C9BC83; border-radius: 2px; background: #fff; cursor: pointer; }
.nr-cell.hot { background: #E0B93A; border-color: #B8921A; }
.nr-grid-lbl { margin-top: 6px; font-size: 12px; color: #5A5434; text-align: center; }

.nr-doc { flex: 1; display: flex; min-width: 0; }
.nr-doc > .ProseMirror { flex: 1; min-height: 380px; outline: none; color: #2A2716; font: 400 16px/1.6 'DM Sans', sans-serif; overflow-wrap: anywhere; min-width: 0; }
.ProseMirror p.is-editor-empty:first-child::before { content: attr(data-placeholder); color: #9A9066; float: left; height: 0; pointer-events: none; }
.nr-doc .ProseMirror p { margin: 0 0 0.6em; }
.nr-doc .ProseMirror h1, .nr-doc .ProseMirror h2, .nr-doc .ProseMirror h3, .nr-doc .ProseMirror h4 { font-family: 'Space Grotesk', sans-serif; line-height: 1.25; margin: 1em 0 0.4em; color: #1F1B0B; }
.nr-doc .ProseMirror h1 { font-size: 1.9em; } .nr-doc .ProseMirror h2 { font-size: 1.5em; } .nr-doc .ProseMirror h3 { font-size: 1.25em; } .nr-doc .ProseMirror h4 { font-size: 1.05em; }
.nr-doc .ProseMirror ul, .nr-doc .ProseMirror ol { padding-left: 1.6em; margin: 0 0 0.6em; }
.nr-doc .ProseMirror li > p { margin: 0; }
.nr-doc .ProseMirror ul[data-type="taskList"] { list-style: none; padding-left: 0.2em; }
.nr-doc .ProseMirror ul[data-type="taskList"] li { display: flex; gap: 8px; align-items: flex-start; }
.nr-doc .ProseMirror ul[data-type="taskList"] li > label { flex: none; margin-top: 0.35em; }
.nr-doc .ProseMirror ul[data-type="taskList"] li > div { flex: 1; }
.nr-doc .ProseMirror blockquote { margin: 0 0 0.8em; padding: 2px 0 2px 14px; border-left: 4px solid #E0B93A; color: #5A5434; }
.nr-doc .ProseMirror hr { border: 0; border-top: 2px solid #E0D49C; margin: 1.2em 0; }
.nr-doc .ProseMirror a { color: #1F6FB5; text-decoration: underline; cursor: pointer; }
.nr-doc .ProseMirror code { background: #EFE7C6; padding: 1px 5px; border-radius: 4px; font-size: 0.9em; }
.nr-doc .ProseMirror pre { background: #2A2716; color: var(--k-F3EBC8,#F3EBC8); padding: 12px 14px; border-radius: 8px; overflow-x: auto; margin: 0 0 0.8em; }
.nr-doc .ProseMirror pre code { background: transparent; padding: 0; color: inherit; }
.nr-doc .ProseMirror mark { border-radius: 2px; padding: 0 1px; }
.nr-doc .ProseMirror img { display: block; max-width: 100%; height: auto; margin: 8px 0; border-radius: 6px; }
.nr-doc .ProseMirror img.ProseMirror-selectednode { outline: 2px solid #7CC4FF; }
.nr-doc .ProseMirror [data-panel] { margin: 8px 0; padding: 8px 12px; border-radius: 8px; border-left: 4px solid #1F6FB5; background: #1F6FB522; }
.nr-doc .ProseMirror [data-panel="warning"] { border-color: #C0392B; background: #C0392B22; }
.nr-doc .ProseMirror [data-panel="note"] { border-color: #E0A800; background: #E0A80022; }
.nr-doc .ProseMirror [data-panel="success"] { border-color: #3C8D2F; background: #3C8D2F22; }
.nr-doc .ProseMirror [data-expand] { margin: 8px 0; padding: 8px 12px; border: 1px solid #8884; border-radius: 8px; }
.nr-doc .ProseMirror [data-expand]::before { content: '▾ ' attr(data-expand); display: block; font-weight: 700; cursor: pointer; }
.nr-doc .ProseMirror [data-expand].nr-closed::before { content: '▸ ' attr(data-expand); }
.nr-doc .ProseMirror [data-expand].nr-closed > * { display: none; }
.nr-doc .ProseMirror [data-macro] { margin: 8px 0; padding: 10px; border: 1px dashed #8886; border-radius: 8px; font-size: 13px; }
.nr-doc .ProseMirror [data-macro]::before { content: 'Macro: ' attr(data-macro) ' ' attr(data-arg) ' (shows when the page is read)'; }
.nr-doc .ProseMirror [data-status] { padding: 1px 6px; border-radius: 4px; font-size: 12px; font-weight: 700; background: #1F6FB5; color: #fff; }
.nr-doc .ProseMirror [data-status="green"] { background: #3C8D2F; } .nr-doc .ProseMirror [data-status="yellow"] { background: #E0A800; color: #111; } .nr-doc .ProseMirror [data-status="red"] { background: #C0392B; } .nr-doc .ProseMirror [data-status="grey"] { background: #777; }
.nr-doc .ProseMirror [data-ref] { color: #1F6FB5; text-decoration: underline; cursor: pointer; }
.nr-doc .ProseMirror [data-comment] { background: #FFD66B66; border-bottom: 2px solid #E0A800; cursor: pointer; }

.nr-doc .tableWrapper { overflow-x: auto; margin: 0.4em 0 1em; max-width: 100%; }
.nr-doc .ProseMirror table { border-collapse: collapse; table-layout: fixed; width: 100%; margin: 0; overflow: hidden; }
.nr-doc .ProseMirror td, .nr-doc .ProseMirror th { position: relative; min-width: 60px; padding: 6px 9px; border: 1px solid #B9AC72; vertical-align: top; box-sizing: border-box; text-align: left; }
.nr-doc .ProseMirror th { background: #EFE7C6; font-weight: 700; }
.nr-doc .ProseMirror td > p, .nr-doc .ProseMirror th > p { margin: 0 0 0.3em; }
.nr-doc .ProseMirror td > *:last-child, .nr-doc .ProseMirror th > *:last-child { margin-bottom: 0; }
.nr-doc .ProseMirror .selectedCell::after { content: ''; position: absolute; inset: 0; background: rgba(224, 185, 58, 0.3); pointer-events: none; }
.nr-doc .ProseMirror .column-resize-handle { position: absolute; right: -2px; top: 0; bottom: -2px; width: 4px; background: #E0B93A; pointer-events: none; }
.nr-doc .ProseMirror.resize-cursor { cursor: col-resize; }
</style>
