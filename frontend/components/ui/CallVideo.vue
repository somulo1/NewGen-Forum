<!-- One participant's live video in a call tile. The stream lives in logic/server/calls.js (outside reactive state). -->
<template>
<video ref="v" autoplay playsinline muted :style="`position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; background: #000; ${mirror ? 'transform: scaleX(-1)' : ''}`"></video>
</template>

<script>
import { streamFor, onMedia } from '~/logic/server/calls.js';

export default {
  name: 'CallVideo',
  props: { streamKey: { type: String, required: true } },
  computed: { mirror() { return this.streamKey === 'local'; } },
  mounted() { this.attach(); this.off = onMedia(() => this.attach()); },
  beforeUnmount() { if (this.off) this.off(); },
  watch: { streamKey() { this.attach(); } },
  methods: {
    attach() {
      const s = streamFor(this.streamKey);
      const v = this.$refs.v;
      if (v && v.srcObject !== s) { v.srcObject = s; if (s) v.play().catch(() => {}); }
    }
  }
};
</script>
