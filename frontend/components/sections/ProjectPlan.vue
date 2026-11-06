<!-- The open project's plan: backlog or work breakdown, timeline, reports, documents, and its type and workflow. -->
<template>
<section v-if="plan" aria-label="Project plan" class="pp" style="padding: 12px 18px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 14px; min-width: 0">
<div v-if="plan.stuck" role="alert" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center; padding: 10px 12px; border-radius: 10px; background: var(--k-2A2314,#2A2314); color: var(--k-FFD66B,#FFD66B); font-size: 13px"><span style="flex: 1 1 260px">This project still has its older encrypted plan, so changes cannot be saved yet. {{ plan.stuck.unread }} old item{{ plan.stuck.unread === 1 ? '' : 's' }} cannot be read on this device. Try from another team member's device first. Moving now leaves those items behind.</span><button type="button" class="pp-btn" @click="plan.stuck.move">Move what can be read</button></div>
<ScrollTabs :active="plan.tab">
<button v-for="t in plan.tabs" :key="t.key" role="tab" :aria-selected="t.on" @click="t.go" :style="`min-height: 40px; padding: 0 14px; border: 0; background: transparent; cursor: pointer; font: inherit; font-size: 14px; border-bottom: 2px solid ${t.on ? 'var(--k-B8F36B,#B8F36B)' : 'transparent'}; color: ${t.on ? 'var(--k-E8EFEA,#E8EFEA)' : 'var(--k-9DB0A6,#9DB0A6)'}; font-weight: ${t.on ? 600 : 500}`">{{ t.label }}<span v-if="t.count !== null" style="margin-left: 6px; color: var(--k-9DB0A6,#9DB0A6); font-weight: 400">{{ t.count }}</span></button>
</ScrollTabs>

<!-- Backlog / work breakdown -->
<template v-if="plan.isBacklog">
<slot name="plan-tables" />
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
<div v-for="r in plan.backlog" :key="r.id" :class="'pp-row'" :draggable="plan.canRank" @dragstart="drag = r.n" @dragover.prevent @drop.prevent="drag !== null && r.dropOn(drag); drag = null" style="cursor: grab"><span v-if="plan.canRank" style="display: flex; gap: 2px"><button type="button" class="pp-mini" :disabled="r.first" :aria-label="`Move ${r.title} up`" @click="r.up">↑</button><button type="button" class="pp-mini" :disabled="r.last" :aria-label="`Move ${r.title} down`" @click="r.down">↓</button></span><button type="button" class="pp-open" @click="r.open"><span class="pp-key">{{ r.key }}</span> {{ r.title }}</button><span class="pp-mut">{{ r.kind }}<template v-if="r.points !== ''"> · {{ r.points }} pts</template> · {{ r.owner }}</span><span v-if="r.blocked" class="pp-blocked">Blocked</span>
<label style="display: flex; flex-direction: column; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)"><span class="pp-sr">Plan {{ r.title }} into a sprint</span><select :value="r.sprint" :disabled="!plan.canPlan" @change="r.setSprint" class="pp-input" style="height: 32px"><option v-for="o in plan.sprintOpts" :key="o.v" :value="o.v">{{ o.l }}</option></select></label></div>
</template>
</template>

<!-- Timeline -->
<template v-if="plan.isTimeline">
<span v-if="plan.timeline.empty" class="pp-mut">Nothing to draw yet. Add phases and tasks with dates.</span>
<template v-else>
<div style="display: flex; flex-wrap: wrap; gap: 12px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)"><span v-for="l in legend" :key="l.name"><span :style="`display: inline-block; width: 10px; height: 10px; border-radius: 3px; margin-right: 5px; background: ${l.color}`"></span>{{ l.name }}</span><span>◆ milestone or gate</span><span>→ depends on</span><span v-if="plan.timeline.hasBaseline">▁ baseline</span></div>
<div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center"><span class="pp-mut" style="flex: 1 1 240px">Drag a bar to change its dates. Work that depends on it moves with it. <template v-if="plan.timeline.hasCritical">The outlined bars are the critical path.</template></span><button v-if="plan.timeline.canBaseline" type="button" class="pp-btn" @click="plan.timeline.saveBaseline">{{ plan.timeline.hasBaseline ? 'Update baseline' : 'Save baseline' }}</button></div>
<div style="overflow-x: auto; scrollbar-width: thin">
<svg :width="plan.timeline.width" :height="plan.timeline.height" role="img" aria-label="Project timeline" style="display: block; font-family: inherit">
<template v-for="k in plan.timeline.ticks" :key="k.x"><line :x1="k.x" :x2="k.x" y1="20" :y2="plan.timeline.height" stroke="var(--k-24322C,#24322C)"/><text :x="k.x + 3" y="13" font-size="10" fill="var(--k-9DB0A6,#9DB0A6)">{{ k.label }}</text></template>
<line v-if="plan.timeline.todayX !== null" :x1="plan.timeline.todayX" :x2="plan.timeline.todayX" y1="0" :y2="plan.timeline.height" stroke="var(--k-FF9B8A,#FF9B8A)" stroke-width="2"/>
<template v-for="r in plan.timeline.rows" :key="r.key">
<g v-if="r.type === 'phase'"><text x="0" :y="r.y + 19" font-size="12" font-weight="700" fill="var(--k-E8EFEA,#E8EFEA)">{{ r.label }}</text><path :d="`M${r.x} ${r.y + 8} l7 7 l-7 7 l-7 -7 z`" fill="var(--k-E8EFEA,#E8EFEA)"><title>{{ r.tip }}</title></path></g>
<g v-else><title>{{ r.tip }}</title><text x="10" :y="r.y + 19" font-size="12" fill="var(--k-C9D6CF,#C9D6CF)" style="cursor: pointer" @click="r.open">{{ r.label }}</text><rect v-if="r.baseX !== null" :x="r.baseX" :y="r.y + 23" :width="r.baseW" height="4" rx="2" fill="var(--k-6F8178,#6F8178)"/><rect :x="r.x + (dragKey === r.key ? dragPx : 0)" :y="r.y + 7" :width="r.w" height="14" rx="4" :fill="r.color" :stroke="r.critical ? 'var(--k-E8EFEA,#E8EFEA)' : 'none'" :stroke-width="r.critical ? 2 : 0" :style="`cursor: ${r.canDrag ? 'grab' : 'pointer'}; touch-action: none`" @pointerdown="barDown($event, r)" @pointermove="barMove" @pointerup="barUp($event, r)" @pointercancel="dragKey = null"/></g>
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
<div v-if="plan.docs.labels.length" style="display: flex; flex-wrap: wrap; gap: 4px"><button v-for="l in plan.docs.labels" :key="l.name" type="button" class="pp-mini" :aria-pressed="l.on" :style="l.on ? 'width: auto; padding: 0 8px; background: var(--k-22302A,#22302A)' : 'width: auto; padding: 0 8px'" @click="l.pick">{{ l.name }}</button></div>
<span v-if="!plan.docs.list.length" class="pp-mut">No pages found.</span>
<button v-for="d in plan.docs.list" :key="d.id" type="button" @click="d.open" :aria-current="d.on" :style="`text-align: left; padding: 8px 10px; padding-left: ${10 + d.indent}px; border: 0; border-radius: 8px; cursor: pointer; font: inherit; font-size: 13px; color: var(--k-E8EFEA,#E8EFEA); background: ${d.on ? 'var(--k-22302A,#22302A)' : 'transparent'}`">{{ d.title }}</button>
<label v-if="plan.docs.canImport" class="pp-btn" style="position: relative; align-self: flex-start; display: inline-flex; align-items: center; margin-top: 8px; overflow: hidden">Import from Confluence<input type="file" accept=".zip" aria-label="Import a Confluence HTML space export" :disabled="plan.docs.importing.busy" @change="plan.docs.importFile" style="position: absolute; inset: 0; opacity: 0; cursor: pointer"></label>
<span v-if="plan.docs.importing.text" role="status" class="pp-mut">{{ plan.docs.importing.text }}</span>
<ul v-if="plan.docs.importing.problems.length" class="pp-mut" style="margin: 0; padding-left: 16px"><li v-for="x in plan.docs.importing.problems" :key="x">{{ x }}</li></ul>
<span v-if="plan.docs.canCreate" class="pp-mut" style="margin-top: 8px">New page from</span>
<div v-if="plan.docs.canCreate" style="display: flex; flex-wrap: wrap; gap: 6px"><button v-for="t in plan.docs.templates" :key="t.key" type="button" class="pp-btn" @click="t.pick">{{ t.name }}</button></div>
</div>
<div style="flex: 999 1 320px; min-width: 0; display: flex; flex-direction: column; gap: 10px">
<form v-if="plan.docs.draft" novalidate @submit.prevent="plan.docs.draft.save" aria-label="Edit page" style="display: flex; flex-direction: column; gap: 10px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Title<input type="text" maxlength="200" :value="plan.docs.draft.title" @input="plan.docs.draft.onTitle" class="pp-input"></label>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 10px"><label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Sits under<select :value="plan.docs.draft.parent" @change="plan.docs.draft.onParent" class="pp-input"><option value="">Top level</option><option v-for="o in plan.docs.draft.parentOpts" :key="o.v" :value="o.v">{{ o.l }}</option></select></label>
<label v-if="!plan.docs.draft.isBoard" style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Background<select :value="plan.docs.draft.paper || 'theme'" @change="plan.docs.draft.onPaper" class="pp-input"><option v-for="o in plan.docs.papers" :key="o.v" :value="o.v">{{ o.l }}</option></select></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Labels, separated by commas<input type="text" :value="plan.docs.draft.labels" @input="plan.docs.draft.onLabels" class="pp-input"></label></div>
<BoardEditor v-if="plan.docs.draft.isBoard" :modelValue="plan.docs.draft.body" @update:modelValue="plan.docs.draft.setBody"/>
<template v-else>
<span v-if="editingWith.length" role="status" class="pp-mut">Also editing now: {{ editingWith.join(', ') }}. Your changes merge as you type.</span>
<NoteEditor :key="liveSession ? liveSession.key : 'solo'" macros :live="liveSession && liveSession.conn" :paper="plan.docs.draft.paper || 'theme'" :liveSeed="plan.docs.draft.seed || ''" :modelValue="plan.docs.draft.body" @update:modelValue="plan.docs.draft.setBody"/>
</template>
<span class="pp-mut">Link another page with [[Page title]], and a task with its key, such as PRJ-01-12.</span>
<span v-if="plan.docs.draft.err" role="alert" style="font-size: 13px; color: var(--k-FF9B8A,#FF9B8A)">{{ plan.docs.draft.err }}</span>
<div style="display: flex; gap: 8px"><button type="submit" class="pp-btn pp-primary">Save page</button><button type="button" class="pp-btn" @click="plan.docs.draft.cancel">Cancel</button></div>
</form>
<template v-else-if="plan.docs.page">
<nav v-if="plan.docs.page.crumbs.length" aria-label="Page path" style="display: flex; flex-wrap: wrap; gap: 4px" class="pp-mut"><template v-for="c in plan.docs.page.crumbs" :key="c.id"><button type="button" class="pp-link" @click="c.open">{{ c.title }}</button><span>›</span></template></nav>
<h3 style="margin: 0; font-size: 20px; overflow-wrap: anywhere">{{ plan.docs.page.title }}<span v-if="plan.docs.page.locked" class="pp-mut"> · locked</span></h3>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><button v-if="plan.docs.page.mayEdit" type="button" class="pp-btn pp-primary" @click="plan.docs.page.edit">Edit</button><button type="button" class="pp-btn" @click="plan.docs.page.toggleWatch">{{ plan.docs.page.watching ? 'Watching' : 'Watch' }}</button><button v-if="plan.docs.canCreate" type="button" class="pp-btn" @click="plan.docs.page.copy">Copy</button><template v-if="plan.docs.canExport"><button type="button" class="pp-btn" @click="plan.docs.page.exportDoc">Export Word</button><button type="button" class="pp-btn" @click="plan.docs.page.exportHtml">Export HTML</button><button type="button" class="pp-btn" @click="plan.docs.page.exportPdf">Export PDF</button></template><button v-if="plan.docs.page.canLock" type="button" class="pp-btn" @click="plan.docs.page.toggleLock">{{ plan.docs.page.locked ? 'Unlock' : 'Lock' }}</button><button v-if="plan.docs.page.canDelete" type="button" class="pp-btn" @click="plan.docs.page.remove">Delete</button></div>
<span class="pp-mut"><template v-if="plan.docs.page.updated">Last changed {{ plan.docs.page.updated }}</template><template v-for="l in plan.docs.page.labels" :key="l"> · {{ l }}</template></span>
<div ref="reading" @mouseup="readSelection" @keyup="readSelection" @touchend="readSelection"><div v-if="plan.docs.page.isBoard" v-html="plan.docs.page.boardHtml" style="overflow: auto"></div>
<NoteEditor v-else :key="plan.docs.page.id" :paper="plan.docs.page.paper" :editable="false" :modelValue="plan.docs.page.body" @ref="plan.docs.page.followRef" @comment="showComment"/></div>
<div v-if="plan.docs.canComment && selected" style="position: sticky; bottom: 8px; display: flex; flex-wrap: wrap; gap: 8px; align-items: center; padding: 8px 10px; border-radius: 10px; background: var(--k-22302A,#22302A)"><span class="pp-mut" style="flex: 1 1 160px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">“{{ selected }}”</span><button type="button" class="pp-btn pp-primary" @mousedown.prevent @click="commentOnSelection">Comment on this</button></div>
<template v-if="plan.docs.page.children.length"><strong style="font-size: 13px">Pages under this one</strong><button v-for="c in plan.docs.page.children" :key="c.id" type="button" class="pp-link" style="text-align: left" @click="c.open">{{ c.title }}</button></template>
<template v-if="plan.docs.page.backlinks.length"><strong style="font-size: 13px">Linked from</strong><button v-for="c in plan.docs.page.backlinks" :key="c.id" type="button" class="pp-link" style="text-align: left" @click="c.open">{{ c.title }}</button></template>
<details v-if="plan.docs.page.canRestrict"><summary style="cursor: pointer; font-size: 13px; font-weight: 700">Restrictions<template v-if="plan.docs.page.restricted"> · on</template></summary><div v-for="l in plan.docs.page.restrictLists" :key="l.key" style="margin-top: 8px"><span class="pp-mut">{{ l.label }}</span><div style="display: flex; flex-wrap: wrap; gap: 4px 12px; margin-top: 4px; font-size: 13px"><label v-for="m in l.people" :key="m.person" style="display: flex !important; flex-direction: row !important; width: auto; gap: 4px; align-items: center"><input type="checkbox" :checked="m.on" @change="m.toggle">{{ m.name }}</label></div></div><span class="pp-mut">Applies to the pages under this one too. The page's creator and the project manager always keep access.</span></details>
<strong style="font-size: 13px">Attachments</strong>
<div v-for="f in plan.docs.page.files" :key="f.key" style="display: flex; gap: 8px; align-items: center"><button type="button" class="pp-link" style="flex: 1; text-align: left; overflow-wrap: anywhere" @click="f.open">{{ f.name }}</button><button v-if="plan.docs.canRemoveFile" type="button" class="pp-mini" :aria-label="`Remove ${f.name}`" @click="f.remove">×</button></div>
<label v-if="plan.docs.canAttach" class="pp-btn" style="position: relative; align-self: flex-start; display: inline-flex; align-items: center; height: 32px; overflow: hidden">Attach files<input type="file" multiple aria-label="Attach files to this page" @change="plan.docs.page.onFile" style="position: absolute; inset: 0; opacity: 0; cursor: pointer"></label>
<strong style="font-size: 13px">Comments</strong>
<div v-for="c in plan.docs.page.comments" :key="c.id" :id="`pc-${c.id}`" :style="c.resolved ? 'opacity: .55' : ''" style="padding: 8px 10px; border-radius: 10px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; gap: 2px"><span class="pp-mut">{{ c.who }} · {{ c.when }}<button v-if="c.quote" type="button" class="pp-link" style="margin-left: 8px" @click="c.toggleResolved">{{ c.resolved ? 'Reopen' : 'Resolve' }}</button><button v-if="c.mine" type="button" class="pp-link" style="margin-left: 8px" @click="c.remove">Delete</button></span><span v-if="c.quote" style="font-size: 12px; font-style: italic; padding-left: 8px; border-left: 3px solid #E0A800; overflow-wrap: anywhere">{{ c.quote }}</span><span style="font-size: 13px; white-space: pre-line; overflow-wrap: anywhere">{{ c.text }}</span></div>
<span v-if="plan.docs.page.commentQuote" class="pp-mut">On “{{ plan.docs.page.commentQuote }}” <button type="button" class="pp-link" @click="plan.docs.page.clearQuote">Remove</button></span>
<form v-if="plan.docs.canComment" novalidate @submit.prevent="plan.docs.page.addComment" style="display: flex; flex-wrap: wrap; gap: 8px"><label style="flex: 1 1 200px; display: flex; flex-direction: column"><span class="pp-sr">Add a comment</span><input ref="commentBox" type="text" :value="plan.docs.page.comment" @input="plan.docs.page.onComment" placeholder="Add a comment" class="pp-input"></label><button type="submit" class="pp-btn">Comment</button></form>
<template v-if="plan.docs.page.versions.length"><strong style="font-size: 13px">Version history</strong><div v-for="v in plan.docs.page.versions" :key="v.key" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center"><span class="pp-mut" style="flex: 1 1 200px">{{ v.who }}, {{ v.when }}</span><button type="button" class="pp-btn" @click="v.compare">Compare</button><button v-if="plan.docs.page.mayEdit" type="button" class="pp-btn" @click="v.restore">Restore earlier</button></div></template>
<div v-if="plan.docs.page.diff" style="padding: 10px; border-radius: 10px; background: var(--k-121B17,#121B17); font-size: 13px; overflow-x: auto"><button type="button" class="pp-link" @click="plan.docs.page.closeDiff">Close comparison</button><div v-for="(d, di) in plan.docs.page.diff" :key="di" :style="`white-space: pre-wrap; color: ${d.k === 'add' ? 'var(--k-B8F36B,#B8F36B)' : d.k === 'del' ? 'var(--k-FF9B8A,#FF9B8A)' : 'inherit'}`">{{ d.k === 'add' ? '+ ' : d.k === 'del' ? '− ' : '  ' }}{{ d.t }}</div></div>
</template>
<template v-else><span class="pp-mut">Choose a page, or start a new one.</span><template v-if="plan.docs.recent.length"><strong style="font-size: 13px">Recently changed</strong><button v-for="c in plan.docs.recent" :key="c.id" type="button" class="pp-link" style="text-align: left" @click="c.open">{{ c.title }} <span class="pp-mut">· {{ c.when }}</span></button></template></template>
</div>
</div>
</template>

<!-- Type and workflow -->
<template v-if="plan.isSettings">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); max-width: 360px">Project type<select :value="plan.settings.method" :disabled="!plan.settings.canMethod" @change="plan.settings.onMethod" class="pp-input"><option v-for="o in plan.settings.methods" :key="o.v" :value="o.v">{{ o.l }}</option></select></label>
<strong style="font-size: 14px">Workflow</strong>
<span class="pp-mut">The columns of the task board, in order. Work must end in a finished status. Only the project manager moves tasks into or out of a finished status. A limit warns when a column holds too many tasks.</span>
<div v-for="r in plan.settings.rows" :key="r.key" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: flex-end; padding: 10px; border-radius: 12px; background: var(--k-121B17,#121B17)">
<label style="flex: 1 1 150px; display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Name<input type="text" maxlength="40" :value="r.name" :disabled="!plan.settings.canEdit" @input="r.onName" class="pp-input"></label>
<label style="flex: 1 1 130px; display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Means<select :value="r.cat" :disabled="!plan.settings.canEdit" @change="r.onCat" class="pp-input"><option v-for="c in plan.settings.cats" :key="c.v" :value="c.v">{{ c.l }}</option></select></label>
<label style="flex: 0 1 80px; display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Limit<input type="number" min="0" :value="r.wip" :disabled="!plan.settings.canEdit" @input="r.onWip" class="pp-input"></label>
<fieldset style="flex: 1 1 100%; margin: 0; padding: 0; border: 0; display: flex; flex-wrap: wrap; gap: 4px 10px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)"><legend style="padding: 0 0 4px">Can move to</legend><label v-for="x in r.targets" :key="x.key" style="display: flex !important; flex-direction: row !important; width: auto; gap: 4px; align-items: center"><input type="checkbox" :checked="x.on" :disabled="!plan.settings.canEdit" @change="x.toggle">{{ x.name }}</label></fieldset>
<div v-if="plan.settings.canEdit" style="display: flex; gap: 4px"><button type="button" class="pp-btn" :disabled="r.first" :aria-label="`Move ${r.name} up`" @click="r.up">↑</button><button type="button" class="pp-btn" :disabled="r.last" :aria-label="`Move ${r.name} down`" @click="r.down">↓</button><button v-if="r.canRemove" type="button" class="pp-btn" :disabled="r.inUse > 0" :title="r.inUse ? r.inUse + ' tasks are here' : 'Remove'" :aria-label="`Remove ${r.name}`" @click="r.remove">×</button></div>
</div>
<span v-if="plan.settings.err" role="alert" style="font-size: 13px; color: var(--k-FF9B8A,#FF9B8A)">{{ plan.settings.err }}</span>
<div v-if="plan.settings.canEdit" style="display: flex; flex-wrap: wrap; gap: 8px"><button type="button" class="pp-btn" @click="plan.settings.add">Add a status</button><template v-if="plan.settings.dirty"><button type="button" class="pp-btn pp-primary" @click="plan.settings.save">Save workflow</button><button type="button" class="pp-btn" @click="plan.settings.cancel">Cancel</button></template></div>
<div v-if="plan.settings.canDeleteProject" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center; padding-top: 10px; border-top: 1px solid var(--k-4A2A2A,#4A2A2A)"><span class="pp-mut" style="flex: 1 1 240px">Deleting this project removes its whole plan and documents for good. Its discussion stays.</span><button type="button" class="pp-btn" style="color: var(--k-FF9B8A,#FF9B8A); border-color: var(--k-6A3A34,#6A3A34)" @click="plan.settings.deleteProject">Delete project</button></div>
</template>
</section>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

const bound = hubBindings(['pj']);
export default {
  name: 'ProjectPlan',
  data() { return { liveSession: null, editingWith: [], selected: '', drag: null, dragKey: null, dragPx: 0, dragX: 0, moved: false }; },
  watch: {
    // a session opens when an existing page starts being edited, and closes with the editor
    liveKey: { immediate: true, handler(k) {
      if (this.liveSession) { this.liveSession.conn.leave(); this.liveSession = null; this.editingWith = []; }
      const d = this.plan && this.plan.docs && this.plan.docs.draft;
      if (k && d) this.liveSession = { key: k, conn: d.joinLive((names) => { this.editingWith = names; }) };
    } }
  },
  beforeUnmount() { if (this.liveSession) this.liveSession.conn.leave(); },
  methods: {
    readSelection() {
      const sel = window.getSelection();
      const inside = sel && sel.rangeCount && this.$refs.reading && this.$refs.reading.contains(sel.anchorNode);
      this.selected = inside ? sel.toString().trim().slice(0, 500) : '';
    },
    commentOnSelection() {
      this.plan.docs.page.setQuote(this.selected);
      this.selected = '';
      this.$nextTick(() => this.$refs.commentBox && this.$refs.commentBox.focus());
    },
    showComment(id) { const el = document.getElementById('pc-' + id); if (el) el.scrollIntoView({ behavior: 'smooth', block: 'center' }); },
    barDown(e, r) {
      this.moved = false;
      if (!r.canDrag) return;
      this.dragKey = r.key; this.dragPx = 0; this.dragX = e.clientX;
      e.target.setPointerCapture(e.pointerId);
    },
    barMove(e) { if (this.dragKey) { this.dragPx = e.clientX - this.dragX; if (Math.abs(this.dragPx) > 4) this.moved = true; } },
    barUp(e, r) {
      const px = this.dragPx, moved = this.moved;
      this.dragKey = null; this.dragPx = 0; this.moved = false;
      if (moved) r.shift(px); else r.open();
    }
  },
  computed: {
    ...bound.computed,
    plan() { return this.pj && this.pj.plan; },
    liveKey() { const d = this.plan && this.plan.isDocs && this.plan.docs.draft; return d && d.liveOn ? d.id : ''; },
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
.pp-mini { width: 26px; height: 26px; padding: 0; border-radius: 7px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font: inherit; font-size: 12px; }
.pp-mini:disabled { opacity: .35; cursor: default; }
.pp-link { padding: 0; border: 0; background: transparent; color: var(--k-7CC4FF,#7CC4FF); cursor: pointer; font: inherit; font-size: 13px; }
.pp-sr { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); }
</style>
