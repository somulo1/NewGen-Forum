<!-- Call history of the open chat (forum or private). Works in the docked panel and the expanded screens. -->
<template>
<div style="display: flex; flex-direction: column; gap: 8px; min-width: 0">
<strong style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Calls with {{ callHistory.who }}</strong>
<template v-if="callHistory.empty"><p style="margin: 8px 0; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">No calls yet.</p></template>
<template v-for="r in callHistory.rows" :key="r.id">
<div style="display: flex; align-items: center; gap: 10px; padding: 8px 10px; border-radius: 10px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); min-width: 0">
<span aria-hidden="true" :style="`width: 30px; height: 30px; flex: none; border-radius: 50%; background: var(--k-1D2A25,#1D2A25); color: ${r.color}; display: flex; align-items: center; justify-content: center`"><svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 4h3.5l1.6 4-2 1.4a11 11 0 0 0 5.5 5.5l1.4-2 4 1.6V18a2 2 0 0 1-2 2A15 15 0 0 1 3 6a2 2 0 0 1 2-2z" stroke-width="1.6"/><template v-if="r.icon === 'in'"><path d="M21 3l-6 6M15 5v4h4"/></template><template v-else-if="r.icon === 'out'"><path d="M15 9l6-6M17 3h4v4"/></template><template v-else-if="r.icon === 'missedIn'"><path d="M21 3l-6 6M15 5v4h4"/><path d="M16.5 14.5l4.5 4.5M21 14.5L16.5 19" stroke-width="1.6"/></template><template v-else-if="r.icon === 'missedOut'"><path d="M15 9l6-6M17 3h4v4"/><path d="M16.5 14.5l4.5 4.5M21 14.5L16.5 19" stroke-width="1.6"/></template><template v-else-if="r.icon === 'declined'"><path d="M15 3l6 6M21 3l-6 6"/></template><template v-else><circle cx="18" cy="6" r="3.5"/><path d="M15.5 8.5l5-5"/></template></svg></span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><span :style="`font-size: 13px; font-weight: 600; color: ${r.color}; overflow: hidden; text-overflow: ellipsis; white-space: nowrap`">{{ r.title }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ r.sub }}</span></div>
</div>
</template>
</div>
</template>

<script>
import { hubBindings } from '~/composables/useHub';
export default { name: 'CallHistory', ...hubBindings(['callHistory']) };
</script>
