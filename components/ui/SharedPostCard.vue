<!-- A wall post shared into a forum, shown as it was posted. Polls keep voting here; other reactions stay on the wall. -->
<template>
<div v-if="v" role="group" :aria-label="`Shared post by ${v.author}`" style="margin-top: 6px; padding: 10px; border-radius: 12px; background: #121B17; border: 1px solid #2A3832; display: flex; flex-direction: column; gap: 8px; min-width: 0">
<div style="display: flex; align-items: center; gap: 8px; min-width: 0"><span aria-hidden="true" :style="`width: 28px; height: 28px; flex: none; border-radius: 50%; background: ${v.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 11px`">{{ v.initials }}</span><span style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ v.author }}</strong><span style="font-size: 11px; color: #9DB0A6">Shared from the Community wall · {{ v.meta }}</span></span></div>
<p v-if="v.text" style="margin: 0; font-size: 14px; line-height: 1.5; white-space: pre-line; overflow-wrap: anywhere">{{ v.text }}</p>
<MediaCollage v-if="v.hasMedia" :files="v.media" framed :max-h="300" :row="130"/>
<template v-if="v.isPoll"><div style="display: flex; flex-direction: column; gap: 6px"><template v-for="(o, i) in v.poll" :key="i"><button type="button" @click="o.vote" :style="o.style + ';min-height:38px;width:100%'"><span style="overflow-wrap: anywhere">{{ o.label }}</span><span style="margin-left: auto; font-weight: 600">{{ o.pct }}</span></button></template><span style="font-size: 12px; color: #9DB0A6">{{ v.pollNote }}</span></div></template>
<span style="font-size: 11px; color: #9DB0A6">{{ v.syncNote }}</span>
<div style="display: flex; flex-direction: column; gap: 6px; border-top: 1px solid #24322C; padding-top: 8px">
<div v-if="v.hasComments" style="max-height: 160px; overflow-y: auto; display: flex; flex-direction: column; gap: 6px">
<template v-for="(c, i) in v.comments" :key="i"><div style="display: flex; gap: 6px; min-width: 0"><span aria-hidden="true" :style="`width: 22px; height: 22px; flex: none; border-radius: 50%; background: ${c.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 9px`">{{ c.initials }}</span><div style="min-width: 0; padding: 5px 10px; border-radius: 4px 12px 12px 12px; background: #1F2C26; font-size: 13px; overflow-wrap: anywhere"><strong>{{ c.who }}</strong> {{ c.text }}</div></div></template>
</div>
<CommentBox placeholder="Comment in this forum…" @send="v.addComment"/>
</div>
</div>
</template>

<script>
import { hubBindings } from '~/composables/useHub';
export default { name: 'SharedPostCard', props: { sid: { type: String, required: true } }, computed: { ...hubBindings(['sharedViews']).computed, v() { return this.sharedViews[this.sid]; } } };
</script>
