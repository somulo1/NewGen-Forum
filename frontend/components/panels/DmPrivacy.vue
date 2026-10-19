<!-- Privacy for one private chat: block a person, or stop chats, calls or video from them -->
<template><template v-if="dmPriv.open">
<div role="region" :aria-label="`Privacy for ${dmPriv.name}`" style="margin: 10px 14px 0; padding: 14px; border-radius: 14px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); display: flex; flex-direction: column; gap: 10px">
<div style="display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 14px">Privacy for {{ dmPriv.name }}</strong><span :style="`font-size: 12px; color: ${dmPriv.anyOn ? 'var(--k-FFB86B,#FFB86B)' : 'var(--k-9DB0A6,#9DB0A6)'}`">{{ dmPriv.summary }}</span></div>
<template v-for="(fl, fl__i) in dmPriv.flags" :key="fl.key">
<label :style="`display: flex; align-items: flex-start; gap: 10px; cursor: ${fl.locked ? 'default' : 'pointer'}; opacity: ${fl.locked ? 0.6 : 1}`"><input type="checkbox" :checked="fl.on" :disabled="fl.locked" @change="fl.toggle" style="margin-top: 3px; accent-color: var(--k-B8F36B,#B8F36B)"><span style="display: flex; flex-direction: column; gap: 1px"><span style="font-size: 13px; font-weight: 600">{{ fl.label }}</span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ fl.desc }}</span></span></label>
</template>
<div style="display: flex; flex-direction: column; gap: 6px; padding-top: 4px; border-top: 1px solid var(--k-34453D,#34453D)">
<span style="font-size: 11px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">See what happens when they reach you</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(tr, tr__i) in dmPriv.tries" :key="tr__i"><button @click="tr.run" style="height: 30px; padding: 0 12px; border-radius: 15px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 12px">{{ tr.label }}</button></template></div>
</div>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'DmPrivacy',
  ...hubBindings(["dmPriv"])
};
</script>
