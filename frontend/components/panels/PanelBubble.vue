<!-- Minimized forum / chat bubble with its menu -->
<template><template v-if="isMin">
<div class="hub-bubble" style="position: fixed; right: 20px; bottom: 20px; z-index: 50; display: flex; flex-direction: column; align-items: flex-end; gap: 10px">
<template v-if="menuOpen">
<div role="menu" style="width: 250px; padding: 8px; border-radius: 14px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 16px 40px rgba(0,0,0,0.45); display: flex; flex-direction: column; gap: 2px">
<span class="hub-d" style="padding: 8px 10px 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Open {{ cur.name }} in…</span>
<button class="hub-d" role="menuitem" @click="toDocked" style="display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 0 10px; border-radius: 9px; border: 0; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left; font-size: 14px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M14 4v16"></path><rect x="14" y="4" width="7" height="16" rx="1" fill="currentColor" stroke="none" opacity="0.55"></rect></svg>Side panel</button>
<button class="hub-d" role="menuitem" @click="toPage" style="display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 0 10px; border-radius: 9px; border: 0; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left; font-size: 14px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M7 12h10M9.5 9.5L7 12l2.5 2.5M14.5 9.5L17 12l-2.5 2.5"></path></svg>Fit to page</button>
<button class="hub-m" role="menuitem" @click="toPage" style="display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 0 10px; border-radius: 9px; border: 0; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 14px; text-align: left"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M14 4h6v6M20 4l-8 8M18 14v5a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V7a1 1 0 0 1 1-1h5"></path></svg>Open {{ cur.name }}</button>
<button role="menuitem" @click="toClosed" style="display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 0 10px; border-radius: 9px; border: 0; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left; font-size: 14px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg>Close</button>
</div>
</template>
<div :class="{ 'hub-live-glow': !!curLive }" :style="`display: flex; align-items: center; gap: 10px; padding: 8px 8px 8px 10px; border-radius: 30px; background: var(--k-1D2A25,#1D2A25); border: 1px solid ${curLive ? '#B8F36B' : 'var(--k-34453D,#34453D)'}; box-shadow: 0 12px 30px rgba(0,0,0,0.4)`">
<span :style="`width: 36px; height: 36px; border-radius: 50%; background: ${cur.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 13px`">{{ cur.initials }}</span>
<div style="display: flex; flex-direction: column; padding-right: 4px"><strong style="font-size: 13px">{{ cur.name }}</strong><span style="font-size: 11px; color: var(--k-B8F36B,#B8F36B)">● {{ curLive ? curLive.text : cur.unread ? cur.unread + (cur.unread === 1 ? ' new message' : ' new messages') : 'Minimized' }}</span></div>
<button v-if="curLive" type="button" class="hub-live-glow" @click="curLive.join" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; cursor: pointer">Join</button>
<button aria-label="Forum view options" aria-haspopup="menu" @click="toggleMenu" style="width: 44px; height: 44px; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><circle cx="5" cy="12" r="2"></circle><circle cx="12" cy="12" r="2"></circle><circle cx="19" cy="12" r="2"></circle></svg></button>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'PanelBubble',
  ...hubBindings(["curLive","cur","isMin","menuOpen","toClosed","toDocked","toFull","toPage","toggleMenu"])
};
</script>
