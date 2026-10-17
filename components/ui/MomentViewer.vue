<!-- A moment opened wide: photo, video or text. Double tap sends love; like and love buttons; the owner sees who reacted. -->
<template>
<div v-if="momentView.open" class="hub-overlay" role="dialog" aria-modal="true" aria-label="Moment" tabindex="-1" ref="root" @keydown="onKey" @click.self="momentView.close" style="position: fixed; inset: 0; z-index: 70; background: rgba(5,10,8,0.85); display: flex; align-items: center; justify-content: center; padding: 16px; outline: none">
<div style="position: relative; width: min(560px, 100%); height: min(680px, 100%); display: flex; flex-direction: column; border-radius: 22px; overflow: hidden; background: #131D19; border: 1px solid #2A3832">
<div v-if="momentView.total > 1" role="progressbar" :aria-valuenow="Math.round(progress * 100)" aria-valuemin="0" aria-valuemax="100" aria-label="Moment progress" style="flex: none; display: flex; gap: 4px; padding: 10px 14px 0">
<template v-for="n in momentView.total" :key="n"><span style="flex: 1; height: 3px; border-radius: 2px; background: rgba(255,255,255,0.22); overflow: hidden"><span :style="`display: block; height: 100%; background: #fff; width: ${n - 1 < momentView.index ? 100 : n - 1 === momentView.index ? Math.round(progress * 100) : 0}%`"></span></span></template>
</div>
<div style="flex: none; display: flex; align-items: center; gap: 10px; padding: 12px 14px">
<UserAvatar :name="item.owner" :initials="item.initials" :color="item.color" :size="38"/>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ item.owner === 'Sam' ? 'Your moment' : item.owner }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ momentView.counter }}</span></div>
<button type="button" aria-label="Close moment" @click="momentView.close" style="width: 36px; height: 36px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">✕</button>
</div>
<div @dblclick="onDouble" @pointerdown="hold = true" @pointerup="hold = false" @pointerleave="hold = false" @pointercancel="hold = false" title="Double tap to send love" :style="`position: relative; flex: 1; min-height: 0; display: flex; align-items: center; justify-content: center; background: ${item.bg}; user-select: none; touch-action: manipulation`">
<img v-if="item.mediaIsImg" :src="item.mediaUrl" :alt="item.caption || 'Moment photo'" draggable="false" style="width: 100%; height: 100%; object-fit: contain; background: #0F1714">
<video v-else-if="item.mediaIsVideo" ref="vid" :key="item.mediaUrl" :src="item.mediaUrl" controls autoplay playsinline @timeupdate="onTime" @ended="advance" style="width: 100%; height: 100%; object-fit: contain; background: #000"></video>
<strong v-else style="padding: 28px; font-family: 'Space Grotesk', sans-serif; font-size: 26px; line-height: 1.3; text-align: center; color: #fff; overflow-wrap: anywhere">{{ item.caption }}</strong>
<span v-if="burst" aria-hidden="true" class="mv-burst"><svg width="96" height="96" viewBox="0 0 24 24" fill="#FF6B81" stroke="#FF6B81" stroke-width="1.5" stroke-linejoin="round"><path d="M12 20s-7-4.4-7-10a4 4 0 0 1 7-2.6A4 4 0 0 1 19 10c0 5.6-7 10-7 10z"></path></svg></span>
<button type="button" aria-label="Previous moment" @click.stop="momentView.prev" style="position: absolute; left: 8px; top: 50%; transform: translateY(-50%); width: 38px; height: 38px; border-radius: 50%; border: 1px solid rgba(255,255,255,.35); background: rgba(15,23,20,.6); color: #fff; cursor: pointer; font-size: 18px">‹</button>
<button type="button" aria-label="Next moment" @click.stop="momentView.next" style="position: absolute; right: 8px; top: 50%; transform: translateY(-50%); width: 38px; height: 38px; border-radius: 50%; border: 1px solid rgba(255,255,255,.35); background: rgba(15,23,20,.6); color: #fff; cursor: pointer; font-size: 18px">›</button>
</div>
<p v-if="item.caption && (item.mediaIsImg || item.mediaIsVideo)" style="flex: none; margin: 0; padding: 10px 16px 0; font-size: 14px; line-height: 1.45; overflow-wrap: anywhere">{{ item.caption }}</p>
<div style="flex: none; display: flex; flex-wrap: wrap; align-items: center; gap: 8px; padding: 10px 14px">
<button type="button" @click="momentView.like" :aria-pressed="momentView.iLiked" :style="pill(momentView.iLiked, '#7CC4FF')"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" :fill="momentView.iLiked ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M7 11v9H4v-9zM7 11l4-7a2 2 0 0 1 2 2l-.5 3H19a2 2 0 0 1 2 2.3l-1.2 6A2 2 0 0 1 17.8 20H7"></path></svg>Like · {{ item.likeCount }}</button>
<button type="button" @click="momentView.love" :aria-pressed="momentView.iLoved" :style="pill(momentView.iLoved, '#FF6B81')"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" :fill="momentView.iLoved ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M12 20s-7-4.4-7-10a4 4 0 0 1 7-2.6A4 4 0 0 1 19 10c0 5.6-7 10-7 10z"></path></svg>Love · {{ item.loveCount }}</button>
<button v-if="momentView.mine" type="button" @click="momentView.toggleWho" :aria-expanded="momentView.whoOpen" style="margin-left: auto; height: 34px; padding: 0 14px; border-radius: 17px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">{{ momentView.whoOpen ? 'Hide' : 'Who reacted' }}</button>
<span v-else style="margin-left: auto; font-size: 12px; color: #9DB0A6">Double tap to send love</span>
</div>
<div v-if="momentView.mine && momentView.whoOpen" role="list" aria-label="People who reacted" style="flex: none; max-height: 150px; overflow-y: auto; padding: 0 14px 14px; display: flex; flex-direction: column; gap: 6px">
<span v-if="!momentView.hasWho" style="font-size: 13px; color: #9DB0A6">No reactions yet.</span>
<template v-for="(w, i) in momentView.who" :key="i"><div role="listitem" style="display: flex; align-items: center; gap: 8px; font-size: 14px"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" :fill="w.kind === 'Loved' ? '#FF6B81' : 'none'" :stroke="w.kind === 'Loved' ? '#FF6B81' : '#7CC4FF'" stroke-width="2" stroke-linejoin="round"><path v-if="w.kind === 'Loved'" d="M12 20s-7-4.4-7-10a4 4 0 0 1 7-2.6A4 4 0 0 1 19 10c0 5.6-7 10-7 10z"></path><path v-else d="M7 11v9H4v-9zM7 11l4-7a2 2 0 0 1 2 2l-.5 3H19a2 2 0 0 1 2 2.3l-1.2 6A2 2 0 0 1 17.8 20H7"></path></svg><span style="flex: 1">{{ w.name }}</span><span style="font-size: 12px; color: #9DB0A6">{{ w.kind }}</span></div></template>
</div>
</div>
</div>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'MomentViewer',
  data() { return { burst: false, progress: 0, hold: false }; },
  computed: { ...hubBindings(['momentView']).computed, item() { return this.momentView.item || {}; }, paused() { return this.hold || this.momentView.whoOpen; } },
  watch: {
    'momentView.open'(v) { if (v) { this.$nextTick(() => this.$refs.root && this.$refs.root.focus()); this.start(); } else this.stop(); },
    'momentView.index'() { this.progress = 0; this.last = 0; },
    paused(p) { const v = this.$refs.vid; if (v) { if (p) v.pause(); else v.play().catch(() => {}); } }
  },
  beforeUnmount() { this.stop(); },
  methods: {
    // Like a status: each moment plays for a few seconds (a video plays to its end), then the next one starts. Hold to pause.
    start() { this.stop(); this.progress = 0; this.last = 0; const tick = (now) => { if (!this.momentView.open) return; const dt = this.last ? now - this.last : 0; this.last = now; const isVid = this.item.mediaIsVideo; if (!this.paused && !isVid && !document.hidden) { this.progress = Math.min(1, this.progress + dt / 5000); if (this.progress >= 1) { this.advance(); } } this._raf = requestAnimationFrame(tick); }; this._raf = requestAnimationFrame(tick); },
    stop() { if (this._raf) cancelAnimationFrame(this._raf); this._raf = null; },
    advance() { this.progress = 0; this.last = 0; this.momentView.next(); },
    onTime(e) { const v = e.target; if (v.duration) this.progress = v.currentTime / v.duration; },
    pill(on, c) { return `height: 34px; padding: 0 14px; border-radius: 17px; display: flex; align-items: center; gap: 6px; cursor: pointer; font-size: 13px; font-weight: 600; border: 1px solid ${on ? c : '#2A3832'}; background: ${on ? c + '22' : 'transparent'}; color: ${on ? c : '#E8EFEA'}`; },
    onDouble() { this.momentView.loveOn(); this.burst = false; this.$nextTick(() => { this.burst = true; clearTimeout(this._b); this._b = setTimeout(() => { this.burst = false; }, 800); }); },
    onKey(e) { if (e.key === 'Escape') this.momentView.close(); else if (e.key === 'ArrowRight') this.momentView.next(); else if (e.key === 'ArrowLeft') this.momentView.prev(); }
  }
};
</script>

<style scoped>
.mv-burst { position: absolute; left: 50%; top: 50%; transform: translate(-50%, -50%); pointer-events: none; animation: mvpop .8s ease-out forwards; filter: drop-shadow(0 6px 18px rgba(0,0,0,.45)); }
@keyframes mvpop { 0% { transform: translate(-50%, -50%) scale(.3); opacity: 0 } 30% { transform: translate(-50%, -50%) scale(1.15); opacity: 1 } 100% { transform: translate(-50%, -80%) scale(1); opacity: 0 } }
</style>
