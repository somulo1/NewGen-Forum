<!-- Minimized forum / chat bubble with its menu -->
<template><template v-if="isMin">
<div class="hub-bubble" style="position: fixed; right: 20px; bottom: 20px; z-index: 50; display: flex; flex-direction: column; align-items: flex-end; gap: 10px">
<template v-if="menuOpen">
<div role="menu" style="width: 250px; padding: 8px; border-radius: 14px; background: #1D2A25; border: 1px solid #34453D; box-shadow: 0 16px 40px rgba(0,0,0,0.45); display: flex; flex-direction: column; gap: 2px">
<span style="padding: 8px 10px 6px; font-size: 12px; color: #9DB0A6">Open {{ cur.name }} in…</span>
<button role="menuitem" @click="toDocked" style="display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 0 10px; border-radius: 9px; border: 0; background: transparent; color: #E8EFEA; cursor: pointer; text-align: left; font-size: 14px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M15 4v16"></path></svg>Side panel</button>
<button role="menuitem" @click="toPage" style="display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 0 10px; border-radius: 9px; border: 0; background: transparent; color: #E8EFEA; cursor: pointer; text-align: left; font-size: 14px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M9 4v16"></path></svg>Fit to page</button>
<button role="menuitem" @click="toFull" style="display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 0 10px; border-radius: 9px; border: 0; background: transparent; color: #E8EFEA; cursor: pointer; text-align: left; font-size: 14px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"></path></svg>Full screen</button>
<button role="menuitem" @click="toClosed" style="display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 0 10px; border-radius: 9px; border: 0; background: transparent; color: #E8EFEA; cursor: pointer; text-align: left; font-size: 14px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg>Close</button>
</div>
</template>
<div style="display: flex; align-items: center; gap: 10px; padding: 8px 8px 8px 10px; border-radius: 30px; background: #1D2A25; border: 1px solid #34453D; box-shadow: 0 12px 30px rgba(0,0,0,0.4)">
<span :style="`width: 36px; height: 36px; border-radius: 50%; background: ${cur.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 13px`">{{ cur.initials }}</span>
<div style="display: flex; flex-direction: column; padding-right: 4px"><strong style="font-size: 13px">{{ cur.name }}</strong><span style="font-size: 11px; color: #B8F36B">● 2 new updates</span></div>
<button aria-label="Forum view options" aria-haspopup="menu" @click="toggleMenu" style="width: 44px; height: 44px; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><circle cx="5" cy="12" r="2"></circle><circle cx="12" cy="12" r="2"></circle><circle cx="19" cy="12" r="2"></circle></svg></button>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'PanelBubble',
  ...hubBindings(["cur","isMin","menuOpen","toClosed","toDocked","toFull","toPage","toggleMenu"])
};
</script>
