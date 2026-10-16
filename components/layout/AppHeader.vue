<!-- Top bar: menu (mobile), logo, search, theme, notifications, profile -->
<template><header class="hub-header" style="flex: none; display: flex; flex-wrap: wrap; align-items: center; gap: 16px; padding: 12px 20px; border-bottom: 1px solid #24322C; background: #121B17">
<div style="display: flex; align-items: center; gap: 10px; flex: none">
<button class="hub-m" aria-label="Open menu" @click="openNav" style="width: 40px; height: 40px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; align-items: center; justify-content: center"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M4 7h16M4 12h16M4 17h16"></path></svg></button>
<div style="width: 34px; height: 34px; border-radius: 9px; background: #B8F36B; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 17px">H</div>
<div class="hub-d" style="font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 18px; letter-spacing: -0.01em">Hub</div>
<div class="hub-d" style="color: #9DB0A6; font-size: 13px; padding-left: 10px; border-left: 1px solid #2A3832">[YOUR ORGANIZATION]</div>
</div>
<div class="hub-searchwrap" style="position: relative; flex: 1 1 280px; max-width: 560px; min-width: 0">
<label class="hub-search" style="display: flex; align-items: center; gap: 10px; height: 42px; padding: 0 14px; border-radius: 21px; background: #1A2520; border: 1px solid #2A3832; color: #9DB0A6">
<svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="11" cy="11" r="7"></circle><path d="M20 20l-4-4"></path></svg>
<span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Search</span>
<input type="search" :value="headerSearch.q" @input="headerSearch.onInput" @keydown.esc="headerSearch.clear" placeholder="Search people, forums, pages…" style="flex: 1; min-width: 0; background: transparent; border: 0; outline: none; color: #E8EFEA; font-size: 14px">
</label>
<template v-if="headerSearch.active">
<div class="hub-searchres" role="listbox" aria-label="Search results" style="position: absolute; left: 0; right: 0; top: 48px; z-index: 70; max-height: 420px; overflow-y: auto; padding: 6px; border-radius: 16px; background: #1D2A25; border: 1px solid #34453D; box-shadow: 0 18px 44px rgba(0,0,0,0.5)">
<template v-for="(r, r__i) in headerSearch.results" :key="r__i"><button @click="r.run" style="display: flex; align-items: center; gap: 10px; width: 100%; min-height: 44px; padding: 4px 10px; border: 0; border-radius: 10px; background: transparent; color: #E8EFEA; cursor: pointer; text-align: left; font-family: inherit"><span :style="`width: 30px; height: 30px; flex: none; border-radius: 9px; background: #22302A; color: ${r.color}; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 11px`">{{ r.initials }}</span><span style="flex: 1; min-width: 0; display: flex; flex-direction: column"><span style="font-size: 14px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ r.label }}</span><span style="font-size: 11px; color: #9DB0A6">{{ r.group }}<template v-if="r.sub"> · {{ r.sub }}</template></span></span></button></template>
<template v-if="headerSearch.none"><span style="display: block; padding: 12px; font-size: 13px; color: #9DB0A6">Nothing matches "{{ headerSearch.q }}".</span></template>
</div>
</template>
</div>
<div style="margin-left: auto; flex: none; display: flex; align-items: center; gap: 8px">
<button @click="toggleTheme" :aria-label="themeAria" :title="themeAria" style="height: 40px; padding: 0 14px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-size: 13px; font-weight: 600; cursor: pointer; display: flex; align-items: center; gap: 8px">
<template v-if="isDark"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="4"></circle><path d="M12 2v2M12 20v2M2 12h2M20 12h2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"></path></svg><span class="hub-d">Light</span></template>
<template v-if="isLight"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z"></path></svg><span class="hub-d">Dark</span></template>
</button>
<div style="position: relative">
<button aria-label="Notifications" @click="toggleNotifs" style="position: relative; width: 40px; height: 40px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M6 16v-5a6 6 0 0 1 12 0v5l1.5 2h-15z"></path><path d="M10 20.5a2 2 0 0 0 4 0"></path></svg>
<template v-if="hasBadge"><span style="position: absolute; top: -4px; right: -4px; min-width: 18px; height: 18px; padding: 0 5px; box-sizing: border-box; border-radius: 9px; background: #FF9B8A; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ badgeCount }}</span></template>
</button>
<template v-if="notifOpen">
<div class="hub-notif" role="menu" aria-label="Notifications" style="position: absolute; right: 0; top: 48px; z-index: 70; width: 360px; max-width: 86vw; max-height: 440px; overflow-y: auto; padding: 8px; border-radius: 16px; background: #1D2A25; border: 1px solid #34453D; box-shadow: 0 18px 44px rgba(0,0,0,0.5)">
<div style="padding: 6px 10px 8px; font-weight: 600; font-size: 14px">Notifications</div>
<template v-if="noNotifs"><div style="padding: 16px 10px; color: #9DB0A6; font-size: 13px">You're all caught up.</div></template>
<template v-for="(n, n__i) in notifs" :key="n__i">
<button role="menuitem" @click="n.open" style="display: flex; gap: 10px; width: 100%; text-align: left; padding: 10px; border: 0; border-radius: 10px; background: transparent; color: #E8EFEA; cursor: pointer">
<span :style="`width: 8px; height: 8px; margin-top: 6px; flex: none; border-radius: 50%; background: ${n.color}`"></span>
<span style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 13px; line-height: 1.4">{{ n.text }}</span><span style="font-size: 11px; color: #9DB0A6">{{ n.time }}</span></span>
</button>
</template>
</div>
</template>
</div>
<div ref="acct" style="position: relative; flex: none">
<button aria-label="Account menu" aria-haspopup="menu" :aria-expanded="acctOpen" @click.stop="acctOpen = !acctOpen" :style="`position: relative; width: 40px; height: 40px; padding: 0; border-radius: 50%; border: 2px solid ${me.statusColor}; background: #C2185B; color: #fff; font-weight: 600; font-size: 15px; cursor: pointer; display: flex; align-items: center; justify-content: center; overflow: hidden`">
<template v-if="me.hasPhoto"><img :src="me.photo" alt="" style="width: 100%; height: 100%; object-fit: cover"></template>
<template v-if="me.noPhoto">S</template>
</button>
<template v-if="acctOpen"><div role="menu" aria-label="Account" style="position: absolute; top: 48px; right: 0; z-index: 60; width: 200px; padding: 6px; border-radius: 12px; background: #1D2A25; border: 1px solid #34453D; box-shadow: 0 14px 30px rgba(0,0,0,0.45); display: flex; flex-direction: column; gap: 2px">
<button role="menuitem" @click="openProfile" style="display: flex; align-items: center; gap: 10px; min-height: 42px; padding: 0 12px; border: 0; border-radius: 8px; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 14px; text-align: left">My profile</button>
<button role="menuitem" @click="logout" style="display: flex; align-items: center; gap: 10px; min-height: 42px; padding: 0 12px; border: 0; border-radius: 8px; background: transparent; color: #FF9B8A; cursor: pointer; font-size: 14px; text-align: left">Sign out</button>
</div></template>
</div>
</div>
</header>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'AppHeader',
  data() { return { acctOpen: false }; },
  mounted() { this._off = (e) => { if (this.acctOpen && this.$refs.acct && !this.$refs.acct.contains(e.target)) this.acctOpen = false; }; this._esc = (e) => { if (e.key === 'Escape') this.acctOpen = false; }; document.addEventListener('click', this._off); document.addEventListener('keydown', this._esc); },
  beforeUnmount() { document.removeEventListener('click', this._off); document.removeEventListener('keydown', this._esc); },
  methods: { openProfile() { this.acctOpen = false; this.go.profile(); }, logout() { this.acctOpen = false; useAuth().signOut(); } },
  ...hubBindings(["badgeCount","headerSearch","go","hasBadge","isDark","isLight","me","noNotifs","notifOpen","notifs","openNav","themeAria","toggleNotifs","toggleTheme"])
};
</script>
