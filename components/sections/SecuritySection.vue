<!-- Security & logs -->
<template><template v-if="show.security">
<div style="max-width: 1280px; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Security &amp; logs</h1><p style="margin: 4px 0 0; color: #9DB0A6">Every action in Hub is logged. Rules flag anything suspicious automatically.</p></div>
<div role="tablist" style="display: flex; padding: 3px; border-radius: 12px; background: #16201C; border: 1px solid #24322C"><button role="tab" @click="secTab.alertsGo" :style="secTab.alerts">Alerts</button><button role="tab" @click="secTab.logsGo" :style="secTab.logs">Audit logs</button><button role="tab" @click="secTab.rulesGo" :style="secTab.rules">Detection rules</button></div>
</div>
<template v-if="secTab.isAlerts">
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); gap: 12px">
<template v-for="(kp, kp__i) in sec.kpis" :key="kp__i"><div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 4px"><span style="font-size: 12px; color: #9DB0A6">{{ kp.label }}</span><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700; color: ${kp.color}`">{{ kp.n }}</span><span style="font-size: 12px; color: #9DB0A6">{{ kp.sub }}</span></div></template>
</div>
<div style="padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; justify-content: space-between; align-items: baseline; flex-wrap: wrap; gap: 8px"><strong style="font-size: 15px">Activity in the last 24 hours</strong><span style="font-size: 12px; color: #9DB0A6"><span style="color: #FF9B8A">■</span> hours with flagged events</span></div>
<div role="img" aria-label="Events per hour for the last 24 hours" style="height: 80px; display: flex; align-items: flex-end; gap: 3px"><template v-for="(br, br__i) in sec.bars" :key="br__i"><span :style="br.style"></span></template></div>
</div>
<strong style="font-size: 15px">Flagged alerts</strong>
<DataTable :t="alertsTable" />
</template>
<template v-if="secTab.isLogs">
<DataTable :t="logsTable" />
<button @click="sec.export" style="align-self: flex-start; height: 40px; padding: 0 16px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Export all logs (CSV)</button>
</template>
<template v-if="secTab.isRules">
<span style="font-size: 13px; color: #C9D6CF">When a rule matches, the event is flagged in the logs, an alert is raised and the security team is notified.</span>
<DataTable :t="rulesTable" />
</template>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'SecuritySection',
  ...hubBindings(["alertsTable","logsTable","rulesTable","sec","secTab","show"])
};
</script>
