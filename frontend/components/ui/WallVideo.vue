<!-- A video in a wall post. It plays by itself while most of it is on screen and stops once scrolled past; one plays at a time.
     Sound is one choice for the whole wall: unmute one and the next ones play with sound too, mute one and they all stay muted.
     Click the video to pause or play it; the corner button opens it full size. -->
<template>
<div ref="box" class="wv" @click.stop.prevent="toggle">
<video ref="v" class="wv-v" :src="src" :muted="sound.muted" playsinline loop preload="metadata" :aria-label="name" :style="maxH ? `max-height: ${maxH}px` : ''" @play="playing = true" @pause="playing = false"></video>
<span v-if="!playing" class="wv-play" aria-hidden="true">▶</span>
<button type="button" class="wv-btn wv-mute" :aria-label="sound.muted ? 'Turn sound on' : 'Mute'" :title="sound.muted ? 'Turn sound on' : 'Mute'" @click.stop.prevent="toggleSound">
<svg v-if="sound.muted" aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 5 6 9H3v6h3l5 4zM22 9l-6 6M16 9l6 6"/></svg>
<svg v-else aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 5 6 9H3v6h3l5 4zM15.5 8.5a5 5 0 0 1 0 7M19 5a10 10 0 0 1 0 14"/></svg>
</button>
<button type="button" class="wv-btn wv-open" aria-label="Open full size" title="Open full size" @click.stop.prevent="$emit('open')"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h6v6M9 21H3v-6M21 3l-7 7M3 21l7-7"/></svg></button>
</div>
</template>

<script>
import { reactive } from 'vue';

// shared by every wall video, and remembered in this browser
const sound = reactive({ muted: (() => { try { return localStorage.getItem('hub-wall-sound') !== 'on'; } catch (e) { return true; } })() });
const all = new Set();
let active = null;

export default {
  name: 'WallVideo',
  props: { src: { type: String, default: '' }, name: { type: String, default: 'Video' }, maxH: { type: Number, default: 0 } },
  emits: ['open'],
  data() { return { sound, playing: false, seen: false, held: false }; },
  mounted() {
    all.add(this);
    this.io = new IntersectionObserver(([e]) => {
      this.seen = e.isIntersecting && e.intersectionRatio >= 0.6;
      if (this.seen) { if (!this.held) this.start(); } else { this.held = false; this.stop(); }   // scrolled past: a pause by hand is forgotten
    }, { threshold: [0, 0.6, 1] });
    this.io.observe(this.$refs.box);
    this.onHide = () => { if (document.hidden) this.stop(); else if (this.seen && !this.held) this.start(); };
    document.addEventListener('visibilitychange', this.onHide);
  },
  beforeUnmount() { all.delete(this); if (active === this) active = null; this.io && this.io.disconnect(); document.removeEventListener('visibilitychange', this.onHide); },
  methods: {
    start() {
      const v = this.$refs.v; if (!v || document.hidden) return;
      all.forEach((o) => { if (o !== this) o.stop(); });   // one at a time
      active = this;
      v.muted = sound.muted;
      v.play().catch(() => {   // the browser refused sound before the first click on the page: play muted instead
        if (!v.muted) { v.muted = true; sound.muted = true; v.play().catch(() => {}); }
      });
    },
    stop() { const v = this.$refs.v; if (v && !v.paused) v.pause(); if (active === this) active = null; },
    toggle() { if (this.playing) { this.held = true; this.stop(); } else { this.held = false; this.start(); } },
    toggleSound() {
      sound.muted = !sound.muted;
      try { localStorage.setItem('hub-wall-sound', sound.muted ? 'off' : 'on'); } catch (e) { /* only this visit */ }
      all.forEach((o) => { if (o.$refs.v) o.$refs.v.muted = sound.muted; });
      if (!this.playing && this.seen && !this.held) this.start();   // never restarts one paused by hand
    }
  }
};
</script>

<style scoped>
.wv { position: relative; display: flex; align-items: center; justify-content: center; width: 100%; height: 100%; cursor: pointer; }
.wv-v { position: relative; display: block; max-width: 100%; width: 100%; height: 100%; object-fit: contain; background: transparent; }
.wv-play { position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%); width: 46px; height: 46px; border-radius: 50%; background: rgba(15, 23, 20, 0.72); color: #fff; display: flex; align-items: center; justify-content: center; font-size: 18px; pointer-events: none; }
.wv-btn { position: absolute; width: 34px; height: 34px; border-radius: 50%; border: 0; background: rgba(15, 23, 20, 0.72); color: #fff; display: flex; align-items: center; justify-content: center; cursor: pointer; }
.wv-mute { right: 10px; bottom: 10px; }
.wv-open { right: 10px; top: 10px; }
</style>
