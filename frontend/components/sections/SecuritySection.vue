<!-- Security & logs -->
<template><template v-if="show.security">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Security &amp; logs</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">Every action in Hub is logged. Rules flag anything suspicious automatically.</p></div>
<div role="tablist" style="display: flex; padding: 3px; border-radius: 12px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)"><button role="tab" @click="secTab.alertsGo" :style="secTab.alerts">Alerts</button><button role="tab" @click="secTab.logsGo" :style="secTab.logs">Audit logs</button><button role="tab" @click="secTab.blocksGo" :style="secTab.blocks">Blocked sign-ins</button><button role="tab" @click="secTab.rulesGo" :style="secTab.rules">Detection rules</button></div>
</div>
<div v-if="!secTab.isBlocks" style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px 10px; padding: 10px 12px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)">
<template v-if="!secTab.isRules">
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Time range</span>
<template v-for="(pr, pr__i) in sec.range.presets" :key="pr__i"><button @click="pr.pick" :aria-pressed="pr.on" :style="`height: 34px; padding: 0 14px; border-radius: 17px; cursor: pointer; font-size: 13px; ${pr.on ? 'border: 1px solid var(--k-B8F36B,#B8F36B); background: var(--k-2C3D34,#2C3D34); color: var(--k-E8EFEA,#E8EFEA); font-weight: 600' : 'border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-C9D6CF,#C9D6CF)'}`">{{ pr.label }}</button></template>
<template v-if="sec.range.isCustom"><label style="display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">From<input type="date" :value="sec.range.from" @input="sec.range.onFrom" style="height: 34px; padding: 0 8px; border-radius: 8px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px"></label>
<label style="display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">To<input type="date" :value="sec.range.to" @input="sec.range.onTo" style="height: 34px; padding: 0 8px; border-radius: 8px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px"></label></template>
</template>
<span style="flex: 1 1 160px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); min-width: 0">{{ secTab.isRules ? 'All detection rules are exported.' : sec.range.summary }}</span>
<button @click="sec.exportXlsx" style="height: 38px; padding: 0 18px; border-radius: 19px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; font-size: 14px; cursor: pointer">Export to Excel</button>
</div>
<template v-if="secTab.isAlerts">
<UsageCard />
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); gap: 12px">
<template v-for="(kp, kp__i) in sec.kpis" :key="kp__i"><div style="padding: 16px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 4px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ kp.label }}</span><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700; color: ${kp.color}`">{{ kp.n }}</span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ kp.sub }}</span></div></template>
</div>
<div style="padding: 18px; border-radius: 18px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; justify-content: space-between; align-items: baseline; flex-wrap: wrap; gap: 8px"><strong style="font-size: 15px">Activity in the last 24 hours</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)"><span style="color: var(--k-FF9B8A,#FF9B8A)">■</span> hours with flagged events</span></div>
<div role="img" aria-label="Events per hour for the last 24 hours" style="height: 80px; display: flex; align-items: flex-end; gap: 3px"><template v-for="(br, br__i) in sec.bars" :key="br__i"><span :style="br.style"></span></template></div>
</div>
<strong style="font-size: 15px">Flagged alerts</strong>
<DataTable :t="alertsTable" />
</template>
<template v-if="secTab.isBlocks"><p style="margin: 0 0 8px; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Sign-ins blocked this week: people of your organization locked out after wrong passwords, and addresses that kept trying (including emails that belong to nobody). Blocks lift by themselves after 2 minutes; this record stays.</p><DataTable :t="blocksTable" /></template>
<template v-if="secTab.isLogs">
<DataTable :t="logsTable" />
</template>
<template v-if="secTab.isRules">
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">When a rule matches, the event is flagged in the logs, an alert is raised and the security team is notified.</span>
<DataTable :t="rulesTable" />
</template>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'SecuritySection',
  ...hubBindings(["alertsTable","logsTable","blocksTable","rulesTable","sec","secTab","show"])
};
</script>
