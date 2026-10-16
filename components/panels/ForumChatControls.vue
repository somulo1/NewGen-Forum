<!-- Who may chat in this forum. Managers can limit it by role or by person. -->
<template>
<div style="padding: 16px; border-radius: 14px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 15px">Chat permissions</strong><span style="font-size: 13px; color: #C9D6CF">{{ chatCtl.summary }}</span><span :style="`font-size: 12px; color: ${chatGate.allowed ? '#9DB0A6' : '#FFB86B'}`">{{ chatCtl.youLine }}</span></div>
<template v-if="chatCtl.canManage">
<label style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px; font-size: 13px; color: #9DB0A6">Who can chat<select :value="chatCtl.mode" @change="chatCtl.setMode" aria-label="Who can chat in this forum" style="height: 40px; min-width: 260px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><template v-for="(mo, mo__i) in chatCtl.modeOpts" :key="mo__i"><option :value="mo.v">{{ mo.l }}</option></template></select></label>
<template v-if="chatCtl.isRoles">
<div role="group" aria-label="Roles allowed to chat" style="display: flex; flex-wrap: wrap; gap: 8px">
<template v-for="(rl, rl__i) in chatCtl.roles" :key="rl__i"><label :style="`display: flex; align-items: center; gap: 8px; height: 34px; padding: 0 12px; border-radius: 17px; cursor: pointer; font-size: 13px; ${rl.on ? 'background: #2C3D34; border: 1px solid #B8F36B' : 'background: #121B17; border: 1px solid #2A3832'}`"><input type="checkbox" :checked="rl.on" @change="rl.toggle" style="accent-color: #B8F36B">{{ rl.name }}</label></template>
</div>
</template>
<span style="font-size: 12px; color: #9DB0A6">Block or allow single people below. A name on the allowed list can chat even when their role cannot.</span>
<div style="display: flex; flex-direction: column; gap: 6px">
<template v-for="(mb, mb__i) in chatCtl.members" :key="mb__i">
<div style="display: flex; align-items: center; gap: 10px; padding: 8px 10px; border-radius: 10px; background: #121B17">
<span aria-hidden="true" :style="`width: 32px; height: 32px; border-radius: 50%; background: ${mb.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 11px`">{{ mb.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px; font-weight: 600">{{ mb.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ mb.forumRole }}</span></div>
<span :style="`font-size: 12px; font-weight: 600; color: ${mb.stateColor}`">{{ mb.stateLabel }}</span>
<button @click="mb.toggle" :aria-label="`${mb.actionLabel}: ${mb.name}`" :style="`height: 30px; padding: 0 12px; border-radius: 15px; cursor: pointer; font-size: 12px; ${mb.can ? 'border: 1px solid #6A3A34; background: transparent; color: #FF9B8A' : 'border: 0; background: #B8F36B; color: #0F1714; font-weight: 600'}`">{{ mb.actionLabel }}</button>
</div>
</template>
</div>
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
