<!-- Approvals inbox -->
<template><template v-if="show.approvals">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Approvals</h1><p style="margin: 4px 0 0; color: #9DB0A6">Requests move through each approval level in order. Some need one approver, some two or three.</p></div>
<div role="tablist" style="display: flex; padding: 3px; border-radius: 12px; background: #16201C; border: 1px solid #24322C"><button role="tab" @click="apTab.waitingGo" :style="apTab.waiting">Waiting ({{ pendingApprovals }})</button><button role="tab" @click="apTab.allGo" :style="apTab.all">All</button></div>
</div>
<template v-if="noRequests"><div style="padding: 40px 20px; border-radius: 16px; border: 1px dashed #3F524A; text-align: center; color: #9DB0A6">Nothing is waiting for approval.</div></template>
<template v-for="(rq, rq__i) in requests" :key="rq__i">
<article style="padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: center">
<span style="font-family: 'Space Grotesk', sans-serif; color: #9DB0A6; font-size: 13px">{{ rq.id }}</span>
<span style="padding: 2px 8px; border-radius: 8px; background: #22302A; color: #C9D6CF; font-size: 11px; font-weight: 600">{{ rq.modName }}</span>
<strong style="flex: 1 1 260px; font-size: 16px">{{ rq.title }}</strong>
<span :style="rq.statusStyle">{{ rq.statusLabel }}</span>
</div>
<span style="font-size: 13px; color: #9DB0A6">Requested by {{ rq.by }} · {{ rq.when }} · {{ rq.levelsLabel }}</span>
<template v-if="rq.hasNote"><p style="margin: 0; font-size: 14px; color: #C9D6CF; line-height: 1.5">{{ rq.note }}</p></template>
<ol style="list-style: none; margin: 0; padding: 0; display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 10px">
<template v-for="(sp, sp__i) in rq.steps" :key="sp__i">
<li style="display: flex; flex-direction: column; gap: 6px"><span :style="sp.bar"></span><span style="display: flex; align-items: center; gap: 8px; font-size: 13px"><span :style="sp.dot">{{ sp.n }}</span><strong style="font-weight: 600">{{ sp.who }}</strong></span><span :style="`font-size: 12px; color: ${sp.color}`">{{ sp.state }}</span></li>
</template>
</ol>
<template v-if="rq.canAct">
<div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center"><span style="flex: 1 1 200px; font-size: 13px; color: #FFB86B">Waiting for you as {{ rq.currentApprover }}</span><button @click="rq.reject" style="height: 40px; padding: 0 18px; border-radius: 10px; border: 1px solid #5A3A36; background: transparent; color: #FF9B8A; font-weight: 600; cursor: pointer">Reject</button><button @click="rq.approve" style="height: 40px; padding: 0 22px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">{{ rq.approveLabel }}</button></div>
</template>
</article>
</template>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'ApprovalsSection',
  ...hubBindings(["apTab","noRequests","pendingApprovals","requests","show"])
};
</script>
