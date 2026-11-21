<!-- Who has read a message of mine in a forum, and who has not: a small scrollable list, readers first. -->
<template>
<div v-if="seen" style="flex: 1 1 100%; min-width: 0; display: flex; flex-direction: column; gap: 4px">
<button type="button" :aria-expanded="open" @click.stop="open = !open" style="align-self: flex-start; height: 30px; padding: 0 10px; border: 1px solid var(--k-526259,#526259); border-radius: 8px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 12px">Seen by {{ seen.read.length }} of {{ seen.total }}</button>
<div v-if="open" @click.stop style="max-height: 96px; overflow-y: auto; padding: 4px 6px; border-radius: 8px; background: rgba(0, 0, 0, 0.18); font-size: 11px; line-height: 1.55">
<div v-for="p in seen.read" :key="'r' + p.person" style="display: flex; align-items: center; gap: 2px; white-space: nowrap"><MsgTicks state="read" :size="12"/><span style="overflow: hidden; text-overflow: ellipsis">{{ p.name }}</span></div>
<div v-for="p in seen.unread" :key="'u' + p.person" style="display: flex; align-items: center; gap: 2px; white-space: nowrap; color: var(--k-9DB0A6,#9DB0A6)"><MsgTicks state="sent" :size="12"/><span style="overflow: hidden; text-overflow: ellipsis">{{ p.name }}</span></div>
</div>
</div>
</template>

<script setup>
import { ref } from 'vue';

defineProps({ seen: { type: Object, default: null } });
const open = ref(false);
</script>
