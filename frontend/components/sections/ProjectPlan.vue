<!-- The open project's plan: backlog or work breakdown, timeline, reports, documents, and its type and workflow. -->
<template>
<section v-if="plan" aria-label="Project plan" class="pp" style="padding: 16px 18px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 14px; min-width: 0">
<ScrollTabs :active="plan.tab">
<button v-for="t in plan.tabs" :key="t.key" role="tab" :aria-selected="t.on" @click="t.go" :style="`min-height: 40px; padding: 0 14px; border: 0; background: transparent; cursor: pointer; font: inherit; font-size: 14px; border-bottom: 2px solid ${t.on ? 'var(--k-B8F36B,#B8F36B)' : 'transparent'}; color: ${t.on ? 'var(--k-E8EFEA,#E8EFEA)' : 'var(--k-9DB0A6,#9DB0A6)'}; font-weight: ${t.on ? 600 : 500}`">{{ t.label }}<span v-if="t.count !== null" style="color: var(--k-9DB0A6,#9DB0A6); font-weight: 400"> {{ t.count }}</span></button>
</ScrollTabs>

<!-- Backlog / work breakdown -->
<template v-if="plan.isBacklog">
<template v-if="plan.epics.length">
<strong style="font-size: 14px">Epics</strong>
<button v-for="e in plan.epics" :key="e.id" type="button" class="pp-row" @click="e.open" style="flex-wrap: wrap">
<span class="pp-key">{{ e.key }}</span><strong style="flex: 1 1 160px; min-width: 0; font-size: 13px">{{ e.title }}</strong><span class="pp-mut">{{ e.done }} of {{ e.count }} done<template v-if="e.points"> · {{ e.points }} pts</template></span>
<span style="flex: 1 1 100%; height: 6px; border-radius: 3px; background: var(--k-22302A,#22302A); overflow: hidden"><span :style="`display: block; ${e.bar}`"></span></span>
</button>
</template>
<template v-if="plan.waterfall">
<div v-for="ph in plan.phases" :key="ph.id" style="display: flex; flex-direction: column; gap: 6px">
<div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: baseline"><strong style="font-size: 14px">{{ ph.name }}</strong><span class="pp-mut">gate {{ ph.due }} · {{ ph.done }} of {{ ph.total }} done</span></div>
<span v-if="!ph.rows.length" class="pp-mut">No tasks in this phase yet.</span>
<div v-for="r in ph.rows" :key="r.id" class="pp-row" style="cursor: default"><button type="button" class="pp-open" @click="r.open"><span class="pp-key">{{ r.key }}</span> {{ r.title }}</button><span v-if="r.blocked" class="pp-blocked">Blocked</span><span class="pp-mut"><span :style="`color: ${r.statusColor}`" aria-hidden="true">●</span> {{ r.status }} · {{ r.owner }}</span></div>
</div>
</template>
<template v-else>
<div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center"><strong style="font-size: 14px; flex: 1 1 200px">Backlog<span class="pp-mut" style="font-weight: 400"> · waiting to be planned into a sprint</span></strong><button v-if="plan.canAdd" type="button" class="pp-btn" @click="plan.addBacklog">Add to backlog</button></div>
<span v-if="!plan.backlog.length" class="pp-mut">The backlog is empty. Tasks with no sprint wait here.</span>
<div v-for="r in plan.backlog" :key="r.id" class="pp-row" style="cursor: default"><button type="button" class="pp-open" @click="r.open"><span class="pp-key">{{ r.key }}</span> {{ r.title }}</button><span class="pp-mut">{{ r.kind }}<template v-if="r.points !== ''"> · {{ r.points }} pts</template> · {{ r.owner }}</span><span v-if="r.blocked" class="pp-blocked">Blocked</span>
<label style="display: flex; flex-direction: column; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)"><span class="pp-sr">Plan {{ r.title }} into a sprint</span><select :value="r.sprint" @change="r.setSprint" class="pp-input" style="height: 32px"><option v-for="o in plan.sprintOpts" :key="o.v" :value="o.v">{{ o.l }}</option></select></label></div>
</template>
</template>

<!-- Timeline -->
<template v-if="plan.isTimeline">
<span v-if="plan.timeline.empty" class="pp-mut">Nothing to draw yet. Add phases and tasks with dates.</span>
<template v-else>
<div style="display: flex; flex-wrap: wrap; gap: 12px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)"><span v-for="l in legend" :key="l.name"><span :style="`display: inline-block; width: 10px; height: 10px; border-radius: 3px; margin-right: 5px; background: ${l.color}`"></span>{{ l.name }}</span><span>◆ milestone or gate</span><span>→ depends on</span></div>
<div style="overflow-x: auto; scrollbar-width: thin">
<svg :width="plan.timeline.width" :height="plan.timeline.height" role="img" aria-label="Project timeline" style="display: block; font-family: inherit">
<template v-for="k in plan.timeline.ticks" :key="k.x"><line :x1="k.x" :x2="k.x" y1="20" :y2="plan.timeline.height" stroke="var(--k-24322C,#24322C)"/><text :x="k.x + 3" y="13" font-size="10" fill="var(--k-9DB0A6,#9DB0A6)">{{ k.label }}</text></template>
<line v-if="plan.timeline.todayX !== null" :x1="plan.timeline.todayX" :x2="plan.timeline.todayX" y1="0" :y2="plan.timeline.height" stroke="var(--k-FF9B8A,#FF9B8A)" stroke-width="2"/>
<template v-for="r in plan.timeline.rows" :key="r.key">
<g v-if="r.type === 'phase'"><text x="0" :y="r.y + 19" font-size="12" font-weight="700" fill="var(--k-E8EFEA,#E8EFEA)">{{ r.label }}</text><path :d="`M${r.x} ${r.y + 8} l7 7 l-7 7 l-7 -7 z`" fill="var(--k-E8EFEA,#E8EFEA)"><title>{{ r.tip }}</title></path></g>
<g v-else style="cursor: pointer" @click="r.open"><title>{{ r.tip }}</title><rect x="0" :y="r.y" :width="plan.timeline.width" height="30" fill="transparent"/><text x="10" :y="r.y + 19" font-size="12" fill="var(--k-C9D6CF,#C9D6CF)">{{ r.label }}</text><rect :x="r.x" :y="r.y + 8" :width="r.w" height="14" rx="4" :fill="r.color"/></g>
</template>
<path v-for="l in plan.timeline.links" :key="l.key" :d="l.d" fill="none" stroke="var(--k-9DB0A6,#9DB0A6)" stroke-width="1.5"/>
</svg>
</div>
</template>
</template>

<!-- Reports -->
<template v-if="plan.isReports">
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap: 10px"><div v-for="st in plan.reports.stats" :key="st.label" style="padding: 10px 12px; border-radius: 12px; background: var(--k-121B17,#121B17)"><span class="pp-mut" style="display: block">{{ st.label }}</span><strong style="font-size: 15px">{{ st.value }}</strong></div></div>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); max-width: 280px">Burndown of<select :value="plan.reports.rangeValue" @change="plan.reports.onRange" class="pp-input"><option v-for="o in plan.reports.ranges" :key="o.v" :value="o.v">{{ o.l }}</option></select></label>
<strong style="font-size: 14px">Burndown · {{ plan.reports.title }}</strong>
<span v-if="plan.reports.empty" class="pp-mut">No work in this range yet.</span>
<template v-else>
<div style="display: flex; flex-wrap: wrap; gap: 12px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)"><span><span style="display: inline-block; width: 14px; height: 2px; margin: 0 5px 3px 0; background: #7CC4FF"></span>Remaining {{ plan.reports.unit }}</span><span><span style="display: inline-block; width: 14px; margin: 0 5px 3px 0; border-top: 2px dashed var(--k-9DB0A6,#9DB0A6)"></span>Ideal</span></div>
<svg :viewBox="`0 0 ${plan.reports.W} ${plan.reports.H}`" role="img" :aria-label="`Burndown of ${plan.reports.title}`" style="width: 100%; max-width: 680px; height: auto; font-family: inherit">
<line :x1="plan.reports.L" :x2="plan.reports.W - plan.reports.R" :y1="plan.reports.H - plan.reports.B" :y2="plan.reports.H - plan.reports.B" stroke="var(--k-34453D,#34453D)"/>
<text :x="plan.reports.L - 6" :y="plan.reports.T + 8" text-anchor="end" font-size="11" fill="var(--k-9DB0A6,#9DB0A6)">{{ plan.reports.total }}</text><text :x="plan.reports.L - 6" :y="plan.reports.H - plan.reports.B" text-anchor="end" font-size="11" fill="var(--k-9DB0A6,#9DB0A6)">0</text>
<text :x="plan.reports.L" :y="plan.reports.H - 8" font-size="11" fill="var(--k-9DB0A6,#9DB0A6)">{{ plan.reports.startLabel }}</text><text :x="plan.reports.W - plan.reports.R" :y="plan.reports.H - 8" text-anchor="end" font-size="11" fill="var(--k-9DB0A6,#9DB0A6)">{{ plan.reports.endLabel }}</text>
<line :x1="plan.reports.idealX1" :y1="plan.reports.idealY1" :x2="plan.reports.idealX2" :y2="plan.reports.idealY2" stroke="var(--k-9DB0A6,#9DB0A6)" stroke-width="2" stroke-dasharray="5 5"/>
<path :d="plan.reports.line" fill="none" stroke="#7CC4FF" stroke-width="2" stroke-linejoin="round"/>
<g v-for="q in plan.reports.points" :key="q.x"><circle :cx="q.x" :cy="q.y" r="10" fill="transparent"><title>{{ q.tip }}</title></circle></g>
<circle v-if="plan.reports.last" :cx="plan.reports.last.x" :cy="plan.reports.last.y" r="4" fill="#7CC4FF" stroke="var(--k-16201C,#16201C)" stroke-width="2"/>
</svg>
<details><summary class="pp-mut" style="cursor: pointer">Show as a table</summary><table style="font-size: 12px; margin-top: 6px; border-collapse: collapse"><tr><th style="text-align: left; padding-right: 16px">Day</th><th style="text-align: right">Left ({{ plan.reports.unit }})</th></tr><tr v-for="q in plan.reports.points" :key="q.x"><td>{{ q.date }}</td><td style="text-align: right">{{ q.left }}</td></tr></table></details>
</template>
<template v-if="plan.reports.showVelocity">
<strong style="font-size: 14px">Velocity<template v-if="plan.reports.average"> · average {{ plan.reports.average }} {{ plan.reports.unit }} a sprint</template></strong>
<span v-if="plan.reports.velEmpty" class="pp-mut">Velocity appears once a sprint has started.</span>
<svg v-else :viewBox="`0 0 ${plan.reports.W} ${plan.reports.H}`" role="img" aria-label="Finished work per sprint" style="width: 100%; max-width: 680px; height: auto; font-family: inherit">
<line :x1="plan.reports.L" :x2="plan.reports.W - plan.reports.R" :y1="plan.reports.H - plan.reports.B" :y2="plan.reports.H - plan.reports.B" stroke="var(--k-34453D,#34453D)"/>
<g v-for="b in plan.reports.bars" :key="b.id"><rect :x="b.x" :y="b.y" :width="b.w" :height="b.h" rx="4" fill="#7CC4FF"><title>{{ b.name }}: {{ b.value }} {{ plan.reports.unit }}</title></rect><text :x="b.x + b.w / 2" :y="b.y - 4" text-anchor="middle" font-size="11" fill="var(--k-E8EFEA,#E8EFEA)">{{ b.value }}</text><text :x="b.x + b.w / 2" :y="plan.reports.H - 8" text-anchor="middle" font-size="10" fill="var(--k-9DB0A6,#9DB0A6)">{{ b.name }}</text></g>
</svg>
</template>
</template>

<!-- Documents -->
<template v-if="plan.isDocs">
<div style="display: flex; flex-wrap: wrap; gap: 14px; align-items: flex-start">
<div style="flex: 1 1 220px; min-width: 0; display: flex; flex-direction: column; gap: 6px">
<label style="display: flex; flex-direction: column; font-size: 12px"><span class="pp-sr">Search pages</span><input type="search" :value="plan.docs.query" @input="plan.docs.onQuery" placeholder="Search pages" class="pp-input"></label>
<span v-if="!plan.docs.list.length" class="pp-mut">No pages yet. Start one below.</span>
<button v-for="d in plan.docs.list" :key="d.id" type="button" @click="d.open" :aria-current="d.on" :style="`text-align: left; padding: 8px 10px; padding-left: ${10 + d.indent}px; border: 0; border-radius: 8px; cursor: pointer; font: inherit; font-size: 13px; color: var(--k-E8EFEA,#E8EFEA); background: ${d.on ? 'var(--k-22302A,#22302A)' : 'transparent'}`">{{ d.title }}</button>
<span class="pp-mut" style="margin-top: 8px">New page from</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><button v-for="t in plan.docs.templates" :key="t.key" type="button" class="pp-btn" @click="t.pick">{{ t.name }}</button></div>
</div>
<div style="flex: 999 1 320px; min-width: 0; display: flex; flex-direction: column; gap: 10px">
<form v-if="plan.docs.draft" novalidate @submit.prevent="plan.docs.draft.save" aria-label="Edit page" style="display: flex; flex-direction: column; gap: 10px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Title<input type="text" maxlength="200" :value="plan.docs.draft.title" @input="plan.docs.draft.onTitle" class="pp-input"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Sits under<select :value="plan.docs.draft.parent" @change="plan.docs.draft.onParent" class="pp-input"><option value="">Top level</option><option v-for="o in plan.docs.draft.parentOpts" :key="o.v" :value="o.v">{{ o.l }}</option></select></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Page<textarea rows="14" :value="plan.docs.draft.body" @input="plan.docs.draft.onBody" class="pp-input" style="height: auto; padding: 10px 12px; font: inherit; font-size: 14px; line-height: 1.5; resize: vertical"></textarea></label>
<span v-if="plan.docs.draft.err" role="alert" style="font-size: 13px; color: var(--k-FF9B8A,#FF9B8A)">{{ plan.docs.draft.err }}</span>
<div style="display: flex; gap: 8px"><button type="submit" class="pp-btn pp-primary">Save page</button><button type="button" class="pp-btn" @click="plan.docs.draft.cancel">Cancel</button></div>
</form>
<template v-else-if="plan.docs.page">
<div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center"><h3 style="margin: 0; flex: 1 1 200px; font-size: 18px; overflow-wrap: anywhere">{{ plan.docs.page.title }}</h3><button type="button" class="pp-btn" @click="plan.docs.page.edit">Edit</button><button type="button" class="pp-btn" @click="plan.docs.page.remove">Delete</button></div>
<span v-if="plan.docs.page.updated" class="pp-mut">Last changed {{ plan.docs.page.updated }}</span>
<div v-html="plan.docs.page.html" style="font-size: 14px; line-height: 1.6; white-space: pre-wrap; overflow-wrap: anywhere"></div>
<template v-if="plan.docs.page.versions.length"><strong style="font-size: 13px; margin-top: 8px">Earlier versions</strong><div v-for="v in plan.docs.page.versions" :key="v.key" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center"><span class="pp-mut" style="flex: 1 1 200px">Replaced by {{ v.who }}, {{ v.when }}</span><button type="button" class="pp-btn" @click="v.restore">Restore</button></div></template>
</template>
<span v-else class="pp-mut">Choose a page, or start a new one.</span>
</div>
</div>
</template>

<!-- Type and workflow -->
<template v-if="plan.isSettings">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); max-width: 360px">Project type<select :value="plan.settings.method" :disabled="!plan.settings.canEdit" @change="plan.settings.onMethod" class="pp-input"><option v-for="o in plan.settings.methods" :key="o.v" :value="o.v">{{ o.l }}</option></select></label>
<strong style="font-size: 14px">Workflow</strong>
<span class="pp-mut">The columns of the task board, in order. Work must end in a finished status. Only the project manager moves tasks into or out of a finished status. A limit warns when a column holds too many tasks.</span>
<div v-for="r in plan.settings.rows" :key="r.key" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: flex-end; padding: 10px; border-radius: 12px; background: var(--k-121B17,#121B17)">
<label style="flex: 1 1 150px; display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Name<input type="text" maxlength="40" :value="r.name" :disabled="!plan.settings.canEdit" @input="r.onName" class="pp-input"></label>
<label style="flex: 1 1 130px; display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Means<select :value="r.cat" :disabled="!plan.settings.canEdit" @change="r.onCat" class="pp-input"><option v-for="c in plan.settings.cats" :key="c.v" :value="c.v">{{ c.l }}</option></select></label>
<label style="flex: 0 1 80px; display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Limit<input type="number" min="0" :value="r.wip" :disabled="!plan.settings.canEdit" @input="r.onWip" class="pp-input"></label>
<div v-if="plan.settings.canEdit" style="display: flex; gap: 4px"><button type="button" class="pp-btn" :disabled="r.first" :aria-label="`Move ${r.name} up`" @click="r.up">↑</button><button type="button" class="pp-btn" :disabled="r.last" :aria-label="`Move ${r.name} down`" @click="r.down">↓</button><button type="button" class="pp-btn" :disabled="r.inUse > 0" :title="r.inUse ? r.inUse + ' tasks are here' : 'Remove'" :aria-label="`Remove ${r.name}`" @click="r.remove">×</button></div>
</div>
<span v-if="plan.settings.err" role="alert" style="font-size: 13px; color: var(--k-FF9B8A,#FF9B8A)">{{ plan.settings.err }}</span>
<div v-if="plan.settings.canEdit" style="display: flex; flex-wrap: wrap; gap: 8px"><button type="button" class="pp-btn" @click="plan.settings.add">Add a status</button><template v-if="plan.settings.dirty"><button type="button" class="pp-btn pp-primary" @click="plan.settings.save">Save workflow</button><button type="button" class="pp-btn" @click="plan.settings.cancel">Cancel</button></template></div>
</template>
</section>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

const bound = hubBindings(['pj']);
export default {
  name: 'ProjectPlan',
  computed: {
    ...bound.computed,
    plan() { return this.pj && this.pj.plan; },
    legend() { return [{ name: 'Not started', color: '#9DB0A6' }, { name: 'In progress', color: '#7CC4FF' }, { name: 'Being checked', color: '#FFB86B' }, { name: 'Finished', color: '#B8F36B' }, { name: 'Overdue', color: '#FF9B8A' }]; }
  }
};
</script>

<style>
.pp-row { display: flex; flex-wrap: wrap; gap: 6px 10px; align-items: center; padding: 8px 10px; border-radius: 10px; border: 1px solid var(--k-24322C,#24322C); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font: inherit; text-align: left; }
.pp-open { flex: 1 1 180px; min-width: 0; padding: 0; border: 0; background: transparent; color: inherit; font: inherit; font-size: 13px; text-align: left; cursor: pointer; overflow-wrap: anywhere; }
.pp-key { font-size: 11px; font-weight: 700; color: var(--k-9DB0A6,#9DB0A6); }
.pp-mut { font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); }
.pp-blocked { padding: 2px 7px; border-radius: 6px; font-size: 10px; font-weight: 700; background: #FF9B8A22; color: var(--k-FF9B8A,#FF9B8A); }
.pp-input { height: 38px; padding: 0 10px; box-sizing: border-box; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px; }
.pp-btn { min-height: 32px; padding: 0 12px; border-radius: 16px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font: inherit; font-size: 13px; }
.pp-btn:disabled { opacity: .4; cursor: default; }
.pp-primary { border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; }
.pp-sr { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); }
</style>
