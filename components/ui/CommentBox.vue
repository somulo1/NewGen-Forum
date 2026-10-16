<!-- A one-line text box with emoji picker. Used for new comments and for replies. -->
<template>
<div style="position: relative; display: flex; gap: 6px; align-items: center; min-width: 0">
<template v-if="emojiOpen">
<div role="menu" aria-label="Emoji" style="position: absolute; left: 0; bottom: 44px; z-index: 6; display: grid; grid-template-columns: repeat(5, 1fr); gap: 2px; padding: 6px; border-radius: 12px; background: #1D2A25; border: 1px solid #34453D; box-shadow: 0 12px 26px rgba(0,0,0,.45)">
<template v-for="e in emojis" :key="e"><button type="button" :aria-label="`Insert ${e}`" @click="add(e)" style="width: 34px; height: 34px; border: 0; border-radius: 8px; background: transparent; font-size: 18px; cursor: pointer">{{ e }}</button></template>
</div>
</template>
<button type="button" aria-label="Insert emoji" :aria-expanded="emojiOpen" @click="emojiOpen = !emojiOpen" style="width: 34px; height: 34px; flex: none; border-radius: 50%; border: 0; background: transparent; font-size: 17px; cursor: pointer">😊</button>
<input ref="inp" v-model="text" type="text" :placeholder="placeholder" :aria-label="placeholder" @keydown.enter.prevent="send" @keydown.esc="$emit('cancel')" style="flex: 1; min-width: 0; height: 36px; padding: 0 14px; border-radius: 18px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px">
<button type="button" :disabled="!text.trim()" @click="send" aria-label="Send" :style="`height: 34px; padding: 0 14px; border-radius: 17px; border: 0; font-weight: 600; font-size: 13px; cursor: ${text.trim() ? 'pointer' : 'default'}; background: ${text.trim() ? '#B8F36B' : '#2A3832'}; color: ${text.trim() ? '#0F1714' : '#9DB0A6'}`">{{ sendLabel }}</button>
<button v-if="cancellable" type="button" aria-label="Cancel reply" @click="$emit('cancel')" style="height: 34px; padding: 0 10px; border-radius: 17px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">Cancel</button>
</div>
</template>

<script>
export default {
  name: 'CommentBox',
  props: { placeholder: { type: String, default: 'Write a comment…' }, sendLabel: { type: String, default: 'Send' }, cancellable: Boolean, autofocus: Boolean },
  emits: ['send', 'cancel'],
  data() { return { text: '', emojiOpen: false, emojis: ['😀', '😂', '😍', '👍', '👏', '🎉', '🙏', '❤️', '😮', '😢'] }; },
  mounted() { if (this.autofocus) this.$refs.inp.focus(); },
  methods: {
    add(e) { this.text += e; this.emojiOpen = false; this.$refs.inp.focus(); },
    send() { const v = this.text.trim(); if (!v) return; this.$emit('send', v); this.text = ''; this.emojiOpen = false; }
  }
};
</script>
