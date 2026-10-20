<!-- One participant's live video in a call tile. The stream lives in logic/server/calls.js (outside reactive state). -->
<template>
<video ref="v" autoplay playsinline muted :style="`position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover; background: #000; ${mirror ? 'transform: scaleX(-1)' : ''}`"></video>
</template>

<script>
import { streamFor, trackFor, onMedia } from '~/logic/server/calls.js';

export default {
  name: 'CallVideo',
  props: { streamKey: { type: String, required: true } },
  computed: { mirror() { return this.streamKey === 'local'; } },
  mounted() { this.attach(); this.off = onMedia(() => this.attach()); },
  beforeUnmount() { if (this.off) this.off(); this.release(); },
  watch: { streamKey() { this.attach(); } },
  methods: {
    release() { if (this.cur && this.$refs.v) { try { this.cur.detach(this.$refs.v); } catch (e) { /* gone */ } } this.cur = null; },
    attach() {
      const v = this.$refs.v;
      // a big room: a media-server track, attached so the server sends the quality that fits this tile
      const t = trackFor(this.streamKey);
      if (t) {
        if (this.cur !== t) { this.release(); t.attach(v); this.cur = t; }
        return;
      }
      this.release();
      const s = streamFor(this.streamKey);
      if (v && v.srcObject !== s) { v.srcObject = s; if (s) v.play().catch(() => {}); }
    }
  }
};
</script>
