<!-- Ticket desk: every ticket received, when the organization has no department, unit, branch or sector -->
<template><template v-if="showDeskList">
<div style="width: 100%; display: flex; flex-direction: column; gap: 18px">
<div style="display: flex; flex-direction: column; gap: 4px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em; line-height: 1.15">Ticket desk</h1><p style="margin: 0; max-width: 760px; line-height: 1.5; color: var(--k-9DB0A6,#9DB0A6)">Every ticket sent to the organization lands here, most urgent first. With no departments, units, branches or sectors, the desk handles them all.</p></div>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 12px">
<template v-for="(ds, ds__i) in deskStats" :key="ds__i">
<div :style="`position: relative; overflow: hidden; padding: 12px 16px 12px 20px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; align-items: baseline; gap: 10px`"><span :style="`position: absolute; left: 0; top: 0; bottom: 0; width: 5px; background: ${ds.color}`"></span><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700; line-height: 1; color: ${ds.color}`">{{ ds.n }}</span><span style="font-size: 13px; font-weight: 600; color: var(--k-C9D6CF,#C9D6CF)">{{ ds.label }}</span></div>
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
<section aria-label="Tickets received" style="display: flex; flex-direction: column; gap: 10px"><strong style="font-size: 16px">Tickets received</strong><DataTable :t="deskTable" /></section>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'TicketDeskSection',
  ...hubBindings(["deskCats","deskStats","deskTable","showDeskList"])
};
</script>
