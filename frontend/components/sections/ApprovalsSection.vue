<!-- Approvals inbox -->
<template><template v-if="show.approvals">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Approvals</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">Requests move through each approval level in order. Some need one approver, some two or three.</p></div>
<div role="tablist" style="display: flex; padding: 3px; border-radius: 12px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)"><button role="tab" @click="apTab.waitingGo" :style="apTab.waiting">Waiting ({{ pendingApprovals }})</button><button role="tab" @click="apTab.allGo" :style="apTab.all">All</button></div>
</div>
<template v-for="g in apGroups" :key="g.key">
<section :aria-label="g.title" style="border-radius: 16px; border: 2px solid var(--k-6B7A72,#6B7A72); background: var(--k-16201C,#16201C); overflow: hidden">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px 14px; padding: 14px 16px">
<div style="flex: 1 1 280px; min-width: 0"><strong style="font-size: 15px">{{ g.title }}</strong><div style="margin-top: 2px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ g.sub }}</div></div>
<button type="button" :aria-expanded="g.open" @click="g.toggle" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 1px solid var(--k-6B7A72,#6B7A72); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px; font-weight: 600">{{ g.toggleLabel }} {{ g.open ? '▴' : '▾' }}</button>
<button type="button" :disabled="g.busy" @click="g.approve" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; font-size: 13px; font-weight: 700">{{ g.busy ? 'Working…' : g.approveLabel }}</button>
<button type="button" :disabled="g.busy" @click="g.reject" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 1px solid #FF9B8A; background: transparent; color: #FF9B8A; cursor: pointer; font-size: 13px; font-weight: 600">{{ g.rejectLabel }}</button>
</div>
<template v-if="g.open">
<div style="border-top: 1px solid var(--k-24322C,#24322C); padding: 10px 16px; display: flex; flex-wrap: wrap; align-items: center; gap: 10px 16px">
<label style="display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer"><input type="checkbox" :checked="g.allChecked" @change="g.toggleAll"> Select all waiting ({{ g.count }})</label>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ g.selectedCount }} selected{{ g.selectedCount ? '' : '. With none ticked, approving covers all of them.' }}</span>
<input type="text" :value="g.note" @input="g.onNote" placeholder="Comment (optional)" aria-label="Comment for this decision" style="flex: 1 1 220px; min-width: 0; height: 36px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px">
</div>
<template v-for="it in g.items" :key="it.uid">
<label style="display: flex; align-items: center; gap: 12px; padding: 9px 16px; border-top: 1px solid var(--k-24322C,#24322C); cursor: pointer; font-size: 14px">
<input type="checkbox" :checked="it.checked" @change="it.toggle">
<span :style="`flex: none; width: 18px; text-align: center; font-weight: 700; color: ${it.signColor}`">{{ it.sign }}</span>
<span style="flex: 1; min-width: 0">{{ it.right }} <span style="color: var(--k-9DB0A6,#9DB0A6); font-size: 12px">· {{ it.module }} · {{ it.verbLabel }}</span></span>
<span style="flex: none; padding: 3px 10px; border-radius: 10px; font-size: 11px; font-weight: 600; background: #3A2F18; color: #FFB86B">Waiting</span>
</label>
</template>
</template>
</section>
</template>
<DataTable :t="approvalsTable" />
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'ApprovalsSection',
  ...hubBindings(["apGroups","apTab","approvalsTable","pendingApprovals","show"])
};
</script>
