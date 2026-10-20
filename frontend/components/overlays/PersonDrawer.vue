<!-- Person profile drawer -->
<template><template v-if="personOpen">
<div class="hub-overlay hub-right-drawer" style="position: fixed; inset: 0; z-index: 64; background: rgba(5,10,8,0.55); display: flex; justify-content: flex-end">
<div role="dialog" :aria-label="psn.name" style="width: 420px; max-width: 100%; height: 100%; overflow-y: auto; background: var(--k-16201C,#16201C); border-left: 1px solid var(--k-2A3832,#2A3832); display: flex; flex-direction: column">
<div :style="`position: relative; height: 110px; flex: none; background: ${psn.cover}`"><button aria-label="Close" @click="closePerson" style="position: absolute; top: 12px; right: 12px; width: 38px; height: 38px; border-radius: 50%; border: 0; background: rgba(10,16,13,0.6); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button></div>
<div style="padding: 0 22px 22px; margin-top: -42px; display: flex; flex-direction: column; gap: 14px">
<span style="position: relative; flex: none"><UserAvatar :name="psn.name" :initials="psn.initials" :color="psn.color" :size="84" :font-size="28" border="4px solid #16201C"/><span :style="`position: absolute; right: 2px; bottom: 2px; width: 16px; height: 16px; border-radius: 50%; background: ${psn.statusColor}; border: 3px solid var(--k-16201C,#16201C)`"></span></span>
<div style="display: flex; flex-direction: column; gap: 4px"><h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 22px">{{ psn.name }}</h2><span style="font-size: 14px; color: var(--k-C9D6CF,#C9D6CF)">{{ psn.title }}</span><span :style="`font-size: 13px; color: ${psn.statusColor}`">● {{ psn.status }}</span></div>
<template v-if="psn.canMessage"><div style="display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px"><button @click="psn.message" style="height: 40px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Message</button><button @click="psn.voice" style="height: 40px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Call</button><button @click="psn.video" style="height: 40px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Video</button></div></template>
<div style="display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 10px 16px; padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17); font-size: 13px">
<span style="color: var(--k-9DB0A6,#9DB0A6)">Department</span><span>{{ psn.dept }}</span>
<span style="color: var(--k-9DB0A6,#9DB0A6)">Location</span><span>{{ psn.loc }}</span>
<span style="color: var(--k-9DB0A6,#9DB0A6)">Reports to</span><span>{{ psn.manager }}</span>
<span style="color: var(--k-9DB0A6,#9DB0A6)">Role in Hub</span><span>{{ psn.hubRole }}</span>
<span style="color: var(--k-9DB0A6,#9DB0A6)">Email</span><span>{{ psn.email }}</span>
<template v-if="psn.securityCode"><span style="color: var(--k-9DB0A6,#9DB0A6)">Security code</span><span style="font-family: monospace; font-size: 12px; word-spacing: 4px; overflow-wrap: anywhere">{{ psn.securityCode }}</span></template>
</div>
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Skills</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(sk, sk__i) in psn.skills" :key="sk__i"><span style="padding: 6px 12px; border-radius: 16px; background: var(--k-1F2C26,#1F2C26); border: 1px solid var(--k-2A3832,#2A3832); font-size: 12px">{{ sk.name }}</span></template></div>
<template v-if="psn.hasAwards">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">Badges &amp; awards</span>
<div style="display: flex; flex-direction: column; gap: 8px"><template v-for="(pa, pa__i) in psn.awards" :key="pa__i"><div style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: var(--k-24220F,#24220F); border: 1px solid var(--k-4A4426,#4A4426)"><span style="width: 38px; height: 38px; flex: none; border-radius: 50%; background: var(--k-3A3416,#3A3416); color: var(--k-FFD66B,#FFD66B); display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M8 4h8v5a4 4 0 0 1-8 0z"></path><path d="M8 6H5a3 3 0 0 0 3 4M16 6h3a3 3 0 0 1-3 4M12 13v4M8 20h8M10 17h4"></path></svg></span><span style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 13px; color: var(--k-F3EBC8,#F3EBC8)">{{ pa.type }}</strong><span style="font-size: 12px; color: var(--k-C9BE8F,#C9BE8F)">{{ pa.citation }}</span></span><span style="flex: none; font-size: 11px; color: var(--k-C9BE8F,#C9BE8F)">{{ pa.when }}</span></div></template></div>
</template>
</div>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'PersonDrawer',
  ...hubBindings(["closePerson","personOpen","psn"])
};
</script>
