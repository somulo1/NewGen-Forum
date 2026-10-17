<!-- A one-line text box with emoji picker. Used for new comments and for replies. -->
<template>
<div style="position: relative; display: flex; gap: 6px; align-items: center; min-width: 0">
<EmojiButton :size="38" @pick="add"/>
<input ref="inp" v-model="text" type="text" :placeholder="placeholder" :aria-label="placeholder" @keydown.enter.prevent="send" @keydown.esc="$emit('cancel')" style="flex: 1; min-width: 0; height: 42px; padding: 0 14px; border-radius: 21px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px">
<button type="button" :disabled="!text.trim()" @click="send" :aria-label="sendLabel" :title="sendLabel" :style="`width: 42px; height: 42px; flex: none; padding: 0; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: ${text.trim() ? 'pointer' : 'default'}; display: flex; align-items: center; justify-content: center`"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M4 12l16-8-6 16-3-7z"></path></svg></button>
<button v-if="cancellable" type="button" aria-label="Cancel reply" @click="$emit('cancel')" style="height: 34px; padding: 0 10px; border-radius: 17px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">Cancel</button>
</div>
</template>

<script>
export default {
  name: 'CommentBox',
  props: { placeholder: { type: String, default: 'Write a comment…' }, sendLabel: { type: String, default: 'Send' }, cancellable: Boolean, autofocus: Boolean },
  emits: ['send', 'cancel'],
  data() { return { text: '' }; },
  mounted() { if (this.autofocus) this.$refs.inp.focus(); },
  methods: {
    add(e) { this.text += e; this.$refs.inp.focus(); },
    send() { const v = this.text.trim(); if (!v) return; this.$emit('send', v); this.text = ''; }
  }
};
</script>
