<!-- Widget settings drawer -->
<template><template v-if="settingsOpen">
<div class="hub-overlay hub-right-drawer" style="position: fixed; inset: 0; z-index: 65; background: rgba(5,10,8,0.6); display: flex; justify-content: flex-end">
<div role="dialog" aria-label="Widget settings" style="width: 540px; max-width: 100%; height: 100%; background: var(--k-16201C,#16201C); border-left: 1px solid var(--k-2A3832,#2A3832); box-shadow: -20px 0 50px rgba(0,0,0,0.4); display: flex; flex-direction: column">
<div style="flex: none; display: flex; align-items: center; gap: 12px; padding: 18px 20px; border-bottom: 1px solid var(--k-24322C,#24322C)">
<span :style="`width: 12px; height: 12px; border-radius: 50%; background: ${sw.tint}`"></span>
<div style="flex: 1; min-width: 0"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Widget settings</span><h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 22px">{{ sw.baseTitle }}</h2></div>
<button aria-label="Close settings" @click="closeSettings" style="width: 40px; height: 40px; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 20px; display: flex; flex-direction: column; gap: 22px">

<div style="display: flex; flex-direction: column; gap: 14px">
<button type="button" role="switch" :aria-checked="sw.orgOn" @click="sw.toggleOrgOn" :style="`display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 0 14px; border-radius: 12px; cursor: pointer; text-align: left; font-size: 14px; font-weight: 600; color: var(--k-E8EFEA,#E8EFEA); background: var(--k-121B17,#121B17); border: 2px solid ${sw.orgOn ? '#7BB93A' : 'var(--k-6B7A72,#6B7A72)'}`"><span :style="`flex: none; width: 36px; height: 20px; border-radius: 10px; position: relative; background: ${sw.orgOn ? '#B8F36B' : 'var(--k-3F524A,#3F524A)'}`"><span :style="`position: absolute; top: 2px; ${sw.orgOn ? 'right: 2px' : 'left: 2px'}; width: 16px; height: 16px; border-radius: 50%; background: #0F1714`"></span></span>{{ sw.orgOnLabel }}</button>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Widget name<input type="text" maxlength="60" :value="sw.title" @input="sw.setTitle" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<div style="display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Who manages it <span style="color: var(--k-9DB0A6,#9DB0A6)">(roles that may change these settings)</span></span>
<div v-if="sw.editManagers" style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="mr in sw.managerRoles" :key="mr.name"><button type="button" :aria-pressed="mr.on" :title="mr.title" @click="mr.toggle" :disabled="mr.locked" :style="mr.style">{{ mr.name }}</button></template></div>
<span v-else style="font-size: 14px">{{ sw.managerNames }}</span>
</div>
<label v-if="sw.items" style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Items to show on Home<select :value="sw.items.value" @change="sw.items.set" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><template v-for="o in sw.items.options" :key="o"><option :value="o">{{ o }}</option></template></select></label>
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding: 12px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C)">
<span style="flex: 1 1 220px; font-size: 13px; line-height: 1.5; color: var(--k-C9D6CF,#C9D6CF)">{{ sw.seeHint }}</span>
<button v-if="sw.canRights" type="button" @click="sw.openRights" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px">Open rights</button>
</div>
</div>
</div>
<div style="flex: none; display: flex; justify-content: flex-end; gap: 8px; padding: 14px 20px; border-top: 1px solid var(--k-24322C,#24322C)">
<button @click="closeSettings" style="height: 42px; padding: 0 18px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Cancel</button>
<button @click="saveSettings" style="height: 42px; padding: 0 22px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Save settings</button>
</div>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'WidgetSettingsDrawer',
  ...hubBindings(["closeSettings","saveSettings","settingsOpen","sw"])
};
</script>
