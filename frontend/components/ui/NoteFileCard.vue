<!-- A shared note in a chat bubble, shaped like its card in the chat's Files tab: a small sheet of the note, its name, who may do what, and a button that opens it. -->
<template>
<button type="button" :disabled="!file.live" @click.stop="file.noteOpen" :aria-label="file.live ? `Open note ${title}` : `${title} is no longer shared`" :style="`display: flex; gap: 10px; align-items: stretch; width: 100%; max-width: 340px; margin-top: 4px; padding: 8px; box-sizing: border-box; border-radius: 12px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); text-align: left; font-family: inherit; cursor: ${file.live ? 'pointer' : 'default'}`">
<span :style="`flex: none; width: 78px; height: 92px; border-radius: 4px; background: #FBF6E3; padding: 8px; box-sizing: border-box; display: flex; flex-direction: column; gap: 4px; overflow: hidden; box-shadow: 0 4px 10px rgba(0,0,0,0.3); opacity: ${file.live ? 1 : 0.45}`"><span style="font-size: 8px; font-weight: 600; line-height: 1.3; color: #2A2716; overflow: hidden; max-height: 2.6em">{{ title }}</span><span style="height: 2px; width: 36%; background: #E0B93A"></span><span style="font-size: 6.5px; line-height: 1.5; color: #5A5434; white-space: pre-line; overflow: hidden; flex: 1">{{ file.live ? file.notePreview : '' }}</span></span>
<span style="flex: 1; min-width: 0; display: flex; flex-direction: column; justify-content: space-between; gap: 6px">
<span style="display: flex; flex-direction: column; gap: 2px; min-width: 0"><span style="align-self: flex-start; padding: 1px 6px; border-radius: 4px; background: #FFD66B; color: #0F1714; font-size: 10px; font-weight: 700">NOTE</span><strong style="font-size: 14px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ title }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ file.meta }}</span></span>
<span :style="`align-self: flex-start; height: 30px; padding: 0 14px; border-radius: 15px; border: 1px solid var(--k-2A3832,#2A3832); display: flex; align-items: center; font-size: 13px; color: ${file.live ? 'var(--k-B8F36B,#B8F36B)' : 'var(--k-6C7B73,#6C7B73)'}`">{{ file.live ? 'Open note' : 'Disconnected' }}</span>
</span>
</button>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({ file: { type: Object, required: true } });
const title = computed(() => String(props.file.name || 'Untitled').replace(/\.note$/, ''));
</script>
