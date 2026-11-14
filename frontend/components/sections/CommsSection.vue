<!-- Communications (memos, newsletter, announcements) -->
<template><template v-if="show.comms">
<template v-if="commView.open">
<div style="width: 100%; max-width: 1100px; margin: 0 auto; display: flex; flex-direction: column; gap: 18px">
<button @click="commView.close" style="align-self: flex-start; height: 38px; padding: 0 16px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">‹ Communications</button>

<header style="padding: 22px 24px; display: flex; flex-direction: column; gap: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px">
<span :style="`padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: 700; background: ${commView.typeColor}22; color: ${commView.typeColor}`">{{ commView.type }}</span>
<span :style="`padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: 700; background: ${commView.statusColor}22; color: ${commView.statusColor}`">{{ commView.status }}</span>
<button v-if="commView.canRepost" type="button" :disabled="commView.reposting" @click="commView.repost" style="margin-left: auto; height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 13px">{{ commView.reposting ? 'Reposting…' : '↻ Repost' }}</button>
<span style="font-family: 'Space Grotesk', sans-serif; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ commView.ref }}</span>
</div>
<h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 30px; letter-spacing: -0.02em; line-height: 1.2; overflow-wrap: anywhere">{{ commView.title }}</h1>
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px 14px">
<span aria-hidden="true" :style="`width: 36px; height: 36px; border-radius: 50%; background: ${commView.typeColor}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 13px`">{{ commView.fromInitials }}</span>
<div style="display: flex; flex-direction: column"><strong style="font-size: 14px">{{ commView.from }}<template v-if="commView.hasDept"><span style="font-weight: 400; color: var(--k-9DB0A6,#9DB0A6)"> · {{ commView.byDept }}</span></template></strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ commView.dateLabel }}</span></div>
<div v-if="commView.showStats" style="margin-left: auto; display: flex; align-items: center; gap: 10px; min-width: 190px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); white-space: nowrap">Read by {{ commView.readPct }}</span><div role="progressbar" :aria-valuenow="commView.readNum" aria-valuemin="0" aria-valuemax="100" aria-label="Share of people who read it" style="flex: 1; height: 8px; border-radius: 4px; background: var(--k-2A3832,#2A3832); overflow: hidden"><div :style="`height: 100%; width: ${commView.readNum}%; background: ${commView.readNum >= 80 ? '#B8F36B' : '#FFB86B'}; border-radius: 4px`"></div></div></div>
</div>
</header>

<template v-if="commView.canInsights && commView.hasPlaces">
<div :style="`display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px; padding: 12px 16px; border-radius: 14px; font-size: 14px; font-weight: 600; background: ${commView.measure.landed ? 'rgba(184,243,107,0.12)' : 'rgba(255,184,107,0.12)'}; border: 1px solid ${commView.measure.landed ? '#B8F36B' : '#FFB86B'}`" role="status">
<span>{{ commView.measure.landed ? 'Landed: ' + commView.measure.readPct + '% of its audience has read it (target ' + commView.measure.target + '%).' : 'Not landed yet: ' + commView.measure.readPct + '% read, below the ' + commView.measure.target + '% target. ' + commView.measure.unread.length + ' people still to read it.' }}</span>
</div>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 12px">
<div style="padding: 14px 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px; display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 24px; font-weight: 700; font-family: 'Space Grotesk', sans-serif">{{ commView.measure.audience }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">People it reached</span></div>
<div style="padding: 14px 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px; display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 24px; font-weight: 700; font-family: 'Space Grotesk', sans-serif">{{ commView.measure.readPct }}%</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Read · {{ commView.measure.read }} people</span></div>
<div style="padding: 14px 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px; display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 24px; font-weight: 700; font-family: 'Space Grotesk', sans-serif">{{ commView.measure.read48Pct }}%</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Read within 48 hours</span></div>
<div style="padding: 14px 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px; display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 24px; font-weight: 700; font-family: 'Space Grotesk', sans-serif">{{ commView.measure.ackPct }}%</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Acknowledged · {{ commView.measure.acked }} people</span></div>
<div style="padding: 14px 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px; display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 24px; font-weight: 700; font-family: 'Space Grotesk', sans-serif">{{ commView.measure.median }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Median time to read</span></div>
<div style="padding: 14px 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px; display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 24px; font-weight: 700; font-family: 'Space Grotesk', sans-serif">{{ commView.totals.likes }} · {{ commView.totals.replies }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Reactions · replies</span></div>
</div>
</template>

<div style="display: flex; flex-wrap: wrap; gap: 18px; align-items: flex-start">
<div style="flex: 999 1 460px; min-width: 0; display: flex; flex-direction: column; gap: 18px">
<article style="padding: 24px; display: flex; flex-direction: column; gap: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px">
<p v-if="commView.hasSummary" style="margin: 0; font-size: 17px; font-weight: 600; line-height: 1.5"><LinkText :text="commView.summary"/></p>
<div style="font-size: 15px; line-height: 1.7; white-space: pre-wrap; overflow-wrap: anywhere; max-width: 72ch; color: var(--k-E8EFEA,#E8EFEA)"><LinkText :text="commView.body"/></div>
<div v-if="commView.canAck || commView.acked" style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding-top: 10px; border-top: 1px solid var(--k-24322C,#24322C)">
<button v-if="commView.canAck" type="button" @click="commView.acknowledge" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Acknowledge</button>
<span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ commView.acked ? 'You acknowledged this.' : 'Let the sender know you received and noted it.' }}</span>
</div>
</article>
<div v-if="commView.hasFiles" style="padding: 18px 20px; display: flex; flex-direction: column; gap: 8px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px"><strong style="font-size: 15px">Attachments</strong>
<template v-for="f in commView.files" :key="f.name"><button type="button" @click="f.open" style="display: flex; gap: 10px; align-items: center; padding: 10px 14px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ f.name }}</span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); white-space: nowrap">{{ f.meta }} · Open</span></button></template></div>
<div v-if="commView.canInsights" style="padding: 18px 20px; display: flex; flex-direction: column; gap: 12px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px"><strong style="font-size: 15px">Where it went and how it landed</strong>
<div v-if="commView.hasPlaces" style="display: flex; flex-direction: column; gap: 6px; padding: 10px 12px; border-radius: 12px; background: var(--k-121B17,#121B17); font-size: 13px"><strong>Acknowledged by {{ commView.totals.acked }}<template v-if="commView.totals.audience"> of {{ commView.totals.audience }}</template></strong><span v-if="commView.ackedBy.length" style="color: var(--k-C9D6CF,#C9D6CF)">{{ commView.ackedLine }}</span><span v-else style="color: var(--k-9DB0A6,#9DB0A6)">Nobody has acknowledged it yet.</span></div>
<span v-if="!commView.hasPlaces && !commView.ins.loading && !commView.ins.err" style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Only people with the right to see insights can see this.</span>
<PageSkeleton v-if="commView.ins.loading" kind="lines" bare />
<span v-if="commView.ins.err" role="alert" style="color: #FF9B8A">{{ commView.ins.err }}</span>
<div v-if="commView.hasPlaces" style="overflow-x: auto"><table style="width: 100%; border-collapse: collapse; font-size: 13px"><thead><tr style="text-align: left; color: var(--k-9DB0A6,#9DB0A6); font-size: 12px"><th style="padding: 6px">Where</th><th style="padding: 6px">Status</th><th style="padding: 6px">Read</th><th style="padding: 6px">Likes</th><th style="padding: 6px">Replies</th></tr></thead><tbody>
<template v-for="x in commView.places" :key="x.comm"><tr style="border-top: 1px solid var(--k-24322C,#24322C)"><td style="padding: 10px 6px"><strong>{{ x.place }}</strong><div style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ x.kindLabel }} · {{ x.audience }} people</div></td><td style="padding: 10px 6px">{{ x.status }}</td><td style="padding: 10px 6px">{{ x.read }} ({{ x.readPct }})</td><td style="padding: 10px 6px">{{ x.likes }}</td><td style="padding: 10px 6px">{{ x.replies }}</td></tr></template></tbody></table></div></div>

<div v-if="commView.canInsights && commView.hasPlaces" style="padding: 18px 20px; display: flex; flex-direction: column; gap: 12px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px"><strong style="font-size: 15px; flex: 1 1 200px">Not read yet <span style="font-weight: 400; color: var(--k-9DB0A6,#9DB0A6)">· {{ commView.measure.unread.length }} people</span></strong>
<button v-if="commView.measure.unread.length" type="button" :disabled="commView.measure.reminding" @click="commView.measure.remind" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">{{ commView.measure.reminding ? 'Sending…' : 'Remind only these ' + commView.measure.unread.length }}</button></div>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Read means: opened it in Communications, acknowledged it, or had it on screen for at least 2 seconds on the wall or in a forum. The reminder goes to their Hub notifications, not to everyone and not to WhatsApp. At most once every 6 hours.</span>
<span v-if="commView.measure.remindNote" role="status" style="font-size: 13px; color: var(--k-B8F36B,#B8F36B)">{{ commView.measure.remindNote }}</span>
<span v-if="!commView.measure.unread.length" style="font-size: 13px; color: var(--k-B8F36B,#B8F36B)">Everyone it reached has read it.</span>
<div v-else style="display: flex; flex-wrap: wrap; gap: 6px"><span v-for="u in commView.measure.unreadShown" :key="u.person" :title="u.department" style="padding: 4px 10px; border-radius: 12px; font-size: 12px; background: var(--k-22302A,#22302A); color: var(--k-C9D6CF,#C9D6CF)">{{ u.name }} <span style="color: var(--k-9DB0A6,#9DB0A6)">· {{ u.department }}</span></span><span v-if="commView.measure.moreUnread" style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); align-self: center">+ {{ commView.measure.moreUnread }} more</span></div>
</div>
<div v-if="commView.canInsights && commView.measure.byDept.length > 0" style="padding: 18px 20px; display: flex; flex-direction: column; gap: 10px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px"><strong style="font-size: 15px">By department <span style="font-weight: 400; color: var(--k-9DB0A6,#9DB0A6)">· weakest first</span></strong>
<div style="overflow-x: auto"><table style="width: 100%; border-collapse: collapse; font-size: 13px"><thead><tr style="text-align: left; color: var(--k-9DB0A6,#9DB0A6); font-size: 12px"><th style="padding: 5px 6px">Department</th><th style="padding: 5px 6px">Reached</th><th style="padding: 5px 6px; min-width: 160px">Read</th><th style="padding: 5px 6px">Acknowledged</th></tr></thead><tbody>
<tr v-for="d in commView.measure.byDept" :key="d.department" style="border-top: 1px dotted var(--k-2A3832,#2A3832)"><td style="padding: 6px"><strong>{{ d.department }}</strong></td><td style="padding: 6px">{{ d.audience }}</td>
<td style="padding: 6px"><div style="display: flex; align-items: center; gap: 8px"><div style="flex: 1; height: 8px; border-radius: 4px; background: var(--k-24322C,#24322C); overflow: hidden"><div :style="`height: 100%; width: ${d.readPct}%; background: ${d.readPct >= 80 ? '#B8F36B' : d.readPct >= 40 ? '#FFB86B' : '#FF9B8A'}`"></div></div><span style="white-space: nowrap">{{ d.read }} ({{ d.readPct }}%)</span></div></td>
<td style="padding: 6px">{{ d.acked }} ({{ d.ackPct }}%)</td></tr></tbody></table></div>
</div>
</div>
<aside aria-label="Details" style="flex: 1 1 280px; min-width: 0; display: flex; flex-direction: column; gap: 18px">
<div v-if="commView.hasTo && commView.showStats" style="padding: 18px 20px; display: flex; flex-direction: column; gap: 10px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px"><strong style="font-size: 15px">Sent to <span style="font-weight: 400; color: var(--k-9DB0A6,#9DB0A6)">· {{ commView.toList.length }}</span></strong>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(t, ti) in (showAllTo ? commView.toList : commView.toList.slice(0, 6))" :key="ti"><span style="padding: 4px 10px; border-radius: 12px; font-size: 12px; background: var(--k-22302A,#22302A); color: var(--k-C9D6CF,#C9D6CF)">{{ t }}</span></template></div>
<button v-if="commView.toList.length > 6" type="button" @click="showAllTo = !showAllTo" style="align-self: flex-start; height: 30px; padding: 0 12px; border-radius: 15px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 12px">{{ showAllTo ? 'Show fewer' : 'Show all ' + commView.toList.length }}</button></div>
<div v-if="commView.fields.length" style="padding: 18px 20px; display: flex; flex-direction: column; gap: 12px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); border-radius: 16px"><strong style="font-size: 15px">Details</strong>
<template v-for="f in commView.fields" :key="f.label"><div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 11px; letter-spacing: 0.05em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">{{ f.label }}</span><span style="font-size: 14px; white-space: pre-wrap; overflow-wrap: anywhere"><LinkText :text="f.value"/></span></div></template></div>
</aside>
</div>
</div>
</template>
<template v-else-if="engagement.open">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<button @click="engagement.close" style="align-self: flex-start; height: 38px; padding: 0 16px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">‹ Communications</button>
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px"><div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Engagement</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">Did our communication land? Who it reached, how many read it and how fast, who acknowledged it, and who still has not. A communication lands when 80% of its audience has read it.</p></div>
<button @click="engagement.refresh" :disabled="engagement.loading" style="height: 38px; padding: 0 16px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">{{ engagement.loading ? 'Measuring…' : 'Refresh' }}</button></div>
<div v-if="engagement.hasErr" role="alert" style="padding: 10px 14px; border-radius: 12px; background: var(--k-3A2420,#3A2420); color: var(--k-FFD3CA,#FFD3CA); font-size: 14px">{{ engagement.err }}</div>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap: 12px">
<template v-for="k in engagement.kpis" :key="k.label"><div style="padding: 16px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); text-align: center"><strong style="font-size: 26px; font-family: 'Space Grotesk', sans-serif">{{ k.n }}</strong><div style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ k.label }}</div></div></template>
</div>
<div v-if="engagement.trend.length === 1" style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Week of {{ engagement.trend[0].label }}: <strong :style="`color: ${engagement.trend[0].color}`">{{ engagement.trend[0].pct }}% read</strong> (target 80%). The week-by-week chart appears once there are two weeks to compare.</div>
<div v-if="engagement.trend.length > 1" style="padding: 16px 18px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 10px">
<div style="display: flex; flex-wrap: wrap; align-items: baseline; gap: 6px 12px"><strong style="font-size: 15px">Read rate, week by week</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Of everyone reached by what was published that week, how many read it. Dashed line: the 80% target.</span></div>
<div style="position: relative; display: flex; align-items: flex-end; gap: 8px; height: 120px; padding-top: 6px">
<div aria-hidden="true" style="position: absolute; left: 0; right: 0; bottom: calc(18px + 80% * 0.82); border-top: 1px dashed #B8F36B; opacity: .7"></div>
<div v-for="w in engagement.trend" :key="w.label" :title="w.title" style="flex: 1; min-width: 0; height: 100%; display: flex; flex-direction: column; justify-content: flex-end; align-items: center; gap: 4px">
<span style="font-size: 11px; font-weight: 600">{{ w.pct }}%</span><div :style="`width: 100%; max-width: 46px; height: ${Math.max(2, w.pct * 0.82)}%; border-radius: 6px 6px 0 0; background: ${w.color}`"></div><span style="font-size: 10px; color: var(--k-9DB0A6,#9DB0A6)">{{ w.label }}</span></div>
</div>
</div>
<p style="margin: 0; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">How it is counted: <strong>reached</strong> = people in the place it was posted to (the author is not counted). <strong>Read</strong> = opened it in Communications, acknowledged it, or had it on screen for at least 2 seconds on the wall or in a forum. Scrolling past it does not count.</p>
<DataTable :t="engagementTable" />
</div>
</template>
<template v-else>
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Communications</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">Send announcements, memos, newsletters and job posts. You choose where each one goes.</p></div>
<div style="display: flex; flex-wrap: wrap; gap: 8px"><button v-if="engagement.canOpen" @click="engagement.openIt" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; cursor: pointer">Engagement</button><template v-for="(cn, cn__i) in commNew" :key="cn__i"><button @click="cn.run" :style="cn.style">{{ cn.label }}</button></template></div>
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
  data() { return { showAllTo: false }; },
  ...hubBindings(["commNew","commsTable","commsTabs","commView","engagement","engagementTable","show"])
};
</script>
