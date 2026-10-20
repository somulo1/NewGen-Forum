<!-- Communications (memos, newsletter, announcements) -->
<template><template v-if="show.comms">
<template v-if="commView.open">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<button @click="commView.close" style="align-self: flex-start; height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">‹ Communications</button>
<div style="display: flex; flex-wrap: wrap; gap: 10px 14px; align-items: center">
<span style="font-family: 'Space Grotesk', sans-serif; color: var(--k-9DB0A6,#9DB0A6); font-size: 15px">{{ commView.ref }}</span>
<h1 style="margin: 0; flex: 1 1 320px; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em; line-height: 1.2; overflow-wrap: anywhere">{{ commView.title }}</h1>
<span :style="`padding: 3px 12px; border-radius: 12px; font-size: 12px; font-weight: 700; background: ${commView.typeColor}22; color: ${commView.typeColor}`">{{ commView.type }}</span>
<span :style="`padding: 3px 12px; border-radius: 12px; font-size: 12px; font-weight: 700; background: ${commView.statusColor}22; color: ${commView.statusColor}`">{{ commView.status }}</span>
</div>
<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-start">
<div style="flex: 999 1 460px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 18px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 10px"><strong style="font-size: 15px">Message</strong>
<p v-if="commView.hasSummary" style="margin: 0; font-size: 15px; font-weight: 600; line-height: 1.5">{{ commView.summary }}</p>
<div style="font-size: 14px; line-height: 1.6; white-space: pre-wrap; overflow-wrap: anywhere">{{ commView.body }}</div></div>
<div v-if="commView.hasFiles" style="padding: 18px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 8px"><strong style="font-size: 15px">Attachments</strong>
<template v-for="f in commView.files" :key="f.name"><button type="button" @click="f.open" style="display: flex; gap: 10px; align-items: center; padding: 8px 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ f.name }}</span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ f.meta }} · Open</span></button></template></div>
<div v-if="commView.canInsights" style="padding: 18px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 12px"><strong style="font-size: 15px">Interaction</strong>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">How people received it in each place it was posted. Only people with the right to see insights can see this.</span>
<span v-if="commView.ins.loading" role="status" style="color: var(--k-9DB0A6,#9DB0A6)">Loading…</span>
<span v-if="commView.ins.err" role="alert" style="color: #FF9B8A">{{ commView.ins.err }}</span>
<div v-if="commView.hasPlaces" style="overflow-x: auto"><table style="width: 100%; border-collapse: collapse; font-size: 13px"><thead><tr style="text-align: left; color: var(--k-9DB0A6,#9DB0A6); font-size: 12px"><th style="padding: 6px">Where</th><th style="padding: 6px">Status</th><th style="padding: 6px">Read</th><th style="padding: 6px">Likes</th><th style="padding: 6px">Replies</th></tr></thead><tbody>
<template v-for="x in commView.places" :key="x.comm"><tr style="border-top: 1px solid var(--k-24322C,#24322C)"><td style="padding: 8px 6px"><strong>{{ x.place }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ x.kindLabel }} · {{ x.audience }} people</div></td><td style="padding: 8px 6px">{{ x.status }}</td><td style="padding: 8px 6px">{{ x.read }} ({{ x.readPct }})</td><td style="padding: 8px 6px">{{ x.likes }}</td><td style="padding: 8px 6px">{{ x.replies }}</td></tr></template></tbody></table></div></div>
</div>
<aside aria-label="Details" style="flex: 1 1 280px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 16px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 10px"><strong style="font-size: 15px">Details</strong>
<div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">From</span><span style="font-size: 14px">{{ commView.from }}<template v-if="commView.hasDept"> · {{ commView.byDept }}</template></span></div>
<div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">Date</span><span style="font-size: 14px">{{ commView.dateLabel }}</span></div>
<div v-if="commView.hasTo" style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">Sent to</span><span style="font-size: 14px">{{ commView.to }}</span></div>
<div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">Read by</span><span style="font-size: 14px">{{ commView.readPct }} of people</span></div>
<template v-for="f in commView.fields" :key="f.label"><div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ f.label }}</span><span style="font-size: 14px; white-space: pre-wrap">{{ f.value }}</span></div></template></div>
<div v-if="commView.canInsights && commView.hasPlaces" style="padding: 16px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: grid; grid-template-columns: 1fr 1fr; gap: 8px"><strong style="grid-column: 1 / -1; font-size: 15px">Reach</strong>
<div style="text-align: center"><strong style="font-size: 20px">{{ commView.totals.read }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">Read</div></div>
<div style="text-align: center"><strong style="font-size: 20px">{{ commView.totals.audience }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">People reached</div></div>
<div style="text-align: center"><strong style="font-size: 20px">{{ commView.totals.likes }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">Likes</div></div>
<div style="text-align: center"><strong style="font-size: 20px">{{ commView.totals.replies }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">Replies</div></div></div>
</aside>
</div>
</div>
</template>
<template v-else>
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Communications</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">Send announcements, memos, newsletters and job posts. You choose where each one goes.</p></div>
<div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(cn, cn__i) in commNew" :key="cn__i"><button @click="cn.run" :style="cn.style">{{ cn.label }}</button></template></div>
</div>
<CommComposer inline />
<div role="tablist" aria-label="Kinds of communication" style="display: flex; flex-wrap: wrap; gap: 4px; border-bottom: 1px solid var(--k-24322C,#24322C)"><template v-for="t in commsTabs" :key="t.label"><button role="tab" :aria-selected="t.on" @click="t.pick" :style="t.style">{{ t.label }} <span style="font-size: 12px; opacity: 0.8">{{ t.count }}</span></button></template></div>
<DataTable :t="commsTable" />
</div>
</template>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'CommsSection',
  ...hubBindings(["commNew","commsTable","commsTabs","commView","show"])
};
</script>
