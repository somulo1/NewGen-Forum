<!-- Ticket desk: every ticket received, when the organization has no department, unit, branch or sector -->
<template><template v-if="showDeskList">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Ticket desk</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">Every ticket sent to the organization lands here, sorted by priority and how soon it is due. Your organization has no departments, units, branches or sectors, so the desk handles them all.</p></div>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 10px">
<template v-for="(ds, ds__i) in deskStats" :key="ds__i">
<div style="padding: 14px 16px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 2px"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 28px; font-weight: 700; color: ${ds.color}`">{{ ds.n }}</span><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ ds.label }}</span></div>
</template>
</div>
<template v-if="deskCats.length">
<div style="display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">By category</span>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 10px">
<template v-for="(dc, dc__i) in deskCats" :key="dc__i">
<div style="padding: 12px 14px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 14px">{{ dc.name }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ dc.open }} open of {{ dc.total }}<template v-if="dc.overdue"> · <span style="color: var(--k-FF9B8A,#FF9B8A)">{{ dc.overdue }} overdue</span></template></span></div>
</template>
</div>
</div>
</template>
<TicketCategories />
<DataTable :t="deskTable" />
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'TicketDeskSection',
  ...hubBindings(["deskCats","deskStats","deskTable","showDeskList"])
};
</script>
