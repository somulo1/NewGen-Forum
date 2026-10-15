<!-- Department / sector / branch / unit dashboard -->
<template><template v-if="show.orgunit">
<div style="max-width: 1240px; display: flex; flex-direction: column; gap: 16px">
<button @click="go.home" style="align-self: flex-start; height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">‹ Home</button>
<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: center; padding: 20px; border-radius: 20px; background: #16201C; border: 1px solid #24322C">
<span :style="`width: 56px; height: 56px; flex: none; border-radius: 16px; background: #22302A; color: ${ou.color}; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 18px`">{{ ou.short }}</span>
<div style="flex: 1 1 280px; min-width: 0; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: center"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">{{ ou.name }}</h1><span :style="ou.badge">{{ ou.type }}</span></div>
<span style="font-size: 13px; color: #9DB0A6">{{ ou.path }}</span>
<span style="font-size: 13px; color: #C9D6CF">Managed by the {{ ou.managerRole }} role · {{ ou.manager }} · {{ ou.members }} people</span>
</div>
<div style="flex: 1 1 260px; max-width: 420px; padding: 12px 14px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 4px"><span style="font-size: 11px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Communication rules</span><template v-for="(cr, cr__i) in ou.commRules" :key="cr__i"><span style="font-size: 12px; color: #C9D6CF; line-height: 1.45">{{ cr.text }}</span></template><span style="font-size: 11px; color: #9DB0A6">Set in the {{ ou.typePlural }} widget settings.</span></div>
</div>

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(210px, 1fr)); gap: 12px">
<template v-for="(kp, kp__i) in ou.kpis" :key="kp__i">
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 12px; color: #9DB0A6">{{ kp.label }}</span>
<div style="display: flex; gap: 14px; flex-wrap: wrap"><template v-for="(pt, pt__i) in kp.parts" :key="pt__i"><div style="display: flex; flex-direction: column"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 24px; font-weight: 700; color: ${pt.color}`">{{ pt.n }}</span><span style="font-size: 11px; color: #9DB0A6">{{ pt.l }}</span></div></template></div>
</div>
</template>
</div>

<div style="padding: 20px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-end">
<div style="flex: 1 1 260px; display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 15px">How we resolve what we receive</strong><span style="font-size: 13px; color: #9DB0A6">Share of received items resolved within their time limit, this month.</span></div>
<label style="display: flex; align-items: center; gap: 8px; font-size: 12px; color: #9DB0A6">Target<input type="number" min="50" max="100" :value="ou.target" @input="ou.setTarget" style="width: 70px; height: 34px; padding: 0 8px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px">%</label>
</div>
<div style="display: flex; flex-wrap: wrap; gap: 24px; align-items: center">
<div style="display: flex; flex-direction: column; gap: 2px; min-width: 150px"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 48px; font-weight: 700; line-height: 1; color: ${ou.perfColor}`">{{ ou.perf }}%</span><span :style="`font-size: 13px; color: ${ou.perfColor}`">{{ ou.perfNote }}</span></div>
<div style="flex: 1 1 320px; display: flex; flex-direction: column; gap: 14px">
<template v-for="(bk, bk__i) in ou.breakdown" :key="bk__i">
<div style="display: flex; flex-direction: column; gap: 6px">
<div style="display: flex; justify-content: space-between; font-size: 13px"><span>{{ bk.label }}</span><span :style="`color: ${bk.color}; font-weight: 600`">{{ bk.pct }}% <span style="color: #9DB0A6; font-weight: 400">· {{ bk.detail }}</span></span></div>
<div style="position: relative; height: 10px; border-radius: 5px; background: #22302A"><div :style="bk.bar"></div><span title="Target" :style="bk.marker"></span></div>
</div>
</template>
</div>
</div>
</div>

<div role="tablist" style="display: flex; flex-wrap: wrap; gap: 4px; border-bottom: 1px solid #24322C"><template v-for="(ot, ot__i) in ou.tabs" :key="ot__i"><button role="tab" @click="ot.go" :style="ot.style">{{ ot.label }} <span style="color: #9DB0A6; font-weight: 400">{{ ot.count }}</span></button></template></div>
<DataTable :t="ouTable" />
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
