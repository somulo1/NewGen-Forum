<!-- Double-click (or double-tap) what is inside to like the post, with hearts rising where it was touched. Taps that follow keep the
     hearts coming and never undo the like: only the Like button takes a like back. The server keeps one like per person. -->
<template>
<div style="display: contents" @click.capture="onClick" @mousedown.capture="noSelect">
<slot />
<svg v-for="h in hearts" :key="h.id" class="lb-heart" aria-hidden="true" width="64" height="64" viewBox="0 0 24 24" fill="#FF4D6D" :style="`left: ${h.x - 32}px; top: ${h.y - 32}px; --lb-tilt: ${h.tilt}deg; --lb-drift: ${h.drift}px`"><path d="M12 21s-7.5-4.7-7.5-10.5a4.3 4.3 0 0 1 7.5-2.8 4.3 4.3 0 0 1 7.5 2.8C19.5 16.3 12 21 12 21z"/></svg>
</div>
</template>

<script setup>
import { ref, onBeforeUnmount } from 'vue';

const emit = defineEmits(['like']);
const hearts = ref([]);
const DOUBLE = 300, KEEP = 700;
let last = 0, burstUntil = 0, wait = null, seq = 0, replay = false;
const timers = new Set();

const skip = (t) => !(t instanceof Element) || !!t.closest('a, input, textarea, select, video, [data-noburst]');
const noSelect = (e) => { if (e.detail > 1 && !skip(e.target)) e.preventDefault(); };

const spawn = (e) => {
  const h = { id: ++seq, x: e.clientX || 0, y: e.clientY || 0, tilt: Math.round(Math.random() * 40 - 20), drift: Math.round(Math.random() * 60 - 30) };
  hearts.value = hearts.value.concat([h]).slice(-12);
  const t = setTimeout(() => { hearts.value = hearts.value.filter((x) => x.id !== h.id); timers.delete(t); }, 900);
  timers.add(t);
};

// a single tap on a picture is held back a moment (it opens the viewer) so that a second tap can turn it into a like instead
const onClick = (e) => {
  if (replay || skip(e.target)) return;
  const now = Date.now();
  if (now < burstUntil || (wait && now - last < DOUBLE)) {
    e.stopPropagation(); e.preventDefault();
    clearTimeout(wait); wait = null;
    if (now >= burstUntil) emit('like');
    burstUntil = now + KEEP; last = now;
    spawn(e);
    return;
  }
  last = now;
  const button = e.target.closest('button');
  e.stopPropagation(); e.preventDefault();
  const target = button || e.target;
  wait = setTimeout(() => { wait = null; replay = true; try { target.click(); } finally { replay = false; } }, DOUBLE - 20);
};

onBeforeUnmount(() => { clearTimeout(wait); timers.forEach(clearTimeout); });
</script>

<style>
.lb-heart { position: fixed; z-index: 400; pointer-events: none; filter: drop-shadow(0 4px 10px rgba(0, 0, 0, 0.4)); animation: lb-rise 0.9s ease-out forwards; }
@keyframes lb-rise {
  0% { opacity: 0; transform: translate(0, 10px) scale(0.3) rotate(var(--lb-tilt)); }
  18% { opacity: 1; transform: translate(0, 0) scale(1.15) rotate(var(--lb-tilt)); }
  100% { opacity: 0; transform: translate(var(--lb-drift), -130px) scale(1) rotate(var(--lb-tilt)); }
}
@media (prefers-reduced-motion: reduce) { .lb-heart { animation: lb-fade 0.9s ease-out forwards; } @keyframes lb-fade { 0%, 60% { opacity: 1; } 100% { opacity: 0; } } }
</style>
