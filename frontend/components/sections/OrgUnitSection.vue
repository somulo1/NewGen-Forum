<!-- Department / sector / branch / unit dashboard -->
<template><template v-if="show.orgunit && !ou.name">
<div style="padding: 40px 20px; border-radius: 16px; border: 1px dashed var(--k-3F524A,#3F524A); text-align: center; color: var(--k-9DB0A6,#9DB0A6)">You do not have access to this dashboard. Ask an admin to give your role the "View dashboard" right.</div>
</template>
<template v-if="show.orgunit && ou.name && ou.loading"><PageSkeleton kind="cards" /></template>
<template v-if="show.orgunit && ou.name && !ou.loading && ou.hasDetail">
<div style="display: flex; flex-direction: column; gap: 20px; width: 100%; min-width: 0">
<button @click="ou.detail.back" style="align-self: flex-start; height: 38px; padding: 0 16px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 14px">{{ ou.detail.backLabel }}</button>
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px 14px">
<span style="font-family: ui-monospace, monospace; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ ou.detail.ref }}</span>
<h1 style="margin: 0; flex: 1 1 260px; min-width: 0; font-family: 'Space Grotesk', sans-serif; font-size: clamp(22px, 3.4vw, 30px); letter-spacing: -0.02em; overflow-wrap: anywhere">{{ ou.detail.title }}</h1>
<template v-for="(b, b__i) in ou.detail.badges" :key="b__i"><span :style="`padding: 4px 12px; border-radius: 14px; font-size: 12px; font-weight: 600; border: 1px solid ${b.color}; color: ${b.color}`">{{ b.text }}</span></template>
</div>
<ol aria-label="Progress" style="list-style: none; margin: 0; padding: 0; display: grid; grid-template-columns: repeat(auto-fit, minmax(110px, 1fr)); gap: 10px">
<template v-for="(st, st__i) in ou.detail.steps" :key="st__i"><li style="display: flex; flex-direction: column; gap: 8px; min-width: 0"><span :style="st.barStyle"></span><span style="display: flex; align-items: center; gap: 8px; font-size: 13px; min-width: 0"><span :style="st.dotStyle">{{ st.num }}</span><span :style="st.labelStyle">{{ st.label }}</span></span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ st.note }}</span></li></template>
</ol>
<div style="display: flex; flex-wrap: wrap; gap: 20px; align-items: flex-start">
<div style="flex: 2 1 380px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 18px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 8px"><strong style="font-size: 15px">Description</strong><p style="margin: 0; font-size: 14px; line-height: 1.55; color: var(--k-C9D6CF,#C9D6CF); overflow-wrap: anywhere">{{ ou.detail.about }}</p></div>
<div style="padding: 18px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 14px">
<strong style="font-size: 15px">Activity</strong>
<template v-for="(ac, ac__i) in ou.detail.activity" :key="ac__i">
<div style="display: flex; gap: 10px; min-width: 0"><UserAvatar :name="ac.who" :initials="ac.initials" :color="ac.color" :size="32" :font-size="11"/>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px"><div style="display: flex; flex-wrap: wrap; gap: 4px 8px; align-items: baseline"><strong style="font-size: 13px">{{ ac.who }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ ac.time }}</span></div>
<template v-if="ac.isEvent"><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ ac.text }}</span></template>
<template v-if="ac.isComment"><div style="padding: 8px 12px; border-radius: 4px 12px 12px 12px; background: var(--k-1F2C26,#1F2C26); font-size: 14px; line-height: 1.45; overflow-wrap: anywhere">{{ ac.text }}</div></template></div></div>
</template>
<template v-if="ou.detail.closed"><div role="status" style="padding: 12px 14px; border-radius: 12px; background: var(--k-121B17,#121B17); color: var(--k-9DB0A6,#9DB0A6); font-size: 14px">This is closed. Status changes and comments are turned off.</div></template>
<template v-else-if="ou.detail.noComments"><div role="status" style="padding: 12px 14px; border-radius: 12px; background: var(--k-121B17,#121B17); color: var(--k-9DB0A6,#9DB0A6); font-size: 14px">{{ ou.detail.noCommentsText }}</div></template>
<template v-else><CommentBox placeholder="Write a comment…" send-label="Comment" @send="ou.detail.comment"/></template>
</div>
</div>
<aside aria-label="Details" style="flex: 1 1 280px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 18px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Actions</span>
<template v-for="(act, act__i) in ou.detail.actions" :key="act__i"><button @click="act.run" :style="act.style">{{ act.label }}</button></template>
<template v-if="ou.detail.closed"><span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Closed. Nothing more can be changed.</span></template>
</div>
<div style="padding: 18px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: grid; grid-template-columns: auto 1fr; gap: 8px 14px; font-size: 13px">
<template v-for="(m, m__i) in ou.detail.side" :key="m__i"><span style="color: var(--k-9DB0A6,#9DB0A6)">{{ m.label }}</span><span style="overflow-wrap: anywhere">{{ m.value }}</span></template>
</div>
</aside>
</div>
</div>
</template>
<template v-if="show.orgunit && ou.name && !ou.loading && !ou.hasDetail">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<button @click="go.home" style="align-self: flex-start; height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">‹ Home</button>
<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: center; padding: 20px; border-radius: 20px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)">
<span :style="`width: 56px; height: 56px; flex: none; border-radius: 16px; background: var(--k-22302A,#22302A); color: ${ou.color}; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 18px`">{{ ou.short }}</span>
<div style="flex: 1 1 280px; min-width: 0; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: center"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">{{ ou.name }}</h1><span :style="ou.badge">{{ ou.type }}</span></div>
<span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ ou.path }}</span>
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Managed by the {{ ou.managerRole }} role · {{ ou.manager }} · {{ ou.members }} people</span>
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px; margin-top: 6px"><label style="display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Performance for <select :value="ou.period" @change="ou.onPeriod" aria-label="Period" style="height: 34px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px"><template v-for="o in ou.periodOpts" :key="o.v"><option :value="o.v">{{ o.l }}</option></template></select></label><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ ou.perfFormula }}</span></div>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 150px), 1fr)); gap: 10px; margin-top: 6px">
<template v-for="(pt, pt__i) in ou.perfTiles" :key="pt__i"><div role="group" :aria-label="`${pt.label} performance`" :style="`padding: 10px 12px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid ${pt.color}; display: flex; flex-direction: column; gap: 2px; min-width: 0`"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 24px; font-weight: 700; color: ${pt.color}`">{{ pt.pctText }}</span><span style="font-size: 13px; font-weight: 600">{{ pt.label }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6); overflow-wrap: anywhere">{{ pt.note }}</span></div></template>
</div>
</div>
</div>

<div class="ou-kpis" style="display: grid; grid-template-columns: repeat(auto-fit, minmax(210px, 1fr)); gap: 12px">
<template v-for="(kp, kp__i) in ou.kpis" :key="kp__i">
<div class="ou-kpi" style="padding: 16px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ kp.label }}</span>
<div class="ou-kpi-parts" style="display: flex; gap: 14px; flex-wrap: wrap"><template v-for="(pt, pt__i) in kp.parts" :key="pt__i"><div style="display: flex; flex-direction: column; min-width: 0"><span class="ou-kpi-n" :style="`font-family: 'Space Grotesk', sans-serif; font-size: 24px; font-weight: 700; color: ${pt.color}`">{{ pt.n }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ pt.l }}</span></div></template></div>
</div>
</template>
</div>


<ScrollTabs :active="(ou.tabs.find((t) => t.on) || {}).label"><template v-for="(ot, ot__i) in ou.tabs" :key="ot__i"><button role="tab" :aria-selected="!!ot.on" @click="ot.go" :style="ot.style + ';position:relative'" :aria-label="ot.fresh ? ot.label + ' (new)' : undefined">{{ ot.label }}<span v-if="ot.fresh" class="hub-blink" aria-hidden="true"></span> <span style="color: var(--k-9DB0A6,#9DB0A6); font-weight: 400">{{ ot.count }}</span></button></template></ScrollTabs>
<template v-if="ou.isCats"><TicketCategories /></template>
<div v-if="ou.isComms || ou.isReqs" style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px 10px">
<div v-if="ou.isComms" role="radiogroup" aria-label="Communications" style="display: inline-flex; padding: 3px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-2A3832,#2A3832)"><button v-for="d in ou.commDirs" :key="d.label" type="button" role="radio" :aria-checked="d.on" @click="d.pick" :style="`height: 34px; padding: 0 14px; border-radius: 9px; border: 0; cursor: pointer; font-size: 13px; font-weight: 600; background: ${d.on ? '#B8F36B' : 'transparent'}; color: ${d.on ? '#0F1714' : 'var(--k-C9D6CF,#C9D6CF)'}`">{{ d.label }} · {{ d.count }}</button></div>
<span style="flex: 1"></span>
<button v-for="b in (ou.isReqs ? ou.commSend.filter((x) => x.label === 'New business request') : ou.commSend)" :key="b.label" type="button" @click="b.run" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">{{ b.label }}</button>
</div>
<div v-if="ou.isTks" role="radiogroup" aria-label="Tickets" style="display: inline-flex; padding: 3px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-2A3832,#2A3832); align-self: flex-start"><button v-for="d in ou.tkDirs" :key="d.label" type="button" role="radio" :aria-checked="d.on" @click="d.pick" :style="`height: 34px; padding: 0 14px; border-radius: 9px; border: 0; cursor: pointer; font-size: 13px; font-weight: 600; background: ${d.on ? '#B8F36B' : 'transparent'}; color: ${d.on ? '#0F1714' : 'var(--k-C9D6CF,#C9D6CF)'}`">{{ d.label }} · {{ d.count }}</button></div>
<div v-if="(ou.isReqs && ou.reqTat) || (ou.isComms && ou.commTat)" style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px 10px; padding: 10px 12px; border-radius: 12px; background: var(--k-121B17,#121B17); font-size: 13px">
<span style="flex: 1 1 220px">{{ tatOf(ou).what }} here are due within <strong>{{ tatOf(ou).label }}</strong>.</span>
<template v-if="tatOf(ou).canSet"><label style="display: inline-flex; align-items: center; gap: 6px">Hours<input type="number" min="1" max="2160" :value="tatOf(ou).draft" @input="tatOf(ou).onInput" style="width: 90px; height: 34px; padding: 0 8px; border-radius: 8px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA)"></label><button type="button" @click="tatOf(ou).save" style="height: 34px; padding: 0 14px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Save</button><span style="flex: 1 1 100%; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">1 hour to 3 months (2160 hours). Within 8 hours is high priority, within 24 hours medium, longer low.</span></template>
</div>
<div v-if="ou.isPeople" style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px 10px; font-size: 13px">
<label style="display: inline-flex; align-items: center; gap: 8px">Person<select :value="ou.person" @change="ou.onPerson" style="height: 36px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA)"><option value="">Everyone</option><option v-for="o in ou.personOpts" :key="o.v" :value="o.v">{{ o.l }}</option></select></label>
<span style="color: var(--k-9DB0A6,#9DB0A6)">Only people who were assigned tickets, business requests or communications from other places. On time means resolved, or escalated on, within the place's turnaround, out of what was decided in the period. Items still within their time show as not due yet.</span>
</div>
<div v-if="ou.send" class="hub-overlay" style="position: fixed; inset: 0; z-index: 80; background: rgba(5,10,8,0.6); display: flex; align-items: center; justify-content: center; padding: 16px" @click.self="ou.send.close">
<form role="dialog" aria-modal="true" :aria-label="ou.send.title" novalidate @submit.prevent="ou.send.submit" style="width: 560px; max-width: 100%; max-height: 100%; overflow-y: auto; box-sizing: border-box; display: flex; flex-direction: column; gap: 12px; padding: 20px; border-radius: 20px; background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); border: 1px solid var(--k-2A3832,#2A3832)">
<h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 20px">{{ ou.send.title }}</h2>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 200px), 1fr)); gap: 10px">
<label v-if="ou.send.levels.length > 1" style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); min-width: 0">Send to a<select :value="ou.send.level" @change="ou.send.onLevel" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; width: 100%; box-sizing: border-box; min-width: 0"><option v-for="o in ou.send.levels" :key="o.v" :value="o.v">{{ o.l }}</option></select></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); min-width: 0">Send to<select :value="ou.send.place" @change="ou.send.onPlace" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; width: 100%; box-sizing: border-box; min-width: 0"><option value="" disabled>Choose…</option><option v-for="o in ou.send.places" :key="o" :value="o">{{ o }}</option></select></label>
</div>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); min-width: 0">Subject<input type="text" maxlength="200" :value="ou.send.subject" @input="ou.send.onSubject" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; width: 100%; box-sizing: border-box; min-width: 0"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); min-width: 0">Details<textarea rows="5" maxlength="4000" :value="ou.send.details" @input="ou.send.onDetails" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; width: 100%; box-sizing: border-box; min-width: 0; height: auto; padding: 10px; resize: vertical"></textarea></label>
<span v-if="ou.send.tatNote" style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ ou.send.tatNote }}</span>
<span v-if="ou.send.err" role="alert" style="font-size: 13px; color: var(--k-FF9B8A,#FF9B8A)">{{ ou.send.err }}</span>
<div style="display: flex; justify-content: flex-end; gap: 8px"><button type="button" @click="ou.send.close" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Cancel</button><button type="submit" :disabled="ou.send.busy" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">{{ ou.send.busy ? 'Sending…' : 'Send' }}</button></div>
</form>
</div>
<DataTable v-if="!ou.isCats" :t="ouTable" />
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'OrgUnitSection',
  ...hubBindings(["go","ou","ouTable","show"]),
  methods: { tatOf(ou) { return ou.isReqs ? ou.reqTat : ou.commTat; } }
};
</script>
