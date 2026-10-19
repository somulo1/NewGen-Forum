<!-- Communications (memos, newsletter, announcements) -->
<template><template v-if="show.comms">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Communications</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">Send announcements, memos and newsletters. You choose where each one goes.</p></div>
<div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(cn, cn__i) in commNew" :key="cn__i"><button @click="cn.run" :style="cn.style">{{ cn.label }}</button></template></div>
</div>
<DataTable :t="commsTable" />
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'CommsSection',
  ...hubBindings(["commNew","commsTable","show"])
};
</script>
