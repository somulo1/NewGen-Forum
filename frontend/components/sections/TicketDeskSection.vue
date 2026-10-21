<!-- Ticket desk: every ticket received, when the organization has no department, unit, branch or sector -->
<template><template v-if="showDeskList">
<div style="width: 100%; display: flex; flex-direction: column; gap: 18px">
<div style="display: flex; flex-direction: column; gap: 4px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em; line-height: 1.15">Ticket desk</h1><p style="margin: 0; max-width: 760px; line-height: 1.5; color: var(--k-9DB0A6,#9DB0A6)">Every ticket sent to the organization lands here, most urgent first. With no departments, units, branches or sectors, the desk handles them all.</p></div>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 12px">
<template v-for="(ds, ds__i) in deskStats" :key="ds__i">
<div :style="`position: relative; overflow: hidden; padding: 12px 16px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; align-items: baseline; gap: 10px`"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700; line-height: 1; color: ${ds.color}`">{{ ds.n }}</span><span style="font-size: 13px; font-weight: 600; color: var(--k-C9D6CF,#C9D6CF)">{{ ds.label }}</span></div>
</template>
</div>
<div role="tablist" aria-label="Ticket desk" style="display: flex; flex-wrap: wrap; gap: 4px; border-bottom: 1px solid var(--k-24322C,#24322C)">
<button role="tab" :aria-selected="tab === 'tickets'" @click="tab = 'tickets'" :style="tabStyle('tickets')">Tickets received <span style="color: var(--k-9DB0A6,#9DB0A6); font-weight: 400">{{ deskTable.rows.length }}</span></button>
<button role="tab" :aria-selected="tab === 'categories'" @click="tab = 'categories'" :style="tabStyle('categories')">Ticket categories <span style="color: var(--k-9DB0A6,#9DB0A6); font-weight: 400">{{ catMgr.rows ? catMgr.rows.length : 0 }}</span></button>
</div>
<template v-if="tab === 'tickets'">
<template v-if="deskCats.length">
<div style="display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">By category</span>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 10px">
<template v-for="(dc, dc__i) in deskCats" :key="dc__i">
<div style="padding: 12px 14px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 14px">{{ dc.name }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ dc.open }} open of {{ dc.total }}<template v-if="dc.overdue"> · <span style="color: var(--k-FF9B8A,#FF9B8A)">{{ dc.overdue }} overdue</span></template></span></div>
</template>
</div>
</div>
</template>
<DataTable :t="deskTable" />
</template>
<template v-else><TicketCategories /></template>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'TicketDeskSection',
  data() { return { tab: 'tickets' }; },
  methods: {
    // the same tab look as a department's page: a thin line under the open one
    tabStyle(k) { const on = this.tab === k; return 'min-height:44px;padding:0 16px;border:0;background:transparent;cursor:pointer;font-size:14px;white-space:nowrap;font-family:inherit;border-bottom:2px solid ' + (on ? '#B8F36B' : 'transparent') + ';color:' + (on ? 'var(--k-E8EFEA,#E8EFEA)' : 'var(--k-9DB0A6,#9DB0A6)') + ';font-weight:' + (on ? '600' : '500'); }
  },
  ...hubBindings(["deskCats","deskStats","deskTable","showDeskList","catMgr"])
};
</script>
