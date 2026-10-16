<!-- People directory -->
<template><template v-if="show.people">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">People</h1><p style="margin: 4px 0 0; color: #9DB0A6">{{ peopleSummary }}</p></div>
<template v-if="onb.canOnboard"><div style="display: flex; flex-wrap: wrap; gap: 8px"><button @click="onb.openBulk" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">Bulk onboarding</button><button @click="onb.openOne" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Onboard person</button></div></template>
</div>
<DataTable :t="peopleTable" />
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'PeopleSection',
  ...hubBindings(["onb","peopleSummary","peopleTable","show"])
};
</script>
