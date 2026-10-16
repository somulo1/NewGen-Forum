<!-- Organization structure -->
<template><template v-if="show.structure">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Organization structure</h1><p style="margin: 4px 0 0; color: #9DB0A6; line-height: 1.5">Built from the structure widgets you use. Add or remove Sectors, Branches, Units and Departments from Home → Add widget and this view rearranges itself.</p></div>
<div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center"><template v-for="(ch, ch__i) in org.chain" :key="ch__i"><span :style="ch.style">{{ ch.label }}</span><template v-if="ch.arrow"><span style="color: #4A5C54">›</span></template></template></div>
<template v-if="org.deptNote"><span style="font-size: 13px; color: #C9D6CF">{{ org.deptNoteText }}</span></template>
<div style="border-radius: 16px; border: 1px solid #24322C; background: #16201C; overflow: hidden">
<template v-for="(on, on__i) in org.rows" :key="on__i">
<div :style="on.rowStyle">
<span :style="on.indent"></span>
<span :style="on.badge">{{ on.type }}</span>
<div style="flex: 1 1 200px; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px">{{ on.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ on.sub }}</span></div>
<template v-if="on.canOpen"><button @click="on.open" style="height: 34px; padding: 0 14px; border-radius: 17px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px; font-weight: 600">Dashboard</button></template>
</div>
</template>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'StructureSection',
  ...hubBindings(["org","show"])
};
</script>
