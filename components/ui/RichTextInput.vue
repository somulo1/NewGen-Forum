<template>
<div :style="toolbarInActions ? 'display: contents' : 'width: 100%; min-width: 0; box-sizing: border-box; display: flex; flex-direction: column; align-items: flex-start; gap: 5px'">
<textarea ref="editor" :value="modelValue" :rows="rows" :aria-label="ariaLabel" :placeholder="placeholder" :disabled="disabled" @input="$emit('update:modelValue', $event.target.value)" @keydown="$emit('keydown', $event)" :style="`display: block; width: 100%; min-width: 0; box-sizing: border-box; padding: 10px 12px; border: 1px solid #2A3832; border-radius: 10px; outline: none; background: #121B17; color: #E8EFEA; font: inherit; resize: vertical; ${toolbarInActions ? 'grid-column: 2; grid-row: 1 / 3; min-height: 80px; height: 100%;' : ''} ${inputStyle}`"></textarea>
<div role="toolbar" aria-label="Text formatting" :style="`display: flex; align-items: center; gap: 3px; width: max-content; max-width: 100%; box-sizing: border-box; padding: 2px 5px; border: 1px solid #2A3832; border-radius: 8px; background: #121B17; overflow-x: auto; ${toolbarInActions ? 'grid-column: 1; grid-row: 2; align-self: start;' : ''}`">
<button v-for="tool in tools" :key="tool.key" type="button" :aria-label="tool.label" :title="tool.label" :disabled="disabled" @mousedown.prevent @click="format(tool.key)" :style="`flex: none; min-width: 32px; height: 30px; padding: 0 7px; border: 0; border-radius: 6px; background: transparent; color: ${disabled ? '#66736C' : '#E8EFEA'}; cursor: ${disabled ? 'not-allowed' : 'pointer'}; font-weight: 700`">
<svg v-if="tool.key === 'link'" aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.07 0l3-3A5 5 0 0 0 13 2.93l-1.72 1.71"></path><path d="M14 11a5 5 0 0 0-7.07 0l-3 3A5 5 0 0 0 11 21.07l1.71-1.71"></path></svg>
<span v-else>{{ tool.icon }}</span>
</button>
</div>
</div>
</template>

<script>
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
    toolbarInActions: Boolean
  },
  emits: ['update:modelValue', 'keydown'],
  data() {
    return {
      tools: [
        { key: 'bold', label: 'Bold', icon: 'B' },
        { key: 'italic', label: 'Italic', icon: 'I' },
        { key: 'underline', label: 'Underline', icon: 'U' },
        { key: 'link', label: 'Insert link', icon: '' }
      ]
    };
  },
  methods: {
    format(kind) {
      if (this.disabled) return;
      const input = this.$refs.editor;
      if (!input) return;
      const start = input.selectionStart, end = input.selectionEnd;
      const value = this.modelValue || '';
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
        const cursor = start + insertion.length;
        input.setSelectionRange(cursor, cursor);
      });
    }
  }
};
</script>
