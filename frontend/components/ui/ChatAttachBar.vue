<!-- Files added to the chat message that is being written (up to 5); they are sent with it when Send is pressed. -->
<template>
<div v-if="cfiles.length" style="display: flex; flex-wrap: wrap; align-items: center; gap: 6px; padding: 8px 14px 0 18px">
<span v-for="(f, i) in cfiles" :key="i" style="display: inline-flex; align-items: center; gap: 6px; max-width: 100%; height: 30px; padding: 0 6px 0 10px; border-radius: 15px; background: var(--k-1F2C26,#1F2C26); border: 1px solid var(--k-2A3832,#2A3832); font-size: 12px">
<span style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 160px">{{ f.name }}</span><span style="color: var(--k-9DB0A6,#9DB0A6)">{{ f.meta }}</span>
<button type="button" :aria-label="`Remove ${f.name}`" @click="removeCFile(i)" style="width: 20px; height: 20px; border-radius: 50%; border: 0; background: transparent; color: inherit; cursor: pointer; line-height: 1">&times;</button></span>
<span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ cfiles.length }} of 5</span>
</div>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'ChatAttachBar',
  ...hubBindings(["cfiles", "removeCFile"])
};
</script>
