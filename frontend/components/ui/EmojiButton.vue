<!-- An emoji button for any text box. Opens the full emoji library; emits the chosen emoji. -->
<template>
<span style="display: inline-flex; flex: none">
<button ref="btn" type="button" aria-label="Insert emoji" aria-haspopup="dialog" :aria-expanded="open" @click.stop="toggle" :style="`width: ${size}px; height: ${size}px; flex: none; border-radius: ${shape}; border: 0; background: ${open ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; display: flex; align-items: center; justify-content: center`">
<svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="12" cy="12" r="9"></circle><path d="M8.5 14c.8 1.2 2 2 3.5 2s2.7-.8 3.5-2M9 10h.01M15 10h.01"></path></svg>
</button>
<template v-if="open">
<div ref="pop" role="dialog" aria-label="Emoji picker" @click.stop :style="`position: fixed; z-index: 120; left: ${pos.left}px; bottom: ${pos.bottom}px; width: ${pos.width}px`"><EmojiPicker @pick="onPick"/></div>
</template>
</span>
</template>

<script>
export default {
  name: 'EmojiButton',
  props: { size: { type: Number, default: 38 }, shape: { type: String, default: '50%' } },
  emits: ['pick'],
  data() { return { open: false, pos: { left: 8, bottom: 80, width: 320 } }; },
  mounted() { this._close = (e) => { if (this.open && !(this.$refs.pop && this.$refs.pop.contains(e.target))) this.open = false; }; this._esc = (e) => { if (e.key === 'Escape') this.open = false; }; document.addEventListener('click', this._close); document.addEventListener('keydown', this._esc); },
  beforeUnmount() { document.removeEventListener('click', this._close); document.removeEventListener('keydown', this._esc); },
  methods: {
    toggle() {
      if (!this.open) {
        const r = this.$refs.btn.getBoundingClientRect(), w = Math.min(340, window.innerWidth - 16);
        const fitsAbove = r.top > 340;
        this.pos = { left: Math.max(8, Math.min(window.innerWidth - w - 8, r.left)), bottom: fitsAbove ? window.innerHeight - r.top + 6 : Math.max(8, window.innerHeight - 340 - 8), width: w };
      }
      this.open = !this.open;
    },
    onPick(ch) { this.$emit('pick', ch); }
  }
};
</script>
