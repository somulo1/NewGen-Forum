<!-- The card that confirms anything destructive (see logic/confirm.js). -->
<template>
<div v-if="c.open" class="hub-overlay" style="position: fixed; inset: 0; z-index: 300; background: rgba(5,10,8,0.6); display: flex; align-items: center; justify-content: center; padding: 16px" @click.self="answer(false)" @keydown.esc.stop="answer(false)">
<div ref="card" role="alertdialog" aria-modal="true" :aria-label="c.title" tabindex="-1" style="width: 440px; max-width: 100%; max-height: 100%; overflow-y: auto; box-sizing: border-box; display: flex; flex-direction: column; gap: 14px; padding: 20px; border-radius: 20px; background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); border: 1px solid var(--k-2A3832,#2A3832); box-shadow: 0 16px 48px rgba(0,0,0,0.35)">
<h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 20px; overflow-wrap: anywhere">{{ c.title }}</h2>
<p v-if="c.body" style="margin: 0; font-size: 14px; line-height: 1.5; color: var(--k-C9D6CF,#C9D6CF); overflow-wrap: anywhere">{{ c.body }}</p>
<label v-if="c.typed" style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Type {{ c.typed }} to confirm<input ref="field" v-model="c.value" autocomplete="off" spellcheck="false" @keydown.enter="ready && answer(true)" style="height: 40px; padding: 0 12px; box-sizing: border-box; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<div style="display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px">
<button ref="cancel" type="button" @click="answer(false)" style="min-height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; cursor: pointer">Cancel</button>
<button type="button" :disabled="!ready" @click="answer(true)" :style="`min-height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600; cursor: ${ready ? 'pointer' : 'not-allowed'}; opacity: ${ready ? 1 : 0.5}`">{{ c.action }}</button>
</div>
</div>
</div>
</template>
<script>
import { confirmState, answerConfirm } from '~/logic/confirm.js';
export default {
  data() { return { c: confirmState }; },
  computed: { ready() { return !this.c.typed || this.c.value.trim() === this.c.typed; } },
  watch: {
    'c.open'(on) {
      if (on) { this.back = document.activeElement; this.$nextTick(() => { (this.$refs.field || this.$refs.cancel)?.focus(); }); }
      else if (this.back && this.back.focus) { this.back.focus(); this.back = null; }
    }
  },
  mounted() { this.onKey = (e) => { if (e.key === 'Escape' && this.c.open) this.answer(false); }; window.addEventListener('keydown', this.onKey); },
  beforeUnmount() { window.removeEventListener('keydown', this.onKey); if (this.c.open) answerConfirm(false); },
  methods: { answer(ok) { answerConfirm(ok); } }
};
</script>
