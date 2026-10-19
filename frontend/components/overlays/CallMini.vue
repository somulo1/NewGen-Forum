<!-- Small floating card for a call that is minimized or whose panel is not on screen -->
<template><template v-if="showCallMini">
<div class="hub-minicall" role="region" aria-label="Call in progress" :style="callMiniStyle + '; width: 300px; max-width: calc(100vw - 28px); box-sizing: border-box; padding: 10px 10px 10px 12px; border-radius: 22px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 12px 30px rgba(0,0,0,0.4); display: flex; align-items: center; gap: 10px'">
<span :style="`width: 36px; height: 36px; flex: none; border-radius: 10px; background: ${call.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 13px`">{{ call.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 13px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ call.title }}</strong><span style="font-size: 11px; color: var(--k-B8F36B,#B8F36B)">● {{ call.timer }} · {{ call.modeLabel }}</span></div>
<button :aria-label="call.micLabel" :title="call.micLabel" @click="toggleMic" :style="`width: 34px; height: 34px; flex: none; border-radius: 50%; border: 0; cursor: pointer; display: flex; align-items: center; justify-content: center; background: ${call.micLabel === 'Unmute' ? '#FF9B8A' : 'var(--k-2A3832,#2A3832)'}; color: ${call.micLabel === 'Unmute' ? '#0F1714' : 'var(--k-E8EFEA,#E8EFEA)'}`"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round"><rect x="9" y="3" width="6" height="11" rx="3"></rect><path d="M5 11a7 7 0 0 0 14 0M12 18v3"></path></svg></button>
<button aria-label="Back to call" title="Back to call" @click="restoreCall" style="width: 34px; height: 34px; flex: none; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"></path></svg></button>
<button aria-label="Leave call" title="Leave" @click="endCall" style="width: 34px; height: 34px; flex: none; border-radius: 50%; border: 0; background: #E5484D; color: #fff; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M3 14.5c5-4.5 13-4.5 18 0l-2 2.5-3.5-1.5V13a12 12 0 0 0-7 0v2.5L5 17z"></path></svg></button>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'CallMini',
  ...hubBindings(["call","callMiniStyle","endCall","restoreCall","showCallMini","toggleMic"])
};
</script>
