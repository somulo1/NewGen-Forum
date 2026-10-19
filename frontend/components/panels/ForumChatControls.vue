<!-- Who may chat in this forum. Managers can limit it by role or by person. -->
<template>
<div style="padding: 16px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 15px">Chat permissions</strong><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ chatCtl.summary }}</span><span :style="`font-size: 12px; color: ${chatGate.allowed ? 'var(--k-9DB0A6,#9DB0A6)' : 'var(--k-FFB86B,#FFB86B)'}`">{{ chatCtl.youLine }}</span></div>
<template v-if="chatCtl.canManage">
<label style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Who can chat<select :value="chatCtl.mode" @change="chatCtl.setMode" aria-label="Who can chat in this forum" style="height: 40px; min-width: 260px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><template v-for="(mo, mo__i) in chatCtl.modeOpts" :key="mo__i"><option :value="mo.v">{{ mo.l }}</option></template></select></label>
<template v-if="chatCtl.isRoles">
<div role="group" aria-label="Roles allowed to chat" style="display: flex; flex-wrap: wrap; gap: 8px">
<template v-for="(rl, rl__i) in chatCtl.roles" :key="rl__i"><label :style="`display: flex; align-items: center; gap: 8px; height: 34px; padding: 0 12px; border-radius: 17px; cursor: pointer; font-size: 13px; ${rl.on ? 'background: var(--k-2C3D34,#2C3D34); border: 1px solid var(--k-B8F36B,#B8F36B)' : 'background: var(--k-121B17,#121B17); border: 1px solid var(--k-2A3832,#2A3832)'}`"><input type="checkbox" :checked="rl.on" @change="rl.toggle" style="accent-color: var(--k-B8F36B,#B8F36B)">{{ rl.name }}</label></template>
</div>
</template>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">To block or allow one person, use the button on their card below. A name on the allowed list can chat even when their role cannot.</span>
</template>
</div>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'ForumChatControls',
  ...hubBindings(["chatCtl","chatGate"])
};
</script>
