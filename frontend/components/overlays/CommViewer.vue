<!-- Reading one communication, and (for people allowed) how it was received where it was posted -->
<template><template v-if="commView.open">
<div class="hub-overlay" style="position: fixed; inset: 0; z-index: 66; background: rgba(5,10,8,0.65); display: flex; align-items: center; justify-content: center; padding: 16px; box-sizing: border-box" @click.self="commView.close">
<div role="dialog" :aria-label="commView.title" style="width: 100%; max-width: 720px; max-height: 92vh; display: flex; flex-direction: column; border-radius: 20px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-2A3832,#2A3832); box-shadow: 0 24px 60px rgba(0,0,0,0.5); overflow: hidden">
<div style="flex: none; display: flex; align-items: flex-start; gap: 12px; padding: 16px 20px; border-bottom: 1px solid var(--k-24322C,#24322C)">
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6px"><span :style="`align-self: flex-start; padding: 2px 10px; border-radius: 12px; font-size: 11px; font-weight: 700; background: ${commView.typeColor}22; color: ${commView.typeColor}`">{{ commView.type }}</span><h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 22px; overflow-wrap: anywhere">{{ commView.title }}</h2><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ commView.meta }}</span></div>
<button aria-label="Close" @click="commView.close" style="width: 36px; height: 36px; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 18px">×</button>
</div>
<template v-if="commView.canInsights"><div role="tablist" style="flex: none; display: flex; gap: 6px; padding: 0 14px; border-bottom: 1px solid var(--k-24322C,#24322C)">
<button role="tab" :aria-selected="commView.isRead" @click="commView.showRead" :style="`height: 40px; padding: 0 14px; border: 0; border-bottom: 3px solid ${commView.isRead ? '#B8F36B' : 'transparent'}; background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; cursor: pointer`">Read</button>
<button role="tab" :aria-selected="commView.isInsights" @click="commView.showInsights" :style="`height: 40px; padding: 0 14px; border: 0; border-bottom: 3px solid ${commView.isInsights ? '#B8F36B' : 'transparent'}; background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; cursor: pointer`">Interaction</button>
</div></template>
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 18px 20px; display: flex; flex-direction: column; gap: 14px">
<template v-if="commView.isRead">
<p v-if="commView.hasSummary" style="margin: 0; font-size: 15px; font-weight: 600; line-height: 1.5">{{ commView.summary }}</p>
<div v-if="commView.fields.length" style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 10px; padding: 12px 14px; border-radius: 12px; background: var(--k-121B17,#121B17)">
<template v-for="f in commView.fields" :key="f.label"><div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ f.label }}</span><span style="font-size: 14px; white-space: pre-wrap">{{ f.value }}</span></div></template>
</div>
<div style="font-size: 14px; line-height: 1.6; white-space: pre-wrap; overflow-wrap: anywhere">{{ commView.body }}</div>
<div v-if="commView.hasFiles" style="display: flex; flex-direction: column; gap: 6px"><strong style="font-size: 13px">Attachments</strong>
<template v-for="f in commView.files" :key="f.name"><button type="button" @click="f.open" style="display: flex; gap: 10px; align-items: center; padding: 8px 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ f.name }}</span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ f.meta }} · Open</span></button></template></div>
<span v-if="commView.hasTo" style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Sent to {{ commView.to }}</span>
</template>
<template v-if="commView.isInsights">
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); line-height: 1.5">How people received it in each place it was posted. Only people with the right to see insights can open this page.</span>
<span v-if="commView.ins.loading" role="status" style="color: var(--k-9DB0A6,#9DB0A6)">Loading…</span>
<span v-if="commView.ins.err" role="alert" style="color: #FF9B8A">{{ commView.ins.err }}</span>
<div v-if="commView.hasPlaces" style="display: grid; grid-template-columns: repeat(4, 1fr); gap: 8px">
<div style="padding: 10px; border-radius: 12px; background: var(--k-121B17,#121B17); text-align: center"><strong style="font-size: 20px">{{ commView.totals.read }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">Read</div></div>
<div style="padding: 10px; border-radius: 12px; background: var(--k-121B17,#121B17); text-align: center"><strong style="font-size: 20px">{{ commView.totals.likes }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">Likes</div></div>
<div style="padding: 10px; border-radius: 12px; background: var(--k-121B17,#121B17); text-align: center"><strong style="font-size: 20px">{{ commView.totals.replies }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">Replies</div></div>
<div style="padding: 10px; border-radius: 12px; background: var(--k-121B17,#121B17); text-align: center"><strong style="font-size: 20px">{{ commView.totals.audience }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">People reached</div></div>
</div>
<div v-if="commView.hasPlaces" style="overflow-x: auto"><table style="width: 100%; border-collapse: collapse; font-size: 13px"><thead><tr style="text-align: left; color: var(--k-9DB0A6,#9DB0A6); font-size: 12px"><th style="padding: 6px">Where</th><th style="padding: 6px">Status</th><th style="padding: 6px">Read</th><th style="padding: 6px">Likes</th><th style="padding: 6px">Replies</th></tr></thead><tbody>
<template v-for="x in commView.places" :key="x.comm"><tr style="border-top: 1px solid var(--k-24322C,#24322C)"><td style="padding: 8px 6px"><strong>{{ x.place }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ x.kindLabel }} · {{ x.audience }} people</div></td><td style="padding: 8px 6px">{{ x.status }}</td><td style="padding: 8px 6px">{{ x.read }} ({{ x.readPct }})</td><td style="padding: 8px 6px">{{ x.likes }}</td><td style="padding: 8px 6px">{{ x.replies }}</td></tr></template>
</tbody></table></div>
</template>
</div>
</div>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default { name: 'CommViewer', ...hubBindings(["commView"]) };
</script>
