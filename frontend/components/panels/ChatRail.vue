<!-- The list beside a chat that is fitted to the page or full screen: the other forums, or the other private chats, to switch between -->
<template>
<aside data-no-busy :class="['hub-chatrail', { 'is-open': open }]" :aria-label="kind === 'dm' ? 'Private chats' : 'Forums'" style="flex: 0 0 280px; min-width: 0; display: flex; flex-direction: column; border-right: 1px solid var(--k-24322C,#24322C); background: var(--k-121B17,#121B17)">
<div style="flex: none; padding: 14px 14px 10px; display: flex; flex-direction: column; gap: 10px">
<div style="display: flex; align-items: center; gap: 8px"><button type="button" class="hub-chatrail-back" aria-label="Back to the chat" @click="$emit('close')" style="width: 32px; height: 32px; border-radius: 9px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6"></path></svg></button><strong style="flex: 1; font-family: 'Space Grotesk', sans-serif; font-size: 17px; color: var(--k-E8EFEA,#E8EFEA)">{{ kind === 'dm' ? 'Private chats' : 'Forums' }}</strong>
<button type="button" :aria-label="adding ? 'Cancel' : kind === 'dm' ? 'New private chat' : 'Create a forum'" :title="adding ? 'Cancel' : kind === 'dm' ? 'New private chat' : 'Create a forum'" :aria-expanded="adding" @click="toggleAdding" :style="`width: 32px; height: 32px; border-radius: 9px; border: 1px solid var(--k-2A3832,#2A3832); background: ${adding ? '#B8F36B' : 'transparent'}; color: ${adding ? '#0F1714' : 'var(--k-E8EFEA,#E8EFEA)'}; cursor: pointer; display: flex; align-items: center; justify-content: center`"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path :d="adding ? 'M6 6l12 12M18 6L6 18' : 'M12 5v14M5 12h14'"></path></svg></button></div>
<div v-if="adding && kind === 'forum'" style="padding: 12px; border-radius: 12px; background: var(--k-1A2520,#1A2520); border: 1px solid var(--k-2A3832,#2A3832); display: flex; flex-direction: column; gap: 10px">
<label style="display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Forum name<input type="text" :value="forumForm.name" @input="onForumName" placeholder="e.g. Finance Team" style="height: 34px; padding: 0 10px; border-radius: 8px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA)"></label>
<div style="display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Invite members<PeoplePicker :people="pickerPeople" :selected="forumPicked" placeholder="Search people by name…" @pick="pickForumMember" @remove="unpickForumMember"/></div>
<span v-if="forumForm.err" role="alert" style="font-size: 12px; color: #FF9B8A">{{ forumForm.err }}</span>
<button type="button" :disabled="forumForm.busy" @click="create" style="height: 38px; border-radius: 9px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">{{ forumForm.busy ? 'Creating…' : 'Create forum' }}</button>
</div>
<input type="search" :value="q" @input="onSearch" :placeholder="adding && kind === 'dm' ? 'Search anyone…' : kind === 'dm' ? 'Search chats…' : 'Search forums…'" :aria-label="adding && kind === 'dm' ? 'Search anyone to message' : kind === 'dm' ? 'Search private chats' : 'Search forums'" style="height: 36px; padding: 0 12px; border-radius: 18px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px">
</div>
<div v-if="confirming" role="alertdialog" aria-label="Delete" style="margin: 0 14px 8px; padding: 12px; border-radius: 12px; background: var(--k-3A2420,#3A2420); border: 1px solid var(--k-5A3A33,#5A3A33); display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 13px; line-height: 1.45; color: var(--k-FFD3CA,#FFD3CA)">Delete the {{ confirming.kind }} <strong>{{ confirming.name }}</strong>? Its history is erased from your account for good. Other people keep theirs.</span>
<span v-if="confirming.err" role="alert" style="font-size: 12px; color: #FF9B8A">{{ confirming.err }}</span>
<div style="display: flex; gap: 8px"><button type="button" :disabled="confirming.busy" @click="doDelete" style="flex: 1; height: 34px; border-radius: 17px; border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600; cursor: pointer">{{ confirming.busy ? 'Deleting…' : 'Delete' }}</button><button type="button" @click="confirming = null" style="flex: 1; height: 34px; border-radius: 17px; border: 1px solid var(--k-5A3A33,#5A3A33); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Keep</button></div>
</div>
<div class="nav-scroll" style="flex: 1; min-height: 0; overflow-y: auto; padding: 0 8px 12px; display: flex; flex-direction: column; gap: 2px">
<template v-if="kind === 'dm'">
<button v-for="d in shownDms" :key="d.key" type="button" class="nav-row" title="Press and hold, or right-click, to delete" @click="rowClick($event, d)" @pointerdown="pressStart(d)" @pointerup="pressEnd" @pointerleave="pressEnd" @pointercancel="pressEnd" @contextmenu.prevent="ask(d)" @keydown.delete="ask(d)" :style="d.style + ';min-height:56px'">
<span style="position: relative; width: 36px; height: 36px; flex: none"><UserAvatar :name="d.name" :initials="d.initials" :color="d.color" :size="36" :font-size="12"/><span :style="`position: absolute; right: -1px; bottom: -1px; width: 10px; height: 10px; border-radius: 50%; background: ${d.dot}; border: 2px solid var(--k-121B17,#121B17)`"></span></span>
<span style="flex: 1; min-width: 0; display: flex; flex-direction: column"><span style="font-size: 14px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ d.name }}</span><span :style="`font-size: 12px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: ${d.typing ? '#7CC4FF' : 'var(--k-9DB0A6,#9DB0A6)'}`">{{ d.last }}</span></span>
<span v-if="d.hasUnread" style="min-width: 20px; height: 20px; padding: 0 6px; box-sizing: border-box; border-radius: 10px; background: #C9B6FF; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ d.unread }}</span>
</button>
<template v-if="adding">
<span style="padding: 10px 6px 4px; font-size: 11px; font-weight: 600; letter-spacing: 0.08em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Start a new chat</span>
<button v-for="m in freshPeople" :key="m.name" type="button" @click="startChat(m)" style="display: flex; align-items: center; gap: 10px; width: 100%; min-height: 48px; padding: 4px 10px; border: 0; border-radius: 10px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left; font-family: inherit">
<UserAvatar :name="m.name" :initials="m.initials" :color="m.color" :size="32" :font-size="11"/>
<span style="flex: 1; min-width: 0; display: flex; flex-direction: column"><span style="font-size: 14px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ m.name }}</span><span v-if="m.title" style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ m.title }}</span></span>
</button>
<span v-if="!freshPeople.length" style="padding: 6px 10px; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ q ? 'Nobody new matches "' + q + '".' : 'Type a name to find someone.' }}</span>
</template>
</template>
<template v-else>
<button v-for="f in shownForums" :key="f.key" type="button" class="nav-row" title="Press and hold, or right-click, to delete" @click="rowClick($event, f)" @pointerdown="pressStart(f)" @pointerup="pressEnd" @pointerleave="pressEnd" @pointercancel="pressEnd" @contextmenu.prevent="ask(f)" @keydown.delete="ask(f)" :style="f.rowStyle + ';min-height:52px'">
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
  props: { kind: { type: String, default: 'forum' }, open: { type: Boolean, default: false } },
  emits: ['close'],
  data() { return { q: '', adding: false, confirming: null, timer: 0, held: false }; },
  methods: {
    toggleAdding() { this.adding = !this.adding; this.q = ''; if (this.onDmq) this.onDmq({ target: { value: '' } }); },
    // in "new chat" the same box also searches everyone in the organization
    onSearch(e) { this.q = e.target.value; if (this.adding && this.kind === 'dm' && this.onDmq) this.onDmq(e); },
    startChat(m) { this.adding = false; this.q = ''; m.open(); this.$emit('close'); },
    // press and hold (or right-click, or the Delete key) asks to delete; a normal click opens
    pressStart(item) { this.held = false; clearTimeout(this.timer); this.timer = setTimeout(() => { this.held = true; this.ask(item); }, 600); },
    pressEnd() { clearTimeout(this.timer); },
    rowClick(e, item) { if (this.held) { this.held = false; e.preventDefault(); return; } item.open(); this.$emit('close'); },
    ask(item) { clearTimeout(this.timer); if (item && item.remove) this.confirming = { kind: item.kindLabel, name: item.name, remove: item.remove, busy: false, err: '' }; },
    async doDelete() {
      const c = this.confirming; if (!c || c.busy) return;
      c.busy = true; c.err = '';
      try { await c.remove(); this.confirming = null; } catch (e) { c.busy = false; c.err = (e && e.message) || 'Could not delete it.'; }
    },
    async create() { await this.createForumNow(); if (!this.forumForm.err) this.adding = false; }
  },
  computed: {
    ...hubBindings(['dmList', 'forums', 'dmMatches', 'onDmq', 'forumForm', 'onForumName', 'createForumNow', 'pickerPeople', 'forumPicked', 'pickForumMember', 'unpickForumMember']).computed,
    // people found by the search who have no chat with me yet: picking one opens a fresh chat
    freshPeople() { const have = new Set((this.dmList || []).map((d) => d.name)); return (this.dmMatches || []).filter((m) => !have.has(m.name)); },
    needle() { return this.q.trim().toLowerCase(); },
    shownDms() { return (this.dmList || []).filter((d) => !this.needle || d.name.toLowerCase().indexOf(this.needle) !== -1); },
    shownForums() { return (this.forums || []).filter((f) => !f.notMember && (!this.needle || f.name.toLowerCase().indexOf(this.needle) !== -1)); },
    empty() { return this.kind === 'dm' ? !this.shownDms.length : !this.shownForums.length; }
  }
};
</script>

<style>
/* on a narrow screen the chat needs all the width: the list opens over the chat from its header's list button */
.hub-chatrail-back { display: none; }
@media (max-width: 900px) {
  .hub-chatrail { display: none !important; }
  .hub-chatrail.is-open { display: flex !important; position: absolute; inset: 0; z-index: 25; flex-basis: auto !important; width: 100%; border-right: 0 !important; }
  .hub-chatrail.is-open .hub-chatrail-back { display: flex; }
}
.hub-chatlist-btn { display: none !important; }
@media (max-width: 900px) { .hub-chatlist-btn { display: flex !important; } }
</style>
