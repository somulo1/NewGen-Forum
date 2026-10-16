<!-- Call history of the open chat (forum or private). Works in the docked panel and the expanded screens. -->
<template>
<div style="display: flex; flex-direction: column; gap: 8px; min-width: 0">
<strong style="font-size: 13px; color: #9DB0A6">Calls with {{ callHistory.who }}</strong>
<template v-if="callHistory.empty"><p style="margin: 8px 0; font-size: 13px; color: #9DB0A6">No calls yet.</p></template>
<template v-for="r in callHistory.rows" :key="r.id">
<div style="display: flex; align-items: center; gap: 10px; padding: 8px 10px; border-radius: 10px; background: #16201C; border: 1px solid #24322C; min-width: 0">
<span aria-hidden="true" :style="`width: 30px; height: 30px; flex: none; border-radius: 50%; background: #1D2A25; color: ${r.color}; display: flex; align-items: center; justify-content: center; font-size: 14px`">{{ r.arrow }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><span :style="`font-size: 13px; font-weight: 600; color: ${r.color}; overflow: hidden; text-overflow: ellipsis; white-space: nowrap`">{{ r.title }}</span><span style="font-size: 11px; color: #9DB0A6">{{ r.sub }}</span></div>
</div>
</template>
</div>
</template>

<script>
import { hubBindings } from '~/composables/useHub';
export default { name: 'CallHistory', ...hubBindings(['callHistory']) };
</script>
