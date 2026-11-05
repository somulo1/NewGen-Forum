<template><span v-html="rich ? richTextHtml(text) : linkTextHtml(text)"></span></template>

<script>
import { richTextHtml, linkTextHtml } from '~/logic/richText.js';

// Text someone wrote, with its links clickable. `rich` also applies the **bold** / *italic* / ++underline++ marks.
export default {
  name: 'LinkText',
  props: { text: { type: [String, Number], default: '' }, rich: Boolean },
  methods: { richTextHtml, linkTextHtml }
};
</script>
