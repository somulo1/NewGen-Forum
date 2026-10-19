<!-- Approvals inbox -->
<template><template v-if="show.approvals">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Approvals</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">Requests move through each approval level in order. Some need one approver, some two or three.</p></div>
<div role="tablist" style="display: flex; padding: 3px; border-radius: 12px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)"><button role="tab" @click="apTab.waitingGo" :style="apTab.waiting">Waiting ({{ pendingApprovals }})</button><button role="tab" @click="apTab.allGo" :style="apTab.all">All</button></div>
</div>
<DataTable :t="approvalsTable" />
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'ApprovalsSection',
  ...hubBindings(["apTab","approvalsTable","pendingApprovals","show"])
};
</script>
