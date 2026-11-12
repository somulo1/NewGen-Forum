<!-- The list beside a chat that is fitted to the page or full screen: the other forums, or the other private chats, to switch between -->
<template>
<aside class="hub-chatrail" :aria-label="kind === 'dm' ? 'Private chats' : 'Forums'" style="flex: 0 0 280px; min-width: 0; display: flex; flex-direction: column; border-right: 1px solid var(--k-24322C,#24322C); background: var(--k-121B17,#121B17)">
<div style="flex: none; padding: 14px 14px 10px; display: flex; flex-direction: column; gap: 10px">
<strong style="font-family: 'Space Grotesk', sans-serif; font-size: 17px; color: var(--k-E8EFEA,#E8EFEA)">{{ kind === 'dm' ? 'Private chats' : 'Forums' }}</strong>
<input type="search" v-model="q" :placeholder="kind === 'dm' ? 'Search chats…' : 'Search forums…'" :aria-label="kind === 'dm' ? 'Search private chats' : 'Search forums'" style="height: 36px; padding: 0 12px; border-radius: 18px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px">
</div>
<div class="nav-scroll" style="flex: 1; min-height: 0; overflow-y: auto; padding: 0 8px 12px; display: flex; flex-direction: column; gap: 2px">
<template v-if="kind === 'dm'">
<button v-for="d in shownDms" :key="d.key" type="button" @click="d.open" :style="d.style + ';min-height:56px'">
<span style="position: relative; width: 36px; height: 36px; flex: none"><UserAvatar :name="d.name" :initials="d.initials" :color="d.color" :size="36" :font-size="12"/><span :style="`position: absolute; right: -1px; bottom: -1px; width: 10px; height: 10px; border-radius: 50%; background: ${d.dot}; border: 2px solid var(--k-121B17,#121B17)`"></span></span>
<span style="flex: 1; min-width: 0; display: flex; flex-direction: column"><span style="font-size: 14px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ d.name }}</span><span :style="`font-size: 12px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: ${d.typing ? '#7CC4FF' : 'var(--k-9DB0A6,#9DB0A6)'}`">{{ d.last }}</span></span>
<span v-if="d.hasUnread" style="min-width: 20px; height: 20px; padding: 0 6px; box-sizing: border-box; border-radius: 10px; background: #C9B6FF; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ d.unread }}</span>
</button>
</template>
<template v-else>
<button v-for="f in shownForums" :key="f.key" type="button" @click="f.open" :style="f.rowStyle + ';min-height:52px'">
<span :style="`width: 34px; height: 34px; flex: none; border-radius: 10px; background: ${f.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 12px`">{{ f.initials }}</span>
<span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ f.name }}</span>
<span v-if="f.hasUnread" style="min-width: 20px; height: 20px; padding: 0 6px; box-sizing: border-box; border-radius: 10px; background: #B8F36B; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ f.unread }}</span>
</button>
</template>
<span v-if="empty" style="padding: 12px; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ q ? 'Nothing matches "' + q + '".' : 'Nothing here yet.' }}</span>
</div>
</aside>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'ChatRail',
  props: { kind: { type: String, default: 'forum' } },
  data() { return { q: '' }; },
  computed: {
    ...hubBindings(['dmList', 'forums']).computed,
    needle() { return this.q.trim().toLowerCase(); },
    shownDms() { return (this.dmList || []).filter((d) => !this.needle || d.name.toLowerCase().indexOf(this.needle) !== -1); },
    shownForums() { return (this.forums || []).filter((f) => !f.notMember && (!this.needle || f.name.toLowerCase().indexOf(this.needle) !== -1)); },
    empty() { return this.kind === 'dm' ? !this.shownDms.length : !this.shownForums.length; }
  }
};
</script>

<style>
/* on a narrow screen the chat needs all the width: the list hides */
@media (max-width: 900px) { .hub-chatrail { display: none !important; } }
</style>
