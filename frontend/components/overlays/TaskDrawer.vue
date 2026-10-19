<!-- Project task drawer -->
<template><template v-if="taskOpen">
<div class="hub-overlay hub-right-drawer" style="position: fixed; inset: 0; z-index: 64; background: rgba(5,10,8,0.55); display: flex; justify-content: flex-end">
<div role="dialog" aria-label="Task details" style="width: 520px; max-width: 100%; height: 100%; background: var(--k-16201C,#16201C); border-left: 1px solid var(--k-2A3832,#2A3832); display: flex; flex-direction: column">
<div style="flex: none; display: flex; align-items: center; gap: 12px; padding: 16px 20px; border-bottom: 1px solid var(--k-24322C,#24322C)">
<span :style="`width: 34px; height: 34px; border-radius: 10px; background: ${td.pcolor}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 12px`">{{ td.pinitials }}</span>
<span style="flex: 1; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ td.project }} · Task</span>
<button aria-label="Close task" @click="closeTask" style="width: 38px; height: 38px; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 20px; display: flex; flex-direction: column; gap: 18px">
<h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 24px; line-height: 1.25">{{ td.title }}</h2>
<div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center">
<template v-if="td.canBack"><button @click="td.back" style="height: 36px; padding: 0 12px; border-radius: 18px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px">◂ {{ td.prevCol }}</button></template>
<span :style="td.colStyle">{{ td.col }}</span>
<template v-if="td.canFwd"><button @click="td.fwd" style="height: 36px; padding: 0 12px; border-radius: 18px; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; font-size: 13px; font-weight: 600">{{ td.nextCol }} ▸</button></template>
</div>
<template v-if="td.overdue">
<div role="alert" style="padding: 12px 14px; border-radius: 12px; background: var(--k-3A1E1A,#3A1E1A); border: 1px solid var(--k-6A3A34,#6A3A34); display: flex; flex-wrap: wrap; gap: 10px; align-items: center"><span style="flex: 1 1 220px; font-size: 14px; color: #FFC2B8"><strong>Overdue by {{ td.overDays }}.</strong> {{ td.reminderNote }}</span><button @click="td.remind" style="height: 34px; padding: 0 12px; border-radius: 17px; border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 12px">Remind owners</button></div>
</template>
<div style="padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Timeline</span>
<div style="display: flex; justify-content: space-between; font-size: 13px"><span>Start · <strong>{{ td.startLabel }}</strong></span><span>Due · <strong :style="`color: ${td.dueColor}`">{{ td.dueLabel }}</strong></span></div>
<div style="height: 8px; border-radius: 4px; background: var(--k-22302A,#22302A); overflow: hidden"><div :style="td.timeBar"></div></div>
<span :style="`font-size: 12px; color: ${td.dueColor}`">{{ td.timeNote }}</span>
</div>
<div style="padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Where this task fits</span>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Objective<input type="text" :value="td.objective" :disabled="!td.canEdit" @change="td.setObjective" placeholder="What should this task achieve?" style="height: 38px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Sprint<select :value="td.sprintId" :disabled="!td.canEdit" @change="td.setSprint" style="height: 38px; padding: 0 8px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px"><template v-for="(so, so__i) in td.sprintOpts" :key="so__i"><option :value="so.v">{{ so.l }}</option></template></select></label>
<div style="display: flex; flex-direction: column; gap: 2px; font-size: 13px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Milestone (from the sprint)</span><strong style="font-weight: 600">{{ td.milestoneName }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Due {{ td.milestoneDue }} · {{ td.milestoneState }} · sprint runs {{ td.sprintDates }}</span></div>
<label style="display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Due date<input type="date" :value="td.dueISO" :min="td.minDue" :max="td.maxDue" :disabled="!td.canEdit" @change="td.setDue" style="height: 36px; padding: 0 8px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px"></label>
</div>
<div style="display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px">
<div style="padding: 12px; border-radius: 12px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Primary owner</span><div style="display: flex; gap: 8px; align-items: center"><span aria-hidden="true" :style="`width: 30px; height: 30px; border-radius: 50%; background: ${td.pColor}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 11px`">{{ td.pIni }}</span><strong style="font-size: 14px; font-weight: 600">{{ td.primary }}</strong></div></div>
<label style="padding: 12px; border-radius: 12px; background: var(--k-121B17,#121B17); display: flex; flex-direction: column; gap: 8px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Secondary owner<select :value="td.secondary" @change="td.setSecondary" style="height: 36px; padding: 0 8px; border-radius: 8px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px"><option value="">None</option><template v-for="(to, to__i) in td.teamOpts" :key="to__i"><option :value="to.name">{{ to.name }}</option></template></select></label>
</div>
<div style="display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Description</span><p style="margin: 0; font-size: 15px; line-height: 1.6; color: var(--k-C9D6CF,#C9D6CF)">{{ td.desc }}</p></div>
<div style="display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Linked tasks</span>
<template v-if="td.noLinks"><span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Not linked to any task yet.</span></template>
<template v-for="(lk, lk__i) in td.links" :key="lk__i"><div style="display: flex; gap: 10px; align-items: center; padding: 10px 12px; border-radius: 10px; background: var(--k-121B17,#121B17)"><button @click="lk.open" style="flex: 1; min-width: 0; padding: 0; border: 0; background: transparent; color: var(--k-E8EFEA,#E8EFEA); text-align: left; cursor: pointer; display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 13px; font-weight: 600">{{ lk.title }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ lk.col }} · due {{ lk.due }}</span></button><button :aria-label="`Unlink ${lk.title}`" @click="lk.unlink" style="width: 30px; height: 30px; border-radius: 8px; border: 0; background: transparent; color: var(--k-FF9B8A,#FF9B8A); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button></div></template>
<label style="display: flex; gap: 8px; align-items: center; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Link to<select value="" @change="td.addLink" style="flex: 1; min-width: 0; height: 38px; padding: 0 8px; border-radius: 8px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px"><option value="">Choose a task…</option><template v-for="(lo, lo__i) in td.linkOpts" :key="lo__i"><option :value="lo.i">{{ lo.title }}</option></template></select></label>
</div>
<div style="display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Reminders</span>
<template v-for="(rmd, rmd__i) in td.reminders" :key="rmd__i"><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">• {{ rmd.text }}</span></template>
</div>
</div>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'TaskDrawer',
  ...hubBindings(["closeTask","taskOpen","td"])
};
</script>
