<template>
<div :style="toolbarInActions ? 'display: contents' : 'width: 100%; min-width: 0; box-sizing: border-box; display: flex; flex-direction: column; align-items: flex-start; gap: 5px'">
<div :style="toolbarInActions ? 'position: relative; min-width: 0; grid-column: 2; grid-row: 1 / 3; display: flex; align-self: stretch' : 'position: relative; width: 100%; min-width: 0; display: block'"><div v-if="mq && matches.length" role="listbox" aria-label="People you can tag" style="position: absolute; left: 0; bottom: calc(100% + 6px); z-index: 90; min-width: 240px; max-width: 100%; max-height: 240px; overflow-y: auto; padding: 4px; border-radius: 12px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 14px 30px rgba(0,0,0,0.45)">
<button v-for="(m, i) in matches" :key="m.uid" type="button" role="option" :aria-selected="i === mActive" @mousedown.prevent="chooseMention(m)" @mouseenter="mActive = i" :style="`display: flex; align-items: center; gap: 10px; width: 100%; padding: 6px 8px; border: 0; border-radius: 8px; background: ${i === mActive ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left`"><span :style="`flex: none; width: 26px; height: 26px; border-radius: 50%; background: ${m.hash ? '#7CC4FF' : (m.color || '#7CC4FF')}; color: #0F1714; font-size: 10px; font-weight: 700; display: flex; align-items: center; justify-content: center`">{{ m.hash ? '#' : m.initials }}</span><span style="flex: 1; min-width: 0; font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ m.hash ? '#' : '' }}{{ m.name }}</span><span v-if="m.hint" style="flex: none; font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ m.hint }}</span></button>
</div><textarea ref="editor" :value="local" :rows="rows" :aria-label="ariaLabel" :placeholder="placeholder" :disabled="disabled" @input="onInput" @keydown="onKeydown" @keyup="onKeyup" @click="watchMention" @blur="flush" :style="`display: block; width: 100%; min-width: 0; box-sizing: border-box; padding: 10px 12px; border: 1px solid var(--k-2A3832,#2A3832); border-radius: 10px; outline: none; background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font: inherit; resize: ${toolbarInActions ? 'none' : 'vertical'}; ${inputStyle} ${toolbarInActions ? 'min-height: 84px; height: 100%;' : ''}`"></textarea></div>
<div v-if="toolbar" role="toolbar" aria-label="Text formatting" :style="`display: flex; align-items: center; gap: 3px; width: max-content; max-width: 100%; box-sizing: border-box; padding: 2px 5px; border: 1px solid var(--k-2A3832,#2A3832); border-radius: 8px; background: var(--k-121B17,#121B17); overflow-x: auto; ${toolbarInActions ? 'grid-column: 1; grid-row: 2; align-self: start;' : ''}`">
<button v-for="tool in tools" :key="tool.key" type="button" :aria-label="tool.label" :title="tool.label" :disabled="disabled" @mousedown.prevent @click="format(tool.key)" :style="`flex: none; min-width: 32px; height: 30px; padding: 0 7px; border: 0; border-radius: 6px; background: transparent; color: ${disabled ? 'var(--k-66736C,#66736C)' : 'var(--k-E8EFEA,#E8EFEA)'}; cursor: ${disabled ? 'not-allowed' : 'pointer'}; font-weight: 700`">
<svg v-if="tool.key === 'link'" aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.07 0l3-3A5 5 0 0 0 13 2.93l-1.72 1.71"></path><path d="M14 11a5 5 0 0 0-7.07 0l-3 3A5 5 0 0 0 11 21.07l1.71-1.71"></path></svg>
<span v-else>{{ tool.icon }}</span>
</button>
</div>
</div>
</template>

<script>
import { useHubStore } from '~/composables/useHub';
import { hashtagsOf } from '~/logic/richText.js';

const normalizedLink = (value) => {
  const trimmed = value.trim();
  if (!trimmed) return '';
  const candidate = /^[a-z][a-z\d+.-]*:/i.test(trimmed) ? trimmed : 'https://' + trimmed;
  try {
    const url = new URL(candidate);
    return ['http:', 'https:', 'mailto:'].includes(url.protocol) ? url.toString() : '';
  } catch {
    return '';
  }
};

export default {
  name: 'RichTextInput',
  props: {
    modelValue: { type: String, default: '' },
    rows: { type: Number, default: 3 },
    ariaLabel: { type: String, default: 'Formatted text' },
    placeholder: { type: String, default: '' },
    disabled: Boolean,
    inputStyle: { type: String, default: '' },
    toolbarInActions: Boolean,
    toolbar: { type: Boolean, default: true },   // the Bold / Italic / Underline / Link strip (shortcuts still work without it)
    hashtags: Boolean,   // typing # offers the hashtags already used on the wall, and the one being typed
    mentions: { type: [Array, Boolean], default: false }   // an array of { uid, name, initials, color }, or true for everyone in Hub: typing @ offers them
  },
  emits: ['update:modelValue', 'keydown'],
  // Typing stays in this box and reaches the app a moment later (or at once on any key handled above, blur or format).
  // Each update re-renders the whole Hub, so sending one per keystroke made typing lag.
  data() {
    return {
      local: this.modelValue || '',
      sent: this.modelValue || '',
      timer: null,
      mq: null,      // the @word being typed: { start, query }
      mActive: 0,
      tools: [
        { key: 'bold', label: 'Bold (Ctrl/⌘+B)', icon: 'B' },
        { key: 'italic', label: 'Italic (Ctrl/⌘+I)', icon: 'I' },
        { key: 'underline', label: 'Underline (Ctrl/⌘+U)', icon: 'U' },
        { key: 'link', label: 'Insert link', icon: '' }
      ]
    };
  },
  watch: {
    // the app changed the text (cleared after sending, a draft restored): show it, unless it is our own pending value
    modelValue(v) { if ((v || '') !== this.sent) { clearTimeout(this.timer); this.timer = null; this.local = this.sent = v || ''; } }
  },
  computed: {
    matches() {
      if (!this.mq) return [];
      const t = this.mq.query.toLowerCase();
      if (this.mq.kind === '#') return this.tagPool(t);
      return this.pool().filter((m) => !t || m.name.toLowerCase().includes(t))
        .sort((a, b) => (a.name.toLowerCase().startsWith(t) ? 0 : 1) - (b.name.toLowerCase().startsWith(t) ? 0 : 1) || a.name.localeCompare(b.name)).slice(0, 6);
    }
  },
  beforeUnmount() { this.flush(); },
  methods: {
    // who can be tagged: the given list, or everyone already in Hub (except you)
    pool() {
      if (Array.isArray(this.mentions)) return this.mentions;
      if (!this.mentions) return [];
      const l = useHubStore().logic;
      return (l.dir || []).filter((p) => p.hasAccount && p.uid).map((p) => ({ uid: p.uid, name: p.name, initials: p.initials, color: p.color }));
    },
    // hashtags used on the wall (most used first) that start with, or contain, what is typed; plus the typed one as a new tag
    tagPool(t) {
      const counts = new Map();
      ((useHubStore().logic.state || {}).posts || []).forEach((p) => hashtagsOf(p.text).forEach((x) => { const k = x.toLowerCase(); const c = counts.get(k) || { name: x, n: 0 }; c.n++; counts.set(k, c); }));
      const list = [...counts.values()].filter((c) => !t || c.name.toLowerCase().includes(t))
        .sort((a, b) => (a.name.toLowerCase().startsWith(t) ? 0 : 1) - (b.name.toLowerCase().startsWith(t) ? 0 : 1) || b.n - a.n).slice(0, 5)
        .map((c) => ({ uid: '#' + c.name, name: c.name, hash: true, hint: c.n + (c.n === 1 ? ' post' : ' posts') }));
      if (t.length >= 2 && !counts.has(t)) list.push({ uid: '#new', name: this.mq.query, hash: true, hint: 'New hashtag' });
      return list.slice(0, 6);
    },
    // is the caret right after "@something" (or "#something")? then offer the people (or hashtags)
    watchMention() {
      const el = this.$refs.editor;
      if (!el || !(this.mentions || this.hashtags)) { this.mq = null; return; }
      const upto = this.local.slice(0, el.selectionStart);
      const m = this.mentions ? /(^|\s)@([^\s@]{0,30})$/.exec(upto) : null;
      const h = !m && this.hashtags ? /(^|\s)#([\p{L}\p{N}_]{0,40})$/u.exec(upto) : null;
      if (m) { this.mq = { kind: '@', start: upto.length - m[2].length - 1, query: m[2] }; this.mActive = 0; }
      else if (h) { this.mq = { kind: '#', start: upto.length - h[2].length - 1, query: h[2] }; this.mActive = 0; }
      else this.mq = null;
    },
    chooseMention(person) {
      const el = this.$refs.editor;
      if (!el || !this.mq) return;
      const caret = el.selectionStart;
      const insert = (this.mq.kind === '#' ? '#' : '@') + person.name + ' ';
      const pos = this.mq.start + insert.length;
      this.local = this.local.slice(0, this.mq.start) + insert + this.local.slice(caret);
      this.mq = null;
      this.flush();
      this.$nextTick(() => { el.focus(); el.setSelectionRange(pos, pos); });
    },
    onInput(event) {
      this.local = event.target.value;
      this.$nextTick(() => this.watchMention());
      clearTimeout(this.timer);
      this.timer = setTimeout(this.flush, 150);
    },
    flush() {
      clearTimeout(this.timer);
      this.timer = null;
      if (this.local === this.sent) return;
      this.sent = this.local;
      this.$emit('update:modelValue', this.local);
    },
    onKeyup(event) { if (!['ArrowDown', 'ArrowUp', 'Enter', 'Tab', 'Escape'].includes(event.key)) this.watchMention(); },
    onKeydown(event) {
      // while the list of people is open, these keys pick from it instead of sending or moving the caret
      if (this.mq && this.matches.length) {
        if (event.key === 'ArrowDown' || event.key === 'ArrowUp') { event.preventDefault(); this.mActive = (this.mActive + (event.key === 'ArrowDown' ? 1 : -1) + this.matches.length) % this.matches.length; return; }
        if (event.key === 'Enter' || event.key === 'Tab') { event.preventDefault(); event.stopPropagation(); this.chooseMention(this.matches[this.mActive]); return; }
        if (event.key === 'Escape') { event.preventDefault(); this.mq = null; return; }
      }
      // a key the app handles (Enter to send) must see everything typed so far
      if (event.key === 'Enter' || event.ctrlKey || event.metaKey) this.flush();
      const modifier = event.ctrlKey || event.metaKey;
      const key = event.key.toLowerCase();
      const shortcut = modifier && !event.altKey && !event.shiftKey
        ? ({ b: 'bold', i: 'italic', u: 'underline', k: 'link' })[key]
        : null;
      if (shortcut) {
        event.preventDefault();
        event.stopPropagation();
        this.format(shortcut, true);
      }
      this.$emit('keydown', event);
    },
    format(kind, fromShortcut = false) {
      if (this.disabled) return;
      const input = this.$refs.editor;
      if (!input) return;
      const start = input.selectionStart, end = input.selectionEnd;
      this.flush();
      const value = this.local;
      const selected = value.slice(start, end);
      let before = '', after = '', text = selected;
      if (kind === 'bold') { before = '**'; after = '**'; }
      else if (kind === 'italic') { before = '*'; after = '*'; }
      else if (kind === 'underline') { before = '++'; after = '++'; }
      else {
        const rawUrl = typeof window !== 'undefined' ? window.prompt('Enter the link URL') : '';
        const url = rawUrl ? normalizedLink(rawUrl) : '';
        if (!url) return;
        text = '[' + (selected || rawUrl.trim()) + '](' + url + ')';
      }
      const insertion = before + text + after;
      const next = value.slice(0, start) + insertion + value.slice(end);
      this.$emit('update:modelValue', next);
      this.$nextTick(() => {
        input.focus();
        const cursor = fromShortcut && !selected && kind !== 'link'
          ? start + before.length
          : start + insertion.length;
        input.setSelectionRange(cursor, cursor);
      });
    }
  }
};
</script>
