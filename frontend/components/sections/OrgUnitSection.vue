<!-- Department / sector / branch / unit dashboard -->
<template><template v-if="show.orgunit && !ou.name">
<div style="padding: 40px 20px; border-radius: 16px; border: 1px dashed var(--k-3F524A,#3F524A); text-align: center; color: var(--k-9DB0A6,#9DB0A6)">You do not have access to this dashboard. Ask an admin to give your role the "View dashboard" right.</div>
</template>
<template v-if="show.orgunit && ou.name && ou.hasDetail">
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
<template v-if="show.orgunit && ou.name && !ou.hasDetail">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<button @click="go.home" style="align-self: flex-start; height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">‹ Home</button>
<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: center; padding: 20px; border-radius: 20px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)">
<span :style="`width: 56px; height: 56px; flex: none; border-radius: 16px; background: var(--k-22302A,#22302A); color: ${ou.color}; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 18px`">{{ ou.short }}</span>
<div style="flex: 1 1 280px; min-width: 0; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: center"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">{{ ou.name }}</h1><span :style="ou.badge">{{ ou.type }}</span></div>
<span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ ou.path }}</span>
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Managed by the {{ ou.managerRole }} role · {{ ou.manager }} · {{ ou.members }} people</span>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 150px), 1fr)); gap: 10px; margin-top: 6px">
<template v-for="(pt, pt__i) in ou.perfTiles" :key="pt__i"><div role="group" :aria-label="`${pt.label} performance`" :style="`padding: 10px 12px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid ${pt.color}; display: flex; flex-direction: column; gap: 2px; min-width: 0`"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 24px; font-weight: 700; color: ${pt.color}`">{{ pt.pct }}%</span><span style="font-size: 13px; font-weight: 600">{{ pt.label }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6); overflow-wrap: anywhere">{{ pt.note }}</span></div></template>
</div>
</div>
</div>

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(210px, 1fr)); gap: 12px">
<template v-for="(kp, kp__i) in ou.kpis" :key="kp__i">
<div style="padding: 16px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ kp.label }}</span>
<div style="display: flex; gap: 14px; flex-wrap: wrap"><template v-for="(pt, pt__i) in kp.parts" :key="pt__i"><div style="display: flex; flex-direction: column"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 24px; font-weight: 700; color: ${pt.color}`">{{ pt.n }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ pt.l }}</span></div></template></div>
</div>
</template>
</div>


<div role="tablist" style="display: flex; flex-wrap: wrap; gap: 4px; border-bottom: 1px solid var(--k-24322C,#24322C)"><template v-for="(ot, ot__i) in ou.tabs" :key="ot__i"><button role="tab" @click="ot.go" :style="ot.style">{{ ot.label }} <span style="color: var(--k-9DB0A6,#9DB0A6); font-weight: 400">{{ ot.count }}</span></button></template></div>
<template v-if="ou.isCats"><TicketCategories /></template>
<DataTable v-if="!ou.isCats" :t="ouTable" />
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'OrgUnitSection',
  ...hubBindings(["go","ou","ouTable","show"])
};
</script>
